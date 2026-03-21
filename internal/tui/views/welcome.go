package views

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"taxpilot/internal/tui"
)

const logo = `
 _____          ____  _ _       _
|_   _|_ ___  _|  _ \(_) | ___ | |_
  | |/ _` + "`" + ` \ \/ / |_) | | |/ _ \| __|
  | | (_| |>  <|  __/| | | (_) | |_
  |_|\__,_/_/\_\_|   |_|_|\___/ \__|
`

// WelcomeModel is the Bubble Tea model for the welcome screen.
type WelcomeModel struct {
	taxYear          int
	state            string
	width            int
	height           int
	priorYearLoaded  bool     // true if prior-year data is available
	priorYearLabel   string   // e.g. "2024 return loaded"
	priorFormNames   []string // names of loaded forms
	loadingPriorYear bool     // true when prompting for file path
	filePathInput    string   // text input for file path
}

// NewWelcomeModel creates a WelcomeModel with the given tax year and state.
func NewWelcomeModel(taxYear int, stateCode string) WelcomeModel {
	return WelcomeModel{
		taxYear: taxYear,
		state:   stateCode,
	}
}

// SetPriorYearLoaded marks prior-year data as available (single form).
func (m *WelcomeModel) SetPriorYearLoaded(year int) {
	m.priorYearLoaded = true
	m.priorYearLabel = fmt.Sprintf("%d return loaded", year)
}

// SetPriorYearLoadedMulti marks prior-year data as available with form names.
func (m *WelcomeModel) SetPriorYearLoadedMulti(year int, formNames []string) {
	m.priorYearLoaded = true
	m.priorFormNames = appendUniqueStrings(m.priorFormNames, formNames)
	if year > 0 {
		m.priorYearLabel = fmt.Sprintf("%d return loaded (%d forms)", year, len(m.priorFormNames))
	} else {
		m.priorYearLabel = fmt.Sprintf("%d forms loaded", len(m.priorFormNames))
	}
}

// Init satisfies tea.Model.
func (m WelcomeModel) Init() tea.Cmd {
	return nil
}

// Update satisfies tea.Model.
func (m WelcomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tui.PriorYearImportedMsg:
		if msg.Err != nil {
			m.loadingPriorYear = false
			m.filePathInput = ""
			return m, nil
		}
		m.priorYearLoaded = true
		m.priorFormNames = appendUniqueStrings(m.priorFormNames, msg.FormNames)
		if msg.TaxYear > 0 {
			m.priorYearLabel = fmt.Sprintf("%d return loaded (%d forms)", msg.TaxYear, len(m.priorFormNames))
		} else {
			m.priorYearLabel = fmt.Sprintf("%d forms loaded", len(m.priorFormNames))
		}
		m.loadingPriorYear = false
		m.filePathInput = ""
		return m, nil

	case tea.KeyMsg:
		if m.loadingPriorYear {
			// Handle file path input mode
			switch msg.Type {
			case tea.KeyEnter:
				path := m.filePathInput
				m.filePathInput = ""
				if path == "" {
					m.loadingPriorYear = false
					return m, nil
				}
				return m, func() tea.Msg {
					return tui.ImportPriorYearMsg{FilePaths: []string{path}}
				}
			case tea.KeyEsc:
				m.loadingPriorYear = false
				m.filePathInput = ""
				return m, nil
			case tea.KeyBackspace, tea.KeyDelete:
				if len(m.filePathInput) > 0 {
					m.filePathInput = m.filePathInput[:len(m.filePathInput)-1]
				}
				return m, nil
			case tea.KeyRunes:
				m.filePathInput += msg.String()
				return m, nil
			case tea.KeySpace:
				m.filePathInput += " "
				return m, nil
			case tea.KeyCtrlC:
				return m, tea.Quit
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "n":
			return m, func() tea.Msg {
				return tui.StartInterviewMsg{
					TaxYear:   m.taxYear,
					StateCode: m.state,
					Continue:  false,
				}
			}
		case "l":
			m.loadingPriorYear = true
			m.filePathInput = ""
			return m, nil
		case "c":
			return m, func() tea.Msg {
				return tui.StartInterviewMsg{
					TaxYear:   m.taxYear,
					StateCode: m.state,
					Continue:  true,
				}
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

// View satisfies tea.Model.
func (m WelcomeModel) View() string {
	titleBlock := tui.TitleStyle.Render(logo)

	stateName := m.state
	if stateName == "CA" {
		stateName = "California"
	}

	info := fmt.Sprintf(
		"Tax Year: %s    State: %s",
		tui.HighlightStyle.Render(fmt.Sprintf("%d", m.taxYear)),
		tui.HighlightStyle.Render(stateName),
	)

	var menuParts []string
	menuParts = append(menuParts,
		tui.PromptStyle.Render("What would you like to do?"),
		"",
		"  [N] New return",
		"  [L] Load prior-year return (PDF)",
		"  [C] Continue saved session",
	)

	if m.priorYearLoaded {
		menuParts = append(menuParts, "",
			tui.SuccessStyle.Render("  Prior-year data: "+m.priorYearLabel),
		)
		for _, name := range m.priorFormNames {
			menuParts = append(menuParts, tui.SuccessStyle.Render("    - "+name))
		}
		menuParts = append(menuParts, "",
			"  [L] Load additional prior-year return (PDF)",
		)
	}

	menu := lipgloss.JoinVertical(lipgloss.Left, menuParts...)

	var help string
	if m.loadingPriorYear {
		help = tui.PromptStyle.Render("Enter path to prior-year PDF (file or directory):") + "\n" +
			tui.HighlightStyle.Render("▸ ") + tui.InputStyle.Render(m.filePathInput) +
			tui.HighlightStyle.Render("█") + "\n" +
			tui.HelpStyle.Render("Enter: import  |  Esc: cancel")
	} else {
		help = tui.HelpStyle.Render("q: quit  |  ?: help")
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		titleBlock,
		info,
		"",
		menu,
		"",
		help,
	)

	return tui.BorderStyle.Render(content) + "\n"
}

// appendUniqueStrings appends items from add to base, skipping duplicates.
func appendUniqueStrings(base, add []string) []string {
	seen := make(map[string]bool, len(base))
	for _, s := range base {
		seen[s] = true
	}
	for _, s := range add {
		if !seen[s] {
			base = append(base, s)
			seen[s] = true
		}
	}
	return base
}
