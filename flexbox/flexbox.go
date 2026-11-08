package flexbox

import "github.com/charmbracelet/lipgloss"

// FlexBox is a responsive grid of rows, each holding cells side by side. Rows stack
// vertically and share the box height by ratio or fixed size.
type FlexBox struct {
	// style to apply to the gridbox itself
	style         lipgloss.Style
	styleAncestor bool

	// width is fixed width of the box
	width int
	// height is fixed height of the box
	height int
	// fixedRowHeight will lock row height to a number, this disabless responsivness
	fixedRowHeight int

	rows []*Row

	// recalculateFlag indicates if next render should make calculations regarding
	// the rows objects height
	recalculateFlag bool

	// rowAlign controls horizontal alignment when joining rows of varying width
	// Default: lipgloss.Left (for backwards compatibility)
	rowAlign lipgloss.Position
}

// New initializes a FlexBox of the given size with no rows and an empty style.
func New(width, height int) *FlexBox {
	r := &FlexBox{
		width:           width,
		height:          height,
		fixedRowHeight:  -1,
		style:           lipgloss.NewStyle(),
		recalculateFlag: false,
		rowAlign:        lipgloss.Left,
	}
	return r
}

// SetStyle replaces the style. Width and height keys are removed because the box's size
// is set with SetWidth and SetHeight. A border, padding or margin shrinks the area the
// rows share.
func (r *FlexBox) SetStyle(style lipgloss.Style) *FlexBox {
	r.style = style.
		UnsetWidth().
		UnsetMaxWidth().
		UnsetHeight().
		UnsetMaxHeight()
	return r
}

// StylePassing sets whether the box's style is passed to its rows. A row passes it on to
// its cells only when its own StylePassing is enabled.
func (r *FlexBox) StylePassing(value bool) *FlexBox {
	r.styleAncestor = value
	return r
}

// UseStyle sets the style and enables StylePassing in one call.
func (r *FlexBox) UseStyle(style lipgloss.Style) *FlexBox {
	r.SetStyle(style)
	r.StylePassing(true)
	return r
}

// SetRowAlign places rows narrower than the box, such as rows of fixed-width cells, at
// lipgloss.Left, lipgloss.Center or lipgloss.Right. Full-width rows are unaffected. The
// default is lipgloss.Left.
func (r *FlexBox) SetRowAlign(align lipgloss.Position) *FlexBox {
	r.rowAlign = align
	return r
}

// NewRow initializes an empty Row of the box's width. Add it with AddRows or SetRows.
func (r *FlexBox) NewRow() *Row {
	rw := &Row{
		cells: []*Cell{},
		width: r.width,
		style: lipgloss.NewStyle(),
	}
	return rw
}

// AddRows appends rows to the box.
func (r *FlexBox) AddRows(rows []*Row) *FlexBox {
	r.rows = append(r.rows, rows...)
	r.setRecalculate()
	return r
}

// SetRows replaces all rows of the box.
func (r *FlexBox) SetRows(rows []*Row) *FlexBox {
	r.rows = rows
	r.setRecalculate()
	return r
}

// RowsLen returns the len of the rows slice
func (r *FlexBox) RowsLen() int {
	return len(r.rows)
}

// GetRow returns the row at the index, or nil. The box recalculates on its next render.
func (r *FlexBox) GetRow(index int) *Row {
	if index >= 0 && index < len(r.rows) {
		r.setRecalculate()
		return r.rows[index]
	}
	return nil
}

// GetRowCopy returns a copy of the row at the index, with copies of its cells, or nil.
// Reading the copy does not trigger a recalculation.
func (r *FlexBox) GetRowCopy(index int) *Row {
	if index >= 0 && index < len(r.rows) {
		rowCopy := r.rows[index].copy()
		return &rowCopy
	}
	return nil
}

// GetRowCellCopy returns a copy of the cell at cellIndex in the row at rowIndex, or nil.
// Reading the copy does not trigger a recalculation.
func (r *FlexBox) GetRowCellCopy(rowIndex, cellIndex int) *Cell {
	if rowIndex >= 0 && rowIndex < len(r.rows) {
		if cellIndex >= 0 && cellIndex < len(r.rows[rowIndex].cells) {
			cellCopy := r.rows[rowIndex].cells[cellIndex].copy()
			return &cellCopy
		}
	}
	return nil
}

// UpdateRow replaces the row at the index. The index must exist.
func (r *FlexBox) UpdateRow(index int, row *Row) *FlexBox {
	r.rows[index] = row
	r.setRecalculate()
	return r
}

// LockRowHeight pins every row to the given number of lines, which disables vertical
// scaling. A value of 0 or less restores ratio sizing.
func (r *FlexBox) LockRowHeight(value int) *FlexBox {
	r.fixedRowHeight = value
	return r
}

// SetHeight sets the FlexBox height
func (r *FlexBox) SetHeight(value int) *FlexBox {
	r.height = value
	r.setRecalculate()
	return r
}

