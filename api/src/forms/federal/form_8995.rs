use crate::domain::builders::*;
use crate::domain::field::*;
use crate::domain::form::*;

const QBI_RATE: f64 = 0.20;
const QBI_THRESHOLD_SINGLE: f64 = 191_950.0;
const QBI_THRESHOLD_MFJ: f64 = 383_900.0;

fn get_qbi_threshold(fs: &str) -> f64 {
    match fs {
        "mfj" | "qss" => QBI_THRESHOLD_MFJ,
        _ => QBI_THRESHOLD_SINGLE,
    }
}

pub fn form_8995() -> FormDef {
    FormDef {
        id: FORM_F8995.to_string(),
        name: "Qualified Business Income Deduction (Simplified)".to_string(),
        jurisdiction: Jurisdiction::Federal,
        tax_years: vec![2024, 2025, 2026],
        question_group: "business".to_string(),
        question_order: 5,
        fields: vec![
            // Line 1: Total QBI
            ref_field("1", "Total qualified business income", SCHED_C_LINE_31),
            // Line 2: REIT dividends
            wildcard_sum_field("2", "Qualified REIT dividends and PTP income", F1099_DIV_WILDCARD_SEC_199A),
            // Line 3
            sum_field("3", "Combinable QBI and REIT/PTP amounts", vec!["form_8995:1", "form_8995:2"]),
            // Line 4: QBI component (20%)
            {
                let deps = vec![F8995_LINE_3.to_string()];
                FieldDef::new_computed("4", "QBI component (20% of qualified income)", deps, Box::new(|dv: &DepValues| {
                    let qbi = dv.get(F8995_LINE_3);
                    if qbi <= 0.0 { 0.0 } else { qbi * QBI_RATE }
                }))
            },
            // Line 5: Taxable income before QBI
            max_zero_field("5", "Taxable income before QBI deduction", F1040_LINE_11, F1040_LINE_12),
            // Line 6: Net capital gain
            {
                let deps = vec![SCHED_D_LINE_16.to_string()];
                FieldDef::new_computed("6", "Net capital gain", deps, Box::new(|dv: &DepValues| {
                    dv.get(SCHED_D_LINE_16).max(0.0)
                }))
            },
            // Line 7
            max_zero_field("7", "Taxable income minus net capital gain", F8995_LINE_5, "form_8995:6"),
            // Line 8: Income limitation
            {
                let deps = vec!["form_8995:7".to_string()];
                FieldDef::new_computed("8", "Income limitation (20% of adjusted taxable income)", deps, Box::new(|dv: &DepValues| {
                    dv.get("form_8995:7") * QBI_RATE
                }))
            },
            // Line 10: QBI deduction
            {
                let deps = vec![
                    F8995_LINE_4.to_string(),
                    F8995_LINE_5.to_string(),
                    F8995_LINE_8.to_string(),
                    F1040_FILING_STATUS.to_string(),
                ];
                FieldDef::new_computed("10", "Qualified business income deduction", deps, Box::new(|dv: &DepValues| {
                    let taxable_income = dv.get(F8995_LINE_5);
                    let fs = dv.get_string(F1040_FILING_STATUS);
                    let threshold = get_qbi_threshold(&fs);

                    if taxable_income > threshold {
                        return 0.0;
                    }

                    let qbi_component = dv.get(F8995_LINE_4);
                    let income_limit = dv.get(F8995_LINE_8);
                    qbi_component.min(income_limit).max(0.0)
                }))
            },
        ],
    }
}
