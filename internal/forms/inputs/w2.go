package inputs

import "taxpilot/internal/forms"

func init() { forms.RegisterForm(W2) }

// W2 returns the FormDef for a W-2 Wage and Tax Statement.
// This captures both federal and state boxes from a single W-2.
// W-2 forms are issued by US employers only. Foreign employers do not issue W-2s;
// foreign wages are entered separately on Form 1040.
func W2() *forms.FormDef {
	return &forms.FormDef{
		ID:           forms.FormW2,
		Name:         "W-2 Wage and Tax Statement (US employers only)",
		Jurisdiction: forms.Federal, // W-2 is federally defined but contains state info
		TaxYears:      []int{2025},
		QuestionGroup: forms.GroupIncomeW2,
		QuestionOrder: 2,
		Fields: []forms.FieldDef{
			{
				Line:   forms.LineEmployerName,
				Type:   forms.UserInput,
				ValueType: forms.StringValue,
				Label:  "Employer name",
				Prompt: "What is the US employer's name? (Skip this form if your employer is foreign — foreign wages are entered separately)",
			},
			{
				Line:   forms.LineEmployerEIN,
				Type:   forms.UserInput,
				ValueType: forms.StringValue,
				Label:  "Employer EIN",
				Prompt: "What is the employer's EIN (XX-XXXXXXX)?",
			},
			{
				Line:   "wages",
				Type:   forms.UserInput,
				Label:  "Box 1: Wages, tips, other compensation",
				Prompt: "Enter Box 1 — Wages, tips, other compensation:",
			},
			{
				Line:   "federal_tax_withheld",
				Type:   forms.UserInput,
				Label:  "Box 2: Federal income tax withheld",
				Prompt: "Enter Box 2 — Federal income tax withheld:",
			},
			{
				Line:   "ss_wages",
				Type:   forms.UserInput,
				Label:  "Box 3: Social security wages",
				Prompt: "Enter Box 3 — Social security wages:",
			},
			{
				Line:   "ss_tax_withheld",
				Type:   forms.UserInput,
				Label:  "Box 4: Social security tax withheld",
				Prompt: "Enter Box 4 — Social security tax withheld:",
			},
			{
				Line:   "medicare_wages",
				Type:   forms.UserInput,
				Label:  "Box 5: Medicare wages and tips",
				Prompt: "Enter Box 5 — Medicare wages and tips:",
			},
			{
				Line:   "medicare_tax_withheld",
				Type:   forms.UserInput,
				Label:  "Box 6: Medicare tax withheld",
				Prompt: "Enter Box 6 — Medicare tax withheld:",
			},
			{
				Line:   "state_wages",
				Type:   forms.UserInput,
				Label:  "Box 16: State wages, tips, etc.",
				Prompt: "Enter Box 16 — State wages, tips, etc.:",
			},
			{
				Line:   "state_tax_withheld",
				Type:   forms.UserInput,
				Label:  "Box 17: State income tax withheld",
				Prompt: "Enter Box 17 — State income tax withheld:",
			},
		},
	}
}
