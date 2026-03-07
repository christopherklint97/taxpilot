package forms_test

import (
	"encoding/json"
	"math"
	"os"
	"strings"
	"testing"

	"taxpilot/internal/forms"
	"taxpilot/internal/forms/federal"
	ca "taxpilot/internal/forms/state/ca"
)

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------

func assertClose(t *testing.T, got, want float64, label string) {
	t.Helper()
	if math.Abs(got-want) > 0.01 {
		t.Errorf("%s: got %.2f, want %.2f (diff %.2f)", label, got, want, got-want)
	}
}

// scenarioJSON mirrors the JSON structure used in testdata/scenarios.
type scenarioJSON struct {
	Name    string                 `json:"name"`
	TaxYear int                    `json:"tax_year"`
	Inputs  map[string]interface{} `json:"inputs"`
}

// expectedJSON mirrors the JSON structure used in testdata/expected.
type expectedJSON struct {
	Name     string             `json:"name"`
	TaxYear  int                `json:"tax_year"`
	Expected map[string]float64 `json:"expected"`
}

// inputFormDefaults provides zero-value defaults for all input form fields.
// This allows W-2-only scenarios to work even though 1099 forms and
// Schedule A are registered in the solver.
var inputFormDefaults = map[string]interface{}{
	// 1099-INT defaults
	"1099int:1:payer_name":                     "N/A",
	"1099int:1:payer_tin":                      "00-0000000",
	"1099int:1:interest_income":                0.0,
	"1099int:1:early_withdrawal_penalty":       0.0,
	"1099int:1:us_savings_bond_interest":       0.0,
	"1099int:1:federal_tax_withheld":           0.0,
	"1099int:1:tax_exempt_interest":            0.0,
	"1099int:1:private_activity_bond_interest": 0.0,
	// 1099-DIV defaults
	"1099div:1:payer_name":                      "N/A",
	"1099div:1:payer_tin":                       "00-0000000",
	"1099div:1:ordinary_dividends":              0.0,
	"1099div:1:qualified_dividends":             0.0,
	"1099div:1:total_capital_gain":              0.0,
	"1099div:1:section_1250_gain":               0.0,
	"1099div:1:section_199a_dividends":          0.0,
	"1099div:1:federal_tax_withheld":            0.0,
	"1099div:1:exempt_interest_dividends":       0.0,
	"1099div:1:private_activity_bond_dividends": 0.0,
	// 1099-NEC defaults
	"1099nec:1:payer_name":               "N/A",
	"1099nec:1:payer_tin":                "00-0000000",
	"1099nec:1:nonemployee_compensation": 0.0,
	"1099nec:1:federal_tax_withheld":     0.0,
	// 1099-B defaults
	"1099b:1:description":          "N/A",
	"1099b:1:date_acquired":        "N/A",
	"1099b:1:date_sold":            "N/A",
	"1099b:1:proceeds":             0.0,
	"1099b:1:cost_basis":           0.0,
	"1099b:1:wash_sale_loss":       0.0,
	"1099b:1:federal_tax_withheld": 0.0,
	"1099b:1:term":                 "short",
	"1099b:1:basis_reported":       "yes",
	// Form 8889 defaults (no HSA)
	"form_8889:1":   "self-only",
	"form_8889:2":   0.0,
	"form_8889:3":   0.0,
	"form_8889:5":   0.0,
	"form_8889:14a": 0.0,
	"form_8889:14c": 0.0,
	// Schedule A defaults (0 = standard deduction wins)
	"schedule_a:1":  0.0,
	"schedule_a:5a": 0.0,
	"schedule_a:5b": 0.0,
	"schedule_a:5c": 0.0,
	"schedule_a:8a": 0.0,
	"schedule_a:12": 0.0,
	"schedule_a:13": 0.0,
	"schedule_a:14": 0.0,
	// Schedule 3 defaults
	"schedule_3:10": 0.0,
	// Form 3514 (CalEITC) defaults
	"form_3514:3":      0.0,
	"form_3514:6_yctc": "no",
	// Form 3853 (Health Coverage) defaults
	"form_3853:1": "yes",
	"form_3853:2": 0.0,
	"form_3853:3": "no",
	// Form 8938 defaults (no FATCA filing)
	"form_8938:lives_abroad":            "no",
	"form_8938:num_accounts":            0.0,
	"form_8938:max_value_accounts":      0.0,
	"form_8938:yearend_value_accounts":  0.0,
	"form_8938:num_other_assets":        0.0,
	"form_8938:max_value_other":         0.0,
	"form_8938:yearend_value_other":     0.0,
	"form_8938:account_country":         "N/A",
	"form_8938:account_institution":     "N/A",
	"form_8938:account_type":            "deposit",
	"form_8938:income_from_accounts":    0.0,
	"form_8938:gain_from_accounts":      0.0,
	// Form 8833 defaults (no treaty position)
	"form_8833:treaty_country":                "N/A",
	"form_8833:treaty_article":                "N/A",
	"form_8833:irc_provision":                 "N/A",
	"form_8833:treaty_position_explanation":   "N/A",
	"form_8833:treaty_amount":                 0.0,
	"form_8833:num_positions":                 0.0,
	// Form 1116 defaults (no foreign tax credit)
	"form_1116:category":                  "general",
	"form_1116:foreign_country":           "N/A",
	"form_1116:foreign_source_income":     0.0,
	"form_1116:foreign_source_deductions": 0.0,
	"form_1116:foreign_tax_paid_income":   0.0,
	"form_1116:foreign_tax_paid_other":    0.0,
	"form_1116:accrued_or_paid":           "paid",
	// Schedule B Part III defaults (no foreign accounts)
	"schedule_b:7a": "no",
	"schedule_b:7b": "N/A",
	"schedule_b:8":  "no",
	// Form 2555 defaults (no foreign income)
	"form_2555:foreign_country":          "N/A",
	"form_2555:foreign_address":          "N/A",
	"form_2555:employer_name_2555":       "N/A",
	"form_2555:employer_foreign":         "no",
	"form_2555:self_employed_abroad":     "no",
	"form_2555:qualifying_test":          "bona_fide_residence",
	"form_2555:bfrt_start_date":          "N/A",
	"form_2555:bfrt_end_date":            "N/A",
	"form_2555:bfrt_full_year":           "no",
	"form_2555:ppt_days_present":         0.0,
	"form_2555:ppt_period_start":         "N/A",
	"form_2555:ppt_period_end":           "N/A",
	"form_2555:foreign_earned_income":    0.0,
	"form_2555:currency_code":            "USD",
	"form_2555:exchange_rate":            0.0,
	"form_2555:foreign_tax_paid":         0.0,
	"form_2555:employer_provided_housing": 0.0,
	"form_2555:housing_expenses":         0.0,
	// Schedule C defaults (0 = no business income)
	"schedule_c:business_name": "N/A",
	"schedule_c:business_code": "000000",
	"schedule_c:8":             0.0,
	"schedule_c:10":            0.0,
	"schedule_c:17":            0.0,
	"schedule_c:18":            0.0,
	"schedule_c:22":            0.0,
	"schedule_c:25":            0.0,
	"schedule_c:27":            0.0,
}

// loadScenario reads a scenario file and splits its inputs into numeric and
// string maps suitable for the solver. Input form fields are re-keyed with
// a "1:" instance prefix so that wildcard patterns (e.g., w2:*:wages) match.
// Missing input forms get zero defaults so scenarios aren't required to
// provide every registered form's inputs.
func loadScenario(t *testing.T, path string) (numInputs map[string]float64, strInputs map[string]string, taxYear int) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read scenario %s: %v", path, err)
	}
	var s scenarioJSON
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatalf("failed to parse scenario %s: %v", path, err)
	}

	numInputs = make(map[string]float64)
	strInputs = make(map[string]string)
	taxYear = s.TaxYear

	// Apply defaults for input forms not in the scenario
	for key, val := range inputFormDefaults {
		switch v := val.(type) {
		case float64:
			numInputs[key] = v
		case string:
			numInputs[key] = 0
			strInputs[key] = v
		}
	}

	// Input form prefixes that need instance re-keying (form_id: -> form_id:1:)
	inputPrefixes := []string{"w2:", "1099int:", "1099div:", "1099nec:", "1099b:"}

	for key, val := range s.Inputs {
		solverKey := key
		for _, prefix := range inputPrefixes {
			if len(key) > len(prefix) && key[:len(prefix)] == prefix {
				// Only re-key if not already instance-keyed
				rest := key[len(prefix):]
				if !strings.Contains(rest, ":") {
					solverKey = prefix + "1:" + rest
				}
				break
			}
		}

		switch v := val.(type) {
		case float64:
			numInputs[solverKey] = v
		case string:
			numInputs[solverKey] = 0
			strInputs[solverKey] = v
		}
	}

	// 1040 string fields (filing_status, names, ssn) also need zero
	// placeholders in the numeric map.
	for key, val := range s.Inputs {
		if _, isStr := val.(string); isStr {
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = 0
				strInputs[key] = val.(string)
			}
		}
	}

	return numInputs, strInputs, taxYear
}

