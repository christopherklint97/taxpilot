use crate::domain::builders::*;
use crate::domain::field::*;
use crate::domain::form::*;

/// California Form 3514 -- California Earned Income Tax Credit (CalEITC).
///
/// Computes the CalEITC and the Young Child Tax Credit (YCTC) for
/// low-income California filers.
pub fn form_3514() -> FormDef {
    FormDef {
        id: FORM_F3514.to_string(),
        name: "Form 3514 -- California Earned Income Tax Credit".to_string(),
        jurisdiction: Jurisdiction::StateCA,
        tax_years: vec![2024, 2025, 2026],
        question_group: "ca".to_string(),
        question_order: 7,
        fields: vec![
            // Line 1: Earned income (wages + positive self-employment income)
            {
                let deps = vec![
                    F1040_LINE_1A.to_string(),
                    SCHED_C_LINE_31.to_string(),
                ];
                let deps_c = deps.clone();
                FieldDef {
                    line: "1".to_string(),
                    field_type: FieldType::FederalRef,
                    value_type: FieldValueType::Numeric,
                    label: "Earned income".to_string(),
                    prompt: String::new(),
                    depends_on: deps,
                    options: Vec::new(),
                    compute: Some(Box::new(move |dv: &DepValues| {
                        let wages = dv.get(&deps_c[0]);
                        let se = dv.get(&deps_c[1]).max(0.0);
                        wages + se
                    })),
                    compute_str: None,
                }
            },

            // Line 2: Filing status factor (1 = single/HOH, 2 = MFJ)
            {
                let dep = F1040_FILING_STATUS.to_string();
                let dep_c = dep.clone();
                FieldDef {
                    line: "2".to_string(),
                    field_type: FieldType::Computed,
                    value_type: FieldValueType::Numeric,
                    label: "Filing status factor".to_string(),
                    prompt: String::new(),
                    depends_on: vec![dep],
                    options: Vec::new(),
                    compute: Some(Box::new(move |dv: &DepValues| {
                        let fs = dv.get_string(&dep_c);
                        if fs == "mfj" { 2.0 } else { 1.0 }
                    })),
                    compute_str: None,
                }
            },

            // Line 3: Number of qualifying children (0-3+)
            input_field(
                "3",
                "Number of qualifying children for CalEITC",
                "How many qualifying children do you have for the California EITC?",
            ),

            // Line 4: Income limit check (earned income must be <= $30,950)
            {
                let dep = "form_3514:1".to_string();
                let dep_c = dep.clone();
                FieldDef {
                    line: "4".to_string(),
                    field_type: FieldType::Computed,
                    value_type: FieldValueType::Numeric,
                    label: "CalEITC income limit check".to_string(),
                    prompt: String::new(),
                    depends_on: vec![dep],
                    options: Vec::new(),
                    compute: Some(Box::new(move |dv: &DepValues| {
                        if dv.get(&dep_c) <= 30950.0 { 1.0 } else { 0.0 }
                    })),
                    compute_str: None,
                }
            },

            // Line 5: CalEITC amount based on income and children
            {
                let deps = vec![
                    "form_3514:1".to_string(),
                    "form_3514:3".to_string(),
                    "form_3514:4".to_string(),
                ];
                let deps_c = deps.clone();
                FieldDef {
                    line: "5".to_string(),
                    field_type: FieldType::Computed,
                    value_type: FieldValueType::Numeric,
                    label: "CalEITC amount".to_string(),
                    prompt: String::new(),
                    depends_on: deps,
                    options: Vec::new(),
                    compute: Some(Box::new(move |dv: &DepValues| {
                        if dv.get(&deps_c[2]) == 0.0 {
                            return 0.0; // over income limit
                        }
                        let earned = dv.get(&deps_c[0]);
                        let children = dv.get(&deps_c[1]) as i32;
                        compute_cal_eitc(earned, children)
                    })),
                    compute_str: None,
                }
            },

            // Line 6_yctc: Whether taxpayer has a qualifying child under age 6
            enum_field(
                "6_yctc",
                "Qualifying child under age 6",
                "Do you have a qualifying child under age 6?",
                vec!["yes", "no"],
            ),

            // Line 6: Young Child Tax Credit ($1,117 if child under 6)
            {
                let deps = vec![
                    F3514_LINE_6_YCTC.to_string(),
                    "form_3514:4".to_string(),
                ];
                let deps_c = deps.clone();
                FieldDef {
                    line: "6".to_string(),
                    field_type: FieldType::Computed,
                    value_type: FieldValueType::Numeric,
                    label: "Young Child Tax Credit".to_string(),
                    prompt: String::new(),
                    depends_on: deps,
                    options: Vec::new(),
                    compute: Some(Box::new(move |dv: &DepValues| {
                        if dv.get(&deps_c[1]) == 0.0 {
                            return 0.0; // over income limit
                        }
                        if dv.get_string(&deps_c[0]) == "yes" {
                            1117.0
                        } else {
                            0.0
                        }
                    })),
                    compute_str: None,
                }
            },

            // Line 7: Total CalEITC = Line 5 + Line 6
            sum_field("7", "Total CalEITC", vec!["form_3514:5", "form_3514:6"]),
        ],
    }
}

/// Computes the CalEITC credit based on earned income and number of qualifying
/// children. Uses simplified 2025 phase-in/phase-out schedule.
fn compute_cal_eitc(earned: f64, children: i32) -> f64 {
    let children = children.clamp(0, 3) as usize;

    struct EitcParams {
        max_credit: f64,
        phase_out_start: f64,
    }

    let params = [
        EitcParams { max_credit: 275.0,  phase_out_start: 7500.0 },   // 0 children
        EitcParams { max_credit: 1843.0, phase_out_start: 11000.0 },  // 1 child
        EitcParams { max_credit: 3037.0, phase_out_start: 15500.0 },  // 2 children
        EitcParams { max_credit: 3417.0, phase_out_start: 15500.0 },  // 3+ children
    ];

    let p = &params[children];
    let phase_out_end = 30950.0;

    if earned <= 0.0 {
        return 0.0;
    }

    if earned <= p.phase_out_start {
        // Phase-in range: credit grows proportionally up to max
        return (p.max_credit * earned / p.phase_out_start * 100.0).round() / 100.0;
    }

    if earned > phase_out_end {
        return 0.0;
    }

    // Phase-out range: credit reduces linearly from max to 0
    let remaining = phase_out_end - earned;
    let phase_out_range = phase_out_end - p.phase_out_start;
    let credit = p.max_credit * remaining / phase_out_range;
    (credit * 100.0).round() / 100.0
}
