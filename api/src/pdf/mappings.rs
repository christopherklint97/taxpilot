/// Maps a field_key (e.g., "1040:1a") to the PDF AcroForm field name.
pub struct FieldMapping {
    pub field_key: &'static str,
    pub pdf_field: &'static str,
    pub format: FieldFormat,
}

pub enum FieldFormat {
    Currency,
    String,
    Integer,
    Ssn,
    Ein,
    Checkbox,
    /// Filing status radio button — index 0=Single, 1=MFJ, 2=MFS, 3=HOH, 4=QSS
    FilingStatus(usize),
}

/// Returns all PDF field mappings for a given form.
pub fn get_mappings(form_id: &str) -> Vec<FieldMapping> {
    match form_id {
        "1040" => f1040_mappings(),
        "schedule_1" => schedule_1_mappings(),
        "schedule_2" => schedule_2_mappings(),
        "schedule_3" => schedule_3_mappings(),
        "schedule_a" => schedule_a_mappings(),
        "schedule_b" => schedule_b_mappings(),
        "schedule_c" => schedule_c_mappings(),
        "schedule_d" => schedule_d_mappings(),
        "schedule_se" => schedule_se_mappings(),
        "form_8995" => form_8995_mappings(),
        "form_8889" => form_8889_mappings(),
        "form_8949" => form_8949_mappings(),
        "form_2555" => form_2555_mappings(),
        "form_1116" => form_1116_mappings(),
        "form_8938" => form_8938_mappings(),
        "form_8833" => form_8833_mappings(),
        "ca_540" => ca_540_mappings(),
        "ca_schedule_ca" => ca_schedule_ca_mappings(),
        "form_3514" => form_3514_mappings(),
        "form_3853" => form_3853_mappings(),
        _ => vec![],
    }
}

// ──────────────────────────────────────────────────────────────────────
// Form 1040 — U.S. Individual Income Tax Return (2024)
// Field IDs from visual field map of 2024 IRS PDF
// ──────────────────────────────────────────────────────────────────────
fn f1040_mappings() -> Vec<FieldMapping> {
    use FieldFormat::*;
    vec![
        // Filing status checkboxes (c1_3[0]=Single, c1_3[1]=MFJ, c1_3[2]=MFS)
        // Note: HOH and QSS use separate checkbox fields on the 2024 form
        m("1040:filing_status", "topmostSubform[0].Page1[0].FilingStatus_ReadOrder[0].c1_3[0]", FilingStatus(0)),
        m("1040:filing_status", "topmostSubform[0].Page1[0].FilingStatus_ReadOrder[0].c1_3[1]", FilingStatus(1)),
        m("1040:filing_status", "topmostSubform[0].Page1[0].FilingStatus_ReadOrder[0].c1_3[2]", FilingStatus(2)),
        // Personal info
        m("1040:first_name",  "topmostSubform[0].Page1[0].f1_04[0]", String),
        m("1040:last_name",   "topmostSubform[0].Page1[0].f1_05[0]", String),
        m("1040:ssn",         "topmostSubform[0].Page1[0].f1_06[0]", Ssn),
        // Page 1 income lines
        m("1040:1a", "topmostSubform[0].Page1[0].f1_32[0]", Currency),
        m("1040:1z", "topmostSubform[0].Page1[0].f1_41[0]", Currency),
        m("1040:2a", "topmostSubform[0].Page1[0].f1_42[0]", Currency),
        m("1040:2b", "topmostSubform[0].Page1[0].f1_43[0]", Currency),
        m("1040:3a", "topmostSubform[0].Page1[0].f1_44[0]", Currency),
        m("1040:3b", "topmostSubform[0].Page1[0].f1_45[0]", Currency),
        m("1040:7",  "topmostSubform[0].Page1[0].f1_52[0]", Currency),
        m("1040:8",  "topmostSubform[0].Page1[0].f1_53[0]", Currency),
        m("1040:9",  "topmostSubform[0].Page1[0].f1_54[0]", Currency),
        m("1040:10", "topmostSubform[0].Page1[0].f1_55[0]", Currency),
        m("1040:11", "topmostSubform[0].Page1[0].f1_56[0]", Currency),
        m("1040:12", "topmostSubform[0].Page1[0].f1_57[0]", Currency),
        m("1040:13", "topmostSubform[0].Page1[0].f1_58[0]", Currency),
        m("1040:14", "topmostSubform[0].Page1[0].f1_59[0]", Currency),
        m("1040:15", "topmostSubform[0].Page1[0].f1_60[0]", Currency),
        // Page 2
        m("1040:16", "topmostSubform[0].Page2[0].f2_02[0]", Currency),
        m("1040:17", "topmostSubform[0].Page2[0].f2_03[0]", Currency),
        m("1040:22", "topmostSubform[0].Page2[0].f2_08[0]", Currency),
        m("1040:23", "topmostSubform[0].Page2[0].f2_09[0]", Currency),
        m("1040:24", "topmostSubform[0].Page2[0].f2_10[0]", Currency),
        m("1040:25a","topmostSubform[0].Page2[0].f2_11[0]", Currency),
        m("1040:25b","topmostSubform[0].Page2[0].f2_12[0]", Currency),
        m("1040:25d","topmostSubform[0].Page2[0].f2_14[0]", Currency),
        m("1040:33", "topmostSubform[0].Page2[0].f2_22[0]", Currency),
        m("1040:34", "topmostSubform[0].Page2[0].f2_23[0]", Currency),
        m("1040:37", "topmostSubform[0].Page2[0].f2_28[0]", Currency),
    ]
}