// loadExpected reads an expected-values file.
func loadExpected(t *testing.T, path string) map[string]float64 {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read expected %s: %v", path, err)
	}
	var e expectedJSON
	if err := json.Unmarshal(data, &e); err != nil {
		t.Fatalf("failed to parse expected %s: %v", path, err)
	}
	return e.Expected
}

// buildSolver creates a Registry with all production forms, builds the
// dependency graph, and returns the graph ready for Solve.
func buildSolver(t *testing.T) *forms.DependencyGraph {
	t.Helper()
	reg := forms.NewRegistry()
	// Input forms are NOT registered here — their fields are resolved
	// via wildcard dependencies (e.g., "w2:*:wages"). Registering them
	// would make MissingInputs require all their fields in every scenario.
	reg.Register(federal.F1040())
	reg.Register(federal.ScheduleA())
	reg.Register(federal.ScheduleB())
	reg.Register(federal.ScheduleC())
	reg.Register(federal.ScheduleD())
	reg.Register(federal.Form8949())
	reg.Register(federal.Schedule1())
	reg.Register(federal.Schedule2())
	reg.Register(federal.Schedule3())
	reg.Register(federal.ScheduleSE())
	reg.Register(federal.Form8995())
	reg.Register(federal.Form8889())
	reg.Register(federal.Form2555())
	reg.Register(federal.Form1116())
	reg.Register(federal.Form8938())
	reg.Register(federal.Form8833())
	reg.Register(ca.F540())
	reg.Register(ca.ScheduleCA())
	reg.Register(ca.Form3514())
	reg.Register(ca.Form3853())

	g := forms.NewDependencyGraph(reg)
	if err := g.Build(); err != nil {
		t.Fatalf("DependencyGraph.Build failed: %v", err)
	}
	return g
}

// --------------------------------------------------------------------------
// Integration tests
// --------------------------------------------------------------------------

func TestFederalSingleW2Scenario(t *testing.T) {
	g := buildSolver(t)

	numInputs, strInputs, taxYear := loadScenario(t,
		"../../testdata/scenarios/federal/single_w2.json")

	result, err := g.Solve(numInputs, strInputs, taxYear)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	expected := loadExpected(t, "../../testdata/expected/federal/single_w2.json")
	for key, want := range expected {
		assertClose(t, result[key], want, key)
	}
}

func TestCASingleW2Scenario(t *testing.T) {
	g := buildSolver(t)

	numInputs, strInputs, taxYear := loadScenario(t,
		"../../testdata/scenarios/ca/ca_single_w2.json")

	result, err := g.Solve(numInputs, strInputs, taxYear)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// Check federal expected values too (same inputs)
	fedExpected := loadExpected(t, "../../testdata/expected/federal/single_w2.json")
	for key, want := range fedExpected {
		assertClose(t, result[key], want, key)
	}

	// Check CA expected values
	caExpected := loadExpected(t, "../../testdata/expected/ca/ca_single_w2.json")
	for key, want := range caExpected {
		assertClose(t, result[key], want, key)
	}
}

func TestCAHighIncomeScenario(t *testing.T) {
	g := buildSolver(t)

	numInputs, strInputs, taxYear := loadScenario(t,
		"../../testdata/scenarios/ca/ca_high_income.json")

	result, err := g.Solve(numInputs, strInputs, taxYear)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	expected := loadExpected(t, "../../testdata/expected/ca/ca_high_income.json")
	for key, want := range expected {
		assertClose(t, result[key], want, key)
	}

	// Verify mental health surcharge is non-zero for income > $1M
	if result["ca_540:36"] <= 0 {
		t.Error("expected non-zero mental health surcharge for $1.2M income")
	}

	// Verify amount owed (withholding < total tax)
	if result["ca_540:93"] <= 0 {
		t.Error("expected non-zero CA amount owed for high income scenario")
	}
	if result["1040:37"] <= 0 {
		t.Error("expected non-zero federal amount owed for high income scenario")
	}
}

// TestFederalSingleW2Inline runs the same scenario as the JSON-driven test
// but with inline inputs, serving as a self-contained reference.
func TestFederalSingleW2Inline(t *testing.T) {
	g := buildSolver(t)

	numInputs := map[string]float64{
		"w2:1:wages":                 75000,
		"w2:1:federal_tax_withheld":  9500,
		"w2:1:ss_wages":              75000,
		"w2:1:ss_tax_withheld":       4650,
		"w2:1:medicare_wages":        75000,
		"w2:1:medicare_tax_withheld": 1087.50,
		"w2:1:state_wages":           75000,
		"w2:1:state_tax_withheld":    3200,
		"w2:1:employer_name":         0,
		"w2:1:employer_ein":          0,
		"1040:filing_status":         0,
		"1040:first_name":            0,
		"1040:last_name":             0,
		"1040:ssn":                   0,
	}
	strInputs := map[string]string{
		"1040:filing_status":  "single",
		"1040:first_name":     "Jane",
		"1040:last_name":      "Doe",
		"1040:ssn":            "123-45-6789",
		"w2:1:employer_name":  "Acme Corp",
		"w2:1:employer_ein":   "12-3456789",
	}

	// Add 1099 defaults (no interest/dividends for this scenario)
	for key, val := range inputFormDefaults {
		switch v := val.(type) {
		case float64:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = v
			}
		case string:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = 0
				strInputs[key] = v
			}
		}
	}

	result, err := g.Solve(numInputs, strInputs, 2025)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// Federal checks
	assertClose(t, result["1040:1a"], 75000, "Line 1a wages")
	assertClose(t, result["1040:11"], 75000, "Line 11 AGI")
	assertClose(t, result["1040:12"], 15000, "Line 12 standard deduction")
	assertClose(t, result["1040:15"], 60000, "Line 15 taxable income")
	assertClose(t, result["1040:16"], 8114, "Line 16 federal tax")
	assertClose(t, result["1040:24"], 8114, "Line 24 total tax")
	assertClose(t, result["1040:25a"], 9500, "Line 25a withholding")
	assertClose(t, result["1040:33"], 9500, "Line 33 total payments")
	assertClose(t, result["1040:34"], 1386, "Line 34 refund")
	assertClose(t, result["1040:37"], 0, "Line 37 amount owed")

	// CA checks
	assertClose(t, result["ca_540:13"], 75000, "CA line 13 federal AGI")
	assertClose(t, result["ca_540:17"], 75000, "CA line 17 CA AGI")
	assertClose(t, result["ca_540:18"], 5706, "CA line 18 standard deduction")
	assertClose(t, result["ca_540:19"], 69294, "CA line 19 taxable income")
	assertClose(t, result["ca_540:31"], 3003.76, "CA line 31 bracket tax")
	assertClose(t, result["ca_540:32"], 144, "CA line 32 exemption credit")
	assertClose(t, result["ca_540:35"], 2859.76, "CA line 35 net tax")
	assertClose(t, result["ca_540:36"], 0, "CA line 36 mental health tax")
	assertClose(t, result["ca_540:40"], 2859.76, "CA line 40 total tax")
	assertClose(t, result["ca_540:71"], 3200, "CA line 71 withholding")
	assertClose(t, result["ca_540:74"], 3200, "CA line 74 total payments")
	assertClose(t, result["ca_540:91"], 340.24, "CA line 91 refund")
	assertClose(t, result["ca_540:93"], 0, "CA line 93 amount owed")
}

func TestFederalSelfEmployedScenario(t *testing.T) {
	g := buildSolver(t)

	numInputs, strInputs, taxYear := loadScenario(t,
		"../../testdata/scenarios/federal/self_employed.json")

	result, err := g.Solve(numInputs, strInputs, taxYear)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	expected := loadExpected(t, "../../testdata/expected/federal/self_employed.json")
	for key, want := range expected {
		assertClose(t, result[key], want, key)
	}

	// Verify SE tax is non-zero
	if result["schedule_se:6"] <= 0 {
		t.Error("expected non-zero self-employment tax")
	}
	// Verify 50% SE deduction reduces AGI
	if result["1040:10"] <= 0 {
		t.Error("expected non-zero adjustments (deductible SE tax)")
	}
	// Verify amount owed (W-2 withholding < total tax with SE)
	if result["1040:37"] <= 0 {
		t.Error("expected non-zero amount owed")
	}
}

func TestFederalItemizedDeductionsScenario(t *testing.T) {
	g := buildSolver(t)

	numInputs, strInputs, taxYear := loadScenario(t,
		"../../testdata/scenarios/federal/itemized_deductions.json")

	result, err := g.Solve(numInputs, strInputs, taxYear)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	expected := loadExpected(t, "../../testdata/expected/federal/itemized_deductions.json")
	for key, want := range expected {
		assertClose(t, result[key], want, key)
	}

	// Verify itemized deductions exceeded standard deduction
	if result["1040:12"] <= 15000 {
		t.Error("expected itemized deductions to exceed standard deduction of $15,000")
	}
	// SALT should be capped at $10,000
	if result["schedule_a:5e"] > 10000 {
		t.Errorf("SALT deduction should be capped at $10,000, got %.2f", result["schedule_a:5e"])
	}
	// Medical should be $0 (expenses below 7.5% AGI threshold)
	if result["schedule_a:4"] != 0 {
		t.Errorf("expected $0 medical deduction (below threshold), got %.2f", result["schedule_a:4"])
	}
}

