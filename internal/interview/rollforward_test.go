package interview

import (
	"testing"

	"taxpilot/internal/state"
)

func TestNewRollforward_BasicSingleFiler(t *testing.T) {
	registry := SetupRegistry()

	// Simulate a completed 2025 return (single filer, $75K wages)
	prior := state.NewTaxReturn(2025, "CA")
	prior.Inputs = map[string]float64{
		"w2:1:wages":                  75000,
		"w2:1:fed_withholding":        8500,
		"w2:1:state_withholding":      3200,
		"w2:1:social_security":        4650,
		"w2:1:medicare":               1087.50,
		"form_8889:2":                 0,
		"schedule_a:1":                0,
		"schedule_a:5a":               0,
		"schedule_a:5c":               0,
		"schedule_a:5d":               0,
		"schedule_a:8a":               0,
		"schedule_a:12":               0,
		"schedule_a:15":               0,
		"schedule_a:16":               0,
		"schedule_c:1":                0,
		"schedule_c:8":                0,
		"schedule_c:9":                0,
		"schedule_c:10":               0,
		"schedule_c:11":               0,
		"schedule_c:17":               0,
		"schedule_c:18":               0,
		"schedule_c:20":               0,
		"schedule_c:21":               0,
		"schedule_c:22":               0,
		"schedule_c:23":               0,
		"schedule_c:24":               0,
		"schedule_c:25":               0,
		"schedule_c:26":               0,
		"schedule_c:27":               0,
		"form_8949:st_proceeds":       0,
		"form_8949:st_basis":          0,
		"form_8949:lt_proceeds":       0,
		"form_8949:lt_basis":          0,
		"1099_int:1:interest":         0,
		"1099_div:1:ordinary_div":     0,
		"1099_div:1:qualified_div":    0,
		"1099_nec:1:nonemployee":      0,
		"schedule_b:foreign_interest": 0,
		"schedule_b:foreign_accounts": 0,
	}
	prior.StrInputs = map[string]string{
		"1040:filing_status": "single",
		"1040:first_name":    "John",
		"1040:last_name":     "Doe",
		"1040:ssn":           "123-45-6789",
		"w2:1:employer_name": "Acme Corp",
		"w2:1:employer_ein":  "12-3456789",
	}
	prior.Complete = true

	// Create rollforward to 2026
	rf, err := NewRollforward(registry, 2026, prior)
	if err != nil {
		t.Fatalf("NewRollforward: %v", err)
	}

	// Verify basic properties
	if rf.TaxYear != 2026 {
		t.Errorf("TaxYear = %d, want 2026", rf.TaxYear)
	}
	if rf.PriorYear != 2025 {
		t.Errorf("PriorYear = %d, want 2025", rf.PriorYear)
	}

	// Verify inputs were copied
	if rf.Inputs["w2:1:wages"] != 75000 {
		t.Errorf("wages = %v, want 75000", rf.Inputs["w2:1:wages"])
	}
	if rf.StrInputs["1040:first_name"] != "John" {
		t.Errorf("first_name = %q, want John", rf.StrInputs["1040:first_name"])
	}

	// Verify solve produced computed values
	if rf.Computed["1040:11"] == 0 {
		t.Error("AGI (1040:11) should be non-zero after solve")
	}

	// Verify fields were built
	if len(rf.Fields) == 0 {
		t.Error("Fields should not be empty")
	}

	// The standard deduction should differ between 2025 and 2026
	// 2025 single: $15,000, 2026 single: $15,400
	fedDeduction2026 := rf.Computed["1040:12"]
	priorDeduction := rf.PriorComputed["1040:12"]
	if fedDeduction2026 == priorDeduction {
		t.Errorf("Standard deduction should differ between years: 2026=%v, 2025=%v",
			fedDeduction2026, priorDeduction)
	}

	// Should have at least one change flagged (standard deduction, tax, etc.)
	if len(rf.Changes) == 0 {
		t.Error("Expected at least one FieldChange due to year-over-year tax parameter changes")
	}

	// Should have parameter changes
	if len(rf.ParamChanges) == 0 {
		t.Error("Expected parameter changes between 2025 and 2026")
	}

	t.Logf("Fields: %d, Changes: %d, ParamChanges: %d, Flagged: %d",
		len(rf.Fields), len(rf.Changes), len(rf.ParamChanges), rf.CountFlagged())
}

