package pdf

import (
	"math"
	"testing"

	"taxpilot/internal/forms"
)

func TestParseCurrency(t *testing.T) {
	tests := []struct {
		input   string
		want    float64
		wantErr bool
	}{
		{"$75,000.00", 75000.00, false},
		{"$75,000", 75000, false},
		{"$1,234.56", 1234.56, false},
		{"75000", 75000, false},
		{"75000.50", 75000.50, false},
		{"($500)", -500, false},
		{"($1,234.56)", -1234.56, false},
		{"-$500", -500, false},
		{"-500", -500, false},
		{"-$1,234.56", -1234.56, false},
		{"$0", 0, false},
		{"0", 0, false},
		{"  $75,000.00  ", 75000.00, false},
		{"", 0, true},
		{"   ", 0, true},
		{"abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseCurrency(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseCurrency(%q) expected error, got %v", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseCurrency(%q) unexpected error: %v", tt.input, err)
				return
			}
			if math.Abs(got-tt.want) > 0.001 {
				t.Errorf("ParseCurrency(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseSSN(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"123-45-6789", "123456789"},
		{"123456789", "123456789"},
		{"123 45 6789", "123456789"},
		{"12-34-56789", "1234 56789"}, // not standard but strips dashes
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseSSN(tt.input)
			// Re-derive expected: strip dashes, strip spaces.
			expected := tt.input
			expected = replaceAll(expected, "-", "")
			expected = replaceAll(expected, " ", "")
			if got != expected {
				t.Errorf("ParseSSN(%q) = %q, want %q", tt.input, got, expected)
			}
		})
	}
}

// replaceAll is a test helper to avoid importing strings in test logic readability.
func replaceAll(s, old, new string) string {
	for {
		i := indexOf(s, old)
		if i < 0 {
			return s
		}
		s = s[:i] + new + s[i+len(old):]
	}
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestReverseMapping(t *testing.T) {
	config := &FormPDFConfig{
		FormID:   forms.FormF1040,
		FormName: "Form 1040",
		Mappings: []FieldMapping{
			{FieldKey: "1040:first_name", PDFField: "f1_02", Format: "string"},
			{FieldKey: "1040:ssn", PDFField: "f1_04", Format: "ssn"},
			{FieldKey: "1040:1a", PDFField: "f1_07", Format: "currency"},
			{FieldKey: "1040:filing_status", PDFField: "c1_1", Format: "checkbox"},
		},
	}

	rev := ReverseMapping(config)

	// Check that we can look up by PDF field name.
	tests := []struct {
		pdfField string
		wantKey  string
		wantFmt  string
	}{
		{"f1_02", "1040:first_name", "string"},
		{"f1_04", "1040:ssn", "ssn"},
		{"f1_07", "1040:1a", "currency"},
		{"c1_1", "1040:filing_status", "checkbox"},
	}

	for _, tt := range tests {
		m, ok := rev[tt.pdfField]
		if !ok {
			t.Errorf("ReverseMapping: missing key %q", tt.pdfField)
			continue
		}
		if m.FieldKey != tt.wantKey {
			t.Errorf("ReverseMapping[%q].FieldKey = %q, want %q", tt.pdfField, m.FieldKey, tt.wantKey)
		}
		if m.Format != tt.wantFmt {
			t.Errorf("ReverseMapping[%q].Format = %q, want %q", tt.pdfField, m.Format, tt.wantFmt)
		}
	}

	// Verify length matches.
	if len(rev) != len(config.Mappings) {
		t.Errorf("ReverseMapping length = %d, want %d", len(rev), len(config.Mappings))
	}
}

func TestExtractYear(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"Form 1040 - Tax Year 2025", 2025},
		{"f1040_2024.pdf", 2024},
		{"no year here", 0},
		{"2025", 2025},
		{"before 2025 after", 2025},
		{"12345", 0}, // 5 digits, not a standalone year
		{"year20251", 0}, // trailing digit
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := extractYear(tt.input)
			if got != tt.want {
				t.Errorf("extractYear(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestDetectFormFromFields(t *testing.T) {
	p := NewParser()
	p.RegisterForm(&FormPDFConfig{
		FormID:   forms.FormF1040,
		FormName: "Form 1040",
		Mappings: []FieldMapping{
			{FieldKey: "1040:1a", PDFField: "f1_07", Format: "currency"},
			{FieldKey: "1040:9", PDFField: "f1_22", Format: "currency"},
			{FieldKey: "1040:11", PDFField: "f1_24", Format: "currency"},
		},
	})
	p.RegisterForm(&FormPDFConfig{
		FormID:   forms.FormCA540,
		FormName: "Form 540",
		Mappings: []FieldMapping{
			{FieldKey: "ca_540:7", PDFField: "Line_7", Format: "currency"},
			{FieldKey: "ca_540:13", PDFField: "Line_13", Format: "currency"},
		},
	})

	// Simulate PDF fields matching 1040.
	pdfFields := map[string]string{
		"f1_07": "75000",
		"f1_22": "75000",
		"f1_24": "75000",
	}
	formID, config := p.detectFormFromFields(pdfFields, nil, "")
	if formID != "1040" {
		t.Errorf("detectFormFromFields: got formID=%q, want %q", formID, "1040")
	}
	if config == nil {
		t.Fatal("detectFormFromFields: config is nil")
	}

	// Simulate PDF fields matching ca_540.
	pdfFields = map[string]string{
		"Line_7":  "50000",
		"Line_13": "75000",
	}
	formID, config = p.detectFormFromFields(pdfFields, nil, "")
	if formID != "ca_540" {
		t.Errorf("detectFormFromFields: got formID=%q, want %q", formID, "ca_540")
	}

	// No matching fields.
	pdfFields = map[string]string{
		"unknown_field": "123",
	}
	formID, config = p.detectFormFromFields(pdfFields, nil, "")
	if formID != "" || config != nil {
		t.Errorf("detectFormFromFields: expected no match, got formID=%q", formID)
	}
}

func TestNewParser(t *testing.T) {
	p := NewParser()
	if p == nil {
		t.Fatal("NewParser() returned nil")
	}
	if p.configs == nil {
		t.Fatal("NewParser().configs is nil")
	}
}