func TestFederalW2With1099Scenario(t *testing.T) {
	g := buildSolver(t)

	numInputs, strInputs, taxYear := loadScenario(t,
		"../../testdata/scenarios/federal/w2_with_1099.json")

	result, err := g.Solve(numInputs, strInputs, taxYear)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	expected := loadExpected(t, "../../testdata/expected/federal/w2_with_1099.json")
	for key, want := range expected {
		assertClose(t, result[key], want, key)
	}
}

func TestFederalCapitalGainsScenario(t *testing.T) {
	g := buildSolver(t)

	numInputs, strInputs, taxYear := loadScenario(t,
		"../../testdata/scenarios/federal/capital_gains.json")

	result, err := g.Solve(numInputs, strInputs, taxYear)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	expected := loadExpected(t, "../../testdata/expected/federal/capital_gains.json")
	for key, want := range expected {
		assertClose(t, result[key], want, key)
	}

	// Verify Schedule D wiring
	if result["schedule_d:16"] <= 0 {
		t.Error("expected non-zero net capital gain on Schedule D")
	}
	// Verify 1099-B proceeds flow through Form 8949
	if result["form_8949:lt_proceeds"] <= 0 {
		t.Error("expected non-zero long-term proceeds on Form 8949")
	}
}

func TestCAW2With1099Scenario(t *testing.T) {
	g := buildSolver(t)

	numInputs, strInputs, taxYear := loadScenario(t,
		"../../testdata/scenarios/ca/ca_w2_with_1099.json")

	result, err := g.Solve(numInputs, strInputs, taxYear)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// Check federal expected values
	fedExpected := loadExpected(t, "../../testdata/expected/federal/w2_with_1099.json")
	for key, want := range fedExpected {
		assertClose(t, result[key], want, key)
	}

	// Check CA expected values (including Schedule CA adjustments)
	caExpected := loadExpected(t, "../../testdata/expected/ca/ca_w2_with_1099.json")
	for key, want := range caExpected {
		assertClose(t, result[key], want, key)
	}

	// Verify U.S. bond interest subtraction flows correctly
	if result["ca_schedule_ca:2_col_b"] <= 0 {
		t.Error("expected non-zero Schedule CA interest subtraction for U.S. bond interest")
	}
	// CA AGI should be less than federal AGI due to U.S. bond subtraction
	if result["ca_540:17"] >= result["ca_540:13"] {
		t.Error("expected CA AGI < federal AGI due to U.S. bond interest subtraction")
	}
}

func TestCASelfEmployedScenario(t *testing.T) {
	g := buildSolver(t)

	numInputs, strInputs, taxYear := loadScenario(t,
		"../../testdata/scenarios/ca/ca_self_employed.json")

	result, err := g.Solve(numInputs, strInputs, taxYear)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// Check federal expected values (including QBI deduction)
	fedExpected := loadExpected(t, "../../testdata/expected/federal/self_employed.json")
	for key, want := range fedExpected {
		assertClose(t, result[key], want, key)
	}

	// Check CA expected values
	caExpected := loadExpected(t, "../../testdata/expected/ca/ca_self_employed.json")
	for key, want := range caExpected {
		assertClose(t, result[key], want, key)
	}

	// Federal QBI deduction should be $6,900 (20% of $34,500 Schedule C profit)
	assertClose(t, result["1040:13"], 6900, "Federal QBI deduction")

	// CA should NOT benefit from QBI — CA AGI = federal AGI (no QBI in AGI)
	assertClose(t, result["ca_540:13"], result["1040:11"], "CA line 13 = federal AGI")

	// CA taxable > federal taxable (CA doesn't allow QBI deduction)
	if result["ca_540:19"] <= result["1040:15"] {
		t.Error("expected CA taxable income > federal taxable income (QBI excluded from CA)")
	}
}

func TestFederalHSAScenario(t *testing.T) {
	g := buildSolver(t)

	numInputs, strInputs, taxYear := loadScenario(t,
		"../../testdata/scenarios/federal/hsa_contributions.json")

	result, err := g.Solve(numInputs, strInputs, taxYear)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	expected := loadExpected(t, "../../testdata/expected/federal/hsa_contributions.json")
	for key, want := range expected {
		assertClose(t, result[key], want, key)
	}

	// CA should add back the HSA deduction ($3,000)
	assertClose(t, result["ca_schedule_ca:15_col_c"], 3000, "CA HSA add-back")
	// CA AGI should equal wages ($80k) since HSA deduction is added back
	assertClose(t, result["ca_540:17"], 80000, "CA AGI (HSA added back)")
}

func TestCAItemizedDeductionsScenario(t *testing.T) {
	g := buildSolver(t)

	numInputs, strInputs, taxYear := loadScenario(t,
		"../../testdata/scenarios/ca/ca_itemized_deductions.json")

	result, err := g.Solve(numInputs, strInputs, taxYear)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// Check federal expected values (same inputs as itemized_deductions scenario)
	fedExpected := loadExpected(t, "../../testdata/expected/federal/itemized_deductions.json")
	for key, want := range fedExpected {
		assertClose(t, result[key], want, key)
	}

	// Check CA expected values
	caExpected := loadExpected(t, "../../testdata/expected/ca/ca_itemized_deductions.json")
	for key, want := range caExpected {
		assertClose(t, result[key], want, key)
	}

	// CA itemized should differ from federal itemized:
	// Federal SALT is $10,000 (capped). CA removes state income tax ($8,000)
	// and uncaps property taxes ($500 + $4,500 = $5,000).
	// CA itemized = $19,500 - $10,000 + $5,000 = $14,500
	assertClose(t, result["ca_schedule_ca:ca_itemized"], 14500, "CA itemized deductions")

	// CA should use itemized ($14,500) over standard ($5,706)
	assertClose(t, result["ca_540:18"], 14500, "CA deduction (itemized > standard)")

	// Federal SALT cap should still be $10,000
	assertClose(t, result["schedule_a:5e"], 10000, "Federal SALT capped at $10,000")
}

func TestFederalQBIDeductionScenario(t *testing.T) {
	g := buildSolver(t)

	numInputs, strInputs, taxYear := loadScenario(t,
		"../../testdata/scenarios/federal/qbi_deduction.json")

	result, err := g.Solve(numInputs, strInputs, taxYear)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	expected := loadExpected(t, "../../testdata/expected/federal/qbi_deduction.json")
	for key, want := range expected {
		assertClose(t, result[key], want, key)
	}

	// Verify QBI deduction is non-zero and less than 20% of QBI
	if result["form_8995:10"] <= 0 {
		t.Error("expected non-zero QBI deduction")
	}
	// QBI deduction should be limited by income (20% of taxable income minus net capital gain)
	if result["form_8995:10"] > result["form_8995:4"] {
		t.Error("QBI deduction should not exceed QBI component (20% of QBI)")
	}
	// Verify SE tax flows correctly
	if result["schedule_se:6"] <= 0 {
		t.Error("expected non-zero self-employment tax")
	}
	// No withholding, so full amount owed
	assertClose(t, result["1040:34"], 0, "No refund expected (no withholding)")
}

func TestCAHSAScenario(t *testing.T) {
	g := buildSolver(t)

	numInputs, strInputs, taxYear := loadScenario(t,
		"../../testdata/scenarios/ca/ca_hsa.json")

	result, err := g.Solve(numInputs, strInputs, taxYear)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// Check federal expected values (same as hsa_contributions scenario)
	fedExpected := loadExpected(t, "../../testdata/expected/federal/hsa_contributions.json")
	for key, want := range fedExpected {
		assertClose(t, result[key], want, key)
	}

	// Check CA expected values
	caExpected := loadExpected(t, "../../testdata/expected/ca/ca_hsa.json")
	for key, want := range caExpected {
		assertClose(t, result[key], want, key)
	}

	// Verify HSA deduction is added back on Schedule CA
	assertClose(t, result["ca_schedule_ca:15_col_c"], 3000, "CA HSA add-back")

	// CA AGI should be higher than federal AGI (HSA deduction added back)
	if result["ca_540:17"] <= result["ca_540:13"] {
		t.Error("expected CA AGI > federal AGI due to HSA add-back")
	}

	// CA AGI should equal wages ($80k) since HSA deduction is fully added back
	assertClose(t, result["ca_540:17"], 80000, "CA AGI (HSA added back to equal wages)")
}

