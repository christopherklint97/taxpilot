use crate::domain::builders::*;
use crate::domain::field::*;
use crate::domain::form::*;

pub fn schedule_b() -> FormDef {
    FormDef {
        id: FORM_SCHEDULE_B.to_string(),
        name: "Schedule B — Interest and Ordinary Dividends".to_string(),
        jurisdiction: Jurisdiction::Federal,
        tax_years: vec![2024, 2025, 2026],
        question_group: "income_1099".to_string(),
        question_order: 3,
        fields: vec![
            // --- Part I: Interest ---
            input_field(
                "foreign_interest",
                "Foreign interest income (not on 1099-INT)",
                "Enter interest income from foreign banks or institutions (not reported on a 1099-INT), converted to USD:",
            ),
            string_input_field(
                "foreign_interest_payer",
                "Foreign interest payer(s)",
                "Describe the foreign payer(s) of interest income (e.g., \"Nordea Bank, Sweden\"):",
            ),
            // Line 1: Interest income from all 1099-INT + foreign
            {
                let deps = vec![
                    F1099_INT_WILDCARD_INTEREST.to_string(),
                    SCHED_B_FOREIGN_INTEREST.to_string(),
                ];
                FieldDef::new_computed("1", "Interest income (1099-INT + foreign sources)", deps, Box::new(|dv: &DepValues| {
                    dv.sum_all(F1099_INT_WILDCARD_INTEREST) + dv.get(SCHED_B_FOREIGN_INTEREST)
                }))
            },
            // Line 2: Excludable savings bond interest
            zero_field("2", "Excludable savings bond interest"),
            // Line 4: Total interest
            diff_field("4", "Total interest", SCHED_B_LINE_1, "schedule_b:2"),
            // --- Part II: Ordinary Dividends ---
            // Line 5
            wildcard_sum_field("5", "Ordinary dividends (from 1099-DIV Box 1a)", F1099_DIV_WILDCARD_ORDINARY),
            // Line 6
            ref_field("6", "Total ordinary dividends", SCHED_B_LINE_5),
            // --- Part III: Foreign Accounts and Trusts ---
            enum_field(
                "7a",
                "Foreign financial accounts",
                "At any time during 2025, did you have a financial interest in or signature authority over a financial account in a foreign country (e.g., bank account, securities account)?",
                vec!["yes", "no"],
            ),
            string_input_field(
                "7b",
                "Country of foreign accounts",
                "In which country or countries are the foreign accounts located?",
            ),
            enum_field(
                "8",
                "Foreign trust",
                "Did you receive a distribution from, or were you a grantor of, or transferor to, a foreign trust?",
                vec!["yes", "no"],
            ),
            // FBAR required flag
            {
                let deps = vec![SCHED_B_LINE_7A.to_string()];
                FieldDef::new_computed("fbar_required", "FBAR filing required", deps, Box::new(|dv: &DepValues| {
                    if dv.get_string(SCHED_B_LINE_7A) == "yes" { 1.0 } else { 0.0 }
                }))
            },
        ],
    }
}
