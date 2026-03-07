package inputs

import "taxpilot/internal/forms"

// F1099NEC returns the FormDef for a 1099-NEC Nonemployee Compensation.
func F1099NEC() *forms.FormDef {
	return &forms.FormDef{
		ID:           "1099nec",
		Name:         "1099-NEC Nonemployee Compensation",
		Jurisdiction: forms.Federal,
		TaxYears:     []int{2025},
		Fields: []forms.FieldDef{
			{
				Line:   "payer_name",
				Type:   forms.UserInput,
				Label:  "Payer name",
				Prompt: "What is the payer's name (from 1099-NEC)?",
			},
			{
				Line:   "payer_tin",
				Type:   forms.UserInput,
				Label:  "Payer TIN",
				Prompt: "What is the payer's TIN (XX-XXXXXXX)?",
			},
			{
				Line:   "nonemployee_compensation",
				Type:   forms.UserInput,
				Label:  "Box 1: Nonemployee compensation",
				Prompt: "Enter Box 1 — Nonemployee compensation:",
			},
			{
				Line:   "federal_tax_withheld",
				Type:   forms.UserInput,
				Label:  "Box 4: Federal income tax withheld",
				Prompt: "Enter Box 4 — Federal income tax withheld:",
			},
		},
	}
}