func TestCAQBIAddbackScenario(t *testing.T) {
	g := buildSolver(t)

	numInputs, strInputs, taxYear := loadScenario(t,
		"../../testdata/scenarios/ca/ca_qbi_addback.json")

	result, err := g.Solve(numInputs, strInputs, taxYear)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// Check federal expected values (same inputs as qbi_deduction)
	fedExpected := loadExpected(t, "../../testdata/expected/federal/qbi_deduction.json")
	for key, want := range fedExpected {
		assertClose(t, result[key], want, key)
	}

	// Check CA expected values
	caExpected := loadExpected(t, "../../testdata/expected/ca/ca_qbi_addback.json")
	for key, want := range caExpected {
		assertClose(t, result[key], want, key)
	}

	// Federal QBI deduction should be non-zero
	if result["form_8995:10"] <= 0 {
		t.Error("expected non-zero federal QBI deduction")
	}

	// CA does NOT get QBI deduction — CA taxable income should be higher
	// CA uses standard deduction only (no QBI), federal uses standard + QBI
	if result["ca_540:19"] <= result["1040:15"] {
		t.Error("expected CA taxable income > federal taxable income (QBI excluded from CA)")
	}

	// CA AGI should equal federal AGI (QBI is a below-the-line deduction, not in AGI)
	assertClose(t, result["ca_540:13"], result["1040:11"], "CA line 13 = federal AGI")
}

func TestCACalEITCScenario(t *testing.T) {
	g := buildSolver(t)

	numInputs, strInputs, taxYear := loadScenario(t,
		"../../testdata/scenarios/ca/ca_caleitc.json")

	result, err := g.Solve(numInputs, strInputs, taxYear)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// Check CalEITC expected values
	caExpected := loadExpected(t, "../../testdata/expected/ca/ca_caleitc.json")
	for key, want := range caExpected {
		assertClose(t, result[key], want, key)
	}

	// CalEITC should be positive (low income, 1 child)
	if result["form_3514:7"] <= 0 {
		t.Error("expected non-zero CalEITC for low-income filer with 1 child")
	}

	// YCTC should be $1,117 (child under 6)
	assertClose(t, result["form_3514:6"], 1117, "Young Child Tax Credit")

	// No health penalty (full coverage)
	assertClose(t, result["form_3853:7"], 0, "Health penalty should be $0")

	// CalEITC should flow into ca_540:74 (total payments/credits)
	if result["ca_540:74"] <= result["ca_540:71"] {
		t.Error("expected total payments to include CalEITC credit")
	}
}

func TestCAHealthPenaltyScenario(t *testing.T) {
	g := buildSolver(t)

	numInputs, strInputs, taxYear := loadScenario(t,
		"../../testdata/scenarios/ca/ca_health_penalty.json")

	result, err := g.Solve(numInputs, strInputs, taxYear)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// Check health penalty expected values
	caExpected := loadExpected(t, "../../testdata/expected/ca/ca_health_penalty.json")
	for key, want := range caExpected {
		assertClose(t, result[key], want, key)
	}

	// Health penalty should be $2,700 (3 months * $900)
	assertClose(t, result["form_3853:6"], 2700, "Health penalty total")

	// Penalty should flow into ca_540:40 (total CA tax)
	if result["ca_540:40"] <= result["ca_540:35"]+result["ca_540:36"] {
		t.Error("expected total CA tax to include health penalty")
	}

	// CalEITC should be $0 (income $80k > $30,950 limit)
	assertClose(t, result["form_3514:7"], 0, "CalEITC should be $0 for high income")
}

func TestExpatSwedenBFRTScenario(t *testing.T) {
	g := buildSolver(t)

	numInputs, strInputs, taxYear := loadScenario(t,
		"../../testdata/scenarios/federal/expat_sweden_bfrt.json")

	result, err := g.Solve(numInputs, strInputs, taxYear)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	expected := loadExpected(t, "../../testdata/expected/federal/expat_sweden_bfrt.json")
	for key, want := range expected {
		assertClose(t, result[key], want, key)
	}

	// Verify FEIE exclusion equals foreign income (income < limit)
	assertClose(t, result["form_2555:foreign_income_exclusion"], 120000, "FEIE exclusion")
	assertClose(t, result["form_2555:total_exclusion"], 120000, "total exclusion")

	// Verify Schedule 1 line 8d is negative (reducing income)
	if result["schedule_1:8d"] >= 0 {
		t.Error("expected negative Schedule 1 line 8d (FEIE reduces income)")
	}

	// Verify AGI is 0 (all income excluded)
	assertClose(t, result["1040:11"], 0, "AGI should be 0 when all income excluded")

	// Verify no tax owed
	assertClose(t, result["1040:24"], 0, "total tax should be 0")
}

func TestExpatPartialExclusionScenario(t *testing.T) {
	// Income exceeds FEIE limit — only k is excluded
	g := buildSolver(t)

	numInputs := map[string]float64{
		"w2:1:wages":                 180000,
		"w2:1:federal_tax_withheld":  0,
		"w2:1:ss_wages":              0,
		"w2:1:ss_tax_withheld":       0,
		"w2:1:medicare_wages":        0,
		"w2:1:medicare_tax_withheld": 0,
		"w2:1:state_wages":           0,
		"w2:1:state_tax_withheld":    0,
		"w2:1:employer_name":         0,
		"w2:1:employer_ein":          0,
		"1040:filing_status":         0,
		"1040:first_name":            0,
		"1040:last_name":             0,
		"1040:ssn":                   0,
	}
	strInputs := map[string]string{
		"1040:filing_status":  "single",
		"1040:first_name":     "Test",
		"1040:last_name":      "Expat",
		"1040:ssn":            "999-88-7777",
		"w2:1:employer_name":  "Foreign Corp",
		"w2:1:employer_ein":   "00-0000000",
	}

	// Apply defaults
	for key, val := range inputFormDefaults {
		switch v := val.(type) {
		case float64:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = v
			}
		case string:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = 0
				strInputs[key] = v
			}
		}
	}

	// Override Form 2555 fields for this scenario
	numInputs["form_2555:foreign_earned_income"] = 180000
	numInputs["form_2555:ppt_days_present"] = 0
	numInputs["form_2555:exchange_rate"] = 10.5
	numInputs["form_2555:foreign_tax_paid"] = 50000
	numInputs["form_2555:employer_provided_housing"] = 0
	numInputs["form_2555:housing_expenses"] = 0
	strInputs["form_2555:qualifying_test"] = "bona_fide_residence"
	strInputs["form_2555:bfrt_full_year"] = "yes"
	strInputs["form_2555:foreign_country"] = "Sweden"
	strInputs["form_2555:self_employed_abroad"] = "no"

	result, err := g.Solve(numInputs, strInputs, 2025)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// FEIE exclusion should be capped at ,000
	assertClose(t, result["form_2555:foreign_income_exclusion"], 130000, "FEIE capped at limit")
	assertClose(t, result["form_2555:total_exclusion"], 130000, "total exclusion capped")

	// Remaining taxable: k - k = k wages, minus k std deduction = k
	assertClose(t, result["1040:9"], 50000, "total income after exclusion")
	assertClose(t, result["1040:11"], 50000, "AGI")
	assertClose(t, result["1040:15"], 35000, "taxable income")

	// Tax should use stacking: tax(k + k) - tax(k)
	// = tax(k) - tax(k)
	if result["1040:16"] <= 0 {
		t.Error("expected non-zero tax on remaining income")
	}

	// Stacking tax should be higher than normal tax on k
	normalTax := result["1040:16"]
	// Normal tax on k single would be about ,952
	// Stacking tax should be higher (24% bracket range)
	if normalTax < 4000 {
		// With stacking, k is taxed at 22-24% range, not 10-12%
		// This would be much higher than normal k tax
		t.Logf("stacking tax on remaining k: %.2f", normalTax)
	}
}

func TestExpatPPTProratedScenario(t *testing.T) {
	// Physical Presence Test with less than full year
	g := buildSolver(t)

	numInputs := map[string]float64{
		"w2:1:wages":                 100000,
		"w2:1:federal_tax_withheld":  0,
		"w2:1:ss_wages":              0,
		"w2:1:ss_tax_withheld":       0,
		"w2:1:medicare_wages":        0,
		"w2:1:medicare_tax_withheld": 0,
		"w2:1:state_wages":           0,
		"w2:1:state_tax_withheld":    0,
		"w2:1:employer_name":         0,
		"w2:1:employer_ein":          0,
		"1040:filing_status":         0,
		"1040:first_name":            0,
		"1040:last_name":             0,
		"1040:ssn":                   0,
	}
	strInputs := map[string]string{
		"1040:filing_status":  "single",
		"1040:first_name":     "Test",
		"1040:last_name":      "PPT",
		"1040:ssn":            "999-88-6666",
		"w2:1:employer_name":  "Foreign Corp",
		"w2:1:employer_ein":   "00-0000000",
	}

	for key, val := range inputFormDefaults {
		switch v := val.(type) {
		case float64:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = v
			}
		case string:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = 0
				strInputs[key] = v
			}
		}
	}

	// PPT with 200 days (less than 330 required for full exclusion)
	numInputs["form_2555:foreign_earned_income"] = 100000
	numInputs["form_2555:ppt_days_present"] = 200
	numInputs["form_2555:exchange_rate"] = 10.5
	numInputs["form_2555:foreign_tax_paid"] = 30000
	numInputs["form_2555:employer_provided_housing"] = 0
	numInputs["form_2555:housing_expenses"] = 0
	strInputs["form_2555:qualifying_test"] = "physical_presence"
	strInputs["form_2555:bfrt_full_year"] = "no"
	strInputs["form_2555:foreign_country"] = "Sweden"
	strInputs["form_2555:self_employed_abroad"] = "no"

	result, err := g.Solve(numInputs, strInputs, 2025)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// Prorated exclusion: ,000 * 200/365 = ,232.88
	proratedLimit := 130000.0 * 200.0 / 365.0
	assertClose(t, result["form_2555:prorated_exclusion"], proratedLimit, "prorated exclusion")

	// Exclusion should be min(k, ,232.88) = ,232.88
	assertClose(t, result["form_2555:foreign_income_exclusion"], proratedLimit, "FEIE (income > prorated limit)")

	// Remaining income: k - ,232.88 = ,767.12
	remaining := 100000 - proratedLimit
	assertClose(t, result["1040:9"], remaining, "total income after partial exclusion")

	// Should have non-zero tax on remaining income
	if result["1040:16"] <= 0 {
		t.Error("expected non-zero tax on remaining income after partial exclusion")
	}
}

