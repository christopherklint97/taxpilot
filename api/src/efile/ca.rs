//! CA FTB XML generation for California state returns.
//!
//! **Deterministic**: identical inputs always produce identical XML output.

use std::collections::HashMap;
use std::fmt::Write as FmtWrite;

use crate::domain::form::*;

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

fn round_to_int(f: f64) -> i64 {
    f.round() as i64
}

fn get_num(values: &HashMap<String, f64>, key: &str) -> f64 {
    values.get(key).copied().unwrap_or(0.0)
}

fn get_str<'a>(str_values: &'a HashMap<String, String>, key: &str) -> &'a str {
    str_values.get(key).map(|s| s.as_str()).unwrap_or("")
}

fn filing_status_code(fs: &str) -> &'static str {
    match fs {
        "single" => "1",
        "mfj" => "2",
        "mfs" => "3",
        "hoh" => "4",
        "qss" => "5",
        _ => "1",
    }
}

fn write_elem(xml: &mut String, indent: &str, tag: &str, value: i64, omit_zero: bool) {
    if omit_zero && value == 0 {
        return;
    }
    let _ = writeln!(xml, "{indent}<{tag}>{value}</{tag}>");
}

fn write_str_elem(xml: &mut String, indent: &str, tag: &str, value: &str, omit_empty: bool) {
    if omit_empty && value.is_empty() {
        return;
    }
    let escaped = value
        .replace('&', "&amp;")
        .replace('<', "&lt;")
        .replace('>', "&gt;")
        .replace('"', "&quot;")
        .replace('\'', "&apos;");
    let _ = writeln!(xml, "{indent}<{tag}>{escaped}</{tag}>");
}

// ---------------------------------------------------------------------------
// Schedule CA adjustment struct for zero-check
// ---------------------------------------------------------------------------

struct SchedCAValues {
    interest_sub: i64,
    interest_add: i64,
    dividend_sub: i64,
    dividend_add: i64,
    cap_gain_sub: i64,
    cap_gain_add: i64,
    hsa_add_back: i64,
    feie_add_back: i64,
    housing_add_back: i64,
    salt_sub: i64,
    property_tax_add: i64,
    ca_itemized: i64,
    total_sub: i64,
    total_add: i64,
}

