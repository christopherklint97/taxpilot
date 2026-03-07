# Expat/Foreign Income Support — Implementation Plan

## Overview

This plan adds full support for Americans living abroad filing US federal + California taxes. It covers 6 new forms, updates to 25+ existing files, new CA conformity rules, interview flows, knowledge base entries, test scenarios, PDF mappings, e-file XML structures, and validation updates.

**Target user profile:** US citizen living in Sweden with foreign earned income only, foreign housing, foreign savings accounts, and foreign pension plans.

---

## Phase 1: Form 2555 (Foreign Earned Income Exclusion)

**Status:** Not started
**Complexity:** HIGH — this is the foundation for everything else
**Dependencies:** None

### Summary

Form 2555 is the core expat form. It allows US citizens/residents living abroad to exclude up to $130,000 (2025) of foreign earned income from US tax, plus a housing exclusion/deduction.

### New Files

#### 1. `internal/forms/federal/form_2555.go`

Form ID: `"form_2555"`

**UserInput fields:**
- `foreign_country` (string) — country of residence (e.g., "Sweden")
- `foreign_address` (string) — foreign address
- `employer_name` (string) — foreign employer name
- `employer_foreign` (string, "yes"/"no") — is employer foreign?
- `qualifying_test` (string, options: `["bona_fide_residence", "physical_presence"]`)
- `bfrt_start_date` (string) — date bona fide residence began
- `bfrt_end_date` (string) — date it ends (usually ongoing)
- `bfrt_full_year` (string, "yes"/"no") — full tax year?
- `ppt_days_present` (numeric) — days in foreign country during 12-month period
- `ppt_period_start` (string) — start of 12-month period
- `ppt_period_end` (string) — end of 12-month period
- `foreign_earned_income` (numeric) — total foreign earned income
- `employer_provided_housing` (numeric) — employer-provided housing amounts
- `housing_expenses` (numeric) — qualifying housing expenses
- `self_employed_abroad` (string, "yes"/"no")
- `foreign_tax_paid` (numeric) — taxes paid to foreign government
- `currency_code` (string) — e.g., "SEK"
- `exchange_rate` (numeric) — average exchange rate used

**Computed fields:**
- `exclusion_limit` — $130,000 for 2025 (from taxmath)
- `qualifying_days` — from PPT days or 365 for full-year BFRT
- `prorated_exclusion` — `exclusion_limit * qualifying_days / 365`
- `foreign_income_exclusion` — `min(foreign_earned_income, prorated_exclusion)`
- `housing_base_amount` — 16% of exclusion limit ($20,800)
- `housing_max` — 30% of exclusion limit ($39,000) default
- `housing_qualifying_amount` — `max(0, min(housing_expenses, housing_max) - housing_base_amount)`
- `housing_exclusion` — `min(housing_qualifying_amount, employer_provided_housing)`
- `housing_deduction` — `max(0, housing_qualifying_amount - housing_exclusion)` (self-employed only)
- `total_exclusion` — `foreign_income_exclusion + housing_exclusion`

**Integration with 1040:**
- Total exclusion flows to Schedule 1 line 8d (new line)
- Schedule 1 line 8d reduces total income

#### 2. `pkg/taxmath/expat.go`

Pure math utilities:
- `FEIELimit(taxYear int) float64` — $130,000 for 2025
- `HousingBaseAmount(taxYear int) float64` — 16% of FEIE limit
- `HousingMaxAmount(taxYear int) float64` — 30% of FEIE limit (default)
- `ProrateExclusion(limit float64, qualifyingDays, totalDays int) float64`
- `ComputeTaxWithStacking(taxableIncome, excludedIncome float64, fs FilingStatus, year int, jurisdiction JurisdictionType) float64` — tax stacking method

### Modified Files

#### 3. `internal/forms/federal/schedule_1.go`
- Add line `8d`: Foreign earned income exclusion (from Form 2555 total_exclusion)
- Update line `10` (total additional income) to include `schedule_1:8d`

#### 4. `internal/forms/federal/f1040.go`
- Line 16 tax computation: use stacking method when FEIE is claimed
- Tax on remaining income computed at the rate as if excluded income were still included

