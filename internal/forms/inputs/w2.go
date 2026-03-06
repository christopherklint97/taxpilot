package inputs

import "taxpilot/internal/forms"

// W2 returns the FormDef for a W-2 Wage and Tax Statement.
// This captures both federal and state boxes from a single W-2.
func W2() *forms.FormDef {
	return &forms.FormDef{
		ID:           "w2",
		Name:         "W-2 Wage and Tax Statement",
		Jurisdiction: forms.Federal, // W-2 is federally defined but contains state info
		TaxYears:     []int{2025},
		Fields: []forms.FieldDef{
			{
				Line:   "employer_name",
				Type:   forms.UserInput,
				Label:  "Employer name",
				Prompt: "What is the employer's name?",
			},
			{
				Line:   "employer_ein",
				Type:   forms.UserInput,
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