// ──────────────────────────────────────────────────────────────────────
// Schedule 1 — Additional Income and Adjustments to Income
// ──────────────────────────────────────────────────────────────────────
fn schedule_1_mappings() -> Vec<FieldMapping> {
    use FieldFormat::*;
    vec![
        // Part I — Additional Income
        m("schedule_1:1",   "topmostSubform[0].Page1[0].f1_04[0]", Currency),
        m("schedule_1:2a",  "topmostSubform[0].Page1[0].f1_05[0]", Currency),
        m("schedule_1:3",   "topmostSubform[0].Page1[0].f1_07[0]", Currency),
        m("schedule_1:7",   "topmostSubform[0].Page1[0].f1_11[0]", Currency),
        m("schedule_1:8d",  "topmostSubform[0].Page1[0].f1_15[0]", Currency), // FEIE (negative)
        m("schedule_1:10",  "topmostSubform[0].Page1[0].f1_38[0]", Currency),
        // Part II — Adjustments to Income (page 2)
        m("schedule_1:11",  "topmostSubform[0].Page2[0].f2_01[0]", Currency),
        m("schedule_1:15",  "topmostSubform[0].Page2[0].f2_05[0]", Currency),
        m("schedule_1:16",  "topmostSubform[0].Page2[0].f2_06[0]", Currency),
        m("schedule_1:20",  "topmostSubform[0].Page2[0].f2_12[0]", Currency),
        m("schedule_1:21",  "topmostSubform[0].Page2[0].f2_13[0]", Currency),
        m("schedule_1:26",  "topmostSubform[0].Page2[0].f2_31[0]", Currency),
    ]
}

// ──────────────────────────────────────────────────────────────────────
// Schedule 2 — Additional Taxes
// ──────────────────────────────────────────────────────────────────────
fn schedule_2_mappings() -> Vec<FieldMapping> {
    use FieldFormat::*;
    vec![
        // Part I
        m("schedule_2:1",  "topmostSubform[0].Page1[0].f1_12[0]", Currency),
        m("schedule_2:3",  "topmostSubform[0].Page1[0].f1_13[0]", Currency),
        // Part II
        m("schedule_2:6",  "topmostSubform[0].Page1[0].f1_14[0]", Currency),
        m("schedule_2:12", "topmostSubform[0].Page1[0].f1_21[0]", Currency),
        m("schedule_2:18", "topmostSubform[0].Page1[0].f1_22[0]", Currency),
        // Page 2
        m("schedule_2:17c","topmostSubform[0].Page2[0].f2_04[0]", Currency),
        m("schedule_2:21", "topmostSubform[0].Page2[0].f2_25[0]", Currency),
    ]
}

// ──────────────────────────────────────────────────────────────────────
// Schedule 3 — Additional Credits and Payments
// ──────────────────────────────────────────────────────────────────────
fn schedule_3_mappings() -> Vec<FieldMapping> {
    use FieldFormat::*;
    vec![
        m("schedule_3:1",  "topmostSubform[0].Page1[0].f1_03[0]", Currency),
        m("schedule_3:8",  "topmostSubform[0].Page1[0].f1_26[0]", Currency),
        m("schedule_3:10", "topmostSubform[0].Page1[0].f1_28[0]", Currency),
        m("schedule_3:15", "topmostSubform[0].Page1[0].f1_39[0]", Currency),
    ]
}

