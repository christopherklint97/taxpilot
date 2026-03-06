package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"taxpilot/internal/interview"
	"taxpilot/internal/state"
	"taxpilot/internal/tui"
)

// InterviewView is the Bubble Tea model for the interview screen.
type InterviewView struct {
	engine       *interview.Engine
	input        string // current text input
	err          string // error message to display
	helpText     string // contextual help text shown after "?" command
	aiHelpText   string // RAG-powered explanation shown after "??" command
	aiLoading    bool   // true while waiting for AI explanation
	done         bool   // all questions answered
	taxYear      int
	stateCode    string
	width        int
	height       int
}

// NewInterviewView creates a new InterviewView with the given engine.
func NewInterviewView(engine *interview.Engine, taxYear int, stateCode string) InterviewView {
	return InterviewView{
		engine:    engine,
		taxYear:   taxYear,
		stateCode: stateCode,
	}
}

// Init satisfies tea.Model.
func (m InterviewView) Init() tea.Cmd {
	return nil
}

// Update satisfies tea.Model.
func (m InterviewView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tui.ExplanationResponseMsg:
		m.aiLoading = false
		if msg.Err != nil {
			m.aiHelpText = "Error loading explanation: " + msg.Err.Error()
		} else {
			m.aiHelpText = msg.Explanation
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			// Save state and quit
			m.saveState()
			return m, tea.Quit

		case tea.KeyEnter:
			if m.done {
				return m, nil
			}
			// Handle "??" RAG-powered explanation command
			if m.input == "??" {
				q := m.engine.Current()
				if q != nil {
					m.aiLoading = true
					m.aiHelpText = ""
					m.input = ""
					return m, func() tea.Msg {
						return tui.RequestExplanationMsg{
							FieldKey: q.Key,
							Label:    q.Prompt,
							FormName: q.FormName,
						}
					}
				}
				m.input = ""
				return m, nil
			}
			// Handle "?" help command
			if m.input == "?" {
				q := m.engine.Current()
				if q != nil {
					cp := interview.GetContextualPrompt(q.Key, q.Prompt, m.stateCode)
					m.helpText = cp.HelpText
					if cp.CANote != "" && m.stateCode == "CA" {
						m.helpText += "\n\nCalifornia note: " + cp.CANote
					}
					if m.helpText == "" {
						m.helpText = "No additional help available for this question."
					}
				}
				m.input = ""
				return m, nil
			}
			// Clear help text on next answer
			m.helpText = ""
			m.aiHelpText = ""
			m.aiLoading = false
			// If input is empty and a prior-year default exists, accept the default
			if m.input == "" && m.engine.GetPriorYearDefault() != nil {
				if err := m.engine.AcceptDefault(); err != nil {
					m.err = err.Error()
					return m, nil
				}
			} else {
				if err := m.engine.Answer(m.input); err != nil {
					m.err = err.Error()
					return m, nil
				}
			}
			m.input = ""
			m.err = ""

			if !m.engine.HasNext() {
				m.done = true
				// Solve and transition to summary
				results, err := m.engine.Solve()
				if err != nil {
					m.err = fmt.Sprintf("Error computing results: %v", err)
					m.done = false
					return m, nil
				}
				return m, func() tea.Msg {
					return tui.ShowSummaryMsg{
						Results:   results,
						StrInputs: m.engine.StrInputs(),
						TaxYear:   m.taxYear,
						State:     m.stateCode,
					}
				}
			}
			return m, nil

		case tea.KeyBackspace, tea.KeyDelete:
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
				m.err = ""
			} else {
				// Go back to previous question
				if m.engine.Back() {
					m.err = ""
				}
			}
			return m, nil

		case tea.KeyEsc:
			// Save state and quit
			m.saveState()
			return m, tea.Quit

		case tea.KeyRunes:
			key := msg.String()
			if key == "q" && m.input == "" {
				// Save state and quit
				m.saveState()
				return m, tea.Quit
			}
			m.input += key
			m.err = ""
			return m, nil

		case tea.KeySpace:
			m.input += " "
			return m, nil
		}
	}

	return m, nil
}

// saveState persists the current interview state to disk.
func (m *InterviewView) saveState() {
	ret := state.NewTaxReturn(m.taxYear, m.stateCode)
	ret.Inputs = m.engine.Inputs()
	ret.StrInputs = m.engine.StrInputs()
	if fs, ok := ret.StrInputs["1040:filing_status"]; ok {
		ret.FilingStatus = fs
	}
	_ = state.Save(state.DefaultStorePath(), ret)
}

