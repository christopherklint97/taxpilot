use crate::domain::builders::*;
use crate::domain::field::*;
use crate::domain::form::*;

pub fn schedule_1() -> FormDef {
    FormDef {
        id: FORM_SCHEDULE_1.to_string(),
        name: "Schedule 1 — Additional Income and Adjustments to Income".to_string(),
        jurisdiction: Jurisdiction::Federal,
        tax_years: vec![2024, 2025, 2026],
        question_group: "income_w2".to_string(),
        question_order: 2,
        fields: vec![
            // --- Part I: Additional Income ---
            // Line 1: Taxable refunds (deferred)
            zero_field("1", "Taxable refunds of state/local income taxes"),
            // Line 2a: Alimony (deferred)
            zero_field("2a", "Alimony received"),
            // Line 3: Business income
            ref_field("3", "Business income or (loss)", SCHED_C_LINE_31),
            // Line 7: Capital gain or loss
            ref_field("7", "Capital gain or (loss)", SCHED_D_LINE_16),
            // Line 8d: FEIE
            neg_field("8d", "Foreign earned income exclusion (Form 2555)", F2555_TOTAL_EXCLUSION),
            // Line 10: Total additional income
            sum_field("10", "Total additional income", vec![
                SCHED_1_LINE_1, "schedule_1:2a", SCHED_1_LINE_3, SCHED_1_LINE_7, SCHED_1_LINE_8D,
            ]),
            // --- Part II: Adjustments to Income ---
            // Line 11: Educator expenses (deferred)
            zero_field("11", "Educator expenses"),
            // Line 15: HSA deduction
            ref_field("15", "HSA deduction", F8889_LINE_9),
            // Line 16: SE tax deduction
            ref_field("16", "Deductible part of self-employment tax", SCHED_SE_LINE_7),
            // Line 20: IRA deduction (deferred)
            zero_field("20", "IRA deduction"),
            // Line 21: Student loan interest (deferred)
            zero_field("21", "Student loan interest deduction"),
            // Line 24: Early withdrawal penalty
            wildcard_sum_field("24", "Penalty on early withdrawal of savings", F1099_INT_WILDCARD_PENALTY),
            // Line 26: Total adjustments
            sum_field("26", "Total adjustments to income", vec![
                "schedule_1:11", SCHED_1_LINE_15, SCHED_1_LINE_16,
                "schedule_1:20", "schedule_1:21", SCHED_1_LINE_24,
            ]),
        ],
    }
}
