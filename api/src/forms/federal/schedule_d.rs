use crate::domain::builders::*;
use crate::domain::field::*;
use crate::domain::form::*;

pub fn schedule_d() -> FormDef {
    FormDef {
        id: FORM_SCHEDULE_D.to_string(),
        name: "Schedule D — Capital Gains and Losses".to_string(),
        jurisdiction: Jurisdiction::Federal,
        tax_years: vec![2024, 2025, 2026],
        question_group: "income_1099".to_string(),
        question_order: 3,
        fields: vec![
            // --- Part I: Short-Term ---
            ref_field("1", "Short-term from Form 8949", F8949_ST_GAIN_LOSS),
            ref_field("7", "Net short-term capital gain or (loss)", SCHED_D_LINE_1),
            // --- Part II: Long-Term ---
            ref_field("8", "Long-term from Form 8949", F8949_LT_GAIN_LOSS),
            // Line 13: Capital gain distributions
            wildcard_sum_field("13", "Capital gain distributions", F1099_DIV_WILDCARD_CAP_GAIN),
            // Line 15
            sum_field("15", "Net long-term capital gain or (loss)", vec![SCHED_D_LINE_8, SCHED_D_LINE_13]),
            // --- Part III: Summary ---
            // Line 16
            sum_field("16", "Net capital gain or (loss)", vec![SCHED_D_LINE_7, SCHED_D_LINE_15]),
        ],
    }
}
