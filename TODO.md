# TaxPilot — Project TODO

> Master tracking document for all implementation phases.
> Mark items `[x]` when complete. Add date of completion.

---

## Phase 1 — Foundation
**Goal:** Single simple federal + CA return (single filer, one W-2, standard deduction) end-to-end with PDF output.

### 1.1 Project Scaffolding
- [x] Initialize Go module + directory structure
- [x] Create CLAUDE.md
- [x] Set up Cobra CLI with root command and flags (`start`, `continue`, `export`, `efile`, `validate`)
- [x] Set up Bubble Tea app shell with view routing
- [x] Add .gitignore
- [x] Install core dependencies (cobra, bubbletea, lipgloss, pdfcpu, yaml.v3)

### 1.2 Core Form Engine
- [x] Define `FieldType` enum (UserInput, Computed, Lookup, PriorYear, FederalRef)
- [x] Define `FieldDef` struct (Line, Type, Label, Prompt, DependsOn, Options, Compute)
- [x] Define `FormDef` struct (ID, Name, Jurisdiction, TaxYears, Fields)
- [x] Define `Jurisdiction` type (Federal, StateCA, ...)
- [x] Implement `Registry` — register forms, lookup by ID
- [x] Implement `DependencyGraph` — build DAG from all registered form fields
- [x] Implement `Solver` — topological sort, resolve computed fields, collect missing UserInputs
- [x] Unit tests for graph solver (cycle detection, cross-form deps, cross-jurisdiction deps)

### 1.3 Tax Math Package
- [x] `pkg/taxmath/brackets.go` — bracket-based tax computation (federal + CA)
- [x] `pkg/taxmath/rounding.go` — IRS/FTB rounding rules (round to nearest dollar)
- [x] `pkg/taxmath/tables.go` — standard deductions, exemption amounts, limits by year + jurisdiction
- [x] YAML reference data files in `data/tax_years/` (Go uses hardcoded values)
- [x] Unit tests for bracket calculations (federal + CA) — 15 tests

### 1.4 Federal Forms (MVP)
- [x] W-2 input form (`internal/forms/inputs/w2.go`) — wages, withholding, state boxes
- [x] Form 1040 — lines 1-37 (income through tax/refund)
- [x] Register forms in registry (via interview.SetupRegistry)
- [x] Test scenario: `testdata/scenarios/federal/single_w2.json`
- [x] Expected output: `testdata/expected/federal/single_w2.json`

### 1.5 California Forms (MVP)
- [x] CA Form 540 — lines through tax computation (including mental health surcharge)
- [x] Schedule CA (540) — basic version (passthrough when no adjustments needed)
- [x] CA brackets data (`data/tax_years/2025/ca/brackets.yaml`)
- [x] CA conformity differences (`internal/forms/state/ca/conformity.go`)
- [x] Register CA forms in registry (via interview.SetupRegistry)
- [x] Test scenario: `testdata/scenarios/ca/ca_single_w2.json` + `ca_high_income.json`
- [x] Expected output: `testdata/expected/ca/ca_single_w2.json` + `ca_high_income.json`

### 1.6 Basic Interview Loop
- [x] Walk dependency graph for missing `UserInput` fields
- [x] Ask federal questions first, then state-specific (ordered: filing status, personal, W-2)
- [x] Present questions in TUI with progress bar and formatting
- [x] Save/load state to `~/.taxpilot/state.json`
- [x] Resume from saved state (`--continue`)

### 1.7 PDF Output
- [x] AcroForm field mappings for 1040 and 540 (`internal/pdf/mappings.go`)
- [x] Implement PDF filler with text fallback (`internal/pdf/filler.go`)
- [x] `--export` flag generates filled text exports for both jurisdictions
- [x] 7 PDF/export tests passing
- [ ] Download blank 1040 PDF template (when 2025 forms published)
- [ ] Download blank 540 PDF template (when 2025 forms published)

### 1.8 Year-Specific Data Files
- [x] `data/tax_years/2025/federal/brackets.yaml`
- [x] `data/tax_years/2025/federal/limits.yaml`
- [x] `data/tax_years/2025/ca/brackets.yaml`
- [x] `data/tax_years/2025/ca/limits.yaml`
- [ ] `data/tax_years/2025/ca/conformity.yaml` (documented in Go, YAML not yet needed)

---

## Phase 2 — Prior-Year Context
**Goal:** Import prior-year return PDFs (federal + CA) and use them to pre-fill and contextualize the current year.

### 2.1 PDF Parser
- [x] Extract field values from filled IRS 1040 PDFs (`internal/pdf/parser.go`)
- [x] Extract field values from filled CA 540 PDFs (auto-detection via field matching)
- [x] Map extracted values to internal state format (ReverseMapping)
- [x] Handle AcroForm fields (pdfcpu ExportForm)
- [x] ParseCurrency, ParseSSN helpers with 22 parser tests
- [ ] OCR fallback for printed/scanned forms (stretch)

