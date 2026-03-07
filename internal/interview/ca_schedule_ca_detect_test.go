package interview

import (
	"testing"
)

func TestDetectCANoAdjustments(t *testing.T) {
	inputs := map[string]float64{
		"w2:1:wages":       50000,
		"w2:1:state_wages": 50000,
	}
	strInputs := map[string]string{}

	needed, reasons := DetectCAScheduleCANeeded(inputs, strInputs)
	if needed {
		t.Errorf("expected no adjustments needed, got reasons: %v", reasons)
	}
	if len(reasons) != 0 {
		t.Errorf("expected 0 reasons, got %d: %v", len(reasons), reasons)
	}
}

func TestDetectCAEmptyInputs(t *testing.T) {
	inputs := map[string]float64{}
	strInputs := map[string]string{}

	needed, reasons := DetectCAScheduleCANeeded(inputs, strInputs)
	if needed {
		t.Errorf("expected no adjustments for empty inputs, got reasons: %v", reasons)
	}
}

func TestDetectCAHSA(t *testing.T) {
	inputs := map[string]float64{
		"form_8889:2": 3500,
	}
	strInputs := map[string]string{}

	needed, reasons := DetectCAScheduleCANeeded(inputs, strInputs)
	if !needed {
		t.Fatal("expected Schedule CA needed for HSA contributions")
	}
	if len(reasons) != 1 {
		t.Fatalf("expected 1 reason, got %d: %v", len(reasons), reasons)
	}
	if reasons[0] != "HSA deduction add-back" {
		t.Errorf("expected 'HSA deduction add-back', got %q", reasons[0])
	}
}

func TestDetectCASALT(t *testing.T) {
	inputs := map[string]float64{
		"schedule_a:5a": 10000,
	}
	strInputs := map[string]string{}

	needed, reasons := DetectCAScheduleCANeeded(inputs, strInputs)
	if !needed {
		t.Fatal("expected Schedule CA needed for SALT deduction")
	}
	if len(reasons) != 1 {
		t.Fatalf("expected 1 reason, got %d: %v", len(reasons), reasons)
	}
	if reasons[0] != "State income tax deduction removal" {
		t.Errorf("expected 'State income tax deduction removal', got %q", reasons[0])
	}
}

func TestDetectCAQBI(t *testing.T) {
	inputs := map[string]float64{
		"schedule_c:31": 75000,
	}
	strInputs := map[string]string{}

	needed, reasons := DetectCAScheduleCANeeded(inputs, strInputs)
	if !needed {
		t.Fatal("expected Schedule CA needed for QBI")
	}
	if len(reasons) != 1 {
		t.Fatalf("expected 1 reason, got %d: %v", len(reasons), reasons)
	}
	if reasons[0] != "QBI deduction add-back" {
		t.Errorf("expected 'QBI deduction add-back', got %q", reasons[0])
	}
}

func TestDetectCAUSBondInterest(t *testing.T) {
	inputs := map[string]float64{
		"1099int:1:us_savings_bond_interest": 500,
	}
	strInputs := map[string]string{}

	needed, reasons := DetectCAScheduleCANeeded(inputs, strInputs)
	if !needed {
		t.Fatal("expected Schedule CA needed for US bond interest")
	}
	if len(reasons) != 1 {
		t.Fatalf("expected 1 reason, got %d: %v", len(reasons), reasons)
	}
	if reasons[0] != "U.S. bond interest subtraction" {
		t.Errorf("expected 'U.S. bond interest subtraction', got %q", reasons[0])
	}
}

func TestDetectCAOutOfStateMuni(t *testing.T) {
	inputs := map[string]float64{
		"1099int:1:tax_exempt_interest": 1000,
	}
	strInputs := map[string]string{
		"1099int:1:bond_state": "NY",
	}

	needed, reasons := DetectCAScheduleCANeeded(inputs, strInputs)
	if !needed {
		t.Fatal("expected Schedule CA needed for out-of-state muni interest")
	}
	if len(reasons) != 1 {
		t.Fatalf("expected 1 reason, got %d: %v", len(reasons), reasons)
	}
	if reasons[0] != "Out-of-state municipal bond interest" {
		t.Errorf("expected 'Out-of-state municipal bond interest', got %q", reasons[0])
	}
}

func TestDetectCAMuniFromCA(t *testing.T) {
	// CA muni interest should NOT trigger an adjustment
	inputs := map[string]float64{
		"1099int:1:tax_exempt_interest": 1000,
	}
	strInputs := map[string]string{
		"1099int:1:bond_state": "CA",
	}

	needed, reasons := DetectCAScheduleCANeeded(inputs, strInputs)
	if needed {
		t.Errorf("CA muni interest should not need adjustment, got reasons: %v", reasons)
	}
}

func TestDetectCAWageDifference(t *testing.T) {
	inputs := map[string]float64{
		"w2:1:wages":       100000,
		"w2:1:state_wages": 95000,
	}
	strInputs := map[string]string{}

	needed, reasons := DetectCAScheduleCANeeded(inputs, strInputs)
	if !needed {
		t.Fatal("expected Schedule CA needed for wage difference")
	}
	if len(reasons) != 1 {
		t.Fatalf("expected 1 reason, got %d: %v", len(reasons), reasons)
	}
	if reasons[0] != "State wage adjustment" {
		t.Errorf("expected 'State wage adjustment', got %q", reasons[0])
	}
}

func TestDetectCAMultipleAdjustments(t *testing.T) {
	inputs := map[string]float64{
		"form_8889:2":                       3500,
		"schedule_a:5a":                     10000,
		"schedule_c:31":                     75000,
		"1099int:1:us_savings_bond_interest": 500,
		"w2:1:wages":                        100000,
		"w2:1:state_wages":                  95000,
	}
	strInputs := map[string]string{}

	needed, reasons := DetectCAScheduleCANeeded(inputs, strInputs)
	if !needed {
		t.Fatal("expected Schedule CA needed for multiple adjustments")
	}
	if len(reasons) != 5 {
		t.Errorf("expected 5 reasons, got %d: %v", len(reasons), reasons)
	}

	// Verify all expected reasons are present
	reasonSet := make(map[string]bool)
	for _, r := range reasons {
		reasonSet[r] = true
	}
	expectedReasons := []string{
		"HSA deduction add-back",
		"State income tax deduction removal",
		"QBI deduction add-back",
		"U.S. bond interest subtraction",
		"State wage adjustment",
	}
	for _, er := range expectedReasons {
		if !reasonSet[er] {
			t.Errorf("expected reason %q not found in %v", er, reasons)
		}
	}
}
