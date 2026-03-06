package pdf

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	pdfcpuapi "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/form"
)

// ParsedReturn holds field values extracted from a filled PDF return.
type ParsedReturn struct {
	FormID    string             // detected form (e.g., "1040", "ca_540")
	TaxYear   int                // detected or inferred tax year
	Fields    map[string]float64 // numeric field values keyed by internal field key
	StrFields map[string]string  // string field values
	RawFields map[string]string  // raw PDF field name -> value (for debugging)
}

// Parser extracts field values from filled tax return PDFs.
type Parser struct {
	configs map[string]*FormPDFConfig // reuse existing FormPDFConfig from mappings
}

// NewParser creates a new Parser.
func NewParser() *Parser {
	return &Parser{
		configs: make(map[string]*FormPDFConfig),
	}
}

// RegisterForm adds a form's field mappings for parsing.
// Reuses the same FormPDFConfig as the filler — the PDFField names
// map from AcroForm field names back to internal field keys.
func (p *Parser) RegisterForm(config *FormPDFConfig) {
	p.configs[config.FormID] = config
}

// ParseFile reads a filled PDF and extracts field values.
// It tries to match the PDF against registered form configs by checking
// which config's field names are present in the PDF.
func (p *Parser) ParseFile(path string) (*ParsedReturn, error) {
	// Extract form fields from the PDF using pdfcpu.
	fg, err := exportFormFields(path)
	if err != nil {
		return nil, fmt.Errorf("extract form fields from %s: %w", path, err)
	}

	// Collect all PDF field values into a flat map: PDF field ID -> string value.
	pdfFields := flattenFormGroup(fg)

	// Detect which form this PDF is.
	formID, config := p.detectFormFromFields(pdfFields, fg)
	if config == nil {
		return nil, fmt.Errorf("could not detect form type for %s", path)
	}

	// Build the reverse mapping: PDFField -> FieldMapping.
	revMap := ReverseMapping(config)

	result := &ParsedReturn{
		FormID:    formID,
		TaxYear:   detectTaxYear(fg),
		Fields:    make(map[string]float64),
		StrFields: make(map[string]string),
		RawFields: pdfFields,
	}

	// Map PDF field values to internal field keys.
	for pdfFieldID, rawValue := range pdfFields {
		mapping, ok := revMap[pdfFieldID]
		if !ok {
			continue
		}

		switch mapping.Format {
		case "currency", "integer":
			if rawValue == "" {
				continue
			}
			v, err := ParseCurrency(rawValue)
			if err != nil {
				// Store as string if we can't parse it.
				result.StrFields[mapping.FieldKey] = rawValue
				continue
			}
			result.Fields[mapping.FieldKey] = v

		case "ssn":
			result.StrFields[mapping.FieldKey] = ParseSSN(rawValue)

		case "checkbox":
			// Checkbox values come through the pdfFields map as "true"/"false".
			result.StrFields[mapping.FieldKey] = rawValue

		default: // "string", "ein", etc.
			result.StrFields[mapping.FieldKey] = rawValue
		}
	}

	return result, nil
}

// DetectForm examines the PDF's form fields and metadata to determine
// which tax form it is. Returns the form ID or an error.
func (p *Parser) DetectForm(path string) (string, error) {
	fg, err := exportFormFields(path)
	if err != nil {
		return "", fmt.Errorf("extract form fields from %s: %w", path, err)
	}

	pdfFields := flattenFormGroup(fg)
	formID, config := p.detectFormFromFields(pdfFields, fg)
	if config == nil {
		return "", fmt.Errorf("could not detect form type for %s", path)
	}
	return formID, nil
}

// exportFormFields opens a PDF and extracts its AcroForm fields via pdfcpu.
func exportFormFields(path string) (*form.FormGroup, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fg, err := pdfcpuapi.ExportForm(f, path, nil)
	if err != nil {
		return nil, err
	}
	return fg, nil
}

