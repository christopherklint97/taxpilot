package inputs

import "taxpilot/internal/forms"

// F1099B returns the FormDef for a 1099-B Proceeds From Broker and Barter Exchange Transactions.
// Each 1099-B represents a single sale or disposition of a security.
// Multiple sales are handled via instance prefixes (e.g., 1099b:1:, 1099b:2:).
func F1099B() *forms.FormDef {
	return &forms.FormDef{
		ID:           "1099b",
		Name:         "1099-B Proceeds From Broker Transactions",
		Jurisdiction: forms.Federal,
		TaxYears:     []int{2025},
		Fields: []forms.FieldDef{
			{
				Line:   "description",
				Type:   forms.UserInput,
				Label:  "Description of property",
				Prompt: "Describe the security sold (e.g., \"100 sh AAPL\"):",
			},
			{
				Line:   "date_acquired",
				Type:   forms.UserInput,
				Label:  "Date acquired",
				Prompt: "When did you acquire this security (MM/DD/YYYY or VARIOUS)?",
			},
			{
				Line:   "date_sold",
				Type:   forms.UserInput,
				Label:  "Date sold",
				Prompt: "When did you sell this security (MM/DD/YYYY)?",
			},
			{
				Line:   "proceeds",
				Type:   forms.UserInput,
				Label:  "Box 1d: Proceeds",
				Prompt: "Enter Box 1d — Proceeds (sale price):",
			},
			{
				Line:   "cost_basis",
				Type:   forms.UserInput,
				Label:  "Box 1e: Cost or other basis",
				Prompt: "Enter Box 1e — Cost or other basis:",
			},
			{
				Line:   "wash_sale_loss",
				Type:   forms.UserInput,
				Label:  "Box 1g: Wash sale loss disallowed",
				Prompt: "Enter Box 1g — Wash sale loss disallowed (0 if none):",
			},
			{
				Line:   "federal_tax_withheld",
				Type:   forms.UserInput,
				Label:  "Box 4: Federal income tax withheld",
				Prompt: "Enter Box 4 — Federal income tax withheld (0 if none):",
			},
			{
				Line:    "term",
				Type:    forms.UserInput,
				Label:   "Short-term or long-term",
				Prompt:  "Was this a short-term or long-term holding?",
				Options: []string{"short", "long"},
			},
			{
				Line:    "basis_reported",
				Type:    forms.UserInput,
				Label:   "Basis reported to IRS",
				Prompt:  "Was cost basis reported to the IRS by your broker?",
				Options: []string{"yes", "no"},
			},
		},
	}
}
