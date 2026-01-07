package main

import (
	"fmt"
	"os"

	demo1 "github.com/x85446/stickers/example/demo-1-flex-box-simple"
	demo10 "github.com/x85446/stickers/example/demo-10-flex-box-mixed-fixed"
	demo11 "github.com/x85446/stickers/example/demo-11-flex-box-cell-config"
	demo12 "github.com/x85446/stickers/example/demo-12-flex-box-row-align"
	demo2 "github.com/x85446/stickers/example/demo-2-flex-box-horizonal"
	demo3 "github.com/x85446/stickers/example/demo-3-flex-box-with-table"
	demo4 "github.com/x85446/stickers/example/demo-4-table-simple-string"
	demo5 "github.com/x85446/stickers/example/demo-5-table-multi-type"
	demo6 "github.com/x85446/stickers/example/demo-6-flex-box-nested-borders"
	demo7 "github.com/x85446/stickers/example/demo-7-flex-box-simple-borders"
	demo8 "github.com/x85446/stickers/example/demo-8-flex-box-fixed-rows"
	demo9 "github.com/x85446/stickers/example/demo-9-flex-box-fixed-width"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type demoInfo struct {
	name      string
	hasBorder bool // has 'b' for border toggle
}

var demoInfos = []demoInfo{
	{"1: FlexBox Simple", false},
	{"2: FlexBox Horizontal", false},
	{"3: FlexBox with Table", false},
	{"4: Table Simple String", false},
	{"5: Table Multi-Type", false},
	{"6: Nested Borders", true},
	{"7: Simple Borders", true},
	{"8: Fixed Row Heights", false},
	{"9: Fixed Column Widths", false},
	{"10: Mixed Fixed/Dynamic", true},
	{"11: Cell Config", true},
	{"12: Row Alignment", false},
}

type model struct {
	current int
	demos   []tea.Model
	width   int
	height  int
}

func main() {
	m := &model{
		current: 0,
		demos: []tea.Model{
			demo1.New(),
			demo2.New(),
			demo3.New(),
			demo4.New(),
			demo5.New(),
			demo6.New(),
			demo7.New(),
			demo8.New(),
			demo9.New(),
			demo10.New(),
			demo11.New(),
			demo12.New(),
		},
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running demo-all: %v", err)
		os.Exit(1)
	}
}

func (m *model) Init() tea.Cmd {
	// Initialize the current demo
	return m.demos[m.current].Init()
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Pass window size to current demo
		var cmd tea.Cmd
		m.demos[m.current], cmd = m.demos[m.current].Update(msg)
		return m, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "shift+tab":
			// Previous demo (wrap around)
			m.current = (m.current - 1 + len(m.demos)) % len(m.demos)
			// Send window size to new demo
			var cmd tea.Cmd
			m.demos[m.current], cmd = m.demos[m.current].Update(tea.WindowSizeMsg{
				Width:  m.width,
				Height: m.height - 1, // Account for header
			})
			return m, cmd
		case "tab":
			// Next demo (wrap around)
			m.current = (m.current + 1) % len(m.demos)
			// Send window size to new demo
			var cmd tea.Cmd
			m.demos[m.current], cmd = m.demos[m.current].Update(tea.WindowSizeMsg{
				Width:  m.width,
				Height: m.height - 1, // Account for header
			})
			return m, cmd
		case "q":
			// Let q go through to demo for its about toggle, but intercept if needed
			// Most demos use 'q' to quit, so we handle it here to switch demos instead
			// If you want demos to handle 'q', remove this case
			return m, tea.Quit
		}
	}

	// Pass all other messages to the current demo
	var cmd tea.Cmd
	m.demos[m.current], cmd = m.demos[m.current].Update(msg)
	return m, cmd
}

var headerStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#ffffff")).
	Background(lipgloss.Color("#7D56F4"))

func (m *model) View() string {
	info := demoInfos[m.current]

	// Build feature hints
	features := "[a]bout"
	if info.hasBorder {
		features += " [b]order"
	}

	header := headerStyle.Width(m.width).Align(lipgloss.Center).Render(
		fmt.Sprintf("Demo %s | %s | [tab] next | [shift+tab] prev | [q] quit", info.name, features),
	)

	return header + "\n" + m.demos[m.current].View()
}
