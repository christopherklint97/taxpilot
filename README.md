# TaxPilot

A Go CLI that helps US taxpayers fill out and e-file federal and California state income tax forms through an AI-assisted interview process. Built with deterministic tax math — the LLM handles UX, never the numbers.

## What It Does

- **Interactive interview** walks you through your tax return, asking only the questions that apply to your situation
- **Dependency graph solver** computes all form fields in topological order across federal and state forms
- **Prior-year import** parses last year's PDF return for pre-fill and context
- **AI explanations** provide plain-English help for any question, with IRC/RTC references
- **PDF export** generates filled federal and CA tax forms
- **E-file** transmits returns to IRS (MeF) and CA FTB electronically
- **Full expat support** for Americans living abroad (FEIE, FTC, FATCA, FBAR, treaty disclosure)

## Supported Forms

### Federal (IRS)
| Form | Description |
|------|-------------|
| **1040** | Individual Income Tax Return |
| **Schedule 1** | Additional Income and Adjustments |
| **Schedule 2** | Additional Taxes (SE, NIIT, Additional Medicare) |
| **Schedule 3** | Additional Credits and Payments |
| **Schedule A** | Itemized Deductions |
| **Schedule B** | Interest and Dividends (+ Part III foreign accounts) |
| **Schedule C** | Profit or Loss From Business |
| **Schedule D** | Capital Gains and Losses |
| **Schedule SE** | Self-Employment Tax |
| **Form 8949** | Sales and Dispositions of Capital Assets |
| **Form 8889** | Health Savings Accounts |
| **Form 8995** | Qualified Business Income Deduction |
| **Form 2555** | Foreign Earned Income Exclusion |
| **Form 1116** | Foreign Tax Credit |
| **Form 8938** | FATCA (Foreign Financial Assets) |
| **Form 8833** | Treaty-Based Return Position Disclosure |

### Input Forms
W-2, 1099-INT, 1099-DIV, 1099-NEC, 1099-B

### California (FTB)
| Form | Description |
|------|-------------|
| **Form 540** | California Resident Income Tax Return |
| **Schedule CA** | California Adjustments (FEIE add-back, HSA, SALT, QBI) |
| **Form 3514** | California Earned Income Tax Credit |
| **Form 3853** | Health Coverage (Individual Mandate) |

## Architecture

```
cmd/taxpilot/          CLI entrypoint (Cobra)
internal/
  forms/               Deterministic form logic (THE CORE)
    federal/            One file per IRS form
    state/ca/           One file per CA FTB form
    inputs/             W-2, 1099s
  interview/            Interactive Q&A engine, situation detection
  knowledge/            RAG over federal + CA tax code (100+ documents)
  efile/                E-file infrastructure
    mef/                IRS MeF XML + SOAP client
    ca/                 CA FTB XML + REST client
  pdf/                  PDF parsing and filling
  llm/                  OpenRouter API client with caching
  security/             AES-256-GCM encryption, audit trail
  state/                JSON state persistence
  tui/                  Bubble Tea views
  errors/               Typed errors (unsupported, incomplete, CPA referral)
pkg/taxmath/            Pure math (brackets, rounding, tables, expat stacking)
testdata/               20 test scenarios with known-good inputs/outputs
data/                   Year-specific tax data, prompts, schemas
```

### Key Design Principle: AI Does UX, Not Math

The LLM is used for generating explanations, contextual help, and searching tax code. It is **never** used for tax calculations, form field values, or XML generation. All math is deterministic Go code with table-driven tests.

### Form Logic is a Dependency Graph

Each form field is typed as `UserInput`, `Computed`, `FederalRef`, `Lookup`, or `PriorYear`. The solver walks the DAG across all registered forms, computing values in topological order. State forms reference federal fields via `FederalRef` dependencies.

## Quick Start

```bash
# Prerequisites: Go 1.26+

# Start interactive interview
go run ./cmd/taxpilot

# Import prior-year return for pre-fill
go run ./cmd/taxpilot --import prior.pdf

# Resume from saved state
go run ./cmd/taxpilot --continue

# Export filled PDFs
go run ./cmd/taxpilot --export out/

# Validate return
go run ./cmd/taxpilot --validate

# E-file (requires provider registration)
go run ./cmd/taxpilot --efile
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `OPENROUTER_API_KEY` | Required for AI-powered explanations (interview works without it — static help text instead) |
| `TAXPILOT_MODEL` | Override the default LLM model (default: `anthropic/claude-sonnet-4.6`) |

You can also pass `--model <model>` on the command line, which takes precedence over both the default and `TAXPILOT_MODEL`.

All API requests use OpenRouter's Zero Data Retention (ZDR) routing — prompts and completions are never stored by providers.

## Testing

```bash
go test ./...                                  # Full suite (12 packages)
go test ./internal/forms/... -v                # All form logic
go test ./internal/forms/... -run TestExpat    # Expat scenarios
go test ./internal/efile/... -v                # E-file XML generation
go test ./internal/interview/... -v            # Interview + screening
go test ./internal/knowledge/... -v            # Knowledge base + RAG
go test ./pkg/taxmath/... -v                   # Tax math utilities
```

## Expat Support

Full support for Americans living abroad:

- **FEIE (Form 2555)**: Up to $130,000 exclusion (2025), bona fide residence and physical presence tests, prorated exclusion, housing exclusion/deduction, tax stacking
- **FTC (Form 1116)**: Credit limitation formula, carryforward, no double-dip with FEIE
- **FATCA (Form 8938)**: Dual thresholds for abroad vs US residents
- **FBAR guidance**: Detects when FinCEN 114 filing is required, provides instructions
- **Treaty disclosure (Form 8833)**: US-Sweden treaty and other treaty positions
- **CA non-conformity**: California does not allow the FEIE — automatic add-back on Schedule CA

## Installation

```bash
# Via Homebrew
brew install christopherklint97/tap/taxpilot

# Or build from source (requires Go 1.26+)
go install github.com/christopherklint97/taxpilot/cmd/taxpilot@latest
```

## Security

- All data stays local (no cloud sync)
- State files encrypted at rest (AES-256-GCM + Argon2id)
- Audit trail distinguishes AI-suggested vs user-entered values
- E-file credentials stored encrypted

## Status

This is a working implementation with all core tax logic, interview flow, PDF export, and e-file infrastructure complete. Production e-filing requires IRS EFIN registration and CA FTB provider enrollment (bureaucratic processes documented in `docs/efile-provider-guide.md`).

## License

MIT
