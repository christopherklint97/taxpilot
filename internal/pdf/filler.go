package pdf

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"taxpilot/internal/forms"

	pdfcpuapi "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/form"
)

// FieldMapping maps internal field keys to PDF AcroForm field names.
type FieldMapping struct {
	FieldKey string // internal key like "1040:1a"
	PDFField string // AcroForm field name in the PDF
	Format   string // "currency", "string", "integer", "ssn", "ein", "checkbox"
}

// FormPDFConfig holds the PDF template path and field mappings for a form.
type FormPDFConfig struct {
	FormID       forms.FormID
	FormName     string
	TemplatePath string // path to blank PDF template
	Mappings     []FieldMapping
}

// Filler fills PDF forms with computed values.
type Filler struct {
	configs   map[forms.FormID]*FormPDFConfig
	outputDir string
}

// NewFiller creates a new Filler that writes output to the given directory.
func NewFiller(outputDir string) *Filler {
	return &Filler{
		configs:   make(map[forms.FormID]*FormPDFConfig),
		outputDir: outputDir,
	}
}

// RegisterForm registers a form's PDF configuration with the filler.
func (f *Filler) RegisterForm(config *FormPDFConfig) {
	f.configs[config.FormID] = config
}

// FillForm fills a single PDF form with values using pdfcpu.
// Returns the output file path.
func (f *Filler) FillForm(formID forms.FormID, values map[string]float64, strValues map[string]string) (string, error) {
	config, ok := f.configs[formID]
	if !ok {
		return "", fmt.Errorf("no PDF config registered for form %q", formID)
	}

	// Check if the PDF template exists
	if _, err := os.Stat(config.TemplatePath); os.IsNotExist(err) {
		// Fall back to text export
		return f.FillFormText(formID, values, strValues)
	}

	if err := os.MkdirAll(f.outputDir, 0o755); err != nil {
		return "", fmt.Errorf("create output dir: %w", err)
	}

	outPath := filepath.Join(f.outputDir, fmt.Sprintf("%s_filled.pdf", formID))

	// Build the pdfcpu FormGroup JSON
	var textFields []*form.TextField
	var checkBoxes []*form.CheckBox
	for _, m := range config.Mappings {
		val := formatFieldValue(m, values, strValues)
		if val == "" {
			continue
		}
		if m.Format == "checkbox" {
			checkBoxes = append(checkBoxes, &form.CheckBox{
				ID:    m.PDFField,
				Value: val != "" && val != "false" && val != "0",
			})
		} else {
			textFields = append(textFields, &form.TextField{
				ID:    m.PDFField,
				Value: val,
			})
		}
	}

	fg := form.FormGroup{
		Header: form.Header{
			Source:   config.TemplatePath,
			Version:  "pdfcpu",
			Creation: time.Now().Format("2006-01-02 15:04:05 MST"),
		},
		Forms: []form.Form{
			{
				TextFields: textFields,
				CheckBoxes: checkBoxes,
			},
		},
	}

	// Write the JSON data to a temp file
	jsonData, err := json.MarshalIndent(fg, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal form data: %w", err)
	}

	jsonPath := filepath.Join(f.outputDir, fmt.Sprintf("%s_data.json", formID))
	if err := os.WriteFile(jsonPath, jsonData, 0o644); err != nil {
		return "", fmt.Errorf("write form data JSON: %w", err)
	}
	defer os.Remove(jsonPath)

	// Use pdfcpu to fill the form fields
	if err := pdfcpuapi.FillFormFile(config.TemplatePath, jsonPath, outPath, nil); err != nil {
		return "", fmt.Errorf("fill PDF form %s: %w", formID, err)
	}

	return outPath, nil
}

