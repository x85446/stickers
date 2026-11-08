package table

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/x85446/stickers/flexbox"
)

var (
	tableDefaultHeaderStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#7158e2")).
				Foreground(lipgloss.Color("#ffffff"))
	tableDefaultFooterStyle = tableDefaultHeaderStyle.Align(lipgloss.Right).Height(1)
	tableDefaultRowsStyle   = lipgloss.NewStyle().
				Background(lipgloss.Color("#4b4b4b")).
				Foreground(lipgloss.Color("#ffffff"))
	tableDefaultRowsSubsequentStyle = lipgloss.NewStyle().
					Background(lipgloss.Color("#3d3d3d")).
					Foreground(lipgloss.Color("#ffffff"))
	tableDefaultRowsCursorStyle = lipgloss.NewStyle().
					Background(lipgloss.Color("#f7b731")).
					Foreground(lipgloss.Color("#000000")).
					Bold(true)
	tableDefaultCellCursorStyle = lipgloss.NewStyle().
					Background(lipgloss.Color("#f6e58d")).
					Foreground(lipgloss.Color("#000000"))

	tableDefaultSortAscChar  = "▲"
	tableDefaultSortDescChar = "▼"
	tableDefaultFilterChar   = "⑂"

	tableDefaultStyles = map[StyleKey]lipgloss.Style{
		StyleKeyHeader:         tableDefaultHeaderStyle,
		StyleKeyFooter:         tableDefaultFooterStyle,
		StyleKeyRows:           tableDefaultRowsStyle,
		StyleKeyRowsSubsequent: tableDefaultRowsSubsequentStyle,
		StyleKeyRowsCursor:     tableDefaultRowsCursorStyle,
		StyleKeyCellCursor:     tableDefaultCellCursorStyle,
	}
)

// StyleKey names a part of the table whose style SetStyles can override.
type StyleKey int

// The parts of the table that take a style.
const (
	StyleKeyHeader         StyleKey = iota // the header line
	StyleKeyFooter                         // the footer line with the cursor position
	StyleKeyRows                           // odd rows
	StyleKeyRowsSubsequent                 // even rows, so rows alternate for readability
	StyleKeyRowsCursor                     // the row under the cursor
	StyleKeyCellCursor                     // the cell under the cursor
)

// Table is a responsive table with a one-line header, one-line rows and a one-line
// footer, laid out with flexbox. It scrolls vertically with the cursor and horizontally
// when the columns' minimum widths do not fit the table width.
type Table struct {
	// columnRatio ratio of the columns, is applied to rows as well
	columnRatio []int
	// columnMinWidth minimal width of the column
	columnMinWidth []int

	// columnHeaders column text headers
	columnHeaders []string
	columnType    []any
	rows          [][]any

	// filteredRows is the rows that are visible after filtering
	filteredRows   [][]any
	filteredColumn int
	filterString   string

	// orderColumnIndex notes which column is used for sorting
	// -1 means that no column is sorted
	orderedColumnIndex int
	// orderedColumnPhase remarks if the sort is asc or desc, basically works like a toggle
	// 0 indicates desc sorting, 1 indicates
	orderedColumnPhase SortingOrderKey

	// rowsTopIndex top visible index
	rowsTopIndex int
	cursorIndexY int
	cursorIndexX int
	// cursorDirection is the direction of the last cursor move; it decides from which side
	// the visible column range is built
	cursorDirection cursorDirection

	// columnVisibleLeftIndex and columnVisibleRightIndex are used to calculate the columns on the screen
	columnVisibleLeftIndex  int
	columnVisibleRightIndex int

	height int
	width  int

	rowsBoxHeight int
	// rowHeight fixed row height value, maybe this should be optional?
	rowHeight int

	styles map[StyleKey]lipgloss.Style
	// stylePassing if true, styles are passed all the way down from box to cell
	stylePassing bool

	headerBox *flexbox.FlexBox
	rowsBox   *flexbox.FlexBox

	// these flags indicate weather we should update rows and headers flex boxes
	updateRowsFlag    bool
	updateHeadersFlag bool
}

