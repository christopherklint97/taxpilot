package pdf

import (
	"testing"
)

func TestOCRAvailable(t *testing.T) {
	// Just verify it doesn't panic. The result depends on the environment.
	result := OCRAvailable()
	t.Logf("OCRAvailable() = %v", result)
}

func TestDetectFormTypeFromText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "1040 from title",
			text:     "Form 1040 U.S. Individual Income Tax Return 2025",
			expected: "1040",
		},
		{
			name:     "1040 from department",
			text:     "Department of the Treasury\nInternal Revenue Service\nIncome Tax Return",
			expected: "1040",
		},
		{
			name:     "CA 540 from title",
			text:     "Form 540 California Resident Income Tax Return 2025",
			expected: "ca_540",
		},
		{
			name:     "CA 540 from FTB",
			text:     "Franchise Tax Board\nForm 540\nCalifornia Resident",
			expected: "ca_540",
		},
		{
			name:     "unknown form",
			text:     "Some random text with no form indicators",
			expected: "unknown",
		},
		{
			name:     "case insensitive 1040",
			text:     "FORM 1040 u.s. individual income tax return",
			expected: "1040",
		},
		{
			name:     "case insensitive CA 540",
			text:     "CALIFORNIA RESIDENT INCOME TAX RETURN",
			expected: "ca_540",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectFormTypeFromText(tt.text)
			if got != tt.expected {
				t.Errorf("detectFormTypeFromText() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestDetect1040Fields(t *testing.T) {
	sampleText := `Form 1040 U.S. Individual Income Tax Return 2025
Your first name and initial   John
Last name   Smith
Social security number   123-45-6789

1  Wages, salaries, tips ............................... $85,000
7  Capital gain or (loss) .............................  $2,500
9  Total income ....................................... $87,500
11 Adjusted gross income .............................. $82,000
15 Taxable income ..................................... $68,100
16 Tax ................................................ $10,294
24 Total tax .......................................... $10,294
25a Federal income tax withheld ........................ $12,000
33 Total payments ..................................... $12,000
34 Overpaid ........................................... $1,706
37 Amount you owe ..................................... $0
`

	numFields, strFields := detect1040Fields(sampleText)

	// Check numeric fields.
	numChecks := map[string]float64{
		"1040:9":   87500,
		"1040:11":  82000,
		"1040:15":  68100,
		"1040:24":  10294,
		"1040:25a": 12000,
		"1040:33":  12000,
		"1040:34":  1706,
		"1040:37":  0,
	}

	for key, expected := range numChecks {
		got, ok := numFields[key]
		if !ok {
			t.Errorf("missing numeric field %s", key)
			continue
		}
		if got != expected {
			t.Errorf("field %s = %v, want %v", key, got, expected)
		}
	}

	// Check SSN.
	if ssn, ok := strFields["1040:ssn"]; !ok {
		t.Error("missing 1040:ssn")
	} else if ssn != "123456789" {
		t.Errorf("1040:ssn = %q, want %q", ssn, "123456789")
	}

	// Check line 16 tax.
	if v, ok := numFields["1040:16"]; !ok {
		t.Error("missing 1040:16")
	} else if v != 10294 {
		t.Errorf("1040:16 = %v, want %v", v, 10294)
	}
}

func TestDetect1040Fields_NoData(t *testing.T) {
	numFields, strFields := detect1040Fields("nothing useful here")
	if len(numFields) != 0 {
		t.Errorf("expected no numeric fields, got %d", len(numFields))
	}
	if len(strFields) != 0 {
		t.Errorf("expected no string fields, got %d", len(strFields))
	}
}

func TestDetect540Fields(t *testing.T) {
	sampleText := `Form 540 California Resident Income Tax Return 2025
Franchise Tax Board

Social security number   987-65-4321

17 California adjusted gross income .................. $82,000
19 California taxable income .......................... $68,100
40 Total tax .......................................... $4,250
71 California income tax withheld ..................... $5,000
91 Overpaid ........................................... $750
93 Amount you owe ..................................... $0
`

	numFields, strFields := detect540Fields(sampleText)

	numChecks := map[string]float64{
		"ca_540:17": 82000,
		"ca_540:19": 68100,
		"ca_540:40": 4250,
		"ca_540:71": 5000,
		"ca_540:91": 750,
		"ca_540:93": 0,
	}

	for key, expected := range numChecks {
		got, ok := numFields[key]
		if !ok {
			t.Errorf("missing numeric field %s", key)
			continue
		}
		if got != expected {
			t.Errorf("field %s = %v, want %v", key, got, expected)
		}
	}

	// Check SSN.
	if ssn, ok := strFields["ca_540:ssn"]; !ok {
		t.Error("missing ca_540:ssn")
	} else if ssn != "987654321" {
		t.Errorf("ca_540:ssn = %q, want %q", ssn, "987654321")
	}
}

func TestDetect540Fields_NoData(t *testing.T) {
	numFields, strFields := detect540Fields("nothing useful here")
	if len(numFields) != 0 {
		t.Errorf("expected no numeric fields, got %d", len(numFields))
	}
	if len(strFields) != 0 {
		t.Errorf("expected no string fields, got %d", len(strFields))
	}
}

func TestExtractFieldsFromText(t *testing.T) {
	tests := []struct {
		name     string
		formType string
		text     string
		wantKey  string // a key we expect to find
	}{
		{
			name:     "routes to 1040",
			formType: "1040",
			text:     "Adjusted gross income .... $50,000",
			wantKey:  "1040:11",
		},
		{
			name:     "routes to ca_540",
			formType: "ca_540",
			text:     "California adjusted gross income .... $50,000",
			wantKey:  "ca_540:17",
		},
		{
			name:     "unknown defaults to 1040",
			formType: "unknown",
			text:     "Total income .... $75,000",
			wantKey:  "1040:9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			numFields, _ := extractFieldsFromText(tt.text, tt.formType)
			if _, ok := numFields[tt.wantKey]; !ok {
				t.Errorf("expected key %s in result, got keys: %v", tt.wantKey, numFields)
			}
		})
	}
}

func TestParseFileOCR_NoTesseract(t *testing.T) {
	// If tesseract is not installed, ParseFileOCR should return a clear error.
	if OCRAvailable() {
		t.Skip("tesseract is installed; skipping no-tesseract test")
	}

	p := NewParser()
	_, err := p.ParseFileOCR("/nonexistent/file.pdf")
	if err == nil {
		t.Fatal("expected error when tesseract is not installed")
	}
	if got := err.Error(); got != "tesseract is not installed; cannot OCR /nonexistent/file.pdf" {
		t.Errorf("unexpected error message: %s", got)
	}
}

func TestDetectTaxYearFromText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{"explicit tax year", "Tax Year 2025 Form 1040", 2025},
		{"for the year", "For the year 2024", 2024},
		{"standalone year", "2025", 2025},
		{"no year", "Form 1040 Income Tax", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectTaxYearFromText(tt.text)
			if got != tt.expected {
				t.Errorf("detectTaxYearFromText() = %d, want %d", got, tt.expected)
			}
		})
	}
}