// ──────────────────────────────────────────────────────────────────────
// Schedule A — Itemized Deductions
// ──────────────────────────────────────────────────────────────────────
fn schedule_a_mappings() -> Vec<FieldMapping> {
    use FieldFormat::*;
    vec![
        m("schedule_a:1",   "topmostSubform[0].Page1[0].f1_3[0]",  Currency),
        m("schedule_a:2",   "topmostSubform[0].Page1[0].f1_4[0]",  Currency),
        m("schedule_a:3",   "topmostSubform[0].Page1[0].f1_5[0]",  Currency),
        m("schedule_a:4",   "topmostSubform[0].Page1[0].f1_6[0]",  Currency),
        m("schedule_a:5a",  "topmostSubform[0].Page1[0].f1_7[0]",  Currency),
        m("schedule_a:5b",  "topmostSubform[0].Page1[0].f1_8[0]",  Currency),
        m("schedule_a:5c",  "topmostSubform[0].Page1[0].f1_9[0]",  Currency),
        m("schedule_a:5d",  "topmostSubform[0].Page1[0].f1_10[0]", Currency),
        m("schedule_a:5e",  "topmostSubform[0].Page1[0].f1_11[0]", Currency),
        m("schedule_a:8a",  "topmostSubform[0].Page1[0].f1_16[0]", Currency),
        m("schedule_a:10",  "topmostSubform[0].Page1[0].f1_23[0]", Currency),
        m("schedule_a:11",  "topmostSubform[0].Page1[0].f1_24[0]", Currency),
        m("schedule_a:12",  "topmostSubform[0].Page1[0].f1_25[0]", Currency),
        m("schedule_a:13",  "topmostSubform[0].Page1[0].f1_26[0]", Currency),
        m("schedule_a:14",  "topmostSubform[0].Page1[0].f1_27[0]", Currency),
        m("schedule_a:15",  "topmostSubform[0].Page1[0].f1_28[0]", Currency),
        m("schedule_a:16",  "topmostSubform[0].Page1[0].f1_33[0]", Currency),
        m("schedule_a:17",  "topmostSubform[0].Page1[0].f1_34[0]", Currency),
    ]
}

// ──────────────────────────────────────────────────────────────────────
// Schedule B — Interest and Ordinary Dividends
// ──────────────────────────────────────────────────────────────────────
fn schedule_b_mappings() -> Vec<FieldMapping> {
    use FieldFormat::*;
    vec![
        m("schedule_b:1",  "topmostSubform[0].Page1[0].f1_31[0]", Currency), // Line 2 total interest
        m("schedule_b:4",  "topmostSubform[0].Page1[0].f1_33[0]", Currency), // Line 4 subtract
        m("schedule_b:5",  "topmostSubform[0].Page1[0].f1_64[0]", Currency), // Line 6 total dividends
        m("schedule_b:6",  "topmostSubform[0].Page1[0].f1_64[0]", Currency),
    ]
}

// ──────────────────────────────────────────────────────────────────────
// Schedule C — Profit or Loss From Business
// ──────────────────────────────────────────────────────────────────────
fn schedule_c_mappings() -> Vec<FieldMapping> {
    use FieldFormat::*;
    vec![
        m("schedule_c:business_name", "topmostSubform[0].Page1[0].f1_5[0]", String),
        m("schedule_c:1",   "topmostSubform[0].Page1[0].f1_10[0]", Currency),
        m("schedule_c:4",   "topmostSubform[0].Page1[0].f1_13[0]", Currency),
        m("schedule_c:5",   "topmostSubform[0].Page1[0].f1_14[0]", Currency),
        m("schedule_c:7",   "topmostSubform[0].Page1[0].f1_16[0]", Currency),
        m("schedule_c:8",   "topmostSubform[0].Page1[0].f1_17[0]", Currency),
        m("schedule_c:10",  "topmostSubform[0].Page1[0].f1_18[0]", Currency),
        m("schedule_c:17",  "topmostSubform[0].Page1[0].f1_27[0]", Currency),
        m("schedule_c:18",  "topmostSubform[0].Page1[0].f1_28[0]", Currency),
        m("schedule_c:22",  "topmostSubform[0].Page1[0].f1_33[0]", Currency),
        m("schedule_c:25",  "topmostSubform[0].Page1[0].f1_37[0]", Currency),
        m("schedule_c:27",  "topmostSubform[0].Page1[0].f1_39[0]", Currency),
        m("schedule_c:28",  "topmostSubform[0].Page1[0].f1_41[0]", Currency),
        m("schedule_c:31",  "topmostSubform[0].Page1[0].f1_46[0]", Currency),
    ]
}

