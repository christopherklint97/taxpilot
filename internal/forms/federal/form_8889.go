package federal

import (
	"math"

	"taxpilot/internal/forms"
)

func init() { forms.RegisterForm(Form8889) }

// HSA contribution limits for 2025 (IRS Rev. Proc. 2024-25)
const (
	hsaSelfOnly2025 = 4300  // Self-only coverage
	hsaFamily2025   = 8550  // Family coverage
	hsaCatchUp      = 1000  // Additional for age 55+
	hsaPenaltyRate  = 0.20  // 20% penalty on non-qualified distributions
)

// Form8889 returns the FormDef for Form 8889 — Health Savings Accounts.
// This computes the HSA deduction (Part I) and taxable distributions (Part II).
// The deduction flows to Schedule 1 line 15 → Form 1040 line 10.
//
// CA does NOT conform to federal HSA treatment — contributions are not
// deductible for CA, requiring an add-back on Schedule CA.
func Form8889() *forms.FormDef {
	return &forms.FormDef{
		ID:           forms.FormF8889,
		Name:         "Form 8889 — Health Savings Accounts",
		Jurisdiction: forms.Federal,
		TaxYears:      []int{2025},
		QuestionGroup: "deductions",
		QuestionOrder: 6,
		Fields: []forms.FieldDef{
			// --- Part I: HSA Contributions and Deduction ---

			// Line 1: Coverage type (self-only or family)
			{
				Line:    "1",
				Type:    forms.UserInput,
				Label:   "HSA coverage type",
				Prompt:  "What type of HDHP coverage do you have?",
				Options: []string{"self-only", "family"},
			},
			// Line 2: HSA contributions you made for 2025
			{
				Line:   "2",
				Type:   forms.UserInput,
				Label:  "HSA contributions made for 2025",
				Prompt: "How much did you contribute to your HSA for 2025?",
			},
			// Line 3: Employer contributions (including pre-tax payroll)
			{
				Line:   "3",
				Type:   forms.UserInput,
				Label:  "Employer contributions (including pre-tax payroll)",
				Prompt: "How much did your employer contribute to your HSA (W-2 Box 12, code W)?",
			},
			// Line 5: Age 55+ catch-up contribution
			{
				Line:   "5",
				Type:   forms.UserInput,
				Label:  "Additional catch-up contribution (age 55+)",
				Prompt: "Are you age 55 or older? Enter catch-up contribution amount ($1,000 max, or 0):",
			},
			// Line 6: HSA deduction limit
			{
				Line:      "6",
				Type:      forms.Computed,
				Label:     "HSA contribution limit",
				DependsOn: []string{forms.F8889Line1, forms.F8889Line5},
				Compute: func(dv forms.DepValues) float64 {
					coverageType := dv.GetString(forms.F8889Line1)
					var limit float64
					if coverageType == "family" {
						limit = hsaFamily2025
					} else {
						limit = hsaSelfOnly2025
					}
					catchUp := math.Min(dv.Get(forms.F8889Line5), hsaCatchUp)
					return limit + catchUp
				},
			},
			// Line 9: HSA deduction (contributions minus employer, limited by line 6)
			{
				Line:      "9",
				Type:      forms.Computed,
				Label:     "HSA deduction",
				DependsOn: []string{forms.F8889Line2, forms.F8889Line3, forms.F8889Line6},
				Compute: func(dv forms.DepValues) float64 {
					contributions := dv.Get(forms.F8889Line2)
					employer := dv.Get(forms.F8889Line3)
					limit := dv.Get(forms.F8889Line6)

					// Total contributions (yours + employer) can't exceed limit
					total := contributions + employer
					if total > limit {
						// Your deduction is the limit minus employer portion
						return math.Max(0, limit-employer)
					}
					return math.Max(0, contributions)
				},
			},

			// --- Part II: HSA Distributions ---

			// Line 14a: Total HSA distributions received
			{
				Line:   "14a",
				Type:   forms.UserInput,
				Label:  "Total HSA distributions",
				Prompt: "How much did you receive in HSA distributions in 2025?",
			},
			// Line 14c: Qualified medical expenses paid with HSA
			{
				Line:   "14c",
				Type:   forms.UserInput,
				Label:  "Qualified medical expenses paid from HSA",
				Prompt: "How much of your HSA distributions were for qualified medical expenses?",
			},
			// Line 15: Taxable HSA distributions
			{
				Line:      "15",
				Type:      forms.Computed,
				Label:     "Taxable HSA distributions",
				DependsOn: []string{forms.F8889Line14a, forms.F8889Line14c},
				Compute: func(dv forms.DepValues) float64 {
					return math.Max(0, dv.Get(forms.F8889Line14a)-dv.Get(forms.F8889Line14c))
				},
			},
			// Line 17b: Additional 20% tax on non-qualified distributions
			{
				Line:      "17b",
				Type:      forms.Computed,
				Label:     "Additional tax on non-qualified HSA distributions (20%)",
				DependsOn: []string{forms.F8889Line15},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.F8889Line15) * hsaPenaltyRate
				},
			},
		},
	}
}
