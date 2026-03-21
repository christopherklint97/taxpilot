package inputs

import "taxpilot/internal/forms"

func init() { forms.RegisterForm(F1099NEC) }

// F1099NEC returns the FormDef for a 1099-NEC Nonemployee Compensation.
func F1099NEC() *forms.FormDef {
	return &forms.FormDef{
		ID:           forms.Form1099NEC,
		Name:         "1099-NEC Nonemployee Compensation",
		Jurisdiction: forms.Federal,
		TaxYears:      []int{2025},
		QuestionGroup: forms.GroupIncome1099,
		QuestionOrder: 3,
		Fields: []forms.FieldDef{
			{
				Line:   forms.LinePayerName,
				Type:   forms.UserInput,
				ValueType: forms.StringValue,
				Label:  "Payer name",
				Prompt: "What is the payer's name (from 1099-NEC)?",
			},
			{
				Line:   forms.LinePayerTIN,
				Type:   forms.UserInput,
				ValueType: forms.StringValue,
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