// FillFormText generates a text representation of the filled form (fallback when no PDF template).
func (f *Filler) FillFormText(formID forms.FormID, values map[string]float64, strValues map[string]string) (string, error) {
	config, ok := f.configs[formID]
	if !ok {
		return "", fmt.Errorf("no PDF config registered for form %q", formID)
	}

	if err := os.MkdirAll(f.outputDir, 0o755); err != nil {
		return "", fmt.Errorf("create output dir: %w", err)
	}

	outPath := filepath.Join(f.outputDir, fmt.Sprintf("%s_filled.txt", formID))

	var text string
	switch formID {
	case forms.FormF1040:
		text = render1040Text(config, values, strValues)
	case forms.FormCA540:
		text = renderCA540Text(config, values, strValues)
	case forms.FormScheduleB:
		text = renderScheduleBText(config, values, strValues)
	case forms.FormScheduleD:
		text = renderScheduleDText(config, values, strValues)
	case forms.FormSchedule1:
		text = renderSchedule1Text(config, values, strValues)
	default:
		text = renderGenericText(config, values, strValues)
	}

	if err := os.WriteFile(outPath, []byte(text), 0o644); err != nil {
		return "", fmt.Errorf("write text export: %w", err)
	}

	return outPath, nil
}

// FillAll fills all registered forms.
// Returns list of output file paths.
func (f *Filler) FillAll(values map[string]float64, strValues map[string]string) ([]string, error) {
	var paths []string
	for formID := range f.configs {
		path, err := f.FillForm(formID, values, strValues)
		if err != nil {
			return paths, fmt.Errorf("fill %s: %w", formID, err)
		}
		paths = append(paths, path)
	}
	return paths, nil
}

// FormatCurrency formats a float as "$1,234.00" for display.
func FormatCurrency(amount float64) string {
	negative := amount < 0
	if negative {
		amount = -amount
	}

	whole := int64(amount)
	cents := int64((amount - float64(whole)) * 100 + 0.5)
	if cents >= 100 {
		whole++
		cents -= 100
	}

	// Format with commas
	s := fmt.Sprintf("%d", whole)
	if len(s) > 3 {
		var parts []string
		for len(s) > 3 {
			parts = append([]string{s[len(s)-3:]}, parts...)
			s = s[:len(s)-3]
		}
		parts = append([]string{s}, parts...)
		s = strings.Join(parts, ",")
	}

	result := fmt.Sprintf("$%s.%02d", s, cents)
	if negative {
		result = "-" + result
	}
	return result
}

// FormatSSN formats "123456789" as "123-45-6789".
func FormatSSN(ssn string) string {
	// Strip any existing dashes or spaces
	clean := strings.ReplaceAll(ssn, "-", "")
	clean = strings.ReplaceAll(clean, " ", "")
	if len(clean) != 9 {
		return ssn // return as-is if not a valid 9-digit SSN
	}
	return fmt.Sprintf("%s-%s-%s", clean[0:3], clean[3:5], clean[5:9])
}

// formatFieldValue formats a field value based on its mapping format.
func formatFieldValue(m FieldMapping, values map[string]float64, strValues map[string]string) string {
	switch m.Format {
	case "currency":
		if v, ok := values[m.FieldKey]; ok {
			return fmt.Sprintf("%.0f", v)
		}
	case "integer":
		if v, ok := values[m.FieldKey]; ok {
			return fmt.Sprintf("%.0f", v)
		}
	case "string":
		if v, ok := strValues[m.FieldKey]; ok {
			return v
		}
	case "ssn":
		if v, ok := strValues[m.FieldKey]; ok {
			return FormatSSN(v)
		}
	case "ein":
		if v, ok := strValues[m.FieldKey]; ok {
			return v
		}
	case "checkbox":
		if v, ok := strValues[m.FieldKey]; ok && v != "" {
			return v
		}
	}
	return ""
}

// getVal is a helper to get a float64 value with a default of 0.
func getVal(values map[string]float64, key string) float64 {
	return values[key]
}

// getStr is a helper to get a string value with a default of "".
func getStr(strValues map[string]string, key string) string {
	return strValues[key]
}

