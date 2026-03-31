// ---------------------------------------------------------------------------
// FormID constants (as &str)
// ---------------------------------------------------------------------------

// --- Input Forms ---
pub const FORM_W2: &str = "w2";
pub const FORM_1099_INT: &str = "1099int";
pub const FORM_1099_DIV: &str = "1099div";
pub const FORM_1099_NEC: &str = "1099nec";
pub const FORM_1099_B: &str = "1099b";

// --- Federal Forms ---
pub const FORM_F1040: &str = "1040";
pub const FORM_SCHEDULE_A: &str = "schedule_a";
pub const FORM_SCHEDULE_B: &str = "schedule_b";
pub const FORM_SCHEDULE_C: &str = "schedule_c";
pub const FORM_SCHEDULE_D: &str = "schedule_d";
pub const FORM_SCHEDULE_1: &str = "schedule_1";
pub const FORM_SCHEDULE_2: &str = "schedule_2";
pub const FORM_SCHEDULE_3: &str = "schedule_3";
pub const FORM_SCHEDULE_SE: &str = "schedule_se";
pub const FORM_F8949: &str = "form_8949";
pub const FORM_F8995: &str = "form_8995";
pub const FORM_F8889: &str = "form_8889";
pub const FORM_F2555: &str = "form_2555";
pub const FORM_F1116: &str = "form_1116";
pub const FORM_F8938: &str = "form_8938";
pub const FORM_F8833: &str = "form_8833";

// --- California State Forms ---
pub const FORM_CA540: &str = "ca_540";
pub const FORM_CA540_NR: &str = "ca_540nr";
pub const FORM_SCHEDULE_CA: &str = "ca_schedule_ca";
pub const FORM_F3514: &str = "form_3514";
pub const FORM_F3853: &str = "form_3853";

/// Returns every known FormID.
pub fn all_form_ids() -> Vec<&'static str> {
    vec![
        // Input
        FORM_W2,
        FORM_1099_INT,
        FORM_1099_DIV,
        FORM_1099_NEC,
        FORM_1099_B,
        // Federal
        FORM_F1040,
        FORM_SCHEDULE_A,
        FORM_SCHEDULE_B,
        FORM_SCHEDULE_C,
        FORM_SCHEDULE_D,
        FORM_SCHEDULE_1,
        FORM_SCHEDULE_2,
        FORM_SCHEDULE_3,
        FORM_SCHEDULE_SE,
        FORM_F8949,
        FORM_F8995,
        FORM_F8889,
        FORM_F2555,
        FORM_F1116,
        FORM_F8938,
        FORM_F8833,
        // CA
        FORM_CA540,
        FORM_SCHEDULE_CA,
        FORM_F3514,
        FORM_F3853,
    ]
}

/// Returns FormIDs that are input forms (W-2, 1099s).
pub fn input_form_ids() -> Vec<&'static str> {
    vec![FORM_W2, FORM_1099_INT, FORM_1099_DIV, FORM_1099_NEC, FORM_1099_B]
}

/// Returns string prefixes used for instance re-keying.
pub fn input_form_prefixes() -> Vec<String> {
    input_form_ids()
        .iter()
        .map(|id| format!("{}:", id))
        .collect()
}

// ---------------------------------------------------------------------------
// Form 1040 Field Keys
// ---------------------------------------------------------------------------

// Identification
pub const F1040_FILING_STATUS: &str = "1040:filing_status";
pub const F1040_FIRST_NAME: &str = "1040:first_name";
pub const F1040_LAST_NAME: &str = "1040:last_name";
pub const F1040_SSN: &str = "1040:ssn";

// Foreign wages
pub const F1040_FOREIGN_WAGES: &str = "1040:foreign_wages";
pub const F1040_FOREIGN_EMPLOYER: &str = "1040:foreign_employer";

// Income
pub const F1040_LINE_1A: &str = "1040:1a";
pub const F1040_LINE_1Z: &str = "1040:1z";
pub const F1040_LINE_2A: &str = "1040:2a";
pub const F1040_LINE_2B: &str = "1040:2b";
pub const F1040_LINE_3A: &str = "1040:3a";
pub const F1040_LINE_3B: &str = "1040:3b";
pub const F1040_LINE_7: &str = "1040:7";
pub const F1040_LINE_8: &str = "1040:8";
pub const F1040_LINE_9: &str = "1040:9";

