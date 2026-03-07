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
			// Line 2a: Tax-exempt interest (from 1099-INT Box 8)
			{
				Line:      "2a",
				Type:      forms.Computed,
				Label:     "Tax-exempt interest",
				DependsOn: []string{"1099int:*:tax_exempt_interest"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.SumAll("1099int:*:tax_exempt_interest")
				},
			},
			// Line 2b: Taxable interest (from Schedule B line 4)
			{
				Line:      "2b",
				Type:      forms.Computed,
				Label:     "Taxable interest",
				DependsOn: []string{"schedule_b:4"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_b:4")
				},
			},
			// Line 3a: Qualified dividends (from 1099-DIV Box 1b)
			{
				Line:      "3a",
				Type:      forms.Computed,
				Label:     "Qualified dividends",
				DependsOn: []string{"1099div:*:qualified_dividends"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.SumAll("1099div:*:qualified_dividends")
				},
			},
			// Line 3b: Ordinary dividends (from Schedule B line 6)
			{
				Line:      "3b",
				Type:      forms.Computed,
				Label:     "Ordinary dividends",
				DependsOn: []string{"schedule_b:6"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_b:6")
				},
			},
			// Line 7: Capital gain or (loss) from Schedule 1 line 7
			{
				Line:      "7",
				Type:      forms.Computed,
				Label:     "Capital gain or (loss)",
				DependsOn: []string{"schedule_1:7"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_1:7")
				},
			},
			// Line 8: Other income from Schedule 1 line 10
			{
				Line:      "8",
				Type:      forms.Computed,
				Label:     "Other income from Schedule 1",
				DependsOn: []string{"schedule_1:10"},
				Compute: func(dv forms.DepValues) float64 {
					// Schedule 1 line 10 minus capital gains (already on line 7)
					return dv.Get("schedule_1:10") - dv.Get("schedule_1:7")
				},
			},
			// Line 9: Total income
			{
				Line:      "9",
				Type:      forms.Computed,
				Label:     "Total income",
				DependsOn: []string{"1040:1z", "1040:2b", "1040:3b", "1040:7", "1040:8"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("1040:1z") +
						dv.Get("1040:2b") +
						dv.Get("1040:3b") +
						dv.Get("1040:7") +
						dv.Get("1040:8")
				},
			},
			// Line 10: Adjustments to income (from Schedule 1 line 26)
			{
				Line:      "10",
				Type:      forms.Computed,
				Label:     "Adjustments to income",
				DependsOn: []string{"schedule_1:26"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_1:26")
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

			// Line 12: Deduction — larger of standard deduction or itemized (Schedule A)
			{
				Line:      "12",
				Type:      forms.Computed,
				Label:     "Standard deduction or itemized deductions",
				DependsOn: []string{"1040:filing_status", "schedule_a:17"},
				Compute: func(dv forms.DepValues) float64 {
					fs := taxmath.FilingStatus(dv.GetString("1040:filing_status"))
					standard := taxmath.GetStandardDeduction(dv.TaxYear(), taxmath.Federal, fs)
					itemized := dv.Get("schedule_a:17")
					return math.Max(standard, itemized)
				},
			},
			// Line 13: Qualified business income deduction (from Form 8995)
			{
				Line:      "13",
				Type:      forms.Computed,
				Label:     "Qualified business income deduction",
				DependsOn: []string{"form_8995:10"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("form_8995:10")
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
			// When FEIE (Form 2555) is claimed, uses the "stacking" method:
			// tax is computed at the rate that would apply if the excluded
			// income were still included, preventing bracket manipulation.
			{
				Line:      "16",
				Type:      forms.Computed,
				Label:     "Tax",
				DependsOn: []string{"1040:15", "1040:filing_status", "form_2555:total_exclusion"},
				Compute: func(dv forms.DepValues) float64 {
					fs := taxmath.FilingStatus(dv.GetString("1040:filing_status"))
					taxableIncome := dv.Get("1040:15")
					excludedIncome := dv.Get("form_2555:total_exclusion")
					if excludedIncome > 0 {
						return taxmath.ComputeTaxWithStacking(taxableIncome, excludedIncome, fs, dv.TaxYear(), taxmath.Federal)
					}
					return taxmath.ComputeTax(taxableIncome, fs, dv.TaxYear(), taxmath.Federal)
				},
			},
			// Line 17: Amount from Schedule 2 Part I (AMT, etc.)
			{
				Line:      "17",
				Type:      forms.Computed,
				Label:     "Amount from Schedule 2, Part I",
				DependsOn: []string{"schedule_2:3"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_2:3")
				},
			},
			// Line 20: Amount from Schedule 3 Part I (nonrefundable credits)
			{
				Line:      "20",
				Type:      forms.Computed,
				Label:     "Amount from Schedule 3, Part I",
				DependsOn: []string{"schedule_3:8"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_3:8")
				},
			},
			// Line 22: Tax after credits (line 16 + 17 - 20, but not less than 0)
			{
				Line:      "22",
				Type:      forms.Computed,
				Label:     "Tax after credits",
				DependsOn: []string{"1040:16", "1040:17", "1040:20"},
				Compute: func(dv forms.DepValues) float64 {
					return math.Max(0, dv.Get("1040:16")+dv.Get("1040:17")-dv.Get("1040:20"))
				},
			},
			// Line 23: Other taxes from Schedule 2 Part II
			{
				Line:      "23",
				Type:      forms.Computed,
				Label:     "Other taxes from Schedule 2",
				DependsOn: []string{"schedule_2:21"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_2:21")
				},
			},
			// Line 24: Total tax (line 22 + line 23)
			{
				Line:      "24",
				Type:      forms.Computed,
				Label:     "Total tax",
				DependsOn: []string{"1040:22", "1040:23"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("1040:22") + dv.Get("1040:23")
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
			// Line 25b: Federal income tax withheld from 1099s
			{
				Line:      "25b",
				Type:      forms.Computed,
				Label:     "Federal income tax withheld from 1099s",
				DependsOn: []string{"1099int:*:federal_tax_withheld", "1099div:*:federal_tax_withheld", "1099nec:*:federal_tax_withheld", "1099b:*:federal_tax_withheld"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.SumAll("1099int:*:federal_tax_withheld") +
						dv.SumAll("1099div:*:federal_tax_withheld") +
						dv.SumAll("1099nec:*:federal_tax_withheld") +
						dv.SumAll("1099b:*:federal_tax_withheld")
				},
			},
			// Line 25d: Total federal tax withheld
			{
				Line:      "25d",
				Type:      forms.Computed,
				Label:     "Total federal tax withheld",
				DependsOn: []string{"1040:25a", "1040:25b"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("1040:25a") + dv.Get("1040:25b")
				},
			},
			// Line 31: Other payments from Schedule 3 (estimated tax, etc.)
			{
				Line:      "31",
				Type:      forms.Computed,
				Label:     "Other payments from Schedule 3",
				DependsOn: []string{"schedule_3:15"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_3:15")
				},
			},
			// Line 33: Total payments
			{
				Line:      "33",
				Type:      forms.Computed,
				Label:     "Total payments",
				DependsOn: []string{"1040:25d", "1040:31"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("1040:25d") + dv.Get("1040:31")
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
