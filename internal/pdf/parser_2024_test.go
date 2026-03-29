package pdf

import (
	"math"
	"os"
	"testing"

	"taxpilot/internal/forms"
)

// skipIfNoFile skips a test if the given file does not exist.
func skipIfNoFile(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skipf("test PDF not found: %s", path)
	}
}

func TestParse2024Form1040(t *testing.T) {
	path := "../../returns/f1040.pdf"
	skipIfNoFile(t, path)

	parser := NewParser()
	registerAllParseForms(parser)

	result, err := parser.ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile(%s): %v", path, err)
	}

	if result.FormID != forms.FormF1040 {
		t.Errorf("FormID = %q, want %q", result.FormID, forms.FormF1040)
	}
	if result.TaxYear != 2024 {
		t.Errorf("TaxYear = %d, want 2024", result.TaxYear)
	}

	// Verify ALL mapped numeric fields.
	numericTests := []struct {
		key  string
		want float64
	}{
		{"1040:1a", 85000},
		{"1040:1z", 85000},
		{"1040:8", -85000},
		{"1040:9", 0},
		// 1040:10 (adjustments) and 1040:13 (QBI) are not present in this
		// PDF because the filer has zero for those lines — the 2024 IRS PDF
		// omits AcroForm fields for blank lines.
		{"1040:11", 0},
		{"1040:12", 14600},
		{"1040:14", 14600},
		{"1040:15", 0},
		{"1040:16", 0},
		{"1040:24", 0},
		{"1040:25a", 0},
		{"1040:25b", 0},
		{"1040:25d", 0},
		{"1040:33", 0},
		{"1040:34", 0},
		{"1040:37", 0},
	}

	for _, tt := range numericTests {
		got, ok := result.Fields[tt.key]
		if !ok {
			t.Errorf("Fields[%q] not found", tt.key)
			continue
		}
		if math.Abs(got-tt.want) > 0.01 {
			t.Errorf("Fields[%q] = %v, want %v", tt.key, got, tt.want)
		}
	}

	// Verify ALL mapped string fields.
	strTests := []struct {
		key  string
		want string
	}{
		{"1040:first_name", "Christopher K"},
		{"1040:last_name", "Klint"},
		{"1040:ssn", "603986677"},
		{"1040:address", "1502 E Carson St SPC"},
		{"1040:apt", "110"},
		{"1040:city", "Carson"},
		{"1040:state", "California"},
		{"1040:zip", "90745-2326"},
		{"1040:spouse_name", "Sofie Matilde Ovesen"},
		{"1040:occupation", "Director of Technology & Innovation"},
	}

	for _, tt := range strTests {
		got, ok := result.StrFields[tt.key]
		if !ok {
			t.Errorf("StrFields[%q] not found", tt.key)
			continue
		}
		if got != tt.want {
			t.Errorf("StrFields[%q] = %q, want %q", tt.key, got, tt.want)
		}
	}

	// Verify filing status checkboxes — QSS (field 37) should be true.
	if v, ok := result.StrFields["1040:filing_status_qss"]; !ok {
		t.Error("StrFields[1040:filing_status_qss] not found")
	} else if v != "true" {
		t.Errorf("StrFields[1040:filing_status_qss] = %q, want %q", v, "true")
	}

	// Other filing status checkboxes should be false.
	for _, key := range []string{"1040:filing_status_single", "1040:filing_status_mfj", "1040:filing_status_mfs", "1040:filing_status_hoh"} {
		if v, ok := result.StrFields[key]; ok && v == "true" {
			t.Errorf("StrFields[%q] = %q, want %q", key, v, "false")
		}
	}

	// Verify post-processed filing_status enum.
	if v := result.StrFields["1040:filing_status"]; v != "qualifying_surviving_spouse" {
		t.Errorf("StrFields[1040:filing_status] = %q, want %q", v, "qualifying_surviving_spouse")
	}

	// Verify completeness: every mapped field should appear in Fields or StrFields.
	config := Federal1040Mappings2024()
	for _, m := range config.Mappings {
		_, inNum := result.Fields[m.FieldKey]
		_, inStr := result.StrFields[m.FieldKey]
		if !inNum && !inStr {
			// Only flag if the PDF field actually has a value.
			if raw, ok := result.RawFields[m.PDFField]; ok && raw != "" {
				t.Errorf("mapped field %q (PDF %q) has raw value %q but was not extracted", m.FieldKey, m.PDFField, raw)
			}
		}
	}
}