// ──────────────────────────────────────────────────────────────────────
// Schedule D — Capital Gains and Losses
// ──────────────────────────────────────────────────────────────────────
fn schedule_d_mappings() -> Vec<FieldMapping> {
    use FieldFormat::*;
    vec![
        // Part I short-term
        m("form_8949:st_gain_loss", "topmostSubform[0].Page1[0].f1_10[0]", Currency), // Line 1b col h
        m("schedule_d:7",           "topmostSubform[0].Page1[0].f1_22[0]", Currency),
        // Part II long-term
        m("form_8949:lt_gain_loss", "topmostSubform[0].Page1[0].f1_30[0]", Currency), // Line 8b col h
        m("schedule_d:13",          "topmostSubform[0].Page1[0].f1_41[0]", Currency),
        m("schedule_d:15",          "topmostSubform[0].Page1[0].f1_43[0]", Currency),
        // Part III summary (page 2)
        m("schedule_d:16",          "topmostSubform[0].Page2[0].f2_01[0]", Currency),
    ]
}

// ──────────────────────────────────────────────────────────────────────
// Schedule SE — Self-Employment Tax
// ──────────────────────────────────────────────────────────────────────
fn schedule_se_mappings() -> Vec<FieldMapping> {
    use FieldFormat::*;
    vec![
        m("schedule_se:2", "topmostSubform[0].Page1[0].f1_5[0]",  Currency),
        m("schedule_se:3", "topmostSubform[0].Page1[0].f1_6[0]",  Currency),
        m("schedule_se:4", "topmostSubform[0].Page1[0].f1_7[0]",  Currency), // 4a
        m("schedule_se:5", "topmostSubform[0].Page1[0].f1_19[0]", Currency), // line 10 (SS tax)
        m("schedule_se:6", "topmostSubform[0].Page1[0].f1_21[0]", Currency), // line 12
        m("schedule_se:7", "topmostSubform[0].Page1[0].f1_22[0]", Currency), // line 13
    ]
}

// ──────────────────────────────────────────────────────────────────────
// Form 8949 — Sales and Other Dispositions of Capital Assets
// ──────────────────────────────────────────────────────────────────────
fn form_8949_mappings() -> Vec<FieldMapping> {
    use FieldFormat::*;
    vec![
        m("form_8949:st_proceeds",  "topmostSubform[0].Page1[0].f1_3[0]",  Currency),
        m("form_8949:st_basis",     "topmostSubform[0].Page1[0].f1_4[0]",  Currency),
        m("form_8949:st_gain_loss", "topmostSubform[0].Page1[0].f1_6[0]",  Currency),
        m("form_8949:lt_proceeds",  "topmostSubform[0].Page2[0].f2_3[0]",  Currency),
        m("form_8949:lt_basis",     "topmostSubform[0].Page2[0].f2_4[0]",  Currency),
        m("form_8949:lt_gain_loss", "topmostSubform[0].Page2[0].f2_6[0]",  Currency),
    ]
}

// ──────────────────────────────────────────────────────────────────────
// Form 8995 — Qualified Business Income Deduction (Simplified)
// ──────────────────────────────────────────────────────────────────────
fn form_8995_mappings() -> Vec<FieldMapping> {
    use FieldFormat::*;
    vec![
        m("form_8995:1",  "topmostSubform[0].Page1[0].f1_18[0]", Currency), // Line 2
        m("form_8995:2",  "topmostSubform[0].Page1[0].f1_22[0]", Currency), // Line 6
        m("form_8995:3",  "topmostSubform[0].Page1[0].f1_20[0]", Currency), // Line 4
        m("form_8995:4",  "topmostSubform[0].Page1[0].f1_21[0]", Currency), // Line 5
        m("form_8995:5",  "topmostSubform[0].Page1[0].f1_27[0]", Currency), // Line 11
        m("form_8995:6",  "topmostSubform[0].Page1[0].f1_28[0]", Currency), // Line 12
        m("form_8995:7",  "topmostSubform[0].Page1[0].f1_29[0]", Currency), // Line 13
        m("form_8995:8",  "topmostSubform[0].Page1[0].f1_30[0]", Currency), // Line 14
        m("form_8995:10", "topmostSubform[0].Page1[0].f1_31[0]", Currency), // Line 15
    ]
}

