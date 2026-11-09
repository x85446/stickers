package table

// CursorDown moves the cursor one row down, scrolling when it leaves the visible rows.
func (t *Table) CursorDown() *Table {
	t.settle()
	if t.cursorIndexY+1 < len(t.filteredRows) {
		t.moveCursorTo(t.cursorIndexY + 1)
	}
	return t
}

// CursorUp moves the cursor one row up, scrolling when it leaves the visible rows.
func (t *Table) CursorUp() *Table {
	t.settle()
	if t.cursorIndexY > 0 {
		t.moveCursorTo(t.cursorIndexY - 1)
	}
	return t
}

// CursorLeft moves the cursor one column left, scrolling when it leaves the visible
// columns.
func (t *Table) CursorLeft() *Table {
	if t.cursorIndexX > 0 {
		t.cursorIndexX--
		t.cursorMovedRight = false
		t.showCursorColumn()
		t.rowsStale = true
	}
	return t
}

// CursorRight moves the cursor one column right, scrolling when it leaves the visible
// columns.
func (t *Table) CursorRight() *Table {
	if t.cursorIndexX+1 < len(t.columnHeaders) {
		t.cursorIndexX++
		t.cursorMovedRight = true
		t.showCursorColumn()
		t.rowsStale = true
	}
	return t
}

// GetCursorLocation returns the cursor's column and row, counted from 0. The row counts
// over the rows the filter keeps, in sorted order.
func (t *Table) GetCursorLocation() (int, int) {
	t.settle()
	return t.cursorIndexX, t.cursorIndexY
}

// GetCursorValue returns the text of the cell under the cursor, or "" when the table has
// no rows.
func (t *Table) GetCursorValue() string {
	row := t.rowUnderCursor()
	if row == nil || t.cursorIndexX >= len(row) {
		return ""
	}
	return format(row[t.cursorIndexX])
}

// GetCursorRow returns a copy of the values of the row under the cursor, or nil when the
// table has no rows.
func (t *Table) GetCursorRow() []any {
	row := t.rowUnderCursor()
	if row == nil {
		return nil
	}
	return append([]any(nil), row...)
}

// rowUnderCursor returns the row under the cursor, or nil when the table has no rows.
func (t *Table) rowUnderCursor() []any {
	t.settle()
	if len(t.filteredRows) == 0 {
		return nil
	}
	return t.filteredRows[t.cursorIndexY]
}

// moveCursorTo puts the cursor on the row at index and scrolls the rows to show it.
func (t *Table) moveCursorTo(index int) {
	t.cursorRow, t.topRow = index, t.rowsTopIndex
	t.cursorIndexY, t.rowsTopIndex = t.position()
	t.topRow = t.rowsTopIndex
	t.rowsStale = true
}

// position returns the row the cursor is on and the first visible row for the rows the
// filter keeps and the height of the rows area.
//
// The cursor is on the row it was moved to, or on the last row while that one is gone.
// The rows scroll as little as they have to from where the last cursor move left them
// to show the cursor, and leave no blank lines under the last row while rows are hidden
// above it. Nothing is stored, so the same changes always lead to the same position,
// however often the table is read in between.
func (t *Table) position() (cursor, top int) {
	rows, height := len(t.filteredRows), t.rowsBoxHeight
	cursor = max(min(t.cursorRow, rows-1), 0)
	top = t.topRow
	if cursor < top {
		top = cursor
	} else if cursor >= top+height {
		top = cursor - height + 1
	}
	return cursor, max(min(top, rows-height), 0)
}

// GetVisibleColumnRange returns the first and last visible column. Columns are visible
// while their minimum widths fit the table width, starting from the cursor's column;
// with all minimum widths at 0 every column is visible. The cursor's column is always
// visible.
func (t *Table) GetVisibleColumnRange() (int, int) {
	return t.columnVisibleLeftIndex, t.columnVisibleRightIndex
}

// showCursorColumn scrolls horizontally when the cursor left the visible columns.
func (t *Table) showCursorColumn() {
	if t.cursorIndexX < t.columnVisibleLeftIndex || t.cursorIndexX > t.columnVisibleRightIndex {
		t.fitColumns()
	}
}

// fitColumns chooses the visible columns again and, when they change, has the header
// and the rows built for them.
func (t *Table) fitColumns() {
	left, right := t.columnsThatFit()
	if left != t.columnVisibleLeftIndex || right != t.columnVisibleRightIndex {
		t.columnVisibleLeftIndex, t.columnVisibleRightIndex = left, right
		t.headerStale, t.rowsStale, t.columnsStale = true, true, true
	}
}

// columnsThatFit returns the first and last column to show: the cursor's column, then
// as many columns as fit the table width by their minimum widths. A cursor that moved
// right is the last visible column and columns are added to its left first; otherwise
// it is the first and columns are added to its right first. A table without columns
// shows none, as 0 and -1.
func (t *Table) columnsThatFit() (left, right int) {
	if len(t.columnHeaders) == 0 {
		return 0, -1
	}
	left, right = t.cursorIndexX, t.cursorIndexX
	used := t.columnMinWidth[t.cursorIndexX]
	growLeft := func() {
		for left > 0 && used+t.columnMinWidth[left-1] <= t.width {
			left--
			used += t.columnMinWidth[left]
		}
	}
	growRight := func() {
		for right < len(t.columnHeaders)-1 && used+t.columnMinWidth[right+1] <= t.width {
			right++
			used += t.columnMinWidth[right]
		}
	}
	if t.cursorMovedRight {
		growLeft()
		growRight()
	} else {
		growRight()
		growLeft()
	}
	return left, right
}