// render1040Text renders Form 1040 as formatted text.
func render1040Text(config *FormPDFConfig, values map[string]float64, strValues map[string]string) string {
	sep := "============================================================"

	filingStatus := getStr(strValues, "1040:filing_status")
	firstName := getStr(strValues, "1040:first_name")
	lastName := getStr(strValues, "1040:last_name")
	ssn := getStr(strValues, "1040:ssn")

	// Filing status checkboxes
	statuses := map[string][2]string{
		"single": {"Single", "single"},
		"mfj":    {"MFJ", "mfj"},
		"mfs":    {"MFS", "mfs"},
		"hoh":    {"HOH", "hoh"},
		"qss":    {"QSS", "qss"},
	}
	statusOrder := []string{"single", "mfj", "mfs", "hoh", "qss"}
	var statusParts []string
	for _, code := range statusOrder {
		label := statuses[code][0]
		if code == filingStatus {
			statusParts = append(statusParts, fmt.Sprintf("[X] %s", label))
		} else {
			statusParts = append(statusParts, fmt.Sprintf("[ ] %s", label))
		}
	}

	var b strings.Builder
	b.WriteString(sep + "\n")
	b.WriteString("        Form 1040 - U.S. Individual Income Tax Return\n")
	b.WriteString("                     Tax Year 2025\n")
	b.WriteString(sep + "\n")
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Filing Status: %s\n", strings.Join(statusParts, "  ")))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("First Name: %-20s Last Name: %s\n", firstName, lastName))
	b.WriteString(fmt.Sprintf("SSN: %s\n", FormatSSN(ssn)))
	b.WriteString("\n")

	b.WriteString("INCOME\n")
	b.WriteString("------\n")
	b.WriteString(fmt.Sprintf("1a. Wages, salaries, tips ................ %s\n", FormatCurrency(getVal(values, "1040:1a"))))
	b.WriteString(fmt.Sprintf("1z. Total from W-2s ..................... %s\n", FormatCurrency(getVal(values, "1040:1z"))))
	if v := getVal(values, "1040:2a"); v > 0 {
		b.WriteString(fmt.Sprintf("2a. Tax-exempt interest .................. %s\n", FormatCurrency(v)))
	}
	if v := getVal(values, "1040:2b"); v > 0 {
		b.WriteString(fmt.Sprintf("2b. Taxable interest ..................... %s\n", FormatCurrency(v)))
	}
	if v := getVal(values, "1040:3a"); v > 0 {
		b.WriteString(fmt.Sprintf("3a. Qualified dividends .................. %s\n", FormatCurrency(v)))
	}
	if v := getVal(values, "1040:3b"); v > 0 {
		b.WriteString(fmt.Sprintf("3b. Ordinary dividends ................... %s\n", FormatCurrency(v)))
	}
	if v := getVal(values, "1040:7"); v != 0 {
		b.WriteString(fmt.Sprintf("7.  Capital gain or (loss) ............... %s\n", FormatCurrency(v)))
	}
	if v := getVal(values, "1040:8"); v != 0 {
		b.WriteString(fmt.Sprintf("8.  Other income (Schedule 1) ............ %s\n", FormatCurrency(v)))
	}
	b.WriteString(fmt.Sprintf("9.  Total income ........................ %s\n", FormatCurrency(getVal(values, "1040:9"))))
	b.WriteString("\n")

	b.WriteString("ADJUSTED GROSS INCOME\n")
	b.WriteString("---------------------\n")
	b.WriteString(fmt.Sprintf("11. Adjusted gross income ............... %s\n", FormatCurrency(getVal(values, "1040:11"))))
	b.WriteString("\n")

	b.WriteString("DEDUCTIONS\n")
	b.WriteString("----------\n")
	b.WriteString(fmt.Sprintf("12. Standard deduction .................. %s\n", FormatCurrency(getVal(values, "1040:12"))))
	b.WriteString(fmt.Sprintf("14. Total deductions .................... %s\n", FormatCurrency(getVal(values, "1040:14"))))
	b.WriteString(fmt.Sprintf("15. Taxable income ...................... %s\n", FormatCurrency(getVal(values, "1040:15"))))
	b.WriteString("\n")

	b.WriteString("TAX AND CREDITS\n")
	b.WriteString("---------------\n")
	b.WriteString(fmt.Sprintf("16. Tax ................................. %s\n", FormatCurrency(getVal(values, "1040:16"))))
	b.WriteString(fmt.Sprintf("24. Total tax ........................... %s\n", FormatCurrency(getVal(values, "1040:24"))))
	b.WriteString("\n")

	b.WriteString("PAYMENTS\n")
	b.WriteString("--------\n")
	b.WriteString(fmt.Sprintf("25a. Federal tax withheld (W-2) ......... %s\n", FormatCurrency(getVal(values, "1040:25a"))))
	if v := getVal(values, "1040:25b"); v > 0 {
		b.WriteString(fmt.Sprintf("25b. Federal tax withheld (1099) ........ %s\n", FormatCurrency(v)))
	}
	b.WriteString(fmt.Sprintf("25d. Total federal tax withheld ......... %s\n", FormatCurrency(getVal(values, "1040:25d"))))
	b.WriteString(fmt.Sprintf("33. Total payments ...................... %s\n", FormatCurrency(getVal(values, "1040:33"))))
	b.WriteString("\n")

	refund := getVal(values, "1040:34")
	owe := getVal(values, "1040:37")
	if refund > 0 {
		b.WriteString("REFUND\n")
		b.WriteString("------\n")
		b.WriteString(fmt.Sprintf("34. Amount overpaid ..................... %s\n", FormatCurrency(refund)))
	} else if owe > 0 {
		b.WriteString("AMOUNT YOU OWE\n")
		b.WriteString("--------------\n")
		b.WriteString(fmt.Sprintf("37. Amount you owe ...................... %s\n", FormatCurrency(owe)))
	} else {
		b.WriteString("BALANCE\n")
		b.WriteString("-------\n")
		b.WriteString("    No refund or amount owed ............ $0.00\n")
	}

	b.WriteString("\n")
	b.WriteString(sep + "\n")

	return b.String()
}

