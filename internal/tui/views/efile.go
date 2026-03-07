package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"taxpilot/internal/efile"
	"taxpilot/internal/tui"
)

// EFileStep tracks the current step in the e-file flow.
type EFileStep int

const (
	StepReview     EFileStep = iota // review return summary
	StepValidation                  // show validation results
	StepPIN                         // enter self-select PIN
	StepConfirm                     // final confirmation before submit
	StepSubmitting                  // submission in progress
	StepResult                      // show submission result
)

// EFileSubmitMsg triggers the actual e-file submission.
type EFileSubmitMsg struct {
	FederalXML []byte
	CAXML      []byte
	Auth       *efile.EFileAuth
}

// EFileResultMsg carries the submission result back to the view.
type EFileResultMsg struct {
	FederalResult *EFileSubmissionResult
	CAResult      *EFileSubmissionResult
	Err           error
}

// EFileSubmissionResult holds per-jurisdiction submission result.
type EFileSubmissionResult struct {
	SubmissionID string
	Status       string
	Message      string
}

// EFileView is the Bubble Tea model for the e-file submission flow.
type EFileView struct {
	results    map[string]float64
	strResults map[string]string
	taxYear    int
	state      string
	width      int
	height     int

	step       EFileStep
	validation efile.ValidationReport
	auth       *efile.EFileAuth
	pinInput   string
	pinField   string // "federal" or "ca"
	federalXML []byte
	caXML      []byte

	federalOnly bool
	stateOnly   bool

	submitResult *EFileResultMsg
	errMsg       string
}

// NewEFileView creates a new e-file view.
func NewEFileView(
	results map[string]float64,
	strResults map[string]string,
	taxYear int,
	stateCode string,
	federalOnly bool,
	stateOnly bool,
) EFileView {
	return EFileView{
		results:     results,
		strResults:  strResults,
		taxYear:     taxYear,
		state:       stateCode,
		step:        StepReview,
		federalOnly: federalOnly,
		stateOnly:   stateOnly,
	}
}

// Init satisfies tea.Model.
func (m EFileView) Init() tea.Cmd {
	return nil
}

// Update satisfies tea.Model.
func (m EFileView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case EFileResultMsg:
		m.submitResult = &msg
		m.step = StepResult
		return m, nil

	case tea.KeyMsg:
		switch m.step {
		case StepReview:
			return m.handleReview(msg)
		case StepValidation:
			return m.handleValidation(msg)
		case StepPIN:
			return m.handlePIN(msg)
		case StepConfirm:
			return m.handleConfirm(msg)
		case StepResult:
			return m.handleResult(msg)
		}
	}
	return m, nil
}

func (m EFileView) handleReview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "v", "enter":
		// Run validation
		includeCA := m.state == "CA" && !m.federalOnly
		report := efile.ValidateFull(m.results, m.strResults, m.taxYear, includeCA)
		m.validation = report
		m.step = StepValidation
		return m, nil
	case "b":
		// Go back to summary
		return m, func() tea.Msg {
			return tui.ShowSummaryMsg{
				Results:   m.results,
				StrInputs: m.strResults,
				TaxYear:   m.taxYear,
				State:     m.state,
			}
		}
	}
	return m, nil
}

func (m EFileView) handleValidation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "b":
		m.step = StepReview
		return m, nil
	case "enter":
		if !m.validation.IsValid {
			// Can't proceed with errors
			return m, nil
		}
		// Create auth and move to PIN entry
		m.auth = efile.NewEFileAuth(m.results, m.strResults, m.taxYear)
		m.pinInput = ""
		if m.stateOnly {
			m.pinField = "ca"
		} else {
			m.pinField = "federal"
		}
		m.step = StepPIN
		return m, nil
	}
	return m, nil
}