#### 5. `internal/interview/engine.go`
- Register Form 2555 in `SetupRegistry()`
- Add Form 2555 fields to `buildQuestions()` (new `form2555Questions` category)
- Add string fields: `foreign_country`, `foreign_address`, `bfrt_start_date`, `bfrt_end_date`, `ppt_period_start`, `ppt_period_end`, `qualifying_test`, `employer_foreign`, `bfrt_full_year`, `self_employed_abroad`, `currency_code`

#### 6. `internal/forms/integration_test.go`
- Register `federal.Form2555()` in `buildSolver()`
- Add Form 2555 defaults to `inputFormDefaults`

### Tests

- `TestForm2555BFRTFullYear` — full year bona fide residence, $130k exclusion
- `TestForm2555PPT330Days` — physical presence test with 330 days
- `TestForm2555PPTProrated` — PPT with less than full year (prorated exclusion)
- `TestForm2555HousingExclusion` — employer-provided housing
- `TestForm2555HousingDeduction` — self-employed housing deduction
- `TestForm2555ExclusionLimit` — income exceeds exclusion limit
- `TestForm2555Integration` — full 1040 flow with FEIE
- `TestTaxStacking` — verify stacking method produces correct tax
- Unit tests in `pkg/taxmath/expat_test.go` for all pure math functions

### Test Scenarios

- `testdata/scenarios/federal/expat_sweden_bfrt.json` — Sweden expat, BFRT, full year, $120k foreign income
- `testdata/expected/federal/expat_sweden_bfrt.json` — expected values

---

## Phase 2: Form 1116 (Foreign Tax Credit) + Schedule B Part III

**Status:** Not started
**Complexity:** HIGH
**Dependencies:** Phase 1

### Summary

Form 1116 computes the foreign tax credit. A taxpayer cannot claim both FEIE and FTC on the same income — FTC only applies to income NOT excluded by FEIE.

### New Files

#### 1. `internal/forms/federal/form_1116.go`

Form ID: `"form_1116"`

**UserInput fields:**
- `category` (string, options: `["general", "passive", "section_901j", "treaty_sourced"]`)
- `foreign_country` (string)
- `foreign_tax_paid_income` (numeric)
- `foreign_tax_paid_other` (numeric)
- `foreign_source_income` (numeric) — gross foreign source income
- `foreign_source_deductions` (numeric) — expenses allocated to foreign source
- `accrued_or_paid` (string, options: `["paid", "accrued"]`)

**Computed fields:**
- `net_foreign_source` — `foreign_source_income - foreign_source_deductions`
- `total_foreign_tax` — sum of tax paid fields
- `credit_limitation` — `US_tax * (foreign_source_income / worldwide_income)`
- `credit_allowed` — `min(total_foreign_tax, credit_limitation)`

**Cross-form interaction:** `foreign_source_income` must exclude FEIE-excluded income.

### Modified Files

#### 2. `internal/forms/federal/schedule_b.go`
Add Part III fields:
- `7a` (UserInput, "yes"/"no") — foreign accounts question
- `7b` (UserInput, string) — country name(s)
- `8` (UserInput, "yes"/"no") — foreign trust question
- `fbar_required` (Computed) — true if 7a is "yes"

#### 3. `internal/forms/federal/schedule_3.go`
- Wire line 1 (foreign tax credit) to Form 1116 output

#### 4. `internal/interview/engine.go`
- Register Form 1116, add string fields

### Tests
- `TestForm1116BasicCredit` — simple foreign tax credit calculation
- `TestForm1116CreditLimitation` — credit limited by US tax ratio
- `TestForm1116With2555Interaction` — FTC on non-excluded income
- `TestScheduleBPart3ForeignAccounts` — triggers FBAR requirement

### Test Scenarios
- `testdata/scenarios/federal/foreign_tax_credit.json` — FTC-only scenario

---

## Phase 3: Form 8938 (FATCA), Form 8833 (Treaty), FBAR Guidance

**Status:** Not started
**Complexity:** MEDIUM
**Dependencies:** Phases 1-2

### New Files

#### 1. `internal/forms/federal/form_8938.go`

Form ID: `"form_8938"` — FATCA Statement of Specified Foreign Financial Assets