// renderCA540Text renders CA Form 540 as formatted text.
func renderCA540Text(config *FormPDFConfig, values map[string]float64, strValues map[string]string) string {
	sep := "============================================================"

	filingStatus := getStr(strValues, "ca_540:filing_status")
	if filingStatus == "" {
		filingStatus = getStr(strValues, "1040:filing_status")
	}

	var b strings.Builder
	b.WriteString(sep + "\n")
	b.WriteString("    Form 540 - California Resident Income Tax Return\n")
	b.WriteString("                     Tax Year 2025\n")
	b.WriteString(sep + "\n")
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Filing Status: %s\n", strings.ToUpper(filingStatus)))
	b.WriteString("\n")

	b.WriteString("INCOME\n")
	b.WriteString("------\n")
	b.WriteString(fmt.Sprintf("7.  Wages, salaries, tips (CA) .......... %s\n", FormatCurrency(getVal(values, "ca_540:7"))))
	b.WriteString(fmt.Sprintf("13. Federal adjusted gross income ....... %s\n", FormatCurrency(getVal(values, "ca_540:13"))))
	b.WriteString(fmt.Sprintf("14. CA subtractions ..................... %s\n", FormatCurrency(getVal(values, "ca_540:14"))))
	b.WriteString(fmt.Sprintf("15. CA additions ........................ %s\n", FormatCurrency(getVal(values, "ca_540:15"))))
	b.WriteString(fmt.Sprintf("17. California AGI ...................... %s\n", FormatCurrency(getVal(values, "ca_540:17"))))
	b.WriteString("\n")

	b.WriteString("DEDUCTIONS\n")
	b.WriteString("----------\n")
	b.WriteString(fmt.Sprintf("18. Standard deduction .................. %s\n", FormatCurrency(getVal(values, "ca_540:18"))))
	b.WriteString(fmt.Sprintf("19. Taxable income ...................... %s\n", FormatCurrency(getVal(values, "ca_540:19"))))
	b.WriteString("\n")

	b.WriteString("TAX\n")
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("31. California tax ...................... %s\n", FormatCurrency(getVal(values, "ca_540:31"))))
	b.WriteString(fmt.Sprintf("32. Exemption credits ................... %s\n", FormatCurrency(getVal(values, "ca_540:32"))))
	b.WriteString(fmt.Sprintf("35. Net tax ............................. %s\n", FormatCurrency(getVal(values, "ca_540:35"))))
	b.WriteString(fmt.Sprintf("36. Mental Health Services Tax .......... %s\n", FormatCurrency(getVal(values, "ca_540:36"))))
	b.WriteString(fmt.Sprintf("40. Total California tax ................ %s\n", FormatCurrency(getVal(values, "ca_540:40"))))
	b.WriteString("\n")

	b.WriteString("PAYMENTS\n")
	b.WriteString("--------\n")
	b.WriteString(fmt.Sprintf("71. CA tax withheld ..................... %s\n", FormatCurrency(getVal(values, "ca_540:71"))))
	b.WriteString(fmt.Sprintf("74. Total payments ...................... %s\n", FormatCurrency(getVal(values, "ca_540:74"))))
	b.WriteString("\n")

	refund := getVal(values, "ca_540:91")
	owe := getVal(values, "ca_540:93")
	if refund > 0 {
		b.WriteString("REFUND\n")
		b.WriteString("------\n")
		b.WriteString(fmt.Sprintf("91. Amount overpaid ..................... %s\n", FormatCurrency(refund)))
	} else if owe > 0 {
		b.WriteString("AMOUNT YOU OWE\n")
		b.WriteString("--------------\n")
		b.WriteString(fmt.Sprintf("93. Amount you owe ...................... %s\n", FormatCurrency(owe)))
	} else {
		b.WriteString("BALANCE\n")
		b.WriteString("-------\n")
		b.WriteString("    No refund or amount owed ............ $0.00\n")
	}

	b.WriteString("\n")
	b.WriteString(sep + "\n")

	return b.String()
}