// AGI
pub const F1040_LINE_10: &str = "1040:10";
pub const F1040_LINE_11: &str = "1040:11";

// Deductions
pub const F1040_LINE_12: &str = "1040:12";
pub const F1040_LINE_13: &str = "1040:13";
pub const F1040_LINE_14: &str = "1040:14";
pub const F1040_LINE_15: &str = "1040:15";

// Tax
pub const F1040_LINE_16: &str = "1040:16";
pub const F1040_LINE_17: &str = "1040:17";
pub const F1040_LINE_20: &str = "1040:20";
pub const F1040_LINE_22: &str = "1040:22";
pub const F1040_LINE_23: &str = "1040:23";
pub const F1040_LINE_24: &str = "1040:24";

// Payments
pub const F1040_LINE_25A: &str = "1040:25a";
pub const F1040_LINE_25B: &str = "1040:25b";
pub const F1040_LINE_25D: &str = "1040:25d";
pub const F1040_LINE_31: &str = "1040:31";
pub const F1040_LINE_33: &str = "1040:33";

// Refund / Owed
pub const F1040_LINE_34: &str = "1040:34";
pub const F1040_LINE_37: &str = "1040:37";

// ---------------------------------------------------------------------------
// W-2 Field Keys (line names, used with instance prefix)
// ---------------------------------------------------------------------------

pub const W2_EMPLOYER_NAME: &str = "employer_name";
pub const W2_EMPLOYER_EIN: &str = "employer_ein";
pub const W2_WAGES: &str = "wages";
pub const W2_FED_TAX_WITHHELD: &str = "federal_tax_withheld";
pub const W2_SS_WAGES: &str = "ss_wages";
pub const W2_SS_TAX_WITHHELD: &str = "ss_tax_withheld";
pub const W2_MEDICARE_WAGES: &str = "medicare_wages";
pub const W2_MEDICARE_TAX_WH: &str = "medicare_tax_withheld";
pub const W2_STATE_WAGES: &str = "state_wages";
pub const W2_STATE_TAX_WITHHELD: &str = "state_tax_withheld";

// W-2 wildcard patterns
pub const W2_WILDCARD_WAGES: &str = "w2:*:wages";
pub const W2_WILDCARD_FED_TAX_WH: &str = "w2:*:federal_tax_withheld";
pub const W2_WILDCARD_SS_WAGES: &str = "w2:*:ss_wages";
pub const W2_WILDCARD_SS_TAX_WH: &str = "w2:*:ss_tax_withheld";
pub const W2_WILDCARD_MEDICARE_WAGES: &str = "w2:*:medicare_wages";
pub const W2_WILDCARD_MEDICARE_TAX_WH: &str = "w2:*:medicare_tax_withheld";
pub const W2_WILDCARD_STATE_WAGES: &str = "w2:*:state_wages";
pub const W2_WILDCARD_STATE_TAX_WH: &str = "w2:*:state_tax_withheld";

// ---------------------------------------------------------------------------
// Schedule A Field Keys
// ---------------------------------------------------------------------------

pub const SCHED_A_LINE_1: &str = "schedule_a:1";
pub const SCHED_A_LINE_2: &str = "schedule_a:2";
pub const SCHED_A_LINE_3: &str = "schedule_a:3";
pub const SCHED_A_LINE_4: &str = "schedule_a:4";
pub const SCHED_A_LINE_5A: &str = "schedule_a:5a";
pub const SCHED_A_LINE_5B: &str = "schedule_a:5b";
pub const SCHED_A_LINE_5C: &str = "schedule_a:5c";
pub const SCHED_A_LINE_5D: &str = "schedule_a:5d";
pub const SCHED_A_LINE_5E: &str = "schedule_a:5e";
pub const SCHED_A_LINE_8A: &str = "schedule_a:8a";
pub const SCHED_A_LINE_11: &str = "schedule_a:11";
pub const SCHED_A_LINE_12: &str = "schedule_a:12";
pub const SCHED_A_LINE_13: &str = "schedule_a:13";
pub const SCHED_A_LINE_14: &str = "schedule_a:14";
pub const SCHED_A_LINE_15: &str = "schedule_a:15";
pub const SCHED_A_LINE_17: &str = "schedule_a:17";

// ---------------------------------------------------------------------------
// Schedule B Field Keys
// ---------------------------------------------------------------------------