// View satisfies tea.Model.
func (m InterviewView) View() string {
	if m.done {
		return tui.BorderStyle.Render(
			tui.SuccessStyle.Render("All questions answered! Computing results..."),
		) + "\n"
	}

	q := m.engine.Current()
	if q == nil {
		return ""
	}

	cur, total := m.engine.Progress()

	// Progress indicator
	progress := tui.HelpStyle.Render(
		fmt.Sprintf("Question %d of %d", cur+1, total),
	)

	// Progress bar
	barWidth := 30
	filled := 0
	if total > 0 {
		filled = (cur * barWidth) / total
	}
	bar := tui.HighlightStyle.Render(strings.Repeat("█", filled)) +
		tui.HelpStyle.Render(strings.Repeat("░", barWidth-filled))

	// Form context
	formContext := tui.TitleStyle.Render(q.FormName)

	// Get contextual prompt for enhanced question text
	cp := interview.GetContextualPrompt(q.Key, q.Prompt, m.stateCode)

	// Question prompt (use contextual prompt instead of raw prompt)
	prompt := tui.PromptStyle.Render(cp.Prompt)

	// Contextual help text below the prompt
	var contextHelp string
	if cp.HelpText != "" {
		contextHelp = tui.HelpStyle.Render(cp.HelpText)
	}

	// CA-specific note
	var caNote string
	if cp.CANote != "" && m.stateCode == "CA" {
		caNote = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5C07B")).
			Italic(true).
			Render("CA: " + cp.CANote)
	}

	// User-triggered help text (from "?" command)
	var userHelp string
	if m.helpText != "" {
		userHelp = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#98C379")).
			Italic(true).
			Render(m.helpText)
	}

	// AI-powered explanation (from "??" command)
	var aiHelp string
	if m.aiLoading {
		aiHelp = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5C07B")).
			Italic(true).
			Render("Loading AI explanation...")
	} else if m.aiHelpText != "" {
		aiHelp = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#56B6C2")).
			Italic(true).
			Render(m.aiHelpText)
	}

	// Prior-year default indicator
	var priorYearBlock string
	pyd := m.engine.GetPriorYearDefault()
	if pyd != nil {
		priorYearBlock = tui.SuccessStyle.Render(
			fmt.Sprintf("Last year: %s", pyd.PriorValue),
		) + "\n" + tui.HelpStyle.Render(
			"Press Enter to keep last year's value, or type a new one",
		)
	}

	// Options for enum fields
	var optionsBlock string
	if len(q.Options) > 0 {
		var lines []string
		for i, opt := range q.Options {
			label := formatOptionLabel(opt)
			lines = append(lines, fmt.Sprintf("  %s %s",
				tui.HighlightStyle.Render(fmt.Sprintf("[%d]", i+1)),
				label,
			))
		}
		optionsBlock = strings.Join(lines, "\n")
	}

	// Input area
	cursor := tui.HighlightStyle.Render("▸ ")
	inputLine := cursor + tui.InputStyle.Render(m.input) +
		tui.HighlightStyle.Render("█")

	// Error message
	var errBlock string
	if m.err != "" {
		errBlock = tui.ErrorStyle.Render("⚠ " + m.err)
	}

	// Help text
	help := tui.HelpStyle.Render("Enter: submit  |  Backspace: go back  |  ?: help  |  ??: AI explain  |  q: save & quit")

	// Compose layout
	parts := []string{
		progress,
		bar,
		"",
		formContext,
		"",
		prompt,
	}
	if contextHelp != "" {
		parts = append(parts, contextHelp)
	}
	if caNote != "" {
		parts = append(parts, caNote)
	}
	if priorYearBlock != "" {
		parts = append(parts, priorYearBlock)
	}
	if optionsBlock != "" {
		parts = append(parts, optionsBlock)
	}
	if userHelp != "" {
		parts = append(parts, "", userHelp)
	}
	if aiHelp != "" {
		parts = append(parts, "", aiHelp)
	}
	parts = append(parts, "", inputLine)
	if errBlock != "" {
		parts = append(parts, errBlock)
	}
	parts = append(parts, "", help)

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return tui.BorderStyle.Render(content) + "\n"
}

// formatOptionLabel converts a filing status code to a human-readable label.
func formatOptionLabel(opt string) string {
	switch opt {
	case "single":
		return "Single"
	case "mfj":
		return "Married Filing Jointly"
	case "mfs":
		return "Married Filing Separately"
	case "hoh":
		return "Head of Household"
	case "qss":
		return "Qualifying Surviving Spouse"
	default:
		return opt
	}
}
