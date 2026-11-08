package flexbox

import (
	"strconv"

	"github.com/charmbracelet/lipgloss"
)

// Row holds the cells of one row of a FlexBox. Its height comes from the box; its cells'
// widths and heights come from their ratios and fixed sizes. Rows stack vertically.
type Row struct {
	// style of the row
	style         lipgloss.Style
	styleAncestor bool

	cells []*Cell

	height int
	width  int

	// fixedHeight if > 0, this row will have a fixed height instead of ratio-based
	fixedHeight int

	// recalculateFlag indicates if next render should make calculations regarding
	// the cells objects height/width
	recalculateFlag bool
}

// AddCells appends cells to the row. A cell without an ID gets its index as its ID.
func (r *Row) AddCells(cells ...*Cell) *Row {
	r.cells = append(r.cells, cells...)
	for i, cell := range r.cells {
		if cell.id == "" {
			cell.id = strconv.Itoa(i)
		}
	}
	r.setRecalculate()
	return r
}

// SetFixedHeight gives the row a fixed height instead of a ratio-based one. A value of 0
// or less restores ratio sizing.
func (r *Row) SetFixedHeight(value int) *Row {
	r.fixedHeight = value
	r.setRecalculate()
	return r
}

// CellsLen returns the number of cells.
func (r *Row) CellsLen() int {
	return len(r.cells)
}

// GetCell returns the cell at the index, or nil. The row recalculates on its next render.
func (r *Row) GetCell(index int) *Cell {
	if index >= 0 && index < len(r.cells) {
		r.setRecalculate()
		return r.cells[index]
	}
	return nil
}

// GetCellCopy returns a copy of the cell at the index, or nil. Reading the copy does not
// trigger a recalculation.
func (r *Row) GetCellCopy(index int) *Cell {
	if index >= 0 && index < len(r.cells) {
		c := r.cells[index].copy()
		return &c
	}
	return nil
}

// GetCellWithID returns the cell with the ID, or nil. The row recalculates on its next
// render.
func (r *Row) GetCellWithID(id string) *Cell {
	for _, c := range r.cells {
		if c.id == id {
			r.setRecalculate()
			return c
		}
	}
	return nil
}

// UpdateCellWithIndex replaces the cell at the index. An index out of range changes
// nothing.
func (r *Row) UpdateCellWithIndex(index int, cell *Cell) {
	if index >= 0 && len(r.cells) > 0 && index < len(r.cells) {
		r.cells[index] = cell
		r.setRecalculate()
	}
}

// SetStyle replaces the style. Width and height keys are removed because the box owns
// the row's size. A border, padding or margin on the row shrinks the area its cells share.
func (r *Row) SetStyle(style lipgloss.Style) *Row {
	r.style = style.
		UnsetWidth().
		UnsetMaxWidth().
		UnsetHeight().
		UnsetMaxHeight()

	return r
}

// StylePassing sets whether the row's style, including what it inherits from the box, is
// passed to its cells.
func (r *Row) StylePassing(value bool) *Row {
	r.styleAncestor = value
	return r
}

// UseStyle sets the style and enables StylePassing in one call.
func (r *Row) UseStyle(style lipgloss.Style) *Row {
	r.SetStyle(style)
	r.StylePassing(true)
	return r
}

func (r *Row) setHeight(value int) {
	r.height = value
	r.setRecalculate()
}

func (r *Row) setWidth(value int) {
	r.width = value
	r.setRecalculate()
}

func (r *Row) render(inherited ...lipgloss.Style) string {
	var inheritedStyle []lipgloss.Style

	for _, style := range inherited {
		r.style = r.style.Inherit(style)
	}

	// intentionally applied after row inherits the box style
	if r.styleAncestor {
		inheritedStyle = append(inheritedStyle, r.style)
	}

	r.recalculate()
	var renderedCells []string
	for _, cell := range r.cells {
		renderedCells = append(renderedCells, cell.render(inheritedStyle...))
	}
	// The cells are laid out already: anything wider than the row, such as a fixed-width
	// cell wider than the box, is clipped at the row's edge, never wrapped. A row narrower
	// than the box keeps its width so the box can align it.
	joined := clipWidth(lipgloss.JoinHorizontal(lipgloss.Top, renderedCells...), r.getContentWidth())
	return r.style.
		Height(r.getContentHeight()).MaxHeight(r.getMaxHeight()).
		Render(joined)
}

func (r *Row) setRecalculate() {
	r.recalculateFlag = true
}

func (r *Row) unsetRecalculate() {
	r.recalculateFlag = false
}

