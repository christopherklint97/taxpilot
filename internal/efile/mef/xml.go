package mef

import (
	"encoding/xml"
	"fmt"
	"math"
	"strings"
)

// MeF XML namespace and version constants.
const (
	MeFNamespace    = "http://www.irs.gov/efile"
	ReturnVersion   = "2025v1.0"
)

// Filing status codes for MeF.
var filingStatusCodes = map[string]int{
	"single": 1,
	"mfj":    2,
	"mfs":    3,
	"hoh":    4,
	"qss":    5,
}

// --- XML structure types ---

// Return is the root element of a MeF XML return.
type Return struct {
	XMLName       xml.Name     `xml:"Return"`
	Xmlns         string       `xml:"xmlns,attr"`
	ReturnVersion string       `xml:"returnVersion,attr"`
	ReturnHeader  ReturnHeader `xml:"ReturnHeader"`
	ReturnData    ReturnData   `xml:"ReturnData"`
}

// ReturnHeader contains taxpayer identification and filing metadata.
type ReturnHeader struct {
	BinaryAttachmentCnt int    `xml:"binaryAttachmentCnt,attr"`
	TaxYr               int    `xml:"TaxYr"`
	TaxPeriodBeginDt    string `xml:"TaxPeriodBeginDt"`
	TaxPeriodEndDt      string `xml:"TaxPeriodEndDt"`
	Filer               Filer  `xml:"Filer"`
}

// Filer contains taxpayer identity information.
type Filer struct {
	PrimarySSN    string   `xml:"PrimarySSN"`
	Name          FilerName `xml:"Name"`
	FilingStatusCd int     `xml:"FilingStatusCd"`
}

// FilerName contains the taxpayer's name.
type FilerName struct {
	FirstName string `xml:"FirstName"`
	LastName  string `xml:"LastName"`
}

// ReturnData contains all form data documents.
type ReturnData struct {
	DocumentCnt         int                  `xml:"documentCnt,attr"`
	IRS1040             *IRS1040             `xml:"IRS1040,omitempty"`
	IRS1040ScheduleA    *IRS1040ScheduleA    `xml:"IRS1040ScheduleA,omitempty"`
	IRS1040Schedule1    *IRS1040Schedule1    `xml:"IRS1040Schedule1,omitempty"`
	IRS1040Schedule2    *IRS1040Schedule2    `xml:"IRS1040Schedule2,omitempty"`
	IRS1040Schedule3    *IRS1040Schedule3    `xml:"IRS1040Schedule3,omitempty"`
	IRS1040ScheduleB    *IRS1040ScheduleB    `xml:"IRS1040ScheduleB,omitempty"`
	IRS1040ScheduleC    *IRS1040ScheduleC    `xml:"IRS1040ScheduleC,omitempty"`
	IRS1040ScheduleD    *IRS1040ScheduleD    `xml:"IRS1040ScheduleD,omitempty"`
	IRS1040ScheduleSE   *IRS1040ScheduleSE   `xml:"IRS1040ScheduleSE,omitempty"`
	IRS8889             *IRS8889             `xml:"IRS8889,omitempty"`
	IRS8949             *IRS8949             `xml:"IRS8949,omitempty"`
	IRS8995             *IRS8995             `xml:"IRS8995,omitempty"`
	IRS2555             *IRS2555             `xml:"IRS2555,omitempty"`
	IRS1116             *IRS1116             `xml:"IRS1116,omitempty"`
	IRS8938             *IRS8938             `xml:"IRS8938,omitempty"`
	IRS8833             *IRS8833             `xml:"IRS8833,omitempty"`
	IRSW2               []IRSW2              `xml:"IRSW2,omitempty"`
}