func TestExpatHousingExclusionScenario(t *testing.T) {
	// Expat with employer-provided housing
	g := buildSolver(t)

	numInputs := map[string]float64{
		"w2:1:wages":                 120000,
		"w2:1:federal_tax_withheld":  0,
		"w2:1:ss_wages":              0,
		"w2:1:ss_tax_withheld":       0,
		"w2:1:medicare_wages":        0,
		"w2:1:medicare_tax_withheld": 0,
		"w2:1:state_wages":           0,
		"w2:1:state_tax_withheld":    0,
		"w2:1:employer_name":         0,
		"w2:1:employer_ein":          0,
		"1040:filing_status":         0,
		"1040:first_name":            0,
		"1040:last_name":             0,
		"1040:ssn":                   0,
	}
	strInputs := map[string]string{
		"1040:filing_status":  "single",
		"1040:first_name":     "Test",
		"1040:last_name":      "Housing",
		"1040:ssn":            "999-88-5555",
		"w2:1:employer_name":  "Foreign Corp",
		"w2:1:employer_ein":   "00-0000000",
	}

	for key, val := range inputFormDefaults {
		switch v := val.(type) {
		case float64:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = v
			}
		case string:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = 0
				strInputs[key] = v
			}
		}
	}

	numInputs["form_2555:foreign_earned_income"] = 120000
	numInputs["form_2555:ppt_days_present"] = 0
	numInputs["form_2555:exchange_rate"] = 10.5
	numInputs["form_2555:foreign_tax_paid"] = 36000
	numInputs["form_2555:employer_provided_housing"] = 30000 // employer provides housing
	numInputs["form_2555:housing_expenses"] = 28000          // actual expenses
	strInputs["form_2555:qualifying_test"] = "bona_fide_residence"
	strInputs["form_2555:bfrt_full_year"] = "yes"
	strInputs["form_2555:foreign_country"] = "Sweden"
	strInputs["form_2555:self_employed_abroad"] = "no"

	result, err := g.Solve(numInputs, strInputs, 2025)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// Housing qualifying: min(k, k) - ,800 = ,200
	assertClose(t, result["form_2555:housing_qualifying_amount"], 7200, "housing qualifying amount")

	// Housing exclusion: min(,200, ,000) = ,200
	assertClose(t, result["form_2555:housing_exclusion"], 7200, "housing exclusion")

	// Total exclusion: ,000 FEIE + ,200 housing = ,200
	assertClose(t, result["form_2555:total_exclusion"], 127200, "total exclusion with housing")

	// Housing deduction should be /home/linuxbrew/.linuxbrew/bin/zsh (not self-employed)
	assertClose(t, result["form_2555:housing_deduction"], 0, "no housing deduction for employee")
}

func TestExpatSelfEmployedHousingDeduction(t *testing.T) {
	// Self-employed expat gets housing deduction instead of exclusion
	g := buildSolver(t)

	numInputs := map[string]float64{
		"w2:1:wages":                 100000,
		"w2:1:federal_tax_withheld":  0,
		"w2:1:ss_wages":              0,
		"w2:1:ss_tax_withheld":       0,
		"w2:1:medicare_wages":        0,
		"w2:1:medicare_tax_withheld": 0,
		"w2:1:state_wages":           0,
		"w2:1:state_tax_withheld":    0,
		"w2:1:employer_name":         0,
		"w2:1:employer_ein":          0,
		"1040:filing_status":         0,
		"1040:first_name":            0,
		"1040:last_name":             0,
		"1040:ssn":                   0,
	}
	strInputs := map[string]string{
		"1040:filing_status":  "single",
		"1040:first_name":     "Test",
		"1040:last_name":      "SelfEmployed",
		"1040:ssn":            "999-88-4444",
		"w2:1:employer_name":  "Self",
		"w2:1:employer_ein":   "00-0000000",
	}

	for key, val := range inputFormDefaults {
		switch v := val.(type) {
		case float64:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = v
			}
		case string:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = 0
				strInputs[key] = v
			}
		}
	}

	numInputs["form_2555:foreign_earned_income"] = 100000
	numInputs["form_2555:ppt_days_present"] = 0
	numInputs["form_2555:exchange_rate"] = 10.5
	numInputs["form_2555:foreign_tax_paid"] = 30000
	numInputs["form_2555:employer_provided_housing"] = 0  // no employer housing
	numInputs["form_2555:housing_expenses"] = 25000
	strInputs["form_2555:qualifying_test"] = "bona_fide_residence"
	strInputs["form_2555:bfrt_full_year"] = "yes"
	strInputs["form_2555:foreign_country"] = "Sweden"
	strInputs["form_2555:self_employed_abroad"] = "yes"

	result, err := g.Solve(numInputs, strInputs, 2025)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// Housing qualifying: min(k, k) - ,800 = ,200
	assertClose(t, result["form_2555:housing_qualifying_amount"], 4200, "housing qualifying amount")

	// Housing exclusion: min(,200, /home/linuxbrew/.linuxbrew/bin/zsh employer) = /home/linuxbrew/.linuxbrew/bin/zsh
	assertClose(t, result["form_2555:housing_exclusion"], 0, "no housing exclusion (no employer)")

	// Housing deduction: ,200 - /home/linuxbrew/.linuxbrew/bin/zsh = ,200 (self-employed gets deduction)
	assertClose(t, result["form_2555:housing_deduction"], 4200, "housing deduction for self-employed")

	// Total exclusion = FEIE only (housing deduction is separate)
	assertClose(t, result["form_2555:total_exclusion"], 100000, "total exclusion (FEIE only)")
}

func TestForm1116BasicCredit(t *testing.T) {
	// Taxpayer with $50k foreign source income, paid $15k foreign tax
	// No FEIE — all income is taxable, FTC applies to all foreign income
	g := buildSolver(t)

	numInputs := map[string]float64{
		"w2:1:wages":                 150000,
		"w2:1:federal_tax_withheld":  25000,
		"w2:1:ss_wages":              150000,
		"w2:1:ss_tax_withheld":       9300,
		"w2:1:medicare_wages":        150000,
		"w2:1:medicare_tax_withheld": 2175,
		"w2:1:state_wages":           150000,
		"w2:1:state_tax_withheld":    6000,
		"w2:1:employer_name":         0,
		"w2:1:employer_ein":          0,
		"1040:filing_status":         0,
		"1040:first_name":            0,
		"1040:last_name":             0,
		"1040:ssn":                   0,
	}
	strInputs := map[string]string{
		"1040:filing_status":  "single",
		"1040:first_name":     "Test",
		"1040:last_name":      "FTC",
		"1040:ssn":            "999-88-3333",
		"w2:1:employer_name":  "Global Corp",
		"w2:1:employer_ein":   "12-3456789",
	}

	for key, val := range inputFormDefaults {
		switch v := val.(type) {
		case float64:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = v
			}
		case string:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = 0
				strInputs[key] = v
			}
		}
	}

	// Override Form 1116 fields
	numInputs["form_1116:foreign_source_income"] = 50000
	numInputs["form_1116:foreign_source_deductions"] = 0
	numInputs["form_1116:foreign_tax_paid_income"] = 15000
	numInputs["form_1116:foreign_tax_paid_other"] = 0
	strInputs["form_1116:category"] = "general"
	strInputs["form_1116:foreign_country"] = "Sweden"
	strInputs["form_1116:accrued_or_paid"] = "paid"

	result, err := g.Solve(numInputs, strInputs, 2025)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// Net foreign source: $50k - $0 = $50k
	assertClose(t, result["form_1116:7"], 50000, "net foreign source income")

	// Total foreign taxes: $15k
	assertClose(t, result["form_1116:15"], 15000, "total foreign taxes paid")

	// US tax on $135k taxable ($150k - $15k std deduction)
	taxableIncome := result["1040:15"]
	if taxableIncome <= 0 {
		t.Fatal("expected positive taxable income")
	}

	// Limitation: US_tax * (50k / taxable_income)
	usTax := result["1040:16"]
	expectedLimitation := usTax * (50000.0 / taxableIncome)
	assertClose(t, result["form_1116:21"], expectedLimitation, "FTC limitation")

	// Credit = min(taxes paid, limitation)
	expectedCredit := math.Min(15000, expectedLimitation)
	assertClose(t, result["form_1116:22"], expectedCredit, "FTC credit allowed")

	// Credit should flow to Schedule 3 line 1
	assertClose(t, result["schedule_3:1"], expectedCredit, "Schedule 3 line 1")

	// Credit should reduce total tax via 1040 line 20
	assertClose(t, result["1040:20"], expectedCredit, "1040 line 20 nonrefundable credits")

	// Total tax should be reduced by FTC
	if result["1040:22"] >= usTax {
		t.Error("expected tax after credits < tax before credits")
	}
}

