package forms_test

import (
	"encoding/json"
	"math"
	"os"
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

// loadScenario reads a scenario file and splits its inputs into numeric and
// string maps suitable for the solver. W-2 numeric inputs are re-keyed with
// a "1:" employer prefix so that the 1040's wildcard patterns (w2:*:field)
// can match them.
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

	for key, val := range s.Inputs {
		// Re-key w2 fields: "w2:wages" -> "w2:1:wages" so that the
		// 1040 wildcard "w2:*:wages" pattern matches.
		solverKey := key
		if len(key) > 3 && key[:3] == "w2:" {
			solverKey = "w2:1:" + key[3:]
		}

		switch v := val.(type) {
		case float64:
			numInputs[solverKey] = v
		case string:
			// String-valued fields still need a zero placeholder in the
			// numeric map so MissingInputs doesn't reject them.
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
	reg.Register(federal.F1040())
	reg.Register(ca.F540())
	reg.Register(ca.ScheduleCA())

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
		"1040:filing_status":         0,
		"1040:first_name":            0,
		"1040:last_name":             0,
		"1040:ssn":                   0,
	}
	strInputs := map[string]string{
		"1040:filing_status": "single",
		"1040:first_name":    "Jane",
		"1040:last_name":     "Doe",
		"1040:ssn":           "123-45-6789",
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
