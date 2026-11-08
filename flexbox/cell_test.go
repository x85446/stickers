package flexbox

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func bordered() lipgloss.Style { return lipgloss.NewStyle().Border(lipgloss.NormalBorder()) }

// renderAt renders a cell at the given size and returns its lines.
func renderAt(c *Cell, width, height int) []string {
	c.width, c.height = width, height
	return strings.Split(c.render(), "\n")
}

func assertLines(t *testing.T, got []string, want ...string) {
	t.Helper()
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("rendered\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestCellRenderFitsContent(t *testing.T) {
	cell := NewCell(1, 1).SetStyle(bordered()).SetContent("ab\ncd")
	assertLines(t, renderAt(cell, 6, 4),
		"┌────┐",
		"│ab  │",
		"│cd  │",
		"└────┘")
}

func TestCellRenderClipsContentNotBorder(t *testing.T) {
	cell := NewCell(1, 1).SetStyle(bordered()).SetContent("ab\ncd")
	assertLines(t, renderAt(cell, 6, 3),
		"┌────┐",
		"│ab  │",
		"└────┘")
}

func TestCellRenderFrameOnly(t *testing.T) {
	cell := NewCell(1, 1).SetStyle(bordered()).SetContent("ab\ncd")
	assertLines(t, renderAt(cell, 6, 2),
		"┌────┐",
		"└────┘")
}

func TestCellRenderBelowFrameKeepsTop(t *testing.T) {
	cell := NewCell(1, 1).SetStyle(bordered()).SetContent("ab\ncd")
	assertLines(t, renderAt(cell, 6, 1), "┌────┐")
}

func TestCellRenderZeroHeightHoldsWidth(t *testing.T) {
	cell := NewCell(1, 1).SetStyle(bordered()).SetContent("ab")
	assertLines(t, renderAt(cell, 5, 0), "     ")
}

func TestCellRenderClipsWrappedContent(t *testing.T) {
	cell := NewCell(1, 1).SetStyle(bordered()).SetContent("aaaa bbbb cccc")
	assertLines(t, renderAt(cell, 6, 4),
		"┌────┐",
		"│aaaa│",
		"│bbbb│",
		"└────┘")
}

func TestCellRenderPaddingLargerThanContent(t *testing.T) {
	cell := NewCell(1, 1).SetStyle(bordered().Padding(1, 0)).SetContent("ab")
	assertLines(t, renderAt(cell, 6, 4),
		"┌────┐",
		"│    │",
		"│    │",
		"└────┘")
}

func TestCellRenderKeepsMargins(t *testing.T) {
	cell := NewCell(1, 1).SetStyle(bordered().Margin(1, 0)).SetContent("ab\ncd")
	assertLines(t, renderAt(cell, 6, 4),
		"      ",
		"┌────┐",
		"└────┘",
		"      ")
}

// TestRowShortCellsKeepBottomBorder is demo-7's bottom row at a 4-line height: the cells
// with the smaller height ratios get 3 and 2 lines and still close their borders.
func TestRowShortCellsKeepBottomBorder(t *testing.T) {
	box := New(70, 4)
	row := box.NewRow()
	for _, c := range []struct {
		label  string
		ratioY int
	}{{"P", 4}, {"Q", 3}, {"R", 2}, {"S", 3}, {"T", 4}} {
		row.AddCells(NewCell(1, c.ratioY).SetStyle(bordered()).
			SetContent(c.label + "\n[1:" + string(rune('0'+c.ratioY)) + "]"))
	}
	box.AddRows([]*Row{row})

	got := strings.Split(box.Render(), "\n")
	assertLines(t, got,
		"┌────────────┐┌────────────┐┌────────────┐┌────────────┐┌────────────┐",
		"│P           ││Q           │└────────────┘│S           ││T           │",
		"│[1:4]       │└────────────┘              └────────────┘│[1:4]       │",
		"└────────────┘                                          └────────────┘")
}

// TestColumnSkipsZeroHeightCells: with 2 lines for ratios 1:1:2 the first two cells get
// no lines and the third gets both, so its content must be visible.
func TestColumnSkipsZeroHeightCells(t *testing.T) {
	box := NewHorizontal(3, 2)
	col := box.NewColumn()
	col.AddCells(
		NewCell(1, 1).SetContent("a"),
		NewCell(1, 1).SetContent("b"),
		NewCell(1, 2).SetContent("c"),
	)
	box.AddColumns([]*Column{col})

	assertLines(t, strings.Split(box.Render(), "\n"),
		"c  ",
		"   ")
}
