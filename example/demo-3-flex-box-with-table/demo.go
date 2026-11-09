// Package demo3 is the FlexBox with Table demo: a Table inside a cell of a FlexBox, filled from CSV data.
package demo3

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/x85446/stickers/example/sample"
	"github.com/x85446/stickers/flexbox"
	"github.com/x85446/stickers/table"
)

var (
	styleRow        = lipgloss.NewStyle().Align(lipgloss.Center).Foreground(lipgloss.Color("#000000")).Bold(true)
	styleBlank      = lipgloss.NewStyle()
	styleBackground = lipgloss.NewStyle().Align(lipgloss.Center).Background(lipgloss.Color("#ffffff"))
	style1          = lipgloss.NewStyle().Align(lipgloss.Center).Background(lipgloss.Color("#f368e0"))
	style2          = lipgloss.NewStyle().Align(lipgloss.Center).Background(lipgloss.Color("#ff9f43"))
	style3          = lipgloss.NewStyle().Align(lipgloss.Center).Background(lipgloss.Color("#ee5253"))
	style4          = lipgloss.NewStyle().Align(lipgloss.Center).Background(lipgloss.Color("#0abde3"))
	style5          = lipgloss.NewStyle().Align(lipgloss.Center).Background(lipgloss.Color("#10ac84"))
	style6          = lipgloss.NewStyle().Align(lipgloss.Center).Background(lipgloss.Color("#222f3e"))
)

// Model is the Bubble Tea model for demo 3
type Model struct {
	flexBox   *flexbox.FlexBox
	table     *table.Table
	headers   []string
	showAbout bool
	width     int
	height    int
}

const aboutText = `Demo 3: FlexBox with Table

Embedded table inside a flexbox layout with CSV data.

Shows how to nest a Table component inside a FlexBox cell.
The table is placed in the center cell of row 2: the cell's
content generator sizes the table to the cell and renders it.

Navigation:
- Arrow keys: Move cursor in table
- Ctrl+S: Sort by current column
- Enter/Space: Copy cell value to random flexbox cells
- Type letters: Filter current column
- Backspace: Clear filter character

Press 'a' to close | 'q' to quit`

// New creates a new demo 3 model
func New() *Model {
	data, err := csv.NewReader(bytes.NewReader(sample.CSV)).ReadAll()
	if err != nil {
		panic(err)
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
		flexBox: flexbox.New(0, 0).SetStyle(styleBackground),
		table:   table.NewTable(0, 0, headers),
		headers: headers,
	}

	m.table.SetRatio(ratio).SetMinWidth(minSize)
	if _, err := m.table.AddRows(rows); err != nil {
		panic(err)
	}
	m.table.SetStylePassing(true)

	r1 := m.flexBox.NewRow().AddCells(
		flexbox.NewCell(5, 5).SetStyle(style2),
		flexbox.NewCell(2, 5).SetStyle(style3),
		flexbox.NewCell(5, 5).SetStyle(style5),
	).SetStyle(styleRow)
	r2 := m.flexBox.NewRow().AddCells(
		flexbox.NewCell(1, 5).SetStyle(style6),
		// the table takes the size of its cell on every render
		flexbox.NewCell(10, 5).SetStyle(styleBlank).SetContentGenerator(func(width, height int) string {
			return m.table.SetWidth(width).SetHeight(height).Render()
		}),
		flexbox.NewCell(1, 5).SetStyle(style6),
	).SetStyle(styleRow)
	r3 := m.flexBox.NewRow().AddCells(
		flexbox.NewCell(1, 5).SetStyle(style5),
		flexbox.NewCell(1, 4).SetStyle(style4),
		flexbox.NewCell(1, 3).SetStyle(style3),
		flexbox.NewCell(1, 4).SetStyle(style2),
		flexbox.NewCell(1, 5).SetStyle(style1),
	).SetStyle(styleRow)

	_rows := []*flexbox.Row{r1, r2, r3}
	m.flexBox.AddRows(_rows)

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
		windowHeight := msg.Height
		windowWidth := msg.Width
		m.width = windowWidth
		m.height = windowHeight
		m.flexBox.SetWidth(windowWidth)
		m.flexBox.SetHeight(windowHeight)
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
			cellString := m.table.GetCursorValue()
			// add content to random boxes on flex box
			for ir := 0; ir < m.flexBox.RowsLen(); ir++ {
				// don't' want it on the middle row
				if ir == 1 {
					continue
				}
				// not handling error for example script
				for ic := 0; ic < m.flexBox.GetRow(ir).CellsLen(); ic++ {
					// adding a bit of randomness for fun
					if rand.Int()%2 == 0 {
						h := int(math.Floor(float64(m.flexBox.GetRowCellCopy(ir, ic).GetHeight()) / 2.0))
						m.flexBox.GetRow(ir).GetCell(ic).SetContent(strings.Repeat("\n", h) + cellString)
					} else {
						m.flexBox.GetRow(ir).GetCell(ic).SetContent("")
					}
				}
			}
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
	content := m.flexBox.Render()
	if m.showAbout {
		overlay := aboutStyle.Render(aboutText)
		content = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, overlay,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#1a1a2e")))
	}
	return content
}
