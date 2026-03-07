package efile

import (
	"math"
	"strings"
)

// Severity indicates how serious a validation issue is.
type Severity int

const (
	SeverityError   Severity = iota // blocks e-filing
	SeverityWarning                 // user should review but can proceed
	SeverityInfo                    // informational
)

// ValidationResult represents a single validation finding.
type ValidationResult struct {
	Code     string   // e.g., "R0001", "W0001"
	Severity Severity
	Field    string   // field key, e.g., "1040:15"
	Message  string   // human-readable message
}

// ValidationReport holds all validation results.
type ValidationReport struct {
	Results []ValidationResult
	IsValid bool // true if no errors (warnings are OK)
}

func (r *ValidationReport) addResult(code string, sev Severity, field, message string) {
	r.Results = append(r.Results, ValidationResult{
		Code:     code,
		Severity: sev,
		Field:    field,
		Message:  message,
	})
}

func (r *ValidationReport) computeValidity() {
	r.IsValid = true
	for _, res := range r.Results {
		if res.Severity == SeverityError {
			r.IsValid = false
			return
		}
	}
}

// hasW2Wages checks whether any w2:*:wages key has a positive value.
func hasW2Wages(results map[string]float64) bool {
	for k, v := range results {
		if strings.HasPrefix(k, "w2:") && strings.HasSuffix(k, ":wages") && v > 0 {
			return true
		}
	}
	return false
}

// getStr retrieves a string input, returning empty string if not found.
func getStr(strInputs map[string]string, key string) string {
	if strInputs == nil {
		return ""
	}
	return strInputs[key]
}

// getNum retrieves a numeric result, returning 0 if not found.
func getNum(results map[string]float64, key string) float64 {
	if results == nil {
		return 0
	}
	return results[key]
}

// numExists checks whether a key exists in the results map.
func numExists(results map[string]float64, key string) bool {
	if results == nil {
		return false
	}
	_, ok := results[key]
	return ok
}

// ValidateReturn validates a federal return for e-filing readiness.
func ValidateReturn(results map[string]float64, strInputs map[string]string, taxYear int) ValidationReport {
	var report ValidationReport

	// R0001: SSN required
	if getStr(strInputs, "1040:ssn") == "" {
		report.addResult("R0001", SeverityError, "1040:ssn", "SSN is required for e-filing")
	}

	// R0002: Filing status required and valid
	fs := getStr(strInputs, "1040:filing_status")
	validStatuses := map[string]bool{
		"single": true, "mfj": true, "mfs": true, "hoh": true, "qss": true,
	}
	if fs == "" || !validStatuses[fs] {
		report.addResult("R0002", SeverityError, "1040:filing_status",
			"Filing status is required and must be one of: single, mfj, mfs, hoh, qss")
	}

	// R0003: First name required
	if getStr(strInputs, "1040:first_name") == "" {
		report.addResult("R0003", SeverityError, "1040:first_name", "First name is required")
	}

	// R0004: Last name required
	if getStr(strInputs, "1040:last_name") == "" {
		report.addResult("R0004", SeverityError, "1040:last_name", "Last name is required")
	}

	// R0005: Total income must be non-negative
	if numExists(results, "1040:9") && getNum(results, "1040:9") < 0 {
		report.addResult("R0005", SeverityError, "1040:9", "Total income must be non-negative")
	}

	// R0006: Taxable income must not exceed total income
	if numExists(results, "1040:15") && numExists(results, "1040:9") {
		if getNum(results, "1040:15") > getNum(results, "1040:9") {
			report.addResult("R0006", SeverityError, "1040:15",
				"Taxable income cannot exceed total income")
		}
	}

	// R0007: Total tax must be non-negative
	if numExists(results, "1040:24") && getNum(results, "1040:24") < 0 {
		report.addResult("R0007", SeverityError, "1040:24", "Total tax must be non-negative")
	}

	// R0008: Withholding must match sum (25d == 25a + 25b within $1)
	if numExists(results, "1040:25d") {
		sum := getNum(results, "1040:25a") + getNum(results, "1040:25b")
		if math.Abs(getNum(results, "1040:25d")-sum) > 1.0 {
			report.addResult("R0008", SeverityError, "1040:25d",
				"Total withholding (line 25d) must equal sum of lines 25a and 25b (within $1)")
		}
	}

	// R0009: Refund and amount owed can't both be positive
	refund := getNum(results, "1040:34")
	owed := getNum(results, "1040:37")
	if refund > 0 && owed > 0 {
		report.addResult("R0009", SeverityError, "1040:34",
			"Refund and amount owed cannot both be positive")
	}

	// R0010: Must have some income source
	hasW2 := hasW2Wages(results)
	hasScheduleC := getNum(results, "schedule_c:31") > 0
	if !hasW2 && !hasScheduleC {
		report.addResult("R0010", SeverityError, "w2:1:wages",
			"At least one W-2 with wages or Schedule C net profit is required")
	}

	// W0001: Charitable donations > 60% of AGI
	agi := getNum(results, "1040:11")
	charitable := getNum(results, "schedule_a:12")
	if agi > 0 && charitable > 0.6*agi {
		report.addResult("W0001", SeverityWarning, "schedule_a:12",
			"Charitable donations exceed 60% of AGI, which may trigger audit scrutiny")
	}

	// W0002: Medical expenses > 20% of AGI
	medical := getNum(results, "schedule_a:1")
	if agi > 0 && medical > 0.2*agi {
		report.addResult("W0002", SeverityWarning, "schedule_a:1",
			"Medical expenses exceed 20% of AGI, which is unusually high")
	}

	// W0003: Business expense ratio > 80% of revenue
	bizRevenue := getNum(results, "schedule_c:7")
	bizExpenses := getNum(results, "schedule_c:28")
	if bizRevenue > 0 && bizExpenses > 0.8*bizRevenue {
		report.addResult("W0003", SeverityWarning, "schedule_c:28",
			"Business expenses exceed 80% of revenue, which may trigger audit scrutiny")
	}

	// W0004: SALT deduction at cap
	salt := getNum(results, "schedule_a:5d")
	if salt > 0 {
		cap := 10000.0
		if fs == "mfs" {
			cap = 5000.0
		}
		if salt >= cap {
			report.addResult("W0004", SeverityInfo, "schedule_a:5d",
				"SALT deduction is at the cap limit")
		}
	}

	report.computeValidity()
	return report
}

