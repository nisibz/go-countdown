package main

import "github.com/charmbracelet/bubbles/key"

type filterMode int

const (
	filterAll filterMode = iota
	filterActive
	filterPaused
	filterDone
)

type bulkActionType int

const (
	bulkPauseAll bulkActionType = iota
	bulkResumeAll
	bulkDeleteDone
	bulkRestartAll
)

// defaultKeyMap defines keybindings for the main timer view
type defaultKeyMap struct {
	Up         key.Binding
	Down       key.Binding
	UpOrder    key.Binding
	DownOrder  key.Binding
	Add        key.Binding
	Delete     key.Binding
	DeleteDone key.Binding
	Edit       key.Binding
	Redo       key.Binding
	RestartAll key.Binding
	Pause      key.Binding
	PauseAll   key.Binding
	ResumeAll  key.Binding
	Filter1    key.Binding
	Filter2    key.Binding
	Filter3    key.Binding
	Filter4    key.Binding
	Help       key.Binding
	Quit       key.Binding
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
		{k.DeleteDone, k.RestartAll, k.PauseAll, k.ResumeAll},
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
		DeleteDone: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "delete all done"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit timer"),
		),
		Redo: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "restart timer"),
		),
		RestartAll: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "restart all"),
		),
		Pause: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "pause/resume"),
		),
		PauseAll: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "pause all"),
		),
		ResumeAll: key.NewBinding(
			key.WithKeys("shift+r"),
			key.WithHelp("shift+R", "resume all"),
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
	Increase  key.Binding // + or = key
	Decrease  key.Binding // - or _ key
}

// ShortHelp returns keybindings for the mini help view
func (k formKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Enter, k.Esc}
}

// FullHelp returns keybindings for the full help view
func (k formKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.NextField, k.PrevField},
		{k.Increase, k.Decrease},
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
		Increase: key.NewBinding(
			key.WithKeys("+", "="),
			key.WithHelp("+/=", "increase duration"),
		),
		Decrease: key.NewBinding(
			key.WithKeys("-", "_"),
			key.WithHelp("-/_", "decrease duration"),
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
