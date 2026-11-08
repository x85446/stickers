package flexbox

import "github.com/charmbracelet/lipgloss"

// HorizontalFlexBox is a responsive grid of columns, each holding cells stacked
// vertically. Columns sit side by side and share the box width by ratio.
type HorizontalFlexBox struct {
	// style to apply to the gridbox itself
	style         lipgloss.Style
	styleAncestor bool

	// width is fixed width of the box
	width int
	// height is fixed height of the box
	height int

	// fixedColumnWidth will lock column width to a number, this disabless responsivness
	fixedColumnWidth int

	columns []*Column

	// recalculateFlag indicates if next render should make calculations regarding
	// the columns objects height
	recalculateFlag bool

	// columnAlign places the columns inside the box when together they are narrower than it
	columnAlign lipgloss.Position
}

// NewHorizontal initializes a HorizontalFlexBox of the given size with no columns and an
// empty style.
func NewHorizontal(width, height int) *HorizontalFlexBox {
	r := &HorizontalFlexBox{
		width:            width,
		height:           height,
		fixedColumnWidth: -1,
		style:            lipgloss.NewStyle(),
		recalculateFlag:  false,
		columnAlign:      lipgloss.Left,
	}
	return r
}

// SetColumnAlign places the columns inside the box when together they are narrower than
// it, such as columns of a locked width, at lipgloss.Left, lipgloss.Center or
// lipgloss.Right. The default is lipgloss.Left.
func (r *HorizontalFlexBox) SetColumnAlign(align lipgloss.Position) *HorizontalFlexBox {
	r.columnAlign = align
	return r
}

// SetStyle replaces the style. Width and height keys are removed because the box's size
// is set with SetWidth and SetHeight. A border, padding or margin shrinks the area the
// columns share.
func (r *HorizontalFlexBox) SetStyle(style lipgloss.Style) *HorizontalFlexBox {
	r.style = style.
		UnsetWidth().
		UnsetMaxWidth().
		UnsetHeight().
		UnsetMaxHeight()
	return r
}

// StylePassing sets whether the box's style is passed to its columns. A column passes it
// on to its cells only when its own StylePassing is enabled.
func (r *HorizontalFlexBox) StylePassing(value bool) *HorizontalFlexBox {
	r.styleAncestor = value
	return r
}

// UseStyle sets the style and enables StylePassing in one call.
func (r *HorizontalFlexBox) UseStyle(style lipgloss.Style) *HorizontalFlexBox {
	r.SetStyle(style)
	r.StylePassing(true)
	return r
}

// NewColumn initializes an empty Column. Add it with AddColumns or SetColumns.
func (r *HorizontalFlexBox) NewColumn() *Column {
	rw := &Column{
		cells: []*Cell{},
		width: r.width,
		style: lipgloss.NewStyle(),
	}
	return rw
}

// AddColumns appends columns to the box.
func (r *HorizontalFlexBox) AddColumns(columns []*Column) *HorizontalFlexBox {
	r.columns = append(r.columns, columns...)
	r.setRecalculate()
	return r
}

// SetColumns replaces all columns of the box.
func (r *HorizontalFlexBox) SetColumns(columns []*Column) *HorizontalFlexBox {
	r.columns = columns
	r.setRecalculate()
	return r
}

// ColumnsLen returns the len of the columns slice
func (r *HorizontalFlexBox) ColumnsLen() int {
	return len(r.columns)
}

// GetColumn returns the column at the index, or nil. The box recalculates on its next
// render.
func (r *HorizontalFlexBox) GetColumn(index int) *Column {
	if index >= 0 && index < len(r.columns) {
		r.setRecalculate()
		return r.columns[index]
	}
	return nil
}

// GetColumnCopy returns a copy of the column at the index, with copies of its cells, or
// nil. Reading the copy does not trigger a recalculation.
func (r *HorizontalFlexBox) GetColumnCopy(index int) *Column {
	if index >= 0 && index < len(r.columns) {
		columnCopy := r.columns[index].copy()
		return &columnCopy
	}
	return nil
}

// GetColumnCellCopy returns a copy of the cell at cellIndex in the column at
// columnIndex, or nil. Reading the copy does not trigger a recalculation.
func (r *HorizontalFlexBox) GetColumnCellCopy(columnIndex, cellIndex int) *Cell {
	if columnIndex >= 0 && columnIndex < len(r.columns) {
		if cellIndex >= 0 && cellIndex < len(r.columns[columnIndex].cells) {
			cellCopy := r.columns[columnIndex].cells[cellIndex].copy()
			return &cellCopy
		}
	}
	return nil
}

