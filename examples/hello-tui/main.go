// Command hello-tui is a minimal Bubble Tea program used by tea-eyes
// integration tests and demos. It shows a bordered greeting; pressing space
// toggles a counter, and q or ctrl+c quits.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	Counter     int
	ShowCounter bool
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case " ", "space":
			m.ShowCounter = !m.ShowCounter
			if m.ShowCounter {
				m.Counter++
			}
			return m, nil
		}
	}
	return m, nil
}

var borderStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	Padding(0, 2)

func (m model) View() string {
	body := "Hello, tea-eyes!"
	if m.ShowCounter {
		body += fmt.Sprintf("\nCounter: %d", m.Counter)
	}
	return borderStyle.Render(body) + "\n"
}

func main() {
	p := tea.NewProgram(model{})
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "hello-tui: %v\n", err)
		os.Exit(1)
	}
}