func TestRollforward_UpdateInput(t *testing.T) {
	registry := SetupRegistry()

	prior := state.NewTaxReturn(2025, "CA")
	prior.Inputs = map[string]float64{
		"w2:1:wages":                  75000,
		"w2:1:fed_withholding":        8500,
		"w2:1:state_withholding":      3200,
		"w2:1:social_security":        4650,
		"w2:1:medicare":               1087.50,
		"form_8889:2":                 0,
		"schedule_a:1":                0,
		"schedule_a:5a":               0,
		"schedule_a:5c":               0,
		"schedule_a:5d":               0,
		"schedule_a:8a":               0,
		"schedule_a:12":               0,
		"schedule_a:15":               0,
		"schedule_a:16":               0,
		"schedule_c:1":                0,
		"schedule_c:8":                0,
		"schedule_c:9":                0,
		"schedule_c:10":               0,
		"schedule_c:11":               0,
		"schedule_c:17":               0,
		"schedule_c:18":               0,
		"schedule_c:20":               0,
		"schedule_c:21":               0,
		"schedule_c:22":               0,
		"schedule_c:23":               0,
		"schedule_c:24":               0,
		"schedule_c:25":               0,
		"schedule_c:26":               0,
		"schedule_c:27":               0,
		"form_8949:st_proceeds":       0,
		"form_8949:st_basis":          0,
		"form_8949:lt_proceeds":       0,
		"form_8949:lt_basis":          0,
		"1099_int:1:interest":         0,
		"1099_div:1:ordinary_div":     0,
		"1099_div:1:qualified_div":    0,
		"1099_nec:1:nonemployee":      0,
		"schedule_b:foreign_interest": 0,
		"schedule_b:foreign_accounts": 0,
	}
	prior.StrInputs = map[string]string{
		"1040:filing_status": "single",
		"1040:first_name":    "John",
		"1040:last_name":     "Doe",
		"1040:ssn":           "123-45-6789",
		"w2:1:employer_name": "Acme Corp",
		"w2:1:employer_ein":  "12-3456789",
	}
	prior.Complete = true

	rf, err := NewRollforward(registry, 2026, prior)
	if err != nil {
		t.Fatalf("NewRollforward: %v", err)
	}

	oldAGI := rf.Computed["1040:11"]

	// Update wages to $85,000
	changed, err := rf.UpdateInput("w2:1:wages", 85000)
	if err != nil {
		t.Fatalf("UpdateInput: %v", err)
	}

	newAGI := rf.Computed["1040:11"]

	// AGI should have increased
	if newAGI <= oldAGI {
		t.Errorf("AGI should increase: old=%v, new=%v", oldAGI, newAGI)
	}

	// Changed set should include downstream fields
	if len(changed) == 0 {
		t.Error("Expected changed fields after wage update")
	}

	t.Logf("Wage update: AGI %v -> %v, %d fields changed", oldAGI, newAGI, len(changed))
}

func TestRollforward_2024to2025_Comparison(t *testing.T) {
	registry := SetupRegistry()

	// Simulate a completed 2024 return
	prior := state.NewTaxReturn(2024, "CA")
	prior.Inputs = map[string]float64{
		"w2:1:wages":                  75000,
		"w2:1:fed_withholding":        8500,
		"w2:1:state_withholding":      3200,
		"w2:1:social_security":        4650,
		"w2:1:medicare":               1087.50,
		"form_8889:2":                 0,
		"schedule_a:1":                0,
		"schedule_a:5a":               0,
		"schedule_a:5c":               0,
		"schedule_a:5d":               0,
		"schedule_a:8a":               0,
		"schedule_a:12":               0,
		"schedule_a:15":               0,
		"schedule_a:16":               0,
		"schedule_c:1":                0,
		"schedule_c:8":                0,
		"schedule_c:9":                0,
		"schedule_c:10":               0,
		"schedule_c:11":               0,
		"schedule_c:17":               0,
		"schedule_c:18":               0,
		"schedule_c:20":               0,
		"schedule_c:21":               0,
		"schedule_c:22":               0,
		"schedule_c:23":               0,
		"schedule_c:24":               0,
		"schedule_c:25":               0,
		"schedule_c:26":               0,
		"schedule_c:27":               0,
		"form_8949:st_proceeds":       0,
		"form_8949:st_basis":          0,
		"form_8949:lt_proceeds":       0,
		"form_8949:lt_basis":          0,
		"1099_int:1:interest":         0,
		"1099_div:1:ordinary_div":     0,
		"1099_div:1:qualified_div":    0,
		"1099_nec:1:nonemployee":      0,
		"schedule_b:foreign_interest": 0,
		"schedule_b:foreign_accounts": 0,
	}
	prior.StrInputs = map[string]string{
		"1040:filing_status": "single",
		"1040:first_name":    "Jane",
		"1040:last_name":     "Smith",
		"1040:ssn":           "987-65-4321",
		"w2:1:employer_name": "Acme Corp",
		"w2:1:employer_ein":  "12-3456789",
	}
	prior.Complete = true

	rf, err := NewRollforward(registry, 2025, prior)
	if err != nil {
		t.Fatalf("NewRollforward 2024->2025: %v", err)
	}

	// Prior year computed values should be non-zero (2024 tables now exist)
	if rf.PriorComputed["1040:11"] == 0 {
		t.Error("PriorComputed AGI should be non-zero with 2024 tables")
	}

	// Standard deduction should differ: 2024=$14,600, 2025=$15,000
	deduction2025 := rf.Computed["1040:12"]
	deduction2024 := rf.PriorComputed["1040:12"]
	if deduction2024 == 0 {
		t.Fatal("Prior year deduction should be non-zero")
	}
	if deduction2025 <= deduction2024 {
		t.Errorf("2025 deduction ($%.0f) should be > 2024 deduction ($%.0f)",
			deduction2025, deduction2024)
	}

	// Tax should be slightly lower in 2025 (higher deduction + wider brackets)
	tax2025 := rf.Computed["1040:16"]
	tax2024 := rf.PriorComputed["1040:16"]
	if tax2024 == 0 {
		t.Fatal("Prior year tax should be non-zero")
	}
	if tax2025 >= tax2024 {
		t.Errorf("2025 tax ($%.2f) should be < 2024 tax ($%.2f) for same income",
			tax2025, tax2024)
	}

	// Changes should include deduction and tax differences
	if len(rf.Changes) == 0 {
		t.Error("Expected changes between 2024 and 2025")
	}

	// Parameter changes should show deduction increase
	foundDeduction := false
	for _, pc := range rf.ParamChanges {
		if pc.Category == "deduction" && pc.Delta > 0 {
			foundDeduction = true
		}
	}
	if !foundDeduction {
		t.Error("Expected a deduction parameter change between 2024 and 2025")
	}

	t.Logf("2024->2025: Deduction $%.0f->$%.0f, Tax $%.2f->$%.2f, Changes: %d, ParamChanges: %d",
		deduction2024, deduction2025, tax2024, tax2025, len(rf.Changes), len(rf.ParamChanges))
}
