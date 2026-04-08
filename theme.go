package scheduled

import "charm.land/lipgloss/v2"

// Theme holds the lipgloss styles for a dark or light terminal background.
type Theme struct {
	HelpKey  lipgloss.Style
	HelpDesc lipgloss.Style
	HelpSep  lipgloss.Style
	Status   lipgloss.Style
}

// NewTheme returns a Theme configured for a dark or light background.
func NewTheme(dark bool) Theme {
	if dark {
		return Theme{
			HelpKey: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205")),
			HelpDesc: lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")),
			HelpSep: lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")),
			Status: lipgloss.NewStyle().
				Foreground(lipgloss.Color("42")).
				Bold(true).
				Padding(0, 1),
		}
	}
	return Theme{
		HelpKey: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("164")),
		HelpDesc: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
		HelpSep: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")),
		Status: lipgloss.NewStyle().
			Foreground(lipgloss.Color("34")).
			Bold(true).
			Padding(0, 1),
	}
}
