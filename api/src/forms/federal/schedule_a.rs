use crate::domain::builders::*;
use crate::domain::field::*;
use crate::domain::form::*;
use crate::domain::taxmath;

pub fn schedule_a() -> FormDef {
    FormDef {
        id: FORM_SCHEDULE_A.to_string(),
        name: "Schedule A — Itemized Deductions".to_string(),
        jurisdiction: Jurisdiction::Federal,
        tax_years: vec![2024, 2025, 2026],
        question_group: "deductions".to_string(),
        question_order: 6,
        fields: vec![
            // --- Medical and Dental Expenses ---
            input_field("1", "Medical and dental expenses", "Enter your total medical and dental expenses:"),
            // Line 2: AGI
            ref_field("2", "AGI (from Form 1040 line 11)", F1040_LINE_11),
            // Line 3: 7.5% of AGI
            {
                let deps = vec![SCHED_A_LINE_2.to_string()];
                FieldDef::new_computed("3", "7.5% of AGI", deps, Box::new(|dv: &DepValues| {
                    dv.get(SCHED_A_LINE_2) * 0.075
                }))
            },
            // Line 4: Deductible medical expenses
            max_zero_field("4", "Deductible medical and dental expenses", SCHED_A_LINE_1, SCHED_A_LINE_3),
            // --- Taxes You Paid ---
            input_field("5a", "State and local income taxes paid", "Enter state and local income taxes paid (or general sales taxes):"),
            input_field("5b", "State and local personal property taxes", "Enter state and local personal property taxes paid:"),
            input_field("5c", "State and local real estate taxes", "Enter state and local real estate taxes paid:"),
            // Line 5d
            sum_field("5d", "Total state and local taxes", vec![SCHED_A_LINE_5A, SCHED_A_LINE_5B, SCHED_A_LINE_5C]),
            // Line 5e: SALT cap
            {
                let deps = vec![
                    SCHED_A_LINE_5D.to_string(),
                    F1040_FILING_STATUS.to_string(),
                ];
                FieldDef::new_computed("5e", "State and local taxes (SALT) deduction", deps, Box::new(|dv: &DepValues| {
                    let total = dv.get(SCHED_A_LINE_5D);
                    let fs_str = dv.get_string(F1040_FILING_STATUS);
                    let fs = taxmath::FilingStatus::from_str_code(&fs_str).unwrap_or(taxmath::FilingStatus::Single);
                    let cap = if fs == taxmath::FilingStatus::MarriedFilingSep { 5000.0 } else { 10000.0 };
                    total.min(cap)
                }))
            },
            // --- Interest You Paid ---
            input_field("8a", "Home mortgage interest and points (Form 1098)", "Enter home mortgage interest and points reported on Form 1098:"),
            // Line 10: Investment interest (deferred)
            zero_field("10", "Investment interest"),
            // Line 11
            sum_field("11", "Total interest deduction", vec![SCHED_A_LINE_8A, "schedule_a:10"]),
            // --- Gifts to Charity ---
            input_field("12", "Charitable contributions (cash or check)", "Enter charitable contributions paid by cash or check:"),
            input_field("13", "Charitable contributions (other than cash)", "Enter charitable contributions other than cash or check:"),
            input_field("14", "Charitable contribution carryover from prior year", "Enter charitable contribution carryover from prior year (if any):"),
            // Line 15
            sum_field("15", "Total charitable contributions", vec![SCHED_A_LINE_12, SCHED_A_LINE_13, SCHED_A_LINE_14]),
            // Line 16: Casualty losses (deferred)
            zero_field("16", "Casualty and theft losses"),
            // Line 17: Total itemized deductions
            sum_field("17", "Total itemized deductions", vec![
                SCHED_A_LINE_4, SCHED_A_LINE_5E, SCHED_A_LINE_11, SCHED_A_LINE_15, "schedule_a:16",
            ]),
        ],
    }
}
