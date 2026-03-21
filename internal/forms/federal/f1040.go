package federal

import (
	"math"

	"taxpilot/internal/forms"
	"taxpilot/pkg/taxmath"
)

func init() { forms.RegisterForm(F1040) }

// F1040 returns the FormDef for Form 1040 — U.S. Individual Income Tax Return.
// This is a simplified MVP covering a single (or multiple) W-2 filer with
// standard deduction. Additional income sources and itemized deductions will
// be added in future iterations.
func F1040() *forms.FormDef {
	return &forms.FormDef{
		ID:           forms.FormF1040,
		Name:         "U.S. Individual Income Tax Return",
		Jurisdiction: forms.Federal,
		TaxYears:      []int{2025},
		QuestionGroup: forms.GroupPersonal,
		QuestionOrder: 1,
		Fields: []forms.FieldDef{
			// --- Identification ---
			{
				Line:    forms.LineFilingStatus,
				Type:    forms.UserInput,
				ValueType: forms.StringValue,
				Label:   "Filing status",
				Prompt:  "What is your filing status?",
				Options: forms.FilingStatusOptions,
			},
			{
				Line:   forms.LineFirstName,
				Type:   forms.UserInput,
				ValueType: forms.StringValue,
				Label:  "First name",
				Prompt: "What is your first name?",
			},
			{
				Line:   forms.LineLastName,
				Type:   forms.UserInput,
				ValueType: forms.StringValue,
				Label:  "Last name",
				Prompt: "What is your last name?",
			},
			{
				Line:   forms.LineSSN,
				Type:   forms.UserInput,
				ValueType: forms.StringValue,
				Label:  "Social Security number",
				Prompt: "What is your Social Security number (XXX-XX-XXXX)?",
			},

			// --- Income ---

			// Foreign wages not reported on a US W-2 (e.g., from a foreign employer)
			{
				Line:      forms.LineForeignWages,
				Type:      forms.UserInput,
				Label:     "Foreign wages (not on W-2)",
				Prompt:    "Enter wages from foreign employers not reported on a US W-2, converted to USD:",
			},
			// Foreign employer description
			{
				Line:      forms.LineForeignEmployer,
				Type:      forms.UserInput,
				ValueType: forms.StringValue,
				Label:     "Foreign employer(s)",
				Prompt:    "Describe the foreign employer(s) (e.g., \"Volvo AB, Sweden\"):",
			},

			// Line 1a: Wages from W-2s + foreign wages
			{
				Line:      "1a",
				Type:      forms.Computed,
				Label:     "Wages, salaries, tips (W-2 + foreign)",
				DependsOn: []string{forms.W2WildcardWages, forms.F1040ForeignWages},
				Compute: func(dv forms.DepValues) float64 {
					return dv.SumAll(forms.W2WildcardWages) + dv.Get(forms.F1040ForeignWages)
				},
			},
			// Line 1z: Total from adding lines 1a through 1h (MVP: only 1a)
			{
				Line:      "1z",
				Type:      forms.Computed,
				Label:     "Total from W-2s and other wage sources",
				DependsOn: []string{forms.F1040Line1a},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.F1040Line1a)
				},
			},
			// Line 2a: Tax-exempt interest (from 1099-INT Box 8)
			{
				Line:      "2a",
				Type:      forms.Computed,
				Label:     "Tax-exempt interest",
				DependsOn: []string{forms.F1099INTWildcardTaxExempt},
				Compute: func(dv forms.DepValues) float64 {
					return dv.SumAll(forms.F1099INTWildcardTaxExempt)
				},
			},
			// Line 2b: Taxable interest (from Schedule B line 4)
			{
				Line:      "2b",
				Type:      forms.Computed,
				Label:     "Taxable interest",
				DependsOn: []string{forms.SchedBLine4},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.SchedBLine4)
				},
			},
			// Line 3a: Qualified dividends (from 1099-DIV Box 1b)
			{
				Line:      "3a",
				Type:      forms.Computed,
				Label:     "Qualified dividends",
				DependsOn: []string{forms.F1099DIVWildcardQualified},
				Compute: func(dv forms.DepValues) float64 {
					return dv.SumAll(forms.F1099DIVWildcardQualified)
				},
			},
			// Line 3b: Ordinary dividends (from Schedule B line 6)
			{
				Line:      "3b",
				Type:      forms.Computed,
				Label:     "Ordinary dividends",
				DependsOn: []string{forms.SchedBLine6},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.SchedBLine6)
				},
			},
			// Line 7: Capital gain or (loss) from Schedule 1 line 7
			{
				Line:      "7",
				Type:      forms.Computed,
				Label:     "Capital gain or (loss)",
				DependsOn: []string{forms.Sched1Line7},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.Sched1Line7)
				},
			},
			// Line 8: Other income from Schedule 1 line 10
			{
				Line:      "8",
				Type:      forms.Computed,
				Label:     "Other income from Schedule 1",
				DependsOn: []string{forms.Sched1Line10},
				Compute: func(dv forms.DepValues) float64 {
					// Schedule 1 line 10 minus capital gains (already on line 7)
					return dv.Get(forms.Sched1Line10) - dv.Get(forms.Sched1Line7)
				},
			},
			// Line 9: Total income
			{
				Line:      "9",
				Type:      forms.Computed,
				Label:     "Total income",
				DependsOn: []string{forms.F1040Line1z, forms.F1040Line2b, forms.F1040Line3b, forms.F1040Line7, forms.F1040Line8},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.F1040Line1z) +
						dv.Get(forms.F1040Line2b) +
						dv.Get(forms.F1040Line3b) +
						dv.Get(forms.F1040Line7) +
						dv.Get(forms.F1040Line8)
				},
			},
			// Line 10: Adjustments to income (from Schedule 1 line 26)
			{
				Line:      "10",
				Type:      forms.Computed,
				Label:     "Adjustments to income",
				DependsOn: []string{forms.Sched1Line26},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.Sched1Line26)
				},
			},
			// Line 11: Adjusted gross income (AGI)
			{
				Line:      "11",
				Type:      forms.Computed,
				Label:     "Adjusted gross income",
				DependsOn: []string{forms.F1040Line9, forms.F1040Line10},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.F1040Line9) - dv.Get(forms.F1040Line10)
				},
			},

			// --- Deductions ---

			// Line 12: Deduction — larger of standard deduction or itemized (Schedule A)
			{
				Line:      "12",
				Type:      forms.Computed,
				Label:     "Standard deduction or itemized deductions",
				DependsOn: []string{forms.F1040FilingStatus, forms.SchedALine17},
				Compute: func(dv forms.DepValues) float64 {
					fs := taxmath.FilingStatus(dv.GetString(forms.F1040FilingStatus))
					standard := taxmath.GetStandardDeduction(dv.TaxYear(), taxmath.Federal, fs)
					itemized := dv.Get(forms.SchedALine17)
					return math.Max(standard, itemized)
				},
			},
			// Line 13: Qualified business income deduction (from Form 8995)
			{
				Line:      "13",
				Type:      forms.Computed,
				Label:     "Qualified business income deduction",
				DependsOn: []string{forms.F8995Line10},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.F8995Line10)
				},
			},
			// Line 14: Total deductions
			{
				Line:      "14",
				Type:      forms.Computed,
				Label:     "Total deductions",
				DependsOn: []string{forms.F1040Line12, forms.F1040Line13},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.F1040Line12) + dv.Get(forms.F1040Line13)
				},
			},
			// Line 15: Taxable income
			{
				Line:      "15",
				Type:      forms.Computed,
				Label:     "Taxable income",
				DependsOn: []string{forms.F1040Line11, forms.F1040Line14},
				Compute: func(dv forms.DepValues) float64 {
					return math.Max(0, dv.Get(forms.F1040Line11)-dv.Get(forms.F1040Line14))
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
				DependsOn: []string{forms.F1040Line15, forms.F1040FilingStatus, forms.F2555TotalExclusion},
				Compute: func(dv forms.DepValues) float64 {
					fs := taxmath.FilingStatus(dv.GetString(forms.F1040FilingStatus))
					taxableIncome := dv.Get(forms.F1040Line15)
					excludedIncome := dv.Get(forms.F2555TotalExclusion)
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
				DependsOn: []string{forms.Sched2Line3},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.Sched2Line3)
				},
			},
			// Line 20: Amount from Schedule 3 Part I (nonrefundable credits)
			{
				Line:      "20",
				Type:      forms.Computed,
				Label:     "Amount from Schedule 3, Part I",
				DependsOn: []string{forms.Sched3Line8},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.Sched3Line8)
				},
			},
			// Line 22: Tax after credits (line 16 + 17 - 20, but not less than 0)
			{
				Line:      "22",
				Type:      forms.Computed,
				Label:     "Tax after credits",
				DependsOn: []string{forms.F1040Line16, forms.F1040Line17, forms.F1040Line20},
				Compute: func(dv forms.DepValues) float64 {
					return math.Max(0, dv.Get(forms.F1040Line16)+dv.Get(forms.F1040Line17)-dv.Get(forms.F1040Line20))
				},
			},
			// Line 23: Other taxes from Schedule 2 Part II
			{
				Line:      "23",
				Type:      forms.Computed,
				Label:     "Other taxes from Schedule 2",
				DependsOn: []string{forms.Sched2Line21},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.Sched2Line21)
				},
			},
			// Line 24: Total tax (line 22 + line 23)
			{
				Line:      "24",
				Type:      forms.Computed,
				Label:     "Total tax",
				DependsOn: []string{forms.F1040Line22, forms.F1040Line23},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.F1040Line22) + dv.Get(forms.F1040Line23)
				},
			},

			// --- Payments ---

			// Line 25a: Federal income tax withheld from W-2s
			{
				Line:      "25a",
				Type:      forms.Computed,
				Label:     "Federal income tax withheld from W-2s",
				DependsOn: []string{forms.W2WildcardFedTaxWH},
				Compute: func(dv forms.DepValues) float64 {
					return dv.SumAll(forms.W2WildcardFedTaxWH)
				},
			},
			// Line 25b: Federal income tax withheld from 1099s
			{
				Line:      "25b",
				Type:      forms.Computed,
				Label:     "Federal income tax withheld from 1099s",
				DependsOn: []string{forms.F1099INTWildcardFedTaxWH, forms.F1099DIVWildcardFedTaxWH, forms.F1099NECWildcardFedTaxWH, forms.F1099BWildcardFedTaxWH},
				Compute: func(dv forms.DepValues) float64 {
					return dv.SumAll(forms.F1099INTWildcardFedTaxWH) +
						dv.SumAll(forms.F1099DIVWildcardFedTaxWH) +
						dv.SumAll(forms.F1099NECWildcardFedTaxWH) +
						dv.SumAll(forms.F1099BWildcardFedTaxWH)
				},
			},
			// Line 25d: Total federal tax withheld
			{
				Line:      "25d",
				Type:      forms.Computed,
				Label:     "Total federal tax withheld",
				DependsOn: []string{forms.F1040Line25a, forms.F1040Line25b},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.F1040Line25a) + dv.Get(forms.F1040Line25b)
				},
			},
			// Line 31: Other payments from Schedule 3 (estimated tax, etc.)
			{
				Line:      "31",
				Type:      forms.Computed,
				Label:     "Other payments from Schedule 3",
				DependsOn: []string{forms.Sched3Line15},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.Sched3Line15)
				},
			},
			// Line 33: Total payments
			{
				Line:      "33",
				Type:      forms.Computed,
				Label:     "Total payments",
				DependsOn: []string{forms.F1040Line25d, forms.F1040Line31},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.F1040Line25d) + dv.Get(forms.F1040Line31)
				},
			},

			// --- Refund or Amount Owed ---

			// Line 34: Overpayment (refund)
			{
				Line:      "34",
				Type:      forms.Computed,
				Label:     "Overpayment (refund)",
				DependsOn: []string{forms.F1040Line33, forms.F1040Line24},
				Compute: func(dv forms.DepValues) float64 {
					return math.Max(0, dv.Get(forms.F1040Line33)-dv.Get(forms.F1040Line24))
				},
			},
			// Line 37: Amount you owe
			{
				Line:      "37",
				Type:      forms.Computed,
				Label:     "Amount you owe",
				DependsOn: []string{forms.F1040Line24, forms.F1040Line33},
				Compute: func(dv forms.DepValues) float64 {
					return math.Max(0, dv.Get(forms.F1040Line24)-dv.Get(forms.F1040Line33))
				},
			},
		},
	}
}
