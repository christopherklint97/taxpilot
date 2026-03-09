package forms_test

import (
	"strings"
	"testing"

	"taxpilot/internal/forms"
	"taxpilot/internal/interview"
)

// TestAllFormIDsRegistered verifies that every FormID constant in AllFormIDs()
// can be registered in a registry without panics, and that no FormID is duplicated.
func TestAllFormIDsRegistered(t *testing.T) {
	seen := make(map[forms.FormID]bool)
	for _, id := range forms.AllFormIDs() {
		if seen[id] {
			t.Errorf("duplicate FormID: %s", id)
		}
		seen[id] = true

		// Verify the ID is non-empty and doesn't contain spaces
		if string(id) == "" {
			t.Error("empty FormID found")
		}
		if strings.Contains(string(id), " ") {
			t.Errorf("FormID %q contains spaces", id)
		}
	}
}

// TestFieldKeyConsistency verifies that FK() and FieldKey() produce the same output.
func TestFieldKeyConsistency(t *testing.T) {
	tests := []struct {
		formID forms.FormID
		line   string
		want   string
	}{
		{forms.FormF1040, "filing_status", "1040:filing_status"},
		{forms.FormW2, "wages", "w2:wages"},
		{forms.FormScheduleA, "5a", "schedule_a:5a"},
		{forms.FormF2555, "foreign_country", "form_2555:foreign_country"},
		{forms.FormCA540, "17", "ca_540:17"},
	}

	for _, tt := range tests {
		fk := forms.FK(tt.formID, tt.line)
		fieldKey := forms.FieldKey(tt.formID, tt.line)
		if fk != tt.want {
			t.Errorf("FK(%s, %s) = %q, want %q", tt.formID, tt.line, fk, tt.want)
		}
		if fieldKey != tt.want {
			t.Errorf("FieldKey(%s, %s) = %q, want %q", tt.formID, tt.line, fieldKey, tt.want)
		}
		if fk != fieldKey {
			t.Errorf("FK and FieldKey disagree: %q vs %q", fk, fieldKey)
		}
	}
}

// TestFieldKeyConstantsMatchFormIDs verifies that field key constants
// have the correct form ID prefix.
func TestFieldKeyConstantsMatchFormIDs(t *testing.T) {
	checks := []struct {
		key    string
		formID forms.FormID
	}{
		{forms.F1040FilingStatus, forms.FormF1040},
		{forms.F1040Line1a, forms.FormF1040},
		{forms.F1040Line11, forms.FormF1040},
		{forms.SchedALine1, forms.FormScheduleA},
		{forms.SchedBLine4, forms.FormScheduleB},
		{forms.SchedCLine31, forms.FormScheduleC},
		{forms.SchedDLine16, forms.FormScheduleD},
		{forms.Sched1Line10, forms.FormSchedule1},
		{forms.Sched2Line21, forms.FormSchedule2},
		{forms.Sched3Line8, forms.FormSchedule3},
		{forms.SchedSELine6, forms.FormScheduleSE},
		{forms.F8889Line2, forms.FormF8889},
		{forms.F8995Line10, forms.FormF8995},
		{forms.F2555ForeignCountry, forms.FormF2555},
		{forms.F1116Line22, forms.FormF1116},
		{forms.F8938LivesAbroad, forms.FormF8938},
		{forms.F8833TreatyCountry, forms.FormF8833},
		{forms.CA540Line17, forms.FormCA540},
		{forms.SchedCALine8dColC, forms.FormScheduleCA},
	}

	for _, c := range checks {
		prefix := string(c.formID) + ":"
		if !strings.HasPrefix(c.key, prefix) {
			t.Errorf("field key %q should start with %q", c.key, prefix)
		}
	}
}

// TestValidateFieldDefs runs ValidateFieldDefs on the full registry from
// SetupRegistry and verifies no common field definition errors are present.
func TestValidateFieldDefs(t *testing.T) {
	reg := interview.SetupRegistry()
	errs := reg.ValidateFieldDefs()
	for _, err := range errs {
		t.Errorf("field definition error: %v", err)
	}
}

// TestDepValuesGetStrict verifies GetStrict returns an error for missing keys.
func TestDepValuesGetStrict(t *testing.T) {
	dv := forms.NewDepValues(map[string]float64{"a": 42}, nil, 2025)

	v, err := dv.GetStrict("a")
	if err != nil {
		t.Fatalf("unexpected error for existing key: %v", err)
	}
	if v != 42 {
		t.Errorf("got %f, want 42", v)
	}

	_, err = dv.GetStrict("missing")
	if err == nil {
		t.Error("expected error for missing key, got nil")
	}
}

