package ca

import (
	"math"

	"taxpilot/internal/forms"
	"taxpilot/pkg/taxmath"
)

// F540 returns the FormDef for California Form 540 — California Resident
// Income Tax Return. This is a simplified MVP covering a W-2 filer with
// standard deduction and no dependents.
func F540() *forms.FormDef {
	return &forms.FormDef{
		ID:           forms.FormCA540,
		Name:         "California Resident Income Tax Return",
		Jurisdiction: forms.StateCA,
		TaxYears:     []int{2025},
		Fields: []forms.FieldDef{
			// Filing status — references the federal filing status from Form 1040
			{
				Line:      "filing_status",
				Type:      forms.FederalRef,
				Label:     "Filing status (from federal return)",
				DependsOn: []string{"1040:filing_status"},
				ComputeStr: func(dv forms.DepValues) string {
					return dv.GetString("1040:filing_status")
				},
			},

			// --- Income ---

			// Line 7: California wages (from W-2 Box 16)
			{
				Line:      "7",
				Type:      forms.Computed,
				Label:     "Wages, salaries, tips (CA)",
				DependsOn: []string{"w2:*:state_wages"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.SumAll("w2:*:state_wages")
				},
			},
			// Line 13: Federal AGI (from Form 1040 line 11)
			{
				Line:      "13",
				Type:      forms.FederalRef,
				Label:     "Federal adjusted gross income",
				DependsOn: []string{"1040:11"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("1040:11")
				},
			},
			// Line 14: CA subtractions from Schedule CA
			{
				Line:      "14",
				Type:      forms.Computed,
				Label:     "California subtractions (from Schedule CA)",
				DependsOn: []string{"ca_schedule_ca:37_col_b"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("ca_schedule_ca:37_col_b")
				},
			},
			// Line 15: CA additions from Schedule CA
			{
				Line:      "15",
				Type:      forms.Computed,
				Label:     "California additions (from Schedule CA)",
				DependsOn: []string{"ca_schedule_ca:37_col_c"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("ca_schedule_ca:37_col_c")
				},
			},
			// Line 17: California AGI = federal AGI - subtractions + additions
			{
				Line:      "17",
				Type:      forms.Computed,
				Label:     "California adjusted gross income",
				DependsOn: []string{"ca_540:13", "ca_540:14", "ca_540:15"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("ca_540:13") - dv.Get("ca_540:14") + dv.Get("ca_540:15")
				},
			},

			// --- Deductions ---

			// Line 18: CA deduction — larger of CA standard deduction or CA itemized
			{
				Line:      "18",
				Type:      forms.Computed,
				Label:     "California deduction (standard or itemized)",
				DependsOn: []string{"ca_540:filing_status", "ca_schedule_ca:ca_itemized"},
				Compute: func(dv forms.DepValues) float64 {
					fs := taxmath.FilingStatus(dv.GetString("ca_540:filing_status"))
					standard := taxmath.GetStandardDeduction(dv.TaxYear(), taxmath.StateCA, fs)
					caItemized := dv.Get("ca_schedule_ca:ca_itemized")
					return math.Max(standard, caItemized)
				},
			},
			// Line 19: CA taxable income
			{
				Line:      "19",
				Type:      forms.Computed,
				Label:     "California taxable income",
				DependsOn: []string{"ca_540:17", "ca_540:18"},
				Compute: func(dv forms.DepValues) float64 {
					return math.Max(0, dv.Get("ca_540:17")-dv.Get("ca_540:18"))
				},
			},

			// --- Tax computation ---

			// Line 31: CA tax (bracket computation, excluding mental health tax)
			{
				Line:      "31",
				Type:      forms.Computed,
				Label:     "California tax",
				DependsOn: []string{"ca_540:19", "ca_540:filing_status"},
				Compute: func(dv forms.DepValues) float64 {
					fs := taxmath.FilingStatus(dv.GetString("ca_540:filing_status"))
					taxableIncome := dv.Get("ca_540:19")
					// Use bracket computation only (not ComputeTax which includes
					// mental health tax — that is computed separately on line 36).
					brackets := taxmath.GetBrackets(dv.TaxYear(), taxmath.StateCA, fs)
					if brackets == nil {
						return 0
					}
					return taxmath.ComputeBracketTax(taxableIncome, brackets)
				},
			},
			// Line 32: CA exemption credits
			{
				Line:      "32",
				Type:      forms.Computed,
				Label:     "Exemption credits",
				DependsOn: []string{"ca_540:filing_status"},
				Compute: func(dv forms.DepValues) float64 {
					fs := taxmath.FilingStatus(dv.GetString("ca_540:filing_status"))
					// MVP: 0 dependents
					return taxmath.GetCAExemptionCredit(dv.TaxYear(), fs, 0)
				},
			},
			// Line 35: Net tax after exemption credits
			{
				Line:      "35",
				Type:      forms.Computed,
				Label:     "Net tax (after exemption credits)",
				DependsOn: []string{"ca_540:31", "ca_540:32"},
				Compute: func(dv forms.DepValues) float64 {
					return math.Max(0, dv.Get("ca_540:31")-dv.Get("ca_540:32"))
				},
			},
			// Line 36: Mental Health Services Tax (1% on taxable income > $1M)
			{
				Line:      "36",
				Type:      forms.Computed,
				Label:     "Mental Health Services Tax",
				DependsOn: []string{"ca_540:19"},
				Compute: func(dv forms.DepValues) float64 {
					return taxmath.GetCAMentalHealthTax(dv.Get("ca_540:19"))
				},
			},
			// Line 40: Total CA tax (includes health coverage penalty from Form 3853)
			{
				Line:      "40",
				Type:      forms.Computed,
				Label:     "Total California tax",
				DependsOn: []string{"ca_540:35", "ca_540:36", "form_3853:7"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("ca_540:35") + dv.Get("ca_540:36") + dv.Get("form_3853:7")
				},
			},

			// --- Payments ---

			// Line 71: CA tax withheld (from W-2 Box 17)
			{
				Line:      "71",
				Type:      forms.Computed,
				Label:     "California income tax withheld",
				DependsOn: []string{"w2:*:state_tax_withheld"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.SumAll("w2:*:state_tax_withheld")
				},
			},
			// Line 74: Total payments and credits (includes CalEITC from Form 3514)
			{
				Line:      "74",
				Type:      forms.Computed,
				Label:     "Total payments and credits",
				DependsOn: []string{"ca_540:71", "form_3514:7"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("ca_540:71") + dv.Get("form_3514:7")
				},
			},

			// --- Refund or Amount Owed ---

			// Line 91: Overpayment (refund)
			{
				Line:      "91",
				Type:      forms.Computed,
				Label:     "Overpayment (refund)",
				DependsOn: []string{"ca_540:74", "ca_540:40"},
				Compute: func(dv forms.DepValues) float64 {
					return math.Max(0, dv.Get("ca_540:74")-dv.Get("ca_540:40"))
				},
			},
			// Line 93: Amount you owe
			{
				Line:      "93",
				Type:      forms.Computed,
				Label:     "Amount you owe",
				DependsOn: []string{"ca_540:40", "ca_540:74"},
				Compute: func(dv forms.DepValues) float64 {
					return math.Max(0, dv.Get("ca_540:40")-dv.Get("ca_540:74"))
				},
			},
		},
	}
}

// CAFormSet implements the state.StateFormSet interface for California.
type CAFormSet struct{}

// Code returns the two-letter state abbreviation.
func (c CAFormSet) Code() string { return "CA" }

// Name returns the full state name.
func (c CAFormSet) Name() string { return "California" }

// RequiredForms returns the form IDs always needed for a CA filing.
func (c CAFormSet) RequiredForms() []string {
	return []string{"ca_540", "ca_schedule_ca", "form_3514", "form_3853"}
}

// ConditionalForms returns forms that are conditionally required.
func (c CAFormSet) ConditionalForms() map[string]string {
	return map[string]string{
		// Future: Schedule CA-related forms, CA Schedule D, etc.
	}
}
