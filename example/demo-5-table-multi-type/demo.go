// Package demo5 is the Table Multi-Type demo: a Table of typed columns that sort numerically.
package demo5

import (
	"fmt"
	"os"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gocarina/gocsv"
	"github.com/x85446/stickers/example/sample"
	"github.com/x85446/stickers/flexbox"
	"github.com/x85446/stickers/table"
)

var selectedValue string = "\nselect something with spacebar or enter"

// Model is the Bubble Tea model for demo 5
type Model struct {
	table     *table.Table
	infoBox   *flexbox.FlexBox
	headers   []string
	showAbout bool
	width     int
	height    int
}

const aboutText = `Demo 5: Table Multi-Type

Table with typed columns for proper sorting.

Unlike demo-4 (string-only), this demo uses gocsv
to parse CSV data with typed columns (int, string).

This enables proper numeric sorting: integers sort
numerically (1, 2, 10) instead of lexically (1, 10, 2).

Navigation:
- Arrow keys: Move cursor
- Ctrl+S: Sort by column (numeric or alpha)
- Enter/Space: Select cell value
- Type to filter

Press 'a' to close | 'q' to quit`

// SampleData represents a row from the CSV file
type SampleData struct {
	ID         int    `csv:"id"`
	FirstName  string `csv:"First Name"`
	LastName   string `csv:"Last Name"`
	Age        int    `csv:"Age"`
	Occupation string `csv:"Occupation"`
}

// New creates a new demo 5 model
func New() *Model {
	var sampleData []*SampleData
	if err := gocsv.UnmarshalBytes(sample.CSV, &sampleData); err != nil {
		panic(err)
	}

	headers := []string{"id", "First Name", "Last Name", "Age", "Occupation"}
	ratio := []int{1, 10, 10, 5, 10}
	minSize := []int{4, 5, 5, 2, 5}

	var s string
	var i int
	types := []any{i, s, s, i, s}

	m := &Model{
		table:   table.NewTable(0, 0, headers),
		infoBox: flexbox.New(0, 0).SetHeight(7),
		headers: headers,
	}
	// set types
	if _, err := m.table.SetTypes(types...); err != nil {
		panic(err)
	}
	// setup dimensions
	m.table.SetRatio(ratio).SetMinWidth(minSize)
	// set style passing
	m.table.SetStylePassing(true)
	// add rows
	// with multi type table we have to convert our rows to []any first which is a bit of a pain
	var orderedRows [][]any
	for _, row := range sampleData {
		orderedRows = append(orderedRows, []any{
			row.ID, row.FirstName, row.LastName, row.Age, row.Occupation,
		})
	}
	m.table.MustAddRows(orderedRows)

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
	p := tea.NewProgram(m, tea.WithAltScreen())
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
	case tea.KeyMsg:
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
		case "enter", " ":
			selectedValue = m.table.GetCursorValue()
			m.infoBox.GetRow(0).GetCell(1).SetContent("\nselected cell: " + selectedValue)
		case "backspace":
			m.filterWithStr(msg.String())
		default:
			if len(msg.String()) == 1 {
				r := msg.Runes[0]
				if unicode.IsLetter(r) || unicode.IsDigit(r) {
					m.filterWithStr(msg.String())
				}
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
func (m *Model) View() string {
	content := lipgloss.JoinVertical(lipgloss.Left, m.table.Render(), m.infoBox.Render())
	if m.showAbout {
		overlay := aboutStyle.Render(aboutText)
		content = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, overlay,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#1a1a2e")))
	}
	return content
}
