package federal

import (
	"math"

	"taxpilot/internal/forms"
	"taxpilot/pkg/taxmath"
)

func init() { forms.RegisterForm(Form8995) }

// QBI simplified form income thresholds for 2025
const (
	qbiRate            = 0.20   // 20% deduction
	qbiThresholdSingle = 191950 // Single/HOH/MFS
	qbiThresholdMFJ    = 383900 // MFJ/QSS
)

// Form8995 returns the FormDef for Form 8995 — Qualified Business Income
// Deduction Simplified Computation. This is the simplified version used when
// taxable income is below the threshold ($191,950 single / $383,900 MFJ for 2025).
//
// The QBI deduction is 20% of qualified business income, limited to 20% of
// taxable income (before QBI deduction) minus net capital gain.
//
// If taxable income exceeds the threshold, Form 8995-A (not yet implemented)
// applies with W-2 wage/UBIA limitations. In that case, this returns 0.
func Form8995() *forms.FormDef {
	return &forms.FormDef{
		ID:           forms.FormF8995,
		Name:         "Qualified Business Income Deduction (Simplified)",
		Jurisdiction: forms.Federal,
		TaxYears:      []int{2025},
		QuestionGroup: "business",
		QuestionOrder: 5,
		Fields: []forms.FieldDef{
			// Line 1: Total QBI from qualified businesses (Schedule C)
			{
				Line:      "1",
				Type:      forms.Computed,
				Label:     "Total qualified business income",
				DependsOn: []string{"schedule_c:31"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_c:31")
				},
			},
			// Line 2: Qualified REIT dividends (Section 199A dividends from 1099-DIV)
			{
				Line:      "2",
				Type:      forms.Computed,
				Label:     "Qualified REIT dividends and PTP income",
				DependsOn: []string{"1099div:*:section_199a_dividends"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.SumAll("1099div:*:section_199a_dividends")
				},
			},
			// Line 3: Combinable qualified business income (line 1 + line 2)
			{
				Line:      "3",
				Type:      forms.Computed,
				Label:     "Combinable QBI and REIT/PTP amounts",
				DependsOn: []string{"form_8995:1", "form_8995:2"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("form_8995:1") + dv.Get("form_8995:2")
				},
			},
			// Line 4: QBI component (20% of line 3, if positive)
			{
				Line:      "4",
				Type:      forms.Computed,
				Label:     "QBI component (20% of qualified income)",
				DependsOn: []string{"form_8995:3"},
				Compute: func(dv forms.DepValues) float64 {
					qbi := dv.Get("form_8995:3")
					if qbi <= 0 {
						return 0
					}
					return qbi * qbiRate
				},
			},
			// Line 5: Taxable income before QBI deduction
			// This is AGI minus standard/itemized deduction (but NOT minus QBI itself)
			{
				Line:      "5",
				Type:      forms.Computed,
				Label:     "Taxable income before QBI deduction",
				DependsOn: []string{"1040:11", "1040:12"},
				Compute: func(dv forms.DepValues) float64 {
					return math.Max(0, dv.Get("1040:11")-dv.Get("1040:12"))
				},
			},
			// Line 6: Net capital gain (from Schedule D line 16, if positive)
			{
				Line:      "6",
				Type:      forms.Computed,
				Label:     "Net capital gain",
				DependsOn: []string{"schedule_d:16"},
				Compute: func(dv forms.DepValues) float64 {
					return math.Max(0, dv.Get("schedule_d:16"))
				},
			},
			// Line 7: Line 5 minus line 6 (not less than 0)
			{
				Line:      "7",
				Type:      forms.Computed,
				Label:     "Taxable income minus net capital gain",
				DependsOn: []string{"form_8995:5", "form_8995:6"},
				Compute: func(dv forms.DepValues) float64 {
					return math.Max(0, dv.Get("form_8995:5")-dv.Get("form_8995:6"))
				},
			},
			// Line 8: Income limitation (20% of line 7)
			{
				Line:      "8",
				Type:      forms.Computed,
				Label:     "Income limitation (20% of adjusted taxable income)",
				DependsOn: []string{"form_8995:7"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("form_8995:7") * qbiRate
				},
			},
			// Line 10: QBI deduction — lesser of line 4 or line 8, but not less than 0
			// Returns 0 if taxable income exceeds threshold (Form 8995-A needed)
			{
				Line:      "10",
				Type:      forms.Computed,
				Label:     "Qualified business income deduction",
				DependsOn: []string{"form_8995:4", "form_8995:5", "form_8995:8", "1040:filing_status"},
				Compute: func(dv forms.DepValues) float64 {
					taxableIncome := dv.Get("form_8995:5")
					fs := taxmath.FilingStatus(dv.GetString("1040:filing_status"))
					threshold := getQBIThreshold(fs)

					// If above threshold, simplified form doesn't apply
					if taxableIncome > threshold {
						return 0
					}

					qbiComponent := dv.Get("form_8995:4")
					incomeLimit := dv.Get("form_8995:8")
					return math.Max(0, math.Min(qbiComponent, incomeLimit))
				},
			},
		},
	}
}

func getQBIThreshold(fs taxmath.FilingStatus) float64 {
	switch fs {
	case "mfj", "qss":
		return qbiThresholdMFJ
	default: // single, hoh, mfs
		return qbiThresholdSingle
	}
}