// SetWidth sets the FlexBox width
func (r *FlexBox) SetWidth(value int) *FlexBox {
	r.width = value
	for _, row := range r.rows {
		row.setWidth(value)
	}
	return r
}

// GetHeight yields current FlexBox height
func (r *FlexBox) GetHeight() int {
	return r.getMaxHeight()
}

// GetWidth yields current FlexBox width
func (r *FlexBox) GetWidth() int {
	return r.getMaxWidth()
}

// Render lays out the rows if anything changed since the last render, renders each row,
// and joins them vertically with the row alignment, inside the box's own style.
func (r *FlexBox) Render() string {
	var inheritedStyle []lipgloss.Style
	if r.styleAncestor {
		inheritedStyle = append(inheritedStyle, r.style)
	}

	r.recalculate()
	var renderedRows []string
	for _, row := range r.rows {
		renderedRows = append(renderedRows, row.render(inheritedStyle...))
	}
	return r.style.
		Width(r.getContentWidth()).MaxWidth(r.getMaxWidth()).
		Height(r.getContentHeight()).MaxHeight(r.getMaxHeight()).
		Render(lipgloss.JoinVertical(r.rowAlign, renderedRows...))
}

// ForceRecalculate lays out the box and every row now instead of at the next render.
func (r *FlexBox) ForceRecalculate() {
	r.recalculate()
	for _, rw := range r.rows {
		rw.recalculate()
	}
}

// recalculate fetches the row height distribution slice and sets it on the rows
func (r *FlexBox) recalculate() {
	if r.recalculateFlag {
		if len(r.rows) > 0 {
			r.distributeRowsDimensions(r.calculateRowHeight())
		}
		r.unsetRecalculate()
	}
}

func (r *FlexBox) setRecalculate() {
	r.recalculateFlag = true
}

func (r *FlexBox) unsetRecalculate() {
	r.recalculateFlag = false
}

// calculateRowHeight calculates the height of each row and returns the distribution array
func (r *FlexBox) calculateRowHeight() (distribution []int) {
	// Legacy behavior: if global fixedRowHeight is set, use it for all rows
	if r.fixedRowHeight > 0 {
		var fixedRows []int
		for range r.rows {
			fixedRows = append(fixedRows, r.fixedRowHeight)
		}
		return fixedRows
	}

	// New behavior: handle per-row fixed heights
	var dynamicRowIndices []int
	var dynamicRowMatrix [][]int
	totalFixedHeight := 0

	distribution = make([]int, len(r.rows))

	// First pass: identify fixed and dynamic rows
	for i, row := range r.rows {
		// Check for the maximum fixed height among row and its cells
		maxFixedHeight := row.fixedHeight
		for _, cell := range row.cells {
			if cell.fixedHeight > maxFixedHeight {
				maxFixedHeight = cell.fixedHeight
			}
		}

		if maxFixedHeight > 0 {
			distribution[i] = maxFixedHeight
			totalFixedHeight += maxFixedHeight
		} else {
			dynamicRowIndices = append(dynamicRowIndices, i)
			// Get cell ratios for this dynamic row
			var cellValues []int
			for _, cell := range row.cells {
				cellValues = append(cellValues, cell.ratioY)
			}
			dynamicRowMatrix = append(dynamicRowMatrix, cellValues)
		}
	}

	// Second pass: distribute remaining space among dynamic rows
	if len(dynamicRowIndices) > 0 {
		remainingHeight := r.getContentHeight() - totalFixedHeight
		if remainingHeight < 0 {
			remainingHeight = 0
		}
		dynamicDistribution := calculateMatrixRatio(remainingHeight, dynamicRowMatrix)

		for i, rowIndex := range dynamicRowIndices {
			distribution[rowIndex] = dynamicDistribution[i]
		}
	}

	return distribution
}

// distributeRowsDimensions sets height and width of each row per distribution array
func (r *FlexBox) distributeRowsDimensions(ratioDistribution []int) {
	for index, row := range r.rows {
		row.setHeight(ratioDistribution[index])
		row.setWidth(r.getContentWidth())
	}
}

// getRowMatrix return the matrix of the cell heights for all the rows
func (r *FlexBox) getRowMatrix() (rowMatrix [][]int) {
	for _, row := range r.rows {
		var cellValues []int
		for _, cell := range row.cells {
			cellValues = append(cellValues, cell.ratioY)
		}
		rowMatrix = append(rowMatrix, cellValues)
	}
	return rowMatrix
}

func (r *FlexBox) getContentWidth() int {
	return r.getMaxWidth() - r.getExtraWidth()
}

func (r *FlexBox) getContentHeight() int {
	return r.getMaxHeight() - r.getExtraHeight()
}

func (r *FlexBox) getMaxWidth() int {
	return r.width
}

func (r *FlexBox) getMaxHeight() int {
	return r.height
}

func (r *FlexBox) getExtraWidth() int {
	return r.style.GetHorizontalFrameSize()
}

func (r *FlexBox) getExtraHeight() int {
	return r.style.GetVerticalFrameSize()
}
