use crate::domain::builders::*;
use crate::domain::field::*;
use crate::domain::form::*;
use crate::domain::taxmath;

pub fn form_2555() -> FormDef {
    FormDef {
        id: FORM_F2555.to_string(),
        name: "Form 2555 — Foreign Earned Income".to_string(),
        jurisdiction: Jurisdiction::Federal,
        tax_years: vec![2024, 2025, 2026],
        question_group: "expat".to_string(),
        question_order: 4,
        fields: vec![
            // --- Part I: General Information ---
            string_input_field("foreign_country", "Foreign country of residence", "What country do you live in?"),
            string_input_field("foreign_address", "Foreign address", "What is your foreign address?"),
            string_input_field("employer_name_2555", "Employer name (foreign)", "What is your foreign employer's name?"),
            enum_field("employer_foreign", "Is your employer a foreign entity?", "Is your employer a foreign (non-US) entity?", vec!["yes", "no"]),
            enum_field("self_employed_abroad", "Self-employed abroad", "Are you self-employed in your foreign country?", vec!["yes", "no"]),

            // --- Part II: Qualifying Test ---
            enum_field(
                "qualifying_test",
                "Qualifying test for FEIE",
                "Which qualifying test do you meet for the Foreign Earned Income Exclusion?",
                vec!["bona_fide_residence", "physical_presence"],
            ),

            // --- Bona Fide Residence Test ---
            string_input_field("bfrt_start_date", "Bona fide residence start date", "When did your bona fide residence in a foreign country begin? (MM/DD/YYYY)"),
            string_input_field("bfrt_end_date", "Bona fide residence end date", "When did (or will) your bona fide residence end? (MM/DD/YYYY or 'continuing')"),
            enum_field("bfrt_full_year", "Bona fide resident for full tax year", "Were you a bona fide resident of a foreign country for the entire tax year?", vec!["yes", "no"]),

            // --- Physical Presence Test ---
            {
                FieldDef {
                    line: "ppt_days_present".to_string(),
                    field_type: FieldType::UserInput,
                    value_type: FieldValueType::Integer,
                    label: "Days physically present in foreign country".to_string(),
                    prompt: "How many days were you physically present in a foreign country during the 12-month qualifying period?".to_string(),
                    depends_on: Vec::new(),
                    options: Vec::new(),
                    compute: None,
                    compute_str: None,
                }
            },
            string_input_field("ppt_period_start", "Physical presence period start", "What is the start date of your 12-month qualifying period? (MM/DD/YYYY)"),
            string_input_field("ppt_period_end", "Physical presence period end", "What is the end date of your 12-month qualifying period? (MM/DD/YYYY)"),

            // --- Part III: Foreign Earned Income ---
            input_field("foreign_earned_income", "Total foreign earned income (in USD)", "What was your total foreign earned income for 2025 (converted to USD)?"),
            string_input_field("currency_code", "Foreign currency code", "What currency were you paid in? (e.g., SEK, EUR, GBP)"),
            input_field("exchange_rate", "Average exchange rate (foreign currency per USD)", "What average exchange rate did you use to convert to USD? (IRS yearly average recommended)"),
            input_field("foreign_tax_paid", "Foreign income taxes paid (in USD)", "How much foreign income tax did you pay in 2025 (converted to USD)?"),

            // --- Part IV: Housing ---
            input_field("employer_provided_housing", "Employer-provided housing amounts", "How much did your employer provide for housing (included in income)?"),
            input_field("housing_expenses", "Qualifying foreign housing expenses", "What were your total qualifying foreign housing expenses (rent, utilities, insurance, etc.)?"),

            // --- Computed Fields ---

            // FEIE exclusion limit
            {
                FieldDef::new_computed("exclusion_limit", "Foreign earned income exclusion limit", Vec::new(), Box::new(|dv: &DepValues| {
                    taxmath::feie_limit(dv.tax_year())
                }))
            },
            // Qualifying days
            {
                let deps = vec![
                    F2555_QUALIFYING_TEST.to_string(),
                    F2555_BFRT_FULL_YEAR.to_string(),
                    F2555_PPT_DAYS_PRESENT.to_string(),
                ];
                FieldDef::new_computed("qualifying_days", "Qualifying days", deps, Box::new(|dv: &DepValues| {
                    let test = dv.get_string(F2555_QUALIFYING_TEST);
                    if test == "bona_fide_residence" {
                        let full_year = dv.get_string(F2555_BFRT_FULL_YEAR);
                        if full_year == "yes" {
                            return 365.0;
                        }
                        let days = dv.get(F2555_PPT_DAYS_PRESENT);
                        if days > 0.0 { days } else { 365.0 }
                    } else {
                        dv.get(F2555_PPT_DAYS_PRESENT)
                    }
                }))
            },
            // Prorated exclusion
            {
                let deps = vec![
                    F2555_EXCLUSION_LIMIT.to_string(),
                    F2555_QUALIFYING_DAYS.to_string(),
                ];
                FieldDef::new_computed("prorated_exclusion", "Prorated exclusion amount", deps, Box::new(|dv: &DepValues| {
                    let limit = dv.get(F2555_EXCLUSION_LIMIT);
                    let days = dv.get(F2555_QUALIFYING_DAYS) as i32;
                    taxmath::prorate_exclusion(limit, days, 365)
                }))
            },
            // Foreign income exclusion
            {
                let deps = vec![
                    F2555_FOREIGN_EARNED_INCOME.to_string(),
                    "form_2555:prorated_exclusion".to_string(),
                ];
                FieldDef::new_computed("foreign_income_exclusion", "Foreign earned income exclusion", deps, Box::new(|dv: &DepValues| {
                    let income = dv.get(F2555_FOREIGN_EARNED_INCOME);
                    let limit = dv.get("form_2555:prorated_exclusion");
                    income.min(limit)
                }))
            },

            // --- Housing computation ---
            // Housing base amount
            {
                FieldDef::new_computed("housing_base_amount", "Housing base amount (16% of exclusion limit)", Vec::new(), Box::new(|dv: &DepValues| {
                    taxmath::housing_base_amount(dv.tax_year())
                }))
            },
            // Housing max
            {
                FieldDef::new_computed("housing_max", "Maximum housing amount (30% of exclusion limit)", Vec::new(), Box::new(|dv: &DepValues| {
                    taxmath::housing_max_amount(dv.tax_year())
                }))
            },
            // Housing qualifying amount
            {
                let deps = vec![
                    F2555_HOUSING_EXPENSES.to_string(),
                    "form_2555:housing_max".to_string(),
                    "form_2555:housing_base_amount".to_string(),
                ];
                FieldDef::new_computed("housing_qualifying_amount", "Qualifying housing amount", deps, Box::new(|dv: &DepValues| {
                    let expenses = dv.get(F2555_HOUSING_EXPENSES);
                    let max_amt = dv.get("form_2555:housing_max");
                    let base_amt = dv.get("form_2555:housing_base_amount");
                    let limited = expenses.min(max_amt);
                    (limited - base_amt).max(0.0)
                }))
            },
            // Housing exclusion (employees)
            {
                let deps = vec![
                    "form_2555:housing_qualifying_amount".to_string(),
                    F2555_EMPLOYER_HOUSING.to_string(),
                ];
                FieldDef::new_computed("housing_exclusion", "Foreign housing exclusion", deps, Box::new(|dv: &DepValues| {
                    let qualifying = dv.get("form_2555:housing_qualifying_amount");
                    let employer_provided = dv.get(F2555_EMPLOYER_HOUSING);
                    qualifying.min(employer_provided)
                }))
            },
            // Housing deduction (self-employed)
            {
                let deps = vec![
                    "form_2555:housing_qualifying_amount".to_string(),
                    F2555_HOUSING_EXCLUSION.to_string(),
                    F2555_SELF_EMPLOYED_ABROAD.to_string(),
                ];
                FieldDef::new_computed("housing_deduction", "Foreign housing deduction (self-employed)", deps, Box::new(|dv: &DepValues| {
                    let self_employed = dv.get_string(F2555_SELF_EMPLOYED_ABROAD);
                    if self_employed != "yes" {
                        return 0.0;
                    }
                    let qualifying = dv.get("form_2555:housing_qualifying_amount");
                    let exclusion = dv.get(F2555_HOUSING_EXCLUSION);
                    (qualifying - exclusion).max(0.0)
                }))
            },

            // --- Total exclusion ---
            sum_field("total_exclusion", "Total foreign earned income and housing exclusion", vec![
                F2555_FOREIGN_INCOME_EXCL, F2555_HOUSING_EXCLUSION,
            ]),
        ],
    }
}