// IRS1040 represents the Form 1040 XML element.
type IRS1040 struct {
	WagesSalariesTips       int `xml:"WagesSalariesTips"`
	TaxExemptInterestAmt    int `xml:"TaxExemptInterestAmt,omitempty"`
	TaxableInterestAmt      int `xml:"TaxableInterestAmt,omitempty"`
	QualifiedDividendsAmt   int `xml:"QualifiedDividendsAmt,omitempty"`
	OrdinaryDividendsAmt    int `xml:"OrdinaryDividendsAmt,omitempty"`
	CapitalGainLossAmt      int `xml:"CapitalGainLossAmt,omitempty"`
	OtherIncomeAmt          int `xml:"OtherIncomeAmt,omitempty"`
	TotalIncomeAmt          int `xml:"TotalIncomeAmt"`
	AdjustmentsToIncomeAmt  int `xml:"AdjustmentsToIncomeAmt,omitempty"`
	AdjustedGrossIncomeAmt  int `xml:"AdjustedGrossIncomeAmt"`
	TotalDeductionsAmt      int `xml:"TotalDeductionsAmt"`
	TaxableIncomeAmt        int `xml:"TaxableIncomeAmt"`
	TaxAmt                  int `xml:"TaxAmt"`
	Sch2PartIAmt            int `xml:"Sch2PartIAmt,omitempty"`
	Sch3PartIAmt            int `xml:"Sch3PartIAmt,omitempty"`
	TaxAfterCreditsAmt      int `xml:"TaxAfterCreditsAmt"`
	OtherTaxesAmt           int `xml:"OtherTaxesAmt,omitempty"`
	TotalTaxAmt             int `xml:"TotalTaxAmt"`
	WithholdingTaxAmt       int `xml:"WithholdingTaxAmt"`
	EstimatedTaxPaymentsAmt int `xml:"EstimatedTaxPaymentsAmt,omitempty"`
	TotalPaymentsAmt        int `xml:"TotalPaymentsAmt"`
	OverpaidAmt             int `xml:"OverpaidAmt,omitempty"`
	OwedAmt                 int `xml:"OwedAmt,omitempty"`
}

// IRS1040ScheduleA represents Schedule A — Itemized Deductions.
type IRS1040ScheduleA struct {
	MedicalAndDentalExpAmt    int `xml:"MedicalAndDentalExpAmt,omitempty"`
	AGIAmt                    int `xml:"AGIAmt,omitempty"`
	MedicalFloorAmt           int `xml:"MedicalFloorAmt,omitempty"`
	DeductibleMedicalAmt      int `xml:"DeductibleMedicalAmt,omitempty"`
	StateLocalIncomeTaxAmt    int `xml:"StateLocalIncomeTaxAmt,omitempty"`
	PropertyTaxAmt            int `xml:"PropertyTaxAmt,omitempty"`
	RealEstateTaxAmt          int `xml:"RealEstateTaxAmt,omitempty"`
	TotalSALTAmt              int `xml:"TotalSALTAmt,omitempty"`
	SALTDeductionAmt          int `xml:"SALTDeductionAmt,omitempty"`
	MortgageInterestAmt       int `xml:"MortgageInterestAmt,omitempty"`
	TotalInterestDeductionAmt int `xml:"TotalInterestDeductionAmt,omitempty"`
	CashCharityAmt            int `xml:"CashCharityAmt,omitempty"`
	NonCashCharityAmt         int `xml:"NonCashCharityAmt,omitempty"`
	CharityCarryoverAmt       int `xml:"CharityCarryoverAmt,omitempty"`
	TotalCharityAmt           int `xml:"TotalCharityAmt,omitempty"`
	TotalItemizedDeductAmt    int `xml:"TotalItemizedDeductAmt"`
}

// IRS1040Schedule1 represents Schedule 1 — Additional Income and Adjustments.
type IRS1040Schedule1 struct {
	BusinessIncomeLossAmt     int `xml:"BusinessIncomeLossAmt,omitempty"`
	CapitalGainLossAmt        int `xml:"CapitalGainLossAmt,omitempty"`
	TotalAdditionalIncomeAmt  int `xml:"TotalAdditionalIncomeAmt"`
	HSADeductionAmt           int `xml:"HSADeductionAmt,omitempty"`
	SETaxDeductionAmt         int `xml:"SETaxDeductionAmt,omitempty"`
	EarlyWithdrawalPenaltyAmt int `xml:"EarlyWithdrawalPenaltyAmt,omitempty"`
	TotalAdjustmentsAmt       int `xml:"TotalAdjustmentsAmt"`
}

// IRS1040Schedule2 represents Schedule 2 — Additional Taxes.
type IRS1040Schedule2 struct {
	AMTAmt                   int `xml:"AMTAmt,omitempty"`
	TotalPartIAmt            int `xml:"TotalPartIAmt,omitempty"`
	SelfEmploymentTaxAmt     int `xml:"SelfEmploymentTaxAmt,omitempty"`
	AdditionalMedicareTaxAmt int `xml:"AdditionalMedicareTaxAmt,omitempty"`
	HSAPenaltyAmt            int `xml:"HSAPenaltyAmt,omitempty"`
	NIITAmt                  int `xml:"NIITAmt,omitempty"`
	TotalOtherTaxesAmt       int `xml:"TotalOtherTaxesAmt"`
}

