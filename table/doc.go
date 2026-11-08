// Package table is a scrollable, sortable and filterable table for terminal apps, laid
// out with package flexbox and rendered with lipgloss.
//
// A Table renders a one-line header, one-line rows and a one-line footer that shows the
// cursor position and the table size. It scrolls vertically with the cursor and
// horizontally when the columns' minimum widths do not fit the table width.
//
//	t := table.NewTable(width, height, []string{"ID", "Name", "Score"})
//	t, err := t.SetTypes(0, "", 0.0) // one zero value per column
//	t.SetRatio([]int{1, 4, 2}).SetMinWidth([]int{4, 10, 6})
//	t.MustAddRows([][]any{{1, "Ada", 9.5}, {2, "Linus", 8.0}})
//	t.CursorDown()
//	t.OrderByAsc(2)
//	t.SetFilter(1, "li")
//	s := t.Render()
//
// Columns are strings until SetTypes is called. Accepted column types are string, int,
// int8, int16, int32, int64, float32 and float64; rows are validated against them when added,
// and sorting compares values of the column's type. Filtering is a case-insensitive
// substring match on one column.
//
// SetStyles overrides any of the StyleKey entries and keeps the defaults for the rest.
// SetStylePassing passes the row styles down to the cells.
package table
