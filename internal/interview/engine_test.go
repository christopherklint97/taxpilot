package interview

import (
	"testing"

	"taxpilot/internal/forms"
)

// testRegistry creates a minimal registry with a single form containing
// a few UserInput fields for testing.
func testRegistry() *forms.Registry {
	reg := forms.NewRegistry()
	reg.Register(&forms.FormDef{
		ID:           "test_form",
		Name:         "Test Form",
		Jurisdiction: forms.Federal,
		TaxYears:     []int{2025},
		Fields: []forms.FieldDef{
			{
				Line:    "filing_status",
				Type:    forms.UserInput,
				Label:   "Filing Status",
				Prompt:  "Select your filing status:",
				Options: []string{"single", "mfj", "mfs", "hoh", "qss"},
			},
			{
				Line:   "first_name",
				Type:   forms.UserInput,
				Label:  "First Name",
				Prompt: "What is your first name?",
			},
			{
				Line:   "wages",
				Type:   forms.UserInput,
				Label:  "Wages",
				Prompt: "Enter your total wages:",
			},
			{
				Line:      "tax",
				Type:      forms.Computed,
				Label:     "Tax",
				DependsOn: []string{"test_form:wages"},
				Compute: func(d forms.DepValues) float64 {
					return d.Get("test_form:wages") * 0.10
				},
			},
		},
	})
	return reg
}

// TestEngineWithPriorYear verifies prior-year defaults are available.
func TestEngineWithPriorYear(t *testing.T) {
	reg := testRegistry()
	priorNumeric := map[string]float64{
		"test_form:wages": 75000,
	}
	priorStr := map[string]string{
		"test_form:filing_status": "single",
		"test_form:first_name":    "Jane",
	}

	eng, err := NewEngineWithPriorYear(reg, 2025, priorNumeric, priorStr, "CA")
	if err != nil {
		t.Fatalf("NewEngineWithPriorYear failed: %v", err)
	}

	// The first question should be filing_status, and it should have a prior-year default.
	q := eng.Current()
	if q == nil {
		t.Fatal("expected a current question, got nil")
	}
	if q.Key != "test_form:filing_status" {
		t.Fatalf("expected first question to be filing_status, got %s", q.Key)
	}

	pyd := eng.GetPriorYearDefault()
	if pyd == nil {
		t.Fatal("expected prior-year default for filing_status, got nil")
	}
	if pyd.StrValue != "single" {
		t.Fatalf("expected prior-year default 'single', got %q", pyd.StrValue)
	}
}

// TestAcceptDefault verifies accepting a default stores the value correctly.
func TestAcceptDefault(t *testing.T) {
	reg := testRegistry()
	priorNumeric := map[string]float64{
		"test_form:wages": 75000,
	}
	priorStr := map[string]string{
		"test_form:filing_status": "single",
		"test_form:first_name":    "Jane",
	}

	eng, err := NewEngineWithPriorYear(reg, 2025, priorNumeric, priorStr, "CA")
	if err != nil {
		t.Fatalf("NewEngineWithPriorYear failed: %v", err)
	}

	// Accept default for filing_status
	if err := eng.AcceptDefault(); err != nil {
		t.Fatalf("AcceptDefault failed: %v", err)
	}

	// Check that filing_status was stored
	strInputs := eng.StrInputs()
	if strInputs["test_form:filing_status"] != "single" {
		t.Fatalf("expected filing_status='single', got %q", strInputs["test_form:filing_status"])
	}

	// Next question should be first_name
	q := eng.Current()
	if q == nil {
		t.Fatal("expected a current question, got nil")
	}
	if q.Key != "test_form:first_name" {
		t.Fatalf("expected next question to be first_name, got %s", q.Key)
	}

	// Accept default for first_name
	pyd := eng.GetPriorYearDefault()
	if pyd == nil {
		t.Fatal("expected prior-year default for first_name, got nil")
	}
	if pyd.StrValue != "Jane" {
		t.Fatalf("expected prior-year default 'Jane', got %q", pyd.StrValue)
	}
	if err := eng.AcceptDefault(); err != nil {
		t.Fatalf("AcceptDefault for first_name failed: %v", err)
	}
	strInputs = eng.StrInputs()
	if strInputs["test_form:first_name"] != "Jane" {
		t.Fatalf("expected first_name='Jane', got %q", strInputs["test_form:first_name"])
	}

	// Next question should be wages
	q = eng.Current()
	if q == nil {
		t.Fatal("expected a current question, got nil")
	}
	if q.Key != "test_form:wages" {
		t.Fatalf("expected next question to be wages, got %s", q.Key)
	}

	// Accept default for wages (numeric)
	pyd = eng.GetPriorYearDefault()
	if pyd == nil {
		t.Fatal("expected prior-year default for wages, got nil")
	}
	if pyd.RawValue != 75000 {
		t.Fatalf("expected prior-year default 75000, got %f", pyd.RawValue)
	}
	if err := eng.AcceptDefault(); err != nil {
		t.Fatalf("AcceptDefault for wages failed: %v", err)
	}
	numInputs := eng.Inputs()
	if numInputs["test_form:wages"] != 75000 {
		t.Fatalf("expected wages=75000, got %f", numInputs["test_form:wages"])
	}

	// All questions should be done
	if eng.HasNext() {
		t.Fatal("expected no more questions, but HasNext is true")
	}
}