// NewTable initializes a Table of the given size with one string column per header, all
// with ratio 1 and no minimum width.
func NewTable(width, height int, columnHeaders []string) *Table {
	var columnRatio, columnMinWidth []int
	for range columnHeaders {
		columnRatio = append(columnRatio, 1)
		columnMinWidth = append(columnMinWidth, 0)
	}

	// by default all columns are of type string
	var defaultType string
	var defaultTypes []any
	for range columnHeaders {
		defaultTypes = append(defaultTypes, defaultType)
	}

	styles := defaultStyles()

	r := &Table{
		columnHeaders:           columnHeaders,
		columnRatio:             columnRatio,
		columnMinWidth:          columnMinWidth,
		cursorIndexX:            0,
		cursorIndexY:            0,
		cursorDirection:         cursorDirectionUpLeft,
		columnVisibleLeftIndex:  0,
		columnVisibleRightIndex: 0,

		columnType:         defaultTypes,
		orderedColumnIndex: -1,
		orderedColumnPhase: SortingOrderDescending,

		filteredColumn: -1,
		filterString:   "",

		height: height,
		width:  width,
		// when optional header/footer is set rework this
		rowsBoxHeight: height - 2,

		rowsTopIndex: 0,
		rowHeight:    1,

		headerBox: flexbox.New(width, 1).SetStyle(tableDefaultHeaderStyle),
		rowsBox:   flexbox.New(width, height-2),

		styles:       styles,
		stylePassing: false,
	}
	r.recalculateVisibleColumnRange()
	r.setHeadersUpdate()
	return r
}

// SetRatio sets the width ratio of every column. It needs exactly one value per column,
// each greater than 0, and panics otherwise.
func (r *Table) SetRatio(values []int) *Table {
	if len(values) != len(r.columnHeaders) {
		panic(fmt.Sprintf("table: %d ratios for %d columns", len(values), len(r.columnHeaders)))
	}
	for _, val := range values {
		if val < 1 {
			panic("table: a column ratio must be greater than 0")
		}
	}
	r.columnRatio = values
	r.setHeadersUpdate()
	r.setRowsUpdate()
	return r
}

// SetTypes sets the type of every column from one zero value per column, such as 0, ""
// or 0.0, and removes all rows, so call it before adding rows. Accepted types are string,
// int, int8, int16, int32, int64, float32 and float64.
func (r *Table) SetTypes(columnTypes ...any) (*Table, error) {
	if len(columnTypes) != len(r.columnHeaders) {
		return r, errors.New("column types not the same len as headers")
	}
	for i, t := range columnTypes {
		if !isOrdered(t) {
			message := fmt.Sprintf(
				"column of type %s on index %d is not of type Ordered", reflect.TypeOf(t).String(), i,
			)
			return r, ErrorBadType{msg: message}
		}
	}
	r.cursorIndexY, r.cursorIndexX = 0, 0
	r.rows = [][]any{}
	r.columnType = columnTypes
	r.setRowsUpdate()
	return r, nil
}

// SetMinWidth sets the minimum width of every column. It needs exactly one value per
// column and panics otherwise. Minimum widths also decide how many columns are visible at
// once, see GetVisibleColumnRange.
func (r *Table) SetMinWidth(values []int) *Table {
	if len(values) != len(r.columnHeaders) {
		panic(fmt.Sprintf("table: %d minimum widths for %d columns", len(values), len(r.columnHeaders)))
	}
	r.columnMinWidth = values
	r.setHeadersUpdate()
	r.setRowsUpdate()
	return r
}

// SetHeight sets the height of the table. The header and footer take one line each and
// the rows get the rest.
func (r *Table) SetHeight(value int) *Table {
	r.height = value
	// we deduct two to take header/footer into the account
	r.rowsBoxHeight = value - 2
	r.rowsBox.SetHeight(r.rowsBoxHeight)
	r.setRowsUpdate()
	r.setTopRow()
	return r
}

// SetWidth sets the width of the table and recomputes which columns are visible.
func (r *Table) SetWidth(value int) *Table {
	r.width = value
	r.rowsBox.SetWidth(value)
	r.headerBox.SetWidth(value)
	r.recalculateVisibleColumnRange()
	return r
}

// GetVisibleColumnRange returns the first and last visible column. Columns are visible
// while their minimum widths fit the table width, starting from the cursor's column;
// with all minimum widths at 0 every column is visible.
func (r *Table) GetVisibleColumnRange() (int, int) {
	return r.columnVisibleLeftIndex, r.columnVisibleRightIndex
}

