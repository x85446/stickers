package demo10

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestDynamicModeKeepsHeaderRow: with nothing fixed, row 1 shares the height like the
// other rows and still shows its three lines of text inside a closed border.
func TestDynamicModeKeepsHeaderRow(t *testing.T) {
	m := New()
	m.Update(tea.WindowSizeMsg{Width: 127, Height: 22})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})

	lines := strings.Split(m.View(), "\n")
	if len(lines) != 22 {
		t.Fatalf("rendered %d lines, want 22", len(lines))
	}
	row1 := strings.Join(lines[:5], "\n")
	for _, want := range []string{"Sidebar", "Header Area", "Info", "H dynamic, W dynamic", "(25x5)", "└"} {
		if !strings.Contains(row1, want) {
			t.Errorf("row 1 lacks %q:\n%s", want, row1)
		}
	}
}
