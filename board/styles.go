package board

import (
	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"
)

// DoneStyle is the style applied to completed task titles.
var DoneStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

// newDelegate returns a list delegate configured for dark or light backgrounds.
func newDelegate(dark bool) list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.ShowDescription = false
	d.SetSpacing(0)
	if !dark {
		d.Styles.NormalTitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("238")).
			Padding(0, 0, 0, 2)
		d.Styles.SelectedTitle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.Color("164")).
			Foreground(lipgloss.Color("164")).
			Bold(true).
			Padding(0, 0, 0, 1)
		d.Styles.DimmedTitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Padding(0, 0, 0, 2)
	}
	return d
}

// setDoneStyle updates the package-level DoneStyle for dark or light backgrounds.
func setDoneStyle(dark bool) {
	if dark {
		DoneStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	} else {
		DoneStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	}
}
