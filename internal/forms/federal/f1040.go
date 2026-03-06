package federal

import (
	"math"

	"taxpilot/internal/forms"
	"taxpilot/pkg/taxmath"
)

// F1040 returns the FormDef for Form 1040 — U.S. Individual Income Tax Return.
// This is a simplified MVP covering a single (or multiple) W-2 filer with
// standard deduction. Additional income sources and itemized deductions will
// be added in future iterations.
func F1040() *forms.FormDef {
	return &forms.FormDef{
		ID:           "1040",
		Name:         "U.S. Individual Income Tax Return",
		Jurisdiction: forms.Federal,
		TaxYears:     []int{2025},
		Fields: []forms.FieldDef{
			// --- Identification ---
			{
				Line:    "filing_status",
				Type:    forms.UserInput,
				Label:   "Filing status",
				Prompt:  "What is your filing status?",
				Options: []string{"single", "mfj", "mfs", "hoh", "qss"},
			},
			{
				Line:   "first_name",
				Type:   forms.UserInput,
				Label:  "First name",
				Prompt: "What is your first name?",
			},
			{
				Line:   "last_name",
				Type:   forms.UserInput,
				Label:  "Last name",
				Prompt: "What is your last name?",
			},
			{
				Line:   "ssn",
				Type:   forms.UserInput,
				Label:  "Social Security number",
				Prompt: "What is your Social Security number (XXX-XX-XXXX)?",
			},

			// --- Income ---

			// Line 1a: Wages from W-2s
			{
				Line:      "1a",
				Type:      forms.Computed,
				Label:     "Wages, salaries, tips (from W-2 Box 1)",
				DependsOn: []string{"w2:*:wages"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.SumAll("w2:*:wages")
				},
			},
			// Line 1z: Total from adding lines 1a through 1h (MVP: only 1a)
			{
				Line:      "1z",
				Type:      forms.Computed,
				Label:     "Total from W-2s and other wage sources",
				DependsOn: []string{"1040:1a"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("1040:1a")
				},
			},
			// Line 9: Total income
			{
				Line:      "9",
				Type:      forms.Computed,
				Label:     "Total income",
				DependsOn: []string{"1040:1z"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("1040:1z")
				},
			},
			// Line 10: Adjustments to income (0 for MVP)
			{
				Line:      "10",
				Type:      forms.Computed,
				Label:     "Adjustments to income",
				DependsOn: []string{},
				Compute: func(dv forms.DepValues) float64 {
					return 0
				},
			},
			// Line 11: Adjusted gross income (AGI)
			{
				Line:      "11",
				Type:      forms.Computed,
				Label:     "Adjusted gross income",
				DependsOn: []string{"1040:9", "1040:10"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("1040:9") - dv.Get("1040:10")
				},
			},

			// --- Deductions ---

			// Line 12: Standard deduction (or itemized — standard for MVP)
			{
				Line:      "12",
				Type:      forms.Computed,
				Label:     "Standard deduction",
				DependsOn: []string{"1040:filing_status"},
				Compute: func(dv forms.DepValues) float64 {
					fs := taxmath.FilingStatus(dv.GetString("1040:filing_status"))
					return taxmath.GetStandardDeduction(dv.TaxYear(), taxmath.Federal, fs)
				},
			},
			// Line 13: Qualified business income deduction (0 for MVP)
			{
				Line:      "13",
				Type:      forms.Computed,
				Label:     "Qualified business income deduction",
				DependsOn: []string{},
				Compute: func(dv forms.DepValues) float64 {
					return 0
				},
			},
			// Line 14: Total deductions
			{
				Line:      "14",
				Type:      forms.Computed,
				Label:     "Total deductions",
				DependsOn: []string{"1040:12", "1040:13"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("1040:12") + dv.Get("1040:13")
				},
			},
			// Line 15: Taxable income
			{
				Line:      "15",
				Type:      forms.Computed,
				Label:     "Taxable income",
				DependsOn: []string{"1040:11", "1040:14"},
				Compute: func(dv forms.DepValues) float64 {
					return math.Max(0, dv.Get("1040:11")-dv.Get("1040:14"))
				},
			},

			// --- Tax computation ---

			// Line 16: Tax
			{
				Line:      "16",
				Type:      forms.Computed,
				Label:     "Tax",
				DependsOn: []string{"1040:15", "1040:filing_status"},
				Compute: func(dv forms.DepValues) float64 {
					fs := taxmath.FilingStatus(dv.GetString("1040:filing_status"))
					taxableIncome := dv.Get("1040:15")
					return taxmath.ComputeTax(taxableIncome, fs, dv.TaxYear(), taxmath.Federal)
				},
			},
			// Line 24: Total tax (for MVP, same as line 16 — no additional taxes)
			{
				Line:      "24",
				Type:      forms.Computed,
				Label:     "Total tax",
				DependsOn: []string{"1040:16"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("1040:16")
				},
			},

			// --- Payments ---

			// Line 25a: Federal income tax withheld from W-2s
			{
				Line:      "25a",
				Type:      forms.Computed,
				Label:     "Federal income tax withheld from W-2s",
				DependsOn: []string{"w2:*:federal_tax_withheld"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.SumAll("w2:*:federal_tax_withheld")
				},
			},
			// Line 25d: Total federal tax withheld
			{
				Line:      "25d",
				Type:      forms.Computed,
				Label:     "Total federal tax withheld",
				DependsOn: []string{"1040:25a"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("1040:25a")
				},
			},
			// Line 33: Total payments
			{
				Line:      "33",
				Type:      forms.Computed,
				Label:     "Total payments",
				DependsOn: []string{"1040:25d"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("1040:25d")
				},
			},

			// --- Refund or Amount Owed ---

			// Line 34: Overpayment (refund)
			{
				Line:      "34",
				Type:      forms.Computed,
				Label:     "Overpayment (refund)",
				DependsOn: []string{"1040:33", "1040:24"},
				Compute: func(dv forms.DepValues) float64 {
					return math.Max(0, dv.Get("1040:33")-dv.Get("1040:24"))
				},
			},
			// Line 37: Amount you owe
			{
				Line:      "37",
				Type:      forms.Computed,
				Label:     "Amount you owe",
				DependsOn: []string{"1040:24", "1040:33"},
				Compute: func(dv forms.DepValues) float64 {
					return math.Max(0, dv.Get("1040:24")-dv.Get("1040:33"))
				},
			},
		},
	}
}
