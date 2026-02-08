package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg time.Time

type Timer struct {
	Name      string        `json:"name"`
	End       time.Time     `json:"end"`
	Paused    bool          `json:"paused"`
	Remaining time.Duration `json:"remaining"`
	Duration  time.Duration `json:"duration"`
}

type inputField struct {
	label string
	value string
}

type model struct {
	timers          []Timer
	now             time.Time
	cursor          int

	adding          bool
	editing         bool  // true when editing existing timer
	confirmingDelete bool // true when showing delete confirmation
	inputFields     []inputField
	activeField     int

	dirty bool
}

func parseDuration(input string) (time.Duration, error) {
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "" {
		return 0, fmt.Errorf("empty input")
	}

	// Default to seconds if no suffix
	if len(input) > 0 && input[len(input)-1] >= '0' && input[len(input)-1] <= '9' {
		secs, err := strconv.Atoi(input)
		if err != nil {
			return 0, err
		}
		return time.Duration(secs) * time.Second, nil
	}

	// Parse number and suffix
	numStr := input[:len(input)-1]
	suffix := input[len(input)-1:]

	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, err
	}

	if num <= 0 {
		return 0, fmt.Errorf("duration must be positive")
	}

	switch suffix {
	case "s":
		return time.Duration(num) * time.Second, nil
	case "m":
		return time.Duration(num) * time.Minute, nil
	case "h":
		return time.Duration(num) * time.Hour, nil
	case "d":
		return time.Duration(num) * 24 * time.Hour, nil
	case "y":
		return time.Duration(num) * 365 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("invalid suffix: %s (use s, m, h, d, y)", suffix)
	}
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	totalSeconds := int(d.Seconds())

	days := totalSeconds / 86400
	hours := (totalSeconds % 86400) / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	// Calculate months and years for larger durations
	years := days / 365
	months := (days % 365) / 30
	remainingDays := (days % 365) % 30

	var parts []string

	if years > 0 {
		parts = append(parts, fmt.Sprintf("%dy", years))
	}
	if months > 0 {
		parts = append(parts, fmt.Sprintf("%dm", months))
	}
	if remainingDays > 0 {
		parts = append(parts, fmt.Sprintf("%dd", remainingDays))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	if seconds > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%ds", seconds))
	}

	if len(parts) > 2 {
		return strings.Join(parts[:2], " ") + " " + strings.Join(parts[2:], " ")
	}
	return strings.Join(parts, " ")
}

