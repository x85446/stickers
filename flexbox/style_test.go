package flexbox

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// TestUnderlineRunsThroughStyledContent: a cell, or a row, that underlines or strikes
// through keeps the escape sequences of styled content whole and draws its lines under
// all of it, also after the content resets its own color.
func TestUnderlineRunsThroughStyledContent(t *testing.T) {
	useTrueColor(t)
	under := lipgloss.NewStyle().Underline(true)
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Render("red")

	for name, style := range map[string]lipgloss.Style{
		"underline":        under,
		"strikethrough":    lipgloss.NewStyle().Strikethrough(true),
		"both":             lipgloss.NewStyle().Underline(true).Strikethrough(true),
		"underline spaces": lipgloss.NewStyle().UnderlineSpaces(true),
	} {
		for _, content := range []string{red + " and plain", "plain " + red + "\nsecond " + red} {
			// the style on the cell itself, and passed down by the row
			own := New(18, 2)
			own.AddRows([]*Row{own.NewRow().AddCells(NewCell(1, 1).SetStyle(style).SetContent(content))})
			passed := New(18, 2)
			passed.AddRows([]*Row{passed.NewRow().UseStyle(style).AddCells(NewCell(1, 1).SetContent(content))})
			want := New(18, 2)
			want.AddRows([]*Row{want.NewRow().AddCells(NewCell(1, 1).SetContent(content))})
			for kind, box := range map[string]*FlexBox{"on the cell": own, "passed by the row": passed} {
				rendered := box.Render()
				assertSize(t, rendered, 18, 2)
				if got := ansi.Strip(rendered); got != ansi.Strip(want.Render()) {
					t.Errorf("%s %s shows %q for %q", name, kind, got, content)
				}
			}
		}
	}

	// the line is drawn under the colored word and again after its reset
	box := New(14, 1)
	box.AddRows([]*Row{box.NewRow().AddCells(NewCell(1, 1).SetStyle(under).SetContent(red + " after"))})
	rendered := box.Render()
	on := "\x1b[4m"
	if !strings.HasPrefix(rendered, on) || !strings.Contains(rendered, "red\x1b[0m"+on+" after") {
		t.Errorf("the underline does not run through the styled content: %q", rendered)
	}

	// a row keeps the underline on its plain cells when another cell is colored
	row := New(12, 1)
	row.AddRows([]*Row{row.NewRow().SetStyle(under).AddCells(NewCell(1, 1).SetContent("ab"), NewCell(1, 1).SetContent(red))})
	if rendered := row.Render(); !strings.HasPrefix(rendered, on+"ab") || ansi.Strip(rendered) != "ab    red   " {
		t.Errorf("an underlined row with a colored cell renders %q", rendered)
	}
}

// TestNoLinesWithoutAColorProfile: where lipgloss writes no escape sequences, an
// underlined element writes none either.
func TestNoLinesWithoutAColorProfile(t *testing.T) {
	box := New(10, 1)
	box.AddRows([]*Row{box.NewRow().AddCells(
		NewCell(1, 1).SetStyle(lipgloss.NewStyle().Underline(true)).SetContent("\x1b[31mred\x1b[0m"),
	)})
	if got := box.Render(); got != "\x1b[31mred\x1b[0m       " {
		t.Fatalf("rendered %q", got)
	}
}

// TestLinesResumeAfterAnySequenceThatEndsThem: the underline of an element comes back
// after content turns it off in any way.
func TestLinesResumeAfterAnySequenceThatEndsThem(t *testing.T) {
	on, off := "\x1b[4m", "\x1b[24m"
	for _, end := range []string{"\x1b[m", "\x1b[0m", "\x1b[00m", "\x1b[24m", "\x1b[29m", "\x1b[59m", "\x1b[4:0m", "\x1b[01;00m"} {
		if got := relined("a"+end+"b", on, off); got != on+"a"+end+on+"b"+off {
			t.Errorf("after %q the line does not resume: %q", end, got)
		}
	}
	for _, keep := range []string{"\x1b[31m", "\x1b[39m", "\x1b[38;5;0m", "\x1b[38;2;0;0;0m", "\x1b[4:3m", "\x1b[1;34m"} {
		if got := relined("a"+keep+"b", on, off); got != on+"a"+keep+"b"+off {
			t.Errorf("after %q the line is written again: %q", keep, got)
		}
	}
}

func TestRelined(t *testing.T) {
	on, off := "\x1b[4m", "\x1b[24m"
	for _, tc := range []struct{ text, want string }{
		{"plain", on + "plain" + off},
		{"a\n\nb", on + "a" + off + "\n\n" + on + "b" + off},
		{"\x1b[31mred\x1b[m after", on + "\x1b[31mred\x1b[m" + on + " after" + off},
		{"\x1b[0;1mbold", on + "\x1b[0;1m" + on + "bold" + off},
		{"\x1b[38;2;0;0;0mblack\x1b[38;5;0m", on + "\x1b[38;2;0;0;0mblack\x1b[38;5;0m" + off},
		{"\x1b[2Kcleared", on + "\x1b[2Kcleared" + off},
		{"cut \x1b[31", on + "cut \x1b[31" + off},
	} {
		if got := relined(tc.text, on, off); got != tc.want {
			t.Errorf("relined(%q) = %q, want %q", tc.text, got, tc.want)
		}
	}
	if got := relined("text", "", ""); got != "text" {
		t.Errorf("relined without lines changed the text to %q", got)
	}
}