impl SchedCAValues {
    fn has_non_zero(&self) -> bool {
        self.interest_sub != 0
            || self.interest_add != 0
            || self.dividend_sub != 0
            || self.dividend_add != 0
            || self.cap_gain_sub != 0
            || self.cap_gain_add != 0
            || self.hsa_add_back != 0
            || self.feie_add_back != 0
            || self.housing_add_back != 0
            || self.salt_sub != 0
            || self.property_tax_add != 0
            || self.ca_itemized != 0
            || self.total_sub != 0
            || self.total_add != 0
    }
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

/// Generate CA FTB e-file XML from solver results.
pub fn generate_ca_xml(
    values: &HashMap<String, f64>,
    str_values: &HashMap<String, String>,
    tax_year: i32,
) -> Result<String, String> {
    let mut xml = String::with_capacity(4096);
    let _ = writeln!(xml, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>");
    let _ = writeln!(
        xml,
        "<CAReturn xmlns=\"http://www.ftb.ca.gov/efile\" version=\"{}.1\">",
        tax_year
    );

    // Header
    let _ = writeln!(xml, "  <CAReturnHeader>");
    write_elem(&mut xml, "    ", "TaxYear", tax_year as i64, false);
    write_str_elem(&mut xml, "    ", "PrimarySSN", get_str(str_values, F1040_SSN), false);
    write_str_elem(
        &mut xml,
        "    ",
        "FirstName",
        get_str(str_values, F1040_FIRST_NAME),
        false,
    );
    write_str_elem(
        &mut xml,
        "    ",
        "LastName",
        get_str(str_values, F1040_LAST_NAME),
        false,
    );
    write_str_elem(
        &mut xml,
        "    ",
        "FilingStatusCd",
        filing_status_code(get_str(str_values, F1040_FILING_STATUS)),
        false,
    );
    let _ = writeln!(xml, "  </CAReturnHeader>");

    // CA540
    let _ = writeln!(xml, "  <CA540>");
    write_elem(&mut xml, "    ", "FederalAGIAmt", round_to_int(get_num(values, CA540_LINE_13)), false);
    write_elem(&mut xml, "    ", "CASubtractionsAmt", round_to_int(get_num(values, CA540_LINE_14)), false);
    write_elem(&mut xml, "    ", "CAAdditionsAmt", round_to_int(get_num(values, CA540_LINE_15)), false);
    write_elem(&mut xml, "    ", "CAAGIAmt", round_to_int(get_num(values, CA540_LINE_17)), false);
    write_elem(&mut xml, "    ", "CADeductionAmt", round_to_int(get_num(values, CA540_LINE_18)), false);
    write_elem(&mut xml, "    ", "CATaxableIncomeAmt", round_to_int(get_num(values, CA540_LINE_19)), false);
    write_elem(&mut xml, "    ", "CATaxAmt", round_to_int(get_num(values, CA540_LINE_31)), false);
    write_elem(&mut xml, "    ", "ExemptionCreditAmt", round_to_int(get_num(values, CA540_LINE_32)), false);
    write_elem(&mut xml, "    ", "NetTaxAmt", round_to_int(get_num(values, CA540_LINE_35)), false);
    write_elem(&mut xml, "    ", "MentalHealthTaxAmt", round_to_int(get_num(values, CA540_LINE_36)), false);
    write_elem(&mut xml, "    ", "TotalTaxAmt", round_to_int(get_num(values, CA540_LINE_40)), false);
    write_elem(&mut xml, "    ", "WithholdingAmt", round_to_int(get_num(values, CA540_LINE_71)), false);
    write_elem(&mut xml, "    ", "TotalPaymentsAmt", round_to_int(get_num(values, CA540_LINE_74)), false);
    write_elem(&mut xml, "    ", "OverpaidAmt", round_to_int(get_num(values, CA540_LINE_91)), false);
    write_elem(&mut xml, "    ", "OwedAmt", round_to_int(get_num(values, CA540_LINE_93)), false);
    let _ = writeln!(xml, "  </CA540>");

    // Schedule CA (only if there are non-zero adjustments)
    let sca = SchedCAValues {
        interest_sub: round_to_int(get_num(values, "ca_schedule_ca:2_col_b")),
        interest_add: round_to_int(get_num(values, "ca_schedule_ca:2_col_c")),
        dividend_sub: round_to_int(get_num(values, "ca_schedule_ca:3_col_b")),
        dividend_add: round_to_int(get_num(values, "ca_schedule_ca:3_col_c")),
        cap_gain_sub: round_to_int(get_num(values, "ca_schedule_ca:7_col_b")),
        cap_gain_add: round_to_int(get_num(values, "ca_schedule_ca:7_col_c")),
        hsa_add_back: round_to_int(get_num(values, "ca_schedule_ca:15_col_c")),
        feie_add_back: round_to_int(get_num(values, SCHED_CA_LINE_8D_COL_C)),
        housing_add_back: round_to_int(get_num(values, SCHED_CA_LINE_8D_COL_C_HOUSING)),
        salt_sub: round_to_int(get_num(values, "ca_schedule_ca:5e_col_b")),
        property_tax_add: round_to_int(get_num(values, "ca_schedule_ca:5e_col_c")),
        ca_itemized: round_to_int(get_num(values, "ca_schedule_ca:ca_itemized")),
        total_sub: round_to_int(get_num(values, "ca_schedule_ca:37_col_b")),
        total_add: round_to_int(get_num(values, SCHED_CA_LINE_37_COL_C)),
    };

    if sca.has_non_zero() {
        let _ = writeln!(xml, "  <CAScheduleCA>");
        write_elem(&mut xml, "    ", "InterestSubAmt", sca.interest_sub, false);
        write_elem(&mut xml, "    ", "InterestAddAmt", sca.interest_add, false);
        write_elem(&mut xml, "    ", "DividendSubAmt", sca.dividend_sub, false);
        write_elem(&mut xml, "    ", "DividendAddAmt", sca.dividend_add, false);
        write_elem(&mut xml, "    ", "CapGainSubAmt", sca.cap_gain_sub, false);
        write_elem(&mut xml, "    ", "CapGainAddAmt", sca.cap_gain_add, false);
        write_elem(&mut xml, "    ", "HSAAddBackAmt", sca.hsa_add_back, false);
        write_elem(&mut xml, "    ", "FEIEAddBackAmt", sca.feie_add_back, true);
        write_elem(&mut xml, "    ", "HousingAddBackAmt", sca.housing_add_back, true);
        write_elem(&mut xml, "    ", "SALTSubAmt", sca.salt_sub, false);
        write_elem(&mut xml, "    ", "PropertyTaxAddAmt", sca.property_tax_add, false);
        write_elem(&mut xml, "    ", "CAItemizedAmt", sca.ca_itemized, false);
        write_elem(&mut xml, "    ", "TotalSubAmt", sca.total_sub, false);
        write_elem(&mut xml, "    ", "TotalAddAmt", sca.total_add, false);
        let _ = writeln!(xml, "  </CAScheduleCA>");
    }

    let _ = writeln!(xml, "</CAReturn>");

    Ok(xml)
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
            (CA540_LINE_13, 50000.0),
            (CA540_LINE_17, 50000.0),
            (CA540_LINE_19, 40000.0),
            (CA540_LINE_31, 2000.0),
            (CA540_LINE_40, 2000.0),
        ]);
        let str_values = make_str_values(&[
            (F1040_SSN, "123-45-6789"),
            (F1040_FILING_STATUS, "single"),
            (F1040_FIRST_NAME, "John"),
            (F1040_LAST_NAME, "Doe"),
        ]);

