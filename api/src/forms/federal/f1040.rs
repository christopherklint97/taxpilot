use crate::domain::builders::*;
use crate::domain::field::*;
use crate::domain::form::*;
use crate::domain::taxmath;

pub fn form_1040() -> FormDef {
    FormDef {
        id: FORM_F1040.to_string(),
        name: "U.S. Individual Income Tax Return".to_string(),
        jurisdiction: Jurisdiction::Federal,
        tax_years: vec![2024, 2025, 2026],
        question_group: "personal".to_string(),
        question_order: 1,
        fields: vec![
            // --- Identification ---
            enum_field(
                "filing_status",
                "Filing status",
                "What is your filing status?",
                vec!["single", "mfj", "mfs", "hoh", "qss"],
            ),
            string_input_field("first_name", "First name", "What is your first name?"),
            string_input_field("last_name", "Last name", "What is your last name?"),
            string_input_field(
                "ssn",
                "Social Security number",
                "What is your Social Security number (XXX-XX-XXXX)?",
            ),
            // --- Income ---
            input_field(
                "foreign_wages",
                "Foreign wages (not on W-2)",
                "Enter wages from foreign employers not reported on a US W-2, converted to USD:",
            ),
            string_input_field(
                "foreign_employer",
                "Foreign employer(s)",
                "Describe the foreign employer(s) (e.g., \"Volvo AB, Sweden\"):",
            ),
            // Line 1a: Wages from W-2s + foreign wages
            {
                let deps = vec![
                    W2_WILDCARD_WAGES.to_string(),
                    F1040_FOREIGN_WAGES.to_string(),
                ];
                FieldDef::new_computed("1a", "Wages, salaries, tips (W-2 + foreign)", deps, Box::new(|dv: &DepValues| {
                    dv.sum_all(W2_WILDCARD_WAGES) + dv.get(F1040_FOREIGN_WAGES)
                }))
            },
            // Line 1z
            ref_field("1z", "Total from W-2s and other wage sources", F1040_LINE_1A),
            // Line 2a: Tax-exempt interest
            wildcard_sum_field("2a", "Tax-exempt interest", F1099_INT_WILDCARD_TAX_EXEMPT),
            // Line 2b: Taxable interest
            ref_field("2b", "Taxable interest", SCHED_B_LINE_4),
            // Line 3a: Qualified dividends
            wildcard_sum_field("3a", "Qualified dividends", F1099_DIV_WILDCARD_QUALIFIED),
            // Line 3b: Ordinary dividends
            ref_field("3b", "Ordinary dividends", SCHED_B_LINE_6),
            // Line 7: Capital gain or (loss)
            ref_field("7", "Capital gain or (loss)", SCHED_1_LINE_7),
            // Line 8: Other income from Schedule 1
            {
                let deps = vec![SCHED_1_LINE_10.to_string()];
                FieldDef::new_computed("8", "Other income from Schedule 1", deps, Box::new(|dv: &DepValues| {
                    dv.get(SCHED_1_LINE_10) - dv.get(SCHED_1_LINE_7)
                }))
            },
            // Line 9: Total income
            sum_field("9", "Total income", vec![
                F1040_LINE_1Z, F1040_LINE_2B, F1040_LINE_3B, F1040_LINE_7, F1040_LINE_8,
            ]),
            // Line 10: Adjustments to income
            ref_field("10", "Adjustments to income", SCHED_1_LINE_26),
            // Line 11: Adjusted gross income (AGI)
            max_zero_field("11", "Adjusted gross income", F1040_LINE_9, F1040_LINE_10),
            // Line 12: Deduction - larger of standard or itemized
            {
                let deps = vec![
                    F1040_FILING_STATUS.to_string(),
                    SCHED_A_LINE_17.to_string(),
                ];
                FieldDef::new_computed("12", "Standard deduction or itemized deductions", deps, Box::new(|dv: &DepValues| {
                    let fs_str = dv.get_string(F1040_FILING_STATUS);
                    let fs = taxmath::FilingStatus::from_str_code(&fs_str).unwrap_or(taxmath::FilingStatus::Single);
                    let standard = taxmath::get_standard_deduction(dv.tax_year(), taxmath::JurisdictionType::Federal, fs);
                    let itemized = dv.get(SCHED_A_LINE_17);
                    standard.max(itemized)
                }))
            },
            // Line 13: QBI deduction
            ref_field("13", "Qualified business income deduction", F8995_LINE_10),
            // Line 14: Total deductions
            sum_field("14", "Total deductions", vec![F1040_LINE_12, F1040_LINE_13]),
            // Line 15: Taxable income
            max_zero_field("15", "Taxable income", F1040_LINE_11, F1040_LINE_14),
            // Line 16: Tax (with FEIE stacking)
            {
                let deps = vec![
                    F1040_LINE_15.to_string(),
                    F1040_FILING_STATUS.to_string(),
                    F2555_TOTAL_EXCLUSION.to_string(),
                ];
                FieldDef::new_computed("16", "Tax", deps, Box::new(|dv: &DepValues| {
                    let fs_str = dv.get_string(F1040_FILING_STATUS);
                    let fs = taxmath::FilingStatus::from_str_code(&fs_str).unwrap_or(taxmath::FilingStatus::Single);
                    let taxable_income = dv.get(F1040_LINE_15);
                    let excluded_income = dv.get(F2555_TOTAL_EXCLUSION);
                    if excluded_income > 0.0 {
                        taxmath::compute_tax_with_stacking(taxable_income, excluded_income, fs, dv.tax_year(), taxmath::JurisdictionType::Federal)
                    } else {
                        taxmath::compute_tax(taxable_income, fs, dv.tax_year(), taxmath::JurisdictionType::Federal)
                    }
                }))
            },
            // Line 17
            ref_field("17", "Amount from Schedule 2, Part I", SCHED_2_LINE_3),
            // Line 20
            ref_field("20", "Amount from Schedule 3, Part I", SCHED_3_LINE_8),
            // Line 22: Tax after credits
            {
                let deps = vec![
                    F1040_LINE_16.to_string(),
                    F1040_LINE_17.to_string(),
                    F1040_LINE_20.to_string(),
                ];
                FieldDef::new_computed("22", "Tax after credits", deps, Box::new(|dv: &DepValues| {
                    (dv.get(F1040_LINE_16) + dv.get(F1040_LINE_17) - dv.get(F1040_LINE_20)).max(0.0)
                }))
            },
            // Line 23
            ref_field("23", "Other taxes from Schedule 2", SCHED_2_LINE_21),
            // Line 24: Total tax
            sum_field("24", "Total tax", vec![F1040_LINE_22, F1040_LINE_23]),
            // --- Payments ---
            // Line 25a
            wildcard_sum_field("25a", "Federal income tax withheld from W-2s", W2_WILDCARD_FED_TAX_WH),
            // Line 25b
            {
                let deps = vec![
                    F1099_INT_WILDCARD_FED_TAX_WH.to_string(),
                    F1099_DIV_WILDCARD_FED_TAX_WH.to_string(),
                    F1099_NEC_WILDCARD_FED_TAX_WH.to_string(),
                    F1099_B_WILDCARD_FED_TAX_WH.to_string(),
                ];
                FieldDef::new_computed("25b", "Federal income tax withheld from 1099s", deps, Box::new(|dv: &DepValues| {
                    dv.sum_all(F1099_INT_WILDCARD_FED_TAX_WH)
                        + dv.sum_all(F1099_DIV_WILDCARD_FED_TAX_WH)
                        + dv.sum_all(F1099_NEC_WILDCARD_FED_TAX_WH)
                        + dv.sum_all(F1099_B_WILDCARD_FED_TAX_WH)
                }))
            },
            // Line 25d
            sum_field("25d", "Total federal tax withheld", vec![F1040_LINE_25A, F1040_LINE_25B]),
            // Line 31
            ref_field("31", "Other payments from Schedule 3", SCHED_3_LINE_15),
            // Line 33
            sum_field("33", "Total payments", vec![F1040_LINE_25D, F1040_LINE_31]),
            // --- Refund or Amount Owed ---
            // Line 34
            max_zero_field("34", "Overpayment (refund)", F1040_LINE_33, F1040_LINE_24),
            // Line 37
            max_zero_field("37", "Amount you owe", F1040_LINE_24, F1040_LINE_33),
        ],
    }
}
