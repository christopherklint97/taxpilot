# TaxPilot — Claude Code Instructions

## Project Overview
TaxPilot is a full-stack web application that helps US taxpayers fill out and
e-file federal and California state income tax forms. Users upload prior-year
PDFs, view forms with a side-by-side PDF editor, edit fields with automatic
dependency recalculation, and export filled PDFs or e-file to IRS/FTB.

## Tech Stack
- **Backend:** Rust (Axum + rusqlite + SQLite)
- **Frontend:** React 19, TypeScript, TanStack Router, TanStack Query, Zustand, shadcn/ui, TailwindCSS 4
- **PDF handling:** lopdf/pdf-form (Rust) + pdf.js (browser rendering)
- **LLM integration:** OpenRouter API via reqwest
- **XML generation:** quick-xml (for MeF and CA e-file submissions)
- **Deployment:** Docker Compose (api + web + Tailscale sidecar)
- **Ports:** API on 4100, Web on 4101

## Key Principles

### 1. AI Does UX, Not Math
The LLM is used for:
- Generating plain-English explanations of tax questions
- Searching/explaining relevant tax code sections (federal & CA)
- Explaining CA <-> federal differences in plain English

The LLM is NEVER used for:
- Tax calculations (use deterministic Rust code)
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

The solver walks the dependency graph across BOTH federal and state forms.
When a field changes, all dependent fields are recomputed automatically.

### 3. Federal and State Are Separate but Connected
- Federal forms live in `api/src/forms/federal/`
- State forms live in `api/src/forms/state/ca/`
- State forms can reference federal fields via `FederalRef` dependencies
- Conformity module maps which federal values CA accepts as-is
  vs. which require California adjustments (Schedule CA)

### 4. Expat/Foreign Income Is Fully Supported
- Form 2555 (FEIE) with tax stacking
- Form 1116 (FTC) with credit limitation and carryforward
- Form 8938 (FATCA) with dual thresholds
- Form 8833 (treaty disclosure)
- CA FEIE non-conformity (automatic add-back on Schedule CA)

### 5. Test Against Known-Good Returns
Every form module must have table-driven tests using scenarios from
testdata/scenarios/. Each scenario has inputs and expected outputs.

### 6. Year-Agnostic Form Definitions
Keep year-specific data in data/tax_years/YYYY/{federal,ca}/ YAML files.
Form logic in Rust references these via the taxmath module.

### 7. E-File XML Must Be Deterministic
Same inputs always produce identical XML. Never use the LLM for
anything in the e-file pipeline.

## Supported Forms

### Federal
Form 1040, Schedules 1-3/A/B/C/D/SE, Forms 8949/8889/8995,
Form 2555 (FEIE), Form 1116 (FTC), Form 8938 (FATCA), Form 8833 (Treaty)

### Input Forms
W-2, 1099-INT, 1099-DIV, 1099-NEC, 1099-B

### California
Form 540, Schedule CA, Form 3514 (CalEITC), Form 3853 (Health Coverage)

## Directory Guide
- `api/` — Rust backend
  - `api/src/main.rs` — Axum app setup, routes, middleware
  - `api/src/db/` — SQLite connection, schema, migrations
  - `api/src/domain/` — Core types: field.rs, form.rs, solver.rs, taxmath.rs, registry.rs
  - `api/src/forms/` — Form definitions (federal/, inputs/, state/ca/)
  - `api/src/pdf/` — PDF parsing and filling
  - `api/src/efile/` — E-file XML generation (mef/, ca/) and validation
  - `api/src/routes/` — REST API endpoint handlers
  - `api/src/llm/` — OpenRouter LLM client
- `web/` — React frontend
  - `web/src/router.tsx` — TanStack Router setup
  - `web/src/stores/` — Zustand stores (return-store, field-store, ui-store)
  - `web/src/routes/` — Page components
  - `web/src/components/` — Layout, forms, PDF viewer, UI components
  - `web/src/api/` — Typed API client and types
  - `web/src/lib/` — Utility functions
- `data/` — Year-specific tax data, PDF templates, prompts
- `testdata/` — Test scenarios with known-good inputs/outputs

## API Endpoints
```
GET    /api/health                          Health check
GET    /api/returns                         List returns
POST   /api/returns                         Create return
GET    /api/returns/:id                     Get return + fields
DELETE /api/returns/:id                     Delete return
PUT    /api/returns/:id/fields/:key         Update field, recompute dependents
PUT    /api/returns/:id/fields              Batch update fields
GET    /api/forms                           List form metadata
GET    /api/forms/:formId                   Get form definition
POST   /api/returns/:id/pdf/upload          Upload prior-year PDF
GET    /api/returns/:id/pdf/filled/:formId  Generate filled PDF
POST   /api/returns/:id/rollforward         Rollforward from prior year
GET    /api/returns/:id/validate            Run validation checks
POST   /api/returns/:id/efile/mef           Generate MeF XML
POST   /api/returns/:id/efile/ca            Generate CA FTB XML
```

## Common Tasks

### Adding a new federal form
1. Create `api/src/forms/federal/new_form.rs`
2. Define all fields with their types (UserInput/Computed/Lookup/PriorYear)
3. Implement Compute closures for all Computed fields
4. Register in `api/src/forms/mod.rs` via `register_all_forms()`
5. Add PDF field mappings
6. Add MeF XML element mappings in `api/src/efile/mef/xml.rs`
7. Add test scenario in `testdata/scenarios/federal/`
8. Run `cd api && cargo test`

### Adding a new CA state form
1. Create `api/src/forms/state/ca/new_form.rs`
2. Define fields — use `FederalRef` type for fields that pull from federal forms
3. Check conformity rules for CA <-> federal differences
4. Register in `api/src/forms/mod.rs`
5. Add CA e-file XML mappings in `api/src/efile/ca/xml.rs`
6. Add test scenarios — MUST include conformity edge cases
7. Run `cd api && cargo test`

## Running
```bash
# Development
cd api && cargo run                    # Start API on :4100
cd web && pnpm dev                     # Start frontend on :4101 (proxies /api to :4100)

# Docker
docker compose up -d api web           # Start API + web
docker compose up -d                   # Start all (including Tailscale)
docker compose down                    # Stop all

# Testing
cd api && cargo test                   # All Rust tests
cd web && pnpm build                   # TypeScript check + build
```
