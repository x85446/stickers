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
	flexBox    *flexbox.FlexBox
	borderType int  // 0=none, 1=normal, 2=rounded, 3=thick, 4=double
	useFixed   bool // Toggle fixed vs dynamic
}

var borderTypes = []lipgloss.Border{
	{}, // no border
	lipgloss.NormalBorder(),
	lipgloss.RoundedBorder(),
	lipgloss.ThickBorder(),
	lipgloss.DoubleBorder(),
}

var borderNames = []string{
	"Filled Background",
	"Normal Border",
	"Rounded Border",
	"Thick Border",
	"Double Border",
}

func main() {
	m := model{
		flexBox:    flexbox.New(0, 0),
		borderType: 1,    // Start with normal border
		useFixed:   true, // Start with fixed sizes enabled
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
		m.flexBox.SetWidth(msg.Width)
		m.flexBox.SetHeight(msg.Height)
		m.rebuildFlexBox()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "t":
			m.borderType = (m.borderType + 1) % len(borderTypes)
			m.rebuildFlexBox()
		case "f":
			m.useFixed = !m.useFixed
			m.rebuildFlexBox()
		}
	}
	return m, nil
}

func (m *model) rebuildFlexBox() {
	// Clear and rebuild
	m.flexBox.SetRows([]*flexbox.Row{})

	// Row 1: Header-like row (fixed height when enabled)
	row1 := m.flexBox.NewRow()
	if m.useFixed {
		row1.SetFixedHeight(5)
	}

	// Three cells: Fixed sidebar | Dynamic content | Fixed info
	cell1 := flexbox.NewCell(1, 1).SetStyle(m.getCellStyle(0))
	if m.useFixed {
		cell1.SetFixedWidth(25) // Fixed sidebar width
	}
	cell1.SetContentGenerator(func(w, h int) string {
		return lipgloss.NewStyle().
			Width(w).Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Sidebar\n%dx%d", w, h))
	})

	cell2 := flexbox.NewCell(3, 1).SetStyle(m.getCellStyle(1))
	cell2.SetContentGenerator(func(w, h int) string {
		fixedText := "Dynamic"
		if m.useFixed {
			fixedText = "Fixed Mode"
		}
		return lipgloss.NewStyle().
			Width(w).Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Header Area (%s)\n%dx%d", fixedText, w, h))
	})

	cell3 := flexbox.NewCell(1, 1).SetStyle(m.getCellStyle(2))
	if m.useFixed {
		cell3.SetFixedWidth(20) // Fixed info panel width
	}
	cell3.SetContentGenerator(func(w, h int) string {
		return lipgloss.NewStyle().
			Width(w).Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Info\n%dx%d", w, h))
	})

	row1.AddCells(cell1, cell2, cell3)

	// Row 2: Main content row (dynamic height)
	row2 := m.flexBox.NewRow()

	// Seven cells with mixed fixed/dynamic widths
	cells := []struct {
		ratio      int
		fixedWidth int
		label      string
	}{
		{2, 15, "Nav"},      // Fixed narrow nav
		{2, 0, "Col1"},      // Dynamic
		{3, 0, "Main"},      // Dynamic main area
		{3, 0, "Col2"},      // Dynamic
		{3, 0, "Col3"},      // Dynamic
		{4, 0, "Content"},   // Dynamic
		{4, 30, "Details"},  // Fixed details panel
	}

	for i, c := range cells {
		cell := flexbox.NewCell(c.ratio, 4).SetStyle(m.getCellStyle(i + 3))
		if m.useFixed && c.fixedWidth > 0 {
			cell.SetFixedWidth(c.fixedWidth)
		}
		idx := i
		label := c.label
		cell.SetContentGenerator(func(w, h int) string {
			sizeInfo := fmt.Sprintf("%dx%d", w, h)
			if m.useFixed && cells[idx].fixedWidth > 0 {
				sizeInfo = fmt.Sprintf("Fixed: %dx%d", w, h)
			}
			return lipgloss.NewStyle().
				Width(w).Height(h).
				Align(lipgloss.Center, lipgloss.Center).
				Render(fmt.Sprintf("%s\n%s", label, sizeInfo))
		})
		row2.AddCells(cell)
	}

	// Row 3: Mixed row with central focus
	row3 := m.flexBox.NewRow()

	cell31 := flexbox.NewCell(2, 5).SetStyle(m.getCellStyle(10))
	if m.useFixed {
		cell31.SetFixedWidth(20)
	}
	cell31.SetContentGenerator(func(w, h int) string {
		return lipgloss.NewStyle().
			Width(w).Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Left\n%dx%d", w, h))
	})

	cell32 := flexbox.NewCell(3, 5).SetStyle(m.getCellStyle(11))
	cell32.SetContentGenerator(func(w, h int) string {
		return lipgloss.NewStyle().
			Width(w).Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Dynamic Center\n%dx%d", w, h))
	})

	cell33 := flexbox.NewCell(10, 5).SetStyle(m.getCellStyle(12))
	cell33.SetContentGenerator(func(w, h int) string {
		modeText := "All Dynamic"
		if m.useFixed {
			modeText = "Mixed Fixed/Dynamic"
		}
		return lipgloss.NewStyle().
			Width(w).Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("%s\n%s\n%dx%d\n\n't' = toggle border\n'f' = toggle fixed",
				borderNames[m.borderType], modeText, w, h))
	})

	cell34 := flexbox.NewCell(3, 5).SetStyle(m.getCellStyle(13))
	cell34.SetContentGenerator(func(w, h int) string {
		return lipgloss.NewStyle().
			Width(w).Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Dynamic\n%dx%d", w, h))
	})

	cell35 := flexbox.NewCell(2, 5).SetStyle(m.getCellStyle(14))
	if m.useFixed {
		cell35.SetFixedWidth(20)
	}
	cell35.SetContentGenerator(func(w, h int) string {
		return lipgloss.NewStyle().
			Width(w).Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Right\n%dx%d", w, h))
	})

	row3.AddCells(cell31, cell32, cell33, cell34, cell35)

	// Row 4: Footer-like row (fixed height when enabled)
	row4 := m.flexBox.NewRow()
	if m.useFixed {
		row4.SetFixedHeight(4)
	}

	// Five cells with varying ratios
	for i := 0; i < 5; i++ {
		ratios := []int{1, 1, 1, 1, 1}
		fixedWidths := []int{12, 0, 0, 0, 12}  // First and last fixed

		cell := flexbox.NewCell(ratios[i], 4).SetStyle(m.getCellStyle(15 + i))
		if m.useFixed && fixedWidths[i] > 0 {
			cell.SetFixedWidth(fixedWidths[i])
		}

		idx := i
		cell.SetContentGenerator(func(w, h int) string {
			label := fmt.Sprintf("F%d", idx+1)
			if m.useFixed && fixedWidths[idx] > 0 {
				label = fmt.Sprintf("Fix%d", idx+1)
			}
			return lipgloss.NewStyle().
				Width(w).Height(h).
				Align(lipgloss.Center, lipgloss.Center).
				Render(fmt.Sprintf("%s\n%dx%d", label, w, h))
		})
		row4.AddCells(cell)
	}

	m.flexBox.AddRows([]*flexbox.Row{row1, row2, row3, row4})
}

func (m *model) getCellStyle(colorIndex int) lipgloss.Style {
	if m.borderType == 0 {
		// Filled background
		return lipgloss.NewStyle().Background(lipgloss.Color(colors[colorIndex]))
	}
	// Border with type
	return lipgloss.NewStyle().
		Border(borderTypes[m.borderType]).
		BorderForeground(lipgloss.Color(colors[colorIndex]))
}

func (m *model) View() string {
	header := lipgloss.NewStyle().Bold(true).
		Render(fmt.Sprintf("Mixed Fixed/Dynamic Layout | 't' = border | 'f' = toggle fixed | Current: %s, Mode: %s",
			borderNames[m.borderType],
			map[bool]string{true: "Fixed", false: "Dynamic"}[m.useFixed]))
	return header + "\n" + m.flexBox.Render()
}