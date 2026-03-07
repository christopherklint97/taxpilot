package errors

import (
	"fmt"
	"strings"
)

// UnsupportedError indicates a tax situation TaxPilot can't handle.
type UnsupportedError struct {
	Situation  string // e.g., "Non-resident alien filing"
	Reason     string // e.g., "TaxPilot only supports resident filers"
	Suggestion string // e.g., "Consult a CPA or use Form 1040-NR"
}

func (e *UnsupportedError) Error() string {
	return fmt.Sprintf("Unsupported: %s — %s", e.Situation, e.Reason)
}

// IncompleteError indicates missing required data.
type IncompleteError struct {
	MissingFields []string // e.g., ["1040:ssn", "1040:first_name"]
	Message       string
}

func (e *IncompleteError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("Incomplete: missing %d required field(s): %s",
		len(e.MissingFields), strings.Join(e.MissingFields, ", "))
}

// ConformityError indicates a CA conformity issue that needs attention.
type ConformityError struct {
	FederalField string // e.g., "form_8889:2"
	CAField      string // e.g., "ca_schedule_ca:15_col_c"
	Situation    string // e.g., "HSA deduction"
	Message      string // e.g., "California does not conform to federal HSA deduction..."
}

func (e ConformityError) Error() string {
	return fmt.Sprintf("CA Conformity [%s]: %s", e.Situation, e.Message)
}

// CPAReferralError indicates a situation too complex for automated filing.
type CPAReferralError struct {
	Situation  string
	Reason     string
	Complexity string // "high", "very_high"
}

func (e *CPAReferralError) Error() string {
	return fmt.Sprintf("CPA Referral (%s complexity): %s — %s", e.Complexity, e.Situation, e.Reason)
}

// ---------------------------------------------------------------------------
// Helper functions (mirrors conventions in efile/validate.go)
// ---------------------------------------------------------------------------

func getStr(strInputs map[string]string, key string) string {
	if strInputs == nil {
		return ""
	}
	return strInputs[key]
}

func getNum(results map[string]float64, key string) float64 {
	if results == nil {
		return 0
	}
	return results[key]
}

func numExists(results map[string]float64, key string) bool {
	if results == nil {
		return false
	}
	_, ok := results[key]
	return ok
}

