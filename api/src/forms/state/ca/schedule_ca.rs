use crate::domain::builders::*;
use crate::domain::field::*;
use crate::domain::form::*;

/// Schedule CA (540) -- California Adjustments.
///
/// Adjusts federal income for California differences.
///
/// Part I, Section A: Income adjustments
///   - Line 2: Interest -- subtract U.S. obligation interest (CA-exempt);
///     add out-of-state muni bond interest (CA-taxable)
///   - Line 3: Dividends -- adjust for CA conformity differences
///   - Line 7: Capital gains -- CA generally conforms
///
/// Part I, Section B: Adjustments to income
///   - Line 15: HSA deduction add-back (CA does not conform to IRC sec. 223)
///   - Line 8d: Foreign earned income exclusion add-back (CA does not conform to FEIE)
///
/// Part II: Itemized deduction adjustments
///   - Line 5a: Remove state/local income tax deduction
///   - Line 5e: Recompute SALT without state income tax and without federal cap
pub fn form_schedule_ca() -> FormDef {
    FormDef {
        id: FORM_SCHEDULE_CA.to_string(),
        name: "Schedule CA (540) -- California Adjustments".to_string(),
        jurisdiction: Jurisdiction::StateCA,
        tax_years: vec![2024, 2025, 2026],
        question_group: "ca".to_string(),
        question_order: 7,
        fields: vec![
            // =================================================================
            // Part I, Section A: Income
            // =================================================================

            // Line 2, Column A: Federal taxable interest (from 1040 line 2b)
            {
                let dep = F1040_LINE_2B.to_string();
                let dep_c = dep.clone();
                FieldDef {
                    line: "2_col_a".to_string(),
                    field_type: FieldType::FederalRef,
                    value_type: FieldValueType::Numeric,
                    label: "Federal taxable interest".to_string(),
                    prompt: String::new(),
                    depends_on: vec![dep],
                    options: Vec::new(),
                    compute: Some(Box::new(move |dv: &DepValues| dv.get(&dep_c))),
                    compute_str: None,
                }
            },

            // Line 2, Column B: Interest subtractions (U.S. obligation interest
            // is exempt from CA tax)
            wildcard_sum_field(
                "2_col_b",
                "Interest subtractions (U.S. obligations exempt in CA)",
                F1099_INT_WILDCARD_US_BOND,
            ),

            // Line 2, Column C: Interest additions (out-of-state muni bond
            // interest is federally exempt but CA-taxable)
            zero_field("2_col_c", "Interest additions (non-CA muni bond interest)"),

            // Line 3, Column A: Federal ordinary dividends (from 1040 line 3b)
            {
                let dep = F1040_LINE_3B.to_string();
                let dep_c = dep.clone();
                FieldDef {
                    line: "3_col_a".to_string(),
                    field_type: FieldType::FederalRef,
                    value_type: FieldValueType::Numeric,
                    label: "Federal ordinary dividends".to_string(),
                    prompt: String::new(),
                    depends_on: vec![dep],
                    options: Vec::new(),
                    compute: Some(Box::new(move |dv: &DepValues| dv.get(&dep_c))),
                    compute_str: None,
                }
            },

            // Line 3, Column B: Dividend subtractions (CA generally conforms)
            zero_field("3_col_b", "Dividend subtractions"),

            // Line 3, Column C: Dividend additions (CA generally conforms)
            zero_field("3_col_c", "Dividend additions"),

            // Line 7, Column A: Federal capital gain (from 1040 line 7)
            {
                let dep = F1040_LINE_7.to_string();
                let dep_c = dep.clone();
                FieldDef {
                    line: "7_col_a".to_string(),
                    field_type: FieldType::FederalRef,
                    value_type: FieldValueType::Numeric,
                    label: "Federal capital gain or (loss)".to_string(),
                    prompt: String::new(),
                    depends_on: vec![dep],
                    options: Vec::new(),
                    compute: Some(Box::new(move |dv: &DepValues| dv.get(&dep_c))),
                    compute_str: None,
                }
            },

            // Line 7, Column B: Capital gain subtractions
            zero_field("7_col_b", "Capital gain subtractions"),

            // Line 7, Column C: Capital gain additions
            zero_field("7_col_c", "Capital gain additions"),

            // =================================================================
            // Part I, Section B: Adjustments to Income
            // =================================================================

            // Line 12: Business income -- CA generally conforms
            zero_field("12_col_b", "Business income subtractions"),
            zero_field("12_col_c", "Business income additions"),

            // Line 15, Column C: HSA deduction add-back
            // CA does not conform to federal HSA treatment (IRC sec. 223).
            ref_field(
                "15_col_c",
                "HSA deduction add-back (CA does not allow)",
                F8889_LINE_9,
            ),

            // Line 8d, Column C: Foreign earned income exclusion add-back
            // CA does NOT conform to the federal FEIE (IRC sec. 911).
            ref_field(
                "8d_col_c",
                "Foreign earned income exclusion add-back (CA does not allow FEIE)",
                F2555_TOTAL_EXCLUSION,
            ),

            // Line 8d, Column C (housing): Foreign housing deduction add-back
            ref_field(
                "8d_col_c_housing",
                "Foreign housing deduction add-back (CA does not allow)",
                F2555_HOUSING_DEDUCTION,
            ),

            // Line 16: Self-employment tax deduction -- CA conforms
            zero_field("16_col_b", "SE tax deduction subtractions"),

            // =================================================================
            // Part II: Itemized Deduction Adjustments
            // =================================================================

            // Line 5a_col_b: Subtract state/local income tax deduction
            ref_field(
                "5a_col_b",
                "State income tax subtraction (not deductible in CA)",
                SCHED_A_LINE_5A,
            ),

            // Line 5e_col_b: Subtract the federal SALT amount
            ref_field(
                "5e_col_b",
                "Federal SALT subtraction (CA recomputes without cap)",
                SCHED_A_LINE_5E,
            ),

            // Line 5e_col_c: Add back CA-allowed SALT (property taxes only, no cap)
            sum_field(
                "5e_col_c",
                "CA SALT addition (property taxes only, no cap)",
                vec![SCHED_A_LINE_5B, SCHED_A_LINE_5C],
            ),

            // CA itemized deductions total adjustment
            ref_field(
                "itemized_sub",
                "Total itemized deduction subtractions",
                "ca_schedule_ca:5e_col_b",
            ),
            ref_field(
                "itemized_add",
                "Total itemized deduction additions",
                "ca_schedule_ca:5e_col_c",
            ),

            // CA itemized deductions = federal itemized - subtractions + additions
            {
                let deps = vec![
                    SCHED_A_LINE_17.to_string(),
                    "ca_schedule_ca:itemized_sub".to_string(),
                    "ca_schedule_ca:itemized_add".to_string(),
                ];
                let deps_c = deps.clone();
                FieldDef {
                    line: "ca_itemized".to_string(),
                    field_type: FieldType::Computed,
                    value_type: FieldValueType::Numeric,
                    label: "California itemized deductions".to_string(),
                    prompt: String::new(),
                    depends_on: deps,
                    options: Vec::new(),
                    compute: Some(Box::new(move |dv: &DepValues| {
                        let federal = dv.get(&deps_c[0]);
                        let sub = dv.get(&deps_c[1]);
                        let add = dv.get(&deps_c[2]);
                        let result = federal - sub + add;
                        if result < 0.0 { 0.0 } else { result }
                    })),
                    compute_str: None,
                }
            },

            // =================================================================
            // Totals
            // =================================================================

            // Line 37, Column A: Federal amounts (mirrors federal AGI)
            {
                let dep = F1040_LINE_11.to_string();
                let dep_c = dep.clone();
                FieldDef {
                    line: "37_col_a".to_string(),
                    field_type: FieldType::FederalRef,
                    value_type: FieldValueType::Numeric,
                    label: "Federal amounts (from Form 1040)".to_string(),
                    prompt: String::new(),
                    depends_on: vec![dep],
                    options: Vec::new(),
                    compute: Some(Box::new(move |dv: &DepValues| dv.get(&dep_c))),
                    compute_str: None,
                }
            },

            // Line 37, Column B: Total subtractions
            sum_field(
                "37_col_b",
                "Subtractions (Column B)",
                vec![
                    "ca_schedule_ca:2_col_b",
                    "ca_schedule_ca:3_col_b",
                    "ca_schedule_ca:7_col_b",
                    "ca_schedule_ca:12_col_b",
                    "ca_schedule_ca:16_col_b",
                ],
            ),

            // Line 37, Column C: Total additions
            sum_field(
                "37_col_c",
                "Additions (Column C)",
                vec![
                    "ca_schedule_ca:2_col_c",
                    "ca_schedule_ca:3_col_c",
                    "ca_schedule_ca:7_col_c",
                    SCHED_CA_LINE_8D_COL_C,
                    SCHED_CA_LINE_8D_COL_C_HOUSING,
                    "ca_schedule_ca:12_col_c",
                    "ca_schedule_ca:15_col_c",
                ],
            ),
        ],
    }
}
