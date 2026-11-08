package table

import (
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]`)

func plain(s string) string { return ansiRe.ReplaceAllString(s, "") }

func TestInt64ColumnsSortNumerically(t *testing.T) {
	tb := NewTable(40, 6, []string{"ID", "Name"})
	if _, err := tb.SetTypes(int64(0), ""); err != nil {
		t.Fatalf("SetTypes int64: %v", err)
	}
	if _, err := tb.AddRows([][]any{{int64(20), "b"}, {int64(3), "a"}, {int64(100), "c"}}); err != nil {
		t.Fatalf("AddRows int64: %v", err)
	}
	firstColumn := func() string {
		lines := strings.Split(plain(tb.Render()), "\n")
		var ids []string
		for _, l := range lines[1:4] {
			ids = append(ids, strings.Fields(l)[0])
		}
		return strings.Join(ids, " ")
	}
	tb.OrderByAsc(0)
	if got := firstColumn(); got != "3 20 100" {
		t.Fatalf("OrderByAsc order %q, want 3 20 100", got)
	}
	tb.OrderByDesc(0)
	if got := firstColumn(); got != "100 20 3" {
		t.Fatalf("OrderByDesc order %q, want 100 20 3", got)
	}
	tb.OrderByAsc(1)
	if got := firstColumn(); got != "3 20 100" {
		t.Fatalf("OrderByAsc by name gives %q, want 3 20 100", got)
	}
	if col, dir := tb.GetOrder(); col != 1 || dir != SortingOrderAscending {
		t.Fatalf("GetOrder = %d, %d", col, dir)
	}
}

func TestSetStylesKeepsDefaultsForOtherTables(t *testing.T) {
	custom := lipgloss.NewStyle().Background(lipgloss.Color("#123456"))
	a := NewTable(10, 5, []string{"A"})
	a.SetStyles(map[StyleKey]lipgloss.Style{StyleKeyHeader: custom})
	b := NewTable(10, 5, []string{"A"})
	if b.styles[StyleKeyHeader].GetBackground() == custom.GetBackground() {
		t.Fatal("SetStyles on one table changed the header style of another")
	}
	if tableDefaultStyles[StyleKeyHeader].GetBackground() == custom.GetBackground() {
		t.Fatal("SetStyles wrote into the package defaults")
	}
	if a.styles[StyleKeyHeader].GetBackground() != custom.GetBackground() {
		t.Fatal("SetStyles did not apply the override")
	}
}

func TestRenderHeightBeforeSetHeight(t *testing.T) {
	tb := NewTable(20, 6, []string{"A"})
	tb.MustAddRows([][]any{{"x"}})
	if n := strings.Count(tb.Render(), "\n") + 1; n != 6 {
		t.Fatalf("rendered %d lines, want 6", n)
	}
}

// TestScrolledRowsUseTheirColumnsRatios: after scrolling right, the rows must size their
// cells with the ratios and minimum widths of the columns on screen, as the header does.
func TestScrolledRowsUseTheirColumnsRatios(t *testing.T) {
	tb := NewTable(30, 5, []string{"A", "B", "C", "D"})
	tb.SetRatio([]int{1, 1, 2, 1}).SetMinWidth([]int{12, 12, 12, 12})
	tb.SetWidth(30)
	tb.MustAddRows([][]any{{"a1", "b1", "c1", "d1"}})
	tb.CursorRight()
	tb.CursorRight()
	if l, r := tb.GetVisibleColumnRange(); l != 1 || r != 2 {
		t.Fatalf("visible range %d-%d, want 1-2", l, r)
	}
	lines := strings.Split(plain(tb.Render()), "\n")
	header, row := lines[0], lines[1]
	if h, c := strings.Index(header, "C"), strings.Index(row, "c1"); h != c {
		t.Fatalf("column C starts at %d in the header and its value at %d in the row:\n%s\n%s", h, c, header, row)
	}
}

func TestSetRatioAndSetMinWidthPanicOnBadInput(t *testing.T) {
	mustPanic := func(name string, f func()) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Errorf("%s did not panic", name)
			}
		}()
		f()
	}
	mustPanic("SetRatio with one value for two columns", func() { NewTable(10, 5, []string{"A", "B"}).SetRatio([]int{1}) })
	mustPanic("SetRatio with a zero ratio", func() { NewTable(10, 5, []string{"A", "B"}).SetRatio([]int{1, 0}) })
	mustPanic("SetMinWidth with one value for two columns", func() { NewTable(10, 5, []string{"A", "B"}).SetMinWidth([]int{1}) })
}
