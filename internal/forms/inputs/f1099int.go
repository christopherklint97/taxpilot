package inputs

import "taxpilot/internal/forms"

// F1099INT returns the FormDef for a 1099-INT Interest Income.
func F1099INT() *forms.FormDef {
	return &forms.FormDef{
		ID:           forms.Form1099INT,
		Name:         "1099-INT Interest Income",
		Jurisdiction: forms.Federal,
		TaxYears:     []int{2025},
		Fields: []forms.FieldDef{
			{
				Line:   "payer_name",
				Type:   forms.UserInput,
				Label:  "Payer name",
				Prompt: "What is the payer's name (from 1099-INT)?",
			},
			{
				Line:   "payer_tin",
				Type:   forms.UserInput,
				Label:  "Payer TIN",
				Prompt: "What is the payer's TIN (XX-XXXXXXX)?",
			},
			{
				Line:   "interest_income",
				Type:   forms.UserInput,
				Label:  "Box 1: Interest income",
				Prompt: "Enter Box 1 — Interest income:",
			},
			{
				Line:   "early_withdrawal_penalty",
				Type:   forms.UserInput,
				Label:  "Box 2: Early withdrawal penalty",
				Prompt: "Enter Box 2 — Early withdrawal penalty (if any):",
			},
			{
				Line:   "us_savings_bond_interest",
				Type:   forms.UserInput,
				Label:  "Box 3: Interest on U.S. Savings Bonds and Treasury obligations",
				Prompt: "Enter Box 3 — Interest on U.S. Savings Bonds and Treasury obligations:",
			},
			{
				Line:   "federal_tax_withheld",
				Type:   forms.UserInput,
				Label:  "Box 4: Federal income tax withheld",
				Prompt: "Enter Box 4 — Federal income tax withheld:",
			},
			{
				Line:   "tax_exempt_interest",
				Type:   forms.UserInput,
				Label:  "Box 8: Tax-exempt interest",
				Prompt: "Enter Box 8 — Tax-exempt interest:",
			},
			{
				Line:   "private_activity_bond_interest",
				Type:   forms.UserInput,
				Label:  "Box 9: Specified private activity bond interest",
				Prompt: "Enter Box 9 — Specified private activity bond interest:",
			},
		},
	}
}
