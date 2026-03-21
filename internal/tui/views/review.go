package views

import (
	"fmt"
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"taxpilot/internal/efile"
	"taxpilot/internal/tui"
)

// WarningStyle is used for yellow warning text.
var WarningStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#E5C07B")).
	Bold(true)

// InfoStyle is used for blue informational text.
var InfoStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#61AFEF"))

const (
	tabOverview    = 0
	tabFederal     = 1
	tabCA          = 2
	tabPriorYear   = 3
	tabValidation  = 4
	numTabs        = 5
)

// ReviewView displays a detailed multi-tab review of the tax return.
type ReviewView struct {
	results      map[string]float64
	strResults   map[string]string
	priorResults map[string]float64
	taxYear      int
	state        string
	validation   efile.ValidationReport
	tab          int
	width        int
	height       int
	scrollOffset int
	exportMsg    string
}

// NewReviewView creates a ReviewView with the given data.
func NewReviewView(msg tui.ShowReviewMsg) ReviewView {
	includeCA := msg.State == "CA"
	validation := efile.ValidateFull(msg.Results, msg.StrInputs, msg.TaxYear, includeCA)

	return ReviewView{
		results:      msg.Results,
		strResults:   msg.StrInputs,
		priorResults: msg.PriorResults,
		taxYear:      msg.TaxYear,
		state:        msg.State,
		validation:   validation,
		tab:          tabOverview,
	}
}

// Init satisfies tea.Model.
func (m ReviewView) Init() tea.Cmd {
	return nil
}

// Update satisfies tea.Model.
func (m ReviewView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tui.ExportPDFResultMsg:
		if msg.Err != nil {
			m.exportMsg = fmt.Sprintf("Export error: %v", msg.Err)
		} else {
			m.exportMsg = fmt.Sprintf("Exported %d file(s)", len(msg.Paths))
		}
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.tab = (m.tab + 1) % numTabs
			m.scrollOffset = 0
			return m, nil
		case "1":
			m.tab = tabOverview
			m.scrollOffset = 0
			return m, nil
		case "2":
			m.tab = tabFederal
			m.scrollOffset = 0
			return m, nil
		case "3":
			m.tab = tabCA
			m.scrollOffset = 0
			return m, nil
		case "4":
			m.tab = tabPriorYear
			m.scrollOffset = 0
			return m, nil
		case "5":
			m.tab = tabValidation
			m.scrollOffset = 0
			return m, nil
		case "j", "down":
			m.scrollOffset++
			return m, nil
		case "k", "up":
			if m.scrollOffset > 0 {
				m.scrollOffset--
			}
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
		}
	}
	return m, nil
}

// View satisfies tea.Model.
func (m ReviewView) View() string {
	var sections []string

	// Header
	header := tui.TitleStyle.Render(fmt.Sprintf(
		"Detailed Review — Tax Year %d", m.taxYear,
	))
	sections = append(sections, header)

	// Tab bar
	sections = append(sections, m.renderTabBar())
	sections = append(sections, "")

	// Tab content
	var content []string
	switch m.tab {
	case tabOverview:
		content = m.renderOverview()
	case tabFederal:
		content = m.renderFederalDetail()
	case tabCA:
		content = m.renderCADetail()
	case tabPriorYear:
		content = m.renderPriorYearComparison()
	case tabValidation:
		content = m.renderValidation()
	}

	// Apply scroll
	if m.scrollOffset > 0 && m.scrollOffset < len(content) {
		content = content[m.scrollOffset:]
	} else if m.scrollOffset >= len(content) && len(content) > 0 {
		m.scrollOffset = len(content) - 1
		content = content[m.scrollOffset:]
	}

	// Limit visible lines based on height
	maxLines := m.height - 8 // reserve space for header, tab bar, footer
	if maxLines < 5 {
		maxLines = 20
	}
	if len(content) > maxLines {
		content = content[:maxLines]
	}

	sections = append(sections, content...)

	// Export status
	if m.exportMsg != "" {
		sections = append(sections, "")
		sections = append(sections, tui.HighlightStyle.Render(m.exportMsg))
	}

	// Footer
	sections = append(sections, "")
	sections = append(sections, tui.HelpStyle.Render(
		"1-5/tab: switch tabs  |  j/k: scroll  |  b: back  |  e: export  |  f: e-file  |  q: quit",
	))

	body := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return tui.BorderStyle.Render(body) + "\n"
}

