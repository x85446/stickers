# Stickers 👾

Stickers is a set of layout elements for terminal apps built with
[Bubble Tea](https://github.com/charmbracelet/bubbletea) and styled with
[Lip Gloss](https://github.com/charmbracelet/lipgloss). It ships two packages:

- `flexbox`: a responsive grid inspired by CSS flexbox. A `FlexBox` stacks rows
  of cells; a `HorizontalFlexBox` lines up columns of cells. Sizes come from
  ratios, fixed values, or both, and follow the terminal as it resizes.
- `table`: a scrollable, sortable, filterable table that uses `flexbox` for its
  layout.

This module lives at `github.com/x85446/stickers`. It is a fork of
[76creates/stickers](https://github.com/76creates/stickers).

![FlexBox with Table](.github/images/flex-box-with-table.gif)

## Install

```bash
go get github.com/x85446/stickers@latest
```

Requires Go 1.23 or later. Stickers depends on `lipgloss` v0.13 and
`bubbletea` v1.1.

## FlexBox

A `FlexBox` has a width and a height, holds rows, and each row holds cells. You
give the box the terminal size, give rows and cells ratios or fixed sizes, and
call `Render()`.

```go
package main

import (
	"fmt"
	"os"

	"github.com/x85446/stickers/flexbox"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct{ box *flexbox.FlexBox }

func newModel() model {
	box := flexbox.New(0, 0)
	box.AddRows([]*flexbox.Row{
		box.NewRow().AddCells(
			flexbox.NewCell(1, 1).SetContent("sidebar").SetFixedWidth(20),
			flexbox.NewCell(3, 1).SetContent("main"),
		),
		box.NewRow().AddCells(
			flexbox.NewCell(1, 1).SetContent("footer").
				SetStyle(lipgloss.NewStyle().Border(lipgloss.NormalBorder())),
		).SetFixedHeight(3),
	})
	return model{box: box}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.box.SetWidth(msg.Width)
		m.box.SetHeight(msg.Height)
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string { return m.box.Render() }

func main() {
	if _, err := tea.NewProgram(newModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
```

### Sizing rules

Every cell is created with `NewCell(ratioX, ratioY)`.

**Row heights.** A row with a fixed height takes exactly that many lines. The
remaining lines are split between the other rows in proportion to each row's
largest `ratioY`. Leftover lines go one each to the rows with the largest
ratios.

- `row.SetFixedHeight(n)` pins a row. `0` or less restores ratio sizing.
- `cell.SetFixedHeight(n)` asks for a fixed row height from a cell. The row
  takes the largest request among its cells.
- `box.LockRowHeight(n)` pins every row to `n` lines. `0` or less restores
  ratio sizing.

**Cell widths within a row.** Cells with a fixed width take exactly that many
columns. The remaining columns are split between the other cells in proportion
to `ratioX`, honoring `SetMinWidth`. If the fixed widths alone exceed the row,
the other cells get no columns.

- `cell.SetFixedWidth(n)` pins a cell. `0` or less restores ratio sizing.
- `cell.SetMinWidth(n)` keeps a ratio-sized cell at least `n` columns wide.

**Cell heights within a row.** Each cell gets `ratioY / maxRatioY` of the row
height, rounded down, and sits at the top of the row. Cells with the largest
ratio fill the row; the others are shorter. A cell short of its content shows
what fits, see below.

**Rows of different widths.** A row whose cells add up to less than the box
width, because of fixed widths, keeps that width. `box.SetRowAlign(pos)` with
`lipgloss.Left`, `lipgloss.Center` or `lipgloss.Right` places such rows inside
the box. The default is left.

### What a cell shows when it does not fit

A cell renders its frame first and its content second.

- The border and margins render complete whenever the cell has room for them.
  Content taller than the cell loses its bottom lines; the bottom border stays.
- Content is wrapped at the content width and clipped to the lines between the
  paddings.
- A cell too short for its frame shows its top lines. A bordered cell with two
  lines is a closed, empty frame; with one line it is the top border.
- A cell with zero lines renders as a blank line of its width, so a row keeps
  its columns. A cell with zero columns adds nothing to a row.
- A row never wraps its cells. Anything wider than the row, such as a
  fixed-width cell wider than the box, is cut at the row's edge.

### Content

- `cell.SetContent(s)` sets static text.
- `cell.SetContentGenerator(func(w, h int) string)` builds the text each time
  the cell renders. `w` and `h` are the area inside the border and margins,
  so the generator can lay out text for the space it actually has. Padding, if
  the style has any, is inside that area.
- `cell.GetWidth()` and `cell.GetHeight()` return the full allocated size,
  including the frame.

### Styles

Styles are `lipgloss.Style` values. Width and height keys are removed when a
style is set; the grid owns sizes.

- `SetStyle(style)` on a box, row, column or cell. Borders, padding and
  margins on a box or row shrink the area its children share.
- `StylePassing(true)` on a box passes its style to its rows; on a row or
  column it passes the style to its cells. A row with passing enabled inherits
  the box style first, then hands the result to its cells. Children inherit
  only keys they have not set themselves; margins and padding never pass.
- `UseStyle(style)` sets the style and enables passing in one call.
- `NewCell(x, y, flexbox.WithStyle(style))` sets a cell style at creation.

### Reading the grid back

`GetRow`, `GetCell` and `GetCellWithID` return pointers and mark the box or row
for recalculation on the next render, so use them when you intend to change
something. `GetRowCopy`, `GetRowCellCopy` and `GetCellCopy` return copies and
leave the layout alone, so use them to read sizes or content. `UpdateRow` and
`UpdateCellWithIndex` replace an element in place. Cell IDs default to the
cell's index in its row; `SetID` names one for `GetCellWithID`.

### HorizontalFlexBox

`NewHorizontal(width, height)` lines up columns side by side; cells stack
vertically inside each column.

- Column widths follow each column's largest cell `ratioX`.
  `LockColumnWidth(n)` pins every column to `n` columns.
- Inside a column, cell heights are split by `ratioY`, honoring
  `SetMinHeight`. Each cell's width is `ratioX / maxRatioX` of the column.
- `SetFixedWidth`, `SetFixedHeight` and `SetMinWidth` apply to cells of a
  `FlexBox` only. `SetMinHeight` applies to cells of a `HorizontalFlexBox`
  only.
- `GetColumn`, `GetColumnCopy`, `GetColumnCellCopy`, `UpdateColumn` and
  `ColumnsLen` mirror the `FlexBox` row methods.

## Table

A `Table` renders a one-line header, one-line rows, and a one-line footer with
the cursor position and table size. It scrolls vertically with the cursor and
horizontally when the columns' minimum widths do not fit.

```go
t := table.NewTable(0, 0, []string{"ID", "Name", "Score"})
t, err := t.SetTypes(0, "", 0.0) // int, string, float64 columns
if err != nil {
	return err
}
t.SetRatio([]int{1, 4, 2}).SetMinWidth([]int{4, 10, 6})
t.MustAddRows([][]any{
	{1, "Ada", 9.5},
	{2, "Linus", 8.0},
})

// on tea.WindowSizeMsg
t.SetWidth(w)
t.SetHeight(h) // includes the header and footer lines

// on keys
t.CursorDown()    // also CursorUp, CursorLeft, CursorRight
t.OrderByAsc(2)   // or OrderByDesc; GetOrder() reports the current sort
t.SetFilter(1, "li") // case-insensitive substring on column 1; UnsetFilter() clears

// in View
s := t.Render()
```

- **Types.** Every column is `string` until `SetTypes` is called with one zero
  value per column. Accepted types are `string`, `int`, `int8`, `int16`,
  `int32`, `float32` and `float64`. `SetTypes` clears existing rows.
- **Rows.** `AddRows` validates each row's length and cell types and returns
  the first error without adding anything. `MustAddRows` panics instead.
  `ClearRows` empties the table.
- **Sorting.** `OrderByAsc(i)` and `OrderByDesc(i)` sort by column `i` using
  the column's type, so numbers sort numerically. An unknown index is ignored.
  The header marks the sorted column with ▲ or ▼.
- **Filtering.** `SetFilter(i, s)` keeps rows whose column `i` contains `s`,
  ignoring case. The header marks the filtered column with ⑂ and the footer
  names the filter while the cursor is on that column. `GetFilter` returns the
  column, `-1` when none, and the string.
- **Cursor.** `GetCursorLocation` returns `x, y`; `GetCursorValue` returns the
  text under the cursor, or `""` when the table is empty.
- **Horizontal scrolling.** Columns are shown while their minimum widths fit
  the table width; moving the cursor past the visible range scrolls.
  `GetVisibleColumnRange` returns the first and last visible column. With all
  minimum widths at `0`, every column is always visible.
- **Styles.** `SetStyles(map[table.StyleKey]lipgloss.Style)` overrides any of
  `StyleKeyHeader`, `StyleKeyFooter`, `StyleKeyRows`, `StyleKeyRowsSubsequent`
  (every other row), `StyleKeyRowsCursor` and `StyleKeyCellCursor`; keys you
  leave out keep their defaults. `SetStylePassing(true)` passes the row styles
  down to the cells.
- **Sizing checks.** `SetRatio` and `SetMinWidth` must receive exactly one
  value per column, and ratios must be greater than zero. A mismatch calls
  `log.Fatalf` and ends the program.

![Table Multi-Type](.github/images/table-multi-type.gif)

## Demos

The `example` directory holds twelve demos and a combined program that cycles
through them.

```bash
make help       # lists every demo with a one-line description
make demo-all   # all demos: tab / shift+tab cycle, a about, q quit
make demo-7     # one demo
```

Every demo answers `a` for an about box and `q` to quit. In `demo-all` the
purple header names the demo, lists its own keys, then the shared keys.

| Demo | Shows | Keys |
|---|---|---|
| 1 FlexBox Simple | ratio-based rows and cells | |
| 2 FlexBox Horizontal | `HorizontalFlexBox` columns | |
| 3 FlexBox with Table | a `Table` inside a cell, CSV data | arrows, ctrl+s sort, enter copies the cell into the grid, letters filter |
| 4 Table Simple String | string columns, sorting, filtering | arrows, ctrl+s, enter, letters |
| 5 Table Multi-Type | typed columns via gocsv, numeric sorting | arrows, ctrl+s, enter, letters |
| 6 Nested Borders | recursive boxes through a content generator | b border |
| 7 Simple Borders | labeled cells, every border style, short cells | b border |
| 8 Fixed Row Heights | `SetFixedHeight` on rows | m fixed/dynamic |
| 9 Fixed Column Widths | `SetFixedWidth` on cells | m fixed/dynamic |
| 10 Mixed Fixed/Dynamic | fixed and ratio sizing in one grid | b border, m fixed/dynamic, h text |
| 11 Cell Config | changing a cell's fixed sizes at run time | b border, h text, C-W select a cell, then w, h, digits, enter, esc |
| 12 Row Alignment | `SetRowAlign` with rows narrower than the box | r row align |

Demos 3, 4 and 5 read `example/sample.csv` relative to the working directory,
so run them through `make` or from inside `example`.

![FlexBox Simple](.github/images/flex-box-simple.gif)

## Tests

```bash
go test ./...
```

The `flexbox` package tests cover the cell and row rendering rules above.
`example/demo-all` has a test that cycles every demo at several terminal sizes
and checks that each fits under the header.

## License

MIT, see [LICENSE](LICENSE).

***
Made with [Charm](https://charm.sh).

<a href="https://charm.sh/"><img alt="The Charm logo" src="https://stuff.charm.sh/charm-badge-unrounded.jpg" width="400"></a>
