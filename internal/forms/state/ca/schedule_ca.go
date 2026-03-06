package ca

import "taxpilot/internal/forms"

// ScheduleCA returns the FormDef for Schedule CA (540) — California Adjustments.
// For the MVP, this is mostly a passthrough for simple W-2 filers with no
// California-specific adjustments. As more income types are supported,
// the conformity differences documented in conformity.go will drive
// actual subtraction and addition computations here.
func ScheduleCA() *forms.FormDef {
	return &forms.FormDef{
		ID:           "ca_schedule_ca",
		Name:         "Schedule CA (540) — California Adjustments",
		Jurisdiction: forms.StateCA,
		TaxYears:     []int{2025},
		Fields: []forms.FieldDef{
			// Line 37, Column A: Federal amounts (mirrors federal AGI for the
			// bottom-line adjustment row)
			{
				Line:      "37_col_a",
				Type:      forms.FederalRef,
				Label:     "Federal amounts (from Form 1040)",
				DependsOn: []string{"1040:11"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("1040:11")
				},
			},
			// Line 37, Column B: Subtractions from federal income.
			// For MVP (simple W-2 filer), there are no subtractions.
			{
				Line:      "37_col_b",
				Type:      forms.Computed,
				Label:     "Subtractions (Column B)",
				DependsOn: []string{},
				Compute: func(dv forms.DepValues) float64 {
					return 0
				},
			},
			// Line 37, Column C: Additions to federal income.
			// For MVP (simple W-2 filer), there are no additions.
			{
				Line:      "37_col_c",
				Type:      forms.Computed,
				Label:     "Additions (Column C)",
				DependsOn: []string{},
				Compute: func(dv forms.DepValues) float64 {
					return 0
				},
			},
		},
	}
}
