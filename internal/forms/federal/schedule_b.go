package federal

import (
	"taxpilot/internal/forms"
)

func init() { forms.RegisterForm(ScheduleB) }

// ScheduleB returns the FormDef for Schedule B — Interest and Ordinary Dividends.
// Part I totals interest income from all 1099-INT forms.
// Part II totals ordinary dividends from all 1099-DIV forms.
// Part III: Foreign Accounts and Trusts — required disclosures for
// taxpayers with foreign financial accounts or foreign trust interests.
func ScheduleB() *forms.FormDef {
	return &forms.FormDef{
		ID:           forms.FormScheduleB,
		Name:         "Schedule B — Interest and Ordinary Dividends",
		Jurisdiction: forms.Federal,
		TaxYears:      []int{2025},
		QuestionGroup: forms.GroupIncome1099,
		QuestionOrder: 3,
		Fields: []forms.FieldDef{
			// --- Part I: Interest ---

			// Foreign interest income not reported on a 1099-INT
			// (e.g., interest from foreign banks, pension funds, brokerage accounts)
			{
				Line:      forms.LineForeignInterest,
				Type:      forms.UserInput,
				Label:     "Foreign interest income (not on 1099-INT)",
				Prompt:    "Enter interest income from foreign banks or institutions (not reported on a 1099-INT), converted to USD:",
			},
			// Foreign interest payer description (for Schedule B listing)
			{
				Line:      forms.LineForeignInterestPayer,
				Type:      forms.UserInput,
				ValueType: forms.StringValue,
				Label:     "Foreign interest payer(s)",
				Prompt:    "Describe the foreign payer(s) of interest income (e.g., \"Nordea Bank, Sweden\"):",
			},

			// Line 1: Interest income from all 1099-INT forms + foreign interest
			{
				Line:      "1",
				Type:      forms.Computed,
				Label:     "Interest income (1099-INT + foreign sources)",
				DependsOn: []string{forms.F1099INTWildcardInterest, forms.SchedBForeignInterest},
				Compute: func(dv forms.DepValues) float64 {
					return dv.SumAll(forms.F1099INTWildcardInterest) + dv.Get(forms.SchedBForeignInterest)
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
				DependsOn: []string{forms.SchedBLine1, "schedule_b:2"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.SchedBLine1) - dv.Get("schedule_b:2")
				},
			},

			// --- Part II: Ordinary Dividends ---

			// Line 5: Ordinary dividends from all 1099-DIV forms
			{
				Line:      "5",
				Type:      forms.Computed,
				Label:     "Ordinary dividends (from 1099-DIV Box 1a)",
				DependsOn: []string{forms.F1099DIVWildcardOrdinary},
				Compute: func(dv forms.DepValues) float64 {
					return dv.SumAll(forms.F1099DIVWildcardOrdinary)
				},
			},
			// Line 6: Total ordinary dividends (Part II result)
			{
				Line:      "6",
				Type:      forms.Computed,
				Label:     "Total ordinary dividends",
				DependsOn: []string{forms.SchedBLine5},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.SchedBLine5)
				},
			},

			// --- Part III: Foreign Accounts and Trusts ---

			// Line 7a: Foreign financial accounts
			{
				Line:    "7a",
				Type:    forms.UserInput,
				ValueType: forms.StringValue,
				Label:   "Foreign financial accounts",
				Prompt:  "At any time during 2025, did you have a financial interest in or signature authority over a financial account in a foreign country (e.g., bank account, securities account)?" ,
				Options: forms.YesNoOptions,
			},
			// Line 7b: Country of foreign accounts
			{
				Line:   "7b",
				Type:   forms.UserInput,
				ValueType: forms.StringValue,
				Label:  "Country of foreign accounts",
				Prompt: "In which country or countries are the foreign accounts located?",
			},
			// Line 8: Foreign trusts
			{
				Line:    "8",
				Type:    forms.UserInput,
				ValueType: forms.StringValue,
				Label:   "Foreign trust",
				Prompt:  "Did you receive a distribution from, or were you a grantor of, or transferor to, a foreign trust?",
				Options: forms.YesNoOptions,
			},
			// FBAR required flag (computed from 7a)
			{
				Line:      "fbar_required",
				Type:      forms.Computed,
				Label:     "FBAR filing required",
				DependsOn: []string{forms.SchedBLine7a},
				Compute: func(dv forms.DepValues) float64 {
					if dv.GetString(forms.SchedBLine7a) == forms.OptionYes {
						return 1
					}
					return 0
				},
			},
		},
	}
}
