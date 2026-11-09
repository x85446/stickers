package about

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestOverKeepsTheContentAroundTheBox(t *testing.T) {
	content := strings.TrimSuffix(strings.Repeat(strings.Repeat("x", 40)+"\n", 12), "\n")
	got := Over(content, "hello", 40, 12)
	if w, h := lipgloss.Size(got); w != 40 || h != 12 {
		t.Fatalf("the result is %dx%d, want 40x12", w, h)
	}
	lines := strings.Split(ansi.Strip(got), "\n")
	if lines[0] != strings.Repeat("x", 40) || !strings.HasPrefix(lines[6], "xxxxxxxxxxxx") {
		t.Fatalf("the content does not show around the box:\n%s", strings.Join(lines, "\n"))
	}
	if !strings.Contains(lines[5], "hello") && !strings.Contains(lines[6], "hello") {
		t.Fatalf("the box does not show at the center:\n%s", strings.Join(lines, "\n"))
	}
}

// TestOverCutsABoxLargerThanTheArea: the result is the size of the area whatever the
// size of the box.
func TestOverCutsABoxLargerThanTheArea(t *testing.T) {
	for _, size := range [][2]int{{10, 3}, {1, 1}, {30, 4}} {
		content := strings.TrimSuffix(strings.Repeat(strings.Repeat("x", size[0])+"\n", size[1]), "\n")
		got := Over(content, "a long text in a box\nof several lines\nthat does not fit", size[0], size[1])
		if w, h := lipgloss.Size(got); w != size[0] || h != size[1] {
			t.Fatalf("the result is %dx%d, want %dx%d", w, h, size[0], size[1])
		}
	}
}
