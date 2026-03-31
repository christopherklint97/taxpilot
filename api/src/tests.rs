use std::collections::HashMap;

use crate::domain::field::{FieldType, FieldValueType};
use crate::domain::solver::DependencyGraph;
use crate::forms;

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/// Build default zero/empty inputs for every UserInput field in the registry,
/// plus zero placeholders for string fields (the solver checks numeric map).
fn build_default_inputs(
    registry: &crate::domain::registry::Registry,
) -> (HashMap<String, f64>, HashMap<String, String>) {
    let mut num = HashMap::new();
    let mut str_inputs = HashMap::new();

    for form in registry.all_forms() {
        for field in &form.fields {
            if field.field_type == FieldType::UserInput {
                let key = format!("{}:{}", form.id, field.line);
                // Every UserInput needs a numeric entry (solver checks this map).
                num.insert(key.clone(), 0.0);
                if field.value_type == FieldValueType::String {
                    str_inputs.insert(key, String::new());
                }
            }
        }
    }
    (num, str_inputs)
}

/// Solve with the given overrides on top of default-zero inputs.
fn solve_with(
    num_overrides: Vec<(&str, f64)>,
    str_overrides: Vec<(&str, &str)>,
    tax_year: i32,
) -> HashMap<String, f64> {
    let registry = forms::register_all_forms();
    let mut graph = DependencyGraph::new(&registry);
    graph.build().expect("graph build failed");

    let (mut num, mut str_inputs) = build_default_inputs(&registry);

    for (k, v) in num_overrides {
        num.insert(k.to_string(), v);
    }
    for (k, v) in str_overrides {
        str_inputs.insert(k.to_string(), v.to_string());
        // Also ensure the key exists in the numeric map (solver requirement).
        num.entry(k.to_string()).or_insert(0.0);
    }

    graph
        .solve(&num, &str_inputs, tax_year)
        .expect("solve failed")
}

/// Assert a value is within $1 of expected.
fn assert_approx(result: &HashMap<String, f64>, key: &str, expected: f64) {
    let actual = result.get(key).copied().unwrap_or(0.0);
    assert!(
        (actual - expected).abs() < 1.0,
        "key {}: expected ~{:.2}, got {:.2} (diff {:.2})",
        key,
        expected,
        actual,
        (actual - expected).abs()
    );
}

/// Assert a value equals expected exactly.
fn assert_exact(result: &HashMap<String, f64>, key: &str, expected: f64) {
    let actual = result.get(key).copied().unwrap_or(0.0);
    assert!(
        (actual - expected).abs() < 0.01,
        "key {}: expected {:.2}, got {:.2}",
        key,
        expected,
        actual
    );
}

// ---------------------------------------------------------------------------
// Test 1: Registry builds without cycles
// ---------------------------------------------------------------------------

#[test]
fn test_registry_builds_without_cycles() {
    let registry = forms::register_all_forms();
    let mut graph = DependencyGraph::new(&registry);
    assert!(graph.build().is_ok(), "dependency graph should build without cycles");
}

// ---------------------------------------------------------------------------
// Test 2: Single W-2 filer (simple)
// ---------------------------------------------------------------------------