// recalculate fetches the cell's height/width distribution slices and sets it on the cells
func (r *Row) recalculate() {
	if r.recalculateFlag {
		if len(r.cells) > 0 {
			r.distributeCellDimensions(r.calculateCellsDimensions())
		}
		r.unsetRecalculate()
	}
}

// distributeCellDimensions sets height of each row per distribution array
func (r *Row) distributeCellDimensions(xMatrix, yMatrix []int) {
	for index, x := range xMatrix {
		r.cells[index].width = x
		r.cells[index].height = yMatrix[index]
	}
}

// calculateCellsDimensions calculates the height and width of the each cell
func (r *Row) calculateCellsDimensions() (xMatrix, yMatrix []int) {
	// calculate the cell height, it uses fixed combined ratio since the height of each cell
	// is individual and does not stack, row height will be calculated using the ratio of the
	// highest cell in the slice
	cellYMatrix, cellYMatrixMax := r.getCellHeightMatrix()
	// reminder not needed here due to how combined ratio is passed
	yMatrix, _ = distributeToMatrix(r.getContentHeight(), cellYMatrixMax, cellYMatrix)

	// calculate the cell width matrix with support for fixed widths
	xMatrix = r.calculateCellWidth()

	return xMatrix, yMatrix
}

// calculateCellWidth calculates width distribution for cells, respecting fixed widths
func (r *Row) calculateCellWidth() []int {
	totalWidth := r.getContentWidth()
	widthMatrix := make([]int, len(r.cells))

	// First pass: allocate fixed widths and calculate remaining space
	remainingWidth := totalWidth
	var dynamicCells []int  // indices of cells with dynamic widths
	var dynamicRatios []int // ratios for dynamic cells
	var minWidths []int     // min widths for dynamic cells
	hasMinWidth := false

	for i, cell := range r.cells {
		if cell.fixedWidth > 0 {
			widthMatrix[i] = cell.fixedWidth
			remainingWidth -= cell.fixedWidth
		} else {
			dynamicCells = append(dynamicCells, i)
			dynamicRatios = append(dynamicRatios, cell.ratioX)
			minWidths = append(minWidths, cell.minWidth)
			if cell.minWidth > 0 {
				hasMinWidth = true
			}
		}
	}

	// Second pass: distribute remaining width among dynamic cells
	if len(dynamicCells) > 0 && remainingWidth > 0 {
		var dynamicWidths []int
		if hasMinWidth {
			dynamicWidths = calculateRatioWithMinimum(remainingWidth, dynamicRatios, minWidths)
		} else {
			dynamicWidths = calculateRatio(remainingWidth, dynamicRatios)
		}

		// Assign calculated widths to dynamic cells
		for i, cellIdx := range dynamicCells {
			widthMatrix[cellIdx] = dynamicWidths[i]
		}
	} else if len(dynamicCells) > 0 {
		// No remaining width or negative - distribute zeros
		for _, cellIdx := range dynamicCells {
			widthMatrix[cellIdx] = 0
		}
	}

	return widthMatrix
}

// getCellHeightMatrix return the matrix of the cell height in cells and the max value in it
func (r *Row) getCellHeightMatrix() (cellHeightMatrix []int, max int) {
	max = 0
	for _, cell := range r.cells {
		if cell.ratioY > max {
			max = cell.ratioY
		}
		cellHeightMatrix = append(cellHeightMatrix, cell.ratioY)
	}
	return cellHeightMatrix, max
}

// getCellWidthMatrix return the matrix of the cell width in cells
func (r *Row) getCellWidthMatrix() (cellWidthMatrix []int) {
	for _, cell := range r.cells {
		cellWidthMatrix = append(cellWidthMatrix, cell.ratioX)
	}
	return cellWidthMatrix
}

func (r *Row) getContentWidth() int {
	return r.getMaxWidth() - r.getExtraWidth()
}

func (r *Row) getContentHeight() int {
	return r.getMaxHeight() - r.getExtraHeight()
}

func (r *Row) getMaxWidth() int {
	return r.width
}

func (r *Row) getMaxHeight() int {
	return r.height
}

func (r *Row) getExtraWidth() int {
	return r.style.GetHorizontalFrameSize()
}

func (r *Row) getExtraHeight() int {
	return r.style.GetVerticalFrameSize()
}

func (r *Row) copy() Row {
	var cells []*Cell
	for _, cell := range r.cells {
		cellCopy := cell.copy()
		cells = append(cells, &cellCopy)
	}
	rowCopy := *r
	rowCopy.cells = cells
	rowCopy.style = r.style

	return rowCopy
}
