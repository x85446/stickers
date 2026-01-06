package main

import (
	"fmt"
	"os"

	"github.com/76creates/stickers/flexbox"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	colors = []string{
		"#fc5c65", "#fd9644", "#fed330", "#26de81", "#2bcbba",
		"#eb3b5a", "#fa8231", "#f7b731", "#20bf6b", "#0fb9b1",
		"#45aaf2", "#4b7bec", "#a55eea", "#d1d8e0", "#778ca3",
		"#fc5c65", "#fd9644", "#fed330", "#26de81", "#2bcbba",
		"#eb3b5a", "#fa8231", "#f7b731", "#20bf6b", "#0fb9b1",
		"#45aaf2", "#4b7bec", "#a55eea", "#d1d8e0", "#778ca3",
	}

	borderTypes = []lipgloss.Border{
		lipgloss.NormalBorder(),
		lipgloss.RoundedBorder(),
		lipgloss.ThickBorder(),
		lipgloss.DoubleBorder(),
	}

	borderNames = []string{
		"Normal Border",
		"Rounded Border",
		"Thick Border",
		"Double Border",
	}
)

type thingy struct {
	flexBox    *flexbox.FlexBox
	borderType int // 0=normal, 1=rounded, 2=thick, 3=double
}

func main() {
	m := thingy{
		flexBox:    flexbox.New(0, 0),
		borderType: 1, // Start with rounded border
	}

	// Create a single cell that will contain nested FlexBoxes
	row := m.flexBox.NewRow()
	cell := flexbox.NewCell(1, 1)

	// Use ContentGenerator to create nested borders
	cell.SetContentGenerator(func(maxX, maxY int) string {
		return createNestedFlexBox(maxX, maxY, 0, m.borderType)
	})

	row.AddCells(cell)
	m.flexBox.AddRows([]*flexbox.Row{row})

	p := tea.NewProgram(&m, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

// createNestedFlexBox recursively creates FlexBoxes with borders
// depth tracks which color to use and when to stop
// borderType selects which border style to use
func createNestedFlexBox(width, height, depth, borderType int) string {
	// Stop when we run out of space or colors
	// Minimum needed: 2 chars width (left+right border), 2 chars height (top+bottom border)
	if width < 4 || height < 4 || depth >= len(colors) {
		// Return centered text at the deepest level
		style := lipgloss.NewStyle().
			Width(width).
			Height(height).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(lipgloss.Color("#ffffff"))
		return style.Render(fmt.Sprintf("Depth: %d\n%s (t)", depth, borderNames[borderType]))
	}

	// Create a FlexBox for this level
	box := flexbox.New(width, height)
	row := box.NewRow()
	cell := flexbox.NewCell(1, 1)

	// Create border style with current color and selected border type
	borderStyle := lipgloss.NewStyle().
		Border(borderTypes[borderType]).
		BorderForeground(lipgloss.Color(colors[depth]))

	cell.SetStyle(borderStyle)

	// Set content generator to recursively create inner FlexBox
	cell.SetContentGenerator(func(innerWidth, innerHeight int) string {
		return createNestedFlexBox(innerWidth, innerHeight, depth+1, borderType)
	})

	row.AddCells(cell)
	box.AddRows([]*flexbox.Row{row})

	return box.Render()
}

func (m *thingy) Init() tea.Cmd { return nil }

func (m *thingy) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.flexBox.SetWidth(msg.Width)
		m.flexBox.SetHeight(msg.Height)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "t":
			m.borderType = (m.borderType + 1) % len(borderTypes)
		}
	}
	return m, nil
}

func (m *thingy) View() string {
	return m.flexBox.Render()
}