func (m ReviewView) renderTabBar() string {
	tabs := []string{"Overview", "Federal", "CA", "Prior Year", "Validation"}
	var parts []string
	for i, name := range tabs {
		label := fmt.Sprintf(" %d:%s ", i+1, name)
		if i == m.tab {
			parts = append(parts, tui.HighlightStyle.Render("["+label+"]"))
		} else {
			parts = append(parts, tui.HelpStyle.Render(" "+label+" "))
		}
	}
	return strings.Join(parts, "")
}

// --- Tab 0: Overview ---

func (m ReviewView) renderOverview() []string {
	var lines []string

	// Taxpayer info
	firstName := m.strResults["1040:first_name"]
	lastName := m.strResults["1040:last_name"]
	filingStatus := formatOptionLabel(m.strResults["1040:filing_status"])
	if firstName != "" || lastName != "" {
		lines = append(lines, tui.PromptStyle.Render(fmt.Sprintf("Taxpayer: %s %s", firstName, lastName)))
	}
	lines = append(lines, formatLine("Filing Status", filingStatus))
	lines = append(lines, "")

	// Federal summary
	lines = append(lines, tui.HighlightStyle.Render("--- Federal (Form 1040) ---"))
	lines = append(lines, formatMoney("AGI", m.results["1040:11"]))
	lines = append(lines, formatMoney("Taxable Income", m.results["1040:15"]))
	lines = append(lines, formatMoney("Total Tax", m.results["1040:24"]))
	lines = append(lines, formatMoney("Withholding", m.results["1040:25d"]))

	fedRefund := m.results["1040:34"]
	fedOwed := m.results["1040:37"]
	if fedRefund > 0 {
		lines = append(lines, tui.SuccessStyle.Render(fmt.Sprintf("%-25s %s", "REFUND:", formatDollar(fedRefund))))
	} else if fedOwed > 0 {
		lines = append(lines, tui.ErrorStyle.Render(fmt.Sprintf("%-25s %s", "AMOUNT OWED:", formatDollar(fedOwed))))
	} else {
		lines = append(lines, formatMoney("Balance", 0))
	}

	// CA summary
	if m.state == "CA" {
		lines = append(lines, "")
		lines = append(lines, tui.HighlightStyle.Render("--- California (Form 540) ---"))
		lines = append(lines, formatMoney("CA AGI", m.results["ca_540:17"]))
		lines = append(lines, formatMoney("CA Taxable Income", m.results["ca_540:19"]))
		lines = append(lines, formatMoney("CA Total Tax", m.results["ca_540:40"]))
		lines = append(lines, formatMoney("CA Withholding", m.results["ca_540:71"]))

		caRefund := m.results["ca_540:91"]
		caOwed := m.results["ca_540:93"]
		if caRefund > 0 {
			lines = append(lines, tui.SuccessStyle.Render(fmt.Sprintf("%-25s %s", "CA REFUND:", formatDollar(caRefund))))
		} else if caOwed > 0 {
			lines = append(lines, tui.ErrorStyle.Render(fmt.Sprintf("%-25s %s", "CA OWED:", formatDollar(caOwed))))
		} else {
			lines = append(lines, formatMoney("CA Balance", 0))
		}

		// Combined total
		lines = append(lines, "")
		lines = append(lines, tui.HighlightStyle.Render("--- Combined ---"))
		totalRefund := fedRefund + m.results["ca_540:91"]
		totalOwed := fedOwed + m.results["ca_540:93"]
		net := totalRefund - totalOwed
		if net > 0 {
			lines = append(lines, tui.SuccessStyle.Render(fmt.Sprintf("%-25s %s", "TOTAL REFUND:", formatDollar(net))))
		} else if net < 0 {
			lines = append(lines, tui.ErrorStyle.Render(fmt.Sprintf("%-25s %s", "TOTAL OWED:", formatDollar(-net))))
		} else {
			lines = append(lines, formatMoney("Net Balance", 0))
		}
	}

	// Validation summary
	lines = append(lines, "")
	errors, warnings, infos := m.countValidation()
	valSummary := fmt.Sprintf("Validation: %d errors, %d warnings, %d info", errors, warnings, infos)
	if errors > 0 {
		lines = append(lines, tui.ErrorStyle.Render(valSummary))
	} else if warnings > 0 {
		lines = append(lines, WarningStyle.Render(valSummary))
	} else {
		lines = append(lines, tui.SuccessStyle.Render(valSummary))
	}

	return lines
}

