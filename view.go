package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

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
		status := t.StatusEmoji(m.now)
		remainingText := t.StatusText(m.now)
		endTimeText := t.EndTimeText(m.now)

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

	if m.confirmingBulk {
		var title, message string
		switch m.pendingBulkAction {
		case bulkPauseAll:
			title = "⏸️  Pause All Active"
			message = "Pause all active timers?"
		case bulkResumeAll:
			title = "▶️  Resume All Paused"
			message = "Resume all paused timers?"
		case bulkDeleteDone:
			title = "🗑️  Delete Completed"
			message = "Delete all completed timers?"
		case bulkRestartAll:
			title = "🔄  Restart All"
			message = "Restart all timers?"
		}
		return fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			title,
			message,
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

func setupTableStyles(tbl table.Model) table.Model {
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
	return tbl
}