// IRS1040Schedule3 represents Schedule 3 — Additional Credits and Payments.
type IRS1040Schedule3 struct {
	TotalNonrefundableCreditsAmt int `xml:"TotalNonrefundableCreditsAmt,omitempty"`
	EstimatedTaxPaymentsAmt      int `xml:"EstimatedTaxPaymentsAmt,omitempty"`
	TotalOtherPaymentsAmt        int `xml:"TotalOtherPaymentsAmt,omitempty"`
}

// IRS1040ScheduleB represents Schedule B — Interest and Ordinary Dividends.
type IRS1040ScheduleB struct {
	TotalInterestAmt   int `xml:"TotalInterestAmt"`
	TotalDividendsAmt  int `xml:"TotalDividendsAmt"`
}

// IRS1040ScheduleC represents Schedule C — Profit or Loss From Business.
type IRS1040ScheduleC struct {
	BusinessName       string `xml:"BusinessName,omitempty"`
	BusinessCode       string `xml:"BusinessCode,omitempty"`
	GrossReceiptsAmt   int    `xml:"GrossReceiptsAmt"`
	GrossProfitAmt     int    `xml:"GrossProfitAmt"`
	TotalExpensesAmt   int    `xml:"TotalExpensesAmt,omitempty"`
	NetProfitLossAmt   int    `xml:"NetProfitLossAmt"`
}

// IRS1040ScheduleD represents Schedule D — Capital Gains and Losses.
type IRS1040ScheduleD struct {
	STGainLossAmt           int `xml:"STGainLossAmt,omitempty"`
	NetSTGainLossAmt        int `xml:"NetSTGainLossAmt,omitempty"`
	LTGainLossAmt           int `xml:"LTGainLossAmt,omitempty"`
	CapGainDistributionsAmt int `xml:"CapGainDistributionsAmt,omitempty"`
	NetLTGainLossAmt        int `xml:"NetLTGainLossAmt,omitempty"`
	NetCapitalGainLossAmt   int `xml:"NetCapitalGainLossAmt"`
}

// IRS1040ScheduleSE represents Schedule SE — Self-Employment Tax.
type IRS1040ScheduleSE struct {
	NetSEEarningsAmt    int `xml:"NetSEEarningsAmt"`
	TaxableEarningsAmt  int `xml:"TaxableEarningsAmt"`
	SSTaxAmt            int `xml:"SSTaxAmt"`
	MedicareTaxAmt      int `xml:"MedicareTaxAmt"`
	SelfEmploymentTaxAmt int `xml:"SelfEmploymentTaxAmt"`
	DeductibleSETaxAmt  int `xml:"DeductibleSETaxAmt"`
}

// IRS8889 represents Form 8889 — Health Savings Accounts.
type IRS8889 struct {
	CoverageType          string `xml:"CoverageType,omitempty"`
	ContributionsAmt      int    `xml:"ContributionsAmt,omitempty"`
	EmployerContribAmt    int    `xml:"EmployerContribAmt,omitempty"`
	ContributionLimitAmt  int    `xml:"ContributionLimitAmt"`
	HSADeductionAmt       int    `xml:"HSADeductionAmt"`
	DistributionsAmt      int    `xml:"DistributionsAmt,omitempty"`
	QualifiedExpensesAmt  int    `xml:"QualifiedExpensesAmt,omitempty"`
	TaxableDistribAmt     int    `xml:"TaxableDistribAmt,omitempty"`
	PenaltyAmt            int    `xml:"PenaltyAmt,omitempty"`
}

// IRS8949 represents Form 8949 — Sales and Other Dispositions of Capital Assets.
type IRS8949 struct {
	STProceedsAmt   int `xml:"STProceedsAmt,omitempty"`
	STBasisAmt      int `xml:"STBasisAmt,omitempty"`
	STWashSaleAmt   int `xml:"STWashSaleAmt,omitempty"`
	STGainLossAmt   int `xml:"STGainLossAmt,omitempty"`
	LTProceedsAmt   int `xml:"LTProceedsAmt,omitempty"`
	LTBasisAmt      int `xml:"LTBasisAmt,omitempty"`
	LTWashSaleAmt   int `xml:"LTWashSaleAmt,omitempty"`
	LTGainLossAmt   int `xml:"LTGainLossAmt,omitempty"`
}

// IRS8995 represents Form 8995 — Qualified Business Income Deduction (Simplified).
type IRS8995 struct {
	TotalQBIAmt            int `xml:"TotalQBIAmt"`
	QBIComponentAmt        int `xml:"QBIComponentAmt"`
	TaxableIncBeforeQBIAmt int `xml:"TaxableIncBeforeQBIAmt"`
	IncomeLimitationAmt    int `xml:"IncomeLimitationAmt"`
	QBIDeductionAmt        int `xml:"QBIDeductionAmt"`
}