// ValidateCAReturn validates a California state return for e-filing readiness.
func ValidateCAReturn(results map[string]float64, strInputs map[string]string, taxYear int) ValidationReport {
	var report ValidationReport

	// R1001: CA AGI required
	if !numExists(results, "ca_540:17") {
		report.addResult("R1001", SeverityError, "ca_540:17",
			"California AGI is required")
	}

	// R1002: CA total tax must be non-negative
	if numExists(results, "ca_540:40") && getNum(results, "ca_540:40") < 0 {
		report.addResult("R1002", SeverityError, "ca_540:40",
			"California total tax must be non-negative")
	}

	// R1003: CA refund and amount owed can't both be positive
	caRefund := getNum(results, "ca_540:75")
	caOwed := getNum(results, "ca_540:81")
	if caRefund > 0 && caOwed > 0 {
		report.addResult("R1003", SeverityError, "ca_540:75",
			"California refund and amount owed cannot both be positive")
	}

	// W1001: CA AGI differs from federal AGI by > $100,000
	if numExists(results, "ca_540:17") && numExists(results, "1040:11") {
		diff := math.Abs(getNum(results, "ca_540:17") - getNum(results, "1040:11"))
		if diff > 100000 {
			report.addResult("W1001", SeverityWarning, "ca_540:17",
				"California AGI differs from federal AGI by more than $100,000")
		}
	}

	// W1002: CA effective tax rate > 13.3%
	caAGI := getNum(results, "ca_540:17")
	caTax := getNum(results, "ca_540:40")
	if caAGI > 0 && caTax/caAGI > 0.133 {
		report.addResult("W1002", SeverityWarning, "ca_540:40",
			"California effective tax rate exceeds 13.3%; verify mental health surcharge is correct")
	}

	report.computeValidity()
	return report
}

// ValidateFull runs both federal and CA validation, merging results.
func ValidateFull(results map[string]float64, strInputs map[string]string, taxYear int, includeCA bool) ValidationReport {
	federal := ValidateReturn(results, strInputs, taxYear)

	if !includeCA {
		return federal
	}

	ca := ValidateCAReturn(results, strInputs, taxYear)

	merged := ValidationReport{
		Results: append(federal.Results, ca.Results...),
	}
	merged.computeValidity()
	return merged
}
