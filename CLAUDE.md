# TaxPilot — Claude Code Instructions

## Project Overview
TaxPilot is a Go CLI that helps US taxpayers fill out and e-file federal and
California state income tax forms through an AI-assisted interview process.
It ingests prior-year returns for context, asks smart questions, and can
e-file directly to the IRS (MeF) and CA FTB, or export filled PDFs.

## Tech Stack
- **Language:** Go 1.22+
- **CLI framework:** Cobra
- **TUI framework:** Bubble Tea + Lip Gloss
- **PDF handling:** pdfcpu (reading/writing IRS + FTB forms)
- **LLM integration:** Anthropic Claude API (via anthropic-sdk-go)
- **XML generation:** encoding/xml (for MeF and CA e-file submissions)
- **SOAP client:** Custom (for IRS MeF A2A interface)
- **Embedding/search:** SQLite + vec extension (or Qdrant if needed later)
- **Config/state:** JSON files in ~/.taxpilot/
- **Testing:** Go standard testing + table-driven tests

## Key Principles

### 1. AI Does UX, Not Math
The LLM is used for:
- Generating plain-English explanations of tax questions
- Determining which follow-up questions to ask based on context
- Searching/explaining relevant tax code sections (federal & CA)
- Summarizing prior-year context
- Explaining CA <-> federal differences in plain English

The LLM is NEVER used for:
- Tax calculations (use deterministic Go code)
- Deciding tax amounts or deductions
- Anything that affects the numbers on the final forms
- Generating XML for e-file submissions

### 2. Form Logic is a Dependency Graph
Each form field is either:
- `UserInput` — needs to come from the taxpayer
- `Computed` — calculated from other fields (same or different form)
- `Lookup` — from tax tables/brackets for the year
- `PriorYear` — carried from last year's return
- `FederalRef` — (state forms only) references a federal form field

The solver walks the dependency graph across BOTH federal and state forms
and only asks for UserInput fields that are actually needed. Federal forms
are always solved first, then state forms that depend on them.

### 3. Federal and State Are Separate but Connected
- Federal forms live in `internal/forms/federal/`
- State forms live in `internal/forms/state/<state_code>/`
- State forms can reference federal fields via `FederalRef` dependencies
- The `conformity.go` module maps which federal values CA accepts as-is
  vs. which require California adjustments (Schedule CA)
- Input forms (W-2, 1099s) capture BOTH federal and state boxes

### 4. Test Against Known-Good Returns
Every form module must have table-driven tests using scenarios from
testdata/scenarios/. Each scenario has inputs and expected outputs.
CA scenarios must test conformity edge cases (SS benefits, SALT, QBI, etc.)
When adding or modifying form logic, always run the full test suite.

### 5. Year-Agnostic Form Definitions
Tax years change brackets and limits, but form structure changes less often.
Keep year-specific data in data/tax_years/YYYY/{federal,ca}/ YAML files.
Form logic in Go references these via the `taxmath` package.

### 6. E-File XML Must Be Deterministic
The MeF XML generation and CA e-file XML generation must be completely
deterministic — same inputs always produce identical XML. This makes
testing and ATS/PATS certification reliable. Never use the LLM for
anything in the e-file pipeline.

## Directory Guide
- `cmd/` — CLI entrypoint only, minimal logic
- `internal/interview/` — The interactive Q&A engine
- `internal/forms/` — Deterministic form logic (THE CORE)
- `internal/forms/federal/` — One file per IRS form
- `internal/forms/state/ca/` — One file per CA FTB form
- `internal/forms/state/interface.go` — State form interface (for future states)
- `internal/forms/inputs/` — Employer/institution provided forms (W-2, 1099s)
- `internal/knowledge/` — RAG over federal + CA tax code for explanations
- `internal/pdf/` — PDF parsing and filling
- `internal/efile/mef/` — IRS MeF e-file transmission
- `internal/efile/ca/` — CA FTB e-file transmission
- `internal/state/` — JSON state persistence between sessions
- `internal/tui/` — Bubble Tea views
- `pkg/taxmath/` — Pure math utilities (brackets, rounding, tables)
- `data/` — Year-specific tax data, knowledge base, MeF/CA schemas
- `testdata/` — Test scenarios with known-good inputs/outputs

## Common Tasks

### Adding a new federal form
1. Create `internal/forms/federal/new_form.go`
2. Define all fields with their types (UserInput/Computed/Lookup/PriorYear)
3. Implement the Compute() method for all Computed fields
4. Register the form in `internal/forms/registry.go`
5. Add PDF field mappings in `data/tax_years/YYYY/federal/forms.yaml`
6. Add MeF XML element mappings in `internal/efile/mef/xml.go`
7. Add at least one test scenario in `testdata/scenarios/federal/`
8. Run `go test ./internal/forms/...`

### Adding a new CA state form
1. Create `internal/forms/state/ca/new_form.go`
2. Define fields — use `FederalRef` type for fields that pull from federal forms
3. Check `data/tax_years/YYYY/ca/conformity.yaml` for CA <-> federal differences
4. Implement CA-specific Compute() logic (especially Schedule CA adjustments)
5. Register the form in `internal/forms/registry.go`
6. Add PDF field mappings in `data/tax_years/YYYY/ca/forms.yaml`
7. Add CA e-file XML element mappings in `internal/efile/ca/xml.go`
8. Add test scenarios in `testdata/scenarios/ca/` — MUST include conformity edge cases
9. Run `go test ./internal/forms/...`

### Adding a new state (future)
1. Create `internal/forms/state/<code>/` directory
2. Implement the `StateFormSet` interface from `internal/forms/state/interface.go`
3. Create `conformity.go` for that state's IRC conformity rules
4. Add year-specific data in `data/tax_years/YYYY/<code>/`
5. Register in the state form registry

### Adding a new tax year
1. Copy `data/tax_years/PREV_YEAR/` to `data/tax_years/NEW_YEAR/`
2. Update federal brackets, limits, and standard deductions from IRS publications
3. Update CA brackets, standard deduction, exemption credit amounts from FTB
4. Update `conformity.yaml` if CA changed its IRC conformity date
5. Check for form structure changes and update Go form definitions if needed
6. Update MeF schema version references in `internal/efile/mef/schemas.go`
7. Run full test suite against prior-year scenarios (should still pass)
8. Add new-year-specific test scenarios

## Running
```bash
go run ./cmd/taxpilot                        # Start interactive interview
go run ./cmd/taxpilot --import prior.pdf     # Import prior year return(s)
go run ./cmd/taxpilot --continue             # Resume from saved state
go run ./cmd/taxpilot --export out/          # Export filled PDFs (federal + CA)
go run ./cmd/taxpilot --efile                # E-file to IRS + CA FTB
go run ./cmd/taxpilot --efile --federal-only # E-file federal only
go run ./cmd/taxpilot --efile --state-only   # E-file CA only
go run ./cmd/taxpilot --validate             # Validate against MeF business rules
go run ./cmd/taxpilot --state CA             # Set state (default from config)
```

## Testing
```bash
go test ./...                                       # Full suite
go test ./internal/forms/... -v                     # All form logic
go test ./internal/forms/federal/... -v             # Federal forms only
go test ./internal/forms/state/ca/... -v            # CA forms only
go test ./internal/forms/... -run TestScenario/ca_  # CA scenarios
go test ./internal/efile/... -v                     # E-file XML generation
```
