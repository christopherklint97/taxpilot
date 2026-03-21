package federal

import (
	"math"

	"taxpilot/internal/forms"
)

func init() { forms.RegisterForm(ScheduleSE) }

// Self-employment tax constants for 2025
const (
	ssTaxRate2025          = 0.124  // Social Security portion (12.4%)
	medicareTaxRate2025    = 0.029  // Medicare portion (2.9%)
	seTaxRate2025          = 0.9235 // 92.35% of net SE earnings are taxable
	ssWageBase2025         = 176100 // Social Security wage base for 2025
	additionalMedicareRate = 0.009  // Additional Medicare Tax rate (0.9%)
	additionalMedicareBase = 200000 // Threshold for Additional Medicare Tax (single)
)

// ScheduleSE returns the FormDef for Schedule SE — Self-Employment Tax.
// This computes the self-employment tax (Social Security + Medicare) on
// net earnings from self-employment (Schedule C line 31).
func ScheduleSE() *forms.FormDef {
	return &forms.FormDef{
		ID:           forms.FormScheduleSE,
		Name:         "Schedule SE — Self-Employment Tax",
		Jurisdiction: forms.Federal,
		TaxYears:      []int{2025},
		QuestionGroup: "business",
		QuestionOrder: 5,
		Fields: []forms.FieldDef{
			// Line 2: Net earnings from self-employment (from Schedule C)
			forms.RefField("2", "Net earnings from self-employment", forms.SchedCLine31),
			// Line 3: 92.35% of line 2 (if > $400)
			{
				Line:      "3",
				Type:      forms.Computed,
				Label:     "Self-employment earnings subject to tax",
				DependsOn: []string{forms.SchedSELine2},
				Compute: func(dv forms.DepValues) float64 {
					net := dv.Get(forms.SchedSELine2)
					if net < 400 {
						return 0 // SE tax only applies if >= $400
					}
					return net * seTaxRate2025
				},
			},
			// Line 4: Social Security tax portion
			// 12.4% on earnings up to the wage base ($176,100 for 2025),
			// reduced by W-2 Social Security wages.
			{
				Line:      "4",
				Type:      forms.Computed,
				Label:     "Social Security tax",
				DependsOn: []string{forms.SchedSELine3, forms.W2WildcardSSWages},
				Compute: func(dv forms.DepValues) float64 {
					seEarnings := dv.Get(forms.SchedSELine3)
					if seEarnings <= 0 {
						return 0
					}
					w2SSWages := dv.SumAll(forms.W2WildcardSSWages)
					remainingBase := math.Max(0, ssWageBase2025-w2SSWages)
					taxableForSS := math.Min(seEarnings, remainingBase)
					return taxableForSS * ssTaxRate2025
				},
			},
			// Line 5: Medicare tax portion (2.9% on all SE earnings, no cap)
			{
				Line:      "5",
				Type:      forms.Computed,
				Label:     "Medicare tax",
				DependsOn: []string{forms.SchedSELine3},
				Compute: func(dv forms.DepValues) float64 {
					seEarnings := dv.Get(forms.SchedSELine3)
					if seEarnings <= 0 {
						return 0
					}
					return seEarnings * medicareTaxRate2025
				},
			},
			// Line 6: Total self-employment tax
			forms.SumField("6", "Self-employment tax", forms.SchedSELine4, forms.SchedSELine5),
			// Line 7: Deductible part (50% of SE tax — goes to Schedule 1 line 16)
			{
				Line:      "7",
				Type:      forms.Computed,
				Label:     "Deductible part of self-employment tax",
				DependsOn: []string{forms.SchedSELine6},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.SchedSELine6) * 0.5
				},
			},
		},
	}
}