// IRS2555 represents Form 2555 — Foreign Earned Income.
type IRS2555 struct {
	ForeignCountry          string `xml:"ForeignCountry"`
	QualifyingTest          string `xml:"QualifyingTest"`
	QualifyingDays          int    `xml:"QualifyingDays"`
	ForeignEarnedIncomeAmt  int    `xml:"ForeignEarnedIncomeAmt"`
	ExclusionLimitAmt       int    `xml:"ExclusionLimitAmt"`
	ForeignIncomeExclAmt    int    `xml:"ForeignIncomeExclAmt"`
	HousingExclusionAmt     int    `xml:"HousingExclusionAmt,omitempty"`
	HousingDeductionAmt     int    `xml:"HousingDeductionAmt,omitempty"`
	TotalExclusionAmt       int    `xml:"TotalExclusionAmt"`
}

// IRS1116 represents Form 1116 — Foreign Tax Credit.
type IRS1116 struct {
	ForeignCountry         string `xml:"ForeignCountry"`
	Category               string `xml:"Category"`
	ForeignSourceIncomeAmt int    `xml:"ForeignSourceIncomeAmt"`
	ForeignTaxPaidAmt      int    `xml:"ForeignTaxPaidAmt"`
	CreditLimitationAmt    int    `xml:"CreditLimitationAmt"`
	AllowedCreditAmt       int    `xml:"AllowedCreditAmt"`
	CarryforwardAmt        int    `xml:"CarryforwardAmt,omitempty"`
}

// IRS8938 represents Form 8938 — Statement of Specified Foreign Financial Assets.
type IRS8938 struct {
	LivesAbroad             string `xml:"LivesAbroad"`
	MaxValueAccountsAmt     int    `xml:"MaxValueAccountsAmt"`
	YearEndValueAccountsAmt int    `xml:"YearEndValueAccountsAmt"`
	TotalMaxValueAmt        int    `xml:"TotalMaxValueAmt"`
	TotalYearEndValueAmt    int    `xml:"TotalYearEndValueAmt"`
	FilingRequired          int    `xml:"FilingRequired"`
}

// IRS8833 represents Form 8833 — Treaty-Based Return Position Disclosure.
type IRS8833 struct {
	TreatyCountry     string `xml:"TreatyCountry"`
	TreatyArticle     string `xml:"TreatyArticle"`
	IRCProvision      string `xml:"IRCProvision"`
	TreatyAmountAmt   int    `xml:"TreatyAmountAmt,omitempty"`
	TreatyClaimed     int    `xml:"TreatyClaimed"`
}

// IRSW2 represents a W-2 Wage and Tax Statement.
type IRSW2 struct {
	EmployerName   string `xml:"EmployerName"`
	EmployerEIN    string `xml:"EmployerEIN"`
	WagesAmt       int    `xml:"WagesAmt"`
	WithholdingAmt int    `xml:"WithholdingAmt"`
	SSWagesAmt     int    `xml:"SSWagesAmt,omitempty"`
	SSTaxAmt       int    `xml:"SSTaxAmt,omitempty"`
	MedicareWages  int    `xml:"MedicareWagesAmt,omitempty"`
	MedicareTaxAmt int    `xml:"MedicareTaxAmt,omitempty"`
}

// --- Public API ---