pub const SCHED_B_LINE_1: &str = "schedule_b:1";
pub const SCHED_B_LINE_4: &str = "schedule_b:4";
pub const SCHED_B_LINE_5: &str = "schedule_b:5";
pub const SCHED_B_LINE_6: &str = "schedule_b:6";
pub const SCHED_B_FOREIGN_INTEREST: &str = "schedule_b:foreign_interest";
pub const SCHED_B_FOREIGN_INTEREST_PAYER: &str = "schedule_b:foreign_interest_payer";
pub const SCHED_B_LINE_7A: &str = "schedule_b:7a";
pub const SCHED_B_LINE_7B: &str = "schedule_b:7b";
pub const SCHED_B_LINE_8: &str = "schedule_b:8";
pub const SCHED_B_FBAR_REQUIRED: &str = "schedule_b:fbar_required";

// ---------------------------------------------------------------------------
// Schedule C Field Keys
// ---------------------------------------------------------------------------

pub const SCHED_C_BUSINESS_NAME: &str = "schedule_c:business_name";
pub const SCHED_C_BUSINESS_CODE: &str = "schedule_c:business_code";
pub const SCHED_C_LINE_1: &str = "schedule_c:1";
pub const SCHED_C_LINE_5: &str = "schedule_c:5";
pub const SCHED_C_LINE_7: &str = "schedule_c:7";
pub const SCHED_C_LINE_8: &str = "schedule_c:8";
pub const SCHED_C_LINE_10: &str = "schedule_c:10";
pub const SCHED_C_LINE_17: &str = "schedule_c:17";
pub const SCHED_C_LINE_18: &str = "schedule_c:18";
pub const SCHED_C_LINE_22: &str = "schedule_c:22";
pub const SCHED_C_LINE_25: &str = "schedule_c:25";
pub const SCHED_C_LINE_27: &str = "schedule_c:27";
pub const SCHED_C_LINE_28: &str = "schedule_c:28";
pub const SCHED_C_LINE_31: &str = "schedule_c:31";

// ---------------------------------------------------------------------------
// Schedule D Field Keys
// ---------------------------------------------------------------------------

pub const SCHED_D_LINE_1: &str = "schedule_d:1";
pub const SCHED_D_LINE_7: &str = "schedule_d:7";
pub const SCHED_D_LINE_8: &str = "schedule_d:8";
pub const SCHED_D_LINE_13: &str = "schedule_d:13";
pub const SCHED_D_LINE_15: &str = "schedule_d:15";
pub const SCHED_D_LINE_16: &str = "schedule_d:16";

// ---------------------------------------------------------------------------
// Schedule 1 Field Keys
// ---------------------------------------------------------------------------

pub const SCHED_1_LINE_1: &str = "schedule_1:1";
pub const SCHED_1_LINE_3: &str = "schedule_1:3";
pub const SCHED_1_LINE_7: &str = "schedule_1:7";
pub const SCHED_1_LINE_8D: &str = "schedule_1:8d";
pub const SCHED_1_LINE_10: &str = "schedule_1:10";
pub const SCHED_1_LINE_15: &str = "schedule_1:15";
pub const SCHED_1_LINE_16: &str = "schedule_1:16";
pub const SCHED_1_LINE_24: &str = "schedule_1:24";
pub const SCHED_1_LINE_26: &str = "schedule_1:26";

// ---------------------------------------------------------------------------
// Schedule 2 Field Keys
// ---------------------------------------------------------------------------

pub const SCHED_2_LINE_1: &str = "schedule_2:1";
pub const SCHED_2_LINE_3: &str = "schedule_2:3";
pub const SCHED_2_LINE_6: &str = "schedule_2:6";
pub const SCHED_2_LINE_12: &str = "schedule_2:12";
pub const SCHED_2_LINE_17C: &str = "schedule_2:17c";
pub const SCHED_2_LINE_18: &str = "schedule_2:18";
pub const SCHED_2_LINE_21: &str = "schedule_2:21";

// ---------------------------------------------------------------------------
// Schedule 3 Field Keys
// ---------------------------------------------------------------------------

pub const SCHED_3_LINE_1: &str = "schedule_3:1";
pub const SCHED_3_LINE_8: &str = "schedule_3:8";
pub const SCHED_3_LINE_10: &str = "schedule_3:10";
pub const SCHED_3_LINE_15: &str = "schedule_3:15";

// ---------------------------------------------------------------------------
// Schedule SE Field Keys
// ---------------------------------------------------------------------------