func (m EFileView) handlePIN(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		if len(m.pinInput) != 5 {
			m.errMsg = "PIN must be exactly 5 digits"
			return m, nil
		}
		m.errMsg = ""
		if m.pinField == "federal" {
			if err := m.auth.SetFederalPIN(m.pinInput); err != nil {
				m.errMsg = err.Error()
				return m, nil
			}
			// If also filing CA, get CA PIN next
			if m.state == "CA" && !m.federalOnly {
				m.pinInput = ""
				m.pinField = "ca"
				return m, nil
			}
		} else {
			if err := m.auth.SetCAPIN(m.pinInput); err != nil {
				m.errMsg = err.Error()
				return m, nil
			}
		}
		m.pinInput = ""
		m.step = StepConfirm
		return m, nil
	case tea.KeyBackspace, tea.KeyDelete:
		if len(m.pinInput) > 0 {
			m.pinInput = m.pinInput[:len(m.pinInput)-1]
		}
		m.errMsg = ""
		return m, nil
	case tea.KeyEsc:
		m.step = StepReview
		m.pinInput = ""
		m.errMsg = ""
		return m, nil
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyRunes:
		ch := msg.String()
		if len(ch) == 1 && ch[0] >= '0' && ch[0] <= '9' && len(m.pinInput) < 5 {
			m.pinInput += ch
			m.errMsg = ""
		}
		return m, nil
	}
	return m, nil
}

func (m EFileView) handleConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "y", "Y":
		m.step = StepSubmitting
		// Trigger submission
		return m, func() tea.Msg {
			return EFileSubmitMsg{
				FederalXML: m.federalXML,
				CAXML:      m.caXML,
				Auth:       m.auth,
			}
		}
	case "n", "N", "b":
		m.step = StepReview
		return m, nil
	}
	return m, nil
}

func (m EFileView) handleResult(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c", "enter":
		return m, tea.Quit
	}
	return m, nil
}

// View satisfies tea.Model.
func (m EFileView) View() string {
	switch m.step {
	case StepReview:
		return m.viewReview()
	case StepValidation:
		return m.viewValidation()
	case StepPIN:
		return m.viewPIN()
	case StepConfirm:
		return m.viewConfirm()
	case StepSubmitting:
		return m.viewSubmitting()
	case StepResult:
		return m.viewResult()
	}
	return ""
}

func (m EFileView) viewReview() string {
	var sections []string

	sections = append(sections, tui.TitleStyle.Render(
		fmt.Sprintf("E-File Review \u2014 Tax Year %d", m.taxYear),
	))

	// Filing scope
	scope := "Federal + California"
	if m.federalOnly {
		scope = "Federal Only"
	} else if m.stateOnly {
		scope = "California Only"
	}
	sections = append(sections, formatLine("Filing Scope", scope))
	sections = append(sections, "")

	// Federal summary
	if !m.stateOnly {
		sections = append(sections, tui.HighlightStyle.Render(
			"\u2550\u2550\u2550 Federal Return \u2550\u2550\u2550",
		))
		sections = append(sections, formatLine("Taxpayer",
			m.strResults["1040:first_name"]+" "+m.strResults["1040:last_name"]))
		sections = append(sections, formatMoney("AGI", m.results["1040:11"]))
		sections = append(sections, formatMoney("Total Tax", m.results["1040:24"]))
		sections = append(sections, formatMoney("Withholding", m.results["1040:25d"]))
		refund := m.results["1040:34"]
		owed := m.results["1040:37"]
		if refund > 0 {
			sections = append(sections,
				tui.SuccessStyle.Render(fmt.Sprintf("%-25s %s", "REFUND:", formatDollar(refund))))
		} else if owed > 0 {
			sections = append(sections,
				tui.ErrorStyle.Render(fmt.Sprintf("%-25s %s", "AMOUNT OWED:", formatDollar(owed))))
		}
		sections = append(sections, "")
	}

	// CA summary
	if m.state == "CA" && !m.federalOnly {
		sections = append(sections, tui.HighlightStyle.Render(
			"\u2550\u2550\u2550 California Return \u2550\u2550\u2550",
		))
		sections = append(sections, formatMoney("CA AGI", m.results["ca_540:17"]))
		sections = append(sections, formatMoney("CA Total Tax", m.results["ca_540:40"]))
		sections = append(sections, formatMoney("CA Withholding", m.results["ca_540:71"]))
		caRefund := m.results["ca_540:91"]
		caOwed := m.results["ca_540:93"]
		if caRefund > 0 {
			sections = append(sections,
				tui.SuccessStyle.Render(fmt.Sprintf("%-25s %s", "CA REFUND:", formatDollar(caRefund))))
		} else if caOwed > 0 {
			sections = append(sections,
				tui.ErrorStyle.Render(fmt.Sprintf("%-25s %s", "CA AMOUNT OWED:", formatDollar(caOwed))))
		}
		sections = append(sections, "")
	}

	sections = append(sections,
		tui.HelpStyle.Render("v/Enter: validate & continue  |  b: back  |  q: quit"))

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return tui.BorderStyle.Render(content) + "\n"
}

