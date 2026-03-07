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
