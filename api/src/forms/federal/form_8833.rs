use crate::domain::builders::*;
use crate::domain::field::*;
use crate::domain::form::*;

pub fn form_8833() -> FormDef {
    FormDef {
        id: FORM_F8833.to_string(),
        name: "Form 8833 — Treaty-Based Return Position Disclosure".to_string(),
        jurisdiction: Jurisdiction::Federal,
        tax_years: vec![2024, 2025, 2026],
        question_group: "expat".to_string(),
        question_order: 4,
        fields: vec![
            string_input_field("treaty_country", "Treaty country", "Which country's tax treaty are you relying on?"),
            string_input_field("treaty_article", "Treaty article number", "Which article of the tax treaty applies (e.g., 'Article 18 — Pensions')?"),
            string_input_field("irc_provision", "IRC provision being overridden", "Which IRC section is modified by the treaty position (e.g., 'IRC §61' or 'IRC §871')?"),
            string_input_field("treaty_position_explanation", "Explanation of treaty-based position", "Briefly explain your treaty-based return position:"),
            input_field("treaty_amount", "Amount of income affected by treaty position (USD)", "What is the amount of income subject to this treaty position (in USD)?"),
            {
                FieldDef {
                    line: "num_positions".to_string(),
                    field_type: FieldType::UserInput,
                    value_type: FieldValueType::Integer,
                    label: "Number of treaty positions disclosed".to_string(),
                    prompt: "How many separate treaty positions are you disclosing?".to_string(),
                    depends_on: Vec::new(),
                    options: Vec::new(),
                    compute: None,
                    compute_str: None,
                }
            },

            // --- Computed fields ---
            // Treaty benefit indicator
            {
                let deps = vec![F8833_TREATY_AMOUNT.to_string()];
                FieldDef::new_computed("treaty_claimed", "Treaty position claimed", deps, Box::new(|dv: &DepValues| {
                    if dv.get(F8833_TREATY_AMOUNT) > 0.0 { 1.0 } else { 0.0 }
                }))
            },
        ],
    }
}
