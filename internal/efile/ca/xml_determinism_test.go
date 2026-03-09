package ca

import (
	"bytes"
	"testing"

	"taxpilot/internal/forms"
)

// TestXMLDeterminism_WithScheduleCA verifies that a CA return with
// Schedule CA adjustments produces byte-identical XML across runs.
func TestXMLDeterminism_WithScheduleCA(t *testing.T) {
	results := simpleW2Results()
	strInputs := baseStrInputs()

	// Add HSA add-back and interest subtraction to trigger Schedule CA
	results["ca_schedule_ca:15_col_c"] = 3850
	results["ca_schedule_ca:2_col_b"] = 500
	results["ca_schedule_ca:37_col_b"] = 500
	results["ca_schedule_ca:37_col_c"] = 3850
	results["ca_540:14"] = 500
	results["ca_540:15"] = 3850
	results["ca_540:17"] = 78350

	first, err := GenerateReturn(results, strInputs, 2025)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	for i := 0; i < 20; i++ {
		got, err := GenerateReturn(results, strInputs, 2025)
		if err != nil {
			t.Fatalf("call %d failed: %v", i, err)
		}
		if !bytes.Equal(first, got) {
			t.Fatalf("call %d produced different XML with Schedule CA (non-deterministic)", i)
		}
	}
}

// TestXMLDeterminism_FEIEAddBack verifies deterministic output when FEIE
// and housing add-backs are present in Schedule CA.
func TestXMLDeterminism_FEIEAddBack(t *testing.T) {
	results := simpleW2Results()
	strInputs := baseStrInputs()

	// Add FEIE and housing add-backs
	results[forms.SchedCALine8dColC] = 120000
	results[forms.SchedCALine8dColCHousing] = 9768
	results[forms.SchedCALine37ColC] = 129768
	results["ca_540:15"] = 129768
	results["ca_540:17"] = 204768

	first, err := GenerateReturn(results, strInputs, 2025)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	for i := 0; i < 20; i++ {
		got, err := GenerateReturn(results, strInputs, 2025)
		if err != nil {
			t.Fatalf("call %d failed: %v", i, err)
		}
		if !bytes.Equal(first, got) {
			t.Fatalf("call %d produced different XML with FEIE add-backs (non-deterministic)", i)
		}
	}
}

// TestXMLDeterminism_AllFilingStatuses verifies deterministic output across
// all five filing statuses.
func TestXMLDeterminism_AllFilingStatuses(t *testing.T) {
	statuses := []string{"single", "mfj", "mfs", "hoh", "qss"}
	results := simpleW2Results()

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			strInputs := baseStrInputs()
			strInputs["1040:filing_status"] = status

			first, err := GenerateReturn(results, strInputs, 2025)
			if err != nil {
				t.Fatalf("first call failed: %v", err)
			}

			for i := 0; i < 10; i++ {
				got, err := GenerateReturn(results, strInputs, 2025)
				if err != nil {
					t.Fatalf("call %d failed: %v", i, err)
				}
				if !bytes.Equal(first, got) {
					t.Fatalf("call %d produced different XML for status %s (non-deterministic)", i, status)
				}
			}
		})
	}
}

// TestXMLDeterminism_CrossTaxYear verifies deterministic output for
// different tax years.
func TestXMLDeterminism_CrossTaxYear(t *testing.T) {
	results := simpleW2Results()
	strInputs := baseStrInputs()

	for _, year := range []int{2024, 2025, 2026} {
		t.Run("year", func(t *testing.T) {
			first, err := GenerateReturn(results, strInputs, year)
			if err != nil {
				t.Fatalf("year %d first call failed: %v", year, err)
			}
			for i := 0; i < 10; i++ {
				got, err := GenerateReturn(results, strInputs, year)
				if err != nil {
					t.Fatalf("year %d call %d failed: %v", year, i, err)
				}
				if !bytes.Equal(first, got) {
					t.Fatalf("year %d call %d produced different XML (non-deterministic)", year, i)
				}
			}
		})
	}
}