// renderScheduleBText renders Schedule B as formatted text.
func renderScheduleBText(config *FormPDFConfig, values map[string]float64, strValues map[string]string) string {
	sep := "============================================================"

	var b strings.Builder
	b.WriteString(sep + "\n")
	b.WriteString("   Schedule B — Interest and Ordinary Dividends\n")
	b.WriteString("                     Tax Year 2025\n")
	b.WriteString(sep + "\n")
	b.WriteString("\n")

	b.WriteString("PART I: INTEREST\n")
	b.WriteString("----------------\n")
	b.WriteString(fmt.Sprintf("1. Interest income ...................... %s\n", FormatCurrency(getVal(values, "schedule_b:1"))))
	b.WriteString(fmt.Sprintf("4. Total interest ....................... %s\n", FormatCurrency(getVal(values, "schedule_b:4"))))
	b.WriteString("\n")

	b.WriteString("PART II: ORDINARY DIVIDENDS\n")
	b.WriteString("---------------------------\n")
	b.WriteString(fmt.Sprintf("5. Ordinary dividends ................... %s\n", FormatCurrency(getVal(values, "schedule_b:5"))))
	b.WriteString(fmt.Sprintf("6. Total ordinary dividends ............. %s\n", FormatCurrency(getVal(values, "schedule_b:6"))))
	b.WriteString("\n")
	b.WriteString(sep + "\n")

	return b.String()
}