// GenerateReturn takes the solver output maps and produces deterministic
// MeF-compatible XML bytes. Only schedules with non-zero values are included.
func GenerateReturn(results map[string]float64, strInputs map[string]string, taxYear int) ([]byte, error) {
	ret := Return{
		Xmlns:         MeFNamespace,
		ReturnVersion: ReturnVersion,
		ReturnHeader:  buildReturnHeader(results, strInputs, taxYear),
	}

	data := ReturnData{}
	docCount := 0

	// --- IRS 1040 (always included) ---
	data.IRS1040 = buildIRS1040(results)
	docCount++

	// --- Schedule A ---
	if isScheduleNeeded(results, "schedule_a:") {
		data.IRS1040ScheduleA = buildScheduleA(results)
		docCount++
	}

	// --- Schedule 1 ---
	if isScheduleNeeded(results, "schedule_1:") {
		data.IRS1040Schedule1 = buildSchedule1(results)
		docCount++
	}

	// --- Schedule 2 ---
	if isScheduleNeeded(results, "schedule_2:") {
		data.IRS1040Schedule2 = buildSchedule2(results)
		docCount++
	}

	// --- Schedule 3 ---
	if isScheduleNeeded(results, "schedule_3:") {
		data.IRS1040Schedule3 = buildSchedule3(results)
		docCount++
	}

	// --- Schedule B ---
	if isScheduleNeeded(results, "schedule_b:") {
		data.IRS1040ScheduleB = buildScheduleB(results)
		docCount++
	}

	// --- Schedule C ---
	if isScheduleNeeded(results, "schedule_c:") {
		data.IRS1040ScheduleC = buildScheduleC(results, strInputs)
		docCount++
	}

	// --- Schedule D ---
	if isScheduleNeeded(results, "schedule_d:") {
		data.IRS1040ScheduleD = buildScheduleD(results)
		docCount++
	}

	// --- Schedule SE ---
	if isScheduleNeeded(results, "schedule_se:") {
		data.IRS1040ScheduleSE = buildScheduleSE(results)
		docCount++
	}

	// --- Form 8889 ---
	if isScheduleNeeded(results, "form_8889:") {
		data.IRS8889 = buildForm8889(results, strInputs)
		docCount++
	}

	// --- Form 8949 ---
	if isScheduleNeeded(results, "form_8949:") {
		data.IRS8949 = buildForm8949(results)
		docCount++
	}

	// --- Form 8995 ---
	if isScheduleNeeded(results, "form_8995:") {
		data.IRS8995 = buildForm8995(results)
		docCount++
	}

	// --- Form 2555 ---
	if isScheduleNeeded(results, "form_2555:") {
		data.IRS2555 = buildForm2555(results, strInputs)
		docCount++
	}

	// --- Form 1116 ---
	if isScheduleNeeded(results, "form_1116:") {
		data.IRS1116 = buildForm1116(results, strInputs)
		docCount++
	}

	// --- Form 8938 ---
	if isScheduleNeeded(results, "form_8938:") {
		data.IRS8938 = buildForm8938(results, strInputs)
		docCount++
	}

	// --- Form 8833 ---
	if isScheduleNeeded(results, "form_8833:") {
		data.IRS8833 = buildForm8833(results, strInputs)
		docCount++
	}

	// --- W-2s ---
	w2s := buildW2s(results, strInputs)
	if len(w2s) > 0 {
		data.IRSW2 = w2s
		docCount += len(w2s)
	}

	data.DocumentCnt = docCount
	ret.ReturnData = data

	output, err := xml.MarshalIndent(ret, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("mef: marshal XML: %w", err)
	}

	// Prepend XML declaration
	header := []byte(xml.Header)
	return append(header, output...), nil
}

// --- Helper functions ---

// roundToInt rounds a float64 to the nearest integer (IRS requirement).
func roundToInt(f float64) int {
	return int(math.Round(f))
}

// formatSSN strips dashes from an SSN string for XML.
func formatSSN(s string) string {
	return strings.ReplaceAll(s, "-", "")
}

// isScheduleNeeded checks if any field with the given prefix has a non-zero value.
func isScheduleNeeded(results map[string]float64, prefix string) bool {
	for k, v := range results {
		if strings.HasPrefix(k, prefix) && v != 0 {
			return true
		}
	}
	return false
}

// --- Builder functions ---

func buildReturnHeader(results map[string]float64, strInputs map[string]string, taxYear int) ReturnHeader {
	fsStr := strInputs["1040:filing_status"]
	fsCd := filingStatusCodes[fsStr]
	if fsCd == 0 {
		fsCd = 1 // default to single
	}

	return ReturnHeader{
		BinaryAttachmentCnt: 0,
		TaxYr:               taxYear,
		TaxPeriodBeginDt:    fmt.Sprintf("%d-01-01", taxYear),
		TaxPeriodEndDt:      fmt.Sprintf("%d-12-31", taxYear),
		Filer: Filer{
			PrimarySSN:     formatSSN(strInputs["1040:ssn"]),
			Name: FilerName{
				FirstName: strInputs["1040:first_name"],
				LastName:  strInputs["1040:last_name"],
			},
			FilingStatusCd: fsCd,
		},
	}
}