#[test]
fn test_single_w2_filer() {
    let result = solve_with(
        vec![
            ("w2:1:wages", 75_000.0),
            ("w2:1:federal_tax_withheld", 9_500.0),
            ("w2:1:ss_wages", 75_000.0),
            ("w2:1:medicare_wages", 75_000.0),
            ("w2:1:state_wages", 75_000.0),
            ("w2:1:state_tax_withheld", 3_000.0),
        ],
        vec![
            ("1040:filing_status", "single"),
            ("form_3853:1", "yes"),   // full health coverage
            ("form_3853:3", "no"),
            ("schedule_b:7a", "no"),
            ("schedule_b:8", "no"),
            ("form_2555:employer_foreign", "no"),
            ("form_2555:self_employed_abroad", "no"),
            ("form_2555:qualifying_test", "bona_fide_residence"),
            ("form_2555:bfrt_full_year", "no"),
            ("form_1116:category", "general"),
            ("form_1116:accrued_or_paid", "paid"),
            ("form_8889:1", "self-only"),
            ("form_3514:6_yctc", "no"),
            ("form_8938:account_type", ""),
        ],
        2025,
    );

    // 1040:1a = 75000 (wages from W-2)
    assert_exact(&result, "1040:1a", 75_000.0);
    // 1040:11 = AGI = 75000
    assert_exact(&result, "1040:11", 75_000.0);
    // 1040:12 = standard deduction for single 2025 = 15000
    assert_exact(&result, "1040:12", 15_000.0);
    // 1040:15 = taxable income = 75000 - 15000 = 60000
    assert_exact(&result, "1040:15", 60_000.0);
    // 1040:16 = tax on 60000 single 2025
    // 10% on 0-11925 = 1192.50
    // 12% on 11925-48475 = 4386.00
    // 22% on 48475-60000 = 2535.50
    // Total = 8114.00
    assert_approx(&result, "1040:16", 8_114.0);
    // 1040:25a = federal tax withheld = 9500
    assert_exact(&result, "1040:25a", 9_500.0);
    // 1040:34 = refund (overpayment) = payments - total tax > 0
    let refund = result.get("1040:34").copied().unwrap_or(0.0);
    assert!(refund > 0.0, "expected positive refund, got {}", refund);
}

// ---------------------------------------------------------------------------
// Test 3: Self-employed filer (Schedule C + Schedule SE)
// ---------------------------------------------------------------------------

#[test]
fn test_self_employed_filer() {
    let result = solve_with(
        vec![
            ("1099nec:1:nonemployee_compensation", 80_000.0),
        ],
        vec![
            ("1040:filing_status", "single"),
            ("schedule_c:business_name", "Freelance LLC"),
            ("schedule_c:business_code", "541511"),
            ("form_3853:1", "yes"),
            ("form_3853:3", "no"),
            ("schedule_b:7a", "no"),
            ("schedule_b:8", "no"),
            ("form_2555:employer_foreign", "no"),
            ("form_2555:self_employed_abroad", "no"),
            ("form_2555:qualifying_test", "bona_fide_residence"),
            ("form_2555:bfrt_full_year", "no"),
            ("form_1116:category", "general"),
            ("form_1116:accrued_or_paid", "paid"),
            ("form_8889:1", "self-only"),
            ("form_3514:6_yctc", "no"),
            ("form_8938:account_type", ""),
        ],
        2025,
    );

    // Schedule C line 1 = 80000 (from 1099-NEC)
    assert_exact(&result, "schedule_c:1", 80_000.0);
    // Schedule C line 31 = net profit = 80000 (no expenses)
    assert_exact(&result, "schedule_c:31", 80_000.0);

    // Schedule SE line 2 = 80000 (from Schedule C)
    assert_exact(&result, "schedule_se:2", 80_000.0);
    // Schedule SE line 3 = 80000 * 0.9235 = 73880
    assert_approx(&result, "schedule_se:3", 73_880.0);
    // Schedule SE line 6 = SE tax (SS + Medicare on 73880)
    // SS tax: 73880 * 0.124 = 9161.12
    // Medicare tax: 73880 * 0.029 = 2142.52
    // Total SE tax: 11303.64
    let se_tax = result.get("schedule_se:6").copied().unwrap_or(0.0);
    assert_approx(&result, "schedule_se:6", 11_303.64);
    // Schedule SE line 7 = 50% of SE tax
    assert_approx(&result, "schedule_se:7", se_tax / 2.0);

    // Schedule 1 line 16 = SE deduction (50% of SE tax)
    assert_approx(&result, "schedule_1:16", se_tax / 2.0);
    // Schedule 1 line 3 = business income = 80000
    assert_exact(&result, "schedule_1:3", 80_000.0);
}

// ---------------------------------------------------------------------------
// Test 4: Capital gains (short-term and long-term)
// ---------------------------------------------------------------------------

