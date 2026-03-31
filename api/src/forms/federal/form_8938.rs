use crate::domain::builders::*;
use crate::domain::field::*;
use crate::domain::form::*;
use crate::domain::taxmath;

pub fn form_8938() -> FormDef {
    FormDef {
        id: FORM_F8938.to_string(),
        name: "Form 8938 — Statement of Specified Foreign Financial Assets".to_string(),
        jurisdiction: Jurisdiction::Federal,
        tax_years: vec![2024, 2025, 2026],
        question_group: "expat".to_string(),
        question_order: 4,
        fields: vec![
            // --- Taxpayer situation ---
            enum_field("lives_abroad", "Do you live outside the United States?", "Do you meet the bona fide residence or physical presence test for living abroad?", vec!["yes", "no"]),

            // --- Financial account details ---
            {
                FieldDef {
                    line: "num_accounts".to_string(),
                    field_type: FieldType::UserInput,
                    value_type: FieldValueType::Integer,
                    label: "Number of foreign financial accounts".to_string(),
                    prompt: "How many foreign financial accounts do you have (bank, brokerage, pension, etc.)?".to_string(),
                    depends_on: Vec::new(),
                    options: Vec::new(),
                    compute: None,
                    compute_str: None,
                }
            },
            input_field("max_value_accounts", "Maximum value of foreign accounts during the year (USD)", "What was the maximum aggregate value of all your foreign financial accounts at any time during 2025 (in USD)?"),
            input_field("yearend_value_accounts", "Year-end value of foreign accounts (USD)", "What was the total value of all your foreign financial accounts on December 31, 2025 (in USD)?"),

            // --- Other specified foreign financial assets ---
            {
                FieldDef {
                    line: "num_other_assets".to_string(),
                    field_type: FieldType::UserInput,
                    value_type: FieldValueType::Integer,
                    label: "Number of other specified foreign financial assets".to_string(),
                    prompt: "How many other specified foreign financial assets do you have (foreign stocks, partnership interests, etc.)?".to_string(),
                    depends_on: Vec::new(),
                    options: Vec::new(),
                    compute: None,
                    compute_str: None,
                }
            },
            input_field("max_value_other", "Maximum value of other foreign assets during the year (USD)", "What was the maximum aggregate value of your other specified foreign financial assets at any time during 2025 (in USD)?"),
            input_field("yearend_value_other", "Year-end value of other foreign assets (USD)", "What was the total value of your other specified foreign financial assets on December 31, 2025 (in USD)?"),

            // --- Account identification ---
            string_input_field("account_country", "Country of foreign accounts", "In which country are your foreign financial accounts held?"),
            string_input_field("account_institution", "Foreign financial institution name", "What is the name of the foreign financial institution?"),
            enum_field("account_type", "Type of foreign account", "What type of foreign financial account is it?", vec!["deposit", "custodial", "pension", "other"]),

            // --- Income from foreign assets ---
            input_field("income_from_accounts", "Income from foreign financial accounts (USD)", "How much income was earned from your foreign financial accounts in 2025 (interest, dividends, etc. in USD)?"),
            input_field("gain_from_accounts", "Gains from foreign financial assets (USD)", "How much in gains (or losses) did you realize from your foreign financial assets in 2025 (in USD)?"),

            // --- Computed fields ---

            // Total maximum value
            sum_field("total_max_value", "Total maximum value of all foreign assets", vec![F8938_MAX_VALUE_ACCOUNTS, F8938_MAX_VALUE_OTHER]),
            // Total year-end value
            sum_field("total_yearend_value", "Total year-end value of all foreign assets", vec![F8938_YEAREND_ACCOUNTS, F8938_YEAREND_OTHER]),
            // Year-end threshold
            {
                let deps = vec![F8938_LIVES_ABROAD.to_string(), F1040_FILING_STATUS.to_string()];
                FieldDef::new_computed("threshold_yearend", "FATCA year-end filing threshold", deps, Box::new(|dv: &DepValues| {
                    let cfg = taxmath::get_config_or_default(dv.tax_year());
                    let abroad = dv.get_string(F8938_LIVES_ABROAD) == "yes";
                    let fs = dv.get_string(F1040_FILING_STATUS);
                    let mfj = fs == "mfj";

                    if abroad {
                        if mfj { cfg.fatca_abroad_mfj_year_end } else { cfg.fatca_abroad_single_year_end }
                    } else if mfj {
                        cfg.fatca_us_mfj_year_end
                    } else {
                        cfg.fatca_us_single_year_end
                    }
                }))
            },
            // Any-time threshold
            {
                let deps = vec![F8938_LIVES_ABROAD.to_string(), F1040_FILING_STATUS.to_string()];
                FieldDef::new_computed("threshold_anytime", "FATCA any-time filing threshold", deps, Box::new(|dv: &DepValues| {
                    let cfg = taxmath::get_config_or_default(dv.tax_year());
                    let abroad = dv.get_string(F8938_LIVES_ABROAD) == "yes";
                    let fs = dv.get_string(F1040_FILING_STATUS);
                    let mfj = fs == "mfj";

                    if abroad {
                        if mfj { cfg.fatca_abroad_mfj_any_time } else { cfg.fatca_abroad_single_any_time }
                    } else if mfj {
                        cfg.fatca_us_mfj_any_time
                    } else {
                        cfg.fatca_us_single_any_time
                    }
                }))
            },
            // Filing required
            {
                let deps = vec![
                    F8938_TOTAL_MAX_VALUE.to_string(),
                    F8938_TOTAL_YEAREND_VALUE.to_string(),
                    "form_8938:threshold_yearend".to_string(),
                    "form_8938:threshold_anytime".to_string(),
                ];
                FieldDef::new_computed("filing_required", "Form 8938 filing required", deps, Box::new(|dv: &DepValues| {
                    let max_val = dv.get(F8938_TOTAL_MAX_VALUE);
                    let year_end = dv.get(F8938_TOTAL_YEAREND_VALUE);
                    let thresh_ye = dv.get("form_8938:threshold_yearend");
                    let thresh_at = dv.get("form_8938:threshold_anytime");

                    if year_end > thresh_ye || max_val > thresh_at { 1.0 } else { 0.0 }
                }))
            },
        ],
    }
}