Filing thresholds (living abroad):
- Single: $200,000 year-end or $300,000 at any time
- MFJ: $400,000 year-end or $600,000 at any time

**UserInput fields:**
- `num_financial_accounts`, `max_value_accounts`, `yearend_value_accounts`
- `num_other_assets`, `max_value_other`, `yearend_value_other`
- `account_country` (string), `account_institution` (string), `account_type` (string)
- `income_from_accounts`, `gain_from_accounts`

**Computed fields:**
- `total_max_value`, `total_yearend_value`
- `filing_required` — based on thresholds and filing status

#### 2. `internal/forms/federal/form_8833.go`

Form ID: `"form_8833"` — Treaty-Based Return Position Disclosure

US-Sweden treaty key provisions:
- Article 15: Employment income
- Article 18: Pensions
- Article 23: Elimination of double taxation

**UserInput fields:**
- `treaty_country` (string), `treaty_article` (string), `irc_provision` (string)
- `treaty_position_explanation` (string), `treaty_amount` (numeric)

#### 3. `internal/interview/fbar_guidance.go`

FBAR (FinCEN 114) is filed separately with FinCEN, not the IRS.
- `FBARRequired()` — detects if aggregate foreign accounts > $10,000
- `FBARGuidanceMessage()` — filing instructions
- `FBARDeadline()` — April 15 with auto extension to October 15

### Tests
- `TestForm8938ThresholdAbroad` — filing required above $200k/$300k
- `TestForm8833SwedenPension` — treaty disclosure for Swedish pension
- `TestFBARDetection` — triggers FBAR guidance

---

## Phase 4: CA State Conformity for Foreign Income

**Status:** Not started
**Complexity:** HIGH
**Dependencies:** Phases 1-2

### Critical Issue: CA Does NOT Conform to FEIE

California taxes worldwide income and does NOT allow the Foreign Earned Income Exclusion. For a taxpayer excluding $130,000 federally, this results in $130,000 of additional CA taxable income — potentially $17,000+ in additional CA tax.

### New Files

#### 1. `internal/forms/state/ca/schedule_s.go`
CA Schedule S for foreign tax credits — CA's mechanism for FTC.

### Modified Files

#### 2. `internal/forms/state/ca/schedule_ca.go`
- Add FEIE add-back field (`1_col_c_feie`) — full exclusion added back
- Add housing exclusion add-back
- Update total additions line to include FEIE

#### 3. `internal/forms/state/ca/conformity.go`
Add `ConformityDifference` entries:
- Foreign Earned Income Exclusion (not allowed by CA)
- Foreign Housing Exclusion/Deduction (not allowed by CA)
- Foreign Tax Credit (CA allows via Schedule S)

#### 4. `internal/forms/state/ca/f540.go`
Ensure FEIE add-back flows into CA taxable income correctly.

#### 5. `internal/interview/ca_differences.go`
Add `CAFederalDifference` entries for FEIE and housing exclusion.

#### 6. `internal/interview/ca_schedule_ca_detect.go`
Add detection for FEIE add-back trigger.

### Tests
- `TestCAFEIEAddback` — FEIE is fully added back on Schedule CA
- `TestCAForeignHousingAddback` — housing exclusion added back
- `TestCAForeignTaxCredit` — CA allows FTC
- `TestCAExpatFullScenario` — complete Sweden expat

### Test Scenarios
- `testdata/scenarios/ca/ca_expat_feie_addback.json`
- `testdata/scenarios/ca/ca_expat_ftc.json`

---

## Phase 5: Interview, Knowledge Base, and UX

**Status:** Not started
**Complexity:** MEDIUM
**Dependencies:** Phases 1-4

### New Files

#### 1. `internal/knowledge/seed_expat.go`
~12 knowledge base documents:
- IRC Section 911 (FEIE details)
- IRC Section 901/903 (Foreign Tax Credit)
- IRC Section 6038D (FATCA/Form 8938)
- Publication 54 (Tax Guide for US Citizens Abroad)
- Publication 514 (Foreign Tax Credit)
- FBAR/FinCEN 114 requirements
- US-Sweden tax treaty overview
- Swedish pension treaty treatment
- CA FEIE non-conformity
- Currency conversion rules
- Expat filing deadlines (automatic June 15 extension)

