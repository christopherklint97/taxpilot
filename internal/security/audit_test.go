package security

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestRecordAndRetrieveEntries(t *testing.T) {
	trail := NewAuditTrail(2025)

	trail.Record("1040:1a", "85000", SourceUserInput)
	trail.Record("1040:11", "72000", SourceComputed)

	if len(trail.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(trail.Entries))
	}

	e := trail.Entries[0]
	if e.FieldKey != "1040:1a" || e.Value != "85000" || e.Source != SourceUserInput {
		t.Fatalf("unexpected entry: %+v", e)
	}

	e = trail.Entries[1]
	if e.FieldKey != "1040:11" || e.Value != "72000" || e.Source != SourceComputed {
		t.Fatalf("unexpected entry: %+v", e)
	}

	if trail.TaxYear != 2025 {
		t.Fatalf("expected tax year 2025, got %d", trail.TaxYear)
	}
}

func TestRecordChangeTracksPriorValue(t *testing.T) {
	trail := NewAuditTrail(2025)

	trail.Record("1040:1a", "75000", SourcePDFImport)
	trail.RecordChange("1040:1a", "75000", "85000", SourceUserInput)

	if len(trail.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(trail.Entries))
	}

	change := trail.Entries[1]
	if change.PriorValue != "75000" {
		t.Fatalf("expected prior value 75000, got %q", change.PriorValue)
	}
	if change.Value != "85000" {
		t.Fatalf("expected new value 85000, got %q", change.Value)
	}
}

func TestGetHistoryFiltersByField(t *testing.T) {
	trail := NewAuditTrail(2025)

	trail.Record("1040:1a", "75000", SourcePDFImport)
	trail.Record("1040:11", "72000", SourceComputed)
	trail.RecordChange("1040:1a", "75000", "85000", SourceUserInput)
	trail.Record("ca540:17", "70000", SourceComputed)

	history := trail.GetHistory("1040:1a")
	if len(history) != 2 {
		t.Fatalf("expected 2 entries for 1040:1a, got %d", len(history))
	}

	for _, e := range history {
		if e.FieldKey != "1040:1a" {
			t.Fatalf("expected field key 1040:1a, got %q", e.FieldKey)
		}
	}

	// Non-existent field returns nil.
	empty := trail.GetHistory("nonexistent")
	if len(empty) != 0 {
		t.Fatalf("expected 0 entries for nonexistent field, got %d", len(empty))
	}
}

func TestSummaryCountsSources(t *testing.T) {
	trail := NewAuditTrail(2025)

	trail.Record("1040:1a", "85000", SourceUserInput)
	trail.Record("1040:first_name", "John", SourceUserInput)
	trail.Record("1040:ssn", "123-45-6789", SourcePriorYear)
	trail.Record("1040:11", "72000", SourceComputed)
	trail.Record("ca540:17", "70000", SourceComputed)
	trail.Record("w2:1:wages", "85000", SourcePDFImport)

	summary := trail.Summary()

	if !strings.Contains(summary, "Tax Year 2025") {
		t.Fatalf("summary should mention tax year 2025: %s", summary)
	}
	if !strings.Contains(summary, "Total entries: 6") {
		t.Fatalf("summary should show 6 total entries: %s", summary)
	}
	if !strings.Contains(summary, "User input: 2") {
		t.Fatalf("summary should show 2 user input: %s", summary)
	}
	if !strings.Contains(summary, "Prior year: 1") {
		t.Fatalf("summary should show 1 prior year: %s", summary)
	}
	if !strings.Contains(summary, "Computed: 2") {
		t.Fatalf("summary should show 2 computed: %s", summary)
	}
	if !strings.Contains(summary, "PDF import: 1") {
		t.Fatalf("summary should show 1 PDF import: %s", summary)
	}
}

func TestSaveLoadAuditTrailRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	trailPath := filepath.Join(tmpDir, "audit.json")

	original := NewAuditTrail(2025)
	original.Record("1040:1a", "85000", SourceUserInput)
	original.Record("1040:ssn", "123-45-6789", SourcePriorYear)
	original.RecordChange("1040:1a", "85000", "90000", SourceUserInput)

	if err := original.SaveJSON(trailPath); err != nil {
		t.Fatalf("SaveJSON: %v", err)
	}

	loaded, err := LoadAuditTrail(trailPath)
	if err != nil {
		t.Fatalf("LoadAuditTrail: %v", err)
	}

	if loaded.TaxYear != original.TaxYear {
		t.Fatalf("tax year mismatch: got %d, want %d", loaded.TaxYear, original.TaxYear)
	}

	if len(loaded.Entries) != len(original.Entries) {
		t.Fatalf("entries count mismatch: got %d, want %d", len(loaded.Entries), len(original.Entries))
	}

	// Verify the change entry preserved prior value.
	changeEntry := loaded.Entries[2]
	if changeEntry.PriorValue != "85000" {
		t.Fatalf("expected prior value 85000, got %q", changeEntry.PriorValue)
	}
	if changeEntry.Value != "90000" {
		t.Fatalf("expected value 90000, got %q", changeEntry.Value)
	}
}

func TestLoadAuditTrailFileNotFound(t *testing.T) {
	_, err := LoadAuditTrail("/nonexistent/path/audit.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestNewAuditTrailStartsEmpty(t *testing.T) {
	trail := NewAuditTrail(2024)
	if len(trail.Entries) != 0 {
		t.Fatalf("new trail should have 0 entries, got %d", len(trail.Entries))
	}
	if trail.TaxYear != 2024 {
		t.Fatalf("expected tax year 2024, got %d", trail.TaxYear)
	}
	if trail.CreatedAt.IsZero() {
		t.Fatal("CreatedAt should not be zero")
	}
}
