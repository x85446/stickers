package flexbox

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// TestHorizontalColumnAlign: columns of a locked width are narrower than the box and sit
// where SetColumnAlign puts them.
func TestHorizontalColumnAlign(t *testing.T) {
	for _, tc := range []struct {
		align lipgloss.Position
		want  string
	}{
		{lipgloss.Left, "a b       "},
		{lipgloss.Center, "   a b    "},
		{lipgloss.Right, "      a b "},
	} {
		box := NewHorizontal(10, 1).LockColumnWidth(2).SetColumnAlign(tc.align)
		box.AddColumns([]*Column{
			box.NewColumn().AddCells(NewCell(1, 1).SetContent("a")),
			box.NewColumn().AddCells(NewCell(1, 1).SetContent("b")),
		})
		assertLines(t, strings.Split(box.Render(), "\n"), tc.want)
	}
}