// --- Tab 1: Federal Detail ---

func (m ReviewView) renderFederalDetail() []string {
	var lines []string

	lines = append(lines, tui.HighlightStyle.Render("=== Form 1040 ==="))
	form1040Lines := []struct {
		key   string
		label string
	}{
		{"1040:1a", "Line 1a: Wages, salaries, tips"},
		{"1040:1z", "Line 1z: Total from W-2s"},
		{"1040:2a", "Line 2a: Tax-exempt interest"},
		{"1040:2b", "Line 2b: Taxable interest"},
		{"1040:3a", "Line 3a: Qualified dividends"},
		{"1040:3b", "Line 3b: Ordinary dividends"},
		{"1040:4a", "Line 4a: IRA distributions"},
		{"1040:4b", "Line 4b: Taxable IRA distributions"},
		{"1040:5a", "Line 5a: Pensions and annuities"},
		{"1040:5b", "Line 5b: Taxable pensions"},
		{"1040:6a", "Line 6a: Social Security benefits"},
		{"1040:6b", "Line 6b: Taxable Social Security"},
		{"1040:7", "Line 7: Capital gain or loss"},
		{"1040:8", "Line 8: Other income (Schedule 1)"},
		{"1040:9", "Line 9: Total income"},
		{"1040:10", "Line 10: Adjustments (Schedule 1)"},
		{"1040:11", "Line 11: Adjusted gross income"},
		{"1040:12", "Line 12: Standard/itemized deduction"},
		{"1040:13", "Line 13: Qualified business income deduction"},
		{"1040:14", "Line 14: Total deductions"},
		{"1040:15", "Line 15: Taxable income"},
		{"1040:16", "Line 16: Tax"},
		{"1040:17", "Line 17: Amount from Schedule 2 Part I"},
		{"1040:18", "Line 18: Sum of lines 16 and 17"},
		{"1040:19", "Line 19: Child/dependent credit"},
		{"1040:20", "Line 20: Amount from Schedule 3 Part I"},
		{"1040:21", "Line 21: Sum of lines 19 and 20"},
		{"1040:22", "Line 22: Subtract line 21 from 18"},
		{"1040:23", "Line 23: Other taxes (Schedule 2 Part II)"},
		{"1040:24", "Line 24: Total tax"},
		{"1040:25a", "Line 25a: W-2 withholding"},
		{"1040:25b", "Line 25b: 1099 withholding"},
		{"1040:25d", "Line 25d: Total withholding"},
		{"1040:26", "Line 26: Estimated tax payments"},
		{"1040:27", "Line 27: Earned income credit"},
		{"1040:33", "Line 33: Total payments"},
		{"1040:34", "Line 34: Overpayment (refund)"},
		{"1040:37", "Line 37: Amount owed"},
	}
	for _, l := range form1040Lines {
		val := m.results[l.key]
		if val != 0 {
			lines = append(lines, fmt.Sprintf("  %-42s %s", l.label, formatDollar(val)))
		}
	}

	// Schedule A (if itemized)
	if m.results["schedule_a:17"] > 0 {
		lines = append(lines, "")
		lines = append(lines, tui.HighlightStyle.Render("=== Schedule A (Itemized Deductions) ==="))
		schedALines := []struct {
			key   string
			label string
		}{
			{"schedule_a:1", "Line 1: Medical and dental expenses"},
			{"schedule_a:4", "Line 4: Deductible medical expenses"},
			{"schedule_a:5a", "Line 5a: State/local income taxes"},
			{"schedule_a:5b", "Line 5b: State/local sales taxes"},
			{"schedule_a:5c", "Line 5c: Real estate taxes"},
			{"schedule_a:5d", "Line 5d: Personal property taxes"},
			{"schedule_a:5e", "Line 5e: Total SALT"},
			{"schedule_a:7", "Line 7: SALT deduction (capped)"},
			{"schedule_a:8a", "Line 8a: Home mortgage interest"},
			{"schedule_a:10", "Line 10: Total interest"},
			{"schedule_a:12", "Line 12: Charitable cash contributions"},
			{"schedule_a:14", "Line 14: Total charitable"},
			{"schedule_a:15", "Line 15: Casualty/theft losses"},
			{"schedule_a:16", "Line 16: Other deductions"},
			{"schedule_a:17", "Line 17: Total itemized deductions"},
		}
		for _, l := range schedALines {
			val := m.results[l.key]
			if val != 0 {
				lines = append(lines, fmt.Sprintf("  %-42s %s", l.label, formatDollar(val)))
			}
		}
	}

	// Schedule B
	if m.results["schedule_b:4"] > 0 || m.results["schedule_b:6"] > 0 {
		lines = append(lines, "")
		lines = append(lines, tui.HighlightStyle.Render("=== Schedule B (Interest & Dividends) ==="))
		schedBLines := []struct {
			key   string
			label string
		}{
			{"schedule_b:4", "Line 4: Total interest"},
			{"schedule_b:6", "Line 6: Total dividends"},
		}
		for _, l := range schedBLines {
			val := m.results[l.key]
			if val != 0 {
				lines = append(lines, fmt.Sprintf("  %-42s %s", l.label, formatDollar(val)))
			}
		}
	}

	// Schedule C
	if m.results["schedule_c:31"] != 0 || m.results["schedule_c:7"] != 0 {
		lines = append(lines, "")
		lines = append(lines, tui.HighlightStyle.Render("=== Schedule C (Business Income) ==="))
		schedCLines := []struct {
			key   string
			label string
		}{
			{"schedule_c:7", "Line 7: Gross income"},
			{"schedule_c:28", "Line 28: Total expenses"},
			{"schedule_c:29", "Line 29: Tentative profit"},
			{"schedule_c:31", "Line 31: Net profit or loss"},
		}
		for _, l := range schedCLines {
			val := m.results[l.key]
			if val != 0 {
				lines = append(lines, fmt.Sprintf("  %-42s %s", l.label, formatDollar(val)))
			}
		}
	}

	// Schedule D
	if m.results["schedule_d:16"] != 0 || m.results["schedule_d:7"] != 0 {
		lines = append(lines, "")
		lines = append(lines, tui.HighlightStyle.Render("=== Schedule D (Capital Gains) ==="))
		schedDLines := []struct {
			key   string
			label string
		}{
			{"schedule_d:7", "Line 7: Net short-term gain/loss"},
			{"schedule_d:15", "Line 15: Net long-term gain/loss"},
			{"schedule_d:16", "Line 16: Combined gain/loss"},
			{"schedule_d:21", "Line 21: Capital gain tax"},
		}
		for _, l := range schedDLines {
			val := m.results[l.key]
			if val != 0 {
				lines = append(lines, fmt.Sprintf("  %-42s %s", l.label, formatDollar(val)))
			}
		}
	}

	// Schedule SE
	if m.results["schedule_se:12"] != 0 {
		lines = append(lines, "")
		lines = append(lines, tui.HighlightStyle.Render("=== Schedule SE (Self-Employment Tax) ==="))
		schedSELines := []struct {
			key   string
			label string
		}{
			{"schedule_se:4", "Line 4: Net SE earnings"},
			{"schedule_se:12", "Line 12: SE tax"},
			{"schedule_se:13", "Line 13: Deductible part of SE tax"},
		}
		for _, l := range schedSELines {
			val := m.results[l.key]
			if val != 0 {
				lines = append(lines, fmt.Sprintf("  %-42s %s", l.label, formatDollar(val)))
			}
		}
	}

	// Schedule 1
	if m.results["schedule_1:10"] != 0 || m.results["schedule_1:26"] != 0 {
		lines = append(lines, "")
		lines = append(lines, tui.HighlightStyle.Render("=== Schedule 1 (Additional Income/Adjustments) ==="))
		sched1Lines := []struct {
			key   string
			label string
		}{
			{"schedule_1:3", "Line 3: Business income/loss"},
			{"schedule_1:7", "Line 7: Capital gain/loss"},
			{"schedule_1:10", "Line 10: Total additional income"},
			{"schedule_1:15", "Line 15: HSA deduction"},
			{"schedule_1:16", "Line 16: Deductible SE tax"},
			{"schedule_1:20", "Line 20: Student loan interest deduction"},
			{"schedule_1:26", "Line 26: Total adjustments"},
		}
		for _, l := range sched1Lines {
			val := m.results[l.key]
			if val != 0 {
				lines = append(lines, fmt.Sprintf("  %-42s %s", l.label, formatDollar(val)))
			}
		}
	}

	// Schedule 2
	if m.results["schedule_2:4"] != 0 || m.results["schedule_2:21"] != 0 {
		lines = append(lines, "")
		lines = append(lines, tui.HighlightStyle.Render("=== Schedule 2 (Additional Taxes) ==="))
		sched2Lines := []struct {
			key   string
			label string
		}{
			{"schedule_2:4", "Line 4: Self-employment tax"},
			{"schedule_2:6", "Line 6: AMT"},
			{"schedule_2:21", "Line 21: Total additional taxes"},
		}
		for _, l := range sched2Lines {
			val := m.results[l.key]
			if val != 0 {
				lines = append(lines, fmt.Sprintf("  %-42s %s", l.label, formatDollar(val)))
			}
		}
	}

	// Schedule 3
	if m.results["schedule_3:8"] != 0 || m.results["schedule_3:15"] != 0 {
		lines = append(lines, "")
		lines = append(lines, tui.HighlightStyle.Render("=== Schedule 3 (Additional Credits/Payments) ==="))
		sched3Lines := []struct {
			key   string
			label string
		}{
			{"schedule_3:1", "Line 1: Foreign tax credit"},
			{"schedule_3:2", "Line 2: Child/dependent care credit"},
			{"schedule_3:8", "Line 8: Total nonrefundable credits"},
			{"schedule_3:10", "Line 10: Estimated tax payments"},
			{"schedule_3:15", "Line 15: Total other payments"},
		}
		for _, l := range sched3Lines {
			val := m.results[l.key]
			if val != 0 {
				lines = append(lines, fmt.Sprintf("  %-42s %s", l.label, formatDollar(val)))
			}
		}
	}

	// Form 8889 (HSA)
	if m.results["form_8889:13"] != 0 || m.results["form_8889:2"] != 0 {
		lines = append(lines, "")
		lines = append(lines, tui.HighlightStyle.Render("=== Form 8889 (HSA) ==="))
		hsa8889Lines := []struct {
			key   string
			label string
		}{
			{"form_8889:2", "Line 2: HSA contributions"},
			{"form_8889:9", "Line 9: Maximum deduction"},
			{"form_8889:13", "Line 13: HSA deduction"},
		}
		for _, l := range hsa8889Lines {
			val := m.results[l.key]
			if val != 0 {
				lines = append(lines, fmt.Sprintf("  %-42s %s", l.label, formatDollar(val)))
			}
		}
	}

	// Form 8995 (QBI)
	if m.results["form_8995:15"] != 0 || m.results["form_8995:5"] != 0 {
		lines = append(lines, "")
		lines = append(lines, tui.HighlightStyle.Render("=== Form 8995 (QBI Deduction) ==="))
		qbi8995Lines := []struct {
			key   string
			label string
		}{
			{"form_8995:5", "Line 5: Total QBI"},
			{"form_8995:10", "Line 10: QBI component"},
			{"form_8995:15", "Line 15: QBI deduction"},
		}
		for _, l := range qbi8995Lines {
			val := m.results[l.key]
			if val != 0 {
				lines = append(lines, fmt.Sprintf("  %-42s %s", l.label, formatDollar(val)))
			}
		}
	}

	if len(lines) == 1 {
		lines = append(lines, "  No federal form data available.")
	}

	return lines
}

