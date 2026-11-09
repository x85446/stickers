package flexbox

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// TestStyleFeaturesKeepSizeAndText: the style features of lipgloss v2 show when set on a
// cell, a row or a box, and leave sizes and text as they are.
func TestStyleFeaturesKeepSizeAndText(t *testing.T) {
	base := bordered().Padding(0, 1)
	render := func(style lipgloss.Style) string {
		box := New(24, 5)
		box.AddRows([]*Row{box.NewRow().AddCells(
			NewCell(1, 1).SetStyle(style).SetContent("left cell with some text"),
			NewCell(1, 1).SetStyle(base).SetContent("right"),
		)})
		return box.Render()
	}
	want := ansi.Strip(render(base))

	for _, tc := range []struct {
		name   string
		style  lipgloss.Style
		marker string // an escape sequence the feature writes
	}{
		{"gradient border", base.BorderForegroundBlend(lipgloss.Color("#ff0000"), lipgloss.Color("#0000ff")), "\x1b[38;2;"},
		{"curly underline in a color", base.UnderlineStyle(lipgloss.UnderlineCurly).UnderlineColor(lipgloss.Color("#00ff00")), "4:3"},
		{"hyperlink", base.Hyperlink("https://example.com"), "\x1b]8;;https://example.com"},
		{"background and bold", base.Background(lipgloss.Color("#222222")).Bold(true), "\x1b[1;"},
	} {
		rendered := render(tc.style)
		assertSize(t, rendered, 24, 5)
		if got := ansi.Strip(rendered); got != want {
			t.Errorf("%s changes the text:\n%s\nwant\n%s", tc.name, got, want)
		}
		if !strings.Contains(rendered, tc.marker) {
			t.Errorf("%s does not show in the render: no %q in %q", tc.name, tc.marker, rendered)
		}
	}

	// a gradient has more than one color along the border
	gradient := render(base.BorderForegroundBlend(lipgloss.Color("#ff0000"), lipgloss.Color("#0000ff")))
	if !strings.Contains(gradient, "\x1b[38;2;255;0;0m") || strings.Count(gradient, "\x1b[38;2;") < 4 {
		t.Errorf("the gradient border has too few colors: %q", gradient)
	}

	// padding and margin characters fill the frame and keep the size
	dotted := render(lipgloss.NewStyle().Padding(1).PaddingChar('.').Margin(0, 1).MarginChar('~'))
	assertSize(t, dotted, 24, 5)
	if second := strings.Split(ansi.Strip(dotted), "\n")[1]; !strings.HasPrefix(second, "~.left.") {
		t.Errorf("padding and margin characters do not show: %q", second)
	}
}

// TestUnderlinedContainersKeepTheirCellsWhole: a row, column or box that underlines or
// strikes through leaves the escape sequences of its styled cells intact, and cells that
// inherit the style show it themselves.
func TestUnderlinedContainersKeepTheirCellsWhole(t *testing.T) {
	lined := lipgloss.NewStyle().Underline(true).Strikethrough(true)
	colored := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	whole := func(name, rendered string) {
		t.Helper()
		assertSize(t, rendered, 12, 2)
		if got := ansi.Strip(rendered); got != "ab    cd    \n            " {
			t.Errorf("%s shows %q", name, got)
		}
	}
	cells := func() []*Cell {
		return []*Cell{NewCell(1, 1).SetStyle(colored).SetContent("ab"), NewCell(1, 1).SetContent("cd")}
	}

	box := New(12, 2)
	row := box.NewRow().SetStyle(lined).AddCells(cells()...)
	box.AddRows([]*Row{row})
	whole("an underlined row", box.Render())
	row.StylePassing(true)
	whole("an underlined row that passes its style", box.Render())
	if rendered := box.Render(); !strings.Contains(rendered, "\x1b[4;") && !strings.Contains(rendered, ";4;") {
		t.Errorf("cells that inherit an underline do not show it: %q", rendered)
	}
	box.SetStyle(lined)
	whole("an underlined box", box.Render())

	horizontal := NewHorizontal(12, 2).SetStyle(lined)
	horizontal.AddColumns([]*Column{
		horizontal.NewColumn().SetStyle(lined).AddCells(cells()[0]),
		horizontal.NewColumn().AddCells(cells()[1]),
	})
	whole("an underlined column in an underlined box", horizontal.Render())

	// plain cells in an underlined row are underlined by the row
	plainBox := New(4, 1)
	plainBox.AddRows([]*Row{plainBox.NewRow().SetStyle(lipgloss.NewStyle().Underline(true)).AddCells(NewCell(1, 1).SetContent("ab"))})
	if rendered := plainBox.Render(); !strings.Contains(rendered, "\x1b[4") {
		t.Errorf("an underlined row does not underline its plain cells: %q", rendered)
	}
}

