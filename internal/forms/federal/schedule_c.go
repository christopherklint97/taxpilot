package federal

import (
	"math"

	"taxpilot/internal/forms"
)

func init() { forms.RegisterForm(ScheduleC) }

// ScheduleC returns the FormDef for Schedule C — Profit or Loss From Business.
// This is a simplified version covering basic sole proprietor income from
// 1099-NEC with user-entered expenses. Full Schedule C with detailed expense
// categories will be expanded in future iterations.
func ScheduleC() *forms.FormDef {
	return &forms.FormDef{
		ID:           forms.FormScheduleC,
		Name:         "Schedule C — Profit or Loss From Business",
		Jurisdiction: forms.Federal,
		TaxYears:      []int{2025},
		QuestionGroup: "business",
		QuestionOrder: 5,
		Fields: []forms.FieldDef{
			// --- Business Info ---
			{
				Line:   "business_name",
				Type:   forms.UserInput,
				ValueType: forms.StringValue,
				Label:  "Business name",
				Prompt: "What is your business name (or your name if sole proprietor)?",
			},
			{
				Line:    "business_code",
				Type:    forms.UserInput,
				ValueType: forms.StringValue,
				Label:   "Principal business code",
				Prompt:  "What is your principal business activity code (6-digit NAICS)?",
			},

			// --- Income ---

			// Line 1: Gross receipts (from 1099-NEC + other business income)
			{
				Line:      "1",
				Type:      forms.Computed,
				Label:     "Gross receipts or sales",
				DependsOn: []string{forms.F1099NECWildcardComp},
				Compute: func(dv forms.DepValues) float64 {
					return dv.SumAll(forms.F1099NECWildcardComp)
				},
			},
			// Line 4: Cost of goods sold (deferred)
			{
				Line:      "4",
				Type:      forms.Computed,
				Label:     "Cost of goods sold",
				DependsOn: []string{},
				Compute: func(dv forms.DepValues) float64 {
					return 0
				},
			},
			// Line 5: Gross profit
			{
				Line:      "5",
				Type:      forms.Computed,
				Label:     "Gross profit",
				DependsOn: []string{forms.SchedCLine1, "schedule_c:4"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.SchedCLine1) - dv.Get("schedule_c:4")
				},
			},
			// Line 7: Gross income
			{
				Line:      "7",
				Type:      forms.Computed,
				Label:     "Gross income",
				DependsOn: []string{forms.SchedCLine5},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.SchedCLine5)
				},
			},

			// --- Expenses (simplified) ---

			// Line 8: Advertising
			{
				Line:   "8",
				Type:   forms.UserInput,
				Label:  "Advertising expenses",
				Prompt: "Enter advertising expenses:",
			},
			// Line 10: Car and truck expenses
			{
				Line:   "10",
				Type:   forms.UserInput,
				Label:  "Car and truck expenses",
				Prompt: "Enter car and truck expenses (business use only):",
			},
			// Line 17: Legal and professional services
			{
				Line:   "17",
				Type:   forms.UserInput,
				Label:  "Legal and professional services",
				Prompt: "Enter legal and professional service fees:",
			},
			// Line 18: Office expense
			{
				Line:   "18",
				Type:   forms.UserInput,
				Label:  "Office expense",
				Prompt: "Enter office expenses:",
			},
			// Line 22: Supplies
			{
				Line:   "22",
				Type:   forms.UserInput,
				Label:  "Supplies",
				Prompt: "Enter supply expenses:",
			},
			// Line 25: Utilities
			{
				Line:   "25",
				Type:   forms.UserInput,
				Label:  "Utilities",
				Prompt: "Enter utility expenses (business portion):",
			},
			// Line 27: Other expenses
			{
				Line:   "27",
				Type:   forms.UserInput,
				Label:  "Other expenses",
				Prompt: "Enter other business expenses not listed above:",
			},

			// Line 28: Total expenses
			{
				Line:      "28",
				Type:      forms.Computed,
				Label:     "Total expenses",
				DependsOn: []string{forms.SchedCLine8, forms.SchedCLine10, forms.SchedCLine17, forms.SchedCLine18, forms.SchedCLine22, forms.SchedCLine25, forms.SchedCLine27},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.SchedCLine8) +
						dv.Get(forms.SchedCLine10) +
						dv.Get(forms.SchedCLine17) +
						dv.Get(forms.SchedCLine18) +
						dv.Get(forms.SchedCLine22) +
						dv.Get(forms.SchedCLine25) +
						dv.Get(forms.SchedCLine27)
				},
			},

			// --- Net Profit or Loss ---

			// Line 31: Net profit or (loss)
			{
				Line:      "31",
				Type:      forms.Computed,
				Label:     "Net profit or (loss)",
				DependsOn: []string{forms.SchedCLine7, forms.SchedCLine28},
				Compute: func(dv forms.DepValues) float64 {
					return math.Max(0, dv.Get(forms.SchedCLine7)-dv.Get(forms.SchedCLine28))
				},
			},
		},
	}
}
