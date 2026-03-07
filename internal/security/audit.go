package security

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// AuditSource indicates where a value came from.
type AuditSource string

const (
	SourceUserInput    AuditSource = "user_input"     // typed by user
	SourcePriorYear    AuditSource = "prior_year"     // carried from prior year
	SourcePDFImport    AuditSource = "pdf_import"     // extracted from PDF
	SourceComputed     AuditSource = "computed"        // calculated by form logic
	SourceAIDefault    AuditSource = "ai_suggested"    // suggested by LLM (user accepted)
	SourceUserAccepted AuditSource = "user_accepted"   // user accepted a default
)

// AuditEntry records one field value's provenance.
type AuditEntry struct {
	FieldKey   string      `json:"field_key"`
	Value      string      `json:"value"`                  // string repr of value
	Source     AuditSource `json:"source"`
	Timestamp  time.Time   `json:"timestamp"`
	PriorValue string      `json:"prior_value,omitempty"` // if changed
}

// AuditTrail tracks the provenance of all entered/modified values.
type AuditTrail struct {
	Entries   []AuditEntry `json:"entries"`
	CreatedAt time.Time    `json:"created_at"`
	TaxYear   int          `json:"tax_year"`
}

// NewAuditTrail creates a new audit trail for the given tax year.
func NewAuditTrail(taxYear int) *AuditTrail {
	return &AuditTrail{
		Entries:   []AuditEntry{},
		CreatedAt: time.Now().UTC(),
		TaxYear:   taxYear,
	}
}

// Record adds an entry to the audit trail.
func (t *AuditTrail) Record(fieldKey, value string, source AuditSource) {
	t.Entries = append(t.Entries, AuditEntry{
		FieldKey:  fieldKey,
		Value:     value,
		Source:    source,
		Timestamp: time.Now().UTC(),
	})
}

// RecordChange records a value change (tracks prior value).
func (t *AuditTrail) RecordChange(fieldKey, oldValue, newValue string, source AuditSource) {
	t.Entries = append(t.Entries, AuditEntry{
		FieldKey:   fieldKey,
		Value:      newValue,
		Source:     source,
		Timestamp:  time.Now().UTC(),
		PriorValue: oldValue,
	})
}

// GetHistory returns all entries for a given field key.
func (t *AuditTrail) GetHistory(fieldKey string) []AuditEntry {
	var result []AuditEntry
	for _, e := range t.Entries {
		if e.FieldKey == fieldKey {
			result = append(result, e)
		}
	}
	return result
}

// Summary returns a human-readable summary of the audit trail.
func (t *AuditTrail) Summary() string {
	counts := make(map[AuditSource]int)
	for _, e := range t.Entries {
		counts[e.Source]++
	}

	summary := fmt.Sprintf("Audit Trail for Tax Year %d:\n", t.TaxYear)
	summary += fmt.Sprintf("  Total entries: %d\n", len(t.Entries))

	labels := []struct {
		source AuditSource
		label  string
	}{
		{SourceUserInput, "User input"},
		{SourcePriorYear, "Prior year"},
		{SourcePDFImport, "PDF import"},
		{SourceComputed, "Computed"},
		{SourceAIDefault, "AI suggested"},
		{SourceUserAccepted, "User accepted"},
	}

	for _, l := range labels {
		if c, ok := counts[l.source]; ok {
			summary += fmt.Sprintf("  %s: %d\n", l.label, c)
		}
	}

	return summary
}

// SaveJSON saves the audit trail as JSON to the given path.
func (t *AuditTrail) SaveJSON(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal audit trail: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// LoadAuditTrail loads an audit trail from a JSON file.
func LoadAuditTrail(path string) (*AuditTrail, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var trail AuditTrail
	if err := json.Unmarshal(data, &trail); err != nil {
		return nil, fmt.Errorf("unmarshal audit trail: %w", err)
	}
	return &trail, nil
}
