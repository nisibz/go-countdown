package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tickMsg time.Time
type fileWatchMsg struct{}

// defaultKeyMap defines keybindings for the main timer view
type filterMode int

const (
	filterAll filterMode = iota
	filterActive
	filterPaused
	filterDone
)

type defaultKeyMap struct {
	Up        key.Binding
	Down      key.Binding
	UpOrder   key.Binding
	DownOrder key.Binding
	Add       key.Binding
	Delete    key.Binding
	Edit      key.Binding
	Redo      key.Binding
	Pause     key.Binding
	Filter1   key.Binding
	Filter2   key.Binding
	Filter3   key.Binding
	Filter4   key.Binding
	Help      key.Binding
	Quit      key.Binding
}

// ShortHelp returns keybindings for the mini help view
func (k defaultKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the full help view
func (k defaultKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.UpOrder, k.DownOrder},
		{k.Add, k.Delete, k.Edit, k.Redo, k.Pause},
		{k.Filter1, k.Filter2, k.Filter3, k.Filter4},
		{k.Help, k.Quit},
	}
}

func newDefaultKeyMap() defaultKeyMap {
	return defaultKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		UpOrder: key.NewBinding(
			key.WithKeys("ctrl+up", "ctrl+k"),
			key.WithHelp("ctrl+↑", "reorder up"),
		),
		DownOrder: key.NewBinding(
			key.WithKeys("ctrl+down", "ctrl+j"),
			key.WithHelp("ctrl+↓", "reorder down"),
		),
		Add: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add timer"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete timer"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit timer"),
		),
		Redo: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "restart timer"),
		),
		Pause: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "pause/resume"),
		),
		Filter1: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "show all"),
		),
		Filter2: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "show active"),
		),
		Filter3: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "show paused"),
		),
		Filter4: key.NewBinding(
			key.WithKeys("4"),
			key.WithHelp("4", "show done"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

// formKeyMap defines keybindings for adding/editing timers
type formKeyMap struct {
	NextField key.Binding
	PrevField key.Binding
	Enter     key.Binding
	Esc       key.Binding
	Help      key.Binding
}

// ShortHelp returns keybindings for the mini help view
func (k formKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Enter, k.Esc}
}

// FullHelp returns keybindings for the full help view
func (k formKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.NextField, k.PrevField},
		{k.Enter, k.Esc},
	}
}

func newFormKeyMap() formKeyMap {
	return formKeyMap{
		NextField: key.NewBinding(
			key.WithKeys("tab", "down"),
			key.WithHelp("tab/↓", "next field"),
		),
		PrevField: key.NewBinding(
			key.WithKeys("shift+tab", "up"),
			key.WithHelp("↑/shift+tab", "prev field"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm/next"),
		),
		Esc: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
	}
}

// confirmKeyMap defines keybindings for delete confirmation
type confirmKeyMap struct {
	ConfirmYes key.Binding
	ConfirmNo  key.Binding
	Esc        key.Binding
}

// ShortHelp returns keybindings for the mini help view
func (k confirmKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.ConfirmYes, k.ConfirmNo}
}

// FullHelp returns keybindings for the full help view
func (k confirmKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.ConfirmYes, k.ConfirmNo, k.Esc},
	}
}

func newConfirmKeyMap() confirmKeyMap {
	return confirmKeyMap{
		ConfirmYes: key.NewBinding(
			key.WithKeys("y", "Y", "enter"),
			key.WithHelp("y/enter", "yes, delete"),
		),
		ConfirmNo: key.NewBinding(
			key.WithKeys("n", "N"),
			key.WithHelp("n", "no, cancel"),
		),
		Esc: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}
}

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
	timers []Timer
	now    time.Time
	cursor int
	table  table.Model

	adding           bool
	editing          bool // true when editing existing timer
	editingIndex     int  // actual index of timer being edited
	confirmingDelete bool // true when showing delete confirmation
	inputFields      []inputField
	activeField      int

	filter filterMode

	dirty           bool
	lastModTime     time.Time // track file modification time for external changes
	defaultKeys defaultKeyMap // key bindings for default view
	formKeys    formKeyMap    // key bindings for form mode
	confirmKeys confirmKeyMap // key bindings for delete confirmation
	help        help.Model    // help component
}