func hasKeyPrefix(m map[string]float64, prefix string) bool {
	for k := range m {
		if strings.HasPrefix(k, prefix) {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Situation checkers
// ---------------------------------------------------------------------------

// CheckUnsupported examines inputs and returns any UnsupportedErrors.
// Checks for:
//   - MFS with complex situations (marriage CPA referral)
//   - AMT triggers (very high income with specific deductions)
//   - Multiple states (only CA supported)
func CheckUnsupported(results map[string]float64, strInputs map[string]string) []error {
	var errs []error

	// MFS with itemized deductions — complex community property rules
	fs := getStr(strInputs, "1040:filing_status")
	if fs == "mfs" {
		itemized := getNum(results, "schedule_a:17")
		if itemized > 0 {
			errs = append(errs, &UnsupportedError{
				Situation:  "Married Filing Separately with itemized deductions",
				Reason:     "Community property rules in California make MFS with itemized deductions complex",
				Suggestion: "Consider filing jointly, or consult a CPA for MFS with itemized deductions",
			})
		}
	}

	// AMT trigger — very high income with large SALT or other preferences
	agi := getNum(results, "1040:11")
	salt := getNum(results, "schedule_a:5d")
	if agi > 500000 && salt >= 10000 {
		errs = append(errs, &UnsupportedError{
			Situation:  "Potential Alternative Minimum Tax (AMT)",
			Reason:     "High income with capped SALT deduction may trigger AMT, which TaxPilot does not compute",
			Suggestion: "Review Form 6251 with a tax professional to determine if AMT applies",
		})
	}

	// Multiple states — only CA supported
	state := getStr(strInputs, "1040:state")
	if state != "" && state != "CA" && state != "ca" {
		errs = append(errs, &UnsupportedError{
			Situation:  fmt.Sprintf("State filing for %s", strings.ToUpper(state)),
			Reason:     "TaxPilot currently only supports California state returns",
			Suggestion: "File your federal return with TaxPilot and use another tool for your state return",
		})
	}

	return errs
}

// CheckIncomplete validates that all critical fields are present.
// Returns an IncompleteError listing missing fields, or nil if complete.
func CheckIncomplete(results map[string]float64, strInputs map[string]string) *IncompleteError {
	var missing []string

	// Required string fields
	requiredStr := []string{
		"1040:ssn",
		"1040:first_name",
		"1040:last_name",
		"1040:filing_status",
	}
	for _, key := range requiredStr {
		if getStr(strInputs, key) == "" {
			missing = append(missing, key)
		}
	}

	// Required numeric fields (must exist in results)
	requiredNum := []string{
		"1040:9",  // total income
		"1040:11", // AGI
		"1040:15", // taxable income
		"1040:24", // total tax
	}
	for _, key := range requiredNum {
		if !numExists(results, key) {
			missing = append(missing, key)
		}
	}

	if len(missing) == 0 {
		return nil
	}

	return &IncompleteError{
		MissingFields: missing,
		Message:       fmt.Sprintf("Return is incomplete: %d required field(s) missing: %s", len(missing), strings.Join(missing, ", ")),
	}
}

// CheckCAConformity validates CA-federal conformity issues and returns
// informational ConformityErrors about adjustments that were applied.
func CheckCAConformity(results map[string]float64) []ConformityError {
	var errs []ConformityError

	// HSA deduction — California does not conform
	hsaDeduction := getNum(results, "form_8889:13")
	if hsaDeduction > 0 {
		errs = append(errs, ConformityError{
			FederalField: "form_8889:13",
			CAField:      "ca_schedule_ca:13_col_b",
			Situation:    "HSA deduction",
			Message: fmt.Sprintf(
				"California does not conform to the federal HSA deduction of $%.0f. "+
					"An add-back is required on Schedule CA, Line 13, Column B.",
				hsaDeduction),
		})
	}

	// QBI deduction (Section 199A) — California does not conform
	qbiDeduction := getNum(results, "form_8995:15")
	if qbiDeduction > 0 {
		errs = append(errs, ConformityError{
			FederalField: "form_8995:15",
			CAField:      "ca_schedule_ca:13_col_b",
			Situation:    "QBI deduction (Section 199A)",
			Message: fmt.Sprintf(
				"California does not allow the federal Qualified Business Income deduction of $%.0f. "+
					"An add-back is required on Schedule CA, Line 13, Column B.",
				qbiDeduction),
		})
	}

	// Social Security — CA does not tax, so if federal includes SS income, CA subtracts it
	ssIncome := getNum(results, "1040:6b")
	if ssIncome > 0 {
		errs = append(errs, ConformityError{
			FederalField: "1040:6b",
			CAField:      "ca_schedule_ca:6a_col_c",
			Situation:    "Social Security benefits",
			Message: fmt.Sprintf(
				"California does not tax Social Security benefits. The $%.0f included in federal income "+
					"is subtracted on Schedule CA, Line 6a, Column C.",
				ssIncome),
		})
	}

	// Municipal bond interest — out-of-state bonds taxable in CA
	exemptInterest := getNum(results, "1040:2a")
	if exemptInterest > 0 {
		errs = append(errs, ConformityError{
			FederalField: "1040:2a",
			CAField:      "ca_schedule_ca:2a_col_b",
			Situation:    "Tax-exempt interest",
			Message: fmt.Sprintf(
				"California only exempts interest from CA-issued bonds. If the $%.0f in tax-exempt interest "+
					"includes out-of-state municipal bonds, those amounts must be added back on Schedule CA.",
				exemptInterest),
		})
	}

	return errs
}

// CheckComplexity assesses whether the return is too complex for TaxPilot.
// Returns a CPAReferralError if the situation warrants professional help.
// Triggers:
//   - AMT (if AGI > $500k with itemized deductions)
//   - Estate/trust income
//   - Partnership/S-Corp (K-1)
func CheckComplexity(results map[string]float64, strInputs map[string]string) *CPAReferralError {
	agi := getNum(results, "1040:11")

	// K-1 / Partnership / S-Corp income
	if hasKeyPrefix(results, "k1:") || getNum(results, "schedule_1:5_partnership") > 0 {
		return &CPAReferralError{
			Situation:  "Partnership or S-Corporation income (Schedule K-1)",
			Reason:     "K-1 income involves complex allocation rules, basis tracking, and at-risk/passive activity limitations that require professional review",
			Complexity: "very_high",
		}
	}

	// Estate/trust income
	if getNum(results, "schedule_1:5_estate_trust") > 0 {
		return &CPAReferralError{
			Situation:  "Estate or trust income",
			Reason:     "Estate and trust income has complex distribution rules and separate tax brackets that TaxPilot cannot handle",
			Complexity: "high",
		}
	}

	// Very high income with itemized deductions — AMT risk
	itemized := getNum(results, "schedule_a:17")
	if agi > 500000 && itemized > 0 {
		return &CPAReferralError{
			Situation:  "Potential Alternative Minimum Tax",
			Reason:     "High-income filers with itemized deductions may owe AMT; Form 6251 analysis is needed",
			Complexity: "high",
		}
	}

	return nil
}

// ---------------------------------------------------------------------------
// User-friendly formatters
// ---------------------------------------------------------------------------

// FormatForUser converts an error into a clear, actionable message.
// Uses plain English, no tax jargon where possible.
func FormatForUser(err error) string {
	switch e := err.(type) {
	case *UnsupportedError:
		var sb strings.Builder
		sb.WriteString("NOT SUPPORTED: ")
		sb.WriteString(e.Situation)
		sb.WriteString("\n  Why: ")
		sb.WriteString(e.Reason)
		if e.Suggestion != "" {
			sb.WriteString("\n  What to do: ")
			sb.WriteString(e.Suggestion)
		}
		return sb.String()

	case *IncompleteError:
		var sb strings.Builder
		sb.WriteString("MISSING INFORMATION\n")
		sb.WriteString("  The following required fields are not yet filled in:\n")
		for _, f := range e.MissingFields {
			sb.WriteString("    - ")
			sb.WriteString(friendlyFieldName(f))
			sb.WriteString("\n")
		}
		sb.WriteString("  Please go back and provide this information before continuing.")
		return sb.String()

	case ConformityError:
		var sb strings.Builder
		sb.WriteString("CALIFORNIA ADJUSTMENT: ")
		sb.WriteString(e.Situation)
		sb.WriteString("\n  ")
		sb.WriteString(e.Message)
		return sb.String()

	case *ConformityError:
		var sb strings.Builder
		sb.WriteString("CALIFORNIA ADJUSTMENT: ")
		sb.WriteString(e.Situation)
		sb.WriteString("\n  ")
		sb.WriteString(e.Message)
		return sb.String()

	case *CPAReferralError:
		var sb strings.Builder
		sb.WriteString("PROFESSIONAL HELP RECOMMENDED: ")
		sb.WriteString(e.Situation)
		sb.WriteString("\n  Why: ")
		sb.WriteString(e.Reason)
		sb.WriteString("\n  We recommend consulting a CPA or enrolled agent for this situation.")
		return sb.String()

	default:
		return err.Error()
	}
}

// FormatAllIssues takes a slice of errors and formats them into a
// structured report with sections for each severity.
func FormatAllIssues(errs []error) string {
	if len(errs) == 0 {
		return "No issues found. Your return looks good!"
	}

	var unsupported []error
	var incomplete []error
	var conformity []error
	var referral []error
	var other []error

	for _, err := range errs {
		switch err.(type) {
		case *UnsupportedError:
			unsupported = append(unsupported, err)
		case *IncompleteError:
			incomplete = append(incomplete, err)
		case ConformityError:
			conformity = append(conformity, err)
		case *ConformityError:
			conformity = append(conformity, err)
		case *CPAReferralError:
			referral = append(referral, err)
		default:
			other = append(other, err)
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d issue(s) with your return:\n", len(errs)))

	if len(unsupported) > 0 {
		sb.WriteString("\n--- Unsupported Situations ---\n")
		for i, e := range unsupported {
			sb.WriteString(fmt.Sprintf("\n%d. %s\n", i+1, FormatForUser(e)))
		}
	}

	if len(incomplete) > 0 {
		sb.WriteString("\n--- Missing Information ---\n")
		for i, e := range incomplete {
			sb.WriteString(fmt.Sprintf("\n%d. %s\n", i+1, FormatForUser(e)))
		}
	}

	if len(referral) > 0 {
		sb.WriteString("\n--- Professional Help Recommended ---\n")
		for i, e := range referral {
			sb.WriteString(fmt.Sprintf("\n%d. %s\n", i+1, FormatForUser(e)))
		}
	}

	if len(conformity) > 0 {
		sb.WriteString("\n--- California Adjustments (Informational) ---\n")
		for i, e := range conformity {
			sb.WriteString(fmt.Sprintf("\n%d. %s\n", i+1, FormatForUser(e)))
		}
	}

	if len(other) > 0 {
		sb.WriteString("\n--- Other Issues ---\n")
		for i, e := range other {
			sb.WriteString(fmt.Sprintf("\n%d. %s\n", i+1, e.Error()))
		}
	}

	return sb.String()
}

// friendlyFieldName converts internal field keys to human-readable names.
func friendlyFieldName(key string) string {
	names := map[string]string{
		"1040:ssn":           "Social Security Number",
		"1040:first_name":    "First Name",
		"1040:last_name":     "Last Name",
		"1040:filing_status": "Filing Status",
		"1040:9":             "Total Income (Form 1040, Line 9)",
		"1040:11":            "Adjusted Gross Income (Form 1040, Line 11)",
		"1040:15":            "Taxable Income (Form 1040, Line 15)",
		"1040:24":            "Total Tax (Form 1040, Line 24)",
	}
	if friendly, ok := names[key]; ok {
		return friendly
	}
	return key
}
