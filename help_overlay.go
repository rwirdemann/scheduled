package scheduled

import (
	"image/color"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
)

// HelpOverlay renders a centered keybindings box composited over the
// main UI.
type HelpOverlay struct {
	boxStyle    lipgloss.Style
	borderColor color.Color
	keyStyle    lipgloss.Style
	descStyle   lipgloss.Style
}

// NewHelpOverlay returns a HelpOverlay configured from t and s.
func NewHelpOverlay(t Theme, s help.Styles) HelpOverlay {
	return HelpOverlay{
		boxStyle:    t.OverlayBox,
		borderColor: t.OverlayBorderColor,
		keyStyle:    s.FullKey,
		descStyle:   s.FullDesc,
	}
}

// Render produces the overlay string for the given key binding groups.
func (o HelpOverlay) Render(groups [][]key.Binding) string {
	box := o.boxStyle.Render(o.renderColumns(groups))
	return o.insertBorderTitle(box, " Keybindings ")
}

// renderColumns arranges the binding groups into three side-by-side
// columns, two groups each.
func (o HelpOverlay) renderColumns(groups [][]key.Binding) string {
	colStyle := lipgloss.NewStyle().PaddingRight(6)
	col1 := colStyle.Render(o.renderGroups(groups[0:2]))
	col2 := colStyle.Render(o.renderGroups(groups[2:4]))
	col3 := o.renderGroups(groups[4:6])
	return lipgloss.JoinHorizontal(lipgloss.Top, col1, col2, col3)
}

// renderGroups renders a slice of binding groups as "key  desc" lines,
// with a blank line between groups.
func (o HelpOverlay) renderGroups(groups [][]key.Binding) string {
	var parts []string
	for _, group := range groups {
		var lines []string
		for _, b := range group {
			h := b.Help()
			k := o.keyStyle.Width(14).Render(h.Key)
			d := o.descStyle.Render(h.Desc)
			lines = append(lines, k+d)
		}
		parts = append(parts, strings.Join(lines, "\n"))
	}
	return strings.Join(parts, "\n\n")
}

// insertBorderTitle replaces the top border line of a lipgloss-rendered
// box with a version that includes a centered title.
func (o HelpOverlay) insertBorderTitle(box, title string) string {
	lines := strings.Split(box, "\n")
	if len(lines) == 0 {
		return box
	}
	w := lipgloss.Width(lines[0])
	titleW := lipgloss.Width(title)
	if titleW >= w-2 {
		return box
	}
	leftPad := (w - titleW) / 2
	rightPad := w - titleW - leftPad
	newTop := "┌" +
		strings.Repeat("─", leftPad-1) +
		title +
		strings.Repeat("─", rightPad-1) +
		"┐"
	lines[0] = lipgloss.NewStyle().
		Foreground(o.borderColor).
		Render(newTop)
	return strings.Join(lines, "\n")
}
