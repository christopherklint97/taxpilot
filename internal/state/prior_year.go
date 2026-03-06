package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// PriorYearStore manages prior-year return data.
// Stores in ~/.taxpilot/prior_years/<year>/
type PriorYearStore struct {
	baseDir string
}

// NewPriorYearStore creates a PriorYearStore using ~/.taxpilot/prior_years/ as base.
func NewPriorYearStore() *PriorYearStore {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return &PriorYearStore{
		baseDir: filepath.Join(home, ".taxpilot", "prior_years"),
	}
}

// NewPriorYearStoreWithDir creates a PriorYearStore with a custom base directory.
// Useful for testing.
func NewPriorYearStoreWithDir(baseDir string) *PriorYearStore {
	return &PriorYearStore{baseDir: baseDir}
}

// SaveContext saves a PriorYearContext for a given year.
func (s *PriorYearStore) SaveContext(ctx *PriorYearContext) error {
	if ctx == nil {
		return fmt.Errorf("cannot save nil context")
	}

	dir := s.DefaultPriorYearDir(ctx.PriorYear)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal prior year context: %w", err)
	}

	path := filepath.Join(dir, "context.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// LoadContext loads a PriorYearContext for a given year.
func (s *PriorYearStore) LoadContext(year int) (*PriorYearContext, error) {
	path := filepath.Join(s.DefaultPriorYearDir(year), "context.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var ctx PriorYearContext
	if err := json.Unmarshal(data, &ctx); err != nil {
		return nil, fmt.Errorf("unmarshal prior year context: %w", err)
	}
	return &ctx, nil
}

// HasPriorYear checks if prior-year data exists.
func (s *PriorYearStore) HasPriorYear(year int) bool {
	path := filepath.Join(s.DefaultPriorYearDir(year), "context.json")
	_, err := os.Stat(path)
	return err == nil
}

// DefaultPriorYearDir returns the path for a given year's data.
func (s *PriorYearStore) DefaultPriorYearDir(year int) string {
	return filepath.Join(s.baseDir, fmt.Sprintf("%d", year))
}
