package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// TaxReturn holds all user-provided and computed values for a return.
type TaxReturn struct {
	TaxYear      int                `json:"tax_year"`
	State        string             `json:"state"`
	FilingStatus string             `json:"filing_status"`
	Inputs       map[string]float64 `json:"inputs"`
	StrInputs    map[string]string  `json:"str_inputs"`
	Computed     map[string]float64 `json:"computed"`
	PriorYear    map[string]float64 `json:"prior_year"`
	LastUpdated  string             `json:"last_updated"`
	Complete     bool               `json:"complete"`
	PriorYearCtx *PriorYearContext  `json:"prior_year_ctx,omitempty"` // extracted prior-year context
}

// DefaultStorePath returns ~/.taxpilot/state.json.
func DefaultStorePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".taxpilot", "state.json")
}

// Save persists a TaxReturn to the given path as JSON.
func Save(path string, ret *TaxReturn) error {
	ret.LastUpdated = time.Now().UTC().Format(time.RFC3339)

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(ret, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal tax return: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// Load reads a TaxReturn from the given JSON file.
func Load(path string) (*TaxReturn, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var ret TaxReturn
	if err := json.Unmarshal(data, &ret); err != nil {
		return nil, fmt.Errorf("unmarshal tax return: %w", err)
	}
	return &ret, nil
}

// NewTaxReturn creates a fresh TaxReturn with sensible defaults.
func NewTaxReturn(year int, stateCode string) *TaxReturn {
	return &TaxReturn{
		TaxYear:      year,
		State:        stateCode,
		FilingStatus: "",
		Inputs:       make(map[string]float64),
		StrInputs:    make(map[string]string),
		Computed:     make(map[string]float64),
		PriorYear:    make(map[string]float64),
		LastUpdated:  time.Now().UTC().Format(time.RFC3339),
		Complete:     false,
	}
}