func TestParse2024Schedule1(t *testing.T) {
	path := "../../returns/f1040s1.pdf"
	skipIfNoFile(t, path)

	parser := NewParser()
	registerAllParseForms(parser)

	result, err := parser.ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile(%s): %v", path, err)
	}

	if result.FormID != forms.FormSchedule1 {
		t.Errorf("FormID = %q, want %q", result.FormID, forms.FormSchedule1)
	}

	// Verify ALL mapped numeric fields.
	numericTests := []struct {
		key  string
		want float64
	}{
		{"schedule_1:1", 0},
		{"schedule_1:8z", 85000},
		{"schedule_1:8d", -85000},
		{"schedule_1:10", -85000},
		{"schedule_1:25", 0},
		{"schedule_1:26", 0},
	}

	for _, tt := range numericTests {
		got, ok := result.Fields[tt.key]
		if !ok {
			t.Errorf("Fields[%q] not found", tt.key)
			continue
		}
		if math.Abs(got-tt.want) > 0.01 {
			t.Errorf("Fields[%q] = %v, want %v", tt.key, got, tt.want)
		}
	}

	// Verify string fields.
	if v := result.StrFields["schedule_1:name"]; v != "Christopher K Klint" {
		t.Errorf("StrFields[schedule_1:name] = %q, want %q", v, "Christopher K Klint")
	}
}

func TestParse2024ScheduleB(t *testing.T) {
	path := "../../returns/f1040sb.pdf"
	skipIfNoFile(t, path)

	parser := NewParser()
	registerAllParseForms(parser)

	result, err := parser.ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile(%s): %v", path, err)
	}

	if result.FormID != forms.FormScheduleB {
		t.Errorf("FormID = %q, want %q", result.FormID, forms.FormScheduleB)
	}

	// Verify all checkboxes.
	for _, key := range []string{"schedule_b:7a_yes", "schedule_b:8_yes", "schedule_b:7b_yes"} {
		if v, ok := result.StrFields[key]; !ok || v != "true" {
			t.Errorf("StrFields[%q] = %q, want %q", key, v, "true")
		}
	}

	// Verify post-processed enum fields.
	if v := result.StrFields["schedule_b:7a"]; v != "yes" {
		t.Errorf("StrFields[schedule_b:7a] = %q, want %q", v, "yes")
	}
	if v := result.StrFields["schedule_b:8"]; v != "yes" {
		t.Errorf("StrFields[schedule_b:8] = %q, want %q", v, "yes")
	}
}

