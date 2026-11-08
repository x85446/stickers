// Package flexbox is a responsive grid for terminal apps, inspired by CSS flexbox and
// rendered with lipgloss.
//
// A FlexBox has a width and a height, holds rows, and each row holds cells. A
// HorizontalFlexBox holds columns of cells instead. Give the box the terminal size, give
// rows and cells ratios or fixed sizes, and call Render.
//
//	box := flexbox.New(width, height)
//	box.AddRows([]*flexbox.Row{
//		box.NewRow().AddCells(
//			flexbox.NewCell(1, 1).SetContent("sidebar").SetFixedWidth(20),
//			flexbox.NewCell(3, 1).SetContent("main"),
//		),
//		box.NewRow().AddCells(flexbox.NewCell(1, 1).SetContent("footer")).SetFixedHeight(3),
//	})
//	s := box.Render()
//
// # Sizing
//
// Rows with a fixed height take it; the remaining lines are split between the other rows
// in proportion to each row's largest cell ratioY, leftover lines going one each to the
// rows with the largest ratios. Inside a row, fixed-width cells take their columns and
// the rest are split by ratioX, honoring SetMinWidth. Each cell is ratioY/maxRatioY of
// the row height, rounded down, and sits at the top of the row.
//
// In a HorizontalFlexBox, column widths follow each column's largest cell ratioX, cell
// heights inside a column are split by ratioY honoring SetMinHeight, and each cell is
// ratioX/maxRatioX of the column width. SetFixedWidth, SetFixedHeight and SetMinWidth
// apply to cells of a FlexBox only; SetMinHeight applies to cells of a
// HorizontalFlexBox only.
//
// # Rendering
//
// A cell renders its frame first and its content second. The border and margins render
// complete whenever the cell has room for them; content is wrapped at the content width
// and clipped to the lines between the paddings, so content taller than the cell loses
// its bottom lines, never the border. A cell with zero lines renders as a blank line of
// its width in a row and takes no room in a column; a cell with zero columns adds
// nothing to a row. A row never wraps its cells: anything wider than the row is cut at
// the row's edge, and a row narrower than the box keeps its width so SetRowAlign can
// place it.
//
// # Styles
//
// Styles are lipgloss values. Width and height keys are removed when a style is set,
// because the grid owns sizes. StylePassing on a box passes its style to the rows, and
// on a row or column to the cells; UseStyle sets a style and enables passing in one call.
// Children inherit only the keys they have not set; margins and padding never pass.
//
// # Reading the grid
//
// GetRow, GetCell and GetCellWithID return pointers and mark the layout for
// recalculation on the next render. GetRowCopy, GetRowCellCopy and GetCellCopy return
// copies and leave the layout alone.
package flexbox
