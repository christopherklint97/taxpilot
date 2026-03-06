package pdf

import (
	"os"
	"strings"
	"testing"
)

func TestFormatCurrency(t *testing.T) {
	tests := []struct {
		amount float64
		want   string
	}{
		{0, "$0.00"},
		{1234, "$1,234.00"},
		{75000, "$75,000.00"},
		{8114, "$8,114.00"},
		{1386, "$1,386.00"},
		{100, "$100.00"},
		{1000000, "$1,000,000.00"},
		{99.99, "$99.99"},
		{0.50, "$0.50"},
	}

	for _, tt := range tests {
		got := FormatCurrency(tt.amount)
		if got != tt.want {
			t.Errorf("FormatCurrency(%v) = %q, want %q", tt.amount, got, tt.want)
		}
	}
}

func TestFormatSSN(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"123456789", "123-45-6789"},
		{"123-45-6789", "123-45-6789"},
		{"12345", "12345"}, // invalid, returned as-is
		{"", ""},
	}

	for _, tt := range tests {
		got := FormatSSN(tt.input)
		if got != tt.want {
			t.Errorf("FormatSSN(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestTextExport1040(t *testing.T) {
	filler := NewFiller(t.TempDir())
	filler.RegisterForm(Federal1040Mappings())

	values := map[string]float64{
		"1040:1a":  75000,
		"1040:1z":  75000,
		"1040:9":   75000,
		"1040:10":  0,
		"1040:11":  75000,
		"1040:12":  15000,
		"1040:13":  0,
		"1040:14":  15000,
		"1040:15":  60000,
		"1040:16":  8114,
		"1040:24":  8114,
		"1040:25a": 9500,
		"1040:25d": 9500,
		"1040:33":  9500,
		"1040:34":  1386,
		"1040:37":  0,
	}
	strValues := map[string]string{
		"1040:filing_status": "single",
		"1040:first_name":    "Jane",
		"1040:last_name":     "Doe",
		"1040:ssn":           "123456789",
	}

	path, err := filler.FillFormText("1040", values, strValues)
	if err != nil {
		t.Fatalf("FillFormText returned error: %v", err)
	}

	// Verify file exists
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("output file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("output file is empty")
	}

	// Read and verify content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	content := string(data)

	// Check for expected content
	expectedStrings := []string{
		"Form 1040",
		"Tax Year 2025",
		"[X] Single",
		"Jane",
		"Doe",
		"123-45-6789",
		"$75,000.00",
		"$60,000.00",
		"$8,114.00",
		"$9,500.00",
		"$1,386.00",
		"$15,000.00",
		"INCOME",
		"DEDUCTIONS",
		"PAYMENTS",
		"REFUND",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(content, expected) {
			t.Errorf("output missing expected string %q", expected)
		}
	}

	// Verify the file ends with the separator
	if !strings.HasSuffix(strings.TrimSpace(content), "============================================================") {
		t.Error("output does not end with separator line")
	}
}

func TestTextExportCA540(t *testing.T) {
	filler := NewFiller(t.TempDir())
	filler.RegisterForm(CA540Mappings())

	values := map[string]float64{
		"ca_540:7":  75000,
		"ca_540:13": 75000,
		"ca_540:14": 0,
		"ca_540:15": 0,
		"ca_540:17": 75000,
		"ca_540:18": 5540,
		"ca_540:19": 69460,
		"ca_540:31": 3200,
		"ca_540:32": 144,
		"ca_540:35": 3056,
		"ca_540:36": 0,
		"ca_540:40": 3056,
		"ca_540:71": 4000,
		"ca_540:74": 4000,
		"ca_540:91": 944,
		"ca_540:93": 0,
	}
	strValues := map[string]string{
		"1040:filing_status":   "single",
		"ca_540:filing_status": "single",
	}

	path, err := filler.FillFormText("ca_540", values, strValues)
	if err != nil {
		t.Fatalf("FillFormText returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	content := string(data)

	expectedStrings := []string{
		"Form 540",
		"California",
		"Tax Year 2025",
		"SINGLE",
		"$75,000.00",
		"$69,460.00",
		"$3,056.00",
		"$4,000.00",
		"$944.00",
		"INCOME",
		"TAX",
		"PAYMENTS",
		"REFUND",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(content, expected) {
			t.Errorf("output missing expected string %q", expected)
		}
	}
}

func TestFillFormFallsBackToText(t *testing.T) {
	filler := NewFiller(t.TempDir())
	filler.RegisterForm(Federal1040Mappings())

	values := map[string]float64{
		"1040:1a": 50000,
		"1040:15": 35000,
	}
	strValues := map[string]string{
		"1040:filing_status": "single",
		"1040:first_name":    "Test",
	}

	// FillForm should fall back to text since no PDF template exists
	path, err := filler.FillForm("1040", values, strValues)
	if err != nil {
		t.Fatalf("FillForm returned error: %v", err)
	}

	// Should have generated a .txt file
	if !strings.HasSuffix(path, ".txt") {
		t.Errorf("expected .txt fallback, got path: %s", path)
	}

	_, err = os.Stat(path)
	if err != nil {
		t.Fatalf("output file not found: %v", err)
	}
}

func TestExportReturn(t *testing.T) {
	outputDir := t.TempDir()

	values := map[string]float64{
		"1040:1a":   75000,
		"1040:1z":   75000,
		"1040:9":    75000,
		"1040:11":   75000,
		"1040:12":   15000,
		"1040:14":   15000,
		"1040:15":   60000,
		"1040:16":   8114,
		"1040:24":   8114,
		"1040:25a":  9500,
		"1040:25d":  9500,
		"1040:33":   9500,
		"1040:34":   1386,
		"ca_540:7":  75000,
		"ca_540:13": 75000,
		"ca_540:17": 75000,
		"ca_540:19": 69460,
		"ca_540:31": 3200,
		"ca_540:35": 3056,
		"ca_540:40": 3056,
		"ca_540:71": 4000,
		"ca_540:74": 4000,
		"ca_540:91": 944,
	}
	strValues := map[string]string{
		"1040:filing_status":   "single",
		"1040:first_name":      "Jane",
		"1040:last_name":       "Doe",
		"1040:ssn":             "123456789",
		"ca_540:filing_status": "single",
	}

	paths, err := ExportReturn(outputDir, values, strValues, 2025)
	if err != nil {
		t.Fatalf("ExportReturn returned error: %v", err)
	}

	if len(paths) != 2 {
		t.Fatalf("expected 2 exported files, got %d", len(paths))
	}

	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			t.Errorf("exported file not found: %s", p)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("exported file is empty: %s", p)
		}
	}
}

func TestUnregisteredForm(t *testing.T) {
	filler := NewFiller(t.TempDir())

	_, err := filler.FillFormText("unknown_form", nil, nil)
	if err == nil {
		t.Fatal("expected error for unregistered form, got nil")
	}
}
