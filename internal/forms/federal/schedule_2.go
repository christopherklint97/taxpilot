package federal

import (
	"math"

	"taxpilot/internal/forms"
)

func init() { forms.RegisterForm(Schedule2) }

// Schedule 2 thresholds for 2025
const (
	// Net Investment Income Tax (NIIT) — IRC §1411
	niitRate         = 0.038  // 3.8%
	niitSingle       = 200000 // MAGI threshold for single/HOH
	niitMFJ          = 250000 // MAGI threshold for MFJ/QSS
	niitMFS          = 125000 // MAGI threshold for MFS

	// Additional Medicare Tax — IRC §3101(b)(2)
	addlMedicareRate       = 0.009  // 0.9%
	addlMedicareSingle     = 200000 // threshold for single/HOH
	addlMedicareMFJ        = 250000 // threshold for MFJ
	addlMedicareMFS        = 125000 // threshold for MFS
)

// Schedule2 returns the FormDef for Schedule 2 — Additional Taxes.
// This captures taxes beyond the basic income tax:
//   - Line 6: Self-employment tax (from Schedule SE) — already wired via 1040:23
//   - Line 18: Net Investment Income Tax (3.8% on investment income above threshold)
//   - Line 23: Additional Medicare Tax (0.9% on earned income above threshold)
//   - Line 21: Total additional taxes -> flows to 1040 line 23
func Schedule2() *forms.FormDef {
	return &forms.FormDef{
		ID:           forms.FormSchedule2,
		Name:         "Schedule 2 — Additional Taxes",
		Jurisdiction: forms.Federal,
		TaxYears:      []int{2025},
		QuestionGroup: "deductions",
		QuestionOrder: 6,
		Fields: []forms.FieldDef{
			// --- Part I: Tax ---

			// Line 1: AMT (Alternative Minimum Tax) — deferred
			{
				Line:      "1",
				Type:      forms.Computed,
				Label:     "Alternative minimum tax (Form 6251)",
				DependsOn: []string{},
				Compute: func(dv forms.DepValues) float64 {
					return 0 // AMT deferred
				},
			},
			// Line 2: Excess advance premium tax credit repayment — deferred
			{
				Line:      "2",
				Type:      forms.Computed,
				Label:     "Excess advance premium tax credit repayment",
				DependsOn: []string{},
				Compute: func(dv forms.DepValues) float64 {
					return 0
				},
			},
			// Line 3: Total Part I (AMT + excess PTC)
			{
				Line:      "3",
				Type:      forms.Computed,
				Label:     "Total Part I additional taxes",
				DependsOn: []string{forms.Sched2Line1, "schedule_2:2"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.Sched2Line1) + dv.Get("schedule_2:2")
				},
			},

			// --- Part II: Other Taxes ---

			// Line 6: Self-employment tax (from Schedule SE line 6)
			{
				Line:      "6",
				Type:      forms.Computed,
				Label:     "Self-employment tax",
				DependsOn: []string{forms.SchedSELine6},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.SchedSELine6)
				},
			},
			// Line 12: Additional Medicare Tax (0.9% on wages + SE income above threshold)
			{
				Line:      "12",
				Type:      forms.Computed,
				Label:     "Additional Medicare Tax",
				DependsOn: []string{forms.F1040FilingStatus, forms.W2WildcardMedicareWages, forms.SchedSELine3},
				Compute: func(dv forms.DepValues) float64 {
					fs := dv.GetString(forms.F1040FilingStatus)
					threshold := getAddlMedicareThreshold(fs)

					// Combined Medicare wages + SE earnings
					medicareWages := dv.SumAll(forms.W2WildcardMedicareWages)
					seEarnings := dv.Get(forms.SchedSELine3)
					totalEarned := medicareWages + seEarnings

					excess := totalEarned - threshold
					if excess <= 0 {
						return 0
					}

					// Credit for Additional Medicare Tax already withheld by employer
					// (employer withholds on wages > $200k regardless of filing status)
					employerWithheld := 0.0
					if medicareWages > 200000 {
						employerWithheld = (medicareWages - 200000) * addlMedicareRate
					}

					tax := excess * addlMedicareRate
					return math.Max(0, tax-employerWithheld)
				},
			},
			// Line 18: Net Investment Income Tax (3.8%)
			{
				Line:      "18",
				Type:      forms.Computed,
				Label:     "Net investment income tax",
				DependsOn: []string{forms.F1040FilingStatus, forms.F1040Line11, forms.F1040Line2b, forms.F1040Line3b, forms.SchedDLine16},
				Compute: func(dv forms.DepValues) float64 {
					fs := dv.GetString(forms.F1040FilingStatus)
					threshold := getNIITThreshold(fs)

					magi := dv.Get(forms.F1040Line11) // AGI (MAGI ≈ AGI for most filers)
					if magi <= threshold {
						return 0
					}

					// Net investment income = interest + dividends + capital gains
					nii := dv.Get(forms.F1040Line2b) + dv.Get(forms.F1040Line3b) + dv.Get(forms.SchedDLine16)
					if nii <= 0 {
						return 0
					}

					// NIIT applies to the lesser of NII or excess MAGI
					excess := magi - threshold
					taxable := math.Min(nii, excess)
					return taxable * niitRate
				},
			},
			// Line 17c: Additional tax on HSA distributions (from Form 8889 line 17b)
			{
				Line:      "17c",
				Type:      forms.Computed,
				Label:     "Additional tax on HSA distributions",
				DependsOn: []string{forms.F8889Line17b},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.F8889Line17b)
				},
			},
			// Line 21: Total Part II other taxes
			{
				Line:      "21",
				Type:      forms.Computed,
				Label:     "Total other taxes",
				DependsOn: []string{forms.Sched2Line6, forms.Sched2Line12, forms.Sched2Line17c, forms.Sched2Line18},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.Sched2Line6) +
						dv.Get(forms.Sched2Line12) +
						dv.Get(forms.Sched2Line17c) +
						dv.Get(forms.Sched2Line18)
				},
			},
		},
	}
}

func getNIITThreshold(fs string) float64 {
	switch fs {
	case forms.FilingMFJ, forms.FilingQSS:
		return niitMFJ
	case forms.FilingMFS:
		return niitMFS
	default: // single, hoh
		return niitSingle
	}
}

func getAddlMedicareThreshold(fs string) float64 {
	switch fs {
	case forms.FilingMFJ:
		return addlMedicareMFJ
	case forms.FilingMFS:
		return addlMedicareMFS
	default: // single, hoh, qss
		return addlMedicareSingle
	}
}
