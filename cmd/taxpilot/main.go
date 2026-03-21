package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"taxpilot/internal/efile"
	"taxpilot/internal/pdf"
	"taxpilot/internal/state"
	"taxpilot/internal/tui"
	"taxpilot/internal/tui/views"
)

func main() {
	taxYear := flag.Int("tax-year", 2025, "Tax year")
	stateCode := flag.String("state", "CA", "State code")
	importPath := flag.String("import", "", "Prior-year PDF path to import")
	continueSession := flag.Bool("continue", false, "Resume saved session")
	exportDir := flag.String("export", "", "Output directory for PDF export")
	efileMode := flag.Bool("efile", false, "Start in e-file mode")
	validateOnly := flag.Bool("validate", false, "Validate only (no TUI)")
	federalOnly := flag.Bool("federal-only", false, "Federal only")
	stateOnly := flag.Bool("state-only", false, "State only")
	flag.Parse()

	// Non-TUI: --validate
	if *validateOnly {
		runValidate(*stateCode)
		return
	}

	// Non-TUI: --export
	if *exportDir != "" {
		runExport(*exportDir, *taxYear)
		return
	}

	// TUI mode
	factory := buildFactory(*taxYear, *stateCode, *exportDir)

	welcome := views.NewWelcomeModel(*taxYear, *stateCode)

	// If --import was given, pre-parse and mark prior year loaded
	if *importPath != "" {
		parsed := importPriorYear(*importPath)
		if parsed != nil {
			welcome.SetPriorYearLoaded(parsed.TaxYear)
			factory.priorNumeric = parsed.Fields
			factory.priorString = parsed.StrFields
		}
	}

	app := tui.NewApp(&welcome, factory.ViewFactory())

	// Auto-dispatch based on flags
	var initCmd tea.Cmd
	if *continueSession {
		initCmd = func() tea.Msg {
			return tui.StartInterviewMsg{
				TaxYear:   *taxYear,
				StateCode: *stateCode,
				Continue:  true,
			}
		}
	} else if *efileMode {
		// Load state and go straight to e-file
		ret, err := state.Load(state.DefaultStorePath())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading saved state: %v\n", err)
			os.Exit(1)
		}
		initCmd = func() tea.Msg {
			return tui.StartEFileMsg{
				Results:     ret.Computed,
				StrInputs:   ret.StrInputs,
				TaxYear:     ret.TaxYear,
				State:       ret.State,
				FederalOnly: *federalOnly,
				StateOnly:   *stateOnly,
			}
		}
	}

	p := tea.NewProgram(app, tea.WithAltScreen())
	if initCmd != nil {
		go func() {
			p.Send(initCmd())
		}()
	}
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// runValidate loads saved state and prints a validation report.
func runValidate(stateCode string) {
	ret, err := state.Load(state.DefaultStorePath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading saved state: %v\n", err)
		os.Exit(1)
	}

	includeCA := stateCode == "CA" || ret.State == "CA"
	report := efile.ValidateFull(ret.Computed, ret.StrInputs, ret.TaxYear, includeCA)

	if len(report.Results) == 0 {
		fmt.Println("All validation checks passed.")
		return
	}

	for _, r := range report.Results {
		var severity string
		switch r.Severity {
		case efile.SeverityError:
			severity = "ERROR"
		case efile.SeverityWarning:
			severity = "WARN "
		case efile.SeverityInfo:
			severity = "INFO "
		}
		fmt.Printf("[%s] %s: %s (%s)\n", severity, r.Code, r.Message, r.Field)
	}

	if report.IsValid {
		fmt.Println("\nReturn is valid for e-filing.")
	} else {
		fmt.Println("\nReturn has errors that must be fixed before e-filing.")
		os.Exit(1)
	}
}

// runExport loads saved state and exports PDFs.
func runExport(outputDir string, taxYear int) {
	ret, err := state.Load(state.DefaultStorePath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading saved state: %v\n", err)
		os.Exit(1)
	}

	year := ret.TaxYear
	if year == 0 {
		year = taxYear
	}

	paths, err := pdf.ExportReturn(outputDir, ret.Computed, ret.StrInputs, year)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error exporting PDFs: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Exported %d file(s) to %s:\n", len(paths), outputDir)
	for _, p := range paths {
		fmt.Printf("  %s\n", p)
	}
}

// importPriorYear parses a prior-year PDF and returns the parsed data.
func importPriorYear(path string) *pdf.ParsedReturn {
	parser := pdf.NewParser()
	parsed, err := parser.ParseFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not import prior-year PDF: %v\n", err)
		return nil
	}
	return parsed
}