#[test]
fn test_capital_gains() {
    let result = solve_with(
        vec![
            ("w2:1:wages", 50_000.0),
            ("w2:1:federal_tax_withheld", 5_000.0),
            ("w2:1:ss_wages", 50_000.0),
            ("w2:1:medicare_wages", 50_000.0),
            ("w2:1:state_wages", 50_000.0),
            // Short-term sale
            ("1099b:1:proceeds", 10_000.0),
            ("1099b:1:cost_basis", 8_000.0),
            // Long-term sale
            ("1099b:2:proceeds", 20_000.0),
            ("1099b:2:cost_basis", 12_000.0),
        ],
        vec![
            ("1040:filing_status", "single"),
            ("1099b:1:term", "short"),
            ("1099b:2:term", "long"),
            ("form_3853:1", "yes"),
            ("form_3853:3", "no"),
            ("schedule_b:7a", "no"),
            ("schedule_b:8", "no"),
            ("form_2555:employer_foreign", "no"),
            ("form_2555:self_employed_abroad", "no"),
            ("form_2555:qualifying_test", "bona_fide_residence"),
            ("form_2555:bfrt_full_year", "no"),
            ("form_1116:category", "general"),
            ("form_1116:accrued_or_paid", "paid"),
            ("form_8889:1", "self-only"),
            ("form_3514:6_yctc", "no"),
            ("form_8938:account_type", ""),
        ],
        2025,
    );

    // Form 8949 short-term
    assert_exact(&result, "form_8949:st_proceeds", 10_000.0);
    assert_exact(&result, "form_8949:st_basis", 8_000.0);
    assert_exact(&result, "form_8949:st_gain_loss", 2_000.0);

    // Form 8949 long-term
    assert_exact(&result, "form_8949:lt_proceeds", 20_000.0);
    assert_exact(&result, "form_8949:lt_basis", 12_000.0);
    assert_exact(&result, "form_8949:lt_gain_loss", 8_000.0);

    // Schedule D
    assert_exact(&result, "schedule_d:7", 2_000.0);  // net short-term
    assert_exact(&result, "schedule_d:15", 8_000.0);  // net long-term
    assert_exact(&result, "schedule_d:16", 10_000.0); // total net gain

    // Schedule 1 line 7 = capital gain flows to 1040
    assert_exact(&result, "schedule_1:7", 10_000.0);

    // AGI should include cap gains: 50000 + 10000 = 60000
    assert_exact(&result, "1040:11", 60_000.0);
}

// ---------------------------------------------------------------------------
// Test 5: Itemized deductions (Schedule A)
// ---------------------------------------------------------------------------

#[test]
fn test_itemized_deductions() {
    let result = solve_with(
        vec![
            ("w2:1:wages", 120_000.0),
            ("w2:1:federal_tax_withheld", 20_000.0),
            ("w2:1:ss_wages", 120_000.0),
            ("w2:1:medicare_wages", 120_000.0),
            ("w2:1:state_wages", 120_000.0),
            ("w2:1:state_tax_withheld", 8_000.0),
            // Schedule A inputs
            ("schedule_a:1", 2_000.0),     // medical expenses
            ("schedule_a:5a", 8_000.0),    // state/local income tax
            ("schedule_a:5b", 1_000.0),    // personal property tax
            ("schedule_a:5c", 6_000.0),    // real estate tax
            ("schedule_a:8a", 12_000.0),   // mortgage interest
            ("schedule_a:12", 3_000.0),    // charitable cash
        ],
        vec![
            ("1040:filing_status", "single"),
            ("form_3853:1", "yes"),
            ("form_3853:3", "no"),
            ("schedule_b:7a", "no"),
            ("schedule_b:8", "no"),
            ("form_2555:employer_foreign", "no"),
            ("form_2555:self_employed_abroad", "no"),
            ("form_2555:qualifying_test", "bona_fide_residence"),
            ("form_2555:bfrt_full_year", "no"),
            ("form_1116:category", "general"),
            ("form_1116:accrued_or_paid", "paid"),
            ("form_8889:1", "self-only"),
            ("form_3514:6_yctc", "no"),
            ("form_8938:account_type", ""),
        ],
        2025,
    );

    // SALT: 8000 + 1000 + 6000 = 15000, capped at 10000
    assert_exact(&result, "schedule_a:5d", 15_000.0);
    assert_exact(&result, "schedule_a:5e", 10_000.0);

    // Medical: 2000 - 7.5% of AGI (120000 * 0.075 = 9000) = 0 (below threshold)
    assert_exact(&result, "schedule_a:4", 0.0);

    // Total itemized = medical(0) + SALT(10000) + mortgage(12000) + charitable(3000) = 25000
    assert_exact(&result, "schedule_a:17", 25_000.0);

    // 1040:12 should pick itemized (25000) over standard (15000)
    assert_exact(&result, "1040:12", 25_000.0);

    // Taxable income = 120000 - 25000 = 95000
    assert_exact(&result, "1040:15", 95_000.0);
}