// TestHyperlinksStayInTheirElement: a hyperlink set on a cell or a column covers that
// element on every line it has text on, and nothing beside it.
func TestHyperlinksStayInTheirElement(t *testing.T) {
	link := lipgloss.NewStyle().Hyperlink("https://example.com")
	open, closed := "\x1b]8;;https://example.com\a", ansi.ResetHyperlink()
	// linked returns how many columns of line the link covers, from the left edge
	linked := func(line string) int {
		t.Helper()
		before, _, found := strings.Cut(line, closed)
		if strings.Count(line, open) != 1 || strings.Count(line, closed) != 1 || !strings.HasPrefix(line, open) || !found {
			t.Fatalf("the line does not open one link at its start and close it: %q", line)
		}
		return ansi.StringWidth(before)
	}

	box := New(12, 3)
	box.AddRows([]*Row{box.NewRow().AddCells(
		NewCell(1, 1).SetStyle(link).SetContent("link\nmore"),
		NewCell(1, 1).SetContent("plain"),
	)})
	got := lines(box)
	assertLines(t, strings.Split(ansi.Strip(box.Render()), "\n"), "link  plain ", "more        ", "            ")
	for i, line := range got[:2] {
		if columns := linked(line); columns < 4 || columns > 6 {
			t.Errorf("line %d links %d columns, want the text of the first cell and no more than its 6 columns: %q", i, columns, line)
		}
	}
	if strings.Contains(got[2], linkStart) {
		t.Errorf("a line without text carries a link: %q", got[2])
	}

	horizontal := NewHorizontal(12, 2)
	horizontal.AddColumns([]*Column{
		horizontal.NewColumn().SetStyle(link).AddCells(NewCell(1, 1).SetContent("link\nmore")),
		horizontal.NewColumn().AddCells(NewCell(1, 1).SetContent("plain")),
	})
	for i, line := range lines(horizontal) {
		if columns := linked(line); columns != 6 {
			t.Errorf("line %d links %d columns, want the 6 columns of the first column: %q", i, columns, line)
		}
	}
}

func TestLinkAfter(t *testing.T) {
	a, b := "\x1b]8;;https://a\a", "\x1b]8;id=1;https://b\x1b\\"
	for _, tc := range []struct{ text, open, want string }{
		{"plain", "", ""},
		{"plain", a, a},
		{a + "text", "", a},
		{a + "text" + ansi.ResetHyperlink(), "", ""},
		{"text\x1b]8;;\x1b\\", a, ""},
		{a + "x" + ansi.ResetHyperlink() + b + "y", "", b},
		{"cut \x1b]8;;https://", a, a},
	} {
		if got := linkAfter(tc.text, tc.open); got != tc.want {
			t.Errorf("linkAfter(%q, %q) = %q, want %q", tc.text, tc.open, got, tc.want)
		}
	}
}

// TestLinkedContentWrapsInsideItsCell: content that carries its own hyperlink and colors
// wraps like plain text, and every line closes what it opens, so neither a link nor a
// color runs into the next cell.
func TestLinkedContentWrapsInsideItsCell(t *testing.T) {
	linked := lipgloss.NewStyle().Hyperlink("https://example.com").Foreground(lipgloss.Color("#ff0000")).
		Render("a linked text that wraps over lines")
	for _, style := range []*lipgloss.Style{nil, ptr(bordered())} {
		cell := NewCell(1, 1).SetContent(linked)
		width := 12
		if style != nil {
			cell.SetStyle(*style)
			width = 14
		}
		box := New(width+6, 6)
		box.AddRows([]*Row{box.NewRow().AddCells(cell.SetFixedWidth(width), NewCell(1, 1).SetContent("plain"))})
		rendered := box.Render()
		assertSize(t, rendered, width+6, 6)

		text := ansi.Strip(rendered)
		for _, word := range []string{"a linked", "text that", "wraps over", "lines", "plain"} {
			if !strings.Contains(text, word) {
				t.Errorf("the render lacks %q:\n%s", word, text)
			}
		}
		for i, line := range strings.Split(rendered, "\n") {
			opened := strings.Count(line, "\x1b]8;;https://example.com")
			closed := strings.Count(line, "\x1b]8;;\x1b\\") + strings.Count(line, "\x1b]8;;\a")
			if opened != closed {
				t.Errorf("line %d opens %d links and closes %d: %q", i, opened, closed, line)
			}
			if at := strings.Index(line, "plain"); at >= 0 && strings.LastIndex(line[:at], "\x1b[38;2;255;0;0m") > strings.LastIndex(line[:at], "\x1b[m") {
				t.Errorf("line %d colors the next cell: %q", i, line)
			}
		}
	}
}

func ptr[T any](v T) *T { return &v }
