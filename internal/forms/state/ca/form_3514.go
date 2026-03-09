package ca

import (
	"math"

	"taxpilot/internal/forms"
)

func init() { forms.RegisterForm(Form3514) }

// Form3514 returns the FormDef for California Form 3514 — California Earned
// Income Tax Credit (CalEITC). This form computes the CalEITC and the Young
// Child Tax Credit (YCTC) for low-income California filers.
func Form3514() *forms.FormDef {
	return &forms.FormDef{
		ID:           forms.FormF3514,
		Name:         "Form 3514 — California Earned Income Tax Credit",
		Jurisdiction: forms.StateCA,
		TaxYears:      []int{2025},
		QuestionGroup: "ca",
		QuestionOrder: 7,
		Fields: []forms.FieldDef{
			// Line 1: Earned income (wages + positive self-employment income)
			{
				Line:      "1",
				Type:      forms.FederalRef,
				Label:     "Earned income",
				DependsOn: []string{"1040:1a", "schedule_c:31"},
				Compute: func(dv forms.DepValues) float64 {
					wages := dv.Get("1040:1a")
					se := dv.Get("schedule_c:31")
					if se < 0 {
						se = 0
					}
					return wages + se
				},
			},
			// Line 2: Filing status factor (1 = single/HOH, 2 = MFJ)
			{
				Line:      "2",
				Type:      forms.Computed,
				Label:     "Filing status factor",
				DependsOn: []string{"1040:filing_status"},
				Compute: func(dv forms.DepValues) float64 {
					fs := dv.GetString("1040:filing_status")
					if fs == "married_filing_jointly" {
						return 2
					}
					return 1
				},
			},
			// Line 3: Number of qualifying children (0-3+)
			{
				Line:   "3",
				Type:   forms.UserInput,
				Label:  "Number of qualifying children for CalEITC",
				Prompt: "How many qualifying children do you have for the California EITC?",
			},
			// Line 4: Income limit check (earned income must be <= $30,950)
			{
				Line:      "4",
				Type:      forms.Computed,
				Label:     "CalEITC income limit check",
				DependsOn: []string{"form_3514:1"},
				Compute: func(dv forms.DepValues) float64 {
					earned := dv.Get("form_3514:1")
					if earned <= 30950 {
						return 1 // eligible
					}
					return 0 // over limit
				},
			},
			// Line 5: CalEITC amount based on income and children
			{
				Line:      "5",
				Type:      forms.Computed,
				Label:     "CalEITC amount",
				DependsOn: []string{"form_3514:1", "form_3514:3", "form_3514:4"},
				Compute: func(dv forms.DepValues) float64 {
					if dv.Get("form_3514:4") == 0 {
						return 0 // over income limit
					}
					earned := dv.Get("form_3514:1")
					children := int(dv.Get("form_3514:3"))
					return computeCalEITC(earned, children)
				},
			},
			// Line 6_yctc: Whether taxpayer has a qualifying child under age 6
			{
				Line:    "6_yctc",
				Type:    forms.UserInput,
				ValueType: forms.StringValue,
				Label:   "Qualifying child under age 6",
				Prompt:  "Do you have a qualifying child under age 6?",
				Options: []string{"yes", "no"},
			},
			// Line 6: Young Child Tax Credit ($1,117 if child under 6)
			{
				Line:      "6",
				Type:      forms.Computed,
				Label:     "Young Child Tax Credit",
				DependsOn: []string{"form_3514:6_yctc", "form_3514:4"},
				Compute: func(dv forms.DepValues) float64 {
					if dv.Get("form_3514:4") == 0 {
						return 0 // over income limit
					}
					if dv.GetString("form_3514:6_yctc") == "yes" {
						return 1117
					}
					return 0
				},
			},
			// Line 7: Total CalEITC = Line 5 + Line 6
			{
				Line:      "7",
				Type:      forms.Computed,
				Label:     "Total CalEITC",
				DependsOn: []string{"form_3514:5", "form_3514:6"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("form_3514:5") + dv.Get("form_3514:6")
				},
			},
		},
	}
}

// computeCalEITC computes the CalEITC credit based on earned income and
// number of qualifying children. Uses simplified 2025 phase-in/phase-out
// schedule.
func computeCalEITC(earned float64, children int) float64 {
	if children < 0 {
		children = 0
	}
	if children > 3 {
		children = 3
	}

	type eitcParams struct {
		maxCredit    float64
		phaseOutStart float64
	}

	params := []eitcParams{
		{275, 7500},     // 0 children
		{1843, 11000},   // 1 child
		{3037, 15500},   // 2 children
		{3417, 15500},   // 3+ children
	}

	p := params[children]
	phaseOutEnd := 30950.0

	if earned <= 0 {
		return 0
	}

	if earned <= p.phaseOutStart {
		// In phase-in range: credit grows proportionally up to max
		return math.Round(p.maxCredit * earned / p.phaseOutStart * 100) / 100
	}

	if earned > phaseOutEnd {
		return 0
	}

	// Phase-out range: credit reduces linearly from max to 0
	remaining := phaseOutEnd - earned
	phaseOutRange := phaseOutEnd - p.phaseOutStart
	credit := p.maxCredit * remaining / phaseOutRange
	return math.Round(credit*100) / 100
}