### 2.2 State Migration
- [x] Define which fields carry over year-to-year (`CarryoverFields` in `internal/state/migrate.go`)
- [x] Build migration logic: `MigrateToCurrentYear()` carries filing status, personal info, employer info
- [x] Store prior-year CA AGI separately (`PriorYearCAAGI` in `PriorYearContext`)
- [x] `PriorYearStore` — save/load/check prior-year data in `~/.taxpilot/prior_years/<year>/`
- [x] Flag significant changes: `CompareReturns()` with 20% threshold, severity classification
- [x] 17 state/migration tests passing

### 2.3 Pre-fill in Interview
- [x] `NewEngineWithPriorYear()` — engine with prior-year defaults
- [x] `GetPriorYearDefault()` — returns prior-year value for current question
- [x] `AcceptDefault()` — accept prior-year value with Enter (empty input)
- [x] TUI shows "Last year: $X" with "Press Enter to keep" hint
- [x] Welcome screen: [L] load prior-year PDF, shows loaded status
- [x] `--import` CLI flag parses PDF and saves prior-year context
- [x] TUI `ImportPriorYearMsg`/`PriorYearImportedMsg` message flow
- [ ] CA-specific pre-fill messages (e.g., "CA made no adjustments last year")

---

## Phase 3 — LLM Interview Layer
**Goal:** Contextual questions, explanations, and tax code guidance for both federal and CA.

### 3.1 OpenRouter API Integration
- [x] OpenRouter client (`internal/llm/client.go`) — OpenAI-compatible API at openrouter.ai
- [x] `OPENROUTER_API_KEY` env var authentication
- [x] Default model: `anthropic/claude-sonnet-4` (configurable via `SetModel`)
- [x] Response cache (`internal/llm/cache.go`) — SHA-256 keyed, in-memory + disk persistence
- [x] Cache stored in `~/.taxpilot/llm_cache/`
- [x] System prompts: `data/prompts/interview_system.txt`, `explainer_system.txt`, `ca_adjustments.txt`
- [x] 19 LLM tests passing (mock HTTP server, no real API key needed)

### 3.2 Contextual Question Generation
- [x] 14 contextual prompts with HelpText + CANote (`internal/interview/questions.go`)
- [x] Enhanced prompts for all W-2 fields, personal info, and filing status
- [x] Prior-year context builder (`internal/interview/context.go`) — `BuildContextSummary`, `FormatForLLM`
- [x] CA-specific notes shown in TUI when filing in CA (gold italic style)
- [x] "?" command in interview shows contextual help text

### 3.3 Situation Detection & Form Routing
- [x] `Situation` and `ScreeningQuestion` types (`internal/interview/situationdetect.go`)
- [x] `EvaluateScreening()` framework for triggering additional forms
- [ ] Actual screening questions (deferred to Phase 5 when more forms are added)
- [ ] Auto-detect when Schedule CA adjustments are needed

### 3.4 Tax Code Explainer
- [x] `Explainer` with `ExplainField`, `ExplainCADifference`, `ExplainWhyAsked` (`internal/llm/explainer.go`)
- [x] All methods cache-first, then LLM call, then cache result
- [x] CA adjustments context auto-injected for CA difference explanations
- [ ] "Why?" handler in TUI calls explainer via LLM (currently shows static help text)
- [ ] IRC and CA R&TC section references with plain-English summaries
- [ ] CA <-> federal difference explanations

---

## Phase 4 — Knowledge Base / RAG
**Goal:** Index federal and CA tax code and publications for on-demand lookup.

### 4.1 Content Extraction
- [x] 26 federal seed documents (IRC sections, W-2 guide, Form 1040 overview)
- [x] 15 CA seed documents (conformity, rates, Schedule CA, mental health tax, etc.)
- [x] `SeedStore()` convenience function pre-populates knowledge base
- [ ] Script: download IRS form instructions (PDF -> text) — `scripts/extract-instructions.sh` stub created
- [ ] Script: download FTB form instructions (PDF -> text)
- [ ] Extract full IRC sections from Title 26
- [ ] Extract IRS Publications (Pub 17, 334, 505)
- [ ] Extract FTB Publications (Pub 1001, etc.)

### 4.2 Search Pipeline
- [x] TF-IDF keyword search with inverted index (`internal/knowledge/store.go`)
- [x] Weighted scoring: Title 3x, Source/Tags 2x, Content 1x, length-normalized
- [x] Tokenization with stop word filtering
- [x] Jurisdiction-scoped search (federal only, CA only, or all)
- [x] JSON file save/load round-tripping
- [x] 9 knowledge base tests passing
- [ ] Upgrade to vector embeddings via SQLite vec extension (stretch)

