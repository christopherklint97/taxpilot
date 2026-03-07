package interview

import "fmt"

// caPreFillNotes maps field keys to functions that generate CA-specific
// pre-fill messages based on prior-year data.
var caPreFillNotes = map[string]func(priorYear map[string]float64, priorYearStr map[string]string) string{
	"w2:1:wages": func(py map[string]float64, _ map[string]string) string {
		stateWages := py["w2:1:state_wages"]
		wages := py["w2:1:wages"]
		if wages > 0 && stateWages > 0 && wages != stateWages {
			return fmt.Sprintf("Last year your CA wages (%s) differed from federal wages (%s)",
				formatCurrency(stateWages), formatCurrency(wages))
		}
		return ""
	},
	"w2:1:state_wages": func(py map[string]float64, _ map[string]string) string {
		stateWages := py["w2:1:state_wages"]
		wages := py["w2:1:wages"]
		if wages > 0 && stateWages > 0 && wages != stateWages {
			return "Your CA wages differed from federal last year — check your W-2 Box 16"
		}
		return ""
	},
	"schedule_a:5a": func(py map[string]float64, _ map[string]string) string {
		salt := py["schedule_a:5a"]
		if salt > 0 {
			return fmt.Sprintf("Last year CA removed your %s state income tax deduction on Schedule CA",
				formatCurrency(salt))
		}
		return "CA does not allow state income tax deductions — this will be removed on Schedule CA"
	},
	"schedule_a:5c": func(py map[string]float64, _ map[string]string) string {
		propTax := py["schedule_a:5c"]
		if propTax > 0 {
			return fmt.Sprintf("Last year CA allowed your full %s property tax deduction (no SALT cap)",
				formatCurrency(propTax))
		}
		return "CA allows property taxes with no cap (unlike the federal $10,000 SALT limit)"
	},
	"form_8889:2": func(py map[string]float64, _ map[string]string) string {
		hsaContrib := py["form_8889:2"]
		if hsaContrib > 0 {
			return fmt.Sprintf("Last year CA added back your %s HSA deduction on Schedule CA",
				formatCurrency(hsaContrib))
		}
		return "CA does not allow HSA deductions — contributions will be added back on Schedule CA"
	},
	"1099int:1:us_savings_bond_interest": func(py map[string]float64, _ map[string]string) string {
		bondInt := py["1099int:1:us_savings_bond_interest"]
		if bondInt > 0 {
			return fmt.Sprintf("Last year CA subtracted your %s U.S. bond interest on Schedule CA",
				formatCurrency(bondInt))
		}
		return "CA does not tax U.S. government bond interest — it will be subtracted on Schedule CA"
	},
	"1099int:1:tax_exempt_interest": func(_ map[string]float64, _ map[string]string) string {
		return "Only CA-issued municipal bond interest is exempt from CA tax; out-of-state muni interest is taxable"
	},
	"1099div:1:qualified_dividends": func(_ map[string]float64, _ map[string]string) string {
		return "CA taxes qualified dividends as ordinary income — no preferential rate"
	},
	"1099div:1:section_199a_dividends": func(_ map[string]float64, _ map[string]string) string {
		return "CA does not allow the Section 199A (QBI) deduction on these dividends"
	},
	"1099b:1:term": func(_ map[string]float64, _ map[string]string) string {
		return "CA taxes long-term capital gains as ordinary income — holding period still matters for federal"
	},
}

// caScheduleCANote returns a summary of prior-year Schedule CA adjustments.
func caScheduleCANote(priorYear map[string]float64) string {
	subtractions := priorYear["ca_schedule_ca:37_col_b"]
	additions := priorYear["ca_schedule_ca:37_col_c"]

	if subtractions == 0 && additions == 0 {
		return "CA made no Schedule CA adjustments last year"
	}

	msg := ""
	if subtractions > 0 {
		msg += fmt.Sprintf("Last year CA subtracted %s", formatCurrency(subtractions))
	}
	if additions > 0 {
		if msg != "" {
			msg += " and "
		}
		msg += fmt.Sprintf("added back %s on Schedule CA", formatCurrency(additions))
	} else if msg != "" {
		msg += " on Schedule CA"
	}
	return msg
}

// GetCAPreFillNote returns a CA-specific message for a field's prior-year default.
// Returns empty string if no CA note is applicable.
func GetCAPreFillNote(fieldKey string, priorYear map[string]float64, priorYearStr map[string]string) string {
	if fn, ok := caPreFillNotes[fieldKey]; ok {
		return fn(priorYear, priorYearStr)
	}
	return ""
}
