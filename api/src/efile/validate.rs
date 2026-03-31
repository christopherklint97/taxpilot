use std::collections::HashMap;

use crate::domain::form::*;

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

#[derive(Debug, Clone, Copy, PartialEq, Eq, serde::Serialize)]
#[serde(rename_all = "lowercase")]
pub enum Severity {
    Error,
    Warning,
    Info,
}

#[derive(Debug, Clone, serde::Serialize)]
pub struct ValidationResult {
    pub code: String,
    pub severity: Severity,
    pub message: String,
    pub field_key: Option<String>,
}

#[derive(Debug, Clone, serde::Serialize)]
pub struct ValidationReport {
    pub results: Vec<ValidationResult>,
    pub is_valid: bool,
}

impl ValidationReport {
    fn new() -> Self {
        Self {
            results: Vec::new(),
            is_valid: true,
        }
    }

    fn add(&mut self, code: &str, severity: Severity, field: &str, message: &str) {
        self.results.push(ValidationResult {
            code: code.to_string(),
            severity,
            message: message.to_string(),
            field_key: if field.is_empty() {
                None
            } else {
                Some(field.to_string())
            },
        });
    }

    pub fn compute_validity(&mut self) {
        self.is_valid = !self.results.iter().any(|r| r.severity == Severity::Error);
    }
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

fn get_num(values: &HashMap<String, f64>, key: &str) -> f64 {
    values.get(key).copied().unwrap_or(0.0)
}

fn num_exists(values: &HashMap<String, f64>, key: &str) -> bool {
    values.contains_key(key)
}

fn get_str<'a>(str_values: &'a HashMap<String, String>, key: &str) -> &'a str {
    str_values.get(key).map(|s| s.as_str()).unwrap_or("")
}

fn has_w2_wages(values: &HashMap<String, f64>) -> bool {
    for (k, v) in values {
        if k.starts_with(&format!("{}:", FORM_W2))
            && k.ends_with(&format!(":{}", W2_WAGES))
            && *v > 0.0
        {
            return true;
        }
    }
    false
}

// ---------------------------------------------------------------------------
// Filing status constants (matching Go side)
// ---------------------------------------------------------------------------

const FILING_SINGLE: &str = "single";
const FILING_MFJ: &str = "mfj";
const FILING_MFS: &str = "mfs";
const FILING_HOH: &str = "hoh";
const FILING_QSS: &str = "qss";

// Form 2555 qualifying test constants
const QUALIFYING_TEST_BFRT: &str = "bona_fide_residence";
const QUALIFYING_TEST_PPT: &str = "physical_presence";

// Yes option
const OPTION_YES: &str = "yes";

// ---------------------------------------------------------------------------
// Federal validation
// ---------------------------------------------------------------------------

