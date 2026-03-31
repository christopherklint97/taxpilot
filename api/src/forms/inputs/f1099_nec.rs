use crate::domain::builders::*;
use crate::domain::field::*;
use crate::domain::form::*;

/// 1099-NEC Nonemployee Compensation (US payers only).
///
/// 1099-NEC forms are issued by US clients and companies only.
/// Foreign freelance/contractor income is entered as foreign self-employment income.
/// Instance-keyed: fields are prefixed with "1099nec:1:", "1099nec:2:", etc. at runtime.
pub fn form_1099_nec() -> FormDef {
    FormDef {
        id: FORM_1099_NEC.to_string(),
        name: "1099-NEC Nonemployee Compensation (US payers only)".to_string(),
        jurisdiction: Jurisdiction::Federal,
        tax_years: vec![2024, 2025, 2026],
        question_group: "income_1099".to_string(),
        question_order: 3,
        fields: vec![
            string_input_field(
                "payer_name",
                "Payer name",
                "What is the US payer's name (from 1099-NEC)? (Skip this form if the payer is \
                 foreign -- foreign contractor income is entered separately)",
            ),
            string_input_field(
                "payer_tin",
                "Payer TIN",
                "What is the payer's TIN (XX-XXXXXXX)?",
            ),
            input_field(
                "nonemployee_compensation",
                "Box 1: Nonemployee compensation",
                "Enter Box 1 -- Nonemployee compensation:",
            ),
            input_field(
                "federal_tax_withheld",
                "Box 4: Federal income tax withheld",
                "Enter Box 4 -- Federal income tax withheld:",
            ),
        ],
    }
}
