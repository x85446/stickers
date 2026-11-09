package flexbox

import "github.com/charmbracelet/lipgloss"

// FlexBox is a responsive grid of rows, each holding cells side by side. Rows stack
// vertically and share the box height by ratio or fixed size.
//
// A FlexBox is not safe for concurrent use.
type FlexBox struct {
	style styling
	rows  []*Row

	width, height int
	// fixedRowHeight pins every row to a height when greater than 0
	fixedRowHeight int
	// rowAlign places rows narrower than the widest row
	rowAlign lipgloss.Position

	// out is the last render, key and revs what it was rendered from
	out      string
	key      containerKey
	revs     []uint64
	rendered bool

	// tracks, parts and frame are reused between renders
	tracks []track
	parts  []*block
	frame  block
}

// New initializes a FlexBox of the given size with no rows and an empty style.
func New(width, height int) *FlexBox {
	return &FlexBox{
		style:    styling{own: lipgloss.NewStyle()},
		width:    width,
		height:   height,
		rowAlign: lipgloss.Left,
	}
}

// SetStyle replaces the style. Width and height keys are removed because the box's size
// is set with SetWidth and SetHeight. A border, padding or margin shrinks the area the
// rows share.
func (b *FlexBox) SetStyle(style lipgloss.Style) *FlexBox {
	b.style.set(style)
	return b
}

// StylePassing sets whether the box's style is passed to its rows. A row passes it on to
// its cells only when its own StylePassing is enabled.
func (b *FlexBox) StylePassing(value bool) *FlexBox {
	b.style.passing = value
	return b
}

// UseStyle sets the style and enables StylePassing in one call.
func (b *FlexBox) UseStyle(style lipgloss.Style) *FlexBox {
	return b.SetStyle(style).StylePassing(true)
}

// SetRowAlign places rows narrower than the widest row, such as rows of fixed-width
// cells next to full-width rows, at lipgloss.Left, lipgloss.Center or lipgloss.Right.
// The default is lipgloss.Left.
func (b *FlexBox) SetRowAlign(align lipgloss.Position) *FlexBox {
	b.rowAlign = align
	return b
}

// NewRow initializes an empty Row. Add it with AddRows or SetRows.
func (b *FlexBox) NewRow() *Row {
	return &Row{style: styling{own: lipgloss.NewStyle()}}
}

// AddRows appends rows to the box. A nil row is left out.
func (b *FlexBox) AddRows(rows []*Row) *FlexBox {
	b.rows = appendElements(b.rows, rows)
	return b
}

// SetRows replaces all rows of the box. A nil row is left out. The box keeps its own
// list, so later changes to the slice do not reach it.
func (b *FlexBox) SetRows(rows []*Row) *FlexBox {
	b.rows = replaceElements(b.rows, rows)
	return b
}

// RowsLen returns the number of rows.
func (b *FlexBox) RowsLen() int {
	return len(b.rows)
}

// GetRow returns the row at the index, or nil.
func (b *FlexBox) GetRow(index int) *Row {
	if index >= 0 && index < len(b.rows) {
		return b.rows[index]
	}
	return nil
}

// GetRowCopy returns a copy of the row at the index, with copies of its cells, or nil.
// Changes to the copy do not reach the box.
func (b *FlexBox) GetRowCopy(index int) *Row {
	if index >= 0 && index < len(b.rows) {
		rowCopy := b.rows[index].copy()
		return &rowCopy
	}
	return nil
}

// GetRowCellCopy returns a copy of the cell at cellIndex in the row at rowIndex, or nil.
// Changes to the copy do not reach the box.
func (b *FlexBox) GetRowCellCopy(rowIndex, cellIndex int) *Cell {
	if row := b.GetRow(rowIndex); row != nil {
		return row.GetCellCopy(cellIndex)
	}
	return nil
}

// UpdateRow replaces the row at the index. An index out of range or a nil row changes
// nothing.
func (b *FlexBox) UpdateRow(index int, row *Row) *FlexBox {
	if row != nil && index >= 0 && index < len(b.rows) {
		b.rows[index] = row
	}
	return b
}

