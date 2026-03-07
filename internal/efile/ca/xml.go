package ca

import (
	"encoding/xml"
	"fmt"
	"math"
)

// CAReturn is the root element for a California FTB e-file XML return.
type CAReturn struct {
	XMLName   xml.Name       `xml:"CAReturn"`
	Xmlns     string         `xml:"xmlns,attr"`
	Version   string         `xml:"version,attr"`
	Header    CAReturnHeader `xml:"CAReturnHeader"`
	CA540     CA540          `xml:"CA540"`
	ScheduleCA *CAScheduleCA `xml:"CAScheduleCA,omitempty"`
}

// CAReturnHeader contains taxpayer identification and filing metadata.
type CAReturnHeader struct {
	TaxYear       int    `xml:"TaxYear"`
	PrimarySSN    string `xml:"PrimarySSN"`
	FirstName     string `xml:"FirstName"`
	LastName      string `xml:"LastName"`
	FilingStatusCd string `xml:"FilingStatusCd"`
}

// CA540 contains the line amounts for California Form 540.
type CA540 struct {
	FederalAGIAmt      int `xml:"FederalAGIAmt"`
	CASubtractionsAmt  int `xml:"CASubtractionsAmt"`
	CAAdditionsAmt     int `xml:"CAAdditionsAmt"`
	CAAGIAmt           int `xml:"CAAGIAmt"`
	CADeductionAmt     int `xml:"CADeductionAmt"`
	CATaxableIncomeAmt int `xml:"CATaxableIncomeAmt"`
	CATaxAmt           int `xml:"CATaxAmt"`
	ExemptionCreditAmt int `xml:"ExemptionCreditAmt"`
	NetTaxAmt          int `xml:"NetTaxAmt"`
	MentalHealthTaxAmt int `xml:"MentalHealthTaxAmt"`
	TotalTaxAmt        int `xml:"TotalTaxAmt"`
	WithholdingAmt     int `xml:"WithholdingAmt"`
	TotalPaymentsAmt   int `xml:"TotalPaymentsAmt"`
	OverpaidAmt        int `xml:"OverpaidAmt"`
	OwedAmt            int `xml:"OwedAmt"`
}

// CAScheduleCA contains adjustment amounts from Schedule CA (540).
type CAScheduleCA struct {
	InterestSubAmt    int `xml:"InterestSubAmt"`
	InterestAddAmt    int `xml:"InterestAddAmt"`
	DividendSubAmt    int `xml:"DividendSubAmt"`
	DividendAddAmt    int `xml:"DividendAddAmt"`
	CapGainSubAmt     int `xml:"CapGainSubAmt"`
	CapGainAddAmt     int `xml:"CapGainAddAmt"`
	HSAAddBackAmt        int `xml:"HSAAddBackAmt"`
	FEIEAddBackAmt       int `xml:"FEIEAddBackAmt,omitempty"`
	HousingAddBackAmt    int `xml:"HousingAddBackAmt,omitempty"`
	SALTSubAmt           int `xml:"SALTSubAmt"`
	PropertyTaxAddAmt int `xml:"PropertyTaxAddAmt"`
	CAItemizedAmt     int `xml:"CAItemizedAmt"`
	TotalSubAmt       int `xml:"TotalSubAmt"`
	TotalAddAmt       int `xml:"TotalAddAmt"`
}

