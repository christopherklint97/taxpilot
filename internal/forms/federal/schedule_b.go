package federal

import (
	"taxpilot/internal/forms"
)

// ScheduleB returns the FormDef for Schedule B — Interest and Ordinary Dividends.
// Part I totals interest income from all 1099-INT forms.
// Part II totals ordinary dividends from all 1099-DIV forms.
// Part III: Foreign Accounts and Trusts — required disclosures for
// taxpayers with foreign financial accounts or foreign trust interests.
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

			// --- Part III: Foreign Accounts and Trusts ---

			// Line 7a: Foreign financial accounts
			{
				Line:    "7a",
				Type:    forms.UserInput,
				Label:   "Foreign financial accounts",
				Prompt:  "At any time during 2025, did you have a financial interest in or signature authority over a financial account in a foreign country (e.g., bank account, securities account)?" ,
				Options: []string{"yes", "no"},
			},
			// Line 7b: Country of foreign accounts
			{
				Line:   "7b",
				Type:   forms.UserInput,
				Label:  "Country of foreign accounts",
				Prompt: "In which country or countries are the foreign accounts located?",
			},
			// Line 8: Foreign trusts
			{
				Line:    "8",
				Type:    forms.UserInput,
				Label:   "Foreign trust",
				Prompt:  "Did you receive a distribution from, or were you a grantor of, or transferor to, a foreign trust?",
				Options: []string{"yes", "no"},
			},
			// FBAR required flag (computed from 7a)
			{
				Line:      "fbar_required",
				Type:      forms.Computed,
				Label:     "FBAR filing required",
				DependsOn: []string{"schedule_b:7a"},
				Compute: func(dv forms.DepValues) float64 {
					if dv.GetString("schedule_b:7a") == "yes" {
						return 1
					}
					return 0
				},
			},
		},
	}
}
