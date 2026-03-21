package ca

import (
	"math"

	"taxpilot/internal/forms"
	"taxpilot/internal/forms/state"
	"taxpilot/pkg/taxmath"
)

func init() { forms.RegisterForm(F540) }

// F540 returns the FormDef for California Form 540 — California Resident
// Income Tax Return. This is a simplified MVP covering a W-2 filer with
// standard deduction and no dependents.
func F540() *forms.FormDef {
	return &forms.FormDef{
		ID:           forms.FormCA540,
		Name:         "California Resident Income Tax Return",
		Jurisdiction: forms.StateCA,
		TaxYears:      []int{2025},
		QuestionGroup: forms.GroupCA,
		QuestionOrder: 7,
		Fields: []forms.FieldDef{
			// Filing status — references the federal filing status from Form 1040
			{
				Line:      forms.LineFilingStatus,
				Type:      forms.FederalRef,
				Label:     "Filing status (from federal return)",
				DependsOn: []string{forms.F1040FilingStatus},
				ComputeStr: func(dv forms.DepValues) string {
					return dv.GetString(forms.F1040FilingStatus)
				},
			},

			// --- Income ---

			// Line 7: California wages (from W-2 Box 16)
			{
				Line:      "7",
				Type:      forms.Computed,
				Label:     "Wages, salaries, tips (CA)",
				DependsOn: []string{forms.W2WildcardStateWages},
				Compute: func(dv forms.DepValues) float64 {
					return dv.SumAll(forms.W2WildcardStateWages)
				},
			},
			// Line 13: Federal AGI (from Form 1040 line 11)
			{
				Line:      "13",
				Type:      forms.FederalRef,
				Label:     "Federal adjusted gross income",
				DependsOn: []string{forms.F1040Line11},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.F1040Line11)
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
				DependsOn: []string{forms.SchedCALine37ColC},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.SchedCALine37ColC)
				},
			},
			// Line 17: California AGI = federal AGI - subtractions + additions
			{
				Line:      "17",
				Type:      forms.Computed,
				Label:     "California adjusted gross income",
				DependsOn: []string{forms.CA540Line13, forms.CA540Line14, forms.CA540Line15},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.CA540Line13) - dv.Get(forms.CA540Line14) + dv.Get(forms.CA540Line15)
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
				DependsOn: []string{forms.CA540Line17, forms.CA540Line18},
				Compute: func(dv forms.DepValues) float64 {
					return math.Max(0, dv.Get(forms.CA540Line17)-dv.Get(forms.CA540Line18))
				},
			},

			// --- Tax computation ---

			// Line 31: CA tax (bracket computation, excluding mental health tax)
			{
				Line:      "31",
				Type:      forms.Computed,
				Label:     "California tax",
				DependsOn: []string{forms.CA540Line19, "ca_540:filing_status"},
				Compute: func(dv forms.DepValues) float64 {
					fs := taxmath.FilingStatus(dv.GetString("ca_540:filing_status"))
					taxableIncome := dv.Get(forms.CA540Line19)
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
				DependsOn: []string{forms.CA540Line31, forms.CA540Line32},
				Compute: func(dv forms.DepValues) float64 {
					return math.Max(0, dv.Get(forms.CA540Line31)-dv.Get(forms.CA540Line32))
				},
			},
			// Line 36: Mental Health Services Tax (1% on taxable income > $1M)
			{
				Line:      "36",
				Type:      forms.Computed,
				Label:     "Mental Health Services Tax",
				DependsOn: []string{forms.CA540Line19},
				Compute: func(dv forms.DepValues) float64 {
					return taxmath.GetCAMentalHealthTax(dv.Get(forms.CA540Line19))
				},
			},
			// Line 40: Total CA tax (includes health coverage penalty from Form 3853)
			{
				Line:      "40",
				Type:      forms.Computed,
				Label:     "Total California tax",
				DependsOn: []string{forms.CA540Line35, forms.CA540Line36, "form_3853:7"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.CA540Line35) + dv.Get(forms.CA540Line36) + dv.Get("form_3853:7")
				},
			},

			// --- Payments ---

			// Line 71: CA tax withheld (from W-2 Box 17)
			{
				Line:      "71",
				Type:      forms.Computed,
				Label:     "California income tax withheld",
				DependsOn: []string{forms.W2WildcardStateTaxWH},
				Compute: func(dv forms.DepValues) float64 {
					return dv.SumAll(forms.W2WildcardStateTaxWH)
				},
			},
			// Line 74: Total payments and credits (includes CalEITC from Form 3514)
			{
				Line:      "74",
				Type:      forms.Computed,
				Label:     "Total payments and credits",
				DependsOn: []string{forms.CA540Line71, "form_3514:7"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.CA540Line71) + dv.Get("form_3514:7")
				},
			},

			// --- Refund or Amount Owed ---

			// Line 91: Overpayment (refund)
			{
				Line:      "91",
				Type:      forms.Computed,
				Label:     "Overpayment (refund)",
				DependsOn: []string{forms.CA540Line74, forms.CA540Line40},
				Compute: func(dv forms.DepValues) float64 {
					return math.Max(0, dv.Get(forms.CA540Line74)-dv.Get(forms.CA540Line40))
				},
			},
			// Line 93: Amount you owe
			{
				Line:      "93",
				Type:      forms.Computed,
				Label:     "Amount you owe",
				DependsOn: []string{forms.CA540Line40, forms.CA540Line74},
				Compute: func(dv forms.DepValues) float64 {
					return math.Max(0, dv.Get(forms.CA540Line40)-dv.Get(forms.CA540Line74))
				},
			},
		},
	}
}

// CAFormSet implements the state.StateFormSet interface for California.
type CAFormSet struct{}

// Compile-time interface compliance check.
var _ state.StateFormSet = CAFormSet{}

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
