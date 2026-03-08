package forms

// This file defines typed field key constants for all forms.
// Using these constants instead of raw strings catches typos at compile time.
// Field keys are in "form_id:line" format.

// --- Form 1040 Field Keys ---

const (
	// Identification
	F1040FilingStatus = "1040:filing_status"
	F1040FirstName    = "1040:first_name"
	F1040LastName     = "1040:last_name"
	F1040SSN          = "1040:ssn"

	// Income
	F1040Line1a = "1040:1a"
	F1040Line1z = "1040:1z"
	F1040Line2a = "1040:2a"
	F1040Line2b = "1040:2b"
	F1040Line3a = "1040:3a"
	F1040Line3b = "1040:3b"
	F1040Line7  = "1040:7"
	F1040Line8  = "1040:8"
	F1040Line9  = "1040:9"

	// AGI
	F1040Line10 = "1040:10"
	F1040Line11 = "1040:11"

	// Deductions
	F1040Line12 = "1040:12"
	F1040Line13 = "1040:13"
	F1040Line14 = "1040:14"
	F1040Line15 = "1040:15"

	// Tax
	F1040Line16 = "1040:16"
	F1040Line17 = "1040:17"
	F1040Line20 = "1040:20"
	F1040Line22 = "1040:22"
	F1040Line23 = "1040:23"
	F1040Line24 = "1040:24"

	// Payments
	F1040Line25a = "1040:25a"
	F1040Line25b = "1040:25b"
	F1040Line25d = "1040:25d"
	F1040Line31  = "1040:31"
	F1040Line33  = "1040:33"

	// Refund / Owed
	F1040Line34 = "1040:34"
	F1040Line37 = "1040:37"
)

// --- W-2 Field Keys ---
// W-2 uses instance keys (e.g., "w2:1:wages"), but these are the line names.

const (
	W2EmployerName     = "employer_name"
	W2EmployerEIN      = "employer_ein"
	W2Wages            = "wages"
	W2FedTaxWithheld   = "federal_tax_withheld"
	W2SSWages          = "ss_wages"
	W2SSTaxWithheld    = "ss_tax_withheld"
	W2MedicareWages    = "medicare_wages"
	W2MedicareTaxWH    = "medicare_tax_withheld"
	W2StateWages       = "state_wages"
	W2StateTaxWithheld = "state_tax_withheld"
)

// W-2 wildcard patterns for dependency resolution
const (
	W2WildcardWages         = "w2:*:wages"
	W2WildcardFedTaxWH      = "w2:*:federal_tax_withheld"
	W2WildcardSSWages       = "w2:*:ss_wages"
	W2WildcardSSTaxWH       = "w2:*:ss_tax_withheld"
	W2WildcardMedicareWages = "w2:*:medicare_wages"
	W2WildcardMedicareTaxWH = "w2:*:medicare_tax_withheld"
	W2WildcardStateWages    = "w2:*:state_wages"
	W2WildcardStateTaxWH    = "w2:*:state_tax_withheld"
)

// --- Schedule A Field Keys ---

const (
	SchedALine1  = "schedule_a:1"
	SchedALine2  = "schedule_a:2"
	SchedALine3  = "schedule_a:3"
	SchedALine4  = "schedule_a:4"
	SchedALine5a = "schedule_a:5a"
	SchedALine5b = "schedule_a:5b"
	SchedALine5c = "schedule_a:5c"
	SchedALine5d = "schedule_a:5d"
	SchedALine5e = "schedule_a:5e"
	SchedALine8a = "schedule_a:8a"
	SchedALine11 = "schedule_a:11"
	SchedALine12 = "schedule_a:12"
	SchedALine13 = "schedule_a:13"
	SchedALine14 = "schedule_a:14"
	SchedALine15 = "schedule_a:15"
	SchedALine17 = "schedule_a:17"
)

// --- Schedule B Field Keys ---