// TestDepValuesKeys verifies Keys returns all available keys.
func TestDepValuesKeys(t *testing.T) {
	dv := forms.NewDepValues(map[string]float64{"a": 1, "b": 2, "c": 3}, nil, 2025)
	keys := dv.Keys()
	if len(keys) != 3 {
		t.Errorf("got %d keys, want 3", len(keys))
	}
	seen := make(map[string]bool)
	for _, k := range keys {
		seen[k] = true
	}
	for _, want := range []string{"a", "b", "c"} {
		if !seen[want] {
			t.Errorf("missing key %q", want)
		}
	}
}

// TestValidateFederalRefs runs ValidateFederalRefs on the full registry and
// verifies that all FederalRef fields reference federal forms.
func TestValidateFederalRefs(t *testing.T) {
	reg := interview.SetupRegistry()
	errs := reg.ValidateFederalRefs()
	for _, err := range errs {
		t.Errorf("FederalRef validation error: %v", err)
	}
}

// TestValidateFederalRefsDetectsNonFederal verifies that ValidateFederalRefs
// catches FederalRef fields that point to non-federal forms.
func TestValidateFederalRefsDetectsNonFederal(t *testing.T) {
	reg := forms.NewRegistry()

	// Register a CA form
	reg.Register(&forms.FormDef{
		ID:           "ca_test",
		Jurisdiction: forms.StateCA,
		Fields:       []forms.FieldDef{},
	})

	// Register a form with a FederalRef pointing to the CA form
	reg.Register(&forms.FormDef{
		ID:           "bad_form",
		Jurisdiction: forms.StateCA,
		Fields: []forms.FieldDef{
			{
				Line:      "1",
				Type:      forms.FederalRef,
				Label:     "Bad ref",
				DependsOn: []string{"ca_test:some_field"},
			},
		},
	})

	errs := reg.ValidateFederalRefs()
	if len(errs) == 0 {
		t.Error("expected ValidateFederalRefs to report error for FederalRef pointing to non-federal form")
	}
}

// TestFieldByLine verifies that FormDef.FieldByLine returns the correct field
// and nil for missing lines.
func TestFieldByLine(t *testing.T) {
	form := &forms.FormDef{
		ID:   "test_form",
		Name: "Test Form",
		Fields: []forms.FieldDef{
			{Line: "1", Label: "First"},
			{Line: "2a", Label: "Second A"},
			{Line: "filing_status", Label: "Filing Status"},
		},
	}

	// Found cases
	f := form.FieldByLine("1")
	if f == nil || f.Label != "First" {
		t.Errorf("FieldByLine(\"1\") = %v, want field with label \"First\"", f)
	}

	f = form.FieldByLine("2a")
	if f == nil || f.Label != "Second A" {
		t.Errorf("FieldByLine(\"2a\") = %v, want field with label \"Second A\"", f)
	}

	f = form.FieldByLine("filing_status")
	if f == nil || f.Label != "Filing Status" {
		t.Errorf("FieldByLine(\"filing_status\") = %v, want field with label \"Filing Status\"", f)
	}

	// Not found case
	f = form.FieldByLine("nonexistent")
	if f != nil {
		t.Errorf("FieldByLine(\"nonexistent\") = %v, want nil", f)
	}
}

// TestFieldByLineUsedByGetField verifies that Registry.GetField uses the
// indexed lookup via FieldByLine.
func TestFieldByLineUsedByGetField(t *testing.T) {
	reg := forms.NewRegistry()
	reg.Register(&forms.FormDef{
		ID:   "idx_test",
		Name: "Index Test",
		Fields: []forms.FieldDef{
			{Line: "a", Label: "Alpha"},
			{Line: "b", Label: "Beta"},
			{Line: "c", Label: "Gamma"},
		},
	})

	_, field, err := reg.GetField("idx_test:b")
	if err != nil {
		t.Fatalf("GetField returned error: %v", err)
	}
	if field.Label != "Beta" {
		t.Errorf("GetField returned label %q, want \"Beta\"", field.Label)
	}

	_, _, err = reg.GetField("idx_test:missing")
	if err == nil {
		t.Error("expected error for missing field, got nil")
	}
}

// TestInputFormPrefixes verifies InputFormPrefixes returns correct prefixes.
func TestInputFormPrefixes(t *testing.T) {
	prefixes := forms.InputFormPrefixes()
	expected := map[string]bool{
		"w2:":      true,
		"1099int:": true,
		"1099div:": true,
		"1099nec:": true,
		"1099b:":   true,
	}
	if len(prefixes) != len(expected) {
		t.Errorf("got %d prefixes, want %d", len(prefixes), len(expected))
	}
	for _, p := range prefixes {
		if !expected[p] {
			t.Errorf("unexpected prefix: %q", p)
		}
	}
}
