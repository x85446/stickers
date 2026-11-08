package flexbox

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// Cell is the building block of a grid. In a FlexBox the cells of a row sit side by
// side; in a HorizontalFlexBox the cells of a column stack vertically.
type Cell struct {
	// style of the cell, when rendering it will inherit the style of the parent row
	style lipgloss.Style
	// id of the cell, if not set it will default to the index in the row
	id string

	// ratioX width ratio of the cell
	ratioX int
	// ratioY height ratio of the cell
	ratioY int
	// minWidth minimal width of the cell
	minWidth int
	// minHeight minimal height of the cell
	minHeight int

	// fixedWidth if > 0, this cell will have a fixed width instead of ratio-based
	fixedWidth int
	// fixedHeight if > 0, this cell requests a fixed height (row will use max of all cells)
	fixedHeight int

	width  int
	height int
	// contentGenerator is a function that generates the content of the cell depending on the
	// size of the cell, this can be useful for wrapping text or generating dynamic content.
	contentGenerator func(maxX, maxY int) string
}

// CellOption configures a Cell as NewCell creates it.
type CellOption func(*Cell)

// WithStyle sets the cell's style, with width and height keys removed, as NewCell creates
// it.
func WithStyle(style lipgloss.Style) CellOption {
	return func(c *Cell) {
		c.style = style.
			UnsetWidth().
			UnsetMaxWidth().
			UnsetHeight().
			UnsetMaxHeight()
	}
}