const (
	SchedBLine1  = "schedule_b:1"
	SchedBLine4  = "schedule_b:4"
	SchedBLine5  = "schedule_b:5"
	SchedBLine6  = "schedule_b:6"
	SchedBLine7a = "schedule_b:7a"
	SchedBLine7b = "schedule_b:7b"
	SchedBLine8  = "schedule_b:8"
)

// --- Schedule C Field Keys ---

const (
	SchedCBusinessName = "schedule_c:business_name"
	SchedCBusinessCode = "schedule_c:business_code"
	SchedCLine1        = "schedule_c:1"
	SchedCLine5        = "schedule_c:5"
	SchedCLine7        = "schedule_c:7"
	SchedCLine8        = "schedule_c:8"
	SchedCLine10       = "schedule_c:10"
	SchedCLine17       = "schedule_c:17"
	SchedCLine18       = "schedule_c:18"
	SchedCLine22       = "schedule_c:22"
	SchedCLine25       = "schedule_c:25"
	SchedCLine27       = "schedule_c:27"
	SchedCLine28       = "schedule_c:28"
	SchedCLine31       = "schedule_c:31"
)

// --- Schedule D Field Keys ---

const (
	SchedDLine1  = "schedule_d:1"
	SchedDLine7  = "schedule_d:7"
	SchedDLine8  = "schedule_d:8"
	SchedDLine13 = "schedule_d:13"
	SchedDLine15 = "schedule_d:15"
	SchedDLine16 = "schedule_d:16"
)

// --- Schedule 1 Field Keys ---

const (
	Sched1Line1   = "schedule_1:1"
	Sched1Line3   = "schedule_1:3"
	Sched1Line7   = "schedule_1:7"
	Sched1Line8d  = "schedule_1:8d"
	Sched1Line10  = "schedule_1:10"
	Sched1Line15  = "schedule_1:15"
	Sched1Line16  = "schedule_1:16"
	Sched1Line24  = "schedule_1:24"
	Sched1Line26  = "schedule_1:26"
)

// --- Schedule 2 Field Keys ---

const (
	Sched2Line1   = "schedule_2:1"
	Sched2Line3   = "schedule_2:3"
	Sched2Line6   = "schedule_2:6"
	Sched2Line12  = "schedule_2:12"
	Sched2Line17c = "schedule_2:17c"
	Sched2Line18  = "schedule_2:18"
	Sched2Line21  = "schedule_2:21"
)

// --- Schedule 3 Field Keys ---

const (
	Sched3Line1  = "schedule_3:1"
	Sched3Line8  = "schedule_3:8"
	Sched3Line10 = "schedule_3:10"
	Sched3Line15 = "schedule_3:15"
)

// --- Schedule SE Field Keys ---

const (
	SchedSELine2 = "schedule_se:2"
	SchedSELine3 = "schedule_se:3"
	SchedSELine4 = "schedule_se:4"
	SchedSELine5 = "schedule_se:5"
	SchedSELine6 = "schedule_se:6"
	SchedSELine7 = "schedule_se:7"
)

// --- Form 8889 Field Keys ---

const (
	F8889Line1   = "form_8889:1"
	F8889Line2   = "form_8889:2"
	F8889Line3   = "form_8889:3"
	F8889Line5   = "form_8889:5"
	F8889Line6   = "form_8889:6"
	F8889Line9   = "form_8889:9"
	F8889Line14a = "form_8889:14a"
	F8889Line14c = "form_8889:14c"
	F8889Line15  = "form_8889:15"
	F8889Line17b = "form_8889:17b"
)

// --- Form 8949 Field Keys ---

const (
	F8949STProceedsKey  = "form_8949:st_proceeds"
	F8949STBasisKey     = "form_8949:st_basis"
	F8949STWashKey      = "form_8949:st_wash"
	F8949STGainLossKey  = "form_8949:st_gain_loss"
	F8949LTProceedsKey  = "form_8949:lt_proceeds"
	F8949LTBasisKey     = "form_8949:lt_basis"
	F8949LTWashKey      = "form_8949:lt_wash"
	F8949LTGainLossKey  = "form_8949:lt_gain_loss"
)

