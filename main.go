package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg time.Time

type Timer struct {
	Name string
	End  time.Time
}

type model struct {
	timers []Timer
	now    time.Time
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
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	case tickMsg:
		m.now = time.Time(msg)

		allDone := true
		for _, t := range m.timers {
			if m.now.Before(t.End) {
				allDone = false
				break
			}
		}

		if allDone {
			return m, tea.Quit
		}

		return m, tick()
	}

	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString("⏳ Countdowns\n\n")

	for _, t := range m.timers {
		remaining := time.Until(t.End).Round(time.Second)

		if remaining <= 0 {
			b.WriteString(fmt.Sprintf("✅ %s done\n", t.Name))
		} else {
			b.WriteString(fmt.Sprintf("⏳ %s: %s\n", t.Name, remaining))
		}
	}

	b.WriteString("\n(press q to quit)\n")
	return b.String()
}

func main() {
	m := model{
		timers: []Timer{
			{"A", time.Now().Add(10 * time.Second)},
			{"B", time.Now().Add(5 * time.Second)},
			{"C", time.Now().Add(15 * time.Second)},
		},
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