func (m EFileView) viewValidation() string {
	var sections []string

	sections = append(sections, tui.TitleStyle.Render("Pre-Submission Validation"))

	if len(m.validation.Results) == 0 {
		sections = append(sections,
			tui.SuccessStyle.Render("\u2713 All validation checks passed!"))
	} else {
		for _, r := range m.validation.Results {
			var icon, style string
			switch r.Severity {
			case efile.SeverityError:
				icon = "\u2717"
				style = tui.ErrorStyle.Render(fmt.Sprintf("  %s [%s] %s", icon, r.Code, r.Message))
			case efile.SeverityWarning:
				icon = "\u26A0"
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("#E5C07B")).
					Render(fmt.Sprintf("  %s [%s] %s", icon, r.Code, r.Message))
			case efile.SeverityInfo:
				icon = "\u2139"
				style = tui.HelpStyle.Render(fmt.Sprintf("  %s [%s] %s", icon, r.Code, r.Message))
			}
			sections = append(sections, style)
		}
	}

	sections = append(sections, "")

	if m.validation.IsValid {
		sections = append(sections,
			tui.SuccessStyle.Render("\u2713 Return is valid for e-filing"))
		sections = append(sections, "")
		sections = append(sections,
			tui.HelpStyle.Render("Enter: continue to PIN  |  b: back  |  q: quit"))
	} else {
		errorCount := 0
		for _, r := range m.validation.Results {
			if r.Severity == efile.SeverityError {
				errorCount++
			}
		}
		sections = append(sections,
			tui.ErrorStyle.Render(fmt.Sprintf(
				"\u2717 %d error(s) must be fixed before e-filing", errorCount)))
		sections = append(sections, "")
		sections = append(sections,
			tui.HelpStyle.Render("b: back to fix errors  |  q: quit"))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return tui.BorderStyle.Render(content) + "\n"
}

func (m EFileView) viewPIN() string {
	var sections []string

	jurisdiction := "Federal"
	if m.pinField == "ca" {
		jurisdiction = "California"
	}

	sections = append(sections, tui.TitleStyle.Render(
		fmt.Sprintf("E-File Signature \u2014 %s", jurisdiction)))

	sections = append(sections, tui.PromptStyle.Render(
		"Enter your 5-digit self-select PIN:"))
	sections = append(sections, tui.HelpStyle.Render(
		"This PIN serves as your electronic signature (Form 8879)."))
	sections = append(sections, "")

	// PIN display with masked digits
	display := strings.Repeat("\u2022", len(m.pinInput))
	remaining := 5 - len(m.pinInput)
	display += strings.Repeat("_", remaining)

	sections = append(sections, tui.HighlightStyle.Render(
		"  PIN: [ "+display+" ]"))

	if m.errMsg != "" {
		sections = append(sections, "")
		sections = append(sections, tui.ErrorStyle.Render("  "+m.errMsg))
	}

	sections = append(sections, "")
	sections = append(sections,
		tui.HelpStyle.Render("Enter: submit  |  Esc: cancel  |  Digits only"))

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return tui.BorderStyle.Render(content) + "\n"
}