func TestParse2024Form2555(t *testing.T) {
	path := "../../returns/f2555.pdf"
	skipIfNoFile(t, path)

	parser := NewParser()
	registerAllParseForms(parser)

	result, err := parser.ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile(%s): %v", path, err)
	}

	if result.FormID != forms.FormF2555 {
		t.Errorf("FormID = %q, want %q", result.FormID, forms.FormF2555)
	}

	// Verify ALL mapped numeric fields.
	numericTests := []struct {
		key  string
		want float64
	}{
		{"form_2555:foreign_earned_income", 85000},
		{"form_2555:27", 85000},
		{"form_2555:exclusion_limit", 126500},
		{"form_2555:ppt_days_present", 365},
		{"form_2555:prorated_exclusion", 126500},
		{"form_2555:25", 0},
		{"form_2555:27a", 0},
		{"form_2555:28", 85000},
		{"form_2555:28a", 85000},
		{"form_2555:foreign_income_exclusion", 41500},
		{"form_2555:housing_exclusion", 41500},
		{"form_2555:housing_deduction", 41500},
		{"form_2555:total_exclusion", 41500},
	}

	for _, tt := range numericTests {
		got, ok := result.Fields[tt.key]
		if !ok {
			t.Errorf("Fields[%q] not found", tt.key)
			continue
		}
		if math.Abs(got-tt.want) > 0.01 {
			t.Errorf("Fields[%q] = %v, want %v", tt.key, got, tt.want)
		}
	}

	// Verify ALL mapped string fields.
	strTests := []struct {
		key  string
		want string
	}{
		{"form_2555:foreign_country", "Sweden"},
		{"form_2555:name", "Christopher K Klint"},
		{"form_2555:ssn", "603986677"},
		{"form_2555:foreign_address", "Allévägen 10E, 19276 Sollentuna, Sweden"},
		{"form_2555:occupation", "Director of Technology & Innovation"},
		{"form_2555:employer_name_2555", "Nenda AB"},
		{"form_2555:employer_ein", "None"},
		{"form_2555:employer_address", "Stora Nygatan 26, 11127 Stockholm, Sweden"},
		{"form_2555:citizenship_country", "United States"},
		{"form_2555:countries_dates", "Sweden 08/21/2015"},
		{"form_2555:prior_year_2555", "2023"},
		{"form_2555:ppt_period_start", "01/01/2024"},
		{"form_2555:ppt_period_end", "12/31/2024"},
		{"form_2555:ppt_reason", "Physically present in a foreign country or countries"},
		{"form_2555:ppt_duration", "for the entire 12-month period"},
	}

	for _, tt := range strTests {
		got, ok := result.StrFields[tt.key]
		if !ok {
			t.Errorf("StrFields[%q] not found", tt.key)
			continue
		}
		if got != tt.want {
			t.Errorf("StrFields[%q] = %q, want %q", tt.key, got, tt.want)
		}
	}

	// Verify post-processed enum fields.
	if v := result.StrFields["form_2555:qualifying_test"]; v != "physical_presence" {
		t.Errorf("StrFields[form_2555:qualifying_test] = %q, want %q", v, "physical_presence")
	}
	if v := result.StrFields["form_2555:employer_foreign"]; v != "yes" {
		t.Errorf("StrFields[form_2555:employer_foreign] = %q, want %q", v, "yes")
	}

	// Verify checkboxes are present.
	for _, key := range []string{
		"form_2555:qualifying_test_ppt",
		"form_2555:employer_foreign_yes",
		"form_2555:claimed_prior_year",
	} {
		if v, ok := result.StrFields[key]; !ok || v != "true" {
			t.Errorf("StrFields[%q] = %q, want %q", key, v, "true")
		}
	}
}

