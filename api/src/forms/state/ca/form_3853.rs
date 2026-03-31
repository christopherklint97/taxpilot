use crate::domain::builders::*;
use crate::domain::field::*;
use crate::domain::form::*;

/// California Form 3853 -- Health Coverage Exemptions and Individual Shared
/// Responsibility Penalty.
///
/// Computes the penalty for not maintaining qualifying health coverage under
/// California's individual mandate.
pub fn form_3853() -> FormDef {
    FormDef {
        id: FORM_F3853.to_string(),
        name: "Form 3853 -- Health Coverage Exemptions and Individual Shared Responsibility Penalty".to_string(),
        jurisdiction: Jurisdiction::StateCA,
        tax_years: vec![2024, 2025, 2026],
        question_group: "ca".to_string(),
        question_order: 7,
        fields: vec![
            // Line 1: Full-year coverage (yes/no)
            enum_field(
                "1",
                "Full-year qualifying health coverage",
                "Did you have qualifying health coverage for all 12 months of 2025?",
                vec!["yes", "no"],
            ),

            // Line 2: Months without coverage (0-12)
            input_field(
                "2",
                "Months without qualifying health coverage",
                "How many months were you without qualifying health coverage?",
            ),

            // Line 3: Exemption from coverage requirement (yes/no)
            enum_field(
                "3",
                "Exemption from health coverage requirement",
                "Did you have an exemption from the health coverage requirement?",
                vec!["yes", "no"],
            ),

            // Line 4: Applicable household income (CA AGI from Form 540)
            ref_field("4", "Applicable household income", CA540_LINE_17),

            // Line 5: Penalty per month
            {
                let deps = vec![
                    F3853_LINE_1.to_string(),
                    F3853_LINE_3.to_string(),
                    "form_3853:4".to_string(),
                ];
                let deps_c = deps.clone();
                FieldDef {
                    line: "5".to_string(),
                    field_type: FieldType::Computed,
                    value_type: FieldValueType::Numeric,
                    label: "Penalty per month".to_string(),
                    prompt: String::new(),
                    depends_on: deps,
                    options: Vec::new(),
                    compute: Some(Box::new(move |dv: &DepValues| {
                        if dv.get_string(&deps_c[0]) == "yes" {
                            return 0.0; // full coverage, no penalty
                        }
                        if dv.get_string(&deps_c[1]) == "yes" {
                            return 0.0; // exempt from requirement
                        }

                        let ca_agi = dv.get(&deps_c[2]);

                        // Flat penalty: $900/month (2025 state avg bronze plan)
                        let flat_per_month: f64 = 900.0;

                        // Income-based penalty: 2.5% of (CA AGI - filing threshold) / 12
                        let filing_threshold = 21135.0;
                        let income_base = (ca_agi - filing_threshold).max(0.0);
                        let income_per_month = 0.025 * income_base / 12.0;

                        let penalty = flat_per_month.max(income_per_month);
                        (penalty * 100.0).round() / 100.0
                    })),
                    compute_str: None,
                }
            },

            // Line 6: Total penalty = months * per_month_penalty
            {
                let deps = vec![
                    F3853_LINE_1.to_string(),
                    F3853_LINE_2.to_string(),
                    F3853_LINE_3.to_string(),
                    "form_3853:5".to_string(),
                ];
                let deps_c = deps.clone();
                FieldDef {
                    line: "6".to_string(),
                    field_type: FieldType::Computed,
                    value_type: FieldValueType::Numeric,
                    label: "Total penalty".to_string(),
                    prompt: String::new(),
                    depends_on: deps,
                    options: Vec::new(),
                    compute: Some(Box::new(move |dv: &DepValues| {
                        if dv.get_string(&deps_c[0]) == "yes" {
                            return 0.0;
                        }
                        if dv.get_string(&deps_c[2]) == "yes" {
                            return 0.0;
                        }
                        let months = dv.get(&deps_c[1]);
                        let per_month = dv.get(&deps_c[3]);
                        let total = months * per_month;
                        // Cap at $10,800/year (12 * $900)
                        total.min(10800.0)
                    })),
                    compute_str: None,
                }
            },

            // Line 7: Penalty amount that carries to Form 540
            ref_field("7", "Penalty to Form 540", "form_3853:6"),
        ],
    }
}
