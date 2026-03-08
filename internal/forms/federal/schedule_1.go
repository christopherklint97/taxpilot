package federal

import (
	"taxpilot/internal/forms"
)

// Schedule1 returns the FormDef for Schedule 1 — Additional Income and
// Adjustments to Income. This wires additional income sources (interest,
// dividends, capital gains, business income) into Form 1040 lines 8 and 10.
//
// Part I: Additional Income flows to 1040 line 8 (added to total income).
// Part II: Adjustments flows to 1040 line 10 (subtracted from total income).
func Schedule1() *forms.FormDef {
	return &forms.FormDef{
		ID:           forms.FormSchedule1,
		Name:         "Schedule 1 — Additional Income and Adjustments to Income",
		Jurisdiction: forms.Federal,
		TaxYears:     []int{2025},
		Fields: []forms.FieldDef{
			// --- Part I: Additional Income ---

			// Line 1: Taxable refunds of state/local income taxes (deferred)
			{
				Line:      "1",
				Type:      forms.Computed,
				Label:     "Taxable refunds of state/local income taxes",
				DependsOn: []string{},
				Compute: func(dv forms.DepValues) float64 {
					return 0 // deferred — requires prior-year itemized deduction tracking
				},
			},
			// Line 2a: Alimony received (deferred)
			{
				Line:      "2a",
				Type:      forms.Computed,
				Label:     "Alimony received",
				DependsOn: []string{},
				Compute: func(dv forms.DepValues) float64 {
					return 0
				},
			},
			// Line 3: Business income (from Schedule C line 31)
			{
				Line:      "3",
				Type:      forms.Computed,
				Label:     "Business income or (loss)",
				DependsOn: []string{"schedule_c:31"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_c:31")
				},
			},
			// Line 7: Capital gain or (loss) — from Schedule D line 16
			{
				Line:      "7",
				Type:      forms.Computed,
				Label:     "Capital gain or (loss)",
				DependsOn: []string{"schedule_d:16"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_d:16")
				},
			},
			// Line 8d: Foreign earned income exclusion (from Form 2555)
			{
				Line:      "8d",
				Type:      forms.Computed,
				Label:     "Foreign earned income exclusion (Form 2555)",
				DependsOn: []string{"form_2555:total_exclusion"},
				Compute: func(dv forms.DepValues) float64 {
					return -dv.Get("form_2555:total_exclusion")
				},
			},
			// Line 10: Total additional income (sum of Part I lines)
			{
				Line:      "10",
				Type:      forms.Computed,
				Label:     "Total additional income",
				DependsOn: []string{"schedule_1:1", "schedule_1:2a", "schedule_1:3", "schedule_1:7", "schedule_1:8d"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_1:1") +
						dv.Get("schedule_1:2a") +
						dv.Get("schedule_1:3") +
						dv.Get("schedule_1:7") +
						dv.Get("schedule_1:8d")
				},
			},

			// --- Part II: Adjustments to Income ---

			// Line 11: Educator expenses (deferred)
			{
				Line:      "11",
				Type:      forms.Computed,
				Label:     "Educator expenses",
				DependsOn: []string{},
				Compute: func(dv forms.DepValues) float64 {
					return 0
				},
			},
			// Line 15: HSA deduction (from Form 8889 line 9)
			{
				Line:      "15",
				Type:      forms.Computed,
				Label:     "HSA deduction",
				DependsOn: []string{"form_8889:9"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("form_8889:9")
				},
			},
			// Line 16: Self-employment tax deduction (from Schedule SE line 7)
			{
				Line:      "16",
				Type:      forms.Computed,
				Label:     "Deductible part of self-employment tax",
				DependsOn: []string{"schedule_se:7"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_se:7")
				},
			},
			// Line 20: IRA deduction (deferred)
			{
				Line:      "20",
				Type:      forms.Computed,
				Label:     "IRA deduction",
				DependsOn: []string{},
				Compute: func(dv forms.DepValues) float64 {
					return 0
				},
			},
			// Line 21: Student loan interest deduction (deferred)
			{
				Line:      "21",
				Type:      forms.Computed,
				Label:     "Student loan interest deduction",
				DependsOn: []string{},
				Compute: func(dv forms.DepValues) float64 {
					return 0
				},
			},
			// Line 24: Early withdrawal penalty from 1099-INT
			{
				Line:      "24",
				Type:      forms.Computed,
				Label:     "Penalty on early withdrawal of savings",
				DependsOn: []string{"1099int:*:early_withdrawal_penalty"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.SumAll("1099int:*:early_withdrawal_penalty")
				},
			},
			// Line 26: Total adjustments (sum of Part II lines)
			{
				Line:      "26",
				Type:      forms.Computed,
				Label:     "Total adjustments to income",
				DependsOn: []string{"schedule_1:11", "schedule_1:15", "schedule_1:16", "schedule_1:20", "schedule_1:21", "schedule_1:24"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_1:11") +
						dv.Get("schedule_1:15") +
						dv.Get("schedule_1:16") +
						dv.Get("schedule_1:20") +
						dv.Get("schedule_1:21") +
						dv.Get("schedule_1:24")
				},
			},
		},
	}
}
