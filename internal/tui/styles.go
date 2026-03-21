package tui

import "github.com/charmbracelet/lipgloss"

var (
	// TitleStyle is used for bold, colored headers.
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#4A90D9"))

	// PromptStyle is used for question text shown to the user.
	PromptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D7D7D7")).
			Bold(true)

	// InputStyle is used for the user input area.
	InputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	// HelpStyle is used for subtle help text.
	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true)

	// ErrorStyle is used for red error text.
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E06C75")).
			Bold(true)

	// SuccessStyle is used for green success messages.
	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#98C379")).
			Bold(true)

	// BorderStyle wraps content in a rounded border.
	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4A90D9")).
			Padding(1, 2)

	// HighlightStyle is used for emphasized values.
	HighlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#61AFEF")).
			Bold(true)
)

// ContentWidth returns the usable width inside a BorderStyle box.
// BorderStyle has 1-char border + 2-char padding on each side = 6 total.
func ContentWidth(termWidth int) int {
	w := termWidth - 6
	if w < 40 {
		w = 40
	}
	return w
}
