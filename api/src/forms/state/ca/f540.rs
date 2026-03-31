use crate::domain::builders::*;
use crate::domain::field::*;
use crate::domain::form::*;
use crate::domain::taxmath;

/// California Form 540 -- California Resident Income Tax Return.
///
/// Simplified MVP covering a W-2 filer with standard deduction and no dependents.
/// FederalRef fields pull values from the federal return. Computed fields use
/// CA-specific brackets, exemption credits, and mental health tax.
pub fn form_ca_540() -> FormDef {
    FormDef {
        id: FORM_CA540.to_string(),
        name: "California Resident Income Tax Return".to_string(),
        jurisdiction: Jurisdiction::StateCA,
        tax_years: vec![2024, 2025, 2026],
        question_group: "ca".to_string(),
        question_order: 7,
        fields: vec![
            // Filing status -- references the federal filing status from Form 1040
            {
                let dep = F1040_FILING_STATUS.to_string();
                let dep_c = dep.clone();
                FieldDef {
                    line: "filing_status".to_string(),
                    field_type: FieldType::FederalRef,
                    value_type: FieldValueType::String,
                    label: "Filing status (from federal return)".to_string(),
                    prompt: String::new(),
                    depends_on: vec![dep],
                    options: Vec::new(),
                    compute: None,
                    compute_str: Some(Box::new(move |dv: &DepValues| {
                        dv.get_string(&dep_c)
                    })),
                }
            },

            // --- Income ---

            // Line 7: California wages (from W-2 Box 16)
            wildcard_sum_field(
                "7",
                "Wages, salaries, tips (CA)",
                W2_WILDCARD_STATE_WAGES,
            ),

            // Line 13: Federal AGI (from Form 1040 line 11)
            {
                let dep = F1040_LINE_11.to_string();
                let dep_c = dep.clone();
                FieldDef {
                    line: "13".to_string(),
                    field_type: FieldType::FederalRef,
                    value_type: FieldValueType::Numeric,
                    label: "Federal adjusted gross income".to_string(),
                    prompt: String::new(),
                    depends_on: vec![dep],
                    options: Vec::new(),
                    compute: Some(Box::new(move |dv: &DepValues| dv.get(&dep_c))),
                    compute_str: None,
                }
            },

            // Line 14: CA subtractions from Schedule CA
            ref_field("14", "California subtractions (from Schedule CA)", "ca_schedule_ca:37_col_b"),

            // Line 15: CA additions from Schedule CA
            ref_field("15", "California additions (from Schedule CA)", SCHED_CA_LINE_37_COL_C),

            // Line 17: California AGI = federal AGI - subtractions + additions
            {
                let deps = vec![
                    CA540_LINE_13.to_string(),
                    CA540_LINE_14.to_string(),
                    CA540_LINE_15.to_string(),
                ];
                let deps_c = deps.clone();
                FieldDef {
                    line: "17".to_string(),
                    field_type: FieldType::Computed,
                    value_type: FieldValueType::Numeric,
                    label: "California adjusted gross income".to_string(),
                    prompt: String::new(),
                    depends_on: deps,
                    options: Vec::new(),
                    compute: Some(Box::new(move |dv: &DepValues| {
                        dv.get(&deps_c[0]) - dv.get(&deps_c[1]) + dv.get(&deps_c[2])
                    })),
                    compute_str: None,
                }
            },

            // --- Deductions ---

            // Line 18: CA deduction -- larger of CA standard deduction or CA itemized
            {
                let deps = vec![
                    "ca_540:filing_status".to_string(),
                    "ca_schedule_ca:ca_itemized".to_string(),
                ];
                let deps_c = deps.clone();
                FieldDef {
                    line: "18".to_string(),
                    field_type: FieldType::Computed,
                    value_type: FieldValueType::Numeric,
                    label: "California deduction (standard or itemized)".to_string(),
                    prompt: String::new(),
                    depends_on: deps,
                    options: Vec::new(),
                    compute: Some(Box::new(move |dv: &DepValues| {
                        let fs_str = dv.get_string(&deps_c[0]);
                        let fs = taxmath::FilingStatus::from_str_code(&fs_str)
                            .unwrap_or(taxmath::FilingStatus::Single);
                        let standard = taxmath::get_standard_deduction(
                            dv.tax_year(),
                            taxmath::JurisdictionType::StateCA,
                            fs,
                        );
                        let ca_itemized = dv.get(&deps_c[1]);
                        standard.max(ca_itemized)
                    })),
                    compute_str: None,
                }
            },

            // Line 19: CA taxable income
            max_zero_field("19", "California taxable income", CA540_LINE_17, CA540_LINE_18),

            // --- Tax computation ---

            // Line 31: CA tax (bracket computation, excluding mental health tax)
            {
                let deps = vec![
                    CA540_LINE_19.to_string(),
                    "ca_540:filing_status".to_string(),
                ];
                let deps_c = deps.clone();
                FieldDef {
                    line: "31".to_string(),
                    field_type: FieldType::Computed,
                    value_type: FieldValueType::Numeric,
                    label: "California tax".to_string(),
                    prompt: String::new(),
                    depends_on: deps,
                    options: Vec::new(),
                    compute: Some(Box::new(move |dv: &DepValues| {
                        let fs_str = dv.get_string(&deps_c[1]);
                        let fs = taxmath::FilingStatus::from_str_code(&fs_str)
                            .unwrap_or(taxmath::FilingStatus::Single);
                        let taxable_income = dv.get(&deps_c[0]);
                        let brackets = match taxmath::get_brackets(
                            dv.tax_year(),
                            taxmath::JurisdictionType::StateCA,
                            fs,
                        ) {
                            Some(b) => b,
                            None => return 0.0,
                        };
                        taxmath::compute_bracket_tax(taxable_income, brackets)
                    })),
                    compute_str: None,
                }
            },

            // Line 32: CA exemption credits
            {
                let dep = "ca_540:filing_status".to_string();
                let dep_c = dep.clone();
                FieldDef {
                    line: "32".to_string(),
                    field_type: FieldType::Computed,
                    value_type: FieldValueType::Numeric,
                    label: "Exemption credits".to_string(),
                    prompt: String::new(),
                    depends_on: vec![dep],
                    options: Vec::new(),
                    compute: Some(Box::new(move |dv: &DepValues| {
                        let fs_str = dv.get_string(&dep_c);
                        let fs = taxmath::FilingStatus::from_str_code(&fs_str)
                            .unwrap_or(taxmath::FilingStatus::Single);
                        // MVP: 0 dependents
                        taxmath::get_ca_exemption_credit(dv.tax_year(), fs, 0)
                    })),
                    compute_str: None,
                }
            },

            // Line 35: Net tax after exemption credits
            max_zero_field("35", "Net tax (after exemption credits)", CA540_LINE_31, CA540_LINE_32),

            // Line 36: Mental Health Services Tax (1% on taxable income > $1M)
            {
                let dep = CA540_LINE_19.to_string();
                let dep_c = dep.clone();
                FieldDef {
                    line: "36".to_string(),
                    field_type: FieldType::Computed,
                    value_type: FieldValueType::Numeric,
                    label: "Mental Health Services Tax".to_string(),
                    prompt: String::new(),
                    depends_on: vec![dep],
                    options: Vec::new(),
                    compute: Some(Box::new(move |dv: &DepValues| {
                        taxmath::get_ca_mental_health_tax(dv.get(&dep_c))
                    })),
                    compute_str: None,
                }
            },

            // Line 40: Total CA tax (includes health coverage penalty from Form 3853)
            sum_field("40", "Total California tax", vec![CA540_LINE_35, CA540_LINE_36, "form_3853:7"]),

            // --- Payments ---

            // Line 71: CA tax withheld (from W-2 Box 17)
            wildcard_sum_field(
                "71",
                "California income tax withheld",
                W2_WILDCARD_STATE_TAX_WH,
            ),

            // Line 74: Total payments and credits (includes CalEITC from Form 3514)
            sum_field("74", "Total payments and credits", vec![CA540_LINE_71, "form_3514:7"]),

            // --- Refund or Amount Owed ---

            // Line 91: Overpayment (refund)
            max_zero_field("91", "Overpayment (refund)", CA540_LINE_74, CA540_LINE_40),

            // Line 93: Amount you owe
            max_zero_field("93", "Amount you owe", CA540_LINE_40, CA540_LINE_74),
        ],
    }
}
