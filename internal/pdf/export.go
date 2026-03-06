package pdf

import (
	"fmt"
	"os"
)

// ExportReturn exports all forms for a completed return.
// It generates text exports (or PDF when templates are available) to the specified directory.
// Returns the list of generated file paths.
func ExportReturn(outputDir string, values map[string]float64, strValues map[string]string, taxYear int) ([]string, error) {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, fmt.Errorf("create output directory: %w", err)
	}

	filler := NewFiller(outputDir)

	// Register all known form mappings
	f1040 := Federal1040Mappings()
	ca540 := CA540Mappings()

	filler.RegisterForm(f1040)
	filler.RegisterForm(ca540)

	// Fill all registered forms
	paths, err := filler.FillAll(values, strValues)
	if err != nil {
		return paths, fmt.Errorf("export forms: %w", err)
	}

	return paths, nil
}
