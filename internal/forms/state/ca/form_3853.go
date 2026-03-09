package ca

import (
	"math"

	"taxpilot/internal/forms"
)

func init() { forms.RegisterForm(Form3853) }

// Form3853 returns the FormDef for California Form 3853 — Health Coverage
// Exemptions and Individual Shared Responsibility Penalty. This computes the
// penalty for not maintaining qualifying health coverage under California's
// individual mandate.
func Form3853() *forms.FormDef {
	return &forms.FormDef{
		ID:           forms.FormF3853,
		Name:         "Form 3853 — Health Coverage Exemptions and Individual Shared Responsibility Penalty",
		Jurisdiction: forms.StateCA,
		TaxYears:      []int{2025},
		QuestionGroup: "ca",
		QuestionOrder: 7,
		Fields: []forms.FieldDef{
			// Line 1: Full-year coverage (yes/no)
			{
				Line:    "1",
				Type:    forms.UserInput,
				Label:   "Full-year qualifying health coverage",
				Prompt:  "Did you have qualifying health coverage for all 12 months of 2025?",
				Options: []string{"yes", "no"},
			},
			// Line 2: Months without coverage (0-12)
			{
				Line:   "2",
				Type:   forms.UserInput,
				Label:  "Months without qualifying health coverage",
				Prompt: "How many months were you without qualifying health coverage?",
			},
			// Line 3: Exemption from coverage requirement (yes/no)
			{
				Line:    "3",
				Type:    forms.UserInput,
				Label:   "Exemption from health coverage requirement",
				Prompt:  "Did you have an exemption from the health coverage requirement?",
				Options: []string{"yes", "no"},
			},
			// Line 4: Applicable household income (CA AGI from Form 540)
			{
				Line:      "4",
				Type:      forms.Computed,
				Label:     "Applicable household income",
				DependsOn: []string{"ca_540:17"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("ca_540:17")
				},
			},
			// Line 5: Penalty per month
			{
				Line:      "5",
				Type:      forms.Computed,
				Label:     "Penalty per month",
				DependsOn: []string{"form_3853:1", "form_3853:3", "form_3853:4"},
				Compute: func(dv forms.DepValues) float64 {
					if dv.GetString("form_3853:1") == "yes" {
						return 0 // full coverage, no penalty
					}
					if dv.GetString("form_3853:3") == "yes" {
						return 0 // exempt from requirement
					}

					caAGI := dv.Get("form_3853:4")

					// Flat penalty: $900/month (2025 state avg bronze plan)
					flatPerMonth := 900.0

					// Income-based penalty: 2.5% of (CA AGI - filing threshold) / 12
					// Simplified CA filing threshold for 2025: ~$21,135
					filingThreshold := 21135.0
					incomeBase := caAGI - filingThreshold
					if incomeBase < 0 {
						incomeBase = 0
					}
					incomePerMonth := 0.025 * incomeBase / 12

					return math.Round(math.Max(flatPerMonth, incomePerMonth)*100) / 100
				},
			},
			// Line 6: Total penalty = months * per_month_penalty
			{
				Line:      "6",
				Type:      forms.Computed,
				Label:     "Total penalty",
				DependsOn: []string{"form_3853:1", "form_3853:2", "form_3853:3", "form_3853:5"},
				Compute: func(dv forms.DepValues) float64 {
					if dv.GetString("form_3853:1") == "yes" {
						return 0
					}
					if dv.GetString("form_3853:3") == "yes" {
						return 0
					}
					months := dv.Get("form_3853:2")
					perMonth := dv.Get("form_3853:5")
					total := months * perMonth
					// Cap at $10,800/year (12 * $900)
					return math.Min(total, 10800)
				},
			},
			// Line 7: Penalty amount that carries to Form 540
			{
				Line:      "7",
				Type:      forms.Computed,
				Label:     "Penalty to Form 540",
				DependsOn: []string{"form_3853:6"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("form_3853:6")
				},
			},
		},
	}
}
