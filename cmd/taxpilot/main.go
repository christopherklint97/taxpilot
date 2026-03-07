package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"taxpilot/internal/tui"
	"taxpilot/internal/tui/views"
)

func main() {
	welcome := views.NewWelcomeModel(2025, "CA")
	app := tui.NewApp(welcome, tui.ViewFactory{})

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
