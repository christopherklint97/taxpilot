package interview

// FBAR (FinCEN 114) is filed separately with the Financial Crimes
// Enforcement Network (FinCEN), NOT with the IRS as part of the tax return.
// TaxPilot detects when FBAR filing is required and provides guidance,
// but does not generate the FinCEN 114 filing itself.

const fbarThreshold = 10000 // $10,000 aggregate value

// FBARRequired returns true if the taxpayer must file an FBAR based on
// their foreign account values. An FBAR is required if the aggregate
// value of all foreign financial accounts exceeded $10,000 at any time
// during the calendar year.
func FBARRequired(maxAggregateValue float64) bool {
	return maxAggregateValue > fbarThreshold
}

// FBARDeadline returns the FBAR filing deadline for the given tax year.
// The FBAR is due April 15 following the calendar year, with an automatic
// extension to October 15 (no request needed).
func FBARDeadline(taxYear int) string {
	nextYear := taxYear + 1
	return formatFBARDeadline(nextYear)
}

func formatFBARDeadline(year int) string {
	return "April 15, " + itoa(year) + " (automatic extension to October 15, " + itoa(year) + ")"
}

// itoa converts an int to a string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}
	return digits
}

// FBARGuidanceMessage returns guidance text explaining the FBAR requirement.
func FBARGuidanceMessage() string {
	return "IMPORTANT: You are required to file FinCEN Form 114 (FBAR) because " +
		"the aggregate value of your foreign financial accounts exceeded $10,000 " +
		"at some point during the year.\n\n" +
		"The FBAR is filed electronically through the BSA E-Filing System at " +
		"https://bsaefiling.fincen.treas.gov — it is NOT filed with your tax return.\n\n" +
		"The FBAR is due April 15 following the calendar year, with an automatic " +
		"extension to October 15 (no request needed).\n\n" +
		"Failure to file an FBAR can result in civil penalties up to $10,000 per " +
		"violation (or up to $100,000 or 50% of account balances for willful violations)."
}

// FBARNotRequiredMessage returns a message when FBAR is not needed.
func FBARNotRequiredMessage() string {
	return "Based on your foreign account values, you do not need to file an FBAR " +
		"(FinCEN Form 114) for this tax year. The FBAR is required only when the " +
		"aggregate value of all foreign financial accounts exceeds $10,000 at any " +
		"time during the year."
}