pub const SCHED_SE_LINE_2: &str = "schedule_se:2";
pub const SCHED_SE_LINE_3: &str = "schedule_se:3";
pub const SCHED_SE_LINE_4: &str = "schedule_se:4";
pub const SCHED_SE_LINE_5: &str = "schedule_se:5";
pub const SCHED_SE_LINE_6: &str = "schedule_se:6";
pub const SCHED_SE_LINE_7: &str = "schedule_se:7";

// ---------------------------------------------------------------------------
// Form 8889 Field Keys
// ---------------------------------------------------------------------------

pub const F8889_LINE_1: &str = "form_8889:1";
pub const F8889_LINE_2: &str = "form_8889:2";
pub const F8889_LINE_3: &str = "form_8889:3";
pub const F8889_LINE_5: &str = "form_8889:5";
pub const F8889_LINE_6: &str = "form_8889:6";
pub const F8889_LINE_9: &str = "form_8889:9";
pub const F8889_LINE_14A: &str = "form_8889:14a";
pub const F8889_LINE_14C: &str = "form_8889:14c";
pub const F8889_LINE_15: &str = "form_8889:15";
pub const F8889_LINE_17B: &str = "form_8889:17b";

// ---------------------------------------------------------------------------
// Form 8949 Field Keys
// ---------------------------------------------------------------------------

pub const F8949_ST_PROCEEDS: &str = "form_8949:st_proceeds";
pub const F8949_ST_BASIS: &str = "form_8949:st_basis";
pub const F8949_ST_WASH: &str = "form_8949:st_wash";
pub const F8949_ST_GAIN_LOSS: &str = "form_8949:st_gain_loss";
pub const F8949_LT_PROCEEDS: &str = "form_8949:lt_proceeds";
pub const F8949_LT_BASIS: &str = "form_8949:lt_basis";
pub const F8949_LT_WASH: &str = "form_8949:lt_wash";
pub const F8949_LT_GAIN_LOSS: &str = "form_8949:lt_gain_loss";

// ---------------------------------------------------------------------------
// Form 8995 Field Keys
// ---------------------------------------------------------------------------

pub const F8995_LINE_3: &str = "form_8995:3";
pub const F8995_LINE_4: &str = "form_8995:4";
pub const F8995_LINE_5: &str = "form_8995:5";
pub const F8995_LINE_8: &str = "form_8995:8";
pub const F8995_LINE_10: &str = "form_8995:10";

// ---------------------------------------------------------------------------
// Form 2555 (FEIE) Field Keys
// ---------------------------------------------------------------------------

pub const F2555_FOREIGN_COUNTRY: &str = "form_2555:foreign_country";
pub const F2555_FOREIGN_ADDRESS: &str = "form_2555:foreign_address";
pub const F2555_EMPLOYER_NAME: &str = "form_2555:employer_name_2555";
pub const F2555_EMPLOYER_FOREIGN: &str = "form_2555:employer_foreign";
pub const F2555_SELF_EMPLOYED_ABROAD: &str = "form_2555:self_employed_abroad";
pub const F2555_QUALIFYING_TEST: &str = "form_2555:qualifying_test";
pub const F2555_BFRT_START_DATE: &str = "form_2555:bfrt_start_date";
pub const F2555_BFRT_END_DATE: &str = "form_2555:bfrt_end_date";
pub const F2555_BFRT_FULL_YEAR: &str = "form_2555:bfrt_full_year";
pub const F2555_PPT_DAYS_PRESENT: &str = "form_2555:ppt_days_present";
pub const F2555_PPT_PERIOD_START: &str = "form_2555:ppt_period_start";
pub const F2555_PPT_PERIOD_END: &str = "form_2555:ppt_period_end";
pub const F2555_FOREIGN_EARNED_INCOME: &str = "form_2555:foreign_earned_income";
pub const F2555_CURRENCY_CODE: &str = "form_2555:currency_code";
pub const F2555_EXCHANGE_RATE: &str = "form_2555:exchange_rate";
pub const F2555_FOREIGN_TAX_PAID: &str = "form_2555:foreign_tax_paid";
pub const F2555_EMPLOYER_HOUSING: &str = "form_2555:employer_provided_housing";
pub const F2555_HOUSING_EXPENSES: &str = "form_2555:housing_expenses";
pub const F2555_QUALIFYING_DAYS: &str = "form_2555:qualifying_days";
pub const F2555_EXCLUSION_LIMIT: &str = "form_2555:exclusion_limit";
pub const F2555_FOREIGN_INCOME_EXCL: &str = "form_2555:foreign_income_exclusion";
pub const F2555_HOUSING_EXCLUSION: &str = "form_2555:housing_exclusion";
pub const F2555_HOUSING_DEDUCTION: &str = "form_2555:housing_deduction";
pub const F2555_TOTAL_EXCLUSION: &str = "form_2555:total_exclusion";

