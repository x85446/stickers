package flexbox

import (
	"strings"
	"sync/atomic"

	"charm.land/lipgloss/v2"
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
	marginX     int // left and right margin together
	marginY     int // top and bottom margin together
	paddingX    int // left and right padding together
	paddingY    int // top and bottom padding together
}

// edgeY returns the lines the margins and border take, above and below together.
func (f frame) edgeY() int { return f.top + f.bottom }

// sized returns style set to render at exactly width by height, margins included.
// Lipgloss counts the border and the padding into a style's width and height and adds
// the margins outside, so the margins are taken off first. A width of 0 or less leaves
// the width to the content.
func (f frame) sized(style lipgloss.Style, width, height int) lipgloss.Style {
	return style.
		Width(width - f.marginX).MaxWidth(width).
		Height(height - f.marginY)
}

// styling is the style of a box, row, column or cell: the style set on it and the
// effective style it renders with, which adds what its parent passes.
type styling struct {
	own     lipgloss.Style
	ownID   uint64 // identity of own; 0 while own is the empty style
	passing bool   // pass the effective style on to the children

	effective lipgloss.Style
	id        uint64 // identity of effective; 0 while it is the empty style
	frame     frame  // sizes of effective

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
	s.frame = frame{}
	if s.id != 0 {
		s.frame = frame{
			top:      s.effective.GetMarginTop() + s.effective.GetBorderTopSize(),
			bottom:   s.effective.GetBorderBottomSize() + s.effective.GetMarginBottom(),
			edgeX:    s.effective.GetHorizontalMargins() + s.effective.GetHorizontalBorderSize(),
			marginX:  s.effective.GetHorizontalMargins(),
			marginY:  s.effective.GetVerticalMargins(),
			paddingX: s.effective.GetHorizontalPadding(),
			paddingY: s.effective.GetVerticalPadding(),
		}
	}
	s.resolved, s.resolvedOwn, s.resolvedFrom = true, s.ownID, from.id
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
