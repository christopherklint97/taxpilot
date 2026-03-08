package federal

import (
	"math"

	"taxpilot/internal/forms"
)

// Form1116 returns the FormDef for Form 1116 — Foreign Tax Credit.
// This computes the credit for income taxes paid to foreign governments.
//
// Key rule: A taxpayer cannot claim both FEIE (Form 2555) and FTC on the
// same income. The foreign_source_income for FTC purposes must exclude
// any income already excluded via Form 2555.
//
// The credit flows to Schedule 3 line 1 → 1040 line 20 (nonrefundable credits).
//
// For taxes paid ≤ $300 ($600 MFJ), the simplified election allows
// claiming the credit without Form 1116. This form handles the full
// computation for any amount.
func Form1116() *forms.FormDef {
	return &forms.FormDef{
		ID:           forms.FormF1116,
		Name:         "Form 1116 — Foreign Tax Credit",
		Jurisdiction: forms.Federal,
		TaxYears:     []int{2025},
		Fields: []forms.FieldDef{
			// --- Part I: Taxable Income from Sources Outside the US ---

			// Income category
			{
				Line:    "category",
				Type:    forms.UserInput,
				Label:   "Foreign tax credit category",
				Prompt:  "What category of foreign income are you claiming the credit for?",
				Options: []string{"general", "passive", "section_901j", "treaty_sourced"},
			},
			// Foreign country
			{
				Line:   "foreign_country",
				Type:   forms.UserInput,
				Label:  "Country where tax was paid",
				Prompt: "Which country did you pay foreign taxes to?",
			},
			// Gross foreign source income (NOT excluded by FEIE)
			{
				Line:   "foreign_source_income",
				Type:   forms.UserInput,
				Label:  "Gross foreign source income (not excluded by FEIE)",
				Prompt: "What is your gross foreign source income NOT excluded by Form 2555?",
			},
			// Deductions allocated to foreign source income
			{
				Line:   "foreign_source_deductions",
				Type:   forms.UserInput,
				Label:  "Deductions allocated to foreign source income",
				Prompt: "What deductions are definitely allocable to your foreign source income?",
			},
			// Foreign taxes paid — income taxes
			{
				Line:   "foreign_tax_paid_income",
				Type:   forms.UserInput,
				Label:  "Foreign income taxes paid or accrued",
				Prompt: "How much foreign income tax did you pay or accrue (converted to USD)?",
			},
			// Foreign taxes paid — other
			{
				Line:   "foreign_tax_paid_other",
				Type:   forms.UserInput,
				Label:  "Other foreign taxes paid",
				Prompt: "How much in other qualifying foreign taxes did you pay (e.g., war profits tax)?",
			},
			// Paid or accrued
			{
				Line:    "accrued_or_paid",
				Type:    forms.UserInput,
				Label:   "Taxes paid or accrued",
				Prompt:  "Are you claiming foreign taxes on a paid or accrued basis?",
				Options: []string{"paid", "accrued"},
			},

			// --- Computed Fields ---

			// Line 7: Net foreign source taxable income
			{
				Line:      "7",
				Type:      forms.Computed,
				Label:     "Net foreign source taxable income",
				DependsOn: []string{"form_1116:foreign_source_income", "form_1116:foreign_source_deductions"},
				Compute: func(dv forms.DepValues) float64 {
					return math.Max(0,
						dv.Get("form_1116:foreign_source_income")-
							dv.Get("form_1116:foreign_source_deductions"))
				},
			},
			// Line 15: Total foreign taxes paid or accrued
			{
				Line:      "15",
				Type:      forms.Computed,
				Label:     "Total foreign taxes paid or accrued",
				DependsOn: []string{"form_1116:foreign_tax_paid_income", "form_1116:foreign_tax_paid_other"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("form_1116:foreign_tax_paid_income") +
						dv.Get("form_1116:foreign_tax_paid_other")
				},
			},
			// Line 20: US tax on worldwide income (from 1040 line 16)
			{
				Line:      "20",
				Type:      forms.Computed,
				Label:     "US tax liability",
				DependsOn: []string{"1040:16"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("1040:16")
				},
			},
			// Line 21: Foreign tax credit limitation
			// limitation = US_tax * (foreign_source_income / worldwide_income)
			{
				Line:      "21",
				Type:      forms.Computed,
				Label:     "Foreign tax credit limitation",
				DependsOn: []string{"form_1116:20", "form_1116:7", "1040:15"},
				Compute: func(dv forms.DepValues) float64 {
					usTax := dv.Get("form_1116:20")
					foreignSource := dv.Get("form_1116:7")
					worldwideTaxable := dv.Get("1040:15")

					if worldwideTaxable <= 0 || usTax <= 0 {
						return 0
					}

					// Limitation ratio cannot exceed 1.0
					ratio := math.Min(1.0, foreignSource/worldwideTaxable)
					return usTax * ratio
				},
			},
			// Line 22: Credit allowed (lesser of taxes paid or limitation)
			{
				Line:      "22",
				Type:      forms.Computed,
				Label:     "Foreign tax credit allowed",
				DependsOn: []string{"form_1116:15", "form_1116:21"},
				Compute: func(dv forms.DepValues) float64 {
					taxesPaid := dv.Get("form_1116:15")
					limitation := dv.Get("form_1116:21")
					return math.Min(taxesPaid, limitation)
				},
			},
			// Carryforward: excess foreign tax that can be carried to future years
			{
				Line:      "carryforward",
				Type:      forms.Computed,
				Label:     "Foreign tax credit carryforward",
				DependsOn: []string{"form_1116:15", "form_1116:21"},
				Compute: func(dv forms.DepValues) float64 {
					taxesPaid := dv.Get("form_1116:15")
					limitation := dv.Get("form_1116:21")
					return math.Max(0, taxesPaid-limitation)
				},
			},
		},
	}
}