// ──────────────────────────────────────────────────────────────────────
// Form 8889 — Health Savings Accounts
// ──────────────────────────────────────────────────────────────────────
fn form_8889_mappings() -> Vec<FieldMapping> {
    use FieldFormat::*;
    vec![
        m("form_8889:2",    "topmostSubform[0].Page1[0].f1_3[0]",  Currency),
        m("form_8889:3",    "topmostSubform[0].Page1[0].f1_4[0]",  Currency),
        m("form_8889:5",    "topmostSubform[0].Page1[0].f1_7[0]",  Currency), // catch-up
        m("form_8889:6",    "topmostSubform[0].Page1[0].f1_7[0]",  Currency),
        m("form_8889:9",    "topmostSubform[0].Page1[0].f1_14[0]", Currency), // deduction
        m("form_8889:14a",  "topmostSubform[0].Page1[0].f1_15[0]", Currency),
        m("form_8889:14c",  "topmostSubform[0].Page1[0].f1_17[0]", Currency),
        m("form_8889:15",   "topmostSubform[0].Page1[0].f1_19[0]", Currency), // taxable
        m("form_8889:17b",  "topmostSubform[0].Page1[0].f1_20[0]", Currency),
    ]
}

// ──────────────────────────────────────────────────────────────────────
// Form 2555 — Foreign Earned Income
// ──────────────────────────────────────────────────────────────────────
fn form_2555_mappings() -> Vec<FieldMapping> {
    use FieldFormat::*;
    vec![
        m("form_2555:foreign_country",         "topmostSubform[0].Page1[0].f1_07[0]", String),
        m("form_2555:employer_name_2555",       "topmostSubform[0].Page1[0].f1_14[0]", String),
        m("form_2555:foreign_earned_income",    "topmostSubform[0].Page3[0].f3_01[0]", Currency),
        m("form_2555:exclusion_limit",          "topmostSubform[0].Page3[0].f3_07[0]", Currency),
        m("form_2555:foreign_income_exclusion", "topmostSubform[0].Page4[0].f4_01[0]", Currency),
        m("form_2555:housing_exclusion",        "topmostSubform[0].Page4[0].f4_05[0]", Currency),
        m("form_2555:total_exclusion",          "topmostSubform[0].Page4[0].f4_07[0]", Currency),
    ]
}

// ──────────────────────────────────────────────────────────────────────
// Form 1116 — Foreign Tax Credit
// ──────────────────────────────────────────────────────────────────────
fn form_1116_mappings() -> Vec<FieldMapping> {
    use FieldFormat::*;
    vec![
        m("form_1116:foreign_country",        "topmostSubform[0].Page1[0].f1_03[0]", String),
        m("form_1116:foreign_source_income",   "topmostSubform[0].Page1[0].f1_07[0]", Currency),
        m("form_1116:7",                       "topmostSubform[0].Page1[0].f1_15[0]", Currency),
        m("form_1116:15",                      "topmostSubform[0].Page1[0].f1_23[0]", Currency),
        m("form_1116:20",                      "topmostSubform[0].Page2[0].f2_01[0]", Currency),
        m("form_1116:21",                      "topmostSubform[0].Page2[0].f2_06[0]", Currency),
        m("form_1116:22",                      "topmostSubform[0].Page2[0].f2_07[0]", Currency),
    ]
}

// ──────────────────────────────────────────────────────────────────────
// Form 8938 — Statement of Specified Foreign Financial Assets
// ──────────────────────────────────────────────────────────────────────
fn form_8938_mappings() -> Vec<FieldMapping> {
    use FieldFormat::*;
    vec![
        m("form_8938:max_value_accounts",     "topmostSubform[0].Page1[0].f1_05[0]", Currency),
        m("form_8938:yearend_value_accounts",  "topmostSubform[0].Page1[0].f1_06[0]", Currency),
        m("form_8938:account_country",         "topmostSubform[0].Page1[0].f1_10[0]", String),
        m("form_8938:account_institution",     "topmostSubform[0].Page1[0].f1_11[0]", String),
        m("form_8938:total_max_value",         "topmostSubform[0].Page3[0].f3_01[0]", Currency),
        m("form_8938:total_yearend_value",     "topmostSubform[0].Page3[0].f3_02[0]", Currency),
    ]
}

