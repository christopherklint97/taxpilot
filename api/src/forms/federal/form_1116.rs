use crate::domain::builders::*;
use crate::domain::field::*;
use crate::domain::form::*;

pub fn form_1116() -> FormDef {
    FormDef {
        id: FORM_F1116.to_string(),
        name: "Form 1116 — Foreign Tax Credit".to_string(),
        jurisdiction: Jurisdiction::Federal,
        tax_years: vec![2024, 2025, 2026],
        question_group: "expat".to_string(),
        question_order: 4,
        fields: vec![
            // --- Part I: Taxable Income from Sources Outside the US ---
            enum_field(
                "category",
                "Foreign tax credit category",
                "What category of foreign income are you claiming the credit for?",
                vec!["general", "passive", "section_901j", "treaty_sourced"],
            ),
            string_input_field("foreign_country", "Country where tax was paid", "Which country did you pay foreign taxes to?"),
            input_field("foreign_source_income", "Gross foreign source income (not excluded by FEIE)", "What is your gross foreign source income NOT excluded by Form 2555?"),
            input_field("foreign_source_deductions", "Deductions allocated to foreign source income", "What deductions are definitely allocable to your foreign source income?"),
            input_field("foreign_tax_paid_income", "Foreign income taxes paid or accrued", "How much foreign income tax did you pay or accrue (converted to USD)?"),
            input_field("foreign_tax_paid_other", "Other foreign taxes paid", "How much in other qualifying foreign taxes did you pay (e.g., war profits tax)?"),
            enum_field("accrued_or_paid", "Taxes paid or accrued", "Are you claiming foreign taxes on a paid or accrued basis?", vec!["paid", "accrued"]),

            // --- Computed Fields ---

            // Line 7: Net foreign source taxable income
            max_zero_field("7", "Net foreign source taxable income", F1116_FOREIGN_SOURCE_INCOME, F1116_FOREIGN_SOURCE_DEDUCT),
            // Line 15: Total foreign taxes paid or accrued
            sum_field("15", "Total foreign taxes paid or accrued", vec![F1116_FOREIGN_TAX_PAID_INCOME, F1116_FOREIGN_TAX_PAID_OTHER]),
            // Line 20: US tax on worldwide income
            ref_field("20", "US tax liability", F1040_LINE_16),
            // Line 21: Foreign tax credit limitation
            {
                let deps = vec![
                    "form_1116:20".to_string(),
                    F1116_LINE_7.to_string(),
                    F1040_LINE_15.to_string(),
                ];
                FieldDef::new_computed("21", "Foreign tax credit limitation", deps, Box::new(|dv: &DepValues| {
                    let us_tax = dv.get("form_1116:20");
                    let foreign_source = dv.get(F1116_LINE_7);
                    let worldwide_taxable = dv.get(F1040_LINE_15);

                    if worldwide_taxable <= 0.0 || us_tax <= 0.0 {
                        return 0.0;
                    }

                    let ratio = (foreign_source / worldwide_taxable).min(1.0);
                    us_tax * ratio
                }))
            },
            // Line 22: Credit allowed
            {
                let deps = vec![F1116_LINE_15.to_string(), F1116_LINE_21.to_string()];
                FieldDef::new_computed("22", "Foreign tax credit allowed", deps, Box::new(|dv: &DepValues| {
                    let taxes_paid = dv.get(F1116_LINE_15);
                    let limitation = dv.get(F1116_LINE_21);
                    taxes_paid.min(limitation)
                }))
            },
            // Carryforward
            max_zero_field("carryforward", "Foreign tax credit carryforward", F1116_LINE_15, F1116_LINE_21),
        ],
    }
}
