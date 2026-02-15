package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// TUI core types are in tui.go
// Key bindings are in keys.go
// Update helpers are in update.go
// View functions are in view.go

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.help.Width = msg.Width
		m.width = msg.Width
		m.height = msg.Height
		// Adjust table width based on available space (filter panel takes 20 chars)
		tableWidth := msg.Width - 25 // Leave room for filter panel + padding
		m.table.SetWidth(tableWidth)
		m.table.SetHeight(msg.Height - 5) // Leave room for help
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.defaultKeys.Help):
			if !m.adding && !m.editing && !m.confirmingDelete && !m.confirmingRestart {
				m.help.ShowAll = !m.help.ShowAll
			}
			return m, nil
		}

		if m.adding || m.editing {
			// Handle form input with textinput components
			var cmd tea.Cmd

			switch msg.String() {
			case "tab", "shift+tab":
				// Toggle focus between name and duration inputs
				if m.nameInput.Focused() {
					m.nameInput.Blur()
					m.durationInput.Focus()
				} else {
					m.durationInput.Blur()
					m.nameInput.Focus()
				}
				return m, nil

			case "enter":
				// Validate and submit
				name := m.nameInput.Value()
				durationStr := m.durationInput.Value()

				// Name is required
				if name == "" {
					return m, nil
				}

				// Validate duration
				duration, err := parseDuration(durationStr)
				if err != nil {
					return m, nil
				}

				if m.editing {
					// Update existing timer
					m.timers[m.editingIndex].Name = name
					m.timers[m.editingIndex].End = time.Now().Add(duration)
					m.timers[m.editingIndex].Duration = duration
					m.timers[m.editingIndex].Paused = false
					m.timers[m.editingIndex].Remaining = 0
				} else {
					// Add new timer
					newTimer := Timer{
						Name:     name,
						End:      time.Now().Add(duration),
						Duration: duration,
					}
					m.timers = append(m.timers, newTimer)
					visibleTimers := m.getVisibleTimers()
					m.cursor = len(visibleTimers) - 1
				}
				m.dirty = true

				// Reset and close form
				m.adding = false
				m.editing = false
				m.nameInput.Reset()
				m.durationInput.Reset()
				m.nameInput.Focus()
				return m, tick()

			case "esc":
				// Cancel and close form
				m.adding = false
				m.editing = false
				m.nameInput.Reset()
				m.durationInput.Reset()
				m.nameInput.Focus()
				return m, nil

			default:
				// Update the focused input
				if m.nameInput.Focused() {
					m.nameInput, cmd = m.nameInput.Update(msg)
				} else {
					m.durationInput, cmd = m.durationInput.Update(msg)
				}
				return m, cmd
			}
		}

		// Block all keys except y/n/esc/enter when confirming delete, restart, or bulk action
		if m.confirmingDelete || m.confirmingRestart || m.confirmingBulk {
			switch msg.String() {
			case "y", "Y", "n", "N", "d", "esc", "enter":
				// These keys are handled below
			default:
				return m, nil
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			if m.dirty {
				_ = saveToFile(m)
				// Update lastModTime to avoid triggering reload on our own save
				if info, err := os.Stat(saveFile); err == nil {
					m.lastModTime = info.ModTime()
				}
			}
			return m, tea.Quit

		case "p":
			actualIdx := m.getActualTimerIndex(m.cursor)
			if actualIdx >= 0 && len(m.timers) > 0 {
				t := &m.timers[actualIdx]
				if !t.Paused {
					// Pause: only if timer is still running
					if t.End.After(time.Now()) {
						t.Remaining = time.Until(t.End)
						t.Paused = true
						m.dirty = true
					}
				} else {
					// Resume: always allow if we have remaining time
					if t.Remaining > 0 {
						t.End = time.Now().Add(t.Remaining)
						t.Paused = false
						m.dirty = true
					}
				}
			}
			return m, nil

		case "up", "k":
			visibleTimers := m.getVisibleTimers()
			if m.cursor > 0 {
				m.cursor--
				m.table.SetCursor(m.cursor)
			} else if len(visibleTimers) == 0 {
				m.cursor = 0
				m.table.SetCursor(0)
			}
			return m, nil

		case "down", "j":
			visibleTimers := m.getVisibleTimers()
			if m.cursor < len(visibleTimers)-1 {
				m.cursor++
				m.table.SetCursor(m.cursor)
			}
			return m, nil

		case "ctrl+k", "ctrl+up":
			actualIdx := m.getActualTimerIndex(m.cursor)
			if actualIdx > 0 {
				// Swap with previous timer
				m.timers[actualIdx-1], m.timers[actualIdx] = m.timers[actualIdx], m.timers[actualIdx-1]
				if m.cursor > 0 {
					m.cursor--
				}
				m.dirty = true
			}
			return m, nil

		case "ctrl+j", "ctrl+down":
			actualIdx := m.getActualTimerIndex(m.cursor)
			if actualIdx < len(m.timers)-1 && actualIdx >= 0 {
				// Swap with next timer
				m.timers[actualIdx], m.timers[actualIdx+1] = m.timers[actualIdx+1], m.timers[actualIdx]
				visibleTimers := m.getVisibleTimers()
				if m.cursor < len(visibleTimers)-1 {
					m.cursor++
				}
				m.dirty = true
			}
			return m, nil

		case "d":
			actualIdx := m.getActualTimerIndex(m.cursor)
			if actualIdx < 0 || len(m.timers) == 0 {
				return m, nil
			}

			if m.confirmingDelete {
				// Confirm delete
				m.timers = append(
					m.timers[:actualIdx],
					m.timers[actualIdx+1:]...,
				)

				visibleTimers := m.getVisibleTimers()
				if m.cursor >= len(visibleTimers) && m.cursor > 0 {
					m.cursor--
				}

				m.confirmingDelete = false
				m.dirty = true
			} else {
				// Show confirmation
				m.confirmingDelete = true
			}
			return m, nil

		case "y", "Y", "enter":
			if m.confirmingDelete {
				actualIdx := m.getActualTimerIndex(m.cursor)
				// Confirm delete
				if actualIdx >= 0 && len(m.timers) > 0 {
					m.timers = append(
						m.timers[:actualIdx],
						m.timers[actualIdx+1:]...,
					)

					visibleTimers := m.getVisibleTimers()
					if m.cursor >= len(visibleTimers) && m.cursor > 0 {
						m.cursor--
					}

					m.confirmingDelete = false
					m.dirty = true
				}
			}
			if m.confirmingRestart {
				actualIdx := m.getActualTimerIndex(m.cursor)
				// Confirm restart
				if actualIdx >= 0 && len(m.timers) > 0 && m.timers[actualIdx].Duration > 0 {
					t := &m.timers[actualIdx]
					t.End = m.now.Add(t.Duration)
					t.Paused = false
					t.Remaining = 0
					m.confirmingRestart = false
					m.dirty = true
					return m, tick()
				}
			}
			if m.confirmingBulk {
				// Execute bulk action
				switch m.pendingBulkAction {
				case bulkPauseAll:
					count := 0
					for i := range m.timers {
						if !m.timers[i].Paused && m.timers[i].End.After(m.now) {
							m.timers[i].Remaining = time.Until(m.timers[i].End)
							m.timers[i].Paused = true
							count++
						}
					}
					if count > 0 {
						m.dirty = true
					}
				case bulkResumeAll:
					count := 0
					for i := range m.timers {
						if m.timers[i].Paused && m.timers[i].Remaining > 0 {
							m.timers[i].End = m.now.Add(m.timers[i].Remaining)
							m.timers[i].Paused = false
							count++
						}
					}
					if count > 0 {
						m.dirty = true
					}
				case bulkDeleteDone:
					newTimers := make([]Timer, 0, len(m.timers))
					for _, t := range m.timers {
						if t.Paused || t.End.After(m.now) {
							newTimers = append(newTimers, t)
						} else {
							m.dirty = true
						}
					}
					m.timers = newTimers
					// Adjust cursor if needed
					visibleTimers := m.getVisibleTimers()
					if m.cursor >= len(visibleTimers) && m.cursor > 0 {
						m.cursor = len(visibleTimers) - 1
					}
				case bulkRestartAll:
					count := 0
					for i := range m.timers {
						if m.timers[i].Duration > 0 {
							m.timers[i].End = m.now.Add(m.timers[i].Duration)
							m.timers[i].Paused = false
							m.timers[i].Remaining = 0
							count++
						}
					}
					if count > 0 {
						m.dirty = true
					}
				}
				m.confirmingBulk = false
			}
			return m, nil

		case "n", "N", "esc":
			if m.confirmingDelete {
				m.confirmingDelete = false
			}
			if m.confirmingRestart {
				m.confirmingRestart = false
			}
			if m.confirmingBulk {
				m.confirmingBulk = false
			}
			return m, nil

		case "P":
			if !m.adding && !m.editing && !m.confirmingDelete && !m.confirmingRestart {
				m.confirmingBulk = true
				m.pendingBulkAction = bulkPauseAll
			}
			return m, nil

		case "R":
			if !m.adding && !m.editing && !m.confirmingDelete && !m.confirmingRestart {
				m.confirmingBulk = true
				m.pendingBulkAction = bulkRestartAll
			}
			return m, nil

		case "shift+R":
			if !m.adding && !m.editing && !m.confirmingDelete && !m.confirmingRestart {
				m.confirmingBulk = true
				m.pendingBulkAction = bulkResumeAll
			}
			return m, nil

		case "D":
			if !m.adding && !m.editing && !m.confirmingDelete && !m.confirmingRestart {
				m.confirmingBulk = true
				m.pendingBulkAction = bulkDeleteDone
			}
			return m, nil

		case "a":
			if m.adding || m.editing {
				return m, nil
			}

			m.adding = true
			m.nameInput.Reset()
			m.durationInput.Reset()
			m.nameInput.Focus()
			m.durationInput.Blur()
			return m, nil

		case "r":
			actualIdx := m.getActualTimerIndex(m.cursor)
			if actualIdx < 0 || len(m.timers) == 0 || m.timers[actualIdx].Duration == 0 {
				return m, nil
			}

			if m.confirmingRestart {
				// Confirm restart
				t := &m.timers[actualIdx]
				t.End = time.Now().Add(t.Duration)
				t.Paused = false
				t.Remaining = 0
				m.confirmingRestart = false
				m.dirty = true
				return m, tick()
			} else {
				// Show confirmation
				m.confirmingRestart = true
			}
			return m, nil

		case "e":
			actualIdx := m.getActualTimerIndex(m.cursor)
			if actualIdx >= 0 && len(m.timers) > 0 {
				m.editing = true
				m.editingIndex = actualIdx
				m.nameInput.Reset()
				m.durationInput.Reset()
				m.nameInput.SetValue(m.timers[actualIdx].Name)
				m.durationInput.SetValue(formatDuration(m.timers[actualIdx].Duration))
				m.nameInput.Focus()
				m.durationInput.Blur()
			}
			return m, nil

		case "tab":
			m.filter = (m.filter + 1) % 4
			visibleTimers := m.getVisibleTimers()
			if len(visibleTimers) == 0 {
				m.cursor = 0
			} else if m.cursor >= len(visibleTimers) {
				m.cursor = len(visibleTimers) - 1
			}
			return m, nil

		case "1":
			m.filter = filterAll
			m.cursor = 0
			m.table.SetCursor(0)
			m.table.GotoTop()
			return m, nil

		case "2":
			m.filter = filterActive
			m.cursor = 0
			m.table.SetCursor(0)
			m.table.GotoTop()
			return m, nil

		case "3":
			m.filter = filterPaused
			m.cursor = 0
			m.table.SetCursor(0)
			m.table.GotoTop()
			return m, nil

		case "4":
			m.filter = filterDone
			m.cursor = 0
			m.table.SetCursor(0)
			m.table.GotoTop()
			return m, nil

		}

	case tickMsg:
		m.now = time.Time(msg)
		return m, tea.Batch(tick(), fileWatchTick())

	case fileWatchMsg:
		// Check if save file has been modified externally
		if info, err := os.Stat(saveFile); err == nil {
			modTime := info.ModTime()
			if modTime.After(m.lastModTime) {
				// File was modified externally, reload timers
				if s, err := loadFromFile(); err == nil {
					applySaveData(&m, s)
					m.lastModTime = modTime
				}
			}
		}
		return m, nil
	}

	return m, nil
}

// Update helpers are in update.go
// View functions are in view.go
// TUI core is in tui.go

// CLI functions are in cli.go

func main() {
	// If no arguments provided (other than program name), run TUI
	if len(os.Args) < 2 {
		p := tea.NewProgram(initialModel(), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Println("error:", err)
		}
		return
	}

	// CLI mode: parse and execute commands
	cmd := os.Args[1]
	args := os.Args[2:]

	// Command aliases
	aliases := map[string]string{
		"a":   "add",
		"l":   "list",
		"p":   "pause",
		"r":   "resume",
		"d":   "delete",
		"rs":  "restart",
		"e":   "edit",
		"h":   "help",
	}

	// Resolve alias
	if fullCmd, ok := aliases[cmd]; ok {
		cmd = fullCmd
	}

	if err := executeCLICommand(cmd, args); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
