package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"taxpilot/internal/calc"
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

	// Calculator sub-mode
	calcMode      bool               // true when calculator is active
	calcInput     string             // expression being typed in calculator
	calcResult    string             // computed result display
	calcResultVal float64            // numeric result for submitting
	calcHasResult bool               // true when a valid result is available
	calcRates     map[string]float64 // cached exchange rates
	calcRatesErr  string             // error fetching rates
	calcLoading   bool               // true while fetching rates
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

	case tui.WhyAskedResponseMsg:
		m.aiLoading = false
		if msg.Err != nil {
			m.aiHelpText = "Error loading explanation: " + msg.Err.Error()
		} else {
			m.aiHelpText = msg.Explanation
		}
		return m, nil

	case tui.CADiffResponseMsg:
		m.aiLoading = false
		if msg.Err != nil {
			m.aiHelpText = "Error loading explanation: " + msg.Err.Error()
		} else {
			m.aiHelpText = msg.Explanation
		}
		return m, nil

	case tui.ExchangeRatesMsg:
		m.calcLoading = false
		if msg.Err != nil {
			m.calcRatesErr = msg.Err.Error()
		} else {
			m.calcRates = msg.Rates
			m.calcRatesErr = ""
		}
		return m, nil

	case tea.KeyMsg:
		// Calculator sub-mode key handling
		if m.calcMode {
			return m.updateCalcMode(msg)
		}
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
			// Handle "why" command — explain why this question is being asked
			if m.input == "why" {
				q := m.engine.Current()
				if q != nil {
					m.aiLoading = true
					m.aiHelpText = ""
					m.input = ""
					// Build answered keys from string inputs
					answeredKeys := make(map[string]string)
					for k, v := range m.engine.StrInputs() {
						answeredKeys[k] = v
					}
					for k, v := range m.engine.Inputs() {
						if _, exists := answeredKeys[k]; !exists {
							answeredKeys[k] = fmt.Sprintf("%v", v)
						}
					}
					// Get filing status
					filingStatus := ""
					if fs, ok := m.engine.StrInputs()["1040:filing_status"]; ok {
						filingStatus = fs
					}
					return m, func() tea.Msg {
						return tui.RequestWhyAskedMsg{
							FieldKey:     q.Key,
							Label:        q.Prompt,
							FilingStatus: filingStatus,
							AnsweredKeys: answeredKeys,
						}
					}
				}
				m.input = ""
				return m, nil
			}
			// Handle "ca" command — explain CA vs federal difference
			if m.input == "ca" && m.stateCode == "CA" {
				q := m.engine.Current()
				if q != nil {
					m.aiLoading = true
					m.aiHelpText = ""
					m.input = ""
					return m, func() tea.Msg {
						return tui.RequestCADiffMsg{
							FieldKey: q.Key,
							Label:    q.Prompt,
						}
					}
				}
				m.input = ""
				return m, nil
			}
			// Handle "calc" command — enter calculator mode
			if m.input == "calc" {
				m.input = ""
				m.calcMode = true
				m.calcInput = ""
				m.calcResult = ""
				m.calcHasResult = false
				// Fetch exchange rates in the background if not cached
				if m.calcRates == nil && !m.calcLoading {
					m.calcLoading = true
					return m, func() tea.Msg {
						rates, err := calc.FetchRates()
						return tui.ExchangeRatesMsg{Rates: rates, Err: err}
					}
				}
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

// updateCalcMode handles key events when the calculator is active.
func (m InterviewView) updateCalcMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		// Exit calculator without using the result
		m.calcMode = false
		m.calcInput = ""
		m.calcResult = ""
		m.calcHasResult = false
		return m, nil

	case tea.KeyEnter:
		if m.calcInput == "" && m.calcHasResult {
			// Empty input + result available: use the result as the field value
			m.calcMode = false
			m.input = fmt.Sprintf("%.2f", m.calcResultVal)
			m.calcInput = ""
			m.calcResult = ""
			m.calcHasResult = false
			return m, nil
		}
		if m.calcInput != "" {
			// Evaluate the expression
			result, breakdown, err := calc.Eval(m.calcInput, m.calcRates)
			if err != nil {
				m.calcResult = "Error: " + err.Error()
				m.calcHasResult = false
			} else {
				m.calcResultVal = result
				m.calcHasResult = true
				m.calcInput = "" // clear input so next Enter uses the result
				if breakdown != "" {
					m.calcResult = breakdown
				} else {
					m.calcResult = fmt.Sprintf("= %.2f", result)
				}
			}
		}
		return m, nil

	case tea.KeyBackspace, tea.KeyDelete:
		if len(m.calcInput) > 0 {
			m.calcInput = m.calcInput[:len(m.calcInput)-1]
			m.calcHasResult = false
			m.calcResult = ""
		}
		return m, nil

	case tea.KeyRunes:
		m.calcInput += msg.String()
		// Clear previous result when typing new input
		m.calcHasResult = false
		m.calcResult = ""
		return m, nil

	case tea.KeySpace:
		m.calcInput += " "
		return m, nil
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
	if m.calcMode {
		return m.viewCalc()
	}

	if m.done {
		cw := tui.ContentWidth(m.width)
		return tui.BorderStyle.Width(cw).Render(
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

	// Content width inside the border
	contentW := tui.ContentWidth(m.width)

	// Form context
	formContext := tui.TitleStyle.Width(contentW).Render(q.FormName)

	// Get contextual prompt for enhanced question text
	cp := interview.GetContextualPrompt(q.Key, q.Prompt, m.stateCode)

	// Question prompt (use contextual prompt instead of raw prompt)
	prompt := tui.PromptStyle.Width(contentW).Render(cp.Prompt)

	// Contextual help text below the prompt
	var contextHelp string
	if cp.HelpText != "" {
		contextHelp = tui.HelpStyle.Width(contentW).Render(cp.HelpText)
	}

	// CA-specific note
	var caNote string
	if cp.CANote != "" && m.stateCode == "CA" {
		caNote = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5C07B")).
			Italic(true).
			Width(contentW).
			Render("CA: " + cp.CANote)
	}

	// User-triggered help text (from "?" command)
	var userHelp string
	if m.helpText != "" {
		userHelp = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#98C379")).
			Italic(true).
			Width(contentW).
			Render(m.helpText)
	}

	// AI-powered explanation (from "??" command)
	var aiHelp string
	if m.aiLoading {
		aiHelp = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5C07B")).
			Italic(true).
			Width(contentW).
			Render("Loading AI explanation...")
	} else if m.aiHelpText != "" {
		aiHelp = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#56B6C2")).
			Italic(true).
			Width(contentW).
			Render(m.aiHelpText)
	}

	// Prior-year default indicator
	var priorYearBlock string
	pyd := m.engine.GetPriorYearDefault()
	if pyd != nil {
		priorYearBlock = tui.SuccessStyle.Width(contentW).Render(
			fmt.Sprintf("Last year: %s", pyd.PriorValue),
		) + "\n" + tui.HelpStyle.Width(contentW).Render(
			"Press Enter to keep last year's value, or type a new one",
		)
		if pyd.CANote != "" {
			priorYearBlock += "\n" + lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E5C07B")).
				Italic(true).
				Width(contentW).
				Render("CA: "+pyd.CANote)
		}
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

	// Help text — wrap into multiple lines to fit terminal width
	helpItems := []string{
		"Enter: submit", "Backspace: go back", "?: help",
		"??: AI explain", "why: why asked", "calc: calculator",
		"q: save & quit",
	}
	if m.stateCode == "CA" {
		helpItems = append(helpItems, "ca: CA diff")
	}
	sep := "  |  "
	var helpLines []string
	line := ""
	for i, item := range helpItems {
		candidate := line
		if candidate != "" {
			candidate += sep
		}
		candidate += item
		if line != "" && lipgloss.Width(candidate) > contentW {
			helpLines = append(helpLines, line)
			line = item
		} else {
			if i > 0 && line != "" {
				line += sep
			}
			line += item
		}
	}
	if line != "" {
		helpLines = append(helpLines, line)
	}
	help := tui.HelpStyle.Render(strings.Join(helpLines, "\n"))

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
	return tui.BorderStyle.Width(contentW).Render(content) + "\n"
}

// viewCalc renders the calculator overlay.
func (m InterviewView) viewCalc() string {
	contentW := tui.ContentWidth(m.width)

	title := tui.TitleStyle.Width(contentW).Render("Calculator")

	// Instructions
	instructions := tui.HelpStyle.Width(contentW).Render(
		"Type an expression and press Enter to evaluate.\n" +
			"Supports: +, -, *, /  and currency codes (e.g., 1000 EUR, 500 GBP + 200 SEK)")

	// Exchange rate status
	var rateStatus string
	if m.calcLoading {
		rateStatus = tui.HelpStyle.Render("Loading exchange rates...")
	} else if m.calcRatesErr != "" {
		rateStatus = tui.ErrorStyle.Render("Rates: " + m.calcRatesErr)
	} else if m.calcRates != nil {
		rateStatus = tui.SuccessStyle.Render("Exchange rates loaded")
	}

	// Input
	cursor := tui.HighlightStyle.Render("▸ ")
	inputLine := cursor + tui.InputStyle.Render(m.calcInput) +
		tui.HighlightStyle.Render("█")

	// Result
	var resultBlock string
	if m.calcResult != "" {
		resultStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#98C379")).
			Bold(true).
			Width(contentW)
		if m.calcHasResult {
			resultStyle = resultStyle.Foreground(lipgloss.Color("#98C379"))
		} else {
			resultStyle = resultStyle.Foreground(lipgloss.Color("#E06C75"))
		}
		resultBlock = resultStyle.Render(m.calcResult)
	}

	// Help bar
	var helpText string
	if m.calcHasResult {
		helpText = tui.HelpStyle.Render("Enter: use result  |  Esc: cancel  |  Type: new expression")
	} else {
		helpText = tui.HelpStyle.Render("Enter: calculate  |  Esc: cancel")
	}

	parts := []string{title, "", instructions}
	if rateStatus != "" {
		parts = append(parts, rateStatus)
	}
	parts = append(parts, "", inputLine)
	if resultBlock != "" {
		parts = append(parts, "", resultBlock)
	}
	parts = append(parts, "", helpText)

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return tui.BorderStyle.Width(contentW).Render(content) + "\n"
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