// SetStyles overrides the styles of the given parts of the table. Parts left out keep
// their default style.
func (r *Table) SetStyles(styles map[StyleKey]lipgloss.Style) *Table {
	mergedStyles := defaultStyles()
	for key, style := range styles {
		mergedStyles[key] = style
	}
	r.styles = mergedStyles
	r.setRowsUpdate()
	r.setHeadersUpdate()
	return r
}

// SetStylePassing sets whether the header and row styles are passed down to their cells.
func (r *Table) SetStylePassing(value bool) *Table {
	r.stylePassing = value
	r.headerBox.StylePassing(value)
	r.rowsBox.StylePassing(value)
	r.setRowsUpdate()
	r.setHeadersUpdate()
	return r
}

// UnsetFilter removes the filter so every row is visible again.
func (r *Table) UnsetFilter() *Table {
	r.filterString = ""
	r.filteredColumn = -1
	r.setTopRow()
	r.setRowsUpdate()
	r.setHeadersUpdate()
	return r
}

// SetFilter keeps only the rows whose value in the column contains s, ignoring case. An
// index out of range changes nothing. The header marks the column and the footer names
// the filter while the cursor is on that column.
func (r *Table) SetFilter(columnIndex int, s string) *Table {
	if columnIndex < len(r.columnHeaders) {
		r.filterString = s
		r.filteredColumn = columnIndex

		r.setRowsUpdate()
	}
	return r
}

// GetFilter returns the filtered column, -1 when there is none, and the filter string.
func (r *Table) GetFilter() (columnIndex int, s string) {
	return r.filteredColumn, r.filterString
}

// CursorDown moves the cursor one row down, scrolling when it leaves the visible rows.
func (r *Table) CursorDown() *Table {
	if r.cursorIndexY+1 < len(r.filteredRows) {
		r.cursorDirection = r.cursorDirection.setDown()
		r.cursorIndexY++
		r.setTopRow()
		r.setRowsUpdate()
	}
	return r
}

// CursorUp moves the cursor one row up, scrolling when it leaves the visible rows.
func (r *Table) CursorUp() *Table {
	if r.cursorIndexY-1 > -1 {
		r.cursorDirection = r.cursorDirection.setUp()
		r.cursorIndexY--
		r.setTopRow()
		r.setRowsUpdate()
	}
	return r
}

// CursorLeft moves the cursor one column left, scrolling when it leaves the visible
// columns.
func (r *Table) CursorLeft() *Table {
	if r.cursorIndexX-1 > -1 {
		r.cursorDirection = r.cursorDirection.setLeft()
		r.cursorIndexX--
		r.setRowsUpdate()
		r.checkVisibleColumnRange()
	}
	return r
}

// CursorRight moves the cursor one column right, scrolling when it leaves the visible
// columns.
func (r *Table) CursorRight() *Table {
	if r.cursorIndexX+1 < len(r.columnHeaders) {
		r.cursorDirection = r.cursorDirection.setRight()
		r.cursorIndexX++
		r.setRowsUpdate()
		r.checkVisibleColumnRange()
	}
	return r
}

// GetCursorLocation returns the cursor's column and row, counted from 0. The row counts
// over the filtered rows.
func (r *Table) GetCursorLocation() (int, int) {
	return r.cursorIndexX, r.cursorIndexY
}

// GetCursorValue returns the text of the cell under the cursor, or "" when the table has
// no rows.
func (r *Table) GetCursorValue() string {
	// handle 0 rows situation and when table is not active
	if len(r.filteredRows) == 0 || r.cursorIndexX < 0 || r.cursorIndexY < 0 {
		return ""
	}
	return getStringFromOrdered(r.filteredRows[r.cursorIndexY][r.cursorIndexX])
}

// AddRows appends rows. Every row must have one value per column, of the column's type;
// the first row that does not returns an ErrorRowLen, ErrorBadType or ErrorBadCellType
// and nothing is added.
func (r *Table) AddRows(rows [][]any) (*Table, error) {
	// check for errors
	for _, row := range rows {
		if err := r.validateRow(row...); err != nil {
			return r, err
		}
	}
	// append rows
	r.rows = append(r.rows, rows...)

	r.applyFilter()
	r.setRowsUpdate()
	return r, nil
}

