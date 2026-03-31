use crate::domain::builders::*;
use crate::domain::field::*;
use crate::domain::form::*;

/// 1099-DIV Dividends and Distributions (US payers only).
///
/// 1099-DIV forms are issued by US brokerages and fund companies only.
/// Foreign dividends are reported directly on Schedule B and Form 1040.
/// Instance-keyed: fields are prefixed with "1099div:1:", "1099div:2:", etc. at runtime.
pub fn form_1099_div() -> FormDef {
    FormDef {
        id: FORM_1099_DIV.to_string(),
        name: "1099-DIV Dividends and Distributions (US payers only)".to_string(),
        jurisdiction: Jurisdiction::Federal,
        tax_years: vec![2024, 2025, 2026],
        question_group: "income_1099".to_string(),
        question_order: 3,
        fields: vec![
            string_input_field(
                "payer_name",
                "Payer name",
                "What is the US payer's name (from 1099-DIV)? (Skip this form if all dividends \
                 are from foreign sources -- foreign dividends are reported separately)",
            ),
            string_input_field(
                "payer_tin",
                "Payer TIN",
                "What is the payer's TIN (XX-XXXXXXX)?",
            ),
            input_field(
                "ordinary_dividends",
                "Box 1a: Total ordinary dividends",
                "Enter Box 1a -- Total ordinary dividends:",
            ),
            input_field(
                "qualified_dividends",
                "Box 1b: Qualified dividends",
                "Enter Box 1b -- Qualified dividends:",
            ),
            input_field(
                "total_capital_gain",
                "Box 2a: Total capital gain distributions",
                "Enter Box 2a -- Total capital gain distributions:",
            ),
            input_field(
                "section_1250_gain",
                "Box 2b: Unrecaptured Section 1250 gain",
                "Enter Box 2b -- Unrecaptured Section 1250 gain:",
            ),
            input_field(
                "section_199a_dividends",
                "Box 5: Section 199A dividends",
                "Enter Box 5 -- Section 199A dividends:",
            ),
            input_field(
                "federal_tax_withheld",
                "Box 4: Federal income tax withheld",
                "Enter Box 4 -- Federal income tax withheld:",
            ),
            input_field(
                "exempt_interest_dividends",
                "Box 12: Exempt-interest dividends",
                "Enter Box 12 -- Exempt-interest dividends:",
            ),
            input_field(
                "private_activity_bond_dividends",
                "Box 13: Specified private activity bond interest dividends",
                "Enter Box 13 -- Specified private activity bond interest dividends:",
            ),
        ],
    }
}