func parseDuration(input string) (time.Duration, error) {
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "" {
		return 0, fmt.Errorf("empty input")
	}

	var total time.Duration

	// Parse multiple number-suffix pairs (e.g., "30d30m", "1h30m", "2d")
	i := 0
	for i < len(input) {
		// Skip spaces between components
		if input[i] == ' ' {
			i++
			continue
		}

		// Parse number
		numStart := i
		for i < len(input) && input[i] >= '0' && input[i] <= '9' {
			i++
		}
		if i == numStart {
			return 0, fmt.Errorf("expected number at position %d", numStart)
		}
		numStr := input[numStart:i]

		num, err := strconv.Atoi(numStr)
		if err != nil {
			return 0, fmt.Errorf("invalid number %s: %w", numStr, err)
		}

		if num <= 0 {
			return 0, fmt.Errorf("duration must be positive")
		}

		// Parse suffix (default to seconds if at end of input)
		var suffix string
		if i < len(input) {
			suffix = input[i : i+1]
			i++
		} else {
			// No suffix means seconds (e.g., "30" = 30s)
			suffix = "s"
		}

		var d time.Duration
		switch suffix {
		case "s":
			d = time.Duration(num) * time.Second
		case "m":
			d = time.Duration(num) * time.Minute
		case "h":
			d = time.Duration(num) * time.Hour
		case "d":
			d = time.Duration(num) * 24 * time.Hour
		case "y":
			d = time.Duration(num) * 365 * 24 * time.Hour
		default:
			return 0, fmt.Errorf("invalid suffix: %s (use s, m, h, d, y)", suffix)
		}
		total += d
	}

	if total <= 0 {
		return 0, fmt.Errorf("duration must be positive")
	}

	return total, nil
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
	remainingDaysAfterYears := days % 365
	months := remainingDaysAfterYears / 30
	remainingDays := remainingDaysAfterYears % 30

	var parts []string

	// For large durations (> 60 days), show only top 2 units (years, months, or days)
	// This keeps the display compact and prevents truncation
	if days > 60 {
		if years > 0 {
			parts = append(parts, fmt.Sprintf("%dy", years))
		}
		if months > 0 {
			parts = append(parts, fmt.Sprintf("%dmo", months))
		}
		if remainingDays > 0 && len(parts) < 2 {
			parts = append(parts, fmt.Sprintf("%dd", remainingDays))
		}
		// Show at most 2 parts for long durations
		if len(parts) > 2 {
			parts = parts[:2]
		}
		return strings.Join(parts, " ")
	}

	// For shorter durations, show more detail
	if years > 0 {
		parts = append(parts, fmt.Sprintf("%dy", years))
	}
	if months > 0 {
		parts = append(parts, fmt.Sprintf("%dmo", months))
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
	return tea.Batch(tick(), fileWatchTick())
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func fileWatchTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return fileWatchMsg{}
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.help.Width = msg.Width
		// Adjust table width based on available space (filter panel takes 20 chars)
		tableWidth := msg.Width - 25 // Leave room for filter panel + padding
		m.table.SetWidth(tableWidth)
		m.table.SetHeight(msg.Height - 5) // Leave room for help
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.defaultKeys.Help):
			if !m.adding && !m.editing && !m.confirmingDelete {
				m.help.ShowAll = !m.help.ShowAll
			}
			return m, nil
		}

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
						m.timers[m.editingIndex].Name = m.inputFields[0].value
						m.timers[m.editingIndex].End = time.Now().Add(duration)
						m.timers[m.editingIndex].Duration = duration
						m.timers[m.editingIndex].Paused = false
						m.timers[m.editingIndex].Remaining = 0
					} else {
						// Add new timer
						newTimer := Timer{
							Name:     m.inputFields[0].value,
							End:      time.Now().Add(duration),
							Duration: duration,
						}
						m.timers = append(m.timers, newTimer)
						visibleTimers := m.getVisibleTimers()
						m.cursor = len(visibleTimers) - 1
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

		// Block all keys except y/n/esc/enter when confirming delete
		if m.confirmingDelete {
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
			actualIdx := m.getActualTimerIndex(m.cursor)
			if m.confirmingDelete {
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
			actualIdx := m.getActualTimerIndex(m.cursor)
			if actualIdx >= 0 && len(m.timers) > 0 && m.timers[actualIdx].Duration > 0 {
				t := &m.timers[actualIdx]
				t.End = time.Now().Add(t.Duration)
				t.Paused = false
				t.Remaining = 0
				m.dirty = true
				return m, tick()
			}
			return m, nil

		case "e":
			actualIdx := m.getActualTimerIndex(m.cursor)
			if actualIdx >= 0 && len(m.timers) > 0 {
				m.editing = true
				m.editingIndex = actualIdx
				m.inputFields = []inputField{
					{label: "Name", value: m.timers[actualIdx].Name},
					{label: "Duration", value: formatDuration(m.timers[actualIdx].Duration)},
				}
				m.activeField = 0
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

func (m model) getVisibleTimers() []Timer {
	var result []Timer
	for _, t := range m.timers {
		switch m.filter {
		case filterAll:
			result = append(result, t)
		case filterActive:
			if !t.Paused && t.End.After(m.now) {
				result = append(result, t)
			}
		case filterPaused:
			if t.Paused {
				result = append(result, t)
			}
		case filterDone:
			if !t.Paused && !t.End.After(m.now) {
				result = append(result, t)
			}
		}
	}
	return result
}

func (m model) getActualTimerIndex(visibleIndex int) int {
	visibleTimers := m.getVisibleTimers()
	if visibleIndex < 0 || visibleIndex >= len(visibleTimers) {
		return -1
	}
	targetTimer := visibleTimers[visibleIndex]
	for i, t := range m.timers {
		if t.Name == targetTimer.Name && t.End.Equal(targetTimer.End) {
			return i
		}
	}
	return -1
}

func renderFilterPanel(m model) string {
	filters := []struct {
		num   string
		label string
		mode  filterMode
	}{
		{"1", "All", filterAll},
		{"2", "Active", filterActive},
		{"3", "Paused", filterPaused},
		{"4", "Done", filterDone},
	}

	var b strings.Builder
	b.WriteString(" Filters\n")
	for _, f := range filters {
		prefix := "  "
		if m.filter == f.mode {
			prefix = "▶️"
		}
		fmt.Fprintf(&b, "%s %s %s\n", prefix, f.num, f.label)
	}
	return b.String()
}

// updateTableRows populates the table with timer data
func updateTableRows(m *model) {
	visibleTimers := m.getVisibleTimers()

	var rows []table.Row
	for _, t := range visibleTimers {
		var status string
		var remaining time.Duration
		var remainingText string
		var endTimeText string

		// Determine status and calculate remaining
		if t.Paused {
			status = "⏸️"
			remaining = t.Remaining
			remainingText = formatDuration(remaining)
			endTimeText = "(paused)"
		} else {
			remaining = time.Until(t.End)
			if remaining <= 0 {
				status = "✅"
				remainingText = "Done"
				elapsed := t.Duration - remaining
				endTimeText = fmt.Sprintf("+%s", formatDuration(elapsed))
			} else {
				status = "⏳️"
				remainingText = formatDuration(remaining)
				endTimeText = formatEndTime(t.End, m.now)
			}
		}

		// Truncate name if too long
		name := t.Name
		if len(name) > 20 {
			name = name[:19] + "…"
		}

		row := table.Row{status, name, remainingText, endTimeText}
		rows = append(rows, row)
	}

	m.table.SetRows(rows)

	// Sync cursor position
	if m.cursor >= 0 && m.cursor < len(rows) {
		m.table.SetCursor(m.cursor)
	}
}

// formatEndTime formats the end time based on how far in the future it is
func formatEndTime(end, now time.Time) string {
	if end.Day() == now.Day() && end.Month() == now.Month() && end.Year() == now.Year() {
		return end.Format("15:04:05")
	} else if end.Month() == now.Month() && end.Year() == now.Year() {
		return end.Format("2 15:04")
	} else if end.Year() == now.Year() {
		return end.Format("2/01 15:04")
	} else {
		return end.Format("2/01/06 15:04")
	}
}

func (m model) View() string {
	if m.confirmingDelete {
		actualIdx := m.getActualTimerIndex(m.cursor)
		return fmt.Sprintf(
			"🗑️  Delete Timer\n\n"+
				"Delete \"%s\"?\n\n"+
				"%s",
			m.timers[actualIdx].Name,
			m.help.View(m.confirmKeys),
		)
	}

	if m.adding || m.editing {
		var b strings.Builder
		if m.editing {
			b.WriteString("✏️  Edit Timer\n\n")
		} else {
			b.WriteString("➕️ Add Timer\n\n")
		}

		for i, field := range m.inputFields {
			cursor := "  "
			if i == m.activeField {
				cursor = "▶️"
			}
			fmt.Fprintf(&b, "%s %s: %s\n", cursor, field.label, field.value)
		}

		b.WriteString("\n" + m.help.View(m.formKeys))
		b.WriteString("\nDuration: 30s, 5m, 1h, 2d, 1y (e.g., 30d30m)\n")
		return b.String()
	}

	// Build filter panel
	filterPanel := renderFilterPanel(m)

	// Update table rows with current timer data
	updateTableRows(&m)

	// Build timer table
	timerTable := m.table.View()

	// Combine filter panel and table side by side
	filterLines := strings.Split(filterPanel, "\n")
	timerLines := strings.Split(timerTable, "\n")

	var b strings.Builder
	maxFilterLines := len(filterLines)
	for i := 0; i < maxFilterLines || i < len(timerLines); i++ {
		if i < len(filterLines) {
			fmt.Fprintf(&b, "%-20s", filterLines[i])
		} else {
			b.WriteString(strings.Repeat(" ", 20))
		}
		if i < len(timerLines) {
			b.WriteString(timerLines[i])
		}
		if i < maxFilterLines-1 || i < len(timerLines)-1 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n" + m.help.View(m.defaultKeys))
	return b.String()
}

func initialModel() model {
	// Define table columns
	columns := []table.Column{
		{Title: "Stat", Width: 6},
		{Title: "Name", Width: 22},
		{Title: "Remaining", Width: 17},
		{Title: "End Time", Width: 17},
	}

	// Create table with styles
	tbl := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10), // Will be dynamic based on viewport
	)

	// Set table styles
	s := table.DefaultStyles()
	s.Header = s.Header.
		Foreground(lipgloss.Color("15")).
		Bold(true).
		Underline(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("57"))
	tbl.SetStyles(s)

	m := model{
		now:         time.Now(),
		filter:      filterAll,
		defaultKeys: newDefaultKeyMap(),
		formKeys:    newFormKeyMap(),
		confirmKeys: newConfirmKeyMap(),
		help:        help.New(),
		table:       tbl,
	}

	if s, err := loadFromFile(); err == nil {
		applySaveData(&m, s)
	}

	// Get initial file modification time
	if info, err := os.Stat(saveFile); err == nil {
		m.lastModTime = info.ModTime()
	}

	return m
}

// CLI helper functions

func printUsage() {
	fmt.Println("go-countdown - Terminal countdown timer")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  go-countdown              # Launch TUI interface")
	fmt.Println("  go-countdown <command>    # Run CLI command")
	fmt.Println()
	fmt.Println("COMMANDS:")
	fmt.Println("  add <name> <duration>           Add a new timer")
	fmt.Println("  list [--filter]                 List timers (filter: --active, --paused, --done)")
	fmt.Println("  pause [filter] <index>          Pause timer by 1-based index")
	fmt.Println("  resume [filter] <index>         Resume timer by 1-based index")
	fmt.Println("  delete [filter] <index>         Delete timer by 1-based index")
	fmt.Println("  restart [filter] <index>        Restart timer by 1-based index")
	fmt.Println("  edit [filter] <index> <name> <duration>  Edit timer")
	fmt.Println("  help                            Show this help")
	fmt.Println()
	fmt.Println("DURATION FORMAT:")
	fmt.Println("  30s    30 seconds")
	fmt.Println("  5m     5 minutes")
	fmt.Println("  1h     1 hour")
	fmt.Println("  2d     2 days")
	fmt.Println("  1y     1 year")
	fmt.Println("  30d30m Compound: 30 days 30 minutes")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  go-countdown add \"Meeting\" 30m")
	fmt.Println("  go-countdown add \"Project deadline\" 2d")
	fmt.Println("  go-countdown list                    # List all timers")
	fmt.Println("  go-countdown list --active           # List only active timers")
	fmt.Println("  go-countdown pause 1                 # Pause first timer (from all)")
	fmt.Println("  go-countdown pause --active 1        # Pause first active timer")
	fmt.Println("  go-countdown resume --paused 2       # Resume second paused timer")
	fmt.Println("  go-countdown delete --done 1         # Delete first done timer")
	fmt.Println("  go-countdown delete 2                # Delete second timer (from all)")
	fmt.Println("  go-countdown restart 1               # Restart first timer")
}

func parseFilterAndIndex(args []string) (filter, indexStr string, idx int) {
	if len(args) == 0 {
		return "", "", 0
	}
	if strings.HasPrefix(args[0], "--") {
		filter = args[0]
		if len(args) < 2 {
			return filter, "", 0
		}
		indexStr = args[1]
	} else {
		indexStr = args[0]
	}
	var err error
	idx, err = strconv.Atoi(indexStr)
	if err != nil || idx < 1 {
		return filter, indexStr, 0
	}
	return filter, indexStr, idx
}

func getFilteredTimers(timers []Timer, filter string) []Timer {
	now := time.Now()
	var result []Timer
	for _, t := range timers {
		switch filter {
		case "--active":
			if !t.Paused && t.End.After(now) {
				result = append(result, t)
			}
		case "--paused":
			if t.Paused {
				result = append(result, t)
			}
		case "--done":
			if !t.Paused && !t.End.After(now) {
				result = append(result, t)
			}
		default:
			result = append(result, t)
		}
	}
	return result
}

func resolveIndex(timers []Timer, filter string, idx int) (int, error) {
	if idx < 1 {
		return -1, fmt.Errorf("index must be >= 1")
	}
	filtered := getFilteredTimers(timers, filter)
	if idx > len(filtered) {
		return -1, fmt.Errorf("index %d out of range (filter shows %d timer(s))", idx, len(filtered))
	}
	targetTimer := filtered[idx-1]
	for i, t := range timers {
		if t.Name == targetTimer.Name && t.End.Equal(targetTimer.End) {
			return i, nil
		}
	}
	return -1, fmt.Errorf("timer not found")
}

func formatEndTimeCLI(end, now time.Time) string {
	if end.Day() == now.Day() && end.Month() == now.Month() && end.Year() == now.Year() {
		return end.Format("15:04:05")
	} else if end.Month() == now.Month() && end.Year() == now.Year() {
		return end.Format("Jan 2 15:04")
	} else if end.Year() == now.Year() {
		return end.Format("Jan 2")
	} else {
		return end.Format("2006-01-02")
	}
}

func listTimers(timers []Timer, filter string) {
	now := time.Now()
	filtered := getFilteredTimers(timers, filter)

	fmt.Println("Countdown Timers")
	fmt.Println("================")
	fmt.Println()

	if len(filtered) == 0 {
		fmt.Println("No timers found.")
		return
	}

	for i, t := range filtered {
		var statusEmoji, remainingText, endTimeText string

		if t.Paused {
			statusEmoji = "[paused]"
			remainingText = formatDuration(t.Remaining)
			endTimeText = ""
		} else {
			remaining := time.Until(t.End)
			if remaining <= 0 {
				statusEmoji = "[done]"
				remainingText = "Done"
				elapsed := t.Duration - remaining
				endTimeText = fmt.Sprintf("(+%s elapsed)", formatDuration(elapsed))
			} else {
				statusEmoji = "[active]"
				remainingText = formatDuration(remaining)
				endTimeText = fmt.Sprintf("(ends %s)", formatEndTimeCLI(t.End, now))
			}
		}

		fmt.Printf("[%d] %s %-30s %-13s", i+1, statusEmoji, t.Name, remainingText)
		if endTimeText != "" {
			fmt.Printf(" %s", endTimeText)
		}
		fmt.Println()
	}

	fmt.Printf("\nShowing %d timer(s)\n", len(filtered))
}

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

	// Load timers for CLI commands
	timers, err := loadTimers()
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error loading timers: %v\n", err)
		os.Exit(1)
	}
	dirty := false

	switch cmd {
	case "add":
		if len(args) < 2 {
			fmt.Println("Usage: go-countdown add <name> <duration>")
			fmt.Println("\nDuration examples: 30s, 5m, 1h, 2d, 1y, 30d30m, 1h30m")
			os.Exit(1)
		}
		name := args[0]
		duration := args[1]
		d, err := parseDuration(duration)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid duration: %v\n", err)
			os.Exit(1)
		}
		newTimer := Timer{
			Name:     name,
			End:      time.Now().Add(d),
			Duration: d,
		}
		timers = append(timers, newTimer)
		dirty = true
		fmt.Printf("Added timer \"%s\" (%s)\n", name, formatDuration(d))

	case "list":
		filter := ""
		if len(args) > 0 && strings.HasPrefix(args[0], "--") {
			filter = args[0]
		}
		listTimers(timers, filter)

	case "pause":
		filter, _, idx := parseFilterAndIndex(args)
		actualIdx, err := resolveIndex(timers, filter, idx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		if actualIdx >= 0 && len(timers) > 0 {
			t := &timers[actualIdx]
			if !t.Paused {
				if t.End.After(time.Now()) {
					t.Remaining = time.Until(t.End)
					t.Paused = true
					dirty = true
					fmt.Printf("Paused timer \"%s\"\n", t.Name)
				} else {
					fmt.Fprintf(os.Stderr, "Cannot pause: timer already done\n")
					os.Exit(1)
				}
			} else {
				fmt.Printf("Timer \"%s\" is already paused\n", t.Name)
			}
		}

	case "resume":
		filter, _, idx := parseFilterAndIndex(args)
		actualIdx, err := resolveIndex(timers, filter, idx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		if actualIdx >= 0 && len(timers) > 0 {
			t := &timers[actualIdx]
			if t.Paused {
				if t.Remaining > 0 {
					t.End = time.Now().Add(t.Remaining)
					t.Paused = false
					dirty = true
					fmt.Printf("Resumed timer \"%s\"\n", t.Name)
				} else {
					fmt.Fprintf(os.Stderr, "Cannot resume: no remaining time\n")
					os.Exit(1)
				}
			} else {
				fmt.Printf("Timer \"%s\" is already active\n", t.Name)
			}
		}

	case "delete":
		filter, _, idx := parseFilterAndIndex(args)
		actualIdx, err := resolveIndex(timers, filter, idx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		if actualIdx >= 0 && len(timers) > 0 {
			deletedName := timers[actualIdx].Name
			timers = append(timers[:actualIdx], timers[actualIdx+1:]...)
			dirty = true
			fmt.Printf("Deleted timer \"%s\"\n", deletedName)
		}

	case "restart":
		filter, _, idx := parseFilterAndIndex(args)
		actualIdx, err := resolveIndex(timers, filter, idx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		if actualIdx >= 0 && len(timers) > 0 && timers[actualIdx].Duration > 0 {
			t := &timers[actualIdx]
			t.End = time.Now().Add(t.Duration)
			t.Paused = false
			t.Remaining = 0
			dirty = true
			fmt.Printf("Restarted timer \"%s\"\n", t.Name)
		}

	case "edit":
		if len(args) < 3 {
			fmt.Println("Usage: go-countdown edit [--filter] <index> <name> <duration>")
			fmt.Println("\nExamples:")
			fmt.Println("  go-countdown edit 1 \"New Name\" 10m")
			fmt.Println("  go-countdown edit --active 1 \"New Name\" 10m")
			os.Exit(1)
		}

		var filter, indexStr, name, durationStr string
		if strings.HasPrefix(args[0], "--") {
			filter = args[0]
			indexStr = args[1]
			name = args[2]
			if len(args) > 3 {
				durationStr = args[3]
			}
		} else {
			indexStr = args[0]
			name = args[1]
			if len(args) > 2 {
				durationStr = args[2]
			}
		}

		idx, err := strconv.Atoi(indexStr)
		if err != nil || idx < 1 {
			fmt.Fprintf(os.Stderr, "Invalid index: %s\n", indexStr)
			os.Exit(1)
		}

		actualIdx, err := resolveIndex(timers, filter, idx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		if actualIdx >= 0 && len(timers) > 0 {
			t := &timers[actualIdx]
			oldName := t.Name

			// Update name if provided
			if name != "" {
				t.Name = name
			}

			// Update duration if provided
			if durationStr != "" {
				d, err := parseDuration(durationStr)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Invalid duration: %v\n", err)
					os.Exit(1)
				}
				t.Duration = d
				t.End = time.Now().Add(d)
				t.Paused = false
				t.Remaining = 0
			}

			dirty = true
			fmt.Printf("Edited timer: \"%s\" -> \"%s\"\n", oldName, t.Name)
		}

	case "help", "-h", "--help":
		printUsage()

	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		fmt.Println()
		printUsage()
		os.Exit(1)
	}

	// Save if any changes were made
	if dirty {
		if err := saveTimers(timers); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving timers: %v\n", err)
			os.Exit(1)
		}
	}
}
