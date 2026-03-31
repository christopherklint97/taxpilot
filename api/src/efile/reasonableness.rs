use std::collections::HashMap;

use crate::domain::form::*;
use crate::domain::taxmath::get_config_or_default;

use super::validate::{Severity, ValidationReport};

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

fn fk(form_id: &str, line: &str) -> String {
    format!("{form_id}:{line}")
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

/// Run all reasonableness checks (advisory warnings and informational notices).
pub fn check_reasonableness(
    values: &HashMap<String, f64>,
    str_values: &HashMap<String, String>,
    tax_year: i32,
    state_code: &str,
) -> ValidationReport {
    let mut report = ValidationReport {
        results: Vec::new(),
        is_valid: true,
    };

    let cfg = get_config_or_default(tax_year);
    let filing_status = get_str(str_values, F1040_FILING_STATUS);

    // --- Cross-check computed values ---

    // RC001: AGI should equal total income minus adjustments
    if num_exists(values, F1040_LINE_11) && num_exists(values, F1040_LINE_9) {
        let agi = get_num(values, F1040_LINE_11);
        let total_income = get_num(values, F1040_LINE_9);
        let adjustments = get_num(values, SCHED_1_LINE_26);
        let expected = total_income - adjustments;
        if (agi - expected).abs() > 1.0 {
            report.results.push(super::validate::ValidationResult {
                code: "RC001".to_string(),
                severity: Severity::Warning,
                field_key: Some(F1040_LINE_11.to_string()),
                message: format!(
                    "AGI (${:.0}) does not equal total income (${:.0}) minus adjustments (${:.0})",
                    agi, total_income, adjustments
                ),
            });
        }
    }

    // RC002: Taxable income should equal AGI minus deduction
    if num_exists(values, F1040_LINE_15) && num_exists(values, F1040_LINE_11) {
        let taxable_income = get_num(values, F1040_LINE_15);
        let agi = get_num(values, F1040_LINE_11);
        let mut deduction = get_num(values, F1040_LINE_12);
        if num_exists(values, SCHED_A_LINE_17) && get_num(values, SCHED_A_LINE_17) > 0.0 {
            deduction = get_num(values, SCHED_A_LINE_17);
        }
        let expected = (agi - deduction).max(0.0);
        if (taxable_income - expected).abs() > 1.0 {
            report.results.push(super::validate::ValidationResult {
                code: "RC002".to_string(),
                severity: Severity::Warning,
                field_key: Some(F1040_LINE_15.to_string()),
                message: format!(
                    "Taxable income (${:.0}) does not equal AGI (${:.0}) minus deduction (${:.0})",
                    taxable_income, agi, deduction
                ),
            });
        }
    }

    // RC003: Refund should equal payments minus tax when payments > tax
    if num_exists(values, F1040_LINE_34)
        && num_exists(values, F1040_LINE_33)
        && num_exists(values, F1040_LINE_24)
    {
        let refund = get_num(values, F1040_LINE_34);
        let payments = get_num(values, F1040_LINE_33);
        let tax = get_num(values, F1040_LINE_24);
        if payments > tax && refund > 0.0 {
            let expected = payments - tax;
            if (refund - expected).abs() > 1.0 {
                report.results.push(super::validate::ValidationResult {
                    code: "RC003".to_string(),
                    severity: Severity::Warning,
                    field_key: Some(F1040_LINE_34.to_string()),
                    message: format!(
                        "Refund (${:.0}) does not equal payments (${:.0}) minus tax (${:.0})",
                        refund, payments, tax
                    ),
                });
            }
        }
    }

    // RC004: Amount owed should equal tax minus payments when tax > payments
    if num_exists(values, F1040_LINE_37)
        && num_exists(values, F1040_LINE_33)
        && num_exists(values, F1040_LINE_24)
    {
        let owed = get_num(values, F1040_LINE_37);
        let payments = get_num(values, F1040_LINE_33);
        let tax = get_num(values, F1040_LINE_24);
        if tax > payments && owed > 0.0 {
            let expected = tax - payments;
            if (owed - expected).abs() > 1.0 {
                report.results.push(super::validate::ValidationResult {
                    code: "RC004".to_string(),
                    severity: Severity::Warning,
                    field_key: Some(F1040_LINE_37.to_string()),
                    message: format!(
                        "Amount owed (${:.0}) does not equal tax (${:.0}) minus payments (${:.0})",
                        owed, tax, payments
                    ),
                });
            }
        }
    }

    // --- Flag unusual values (audit triggers) ---

    // RC005: Home office deduction > 30% of business income
    let home_office_key = fk(FORM_SCHEDULE_C, "30");
    if num_exists(values, &home_office_key) && num_exists(values, SCHED_C_LINE_7) {
        let home_office = get_num(values, &home_office_key);
        let biz_income = get_num(values, SCHED_C_LINE_7);
        if biz_income > 0.0 && home_office > 0.3 * biz_income {
            report.results.push(super::validate::ValidationResult {
                code: "RC005".to_string(),
                severity: Severity::Warning,
                field_key: Some(home_office_key),
                message: "Home office deduction is large relative to business income".to_string(),
            });
        }
    }

    // RC006: Self-employment income > $400 but no SE tax
    if num_exists(values, SCHED_C_LINE_31) {
        let se_income = get_num(values, SCHED_C_LINE_31);
        let se_tax_key = fk(FORM_SCHEDULE_2, "4");
        let se_tax = get_num(values, &se_tax_key);
        if se_income > 400.0 && se_tax == 0.0 {
            report.results.push(super::validate::ValidationResult {
                code: "RC006".to_string(),
                severity: Severity::Warning,
                field_key: Some(se_tax_key),
                message: "Self-employment income present but no SE tax computed".to_string(),
            });
        }
    }

    // RC007: HSA contributions exceed IRS limits
    if num_exists(values, F8889_LINE_2) {
        let hsa_contrib = get_num(values, F8889_LINE_2);
        if hsa_contrib > 0.0 {
            let (limit, label) = if filing_status == "mfj" || filing_status == "mfs" {
                (cfg.hsa_limit_family, "family")
            } else {
                (cfg.hsa_limit_self_only, "single")
            };
            if hsa_contrib > limit {
                report.results.push(super::validate::ValidationResult {
                    code: "RC007".to_string(),
                    severity: Severity::Warning,
                    field_key: Some(F8889_LINE_2.to_string()),
                    message: format!(
                        "HSA contributions (${:.0}) exceed {} limit (${:.0}) for {}",
                        hsa_contrib, label, limit, tax_year
                    ),
                });
            }
        }
    }

    // RC008: Capital losses exceed limit
    let cap_loss_key = fk(FORM_SCHEDULE_D, "21");
    if num_exists(values, &cap_loss_key) {
        let cap_loss = get_num(values, &cap_loss_key);
        if cap_loss < 0.0 {
            let limit = if filing_status == "mfs" {
                cfg.capital_loss_limit_mfs
            } else {
                cfg.capital_loss_limit
            };
            if cap_loss < limit {
                report.results.push(super::validate::ValidationResult {
                    code: "RC008".to_string(),
                    severity: Severity::Warning,
                    field_key: Some(cap_loss_key),
                    message: format!(
                        "Capital losses (${:.0}) exceed the deductible limit (${:.0})",
                        cap_loss, limit
                    ),
                });
            }
        }
    }

    // RC009: Estimated tax payments exceed total tax
    if num_exists(values, SCHED_3_LINE_8) && num_exists(values, F1040_LINE_24) {
        let estimated = get_num(values, SCHED_3_LINE_8);
        let total_tax = get_num(values, F1040_LINE_24);
        if estimated > 0.0 && estimated > total_tax {
            report.results.push(super::validate::ValidationResult {
                code: "RC009".to_string(),
                severity: Severity::Info,
                field_key: Some(SCHED_3_LINE_8.to_string()),
                message: "Estimated payments exceed total tax; verify amounts".to_string(),
            });
        }
    }

    // RC010: Effective federal tax rate > 37%
    if num_exists(values, F1040_LINE_24) && num_exists(values, F1040_LINE_11) {
        let total_tax = get_num(values, F1040_LINE_24);
        let agi = get_num(values, F1040_LINE_11);
        if agi > 0.0 && total_tax / agi > cfg.max_effective_tax_rate {
            report.results.push(super::validate::ValidationResult {
                code: "RC010".to_string(),
                severity: Severity::Warning,
                field_key: Some(F1040_LINE_24.to_string()),
                message: "Effective tax rate exceeds highest bracket".to_string(),
            });
        }
    }

    // --- CA consistency checks ---
    if state_code == "CA" {
        // RC011: CA AGI differs significantly from federal AGI
        if num_exists(values, CA540_LINE_17) && num_exists(values, F1040_LINE_11) {
            let ca_agi = get_num(values, CA540_LINE_17);
            let fed_agi = get_num(values, F1040_LINE_11);
            let diff = (ca_agi - fed_agi).abs();
            if fed_agi > 0.0 && diff > 0.2 * fed_agi.abs() && diff > 5000.0 {
                report.results.push(super::validate::ValidationResult {
                    code: "RC011".to_string(),
                    severity: Severity::Info,
                    field_key: Some(CA540_LINE_17.to_string()),
                    message: format!(
                        "CA AGI (${:.0}) differs from federal AGI (${:.0}) by more than 20%; review Schedule CA adjustments",
                        ca_agi, fed_agi
                    ),
                });
            }
        }

        // RC012: CA tax should be <= 13.3% of taxable income + 1% mental health surcharge on amount > $1M
        if num_exists(values, CA540_LINE_31) && num_exists(values, CA540_LINE_19) {
            let ca_tax = get_num(values, CA540_LINE_31);
            let ca_taxable_income = get_num(values, CA540_LINE_19);
            if ca_taxable_income > 0.0 {
                let mut max_tax = cfg.ca_max_marginal_rate * ca_taxable_income;
                if ca_taxable_income > cfg.ca_mental_health_threshold {
                    max_tax +=
                        cfg.ca_mental_health_rate * (ca_taxable_income - cfg.ca_mental_health_threshold);
                }
                if ca_tax > max_tax {
                    report.results.push(super::validate::ValidationResult {
                        code: "RC012".to_string(),
                        severity: Severity::Warning,
                        field_key: Some(CA540_LINE_31.to_string()),
                        message: format!(
                            "CA tax (${:.0}) exceeds maximum expected rate for taxable income (${:.0})",
                            ca_tax, ca_taxable_income
                        ),
                    });
                }
            }
        }

        // RC013: HSA add-back on Schedule CA
        if num_exists(values, F8889_LINE_2) && get_num(values, F8889_LINE_2) > 0.0 {
            let hsa_add_back_key = fk(FORM_SCHEDULE_CA, "15_col_c");
            if !num_exists(values, &hsa_add_back_key) || get_num(values, &hsa_add_back_key) <= 0.0
            {
                report.results.push(super::validate::ValidationResult {
                    code: "RC013".to_string(),
                    severity: Severity::Info,
                    field_key: Some(hsa_add_back_key),
                    message: "HSA contributions taken federally but no CA Schedule CA add-back found; CA does not conform to federal HSA deduction".to_string(),
                });
            }
        }

        // RC014: QBI deduction should not carry to CA
        let qbi_key = fk(FORM_F8995, "15");
        if num_exists(values, &qbi_key) && get_num(values, &qbi_key) > 0.0 {
            if num_exists(values, CA540_LINE_18) {
                let ca_deduction = get_num(values, CA540_LINE_18);
                let fed_deduction = get_num(values, F1040_LINE_12);
                let qbi = get_num(values, &qbi_key);
                if fed_deduction > 0.0 && ca_deduction >= fed_deduction && qbi > 0.0 {
                    report.results.push(super::validate::ValidationResult {
                        code: "RC014".to_string(),
                        severity: Severity::Info,
                        field_key: Some(CA540_LINE_18.to_string()),
                        message: "QBI deduction taken federally; verify CA does not include QBI deduction (CA does not conform)".to_string(),
                    });
                }
            }
        }
    }

    // --- Expat reasonableness checks ---

    // RC015: FEIE claimed but low qualifying days
    if num_exists(values, F2555_TOTAL_EXCLUSION) && get_num(values, F2555_TOTAL_EXCLUSION) > 0.0 {
        let days = get_num(values, F2555_QUALIFYING_DAYS);
        if days > 0.0 && days < 330.0 {
            report.results.push(super::validate::ValidationResult {
                code: "RC015".to_string(),
                severity: Severity::Warning,
                field_key: Some(F2555_QUALIFYING_DAYS.to_string()),
                message: format!(
                    "FEIE claimed with only {} qualifying days; full exclusion requires 365 days (BFRT) or 330 days (PPT)",
                    days as i32
                ),
            });
        }
    }

    // RC016: FTC exceeds 50% of foreign income
    if num_exists(values, F1116_LINE_22) && num_exists(values, F1116_LINE_7) {
        let ftc = get_num(values, F1116_LINE_22);
        let foreign_income = get_num(values, F1116_LINE_7);
        if foreign_income > 0.0 && ftc > 0.5 * foreign_income {
            report.results.push(super::validate::ValidationResult {
                code: "RC016".to_string(),
                severity: Severity::Warning,
                field_key: Some(F1116_LINE_22.to_string()),
                message: "Foreign tax credit exceeds 50% of foreign source income; verify foreign taxes paid".to_string(),
            });
        }
    }

    // RC017: FEIE + FTC no double benefit check
    if num_exists(values, F2555_TOTAL_EXCLUSION) && num_exists(values, F1116_LINE_22) {
        let feie = get_num(values, F2555_TOTAL_EXCLUSION);
        let ftc = get_num(values, F1116_LINE_22);
        if feie > 0.0 && ftc > 0.0 {
            report.results.push(super::validate::ValidationResult {
                code: "RC017".to_string(),
                severity: Severity::Info,
                field_key: Some(F1116_LINE_22.to_string()),
                message: "Both FEIE and FTC claimed \u{2014} verify FTC applies only to income NOT excluded by FEIE".to_string(),
            });
        }
    }

    // RC018: FATCA value fluctuation (max value >> year-end value)
    if num_exists(values, F8938_TOTAL_MAX_VALUE) && num_exists(values, F8938_TOTAL_YEAREND_VALUE) {
        let max_val = get_num(values, F8938_TOTAL_MAX_VALUE);
        let year_end = get_num(values, F8938_TOTAL_YEAREND_VALUE);
        if year_end > 0.0 && max_val > 3.0 * year_end {
            report.results.push(super::validate::ValidationResult {
                code: "RC018".to_string(),
                severity: Severity::Info,
                field_key: Some(F8938_TOTAL_MAX_VALUE.to_string()),
                message: format!(
                    "Peak foreign asset value (${:.0}) is more than 3x year-end value (${:.0}); verify values",
                    max_val, year_end
                ),
            });
        }
    }

    // RC019: CA FEIE add-back equals federal FEIE (CA only)
    if state_code == "CA" {
        if num_exists(values, F2555_TOTAL_EXCLUSION)
            && get_num(values, F2555_TOTAL_EXCLUSION) > 0.0
        {
            let feie = get_num(values, F2555_TOTAL_EXCLUSION);
            let add_back = get_num(values, SCHED_CA_LINE_8D_COL_C);
            if (add_back - feie).abs() > 1.0 {
                report.results.push(super::validate::ValidationResult {
                    code: "RC019".to_string(),
                    severity: Severity::Warning,
                    field_key: Some(SCHED_CA_LINE_8D_COL_C.to_string()),
                    message: format!(
                        "CA FEIE add-back (${:.0}) does not match federal FEIE (${:.0})",
                        add_back, feie
                    ),
                });
            }
        }
    }

    report.compute_validity();
    report
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
    fn test_rc001_agi_mismatch() {
        let values = make_values(&[
            (F1040_LINE_9, 100000.0),
            (F1040_LINE_11, 80000.0), // AGI
            // No adjustments, so expected AGI = 100000
        ]);
        let str_values = make_str_values(&[]);
        let report = check_reasonableness(&values, &str_values, 2025, "");
        assert!(report.results.iter().any(|r| r.code == "RC001"));
    }

    #[test]
    fn test_rc002_taxable_income_mismatch() {
        let values = make_values(&[
            (F1040_LINE_11, 100000.0),
            (F1040_LINE_12, 14600.0),
            (F1040_LINE_15, 50000.0), // should be 85400
        ]);
        let str_values = make_str_values(&[]);
        let report = check_reasonableness(&values, &str_values, 2025, "");
        assert!(report.results.iter().any(|r| r.code == "RC002"));
    }

    #[test]
    fn test_rc003_refund_mismatch() {
        let values = make_values(&[
            (F1040_LINE_24, 5000.0),
            (F1040_LINE_33, 8000.0),
            (F1040_LINE_34, 1000.0), // should be 3000
        ]);
        let str_values = make_str_values(&[]);
        let report = check_reasonableness(&values, &str_values, 2025, "");
        assert!(report.results.iter().any(|r| r.code == "RC003"));
    }

    #[test]
    fn test_rc007_hsa_over_limit() {
        let values = make_values(&[(F8889_LINE_2, 10000.0)]);
        let str_values = make_str_values(&[(F1040_FILING_STATUS, "single")]);
        let report = check_reasonableness(&values, &str_values, 2025, "");
        assert!(report.results.iter().any(|r| r.code == "RC007"));
    }

    #[test]
    fn test_rc010_high_effective_rate() {
        let values = make_values(&[
            (F1040_LINE_24, 50000.0),
            (F1040_LINE_11, 100000.0), // 50% effective rate
        ]);
        let str_values = make_str_values(&[]);
        let report = check_reasonableness(&values, &str_values, 2025, "");
        assert!(report.results.iter().any(|r| r.code == "RC010"));
    }

    #[test]
    fn test_rc011_ca_agi_divergence() {
        let values = make_values(&[
            (F1040_LINE_11, 100000.0),
            (CA540_LINE_17, 150000.0), // 50% difference
        ]);
        let str_values = make_str_values(&[]);
        let report = check_reasonableness(&values, &str_values, 2025, "CA");
        assert!(report.results.iter().any(|r| r.code == "RC011"));
    }

    #[test]
    fn test_rc015_feie_low_days() {
        let values = make_values(&[
            (F2555_TOTAL_EXCLUSION, 100000.0),
            (F2555_QUALIFYING_DAYS, 200.0),
        ]);
        let str_values = make_str_values(&[]);
        let report = check_reasonableness(&values, &str_values, 2025, "");
        assert!(report.results.iter().any(|r| r.code == "RC015"));
    }

    #[test]
    fn test_rc016_ftc_high_ratio() {
        let values = make_values(&[
            (F1116_LINE_22, 30000.0),
            (F1116_LINE_7, 50000.0), // FTC = 60% of income
        ]);
        let str_values = make_str_values(&[]);
        let report = check_reasonableness(&values, &str_values, 2025, "");
        assert!(report.results.iter().any(|r| r.code == "RC016"));
    }

    #[test]
    fn test_rc017_feie_and_ftc() {
        let values = make_values(&[
            (F2555_TOTAL_EXCLUSION, 100000.0),
            (F1116_LINE_22, 5000.0),
        ]);
        let str_values = make_str_values(&[]);
        let report = check_reasonableness(&values, &str_values, 2025, "");
        assert!(report.results.iter().any(|r| r.code == "RC017"));
    }

    #[test]
    fn test_rc018_fatca_fluctuation() {
        let values = make_values(&[
            (F8938_TOTAL_MAX_VALUE, 900000.0),
            (F8938_TOTAL_YEAREND_VALUE, 200000.0), // 4.5x
        ]);
        let str_values = make_str_values(&[]);
        let report = check_reasonableness(&values, &str_values, 2025, "");
        assert!(report.results.iter().any(|r| r.code == "RC018"));
    }

    #[test]
    fn test_rc019_ca_feie_add_back_mismatch() {
        let values = make_values(&[
            (F2555_TOTAL_EXCLUSION, 100000.0),
            (SCHED_CA_LINE_8D_COL_C, 50000.0),
        ]);
        let str_values = make_str_values(&[]);
        let report = check_reasonableness(&values, &str_values, 2025, "CA");
        assert!(report.results.iter().any(|r| r.code == "RC019"));
    }

    #[test]
    fn test_clean_return_no_warnings() {
        let values = make_values(&[
            (F1040_LINE_9, 50000.0),
            (F1040_LINE_11, 50000.0),
            (F1040_LINE_12, 14600.0),
            (F1040_LINE_15, 35400.0),
            (F1040_LINE_24, 4000.0),
            (F1040_LINE_33, 5000.0),
            (F1040_LINE_34, 1000.0),
        ]);
        let str_values = make_str_values(&[(F1040_FILING_STATUS, "single")]);
        let report = check_reasonableness(&values, &str_values, 2025, "");
        assert!(
            report.results.is_empty(),
            "Expected no warnings, got: {:?}",
            report.results
        );
    }
}
