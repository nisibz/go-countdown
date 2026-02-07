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
	Name string
	End  time.Time
}

type model struct {
	timers []Timer
	now    time.Time
	paused bool
	cursor int

	adding      bool
	inputName   string
	inputSecs   string
	activeField int // 0 = name, 1 = seconds
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
		if m.adding {
			switch msg.String() {

			case "tab":
				m.activeField = (m.activeField + 1) % 2
				return m, nil

			case "enter":
				secs, err := strconv.Atoi(m.inputSecs)
				if err == nil && secs > 0 && m.inputName != "" {
					newTimer := Timer{
						Name: m.inputName,
						End:  time.Now().Add(time.Duration(secs) * time.Second),
					}
					m.timers = append(m.timers, newTimer)
					m.cursor = len(m.timers) - 1
				}

				m.adding = false
				m.inputName = ""
				m.inputSecs = ""
				return m, nil

			case "esc":
				m.adding = false
				m.inputName = ""
				m.inputSecs = ""
				return m, nil

			case "backspace":
				if m.activeField == 0 && len(m.inputName) > 0 {
					m.inputName = m.inputName[:len(m.inputName)-1]
				}
				if m.activeField == 1 && len(m.inputSecs) > 0 {
					m.inputSecs = m.inputSecs[:len(m.inputSecs)-1]
				}
				return m, nil

			default:
				if msg.Type == tea.KeyRunes {
					switch m.activeField {
					case 0:
						m.inputName += msg.String()

					case 1:
						if msg.String()[0] >= '0' && msg.String()[0] <= '9' {
							m.inputSecs += msg.String()
						}
					}
				}
				return m, nil
			}
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "p":
			m.paused = !m.paused
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

			m.timers = append(
				m.timers[:m.cursor],
				m.timers[m.cursor+1:]...,
			)

			if m.cursor >= len(m.timers) && m.cursor > 0 {
				m.cursor--
			}

			return m, nil

		case "a":
			if m.adding {
				return m, nil
			}

			m.adding = true
			m.inputName = ""
			m.inputSecs = ""
			m.activeField = 0
			return m, nil

		}

	case tickMsg:

		if m.paused {
			return m, tick()
		}

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
	if m.adding {
		nameCursor := " "
		secCursor := " "

		if m.activeField == 0 {
			nameCursor = ">"
		} else {
			secCursor = ">"
		}

		return fmt.Sprintf(
			"➕ Add Timer\n\n"+
				"%s Name   : %s\n"+
				"%s Seconds: %s\n\n"+
				"[tab] switch  [enter] confirm  [esc] cancel\n",
			nameCursor, m.inputName,
			secCursor, m.inputSecs,
		)
	}

	var b strings.Builder

	if m.paused {
		b.WriteString("⏸ PAUSED\n\n")
	} else {
		b.WriteString("⏳ Countdowns\n\n")
	}

	for i, t := range m.timers {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		remaining := time.Until(t.End).Round(time.Second)

		if remaining <= 0 {
			fmt.Fprintf(&b, "%s✅ %s done\n", cursor, t.Name)
		} else {
			fmt.Fprintf(&b, "%s⏳ %s: %s\n", cursor, t.Name, remaining)
		}
	}

	b.WriteString("\n[a] add  d delete  ↑/↓ move  p pause  q quit\n")
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
