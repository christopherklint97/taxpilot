use crate::domain::builders::*;
use crate::domain::field::*;
use crate::domain::form::*;

// HSA contribution limits for 2025
const HSA_SELF_ONLY_2025: f64 = 4300.0;
const HSA_FAMILY_2025: f64 = 8550.0;
const HSA_CATCH_UP: f64 = 1000.0;
const HSA_PENALTY_RATE: f64 = 0.20;

pub fn form_8889() -> FormDef {
    FormDef {
        id: FORM_F8889.to_string(),
        name: "Form 8889 — Health Savings Accounts".to_string(),
        jurisdiction: Jurisdiction::Federal,
        tax_years: vec![2024, 2025, 2026],
        question_group: "deductions".to_string(),
        question_order: 6,
        fields: vec![
            // --- Part I: HSA Contributions and Deduction ---
            enum_field("1", "HSA coverage type", "What type of HDHP coverage do you have?", vec!["self-only", "family"]),
            input_field("2", "HSA contributions made for 2025", "How much did you contribute to your HSA for 2025?"),
            input_field("3", "Employer contributions (including pre-tax payroll)", "How much did your employer contribute to your HSA (W-2 Box 12, code W)?"),
            input_field("5", "Additional catch-up contribution (age 55+)", "Are you age 55 or older? Enter catch-up contribution amount ($1,000 max, or 0):"),
            // Line 6: HSA deduction limit
            {
                let deps = vec![F8889_LINE_1.to_string(), F8889_LINE_5.to_string()];
                FieldDef::new_computed("6", "HSA contribution limit", deps, Box::new(|dv: &DepValues| {
                    let coverage_type = dv.get_string(F8889_LINE_1);
                    let limit = if coverage_type == "family" { HSA_FAMILY_2025 } else { HSA_SELF_ONLY_2025 };
                    let catch_up = dv.get(F8889_LINE_5).min(HSA_CATCH_UP);
                    limit + catch_up
                }))
            },
            // Line 9: HSA deduction
            {
                let deps = vec![F8889_LINE_2.to_string(), F8889_LINE_3.to_string(), F8889_LINE_6.to_string()];
                FieldDef::new_computed("9", "HSA deduction", deps, Box::new(|dv: &DepValues| {
                    let contributions = dv.get(F8889_LINE_2);
                    let employer = dv.get(F8889_LINE_3);
                    let limit = dv.get(F8889_LINE_6);

                    let total = contributions + employer;
                    if total > limit {
                        (limit - employer).max(0.0)
                    } else {
                        contributions.max(0.0)
                    }
                }))
            },
            // --- Part II: HSA Distributions ---
            input_field("14a", "Total HSA distributions", "How much did you receive in HSA distributions in 2025?"),
            input_field("14c", "Qualified medical expenses paid from HSA", "How much of your HSA distributions were for qualified medical expenses?"),
            // Line 15: Taxable HSA distributions
            max_zero_field("15", "Taxable HSA distributions", F8889_LINE_14A, F8889_LINE_14C),
            // Line 17b: Additional 20% tax
            {
                let deps = vec![F8889_LINE_15.to_string()];
                FieldDef::new_computed("17b", "Additional tax on non-qualified HSA distributions (20%)", deps, Box::new(|dv: &DepValues| {
                    dv.get(F8889_LINE_15) * HSA_PENALTY_RATE
                }))
            },
        ],
    }
}
