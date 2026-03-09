# TaxPilot Architecture Improvements

## Status Key
- [ ] Not started
- [x] Complete

---

## 1. Extract Shared Helper Functions
**Impact:** High | **Effort:** Low | **Category:** Readability

Extract duplicated `getStr()`, `getNum()`, `numExists()` from `validate.go`, `reasonableness.go`, and `errors.go` into a shared package.

- [x] Create `internal/forms/helpers.go` with shared helper functions
- [x] Update `internal/efile/validate.go` to use shared helpers
- [x] Update `internal/efile/reasonableness.go` to use shared helpers
- [x] Update `internal/errors/errors.go` to use shared helpers
- [x] Verify all tests pass

## 2. Add ValueType Metadata to FieldDef
**Impact:** High | **Effort:** Medium | **Category:** Architecture

Add `ValueType` (Numeric/String/Enum) to `FieldDef` so string fields are self-documenting instead of tracked in a separate `stringFields` map.

- [x] Add `FieldValueType` type and constants to `internal/forms/field.go`
- [x] Update `FieldDef` struct with `ValueType` field
- [x] Mark all string fields in form definitions with `ValueType: forms.StringValue`
- [x] Update `interview/engine.go` to derive `stringFields` from form metadata
- [x] Verify all tests pass

## 3. Extend Registry Sync Tests to Cover PDF & XML
**Impact:** High | **Effort:** Low | **Category:** Scalability

Add tests verifying all computed forms have PDF mappings and XML builders.

- [x] Add test checking PDF mappings cover all non-input forms
- [x] Add test checking MeF XML covers all federal forms
- [x] Add test checking CA XML covers all CA forms
- [x] Verify all tests pass

## 4. Externalize Year-Specific Constants
**Impact:** High | **Effort:** Medium | **Category:** Architecture

Move hardcoded tax year values (HSA limits, FEIE limit, FATCA thresholds, etc.) into `data/tax_years/YYYY/` YAML files and load them at runtime.

- [x] Define `TaxYearConfig` struct in `pkg/taxmath/config.go`
- [x] Add constants to `data/tax_years/2025/federal/limits.yaml` and `ca/limits.yaml`
- [x] Update `reasonableness.go` to use centralized config
- [x] Update `taxmath/expat.go` to use centralized config
- [x] Update `form_8938.go` to use centralized config
- [x] Verify all tests pass

## 5. Implement StateFormSet Interface
**Impact:** High | **Effort:** Medium | **Category:** Scalability

Implement the existing `StateFormSet` interface for CA, route CA-specific logic through it.

- [x] CA already implements `StateFormSet` in `internal/forms/state/ca/f540.go` — added compile-time check
- [x] Create state registry in `internal/forms/state/registry.go`
- [x] Add `SetupStateRegistry()` to interview engine
- [x] Verify all tests pass

## 6. Declarative Field Builders
**Impact:** Medium | **Effort:** Medium | **Category:** Readability

Create helper functions for common field patterns to reduce form boilerplate.

- [x] Create `internal/forms/builders.go` with field builder functions
- [x] Refactor federal form definitions to use builders where applicable
- [x] Refactor CA form definitions to use builders where applicable
- [x] Verify all tests pass

## 7. Add Strict Mode to DepValues
**Impact:** Medium | **Effort:** Low | **Category:** Architecture

Add `GetStrict()` to `DepValues` that returns an error for missing keys.

- [x] Add `GetStrict()` method to `DepValues` in `internal/forms/field.go`
- [x] Add validation in solver that checks all dependency keys exist
- [x] Add tests for strict mode
- [x] Verify all tests pass

## 8. Unify Validation and Reasonableness Pipeline
**Impact:** Medium | **Effort:** Medium | **Category:** Architecture

Merge `ValidateReturn()` and `ReasonablenessCheck()` into a unified pipeline.

- [ ] Create unified `FullValidation()` function
- [ ] Update callers to use unified function
- [ ] Verify all tests pass

## 9. Metadata-Driven Question Ordering
**Impact:** Medium | **Effort:** Medium | **Category:** Scalability

Move question grouping and ordering into `FormDef` metadata.

- [x] Add `QuestionGroup` and `QuestionOrder` to `FormDef`
- [x] Set metadata on all form definitions
- [x] Refactor `buildQuestions()` to use metadata
- [x] Verify all tests pass

## 10. FederalRef Validation
**Impact:** Low-Medium | **Effort:** Low | **Category:** Readability

Add validation that `FederalRef` fields actually reference federal forms.

- [x] Add validation in registry that FederalRef dependencies point to federal forms
- [x] Add test for this validation
- [x] Verify all tests pass

## 11. Index Fields by Line in FormDef
**Impact:** Low-Medium | **Effort:** Low | **Category:** Scalability

Add `fieldIndex` map for O(1) field lookups instead of linear scan.

- [x] Add field index to `FormDef` with lazy initialization
- [x] Update `GetField()` to use index
- [x] Verify all tests pass

## 12. Add Solver Benchmarks & Optimize
**Impact:** Low | **Effort:** Low | **Category:** Scalability

Replace custom insertion sort with standard library, add benchmarks.

- [x] Replace `sortStrings()` and `insertSorted()` with `slices.Sort`
- [x] Add benchmark tests for solver
- [x] Verify all tests pass

## 13. Add XML Determinism Tests
**Impact:** Low | **Effort:** Low | **Category:** Architecture

Add tests verifying MeF and CA XML generation is byte-for-byte reproducible.

- [x] Add MeF XML determinism test
- [x] Add CA XML determinism test
- [x] Verify all tests pass

## 14. Form Registration via init()
**Impact:** Low | **Effort:** Low | **Category:** Scalability

Use Go `init()` functions for automatic form registration.

- [x] Create auto-registration mechanism
- [x] Convert form files to use init() registration
- [x] Update SetupRegistry() to use auto-registered forms
- [x] Verify all tests pass

## 15. Add Error/Edge-Case Test Scenarios
**Impact:** Low | **Effort:** Medium | **Category:** Architecture

Add test scenarios for error paths and edge cases.

- [x] Add invalid input scenario
- [x] Add FEIE+FTC conflict scenario
- [x] Add missing dependency scenario
- [x] Verify all tests pass

## 16. Remove Unused StateFormSet Interface (Superseded by #5)
**Impact:** Low | **Effort:** Low | **Category:** Readability

This item is superseded by #5 — implementing the interface rather than removing it.

- [x] Superseded by item #5

## 17. Centralized FormRegistry
**Impact:** Low | **Effort:** High | **Category:** Architecture

Create a single FormRegistry tracking all metadata per form.

- [x] Create `FormRegistration` struct with all metadata
- [x] Create centralized `FormRegistry`
- [x] Migrate registration points to use centralized registry
- [x] Verify all tests pass
