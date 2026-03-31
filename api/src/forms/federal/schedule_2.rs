use crate::domain::builders::*;
use crate::domain::field::*;
use crate::domain::form::*;

// Schedule 2 constants
const NIIT_RATE: f64 = 0.038;
const NIIT_SINGLE: f64 = 200_000.0;
const NIIT_MFJ: f64 = 250_000.0;
const NIIT_MFS: f64 = 125_000.0;

const ADDL_MEDICARE_RATE: f64 = 0.009;
const ADDL_MEDICARE_SINGLE: f64 = 200_000.0;
const ADDL_MEDICARE_MFJ: f64 = 250_000.0;
const ADDL_MEDICARE_MFS: f64 = 125_000.0;

fn get_niit_threshold(fs: &str) -> f64 {
    match fs {
        "mfj" | "qss" => NIIT_MFJ,
        "mfs" => NIIT_MFS,
        _ => NIIT_SINGLE,
    }
}

fn get_addl_medicare_threshold(fs: &str) -> f64 {
    match fs {
        "mfj" => ADDL_MEDICARE_MFJ,
        "mfs" => ADDL_MEDICARE_MFS,
        _ => ADDL_MEDICARE_SINGLE,
    }
}

pub fn schedule_2() -> FormDef {
    FormDef {
        id: FORM_SCHEDULE_2.to_string(),
        name: "Schedule 2 — Additional Taxes".to_string(),
        jurisdiction: Jurisdiction::Federal,
        tax_years: vec![2024, 2025, 2026],
        question_group: "deductions".to_string(),
        question_order: 6,
        fields: vec![
            // --- Part I: Tax ---
            zero_field("1", "Alternative minimum tax (Form 6251)"),
            zero_field("2", "Excess advance premium tax credit repayment"),
            // Line 3
            sum_field("3", "Total Part I additional taxes", vec![SCHED_2_LINE_1, "schedule_2:2"]),
            // --- Part II: Other Taxes ---
            // Line 6: Self-employment tax
            ref_field("6", "Self-employment tax", SCHED_SE_LINE_6),
            // Line 12: Additional Medicare Tax
            {
                let deps = vec![
                    F1040_FILING_STATUS.to_string(),
                    W2_WILDCARD_MEDICARE_WAGES.to_string(),
                    SCHED_SE_LINE_3.to_string(),
                ];
                FieldDef::new_computed("12", "Additional Medicare Tax", deps, Box::new(|dv: &DepValues| {
                    let fs = dv.get_string(F1040_FILING_STATUS);
                    let threshold = get_addl_medicare_threshold(&fs);

                    let medicare_wages = dv.sum_all(W2_WILDCARD_MEDICARE_WAGES);
                    let se_earnings = dv.get(SCHED_SE_LINE_3);
                    let total_earned = medicare_wages + se_earnings;

                    let excess = total_earned - threshold;
                    if excess <= 0.0 {
                        return 0.0;
                    }

                    // Credit for Additional Medicare Tax already withheld by employer
                    let employer_withheld = if medicare_wages > 200_000.0 {
                        (medicare_wages - 200_000.0) * ADDL_MEDICARE_RATE
                    } else {
                        0.0
                    };

                    let tax = excess * ADDL_MEDICARE_RATE;
                    (tax - employer_withheld).max(0.0)
                }))
            },
            // Line 18: NIIT
            {
                let deps = vec![
                    F1040_FILING_STATUS.to_string(),
                    F1040_LINE_11.to_string(),
                    F1040_LINE_2B.to_string(),
                    F1040_LINE_3B.to_string(),
                    SCHED_D_LINE_16.to_string(),
                ];
                FieldDef::new_computed("18", "Net investment income tax", deps, Box::new(|dv: &DepValues| {
                    let fs = dv.get_string(F1040_FILING_STATUS);
                    let threshold = get_niit_threshold(&fs);

                    let magi = dv.get(F1040_LINE_11);
                    if magi <= threshold {
                        return 0.0;
                    }

                    let nii = dv.get(F1040_LINE_2B) + dv.get(F1040_LINE_3B) + dv.get(SCHED_D_LINE_16);
                    if nii <= 0.0 {
                        return 0.0;
                    }

                    let excess = magi - threshold;
                    let taxable = nii.min(excess);
                    taxable * NIIT_RATE
                }))
            },
            // Line 17c: Additional tax on HSA distributions
            ref_field("17c", "Additional tax on HSA distributions", F8889_LINE_17B),
            // Line 21: Total other taxes
            sum_field("21", "Total other taxes", vec![
                SCHED_2_LINE_6, SCHED_2_LINE_12, SCHED_2_LINE_17C, SCHED_2_LINE_18,
            ]),
        ],
    }
}
