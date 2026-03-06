package interview

// ContextualPrompt holds an enhanced prompt for a form field.
type ContextualPrompt struct {
	Prompt   string // the question to show the user
	HelpText string // additional context shown below the prompt
	CANote   string // CA-specific note (shown when filing in CA)
}

// contextualPrompts maps field keys to enhanced prompts.
// These are used instead of the raw FieldDef.Prompt values.
var contextualPrompts = map[string]ContextualPrompt{
	"1040:filing_status": {
		Prompt:   "What is your filing status for 2025?",
		HelpText: "Your filing status affects your tax brackets, standard deduction, and eligibility for certain credits.",
		CANote:   "California uses the same filing status as your federal return.",
	},
	"1040:first_name": {
		Prompt:   "What is your first name?",
		HelpText: "As shown on your Social Security card.",
	},
	"1040:last_name": {
		Prompt:   "What is your last name?",
		HelpText: "As shown on your Social Security card.",
	},
	"1040:ssn": {
		Prompt:   "What is your Social Security number?",
		HelpText: "Format: XXX-XX-XXXX. This is required for filing and is kept secure.",
	},
	"w2:1:employer_name": {
		Prompt:   "Who is your employer?",
		HelpText: "Enter the employer name exactly as shown on your W-2 (Box c).",
	},
	"w2:1:employer_ein": {
		Prompt:   "What is your employer's EIN?",
		HelpText: "The 9-digit Employer Identification Number from your W-2 (Box b). Format: XX-XXXXXXX.",
	},
	"w2:1:wages": {
		Prompt:   "What were your total wages from this employer?",
		HelpText: "This is Box 1 on your W-2: Wages, tips, and other compensation. This is your gross pay minus pre-tax deductions (401k, health insurance, etc.).",
		CANote:   "If your CA wages (Box 16) differ from federal wages (Box 1), we'll ask about that separately.",
	},
	"w2:1:federal_tax_withheld": {
		Prompt:   "How much federal income tax was withheld?",
		HelpText: "W-2 Box 2. This is the amount your employer sent to the IRS on your behalf throughout the year.",
	},
	"w2:1:ss_wages": {
		Prompt:   "What were your Social Security wages?",
		HelpText: "W-2 Box 3. Usually the same as Box 1, but may differ if you have pre-tax deductions that are subject to Social Security tax.",
	},
	"w2:1:ss_tax_withheld": {
		Prompt:   "How much Social Security tax was withheld?",
		HelpText: "W-2 Box 4. Should be 6.2% of Box 3 (capped at the Social Security wage base of $176,100 for 2025).",
	},
	"w2:1:medicare_wages": {
		Prompt:   "What were your Medicare wages?",
		HelpText: "W-2 Box 5. Usually the same as Box 1. There is no cap on Medicare wages.",
	},
	"w2:1:medicare_tax_withheld": {
		Prompt:   "How much Medicare tax was withheld?",
		HelpText: "W-2 Box 6. Should be 1.45% of Box 5 (plus 0.9% Additional Medicare Tax on wages over $200,000).",
	},
	"w2:1:state_wages": {
		Prompt:   "What were your state wages?",
		HelpText: "W-2 Box 16. This is your California taxable wages. Usually the same as Box 1 (federal wages).",
		CANote:   "If different from federal wages, this is typically due to items taxed differently by California.",
	},
	"w2:1:state_tax_withheld": {
		Prompt:   "How much California state tax was withheld?",
		HelpText: "W-2 Box 17. This is the amount your employer sent to the California FTB on your behalf.",
	},
}

// GetContextualPrompt returns the enhanced prompt for a field key, falling back
// to the original prompt if no contextual prompt is defined.
func GetContextualPrompt(fieldKey string, originalPrompt string, stateCode string) ContextualPrompt {
	if cp, ok := contextualPrompts[fieldKey]; ok {
		result := ContextualPrompt{
			Prompt:   cp.Prompt,
			HelpText: cp.HelpText,
		}
		// Only include CANote when filing in California
		if stateCode == "CA" {
			result.CANote = cp.CANote
		}
		return result
	}

	// Fall back to the original prompt with no extra help text
	return ContextualPrompt{
		Prompt: originalPrompt,
	}
}