fn validate_federal(
    report: &mut ValidationReport,
    values: &HashMap<String, f64>,
    str_values: &HashMap<String, String>,
    _tax_year: i32,
) {
    // R0001: SSN required
    if get_str(str_values, F1040_SSN).is_empty() {
        report.add("R0001", Severity::Error, F1040_SSN, "SSN is required for e-filing");
    }

    // R0002: Filing status required and valid
    let fs = get_str(str_values, F1040_FILING_STATUS);
    let valid_statuses = [FILING_SINGLE, FILING_MFJ, FILING_MFS, FILING_HOH, FILING_QSS];
    if fs.is_empty() || !valid_statuses.contains(&fs) {
        report.add(
            "R0002",
            Severity::Error,
            F1040_FILING_STATUS,
            "Filing status is required and must be one of: single, mfj, mfs, hoh, qss",
        );
    }

    // R0003: First name required
    if get_str(str_values, F1040_FIRST_NAME).is_empty() {
        report.add("R0003", Severity::Error, F1040_FIRST_NAME, "First name is required");
    }

    // R0004: Last name required
    if get_str(str_values, F1040_LAST_NAME).is_empty() {
        report.add("R0004", Severity::Error, F1040_LAST_NAME, "Last name is required");
    }

    // R0005: Total income must be non-negative
    if num_exists(values, F1040_LINE_9) && get_num(values, F1040_LINE_9) < 0.0 {
        report.add(
            "R0005",
            Severity::Error,
            F1040_LINE_9,
            "Total income must be non-negative",
        );
    }

    // R0006: Taxable income must not exceed total income
    if num_exists(values, F1040_LINE_15) && num_exists(values, F1040_LINE_9) {
        if get_num(values, F1040_LINE_15) > get_num(values, F1040_LINE_9) {
            report.add(
                "R0006",
                Severity::Error,
                F1040_LINE_15,
                "Taxable income cannot exceed total income",
            );
        }
    }

    // R0007: Total tax must be non-negative
    if num_exists(values, F1040_LINE_24) && get_num(values, F1040_LINE_24) < 0.0 {
        report.add(
            "R0007",
            Severity::Error,
            F1040_LINE_24,
            "Total tax must be non-negative",
        );
    }

    // R0008: Withholding must match sum (25d == 25a + 25b within $1)
    if num_exists(values, F1040_LINE_25D) {
        let sum = get_num(values, F1040_LINE_25A) + get_num(values, F1040_LINE_25B);
        if (get_num(values, F1040_LINE_25D) - sum).abs() > 1.0 {
            report.add(
                "R0008",
                Severity::Error,
                F1040_LINE_25D,
                "Total withholding (line 25d) must equal sum of lines 25a and 25b (within $1)",
            );
        }
    }

    // R0009: Refund and amount owed can't both be positive
    let refund = get_num(values, F1040_LINE_34);
    let owed = get_num(values, F1040_LINE_37);
    if refund > 0.0 && owed > 0.0 {
        report.add(
            "R0009",
            Severity::Error,
            F1040_LINE_34,
            "Refund and amount owed cannot both be positive",
        );
    }

    // R0010: Must have some income source
    let has_w2 = has_w2_wages(values);
    let has_schedule_c = get_num(values, SCHED_C_LINE_31) > 0.0;
    let has_foreign_income = get_num(values, F2555_FOREIGN_EARNED_INCOME) > 0.0;
    if !has_w2 && !has_schedule_c && !has_foreign_income {
        report.add(
            "R0010",
            Severity::Error,
            "w2:1:wages",
            "At least one W-2 with wages, Schedule C net profit, or foreign earned income is required",
        );
    }

    // R0011: Form 2555 requires qualifying test
    if has_foreign_income {
        let qt = get_str(str_values, F2555_QUALIFYING_TEST);
        if qt != QUALIFYING_TEST_BFRT && qt != QUALIFYING_TEST_PPT {
            report.add(
                "R0011",
                Severity::Error,
                F2555_QUALIFYING_TEST,
                "Form 2555 requires a qualifying test (bona_fide_residence or physical_presence)",
            );
        }
    }

    // R0012: Physical presence test requires >= 330 days
    if get_str(str_values, F2555_QUALIFYING_TEST) == QUALIFYING_TEST_PPT {
        let days = get_num(values, F2555_QUALIFYING_DAYS);
        if days < 330.0 {
            report.add(
                "R0012",
                Severity::Error,
                F2555_PPT_DAYS_PRESENT,
                "Physical presence test requires at least 330 days in a foreign country",
            );
        }
    }

    // R0013: FEIE cannot exceed the limit
    let exclusion = get_num(values, F2555_TOTAL_EXCLUSION);
    let limit = get_num(values, F2555_EXCLUSION_LIMIT);
    if exclusion > 0.0 && limit > 0.0 && exclusion > limit + 1.0 {
        report.add(
            "R0013",
            Severity::Error,
            F2555_TOTAL_EXCLUSION,
            "FEIE exclusion exceeds the annual limit",
        );
    }

    // R0014: Form 8938 required if threshold met but not filed
    if num_exists(values, F8938_FILING_REQUIRED) && get_num(values, F8938_FILING_REQUIRED) == 1.0 {
        let total_max = get_num(values, F8938_TOTAL_MAX_VALUE);
        if total_max == 0.0 {
            report.add(
                "R0014",
                Severity::Error,
                F8938_TOTAL_MAX_VALUE,
                "Form 8938 filing is required but foreign asset values appear incomplete",
            );
        }
    }

    // W0005: FBAR reminder if foreign accounts
    if get_str(str_values, SCHED_B_LINE_7A) == OPTION_YES {
        report.add(
            "W0005",
            Severity::Info,
            SCHED_B_LINE_7A,
            "You indicated foreign accounts \u{2014} remember to file FBAR (FinCEN 114) separately at bsaefiling.fincen.treas.gov if aggregate value exceeded $10,000",
        );
    }

    // W0001: Charitable donations > 60% of AGI
    let agi = get_num(values, F1040_LINE_11);
    let charitable = get_num(values, SCHED_A_LINE_12);
    if agi > 0.0 && charitable > 0.6 * agi {
        report.add(
            "W0001",
            Severity::Warning,
            SCHED_A_LINE_12,
            "Charitable donations exceed 60% of AGI, which may trigger audit scrutiny",
        );
    }

    // W0002: Medical expenses > 20% of AGI
    let medical = get_num(values, SCHED_A_LINE_1);
    if agi > 0.0 && medical > 0.2 * agi {
        report.add(
            "W0002",
            Severity::Warning,
            SCHED_A_LINE_1,
            "Medical expenses exceed 20% of AGI, which is unusually high",
        );
    }

    // W0003: Business expense ratio > 80% of revenue
    let biz_revenue = get_num(values, SCHED_C_LINE_7);
    let biz_expenses = get_num(values, SCHED_C_LINE_28);
    if biz_revenue > 0.0 && biz_expenses > 0.8 * biz_revenue {
        report.add(
            "W0003",
            Severity::Warning,
            SCHED_C_LINE_28,
            "Business expenses exceed 80% of revenue, which may trigger audit scrutiny",
        );
    }

    // W0004: SALT deduction at cap
    let salt = get_num(values, SCHED_A_LINE_5D);
    if salt > 0.0 {
        let cap = if fs == FILING_MFS { 5000.0 } else { 10000.0 };
        if salt >= cap {
            report.add(
                "W0004",
                Severity::Info,
                SCHED_A_LINE_5D,
                "SALT deduction is at the cap limit",
            );
        }
    }
}

