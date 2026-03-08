package federal

import (
	"taxpilot/internal/forms"
)

// Form8833 returns the FormDef for Form 8833 — Treaty-Based Return
// Position Disclosure Under Section 6114 or 7701(b).
//
// This is primarily a disclosure form — it does not directly affect
// tax computation, but must be filed when a taxpayer takes a position
// that a tax treaty overrides or modifies an Internal Revenue Code
// provision.
//
// Common treaty positions for US-Sweden:
//   - Article 15: Employment income (sourcing rules)
//   - Article 18: Pensions (Swedish pension treatment)
//   - Article 23: Elimination of double taxation
//   - Article 24: Non-discrimination
//
// Failure to disclose carries a $1,000 penalty per position.
func Form8833() *forms.FormDef {
	return &forms.FormDef{
		ID:           forms.FormF8833,
		Name:         "Form 8833 — Treaty-Based Return Position Disclosure",
		Jurisdiction: forms.Federal,
		TaxYears:     []int{2025},
		Fields: []forms.FieldDef{
			// Treaty country
			{
				Line:   "treaty_country",
				Type:   forms.UserInput,
				Label:  "Treaty country",
				Prompt: "Which country's tax treaty are you relying on?",
			},
			// Treaty article
			{
				Line:   "treaty_article",
				Type:   forms.UserInput,
				Label:  "Treaty article number",
				Prompt: "Which article of the tax treaty applies (e.g., 'Article 18 — Pensions')?",
			},
			// IRC provision being overridden
			{
				Line:   "irc_provision",
				Type:   forms.UserInput,
				Label:  "IRC provision being overridden",
				Prompt: "Which IRC section is modified by the treaty position (e.g., 'IRC §61' or 'IRC §871')?",
			},
			// Explanation of the treaty-based position
			{
				Line:   "treaty_position_explanation",
				Type:   forms.UserInput,
				Label:  "Explanation of treaty-based position",
				Prompt: "Briefly explain your treaty-based return position:",
			},
			// Amount of income subject to treaty treatment
			{
				Line:   "treaty_amount",
				Type:   forms.UserInput,
				Label:  "Amount of income affected by treaty position (USD)",
				Prompt: "What is the amount of income subject to this treaty position (in USD)?",
			},
			// Number of treaty positions being disclosed
			{
				Line:   "num_positions",
				Type:   forms.UserInput,
				Label:  "Number of treaty positions disclosed",
				Prompt: "How many separate treaty positions are you disclosing?",
			},

			// --- Computed fields ---

			// Treaty benefit indicator (1 = treaty position taken, 0 = no)
			{
				Line:      "treaty_claimed",
				Type:      forms.Computed,
				Label:     "Treaty position claimed",
				DependsOn: []string{"form_8833:treaty_amount"},
				Compute: func(dv forms.DepValues) float64 {
					if dv.Get("form_8833:treaty_amount") > 0 {
						return 1
					}
					return 0
				},
			},
		},
	}
}
