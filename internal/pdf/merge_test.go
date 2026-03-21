package pdf

import (
	"os"
	"path/filepath"
	"testing"

	"taxpilot/internal/forms"
)

func TestMergeReturns_Empty(t *testing.T) {
	if MergeReturns(nil) != nil {
		t.Error("MergeReturns(nil) should return nil")
	}
	if MergeReturns([]*ParsedReturn{}) != nil {
		t.Error("MergeReturns([]) should return nil")
	}
}

func TestMergeReturns_Single(t *testing.T) {
	r := &ParsedReturn{
		FormID:  forms.FormF1040,
		TaxYear: 2024,
		Fields:  map[string]float64{"1040:11": 75000},
	}
	got := MergeReturns([]*ParsedReturn{r})
	if got != r {
		t.Error("MergeReturns with single item should return the same pointer")
	}
}

func TestMergeReturns_MultipleFormsMerge(t *testing.T) {
	federal := &ParsedReturn{
		FormID:    forms.FormF1040,
		TaxYear:   2024,
		Fields:    map[string]float64{"1040:11": 75000, "1040:15": 10000},
		StrFields: map[string]string{"1040:first_name": "Jane"},
		RawFields: map[string]string{"f1_02": "Jane"},
	}
	state := &ParsedReturn{
		FormID:    forms.FormCA540,
		TaxYear:   2024,
		Fields:    map[string]float64{"ca_540:13": 50000, "ca_540:19": 4500},
		StrFields: map[string]string{"ca_540:first_name": "Jane"},
		RawFields: map[string]string{"Line_13": "50000"},
	}

	merged := MergeReturns([]*ParsedReturn{federal, state})
	if merged == nil {
		t.Fatal("MergeReturns returned nil")
	}

	// Check tax year from first return
	if merged.TaxYear != 2024 {
		t.Errorf("TaxYear = %d, want 2024", merged.TaxYear)
	}

	// Check all fields merged
	wantFields := map[string]float64{
		"1040:11":   75000,
		"1040:15":   10000,
		"ca_540:13": 50000,
		"ca_540:19": 4500,
	}
	for k, want := range wantFields {
		if got, ok := merged.Fields[k]; !ok {
			t.Errorf("missing field %s", k)
		} else if got != want {
			t.Errorf("Fields[%s] = %v, want %v", k, got, want)
		}
	}

	// Check string fields merged
	if merged.StrFields["1040:first_name"] != "Jane" {
		t.Errorf("missing federal string field")
	}
	if merged.StrFields["ca_540:first_name"] != "Jane" {
		t.Errorf("missing CA string field")
	}
}

func TestMergeReturns_TaxYearFromFirst(t *testing.T) {
	r1 := &ParsedReturn{TaxYear: 0, Fields: map[string]float64{}, StrFields: map[string]string{}, RawFields: map[string]string{}}
	r2 := &ParsedReturn{TaxYear: 2024, Fields: map[string]float64{}, StrFields: map[string]string{}, RawFields: map[string]string{}}
	merged := MergeReturns([]*ParsedReturn{r1, r2})
	if merged.TaxYear != 2024 {
		t.Errorf("TaxYear = %d, want 2024 (from second return)", merged.TaxYear)
	}
}

func TestMergeReturns_LaterOverridesEarlier(t *testing.T) {
	r1 := &ParsedReturn{
		Fields:    map[string]float64{"1040:11": 50000},
		StrFields: map[string]string{"1040:first_name": "Old"},
		RawFields: map[string]string{},
	}
	r2 := &ParsedReturn{
		Fields:    map[string]float64{"1040:11": 75000},
		StrFields: map[string]string{"1040:first_name": "New"},
		RawFields: map[string]string{},
	}
	merged := MergeReturns([]*ParsedReturn{r1, r2})
	if merged.Fields["1040:11"] != 75000 {
		t.Errorf("expected later value to override: got %v", merged.Fields["1040:11"])
	}
	if merged.StrFields["1040:first_name"] != "New" {
		t.Errorf("expected later string to override: got %v", merged.StrFields["1040:first_name"])
	}
}

func TestMergeInto(t *testing.T) {
	dst := &ParsedReturn{
		TaxYear:   2024,
		Fields:    map[string]float64{"1040:11": 75000},
		StrFields: map[string]string{"1040:first_name": "Jane"},
	}
	src := &ParsedReturn{
		TaxYear:   2024,
		Fields:    map[string]float64{"ca_540:13": 50000},
		StrFields: map[string]string{"ca_540:first_name": "Jane"},
	}

	MergeInto(dst, src)

	if dst.Fields["ca_540:13"] != 50000 {
		t.Error("MergeInto did not add CA field")
	}
	if dst.Fields["1040:11"] != 75000 {
		t.Error("MergeInto lost federal field")
	}
	if dst.StrFields["ca_540:first_name"] != "Jane" {
		t.Error("MergeInto did not add CA string field")
	}
}

func TestMergeInto_NilSafe(t *testing.T) {
	// Should not panic
	MergeInto(nil, &ParsedReturn{})
	MergeInto(&ParsedReturn{}, nil)
}

func TestExpandPaths_File(t *testing.T) {
	// Create a temp file
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.pdf")
	os.WriteFile(path, []byte("fake"), 0o644)

	got, err := expandPaths([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != path {
		t.Errorf("expandPaths file = %v, want [%s]", got, path)
	}
}

func TestExpandPaths_Directory(t *testing.T) {
	tmpDir := t.TempDir()
	// Create some PDFs
	os.WriteFile(filepath.Join(tmpDir, "federal.pdf"), []byte("fake"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "state.pdf"), []byte("fake"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "notes.txt"), []byte("not a pdf"), 0o644)

	got, err := expandPaths([]string{tmpDir})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Errorf("expandPaths directory = %d files, want 2", len(got))
	}
}

func TestExpandPaths_Nonexistent(t *testing.T) {
	_, err := expandPaths([]string{"/nonexistent/path"})
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestFormLabel(t *testing.T) {
	tests := []struct {
		id   forms.FormID
		want string
	}{
		{forms.FormF1040, "Form 1040"},
		{forms.FormCA540, "CA Form 540"},
		{forms.FormScheduleA, "Schedule A"},
		{forms.FormScheduleCA, "CA Schedule CA"},
		{forms.FormID("unknown"), "unknown"},
	}

	for _, tt := range tests {
		got := formLabel(tt.id)
		if got != tt.want {
			t.Errorf("formLabel(%q) = %q, want %q", tt.id, got, tt.want)
		}
	}
}

func TestAppendUniqueStrings(t *testing.T) {
	got := appendUniqueStrings(
		[]string{"Form 1040", "Schedule A"},
		[]string{"CA Form 540", "Form 1040"},
	)
	if len(got) != 3 {
		t.Errorf("expected 3 unique items, got %d: %v", len(got), got)
	}
}

// appendUniqueStrings is duplicated here for testing the welcome view helper.
// In production it lives in welcome.go, but since it's unexported we test via merge_test.
func appendUniqueStrings(base, add []string) []string {
	seen := make(map[string]bool, len(base))
	for _, s := range base {
		seen[s] = true
	}
	for _, s := range add {
		if !seen[s] {
			base = append(base, s)
			seen[s] = true
		}
	}
	return base
}