// renderSchedule1Text renders Schedule 1 as formatted text.
func renderSchedule1Text(config *FormPDFConfig, values map[string]float64, strValues map[string]string) string {
	sep := "============================================================"

	var b strings.Builder
	b.WriteString(sep + "\n")
	b.WriteString("  Schedule 1 — Additional Income and Adjustments to Income\n")
	b.WriteString("                     Tax Year 2025\n")
	b.WriteString(sep + "\n")
	b.WriteString("\n")

	b.WriteString("PART I: ADDITIONAL INCOME\n")
	b.WriteString("-------------------------\n")
	if v := getVal(values, "schedule_1:1"); v != 0 {
		b.WriteString(fmt.Sprintf("1.  Taxable refunds ..................... %s\n", FormatCurrency(v)))
	}
	if v := getVal(values, "schedule_1:3"); v != 0 {
		b.WriteString(fmt.Sprintf("3.  Business income ..................... %s\n", FormatCurrency(v)))
	}
	if v := getVal(values, "schedule_1:7"); v != 0 {
		b.WriteString(fmt.Sprintf("7.  Capital gain or (loss) .............. %s\n", FormatCurrency(v)))
	}
	b.WriteString(fmt.Sprintf("10. Total additional income ............. %s\n", FormatCurrency(getVal(values, "schedule_1:10"))))
	b.WriteString("\n")

	b.WriteString("PART II: ADJUSTMENTS TO INCOME\n")
	b.WriteString("------------------------------\n")
	if v := getVal(values, "schedule_1:15"); v != 0 {
		b.WriteString(fmt.Sprintf("15. HSA deduction ....................... %s\n", FormatCurrency(v)))
	}
	if v := getVal(values, "schedule_1:24"); v != 0 {
		b.WriteString(fmt.Sprintf("24. Early withdrawal penalty ............ %s\n", FormatCurrency(v)))
	}
	b.WriteString(fmt.Sprintf("26. Total adjustments ................... %s\n", FormatCurrency(getVal(values, "schedule_1:26"))))
	b.WriteString("\n")
	b.WriteString(sep + "\n")

	return b.String()
}

// renderScheduleDText renders Schedule D as formatted text.
func renderScheduleDText(config *FormPDFConfig, values map[string]float64, strValues map[string]string) string {
	sep := "============================================================"

	var b strings.Builder
	b.WriteString(sep + "\n")
	b.WriteString("        Schedule D — Capital Gains and Losses\n")
	b.WriteString("                     Tax Year 2025\n")
	b.WriteString(sep + "\n")
	b.WriteString("\n")

	b.WriteString("PART I: SHORT-TERM CAPITAL GAINS AND LOSSES\n")
	b.WriteString("--------------------------------------------\n")
	if v := getVal(values, "schedule_d:1"); v != 0 {
		b.WriteString(fmt.Sprintf("1.  Short-term from Form 8949 ........... %s\n", FormatCurrency(v)))
	}
	b.WriteString(fmt.Sprintf("7.  Net short-term gain or (loss) ....... %s\n", FormatCurrency(getVal(values, "schedule_d:7"))))
	b.WriteString("\n")

	b.WriteString("PART II: LONG-TERM CAPITAL GAINS AND LOSSES\n")
	b.WriteString("--------------------------------------------\n")
	if v := getVal(values, "schedule_d:8"); v != 0 {
		b.WriteString(fmt.Sprintf("8.  Long-term from Form 8949 ............ %s\n", FormatCurrency(v)))
	}
	if v := getVal(values, "schedule_d:13"); v != 0 {
		b.WriteString(fmt.Sprintf("13. Capital gain distributions .......... %s\n", FormatCurrency(v)))
	}
	b.WriteString(fmt.Sprintf("15. Net long-term gain or (loss) ........ %s\n", FormatCurrency(getVal(values, "schedule_d:15"))))
	b.WriteString("\n")

	b.WriteString("PART III: SUMMARY\n")
	b.WriteString("-----------------\n")
	b.WriteString(fmt.Sprintf("16. Net capital gain or (loss) .......... %s\n", FormatCurrency(getVal(values, "schedule_d:16"))))
	b.WriteString("\n")
	b.WriteString(sep + "\n")

	return b.String()
}

// renderGenericText renders any form as a simple key-value text export.
func renderGenericText(config *FormPDFConfig, values map[string]float64, strValues map[string]string) string {
	sep := "============================================================"

	var b strings.Builder
	b.WriteString(sep + "\n")
	b.WriteString(fmt.Sprintf("        %s\n", config.FormName))
	b.WriteString(sep + "\n")
	b.WriteString("\n")

	for _, m := range config.Mappings {
		val := formatFieldValue(m, values, strValues)
		if val != "" {
			b.WriteString(fmt.Sprintf("%-40s %s\n", m.PDFField, val))
		}
	}

	b.WriteString("\n")
	b.WriteString(sep + "\n")

	return b.String()
}
