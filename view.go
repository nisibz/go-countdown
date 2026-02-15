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
	if m.confirmingDelete || m.confirmingBulk {
		return renderPopupOverlay(m)
	}

	if m.adding || m.editing {
		return renderPopupOverlay(m)
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

func renderMainView(m model) string {
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

func renderPopupForm(m model) string {
	// Define styles
	var (
		borderColor  = lipgloss.Color("99")   // Purple border
		focusedColor = lipgloss.Color("226")  // Bright yellow for focus
		labelColor   = lipgloss.Color("147")  // Light blue for labels
		hintColor    = lipgloss.Color("244")  // Gray for hints

		popupStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 2).
			Width(58)

		titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("213")). // Pink/purple title
			MarginBottom(1)

		labelStyle = lipgloss.NewStyle().
			Width(9).
			Foreground(labelColor)

		focusedLabelStyle = labelStyle.
			Foreground(focusedColor).
			Bold(true)

		hintStyle = lipgloss.NewStyle().
			Foreground(hintColor)

		helpStyle = lipgloss.NewStyle().
			MarginTop(1).
			Foreground(lipgloss.Color("245"))

		divider = lipgloss.NewStyle().
			Foreground(hintColor).
			Render(strings.Repeat("─", 54))
	)

	// Build title
	var title string
	if m.editing {
		title = "✏️  Edit Timer"
	} else {
		title = "➕️ Add Timer"
	}

	// Build form content
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")
	b.WriteString(divider)
	b.WriteString("\n\n")

	// Name input - show focused label if name input is focused
	nameLabel := "Name:"
	if m.nameInput.Focused() {
		nameLabel = focusedLabelStyle.Render(nameLabel)
	} else {
		nameLabel = labelStyle.Render(nameLabel)
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, nameLabel, " ", m.nameInput.View()))
	b.WriteString("\n\n")

	// Duration input
	durationLabel := "Duration:"
	if m.durationInput.Focused() {
		durationLabel = focusedLabelStyle.Render(durationLabel)
	} else {
		durationLabel = labelStyle.Render(durationLabel)
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, durationLabel, " ", m.durationInput.View()))
	b.WriteString("\n\n")

	// Validation hint
	b.WriteString(hintStyle.Render("Examples: 30s, 5m, 1h, 2d, 1y"))
	b.WriteString("\n")

	// Help text
	b.WriteString(helpStyle.Render(m.help.View(m.formKeys)))

	return popupStyle.Render(b.String())
}

func renderConfirmPopup(m model) string {
	// Define styles
	var (
		borderColor = lipgloss.Color("99")   // Purple border
		labelColor  = lipgloss.Color("147")  // Light blue for labels
		hintColor   = lipgloss.Color("244")  // Gray for hints

		popupStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 2).
			Width(58)

		titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("213")). // Pink/purple title
			MarginBottom(1)

		labelStyle = lipgloss.NewStyle().
			Foreground(labelColor)

		helpStyle = lipgloss.NewStyle().
			MarginTop(1).
			Foreground(hintColor)

		divider = lipgloss.NewStyle().
			Foreground(hintColor).
			Render(strings.Repeat("─", 54))
	)

	// Build title and message
	var title, message string
	if m.confirmingDelete {
		actualIdx := m.getActualTimerIndex(m.cursor)
		title = "🗑️  Delete Timer"
		message = fmt.Sprintf("Delete \"%s\"?", m.timers[actualIdx].Name)
	} else {
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
	}

	// Build form content
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")
	b.WriteString(divider)
	b.WriteString("\n\n")

	// Message
	b.WriteString(labelStyle.Render(message))
	b.WriteString("\n")

	// Help text
	b.WriteString(helpStyle.Render(m.help.View(m.confirmKeys)))

	return popupStyle.Render(b.String())
}

func renderPopupOverlay(m model) string {
	// Get dimensions
	width := m.width
	height := m.height
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	// Render the main view (background)
	mainView := renderMainView(m)
	mainLines := strings.Split(mainView, "\n")

	// Render the popup
	var popup string
	if m.confirmingDelete || m.confirmingBulk {
		popup = renderConfirmPopup(m)
	} else {
		popup = renderPopupForm(m)
	}
	popupLines := strings.Split(popup, "\n")
	popupHeight := len(popupLines)

	// Calculate vertical position for popup (centered)
	popupStartRow := max(0, (height-popupHeight)/2)

	// Calculate horizontal position for popup (centered)
	// Popup width is 58 (from style) + borders (4) + padding (4) = ~66 chars
	popupWidth := 66
	popupStartCol := max(0, (width-popupWidth)/2)
	popupEndCol := popupStartCol + popupWidth

	// Build the overlay view line by line
	var result strings.Builder

	for row := 0; row < height; row++ {
		if row >= popupStartRow && row < popupStartRow+popupHeight {
			// This row has the popup overlay
			popupLineIdx := row - popupStartRow
			popupLine := popupLines[popupLineIdx]

			// Get the corresponding background line (if any)
			var bgLine string
			if row < len(mainLines) {
				bgLine = mainLines[row]
			}

			// Build the line: background + popup
			var lineBuilder strings.Builder

			// Part before popup (normal background)
			if len(bgLine) > popupStartCol {
				lineBuilder.WriteString(bgLine[:popupStartCol])
			} else {
				lineBuilder.WriteString(bgLine)
				if len(bgLine) < popupStartCol {
					lineBuilder.WriteString(strings.Repeat(" ", popupStartCol-len(bgLine)))
				}
			}

			// The popup itself
			lineBuilder.WriteString(popupLine)

			// Part after popup (normal background, if needed)
			if len(bgLine) > popupEndCol {
				lineBuilder.WriteString(bgLine[popupEndCol:])
			}

			result.WriteString(lineBuilder.String())
		} else {
			// This row shows only normal background
			if row < len(mainLines) {
				result.WriteString(mainLines[row])
			} else {
				result.WriteString(strings.Repeat(" ", width))
			}
		}

		if row < height-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}