// GenerateReturn produces CA FTB e-file XML from solver results.
// It takes the numeric results, string inputs (for SSN, name, filing status),
// and the tax year. The output is deterministic: identical inputs always
// produce identical XML.
func GenerateReturn(results map[string]float64, strInputs map[string]string, taxYear int) ([]byte, error) {
	if results == nil {
		return nil, fmt.Errorf("results map must not be nil")
	}
	if strInputs == nil {
		return nil, fmt.Errorf("strInputs map must not be nil")
	}

	ret := CAReturn{
		Xmlns:   "http://www.ftb.ca.gov/efile",
		Version: fmt.Sprintf("%d.1", taxYear),
		Header: CAReturnHeader{
			TaxYear:        taxYear,
			PrimarySSN:     strInputs["1040:ssn"],
			FirstName:      strInputs["1040:first_name"],
			LastName:       strInputs["1040:last_name"],
			FilingStatusCd: filingStatusCode(strInputs["1040:filing_status"]),
		},
		CA540: CA540{
			FederalAGIAmt:      roundToInt(results["ca_540:13"]),
			CASubtractionsAmt:  roundToInt(results["ca_540:14"]),
			CAAdditionsAmt:     roundToInt(results["ca_540:15"]),
			CAAGIAmt:           roundToInt(results["ca_540:17"]),
			CADeductionAmt:     roundToInt(results["ca_540:18"]),
			CATaxableIncomeAmt: roundToInt(results["ca_540:19"]),
			CATaxAmt:           roundToInt(results["ca_540:31"]),
			ExemptionCreditAmt: roundToInt(results["ca_540:32"]),
			NetTaxAmt:          roundToInt(results["ca_540:35"]),
			MentalHealthTaxAmt: roundToInt(results["ca_540:36"]),
			TotalTaxAmt:        roundToInt(results["ca_540:40"]),
			WithholdingAmt:     roundToInt(results["ca_540:71"]),
			TotalPaymentsAmt:   roundToInt(results["ca_540:74"]),
			OverpaidAmt:        roundToInt(results["ca_540:91"]),
			OwedAmt:            roundToInt(results["ca_540:93"]),
		},
	}

	// Build Schedule CA and include it only if there are non-zero adjustments.
	sca := CAScheduleCA{
		InterestSubAmt:    roundToInt(results["ca_schedule_ca:2_col_b"]),
		InterestAddAmt:    roundToInt(results["ca_schedule_ca:2_col_c"]),
		DividendSubAmt:    roundToInt(results["ca_schedule_ca:3_col_b"]),
		DividendAddAmt:    roundToInt(results["ca_schedule_ca:3_col_c"]),
		CapGainSubAmt:     roundToInt(results["ca_schedule_ca:7_col_b"]),
		CapGainAddAmt:     roundToInt(results["ca_schedule_ca:7_col_c"]),
		HSAAddBackAmt:        roundToInt(results["ca_schedule_ca:15_col_c"]),
		FEIEAddBackAmt:      roundToInt(results["ca_schedule_ca:8d_col_c"]),
		HousingAddBackAmt:   roundToInt(results["ca_schedule_ca:8d_col_c_housing"]),
		SALTSubAmt:          roundToInt(results["ca_schedule_ca:5e_col_b"]),
		PropertyTaxAddAmt: roundToInt(results["ca_schedule_ca:5e_col_c"]),
		CAItemizedAmt:     roundToInt(results["ca_schedule_ca:ca_itemized"]),
		TotalSubAmt:       roundToInt(results["ca_schedule_ca:37_col_b"]),
		TotalAddAmt:       roundToInt(results["ca_schedule_ca:37_col_c"]),
	}

	if hasNonZeroAdjustments(sca) {
		ret.ScheduleCA = &sca
	}

	out, err := xml.MarshalIndent(ret, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling CA return XML: %w", err)
	}

	// Prepend XML declaration for a well-formed document.
	header := []byte(xml.Header)
	return append(header, out...), nil
}

// roundToInt rounds a float64 to the nearest integer using math.Round
// (round half away from zero), consistent with IRS rounding rules.
func roundToInt(f float64) int {
	return int(math.Round(f))
}

// filingStatusCode maps a filing status string to the CA FTB numeric code.
//
//	1 = Single
//	2 = Married/RDP Filing Jointly
//	3 = Married/RDP Filing Separately
//	4 = Head of Household
//	5 = Qualifying Surviving Spouse/RDP
func filingStatusCode(fs string) string {
	switch fs {
	case "single":
		return "1"
	case "mfj":
		return "2"
	case "mfs":
		return "3"
	case "hoh":
		return "4"
	case "qss":
		return "5"
	default:
		return "1" // default to single
	}
}

// hasNonZeroAdjustments returns true if any Schedule CA field is non-zero.
func hasNonZeroAdjustments(sca CAScheduleCA) bool {
	return sca.InterestSubAmt != 0 ||
		sca.InterestAddAmt != 0 ||
		sca.DividendSubAmt != 0 ||
		sca.DividendAddAmt != 0 ||
		sca.CapGainSubAmt != 0 ||
		sca.CapGainAddAmt != 0 ||
		sca.HSAAddBackAmt != 0 ||
		sca.SALTSubAmt != 0 ||
		sca.PropertyTaxAddAmt != 0 ||
		sca.CAItemizedAmt != 0 ||
		sca.TotalSubAmt != 0 ||
		sca.TotalAddAmt != 0
}
