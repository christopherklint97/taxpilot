package mef

import (
	"encoding/xml"
	"strings"
	"testing"
)

func TestRoundToInt(t *testing.T) {
	tests := []struct {
		input float64
		want  int
	}{
		{0, 0},
		{1.4, 1},
		{1.5, 2},
		{1.6, 2},
		{-1.4, -1},
		{-1.5, -2},
		{99999.999, 100000},
		{75000.50, 75001},
		{75000.49, 75000},
	}
	for _, tc := range tests {
		got := roundToInt(tc.input)
		if got != tc.want {
			t.Errorf("roundToInt(%v) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestFormatSSN(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"123-45-6789", "123456789"},
		{"123456789", "123456789"},
		{"000-00-0000", "000000000"},
		{"", ""},
	}
	for _, tc := range tests {
		got := formatSSN(tc.input)
		if got != tc.want {
			t.Errorf("formatSSN(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestIsScheduleNeeded(t *testing.T) {
	results := map[string]float64{
		"schedule_a:17": 0,
		"schedule_a:5e": 0,
		"schedule_c:31": 50000,
		"schedule_c:1":  60000,
		"schedule_b:4":  0,
		"schedule_b:6":  0,
	}

	if isScheduleNeeded(results, "schedule_a:") {
		t.Error("schedule_a should not be needed when all values are zero")
	}
	if !isScheduleNeeded(results, "schedule_c:") {
		t.Error("schedule_c should be needed when it has non-zero values")
	}
	if isScheduleNeeded(results, "schedule_b:") {
		t.Error("schedule_b should not be needed when all values are zero")
	}
	if isScheduleNeeded(results, "schedule_d:") {
		t.Error("schedule_d should not be needed when no keys exist")
	}
}

// simpleW2Scenario returns solver results for a basic single W-2 filer.
func simpleW2Scenario() (map[string]float64, map[string]string) {
	results := map[string]float64{
		// W-2 inputs
		"w2:1:wages":                75000,
		"w2:1:federal_tax_withheld": 9500,
		"w2:1:ss_wages":             75000,
		"w2:1:ss_tax_withheld":      4650,
		"w2:1:medicare_wages":       75000,
		"w2:1:medicare_tax_withheld": 1087.50,

		// 1040 computed
		"1040:1a":  75000,
		"1040:1z":  75000,
		"1040:2a":  0,
		"1040:2b":  0,
		"1040:3a":  0,
		"1040:3b":  0,
		"1040:7":   0,
		"1040:8":   0,
		"1040:9":   75000,
		"1040:10":  0,
		"1040:11":  75000,
		"1040:12":  15000,
		"1040:13":  0,
		"1040:14":  15000,
		"1040:15":  60000,
		"1040:16":  8114,
		"1040:17":  0,
		"1040:20":  0,
		"1040:22":  8114,
		"1040:23":  0,
		"1040:24":  8114,
		"1040:25a": 9500,
		"1040:25b": 0,
		"1040:25d": 9500,
		"1040:31":  0,
		"1040:33":  9500,
		"1040:34":  1386,
		"1040:37":  0,
	}
	strInputs := map[string]string{
		"1040:filing_status": "single",
		"1040:first_name":    "Jane",
		"1040:last_name":     "Doe",
		"1040:ssn":           "123-45-6789",
		"w2:1:employer_name": "Acme Corp",
		"w2:1:employer_ein":  "12-3456789",
	}
	return results, strInputs
}

func TestGenerateReturn_SimpleW2(t *testing.T) {
	results, strInputs := simpleW2Scenario()
	xmlBytes, err := GenerateReturn(results, strInputs, 2025)
	if err != nil {
		t.Fatalf("GenerateReturn failed: %v", err)
	}

	xmlStr := string(xmlBytes)

	// Verify XML declaration
	if !strings.HasPrefix(xmlStr, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>") {
		t.Error("XML should start with XML declaration")
	}

	// Verify well-formed XML by parsing it back
	var parsed Return
	if err := xml.Unmarshal(xmlBytes, &parsed); err != nil {
		t.Fatalf("Generated XML is not well-formed: %v", err)
	}

	// Verify namespace and version
	if parsed.Xmlns != MeFNamespace {
		t.Errorf("namespace = %q, want %q", parsed.Xmlns, MeFNamespace)
	}
	if parsed.ReturnVersion != ReturnVersion {
		t.Errorf("returnVersion = %q, want %q", parsed.ReturnVersion, ReturnVersion)
	}

	// Verify header
	h := parsed.ReturnHeader
	if h.TaxYr != 2025 {
		t.Errorf("TaxYr = %d, want 2025", h.TaxYr)
	}
	if h.TaxPeriodBeginDt != "2025-01-01" {
		t.Errorf("TaxPeriodBeginDt = %q, want 2025-01-01", h.TaxPeriodBeginDt)
	}
	if h.TaxPeriodEndDt != "2025-12-31" {
		t.Errorf("TaxPeriodEndDt = %q, want 2025-12-31", h.TaxPeriodEndDt)
	}
	if h.Filer.PrimarySSN != "123456789" {
		t.Errorf("SSN = %q, want 123456789", h.Filer.PrimarySSN)
	}
	if h.Filer.Name.FirstName != "Jane" {
		t.Errorf("FirstName = %q, want Jane", h.Filer.Name.FirstName)
	}
	if h.Filer.Name.LastName != "Doe" {
		t.Errorf("LastName = %q, want Doe", h.Filer.Name.LastName)
	}
	if h.Filer.FilingStatusCd != 1 {
		t.Errorf("FilingStatusCd = %d, want 1 (single)", h.Filer.FilingStatusCd)
	}

	// Verify 1040 data
	f := parsed.ReturnData.IRS1040
	if f == nil {
		t.Fatal("IRS1040 should always be present")
	}
	if f.WagesSalariesTips != 75000 {
		t.Errorf("WagesSalariesTips = %d, want 75000", f.WagesSalariesTips)
	}
	if f.TotalIncomeAmt != 75000 {
		t.Errorf("TotalIncomeAmt = %d, want 75000", f.TotalIncomeAmt)
	}
	if f.AdjustedGrossIncomeAmt != 75000 {
		t.Errorf("AdjustedGrossIncomeAmt = %d, want 75000", f.AdjustedGrossIncomeAmt)
	}
	if f.TotalDeductionsAmt != 15000 {
		t.Errorf("TotalDeductionsAmt = %d, want 15000", f.TotalDeductionsAmt)
	}
	if f.TaxableIncomeAmt != 60000 {
		t.Errorf("TaxableIncomeAmt = %d, want 60000", f.TaxableIncomeAmt)
	}
	if f.TaxAmt != 8114 {
		t.Errorf("TaxAmt = %d, want 8114", f.TaxAmt)
	}
	if f.TotalTaxAmt != 8114 {
		t.Errorf("TotalTaxAmt = %d, want 8114", f.TotalTaxAmt)
	}
	if f.WithholdingTaxAmt != 9500 {
		t.Errorf("WithholdingTaxAmt = %d, want 9500", f.WithholdingTaxAmt)
	}
	if f.TotalPaymentsAmt != 9500 {
		t.Errorf("TotalPaymentsAmt = %d, want 9500", f.TotalPaymentsAmt)
	}
	if f.OverpaidAmt != 1386 {
		t.Errorf("OverpaidAmt = %d, want 1386", f.OverpaidAmt)
	}

	// Verify W-2 data
	if len(parsed.ReturnData.IRSW2) != 1 {
		t.Fatalf("expected 1 W-2, got %d", len(parsed.ReturnData.IRSW2))
	}
	w := parsed.ReturnData.IRSW2[0]
	if w.EmployerName != "Acme Corp" {
		t.Errorf("EmployerName = %q, want Acme Corp", w.EmployerName)
	}
	if w.EmployerEIN != "123456789" {
		t.Errorf("EmployerEIN = %q, want 123456789", w.EmployerEIN)
	}
	if w.WagesAmt != 75000 {
		t.Errorf("WagesAmt = %d, want 75000", w.WagesAmt)
	}
	if w.WithholdingAmt != 9500 {
		t.Errorf("WithholdingAmt = %d, want 9500", w.WithholdingAmt)
	}

	// Verify schedules with all-zero values are omitted
	if parsed.ReturnData.IRS1040ScheduleA != nil {
		t.Error("Schedule A should be omitted when not needed")
	}
	if parsed.ReturnData.IRS1040ScheduleC != nil {
		t.Error("Schedule C should be omitted when not needed")
	}
	if parsed.ReturnData.IRS1040ScheduleD != nil {
		t.Error("Schedule D should be omitted when not needed")
	}
	if parsed.ReturnData.IRS1040ScheduleSE != nil {
		t.Error("Schedule SE should be omitted when not needed")
	}
	if parsed.ReturnData.IRS8949 != nil {
		t.Error("Form 8949 should be omitted when not needed")
	}
	if parsed.ReturnData.IRS8889 != nil {
		t.Error("Form 8889 should be omitted when not needed")
	}
	if parsed.ReturnData.IRS8995 != nil {
		t.Error("Form 8995 should be omitted when not needed")
	}

	// Verify XML contains expected element names in output
	if !strings.Contains(xmlStr, "<WagesSalariesTips>75000</WagesSalariesTips>") {
		t.Error("XML should contain WagesSalariesTips element")
	}
	if !strings.Contains(xmlStr, "<EmployerName>Acme Corp</EmployerName>") {
		t.Error("XML should contain EmployerName element")
	}
}

func TestGenerateReturn_Determinism(t *testing.T) {
	results, strInputs := simpleW2Scenario()

	xml1, err := GenerateReturn(results, strInputs, 2025)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	// Generate multiple times and verify identical output
	for i := 0; i < 10; i++ {
		xml2, err := GenerateReturn(results, strInputs, 2025)
		if err != nil {
			t.Fatalf("call %d failed: %v", i+1, err)
		}
		if string(xml1) != string(xml2) {
			t.Fatalf("call %d produced different XML (not deterministic)", i+1)
		}
	}
}

func TestGenerateReturn_SchedulesOmittedWhenZero(t *testing.T) {
	// Minimal scenario: just a W-2, no other forms
	results, strInputs := simpleW2Scenario()
	xmlBytes, err := GenerateReturn(results, strInputs, 2025)
	if err != nil {
		t.Fatalf("GenerateReturn failed: %v", err)
	}
	xmlStr := string(xmlBytes)

	omitted := []string{
		"IRS1040ScheduleA",
		"IRS1040ScheduleC",
		"IRS1040ScheduleD",
		"IRS1040ScheduleSE",
		"IRS8949",
		"IRS8889",
		"IRS8995",
	}
	for _, tag := range omitted {
		if strings.Contains(xmlStr, "<"+tag+">") || strings.Contains(xmlStr, "<"+tag+" ") {
			t.Errorf("%s should be omitted when all values are zero", tag)
		}
	}
}

// selfEmployedScenario returns solver results for a self-employed filer
// with Schedule C + Schedule SE.
func selfEmployedScenario() (map[string]float64, map[string]string) {
	results := map[string]float64{
		// W-2 inputs (also has a W-2 job)
		"w2:1:wages":                50000,
		"w2:1:federal_tax_withheld": 6000,
		"w2:1:ss_wages":             50000,
		"w2:1:ss_tax_withheld":      3100,
		"w2:1:medicare_wages":       50000,
		"w2:1:medicare_tax_withheld": 725,

		// Schedule C
		"schedule_c:1":  80000,
		"schedule_c:4":  0,
		"schedule_c:5":  80000,
		"schedule_c:7":  80000,
		"schedule_c:8":  1000,
		"schedule_c:10": 2000,
		"schedule_c:17": 500,
		"schedule_c:18": 800,
		"schedule_c:22": 300,
		"schedule_c:25": 400,
		"schedule_c:27": 1000,
		"schedule_c:28": 6000,
		"schedule_c:31": 74000,

		// Schedule SE
		"schedule_se:2": 74000,
		"schedule_se:3": 68338.9, // 74000 * 0.9235
		"schedule_se:4": 8474.02, // min(68338.9, 176100-50000) * 0.124
		"schedule_se:5": 1981.83, // 68338.9 * 0.029
		"schedule_se:6": 10455.85,
		"schedule_se:7": 5227.93, // 50% of SE tax

		// Schedule 1
		"schedule_1:1":  0,
		"schedule_1:2a": 0,
		"schedule_1:3":  74000,
		"schedule_1:7":  0,
		"schedule_1:10": 74000,
		"schedule_1:11": 0,
		"schedule_1:15": 0,
		"schedule_1:16": 5227.93,
		"schedule_1:20": 0,
		"schedule_1:21": 0,
		"schedule_1:24": 0,
		"schedule_1:26": 5227.93,

		// Schedule 2
		"schedule_2:1":   0,
		"schedule_2:2":   0,
		"schedule_2:3":   0,
		"schedule_2:6":   10455.85,
		"schedule_2:12":  0,
		"schedule_2:17c": 0,
		"schedule_2:18":  0,
		"schedule_2:21":  10455.85,

		// 1040 computed
		"1040:1a":  50000,
		"1040:1z":  50000,
		"1040:2a":  0,
		"1040:2b":  0,
		"1040:3a":  0,
		"1040:3b":  0,
		"1040:7":   0,
		"1040:8":   74000,
		"1040:9":   124000,
		"1040:10":  5227.93,
		"1040:11":  118772.07,
		"1040:12":  15000,
		"1040:13":  0,
		"1040:14":  15000,
		"1040:15":  103772.07,
		"1040:16":  17650,
		"1040:17":  0,
		"1040:20":  0,
		"1040:22":  17650,
		"1040:23":  10455.85,
		"1040:24":  28105.85,
		"1040:25a": 6000,
		"1040:25b": 0,
		"1040:25d": 6000,
		"1040:31":  0,
		"1040:33":  6000,
		"1040:34":  0,
		"1040:37":  22105.85,
	}
	strInputs := map[string]string{
		"1040:filing_status":      "single",
		"1040:first_name":         "Bob",
		"1040:last_name":          "Builder",
		"1040:ssn":                "987-65-4321",
		"w2:1:employer_name":      "Day Job Inc",
		"w2:1:employer_ein":       "98-7654321",
		"schedule_c:business_name": "Bob's Consulting",
		"schedule_c:business_code": "541611",
	}
	return results, strInputs
}

func TestGenerateReturn_SelfEmployed(t *testing.T) {
	results, strInputs := selfEmployedScenario()
	xmlBytes, err := GenerateReturn(results, strInputs, 2025)
	if err != nil {
		t.Fatalf("GenerateReturn failed: %v", err)
	}

	var parsed Return
	if err := xml.Unmarshal(xmlBytes, &parsed); err != nil {
		t.Fatalf("Generated XML is not well-formed: %v", err)
	}

	// Verify header
	if parsed.ReturnHeader.Filer.PrimarySSN != "987654321" {
		t.Errorf("SSN = %q, want 987654321", parsed.ReturnHeader.Filer.PrimarySSN)
	}
	if parsed.ReturnHeader.Filer.Name.FirstName != "Bob" {
		t.Errorf("FirstName = %q, want Bob", parsed.ReturnHeader.Filer.Name.FirstName)
	}

	// Verify Schedule C is present
	sc := parsed.ReturnData.IRS1040ScheduleC
	if sc == nil {
		t.Fatal("Schedule C should be present for self-employed scenario")
	}
	if sc.BusinessName != "Bob's Consulting" {
		t.Errorf("BusinessName = %q, want Bob's Consulting", sc.BusinessName)
	}
	if sc.BusinessCode != "541611" {
		t.Errorf("BusinessCode = %q, want 541611", sc.BusinessCode)
	}
	if sc.GrossReceiptsAmt != 80000 {
		t.Errorf("GrossReceiptsAmt = %d, want 80000", sc.GrossReceiptsAmt)
	}
	if sc.NetProfitLossAmt != 74000 {
		t.Errorf("NetProfitLossAmt = %d, want 74000", sc.NetProfitLossAmt)
	}
	if sc.TotalExpensesAmt != 6000 {
		t.Errorf("TotalExpensesAmt = %d, want 6000", sc.TotalExpensesAmt)
	}

	// Verify Schedule SE is present
	se := parsed.ReturnData.IRS1040ScheduleSE
	if se == nil {
		t.Fatal("Schedule SE should be present for self-employed scenario")
	}
	if se.NetSEEarningsAmt != 74000 {
		t.Errorf("NetSEEarningsAmt = %d, want 74000", se.NetSEEarningsAmt)
	}
	if se.SelfEmploymentTaxAmt != 10456 { // roundToInt(10455.85)
		t.Errorf("SelfEmploymentTaxAmt = %d, want 10456", se.SelfEmploymentTaxAmt)
	}
	if se.DeductibleSETaxAmt != 5228 { // roundToInt(5227.93)
		t.Errorf("DeductibleSETaxAmt = %d, want 5228", se.DeductibleSETaxAmt)
	}

	// Verify Schedule 1 is present
	s1 := parsed.ReturnData.IRS1040Schedule1
	if s1 == nil {
		t.Fatal("Schedule 1 should be present for self-employed scenario")
	}
	if s1.BusinessIncomeLossAmt != 74000 {
		t.Errorf("BusinessIncomeLossAmt = %d, want 74000", s1.BusinessIncomeLossAmt)
	}
	if s1.SETaxDeductionAmt != 5228 {
		t.Errorf("SETaxDeductionAmt = %d, want 5228", s1.SETaxDeductionAmt)
	}

	// Verify Schedule 2 is present
	s2 := parsed.ReturnData.IRS1040Schedule2
	if s2 == nil {
		t.Fatal("Schedule 2 should be present for self-employed scenario")
	}
	if s2.SelfEmploymentTaxAmt != 10456 {
		t.Errorf("SelfEmploymentTaxAmt = %d, want 10456", s2.SelfEmploymentTaxAmt)
	}

	// Verify 1040 amounts
	f := parsed.ReturnData.IRS1040
	if f.TotalIncomeAmt != 124000 {
		t.Errorf("TotalIncomeAmt = %d, want 124000", f.TotalIncomeAmt)
	}
	if f.OtherIncomeAmt != 74000 {
		t.Errorf("OtherIncomeAmt = %d, want 74000", f.OtherIncomeAmt)
	}
	if f.OwedAmt != 22106 { // roundToInt(22105.85)
		t.Errorf("OwedAmt = %d, want 22106", f.OwedAmt)
	}

	// Verify schedules NOT needed are omitted
	if parsed.ReturnData.IRS1040ScheduleA != nil {
		t.Error("Schedule A should be omitted")
	}
	if parsed.ReturnData.IRS1040ScheduleD != nil {
		t.Error("Schedule D should be omitted")
	}
	if parsed.ReturnData.IRS8949 != nil {
		t.Error("Form 8949 should be omitted")
	}
}

func TestGenerateReturn_FilingStatusCodes(t *testing.T) {
	tests := []struct {
		status string
		code   int
	}{
		{"single", 1},
		{"mfj", 2},
		{"mfs", 3},
		{"hoh", 4},
		{"qss", 5},
	}
	for _, tc := range tests {
		results, strInputs := simpleW2Scenario()
		strInputs["1040:filing_status"] = tc.status
		xmlBytes, err := GenerateReturn(results, strInputs, 2025)
		if err != nil {
			t.Fatalf("GenerateReturn(%s) failed: %v", tc.status, err)
		}
		var parsed Return
		if err := xml.Unmarshal(xmlBytes, &parsed); err != nil {
			t.Fatalf("XML not well-formed for %s: %v", tc.status, err)
		}
		if parsed.ReturnHeader.Filer.FilingStatusCd != tc.code {
			t.Errorf("status %q: FilingStatusCd = %d, want %d",
				tc.status, parsed.ReturnHeader.Filer.FilingStatusCd, tc.code)
		}
	}
}

func TestGenerateReturn_DocumentCount(t *testing.T) {
	// Simple scenario: 1040 + 1 W-2 = 2 documents
	results, strInputs := simpleW2Scenario()
	xmlBytes, err := GenerateReturn(results, strInputs, 2025)
	if err != nil {
		t.Fatalf("GenerateReturn failed: %v", err)
	}
	var parsed Return
	if err := xml.Unmarshal(xmlBytes, &parsed); err != nil {
		t.Fatalf("XML not well-formed: %v", err)
	}
	if parsed.ReturnData.DocumentCnt != 2 {
		t.Errorf("simple scenario DocumentCnt = %d, want 2", parsed.ReturnData.DocumentCnt)
	}

	// Self-employed scenario: 1040 + Sch1 + Sch2 + SchC + SchSE + 1 W-2 = 6
	results2, strInputs2 := selfEmployedScenario()
	xmlBytes2, err := GenerateReturn(results2, strInputs2, 2025)
	if err != nil {
		t.Fatalf("GenerateReturn failed: %v", err)
	}
	var parsed2 Return
	if err := xml.Unmarshal(xmlBytes2, &parsed2); err != nil {
		t.Fatalf("XML not well-formed: %v", err)
	}
	if parsed2.ReturnData.DocumentCnt != 6 {
		t.Errorf("self-employed scenario DocumentCnt = %d, want 6", parsed2.ReturnData.DocumentCnt)
	}
}

func TestGenerateReturn_MultipleW2s(t *testing.T) {
	results := map[string]float64{
		"w2:1:wages":                50000,
		"w2:1:federal_tax_withheld": 5000,
		"w2:2:wages":                30000,
		"w2:2:federal_tax_withheld": 3000,

		"1040:1a":  80000,
		"1040:1z":  80000,
		"1040:9":   80000,
		"1040:11":  80000,
		"1040:14":  15000,
		"1040:15":  65000,
		"1040:16":  9200,
		"1040:22":  9200,
		"1040:24":  9200,
		"1040:25a": 8000,
		"1040:25d": 8000,
		"1040:33":  8000,
		"1040:37":  1200,
	}
	strInputs := map[string]string{
		"1040:filing_status": "single",
		"1040:first_name":    "Multi",
		"1040:last_name":     "Worker",
		"1040:ssn":           "111-22-3333",
		"w2:1:employer_name": "First Corp",
		"w2:1:employer_ein":  "11-1111111",
		"w2:2:employer_name": "Second LLC",
		"w2:2:employer_ein":  "22-2222222",
	}

	xmlBytes, err := GenerateReturn(results, strInputs, 2025)
	if err != nil {
		t.Fatalf("GenerateReturn failed: %v", err)
	}
	var parsed Return
	if err := xml.Unmarshal(xmlBytes, &parsed); err != nil {
		t.Fatalf("XML not well-formed: %v", err)
	}

	if len(parsed.ReturnData.IRSW2) != 2 {
		t.Fatalf("expected 2 W-2s, got %d", len(parsed.ReturnData.IRSW2))
	}
	// Deterministic order: instance "1" before "2"
	if parsed.ReturnData.IRSW2[0].EmployerName != "First Corp" {
		t.Errorf("W-2[0] EmployerName = %q, want First Corp", parsed.ReturnData.IRSW2[0].EmployerName)
	}
	if parsed.ReturnData.IRSW2[1].EmployerName != "Second LLC" {
		t.Errorf("W-2[1] EmployerName = %q, want Second LLC", parsed.ReturnData.IRSW2[1].EmployerName)
	}
}

func TestGenerateReturn_TaxYear(t *testing.T) {
	results, strInputs := simpleW2Scenario()
	xmlBytes, err := GenerateReturn(results, strInputs, 2024)
	if err != nil {
		t.Fatalf("GenerateReturn failed: %v", err)
	}
	var parsed Return
	if err := xml.Unmarshal(xmlBytes, &parsed); err != nil {
		t.Fatalf("XML not well-formed: %v", err)
	}
	if parsed.ReturnHeader.TaxYr != 2024 {
		t.Errorf("TaxYr = %d, want 2024", parsed.ReturnHeader.TaxYr)
	}
	if parsed.ReturnHeader.TaxPeriodBeginDt != "2024-01-01" {
		t.Errorf("TaxPeriodBeginDt = %q, want 2024-01-01", parsed.ReturnHeader.TaxPeriodBeginDt)
	}
	if parsed.ReturnHeader.TaxPeriodEndDt != "2024-12-31" {
		t.Errorf("TaxPeriodEndDt = %q, want 2024-12-31", parsed.ReturnHeader.TaxPeriodEndDt)
	}
}

func TestSortedKeys(t *testing.T) {
	m := map[string]bool{"3": true, "1": true, "2": true, "10": true}
	got := sortedKeys(m)
	want := []string{"1", "10", "2", "3"} // lexicographic
	if len(got) != len(want) {
		t.Fatalf("sortedKeys len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("sortedKeys[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