### Modified Files

#### 2. `internal/interview/situationdetect.go`
4 new screening questions:
- `lives_abroad` — "Do you currently live outside the United States?"
- `has_foreign_income` — "Did you earn income from a foreign employer?"
- `has_foreign_accounts` — "Do you have financial accounts in a foreign country?"
- `has_foreign_pension` — "Do you have a foreign pension plan?"

#### 3. `internal/interview/questions.go`
Contextual prompts for all new form fields with `CANote`, `IRCRef`, `CARef`.

#### 4. `internal/knowledge/seed.go`
Wire `SeedExpatDocuments()` into `SeedStore()`.

### Tests
- `TestExpatScreeningQuestions`
- `TestAutoDetectFromPriorYearForm2555`
- `TestKnowledgeBaseExpatSearch`

---

## Phase 6: E-File XML, PDF Mappings, Validation, Error Handling

**Status:** Not started
**Complexity:** MEDIUM-HIGH
**Dependencies:** Phases 1-5

### Modified Files

#### 1. `internal/efile/mef/xml.go`
Add XML structs: `IRS2555`, `IRS1116`, `IRS8938`, `IRS8833`

#### 2. `internal/efile/ca/xml.go`
Add FEIE add-back fields to `CAScheduleCA` struct.

#### 3. `internal/efile/validate.go`
6 new validation rules:
- Form 2555 requires qualifying test
- PPT requires >= 330 days
- FEIE cannot exceed limit
- Form 8938 required if threshold met
- FBAR reminder if foreign accounts > $10,000
- CA Schedule CA must include FEIE add-back

#### 4. `internal/efile/reasonableness.go`
5 new checks (RC015-RC019):
- FEIE claimed but low days abroad
- FTC exceeds 50% of foreign income
- FATCA value fluctuation
- FEIE + FTC no double benefit
- CA FEIE add-back equals federal FEIE

#### 5. `internal/errors/errors.go`
- **Remove** foreign income `UnsupportedError`
- **Remove** FTC > $300 CPA referral
- **Add** CA FEIE conformity check

#### 6. `internal/pdf/mappings.go`
PDF field mappings for Forms 2555, 1116, 8938, 8833.

#### 7. `internal/pdf/export.go`
Register new form PDF mappings.

### Test Scenarios
- `testdata/scenarios/federal/expat_ppt.json` — PPT, 340 days
- `testdata/scenarios/federal/foreign_tax_credit.json` — FTC only

### Tests
- Deterministic XML for each new form
- CA XML with FEIE add-back
- Validation pass/fail
- Reasonableness checks
- Removed unsupported errors
- Full integration for all scenarios

---

## Summary

| Phase | Description | New Files | Modified Files | Tests |
|-------|-------------|-----------|----------------|-------|
| 1 | Form 2555 (FEIE) | 2 | 4 | ~15 |
| 2 | Form 1116 + Schedule B III | 1 | 3 | ~8 |
| 3 | Form 8938, 8833, FBAR | 3 | 2 | ~8 |
| 4 | CA Conformity | 1 | 5 | ~8 |
| 5 | Interview + Knowledge Base | 1 | 4 | ~6 |
| 6 | E-File, PDF, Validation | 5 scenarios | 7 | ~12 |
| **Total** | | **~8 new** | **~25 modified** | **~57 tests** |

## Key Architectural Decisions

1. **Tax stacking:** When FEIE is claimed, tax on remaining income is computed at the rate that would apply if excluded income were still included. New `ComputeTaxWithStacking()` function.

2. **No double benefit:** Form 2555 and Form 1116 cross-form dependency ensures FTC only applies to non-excluded income.

3. **CA FEIE add-back:** $130k exclusion added back at up to 13.3% = ~$17k+ additional CA tax. Must be clearly communicated in interview.

4. **Currency conversion:** UserInput field — user provides exchange rate or we suggest IRS average rate.

5. **FBAR is guidance only:** Filed separately with FinCEN. TaxPilot detects the requirement and advises but does NOT generate the filing.