func (m EFileView) viewConfirm() string {
	var sections []string

	sections = append(sections, tui.TitleStyle.Render("Confirm E-File Submission"))
	sections = append(sections, "")
	sections = append(sections, tui.PromptStyle.Render(
		"You are about to electronically file your tax return(s)."))
	sections = append(sections, "")

	if !m.stateOnly {
		sections = append(sections, fmt.Sprintf("  Federal: %s (PIN set)",
			tui.SuccessStyle.Render("\u2713")))
	}
	if m.state == "CA" && !m.federalOnly {
		sections = append(sections, fmt.Sprintf("  California: %s (PIN set)",
			tui.SuccessStyle.Render("\u2713")))
	}

	sections = append(sections, "")
	sections = append(sections,
		tui.ErrorStyle.Render("  This action cannot be undone."))
	sections = append(sections, "")
	sections = append(sections,
		tui.PromptStyle.Render("  Proceed with e-filing? (Y/N)"))
	sections = append(sections, "")
	sections = append(sections,
		tui.HelpStyle.Render("y: submit  |  n: go back  |  q: quit"))

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return tui.BorderStyle.Render(content) + "\n"
}

func (m EFileView) viewSubmitting() string {
	var sections []string

	sections = append(sections, tui.TitleStyle.Render("Submitting..."))
	sections = append(sections, "")
	sections = append(sections, tui.HighlightStyle.Render(
		"  Transmitting return to IRS and CA FTB..."))
	sections = append(sections, "")
	sections = append(sections, tui.HelpStyle.Render(
		"  Please wait. Do not close this window."))

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return tui.BorderStyle.Render(content) + "\n"
}

func (m EFileView) viewResult() string {
	var sections []string

	sections = append(sections, tui.TitleStyle.Render("E-File Results"))
	sections = append(sections, "")

	if m.submitResult == nil {
		sections = append(sections, tui.ErrorStyle.Render("No results available"))
	} else if m.submitResult.Err != nil {
		sections = append(sections, tui.ErrorStyle.Render(
			"Submission error: "+m.submitResult.Err.Error()))
	} else {
		if m.submitResult.FederalResult != nil {
			r := m.submitResult.FederalResult
			icon := "\u2713"
			style := tui.SuccessStyle
			if r.Status == "Rejected" || r.Status == "Error" {
				icon = "\u2717"
				style = tui.ErrorStyle
			} else if r.Status == "Pending" {
				icon = "\u231B"
				style = tui.HighlightStyle
			}
			sections = append(sections, style.Render(
				fmt.Sprintf("  %s Federal: %s", icon, r.Status)))
			sections = append(sections, tui.HelpStyle.Render(
				fmt.Sprintf("    ID: %s", r.SubmissionID)))
			sections = append(sections, tui.HelpStyle.Render(
				fmt.Sprintf("    %s", r.Message)))
			sections = append(sections, "")
		}

		if m.submitResult.CAResult != nil {
			r := m.submitResult.CAResult
			icon := "\u2713"
			style := tui.SuccessStyle
			if r.Status == "Rejected" || r.Status == "Error" {
				icon = "\u2717"
				style = tui.ErrorStyle
			} else if r.Status == "Pending" {
				icon = "\u231B"
				style = tui.HighlightStyle
			}
			sections = append(sections, style.Render(
				fmt.Sprintf("  %s California: %s", icon, r.Status)))
			sections = append(sections, tui.HelpStyle.Render(
				fmt.Sprintf("    ID: %s", r.SubmissionID)))
			sections = append(sections, tui.HelpStyle.Render(
				fmt.Sprintf("    %s", r.Message)))
			sections = append(sections, "")
		}
	}

	sections = append(sections,
		tui.HelpStyle.Render("Enter/q: quit"))

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return tui.BorderStyle.Render(content) + "\n"
}
