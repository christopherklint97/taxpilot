use crate::domain::builders::*;
use crate::domain::field::*;
use crate::domain::form::*;

// Self-employment tax constants for 2025
const SS_TAX_RATE: f64 = 0.124;
const MEDICARE_TAX_RATE: f64 = 0.029;
const SE_TAX_RATE: f64 = 0.9235;
const SS_WAGE_BASE_2025: f64 = 176_100.0;

pub fn schedule_se() -> FormDef {
    FormDef {
        id: FORM_SCHEDULE_SE.to_string(),
        name: "Schedule SE — Self-Employment Tax".to_string(),
        jurisdiction: Jurisdiction::Federal,
        tax_years: vec![2024, 2025, 2026],
        question_group: "business".to_string(),
        question_order: 5,
        fields: vec![
            // Line 2
            ref_field("2", "Net earnings from self-employment", SCHED_C_LINE_31),
            // Line 3: 92.35% of line 2 (if > $400)
            {
                let deps = vec![SCHED_SE_LINE_2.to_string()];
                FieldDef::new_computed("3", "Self-employment earnings subject to tax", deps, Box::new(|dv: &DepValues| {
                    let net = dv.get(SCHED_SE_LINE_2);
                    if net < 400.0 { 0.0 } else { net * SE_TAX_RATE }
                }))
            },
            // Line 4: Social Security tax portion
            {
                let deps = vec![
                    SCHED_SE_LINE_3.to_string(),
                    W2_WILDCARD_SS_WAGES.to_string(),
                ];
                FieldDef::new_computed("4", "Social Security tax", deps, Box::new(|dv: &DepValues| {
                    let se_earnings = dv.get(SCHED_SE_LINE_3);
                    if se_earnings <= 0.0 {
                        return 0.0;
                    }
                    let w2_ss_wages = dv.sum_all(W2_WILDCARD_SS_WAGES);
                    let remaining_base = (SS_WAGE_BASE_2025 - w2_ss_wages).max(0.0);
                    let taxable_for_ss = se_earnings.min(remaining_base);
                    taxable_for_ss * SS_TAX_RATE
                }))
            },
            // Line 5: Medicare tax
            {
                let deps = vec![SCHED_SE_LINE_3.to_string()];
                FieldDef::new_computed("5", "Medicare tax", deps, Box::new(|dv: &DepValues| {
                    let se_earnings = dv.get(SCHED_SE_LINE_3);
                    if se_earnings <= 0.0 { 0.0 } else { se_earnings * MEDICARE_TAX_RATE }
                }))
            },
            // Line 6: Total SE tax
            sum_field("6", "Self-employment tax", vec![SCHED_SE_LINE_4, SCHED_SE_LINE_5]),
            // Line 7: Deductible part (50%)
            {
                let deps = vec![SCHED_SE_LINE_6.to_string()];
                FieldDef::new_computed("7", "Deductible part of self-employment tax", deps, Box::new(|dv: &DepValues| {
                    dv.get(SCHED_SE_LINE_6) * 0.5
                }))
            },
        ],
    }
}
