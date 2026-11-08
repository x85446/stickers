package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
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
)

// sharedKeys are the keys every demo answers to, in the order the header shows them.
const sharedKeys = "[a]bout [tab] next [shift+tab] prev [q] quit"

type demoInfo struct {
	name string
	keys string // the demo's own keys, listed after the shared ones
}

var demoInfos = []demoInfo{
	{"1: FlexBox Simple", ""},
	{"2: FlexBox Horizontal", ""},
	{"3: FlexBox with Table", "[arrows] move [ctrl+s] sort [enter] copy cell [letters] filter"},
	{"4: Table Simple String", "[arrows] move [ctrl+s] sort [enter] select [letters] filter"},
	{"5: Table Multi-Type", "[arrows] move [ctrl+s] sort [enter] select [letters] filter"},
	{"6: Nested Borders", "[b] border"},
	{"7: Simple Borders", "[b] border"},
	{"8: Fixed Row Heights", "[m] fixed/dynamic heights"},
	{"9: Fixed Column Widths", "[m] fixed/dynamic widths"},
	{"10: Mixed Fixed/Dynamic", "[b] border [m] fixed/dynamic [h] text"},
	{"11: Cell Config", "[b] border [h] text [C-W] select cell"},
	{"12: Row Alignment", "[r] row align"},
}

type model struct {
	current int
	demos   []tea.Model
	width   int
	height  int
	started bool // false while the splash screen is showing
}

func newModel() *model {
	return &model{
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
}

func main() {
	m := newModel()
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
	if !m.started {
		return m.updateSplash(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		var cmd tea.Cmd
		m.demos[m.current], cmd = m.demos[m.current].Update(m.demoSize())
		return m, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "shift+tab":
			// Previous demo (wrap around)
			m.current = (m.current - 1 + len(m.demos)) % len(m.demos)
			var cmd tea.Cmd
			m.demos[m.current], cmd = m.demos[m.current].Update(m.demoSize())
			return m, cmd
		case "tab":
			// Next demo (wrap around)
			m.current = (m.current + 1) % len(m.demos)
			var cmd tea.Cmd
			m.demos[m.current], cmd = m.demos[m.current].Update(m.demoSize())
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

// demoSize is the window the current demo may draw in: everything under the header line.
func (m *model) demoSize() tea.WindowSizeMsg {
	return tea.WindowSizeMsg{Width: m.width, Height: m.height - 1}
}

// updateSplash handles messages while the splash screen is showing.
// Any key other than q / ctrl+c starts the first demo.
func (m *model) updateSplash(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
		m.started = true
		var cmd tea.Cmd
		m.demos[m.current], cmd = m.demos[m.current].Update(m.demoSize())
		return m, cmd
	}
	// Let ticks and other messages reach the demo so it is ready on start
	var cmd tea.Cmd
	m.demos[m.current], cmd = m.demos[m.current].Update(msg)
	return m, cmd
}

var headerStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#ffffff")).
	Background(lipgloss.Color("#7D56F4"))

var splashTitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#7D56F4"))

var splashKeyStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#ffffff"))

var splashDimStyle = lipgloss.NewStyle().
	Faint(true)

var splashBoxStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#7D56F4")).
	Padding(1, 4)

// viewSplash renders the start screen with the key bindings.
func (m *model) viewSplash() string {
	key := func(k, desc string) string {
		return splashKeyStyle.Render(k) + "  " + desc
	}
	lines := []string{
		splashTitleStyle.Render("Stickers Demos"),
		"",
		key("a          ", "about the current demo"),
		key("tab        ", "next demo"),
		key("shift+tab  ", "previous demo"),
		key("q          ", "quit"),
		"",
		splashDimStyle.Render("each demo lists its own keys in the header"),
		"",
		splashDimStyle.Render("press any key to start"),
	}
	box := splashBoxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m *model) View() string {
	if !m.started {
		return m.viewSplash()
	}

	info := demoInfos[m.current]
	text := fmt.Sprintf("Demo %s | %s", info.name, sharedKeys)
	if info.keys != "" {
		text += " | " + info.keys
	}
	// the header is one line: a narrow terminal cuts the demo's own keys, never wraps them
	if m.width > 0 {
		text = ansi.Truncate(text, m.width, "…")
	}
	header := headerStyle.Width(m.width).Align(lipgloss.Center).Render(text)
	return header + "\n" + m.demos[m.current].View()
}