// ---------------------------------------------------------------------------
// Test 6: CA basic (federal + state)
// ---------------------------------------------------------------------------

#[test]
fn test_ca_basic() {
    let result = solve_with(
        vec![
            ("w2:1:wages", 75_000.0),
            ("w2:1:federal_tax_withheld", 9_500.0),
            ("w2:1:ss_wages", 75_000.0),
            ("w2:1:medicare_wages", 75_000.0),
            ("w2:1:state_wages", 75_000.0),
            ("w2:1:state_tax_withheld", 3_000.0),
        ],
        vec![
            ("1040:filing_status", "single"),
            ("form_3853:1", "yes"),
            ("form_3853:3", "no"),
            ("schedule_b:7a", "no"),
            ("schedule_b:8", "no"),
            ("form_2555:employer_foreign", "no"),
            ("form_2555:self_employed_abroad", "no"),
            ("form_2555:qualifying_test", "bona_fide_residence"),
            ("form_2555:bfrt_full_year", "no"),
            ("form_1116:category", "general"),
            ("form_1116:accrued_or_paid", "paid"),
            ("form_8889:1", "self-only"),
            ("form_3514:6_yctc", "no"),
            ("form_8938:account_type", ""),
        ],
        2025,
    );

    // CA 540 line 7 = state wages
    assert_exact(&result, "ca_540:7", 75_000.0);
    // CA 540 line 13 = federal AGI
    assert_exact(&result, "ca_540:13", 75_000.0);
    // CA 540 line 17 = CA AGI (same as federal since no adjustments)
    assert_exact(&result, "ca_540:17", 75_000.0);
    // CA 540 line 18 = CA standard deduction for single 2025 = 5706
    assert_exact(&result, "ca_540:18", 5_706.0);
    // CA 540 line 19 = CA taxable income = 75000 - 5706 = 69294
    assert_exact(&result, "ca_540:19", 69_294.0);
    // CA 540 line 31 = CA tax (from CA brackets, should be > 0)
    let ca_tax = result.get("ca_540:31").copied().unwrap_or(0.0);
    assert!(ca_tax > 0.0, "CA tax should be positive, got {}", ca_tax);
    // CA 540 line 32 = exemption credit for single 2025 = 144
    assert_exact(&result, "ca_540:32", 144.0);
    // CA 540 line 71 = CA withholding
    assert_exact(&result, "ca_540:71", 3_000.0);
}

// ---------------------------------------------------------------------------
// Test 7: Expat FEIE (Form 2555)
// ---------------------------------------------------------------------------