// MustAddRows is AddRows that panics on error.
func (r *Table) MustAddRows(rows [][]any) *Table {
	if _, err := r.AddRows(rows); err != nil {
		panic(err)
	}
	return r
}

// ClearRows removes every row, for example before reloading data in an update loop.
func (r *Table) ClearRows() *Table {
	r.rows = make([][]any, 0, 10)
	r.setRowsUpdate()
	return r
}

// Render returns the header, the visible rows and a footer that reads "x:y / w:h", the
// cursor position and the rows area size, prefixed by the filter when the cursor is on the
// filtered column.
func (r *Table) Render() string {
	r.updateRows()
	r.updateHeader()

	statusMessage := fmt.Sprintf(
		"%d:%d / %d:%d ",
		r.cursorIndexX,
		r.cursorIndexY,
		r.rowsBox.GetWidth(),
		r.rowsBox.GetHeight(),
	)
	if r.cursorIndexX == r.filteredColumn {
		statusMessage = fmt.Sprintf("filtered by: %q / %s", r.filterString, statusMessage)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		r.headerBox.Render(),
		r.rowsBox.Render(),
		r.styles[StyleKeyFooter].Width(r.width).Render(statusMessage),
	)
}

// defaultStyles returns a copy of the default styles, so overrides on one table never
// reach the package defaults or another table.
func defaultStyles() map[StyleKey]lipgloss.Style {
	styles := make(map[StyleKey]lipgloss.Style, len(tableDefaultStyles))
	for key, style := range tableDefaultStyles {
		styles[key] = style
	}
	return styles
}

func (r *Table) setRowsUpdate() {
	r.updateRowsFlag = true
}

func (r *Table) unsetRowsUpdate() {
	r.updateRowsFlag = false
}

func (r *Table) setHeadersUpdate() {
	r.updateHeadersFlag = true
}

func (r *Table) unsetHeadersUpdate() {
	r.updateHeadersFlag = false
}

// validateRow checks the row for validity, number of cells must match table header length
// and header types per cell as well
func (r *Table) validateRow(cells ...any) error {
	var message string
	// check row len
	if len(cells) != len(r.columnType) {
		message = fmt.Sprintf(
			"len of row[%d] does not equal number of columns[%d]", len(cells), len(r.columnType),
		)
		return ErrorRowLen{msg: message}
	}
	// check cell type
	for i, c := range cells {
		switch c.(type) {
		case string, int, int8, int16, int32, int64, float32, float64:
			// check if the cell matches the type of the column
			if reflect.TypeOf(c) != reflect.TypeOf(r.columnType[i]) {
				message = fmt.Sprintf(
					"type of the cell[%v] on index %d not matching type of the column[%v]",
					reflect.TypeOf(c), i, reflect.TypeOf(r.columnType[i]),
				)
				return ErrorBadCellType{msg: message}
			}
		default:
			message = fmt.Sprintf(
				"type[%v] on index %d not matching Ordered interface types", reflect.TypeOf(c), i,
			)
			return ErrorBadType{msg: message}
		}
	}
	return nil
}

