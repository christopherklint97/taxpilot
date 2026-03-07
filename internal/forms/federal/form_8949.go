package federal

import (
	"taxpilot/internal/forms"
)

// Form8949 returns the FormDef for Form 8949 — Sales and Other Dispositions
// of Capital Assets. This form aggregates 1099-B transactions and computes
// gain/loss for each category, feeding into Schedule D.
//
// The IRS splits Form 8949 into two parts:
//   Part I:  Short-term (held ≤ 1 year)
//   Part II: Long-term (held > 1 year)
//
// Each part has three boxes based on whether basis was reported to IRS:
//   Box A/D: Basis reported (no adjustments needed)
//   Box B/E: Basis not reported
//   Box C/F: Cannot determine / Form 1099-B not received
//
// For the MVP, we aggregate all 1099-B transactions into short-term and
// long-term totals. The user indicates term and basis reporting on each 1099-B.
func Form8949() *forms.FormDef {
	return &forms.FormDef{
		ID:           "form_8949",
		Name:         "Form 8949 — Sales and Other Dispositions of Capital Assets",
		Jurisdiction: forms.Federal,
		TaxYears:     []int{2025},
		Fields: []forms.FieldDef{
			// --- Part I: Short-Term ---

			// Total short-term proceeds (sum of all short-term 1099-B proceeds)
			{
				Line:      "st_proceeds",
				Type:      forms.Computed,
				Label:     "Short-term total proceeds",
				DependsOn: []string{"1099b:*:proceeds", "1099b:*:term"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.SumAllWhere("1099b:*:proceeds", "1099b:*:term", "short")
				},
			},
			// Total short-term cost basis
			{
				Line:      "st_basis",
				Type:      forms.Computed,
				Label:     "Short-term total cost basis",
				DependsOn: []string{"1099b:*:cost_basis", "1099b:*:term"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.SumAllWhere("1099b:*:cost_basis", "1099b:*:term", "short")
				},
			},
			// Total short-term wash sale adjustments
			{
				Line:      "st_wash",
				Type:      forms.Computed,
				Label:     "Short-term wash sale adjustments",
				DependsOn: []string{"1099b:*:wash_sale_loss", "1099b:*:term"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.SumAllWhere("1099b:*:wash_sale_loss", "1099b:*:term", "short")
				},
			},
			// Short-term gain or loss
			{
				Line:      "st_gain_loss",
				Type:      forms.Computed,
				Label:     "Short-term gain or (loss)",
				DependsOn: []string{"form_8949:st_proceeds", "form_8949:st_basis", "form_8949:st_wash"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("form_8949:st_proceeds") - dv.Get("form_8949:st_basis") + dv.Get("form_8949:st_wash")
				},
			},

			// --- Part II: Long-Term ---

			// Total long-term proceeds
			{
				Line:      "lt_proceeds",
				Type:      forms.Computed,
				Label:     "Long-term total proceeds",
				DependsOn: []string{"1099b:*:proceeds", "1099b:*:term"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.SumAllWhere("1099b:*:proceeds", "1099b:*:term", "long")
				},
			},
			// Total long-term cost basis
			{
				Line:      "lt_basis",
				Type:      forms.Computed,
				Label:     "Long-term total cost basis",
				DependsOn: []string{"1099b:*:cost_basis", "1099b:*:term"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.SumAllWhere("1099b:*:cost_basis", "1099b:*:term", "long")
				},
			},
			// Total long-term wash sale adjustments
			{
				Line:      "lt_wash",
				Type:      forms.Computed,
				Label:     "Long-term wash sale adjustments",
				DependsOn: []string{"1099b:*:wash_sale_loss", "1099b:*:term"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.SumAllWhere("1099b:*:wash_sale_loss", "1099b:*:term", "long")
				},
			},
			// Long-term gain or loss
			{
				Line:      "lt_gain_loss",
				Type:      forms.Computed,
				Label:     "Long-term gain or (loss)",
				DependsOn: []string{"form_8949:lt_proceeds", "form_8949:lt_basis", "form_8949:lt_wash"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("form_8949:lt_proceeds") - dv.Get("form_8949:lt_basis") + dv.Get("form_8949:lt_wash")
				},
			},
		},
	}
}
