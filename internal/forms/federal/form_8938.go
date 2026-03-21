package federal

import (
	"taxpilot/internal/forms"
	"taxpilot/pkg/taxmath"
)

func init() { forms.RegisterForm(Form8938) }

// Form8938 returns the FormDef for Form 8938 — Statement of Specified
// Foreign Financial Assets (FATCA).
//
// Required when the total value of specified foreign financial assets
// exceeds the applicable threshold. Thresholds differ based on filing
// status and whether the taxpayer lives abroad or in the US.
//
// This form is informational — it does not affect tax computation but
// failure to file carries significant penalties ($10,000+ per year).
func Form8938() *forms.FormDef {
	return &forms.FormDef{
		ID:           forms.FormF8938,
		Name:         "Form 8938 — Statement of Specified Foreign Financial Assets",
		Jurisdiction: forms.Federal,
		TaxYears:      []int{2025},
		QuestionGroup: forms.GroupExpat,
		QuestionOrder: 4,
		Fields: []forms.FieldDef{
			// --- Taxpayer situation ---

			// Lives abroad (determines which threshold applies)
			{
				Line:    "lives_abroad",
				Type:    forms.UserInput,
				ValueType: forms.StringValue,
				Label:   "Do you live outside the United States?",
				Prompt:  "Do you meet the bona fide residence or physical presence test for living abroad?",
				Options: forms.YesNoOptions,
			},

			// --- Financial account details ---

			// Number of foreign financial accounts
			{
				Line:   "num_accounts",
				Type:   forms.UserInput,
				Label:  "Number of foreign financial accounts",
				Prompt: "How many foreign financial accounts do you have (bank, brokerage, pension, etc.)?",
			},
			// Maximum value of all accounts during the year
			{
				Line:   "max_value_accounts",
				Type:   forms.UserInput,
				Label:  "Maximum value of foreign accounts during the year (USD)",
				Prompt: "What was the maximum aggregate value of all your foreign financial accounts at any time during 2025 (in USD)?",
			},
			// Year-end value of all accounts
			{
				Line:   "yearend_value_accounts",
				Type:   forms.UserInput,
				Label:  "Year-end value of foreign accounts (USD)",
				Prompt: "What was the total value of all your foreign financial accounts on December 31, 2025 (in USD)?",
			},

			// --- Other specified foreign financial assets ---

			// Number of other foreign assets
			{
				Line:   "num_other_assets",
				Type:   forms.UserInput,
				Label:  "Number of other specified foreign financial assets",
				Prompt: "How many other specified foreign financial assets do you have (foreign stocks, partnership interests, etc.)?",
			},
			// Maximum value of other assets
			{
				Line:   "max_value_other",
				Type:   forms.UserInput,
				Label:  "Maximum value of other foreign assets during the year (USD)",
				Prompt: "What was the maximum aggregate value of your other specified foreign financial assets at any time during 2025 (in USD)?",
			},
			// Year-end value of other assets
			{
				Line:   "yearend_value_other",
				Type:   forms.UserInput,
				Label:  "Year-end value of other foreign assets (USD)",
				Prompt: "What was the total value of your other specified foreign financial assets on December 31, 2025 (in USD)?",
			},

			// --- Account identification ---

			// Country of primary account
			{
				Line:   "account_country",
				Type:   forms.UserInput,
				ValueType: forms.StringValue,
				Label:  "Country of foreign accounts",
				Prompt: "In which country are your foreign financial accounts held?",
			},
			// Institution name
			{
				Line:   "account_institution",
				Type:   forms.UserInput,
				ValueType: forms.StringValue,
				Label:  "Foreign financial institution name",
				Prompt: "What is the name of the foreign financial institution?",
			},
			// Account type
			{
				Line:    "account_type",
				Type:    forms.UserInput,
				ValueType: forms.StringValue,
				Label:   "Type of foreign account",
				Prompt:  "What type of foreign financial account is it?",
				Options: []string{"deposit", "custodial", "pension", "other"},
			},

			// --- Income from foreign assets ---

			// Income reported from accounts
			{
				Line:   "income_from_accounts",
				Type:   forms.UserInput,
				Label:  "Income from foreign financial accounts (USD)",
				Prompt: "How much income was earned from your foreign financial accounts in 2025 (interest, dividends, etc. in USD)?",
			},
			// Gains from foreign assets
			{
				Line:   "gain_from_accounts",
				Type:   forms.UserInput,
				Label:  "Gains from foreign financial assets (USD)",
				Prompt: "How much in gains (or losses) did you realize from your foreign financial assets in 2025 (in USD)?",
			},

			// --- Computed fields ---

			// Total maximum value (accounts + other assets)
			{
				Line:      "total_max_value",
				Type:      forms.Computed,
				Label:     "Total maximum value of all foreign assets",
				DependsOn: []string{forms.F8938MaxValueAccounts, forms.F8938MaxValueOther},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.F8938MaxValueAccounts) +
						dv.Get(forms.F8938MaxValueOther)
				},
			},
			// Total year-end value
			{
				Line:      "total_yearend_value",
				Type:      forms.Computed,
				Label:     "Total year-end value of all foreign assets",
				DependsOn: []string{forms.F8938YearEndAccounts, forms.F8938YearEndOther},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.F8938YearEndAccounts) +
						dv.Get(forms.F8938YearEndOther)
				},
			},
			// Year-end threshold (depends on filing status and whether abroad)
			{
				Line:      "threshold_yearend",
				Type:      forms.Computed,
				Label:     "FATCA year-end filing threshold",
				DependsOn: []string{forms.F8938LivesAbroad, forms.F1040FilingStatus},
				Compute: func(dv forms.DepValues) float64 {
					cfg := taxmath.GetConfigOrDefault(dv.TaxYear())
					abroad := dv.GetString(forms.F8938LivesAbroad) == forms.OptionYes
					fs := dv.GetString(forms.F1040FilingStatus)
					mfj := fs == forms.FilingMFJ

					if abroad {
						if mfj {
							return cfg.FATCAAbroadMFJYearEnd
						}
						return cfg.FATCAAbroadSingleYearEnd
					}
					if mfj {
						return cfg.FATCAUSMFJYearEnd
					}
					return cfg.FATCAUSSingleYearEnd
				},
			},
			// Any-time threshold
			{
				Line:      "threshold_anytime",
				Type:      forms.Computed,
				Label:     "FATCA any-time filing threshold",
				DependsOn: []string{forms.F8938LivesAbroad, forms.F1040FilingStatus},
				Compute: func(dv forms.DepValues) float64 {
					cfg := taxmath.GetConfigOrDefault(dv.TaxYear())
					abroad := dv.GetString(forms.F8938LivesAbroad) == forms.OptionYes
					fs := dv.GetString(forms.F1040FilingStatus)
					mfj := fs == forms.FilingMFJ

					if abroad {
						if mfj {
							return cfg.FATCAAbroadMFJAnyTime
						}
						return cfg.FATCAAbroadSingleAnyTime
					}
					if mfj {
						return cfg.FATCAUSMFJAnyTime
					}
					return cfg.FATCAUSSingleAnyTime
				},
			},
			// Filing required (1 = yes, 0 = no)
			{
				Line:      "filing_required",
				Type:      forms.Computed,
				Label:     "Form 8938 filing required",
				DependsOn: []string{forms.F8938TotalMaxValue, forms.F8938TotalYearEndValue, "form_8938:threshold_yearend", "form_8938:threshold_anytime"},
				Compute: func(dv forms.DepValues) float64 {
					maxVal := dv.Get(forms.F8938TotalMaxValue)
					yearEnd := dv.Get(forms.F8938TotalYearEndValue)
					threshYE := dv.Get("form_8938:threshold_yearend")
					threshAT := dv.Get("form_8938:threshold_anytime")

					if yearEnd > threshYE || maxVal > threshAT {
						return 1
					}
					return 0
				},
			},
		},
	}
}