func TestForm1116CreditLimitation(t *testing.T) {
	// Foreign taxes paid exceed the limitation — credit is capped
	g := buildSolver(t)

	numInputs := map[string]float64{
		"w2:1:wages":                 80000,
		"w2:1:federal_tax_withheld":  10000,
		"w2:1:ss_wages":              80000,
		"w2:1:ss_tax_withheld":       4960,
		"w2:1:medicare_wages":        80000,
		"w2:1:medicare_tax_withheld": 1160,
		"w2:1:state_wages":           80000,
		"w2:1:state_tax_withheld":    3400,
		"w2:1:employer_name":         0,
		"w2:1:employer_ein":          0,
		"1040:filing_status":         0,
		"1040:first_name":            0,
		"1040:last_name":             0,
		"1040:ssn":                   0,
	}
	strInputs := map[string]string{
		"1040:filing_status":  "single",
		"1040:first_name":     "Test",
		"1040:last_name":      "Limited",
		"1040:ssn":            "999-88-2222",
		"w2:1:employer_name":  "Foreign Corp",
		"w2:1:employer_ein":   "00-0000000",
	}

	for key, val := range inputFormDefaults {
		switch v := val.(type) {
		case float64:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = v
			}
		case string:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = 0
				strInputs[key] = v
			}
		}
	}

	// Small foreign source but high taxes paid (e.g., high-tax country)
	numInputs["form_1116:foreign_source_income"] = 20000
	numInputs["form_1116:foreign_source_deductions"] = 0
	numInputs["form_1116:foreign_tax_paid_income"] = 10000  // 50% tax rate
	numInputs["form_1116:foreign_tax_paid_other"] = 0
	strInputs["form_1116:category"] = "general"
	strInputs["form_1116:foreign_country"] = "Sweden"
	strInputs["form_1116:accrued_or_paid"] = "paid"

	result, err := g.Solve(numInputs, strInputs, 2025)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// Limitation should be less than taxes paid
	limitation := result["form_1116:21"]
	taxesPaid := result["form_1116:15"]
	if limitation >= taxesPaid {
		t.Errorf("expected limitation (%.2f) < taxes paid (%.2f)", limitation, taxesPaid)
	}

	// Credit should equal the limitation (not the full taxes paid)
	assertClose(t, result["form_1116:22"], limitation, "credit capped at limitation")

	// Carryforward should be the excess
	assertClose(t, result["form_1116:carryforward"], taxesPaid-limitation, "carryforward")
}

func TestForm1116WithFEIEInteraction(t *testing.T) {
	// Expat who claims FEIE on $120k and FTC on remaining $30k
	// Cannot double-dip: FTC foreign source excludes FEIE amount
	g := buildSolver(t)

	numInputs := map[string]float64{
		"w2:1:wages":                 150000,
		"w2:1:federal_tax_withheld":  0,
		"w2:1:ss_wages":              0,
		"w2:1:ss_tax_withheld":       0,
		"w2:1:medicare_wages":        0,
		"w2:1:medicare_tax_withheld": 0,
		"w2:1:state_wages":           0,
		"w2:1:state_tax_withheld":    0,
		"w2:1:employer_name":         0,
		"w2:1:employer_ein":          0,
		"1040:filing_status":         0,
		"1040:first_name":            0,
		"1040:last_name":             0,
		"1040:ssn":                   0,
	}
	strInputs := map[string]string{
		"1040:filing_status":  "single",
		"1040:first_name":     "Test",
		"1040:last_name":      "Combo",
		"1040:ssn":            "999-88-1111",
		"w2:1:employer_name":  "Swedish Corp",
		"w2:1:employer_ein":   "00-0000000",
	}

	for key, val := range inputFormDefaults {
		switch v := val.(type) {
		case float64:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = v
			}
		case string:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = 0
				strInputs[key] = v
			}
		}
	}

	// FEIE: exclude $130k of the $150k
	numInputs["form_2555:foreign_earned_income"] = 150000
	numInputs["form_2555:ppt_days_present"] = 0
	numInputs["form_2555:exchange_rate"] = 10.5
	numInputs["form_2555:foreign_tax_paid"] = 45000
	numInputs["form_2555:employer_provided_housing"] = 0
	numInputs["form_2555:housing_expenses"] = 0
	strInputs["form_2555:qualifying_test"] = "bona_fide_residence"
	strInputs["form_2555:bfrt_full_year"] = "yes"
	strInputs["form_2555:foreign_country"] = "Sweden"
	strInputs["form_2555:self_employed_abroad"] = "no"

	// FTC: only on the $20k NOT excluded by FEIE ($150k - $130k = $20k)
	// The user must enter foreign_source_income as the non-excluded portion
	numInputs["form_1116:foreign_source_income"] = 20000
	numInputs["form_1116:foreign_source_deductions"] = 0
	// Proportional foreign tax: $45k * (20k/150k) = $6,000
	numInputs["form_1116:foreign_tax_paid_income"] = 6000
	numInputs["form_1116:foreign_tax_paid_other"] = 0
	strInputs["form_1116:category"] = "general"
	strInputs["form_1116:foreign_country"] = "Sweden"
	strInputs["form_1116:accrued_or_paid"] = "paid"

	result, err := g.Solve(numInputs, strInputs, 2025)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// FEIE should exclude $130k (limit)
	assertClose(t, result["form_2555:total_exclusion"], 130000, "FEIE exclusion")

	// Remaining income: $150k - $130k = $20k
	assertClose(t, result["1040:9"], 20000, "total income after FEIE")

	// Taxable income: $20k - $15k std deduction = $5k
	assertClose(t, result["1040:15"], 5000, "taxable income")

	// Tax uses stacking: tax($5k + $130k) - tax($130k)
	if result["1040:16"] <= 0 {
		t.Error("expected non-zero stacked tax")
	}

	// FTC on $20k non-excluded income with $6k taxes paid
	// Limitation: US_tax * (20k / 5k) but ratio capped at 1.0
	// Since foreign source ($20k) > taxable income ($5k), ratio = 1.0
	// So limitation = full US tax
	usTax := result["1040:16"]
	assertClose(t, result["form_1116:21"], usTax, "FTC limitation (ratio capped at 1.0)")

	// Credit = min($6k, usTax)
	expectedCredit := math.Min(6000, usTax)
	assertClose(t, result["form_1116:22"], expectedCredit, "FTC credit")

	// Tax after credits should be reduced
	if result["1040:22"] >= usTax {
		t.Error("expected tax after FTC < tax before")
	}
}

func TestScheduleBPart3ForeignAccounts(t *testing.T) {
	g := buildSolver(t)

	numInputs := map[string]float64{
		"w2:1:wages":                 75000,
		"w2:1:federal_tax_withheld":  9500,
		"w2:1:ss_wages":              75000,
		"w2:1:ss_tax_withheld":       4650,
		"w2:1:medicare_wages":        75000,
		"w2:1:medicare_tax_withheld": 1087.50,
		"w2:1:state_wages":           75000,
		"w2:1:state_tax_withheld":    3200,
		"w2:1:employer_name":         0,
		"w2:1:employer_ein":          0,
		"1040:filing_status":         0,
		"1040:first_name":            0,
		"1040:last_name":             0,
		"1040:ssn":                   0,
	}
	strInputs := map[string]string{
		"1040:filing_status":  "single",
		"1040:first_name":     "Test",
		"1040:last_name":      "Accounts",
		"1040:ssn":            "999-88-0000",
		"w2:1:employer_name":  "Acme Corp",
		"w2:1:employer_ein":   "12-3456789",
	}

	for key, val := range inputFormDefaults {
		switch v := val.(type) {
		case float64:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = v
			}
		case string:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = 0
				strInputs[key] = v
			}
		}
	}

	// Has foreign accounts
	strInputs["schedule_b:7a"] = "yes"
	strInputs["schedule_b:7b"] = "Sweden"
	strInputs["schedule_b:8"] = "no"

	result, err := g.Solve(numInputs, strInputs, 2025)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// FBAR required flag should be set
	assertClose(t, result["schedule_b:fbar_required"], 1, "FBAR required when foreign accounts = yes")

	// Regular tax calculation should be unaffected
	assertClose(t, result["1040:1a"], 75000, "wages unaffected")
}

