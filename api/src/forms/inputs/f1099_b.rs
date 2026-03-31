use crate::domain::builders::*;
use crate::domain::field::*;
use crate::domain::form::*;

/// 1099-B Proceeds From Broker and Barter Exchange Transactions (US brokers only).
///
/// 1099-B forms are issued by US brokers only. Sales through foreign brokerages
/// that do not issue 1099-B should be reported directly on Form 8949.
/// Each 1099-B represents a single sale or disposition of a security.
/// Instance-keyed: fields are prefixed with "1099b:1:", "1099b:2:", etc. at runtime.
pub fn form_1099_b() -> FormDef {
    FormDef {
        id: FORM_1099_B.to_string(),
        name: "1099-B Proceeds From Broker Transactions (US brokers only)".to_string(),
        jurisdiction: Jurisdiction::Federal,
        tax_years: vec![2024, 2025, 2026],
        question_group: "income_1099".to_string(),
        question_order: 3,
        fields: vec![
            string_input_field(
                "description",
                "Description of property",
                "Describe the security sold from a US broker (e.g., \"100 sh AAPL\"). \
                 Skip this form if the sale was through a foreign brokerage.",
            ),
            string_input_field(
                "date_acquired",
                "Date acquired",
                "When did you acquire this security (MM/DD/YYYY or VARIOUS)?",
            ),
            string_input_field(
                "date_sold",
                "Date sold",
                "When did you sell this security (MM/DD/YYYY)?",
            ),
            input_field(
                "proceeds",
                "Box 1d: Proceeds",
                "Enter Box 1d -- Proceeds (sale price):",
            ),
            input_field(
                "cost_basis",
                "Box 1e: Cost or other basis",
                "Enter Box 1e -- Cost or other basis:",
            ),
            input_field(
                "wash_sale_loss",
                "Box 1g: Wash sale loss disallowed",
                "Enter Box 1g -- Wash sale loss disallowed (0 if none):",
            ),
            input_field(
                "federal_tax_withheld",
                "Box 4: Federal income tax withheld",
                "Enter Box 4 -- Federal income tax withheld (0 if none):",
            ),
            enum_field(
                "term",
                "Short-term or long-term",
                "Was this a short-term or long-term holding?",
                vec!["short", "long"],
            ),
            enum_field(
                "basis_reported",
                "Basis reported to IRS",
                "Was cost basis reported to the IRS by your broker?",
                vec!["yes", "no"],
            ),
        ],
    }
}