func TestParse2024CA540NR(t *testing.T) {
	path := "../../returns/2024-540nr.pdf"
	skipIfNoFile(t, path)

	parser := NewParser()
	registerAllParseForms(parser)

	result, err := parser.ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile(%s): %v", path, err)
	}

	if result.FormID != forms.FormCA540NR {
		t.Errorf("FormID = %q, want %q", result.FormID, forms.FormCA540NR)
	}

	// Verify ALL mapped numeric fields.
	numericTests := []struct {
		key  string
		want float64
	}{
		{"ca_540nr:total_income", 85000},
		{"ca_540nr:agi", 85000},
		{"ca_540nr:deduction", 5540},
		{"ca_540nr:taxable_income", 79460},
		{"ca_540nr:tax", 4447},
		{"ca_540nr:exemption_credit", 144},
		{"ca_540nr:exemptions", 1},
		{"ca_540nr:subtotal_credits", 140},
		{"ca_540nr:credits", 0},
		{"ca_540nr:use_tax", 0},
		{"ca_540nr:total_payments", 0},
	}

	for _, tt := range numericTests {
		got, ok := result.Fields[tt.key]
		if !ok {
			t.Errorf("Fields[%q] not found", tt.key)
			continue
		}
		if math.Abs(got-tt.want) > 0.01 {
			t.Errorf("Fields[%q] = %v, want %v", tt.key, got, tt.want)
		}
	}

	// Verify ALL mapped string fields.
	strTests := []struct {
		key  string
		want string
	}{
		{"ca_540nr:first_name", "Christopher"},
		{"ca_540nr:middle_initial", "K"},
		{"ca_540nr:last_name", "Klint"},
		{"ca_540nr:city", "Carson"},
		{"ca_540nr:state", "CA"},
		{"ca_540nr:zip", "90745-2326"},
		{"ca_540nr:address", "1502 E Carson St SPC"},
		{"ca_540nr:apt", "110"},
		{"ca_540nr:spouse", "Sofie M Ovesen (Sweden, not USA)"},
		{"ca_540nr:email", "christopher.klint@gmail.com"},
		{"ca_540nr:phone", "3155951458"},
		{"ca_540nr:county_code", "056"},
		{"ca_540nr:name_signature", "Christopher K Klint"},
	}

	for _, tt := range strTests {
		got, ok := result.StrFields[tt.key]
		if !ok {
			t.Errorf("StrFields[%q] not found", tt.key)
			continue
		}
		if got != tt.want {
			t.Errorf("StrFields[%q] = %q, want %q", tt.key, got, tt.want)
		}
	}

	// SSN should be stripped of dashes.
	if ssn := result.StrFields["ca_540nr:ssn"]; ssn != "603986677" {
		t.Errorf("StrFields[ca_540nr:ssn] = %q, want %q", ssn, "603986677")
	}
	if ssn := result.StrFields["ca_540nr:ssn_confirm"]; ssn != "603986677" {
		t.Errorf("StrFields[ca_540nr:ssn_confirm] = %q, want %q", ssn, "603986677")
	}
}

func TestParse2024ScheduleCA(t *testing.T) {
	path := "../../returns/2024-540-ca.pdf"
	skipIfNoFile(t, path)

	parser := NewParser()
	registerAllParseForms(parser)

	result, err := parser.ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile(%s): %v", path, err)
	}

	if result.FormID != forms.FormScheduleCA {
		t.Errorf("FormID = %q, want %q", result.FormID, forms.FormScheduleCA)
	}

	// Verify ALL mapped numeric fields.
	numericTests := []struct {
		key  string
		want float64
	}{
		{"ca_schedule_ca:1a_col_a", 85000},
		{"ca_schedule_ca:1a_col_b", 85000},
		{"ca_schedule_ca:1a_col_c", 85000},
		{"ca_schedule_ca:8d_col_a", -85000},
		{"ca_schedule_ca:8d_col_c", 85000},
		{"ca_schedule_ca:10_col_c", 85000},
		{"ca_schedule_ca:37_col_a", 85000},
		{"ca_schedule_ca:ca_deduction", 5540},
	}

	for _, tt := range numericTests {
		got, ok := result.Fields[tt.key]
		if !ok {
			t.Errorf("Fields[%q] not found", tt.key)
			continue
		}
		if math.Abs(got-tt.want) > 0.01 {
			t.Errorf("Fields[%q] = %v, want %v", tt.key, got, tt.want)
		}
	}

	// Verify string fields.
	if v := result.StrFields["ca_schedule_ca:name"]; v != "Christopher K Klint" {
		t.Errorf("StrFields[ca_schedule_ca:name] = %q, want %q", v, "Christopher K Klint")
	}
}

