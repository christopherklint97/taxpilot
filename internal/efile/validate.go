package efile

import (
	"math"

	"taxpilot/internal/forms"
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
	Field    string   // field key, e.g., forms.F1040Line15
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
		if strings.HasPrefix(k, string(forms.FormW2)+":") && strings.HasSuffix(k, ":"+forms.W2Wages) && v > 0 {
			return true
		}
	}
	return false
}

// ValidateReturn validates a federal return for e-filing readiness.
func ValidateReturn(results map[string]float64, strInputs map[string]string, taxYear int) ValidationReport {
	var report ValidationReport

	// R0001: SSN required
	if forms.GetStr(strInputs, forms.F1040SSN) == "" {
		report.addResult("R0001", SeverityError, forms.F1040SSN, "SSN is required for e-filing")
	}

	// R0002: Filing status required and valid
	fs := forms.GetStr(strInputs, forms.F1040FilingStatus)
	validStatuses := map[string]bool{
		"single": true, "mfj": true, "mfs": true, "hoh": true, "qss": true,
	}
	if fs == "" || !validStatuses[fs] {
		report.addResult("R0002", SeverityError, forms.F1040FilingStatus,
			"Filing status is required and must be one of: single, mfj, mfs, hoh, qss")
	}

	// R0003: First name required
	if forms.GetStr(strInputs, forms.F1040FirstName) == "" {
		report.addResult("R0003", SeverityError, forms.F1040FirstName, "First name is required")
	}

	// R0004: Last name required
	if forms.GetStr(strInputs, forms.F1040LastName) == "" {
		report.addResult("R0004", SeverityError, forms.F1040LastName, "Last name is required")
	}

	// R0005: Total income must be non-negative
	if forms.NumExists(results, forms.F1040Line9) && forms.GetNum(results, forms.F1040Line9) < 0 {
		report.addResult("R0005", SeverityError, forms.F1040Line9, "Total income must be non-negative")
	}

	// R0006: Taxable income must not exceed total income
	if forms.NumExists(results, forms.F1040Line15) && forms.NumExists(results, forms.F1040Line9) {
		if forms.GetNum(results, forms.F1040Line15) > forms.GetNum(results, forms.F1040Line9) {
			report.addResult("R0006", SeverityError, forms.F1040Line15,
				"Taxable income cannot exceed total income")
		}
	}

	// R0007: Total tax must be non-negative
	if forms.NumExists(results, forms.F1040Line24) && forms.GetNum(results, forms.F1040Line24) < 0 {
		report.addResult("R0007", SeverityError, forms.F1040Line24, "Total tax must be non-negative")
	}

	// R0008: Withholding must match sum (25d == 25a + 25b within $1)
	if forms.NumExists(results, forms.F1040Line25d) {
		sum := forms.GetNum(results, forms.F1040Line25a) + forms.GetNum(results, forms.F1040Line25b)
		if math.Abs(forms.GetNum(results, forms.F1040Line25d)-sum) > 1.0 {
			report.addResult("R0008", SeverityError, forms.F1040Line25d,
				"Total withholding (line 25d) must equal sum of lines 25a and 25b (within $1)")
		}
	}

	// R0009: Refund and amount owed can't both be positive
	refund := forms.GetNum(results, forms.F1040Line34)
	owed := forms.GetNum(results, forms.F1040Line37)
	if refund > 0 && owed > 0 {
		report.addResult("R0009", SeverityError, forms.F1040Line34,
			"Refund and amount owed cannot both be positive")
	}

	// R0010: Must have some income source (W-2, Schedule C, or Form 2555 foreign income)
	hasW2 := hasW2Wages(results)
	hasScheduleC := forms.GetNum(results, forms.SchedCLine31) > 0
	hasForeignIncome := forms.GetNum(results, forms.F2555ForeignEarnedIncome) > 0
	if !hasW2 && !hasScheduleC && !hasForeignIncome {
		report.addResult("R0010", SeverityError, forms.FK(forms.FormW2, "1:wages"),
			"At least one W-2 with wages, Schedule C net profit, or foreign earned income is required")
	}

	// R0011: Form 2555 requires qualifying test
	if hasForeignIncome {
		qt := forms.GetStr(strInputs, forms.F2555QualifyingTest)
		if qt != "bona_fide_residence" && qt != "physical_presence" {
			report.addResult("R0011", SeverityError, forms.F2555QualifyingTest,
				"Form 2555 requires a qualifying test (bona_fide_residence or physical_presence)")
		}
	}

	// R0012: Physical presence test requires >= 330 days
	if forms.GetStr(strInputs, forms.F2555QualifyingTest) == "physical_presence" {
		days := forms.GetNum(results, forms.F2555QualifyingDays)
		if days < 330 {
			report.addResult("R0012", SeverityError, forms.F2555PPTDaysPresent,
				"Physical presence test requires at least 330 days in a foreign country")
		}
	}

	// R0013: FEIE cannot exceed the limit
	exclusion := forms.GetNum(results, forms.F2555TotalExclusion)
	limit := forms.GetNum(results, forms.F2555ExclusionLimit)
	if exclusion > 0 && limit > 0 && exclusion > limit+1 {
		report.addResult("R0013", SeverityError, forms.F2555TotalExclusion,
			"FEIE exclusion exceeds the annual limit")
	}

	// R0014: Form 8938 required if threshold met but not filed
	if forms.NumExists(results, forms.F8938FilingRequired) && forms.GetNum(results, forms.F8938FilingRequired) == 1 {
		totalMax := forms.GetNum(results, forms.F8938TotalMaxValue)
		if totalMax == 0 {
			report.addResult("R0014", SeverityError, forms.F8938TotalMaxValue,
				"Form 8938 filing is required but foreign asset values appear incomplete")
		}
	}

	// W0005: FBAR reminder if foreign accounts > $10,000
	if forms.GetStr(strInputs, forms.SchedBLine7a) == "yes" {
		report.addResult("W0005", SeverityInfo, forms.SchedBLine7a,
			"You indicated foreign accounts — remember to file FBAR (FinCEN 114) separately at bsaefiling.fincen.treas.gov if aggregate value exceeded $10,000")
	}

	// W0001: Charitable donations > 60% of AGI
	agi := forms.GetNum(results, forms.F1040Line11)
	charitable := forms.GetNum(results, forms.SchedALine12)
	if agi > 0 && charitable > 0.6*agi {
		report.addResult("W0001", SeverityWarning, forms.SchedALine12,
			"Charitable donations exceed 60% of AGI, which may trigger audit scrutiny")
	}

	// W0002: Medical expenses > 20% of AGI
	medical := forms.GetNum(results, forms.SchedALine1)
	if agi > 0 && medical > 0.2*agi {
		report.addResult("W0002", SeverityWarning, forms.SchedALine1,
			"Medical expenses exceed 20% of AGI, which is unusually high")
	}

	// W0003: Business expense ratio > 80% of revenue
	bizRevenue := forms.GetNum(results, forms.SchedCLine7)
	bizExpenses := forms.GetNum(results, forms.SchedCLine28)
	if bizRevenue > 0 && bizExpenses > 0.8*bizRevenue {
		report.addResult("W0003", SeverityWarning, forms.SchedCLine28,
			"Business expenses exceed 80% of revenue, which may trigger audit scrutiny")
	}

	// W0004: SALT deduction at cap
	salt := forms.GetNum(results, forms.SchedALine5d)
	if salt > 0 {
		cap := 10000.0
		if fs == "mfs" {
			cap = 5000.0
		}
		if salt >= cap {
			report.addResult("W0004", SeverityInfo, forms.SchedALine5d,
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
	if !forms.NumExists(results, forms.CA540Line17) {
		report.addResult("R1001", SeverityError, forms.CA540Line17,
			"California AGI is required")
	}

	// R1002: CA total tax must be non-negative
	if forms.NumExists(results, forms.CA540Line40) && forms.GetNum(results, forms.CA540Line40) < 0 {
		report.addResult("R1002", SeverityError, forms.CA540Line40,
			"California total tax must be non-negative")
	}

	// R1003: CA refund and amount owed can't both be positive
	caRefund := forms.GetNum(results, forms.CA540Line75)
	caOwed := forms.GetNum(results, forms.CA540Line81)
	if caRefund > 0 && caOwed > 0 {
		report.addResult("R1003", SeverityError, forms.CA540Line75,
			"California refund and amount owed cannot both be positive")
	}

	// W1001: CA AGI differs from federal AGI by > $100,000
	if forms.NumExists(results, forms.CA540Line17) && forms.NumExists(results, forms.F1040Line11) {
		diff := math.Abs(forms.GetNum(results, forms.CA540Line17) - forms.GetNum(results, forms.F1040Line11))
		if diff > 100000 {
			report.addResult("W1001", SeverityWarning, forms.CA540Line17,
				"California AGI differs from federal AGI by more than $100,000")
		}
	}

	// R1004: CA Schedule CA must include FEIE add-back when FEIE is claimed
	feieExclusion := forms.GetNum(results, forms.F2555TotalExclusion)
	feieAddBack := forms.GetNum(results, forms.SchedCALine8dColC)
	if feieExclusion > 0 && math.Abs(feieAddBack-feieExclusion) > 1.0 {
		report.addResult("R1004", SeverityError, forms.SchedCALine8dColC,
			"CA Schedule CA FEIE add-back must equal the federal FEIE exclusion amount")
	}

	// W1002: CA effective tax rate > 13.3%
	caAGI := forms.GetNum(results, forms.CA540Line17)
	caTax := forms.GetNum(results, forms.CA540Line40)
	if caAGI > 0 && caTax/caAGI > 0.133 {
		report.addResult("W1002", SeverityWarning, forms.CA540Line40,
			"California effective tax rate exceeds 13.3%; verify mental health surcharge is correct")
	}

	report.computeValidity()
	return report
}

// FullValidation runs validation and reasonableness checks, returning a unified report.
func FullValidation(results map[string]float64, strInputs map[string]string, taxYear int, includeCA bool) ValidationReport {
	validation := ValidateFull(results, strInputs, taxYear, includeCA)
	reasonableness := ReasonablenessCheck(results, strInputs, taxYear, includeCA)
	merged := ValidationReport{
		Results: append(validation.Results, reasonableness.Results...),
	}
	merged.computeValidity()
	return merged
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