func buildIRS1040(r map[string]float64) *IRS1040 {
	return &IRS1040{
		WagesSalariesTips:       roundToInt(r["1040:1a"]),
		TaxExemptInterestAmt:    roundToInt(r["1040:2a"]),
		TaxableInterestAmt:      roundToInt(r["1040:2b"]),
		QualifiedDividendsAmt:   roundToInt(r["1040:3a"]),
		OrdinaryDividendsAmt:    roundToInt(r["1040:3b"]),
		CapitalGainLossAmt:      roundToInt(r["1040:7"]),
		OtherIncomeAmt:          roundToInt(r["1040:8"]),
		TotalIncomeAmt:          roundToInt(r["1040:9"]),
		AdjustmentsToIncomeAmt:  roundToInt(r["1040:10"]),
		AdjustedGrossIncomeAmt:  roundToInt(r["1040:11"]),
		TotalDeductionsAmt:      roundToInt(r["1040:14"]),
		TaxableIncomeAmt:        roundToInt(r["1040:15"]),
		TaxAmt:                  roundToInt(r["1040:16"]),
		Sch2PartIAmt:            roundToInt(r["1040:17"]),
		Sch3PartIAmt:            roundToInt(r["1040:20"]),
		TaxAfterCreditsAmt:      roundToInt(r["1040:22"]),
		OtherTaxesAmt:           roundToInt(r["1040:23"]),
		TotalTaxAmt:             roundToInt(r["1040:24"]),
		WithholdingTaxAmt:       roundToInt(r["1040:25d"]),
		EstimatedTaxPaymentsAmt: roundToInt(r["1040:31"]),
		TotalPaymentsAmt:        roundToInt(r["1040:33"]),
		OverpaidAmt:             roundToInt(r["1040:34"]),
		OwedAmt:                 roundToInt(r["1040:37"]),
	}
}

func buildScheduleA(r map[string]float64) *IRS1040ScheduleA {
	return &IRS1040ScheduleA{
		MedicalAndDentalExpAmt:    roundToInt(r["schedule_a:1"]),
		AGIAmt:                    roundToInt(r["schedule_a:2"]),
		MedicalFloorAmt:           roundToInt(r["schedule_a:3"]),
		DeductibleMedicalAmt:      roundToInt(r["schedule_a:4"]),
		StateLocalIncomeTaxAmt:    roundToInt(r["schedule_a:5a"]),
		PropertyTaxAmt:            roundToInt(r["schedule_a:5b"]),
		RealEstateTaxAmt:          roundToInt(r["schedule_a:5c"]),
		TotalSALTAmt:              roundToInt(r["schedule_a:5d"]),
		SALTDeductionAmt:          roundToInt(r["schedule_a:5e"]),
		MortgageInterestAmt:       roundToInt(r["schedule_a:8a"]),
		TotalInterestDeductionAmt: roundToInt(r["schedule_a:11"]),
		CashCharityAmt:            roundToInt(r["schedule_a:12"]),
		NonCashCharityAmt:         roundToInt(r["schedule_a:13"]),
		CharityCarryoverAmt:       roundToInt(r["schedule_a:14"]),
		TotalCharityAmt:           roundToInt(r["schedule_a:15"]),
		TotalItemizedDeductAmt:    roundToInt(r["schedule_a:17"]),
	}
}

func buildSchedule1(r map[string]float64) *IRS1040Schedule1 {
	return &IRS1040Schedule1{
		BusinessIncomeLossAmt:     roundToInt(r["schedule_1:3"]),
		CapitalGainLossAmt:        roundToInt(r["schedule_1:7"]),
		TotalAdditionalIncomeAmt:  roundToInt(r["schedule_1:10"]),
		HSADeductionAmt:           roundToInt(r["schedule_1:15"]),
		SETaxDeductionAmt:         roundToInt(r["schedule_1:16"]),
		EarlyWithdrawalPenaltyAmt: roundToInt(r["schedule_1:24"]),
		TotalAdjustmentsAmt:       roundToInt(r["schedule_1:26"]),
	}
}

func buildSchedule2(r map[string]float64) *IRS1040Schedule2 {
	return &IRS1040Schedule2{
		AMTAmt:                   roundToInt(r["schedule_2:1"]),
		TotalPartIAmt:            roundToInt(r["schedule_2:3"]),
		SelfEmploymentTaxAmt:     roundToInt(r["schedule_2:6"]),
		AdditionalMedicareTaxAmt: roundToInt(r["schedule_2:12"]),
		HSAPenaltyAmt:            roundToInt(r["schedule_2:17c"]),
		NIITAmt:                  roundToInt(r["schedule_2:18"]),
		TotalOtherTaxesAmt:       roundToInt(r["schedule_2:21"]),
	}
}

func buildSchedule3(r map[string]float64) *IRS1040Schedule3 {
	return &IRS1040Schedule3{
		TotalNonrefundableCreditsAmt: roundToInt(r["schedule_3:8"]),
		EstimatedTaxPaymentsAmt:      roundToInt(r["schedule_3:10"]),
		TotalOtherPaymentsAmt:        roundToInt(r["schedule_3:15"]),
	}
}