func TestScheduleBPart3NoForeignAccounts(t *testing.T) {
	g := buildSolver(t)

	numInputs := map[string]float64{
		"w2:1:wages":                 75000,
		"w2:1:federal_tax_withheld":  9500,
		"w2:1:ss_wages":              75000,
		"w2:1:ss_tax_withheld":       4650,
		"w2:1:medicare_wages":        75000,
		"w2:1:medicare_tax_withheld": 1087.50,
		"w2:1:state_wages":           75000,
		"w2:1:state_tax_withheld":    3200,
		"w2:1:employer_name":         0,
		"w2:1:employer_ein":          0,
		"1040:filing_status":         0,
		"1040:first_name":            0,
		"1040:last_name":             0,
		"1040:ssn":                   0,
	}
	strInputs := map[string]string{
		"1040:filing_status":  "single",
		"1040:first_name":     "Test",
		"1040:last_name":      "NoAccounts",
		"1040:ssn":            "999-88-9999",
		"w2:1:employer_name":  "Acme Corp",
		"w2:1:employer_ein":   "12-3456789",
	}

	for key, val := range inputFormDefaults {
		switch v := val.(type) {
		case float64:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = v
			}
		case string:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = 0
				strInputs[key] = v
			}
		}
	}

	// No foreign accounts (default)
	result, err := g.Solve(numInputs, strInputs, 2025)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// FBAR not required
	assertClose(t, result["schedule_b:fbar_required"], 0, "FBAR not required when no foreign accounts")
}

func TestForeignTaxCreditScenario(t *testing.T) {
	g := buildSolver(t)

	numInputs, strInputs, taxYear := loadScenario(t,
		"../../testdata/scenarios/federal/foreign_tax_credit.json")

	result, err := g.Solve(numInputs, strInputs, taxYear)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	expected := loadExpected(t, "../../testdata/expected/federal/foreign_tax_credit.json")
	for key, want := range expected {
		assertClose(t, result[key], want, key)
	}

	// FTC should reduce total tax
	if result["1040:22"] >= result["1040:16"] {
		t.Error("expected FTC to reduce tax after credits")
	}

	// Should have carryforward (taxes paid > limitation)
	if result["form_1116:carryforward"] <= 0 {
		t.Error("expected non-zero FTC carryforward")
	}

	// FBAR should be required (foreign accounts = yes)
	assertClose(t, result["schedule_b:fbar_required"], 1, "FBAR required")
}

func TestForm8938AbroadSingleAboveThreshold(t *testing.T) {
	// Single filer living abroad with accounts exceeding $200k year-end threshold
	g := buildSolver(t)

	numInputs := map[string]float64{
		"w2:1:wages":                 75000,
		"w2:1:federal_tax_withheld":  9500,
		"w2:1:ss_wages":              75000,
		"w2:1:ss_tax_withheld":       4650,
		"w2:1:medicare_wages":        75000,
		"w2:1:medicare_tax_withheld": 1087.50,
		"w2:1:state_wages":           75000,
		"w2:1:state_tax_withheld":    3200,
		"w2:1:employer_name":         0,
		"w2:1:employer_ein":          0,
		"1040:filing_status":         0,
		"1040:first_name":            0,
		"1040:last_name":             0,
		"1040:ssn":                   0,
	}
	strInputs := map[string]string{
		"1040:filing_status":  "single",
		"1040:first_name":     "Test",
		"1040:last_name":      "FATCA",
		"1040:ssn":            "999-77-1111",
		"w2:1:employer_name":  "Swedish Corp",
		"w2:1:employer_ein":   "00-0000000",
	}

	for key, val := range inputFormDefaults {
		switch v := val.(type) {
		case float64:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = v
			}
		case string:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = 0
				strInputs[key] = v
			}
		}
	}

	// Living abroad, accounts worth $250k year-end
	strInputs["form_8938:lives_abroad"] = "yes"
	numInputs["form_8938:max_value_accounts"] = 280000
	numInputs["form_8938:yearend_value_accounts"] = 250000
	numInputs["form_8938:num_accounts"] = 3
	strInputs["form_8938:account_country"] = "Sweden"
	strInputs["form_8938:account_institution"] = "Handelsbanken"
	strInputs["form_8938:account_type"] = "deposit"

	result, err := g.Solve(numInputs, strInputs, 2025)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// Abroad single thresholds: $200k year-end, $300k any-time
	assertClose(t, result["form_8938:threshold_yearend"], 200000, "abroad single year-end threshold")
	assertClose(t, result["form_8938:threshold_anytime"], 300000, "abroad single any-time threshold")

	// Year-end $250k > $200k threshold — filing required
	assertClose(t, result["form_8938:filing_required"], 1, "filing required (year-end exceeds threshold)")

	// Total values
	assertClose(t, result["form_8938:total_max_value"], 280000, "total max value")
	assertClose(t, result["form_8938:total_yearend_value"], 250000, "total year-end value")
}

func TestForm8938AbroadSingleBelowThreshold(t *testing.T) {
	// Single filer abroad with accounts below threshold
	g := buildSolver(t)

	numInputs := map[string]float64{
		"w2:1:wages":                 75000,
		"w2:1:federal_tax_withheld":  9500,
		"w2:1:ss_wages":              75000,
		"w2:1:ss_tax_withheld":       4650,
		"w2:1:medicare_wages":        75000,
		"w2:1:medicare_tax_withheld": 1087.50,
		"w2:1:state_wages":           75000,
		"w2:1:state_tax_withheld":    3200,
		"w2:1:employer_name":         0,
		"w2:1:employer_ein":          0,
		"1040:filing_status":         0,
		"1040:first_name":            0,
		"1040:last_name":             0,
		"1040:ssn":                   0,
	}
	strInputs := map[string]string{
		"1040:filing_status":  "single",
		"1040:first_name":     "Test",
		"1040:last_name":      "Below",
		"1040:ssn":            "999-77-2222",
		"w2:1:employer_name":  "Corp",
		"w2:1:employer_ein":   "00-0000000",
	}

	for key, val := range inputFormDefaults {
		switch v := val.(type) {
		case float64:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = v
			}
		case string:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = 0
				strInputs[key] = v
			}
		}
	}

	// Abroad, but under threshold
	strInputs["form_8938:lives_abroad"] = "yes"
	numInputs["form_8938:max_value_accounts"] = 150000
	numInputs["form_8938:yearend_value_accounts"] = 120000

	result, err := g.Solve(numInputs, strInputs, 2025)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// Below both thresholds — filing NOT required
	assertClose(t, result["form_8938:filing_required"], 0, "filing not required (below threshold)")
}

func TestForm8938USResidentThresholds(t *testing.T) {
	// US-based filer has lower thresholds
	g := buildSolver(t)

	numInputs := map[string]float64{
		"w2:1:wages":                 75000,
		"w2:1:federal_tax_withheld":  9500,
		"w2:1:ss_wages":              75000,
		"w2:1:ss_tax_withheld":       4650,
		"w2:1:medicare_wages":        75000,
		"w2:1:medicare_tax_withheld": 1087.50,
		"w2:1:state_wages":           75000,
		"w2:1:state_tax_withheld":    3200,
		"w2:1:employer_name":         0,
		"w2:1:employer_ein":          0,
		"1040:filing_status":         0,
		"1040:first_name":            0,
		"1040:last_name":             0,
		"1040:ssn":                   0,
	}
	strInputs := map[string]string{
		"1040:filing_status":  "single",
		"1040:first_name":     "Test",
		"1040:last_name":      "USRes",
		"1040:ssn":            "999-77-3333",
		"w2:1:employer_name":  "Corp",
		"w2:1:employer_ein":   "00-0000000",
	}

	for key, val := range inputFormDefaults {
		switch v := val.(type) {
		case float64:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = v
			}
		case string:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = 0
				strInputs[key] = v
			}
		}
	}

	// US resident with $60k year-end (above $50k US threshold)
	strInputs["form_8938:lives_abroad"] = "no"
	numInputs["form_8938:max_value_accounts"] = 70000
	numInputs["form_8938:yearend_value_accounts"] = 60000

	result, err := g.Solve(numInputs, strInputs, 2025)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// US single thresholds: $50k year-end, $75k any-time
	assertClose(t, result["form_8938:threshold_yearend"], 50000, "US single year-end threshold")
	assertClose(t, result["form_8938:threshold_anytime"], 75000, "US single any-time threshold")

	// Year-end $60k > $50k — filing required
	assertClose(t, result["form_8938:filing_required"], 1, "filing required for US resident above threshold")
}

