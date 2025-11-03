package main

import (
	"fmt"
	"os"

	"github.com/76creates/stickers/flexbox"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	flexBox    *flexbox.FlexBox
	useFixed   bool
	termWidth  int
	termHeight int
}

func main() {
	m := model{
		flexBox:  flexbox.New(0, 0),
		useFixed: true, // Start with fixed heights enabled
	}

	p := tea.NewProgram(&m, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func (m *model) Init() tea.Cmd { return nil }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		m.rebuildFlexBox()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "t":
			m.useFixed = !m.useFixed
			m.rebuildFlexBox()
		}
	}
	return m, nil
}

func (m *model) rebuildFlexBox() {
	// Create new FlexBox with current terminal size
	m.flexBox = flexbox.New(m.termWidth, m.termHeight)

	// Create header row
	headerRow := m.flexBox.NewRow()
	headerCell := flexbox.NewCell(1, 1)
	headerCell.SetStyle(lipgloss.NewStyle().
		Background(lipgloss.Color("#fc5c65")).
		Foreground(lipgloss.Color("#ffffff")).
		Bold(true))
	headerCell.SetContentGenerator(func(w, h int) string {
		return lipgloss.NewStyle().
			Width(w).
			Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("HEADER\n%dx%d", w, h))
	})
	headerRow.AddCells(headerCell)

	// Set fixed height if enabled
	if m.useFixed {
		headerRow.SetFixedHeight(3)
	}

	// Create content row with two columns
	contentRow := m.flexBox.NewRow()

	// Sidebar cell
	sidebarCell := flexbox.NewCell(1, 2) // Width ratio 1, Height ratio 2
	sidebarCell.SetStyle(lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#26de81")))
	sidebarCell.SetContentGenerator(func(w, h int) string {
		return lipgloss.NewStyle().
			Width(w).
			Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Sidebar\n%dx%d", w, h))
	})

	// Main content cell
	mainCell := flexbox.NewCell(3, 2) // Width ratio 3, Height ratio 2
	mainCell.SetStyle(lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#45aaf2")))

	modeText := "Dynamic"
	if m.useFixed {
		modeText = "Fixed (Header=3, Footer=3)"
	}

	mainCell.SetContentGenerator(func(w, h int) string {
		return lipgloss.NewStyle().
			Width(w).
			Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Main Content Area\nMode: %s\nSize: %dx%d\n\nPress 't' to toggle fixed heights", modeText, w, h))
	})

	contentRow.AddCells(sidebarCell, mainCell)
	// Content row remains dynamic (no fixed height)

	// Create footer row
	footerRow := m.flexBox.NewRow()
	footerCell := flexbox.NewCell(1, 1)
	footerCell.SetStyle(lipgloss.NewStyle().
		Background(lipgloss.Color("#fd9644")).
		Foreground(lipgloss.Color("#000000")).
		Bold(true))
	footerCell.SetContentGenerator(func(w, h int) string {
		return lipgloss.NewStyle().
			Width(w).
			Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("FOOTER\n%dx%d", w, h))
	})
	footerRow.AddCells(footerCell)

	// Set fixed height if enabled
	if m.useFixed {
		footerRow.SetFixedHeight(3)
	}

	// Add all rows to FlexBox
	m.flexBox.AddRows([]*flexbox.Row{headerRow, contentRow, footerRow})
}

func (m *model) View() string {
	if m.termWidth == 0 || m.termHeight == 0 {
		return "Initializing..."
	}
	m.flexBox.SetWidth(m.termWidth)
	m.flexBox.SetHeight(m.termHeight)
	return m.flexBox.Render()
}