// ---------------------------------------------------------------------------
// Form 1116 (FTC) Field Keys
// ---------------------------------------------------------------------------

pub const F1116_CATEGORY: &str = "form_1116:category";
pub const F1116_FOREIGN_COUNTRY: &str = "form_1116:foreign_country";
pub const F1116_FOREIGN_SOURCE_INCOME: &str = "form_1116:foreign_source_income";
pub const F1116_FOREIGN_SOURCE_DEDUCT: &str = "form_1116:foreign_source_deductions";
pub const F1116_FOREIGN_TAX_PAID_INCOME: &str = "form_1116:foreign_tax_paid_income";
pub const F1116_FOREIGN_TAX_PAID_OTHER: &str = "form_1116:foreign_tax_paid_other";
pub const F1116_ACCRUED_OR_PAID: &str = "form_1116:accrued_or_paid";
pub const F1116_LINE_7: &str = "form_1116:7";
pub const F1116_LINE_15: &str = "form_1116:15";
pub const F1116_LINE_21: &str = "form_1116:21";
pub const F1116_LINE_22: &str = "form_1116:22";
pub const F1116_CARRYFORWARD: &str = "form_1116:carryforward";

// ---------------------------------------------------------------------------
// Form 8938 (FATCA) Field Keys
// ---------------------------------------------------------------------------

pub const F8938_LIVES_ABROAD: &str = "form_8938:lives_abroad";
pub const F8938_NUM_ACCOUNTS: &str = "form_8938:num_accounts";
pub const F8938_MAX_VALUE_ACCOUNTS: &str = "form_8938:max_value_accounts";
pub const F8938_YEAREND_ACCOUNTS: &str = "form_8938:yearend_value_accounts";
pub const F8938_NUM_OTHER_ASSETS: &str = "form_8938:num_other_assets";
pub const F8938_MAX_VALUE_OTHER: &str = "form_8938:max_value_other";
pub const F8938_YEAREND_OTHER: &str = "form_8938:yearend_value_other";
pub const F8938_ACCOUNT_COUNTRY: &str = "form_8938:account_country";
pub const F8938_ACCOUNT_INSTITUTION: &str = "form_8938:account_institution";
pub const F8938_ACCOUNT_TYPE: &str = "form_8938:account_type";
pub const F8938_INCOME_FROM_ACCOUNTS: &str = "form_8938:income_from_accounts";
pub const F8938_GAIN_FROM_ACCOUNTS: &str = "form_8938:gain_from_accounts";
pub const F8938_TOTAL_MAX_VALUE: &str = "form_8938:total_max_value";
pub const F8938_TOTAL_YEAREND_VALUE: &str = "form_8938:total_yearend_value";
pub const F8938_FILING_REQUIRED: &str = "form_8938:filing_required";

// ---------------------------------------------------------------------------
// Form 8833 (Treaty) Field Keys
// ---------------------------------------------------------------------------

pub const F8833_TREATY_COUNTRY: &str = "form_8833:treaty_country";
pub const F8833_TREATY_ARTICLE: &str = "form_8833:treaty_article";
pub const F8833_IRC_PROVISION: &str = "form_8833:irc_provision";
pub const F8833_TREATY_EXPLAIN: &str = "form_8833:treaty_position_explanation";
pub const F8833_TREATY_AMOUNT: &str = "form_8833:treaty_amount";
pub const F8833_NUM_POSITIONS: &str = "form_8833:num_positions";
pub const F8833_TREATY_CLAIMED: &str = "form_8833:treaty_claimed";

// ---------------------------------------------------------------------------
// CA Form 540 Field Keys
// ---------------------------------------------------------------------------