// flattenFormGroup extracts all field values from a FormGroup into a flat map.
// For text fields the key is the field ID and the value is the string value.
// For checkboxes the value is "true" or "false".
func flattenFormGroup(fg *form.FormGroup) map[string]string {
	fields := make(map[string]string)
	if fg == nil {
		return fields
	}

	for _, f := range fg.Forms {
		for _, tf := range f.TextFields {
			fields[tf.ID] = tf.Value
		}
		for _, df := range f.DateFields {
			fields[df.ID] = df.Value
		}
		for _, cb := range f.CheckBoxes {
			if cb.Value {
				fields[cb.ID] = "true"
			} else {
				fields[cb.ID] = "false"
			}
		}
		for _, rbg := range f.RadioButtonGroups {
			fields[rbg.ID] = rbg.Value
		}
		for _, combo := range f.ComboBoxes {
			fields[combo.ID] = combo.Value
		}
		for _, lb := range f.ListBoxes {
			if len(lb.Values) > 0 {
				fields[lb.ID] = strings.Join(lb.Values, ",")
			}
		}
	}

	return fields
}

// detectFormFromFields determines which registered form config best matches
// the PDF's field names. Returns the form ID and config, or ("", nil) if
// no match is found.
func (p *Parser) detectFormFromFields(pdfFields map[string]string, fg *form.FormGroup) (string, *FormPDFConfig) {
	// First, try metadata-based detection from the source/title.
	if fg != nil {
		source := strings.ToLower(fg.Header.Source)
		title := strings.ToLower(fg.Header.Title)
		combined := source + " " + title
		for formID, config := range p.configs {
			switch formID {
			case "1040":
				if strings.Contains(combined, "1040") && !strings.Contains(combined, "540") {
					return formID, config
				}
			case "ca_540":
				if strings.Contains(combined, "540") {
					return formID, config
				}
			}
		}
	}

	// Fall back to field-name matching: pick the config with the most matching fields.
	var bestID string
	var bestConfig *FormPDFConfig
	bestScore := 0

	for formID, config := range p.configs {
		score := 0
		for _, m := range config.Mappings {
			if _, ok := pdfFields[m.PDFField]; ok {
				score++
			}
		}
		if score > bestScore {
			bestScore = score
			bestID = formID
			bestConfig = config
		}
	}

	if bestScore == 0 {
		return "", nil
	}
	return bestID, bestConfig
}

// detectTaxYear tries to infer the tax year from the FormGroup header metadata.
func detectTaxYear(fg *form.FormGroup) int {
	if fg == nil {
		return 0
	}

	// Check source, title, subject for a 4-digit year.
	for _, s := range []string{fg.Header.Source, fg.Header.Title, fg.Header.Subject} {
		if y := extractYear(s); y > 0 {
			return y
		}
	}
	return 0
}

// extractYear finds the first plausible 4-digit tax year (2000-2099) in a string.
func extractYear(s string) int {
	for i := 0; i <= len(s)-4; i++ {
		chunk := s[i : i+4]
		if y, err := strconv.Atoi(chunk); err == nil && y >= 2000 && y <= 2099 {
			// Make sure it's not part of a longer number.
			if i > 0 && s[i-1] >= '0' && s[i-1] <= '9' {
				continue
			}
			if i+4 < len(s) && s[i+4] >= '0' && s[i+4] <= '9' {
				continue
			}
			return y
		}
	}
	return 0
}

// ParseCurrency converts a PDF currency string to float64.
// Handles: "$75,000", "$75,000.00", "75000", "(500)" -> -500, "-$500"
func ParseCurrency(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty string")
	}

	// Detect negative: parentheses or leading minus.
	negative := false
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		negative = true
		s = s[1 : len(s)-1]
	} else if strings.HasPrefix(s, "-") {
		negative = true
		s = s[1:]
	}

	// Strip dollar signs, commas, spaces.
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimSpace(s)

	if s == "" {
		return 0, fmt.Errorf("no numeric content")
	}

	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("parse %q: %w", s, err)
	}

	if negative {
		v = -v
	}
	return v, nil
}

// ParseSSN strips formatting from an SSN string.
// "123-45-6789" -> "123456789"
func ParseSSN(s string) string {
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, " ", "")
	return s
}

// ReverseMapping creates a PDFField -> FieldMapping lookup from a FormPDFConfig.
func ReverseMapping(config *FormPDFConfig) map[string]FieldMapping {
	rev := make(map[string]FieldMapping, len(config.Mappings))
	for _, m := range config.Mappings {
		rev[m.PDFField] = m
	}
	return rev
}