// NewCell initializes a Cell with the given width and height ratios and an empty style.
// Options such as WithStyle configure the cell before it is returned.
func NewCell(ratioX, ratioY int, opts ...CellOption) *Cell {
	c := &Cell{
		style:    lipgloss.NewStyle(),
		ratioX:   ratioX,
		ratioY:   ratioY,
		minWidth: 0,
		width:    0,
		height:   0,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// SetID names the cell for GetCellWithID. Unnamed cells get their index in the row or
// column as their ID.
func (r *Cell) SetID(id string) *Cell {
	r.id = id
	return r
}

// SetContent sets static content. The cell wraps it at its content width and clips it
// to its content height when it renders.
func (r *Cell) SetContent(content string) *Cell {
	r.contentGenerator = func(_, _ int) string {
		return content
	}
	return r
}

// SetContentGenerator sets a function that builds the content each time the cell
// renders. It receives the width and height inside the border and margins; padding, if
// the style has any, lies inside that area.
func (r *Cell) SetContentGenerator(generator func(maxX, maxY int) string) *Cell {
	r.contentGenerator = generator
	return r
}

// GetContent returns the content as set or as the generator builds it for the cell's
// current content size, before wrapping and clipping.
func (r *Cell) GetContent() string {
	if r.contentGenerator == nil {
		return ""
	}
	return r.contentGenerator(r.getContentWidth(), r.getContentHeight())
}

// SetMinWidth keeps a ratio-sized cell at least this wide while it stays responsive.
// It applies to cells of a FlexBox only, not a HorizontalFlexBox.
func (r *Cell) SetMinWidth(value int) *Cell {
	r.minWidth = value
	return r
}

// SetFixedWidth gives the cell a fixed width instead of a ratio-based one. A value of 0
// or less restores ratio sizing. It applies to cells of a FlexBox only, not a
// HorizontalFlexBox.
func (r *Cell) SetFixedWidth(value int) *Cell {
	r.fixedWidth = value
	return r
}

// Deprecated: use [*Cell.SetMinHeight]
func (r *Cell) SetMinHeigth(value int) *Cell {
	return r.SetMinHeight(value)
}

// SetMinHeight keeps a ratio-sized cell at least this tall while it stays responsive.
// It applies to cells of a HorizontalFlexBox only, not a FlexBox.
func (r *Cell) SetMinHeight(value int) *Cell {
	r.minHeight = value
	return r
}

// SetFixedHeight asks for a fixed height for the cell's row. The row takes the largest
// request among its cells, so every cell in it gets that height. A value of 0 or less
// withdraws the request. It applies to cells of a FlexBox only, not a HorizontalFlexBox.
func (r *Cell) SetFixedHeight(value int) *Cell {
	r.fixedHeight = value
	return r
}

// SetStyle replaces the style. Width and height keys are removed because the grid owns
// the cell's size.
func (r *Cell) SetStyle(style lipgloss.Style) *Cell {
	r.style = style.
		UnsetWidth().
		UnsetMaxWidth().
		UnsetHeight().
		UnsetMaxHeight()
	return r
}

// GetStyle returns a copy of the cell's style.
func (r *Cell) GetStyle() lipgloss.Style {
	return r.style
}

// GetWidth returns the width the grid gave the cell, frame included. A content generator
// receives the content area instead.
func (r *Cell) GetWidth() int {
	return r.getMaxWidth()
}

// GetHeight returns the height the grid gave the cell, frame included. A content
// generator receives the content area instead.
func (r *Cell) GetHeight() int {
	return r.getMaxHeight()
}

// render the cell into string.
//
// The frame (margins and border) renders complete whenever the cell has room for it:
// the content is wrapped the way lipgloss wraps it and clipped to the lines left between
// the paddings, so content taller than the cell loses its bottom lines, never the border.
// A cell with no lines renders as one blank line of its width so a row keeps its columns,
// and a cell with no columns renders as empty lines so a column keeps its rows.
func (r *Cell) render(inherited ...lipgloss.Style) string {
	for _, style := range inherited {
		r.style = r.style.Inherit(style)
	}

	if r.getMaxHeight() <= 0 {
		return strings.Repeat(" ", max(r.getMaxWidth(), 0))
	}
	if r.getMaxWidth() <= 0 {
		// no columns: hold the lines so a column keeps its rows and add nothing to a row
		return strings.Repeat("\n", r.getMaxHeight()-1)
	}

	s := r.GetStyle().
		Width(r.getContentWidth()).MaxWidth(r.getMaxWidth()).
		Height(r.getContentHeight())
	block := s.Render(r.fitContent(r.GetContent()))
	return fitHeight(block, r.getMaxHeight(),
		r.style.GetMarginTop()+r.style.GetBorderTopSize(),
		r.style.GetBorderBottomSize()+r.style.GetMarginBottom())
}

// fitContent wraps content at the width lipgloss wraps it and keeps the lines that fit
// between the vertical paddings.
func (r *Cell) fitContent(content string) string {
	wrapAt := r.getContentWidth() - r.style.GetHorizontalPadding()
	lines := r.getContentHeight() - r.style.GetVerticalPadding()
	if wrapAt <= 0 || lines <= 0 {
		return ""
	}
	return clipLines(ansi.Wrap(content, wrapAt, ""), lines)
}

func (r *Cell) getContentWidth() int {
	return r.getMaxWidth() - r.getExtraWidth()
}

func (r *Cell) getContentHeight() int {
	return r.getMaxHeight() - r.getExtraHeight()
}

func (r *Cell) getMaxWidth() int {
	return r.width
}

func (r *Cell) getMaxHeight() int {
	return r.height
}

func (r *Cell) getExtraWidth() int {
	// Only subtract border and margin, not padding.
	// lipgloss Width() sets the content area INCLUDING padding,
	// so we must not double-subtract padding.
	return r.style.GetHorizontalBorderSize() + r.style.GetHorizontalMargins()
}

func (r *Cell) getExtraHeight() int {
	// Only subtract border and margin, not padding.
	// lipgloss Height() sets the content area INCLUDING padding.
	return r.style.GetVerticalBorderSize() + r.style.GetVerticalMargins()
}

func (r *Cell) copy() Cell {
	cellCopy := *r
	cellCopy.style = r.GetStyle()
	return cellCopy
}
