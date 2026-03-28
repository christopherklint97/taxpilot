package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"taxpilot/internal/efile"
	"taxpilot/internal/pdf"
	"taxpilot/internal/state"
	"taxpilot/internal/tui"
	"taxpilot/internal/tui/views"
)

// stringSlice is a flag.Value that collects repeated -import flags.
type stringSlice []string

func (s *stringSlice) String() string { return strings.Join(*s, ",") }
func (s *stringSlice) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func main() {
	taxYear := flag.Int("tax-year", 2025, "Tax year")
	stateCode := flag.String("state", "CA", "State code")
	var importPaths stringSlice
	flag.Var(&importPaths, "import", "Prior-year PDF file or directory (repeatable)")
	continueSession := flag.Bool("continue", false, "Resume saved session")
	exportDir := flag.String("export", "", "Output directory for PDF export")
	efileMode := flag.Bool("efile", false, "Start in e-file mode")
	validateOnly := flag.Bool("validate", false, "Validate only (no TUI)")
	federalOnly := flag.Bool("federal-only", false, "Federal only")
	stateOnly := flag.Bool("state-only", false, "State only")
	modelOverride := flag.String("model", "", "OpenRouter model (overrides TAXPILOT_MODEL and default)")
	rollforward := flag.Bool("rollforward", false, "Rollforward prior year return to new year")
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
	factory := buildFactory(*taxYear, *stateCode, *exportDir, *modelOverride)

	welcome := views.NewWelcomeModel(*taxYear, *stateCode)

	// If --import was given, parse all files and merge
	if len(importPaths) > 0 {
		debugLog("--import: parsing %d paths: %v", len(importPaths), []string(importPaths))
		merged, formNames, err := pdf.ParseMultipleFiles(importPaths)
		if err != nil {
			debugLog("--import: parse error: %v", err)
			fmt.Fprintf(os.Stderr, "Warning: could not import prior-year PDFs: %v\n", err)
		} else {
			debugLog("--import: parsed OK — TaxYear=%d, Fields=%d, StrFields=%d, forms=%v",
				merged.TaxYear, len(merged.Fields), len(merged.StrFields), formNames)
			welcome.SetPriorYearLoadedMulti(merged.TaxYear, formNames)
			factory.priorNumeric = merged.Fields
			factory.priorString = merged.StrFields

			// Persist so --continue can find it later
			if merged.TaxYear > 0 {
				priorRet := state.NewTaxReturn(merged.TaxYear, *stateCode)
				priorRet.Inputs = merged.Fields
				priorRet.StrInputs = merged.StrFields
				priorRet.Complete = true
				_ = state.Save(state.YearStorePath(merged.TaxYear), priorRet)
			}
		}
	} else {
		debugLog("no --import flag")
	}

	app := tui.NewApp(&welcome, factory.ViewFactory())

	// Auto-dispatch based on flags
	var initCmd tea.Cmd
	if *rollforward {
		initCmd = func() tea.Msg {
			return tui.StartRollforwardMsg{
				TaxYear:   *taxYear,
				StateCode: *stateCode,
			}
		}
	} else if *continueSession {
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