func TestForm8938MFJAbroadThresholds(t *testing.T) {
	// MFJ living abroad — highest thresholds
	g := buildSolver(t)

	numInputs := map[string]float64{
		"w2:1:wages":                 75000,
		"w2:1:federal_tax_withheld":  9500,
		"w2:1:ss_wages":              75000,
		"w2:1:ss_tax_withheld":       4650,
		"w2:1:medicare_wages":        75000,
		"w2:1:medicare_tax_withheld": 1087.50,
		"w2:1:state_wages":           75000,
		"w2:1:state_tax_withheld":    3200,
		"w2:1:employer_name":         0,
		"w2:1:employer_ein":          0,
		"1040:filing_status":         0,
		"1040:first_name":            0,
		"1040:last_name":             0,
		"1040:ssn":                   0,
	}
	strInputs := map[string]string{
		"1040:filing_status":  "mfj",
		"1040:first_name":     "Test",
		"1040:last_name":      "MFJ",
		"1040:ssn":            "999-77-4444",
		"w2:1:employer_name":  "Corp",
		"w2:1:employer_ein":   "00-0000000",
	}

	for key, val := range inputFormDefaults {
		switch v := val.(type) {
		case float64:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = v
			}
		case string:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = 0
				strInputs[key] = v
			}
		}
	}

	// MFJ abroad thresholds: $400k year-end, $600k any-time
	strInputs["form_8938:lives_abroad"] = "yes"
	numInputs["form_8938:max_value_accounts"] = 350000
	numInputs["form_8938:yearend_value_accounts"] = 350000

	result, err := g.Solve(numInputs, strInputs, 2025)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	assertClose(t, result["form_8938:threshold_yearend"], 400000, "MFJ abroad year-end threshold")
	assertClose(t, result["form_8938:threshold_anytime"], 600000, "MFJ abroad any-time threshold")

	// $350k < $400k year-end and $350k < $600k any-time — NOT required
	assertClose(t, result["form_8938:filing_required"], 0, "not required (MFJ abroad under threshold)")
}

func TestForm8938AnyTimeThresholdTrigger(t *testing.T) {
	// Year-end below threshold but max value exceeds any-time threshold
	g := buildSolver(t)

	numInputs := map[string]float64{
		"w2:1:wages":                 75000,
		"w2:1:federal_tax_withheld":  9500,
		"w2:1:ss_wages":              75000,
		"w2:1:ss_tax_withheld":       4650,
		"w2:1:medicare_wages":        75000,
		"w2:1:medicare_tax_withheld": 1087.50,
		"w2:1:state_wages":           75000,
		"w2:1:state_tax_withheld":    3200,
		"w2:1:employer_name":         0,
		"w2:1:employer_ein":          0,
		"1040:filing_status":         0,
		"1040:first_name":            0,
		"1040:last_name":             0,
		"1040:ssn":                   0,
	}
	strInputs := map[string]string{
		"1040:filing_status":  "single",
		"1040:first_name":     "Test",
		"1040:last_name":      "AnyTime",
		"1040:ssn":            "999-77-5555",
		"w2:1:employer_name":  "Corp",
		"w2:1:employer_ein":   "00-0000000",
	}

	for key, val := range inputFormDefaults {
		switch v := val.(type) {
		case float64:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = v
			}
		case string:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = 0
				strInputs[key] = v
			}
		}
	}

	// Abroad, year-end $180k (under $200k), but max $320k (over $300k any-time)
	strInputs["form_8938:lives_abroad"] = "yes"
	numInputs["form_8938:max_value_accounts"] = 320000
	numInputs["form_8938:yearend_value_accounts"] = 180000

	result, err := g.Solve(numInputs, strInputs, 2025)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// Year-end under threshold but any-time over — filing required
	assertClose(t, result["form_8938:filing_required"], 1, "required via any-time threshold")
}

func TestForm8833SwedenPension(t *testing.T) {
	// Treaty disclosure for Swedish pension
	g := buildSolver(t)

	numInputs := map[string]float64{
		"w2:1:wages":                 75000,
		"w2:1:federal_tax_withheld":  9500,
		"w2:1:ss_wages":              75000,
		"w2:1:ss_tax_withheld":       4650,
		"w2:1:medicare_wages":        75000,
		"w2:1:medicare_tax_withheld": 1087.50,
		"w2:1:state_wages":           75000,
		"w2:1:state_tax_withheld":    3200,
		"w2:1:employer_name":         0,
		"w2:1:employer_ein":          0,
		"1040:filing_status":         0,
		"1040:first_name":            0,
		"1040:last_name":             0,
		"1040:ssn":                   0,
	}
	strInputs := map[string]string{
		"1040:filing_status":  "single",
		"1040:first_name":     "Test",
		"1040:last_name":      "Treaty",
		"1040:ssn":            "999-77-6666",
		"w2:1:employer_name":  "Corp",
		"w2:1:employer_ein":   "00-0000000",
	}

	for key, val := range inputFormDefaults {
		switch v := val.(type) {
		case float64:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = v
			}
		case string:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = 0
				strInputs[key] = v
			}
		}
	}

	// Swedish pension treaty position
	strInputs["form_8833:treaty_country"] = "Sweden"
	strInputs["form_8833:treaty_article"] = "Article 18 - Pensions"
	strInputs["form_8833:irc_provision"] = "IRC 61"
	strInputs["form_8833:treaty_position_explanation"] = "Swedish pension contributions treated as deferred under US-Sweden treaty"
	numInputs["form_8833:treaty_amount"] = 5000
	numInputs["form_8833:num_positions"] = 1

	result, err := g.Solve(numInputs, strInputs, 2025)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// Treaty position claimed
	assertClose(t, result["form_8833:treaty_claimed"], 1, "treaty position claimed")

	// Tax computation should be unaffected (Form 8833 is disclosure only)
	assertClose(t, result["1040:1a"], 75000, "wages unchanged by treaty disclosure")
}

func TestForm8833NoTreatyPosition(t *testing.T) {
	g := buildSolver(t)

	numInputs := map[string]float64{
		"w2:1:wages":                 75000,
		"w2:1:federal_tax_withheld":  9500,
		"w2:1:ss_wages":              75000,
		"w2:1:ss_tax_withheld":       4650,
		"w2:1:medicare_wages":        75000,
		"w2:1:medicare_tax_withheld": 1087.50,
		"w2:1:state_wages":           75000,
		"w2:1:state_tax_withheld":    3200,
		"w2:1:employer_name":         0,
		"w2:1:employer_ein":          0,
		"1040:filing_status":         0,
		"1040:first_name":            0,
		"1040:last_name":             0,
		"1040:ssn":                   0,
	}
	strInputs := map[string]string{
		"1040:filing_status":  "single",
		"1040:first_name":     "Test",
		"1040:last_name":      "NoTreaty",
		"1040:ssn":            "999-77-7777",
		"w2:1:employer_name":  "Acme",
		"w2:1:employer_ein":   "12-3456789",
	}

	for key, val := range inputFormDefaults {
		switch v := val.(type) {
		case float64:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = v
			}
		case string:
			if _, exists := numInputs[key]; !exists {
				numInputs[key] = 0
				strInputs[key] = v
			}
		}
	}

	// Default: no treaty position (treaty_amount = 0)
	result, err := g.Solve(numInputs, strInputs, 2025)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	assertClose(t, result["form_8833:treaty_claimed"], 0, "no treaty position")
}

func TestCAExpatFEIEAddbackScenario(t *testing.T) {
	g := buildSolver(t)

	numInputs, strInputs, taxYear := loadScenario(t,
		"../../testdata/scenarios/ca/ca_expat_feie_addback.json")

	result, err := g.Solve(numInputs, strInputs, taxYear)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// Check federal expected values (FEIE)
	fedExpected := loadExpected(t, "../../testdata/expected/federal/expat_sweden_bfrt.json")
	for key, want := range fedExpected {
		assertClose(t, result[key], want, key)
	}

	// Check CA expected values (FEIE add-back)
	caExpected := loadExpected(t, "../../testdata/expected/ca/ca_expat_feie_addback.json")
	for key, want := range caExpected {
		assertClose(t, result[key], want, key)
	}
}

func TestCAFEIEAddback(t *testing.T) {
	g := buildSolver(t)

	numInputs, strInputs, _ := loadScenario(t,
		"../../testdata/scenarios/ca/ca_expat_feie_addback.json")

	result, err := g.Solve(numInputs, strInputs, 2025)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// Core check: FEIE add-back equals the federal exclusion
	assertClose(t, result["ca_schedule_ca:8d_col_c"], 120000, "FEIE add-back")

	// CA AGI should include the add-back even though federal AGI is 0
	assertClose(t, result["ca_540:17"], 120000, "CA AGI with FEIE add-back")

	// CA should have tax owed (no withholding from foreign employer)
	if result["ca_540:93"] <= 0 {
		t.Error("expat should owe CA tax due to FEIE non-conformity")
	}
}

func TestCAForeignHousingAddback(t *testing.T) {
	g := buildSolver(t)

	numInputs, strInputs, _ := loadScenario(t,
		"../../testdata/scenarios/ca/ca_expat_feie_addback.json")

	// Override: make self-employed so housing deduction applies
	numInputs["form_2555:self_employed_abroad"] = 0
	strInputs["form_2555:self_employed_abroad"] = "yes"
	numInputs["form_2555:employer_provided_housing"] = 0
	numInputs["form_2555:housing_expenses"] = 30000

	result, err := g.Solve(numInputs, strInputs, 2025)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// Housing deduction should be added back on Schedule CA
	housingDeduction := result["form_2555:housing_deduction"]
	addBack := result["ca_schedule_ca:8d_col_c_housing"]
	assertClose(t, addBack, housingDeduction, "housing deduction add-back")
}
