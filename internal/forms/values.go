package forms

// This file defines typed constants for common string values used across
// form definitions, interview logic, and rendering. Using these instead of
// raw strings catches typos at compile time and makes refactoring safer.

// --- Filing Status Values ---
// These correspond to the Options on Form 1040's filing_status field.

const (
	FilingSingle = "single"
	FilingMFJ    = "mfj"  // Married Filing Jointly
	FilingMFS    = "mfs"  // Married Filing Separately
	FilingHOH    = "hoh"  // Head of Household
	FilingQSS    = "qss"  // Qualifying Surviving Spouse
)

// FilingStatusOptions is the canonical option list for the filing status field.
var FilingStatusOptions = []string{FilingSingle, FilingMFJ, FilingMFS, FilingHOH, FilingQSS}

// --- Yes/No Option Values ---

const (
	OptionYes = "yes"
	OptionNo  = "no"
)

// YesNoOptions is the canonical option list for yes/no fields.
var YesNoOptions = []string{OptionYes, OptionNo}

// --- State Codes ---

const (
	StateCodeCA = "CA"
)

// --- Question Group Names ---
// These are used in FormDef.QuestionGroup and interview engine routing.

const (
	GroupPersonal   = "personal"
	GroupIncomeW2   = "income_w2"
	GroupIncome1099 = "income_1099"
	GroupExpat      = "expat"
	GroupCA         = "ca"
)

// --- Common Field Line Names ---
// These are used in form definitions and interview engine routing logic.

const (
	LineFilingStatus          = "filing_status"
	LineFirstName             = "first_name"
	LineLastName              = "last_name"
	LineSSN                   = "ssn"
	LineEmployerName          = "employer_name"
	LineEmployerEIN           = "employer_ein"
	LinePayerName             = "payer_name"
	LinePayerTIN              = "payer_tin"
	LineForeignWages          = "foreign_wages"
	LineForeignEmployer       = "foreign_employer"
	LineForeignInterest       = "foreign_interest"
	LineForeignInterestPayer  = "foreign_interest_payer"
	LineDescription           = "description"
	LineDateAcquired          = "date_acquired"
	LineDateSold              = "date_sold"
)

// --- Qualifying Test Values (Form 2555) ---

const (
	QualifyingTestBFRT = "bona_fide_residence"
	QualifyingTestPPT  = "physical_presence"
)

// QualifyingTestOptions is the option list for FEIE qualifying test.
var QualifyingTestOptions = []string{QualifyingTestBFRT, QualifyingTestPPT}

// --- FTC Category Values (Form 1116) ---

const (
	FTCCategoryGeneral = "general"
	FTCCategoryPassive = "passive"
)

// --- Accrued/Paid Values (Form 1116) ---

const (
	AccruedOrPaidAccrued = "accrued"
	AccruedOrPaidPaid    = "paid"
)
