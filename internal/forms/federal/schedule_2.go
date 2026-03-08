package federal

import (
	"math"

	"taxpilot/internal/forms"
)

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
		TaxYears:     []int{2025},
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
				DependsOn: []string{"schedule_2:1", "schedule_2:2"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_2:1") + dv.Get("schedule_2:2")
				},
			},

			// --- Part II: Other Taxes ---

			// Line 6: Self-employment tax (from Schedule SE line 6)
			{
				Line:      "6",
				Type:      forms.Computed,
				Label:     "Self-employment tax",
				DependsOn: []string{"schedule_se:6"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_se:6")
				},
			},
			// Line 12: Additional Medicare Tax (0.9% on wages + SE income above threshold)
			{
				Line:      "12",
				Type:      forms.Computed,
				Label:     "Additional Medicare Tax",
				DependsOn: []string{"1040:filing_status", "w2:*:medicare_wages", "schedule_se:3"},
				Compute: func(dv forms.DepValues) float64 {
					fs := dv.GetString("1040:filing_status")
					threshold := getAddlMedicareThreshold(fs)

					// Combined Medicare wages + SE earnings
					medicareWages := dv.SumAll("w2:*:medicare_wages")
					seEarnings := dv.Get("schedule_se:3")
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
				DependsOn: []string{"1040:filing_status", "1040:11", "1040:2b", "1040:3b", "schedule_d:16"},
				Compute: func(dv forms.DepValues) float64 {
					fs := dv.GetString("1040:filing_status")
					threshold := getNIITThreshold(fs)

					magi := dv.Get("1040:11") // AGI (MAGI ≈ AGI for most filers)
					if magi <= threshold {
						return 0
					}

					// Net investment income = interest + dividends + capital gains
					nii := dv.Get("1040:2b") + dv.Get("1040:3b") + dv.Get("schedule_d:16")
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
				DependsOn: []string{"form_8889:17b"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("form_8889:17b")
				},
			},
			// Line 21: Total Part II other taxes
			{
				Line:      "21",
				Type:      forms.Computed,
				Label:     "Total other taxes",
				DependsOn: []string{"schedule_2:6", "schedule_2:12", "schedule_2:17c", "schedule_2:18"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_2:6") +
						dv.Get("schedule_2:12") +
						dv.Get("schedule_2:17c") +
						dv.Get("schedule_2:18")
				},
			},
		},
	}
}

func getNIITThreshold(fs string) float64 {
	switch fs {
	case "mfj", "qss":
		return niitMFJ
	case "mfs":
		return niitMFS
	default: // single, hoh
		return niitSingle
	}
}

func getAddlMedicareThreshold(fs string) float64 {
	switch fs {
	case "mfj":
		return addlMedicareMFJ
	case "mfs":
		return addlMedicareMFS
	default: // single, hoh, qss
		return addlMedicareSingle
	}
}
