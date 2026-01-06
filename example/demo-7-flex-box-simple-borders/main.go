package main

import (
	"fmt"
	"os"

	"github.com/x85446/stickers/flexbox"
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
	flexBox     *flexbox.FlexBox
	borderType  int       // 0=none, 1=normal, 2=rounded, 3=thick, 4=double, 5=hidden
	cellIndexes [][]int   // Store cell indexes for each row
	showAbout   bool
	width       int
	height      int
}

const aboutText = `Demo 7: Simple Borders

Grid of labeled cells with different border styles.

Each cell shows its label (A-T) and ratio [X:Y].
The ratio determines relative cell size:
- First number: width ratio within the row
- Second number: height ratio for the row

Press 'b' to cycle styles:
Fill → Normal → Rounded → Thick → Double → Hidden

Press 'a' to close | 'q' to quit`

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

// Cell labels
var cellLabels = []string{
	"A", "B", "C",
	"D", "E", "F", "G", "H", "I", "J",
	"K", "L", "M", "N", "O",
	"P", "Q", "R", "S", "T",
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

	// Row 1: 3 cells [1:6]
	row1 := m.flexBox.NewRow()
	row1Ratios := [][]int{{1, 6}, {1, 6}, {1, 6}}
	for i, r := range row1Ratios {
		idx := i
		ratioX, ratioY := r[0], r[1]
		cell := flexbox.NewCell(ratioX, ratioY)
		cell.SetContentGenerator(func(w, h int) string {
			return lipgloss.NewStyle().
				Width(w).Height(h).
				Align(lipgloss.Center, lipgloss.Center).
				Render(fmt.Sprintf("%s\n[%d:%d]", cellLabels[idx], ratioX, ratioY))
		})
		row1.AddCells(cell)
	}

	// Row 2: 7 cells with varying ratios
	row2 := m.flexBox.NewRow()
	row2Ratios := [][]int{{2, 4}, {2, 4}, {3, 4}, {3, 4}, {3, 4}, {4, 4}, {4, 4}}
	for i, r := range row2Ratios {
		idx := 3 + i
		ratioX, ratioY := r[0], r[1]
		cell := flexbox.NewCell(ratioX, ratioY)
		cell.SetContentGenerator(func(w, h int) string {
			return lipgloss.NewStyle().
				Width(w).Height(h).
				Align(lipgloss.Center, lipgloss.Center).
				Render(fmt.Sprintf("%s\n[%d:%d]", cellLabels[idx], ratioX, ratioY))
		})
		row2.AddCells(cell)
	}

	// Row 3: 5 cells, center shows border name
	row3 := m.flexBox.NewRow()
	row3Ratios := [][]int{{2, 5}, {3, 5}, {10, 5}, {3, 5}, {2, 5}}
	for i, r := range row3Ratios {
		idx := 10 + i
		ratioX, ratioY := r[0], r[1]
		cell := flexbox.NewCell(ratioX, ratioY)
		if i == 2 {
			// Center cell shows border name
			cell.SetContentGenerator(func(w, h int) string {
				return lipgloss.NewStyle().
					Width(w).Height(h).
					Align(lipgloss.Center, lipgloss.Center).
					Render(fmt.Sprintf("%s (t)\n[%d:%d]", borderNames[m.borderType], ratioX, ratioY))
			})
		} else {
			cell.SetContentGenerator(func(w, h int) string {
				return lipgloss.NewStyle().
					Width(w).Height(h).
					Align(lipgloss.Center, lipgloss.Center).
					Render(fmt.Sprintf("%s\n[%d:%d]", cellLabels[idx], ratioX, ratioY))
			})
		}
		row3.AddCells(cell)
	}

	// Row 4: 5 cells with varying height ratios
	row4 := m.flexBox.NewRow()
	row4Ratios := [][]int{{1, 4}, {1, 3}, {1, 2}, {1, 3}, {1, 4}}
	for i, r := range row4Ratios {
		idx := 15 + i
		ratioX, ratioY := r[0], r[1]
		cell := flexbox.NewCell(ratioX, ratioY)
		if idx != 17 { // Skip content for cell R (too small for text + border)
			cell.SetContentGenerator(func(w, h int) string {
				return lipgloss.NewStyle().
					Width(w).Height(h).
					Align(lipgloss.Center, lipgloss.Center).
					Render(fmt.Sprintf("%s\n[%d:%d]", cellLabels[idx], ratioX, ratioY))
			})
		}
		row4.AddCells(cell)
	}

	m.flexBox.AddRows([]*flexbox.Row{row1, row2, row3, row4})
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
		m.width = msg.Width
		m.height = msg.Height
		m.flexBox.SetWidth(msg.Width)
		m.flexBox.SetHeight(msg.Height)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "a":
			m.showAbout = !m.showAbout
		case "b":
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

var aboutStyle = lipgloss.NewStyle().
	Padding(2, 4).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#874BFD")).
	Background(lipgloss.Color("#1a1a2e"))

func (m *model) View() string {
	content := m.flexBox.Render()
	if m.showAbout {
		overlay := aboutStyle.Render(aboutText)
		content = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, overlay,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#1a1a2e")))
	}
	return content
}