// --- Tab 2: CA Detail ---

func (m ReviewView) renderCADetail() []string {
	var lines []string

	if m.state != "CA" {
		lines = append(lines, "No California return data. State is set to: "+m.state)
		return lines
	}

	lines = append(lines, tui.HighlightStyle.Render("=== Form 540 (California) ==="))
	ca540Lines := []struct {
		key   string
		label string
	}{
		{"ca_540:7", "Line 7: Federal wages"},
		{"ca_540:8", "Line 8: Federal AGI"},
		{"ca_540:9", "Line 9: CA Schedule CA adjustments"},
		{"ca_540:11", "Line 11: Total income"},
		{"ca_540:13", "Line 13: CA income"},
		{"ca_540:14", "Line 14: CA subtotal income"},
		{"ca_540:15", "Line 15: CA deductions (income-based)"},
		{"ca_540:16", "Line 16: CA adjusted income"},
		{"ca_540:17", "Line 17: CA AGI"},
		{"ca_540:18", "Line 18: CA standard/itemized deduction"},
		{"ca_540:19", "Line 19: CA taxable income"},
		{"ca_540:31", "Line 31: CA tax"},
		{"ca_540:32", "Line 32: Exemption credits"},
		{"ca_540:33", "Line 33: CA tax less exemption credits"},
		{"ca_540:35", "Line 35: CA taxable income for MHT"},
		{"ca_540:36", "Line 36: Mental Health Tax (1%)"},
		{"ca_540:40", "Line 40: Total CA tax"},
		{"ca_540:47", "Line 47: Total tax"},
		{"ca_540:61", "Line 61: CA income tax withheld"},
		{"ca_540:71", "Line 71: Total payments and credits"},
		{"ca_540:74", "Line 74: Tax after withholding"},
		{"ca_540:91", "Line 91: Overpayment (refund)"},
		{"ca_540:93", "Line 93: Amount owed"},
	}
	for _, l := range ca540Lines {
		val := m.results[l.key]
		if val != 0 {
			lines = append(lines, fmt.Sprintf("  %-42s %s", l.label, formatDollar(val)))
		}
	}

	// Schedule CA adjustments
	hasSchedCA := false
	schedCALines := []struct {
		key   string
		label string
	}{
		{"schedule_ca:1a_adj", "Line 1a: Wages adjustment"},
		{"schedule_ca:2b_adj", "Line 2b: Interest adjustment"},
		{"schedule_ca:3b_adj", "Line 3b: Dividends adjustment"},
		{"schedule_ca:4b_adj", "Line 4b: IRA adjustment"},
		{"schedule_ca:5b_adj", "Line 5b: Pensions adjustment"},
		{"schedule_ca:6b_adj", "Line 6b: Social Security adjustment"},
		{"schedule_ca:7_adj", "Line 7: Capital gain adjustment"},
		{"schedule_ca:8_adj", "Line 8: Other income adjustment"},
		{"schedule_ca:total_adj", "Total Schedule CA adjustment"},
	}
	for _, l := range schedCALines {
		val := m.results[l.key]
		if val != 0 {
			if !hasSchedCA {
				lines = append(lines, "")
				lines = append(lines, tui.HighlightStyle.Render("=== Schedule CA (Adjustments) ==="))
				hasSchedCA = true
			}
			lines = append(lines, fmt.Sprintf("  %-42s %s", l.label, formatDollar(val)))
		}
	}

	if len(lines) == 1 {
		lines = append(lines, "  No California form data available.")
	}

	return lines
}

