package forms

import (
	"math"
	"strings"
	"testing"
)

// helper to check float equality within tolerance
func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.01
}

func TestBasicComputationChain(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&FormDef{
		ID:           "1040",
		Name:         "Form 1040",
		Jurisdiction: Federal,
		TaxYears:     []int{2025},
		Fields: []FieldDef{
			{Line: "1", Type: UserInput, Label: "Wages", Prompt: "Enter your wages"},
			{Line: "2", Type: UserInput, Label: "Interest", Prompt: "Enter interest income"},
			{Line: "3", Type: Computed, Label: "Total Income",
				DependsOn: []string{"1040:1", "1040:2"},
				Compute: func(d DepValues) float64 {
					return d.Get("1040:1") + d.Get("1040:2")
				},
			},
		},
	})

	g := NewDependencyGraph(reg)
	if err := g.Build(); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	inputs := map[string]float64{
		"1040:1": 50000,
		"1040:2": 1200,
	}

	result, err := g.Solve(inputs, nil, 2025)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	if !approxEqual(result["1040:3"], 51200) {
		t.Errorf("expected 1040:3 = 51200, got %v", result["1040:3"])
	}
}

func TestLongerChain(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&FormDef{
		ID:           "1040",
		Name:         "Form 1040",
		Jurisdiction: Federal,
		TaxYears:     []int{2025},
		Fields: []FieldDef{
			{Line: "1", Type: UserInput, Label: "Wages"},
			{Line: "2", Type: UserInput, Label: "Interest"},
			{Line: "3", Type: Computed, Label: "Total Income",
				DependsOn: []string{"1040:1", "1040:2"},
				Compute: func(d DepValues) float64 {
					return d.Get("1040:1") + d.Get("1040:2")
				},
			},
			{Line: "4", Type: Computed, Label: "Deduction",
				DependsOn: []string{"1040:3"},
				Compute: func(d DepValues) float64 {
					return d.Get("1040:3") * 0.1
				},
			},
			{Line: "5", Type: Computed, Label: "Taxable Income",
				DependsOn: []string{"1040:3", "1040:4"},
				Compute: func(d DepValues) float64 {
					return d.Get("1040:3") - d.Get("1040:4")
				},
			},
		},
	})

	g := NewDependencyGraph(reg)
	if err := g.Build(); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	inputs := map[string]float64{
		"1040:1": 100000,
		"1040:2": 5000,
	}

	result, err := g.Solve(inputs, nil, 2025)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	if !approxEqual(result["1040:3"], 105000) {
		t.Errorf("expected 1040:3 = 105000, got %v", result["1040:3"])
	}
	if !approxEqual(result["1040:4"], 10500) {
		t.Errorf("expected 1040:4 = 10500, got %v", result["1040:4"])
	}
	if !approxEqual(result["1040:5"], 94500) {
		t.Errorf("expected 1040:5 = 94500, got %v", result["1040:5"])
	}
}

func TestCrossFormDependencies(t *testing.T) {
	reg := NewRegistry()

	// Federal form
	reg.Register(&FormDef{
		ID:           "1040",
		Name:         "Form 1040",
		Jurisdiction: Federal,
		TaxYears:     []int{2025},
		Fields: []FieldDef{
			{Line: "1", Type: UserInput, Label: "Wages"},
			{Line: "11", Type: Computed, Label: "AGI",
				DependsOn: []string{"1040:1"},
				Compute: func(d DepValues) float64 {
					return d.Get("1040:1")
				},
			},
		},
	})

	// California state form referencing federal AGI
	reg.Register(&FormDef{
		ID:           "ca_540",
		Name:         "CA 540",
		Jurisdiction: StateCA,
		TaxYears:     []int{2025},
		Fields: []FieldDef{
			{Line: "13", Type: FederalRef, Label: "Federal AGI",
				DependsOn: []string{"1040:11"},
				Compute: func(d DepValues) float64 {
					return d.Get("1040:11")
				},
			},
			{Line: "14", Type: UserInput, Label: "CA Adjustments"},
			{Line: "15", Type: Computed, Label: "CA Taxable Income",
				DependsOn: []string{"ca_540:13", "ca_540:14"},
				Compute: func(d DepValues) float64 {
					return d.Get("ca_540:13") - d.Get("ca_540:14")
				},
			},
		},
	})

	g := NewDependencyGraph(reg)
	if err := g.Build(); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	inputs := map[string]float64{
		"1040:1":   80000,
		"ca_540:14": 2000,
	}

	result, err := g.Solve(inputs, nil, 2025)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	if !approxEqual(result["1040:11"], 80000) {
		t.Errorf("expected 1040:11 = 80000, got %v", result["1040:11"])
	}
	if !approxEqual(result["ca_540:13"], 80000) {
		t.Errorf("expected ca_540:13 = 80000, got %v", result["ca_540:13"])
	}
	if !approxEqual(result["ca_540:15"], 78000) {
		t.Errorf("expected ca_540:15 = 78000, got %v", result["ca_540:15"])
	}
}

