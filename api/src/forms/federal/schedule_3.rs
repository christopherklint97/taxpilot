use crate::domain::builders::*;
use crate::domain::field::*;
use crate::domain::form::*;

pub fn schedule_3() -> FormDef {
    FormDef {
        id: FORM_SCHEDULE_3.to_string(),
        name: "Schedule 3 — Additional Credits and Payments".to_string(),
        jurisdiction: Jurisdiction::Federal,
        tax_years: vec![2024, 2025, 2026],
        question_group: "deductions".to_string(),
        question_order: 6,
        fields: vec![
            // --- Part I: Nonrefundable Credits ---
            // Line 1: Foreign tax credit
            ref_field("1", "Foreign tax credit", F1116_LINE_22),
            // Line 2: Child and dependent care credit (deferred)
            zero_field("2", "Child and dependent care credit"),
            // Line 3: Education credits (deferred)
            zero_field("3", "Education credits"),
            // Line 8: Total nonrefundable credits
            sum_field("8", "Total nonrefundable credits", vec![SCHED_3_LINE_1, "schedule_3:2", "schedule_3:3"]),
            // --- Part II: Other Payments ---
            // Line 10: Estimated tax payments
            input_field("10", "Estimated tax payments for 2025", "How much did you pay in federal estimated taxes for 2025?"),
            // Line 15
            ref_field("15", "Total other payments", SCHED_3_LINE_10),
        ],
    }
}
