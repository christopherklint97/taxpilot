package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"taxpilot/internal/calc"
	"taxpilot/internal/forms"
	"taxpilot/internal/interview"
	"taxpilot/internal/llm"
	"taxpilot/internal/state"
	"taxpilot/internal/tui"
)

// InterviewView is the Bubble Tea model for the interview screen.
type InterviewView struct {
	engine     *interview.Engine
	input      string // current text input
	err        string // error message to display
	helpText   string // contextual help text shown after "?" command
	aiHelpText string // RAG-powered explanation shown after "??" command
	aiLoading  bool   // true while waiting for AI explanation
	done       bool   // all questions answered
	taxYear    int
	stateCode  string
	width      int
	height     int

	// Cursor position within input (0 = before first char, len(input) = after last)
	cursor int

	// Input history (Up/Down arrow to recall previous inputs)
	history    []string // past inputs, newest last
	historyIdx int      // -1 = not browsing; 0..len-1 = browsing position
	historyBuf string   // saves current input when entering history mode

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
		engine:     engine,
		taxYear:    taxYear,
		stateCode:  stateCode,
		historyIdx: -1,
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
			m.aiHelpText = "Error: " + msg.Err.Error() + " (press ↑ then Enter to retry)"
		} else {
			m.aiHelpText = msg.Explanation
		}
		return m, nil

	case tui.WhyAskedResponseMsg:
		m.aiLoading = false
		if msg.Err != nil {
			m.aiHelpText = "Error: " + msg.Err.Error() + " (press ↑ then Enter to retry)"
		} else {
			m.aiHelpText = msg.Explanation
		}
		return m, nil

	case tui.CADiffResponseMsg:
		m.aiLoading = false
		if msg.Err != nil {
			m.aiHelpText = "Error: " + msg.Err.Error() + " (press ↑ then Enter to retry)"
		} else {
			m.aiHelpText = msg.Explanation
		}
		return m, nil

	case tui.AIPromptResponseMsg:
		m.aiLoading = false
		if msg.Err != nil {
			m.aiHelpText = "Error: " + msg.Err.Error() + " (press ↑ then Enter to retry)"
		} else {
			m.aiHelpText = msg.Answer
		}
		return m, nil

	case tui.AIStreamChunkMsg:
		if msg.Err != nil {
			m.aiLoading = false
			m.aiHelpText += "\nError: " + msg.Err.Error() + " (press ↑ then Enter to retry)"
			return m, nil
		}
		if msg.Done {
			m.aiLoading = false
			return m, nil
		}
		// Append streaming text
		m.aiHelpText += msg.Text
		// Chain: read the next chunk from the channel
		if msg.Ch != nil {
			ch := msg.Ch
			return m, func() tea.Msg {
				raw, ok := <-ch
				if !ok {
					return tui.AIStreamChunkMsg{Done: true}
				}
				chunk := raw.(llm.StreamChunk)
				return tui.AIStreamChunkMsg{
					Text: chunk.Text,
					Err:  chunk.Err,
					Done: chunk.Done,
					Ch:   ch,
				}
			}
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
			// Handle "ai <question>" — free-form AI question about current field
			if strings.HasPrefix(m.input, "ai ") && len(m.input) > 3 {
				userQuestion := strings.TrimSpace(m.input[3:])
				q := m.engine.Current()
				if q != nil && userQuestion != "" {
					m.history = append(m.history, m.input)
					m.historyIdx = -1
					m.aiLoading = true
					m.aiHelpText = ""
					m.input = ""
					m.cursor = 0
					// Build context from answered questions
					answeredKeys := make(map[string]string)
					for k, v := range m.engine.StrInputs() {
						answeredKeys[k] = v
					}
					for k, v := range m.engine.Inputs() {
						if _, exists := answeredKeys[k]; !exists {
							answeredKeys[k] = fmt.Sprintf("%v", v)
						}
					}
					filingStatus := ""
					if fs, ok := m.engine.StrInputs()[forms.F1040FilingStatus]; ok {
						filingStatus = fs
					}
					return m, func() tea.Msg {
						return tui.RequestAIPromptMsg{
							UserQuestion: userQuestion,
							FieldKey:     q.Key,
							Label:        q.Prompt,
							FormName:     q.FormName,
							FilingStatus: filingStatus,
							AnsweredKeys: answeredKeys,
						}
					}
				}
				m.input = ""
				m.cursor = 0
				return m, nil
			}
			// Handle "??" RAG-powered explanation command
			if m.input == "??" {
				q := m.engine.Current()
				if q != nil {
					m.history = append(m.history, m.input)
					m.historyIdx = -1
					m.aiLoading = true
					m.aiHelpText = ""
					m.input = ""
					m.cursor = 0
					return m, func() tea.Msg {
						return tui.RequestExplanationMsg{
							FieldKey: q.Key,
							Label:    q.Prompt,
							FormName: q.FormName,
						}
					}
				}
				m.input = ""
				m.cursor = 0
				return m, nil
			}
			// Handle "why" command — explain why this question is being asked
			if m.input == "why" {
				q := m.engine.Current()
				if q != nil {
					m.history = append(m.history, m.input)
					m.historyIdx = -1
					m.aiLoading = true
					m.aiHelpText = ""
					m.input = ""
					m.cursor = 0
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
					if fs, ok := m.engine.StrInputs()[forms.F1040FilingStatus]; ok {
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
				m.cursor = 0
				return m, nil
			}
			// Handle "ca" command — explain CA vs federal difference
			if m.input == "ca" && m.stateCode == forms.StateCodeCA {
				q := m.engine.Current()
				if q != nil {
					m.history = append(m.history, m.input)
					m.historyIdx = -1
					m.aiLoading = true
					m.aiHelpText = ""
					m.input = ""
					m.cursor = 0
					return m, func() tea.Msg {
						return tui.RequestCADiffMsg{
							FieldKey: q.Key,
							Label:    q.Prompt,
						}
					}
				}
				m.input = ""
				m.cursor = 0
				return m, nil
			}
			// Handle "skip" command — skip all questions for the current form
			if m.input == "skip" {
				m.input = ""
				m.cursor = 0
				m.helpText = ""
				m.aiHelpText = ""
				skipped := m.engine.SkipForm()
				if skipped > 0 {
					m.err = ""
					if !m.engine.HasNext() {
						m.done = true
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
				}
				return m, nil
			}
			// Handle "prior" command — show prior-year value for current question
			if m.input == "prior" {
				m.input = ""
				m.cursor = 0
				m.helpText = ""
				m.aiHelpText = ""
				numCount, strCount := m.engine.PriorYearCount()
				pyd := m.engine.GetPriorYearDefault()
				if pyd != nil {
					display := fmt.Sprintf("Prior year: %s", pyd.PriorValue)
					if pyd.CANote != "" {
						display += "\nCA: " + pyd.CANote
					}
					m.helpText = display
				} else if numCount > 0 || strCount > 0 {
					q := m.engine.Current()
					fieldKey := ""
					if q != nil {
						fieldKey = q.Key
					}
					m.helpText = fmt.Sprintf("No prior-year value for field %q. (%d numeric, %d string values loaded from prior year)", fieldKey, numCount, strCount)
				} else {
					m.helpText = "No prior-year data loaded. Use --import prior.pdf to load a prior return, or complete a return with TaxPilot to auto-save for next year."
				}
				return m, nil
			}
			// Handle "calc" command — enter calculator mode
			if m.input == "calc" {
				m.input = ""
				m.cursor = 0
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
					if cp.CANote != "" && m.stateCode == forms.StateCodeCA {
						m.helpText += "\n\nCalifornia note: " + cp.CANote
					}
					if m.helpText == "" {
						m.helpText = "No additional help available for this question."
					}
				}
				m.input = ""
				m.cursor = 0
				return m, nil
			}
			// Save non-empty input to history for Up-arrow recall
			if m.input != "" {
				m.history = append(m.history, m.input)
			}
			m.historyIdx = -1
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
			m.cursor = 0
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
			if m.cursor > 0 && len(m.input) > 0 {
				m.input = m.input[:m.cursor-1] + m.input[m.cursor:]
				m.cursor--
				m.err = ""
			}
			return m, nil

		case tea.KeyLeft:
			if m.input == "" {
				// Go back to previous question when input is empty
				if m.engine.Back() {
					m.err = ""
					m.helpText = ""
					m.aiHelpText = ""
				}
			} else if m.cursor > 0 {
				// Move cursor left within text
				m.cursor--
			}
			return m, nil

		case tea.KeyRight:
			if m.input == "" {
				// Go forward to next question when input is empty
				if m.engine.Forward() {
					m.err = ""
					m.helpText = ""
					m.aiHelpText = ""
				}
			} else if m.cursor < len(m.input) {
				// Move cursor right within text
				m.cursor++
			}
			return m, nil

		case tea.KeyUp:
			// Recall previous input from history
			if len(m.history) > 0 {
				if m.historyIdx == -1 {
					// Entering history mode — save current input
					m.historyBuf = m.input
					m.historyIdx = len(m.history) - 1
				} else if m.historyIdx > 0 {
					m.historyIdx--
				}
				m.input = m.history[m.historyIdx]
				m.cursor = len(m.input)
			}
			return m, nil

		case tea.KeyDown:
			// Navigate forward in history
			if m.historyIdx >= 0 {
				if m.historyIdx < len(m.history)-1 {
					m.historyIdx++
					m.input = m.history[m.historyIdx]
				} else {
					// Past the end — restore the original input
					m.historyIdx = -1
					m.input = m.historyBuf
				}
				m.cursor = len(m.input)
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
			m.input = m.input[:m.cursor] + key + m.input[m.cursor:]
			m.cursor += len(key)
			m.err = ""
			return m, nil

		case tea.KeySpace:
			m.input = m.input[:m.cursor] + " " + m.input[m.cursor:]
			m.cursor++
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
			m.cursor = len(m.input)
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
	if fs, ok := ret.StrInputs[forms.F1040FilingStatus]; ok {
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
	prompt := tui.PromptStyle.Render(cp.Prompt)

	// Contextual help text below the prompt
	var contextHelp string
	if cp.HelpText != "" {
		contextHelp = tui.HelpStyle.Render(cp.HelpText)
	}

	// CA-specific note
	var caNote string
	if cp.CANote != "" && m.stateCode == forms.StateCodeCA {
		caNote = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5C07B")).
			Italic(true).
			Render("CA: " + cp.CANote)
	}

	// Current answer (shown when navigating back to an answered question)
	var currentAnswer string
	if ans := m.engine.CurrentAnswer(); ans != "" {
		currentAnswer = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#98C379")).
			Bold(true).
			Render("Current answer: " + ans)
	}

	// User-triggered help text (from "?" command)
	var userHelp string
	if m.helpText != "" {
		userHelp = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#98C379")).
			Italic(true).
			Render(m.helpText)
	}

	// AI-powered explanation (from "??" command or streaming)
	var aiHelp string
	if m.aiLoading && m.aiHelpText != "" {
		// Streaming in progress — show accumulated text with a blinking cursor
		aiHelp = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#56B6C2")).
			Italic(true).
			Render(m.aiHelpText + "▍")
	} else if m.aiLoading {
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
		if pyd.CANote != "" {
			priorYearBlock += "\n" + lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E5C07B")).
				Italic(true).
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

	// Input area with cursor at the correct position
	cursorPrefix := tui.HighlightStyle.Render("▸ ")
	beforeCursor := tui.InputStyle.Render(m.input[:m.cursor])
	cursorChar := tui.HighlightStyle.Render("█")
	afterCursor := tui.InputStyle.Render(m.input[m.cursor:])
	inputLine := cursorPrefix + beforeCursor + cursorChar + afterCursor

	// Error message
	var errBlock string
	if m.err != "" {
		errBlock = tui.ErrorStyle.Render("⚠ " + m.err)
	}

	// Help table — aligned columns of key/action pairs
	type helpEntry struct {
		key, action string
	}
	helpEntries := []helpEntry{
		{"Enter", "submit"},
		{"←/→", "navigate/cursor"},
		{"↑↓", "history"},
		{"skip", "skip form"},
		{"?", "help"},
		{"??", "AI explain"},
		{"why", "why asked"},
		{"ai <q>", "ask AI"},
		{"calc", "calculator"},
		{"prior", "last year"},
		{"q", "save & quit"},
	}
	if m.stateCode == forms.StateCodeCA {
		helpEntries = append(helpEntries, helpEntry{"ca", "CA diff"})
	}

	// Find the max key and action widths for alignment
	maxKeyW := 0
	maxActW := 0
	for _, e := range helpEntries {
		if w := lipgloss.Width(e.key); w > maxKeyW {
			maxKeyW = w
		}
		if w := lipgloss.Width(e.action); w > maxActW {
			maxActW = w
		}
	}

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#61AFEF")).
		Bold(true).
		Width(maxKeyW).
		Align(lipgloss.Right)
	actionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Italic(true).
		Width(maxActW)
	sepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#444444"))

	// Each cell: key + " " + action + gap between columns
	cellWidth := maxKeyW + 1 + maxActW + 3 // key + space + action + " | "
	cols := contentW / cellWidth
	if cols < 2 {
		cols = 2
	}
	if cols > len(helpEntries) {
		cols = len(helpEntries)
	}

	var helpLines []string
	for i := 0; i < len(helpEntries); i += cols {
		var cells []string
		for j := 0; j < cols && i+j < len(helpEntries); j++ {
			e := helpEntries[i+j]
			cells = append(cells, keyStyle.Render(e.key)+" "+actionStyle.Render(e.action))
		}
		helpLines = append(helpLines, strings.Join(cells, sepStyle.Render(" · ")))
	}
	help := strings.Join(helpLines, "\n")

	// Compose layout — use single blank lines for separation
	parts := []string{
		progress,
		bar,
		"",
		formContext,
		prompt,
	}
	if contextHelp != "" {
		parts = append(parts, contextHelp)
	}
	if caNote != "" {
		parts = append(parts, caNote)
	}
	if currentAnswer != "" {
		parts = append(parts, currentAnswer)
	}
	if priorYearBlock != "" {
		parts = append(parts, priorYearBlock)
	}
	if optionsBlock != "" {
		parts = append(parts, optionsBlock)
	}
	if userHelp != "" {
		parts = append(parts, userHelp)
	}
	if aiHelp != "" {
		parts = append(parts, aiHelp)
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
	case forms.FilingSingle:
		return "Single"
	case forms.FilingMFJ:
		return "Married Filing Jointly"
	case forms.FilingMFS:
		return "Married Filing Separately"
	case forms.FilingHOH:
		return "Head of Household"
	case forms.FilingQSS:
		return "Qualifying Surviving Spouse"
	default:
		return opt
	}
}
