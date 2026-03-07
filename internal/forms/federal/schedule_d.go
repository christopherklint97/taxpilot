package federal

import (
	"taxpilot/internal/forms"
)

// ScheduleD returns the FormDef for Schedule D — Capital Gains and Losses.
// This form aggregates capital gains/losses from:
//   - Form 8949 (stock/security sales from 1099-B)
//   - 1099-DIV Box 2a (capital gain distributions from mutual funds)
//
// Schedule D line 16 (net capital gain/loss) flows to Schedule 1 line 7,
// which then flows to Form 1040 line 7.
func ScheduleD() *forms.FormDef {
	return &forms.FormDef{
		ID:           "schedule_d",
		Name:         "Schedule D — Capital Gains and Losses",
		Jurisdiction: forms.Federal,
		TaxYears:     []int{2025},
		Fields: []forms.FieldDef{
			// --- Part I: Short-Term Capital Gains and Losses ---

			// Line 1: Short-term from Form 8949 Box A (basis reported)
			// For MVP, all short-term goes here
			{
				Line:      "1",
				Type:      forms.Computed,
				Label:     "Short-term from Form 8949",
				DependsOn: []string{"form_8949:st_gain_loss"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("form_8949:st_gain_loss")
				},
			},
			// Line 7: Net short-term capital gain or (loss)
			{
				Line:      "7",
				Type:      forms.Computed,
				Label:     "Net short-term capital gain or (loss)",
				DependsOn: []string{"schedule_d:1"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_d:1")
				},
			},

			// --- Part II: Long-Term Capital Gains and Losses ---

			// Line 8: Long-term from Form 8949 Box D (basis reported)
			// For MVP, all long-term goes here
			{
				Line:      "8",
				Type:      forms.Computed,
				Label:     "Long-term from Form 8949",
				DependsOn: []string{"form_8949:lt_gain_loss"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("form_8949:lt_gain_loss")
				},
			},
			// Line 13: Capital gain distributions (from 1099-DIV Box 2a)
			{
				Line:      "13",
				Type:      forms.Computed,
				Label:     "Capital gain distributions",
				DependsOn: []string{"1099div:*:total_capital_gain"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.SumAll("1099div:*:total_capital_gain")
				},
			},
			// Line 15: Net long-term capital gain or (loss)
			{
				Line:      "15",
				Type:      forms.Computed,
				Label:     "Net long-term capital gain or (loss)",
				DependsOn: []string{"schedule_d:8", "schedule_d:13"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_d:8") + dv.Get("schedule_d:13")
				},
			},

			// --- Part III: Summary ---

			// Line 16: Combined net gain or (loss)
			{
				Line:      "16",
				Type:      forms.Computed,
				Label:     "Net capital gain or (loss)",
				DependsOn: []string{"schedule_d:7", "schedule_d:15"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_d:7") + dv.Get("schedule_d:15")
				},
			},
		},
	}
}