### 4.3 RAG Query Interface
- [x] `RAG.Query()` — search + LLM synthesis (`internal/knowledge/rag.go`)
- [x] `RAG.QueryForField()` — field-specific knowledge retrieval
- [x] `ExplainWithContext()` — formats context for LLM prompt
- [x] Cross-jurisdiction context (CA queries also pull federal refs)
- [x] "??" command in TUI triggers async RAG query
- [x] Wired in `cmd/taxpilot/main.go` — creates RAG if `OPENROUTER_API_KEY` set
- [x] Graceful fallback: no API key = static help only

---

## Phase 5 — Expand Form Coverage
**Goal:** Support the most common filing scenarios, federal + CA.

### 5.1 Federal Forms
- [ ] Schedule A (Itemized Deductions)
- [ ] Schedule B (Interest & Dividends) + 1099-INT, 1099-DIV inputs
- [ ] Schedule C (Business Income) + 1099-NEC input
- [ ] Schedule D (Capital Gains) + 1099-B input + Form 8949
- [ ] Schedule SE (Self-Employment Tax)
- [ ] Schedule 1-3 (Additional Income, Adjustments, Credits)
- [ ] Form 8889 (HSA)
- [ ] Form 8995 (QBI Deduction)

### 5.2 CA Forms (Corresponding)
- [ ] Schedule CA Part II (Itemized deduction adjustments)
- [ ] CA interest/dividend adjustments in Schedule CA Part I
- [ ] CA Schedule CA business income adjustments
- [ ] Schedule D-1 (CA capital gain differences)
- [ ] CA self-employment conformity verification
- [ ] Form 3514 (CA EITC) + Form 3853 (Health Coverage)
- [ ] CA HSA conformity check
- [ ] Schedule CA QBI add-back (CA does not allow Section 199A)

### 5.3 Per-Form Checklist
For each form pair above:
- [ ] Field definitions and compute logic (federal + CA)
- [ ] PDF field mappings for both jurisdictions
- [ ] Test scenarios — especially conformity edge cases
- [ ] Interview questions and LLM context

---

## Phase 6 — E-File Integration
**Goal:** Transmit returns electronically to IRS and CA FTB.

### 6.1 IRS MeF Integration
- [ ] MeF XML serialization from internal form state
- [ ] SOAP client for MeF A2A interface (SendSubmissions, GetAcknowledgements)
- [ ] Strong Authentication certificate management
- [ ] Pre-submission validation against MeF business rules
- [ ] Form 8879 (IRS e-file Signature Authorization) — self-select PIN
- [ ] ATS test mode for certification testing
- [ ] Rejection handling: parse codes, user messages, correction + resubmission

### 6.2 CA FTB Integration
- [ ] CA e-file XML serialization (FTB specs from SES)
- [ ] FTB transmission client
- [ ] FTB 8879 (CA e-file Signature Authorization)
- [ ] Shared secret authentication (prior-year CA AGI)
- [ ] PATS test mode for certification
- [ ] CA-specific rejection codes and acknowledgement handling

### 6.3 E-File TUI Flow
- [ ] Pre-submission review screen (federal + CA summaries)
- [ ] Validation results with clear error messages
- [ ] Signature authorization flow (self-select PIN)
- [ ] Submission progress indicator
- [ ] Status tracking view (pending -> accepted/rejected)
- [ ] Rejection resolution workflow

### 6.4 Provider Registration
- [ ] `scripts/efile-setup.sh` — interactive guide for IRS e-Services, EFIN, ATS, CA LOI, PATS
- [ ] Begin IRS e-file provider application (bureaucratic process, start early)

---

## Phase 7 — Polish & Safety
**Goal:** Make it reliable and trustworthy.

### 7.1 Validation Layer
- [ ] Cross-check computed values against IRS + FTB reasonableness checks
- [ ] Flag unusual values (charitable > 60% AGI, etc.)
- [ ] Verify federal <-> CA consistency
- [ ] Warn about common audit triggers
- [ ] Run MeF business rules before e-file

### 7.2 Review Mode
- [ ] Summary view — all federal and CA forms with key numbers
- [ ] Side-by-side federal vs. CA comparison
- [ ] Side-by-side with prior year
- [ ] Highlight changes and flag unusual items

### 7.3 Error Handling
- [ ] Graceful handling of missing/incomplete data
- [ ] Clear messages for unsupported situations
- [ ] "I can't handle this — tell your CPA" fallback
- [ ] CA conformity edge cases: clear messaging

### 7.4 Security
- [ ] All data stays local (no cloud sync)
- [ ] State files encrypted at rest with user passphrase
- [ ] E-file credentials in OS keychain
- [ ] Prior-year CA AGI stored encrypted
- [ ] Audit trail: AI-suggested vs. user-entered