func TestCycleDetection(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&FormDef{
		ID:           "cycle",
		Name:         "Cycle Form",
		Jurisdiction: Federal,
		TaxYears:     []int{2025},
		Fields: []FieldDef{
			{Line: "a", Type: Computed, Label: "A",
				DependsOn: []string{"cycle:c"},
				Compute:   func(d DepValues) float64 { return d.Get("cycle:c") },
			},
			{Line: "b", Type: Computed, Label: "B",
				DependsOn: []string{"cycle:a"},
				Compute:   func(d DepValues) float64 { return d.Get("cycle:a") },
			},
			{Line: "c", Type: Computed, Label: "C",
				DependsOn: []string{"cycle:b"},
				Compute:   func(d DepValues) float64 { return d.Get("cycle:b") },
			},
		},
	})

	g := NewDependencyGraph(reg)
	err := g.Build()
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}
	if !strings.Contains(err.Error(), "cycle detected") {
		t.Errorf("expected 'cycle detected' in error, got: %v", err)
	}
}

func TestMissingInputDetection(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&FormDef{
		ID:           "1040",
		Name:         "Form 1040",
		Jurisdiction: Federal,
		TaxYears:     []int{2025},
		Fields: []FieldDef{
			{Line: "1", Type: UserInput, Label: "Wages"},
			{Line: "2", Type: UserInput, Label: "Interest"},
			{Line: "3", Type: Computed, Label: "Total",
				DependsOn: []string{"1040:1", "1040:2"},
				Compute: func(d DepValues) float64 {
					return d.Get("1040:1") + d.Get("1040:2")
				},
			},
		},
	})

	g := NewDependencyGraph(reg)
	if err := g.Build(); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Only provide one of two required inputs
	missing := g.MissingInputs(map[string]float64{"1040:1": 50000})
	if len(missing) != 1 || missing[0] != "1040:2" {
		t.Errorf("expected missing [1040:2], got %v", missing)
	}

	// Solve should fail with missing inputs
	_, err := g.Solve(map[string]float64{"1040:1": 50000}, nil, 2025)
	if err == nil {
		t.Fatal("expected error for missing inputs, got nil")
	}
	if !strings.Contains(err.Error(), "1040:2") {
		t.Errorf("expected error to mention 1040:2, got: %v", err)
	}
}

func TestWildcardSumAll(t *testing.T) {
	reg := NewRegistry()

	// Two W-2 forms with wildcard-matchable keys
	reg.Register(&FormDef{
		ID:       "w2",
		Name:     "W-2",
		TaxYears: []int{2025},
		Fields: []FieldDef{
			{Line: "employer1:wages", Type: UserInput, Label: "W2 Employer 1 Wages"},
			{Line: "employer2:wages", Type: UserInput, Label: "W2 Employer 2 Wages"},
			{Line: "employer1:withholding", Type: UserInput, Label: "W2 Employer 1 Withholding"},
		},
	})

	reg.Register(&FormDef{
		ID:       "1040",
		Name:     "Form 1040",
		TaxYears: []int{2025},
		Fields: []FieldDef{
			{Line: "1", Type: Computed, Label: "Total Wages",
				DependsOn: []string{"w2:*:wages"},
				Compute: func(d DepValues) float64 {
					return d.SumAll("w2:*:wages")
				},
			},
		},
	})

	g := NewDependencyGraph(reg)
	if err := g.Build(); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	inputs := map[string]float64{
		"w2:employer1:wages":       60000,
		"w2:employer2:wages":       25000,
		"w2:employer1:withholding": 12000,
	}

	result, err := g.Solve(inputs, nil, 2025)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	if !approxEqual(result["1040:1"], 85000) {
		t.Errorf("expected 1040:1 = 85000, got %v", result["1040:1"])
	}
}