pub const CA540_LINE_7: &str = "ca_540:7";
pub const CA540_LINE_13: &str = "ca_540:13";
pub const CA540_LINE_14: &str = "ca_540:14";
pub const CA540_LINE_15: &str = "ca_540:15";
pub const CA540_LINE_17: &str = "ca_540:17";
pub const CA540_LINE_18: &str = "ca_540:18";
pub const CA540_LINE_19: &str = "ca_540:19";
pub const CA540_LINE_31: &str = "ca_540:31";
pub const CA540_LINE_32: &str = "ca_540:32";
pub const CA540_LINE_35: &str = "ca_540:35";
pub const CA540_LINE_36: &str = "ca_540:36";
pub const CA540_LINE_40: &str = "ca_540:40";
pub const CA540_LINE_71: &str = "ca_540:71";
pub const CA540_LINE_74: &str = "ca_540:74";
pub const CA540_LINE_75: &str = "ca_540:75";
pub const CA540_LINE_81: &str = "ca_540:81";
pub const CA540_LINE_91: &str = "ca_540:91";
pub const CA540_LINE_93: &str = "ca_540:93";

// ---------------------------------------------------------------------------
// CA Schedule CA Field Keys
// ---------------------------------------------------------------------------

pub const SCHED_CA_LINE_8D_COL_C: &str = "ca_schedule_ca:8d_col_c";
pub const SCHED_CA_LINE_8D_COL_C_HOUSING: &str = "ca_schedule_ca:8d_col_c_housing";
pub const SCHED_CA_LINE_37_COL_C: &str = "ca_schedule_ca:37_col_c";

// ---------------------------------------------------------------------------
// 1099 Wildcard Patterns
// ---------------------------------------------------------------------------

// 1099-INT
pub const F1099_INT_WILDCARD_INTEREST: &str = "1099int:*:interest_income";
pub const F1099_INT_WILDCARD_PENALTY: &str = "1099int:*:early_withdrawal_penalty";
pub const F1099_INT_WILDCARD_FED_TAX_WH: &str = "1099int:*:federal_tax_withheld";
pub const F1099_INT_WILDCARD_TAX_EXEMPT: &str = "1099int:*:tax_exempt_interest";
pub const F1099_INT_WILDCARD_US_BOND: &str = "1099int:*:us_savings_bond_interest";
pub const F1099_INT_WILDCARD_PABI: &str = "1099int:*:private_activity_bond_interest";

// 1099-DIV
pub const F1099_DIV_WILDCARD_ORDINARY: &str = "1099div:*:ordinary_dividends";
pub const F1099_DIV_WILDCARD_QUALIFIED: &str = "1099div:*:qualified_dividends";
pub const F1099_DIV_WILDCARD_CAP_GAIN: &str = "1099div:*:total_capital_gain";
pub const F1099_DIV_WILDCARD_FED_TAX_WH: &str = "1099div:*:federal_tax_withheld";
pub const F1099_DIV_WILDCARD_EXEMPT_INT: &str = "1099div:*:exempt_interest_dividends";
pub const F1099_DIV_WILDCARD_PABI: &str = "1099div:*:private_activity_bond_dividends";
pub const F1099_DIV_WILDCARD_SEC_199A: &str = "1099div:*:section_199a_dividends";
pub const F1099_DIV_WILDCARD_SEC_1250: &str = "1099div:*:section_1250_gain";

// 1099-NEC
pub const F1099_NEC_WILDCARD_COMP: &str = "1099nec:*:nonemployee_compensation";
pub const F1099_NEC_WILDCARD_FED_TAX_WH: &str = "1099nec:*:federal_tax_withheld";

// 1099-B
pub const F1099_B_WILDCARD_PROCEEDS: &str = "1099b:*:proceeds";
pub const F1099_B_WILDCARD_BASIS: &str = "1099b:*:cost_basis";
pub const F1099_B_WILDCARD_WASH_SALE: &str = "1099b:*:wash_sale_loss";
pub const F1099_B_WILDCARD_FED_TAX_WH: &str = "1099b:*:federal_tax_withheld";
pub const F1099_B_WILDCARD_TERM: &str = "1099b:*:term";

// ---------------------------------------------------------------------------
// Form 3514 (CalEITC) Field Keys
// ---------------------------------------------------------------------------

pub const F3514_LINE_3: &str = "form_3514:3";
pub const F3514_LINE_6_YCTC: &str = "form_3514:6_yctc";

// ---------------------------------------------------------------------------
// Form 3853 (Health Coverage) Field Keys
// ---------------------------------------------------------------------------

pub const F3853_LINE_1: &str = "form_3853:1";
pub const F3853_LINE_2: &str = "form_3853:2";
pub const F3853_LINE_3: &str = "form_3853:3";
