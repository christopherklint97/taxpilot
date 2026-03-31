use crate::domain::builders::*;
use crate::domain::field::*;
use crate::domain::form::*;

pub fn schedule_c() -> FormDef {
    FormDef {
        id: FORM_SCHEDULE_C.to_string(),
        name: "Schedule C — Profit or Loss From Business".to_string(),
        jurisdiction: Jurisdiction::Federal,
        tax_years: vec![2024, 2025, 2026],
        question_group: "business".to_string(),
        question_order: 5,
        fields: vec![
            // --- Business Info ---
            string_input_field("business_name", "Business name", "What is your business name (or your name if sole proprietor)?"),
            string_input_field("business_code", "Principal business code", "What is your principal business activity code (6-digit NAICS)?"),
            // --- Income ---
            // Line 1: Gross receipts
            wildcard_sum_field("1", "Gross receipts or sales", F1099_NEC_WILDCARD_COMP),
            // Line 4: COGS (deferred)
            zero_field("4", "Cost of goods sold"),
            // Line 5: Gross profit
            diff_field("5", "Gross profit", SCHED_C_LINE_1, "schedule_c:4"),
            // Line 7: Gross income
            ref_field("7", "Gross income", SCHED_C_LINE_5),
            // --- Expenses ---
            input_field("8", "Advertising expenses", "Enter advertising expenses:"),
            input_field("10", "Car and truck expenses", "Enter car and truck expenses (business use only):"),
            input_field("17", "Legal and professional services", "Enter legal and professional service fees:"),
            input_field("18", "Office expense", "Enter office expenses:"),
            input_field("22", "Supplies", "Enter supply expenses:"),
            input_field("25", "Utilities", "Enter utility expenses (business portion):"),
            input_field("27", "Other expenses", "Enter other business expenses not listed above:"),
            // Line 28: Total expenses
            sum_field("28", "Total expenses", vec![
                SCHED_C_LINE_8, SCHED_C_LINE_10, SCHED_C_LINE_17,
                SCHED_C_LINE_18, SCHED_C_LINE_22, SCHED_C_LINE_25, SCHED_C_LINE_27,
            ]),
            // Line 31: Net profit or loss
            max_zero_field("31", "Net profit or (loss)", SCHED_C_LINE_7, SCHED_C_LINE_28),
        ],
    }
}
