package federal

import (
	"taxpilot/internal/forms"
)

// ScheduleB returns the FormDef for Schedule B — Interest and Ordinary Dividends.
// Part I totals interest income from all 1099-INT forms.
// Part II totals ordinary dividends from all 1099-DIV forms.
// Part III (Foreign Accounts) is a yes/no question deferred to a later phase.
func ScheduleB() *forms.FormDef {
	return &forms.FormDef{
		ID:           "schedule_b",
		Name:         "Schedule B — Interest and Ordinary Dividends",
		Jurisdiction: forms.Federal,
		TaxYears:     []int{2025},
		Fields: []forms.FieldDef{
			// --- Part I: Interest ---

			// Line 1: Interest income from all 1099-INT forms
			{
				Line:      "1",
				Type:      forms.Computed,
				Label:     "Interest income (from 1099-INT Box 1)",
				DependsOn: []string{"1099int:*:interest_income"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.SumAll("1099int:*:interest_income")
				},
			},
			// Line 2: Excludable interest on Series EE/I savings bonds (0 for now)
			{
				Line:      "2",
				Type:      forms.Computed,
				Label:     "Excludable savings bond interest",
				DependsOn: []string{},
				Compute: func(dv forms.DepValues) float64 {
					return 0
				},
			},
			// Line 4: Total interest (Part I result)
			{
				Line:      "4",
				Type:      forms.Computed,
				Label:     "Total interest",
				DependsOn: []string{"schedule_b:1", "schedule_b:2"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_b:1") - dv.Get("schedule_b:2")
				},
			},

			// --- Part II: Ordinary Dividends ---

			// Line 5: Ordinary dividends from all 1099-DIV forms
			{
				Line:      "5",
				Type:      forms.Computed,
				Label:     "Ordinary dividends (from 1099-DIV Box 1a)",
				DependsOn: []string{"1099div:*:ordinary_dividends"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.SumAll("1099div:*:ordinary_dividends")
				},
			},
			// Line 6: Total ordinary dividends (Part II result)
			{
				Line:      "6",
				Type:      forms.Computed,
				Label:     "Total ordinary dividends",
				DependsOn: []string{"schedule_b:5"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_b:5")
				},
			},
		},
	}
}
