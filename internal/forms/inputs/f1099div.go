package inputs

import "taxpilot/internal/forms"

// F1099DIV returns the FormDef for a 1099-DIV Dividends and Distributions.
func F1099DIV() *forms.FormDef {
	return &forms.FormDef{
		ID:           forms.Form1099DIV,
		Name:         "1099-DIV Dividends and Distributions",
		Jurisdiction: forms.Federal,
		TaxYears:     []int{2025},
		Fields: []forms.FieldDef{
			{
				Line:   "payer_name",
				Type:   forms.UserInput,
				Label:  "Payer name",
				Prompt: "What is the payer's name (from 1099-DIV)?",
			},
			{
				Line:   "payer_tin",
				Type:   forms.UserInput,
				Label:  "Payer TIN",
				Prompt: "What is the payer's TIN (XX-XXXXXXX)?",
			},
			{
				Line:   "ordinary_dividends",
				Type:   forms.UserInput,
				Label:  "Box 1a: Total ordinary dividends",
				Prompt: "Enter Box 1a — Total ordinary dividends:",
			},
			{
				Line:   "qualified_dividends",
				Type:   forms.UserInput,
				Label:  "Box 1b: Qualified dividends",
				Prompt: "Enter Box 1b — Qualified dividends:",
			},
			{
				Line:   "total_capital_gain",
				Type:   forms.UserInput,
				Label:  "Box 2a: Total capital gain distributions",
				Prompt: "Enter Box 2a — Total capital gain distributions:",
			},
			{
				Line:   "section_1250_gain",
				Type:   forms.UserInput,
				Label:  "Box 2b: Unrecaptured Section 1250 gain",
				Prompt: "Enter Box 2b — Unrecaptured Section 1250 gain:",
			},
			{
				Line:   "section_199a_dividends",
				Type:   forms.UserInput,
				Label:  "Box 5: Section 199A dividends",
				Prompt: "Enter Box 5 — Section 199A dividends:",
			},
			{
				Line:   "federal_tax_withheld",
				Type:   forms.UserInput,
				Label:  "Box 4: Federal income tax withheld",
				Prompt: "Enter Box 4 — Federal income tax withheld:",
			},
			{
				Line:   "exempt_interest_dividends",
				Type:   forms.UserInput,
				Label:  "Box 12: Exempt-interest dividends",
				Prompt: "Enter Box 12 — Exempt-interest dividends:",
			},
			{
				Line:   "private_activity_bond_dividends",
				Type:   forms.UserInput,
				Label:  "Box 13: Specified private activity bond interest dividends",
				Prompt: "Enter Box 13 — Specified private activity bond interest dividends:",
			},
		},
	}
}
