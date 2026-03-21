package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"taxpilot/internal/forms"
	"taxpilot/internal/tui"
)

// SummaryView displays the computed tax results.
type SummaryView struct {
	results    map[string]float64
	strResults map[string]string
	taxYear    int
	state      string
	width      int
	height     int
	exportMsg  string // status message after export
}

// NewSummaryView creates a SummaryView with the given results.
func NewSummaryView(results map[string]float64, strResults map[string]string, taxYear int, stateCode string) SummaryView {
	return SummaryView{
		results:    results,
		strResults: strResults,
		taxYear:    taxYear,
		state:      stateCode,
	}
}

// Init satisfies tea.Model.
func (m SummaryView) Init() tea.Cmd {
	return nil
}

// Update satisfies tea.Model.
func (m SummaryView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tui.ExportPDFResultMsg:
		if msg.Err != nil {
			m.exportMsg = fmt.Sprintf("Export error: %v", msg.Err)
		} else {
			m.exportMsg = fmt.Sprintf("Exported %d file(s): %s", len(msg.Paths), strings.Join(msg.Paths, ", "))
		}
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "e":
			return m, func() tea.Msg {
				return tui.ExportPDFMsg{
					Results:   m.results,
					StrInputs: m.strResults,
					TaxYear:   m.taxYear,
				}
			}
		case "f":
			// Start e-file flow
			return m, func() tea.Msg {
				return tui.StartEFileMsg{
					Results:   m.results,
					StrInputs: m.strResults,
					TaxYear:   m.taxYear,
					State:     m.state,
				}
			}
		case "r":
			// Open detailed review
			return m, func() tea.Msg {
				return tui.ShowReviewMsg{
					Results:   m.results,
					StrInputs: m.strResults,
					TaxYear:   m.taxYear,
					State:     m.state,
				}
			}
		case "b":
			return m, func() tea.Msg {
				return tui.StartInterviewMsg{
					TaxYear:   m.taxYear,
					StateCode: m.state,
					Continue:  true,
				}
			}
		}
	}
	return m, nil
}

// View satisfies tea.Model.
func (m SummaryView) View() string {
	var sections []string

	// Header
	header := tui.TitleStyle.Render(fmt.Sprintf(
		"Tax Return Summary — Tax Year %d", m.taxYear,
	))
	sections = append(sections, header)

	// Taxpayer info
	firstName := m.strResults[forms.F1040FirstName]
	lastName := m.strResults[forms.F1040LastName]
	filingStatus := formatOptionLabel(m.strResults[forms.F1040FilingStatus])
	if firstName != "" || lastName != "" {
		sections = append(sections,
			tui.PromptStyle.Render(fmt.Sprintf("Taxpayer: %s %s", firstName, lastName)),
		)
	}

	// Federal section
	sections = append(sections, "")
	sections = append(sections, tui.HighlightStyle.Render(
		"═══ Federal Return (Form 1040) ═══",
	))
	sections = append(sections, formatLine("Filing Status", filingStatus))
	sections = append(sections, formatMoney("Total Income", m.results["1040:9"]))
	sections = append(sections, formatMoney("Adjusted Gross Income", m.results["1040:11"]))
	sections = append(sections, formatMoney("Standard Deduction", m.results["1040:12"]))
	sections = append(sections, formatMoney("Taxable Income", m.results["1040:15"]))
	sections = append(sections, formatMoney("Federal Tax", m.results["1040:16"]))
	sections = append(sections, formatMoney("Total Withholding", m.results["1040:25d"]))

	refund := m.results["1040:34"]
	owed := m.results["1040:37"]
	if refund > 0 {
		sections = append(sections,
			tui.SuccessStyle.Render(fmt.Sprintf("%-25s %s", "REFUND:", formatDollar(refund))),
		)
	} else if owed > 0 {
		sections = append(sections,
			tui.ErrorStyle.Render(fmt.Sprintf("%-25s %s", "AMOUNT OWED:", formatDollar(owed))),
		)
	} else {
		sections = append(sections, formatMoney("Balance", 0))
	}

	// State section (CA)
	if m.state == forms.StateCodeCA {
		sections = append(sections, "")
		sections = append(sections, tui.HighlightStyle.Render(
			"═══ California Return (Form 540) ═══",
		))
		sections = append(sections, formatMoney("CA Adjusted Gross Income", m.results["ca_540:17"]))
		sections = append(sections, formatMoney("CA Standard Deduction", m.results["ca_540:18"]))
		sections = append(sections, formatMoney("CA Taxable Income", m.results["ca_540:19"]))
		sections = append(sections, formatMoney("CA Tax", m.results["ca_540:31"]))
		sections = append(sections, formatMoney("Exemption Credits", m.results["ca_540:32"]))
		sections = append(sections, formatMoney("Mental Health Tax", m.results["ca_540:36"]))
		sections = append(sections, formatMoney("Total CA Tax", m.results["ca_540:40"]))
		sections = append(sections, formatMoney("CA Withholding", m.results["ca_540:71"]))

		caRefund := m.results["ca_540:91"]
		caOwed := m.results["ca_540:93"]
		if caRefund > 0 {
			sections = append(sections,
				tui.SuccessStyle.Render(fmt.Sprintf("%-25s %s", "CA REFUND:", formatDollar(caRefund))),
			)
		} else if caOwed > 0 {
			sections = append(sections,
				tui.ErrorStyle.Render(fmt.Sprintf("%-25s %s", "CA AMOUNT OWED:", formatDollar(caOwed))),
			)
		} else {
			sections = append(sections, formatMoney("CA Balance", 0))
		}
	}

	// Export status
	if m.exportMsg != "" {
		sections = append(sections, "")
		sections = append(sections, tui.HighlightStyle.Render(m.exportMsg))
	}

	// Footer
	sections = append(sections, "")
	sections = append(sections, tui.HelpStyle.Render(
		"r: review  |  e: export PDFs  |  f: e-file  |  q: quit  |  b: go back",
	))

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	contentW := tui.ContentWidth(m.width)
	return tui.BorderStyle.Width(contentW).Render(content) + "\n"
}

// formatMoney formats a label and dollar amount on a single line.
func formatMoney(label string, amount float64) string {
	return fmt.Sprintf("%-25s %s", label+":", formatDollar(amount))
}

// formatDollar formats a float64 as a dollar amount with commas.
func formatDollar(amount float64) string {
	negative := amount < 0
	if negative {
		amount = -amount
	}

	whole := int64(amount)
	cents := int64((amount - float64(whole)) * 100 + 0.5)
	if cents >= 100 {
		whole++
		cents -= 100
	}

	// Format with commas
	s := fmt.Sprintf("%d", whole)
	if len(s) > 3 {
		var parts []string
		for len(s) > 3 {
			parts = append([]string{s[len(s)-3:]}, parts...)
			s = s[:len(s)-3]
		}
		parts = append([]string{s}, parts...)
		s = strings.Join(parts, ",")
	}

	result := fmt.Sprintf("$%s.%02d", s, cents)
	if negative {
		result = "-" + result
	}
	return result
}

// formatLine formats a label and string value on a single line.
func formatLine(label, value string) string {
	return fmt.Sprintf("%-25s %s", label+":", value)
}
