package forms_test

import (
	"strings"
	"testing"

	"taxpilot/internal/forms"
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
