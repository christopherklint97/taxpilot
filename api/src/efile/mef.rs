//! MeF (Modernized e-File) XML generation for IRS federal returns.
//!
//! **Deterministic**: identical inputs always produce identical XML output.
//! Uses string-based XML generation (no random ordering, no runtime dependencies).

use std::collections::{BTreeSet, HashMap};
use std::fmt::Write as FmtWrite;

use crate::domain::form::*;

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const MEF_NAMESPACE: &str = "http://www.irs.gov/efile";
const RETURN_VERSION: &str = "2025v1.0";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

fn round_to_int(f: f64) -> i64 {
    f.round() as i64
}

fn format_ssn(s: &str) -> String {
    s.replace('-', "")
}

fn get_num(values: &HashMap<String, f64>, key: &str) -> f64 {
    values.get(key).copied().unwrap_or(0.0)
}

fn get_str<'a>(str_values: &'a HashMap<String, String>, key: &str) -> &'a str {
    str_values.get(key).map(|s| s.as_str()).unwrap_or("")
}

fn filing_status_code(fs: &str) -> i32 {
    match fs {
        "single" => 1,
        "mfj" => 2,
        "mfs" => 3,
        "hoh" => 4,
        "qss" => 5,
        _ => 1,
    }
}

/// Check if any field with the given prefix has a non-zero value.
fn is_schedule_needed(values: &HashMap<String, f64>, prefix: &str) -> bool {
    values
        .iter()
        .any(|(k, v)| k.starts_with(prefix) && *v != 0.0)
}

/// Write an XML element with an integer value, omitting if zero and optional.
fn write_elem(xml: &mut String, indent: &str, tag: &str, value: i64, omit_zero: bool) {
    if omit_zero && value == 0 {
        return;
    }
    let _ = writeln!(xml, "{indent}<{tag}>{value}</{tag}>");
}

/// Write an XML element with a string value, omitting if empty and optional.
fn write_str_elem(xml: &mut String, indent: &str, tag: &str, value: &str, omit_empty: bool) {
    if omit_empty && value.is_empty() {
        return;
    }
    let _ = writeln!(xml, "{indent}<{tag}>{}</{tag}>", xml_escape(value));
}