func TestParse2024Form3853(t *testing.T) {
	path := "../../returns/2024-3853.pdf"
	skipIfNoFile(t, path)

	parser := NewParser()
	registerAllParseForms(parser)

	result, err := parser.ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile(%s): %v", path, err)
	}

	if result.FormID != forms.FormF3853 {
		t.Errorf("FormID = %q, want %q", result.FormID, forms.FormF3853)
	}

	// Verify numeric fields.
	numericTests := []struct {
		key  string
		want float64
	}{
		{"form_3853:income", 85000},
		{"form_3853:penalty", 0},
	}

	for _, tt := range numericTests {
		got, ok := result.Fields[tt.key]
		if !ok {
			t.Errorf("Fields[%q] not found", tt.key)
			continue
		}
		if math.Abs(got-tt.want) > 0.01 {
			t.Errorf("Fields[%q] = %v, want %v", tt.key, got, tt.want)
		}
	}

	// Verify ALL mapped string fields.
	strTests := []struct {
		key  string
		want string
	}{
		{"form_3853:name", "Christopher K Klint"},
		{"form_3853:ssn", "603986677"},
		{"form_3853:first_name", "Christopher"},
		{"form_3853:middle_initial", "K"},
		{"form_3853:last_name", "Klint"},
		{"form_3853:individual_ssn", "603986677"},
		{"form_3853:dob", "05/28/1997"},
		{"form_3853:covered_1_first", "Christopher"},
		{"form_3853:covered_1_middle", "K"},
		{"form_3853:covered_1_last", "Klint"},
		{"form_3853:covered_1_suffix", "E"},
	}

	for _, tt := range strTests {
		got, ok := result.StrFields[tt.key]
		if !ok {
			t.Errorf("StrFields[%q] not found", tt.key)
			continue
		}
		if got != tt.want {
			t.Errorf("StrFields[%q] = %q, want %q", tt.key, got, tt.want)
		}
	}
}

// TestMultiConfigSelection verifies that the parser picks the correct config
// variant (2024 vs 2025) based on which field names match.
func TestMultiConfigSelection(t *testing.T) {
	parser := NewParser()
	registerAllParseForms(parser)

	// Simulate 2024-style PDF fields (numeric IDs).
	pdfFields2024 := map[string]string{
		"19":  "Christopher K",
		"20":  "Klint",
		"73":  "85,000",
		"82":  "85,000",
		"97":  "0",
		"99":  "0",
		"100": "14,600",
	}

	formID, config := parser.detectFormFromFields(pdfFields2024, nil, "f1040.pdf")
	if formID != forms.FormF1040 {
		t.Errorf("2024 fields: formID = %q, want %q", formID, forms.FormF1040)
	}
	if config == nil {
		t.Fatal("2024 fields: config is nil")
	}
	// The selected config should have numeric field IDs (2024 style).
	found := false
	for _, m := range config.Mappings {
		if m.PDFField == "73" {
			found = true
			break
		}
	}
	if !found {
		t.Error("2024 fields: expected config with numeric field '73', got config with XFA-style fields")
	}

	// Simulate 2025-style PDF fields (dot-separated numeric XFA IDs).
	pdfFields2025 := map[string]string{
		"678.677.840": "John",
		"678.677.841": "Doe",
		"678.677.867": "50000",
		"678.677.895": "50000",
	}

	formID, config = parser.detectFormFromFields(pdfFields2025, nil, "f1040.pdf")
	if formID != forms.FormF1040 {
		t.Errorf("2025 fields: formID = %q, want %q", formID, forms.FormF1040)
	}
	if config == nil {
		t.Fatal("2025 fields: config is nil")
	}
	// The selected config should have 2025 dot-separated field IDs.
	found = false
	for _, m := range config.Mappings {
		if m.PDFField == "678.677.867" {
			found = true
			break
		}
	}
	if !found {
		t.Error("2025 fields: expected config with dot-separated field IDs, got config with numeric fields")
	}
}
