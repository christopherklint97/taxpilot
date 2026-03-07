package interview

import (
	"strings"
	"testing"

	"taxpilot/internal/forms"
)

func TestGetCAPreFillNote_HSA(t *testing.T) {
	py := map[string]float64{"form_8889:2": 3500}
	note := GetCAPreFillNote("form_8889:2", py, nil)
	if note == "" {
		t.Fatal("expected non-empty CA note for HSA")
	}
	if !strings.Contains(note, "added back") {
		t.Errorf("HSA note should mention add-back, got: %s", note)
	}
	if !strings.Contains(note, "$3,500") {
		t.Errorf("HSA note should include amount, got: %s", note)
	}
}

func TestGetCAPreFillNote_HSA_NoContributions(t *testing.T) {
	py := map[string]float64{}
	note := GetCAPreFillNote("form_8889:2", py, nil)
	if note == "" {
		t.Fatal("expected generic CA note for HSA field")
	}
	if !strings.Contains(note, "does not allow HSA") {
		t.Errorf("expected generic HSA note, got: %s", note)
	}
}

func TestGetCAPreFillNote_SALT(t *testing.T) {
	py := map[string]float64{"schedule_a:5a": 10000}
	note := GetCAPreFillNote("schedule_a:5a", py, nil)
	if note == "" {
		t.Fatal("expected CA note for SALT")
	}
	if !strings.Contains(note, "removed") {
		t.Errorf("SALT note should mention removal, got: %s", note)
	}
}

func TestGetCAPreFillNote_PropertyTax(t *testing.T) {
	py := map[string]float64{"schedule_a:5c": 8000}
	note := GetCAPreFillNote("schedule_a:5c", py, nil)
	if note == "" {
		t.Fatal("expected CA note for property tax")
	}
	if !strings.Contains(note, "no SALT cap") && !strings.Contains(note, "no cap") {
		t.Errorf("property tax note should mention no cap, got: %s", note)
	}
}

func TestGetCAPreFillNote_USBonds(t *testing.T) {
	py := map[string]float64{"1099int:1:us_savings_bond_interest": 500}
	note := GetCAPreFillNote("1099int:1:us_savings_bond_interest", py, nil)
	if note == "" {
		t.Fatal("expected CA note for US bonds")
	}
	if !strings.Contains(note, "subtracted") {
		t.Errorf("US bond note should mention subtraction, got: %s", note)
	}
}

func TestGetCAPreFillNote_Wages_Different(t *testing.T) {
	py := map[string]float64{"w2:1:wages": 75000, "w2:1:state_wages": 73000}
	note := GetCAPreFillNote("w2:1:wages", py, nil)
	if note == "" {
		t.Fatal("expected CA note when wages differ")
	}
	if !strings.Contains(note, "differed") {
		t.Errorf("wages note should mention difference, got: %s", note)
	}
}

func TestGetCAPreFillNote_Wages_Same(t *testing.T) {
	py := map[string]float64{"w2:1:wages": 75000, "w2:1:state_wages": 75000}
	note := GetCAPreFillNote("w2:1:wages", py, nil)
	if note != "" {
		t.Errorf("no CA note expected when wages are same, got: %s", note)
	}
}

func TestGetCAPreFillNote_UnknownField(t *testing.T) {
	note := GetCAPreFillNote("1040:first_name", nil, nil)
	if note != "" {
		t.Errorf("no CA note expected for name field, got: %s", note)
	}
}

func TestGetCAPreFillNote_QualifiedDividends(t *testing.T) {
	note := GetCAPreFillNote("1099div:1:qualified_dividends", nil, nil)
	if note == "" {
		t.Fatal("expected CA note for qualified dividends")
	}
	if !strings.Contains(note, "ordinary income") {
		t.Errorf("qualified dividends note should mention ordinary income, got: %s", note)
	}
}

func TestCAScheduleCANote_NoAdjustments(t *testing.T) {
	py := map[string]float64{"ca_schedule_ca:37_col_b": 0, "ca_schedule_ca:37_col_c": 0}
	note := caScheduleCANote(py)
	if !strings.Contains(note, "no Schedule CA adjustments") {
		t.Errorf("expected no-adjustments message, got: %s", note)
	}
}

func TestCAScheduleCANote_WithAdjustments(t *testing.T) {
	py := map[string]float64{"ca_schedule_ca:37_col_b": 5000, "ca_schedule_ca:37_col_c": 3500}
	note := caScheduleCANote(py)
	if !strings.Contains(note, "subtracted") {
		t.Errorf("expected subtraction mention, got: %s", note)
	}
	if !strings.Contains(note, "added back") {
		t.Errorf("expected addition mention, got: %s", note)
	}
}

func TestCANoteInPriorYearDefault(t *testing.T) {
	reg := testRegistry()
	// Add a field that has CA notes
	reg.Register(&forms.FormDef{
		ID:           "form_8889",
		Name:         "HSA",
		Jurisdiction: forms.Federal,
		TaxYears:     []int{2025},
		Fields: []forms.FieldDef{
			{Line: "2", Type: forms.UserInput, Label: "HSA Contributions", Prompt: "How much?"},
		},
	})

	py := map[string]float64{"form_8889:2": 4000}
	eng, err := NewEngineWithPriorYear(reg, 2025, py, nil, "CA")
	if err != nil {
		t.Fatalf("NewEngineWithPriorYear failed: %v", err)
	}

	// Advance to the HSA question
	for eng.HasNext() {
		q := eng.Current()
		if q.Key == "form_8889:2" {
			pyd := eng.GetPriorYearDefault()
			if pyd == nil {
				t.Fatal("expected prior year default for HSA")
			}
			if pyd.CANote == "" {
				t.Error("expected CA note in prior year default for HSA field")
			}
			return
		}
		// Answer non-HSA questions to advance
		if q.IsString || len(q.Options) > 0 {
			eng.strInputs[q.Key] = "test"
			eng.inputs[q.Key] = 0
		} else {
			eng.inputs[q.Key] = 0
		}
		eng.current++
	}
	t.Error("never reached HSA question")
}

func TestCANoteNotShownWithoutCA(t *testing.T) {
	reg := testRegistry()
	py := map[string]float64{"w2:1:wages": 75000, "w2:1:state_wages": 73000}
	// Not filing in CA
	eng, err := NewEngineWithPriorYear(reg, 2025, py, nil, "")
	if err != nil {
		t.Fatalf("NewEngineWithPriorYear failed: %v", err)
	}

	for eng.HasNext() {
		pyd := eng.GetPriorYearDefault()
		if pyd != nil && pyd.CANote != "" {
			t.Errorf("should not have CA note when not filing in CA, got: %s for %s", pyd.CANote, pyd.FieldKey)
		}
		if eng.Current().IsString || len(eng.Current().Options) > 0 {
			eng.strInputs[eng.Current().Key] = "test"
			eng.inputs[eng.Current().Key] = 0
		} else {
			eng.inputs[eng.Current().Key] = 0
		}
		eng.current++
	}
}
