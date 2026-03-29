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
	filler.RegisterForm(Federal1040Mappings())
	filler.RegisterForm(ScheduleAMappings())
	filler.RegisterForm(ScheduleBMappings())
	filler.RegisterForm(ScheduleCMappings())
	filler.RegisterForm(ScheduleDMappings())
	filler.RegisterForm(Form8949Mappings())
	filler.RegisterForm(Schedule1Mappings())
	filler.RegisterForm(Schedule2Mappings())
	filler.RegisterForm(Schedule3Mappings())
	filler.RegisterForm(ScheduleSEMappings())
	filler.RegisterForm(Form8995Mappings())
	filler.RegisterForm(Form8889Mappings())
	filler.RegisterForm(CA540Mappings())
	filler.RegisterForm(ScheduleCAMappings())
	filler.RegisterForm(Form3514Mappings())
	filler.RegisterForm(Form3853Mappings())
	filler.RegisterForm(Form2555Mappings())
	filler.RegisterForm(Form1116Mappings())
	filler.RegisterForm(Form8938Mappings())
	filler.RegisterForm(Form8833Mappings())

	// Fill all registered forms
	paths, err := filler.FillAll(values, strValues)
	if err != nil {
		return paths, fmt.Errorf("export forms: %w", err)
	}

	return paths, nil
}