func buildScheduleB(r map[string]float64) *IRS1040ScheduleB {
	return &IRS1040ScheduleB{
		TotalInterestAmt:  roundToInt(r["schedule_b:4"]),
		TotalDividendsAmt: roundToInt(r["schedule_b:6"]),
	}
}

func buildScheduleC(r map[string]float64, s map[string]string) *IRS1040ScheduleC {
	return &IRS1040ScheduleC{
		BusinessName:     s["schedule_c:business_name"],
		BusinessCode:     s["schedule_c:business_code"],
		GrossReceiptsAmt: roundToInt(r["schedule_c:1"]),
		GrossProfitAmt:   roundToInt(r["schedule_c:5"]),
		TotalExpensesAmt: roundToInt(r["schedule_c:28"]),
		NetProfitLossAmt: roundToInt(r["schedule_c:31"]),
	}
}

func buildScheduleD(r map[string]float64) *IRS1040ScheduleD {
	return &IRS1040ScheduleD{
		STGainLossAmt:           roundToInt(r["schedule_d:1"]),
		NetSTGainLossAmt:        roundToInt(r["schedule_d:7"]),
		LTGainLossAmt:           roundToInt(r["schedule_d:8"]),
		CapGainDistributionsAmt: roundToInt(r["schedule_d:13"]),
		NetLTGainLossAmt:        roundToInt(r["schedule_d:15"]),
		NetCapitalGainLossAmt:   roundToInt(r["schedule_d:16"]),
	}
}

func buildScheduleSE(r map[string]float64) *IRS1040ScheduleSE {
	return &IRS1040ScheduleSE{
		NetSEEarningsAmt:     roundToInt(r["schedule_se:2"]),
		TaxableEarningsAmt:   roundToInt(r["schedule_se:3"]),
		SSTaxAmt:             roundToInt(r["schedule_se:4"]),
		MedicareTaxAmt:       roundToInt(r["schedule_se:5"]),
		SelfEmploymentTaxAmt: roundToInt(r["schedule_se:6"]),
		DeductibleSETaxAmt:   roundToInt(r["schedule_se:7"]),
	}
}

func buildForm8889(r map[string]float64, s map[string]string) *IRS8889 {
	return &IRS8889{
		CoverageType:         s["form_8889:1"],
		ContributionsAmt:     roundToInt(r["form_8889:2"]),
		EmployerContribAmt:   roundToInt(r["form_8889:3"]),
		ContributionLimitAmt: roundToInt(r["form_8889:6"]),
		HSADeductionAmt:      roundToInt(r["form_8889:9"]),
		DistributionsAmt:     roundToInt(r["form_8889:14a"]),
		QualifiedExpensesAmt: roundToInt(r["form_8889:14c"]),
		TaxableDistribAmt:    roundToInt(r["form_8889:15"]),
		PenaltyAmt:           roundToInt(r["form_8889:17b"]),
	}
}

func buildForm8949(r map[string]float64) *IRS8949 {
	return &IRS8949{
		STProceedsAmt: roundToInt(r["form_8949:st_proceeds"]),
		STBasisAmt:    roundToInt(r["form_8949:st_basis"]),
		STWashSaleAmt: roundToInt(r["form_8949:st_wash"]),
		STGainLossAmt: roundToInt(r["form_8949:st_gain_loss"]),
		LTProceedsAmt: roundToInt(r["form_8949:lt_proceeds"]),
		LTBasisAmt:    roundToInt(r["form_8949:lt_basis"]),
		LTWashSaleAmt: roundToInt(r["form_8949:lt_wash"]),
		LTGainLossAmt: roundToInt(r["form_8949:lt_gain_loss"]),
	}
}

func buildForm8995(r map[string]float64) *IRS8995 {
	return &IRS8995{
		TotalQBIAmt:            roundToInt(r["form_8995:3"]),
		QBIComponentAmt:        roundToInt(r["form_8995:4"]),
		TaxableIncBeforeQBIAmt: roundToInt(r["form_8995:5"]),
		IncomeLimitationAmt:    roundToInt(r["form_8995:8"]),
		QBIDeductionAmt:        roundToInt(r["form_8995:10"]),
	}
}