// --- Form 8995 Field Keys ---

const (
	F8995Line3  = "form_8995:3"
	F8995Line4  = "form_8995:4"
	F8995Line5  = "form_8995:5"
	F8995Line8  = "form_8995:8"
	F8995Line10 = "form_8995:10"
)

// --- Form 2555 (FEIE) Field Keys ---

const (
	F2555ForeignCountry       = "form_2555:foreign_country"
	F2555ForeignAddress       = "form_2555:foreign_address"
	F2555EmployerName         = "form_2555:employer_name_2555"
	F2555EmployerForeign      = "form_2555:employer_foreign"
	F2555SelfEmployedAbroad   = "form_2555:self_employed_abroad"
	F2555QualifyingTest       = "form_2555:qualifying_test"
	F2555BFRTStartDate        = "form_2555:bfrt_start_date"
	F2555BFRTEndDate          = "form_2555:bfrt_end_date"
	F2555BFRTFullYear         = "form_2555:bfrt_full_year"
	F2555PPTDaysPresent       = "form_2555:ppt_days_present"
	F2555PPTPeriodStart       = "form_2555:ppt_period_start"
	F2555PPTPeriodEnd         = "form_2555:ppt_period_end"
	F2555ForeignEarnedIncome  = "form_2555:foreign_earned_income"
	F2555CurrencyCode         = "form_2555:currency_code"
	F2555ExchangeRate         = "form_2555:exchange_rate"
	F2555ForeignTaxPaid       = "form_2555:foreign_tax_paid"
	F2555EmployerHousing      = "form_2555:employer_provided_housing"
	F2555HousingExpenses      = "form_2555:housing_expenses"
	F2555QualifyingDays       = "form_2555:qualifying_days"
	F2555ExclusionLimit       = "form_2555:exclusion_limit"
	F2555ForeignIncomeExcl    = "form_2555:foreign_income_exclusion"
	F2555HousingExclusion     = "form_2555:housing_exclusion"
	F2555HousingDeduction     = "form_2555:housing_deduction"
	F2555TotalExclusion       = "form_2555:total_exclusion"
)

// --- Form 1116 (FTC) Field Keys ---

const (
	F1116Category             = "form_1116:category"
	F1116ForeignCountry       = "form_1116:foreign_country"
	F1116ForeignSourceIncome  = "form_1116:foreign_source_income"
	F1116ForeignSourceDeduct  = "form_1116:foreign_source_deductions"
	F1116ForeignTaxPaidIncome = "form_1116:foreign_tax_paid_income"
	F1116ForeignTaxPaidOther  = "form_1116:foreign_tax_paid_other"
	F1116AccruedOrPaid        = "form_1116:accrued_or_paid"
	F1116Line7                = "form_1116:7"
	F1116Line15               = "form_1116:15"
	F1116Line21               = "form_1116:21"
	F1116Line22               = "form_1116:22"
	F1116Carryforward         = "form_1116:carryforward"
)

// --- Form 8938 (FATCA) Field Keys ---

const (
	F8938LivesAbroad        = "form_8938:lives_abroad"
	F8938NumAccounts        = "form_8938:num_accounts"
	F8938MaxValueAccounts   = "form_8938:max_value_accounts"
	F8938YearEndAccounts    = "form_8938:yearend_value_accounts"
	F8938NumOtherAssets      = "form_8938:num_other_assets"
	F8938MaxValueOther       = "form_8938:max_value_other"
	F8938YearEndOther        = "form_8938:yearend_value_other"
	F8938AccountCountry      = "form_8938:account_country"
	F8938AccountInstitution  = "form_8938:account_institution"
	F8938AccountType         = "form_8938:account_type"
	F8938IncomeFromAccounts  = "form_8938:income_from_accounts"
	F8938GainFromAccounts    = "form_8938:gain_from_accounts"
	F8938TotalMaxValue       = "form_8938:total_max_value"
	F8938TotalYearEndValue   = "form_8938:total_yearend_value"
	F8938FilingRequired      = "form_8938:filing_required"
)