// LockRowHeight pins every row to the given number of lines, which disables vertical
// scaling. A value of 0 or less restores ratio sizing.
func (b *FlexBox) LockRowHeight(value int) *FlexBox {
	b.fixedRowHeight = value
	return b
}

// SetHeight sets the box height, the lines its rows share.
func (b *FlexBox) SetHeight(value int) *FlexBox {
	b.height = value
	return b
}

// SetWidth sets the box width, the columns its rows share.
func (b *FlexBox) SetWidth(value int) *FlexBox {
	b.width = value
	return b
}

// GetHeight returns the box height.
func (b *FlexBox) GetHeight() int {
	return b.height
}

// GetWidth returns the box width.
func (b *FlexBox) GetWidth() int {
	return b.width
}

// Render lays out the grid and returns it as exactly GetHeight lines of GetWidth columns.
// Only what changed since the last render is rendered again: a cell when its size, style
// or content changed, a row when one of its cells did. An unchanged grid costs a call to
// each content generator.
//
// Text styled by an earlier render is reused as it is. After changing how lipgloss
// renders a style at run time, such as the color profile, set the styles again.
func (b *FlexBox) Render() string {
	b.layout()

	key := containerKey{width: max(b.width, 0), height: max(b.height, 0), style: b.style.id, align: b.rowAlign}
	stale := !b.rendered || key != b.key || len(b.revs) != len(b.rows)
	b.revs, b.parts = resize(b.revs, len(b.rows)), b.parts[:0]
	passed := b.style.passed()
	for i, row := range b.rows {
		part := row.render(passed)
		if part.rev != b.revs[i] {
			b.revs[i], stale = part.rev, true
		}
		b.parts = append(b.parts, part)
	}
	if !stale {
		return b.out
	}
	b.key, b.rendered = key, true

	// Rows narrower than the widest row are aligned against it; the box then pads the
	// group to its own width.
	width, height := b.contentSize()
	group := widest(b.parts)
	if b.style.plain() {
		b.frame.stack(b.parts, group, max(group, width), b.rowAlign, height)
		b.frame.clip(width)
		b.frame.extend(height)
	} else {
		b.frame.stack(b.parts, group, group, b.rowAlign, height)
		f := b.style.frame
		style, content := b.style.dress(b.frame.String())
		style = style.
			Width(key.width - f.edgeX).MaxWidth(key.width).
			Height(key.height - f.edgeY())
		text := renderStyled(style, content)
		b.frame.fill(text, key.width, key.height, f.top, f.bottom)
	}
	b.out = b.frame.String()
	return b.out
}

// ForceRecalculate lays out the box and every row now, so the sizes of rows and cells
// can be read before the next render.
func (b *FlexBox) ForceRecalculate() {
	b.layout()
	passed := b.style.passed()
	for _, row := range b.rows {
		row.layout(passed)
	}
}

// contentSize returns the area the rows share, inside the box's frame.
func (b *FlexBox) contentSize() (width, height int) {
	f := b.style.frame
	return max(b.width-f.edgeX-f.paddingX, 0), max(b.height-f.edgeY()-f.paddingY, 0)
}

// layout sizes the rows. Every row takes the content width, and the rows share the
// content height as tracks.
func (b *FlexBox) layout() {
	b.style.resolve(inherited{})
	width, height := b.contentSize()

	tracks := b.tracks[:0]
	for _, row := range b.rows {
		t := row.track()
		if b.fixedRowHeight > 0 {
			t.fixed = b.fixedRowHeight
		}
		tracks = append(tracks, t)
	}
	distribute(tracks, height)
	for i, row := range b.rows {
		row.width, row.height = width, tracks[i].size
	}
	b.tracks = tracks
}

// appendElements appends the rows, columns or cells that are not nil to dst.
func appendElements[T any](dst, elements []*T) []*T {
	for _, element := range elements {
		if element != nil {
			dst = append(dst, element)
		}
	}
	return dst
}

// replaceElements returns the rows, columns or cells that are not nil in the storage of
// dst, which no longer holds what it held before.
func replaceElements[T any](dst, elements []*T) []*T {
	replaced := appendElements(dst[:0], elements)
	if len(replaced) < len(dst) {
		clear(dst[len(replaced):])
	}
	return replaced
}