// buildW2s extracts W-2 instances from the solver results. W-2 input forms
// use instance keys like "w2:1:wages", "w2:2:wages", etc.
func buildW2s(r map[string]float64, s map[string]string) []IRSW2 {
	// Discover W-2 instances by scanning for w2:N:wages keys
	instances := make(map[string]bool)
	for k := range r {
		if strings.HasPrefix(k, "w2:") && strings.HasSuffix(k, ":wages") {
			parts := strings.SplitN(k, ":", 3)
			if len(parts) == 3 {
				instances[parts[1]] = true
			}
		}
	}
	// Also check string inputs for employer_name
	for k := range s {
		if strings.HasPrefix(k, "w2:") && strings.HasSuffix(k, ":employer_name") {
			parts := strings.SplitN(k, ":", 3)
			if len(parts) == 3 {
				instances[parts[1]] = true
			}
		}
	}

	if len(instances) == 0 {
		return nil
	}

	// Sort instances for deterministic output
	sorted := sortedKeys(instances)

	var w2s []IRSW2
	for _, inst := range sorted {
		prefix := "w2:" + inst + ":"
		w2 := IRSW2{
			EmployerName:   s[prefix+"employer_name"],
			EmployerEIN:    formatSSN(s[prefix+"employer_ein"]),
			WagesAmt:       roundToInt(r[prefix+"wages"]),
			WithholdingAmt: roundToInt(r[prefix+"federal_tax_withheld"]),
			SSWagesAmt:     roundToInt(r[prefix+"ss_wages"]),
			SSTaxAmt:       roundToInt(r[prefix+"ss_tax_withheld"]),
			MedicareWages:  roundToInt(r[prefix+"medicare_wages"]),
			MedicareTaxAmt: roundToInt(r[prefix+"medicare_tax_withheld"]),
		}
		w2s = append(w2s, w2)
	}
	return w2s
}

func buildForm2555(r map[string]float64, s map[string]string) *IRS2555 {
	return &IRS2555{
		ForeignCountry:         getStrVal(s, "form_2555:foreign_country"),
		QualifyingTest:         getStrVal(s, "form_2555:qualifying_test"),
		QualifyingDays:         roundToInt(r["form_2555:qualifying_days"]),
		ForeignEarnedIncomeAmt: roundToInt(r["form_2555:foreign_earned_income"]),
		ExclusionLimitAmt:      roundToInt(r["form_2555:exclusion_limit"]),
		ForeignIncomeExclAmt:   roundToInt(r["form_2555:foreign_income_exclusion"]),
		HousingExclusionAmt:    roundToInt(r["form_2555:housing_exclusion"]),
		HousingDeductionAmt:    roundToInt(r["form_2555:housing_deduction"]),
		TotalExclusionAmt:      roundToInt(r["form_2555:total_exclusion"]),
	}
}

func buildForm1116(r map[string]float64, s map[string]string) *IRS1116 {
	return &IRS1116{
		ForeignCountry:         getStrVal(s, "form_1116:foreign_country"),
		Category:               getStrVal(s, "form_1116:category"),
		ForeignSourceIncomeAmt: roundToInt(r["form_1116:7"]),
		ForeignTaxPaidAmt:      roundToInt(r["form_1116:15"]),
		CreditLimitationAmt:    roundToInt(r["form_1116:21"]),
		AllowedCreditAmt:       roundToInt(r["form_1116:22"]),
		CarryforwardAmt:        roundToInt(r["form_1116:carryforward"]),
	}
}

func buildForm8938(r map[string]float64, s map[string]string) *IRS8938 {
	return &IRS8938{
		LivesAbroad:             getStrVal(s, "form_8938:lives_abroad"),
		MaxValueAccountsAmt:     roundToInt(r["form_8938:max_value_accounts"]),
		YearEndValueAccountsAmt: roundToInt(r["form_8938:yearend_value_accounts"]),
		TotalMaxValueAmt:        roundToInt(r["form_8938:total_max_value"]),
		TotalYearEndValueAmt:    roundToInt(r["form_8938:total_yearend_value"]),
		FilingRequired:          roundToInt(r["form_8938:filing_required"]),
	}
}

func buildForm8833(r map[string]float64, s map[string]string) *IRS8833 {
	return &IRS8833{
		TreatyCountry:   getStrVal(s, "form_8833:treaty_country"),
		TreatyArticle:   getStrVal(s, "form_8833:treaty_article"),
		IRCProvision:    getStrVal(s, "form_8833:irc_provision"),
		TreatyAmountAmt: roundToInt(r["form_8833:treaty_amount"]),
		TreatyClaimed:   roundToInt(r["form_8833:treaty_claimed"]),
	}
}

// getStrVal safely retrieves a string value from the map.
func getStrVal(s map[string]string, key string) string {
	if s == nil {
		return ""
	}
	return s[key]
}

// sortedKeys returns the keys of a map[string]bool in sorted order.
func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// Simple insertion sort — W-2 count is always small
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}
