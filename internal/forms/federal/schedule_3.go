package federal

import (
	"taxpilot/internal/forms"
)

// Schedule3 returns the FormDef for Schedule 3 — Additional Credits and Payments.
//
// Part I: Nonrefundable Credits (deferred — education, child, etc.)
// Part II: Other Payments and Refundable Credits
//   - Line 10: Estimated tax payments -> flows to 1040 line 26
//   - Line 15: Total other payments -> flows to 1040 line 31
func Schedule3() *forms.FormDef {
	return &forms.FormDef{
		ID:           forms.FormSchedule3,
		Name:         "Schedule 3 — Additional Credits and Payments",
		Jurisdiction: forms.Federal,
		TaxYears:     []int{2025},
		Fields: []forms.FieldDef{
			// --- Part I: Nonrefundable Credits ---

			// Line 1: Foreign tax credit (from Form 1116 line 22)
			{
				Line:      "1",
				Type:      forms.Computed,
				Label:     "Foreign tax credit",
				DependsOn: []string{"form_1116:22"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("form_1116:22")
				},
			},
			// Line 2: Child and dependent care credit (deferred)
			{
				Line:      "2",
				Type:      forms.Computed,
				Label:     "Child and dependent care credit",
				DependsOn: []string{},
				Compute: func(dv forms.DepValues) float64 {
					return 0
				},
			},
			// Line 3: Education credits (deferred)
			{
				Line:      "3",
				Type:      forms.Computed,
				Label:     "Education credits",
				DependsOn: []string{},
				Compute: func(dv forms.DepValues) float64 {
					return 0
				},
			},
			// Line 8: Total nonrefundable credits
			{
				Line:      "8",
				Type:      forms.Computed,
				Label:     "Total nonrefundable credits",
				DependsOn: []string{"schedule_3:1", "schedule_3:2", "schedule_3:3"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_3:1") +
						dv.Get("schedule_3:2") +
						dv.Get("schedule_3:3")
				},
			},

			// --- Part II: Other Payments and Refundable Credits ---

			// Line 10: Estimated tax payments (UserInput — taxpayer enters total paid)
			{
				Line:   "10",
				Type:   forms.UserInput,
				Label:  "Estimated tax payments for 2025",
				Prompt: "How much did you pay in federal estimated taxes for 2025?",
			},
			// Line 15: Total other payments and refundable credits
			{
				Line:      "15",
				Type:      forms.Computed,
				Label:     "Total other payments",
				DependsOn: []string{"schedule_3:10"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_3:10")
				},
			},
		},
	}
}
