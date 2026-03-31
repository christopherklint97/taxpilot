use crate::domain::builders::*;
use crate::domain::field::*;
use crate::domain::form::*;

/// 1099-INT Interest Income (US payers only).
///
/// 1099-INT forms are issued by US banks and financial institutions only.
/// Foreign interest income is entered separately on Schedule B.
/// Instance-keyed: fields are prefixed with "1099int:1:", "1099int:2:", etc. at runtime.
pub fn form_1099_int() -> FormDef {
    FormDef {
        id: FORM_1099_INT.to_string(),
        name: "1099-INT Interest Income (US payers only)".to_string(),
        jurisdiction: Jurisdiction::Federal,
        tax_years: vec![2024, 2025, 2026],
        question_group: "income_1099".to_string(),
        question_order: 3,
        fields: vec![
            string_input_field(
                "payer_name",
                "Payer name",
                "What is the US payer's name (from 1099-INT)? (Skip this form if all interest \
                 is from foreign institutions -- foreign interest is entered separately)",
            ),
            string_input_field(
                "payer_tin",
                "Payer TIN",
                "What is the payer's TIN (XX-XXXXXXX)?",
            ),
            input_field(
                "interest_income",
                "Box 1: Interest income",
                "Enter Box 1 -- Interest income:",
            ),
            input_field(
                "early_withdrawal_penalty",
                "Box 2: Early withdrawal penalty",
                "Enter Box 2 -- Early withdrawal penalty (if any):",
            ),
            input_field(
                "us_savings_bond_interest",
                "Box 3: Interest on U.S. Savings Bonds and Treasury obligations",
                "Enter Box 3 -- Interest on U.S. Savings Bonds and Treasury obligations:",
            ),
            input_field(
                "federal_tax_withheld",
                "Box 4: Federal income tax withheld",
                "Enter Box 4 -- Federal income tax withheld:",
            ),
            input_field(
                "tax_exempt_interest",
                "Box 8: Tax-exempt interest",
                "Enter Box 8 -- Tax-exempt interest:",
            ),
            input_field(
                "private_activity_bond_interest",
                "Box 9: Specified private activity bond interest",
                "Enter Box 9 -- Specified private activity bond interest:",
            ),
        ],
    }
}
