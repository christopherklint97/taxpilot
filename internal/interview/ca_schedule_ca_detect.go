package interview

import "taxpilot/internal/forms"

// DetectCAScheduleCANeeded examines the taxpayer's inputs and determines whether
// Schedule CA (California Adjustments) is needed, and returns the reasons why.
// This helps the interview engine decide whether to include Schedule CA questions.
func DetectCAScheduleCANeeded(inputs map[string]float64, strInputs map[string]string) (bool, []string) {
	var reasons []string

	// HSA contributions: CA does not conform to federal HSA treatment.
	// Any HSA deduction must be added back on Schedule CA.
	if inputs[forms.F8889Line2] > 0 {
		reasons = append(reasons, "HSA deduction add-back")
	}

	// SALT deduction: CA does not allow deduction of state/local income taxes.
	// Schedule A line 5a (state and local income taxes) must be removed.
	if inputs[forms.SchedALine5a] > 0 {
		reasons = append(reasons, "State income tax deduction removal")
	}

	// QBI deduction: CA does not conform to the federal Section 199A QBI deduction.
	// Detect via presence of business income on Schedule C.
	if inputs[forms.SchedCLine31] > 0 || inputs[forms.SchedCLine7] > 0 {
		reasons = append(reasons, "QBI deduction add-back")
	}

	// U.S. savings bond interest: exempt from state tax in California.
	// This is a subtraction (reduces CA income).
	if inputs["1099int:1:us_savings_bond_interest"] > 0 {
		reasons = append(reasons, "U.S. bond interest subtraction")
	}

	// Out-of-state municipal bond interest: taxable in CA even though
	// it may be tax-exempt federally. Check for tax-exempt interest
	// that is NOT from California municipal bonds.
	if inputs["1099int:1:tax_exempt_interest"] > 0 {
		// If they have tax-exempt interest, check if it's from out-of-state
		state := strInputs["1099int:1:bond_state"]
		if state != "" && state != "CA" {
			reasons = append(reasons, "Out-of-state municipal bond interest")
		}
	}

	// State vs federal wage differences: if W-2 state wages differ from
	// federal wages, a Schedule CA adjustment may be needed.
	federalWages := inputs["w2:1:wages"]
	stateWages := inputs["w2:1:state_wages"]
	if federalWages > 0 && stateWages > 0 && federalWages != stateWages {
		reasons = append(reasons, "State wage adjustment")
	}

	// Foreign earned income exclusion: CA does not conform to FEIE.
	// The entire exclusion must be added back on Schedule CA.
	if inputs[forms.F2555TotalExclusion] > 0 || inputs[forms.F2555ForeignEarnedIncome] > 0 {
		reasons = append(reasons, "Foreign earned income exclusion add-back (CA does not allow FEIE)")
	}

	// Foreign housing exclusion/deduction: CA does not conform
	if inputs[forms.F2555HousingExclusion] > 0 || inputs[forms.F2555HousingDeduction] > 0 {
		reasons = append(reasons, "Foreign housing exclusion/deduction add-back")
	}

	needed := len(reasons) > 0
	return needed, reasons
}
