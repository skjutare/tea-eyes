// Command multi-pane is a Bubble Tea example used by tea-eyes integration
// tests and the visual rendering demo. It shows two side-by-side panels with
// a focus border and a status bar. Tab switches focus; q or ctrl+c quits.
package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type pane struct {
	title string
	lines []string
}

type model struct {
	panes   [2]pane
	focused int
	width   int
	height  int
}

func initialModel() model {
	return model{
		panes: [2]pane{
			{
				title: "Files",
				lines: []string{
					"  main.go",
					"  model.go",
					"  view.go",
					"  README.md",
				},
			},
			{
				title: "Preview",
				lines: []string{
					"package main",
					"",
					"func main() {",
					"    println(\"hi\")",
					"}",
				},
			},
		},
		focused: 0,
		width:   80,
		height:  24,
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab", "shift+tab":
			m.focused = (m.focused + 1) % 2
		}
	}
	return m, nil
}

var (
	focusedBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("205")).
			Padding(0, 1)
	idleBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("213"))
	statusStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("57")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1)
)

func (m model) View() string {
	paneWidth := max((m.width-4)/2, 16)
	paneHeight := max(m.height-4, 6)

	rendered := make([]string, 2)
	for i, p := range m.panes {
		body := titleStyle.Render(p.title) + "\n" + strings.Join(p.lines, "\n")
		style := idleBorder
		if i == m.focused {
			style = focusedBorder
		}
		rendered[i] = style.
			Width(paneWidth).
			Height(paneHeight).
			Render(body)
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, rendered[0], rendered[1])
	status := statusStyle.
		Width(m.width).
		Render(fmt.Sprintf(" tab: switch pane · q: quit · focus=%s ", m.panes[m.focused].title))

	return row + "\n" + status + "\n"
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "multi-pane: %v\n", err)
		os.Exit(1)
	}
}