func TestFederalRefResolvesCorrectly(t *testing.T) {
	reg := NewRegistry()

	reg.Register(&FormDef{
		ID:           "1040",
		Name:         "Form 1040",
		Jurisdiction: Federal,
		TaxYears:     []int{2025},
		Fields: []FieldDef{
			{Line: "1", Type: UserInput, Label: "Wages"},
			{Line: "11", Type: Computed, Label: "AGI",
				DependsOn: []string{"1040:1"},
				Compute:   func(d DepValues) float64 { return d.Get("1040:1") },
			},
			{Line: "15", Type: Computed, Label: "Taxable Income",
				DependsOn: []string{"1040:11"},
				Compute: func(d DepValues) float64 {
					agi := d.Get("1040:11")
					standardDeduction := 14600.0
					if agi > standardDeduction {
						return agi - standardDeduction
					}
					return 0
				},
			},
		},
	})

	reg.Register(&FormDef{
		ID:           "ca_540",
		Name:         "CA 540",
		Jurisdiction: StateCA,
		TaxYears:     []int{2025},
		Fields: []FieldDef{
			{Line: "13", Type: FederalRef, Label: "Federal AGI",
				DependsOn: []string{"1040:11"},
				Compute:   func(d DepValues) float64 { return d.Get("1040:11") },
			},
			{Line: "18", Type: Computed, Label: "CA Tax",
				DependsOn: []string{"ca_540:13"},
				Compute: func(d DepValues) float64 {
					// Simplified CA tax: 9.3% flat
					return d.Get("ca_540:13") * 0.093
				},
			},
		},
	})

	g := NewDependencyGraph(reg)
	if err := g.Build(); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	inputs := map[string]float64{
		"1040:1": 100000,
	}

	result, err := g.Solve(inputs, nil, 2025)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// Federal AGI
	if !approxEqual(result["1040:11"], 100000) {
		t.Errorf("expected 1040:11 = 100000, got %v", result["1040:11"])
	}
	// Federal taxable income
	if !approxEqual(result["1040:15"], 85400) {
		t.Errorf("expected 1040:15 = 85400, got %v", result["1040:15"])
	}
	// CA 540 line 13 should match federal AGI
	if !approxEqual(result["ca_540:13"], 100000) {
		t.Errorf("expected ca_540:13 = 100000, got %v", result["ca_540:13"])
	}
	// CA tax
	if !approxEqual(result["ca_540:18"], 9300) {
		t.Errorf("expected ca_540:18 = 9300, got %v", result["ca_540:18"])
	}
}

func TestTopologicalSortOrder(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&FormDef{
		ID:       "f",
		Name:     "Test Form",
		TaxYears: []int{2025},
		Fields: []FieldDef{
			{Line: "1", Type: UserInput, Label: "Input"},
			{Line: "2", Type: Computed, Label: "Computed1",
				DependsOn: []string{"f:1"},
				Compute:   func(d DepValues) float64 { return d.Get("f:1") * 2 },
			},
			{Line: "3", Type: Computed, Label: "Computed2",
				DependsOn: []string{"f:2"},
				Compute:   func(d DepValues) float64 { return d.Get("f:2") + 10 },
			},
		},
	})

	g := NewDependencyGraph(reg)
	if err := g.Build(); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	order, err := g.TopologicalSort()
	if err != nil {
		t.Fatalf("TopologicalSort failed: %v", err)
	}

	// Verify f:1 comes before f:2, and f:2 comes before f:3
	indexOf := func(key string) int {
		for i, k := range order {
			if k == key {
				return i
			}
		}
		return -1
	}

	if indexOf("f:1") >= indexOf("f:2") {
		t.Errorf("f:1 should come before f:2 in topological order")
	}
	if indexOf("f:2") >= indexOf("f:3") {
		t.Errorf("f:2 should come before f:3 in topological order")
	}
}

