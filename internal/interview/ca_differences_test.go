package interview

import "testing"

func TestGetCADifference_KnownFields(t *testing.T) {
	tests := []struct {
		fieldKey     string
		expectedArea string
	}{
		{"w2:1:wages", "Wages"},
		{"w2:1:state_wages", "Wages"},
		{"schedule_a:5a", "SALT Deduction"},
		{"form_8889:2", "HSA Deduction"},
		{"form_8889:3", "HSA Deduction"},
		{"1099div:1:section_199a_dividends", "QBI Deduction (Section 199A)"},
		{"1099div:1:qualified_dividends", "Qualified Dividends"},
		{"1099b:1:proceeds", "Capital Gains"},
		{"1099b:1:cost_basis", "Capital Gains"},
		{"1099b:1:term", "Capital Gains"},
		{"1099int:1:us_savings_bond_interest", "U.S. Government Bond Interest"},
		{"1099int:1:tax_exempt_interest", "Municipal Bond Interest"},
		{"1099div:1:exempt_interest_dividends", "Municipal Bond Interest"},
	}

	for _, tt := range tests {
		t.Run(tt.fieldKey, func(t *testing.T) {
			diff := GetCADifference(tt.fieldKey)
			if diff == nil {
				t.Fatalf("GetCADifference(%q) returned nil, expected area %q", tt.fieldKey, tt.expectedArea)
			}
			if diff.Area != tt.expectedArea {
				t.Errorf("GetCADifference(%q).Area = %q, want %q", tt.fieldKey, diff.Area, tt.expectedArea)
			}
		})
	}
}

func TestGetCADifference_NilForUnknownFields(t *testing.T) {
	fields := []string{
		"1040:first_name",
		"1040:last_name",
		"1040:ssn",
		"w2:1:employer_name",
		"nonexistent:field",
		"",
	}

	for _, key := range fields {
		t.Run(key, func(t *testing.T) {
			diff := GetCADifference(key)
			if diff != nil {
				t.Errorf("GetCADifference(%q) returned non-nil (Area=%q), expected nil", key, diff.Area)
			}
		})
	}
}

func TestAllCADifferences_Count(t *testing.T) {
	diffs := AllCADifferences()
	expected := 11
	if len(diffs) != expected {
		t.Errorf("AllCADifferences() returned %d differences, want %d", len(diffs), expected)
	}
}

func TestAllCADifferences_RequiredFieldsNonEmpty(t *testing.T) {
	diffs := AllCADifferences()
	for _, diff := range diffs {
		t.Run(diff.Area, func(t *testing.T) {
			if diff.Area == "" {
				t.Error("Area is empty")
			}
			if diff.Federal == "" {
				t.Error("Federal description is empty")
			}
			if diff.California == "" {
				t.Error("California description is empty")
			}
			if diff.Impact == "" {
				t.Error("Impact description is empty")
			}
			// IRCSection can be empty for CA-only provisions (Mental Health Tax)
			// RTCSection should always be present for CA differences
			if diff.RTCSection == "" {
				t.Error("RTCSection is empty")
			}
		})
	}
}

func TestAllCADifferences_ReturnsCopy(t *testing.T) {
	diffs1 := AllCADifferences()
	diffs2 := AllCADifferences()
	if &diffs1[0] == &diffs2[0] {
		t.Error("AllCADifferences() should return a copy, not the same slice")
	}
}
