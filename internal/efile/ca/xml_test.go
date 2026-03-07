package ca

import (
	"bytes"
	"encoding/xml"
	"math"
	"testing"
)

// baseStrInputs returns string inputs for a simple single filer.
func baseStrInputs() map[string]string {
	return map[string]string{
		"1040:ssn":           "123456789",
		"1040:first_name":    "Jane",
		"1040:last_name":     "Doe",
		"1040:filing_status": "single",
	}
}

// simpleW2Results returns solver results for a single W-2 filer with no
// Schedule CA adjustments.
func simpleW2Results() map[string]float64 {
	return map[string]float64{
		// Form 540
		"ca_540:13": 75000,
		"ca_540:14": 0,
		"ca_540:15": 0,
		"ca_540:17": 75000,
		"ca_540:18": 5706,
		"ca_540:19": 69294,
		"ca_540:31": 3004,
		"ca_540:32": 144,
		"ca_540:35": 2860,
		"ca_540:36": 0,
		"ca_540:40": 2860,
		"ca_540:71": 3200,
		"ca_540:74": 3200,
		"ca_540:91": 340,
		"ca_540:93": 0,
		// Schedule CA — all zeros
		"ca_schedule_ca:2_col_b":  0,
		"ca_schedule_ca:2_col_c":  0,
		"ca_schedule_ca:3_col_b":  0,
		"ca_schedule_ca:3_col_c":  0,
		"ca_schedule_ca:7_col_b":  0,
		"ca_schedule_ca:7_col_c":  0,
		"ca_schedule_ca:15_col_c": 0,
		"ca_schedule_ca:5e_col_b": 0,
		"ca_schedule_ca:5e_col_c": 0,
		"ca_schedule_ca:ca_itemized": 0,
		"ca_schedule_ca:37_col_b": 0,
		"ca_schedule_ca:37_col_c": 0,
	}
}

func TestGenerateReturn_SimpleW2(t *testing.T) {
	xmlBytes, err := GenerateReturn(simpleW2Results(), baseStrInputs(), 2025)
	if err != nil {
		t.Fatalf("GenerateReturn failed: %v", err)
	}

	// Must be well-formed XML.
	var parsed CAReturn
	if err := xml.Unmarshal(xmlBytes, &parsed); err != nil {
		t.Fatalf("output is not well-formed XML: %v", err)
	}

	// Verify header fields.
	if parsed.Header.TaxYear != 2025 {
		t.Errorf("TaxYear = %d, want 2025", parsed.Header.TaxYear)
	}
	if parsed.Header.PrimarySSN != "123456789" {
		t.Errorf("PrimarySSN = %q, want %q", parsed.Header.PrimarySSN, "123456789")
	}
	if parsed.Header.FirstName != "Jane" {
		t.Errorf("FirstName = %q, want %q", parsed.Header.FirstName, "Jane")
	}
	if parsed.Header.LastName != "Doe" {
		t.Errorf("LastName = %q, want %q", parsed.Header.LastName, "Doe")
	}
	if parsed.Header.FilingStatusCd != "1" {
		t.Errorf("FilingStatusCd = %q, want %q", parsed.Header.FilingStatusCd, "1")
	}

	// Verify Form 540 amounts.
	if parsed.CA540.FederalAGIAmt != 75000 {
		t.Errorf("FederalAGIAmt = %d, want 75000", parsed.CA540.FederalAGIAmt)
	}
	if parsed.CA540.CATaxableIncomeAmt != 69294 {
		t.Errorf("CATaxableIncomeAmt = %d, want 69294", parsed.CA540.CATaxableIncomeAmt)
	}
	if parsed.CA540.TotalTaxAmt != 2860 {
		t.Errorf("TotalTaxAmt = %d, want 2860", parsed.CA540.TotalTaxAmt)
	}
	if parsed.CA540.OverpaidAmt != 340 {
		t.Errorf("OverpaidAmt = %d, want 340", parsed.CA540.OverpaidAmt)
	}
	if parsed.CA540.OwedAmt != 0 {
		t.Errorf("OwedAmt = %d, want 0", parsed.CA540.OwedAmt)
	}

	// Schedule CA should be omitted when all adjustments are zero.
	if parsed.ScheduleCA != nil {
		t.Error("ScheduleCA should be nil when all adjustments are zero")
	}

	// Verify the output contains the XML declaration.
	if !bytes.HasPrefix(xmlBytes, []byte("<?xml")) {
		t.Error("output should start with XML declaration")
	}

	// Verify namespace and version attributes.
	if parsed.Xmlns != "http://www.ftb.ca.gov/efile" {
		t.Errorf("xmlns = %q, want %q", parsed.Xmlns, "http://www.ftb.ca.gov/efile")
	}
	if parsed.Version != "2025.1" {
		t.Errorf("version = %q, want %q", parsed.Version, "2025.1")
	}
}