        let xml1 = generate_ca_xml(&values, &str_values, 2025).unwrap();
        let xml2 = generate_ca_xml(&values, &str_values, 2025).unwrap();
        assert_eq!(xml1, xml2, "CA XML must be deterministic");
    }

    #[test]
    fn test_basic_structure() {
        let values = make_values(&[
            (CA540_LINE_13, 75000.0),
            (CA540_LINE_17, 75000.0),
            (CA540_LINE_19, 60000.0),
            (CA540_LINE_31, 3000.0),
            (CA540_LINE_40, 3000.0),
        ]);
        let str_values = make_str_values(&[
            (F1040_SSN, "555-12-3456"),
            (F1040_FILING_STATUS, "mfj"),
            (F1040_FIRST_NAME, "Jane"),
            (F1040_LAST_NAME, "Smith"),
        ]);

        let xml = generate_ca_xml(&values, &str_values, 2025).unwrap();
        assert!(xml.contains("<?xml version=\"1.0\" encoding=\"UTF-8\"?>"));
        assert!(xml.contains("<CAReturn xmlns=\"http://www.ftb.ca.gov/efile\""));
        assert!(xml.contains("version=\"2025.1\""));
        assert!(xml.contains("<TaxYear>2025</TaxYear>"));
        assert!(xml.contains("<FilingStatusCd>2</FilingStatusCd>"));
        assert!(xml.contains("<CAAGIAmt>75000</CAAGIAmt>"));
    }

    #[test]
    fn test_schedule_ca_included_when_needed() {
        let values = make_values(&[
            (CA540_LINE_13, 50000.0),
            (CA540_LINE_17, 55000.0),
            ("ca_schedule_ca:15_col_c", 4000.0), // HSA add-back
        ]);
        let str_values = make_str_values(&[
            (F1040_SSN, "123-45-6789"),
            (F1040_FILING_STATUS, "single"),
            (F1040_FIRST_NAME, "John"),
            (F1040_LAST_NAME, "Doe"),
        ]);

        let xml = generate_ca_xml(&values, &str_values, 2025).unwrap();
        assert!(xml.contains("<CAScheduleCA>"));
        assert!(xml.contains("<HSAAddBackAmt>4000</HSAAddBackAmt>"));
    }

    #[test]
    fn test_schedule_ca_omitted_when_all_zero() {
        let values = make_values(&[
            (CA540_LINE_13, 50000.0),
            (CA540_LINE_17, 50000.0),
        ]);
        let str_values = make_str_values(&[
            (F1040_SSN, "123-45-6789"),
            (F1040_FILING_STATUS, "single"),
            (F1040_FIRST_NAME, "John"),
            (F1040_LAST_NAME, "Doe"),
        ]);

        let xml = generate_ca_xml(&values, &str_values, 2025).unwrap();
        assert!(!xml.contains("<CAScheduleCA>"));
    }

    #[test]
    fn test_feie_add_back() {
        let values = make_values(&[
            (CA540_LINE_13, 150000.0),
            (CA540_LINE_17, 150000.0),
            (SCHED_CA_LINE_8D_COL_C, 100000.0),
            (SCHED_CA_LINE_37_COL_C, 100000.0),
        ]);
        let str_values = make_str_values(&[
            (F1040_SSN, "123-45-6789"),
            (F1040_FILING_STATUS, "single"),
            (F1040_FIRST_NAME, "John"),
            (F1040_LAST_NAME, "Doe"),
        ]);

        let xml = generate_ca_xml(&values, &str_values, 2025).unwrap();
        assert!(xml.contains("<CAScheduleCA>"));
        assert!(xml.contains("<FEIEAddBackAmt>100000</FEIEAddBackAmt>"));
        assert!(xml.contains("<TotalAddAmt>100000</TotalAddAmt>"));
    }
}