// ---------------------------------------------------------------------------
// CA validation
// ---------------------------------------------------------------------------

fn validate_ca(
    report: &mut ValidationReport,
    values: &HashMap<String, f64>,
    _str_values: &HashMap<String, String>,
    _tax_year: i32,
) {
    // R1001: CA AGI required
    if !num_exists(values, CA540_LINE_17) {
        report.add("R1001", Severity::Error, CA540_LINE_17, "California AGI is required");
    }

    // R1002: CA total tax must be non-negative
    if num_exists(values, CA540_LINE_40) && get_num(values, CA540_LINE_40) < 0.0 {
        report.add(
            "R1002",
            Severity::Error,
            CA540_LINE_40,
            "California total tax must be non-negative",
        );
    }

    // R1003: CA refund and amount owed can't both be positive
    let ca_refund = get_num(values, CA540_LINE_75);
    let ca_owed = get_num(values, CA540_LINE_81);
    if ca_refund > 0.0 && ca_owed > 0.0 {
        report.add(
            "R1003",
            Severity::Error,
            CA540_LINE_75,
            "California refund and amount owed cannot both be positive",
        );
    }

    // W1001: CA AGI differs from federal AGI by > $100,000
    if num_exists(values, CA540_LINE_17) && num_exists(values, F1040_LINE_11) {
        let diff = (get_num(values, CA540_LINE_17) - get_num(values, F1040_LINE_11)).abs();
        if diff > 100_000.0 {
            report.add(
                "W1001",
                Severity::Warning,
                CA540_LINE_17,
                "California AGI differs from federal AGI by more than $100,000",
            );
        }
    }

    // R1004: CA Schedule CA must include FEIE add-back when FEIE is claimed
    let feie_exclusion = get_num(values, F2555_TOTAL_EXCLUSION);
    let feie_add_back = get_num(values, SCHED_CA_LINE_8D_COL_C);
    if feie_exclusion > 0.0 && (feie_add_back - feie_exclusion).abs() > 1.0 {
        report.add(
            "R1004",
            Severity::Error,
            SCHED_CA_LINE_8D_COL_C,
            "CA Schedule CA FEIE add-back must equal the federal FEIE exclusion amount",
        );
    }

    // W1002: CA effective tax rate > 13.3%
    let ca_agi = get_num(values, CA540_LINE_17);
    let ca_tax = get_num(values, CA540_LINE_40);
    if ca_agi > 0.0 && ca_tax / ca_agi > 0.133 {
        report.add(
            "W1002",
            Severity::Warning,
            CA540_LINE_40,
            "California effective tax rate exceeds 13.3%; verify mental health surcharge is correct",
        );
    }
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

/// Run all validation rules (federal + optional CA).
pub fn validate_return(
    values: &HashMap<String, f64>,
    str_values: &HashMap<String, String>,
    tax_year: i32,
    state_code: &str,
) -> ValidationReport {
    let mut report = ValidationReport::new();
    validate_federal(&mut report, values, str_values, tax_year);
    if state_code == "CA" {
        validate_ca(&mut report, values, str_values, tax_year);
    }
    report.compute_validity();
    report
}

/// Run both validation and reasonableness checks, returning a unified report.
pub fn full_validation(
    values: &HashMap<String, f64>,
    str_values: &HashMap<String, String>,
    tax_year: i32,
    state_code: &str,
) -> ValidationReport {
    let mut validation = validate_return(values, str_values, tax_year, state_code);
    let reasonableness =
        super::reasonableness::check_reasonableness(values, str_values, tax_year, state_code);
    validation.results.extend(reasonableness.results);
    validation.compute_validity();
    validation
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;

    fn make_values(pairs: &[(&str, f64)]) -> HashMap<String, f64> {
        pairs.iter().map(|(k, v)| (k.to_string(), *v)).collect()
    }

    fn make_str_values(pairs: &[(&str, &str)]) -> HashMap<String, String> {
        pairs
            .iter()
            .map(|(k, v)| (k.to_string(), v.to_string()))
            .collect()
    }

    #[test]
    fn test_valid_minimal_return() {
        let values = make_values(&[
            (F1040_LINE_9, 50000.0),
            (F1040_LINE_11, 50000.0),
            (F1040_LINE_15, 36000.0),
            (F1040_LINE_24, 4000.0),
            (F1040_LINE_25A, 5000.0),
            (F1040_LINE_25D, 5000.0),
            (F1040_LINE_33, 5000.0),
            (F1040_LINE_34, 1000.0),
            ("w2:1:wages", 50000.0),
        ]);
        let str_values = make_str_values(&[
            (F1040_SSN, "123-45-6789"),
            (F1040_FILING_STATUS, "single"),
            (F1040_FIRST_NAME, "John"),
            (F1040_LAST_NAME, "Doe"),
        ]);
        let report = validate_return(&values, &str_values, 2025, "");
        assert!(report.is_valid, "Expected valid return, got: {:?}", report.results);
    }

    #[test]
    fn test_missing_ssn() {
        let values = make_values(&[("w2:1:wages", 50000.0)]);
        let str_values = make_str_values(&[
            (F1040_FILING_STATUS, "single"),
            (F1040_FIRST_NAME, "John"),
            (F1040_LAST_NAME, "Doe"),
        ]);
        let report = validate_return(&values, &str_values, 2025, "");
        assert!(!report.is_valid);
        assert!(report.results.iter().any(|r| r.code == "R0001"));
    }

    #[test]
    fn test_invalid_filing_status() {
        let values = make_values(&[("w2:1:wages", 50000.0)]);
        let str_values = make_str_values(&[
            (F1040_SSN, "123-45-6789"),
            (F1040_FILING_STATUS, "invalid"),
            (F1040_FIRST_NAME, "John"),
            (F1040_LAST_NAME, "Doe"),
        ]);
        let report = validate_return(&values, &str_values, 2025, "");
        assert!(!report.is_valid);
        assert!(report.results.iter().any(|r| r.code == "R0002"));
    }

    #[test]
    fn test_refund_and_owed_both_positive() {
        let values = make_values(&[
            (F1040_LINE_34, 1000.0),
            (F1040_LINE_37, 500.0),
            ("w2:1:wages", 50000.0),
        ]);
        let str_values = make_str_values(&[
            (F1040_SSN, "123-45-6789"),
            (F1040_FILING_STATUS, "single"),
            (F1040_FIRST_NAME, "John"),
            (F1040_LAST_NAME, "Doe"),
        ]);
        let report = validate_return(&values, &str_values, 2025, "");
        assert!(!report.is_valid);
        assert!(report.results.iter().any(|r| r.code == "R0009"));
    }

    #[test]
    fn test_no_income_source() {
        let values = make_values(&[]);
        let str_values = make_str_values(&[
            (F1040_SSN, "123-45-6789"),
            (F1040_FILING_STATUS, "single"),
            (F1040_FIRST_NAME, "John"),
            (F1040_LAST_NAME, "Doe"),
        ]);
        let report = validate_return(&values, &str_values, 2025, "");
        assert!(!report.is_valid);
        assert!(report.results.iter().any(|r| r.code == "R0010"));
    }

    #[test]
    fn test_feie_qualifying_test_required() {
        let values = make_values(&[(F2555_FOREIGN_EARNED_INCOME, 100000.0)]);
        let str_values = make_str_values(&[
            (F1040_SSN, "123-45-6789"),
            (F1040_FILING_STATUS, "single"),
            (F1040_FIRST_NAME, "John"),
            (F1040_LAST_NAME, "Doe"),
        ]);
        let report = validate_return(&values, &str_values, 2025, "");
        assert!(report.results.iter().any(|r| r.code == "R0011"));
    }

    #[test]
    fn test_ppt_insufficient_days() {
        let values = make_values(&[
            (F2555_FOREIGN_EARNED_INCOME, 100000.0),
            (F2555_QUALIFYING_DAYS, 200.0),
        ]);
        let str_values = make_str_values(&[
            (F1040_SSN, "123-45-6789"),
            (F1040_FILING_STATUS, "single"),
            (F1040_FIRST_NAME, "John"),
            (F1040_LAST_NAME, "Doe"),
            (F2555_QUALIFYING_TEST, "physical_presence"),
        ]);
        let report = validate_return(&values, &str_values, 2025, "");
        assert!(report.results.iter().any(|r| r.code == "R0012"));
    }

    #[test]
    fn test_ca_refund_and_owed_both_positive() {
        let values = make_values(&[
            (CA540_LINE_17, 50000.0),
            (CA540_LINE_75, 1000.0),
            (CA540_LINE_81, 500.0),
            ("w2:1:wages", 50000.0),
        ]);
        let str_values = make_str_values(&[
            (F1040_SSN, "123-45-6789"),
            (F1040_FILING_STATUS, "single"),
            (F1040_FIRST_NAME, "John"),
            (F1040_LAST_NAME, "Doe"),
        ]);
        let report = validate_return(&values, &str_values, 2025, "CA");
        assert!(report.results.iter().any(|r| r.code == "R1003"));
    }

    #[test]
    fn test_ca_feie_add_back_mismatch() {
        let values = make_values(&[
            (CA540_LINE_17, 150000.0),
            (F1040_LINE_11, 100000.0),
            (F2555_TOTAL_EXCLUSION, 100000.0),
            (SCHED_CA_LINE_8D_COL_C, 50000.0), // should be 100000
            ("w2:1:wages", 50000.0),
        ]);
        let str_values = make_str_values(&[
            (F1040_SSN, "123-45-6789"),
            (F1040_FILING_STATUS, "single"),
            (F1040_FIRST_NAME, "John"),
            (F1040_LAST_NAME, "Doe"),
            (F2555_QUALIFYING_TEST, "bona_fide_residence"),
        ]);
        let report = validate_return(&values, &str_values, 2025, "CA");
        assert!(report.results.iter().any(|r| r.code == "R1004"));
    }

    #[test]
    fn test_fbar_reminder() {
        let values = make_values(&[("w2:1:wages", 50000.0)]);
        let str_values = make_str_values(&[
            (F1040_SSN, "123-45-6789"),
            (F1040_FILING_STATUS, "single"),
            (F1040_FIRST_NAME, "John"),
            (F1040_LAST_NAME, "Doe"),
            (SCHED_B_LINE_7A, "yes"),
        ]);
        let report = validate_return(&values, &str_values, 2025, "");
        assert!(report.results.iter().any(|r| r.code == "W0005"));
        // W0005 is Info, not Error, so return should still be valid
        let errors: Vec<_> = report
            .results
            .iter()
            .filter(|r| r.severity == Severity::Error)
            .collect();
        // The only errors should be from other rules, not W0005
        assert!(
            !errors.iter().any(|r| r.code == "W0005"),
            "W0005 should be Info, not Error"
        );
    }

    #[test]
    fn test_charitable_warning() {
        let values = make_values(&[
            (F1040_LINE_11, 50000.0),
            (SCHED_A_LINE_12, 40000.0), // 80% of AGI
            ("w2:1:wages", 50000.0),
        ]);
        let str_values = make_str_values(&[
            (F1040_SSN, "123-45-6789"),
            (F1040_FILING_STATUS, "single"),
            (F1040_FIRST_NAME, "John"),
            (F1040_LAST_NAME, "Doe"),
        ]);
        let report = validate_return(&values, &str_values, 2025, "");
        assert!(report.results.iter().any(|r| r.code == "W0001"));
    }

    #[test]
    fn test_withholding_mismatch() {
        let values = make_values(&[
            (F1040_LINE_25A, 3000.0),
            (F1040_LINE_25B, 2000.0),
            (F1040_LINE_25D, 10000.0), // should be 5000
            ("w2:1:wages", 50000.0),
        ]);
        let str_values = make_str_values(&[
            (F1040_SSN, "123-45-6789"),
            (F1040_FILING_STATUS, "single"),
            (F1040_FIRST_NAME, "John"),
            (F1040_LAST_NAME, "Doe"),
        ]);
        let report = validate_return(&values, &str_values, 2025, "");
        assert!(report.results.iter().any(|r| r.code == "R0008"));
    }
}