// --- Tab 3: Prior Year Comparison ---

func (m ReviewView) renderPriorYearComparison() []string {
	var lines []string

	if m.priorResults == nil || len(m.priorResults) == 0 {
		lines = append(lines, "No prior year data available.")
		lines = append(lines, "")
		lines = append(lines, tui.HelpStyle.Render("Import a prior year return to see year-over-year comparisons."))
		return lines
	}

	lines = append(lines, tui.HighlightStyle.Render(
		fmt.Sprintf("=== Year-over-Year Comparison: %d vs %d ===", m.taxYear, m.taxYear-1),
	))
	lines = append(lines, "")

	header := fmt.Sprintf("  %-28s %14s %14s %10s", "Field", "This Year", "Last Year", "Change")
	lines = append(lines, tui.PromptStyle.Render(header))
	lines = append(lines, "  "+strings.Repeat("-", 70))

	comparisonFields := []struct {
		key   string
		label string
	}{
		{"1040:1a", "Wages"},
		{"1040:9", "Total Income"},
		{"1040:11", "AGI"},
		{"1040:12", "Deduction"},
		{"1040:15", "Taxable Income"},
		{"1040:16", "Federal Tax"},
		{"1040:24", "Total Tax"},
		{"1040:25d", "Withholding"},
		{"1040:34", "Refund"},
		{"1040:37", "Amount Owed"},
	}

	// Add CA fields if applicable
	if m.state == "CA" {
		comparisonFields = append(comparisonFields, []struct {
			key   string
			label string
		}{
			{"ca_540:17", "CA AGI"},
			{"ca_540:19", "CA Taxable Income"},
			{"ca_540:40", "CA Total Tax"},
			{"ca_540:71", "CA Withholding"},
			{"ca_540:91", "CA Refund"},
			{"ca_540:93", "CA Amount Owed"},
		}...)
	}

	for _, f := range comparisonFields {
		thisYear := m.results[f.key]
		lastYear := m.priorResults[f.key]

		if thisYear == 0 && lastYear == 0 {
			continue
		}

		changeStr := "---"
		isLarge := false
		if lastYear != 0 {
			changePct := (thisYear - lastYear) / math.Abs(lastYear) * 100
			changeStr = fmt.Sprintf("%+.1f%%", changePct)
			if math.Abs(changePct) > 20 {
				isLarge = true
			}
		} else if thisYear != 0 {
			changeStr = "NEW"
			isLarge = true
		}

		row := fmt.Sprintf("  %-28s %14s %14s %10s",
			f.label,
			formatDollar(thisYear),
			formatDollar(lastYear),
			changeStr,
		)

		if isLarge {
			lines = append(lines, WarningStyle.Render(row))
		} else {
			lines = append(lines, row)
		}
	}

	return lines
}

