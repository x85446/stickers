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
		"#2d98da", "#3867d6", "#8854d0", "#a5b1c2", "#4b6584",
	}
)

type model struct {
	flexBox        *flexbox.FlexBox
	borderType     int // 0=none, 1=normal, 2=rounded, 3=thick, 4=double, 5=hidden
	cellIndexes    [][]int // Store cell indexes for each row
}

var borderTypes = []lipgloss.Border{
	{}, // no border
	lipgloss.NormalBorder(),
	lipgloss.RoundedBorder(),
	lipgloss.ThickBorder(),
	lipgloss.DoubleBorder(),
	lipgloss.HiddenBorder(),
}

var borderNames = []string{
	"Filled Background",
	"Normal Border",
	"Rounded Border",
	"Thick Border",
	"Double Border",
	"Hidden Border",
}

func main() {
	m := model{
		flexBox:     flexbox.New(0, 0),
		borderType:  1, // Start with normal border
		cellIndexes: [][]int{
			{0, 1, 2},
			{3, 4, 5, 6, 7, 8, 9},
			{10, 11, 12, 13, 14},
			{15, 16, 17, 18, 19},
		},
	}

	rows := []*flexbox.Row{
		m.flexBox.NewRow().AddCells(
			flexbox.NewCell(1, 6),
			flexbox.NewCell(1, 6),
			flexbox.NewCell(1, 6),
		),
		m.flexBox.NewRow().AddCells(
			flexbox.NewCell(2, 4),
			flexbox.NewCell(2, 4),
			flexbox.NewCell(3, 4),
			flexbox.NewCell(3, 4),
			flexbox.NewCell(3, 4),
			flexbox.NewCell(4, 4),
			flexbox.NewCell(4, 4),
		),
		m.flexBox.NewRow().AddCells(
			flexbox.NewCell(2, 5),
			flexbox.NewCell(3, 5),
			flexbox.NewCell(10, 5).SetContentGenerator(func(w, h int) string {
				return lipgloss.NewStyle().
					Width(w).
					Height(h).
					Align(lipgloss.Center, lipgloss.Center).
					Render(borderNames[m.borderType] + " (t)")
			}),
			flexbox.NewCell(3, 5),
			flexbox.NewCell(2, 5),
		),
		m.flexBox.NewRow().AddCells(
			flexbox.NewCell(1, 4),
			flexbox.NewCell(1, 3),
			flexbox.NewCell(1, 2),
			flexbox.NewCell(1, 3),
			flexbox.NewCell(1, 4),
		),
	}

	m.flexBox.AddRows(rows)
	m.updateStyles()

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
		m.flexBox.SetWidth(msg.Width)
		m.flexBox.SetHeight(msg.Height)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "t":
			m.borderType = (m.borderType + 1) % len(borderTypes)
			m.updateStyles()
		}

	}
	return m, nil
}

func (m *model) updateStyles() {
	for rowIdx := 0; rowIdx < m.flexBox.RowsLen(); rowIdx++ {
		row := m.flexBox.GetRow(rowIdx)
		if row == nil {
			continue
		}
		for cellIdx := 0; cellIdx < row.CellsLen(); cellIdx++ {
			cell := row.GetCell(cellIdx)
			if cell == nil {
				continue
			}
			colorIdx := m.cellIndexes[rowIdx][cellIdx]
			var style lipgloss.Style
			if m.borderType == 0 {
				// Filled background
				style = lipgloss.NewStyle().Background(lipgloss.Color(colors[colorIdx]))
			} else {
				// Border with type
				style = lipgloss.NewStyle().
					Border(borderTypes[m.borderType]).
					BorderForeground(lipgloss.Color(colors[colorIdx]))
			}
			cell.SetStyle(style)
		}
	}
}

func (m *model) View() string {
	return m.flexBox.Render()
}