func (m model) Init() tea.Cmd {
	return tick()
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		if m.adding || m.editing {
			switch msg.String() {

			case "up", "shift+tab":
				if m.activeField > 0 {
					m.activeField--
				}
				return m, nil

			case "down", "tab":
				if m.activeField < len(m.inputFields)-1 {
					m.activeField++
				}
				return m, nil

			case "enter":
				// If not on last field, move to next field (but name is required)
				if m.activeField < len(m.inputFields)-1 {
					if m.activeField == 0 && m.inputFields[0].value == "" {
						// Name is empty, don't move to next field
						return m, nil
					}
					m.activeField++
					return m, nil
				}
				duration, err := parseDuration(m.inputFields[1].value)
				if err == nil && m.inputFields[0].value != "" {
					if m.editing {
						// Update existing timer
						m.timers[m.cursor].Name = m.inputFields[0].value
						m.timers[m.cursor].End = time.Now().Add(duration)
						m.timers[m.cursor].Duration = duration
						m.timers[m.cursor].Paused = false
						m.timers[m.cursor].Remaining = 0
					} else {
						// Add new timer
						newTimer := Timer{
							Name:     m.inputFields[0].value,
							End:      time.Now().Add(duration),
							Duration: duration,
						}
						m.timers = append(m.timers, newTimer)
						m.cursor = len(m.timers) - 1
					}
					m.dirty = true
				}

				m.adding = false
				m.editing = false
				for i := range m.inputFields {
					m.inputFields[i].value = ""
				}
				return m, tick()

			case "esc":
				m.adding = false
				m.editing = false
				for i := range m.inputFields {
					m.inputFields[i].value = ""
				}
				return m, nil

			case "backspace":
				if len(m.inputFields[m.activeField].value) > 0 {
					m.inputFields[m.activeField].value = m.inputFields[m.activeField].value[:len(m.inputFields[m.activeField].value)-1]
				}
				return m, nil

			default:
				// Handle space key for name field
				if msg.String() == " " && m.activeField == 0 {
					m.inputFields[m.activeField].value += " "
					return m, nil
				}
				if msg.Type == tea.KeyRunes {
					ch := msg.String()[0]
					switch m.activeField {
					case 0:
						// First field: any characters allowed
						m.inputFields[m.activeField].value += msg.String()
					case 1:
						// Duration field: numbers and suffix letters only
						if (ch >= '0' && ch <= '9') || ch == 's' || ch == 'm' || ch == 'h' || ch == 'd' || ch == 'y' {
							m.inputFields[m.activeField].value += msg.String()
						}
					}
				}
				return m, nil
			}
		}

		// Block all keys except y/n/esc when confirming delete
		if m.confirmingDelete {
			switch msg.String() {
			case "y", "Y", "n", "N", "d", "esc":
				// These keys are handled below
			default:
				return m, nil
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			if m.dirty {
				_ = saveToFile(m)
			}
			return m, tea.Quit

		case "p":
			if len(m.timers) > 0 {
				t := &m.timers[m.cursor]
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
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case "down", "j":
			if m.cursor < len(m.timers)-1 {
				m.cursor++
			}
			return m, nil

		case "d":
			if len(m.timers) == 0 {
				return m, nil
			}

			if m.confirmingDelete {
				// Confirm delete
				m.timers = append(
					m.timers[:m.cursor],
					m.timers[m.cursor+1:]...,
				)

				if m.cursor >= len(m.timers) && m.cursor > 0 {
					m.cursor--
				}

				m.confirmingDelete = false
				m.dirty = true
			} else {
				// Show confirmation
				m.confirmingDelete = true
			}
			return m, nil

		case "y", "Y":
			if m.confirmingDelete {
				// Confirm delete
				if len(m.timers) > 0 {
					m.timers = append(
						m.timers[:m.cursor],
						m.timers[m.cursor+1:]...,
					)

					if m.cursor >= len(m.timers) && m.cursor > 0 {
						m.cursor--
					}

					m.confirmingDelete = false
					m.dirty = true
				}
			}
			return m, nil

		case "n", "N", "esc":
			if m.confirmingDelete {
				m.confirmingDelete = false
			}
			return m, nil

		case "a":
			if m.adding || m.editing {
				return m, nil
			}

			m.adding = true
			m.inputFields = []inputField{
				{label: "Name", value: ""},
				{label: "Duration", value: ""},
			}
			m.activeField = 0
			return m, nil

		case "r":
			if len(m.timers) > 0 && m.timers[m.cursor].Duration > 0 {
				t := &m.timers[m.cursor]
				t.End = time.Now().Add(t.Duration)
				t.Paused = false
				t.Remaining = 0
				m.dirty = true
				return m, tick()
			}
			return m, nil

		case "e":
			if len(m.timers) > 0 {
				m.editing = true
				m.inputFields = []inputField{
					{label: "Name", value: m.timers[m.cursor].Name},
					{label: "Duration", value: formatDuration(m.timers[m.cursor].Duration)},
				}
				m.activeField = 0
			}
			return m, nil

		}

	case tickMsg:
		m.now = time.Time(msg)
		return m, tick()
	}

	return m, nil
}

func allDone(timers []Timer, now time.Time) bool {
	for _, t := range timers {
		if t.End.After(now) {
			return false
		}
	}
	return true
}

func activeCount(timers []Timer, now time.Time) int {
	n := 0
	for _, t := range timers {
		if t.Paused || t.End.After(now) {
			n++
		}
	}
	return n
}

func renderStatusBar(m model) string {
	return fmt.Sprintf(
		"%d/%d active\n\n",
		activeCount(m.timers, m.now),
		len(m.timers),
	)
}

func (m model) View() string {
	if m.confirmingDelete {
		return fmt.Sprintf(
			"🗑️  Delete Timer\n\n"+
				"Delete \"%s\"?\n\n"+
				"[y] yes  [n] no  [esc] cancel\n",
			m.timers[m.cursor].Name,
		)
	}

	if m.adding || m.editing {
		var b strings.Builder
		if m.editing {
			b.WriteString("✏️  Edit Timer\n\n")
		} else {
			b.WriteString("➕ Add Timer\n\n")
		}

		for i, field := range m.inputFields {
			cursor := " "
			if i == m.activeField {
				cursor = ">"
			}
			fmt.Fprintf(&b, "%s %s: %s\n", cursor, field.label, field.value)
		}

		b.WriteString("\n[tab/arrows] switch  [enter] next/confirm  [esc] cancel\n")
		b.WriteString("Duration: 30 = 30s, 5m, 1h, 2d, 1y\n")
		return b.String()
	}

	var b strings.Builder

	b.WriteString(renderStatusBar(m))

	for i, t := range m.timers {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		var remaining time.Duration
		if t.Paused {
			remaining = t.Remaining
		} else {
			remaining = time.Until(t.End)
		}

		if remaining <= 0 && !t.Paused {
			fmt.Fprintf(&b, "%s✅ %s done (%s)\n", cursor, t.Name, formatDuration(t.Duration))
		} else if t.Paused {
			fmt.Fprintf(&b, "%s⏸ %s: %s (paused)\n", cursor, t.Name, formatDuration(remaining))
		} else {
			var endTime string
			if t.End.Day() == m.now.Day() && t.End.Month() == m.now.Month() && t.End.Year() == m.now.Year() {
				endTime = t.End.Format("15:04:05")
			} else if t.End.Month() == m.now.Month() && t.End.Year() == m.now.Year() {
				endTime = t.End.Format("2 15:04:05")
			} else if t.End.Year() == m.now.Year() {
				endTime = t.End.Format("2/01 15:04:05")
			} else {
				endTime = t.End.Format("2/01/06 15:04:05")
			}
			fmt.Fprintf(&b, "%s⏳ %s: %s (ends at %s)\n", cursor, t.Name, formatDuration(remaining), endTime)
		}
	}

	b.WriteString("\n[a] add  [d] delete  [e] edit  [r] redo  [↑/↓] move  [p] pause  [q] quit\n")
	return b.String()
}

func initialModel() model {
	m := model{
		now: time.Now(),
	}

	if s, err := loadFromFile(); err == nil {
		applySaveData(&m, s)
	}

	return m
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("error:", err)
	}
}
