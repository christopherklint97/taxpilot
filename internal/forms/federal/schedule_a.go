package federal

import (
	"math"

	"taxpilot/internal/forms"
	"taxpilot/pkg/taxmath"
)

// ScheduleA returns the FormDef for Schedule A — Itemized Deductions.
// Taxpayers choose between standard deduction and itemized deductions.
// Key sections: medical expenses, state/local taxes (SALT), interest,
// charitable contributions, and other deductions.
func ScheduleA() *forms.FormDef {
	return &forms.FormDef{
		ID:           "schedule_a",
		Name:         "Schedule A — Itemized Deductions",
		Jurisdiction: forms.Federal,
		TaxYears:     []int{2025},
		Fields: []forms.FieldDef{
			// --- Medical and Dental Expenses ---

			// Line 1: Medical and dental expenses
			{
				Line:   "1",
				Type:   forms.UserInput,
				Label:  "Medical and dental expenses",
				Prompt: "Enter your total medical and dental expenses:",
			},
			// Line 2: AGI (from 1040 line 11)
			{
				Line:      "2",
				Type:      forms.Computed,
				Label:     "AGI (from Form 1040 line 11)",
				DependsOn: []string{"1040:11"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("1040:11")
				},
			},
			// Line 3: 7.5% of AGI threshold
			{
				Line:      "3",
				Type:      forms.Computed,
				Label:     "7.5% of AGI",
				DependsOn: []string{"schedule_a:2"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_a:2") * 0.075
				},
			},
			// Line 4: Deductible medical expenses (excess over 7.5% AGI)
			{
				Line:      "4",
				Type:      forms.Computed,
				Label:     "Deductible medical and dental expenses",
				DependsOn: []string{"schedule_a:1", "schedule_a:3"},
				Compute: func(dv forms.DepValues) float64 {
					return math.Max(0, dv.Get("schedule_a:1")-dv.Get("schedule_a:3"))
				},
			},

			// --- Taxes You Paid ---

			// Line 5a: State and local income taxes (or sales taxes)
			{
				Line:   "5a",
				Type:   forms.UserInput,
				Label:  "State and local income taxes paid",
				Prompt: "Enter state and local income taxes paid (or general sales taxes):",
			},
			// Line 5b: State and local personal property taxes
			{
				Line:   "5b",
				Type:   forms.UserInput,
				Label:  "State and local personal property taxes",
				Prompt: "Enter state and local personal property taxes paid:",
			},
			// Line 5c: State and local real estate taxes
			{
				Line:   "5c",
				Type:   forms.UserInput,
				Label:  "State and local real estate taxes",
				Prompt: "Enter state and local real estate taxes paid:",
			},
			// Line 5d: Total SALT (sum of 5a-5c)
			{
				Line:      "5d",
				Type:      forms.Computed,
				Label:     "Total state and local taxes",
				DependsOn: []string{"schedule_a:5a", "schedule_a:5b", "schedule_a:5c"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_a:5a") + dv.Get("schedule_a:5b") + dv.Get("schedule_a:5c")
				},
			},
			// Line 5e: SALT deduction (capped at $10,000 / $5,000 MFS)
			{
				Line:      "5e",
				Type:      forms.Computed,
				Label:     "State and local taxes (SALT) deduction",
				DependsOn: []string{"schedule_a:5d", "1040:filing_status"},
				Compute: func(dv forms.DepValues) float64 {
					total := dv.Get("schedule_a:5d")
					fs := taxmath.FilingStatus(dv.GetString("1040:filing_status"))
					cap := 10000.0
					if fs == taxmath.MarriedFilingSep {
						cap = 5000.0
					}
					return math.Min(total, cap)
				},
			},

			// --- Interest You Paid ---

			// Line 8a: Home mortgage interest (from Form 1098)
			{
				Line:   "8a",
				Type:   forms.UserInput,
				Label:  "Home mortgage interest and points (Form 1098)",
				Prompt: "Enter home mortgage interest and points reported on Form 1098:",
			},
			// Line 10: Investment interest (deferred — Form 4952)
			{
				Line:      "10",
				Type:      forms.Computed,
				Label:     "Investment interest",
				DependsOn: []string{},
				Compute: func(dv forms.DepValues) float64 {
					return 0
				},
			},
			// Line 11: Total interest deduction
			{
				Line:      "11",
				Type:      forms.Computed,
				Label:     "Total interest deduction",
				DependsOn: []string{"schedule_a:8a", "schedule_a:10"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_a:8a") + dv.Get("schedule_a:10")
				},
			},

			// --- Gifts to Charity ---

			// Line 12: Gifts by cash or check
			{
				Line:   "12",
				Type:   forms.UserInput,
				Label:  "Charitable contributions (cash or check)",
				Prompt: "Enter charitable contributions paid by cash or check:",
			},
			// Line 13: Gifts other than cash or check
			{
				Line:   "13",
				Type:   forms.UserInput,
				Label:  "Charitable contributions (other than cash)",
				Prompt: "Enter charitable contributions other than cash or check:",
			},
			// Line 14: Carryover from prior year
			{
				Line:   "14",
				Type:   forms.UserInput,
				Label:  "Charitable contribution carryover from prior year",
				Prompt: "Enter charitable contribution carryover from prior year (if any):",
			},
			// Line 15: Total charitable contributions
			{
				Line:      "15",
				Type:      forms.Computed,
				Label:     "Total charitable contributions",
				DependsOn: []string{"schedule_a:12", "schedule_a:13", "schedule_a:14"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_a:12") + dv.Get("schedule_a:13") + dv.Get("schedule_a:14")
				},
			},

			// --- Other Itemized Deductions ---

			// Line 16: Casualty and theft losses (from Form 4684)
			{
				Line:      "16",
				Type:      forms.Computed,
				Label:     "Casualty and theft losses",
				DependsOn: []string{},
				Compute: func(dv forms.DepValues) float64 {
					return 0 // deferred — requires Form 4684
				},
			},

			// --- Total ---

			// Line 17: Total itemized deductions
			{
				Line:      "17",
				Type:      forms.Computed,
				Label:     "Total itemized deductions",
				DependsOn: []string{"schedule_a:4", "schedule_a:5e", "schedule_a:11", "schedule_a:15", "schedule_a:16"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_a:4") +
						dv.Get("schedule_a:5e") +
						dv.Get("schedule_a:11") +
						dv.Get("schedule_a:15") +
						dv.Get("schedule_a:16")
				},
			},
		},
	}
}