func TestGenerateReturn_Determinism(t *testing.T) {
	results := simpleW2Results()
	strInputs := baseStrInputs()

	first, err := GenerateReturn(results, strInputs, 2025)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	// Generate 10 more times and verify identical output.
	for i := 0; i < 10; i++ {
		got, err := GenerateReturn(results, strInputs, 2025)
		if err != nil {
			t.Fatalf("call %d failed: %v", i, err)
		}
		if !bytes.Equal(first, got) {
			t.Fatalf("call %d produced different output (non-deterministic)", i)
		}
	}
}

func TestGenerateReturn_ScheduleCA_USBondSubtraction(t *testing.T) {
	results := simpleW2Results()
	strInputs := baseStrInputs()

	// Simulate U.S. savings bond interest subtraction.
	results["ca_schedule_ca:2_col_b"] = 500
	results["ca_schedule_ca:37_col_b"] = 500
	results["ca_540:14"] = 500
	results["ca_540:17"] = 74500
	results["ca_540:19"] = 68794

	xmlBytes, err := GenerateReturn(results, strInputs, 2025)
	if err != nil {
		t.Fatalf("GenerateReturn failed: %v", err)
	}

	var parsed CAReturn
	if err := xml.Unmarshal(xmlBytes, &parsed); err != nil {
		t.Fatalf("output is not well-formed XML: %v", err)
	}

	// Schedule CA must be included.
	if parsed.ScheduleCA == nil {
		t.Fatal("ScheduleCA should be present when there are adjustments")
	}
	if parsed.ScheduleCA.InterestSubAmt != 500 {
		t.Errorf("InterestSubAmt = %d, want 500", parsed.ScheduleCA.InterestSubAmt)
	}
	if parsed.ScheduleCA.TotalSubAmt != 500 {
		t.Errorf("TotalSubAmt = %d, want 500", parsed.ScheduleCA.TotalSubAmt)
	}

	// Form 540 should reflect the subtraction.
	if parsed.CA540.CASubtractionsAmt != 500 {
		t.Errorf("CASubtractionsAmt = %d, want 500", parsed.CA540.CASubtractionsAmt)
	}
	if parsed.CA540.CAAGIAmt != 74500 {
		t.Errorf("CAAGIAmt = %d, want 74500", parsed.CA540.CAAGIAmt)
	}
}

func TestGenerateReturn_HSAAddBack(t *testing.T) {
	results := simpleW2Results()
	strInputs := baseStrInputs()

	// Simulate HSA deduction add-back.
	results["ca_schedule_ca:15_col_c"] = 3850
	results["ca_schedule_ca:37_col_c"] = 3850
	results["ca_540:15"] = 3850
	results["ca_540:17"] = 78850

	xmlBytes, err := GenerateReturn(results, strInputs, 2025)
	if err != nil {
		t.Fatalf("GenerateReturn failed: %v", err)
	}

	var parsed CAReturn
	if err := xml.Unmarshal(xmlBytes, &parsed); err != nil {
		t.Fatalf("output is not well-formed XML: %v", err)
	}

	if parsed.ScheduleCA == nil {
		t.Fatal("ScheduleCA should be present when HSA add-back exists")
	}
	if parsed.ScheduleCA.HSAAddBackAmt != 3850 {
		t.Errorf("HSAAddBackAmt = %d, want 3850", parsed.ScheduleCA.HSAAddBackAmt)
	}
	if parsed.ScheduleCA.TotalAddAmt != 3850 {
		t.Errorf("TotalAddAmt = %d, want 3850", parsed.ScheduleCA.TotalAddAmt)
	}
	if parsed.CA540.CAAdditionsAmt != 3850 {
		t.Errorf("CAAdditionsAmt = %d, want 3850", parsed.CA540.CAAdditionsAmt)
	}
}