// --- Form 8833 (Treaty) Field Keys ---

const (
	F8833TreatyCountry  = "form_8833:treaty_country"
	F8833TreatyArticle  = "form_8833:treaty_article"
	F8833IRCProvision   = "form_8833:irc_provision"
	F8833TreatyExplain  = "form_8833:treaty_position_explanation"
	F8833TreatyAmount   = "form_8833:treaty_amount"
	F8833NumPositions   = "form_8833:num_positions"
	F8833TreatyClaimed  = "form_8833:treaty_claimed"
)

// --- CA Form 540 Field Keys ---

const (
	CA540Line7  = "ca_540:7"
	CA540Line13 = "ca_540:13"
	CA540Line14 = "ca_540:14"
	CA540Line15 = "ca_540:15"
	CA540Line17 = "ca_540:17"
	CA540Line18 = "ca_540:18"
	CA540Line19 = "ca_540:19"
	CA540Line31 = "ca_540:31"
	CA540Line32 = "ca_540:32"
	CA540Line35 = "ca_540:35"
	CA540Line36 = "ca_540:36"
	CA540Line40 = "ca_540:40"
	CA540Line71 = "ca_540:71"
	CA540Line74 = "ca_540:74"
	CA540Line75 = "ca_540:75"
	CA540Line81 = "ca_540:81"
	CA540Line91 = "ca_540:91"
	CA540Line93 = "ca_540:93"
)

// --- CA Schedule CA Field Keys ---

const (
	SchedCALine8dColC        = "ca_schedule_ca:8d_col_c"
	SchedCALine8dColCHousing = "ca_schedule_ca:8d_col_c_housing"
	SchedCALine37ColC        = "ca_schedule_ca:37_col_c"
)

// --- 1099 Wildcard Patterns ---

const (
	F1099INTWildcardInterest  = "1099int:*:interest_income"
	F1099INTWildcardPenalty   = "1099int:*:early_withdrawal_penalty"
	F1099INTWildcardFedTaxWH  = "1099int:*:federal_tax_withheld"
	F1099INTWildcardTaxExempt = "1099int:*:tax_exempt_interest"
	F1099INTWildcardUSBond    = "1099int:*:us_savings_bond_interest"
	F1099INTWildcardPABI      = "1099int:*:private_activity_bond_interest"

	F1099DIVWildcardOrdinary   = "1099div:*:ordinary_dividends"
	F1099DIVWildcardQualified  = "1099div:*:qualified_dividends"
	F1099DIVWildcardCapGain    = "1099div:*:total_capital_gain"
	F1099DIVWildcardFedTaxWH   = "1099div:*:federal_tax_withheld"
	F1099DIVWildcardExemptInt  = "1099div:*:exempt_interest_dividends"
	F1099DIVWildcardPABI       = "1099div:*:private_activity_bond_dividends"
	F1099DIVWildcardSec199a    = "1099div:*:section_199a_dividends"
	F1099DIVWildcardSec1250    = "1099div:*:section_1250_gain"

	F1099NECWildcardComp      = "1099nec:*:nonemployee_compensation"
	F1099NECWildcardFedTaxWH  = "1099nec:*:federal_tax_withheld"

	F1099BWildcardProceeds    = "1099b:*:proceeds"
	F1099BWildcardBasis       = "1099b:*:cost_basis"
	F1099BWildcardWashSale    = "1099b:*:wash_sale_loss"
	F1099BWildcardFedTaxWH    = "1099b:*:federal_tax_withheld"
	F1099BWildcardTerm        = "1099b:*:term"
)

// --- Form 3514 (CalEITC) Field Keys ---

const (
	F3514Line3    = "form_3514:3"
	F3514Line6YCTC = "form_3514:6_yctc"
)

// --- Form 3853 (Health Coverage) Field Keys ---

const (
	F3853Line1 = "form_3853:1"
	F3853Line2 = "form_3853:2"
	F3853Line3 = "form_3853:3"
)