fn xml_escape(s: &str) -> String {
    s.replace('&', "&amp;")
        .replace('<', "&lt;")
        .replace('>', "&gt;")
        .replace('"', "&quot;")
        .replace('\'', "&apos;")
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

/// Generate MeF-compatible XML from solver results.
pub fn generate_mef_xml(
    values: &HashMap<String, f64>,
    str_values: &HashMap<String, String>,
    tax_year: i32,
) -> Result<String, String> {
    let mut xml = String::with_capacity(8192);
    let _ = writeln!(xml, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>");

    // Count documents for the documentCnt attribute
    let doc_count = count_documents(values, str_values);

    let _ = writeln!(
        xml,
        "<Return xmlns=\"{}\" returnVersion=\"{}\">",
        MEF_NAMESPACE, RETURN_VERSION
    );

    // ReturnHeader
    write_return_header(&mut xml, values, str_values, tax_year);

    // ReturnData
    let _ = writeln!(xml, "  <ReturnData documentCnt=\"{}\">", doc_count);

    // IRS1040 (always included)
    write_irs1040(&mut xml, values);

    // Conditional schedules and forms
    let sched_a_prefix = format!("{}:", FORM_SCHEDULE_A);
    if is_schedule_needed(values, &sched_a_prefix) {
        write_schedule_a(&mut xml, values);
    }

    let sched_1_prefix = format!("{}:", FORM_SCHEDULE_1);
    if is_schedule_needed(values, &sched_1_prefix) {
        write_schedule_1(&mut xml, values);
    }

    let sched_2_prefix = format!("{}:", FORM_SCHEDULE_2);
    if is_schedule_needed(values, &sched_2_prefix) {
        write_schedule_2(&mut xml, values);
    }

    let sched_3_prefix = format!("{}:", FORM_SCHEDULE_3);
    if is_schedule_needed(values, &sched_3_prefix) {
        write_schedule_3(&mut xml, values);
    }

    let sched_b_prefix = format!("{}:", FORM_SCHEDULE_B);
    if is_schedule_needed(values, &sched_b_prefix) {
        write_schedule_b(&mut xml, values);
    }

    let sched_c_prefix = format!("{}:", FORM_SCHEDULE_C);
    if is_schedule_needed(values, &sched_c_prefix) {
        write_schedule_c(&mut xml, values, str_values);
    }

    let sched_d_prefix = format!("{}:", FORM_SCHEDULE_D);
    if is_schedule_needed(values, &sched_d_prefix) {
        write_schedule_d(&mut xml, values);
    }

    let sched_se_prefix = format!("{}:", FORM_SCHEDULE_SE);
    if is_schedule_needed(values, &sched_se_prefix) {
        write_schedule_se(&mut xml, values);
    }

    let f8889_prefix = format!("{}:", FORM_F8889);
    if is_schedule_needed(values, &f8889_prefix) {
        write_form_8889(&mut xml, values, str_values);
    }

    let f8949_prefix = format!("{}:", FORM_F8949);
    if is_schedule_needed(values, &f8949_prefix) {
        write_form_8949(&mut xml, values);
    }

    let f8995_prefix = format!("{}:", FORM_F8995);
    if is_schedule_needed(values, &f8995_prefix) {
        write_form_8995(&mut xml, values);
    }

    let f2555_prefix = format!("{}:", FORM_F2555);
    if is_schedule_needed(values, &f2555_prefix) {
        write_form_2555(&mut xml, values, str_values);
    }

    let f1116_prefix = format!("{}:", FORM_F1116);
    if is_schedule_needed(values, &f1116_prefix) {
        write_form_1116(&mut xml, values, str_values);
    }

    let f8938_prefix = format!("{}:", FORM_F8938);
    if is_schedule_needed(values, &f8938_prefix) {
        write_form_8938(&mut xml, values, str_values);
    }

    let f8833_prefix = format!("{}:", FORM_F8833);
    if is_schedule_needed(values, &f8833_prefix) {
        write_form_8833(&mut xml, values, str_values);
    }

    // W-2s
    write_w2s(&mut xml, values, str_values);

    let _ = writeln!(xml, "  </ReturnData>");
    let _ = writeln!(xml, "</Return>");

    Ok(xml)
}

// ---------------------------------------------------------------------------
// Document count
// ---------------------------------------------------------------------------

fn count_documents(values: &HashMap<String, f64>, str_values: &HashMap<String, String>) -> usize {
    let mut count = 1; // IRS1040 always included

    let prefixes = [
        FORM_SCHEDULE_A,
        FORM_SCHEDULE_1,
        FORM_SCHEDULE_2,
        FORM_SCHEDULE_3,
        FORM_SCHEDULE_B,
        FORM_SCHEDULE_C,
        FORM_SCHEDULE_D,
        FORM_SCHEDULE_SE,
        FORM_F8889,
        FORM_F8949,
        FORM_F8995,
        FORM_F2555,
        FORM_F1116,
        FORM_F8938,
        FORM_F8833,
    ];

    for prefix in &prefixes {
        let p = format!("{}:", prefix);
        if is_schedule_needed(values, &p) {
            count += 1;
        }
    }

    count += count_w2_instances(values, str_values);
    count
}

fn count_w2_instances(values: &HashMap<String, f64>, str_values: &HashMap<String, String>) -> usize {
    let mut instances = BTreeSet::new();
    let wages_suffix = format!(":{}", W2_WAGES);
    let name_suffix = format!(":{}", W2_EMPLOYER_NAME);
    let w2_prefix = format!("{}:", FORM_W2);

    for k in values.keys() {
        if k.starts_with(&w2_prefix) && k.ends_with(&wages_suffix) {
            if let Some(inst) = extract_w2_instance(k) {
                instances.insert(inst);
            }
        }
    }
    for k in str_values.keys() {
        if k.starts_with(&w2_prefix) && k.ends_with(&name_suffix) {
            if let Some(inst) = extract_w2_instance(k) {
                instances.insert(inst);
            }
        }
    }
    instances.len()
}

fn extract_w2_instance(key: &str) -> Option<String> {
    let parts: Vec<&str> = key.splitn(3, ':').collect();
    if parts.len() == 3 {
        Some(parts[1].to_string())
    } else {
        None
    }
}

// ---------------------------------------------------------------------------
// XML writers
// ---------------------------------------------------------------------------

fn write_return_header(
    xml: &mut String,
    _values: &HashMap<String, f64>,
    str_values: &HashMap<String, String>,
    tax_year: i32,
) {
    let fs = get_str(str_values, F1040_FILING_STATUS);
    let fs_cd = filing_status_code(fs);
    let ssn = format_ssn(get_str(str_values, F1040_SSN));

    let _ = writeln!(xml, "  <ReturnHeader binaryAttachmentCnt=\"0\">");
    write_elem(xml, "    ", "TaxYr", tax_year as i64, false);
    write_str_elem(
        xml,
        "    ",
        "TaxPeriodBeginDt",
        &format!("{}-01-01", tax_year),
        false,
    );
    write_str_elem(
        xml,
        "    ",
        "TaxPeriodEndDt",
        &format!("{}-12-31", tax_year),
        false,
    );
    let _ = writeln!(xml, "    <Filer>");
    write_str_elem(xml, "      ", "PrimarySSN", &ssn, false);
    let _ = writeln!(xml, "      <Name>");
    write_str_elem(
        xml,
        "        ",
        "FirstName",
        get_str(str_values, F1040_FIRST_NAME),
        false,
    );
    write_str_elem(
        xml,
        "        ",
        "LastName",
        get_str(str_values, F1040_LAST_NAME),
        false,
    );
    let _ = writeln!(xml, "      </Name>");
    write_elem(xml, "      ", "FilingStatusCd", fs_cd as i64, false);
    let _ = writeln!(xml, "    </Filer>");
    let _ = writeln!(xml, "  </ReturnHeader>");
}

fn write_irs1040(xml: &mut String, r: &HashMap<String, f64>) {
    let _ = writeln!(xml, "    <IRS1040>");
    write_elem(xml, "      ", "WagesSalariesTips", round_to_int(get_num(r, F1040_LINE_1A)), false);
    write_elem(xml, "      ", "TaxExemptInterestAmt", round_to_int(get_num(r, F1040_LINE_2A)), true);
    write_elem(xml, "      ", "TaxableInterestAmt", round_to_int(get_num(r, F1040_LINE_2B)), true);
    write_elem(xml, "      ", "QualifiedDividendsAmt", round_to_int(get_num(r, F1040_LINE_3A)), true);
    write_elem(xml, "      ", "OrdinaryDividendsAmt", round_to_int(get_num(r, F1040_LINE_3B)), true);
    write_elem(xml, "      ", "CapitalGainLossAmt", round_to_int(get_num(r, F1040_LINE_7)), true);
    write_elem(xml, "      ", "OtherIncomeAmt", round_to_int(get_num(r, F1040_LINE_8)), true);
    write_elem(xml, "      ", "TotalIncomeAmt", round_to_int(get_num(r, F1040_LINE_9)), false);
    write_elem(xml, "      ", "AdjustmentsToIncomeAmt", round_to_int(get_num(r, F1040_LINE_10)), true);
    write_elem(xml, "      ", "AdjustedGrossIncomeAmt", round_to_int(get_num(r, F1040_LINE_11)), false);
    write_elem(xml, "      ", "TotalDeductionsAmt", round_to_int(get_num(r, F1040_LINE_14)), false);
    write_elem(xml, "      ", "TaxableIncomeAmt", round_to_int(get_num(r, F1040_LINE_15)), false);
    write_elem(xml, "      ", "TaxAmt", round_to_int(get_num(r, F1040_LINE_16)), false);
    write_elem(xml, "      ", "Sch2PartIAmt", round_to_int(get_num(r, F1040_LINE_17)), true);
    write_elem(xml, "      ", "Sch3PartIAmt", round_to_int(get_num(r, F1040_LINE_20)), true);
    write_elem(xml, "      ", "TaxAfterCreditsAmt", round_to_int(get_num(r, F1040_LINE_22)), false);
    write_elem(xml, "      ", "OtherTaxesAmt", round_to_int(get_num(r, F1040_LINE_23)), true);
    write_elem(xml, "      ", "TotalTaxAmt", round_to_int(get_num(r, F1040_LINE_24)), false);
    write_elem(xml, "      ", "WithholdingTaxAmt", round_to_int(get_num(r, F1040_LINE_25D)), false);
    write_elem(xml, "      ", "EstimatedTaxPaymentsAmt", round_to_int(get_num(r, F1040_LINE_31)), true);
    write_elem(xml, "      ", "TotalPaymentsAmt", round_to_int(get_num(r, F1040_LINE_33)), false);
    write_elem(xml, "      ", "OverpaidAmt", round_to_int(get_num(r, F1040_LINE_34)), true);
    write_elem(xml, "      ", "OwedAmt", round_to_int(get_num(r, F1040_LINE_37)), true);
    let _ = writeln!(xml, "    </IRS1040>");
}

fn write_schedule_a(xml: &mut String, r: &HashMap<String, f64>) {
    let _ = writeln!(xml, "    <IRS1040ScheduleA>");
    write_elem(xml, "      ", "MedicalAndDentalExpAmt", round_to_int(get_num(r, SCHED_A_LINE_1)), true);
    write_elem(xml, "      ", "AGIAmt", round_to_int(get_num(r, SCHED_A_LINE_2)), true);
    write_elem(xml, "      ", "MedicalFloorAmt", round_to_int(get_num(r, SCHED_A_LINE_3)), true);
    write_elem(xml, "      ", "DeductibleMedicalAmt", round_to_int(get_num(r, SCHED_A_LINE_4)), true);
    write_elem(xml, "      ", "StateLocalIncomeTaxAmt", round_to_int(get_num(r, SCHED_A_LINE_5A)), true);
    write_elem(xml, "      ", "PropertyTaxAmt", round_to_int(get_num(r, SCHED_A_LINE_5B)), true);
    write_elem(xml, "      ", "RealEstateTaxAmt", round_to_int(get_num(r, SCHED_A_LINE_5C)), true);
    write_elem(xml, "      ", "TotalSALTAmt", round_to_int(get_num(r, SCHED_A_LINE_5D)), true);
    write_elem(xml, "      ", "SALTDeductionAmt", round_to_int(get_num(r, SCHED_A_LINE_5E)), true);
    write_elem(xml, "      ", "MortgageInterestAmt", round_to_int(get_num(r, SCHED_A_LINE_8A)), true);
    write_elem(xml, "      ", "TotalInterestDeductionAmt", round_to_int(get_num(r, SCHED_A_LINE_11)), true);
    write_elem(xml, "      ", "CashCharityAmt", round_to_int(get_num(r, SCHED_A_LINE_12)), true);
    write_elem(xml, "      ", "NonCashCharityAmt", round_to_int(get_num(r, SCHED_A_LINE_13)), true);
    write_elem(xml, "      ", "CharityCarryoverAmt", round_to_int(get_num(r, SCHED_A_LINE_14)), true);
    write_elem(xml, "      ", "TotalCharityAmt", round_to_int(get_num(r, SCHED_A_LINE_15)), true);
    write_elem(xml, "      ", "TotalItemizedDeductAmt", round_to_int(get_num(r, SCHED_A_LINE_17)), false);
    let _ = writeln!(xml, "    </IRS1040ScheduleA>");
}

fn write_schedule_1(xml: &mut String, r: &HashMap<String, f64>) {
    let _ = writeln!(xml, "    <IRS1040Schedule1>");
    write_elem(xml, "      ", "BusinessIncomeLossAmt", round_to_int(get_num(r, SCHED_1_LINE_3)), true);
    write_elem(xml, "      ", "CapitalGainLossAmt", round_to_int(get_num(r, SCHED_1_LINE_7)), true);
    write_elem(xml, "      ", "TotalAdditionalIncomeAmt", round_to_int(get_num(r, SCHED_1_LINE_10)), false);
    write_elem(xml, "      ", "HSADeductionAmt", round_to_int(get_num(r, SCHED_1_LINE_15)), true);
    write_elem(xml, "      ", "SETaxDeductionAmt", round_to_int(get_num(r, SCHED_1_LINE_16)), true);
    write_elem(xml, "      ", "EarlyWithdrawalPenaltyAmt", round_to_int(get_num(r, SCHED_1_LINE_24)), true);
    write_elem(xml, "      ", "TotalAdjustmentsAmt", round_to_int(get_num(r, SCHED_1_LINE_26)), false);
    let _ = writeln!(xml, "    </IRS1040Schedule1>");
}

fn write_schedule_2(xml: &mut String, r: &HashMap<String, f64>) {
    let _ = writeln!(xml, "    <IRS1040Schedule2>");
    write_elem(xml, "      ", "AMTAmt", round_to_int(get_num(r, SCHED_2_LINE_1)), true);
    write_elem(xml, "      ", "TotalPartIAmt", round_to_int(get_num(r, SCHED_2_LINE_3)), true);
    write_elem(xml, "      ", "SelfEmploymentTaxAmt", round_to_int(get_num(r, SCHED_2_LINE_6)), true);
    write_elem(xml, "      ", "AdditionalMedicareTaxAmt", round_to_int(get_num(r, SCHED_2_LINE_12)), true);
    write_elem(xml, "      ", "HSAPenaltyAmt", round_to_int(get_num(r, SCHED_2_LINE_17C)), true);
    write_elem(xml, "      ", "NIITAmt", round_to_int(get_num(r, SCHED_2_LINE_18)), true);
    write_elem(xml, "      ", "TotalOtherTaxesAmt", round_to_int(get_num(r, SCHED_2_LINE_21)), false);
    let _ = writeln!(xml, "    </IRS1040Schedule2>");
}

fn write_schedule_3(xml: &mut String, r: &HashMap<String, f64>) {
    let _ = writeln!(xml, "    <IRS1040Schedule3>");
    write_elem(xml, "      ", "TotalNonrefundableCreditsAmt", round_to_int(get_num(r, SCHED_3_LINE_8)), true);
    write_elem(xml, "      ", "EstimatedTaxPaymentsAmt", round_to_int(get_num(r, SCHED_3_LINE_10)), true);
    write_elem(xml, "      ", "TotalOtherPaymentsAmt", round_to_int(get_num(r, SCHED_3_LINE_15)), true);
    let _ = writeln!(xml, "    </IRS1040Schedule3>");
}

fn write_schedule_b(xml: &mut String, r: &HashMap<String, f64>) {
    let _ = writeln!(xml, "    <IRS1040ScheduleB>");
    write_elem(xml, "      ", "TotalInterestAmt", round_to_int(get_num(r, SCHED_B_LINE_4)), false);
    write_elem(xml, "      ", "TotalDividendsAmt", round_to_int(get_num(r, SCHED_B_LINE_6)), false);
    let _ = writeln!(xml, "    </IRS1040ScheduleB>");
}

fn write_schedule_c(xml: &mut String, r: &HashMap<String, f64>, s: &HashMap<String, String>) {
    let _ = writeln!(xml, "    <IRS1040ScheduleC>");
    write_str_elem(xml, "      ", "BusinessName", get_str(s, SCHED_C_BUSINESS_NAME), true);
    write_str_elem(xml, "      ", "BusinessCode", get_str(s, SCHED_C_BUSINESS_CODE), true);
    write_elem(xml, "      ", "GrossReceiptsAmt", round_to_int(get_num(r, SCHED_C_LINE_1)), false);
    write_elem(xml, "      ", "GrossProfitAmt", round_to_int(get_num(r, SCHED_C_LINE_5)), false);
    write_elem(xml, "      ", "TotalExpensesAmt", round_to_int(get_num(r, SCHED_C_LINE_28)), true);
    write_elem(xml, "      ", "NetProfitLossAmt", round_to_int(get_num(r, SCHED_C_LINE_31)), false);
    let _ = writeln!(xml, "    </IRS1040ScheduleC>");
}

fn write_schedule_d(xml: &mut String, r: &HashMap<String, f64>) {
    let _ = writeln!(xml, "    <IRS1040ScheduleD>");
    write_elem(xml, "      ", "STGainLossAmt", round_to_int(get_num(r, SCHED_D_LINE_1)), true);
    write_elem(xml, "      ", "NetSTGainLossAmt", round_to_int(get_num(r, SCHED_D_LINE_7)), true);
    write_elem(xml, "      ", "LTGainLossAmt", round_to_int(get_num(r, SCHED_D_LINE_8)), true);
    write_elem(xml, "      ", "CapGainDistributionsAmt", round_to_int(get_num(r, SCHED_D_LINE_13)), true);
    write_elem(xml, "      ", "NetLTGainLossAmt", round_to_int(get_num(r, SCHED_D_LINE_15)), true);
    write_elem(xml, "      ", "NetCapitalGainLossAmt", round_to_int(get_num(r, SCHED_D_LINE_16)), false);
    let _ = writeln!(xml, "    </IRS1040ScheduleD>");
}

fn write_schedule_se(xml: &mut String, r: &HashMap<String, f64>) {
    let _ = writeln!(xml, "    <IRS1040ScheduleSE>");
    write_elem(xml, "      ", "NetSEEarningsAmt", round_to_int(get_num(r, SCHED_SE_LINE_2)), false);
    write_elem(xml, "      ", "TaxableEarningsAmt", round_to_int(get_num(r, SCHED_SE_LINE_3)), false);
    write_elem(xml, "      ", "SSTaxAmt", round_to_int(get_num(r, SCHED_SE_LINE_4)), false);
    write_elem(xml, "      ", "MedicareTaxAmt", round_to_int(get_num(r, SCHED_SE_LINE_5)), false);
    write_elem(xml, "      ", "SelfEmploymentTaxAmt", round_to_int(get_num(r, SCHED_SE_LINE_6)), false);
    write_elem(xml, "      ", "DeductibleSETaxAmt", round_to_int(get_num(r, SCHED_SE_LINE_7)), false);
    let _ = writeln!(xml, "    </IRS1040ScheduleSE>");
}

fn write_form_8889(xml: &mut String, r: &HashMap<String, f64>, s: &HashMap<String, String>) {
    let _ = writeln!(xml, "    <IRS8889>");
    write_str_elem(xml, "      ", "CoverageType", get_str(s, F8889_LINE_1), true);
    write_elem(xml, "      ", "ContributionsAmt", round_to_int(get_num(r, F8889_LINE_2)), true);
    write_elem(xml, "      ", "EmployerContribAmt", round_to_int(get_num(r, F8889_LINE_3)), true);
    write_elem(xml, "      ", "ContributionLimitAmt", round_to_int(get_num(r, F8889_LINE_6)), false);
    write_elem(xml, "      ", "HSADeductionAmt", round_to_int(get_num(r, F8889_LINE_9)), false);
    write_elem(xml, "      ", "DistributionsAmt", round_to_int(get_num(r, F8889_LINE_14A)), true);
    write_elem(xml, "      ", "QualifiedExpensesAmt", round_to_int(get_num(r, F8889_LINE_14C)), true);
    write_elem(xml, "      ", "TaxableDistribAmt", round_to_int(get_num(r, F8889_LINE_15)), true);
    write_elem(xml, "      ", "PenaltyAmt", round_to_int(get_num(r, F8889_LINE_17B)), true);
    let _ = writeln!(xml, "    </IRS8889>");
}

fn write_form_8949(xml: &mut String, r: &HashMap<String, f64>) {
    let _ = writeln!(xml, "    <IRS8949>");
    write_elem(xml, "      ", "STProceedsAmt", round_to_int(get_num(r, F8949_ST_PROCEEDS)), true);
    write_elem(xml, "      ", "STBasisAmt", round_to_int(get_num(r, F8949_ST_BASIS)), true);
    write_elem(xml, "      ", "STWashSaleAmt", round_to_int(get_num(r, F8949_ST_WASH)), true);
    write_elem(xml, "      ", "STGainLossAmt", round_to_int(get_num(r, F8949_ST_GAIN_LOSS)), true);
    write_elem(xml, "      ", "LTProceedsAmt", round_to_int(get_num(r, F8949_LT_PROCEEDS)), true);
    write_elem(xml, "      ", "LTBasisAmt", round_to_int(get_num(r, F8949_LT_BASIS)), true);
    write_elem(xml, "      ", "LTWashSaleAmt", round_to_int(get_num(r, F8949_LT_WASH)), true);
    write_elem(xml, "      ", "LTGainLossAmt", round_to_int(get_num(r, F8949_LT_GAIN_LOSS)), true);
    let _ = writeln!(xml, "    </IRS8949>");
}

fn write_form_8995(xml: &mut String, r: &HashMap<String, f64>) {
    let _ = writeln!(xml, "    <IRS8995>");
    write_elem(xml, "      ", "TotalQBIAmt", round_to_int(get_num(r, F8995_LINE_3)), false);
    write_elem(xml, "      ", "QBIComponentAmt", round_to_int(get_num(r, F8995_LINE_4)), false);
    write_elem(xml, "      ", "TaxableIncBeforeQBIAmt", round_to_int(get_num(r, F8995_LINE_5)), false);
    write_elem(xml, "      ", "IncomeLimitationAmt", round_to_int(get_num(r, F8995_LINE_8)), false);
    write_elem(xml, "      ", "QBIDeductionAmt", round_to_int(get_num(r, F8995_LINE_10)), false);
    let _ = writeln!(xml, "    </IRS8995>");
}

fn write_form_2555(xml: &mut String, r: &HashMap<String, f64>, s: &HashMap<String, String>) {
    let _ = writeln!(xml, "    <IRS2555>");
    write_str_elem(xml, "      ", "ForeignCountry", get_str(s, F2555_FOREIGN_COUNTRY), false);
    write_str_elem(xml, "      ", "QualifyingTest", get_str(s, F2555_QUALIFYING_TEST), false);
    write_elem(xml, "      ", "QualifyingDays", round_to_int(get_num(r, F2555_QUALIFYING_DAYS)), false);
    write_elem(xml, "      ", "ForeignEarnedIncomeAmt", round_to_int(get_num(r, F2555_FOREIGN_EARNED_INCOME)), false);
    write_elem(xml, "      ", "ExclusionLimitAmt", round_to_int(get_num(r, F2555_EXCLUSION_LIMIT)), false);
    write_elem(xml, "      ", "ForeignIncomeExclAmt", round_to_int(get_num(r, F2555_FOREIGN_INCOME_EXCL)), false);
    write_elem(xml, "      ", "HousingExclusionAmt", round_to_int(get_num(r, F2555_HOUSING_EXCLUSION)), true);
    write_elem(xml, "      ", "HousingDeductionAmt", round_to_int(get_num(r, F2555_HOUSING_DEDUCTION)), true);
    write_elem(xml, "      ", "TotalExclusionAmt", round_to_int(get_num(r, F2555_TOTAL_EXCLUSION)), false);
    let _ = writeln!(xml, "    </IRS2555>");
}

fn write_form_1116(xml: &mut String, r: &HashMap<String, f64>, s: &HashMap<String, String>) {
    let _ = writeln!(xml, "    <IRS1116>");
    write_str_elem(xml, "      ", "ForeignCountry", get_str(s, F1116_FOREIGN_COUNTRY), false);
    write_str_elem(xml, "      ", "Category", get_str(s, F1116_CATEGORY), false);
    write_elem(xml, "      ", "ForeignSourceIncomeAmt", round_to_int(get_num(r, F1116_LINE_7)), false);
    write_elem(xml, "      ", "ForeignTaxPaidAmt", round_to_int(get_num(r, F1116_LINE_15)), false);
    write_elem(xml, "      ", "CreditLimitationAmt", round_to_int(get_num(r, F1116_LINE_21)), false);
    write_elem(xml, "      ", "AllowedCreditAmt", round_to_int(get_num(r, F1116_LINE_22)), false);
    write_elem(xml, "      ", "CarryforwardAmt", round_to_int(get_num(r, F1116_CARRYFORWARD)), true);
    let _ = writeln!(xml, "    </IRS1116>");
}

fn write_form_8938(xml: &mut String, r: &HashMap<String, f64>, s: &HashMap<String, String>) {
    let _ = writeln!(xml, "    <IRS8938>");
    write_str_elem(xml, "      ", "LivesAbroad", get_str(s, F8938_LIVES_ABROAD), false);
    write_elem(xml, "      ", "MaxValueAccountsAmt", round_to_int(get_num(r, F8938_MAX_VALUE_ACCOUNTS)), false);
    write_elem(xml, "      ", "YearEndValueAccountsAmt", round_to_int(get_num(r, F8938_YEAREND_ACCOUNTS)), false);
    write_elem(xml, "      ", "TotalMaxValueAmt", round_to_int(get_num(r, F8938_TOTAL_MAX_VALUE)), false);
    write_elem(xml, "      ", "TotalYearEndValueAmt", round_to_int(get_num(r, F8938_TOTAL_YEAREND_VALUE)), false);
    write_elem(xml, "      ", "FilingRequired", round_to_int(get_num(r, F8938_FILING_REQUIRED)), false);
    let _ = writeln!(xml, "    </IRS8938>");
}

fn write_form_8833(xml: &mut String, r: &HashMap<String, f64>, s: &HashMap<String, String>) {
    let _ = writeln!(xml, "    <IRS8833>");
    write_str_elem(xml, "      ", "TreatyCountry", get_str(s, F8833_TREATY_COUNTRY), false);
    write_str_elem(xml, "      ", "TreatyArticle", get_str(s, F8833_TREATY_ARTICLE), false);
    write_str_elem(xml, "      ", "IRCProvision", get_str(s, F8833_IRC_PROVISION), false);
    write_elem(xml, "      ", "TreatyAmountAmt", round_to_int(get_num(r, F8833_TREATY_AMOUNT)), true);
    write_elem(xml, "      ", "TreatyClaimed", round_to_int(get_num(r, F8833_TREATY_CLAIMED)), false);
    let _ = writeln!(xml, "    </IRS8833>");
}

fn write_w2s(xml: &mut String, r: &HashMap<String, f64>, s: &HashMap<String, String>) {
    // Discover W-2 instances deterministically using BTreeSet
    let mut instances = BTreeSet::new();
    let wages_suffix = format!(":{}", W2_WAGES);
    let name_suffix = format!(":{}", W2_EMPLOYER_NAME);
    let w2_prefix = format!("{}:", FORM_W2);

    for k in r.keys() {
        if k.starts_with(&w2_prefix) && k.ends_with(&wages_suffix) {
            if let Some(inst) = extract_w2_instance(k) {
                instances.insert(inst);
            }
        }
    }
    for k in s.keys() {
        if k.starts_with(&w2_prefix) && k.ends_with(&name_suffix) {
            if let Some(inst) = extract_w2_instance(k) {
                instances.insert(inst);
            }
        }
    }

    for inst in &instances {
        let prefix = format!("w2:{}:", inst);
        let _ = writeln!(xml, "    <IRSW2>");
        write_str_elem(
            xml,
            "      ",
            "EmployerName",
            get_str(s, &format!("{}employer_name", prefix)),
            false,
        );
        write_str_elem(
            xml,
            "      ",
            "EmployerEIN",
            &format_ssn(get_str(s, &format!("{}employer_ein", prefix))),
            false,
        );
        write_elem(
            xml,
            "      ",
            "WagesAmt",
            round_to_int(get_num(r, &format!("{}wages", prefix))),
            false,
        );
        write_elem(
            xml,
            "      ",
            "WithholdingAmt",
            round_to_int(get_num(r, &format!("{}federal_tax_withheld", prefix))),
            false,
        );
        write_elem(
            xml,
            "      ",
            "SSWagesAmt",
            round_to_int(get_num(r, &format!("{}ss_wages", prefix))),
            true,
        );
        write_elem(
            xml,
            "      ",
            "SSTaxAmt",
            round_to_int(get_num(r, &format!("{}ss_tax_withheld", prefix))),
            true,
        );
        write_elem(
            xml,
            "      ",
            "MedicareWagesAmt",
            round_to_int(get_num(r, &format!("{}medicare_wages", prefix))),
            true,
        );
        write_elem(
            xml,
            "      ",
            "MedicareTaxAmt",
            round_to_int(get_num(r, &format!("{}medicare_tax_withheld", prefix))),
            true,
        );
        let _ = writeln!(xml, "    </IRSW2>");
    }
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
    fn test_deterministic_output() {
        let values = make_values(&[
            (F1040_LINE_1A, 50000.0),
            (F1040_LINE_9, 50000.0),
            (F1040_LINE_11, 50000.0),
            (F1040_LINE_14, 14600.0),
            (F1040_LINE_15, 35400.0),
            (F1040_LINE_16, 4000.0),
            (F1040_LINE_22, 4000.0),
            (F1040_LINE_24, 4000.0),
            (F1040_LINE_25D, 5000.0),
            (F1040_LINE_33, 5000.0),
            (F1040_LINE_34, 1000.0),
            ("w2:1:wages", 50000.0),
            ("w2:1:federal_tax_withheld", 5000.0),
        ]);
        let str_values = make_str_values(&[
            (F1040_SSN, "123-45-6789"),
            (F1040_FILING_STATUS, "single"),
            (F1040_FIRST_NAME, "John"),
            (F1040_LAST_NAME, "Doe"),
            ("w2:1:employer_name", "ACME Corp"),
        ]);

        let xml1 = generate_mef_xml(&values, &str_values, 2025).unwrap();
        let xml2 = generate_mef_xml(&values, &str_values, 2025).unwrap();
        assert_eq!(xml1, xml2, "MeF XML must be deterministic");
    }

    #[test]
    fn test_basic_structure() {
        let values = make_values(&[
            (F1040_LINE_1A, 75000.0),
            (F1040_LINE_9, 75000.0),
            (F1040_LINE_11, 75000.0),
            (F1040_LINE_14, 14600.0),
            (F1040_LINE_15, 60400.0),
            (F1040_LINE_16, 8000.0),
            (F1040_LINE_22, 8000.0),
            (F1040_LINE_24, 8000.0),
            (F1040_LINE_25D, 10000.0),
            (F1040_LINE_33, 10000.0),
            (F1040_LINE_34, 2000.0),
        ]);
        let str_values = make_str_values(&[
            (F1040_SSN, "555-12-3456"),
            (F1040_FILING_STATUS, "mfj"),
            (F1040_FIRST_NAME, "Jane"),
            (F1040_LAST_NAME, "Smith"),
        ]);

        let xml = generate_mef_xml(&values, &str_values, 2025).unwrap();
        assert!(xml.contains("<?xml version=\"1.0\" encoding=\"UTF-8\"?>"));
        assert!(xml.contains("<Return xmlns=\"http://www.irs.gov/efile\""));
        assert!(xml.contains("<TaxYr>2025</TaxYr>"));
        assert!(xml.contains("<PrimarySSN>555123456</PrimarySSN>"));
        assert!(xml.contains("<FilingStatusCd>2</FilingStatusCd>"));
        assert!(xml.contains("<WagesSalariesTips>75000</WagesSalariesTips>"));
        assert!(xml.contains("<OverpaidAmt>2000</OverpaidAmt>"));
    }

    #[test]
    fn test_schedules_included_when_needed() {
        let values = make_values(&[
            (F1040_LINE_1A, 50000.0),
            (F1040_LINE_9, 50000.0),
            (F1040_LINE_11, 50000.0),
            (F1040_LINE_14, 14600.0),
            (F1040_LINE_15, 35400.0),
            (F1040_LINE_16, 4000.0),
            (F1040_LINE_22, 4000.0),
            (F1040_LINE_24, 4000.0),
            (F1040_LINE_25D, 5000.0),
            (F1040_LINE_33, 5000.0),
            (SCHED_A_LINE_17, 20000.0),
            (SCHED_A_LINE_12, 15000.0),
        ]);
        let str_values = make_str_values(&[
            (F1040_SSN, "123-45-6789"),
            (F1040_FILING_STATUS, "single"),
            (F1040_FIRST_NAME, "John"),
            (F1040_LAST_NAME, "Doe"),
        ]);

        let xml = generate_mef_xml(&values, &str_values, 2025).unwrap();
        assert!(xml.contains("<IRS1040ScheduleA>"));
        assert!(xml.contains("<TotalItemizedDeductAmt>20000</TotalItemizedDeductAmt>"));
    }

    #[test]
    fn test_form_2555_included() {
        let values = make_values(&[
            (F1040_LINE_1A, 0.0),
            (F1040_LINE_9, 100000.0),
            (F1040_LINE_11, 100000.0),
            (F1040_LINE_14, 14600.0),
            (F1040_LINE_15, 85400.0),
            (F1040_LINE_16, 15000.0),
            (F1040_LINE_22, 15000.0),
            (F1040_LINE_24, 15000.0),
            (F1040_LINE_25D, 0.0),
            (F1040_LINE_33, 0.0),
            (F2555_FOREIGN_EARNED_INCOME, 100000.0),
            (F2555_EXCLUSION_LIMIT, 130000.0),
            (F2555_FOREIGN_INCOME_EXCL, 100000.0),
            (F2555_TOTAL_EXCLUSION, 100000.0),
            (F2555_QUALIFYING_DAYS, 365.0),
        ]);
        let str_values = make_str_values(&[
            (F1040_SSN, "123-45-6789"),
            (F1040_FILING_STATUS, "single"),
            (F1040_FIRST_NAME, "John"),
            (F1040_LAST_NAME, "Doe"),
            (F2555_FOREIGN_COUNTRY, "Germany"),
            (F2555_QUALIFYING_TEST, "bona_fide_residence"),
        ]);

        let xml = generate_mef_xml(&values, &str_values, 2025).unwrap();
        assert!(xml.contains("<IRS2555>"));
        assert!(xml.contains("<ForeignCountry>Germany</ForeignCountry>"));
        assert!(xml.contains("<TotalExclusionAmt>100000</TotalExclusionAmt>"));
    }

    #[test]
    fn test_w2s_sorted_deterministically() {
        let values = make_values(&[
            (F1040_LINE_1A, 100000.0),
            (F1040_LINE_9, 100000.0),
            (F1040_LINE_11, 100000.0),
            (F1040_LINE_14, 14600.0),
            (F1040_LINE_15, 85400.0),
            (F1040_LINE_16, 15000.0),
            (F1040_LINE_22, 15000.0),
            (F1040_LINE_24, 15000.0),
            (F1040_LINE_25D, 15000.0),
            (F1040_LINE_33, 15000.0),
            ("w2:1:wages", 60000.0),
            ("w2:1:federal_tax_withheld", 8000.0),
            ("w2:2:wages", 40000.0),
            ("w2:2:federal_tax_withheld", 7000.0),
        ]);
        let str_values = make_str_values(&[
            (F1040_SSN, "123-45-6789"),
            (F1040_FILING_STATUS, "single"),
            (F1040_FIRST_NAME, "John"),
            (F1040_LAST_NAME, "Doe"),
            ("w2:1:employer_name", "First Corp"),
            ("w2:2:employer_name", "Second Corp"),
        ]);

        let xml = generate_mef_xml(&values, &str_values, 2025).unwrap();
        let pos1 = xml.find("First Corp").unwrap();
        let pos2 = xml.find("Second Corp").unwrap();
        assert!(pos1 < pos2, "W-2 instance 1 should appear before instance 2");
    }

    #[test]
    fn test_xml_escaping() {
        let values = make_values(&[
            (F1040_LINE_1A, 50000.0),
            (F1040_LINE_9, 50000.0),
            (F1040_LINE_11, 50000.0),
            (F1040_LINE_14, 14600.0),
            (F1040_LINE_15, 35400.0),
            (F1040_LINE_16, 4000.0),
            (F1040_LINE_22, 4000.0),
            (F1040_LINE_24, 4000.0),
            (F1040_LINE_25D, 5000.0),
            (F1040_LINE_33, 5000.0),
        ]);
        let str_values = make_str_values(&[
            (F1040_SSN, "123-45-6789"),
            (F1040_FILING_STATUS, "single"),
            (F1040_FIRST_NAME, "O'Brien"),
            (F1040_LAST_NAME, "Smith & Jones"),
        ]);

        let xml = generate_mef_xml(&values, &str_values, 2025).unwrap();
        assert!(xml.contains("O&apos;Brien"));
        assert!(xml.contains("Smith &amp; Jones"));
    }
}