func TestFieldKey(t *testing.T) {
	key := FieldKey("1040", "11")
	if key != "1040:11" {
		t.Errorf("expected '1040:11', got %q", key)
	}
}

func TestRegistryGetField(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&FormDef{
		ID:   "1040",
		Name: "Form 1040",
		Fields: []FieldDef{
			{Line: "1", Type: UserInput, Label: "Wages"},
		},
	})

	form, field, err := reg.GetField("1040:1")
	if err != nil {
		t.Fatalf("GetField failed: %v", err)
	}
	if form.ID != "1040" {
		t.Errorf("expected form ID '1040', got %q", form.ID)
	}
	if field.Label != "Wages" {
		t.Errorf("expected label 'Wages', got %q", field.Label)
	}

	_, _, err = reg.GetField("9999:1")
	if err == nil {
		t.Error("expected error for non-existent form")
	}

	_, _, err = reg.GetField("1040:99")
	if err == nil {
		t.Error("expected error for non-existent field")
	}

	_, _, err = reg.GetField("badkey")
	if err == nil {
		t.Error("expected error for invalid key format")
	}
}

func TestSumAllWildcardPatterns(t *testing.T) {
	vals := map[string]float64{
		"w2:emp1:wages":  50000,
		"w2:emp2:wages":  30000,
		"w2:emp1:tips":   1000,
		"w2:emp2:tips":   500,
		"1099:bank:interest": 200,
	}
	dv := NewDepValues(vals, nil, 2025)

	// Sum all wages
	wages := dv.SumAll("w2:*:wages")
	if !approxEqual(wages, 80000) {
		t.Errorf("expected 80000, got %v", wages)
	}

	// Sum all tips
	tips := dv.SumAll("w2:*:tips")
	if !approxEqual(tips, 1500) {
		t.Errorf("expected 1500, got %v", tips)
	}

	// Sum all w2 emp1 fields
	emp1 := dv.SumAll("w2:emp1:*")
	if !approxEqual(emp1, 51000) {
		t.Errorf("expected 51000, got %v", emp1)
	}

	// No match
	none := dv.SumAll("w2:*:bonuses")
	if !approxEqual(none, 0) {
		t.Errorf("expected 0, got %v", none)
	}
}

func TestDepValuesTaxYear(t *testing.T) {
	dv := NewDepValues(nil, nil, 2025)
	if dv.TaxYear() != 2025 {
		t.Errorf("expected tax year 2025, got %d", dv.TaxYear())
	}
}

func TestStringComputedFields(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&FormDef{
		ID:       "1040",
		Name:     "Form 1040",
		TaxYears: []int{2025},
		Fields: []FieldDef{
			{Line: "status", Type: UserInput, Label: "Filing Status"},
			{Line: "1", Type: UserInput, Label: "Wages"},
			{Line: "greeting", Type: Computed, Label: "Greeting",
				DependsOn: []string{"1040:status"},
				ComputeStr: func(d DepValues) string {
					return "Filing as: " + d.GetString("1040:status")
				},
			},
		},
	})

	g := NewDependencyGraph(reg)
	if err := g.Build(); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	inputs := map[string]float64{
		"1040:status": 0,
		"1040:1":      50000,
	}
	strInputs := map[string]string{
		"1040:status": "single",
	}

	result, err := g.Solve(inputs, strInputs, 2025)
	if err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// Numeric results should still work
	if !approxEqual(result["1040:1"], 50000) {
		t.Errorf("expected 1040:1 = 50000, got %v", result["1040:1"])
	}
}
