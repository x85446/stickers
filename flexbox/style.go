package flexbox

import (
	"strings"
	"sync/atomic"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// keeper marks the end of a text whose trailing spaces a style must keep: a zero-width
// space, which takes no columns and is not a space to a word wrap.
const keeper = "\u200b"

// lastID is the number nextID returned last.
var lastID atomic.Uint64

// nextID returns a number no earlier call returned. Styles and rendered blocks carry one
// as their identity, so comparing two numbers tells whether a style or a block changed.
func nextID() uint64 {
	return lastID.Add(1)
}

// sizeless returns the style without its width and height keys, because the grid owns
// sizes.
func sizeless(style lipgloss.Style) lipgloss.Style {
	return style.
		UnsetWidth().
		UnsetMaxWidth().
		UnsetHeight().
		UnsetMaxHeight()
}

// renderStyled renders text with style. Unlike style.Render it keeps the spaces at the
// end of text, which the word wrap inside lipgloss drops; without them the last line of
// a centered or right-aligned block shifts against the lines above it.
func renderStyled(style lipgloss.Style, text string) string {
	if !strings.HasSuffix(text, " ") {
		return style.Render(text)
	}
	rendered := style.Render(text + keeper)
	if i := strings.LastIndex(rendered, keeper); i >= 0 {
		return rendered[:i] + rendered[i+len(keeper):]
	}
	return rendered
}

// inherited is a style a parent passes to a child.
type inherited struct {
	style *lipgloss.Style
	id    uint64 // identity of style; 0 when the parent passes nothing
}

// frame holds the sizes of a style's margins, border and padding.
type frame struct {
	top, bottom int // margin plus border above and below the content
	edgeX       int // margin plus border, left and right together
	paddingX    int // left and right padding together
	paddingY    int // top and bottom padding together
}

// edgeY returns the lines the margins and border take, above and below together.
func (f frame) edgeY() int { return f.top + f.bottom }

// styling is the style of a box, row, column or cell: the style set on it and the
// effective style it renders with, which adds what its parent passes.
type styling struct {
	own     lipgloss.Style
	ownID   uint64 // identity of own; 0 while own is the empty style
	passing bool   // pass the effective style on to the children

	effective lipgloss.Style
	id        uint64 // identity of effective; 0 while it is the empty style
	frame     frame  // sizes of effective
	// runeByRune is true when effective underlines or strikes through, which
	// lipgloss writes one rune at a time. linesOn and linesOff are the escape
	// sequences that start and end those lines.
	runeByRune bool
	linesOn    string
	linesOff   string

	// what effective was resolved from
	resolved     bool
	resolvedOwn  uint64
	resolvedFrom uint64
}

// set replaces the element's own style.
func (s *styling) set(style lipgloss.Style) {
	s.own = sizeless(style)
	s.ownID = nextID()
}

// resolve brings the effective style up to date with the own style and the style the
// parent passes. It does nothing when neither changed.
func (s *styling) resolve(from inherited) {
	if s.resolved && s.resolvedOwn == s.ownID && s.resolvedFrom == from.id {
		return
	}
	s.effective, s.id = s.own, s.ownID
	if from.id != 0 {
		s.effective = s.own.Inherit(*from.style)
		s.id = nextID()
	}
	s.frame, s.runeByRune, s.linesOn, s.linesOff = frame{}, false, "", ""
	if s.id != 0 {
		s.runeByRune = s.effective.GetUnderline() || s.effective.GetStrikethrough() ||
			s.effective.GetUnderlineSpaces() || s.effective.GetStrikethroughSpaces()
		s.linesOn, s.linesOff = lineSequences(s.effective)
		s.frame = frame{
			top:      s.effective.GetMarginTop() + s.effective.GetBorderTopSize(),
			bottom:   s.effective.GetBorderBottomSize() + s.effective.GetMarginBottom(),
			edgeX:    s.effective.GetHorizontalMargins() + s.effective.GetHorizontalBorderSize(),
			paddingX: s.effective.GetHorizontalPadding(),
			paddingY: s.effective.GetVerticalPadding(),
		}
	}
	s.resolved, s.resolvedOwn, s.resolvedFrom = true, s.ownID, from.id
}

// dress returns the style and the content an element renders with.
//
// Lipgloss writes underline and strikethrough one rune at a time, which cuts the escape
// sequences of content that is styled itself apart: a colored value, the cells of a row.
// For such content the style is returned without the two, and the content with the
// sequences that draw them written around every line, and again after every reset inside
// it. Content without escape sequences is left to lipgloss.
func (s *styling) dress(content string) (lipgloss.Style, string) {
	// without a color profile lipgloss writes no escape sequences, so nothing is cut
	if !s.runeByRune || !strings.Contains(content, "\x1b") || lipgloss.ColorProfile() == termenv.Ascii {
		return s.effective, content
	}
	style := s.effective.
		UnsetUnderline().UnsetUnderlineSpaces().
		UnsetStrikethrough().UnsetStrikethroughSpaces()
	return style, relined(content, s.linesOn, s.linesOff)
}

// lineSequences returns the escape sequences that start and end the underline and the
// strikethrough of style, or two empty strings when it has neither.
func lineSequences(style lipgloss.Style) (on, off string) {
	var start, end []string
	if style.GetUnderline() {
		start, end = append(start, "4"), append(end, "24")
	}
	if style.GetStrikethrough() {
		start, end = append(start, "9"), append(end, "29")
	}
	if len(start) == 0 {
		return "", ""
	}
	return "\x1b[" + strings.Join(start, ";") + "m", "\x1b[" + strings.Join(end, ";") + "m"
}

// relined writes on at the start of every line of text that has any, and again after
// every escape sequence in it that resets the text attributes, and off at its end.
func relined(text, on, off string) string {
	if on == "" {
		return text
	}
	var b strings.Builder
	for i, line := range strings.Split(text, "\n") {
		if i > 0 {
			b.WriteByte('\n')
		}
		if line == "" {
			continue
		}
		b.WriteString(on)
		for {
			start := strings.Index(line, "\x1b[")
			if start < 0 {
				break
			}
			// parameters run up to the letter that ends the sequence
			end := start + 2
			for end < len(line) && (line[end] < '@' || line[end] > '~') {
				end++
			}
			if end == len(line) {
				break
			}
			b.WriteString(line[:end+1])
			if line[end] == 'm' && resetsAttributes(line[start+2:end]) {
				b.WriteString(on)
			}
			line = line[end+1:]
		}
		b.WriteString(line)
		b.WriteString(off)
	}
	return b.String()
}

// resetsAttributes reports whether the parameters of a text attribute sequence end an
// underline or a strikethrough: a parameter outside the numbers of a color resets every
// attribute (empty or 0) or turns one of the lines off (24, 29, 59 or 4:0).
func resetsAttributes(params string) bool {
	fields := strings.Split(params, ";")
	for i := 0; i < len(fields); i++ {
		switch strings.TrimLeft(fields[i], "0") {
		case "", "24", "29", "59", "4:0":
			return true
		case "38", "48", "58":
			// a color follows as 5;n or 2;r;g;b
			if i+1 < len(fields) && fields[i+1] == "5" {
				i += 2
			} else if i+1 < len(fields) && fields[i+1] == "2" {
				i += 4
			}
		}
	}
	return false
}

// passed returns what the element passes to its children: its effective style when
// passing is enabled and the style is not empty.
func (s *styling) passed() inherited {
	if !s.passing || s.id == 0 {
		return inherited{}
	}
	return inherited{style: &s.effective, id: s.id}
}

// plain reports whether the effective style is the empty style, which renders its
// content unchanged.
func (s *styling) plain() bool {
	return s.id == 0
}
