// Package demo4 is the Table Simple String demo: a Table of string columns with sorting and filtering.
package demo4

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/x85446/stickers/example/sample"
	"github.com/x85446/stickers/v2/flexbox"
	"github.com/x85446/stickers/v2/table"
)

var selectedValue string = "\nselect something with spacebar or enter"

// Model is the Bubble Tea model for demo 4
type Model struct {
	table     *table.Table
	infoBox   *flexbox.FlexBox
	headers   []string
	showAbout bool
	width     int
	height    int
}

const aboutText = `Demo 4: Table Simple String

Navigable table with sorting and filtering.

Demonstrates the Table component with string data
loaded from a CSV file.

Navigation:
- Arrow keys: Move cursor
- Ctrl+S: Sort by current column (toggle asc/desc)
- Enter/Space: Select cell value
- Type letters/numbers: Filter current column
- Backspace: Clear filter character

Press 'a' to close | 'q' to quit`

// New creates a new demo 4 model
func New() *Model {
	data, err := csv.NewReader(bytes.NewReader(sample.CSV)).ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	headers := data[0]
	rows := make([][]any, len(data[1:]))
	for i, row := range data[1:] {
		rows[i] = make([]any, len(headers))
		for j, cell := range row {
			rows[i][j] = cell
		}
	}
	ratio := []int{1, 10, 10, 5, 10}
	minSize := []int{4, 5, 5, 2, 5}

	m := &Model{
		table:   table.NewTable(0, 0, headers),
		infoBox: flexbox.New(0, 0).SetHeight(7),
		headers: headers,
	}
	m.table.SetStylePassing(true)
	// setup
	m.table.SetRatio(ratio).SetMinWidth(minSize)
	// add rows
	if _, err := m.table.AddRows(rows); err != nil {
		panic(err)
	}

	// setup info box
	infoText := `
use the arrows to navigate
ctrl+s: sort by current column
alphanumerics: filter column
enter, spacebar: get column value
ctrl+c: quit
`
	r1 := m.infoBox.NewRow()
	r1.AddCells(
		flexbox.NewCell(1, 1).
			SetID("info").
			SetContent(infoText),
		flexbox.NewCell(1, 1).
			SetID("info").
			SetContent(selectedValue).
			SetStyle(lipgloss.NewStyle().Bold(true)),
	)
	m.infoBox.AddRows([]*flexbox.Row{r1})

	return m
}

// Run starts the demo as a standalone program
func Run() {
	m := New()
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

// Init implements tea.Model
func (m *Model) Init() tea.Cmd { return nil }

// Update implements tea.Model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetWidth(msg.Width)
		m.table.SetHeight(msg.Height - m.infoBox.GetHeight())
		m.infoBox.SetWidth(msg.Width)
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "a":
			m.showAbout = !m.showAbout
			return m, nil
		case "down":
			m.table.CursorDown()
		case "up":
			m.table.CursorUp()
		case "left":
			m.table.CursorLeft()
		case "right":
			m.table.CursorRight()
		case "ctrl+s":
			x, _ := m.table.GetCursorLocation()
			_, order := m.table.GetOrder()
			switch order {
			case table.SortingOrderAscending:
				m.table.OrderByDesc(x)
			case table.SortingOrderDescending:
				m.table.OrderByAsc(x)
			}
		case "enter", "space":
			selectedValue = m.table.GetCursorValue()
			m.infoBox.GetRow(0).GetCell(1).SetContent("\nselected cell: " + selectedValue)
		case "backspace":
			m.filterWithStr(msg.String())
		default:
			if r := []rune(msg.Text); len(r) == 1 && (unicode.IsLetter(r[0]) || unicode.IsDigit(r[0])) {
				m.filterWithStr(msg.Text)
			}
		}

	}
	return m, nil
}

// filterWithStr edits the filter of the cursor's column: a letter or digit extends it and
// backspace removes its last character.
func (m *Model) filterWithStr(key string) {
	column, filter := m.table.GetFilter()
	x, _ := m.table.GetCursorLocation()
	switch {
	case key != "backspace" && x != column:
		m.table.SetFilter(x, key)
	case key != "backspace":
		m.table.SetFilter(column, filter+key)
	case len(filter) > 1:
		m.table.SetFilter(column, filter[:len(filter)-1])
	case len(filter) == 1:
		m.table.UnsetFilter()
	}
}

var aboutStyle = lipgloss.NewStyle().
	Padding(2, 4).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#874BFD")).
	Background(lipgloss.Color("#1a1a2e"))

// View implements tea.Model
func (m *Model) View() tea.View {
	view := tea.NewView(m.render())
	view.AltScreen = true
	return view
}

// render returns the demo as text.
func (m *Model) render() string {
	content := lipgloss.JoinVertical(lipgloss.Left, m.table.Render(), m.infoBox.Render())
	if m.showAbout {
		overlay := aboutStyle.Render(aboutText)
		content = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, overlay)
	}
	return content
}
