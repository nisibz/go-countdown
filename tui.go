package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type (
	tickMsg      time.Time
	fileWatchMsg struct{}
)

type model struct {
	timers []Timer
	now    time.Time
	cursor int
	table  table.Model

	adding            bool
	editing           bool           // true when editing existing timer
	editingIndex      int            // actual index of timer being edited
	confirmingDelete  bool           // true when showing delete confirmation
	confirmingRestart bool           // true when showing restart confirmation
	confirmingBulk    bool           // true when showing bulk operation confirmation
	pendingBulkAction bulkActionType // which bulk action to execute
	nameInput         textinput.Model
	durationInput     textinput.Model

	filter filterMode

	dirty       bool
	lastModTime time.Time // track file modification time for external changes
	defaultKeys defaultKeyMap
	formKeys    formKeyMap
	confirmKeys confirmKeyMap
	help        help.Model

	width  int // terminal width
	height int // terminal height
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
	tbl = setupTableStyles(tbl)

	// Initialize text inputs
	nameInput := textinput.New()
	nameInput.Placeholder = "Timer name"
	nameInput.Focus()

	durationInput := textinput.New()
	durationInput.Placeholder = "30s, 5m, 1h, 2d, 1y"
	durationInput.Validate = func(s string) error {
		// Allow empty string during typing
		if s == "" {
			return nil
		}
		// Validate: only digits and s/m/h/d/y suffixes allowed
		for _, r := range s {
			if (r < '0' || r > '9') && r != 's' && r != 'm' && r != 'h' && r != 'd' && r != 'y' && r != ' ' {
				return fmt.Errorf("invalid duration format")
			}
		}
		return nil
	}

	m := model{
		now:           time.Now(),
		filter:        filterAll,
		defaultKeys:   newDefaultKeyMap(),
		formKeys:      newFormKeyMap(),
		confirmKeys:   newConfirmKeyMap(),
		help:          help.New(),
		table:         tbl,
		nameInput:     nameInput,
		durationInput: durationInput,
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
