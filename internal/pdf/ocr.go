package pdf

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// OCRAvailable checks if tesseract is installed and accessible.
func OCRAvailable() bool {
	_, err := exec.LookPath("tesseract")
	return err == nil
}

// ParseFileOCR attempts to parse a tax return PDF using OCR.
// Falls back to this when the PDF has no AcroForm fields (scanned/printed).
// Returns a ParsedReturn with extracted values, or an error.
func (p *Parser) ParseFileOCR(path string) (*ParsedReturn, error) {
	if !OCRAvailable() {
		return nil, fmt.Errorf("tesseract is not installed; cannot OCR %s", path)
	}

	images, err := convertPDFToImages(path)
	if err != nil {
		return nil, fmt.Errorf("convert PDF to images: %w", err)
	}
	defer func() {
		for _, img := range images {
			os.Remove(img)
		}
	}()

	// Run tesseract on each page and collect all text.
	var allText strings.Builder
	for _, img := range images {
		text, err := runTesseract(img)
		if err != nil {
			return nil, fmt.Errorf("tesseract on %s: %w", img, err)
		}
		allText.WriteString(text)
		allText.WriteString("\n")
	}

	fullText := allText.String()
	formType := detectFormTypeFromText(fullText)

	numFields, strFields := extractFieldsFromText(fullText, formType)

	result := &ParsedReturn{
		FormID:    formType,
		TaxYear:   detectTaxYearFromText(fullText),
		Fields:    numFields,
		StrFields: strFields,
		RawFields: make(map[string]string),
	}

	return result, nil
}

// convertPDFToImages converts a PDF to PNG images using pdftoppm.
// Returns paths to temporary image files.
func convertPDFToImages(pdfPath string) ([]string, error) {
	// Check that pdftoppm is available.
	if _, err := exec.LookPath("pdftoppm"); err != nil {
		return nil, fmt.Errorf("pdftoppm not found: install poppler-utils")
	}

	tmpDir, err := os.MkdirTemp("", "taxpilot-ocr-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}

	outPrefix := filepath.Join(tmpDir, "page")

	cmd := exec.Command("pdftoppm", "-png", "-r", "300", pdfPath, outPrefix)
	if out, err := cmd.CombinedOutput(); err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("pdftoppm failed: %s: %w", string(out), err)
	}

	// Collect generated PNG files.
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("read temp dir: %w", err)
	}

	var paths []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".png") {
			paths = append(paths, filepath.Join(tmpDir, e.Name()))
		}
	}

	if len(paths) == 0 {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("pdftoppm produced no images")
	}

	return paths, nil
}

// runTesseract runs tesseract OCR on an image file and returns the text.
func runTesseract(imagePath string) (string, error) {
	cmd := exec.Command("tesseract", imagePath, "stdout", "--psm", "6")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("tesseract: %w", err)
	}
	return string(out), nil
}

// extractFieldsFromText parses OCR text to find tax form line values.
// Routes to the appropriate form-specific detector based on formType.
func extractFieldsFromText(text string, formType string) (map[string]float64, map[string]string) {
	switch formType {
	case "ca_540":
		return detect540Fields(text)
	case "1040":
		return detect1040Fields(text)
	default:
		// Try 1040 as default.
		return detect1040Fields(text)
	}
}

// detectFormTypeFromText determines the form type from OCR text content.
func detectFormTypeFromText(text string) string {
	lower := strings.ToLower(text)

	// Check for CA 540 indicators.
	if strings.Contains(lower, "california resident income tax") ||
		strings.Contains(lower, "form 540") ||
		strings.Contains(lower, "franchise tax board") {
		return "ca_540"
	}

	// Check for 1040 indicators.
	if strings.Contains(lower, "form 1040") ||
		strings.Contains(lower, "u.s. individual income tax return") ||
		strings.Contains(lower, "department of the treasury") {
		return "1040"
	}

	return "unknown"
}