// updateHeader recomputes the header of the table
func (r *Table) updateHeader() *Table {
	if !r.updateHeadersFlag {
		return r
	}
	var cells []*flexbox.Cell
	r.headerBox.SetStyle(r.styles[StyleKeyHeader])

	leftmostColumnIndex, rightmostColumnIndex := r.columnVisibleLeftIndex, r.columnVisibleRightIndex
	if r.width == 0 {
		// this is the case when we initialize the table and width is not set yet
		rightmostColumnIndex = len(r.columnHeaders) - 1
	}
	for index := leftmostColumnIndex; index <= rightmostColumnIndex; index++ {
		title := r.columnHeaders[index]
		cells = append(
			cells,
			flexbox.NewCell(r.columnRatio[index], 1).SetMinWidth(r.columnMinWidth[index]).SetContentGenerator(func(maxX, maxY int) string {
				// titleSuffix at the moment can be sort and filter characters
				// filtering symbol should be visible always, if possible of course, and as far right as possible
				// there should be a minimum of space bar between two symbols and symbol and row to the right
				var titleSuffix string
				// add sorting symbol if the sorting is active on the column
				if r.orderedColumnIndex == index {
					if r.orderedColumnPhase == SortingOrderDescending {
						titleSuffix = " " + tableDefaultSortDescChar
					} else if r.orderedColumnPhase == SortingOrderAscending {
						titleSuffix = " " + tableDefaultSortAscChar
					}
				}

				// add filtering symbol if the filtering is active on the column
				if r.filteredColumn == index && r.filterString != "" {
					// add at least one space bar between char to the left, and one to the right
					titleSuffix = titleSuffix + strings.Repeat(
						" ", int(math.Max(
							1,
							float64(
								maxX-utf8.RuneCountInString(title+titleSuffix)-2,
							),
						)),
					) + tableDefaultFilterChar + " "
				}

				// if title and suffix exceed width trim the title
				if maxX-utf8.RuneCountInString(title+titleSuffix) < 0 {
					// this will be the cae only when sort is on and filter is off
					// add one space bar between sort and column to the right
					if utf8.RuneCountInString(titleSuffix) == 2 {
						titleSuffix = titleSuffix + " "
					}
					// trim the title
					title = title[0:int(math.Max(0, float64(maxX-utf8.RuneCountInString(titleSuffix))))]
				}
				return title + titleSuffix
			}),
		)
	}
	r.headerBox.SetRows(
		[]*flexbox.Row{
			r.headerBox.NewRow().StylePassing(r.stylePassing).AddCells(cells...),
		},
	)
	r.unsetHeadersUpdate()
	return r
}

// updateRows recomputes the rows of the table
// calculate the visible rows top/bottom indexes
// create rows and their cells with styles depending on state
func (r *Table) updateRows() {
	if !r.updateRowsFlag {
		return
	}
	if r.rowsBoxHeight < 0 {
		r.unsetRowsUpdate()
		return
	}
	r.applyFilter()

	// calculate the bottom most visible row index
	rowsBottomIndex := r.rowsTopIndex + r.rowsBoxHeight
	if rowsBottomIndex > len(r.filteredRows) {
		rowsBottomIndex = len(r.filteredRows)
	}

	var rows []*flexbox.Row
	for ir, columns := range r.filteredRows[r.rowsTopIndex:rowsBottomIndex] {
		// irCorrected is corrected row index since we iterate only visible rows
		irCorrected := ir + r.rowsTopIndex

		var cells []*flexbox.Cell
		for ic, column := range columns[r.columnVisibleLeftIndex : r.columnVisibleRightIndex+1] {
			icCorrected := ic + r.columnVisibleLeftIndex
			// initialize column cell
			c := flexbox.NewCell(r.columnRatio[icCorrected], r.rowHeight).
				SetMinWidth(r.columnMinWidth[icCorrected]).
				SetContent(getStringFromOrdered(column))
			// update style if cursor is on the cell, otherwise it's inherited from the row
			if irCorrected == r.cursorIndexY && icCorrected == r.cursorIndexX {
				c.SetStyle(r.styles[StyleKeyCellCursor])
			}
			cells = append(cells, c)
		}
		// initialize new row from the rows box and add generated cells
		rw := r.rowsBox.NewRow().StylePassing(r.stylePassing).AddCells(cells...)

		// rows have three styles, normal, subsequent and selected
		// normal and subsequent rows should differ for readability
		if irCorrected == r.cursorIndexY {
			rw.SetStyle(r.styles[StyleKeyRowsCursor])
		} else if irCorrected%2 == 0 || irCorrected == 0 {
			rw.SetStyle(r.styles[StyleKeyRowsSubsequent])
		} else {
			rw.SetStyle(r.styles[StyleKeyRows])
		}

		rows = append(rows, rw)
	}

	// lock row height, this might get optional at some point
	r.rowsBox.LockRowHeight(r.rowHeight)
	r.rowsBox.SetRows(rows)
	r.unsetRowsUpdate()
}

// applyFilter filters column n by a value s
func (r *Table) applyFilter() *Table {
	// sending empty string should reset the filtering
	if r.filterString == "" {
		r.filteredRows = r.rows
		return r
	}
	var filteredRows [][]any
	for _, row := range r.rows {
		cellValue := getStringFromOrdered(row[r.filteredColumn])
		// convert to lower, not sure if anybody needs case-sensitive filtering
		// if you are reading this and need it, open up an issue :zap:
		if strings.Contains(strings.ToLower(cellValue), strings.ToLower(r.filterString)) {
			filteredRows = append(filteredRows, row)
		}
	}
	r.filteredRows = filteredRows
	r.setTopRow()
	r.setHeadersUpdate()
	return r
}

