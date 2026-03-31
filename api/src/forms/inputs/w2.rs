use crate::domain::builders::*;
use crate::domain::field::*;
use crate::domain::form::*;

/// W-2 Wage and Tax Statement (US employers only).
///
/// W-2 forms are issued by US employers only. Foreign employers do not issue W-2s;
/// foreign wages are entered separately on Form 1040.
/// Instance-keyed: fields are prefixed with "w2:1:", "w2:2:", etc. at runtime.
pub fn form_w2() -> FormDef {
    FormDef {
        id: FORM_W2.to_string(),
        name: "W-2 Wage and Tax Statement (US employers only)".to_string(),
        jurisdiction: Jurisdiction::Federal,
        tax_years: vec![2024, 2025, 2026],
        question_group: "income_w2".to_string(),
        question_order: 2,
        fields: vec![
            string_input_field(
                "employer_name",
                "Employer name",
                "What is the US employer's name? (Skip this form if your employer is foreign \
                 -- foreign wages are entered separately)",
            ),
            string_input_field(
                "employer_ein",
                "Employer EIN",
                "What is the employer's EIN (XX-XXXXXXX)?",
            ),
            input_field(
                "wages",
                "Box 1: Wages, tips, other compensation",
                "Enter Box 1 -- Wages, tips, other compensation:",
            ),
            input_field(
                "federal_tax_withheld",
                "Box 2: Federal income tax withheld",
                "Enter Box 2 -- Federal income tax withheld:",
            ),
            input_field(
                "ss_wages",
                "Box 3: Social security wages",
                "Enter Box 3 -- Social security wages:",
            ),
            input_field(
                "ss_tax_withheld",
                "Box 4: Social security tax withheld",
                "Enter Box 4 -- Social security tax withheld:",
            ),
            input_field(
                "medicare_wages",
                "Box 5: Medicare wages and tips",
                "Enter Box 5 -- Medicare wages and tips:",
            ),
            input_field(
                "medicare_tax_withheld",
                "Box 6: Medicare tax withheld",
                "Enter Box 6 -- Medicare tax withheld:",
            ),
            input_field(
                "state_wages",
                "Box 16: State wages, tips, etc.",
                "Enter Box 16 -- State wages, tips, etc.:",
            ),
            input_field(
                "state_tax_withheld",
                "Box 17: State income tax withheld",
                "Enter Box 17 -- State income tax withheld:",
            ),
        ],
    }
}