// detect1040Fields extracts 1040-specific fields from OCR text.
func detect1040Fields(text string) (map[string]float64, map[string]string) {
	numFields := make(map[string]float64)
	strFields := make(map[string]string)

	// Currency pattern: optional $, digits with optional commas, optional decimal.
	currencyPat := `\$?\s*[\d,]+(?:\.\d{2})?`

	// Each pattern maps a regex to a field key.
	type fieldPattern struct {
		pattern  string
		fieldKey string
	}

	patterns := []fieldPattern{
		{`(?i)adjusted\s+gross\s+income\s*[.\s]*?(` + currencyPat + `)`, "1040:11"},
		{`(?i)total\s+income\s*[.\s]*?(` + currencyPat + `)`, "1040:9"},
		{`(?i)taxable\s+income\s*[.\s]*?(` + currencyPat + `)`, "1040:15"},
		{`(?i)total\s+tax\s*[.\s]*?(` + currencyPat + `)`, "1040:24"},
		{`(?i)federal\s+income\s+tax\s+withheld\s*[.\s]*?(` + currencyPat + `)`, "1040:25a"},
		{`(?i)total\s+payments\s*[.\s]*?(` + currencyPat + `)`, "1040:33"},
		{`(?i)(?:overpaid|refund(?:ed)?)\s*[.\s]*?(` + currencyPat + `)`, "1040:34"},
		{`(?i)amount\s+you\s+owe\s*[.\s]*?(` + currencyPat + `)`, "1040:37"},
	}

	// Tax line 16 needs special handling to avoid matching "total tax".
	taxLine16Pat := regexp.MustCompile(`(?i)(?:^|\n)\s*(?:16|16\s*[a-z]?)?\s*tax\s*[.\s]*?(` + currencyPat + `)`)
	// Only match if "total" is not immediately preceding.
	totalTaxPat := regexp.MustCompile(`(?i)total\s+tax`)

	for _, fp := range patterns {
		re := regexp.MustCompile(fp.pattern)
		m := re.FindStringSubmatch(text)
		if m != nil && len(m) > 1 {
			if v, err := ParseCurrency(m[1]); err == nil {
				numFields[fp.fieldKey] = v
			}
		}
	}

	// Line 16 (tax) — match only if not already captured as "total tax".
	if _, hasTotalTax := numFields["1040:24"]; true {
		_ = hasTotalTax
		m := taxLine16Pat.FindStringSubmatch(text)
		if m != nil && len(m) > 1 {
			// Verify this match is not part of "total tax".
			matchIdx := taxLine16Pat.FindStringIndex(text)
			if matchIdx != nil {
				prefix := ""
				if matchIdx[0] > 10 {
					prefix = text[matchIdx[0]-10 : matchIdx[0]]
				} else if matchIdx[0] > 0 {
					prefix = text[:matchIdx[0]]
				}
				if !totalTaxPat.MatchString(prefix) {
					if v, err := ParseCurrency(m[1]); err == nil {
						numFields["1040:16"] = v
					}
				}
			}
		}
	}

	// SSN pattern.
	ssnRe := regexp.MustCompile(`(\d{3})-(\d{2})-(\d{4})`)
	ssnMatch := ssnRe.FindStringSubmatch(text)
	if ssnMatch != nil {
		strFields["1040:ssn"] = ssnMatch[1] + ssnMatch[2] + ssnMatch[3]
	}

	// Name extraction: look for common patterns before SSN.
	nameRe := regexp.MustCompile(`(?i)(?:your\s+)?(?:first\s+name|name)\s+(?:and\s+(?:middle\s+)?initial)?\s*[:\s]*([A-Z][a-z]+)\s+([A-Z][a-z]+)`)
	nameMatch := nameRe.FindStringSubmatch(text)
	if nameMatch != nil && len(nameMatch) > 2 {
		strFields["1040:first_name"] = nameMatch[1]
		strFields["1040:last_name"] = nameMatch[2]
	}

	return numFields, strFields
}

// detect540Fields extracts CA 540-specific fields from OCR text.
func detect540Fields(text string) (map[string]float64, map[string]string) {
	numFields := make(map[string]float64)
	strFields := make(map[string]string)

	currencyPat := `\$?\s*[\d,]+(?:\.\d{2})?`

	type fieldPattern struct {
		pattern  string
		fieldKey string
	}

	patterns := []fieldPattern{
		{`(?i)california\s+adjusted\s+gross\s+income\s*[.\s]*?(` + currencyPat + `)`, "ca_540:17"},
		{`(?i)california\s+taxable\s+income\s*[.\s]*?(` + currencyPat + `)`, "ca_540:19"},
		{`(?i)total\s+tax\s*[.\s]*?(` + currencyPat + `)`, "ca_540:40"},
		{`(?i)california\s+income\s+tax\s+withheld\s*[.\s]*?(` + currencyPat + `)`, "ca_540:71"},
		{`(?i)(?:overpaid|refund(?:ed)?)\s*[.\s]*?(` + currencyPat + `)`, "ca_540:91"},
		{`(?i)amount\s+you\s+owe\s*[.\s]*?(` + currencyPat + `)`, "ca_540:93"},
	}

	for _, fp := range patterns {
		re := regexp.MustCompile(fp.pattern)
		m := re.FindStringSubmatch(text)
		if m != nil && len(m) > 1 {
			if v, err := ParseCurrency(m[1]); err == nil {
				numFields[fp.fieldKey] = v
			}
		}
	}

	// SSN.
	ssnRe := regexp.MustCompile(`(\d{3})-(\d{2})-(\d{4})`)
	ssnMatch := ssnRe.FindStringSubmatch(text)
	if ssnMatch != nil {
		strFields["ca_540:ssn"] = ssnMatch[1] + ssnMatch[2] + ssnMatch[3]
	}

	return numFields, strFields
}

// detectTaxYearFromText tries to extract a tax year from OCR text.
func detectTaxYearFromText(text string) int {
	// Look for year patterns like "2025" or "Tax Year 2025".
	re := regexp.MustCompile(`(?i)(?:tax\s+year|for\s+the\s+year)?\s*(20[2-9]\d)`)
	m := re.FindStringSubmatch(text)
	if m != nil && len(m) > 1 {
		// Use extractYear for validation.
		return extractYear(m[1])
	}
	return 0
}