// setTopRow calculates the row top index used when deciding what is visible
func (r *Table) setTopRow() {
	// if rows are empty set y to 0, retain x pos
	// will be useful for filtering
	if len(r.filteredRows) == 0 {
		r.cursorIndexY = 0
	} else if r.cursorIndexY > len(r.filteredRows) {
		// when filtering if cursor is higher than row length
		// set it to the bottom of the list
		r.cursorIndexY = len(r.filteredRows) - 1
	}

	// case when cursor is in between top or bottom visible row
	if r.cursorIndexY >= r.rowsTopIndex && r.cursorIndexY < r.rowsTopIndex+r.rowsBoxHeight {
		// if cursor is on the last item in row, adjust the row top
		if r.cursorIndexY == len(r.filteredRows)-1 {
			// if all rows can fit on screen
			if len(r.filteredRows) <= r.rowsBoxHeight {
				r.rowsTopIndex = 0
				return
			}
			// fit max rows on the table
			r.rowsTopIndex = r.cursorIndexY - (r.rowsBoxHeight - 1)
		} else if r.cursorIndexY > len(r.filteredRows)-1 && len(r.filteredRows) != 0 {
			r.cursorIndexY = len(r.filteredRows) - 1
		}
		return
	}

	// if cursor is above the top
	if r.cursorIndexY < r.rowsTopIndex {
		if r.cursorIndexY == len(r.filteredRows)-1 {
			// if all rows can fit on screen
			if len(r.filteredRows) <= r.rowsBoxHeight {
				r.rowsTopIndex = 0
				return
			}
			// fit max rows on the table
			r.rowsTopIndex = r.cursorIndexY - (r.rowsBoxHeight - 1)
			return
		}
		r.rowsTopIndex = r.cursorIndexY
		return
	}

	if r.cursorIndexY > r.rowsTopIndex {
		r.rowsTopIndex = r.cursorIndexY - r.rowsBoxHeight + 1
		return
	}
}

func (r *Table) recalculateVisibleColumnRange() {
	var totalWidth int
	r.setRowsUpdate()
	r.setHeadersUpdate()
	if r.cursorDirection.isRight() {
		r.columnVisibleLeftIndex, totalWidth = r.columnIndexSeekLeft(r.cursorIndexX, 0)
		r.columnVisibleRightIndex, totalWidth = r.columnIndexSeekRight(r.cursorIndexX+1, totalWidth)
	} else {
		r.columnVisibleRightIndex, totalWidth = r.columnIndexSeekRight(r.cursorIndexX, totalWidth)
		r.columnVisibleLeftIndex, totalWidth = r.columnIndexSeekLeft(r.cursorIndexX-1, totalWidth)
	}
	return
}

func (r *Table) columnIndexSeekLeft(index int, widthAdded int) (int, int) {
	for i := index; i >= 0; i-- {
		if widthAdded+r.columnMinWidth[i] > r.width {
			return i + 1, widthAdded
		}
		widthAdded += r.columnMinWidth[i]
		if widthAdded == r.width || i == 0 {
			return i, widthAdded
		}
	}
	return 0, widthAdded
}

func (r *Table) columnIndexSeekRight(index int, widthAdded int) (int, int) {
	for i := index; i < len(r.columnHeaders); i++ {
		if widthAdded+r.columnMinWidth[i] > r.width {
			return i - 1, widthAdded
		}
		widthAdded += r.columnMinWidth[i]
		if widthAdded == r.width || i == len(r.columnHeaders)-1 {
			return i, widthAdded
		}
	}
	return len(r.columnHeaders) - 1, widthAdded
}

// checkVisibleColumnRange should be executed only after the cursor is moved left or right
func (r *Table) checkVisibleColumnRange() {
	if r.cursorIndexX < r.columnVisibleLeftIndex || r.cursorIndexX > r.columnVisibleRightIndex {
		r.recalculateVisibleColumnRange()
	}
	return
}
