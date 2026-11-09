// Package about draws the box the demos show on top of themselves when a is pressed.
package about

import "charm.land/lipgloss/v2"

var style = lipgloss.NewStyle().
	Padding(2, 4).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#874BFD")).
	Background(lipgloss.Color("#1a1a2e"))

// Over returns content with text in a box on top of it, at the center of the width by
// height area the content fills. The box is a layer of its own, so the content stays
// visible around it; a box larger than the area is cut at the area's edges.
func Over(content, text string, width, height int) string {
	box := style.Render(text)
	boxWidth, boxHeight := lipgloss.Size(box)
	layers := lipgloss.NewCompositor(
		lipgloss.NewLayer(content),
		lipgloss.NewLayer(box).X((width-boxWidth)/2).Y((height-boxHeight)/2).Z(1),
	)
	return lipgloss.NewCanvas(width, height).Compose(layers).Render()
}