// ──────────────────────────────────────────────────────────────────────
// Form 8833 — Treaty-Based Return Position Disclosure
// ──────────────────────────────────────────────────────────────────────
fn form_8833_mappings() -> Vec<FieldMapping> {
    use FieldFormat::*;
    vec![
        m("form_8833:treaty_country",   "topmostSubform[0].Page1[0].f1_03[0]", String),
        m("form_8833:treaty_article",   "topmostSubform[0].Page1[0].f1_04[0]", String),
        m("form_8833:irc_provision",    "topmostSubform[0].Page1[0].f1_05[0]", String),
        m("form_8833:treaty_amount",    "topmostSubform[0].Page1[0].f1_09[0]", Currency),
    ]
}

// ──────────────────────────────────────────────────────────────────────
// CA Form 540 — California Resident Income Tax Return
// FTB field IDs: 540-XXXX (sequential, page-prefixed)
// ──────────────────────────────────────────────────────────────────────
fn ca_540_mappings() -> Vec<FieldMapping> {
    use FieldFormat::*;
    vec![
        m("ca_540:7",  "540-1037", Currency),
        m("ca_540:13", "540-1041", Currency),
        m("ca_540:14", "540-1042", Currency),
        m("ca_540:15", "540-1043", Currency),
        m("ca_540:17", "540-1044", Currency),
        m("ca_540:18", "540-1045", Currency),
        m("ca_540:19", "540-1046", Currency),
        m("ca_540:31", "540-2001", Currency),
        m("ca_540:32", "540-2002", Currency),
        m("ca_540:35", "540-2003", Currency),
        m("ca_540:36", "540-2004", Currency),
        m("ca_540:40", "540-2005", Currency),
        m("ca_540:71", "540-2015", Currency),
        m("ca_540:74", "540-2018", Currency),
        m("ca_540:91", "540-2022", Currency),
        m("ca_540:93", "540-2024", Currency),
    ]
}

// ──────────────────────────────────────────────────────────────────────
// CA Schedule CA (540) — California Adjustments
// FTB field IDs: 540ca_form - XXXX
// ──────────────────────────────────────────────────────────────────────
fn ca_schedule_ca_mappings() -> Vec<FieldMapping> {
    use FieldFormat::*;
    vec![
        // Section A columns: A=federal, B=subtractions, C=additions
        m("ca_schedule_ca:2_col_a",  "540ca_form - 1005", Currency),
        m("ca_schedule_ca:2_col_b",  "540ca_form - 1006", Currency),
        m("ca_schedule_ca:2_col_c",  "540ca_form - 1007", Currency),
        m("ca_schedule_ca:3_col_a",  "540ca_form - 1008", Currency),
        m("ca_schedule_ca:3_col_b",  "540ca_form - 1009", Currency),
        m("ca_schedule_ca:3_col_c",  "540ca_form - 1010", Currency),
        m("ca_schedule_ca:7_col_a",  "540ca_form - 1017", Currency),
        m("ca_schedule_ca:7_col_b",  "540ca_form - 1018", Currency),
        m("ca_schedule_ca:7_col_c",  "540ca_form - 1019", Currency),
        m("ca_schedule_ca:8d_col_c", "540ca_form - 1028", Currency), // FEIE add-back
        m("ca_schedule_ca:37_col_b", "540ca_form - 1065", Currency),
        m("ca_schedule_ca:37_col_c", "540ca_form - 1066", Currency),
    ]
}

// ──────────────────────────────────────────────────────────────────────
// CA Form 3514 — California Earned Income Tax Credit
// ──────────────────────────────────────────────────────────────────────
fn form_3514_mappings() -> Vec<FieldMapping> {
    use FieldFormat::*;
    vec![
        m("form_3514:1", "3514_Form_1005", Currency),
        m("form_3514:3", "3514_Form_1007", Integer),
        m("form_3514:5", "3514_Form_1011", Currency),
        m("form_3514:7", "3514_Form_1031", Currency),
    ]
}

// ──────────────────────────────────────────────────────────────────────
// CA Form 3853 — Health Coverage
// ──────────────────────────────────────────────────────────────────────
fn form_3853_mappings() -> Vec<FieldMapping> {
    use FieldFormat::*;
    vec![
        m("form_3853:2",  "3853 Form 1005", Integer),
        m("form_3853:4",  "3853 Form 1008", Currency),
        m("form_3853:5",  "3853 Form 1009", Currency),
        m("form_3853:6",  "3853 Form 1010", Currency),
        m("form_3853:7",  "3853 Form 1011", Currency),
    ]
}

// Helper to construct a FieldMapping concisely
fn m(field_key: &'static str, pdf_field: &'static str, format: FieldFormat) -> FieldMapping {
    FieldMapping {
        field_key,
        pdf_field,
        format,
    }
}
