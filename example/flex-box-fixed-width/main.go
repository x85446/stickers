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
		useFixed: true, // Start with fixed widths enabled
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

	// Create header row to show mode
	headerRow := m.flexBox.NewRow()
	headerCell := flexbox.NewCell(1, 1)
	headerCell.SetStyle(lipgloss.NewStyle().
		Background(lipgloss.Color("#45aaf2")).
		Foreground(lipgloss.Color("#ffffff")).
		Bold(true))

	modeText := "Dynamic (all columns use ratios)"
	if m.useFixed {
		modeText = "Fixed (Sidebar=20, Info=25, Content=remaining)"
	}

	headerCell.SetContentGenerator(func(w, h int) string {
		return lipgloss.NewStyle().
			Width(w).
			Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Column Width Demo - Mode: %s\nPress 't' to toggle", modeText))
	})
	headerRow.AddCells(headerCell)

	// Create main content row with three columns
	mainRow := m.flexBox.NewRow()

	// Left sidebar
	sidebarCell := flexbox.NewCell(1, 1) // Ratio 1 (used when not fixed)
	sidebarCell.SetStyle(lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#26de81")))
	sidebarCell.SetContentGenerator(func(w, h int) string {
		return lipgloss.NewStyle().
			Width(w).
			Height(h).
			Padding(1).
			Render(fmt.Sprintf("Sidebar\n\nWidth: %d\nHeight: %d\n\n• Home\n• About\n• Settings\n• Profile", w, h))
	})
	if m.useFixed {
		sidebarCell.SetFixedWidth(20)
	}

	// Main content area
	contentCell := flexbox.NewCell(3, 1) // Ratio 3 (takes most space when dynamic)
	contentCell.SetStyle(lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#fd9644")))
	contentCell.SetContentGenerator(func(w, h int) string {
		fixedInfo := ""
		if m.useFixed {
			fixedInfo = "\n\nWith fixed widths:\n" +
				"• Sidebar stays at 20 chars\n" +
				"• Info panel stays at 25 chars\n" +
				"• Content takes all remaining space"
		} else {
			fixedInfo = "\n\nWith dynamic widths:\n" +
				"• Sidebar gets 1/5 of total width\n" +
				"• Content gets 3/5 of total width\n" +
				"• Info panel gets 1/5 of total width"
		}

		return lipgloss.NewStyle().
			Width(w).
			Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Main Content Area\n%dx%d%s", w, h, fixedInfo))
	})
	// Content cell never has fixed width - always dynamic

	// Right info panel
	infoCell := flexbox.NewCell(1, 1) // Ratio 1 (used when not fixed)
	infoCell.SetStyle(lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("#a55eea")))
	infoCell.SetContentGenerator(func(w, h int) string {
		return lipgloss.NewStyle().
			Width(w).
			Height(h).
			Padding(1).
			Render(fmt.Sprintf("Info Panel\n\nWidth: %d\nHeight: %d\n\n📊 Stats\n📈 Charts\n🔔 Alerts", w, h))
	})
	if m.useFixed {
		infoCell.SetFixedWidth(25)
	}

	// Add all cells to the main row
	mainRow.AddCells(sidebarCell, contentCell, infoCell)

	// Create footer row
	footerRow := m.flexBox.NewRow()
	footerCell := flexbox.NewCell(1, 1)
	footerCell.SetStyle(lipgloss.NewStyle().
		Background(lipgloss.Color("#778ca3")).
		Foreground(lipgloss.Color("#ffffff")))
	footerCell.SetContentGenerator(func(w, h int) string {
		return lipgloss.NewStyle().
			Width(w).
			Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Terminal: %dx%d | Press 'q' to quit | 't' to toggle fixed widths", m.termWidth, m.termHeight))
	})
	footerRow.AddCells(footerCell)

	// Add all rows to FlexBox
	m.flexBox.AddRows([]*flexbox.Row{headerRow, mainRow, footerRow})
}

func (m *model) View() string {
	if m.termWidth == 0 || m.termHeight == 0 {
		return "Initializing..."
	}
	m.flexBox.SetWidth(m.termWidth)
	m.flexBox.SetHeight(m.termHeight)
	return m.flexBox.Render()
}