#[test]
fn test_expat_feie() {
    let result = solve_with(
        vec![
            ("1040:foreign_wages", 100_000.0),
            ("form_2555:foreign_earned_income", 100_000.0),
            ("form_2555:ppt_days_present", 365.0),
        ],
        vec![
            ("1040:filing_status", "single"),
            ("form_2555:foreign_country", "Sweden"),
            ("form_2555:foreign_address", "Stockholm"),
            ("form_2555:employer_name_2555", "Acme AB"),
            ("form_2555:employer_foreign", "yes"),
            ("form_2555:self_employed_abroad", "no"),
            ("form_2555:qualifying_test", "bona_fide_residence"),
            ("form_2555:bfrt_full_year", "yes"),
            ("form_2555:currency_code", "SEK"),
            ("form_3853:1", "yes"),
            ("form_3853:3", "no"),
            ("schedule_b:7a", "no"),
            ("schedule_b:8", "no"),
            ("form_1116:category", "general"),
            ("form_1116:accrued_or_paid", "paid"),
            ("form_8889:1", "self-only"),
            ("form_3514:6_yctc", "no"),
            ("form_8938:account_type", ""),
        ],
        2025,
    );

    // FEIE exclusion limit for 2025 = 130000
    assert_exact(&result, "form_2555:exclusion_limit", 130_000.0);
    // Qualifying days = 365 (full year bona fide)
    assert_exact(&result, "form_2555:qualifying_days", 365.0);
    // Foreign income exclusion = min(100000, 130000) = 100000
    assert_exact(&result, "form_2555:foreign_income_exclusion", 100_000.0);
    // Total exclusion = 100000 (no housing)
    assert_exact(&result, "form_2555:total_exclusion", 100_000.0);

    // 1040:1a = 100000 (foreign wages)
    assert_exact(&result, "1040:1a", 100_000.0);
    // Schedule 1 line 8d = -100000 (negated exclusion)
    assert_exact(&result, "schedule_1:8d", -100_000.0);
    // AGI = 100000 - 100000 = 0
    assert_exact(&result, "1040:11", 0.0);
    // Taxable income = 0 (AGI 0 minus deduction)
    assert_exact(&result, "1040:15", 0.0);
    // Tax should be 0 with zero taxable income
    assert_exact(&result, "1040:16", 0.0);
}

// ---------------------------------------------------------------------------
// Test 8: HSA contributions (Form 8889)
// ---------------------------------------------------------------------------

#[test]
fn test_hsa_contributions() {
    let result = solve_with(
        vec![
            ("w2:1:wages", 60_000.0),
            ("w2:1:federal_tax_withheld", 6_000.0),
            ("w2:1:ss_wages", 60_000.0),
            ("w2:1:medicare_wages", 60_000.0),
            ("w2:1:state_wages", 60_000.0),
            ("form_8889:2", 3_000.0),  // personal HSA contribution
            ("form_8889:3", 1_000.0),  // employer contribution
        ],
        vec![
            ("1040:filing_status", "single"),
            ("form_8889:1", "self-only"),
            ("form_3853:1", "yes"),
            ("form_3853:3", "no"),
            ("schedule_b:7a", "no"),
            ("schedule_b:8", "no"),
            ("form_2555:employer_foreign", "no"),
            ("form_2555:self_employed_abroad", "no"),
            ("form_2555:qualifying_test", "bona_fide_residence"),
            ("form_2555:bfrt_full_year", "no"),
            ("form_1116:category", "general"),
            ("form_1116:accrued_or_paid", "paid"),
            ("form_3514:6_yctc", "no"),
            ("form_8938:account_type", ""),
        ],
        2025,
    );

    // HSA limit for self-only 2025 = 4300
    assert_exact(&result, "form_8889:6", 4_300.0);
    // HSA deduction = personal contribution (3000) since total (3000+1000=4000) <= limit (4300)
    assert_exact(&result, "form_8889:9", 3_000.0);
    // Schedule 1 line 15 = HSA deduction
    assert_exact(&result, "schedule_1:15", 3_000.0);
    // AGI = 60000 - 3000 = 57000
    assert_exact(&result, "1040:11", 57_000.0);
}

// ---------------------------------------------------------------------------
// Test 9: Validate field definitions
// ---------------------------------------------------------------------------

#[test]
fn test_validate_field_defs() {
    let registry = forms::register_all_forms();
    let errors = registry.validate_field_defs();
    assert!(
        errors.is_empty(),
        "field definition validation errors: {:?}",
        errors
    );
}

// ---------------------------------------------------------------------------
// Test 10: Validate federal refs
// ---------------------------------------------------------------------------

#[test]
fn test_validate_federal_refs() {
    let registry = forms::register_all_forms();
    let errors = registry.validate_federal_refs();
    assert!(
        errors.is_empty(),
        "federal ref validation errors: {:?}",
        errors
    );
}