// TestOverrideDefault verifies typing a new value overrides the default.
func TestOverrideDefault(t *testing.T) {
	reg := testRegistry()
	priorNumeric := map[string]float64{
		"test_form:wages": 75000,
	}
	priorStr := map[string]string{
		"test_form:filing_status": "single",
		"test_form:first_name":    "Jane",
	}

	eng, err := NewEngineWithPriorYear(reg, 2025, priorNumeric, priorStr, "CA")
	if err != nil {
		t.Fatalf("NewEngineWithPriorYear failed: %v", err)
	}

	// Override filing_status with a different value
	if err := eng.Answer("mfj"); err != nil {
		t.Fatalf("Answer failed: %v", err)
	}
	strInputs := eng.StrInputs()
	if strInputs["test_form:filing_status"] != "mfj" {
		t.Fatalf("expected filing_status='mfj', got %q", strInputs["test_form:filing_status"])
	}

	// Override first_name
	if err := eng.Answer("John"); err != nil {
		t.Fatalf("Answer failed: %v", err)
	}
	strInputs = eng.StrInputs()
	if strInputs["test_form:first_name"] != "John" {
		t.Fatalf("expected first_name='John', got %q", strInputs["test_form:first_name"])
	}

	// Override wages with a new numeric value
	if err := eng.Answer("85000"); err != nil {
		t.Fatalf("Answer failed: %v", err)
	}
	numInputs := eng.Inputs()
	if numInputs["test_form:wages"] != 85000 {
		t.Fatalf("expected wages=85000, got %f", numInputs["test_form:wages"])
	}
}

// TestNoPriorYear verifies engine works normally without prior-year data.
func TestNoPriorYear(t *testing.T) {
	reg := testRegistry()

	eng, err := NewEngine(reg, 2025)
	if err != nil {
		t.Fatalf("NewEngine failed: %v", err)
	}

	// No prior-year default should be available
	pyd := eng.GetPriorYearDefault()
	if pyd != nil {
		t.Fatal("expected no prior-year default, got one")
	}

	// AcceptDefault should fail
	if err := eng.AcceptDefault(); err == nil {
		t.Fatal("expected AcceptDefault to fail without prior-year data")
	}

	// Normal answer flow should work
	if err := eng.Answer("single"); err != nil {
		t.Fatalf("Answer failed: %v", err)
	}
	if err := eng.Answer("Jane"); err != nil {
		t.Fatalf("Answer failed: %v", err)
	}
	if err := eng.Answer("75000"); err != nil {
		t.Fatalf("Answer failed: %v", err)
	}

	if eng.HasNext() {
		t.Fatal("expected no more questions")
	}
}

// TestEngineWithPriorYear_EmptyMaps verifies engine handles empty prior-year maps.
func TestEngineWithPriorYear_EmptyMaps(t *testing.T) {
	reg := testRegistry()

	eng, err := NewEngineWithPriorYear(reg, 2025, nil, nil, "")
	if err != nil {
		t.Fatalf("NewEngineWithPriorYear with nil maps failed: %v", err)
	}

	// No prior-year defaults should be available
	pyd := eng.GetPriorYearDefault()
	if pyd != nil {
		t.Fatal("expected no prior-year default with empty maps, got one")
	}

	// Normal flow should still work
	if err := eng.Answer("single"); err != nil {
		t.Fatalf("Answer failed: %v", err)
	}
}

// TestFormatCurrency verifies the engine-internal currency formatter.
func TestFormatCurrency(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{75000, "$75,000.00"},
		{0, "$0.00"},
		{1234.56, "$1,234.56"},
		{-500, "-$500.00"},
		{999999.99, "$999,999.99"},
	}

	for _, tc := range tests {
		got := formatCurrency(tc.input)
		if got != tc.expected {
			t.Errorf("formatCurrency(%f) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}
