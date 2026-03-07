package ca

import "taxpilot/internal/forms"

// ScheduleCA returns the FormDef for Schedule CA (540) — California Adjustments.
// This form adjusts federal income for California differences.
//
// Part I, Section A: Income adjustments
//   - Line 2: Interest — subtract U.S. obligation interest (CA-exempt);
//     add out-of-state muni bond interest (CA-taxable)
//   - Line 3: Dividends — adjust for CA conformity differences
//   - Line 7: Capital gains — CA generally conforms (taxes LTCG as ordinary,
//     but that's handled by CA brackets, not a Schedule CA adjustment)
//
// Part I, Section B: Adjustments to income
//   - Line 15: HSA deduction add-back (CA does not conform to IRC §223)
//
// Part II: Itemized deduction adjustments (when itemizing)
//   - Line 5a: Remove state/local income tax deduction (CA does not allow)
//   - Line 5e: Recompute SALT without state income tax and without federal cap
//
// Line 37: Totals flow to Form 540 lines 14 and 15
func ScheduleCA() *forms.FormDef {
	return &forms.FormDef{
		ID:           "ca_schedule_ca",
		Name:         "Schedule CA (540) — California Adjustments",
		Jurisdiction: forms.StateCA,
		TaxYears:     []int{2025},
		Fields: []forms.FieldDef{
			// ===================================================================
			// Part I, Section A: Income
			// ===================================================================

			// Line 2, Column A: Federal taxable interest (from 1040 line 2b)
			{
				Line:      "2_col_a",
				Type:      forms.FederalRef,
				Label:     "Federal taxable interest",
				DependsOn: []string{"1040:2b"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("1040:2b")
				},
			},
			// Line 2, Column B: Interest subtractions (U.S. obligation interest
			// is exempt from CA tax — subtract it here)
			{
				Line:      "2_col_b",
				Type:      forms.Computed,
				Label:     "Interest subtractions (U.S. obligations exempt in CA)",
				DependsOn: []string{"1099int:*:us_savings_bond_interest"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.SumAll("1099int:*:us_savings_bond_interest")
				},
			},
			// Line 2, Column C: Interest additions (out-of-state muni bond
			// interest is federally exempt but CA-taxable)
			{
				Line:      "2_col_c",
				Type:      forms.Computed,
				Label:     "Interest additions (non-CA muni bond interest)",
				DependsOn: []string{},
				Compute: func(dv forms.DepValues) float64 {
					// Deferred: requires user to identify CA vs non-CA munis
					return 0
				},
			},

			// Line 3, Column A: Federal ordinary dividends (from 1040 line 3b)
			{
				Line:      "3_col_a",
				Type:      forms.FederalRef,
				Label:     "Federal ordinary dividends",
				DependsOn: []string{"1040:3b"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("1040:3b")
				},
			},
			// Line 3, Column B: Dividend subtractions (CA generally conforms)
			{
				Line:      "3_col_b",
				Type:      forms.Computed,
				Label:     "Dividend subtractions",
				DependsOn: []string{},
				Compute: func(dv forms.DepValues) float64 {
					return 0
				},
			},
			// Line 3, Column C: Dividend additions (CA generally conforms)
			{
				Line:      "3_col_c",
				Type:      forms.Computed,
				Label:     "Dividend additions",
				DependsOn: []string{},
				Compute: func(dv forms.DepValues) float64 {
					return 0
				},
			},

			// Line 7, Column A: Federal capital gain (from 1040 line 7)
			// CA taxes capital gains as ordinary income, but this is handled
			// by the CA tax brackets — no Schedule CA adjustment needed.
			{
				Line:      "7_col_a",
				Type:      forms.FederalRef,
				Label:     "Federal capital gain or (loss)",
				DependsOn: []string{"1040:7"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("1040:7")
				},
			},
			// Line 7, Column B: Capital gain subtractions
			{
				Line:      "7_col_b",
				Type:      forms.Computed,
				Label:     "Capital gain subtractions",
				DependsOn: []string{},
				Compute: func(dv forms.DepValues) float64 {
					return 0 // CA generally conforms on capital gains
				},
			},
			// Line 7, Column C: Capital gain additions
			{
				Line:      "7_col_c",
				Type:      forms.Computed,
				Label:     "Capital gain additions",
				DependsOn: []string{},
				Compute: func(dv forms.DepValues) float64 {
					return 0 // CA generally conforms on capital gains
				},
			},

			// ===================================================================
			// Part I, Section B: Adjustments to Income
			// ===================================================================

			// Line 12: Business income — CA generally conforms to federal
			// Schedule C. No adjustment needed.
			{
				Line:      "12_col_b",
				Type:      forms.Computed,
				Label:     "Business income subtractions",
				DependsOn: []string{},
				Compute: func(dv forms.DepValues) float64 {
					return 0 // CA conforms to federal Schedule C
				},
			},
			{
				Line:      "12_col_c",
				Type:      forms.Computed,
				Label:     "Business income additions",
				DependsOn: []string{},
				Compute: func(dv forms.DepValues) float64 {
					return 0 // CA conforms to federal Schedule C
				},
			},

			// Line 15, Column C: HSA deduction add-back
			// CA does not conform to federal HSA treatment (IRC §223).
			// The federal HSA deduction (Schedule 1 line 15) must be added back.
			{
				Line:      "15_col_c",
				Type:      forms.Computed,
				Label:     "HSA deduction add-back (CA does not allow)",
				DependsOn: []string{"form_8889:9"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("form_8889:9")
				},
			},

			// Line 16: Self-employment tax deduction — CA conforms to federal
			// treatment. No adjustment needed.
			{
				Line:      "16_col_b",
				Type:      forms.Computed,
				Label:     "SE tax deduction subtractions",
				DependsOn: []string{},
				Compute: func(dv forms.DepValues) float64 {
					return 0 // CA conforms to federal SE tax deduction
				},
			},

			// ===================================================================
			// Part II: Itemized Deduction Adjustments
			// ===================================================================
			// These fields compute CA adjustments to federal itemized deductions.
			// Key difference: CA does NOT allow a deduction for state/local
			// income taxes, and CA does NOT apply the federal $10,000 SALT cap.

			// Line 5a_col_b: Subtract state/local income tax deduction
			// CA does not allow a deduction for state income taxes paid
			{
				Line:      "5a_col_b",
				Type:      forms.Computed,
				Label:     "State income tax subtraction (not deductible in CA)",
				DependsOn: []string{"schedule_a:5a"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_a:5a")
				},
			},

			// Line 5e_col_b: Subtract the federal SALT amount
			// Federal Schedule A line 5e includes the SALT cap; CA needs to
			// remove the entire federal SALT and replace with CA's version
			{
				Line:      "5e_col_b",
				Type:      forms.Computed,
				Label:     "Federal SALT subtraction (CA recomputes without cap)",
				DependsOn: []string{"schedule_a:5e"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("schedule_a:5e")
				},
			},

			// Line 5e_col_c: Add back CA-allowed SALT (property taxes only, no cap)
			// CA allows property taxes (personal property + real estate) with no cap,
			// but does NOT allow state/local income tax deduction
			{
				Line:      "5e_col_c",
				Type:      forms.Computed,
				Label:     "CA SALT addition (property taxes only, no cap)",
				DependsOn: []string{"schedule_a:5b", "schedule_a:5c"},
				Compute: func(dv forms.DepValues) float64 {
					// CA SALT = property taxes only (no state income tax, no cap)
					return dv.Get("schedule_a:5b") + dv.Get("schedule_a:5c")
				},
			},

			// CA itemized deductions total adjustment (net of Part II)
			// This is the net change to apply to federal itemized deductions
			// Subtraction: remove federal SALT
			// Addition: add CA-allowed property taxes
			{
				Line:      "itemized_sub",
				Type:      forms.Computed,
				Label:     "Total itemized deduction subtractions",
				DependsOn: []string{"ca_schedule_ca:5e_col_b"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("ca_schedule_ca:5e_col_b")
				},
			},
			{
				Line:      "itemized_add",
				Type:      forms.Computed,
				Label:     "Total itemized deduction additions",
				DependsOn: []string{"ca_schedule_ca:5e_col_c"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("ca_schedule_ca:5e_col_c")
				},
			},

			// CA itemized deductions = federal itemized - subtractions + additions
			{
				Line:      "ca_itemized",
				Type:      forms.Computed,
				Label:     "California itemized deductions",
				DependsOn: []string{"schedule_a:17", "ca_schedule_ca:itemized_sub", "ca_schedule_ca:itemized_add"},
				Compute: func(dv forms.DepValues) float64 {
					federal := dv.Get("schedule_a:17")
					sub := dv.Get("ca_schedule_ca:itemized_sub")
					add := dv.Get("ca_schedule_ca:itemized_add")
					result := federal - sub + add
					if result < 0 {
						return 0
					}
					return result
				},
			},

			// ===================================================================
			// Totals
			// ===================================================================

			// Line 37, Column A: Federal amounts (mirrors federal AGI)
			{
				Line:      "37_col_a",
				Type:      forms.FederalRef,
				Label:     "Federal amounts (from Form 1040)",
				DependsOn: []string{"1040:11"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("1040:11")
				},
			},
			// Line 37, Column B: Total subtractions
			{
				Line:  "37_col_b",
				Type:  forms.Computed,
				Label: "Subtractions (Column B)",
				DependsOn: []string{
					"ca_schedule_ca:2_col_b",
					"ca_schedule_ca:3_col_b",
					"ca_schedule_ca:7_col_b",
					"ca_schedule_ca:12_col_b",
					"ca_schedule_ca:16_col_b",
				},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("ca_schedule_ca:2_col_b") +
						dv.Get("ca_schedule_ca:3_col_b") +
						dv.Get("ca_schedule_ca:7_col_b") +
						dv.Get("ca_schedule_ca:12_col_b") +
						dv.Get("ca_schedule_ca:16_col_b")
				},
			},
			// Line 37, Column C: Total additions
			{
				Line:  "37_col_c",
				Type:  forms.Computed,
				Label: "Additions (Column C)",
				DependsOn: []string{
					"ca_schedule_ca:2_col_c",
					"ca_schedule_ca:3_col_c",
					"ca_schedule_ca:7_col_c",
					"ca_schedule_ca:12_col_c",
					"ca_schedule_ca:15_col_c",
				},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("ca_schedule_ca:2_col_c") +
						dv.Get("ca_schedule_ca:3_col_c") +
						dv.Get("ca_schedule_ca:7_col_c") +
						dv.Get("ca_schedule_ca:12_col_c") +
						dv.Get("ca_schedule_ca:15_col_c")
				},
			},
		},
	}
}