// UpdateColumn replaces the column at the index. The index must exist.
func (r *HorizontalFlexBox) UpdateColumn(index int, column *Column) *HorizontalFlexBox {
	r.columns[index] = column
	r.setRecalculate()
	return r
}

// LockColumnWidth pins every column to the given number of columns, which disables
// horizontal scaling. A value of 0 or less restores ratio sizing.
func (r *HorizontalFlexBox) LockColumnWidth(value int) *HorizontalFlexBox {
	r.fixedColumnWidth = value
	return r
}

// SetHeight sets the FlexBox height
func (r *HorizontalFlexBox) SetHeight(value int) *HorizontalFlexBox {
	r.height = value
	for _, column := range r.columns {
		column.setHeight(value)
	}
	return r
}

// SetWidth sets the FlexBox width
func (r *HorizontalFlexBox) SetWidth(value int) *HorizontalFlexBox {
	r.width = value
	r.setRecalculate()
	return r
}

// GetHeight yields current FlexBox height
func (r *HorizontalFlexBox) GetHeight() int {
	return r.getMaxHeight()
}

// GetWidth yields current FlexBox width
func (r *HorizontalFlexBox) GetWidth() int {
	return r.getMaxWidth()
}

// Render lays out the columns if anything changed since the last render, renders each
// column, joins them horizontally, places them with the column alignment, and wraps the
// result in the box's own style.
func (r *HorizontalFlexBox) Render() string {
	var inheritedStyle []lipgloss.Style
	if r.styleAncestor {
		inheritedStyle = append(inheritedStyle, r.style)
	}

	r.recalculate()
	var renderedColumns []string
	for _, column := range r.columns {
		renderedColumns = append(renderedColumns, column.render(inheritedStyle...))
	}
	joined := lipgloss.PlaceHorizontal(r.getContentWidth(), r.columnAlign,
		lipgloss.JoinHorizontal(lipgloss.Top, renderedColumns...))
	return r.style.
		Width(r.getContentWidth()).MaxWidth(r.getMaxWidth()).
		Height(r.getContentHeight()).MaxHeight(r.getMaxHeight()).
		Render(joined)
}

// ForceRecalculate lays out the box and every column now instead of at the next render.
func (r *HorizontalFlexBox) ForceRecalculate() {
	r.recalculate()
	for _, rw := range r.columns {
		rw.recalculate()
	}
}

// recalculate fetches the column height distribution slice and sets it on the columns
func (r *HorizontalFlexBox) recalculate() {
	if r.recalculateFlag {
		if len(r.columns) > 0 {
			r.distributeColumnsDimensions(r.calculateColumnWidth())
		}
		r.unsetRecalculate()
	}
}

func (r *HorizontalFlexBox) setRecalculate() {
	r.recalculateFlag = true
}

func (r *HorizontalFlexBox) unsetRecalculate() {
	r.recalculateFlag = false
}

// calculateColumnWidth calculates the width of each column and returns the distribution array
func (r *HorizontalFlexBox) calculateColumnWidth() (distribution []int) {
	if r.fixedColumnWidth > 0 {
		var fixedColumns []int
		for range r.columns {
			fixedColumns = append(fixedColumns, r.fixedColumnWidth)
		}
		return fixedColumns
	}
	return calculateMatrixRatio(r.getContentWidth(), r.getColumnMatrix())
}

// distributeColumnsDimensions sets height and width of each column per distribution array
func (r *HorizontalFlexBox) distributeColumnsDimensions(ratioDistribution []int) {
	for index, column := range r.columns {
		column.setHeight(r.getContentHeight())
		column.setWidth(ratioDistribution[index])
	}
}

// getColumnMatrix return the matrix of the cell widths for all the columns
func (r *HorizontalFlexBox) getColumnMatrix() (columnMatrix [][]int) {
	for _, column := range r.columns {
		var cellValues []int
		for _, cell := range column.cells {
			cellValues = append(cellValues, cell.ratioX)
		}
		columnMatrix = append(columnMatrix, cellValues)
	}
	return columnMatrix
}

func (r *HorizontalFlexBox) getContentWidth() int {
	return r.getMaxWidth() - r.getExtraWidth()
}

func (r *HorizontalFlexBox) getContentHeight() int {
	return r.getMaxHeight() - r.getExtraHeight()
}

func (r *HorizontalFlexBox) getMaxWidth() int {
	return r.width
}

func (r *HorizontalFlexBox) getMaxHeight() int {
	return r.height
}

func (r *HorizontalFlexBox) getExtraWidth() int {
	return r.style.GetHorizontalFrameSize()
}

func (r *HorizontalFlexBox) getExtraHeight() int {
	return r.style.GetVerticalFrameSize()
}