func TestGenerateReturn_Rounding(t *testing.T) {
	results := simpleW2Results()
	strInputs := baseStrInputs()

	// Set fractional amounts to verify rounding.
	results["ca_540:13"] = 75000.49
	results["ca_540:17"] = 75000.49
	results["ca_540:31"] = 3004.50
	results["ca_540:32"] = 144.51
	results["ca_540:71"] = 3200.99

	xmlBytes, err := GenerateReturn(results, strInputs, 2025)
	if err != nil {
		t.Fatalf("GenerateReturn failed: %v", err)
	}

	var parsed CAReturn
	if err := xml.Unmarshal(xmlBytes, &parsed); err != nil {
		t.Fatalf("output is not well-formed XML: %v", err)
	}

	// 75000.49 rounds to 75000
	if parsed.CA540.FederalAGIAmt != 75000 {
		t.Errorf("FederalAGIAmt = %d, want 75000 (rounded from 75000.49)", parsed.CA540.FederalAGIAmt)
	}
	// 3004.50 rounds to 3005 (round half away from zero)
	if parsed.CA540.CATaxAmt != 3005 {
		t.Errorf("CATaxAmt = %d, want 3005 (rounded from 3004.50)", parsed.CA540.CATaxAmt)
	}
	// 144.51 rounds to 145
	if parsed.CA540.ExemptionCreditAmt != 145 {
		t.Errorf("ExemptionCreditAmt = %d, want 145 (rounded from 144.51)", parsed.CA540.ExemptionCreditAmt)
	}
	// 3200.99 rounds to 3201
	if parsed.CA540.WithholdingAmt != 3201 {
		t.Errorf("WithholdingAmt = %d, want 3201 (rounded from 3200.99)", parsed.CA540.WithholdingAmt)
	}
}

func TestRoundToInt(t *testing.T) {
	tests := []struct {
		input float64
		want  int
	}{
		{0, 0},
		{1.4, 1},
		{1.5, 2},
		{2.5, 3}, // round half away from zero
		{-1.4, -1},
		{-1.5, -2},
		{99999.99, 100000},
		{0.001, 0},
	}
	for _, tt := range tests {
		got := roundToInt(tt.input)
		if got != tt.want {
			t.Errorf("roundToInt(%v) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestFilingStatusCode(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"single", "1"},
		{"mfj", "2"},
		{"mfs", "3"},
		{"hoh", "4"},
		{"qss", "5"},
		{"", "1"},       // default
		{"unknown", "1"}, // default
	}
	for _, tt := range tests {
		got := filingStatusCode(tt.input)
		if got != tt.want {
			t.Errorf("filingStatusCode(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestGenerateReturn_NilInputs(t *testing.T) {
	_, err := GenerateReturn(nil, baseStrInputs(), 2025)
	if err == nil {
		t.Error("expected error for nil results")
	}

	_, err = GenerateReturn(simpleW2Results(), nil, 2025)
	if err == nil {
		t.Error("expected error for nil strInputs")
	}
}

func TestGenerateReturn_FilingStatusCodes(t *testing.T) {
	results := simpleW2Results()

	statuses := []struct {
		input string
		want  string
	}{
		{"single", "1"},
		{"mfj", "2"},
		{"mfs", "3"},
		{"hoh", "4"},
		{"qss", "5"},
	}

	for _, tt := range statuses {
		t.Run(tt.input, func(t *testing.T) {
			strInputs := baseStrInputs()
			strInputs["1040:filing_status"] = tt.input

			xmlBytes, err := GenerateReturn(results, strInputs, 2025)
			if err != nil {
				t.Fatalf("GenerateReturn failed: %v", err)
			}

			var parsed CAReturn
			if err := xml.Unmarshal(xmlBytes, &parsed); err != nil {
				t.Fatalf("output is not well-formed XML: %v", err)
			}
			if parsed.Header.FilingStatusCd != tt.want {
				t.Errorf("FilingStatusCd = %q, want %q", parsed.Header.FilingStatusCd, tt.want)
			}
		})
	}
}

func TestGenerateReturn_WellFormedXML(t *testing.T) {
	xmlBytes, err := GenerateReturn(simpleW2Results(), baseStrInputs(), 2025)
	if err != nil {
		t.Fatalf("GenerateReturn failed: %v", err)
	}

	// Verify full round-trip: unmarshal then re-marshal should produce
	// structurally equivalent XML.
	var parsed CAReturn
	if err := xml.Unmarshal(xmlBytes, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	remarshaled, err := xml.MarshalIndent(parsed, "", "  ")
	if err != nil {
		t.Fatalf("re-marshal failed: %v", err)
	}

	// The re-marshaled output (without header) should match the original
	// (without header). We strip the XML declaration for comparison since
	// MarshalIndent does not add it.
	original := xmlBytes[len(xml.Header):]
	if !bytes.Equal(original, remarshaled) {
		t.Error("round-trip XML mismatch — indicates structural issue")
	}
}

// Verify that math.Round is used (not floor/ceil).
func TestRoundToInt_HalfValues(t *testing.T) {
	// math.Round uses "round half away from zero" in Go.
	// Verify our function matches math.Round behavior.
	vals := []float64{0.5, 1.5, 2.5, 3.5, -0.5, -1.5}
	for _, v := range vals {
		got := roundToInt(v)
		want := int(math.Round(v))
		if got != want {
			t.Errorf("roundToInt(%v) = %d, want %d", v, got, want)
		}
	}
}