// --- Tab 4: Validation ---

func (m ReviewView) renderValidation() []string {
	var lines []string

	if len(m.validation.Results) == 0 {
		lines = append(lines, tui.SuccessStyle.Render("No validation issues found. Return is ready for filing."))
		return lines
	}

	errors, warnings, infos := m.countValidation()
	lines = append(lines, fmt.Sprintf("Total: %d errors, %d warnings, %d info", errors, warnings, infos))
	lines = append(lines, "")

	// Group by severity
	if errors > 0 {
		lines = append(lines, tui.ErrorStyle.Render("=== ERRORS (must fix before filing) ==="))
		for _, r := range m.validation.Results {
			if r.Severity == efile.SeverityError {
				lines = append(lines, tui.ErrorStyle.Render(
					fmt.Sprintf("  [%s] %s (%s)", r.Code, r.Message, r.Field),
				))
			}
		}
		lines = append(lines, "")
	}

	if warnings > 0 {
		lines = append(lines, WarningStyle.Render("=== WARNINGS (review recommended) ==="))
		for _, r := range m.validation.Results {
			if r.Severity == efile.SeverityWarning {
				lines = append(lines, WarningStyle.Render(
					fmt.Sprintf("  [%s] %s (%s)", r.Code, r.Message, r.Field),
				))
			}
		}
		lines = append(lines, "")
	}

	if infos > 0 {
		lines = append(lines, InfoStyle.Render("=== INFO ==="))
		for _, r := range m.validation.Results {
			if r.Severity == efile.SeverityInfo {
				lines = append(lines, InfoStyle.Render(
					fmt.Sprintf("  [%s] %s (%s)", r.Code, r.Message, r.Field),
				))
			}
		}
	}

	return lines
}

// --- Helpers ---

func (m ReviewView) countValidation() (errors, warnings, infos int) {
	for _, r := range m.validation.Results {
		switch r.Severity {
		case efile.SeverityError:
			errors++
		case efile.SeverityWarning:
			warnings++
		case efile.SeverityInfo:
			infos++
		}
	}
	return
}
