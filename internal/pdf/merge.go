package pdf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"taxpilot/internal/forms"
)

// MergeReturns combines multiple ParsedReturn results into one.
// Later returns override earlier ones for duplicate keys.
// The tax year is taken from the first return that has one.
func MergeReturns(returns []*ParsedReturn) *ParsedReturn {
	if len(returns) == 0 {
		return nil
	}
	if len(returns) == 1 {
		return returns[0]
	}

	merged := &ParsedReturn{
		Fields:    make(map[string]float64),
		StrFields: make(map[string]string),
		RawFields: make(map[string]string),
	}

	var formIDs []forms.FormID

	for _, r := range returns {
		if r == nil {
			continue
		}

		if merged.TaxYear == 0 && r.TaxYear > 0 {
			merged.TaxYear = r.TaxYear
		}
		if r.FormID != "" {
			formIDs = append(formIDs, r.FormID)
		}

		for k, v := range r.Fields {
			merged.Fields[k] = v
		}
		for k, v := range r.StrFields {
			merged.StrFields[k] = v
		}
		for k, v := range r.RawFields {
			merged.RawFields[k] = v
		}
	}

	// Use the first form ID; caller can check FormIDs via the list.
	if len(formIDs) > 0 {
		merged.FormID = formIDs[0]
	}

	return merged
}

// MergeInto adds fields from src into dst, overwriting on conflict.
func MergeInto(dst, src *ParsedReturn) {
	if src == nil || dst == nil {
		return
	}
	if dst.TaxYear == 0 && src.TaxYear > 0 {
		dst.TaxYear = src.TaxYear
	}
	if dst.Fields == nil {
		dst.Fields = make(map[string]float64)
	}
	if dst.StrFields == nil {
		dst.StrFields = make(map[string]string)
	}
	for k, v := range src.Fields {
		dst.Fields[k] = v
	}
	for k, v := range src.StrFields {
		dst.StrFields[k] = v
	}
}

// ParseMultipleFiles parses all given PDF paths and merges the results.
// Paths can be files or directories (directories are scanned for *.pdf).
// Returns the merged result and a list of successfully parsed form names.
func ParseMultipleFiles(paths []string) (*ParsedReturn, []string, error) {
	expanded, err := expandPaths(paths)
	if err != nil {
		return nil, nil, err
	}
	if len(expanded) == 0 {
		return nil, nil, fmt.Errorf("no PDF files found in provided paths")
	}

	parser := NewParser()
	registerAllParseForms(parser)

	var results []*ParsedReturn
	var formNames []string
	var parseErrors []string

	for _, path := range expanded {
		parsed, err := parser.ParseFile(path)
		if err != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("%s: %v", filepath.Base(path), err))
			continue
		}
		results = append(results, parsed)
		formNames = append(formNames, formLabel(parsed.FormID))
	}

	if len(results) == 0 {
		return nil, nil, fmt.Errorf("could not parse any files: %s", strings.Join(parseErrors, "; "))
	}

	merged := MergeReturns(results)
	return merged, formNames, nil
}

// expandPaths resolves file paths and directories into a list of PDF file paths.
func expandPaths(paths []string) ([]string, error) {
	var result []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, fmt.Errorf("cannot access %s: %w", p, err)
		}
		if info.IsDir() {
			matches, err := filepath.Glob(filepath.Join(p, "*.pdf"))
			if err != nil {
				return nil, fmt.Errorf("glob %s: %w", p, err)
			}
			// Also check uppercase .PDF
			matchesUpper, _ := filepath.Glob(filepath.Join(p, "*.PDF"))
			result = append(result, matches...)
			result = append(result, matchesUpper...)
		} else {
			result = append(result, p)
		}
	}
	return result, nil
}

// registerAllParseForms registers all known form PDF mappings on a parser.
func registerAllParseForms(p *Parser) {
	p.RegisterForm(Federal1040Mappings())
	p.RegisterForm(ScheduleAMappings())
	p.RegisterForm(ScheduleBMappings())
	p.RegisterForm(ScheduleCMappings())
	p.RegisterForm(ScheduleDMappings())
	p.RegisterForm(Form8949Mappings())
	p.RegisterForm(Schedule1Mappings())
	p.RegisterForm(Schedule2Mappings())
	p.RegisterForm(Schedule3Mappings())
	p.RegisterForm(ScheduleSEMappings())
	p.RegisterForm(Form8995Mappings())
	p.RegisterForm(Form8889Mappings())
	p.RegisterForm(CA540Mappings())
	p.RegisterForm(ScheduleCAMappings())
	p.RegisterForm(Form3514Mappings())
	p.RegisterForm(Form3853Mappings())
}

// formLabel returns a human-readable name for a form ID.
func formLabel(id forms.FormID) string {
	switch id {
	case forms.FormF1040:
		return "Form 1040"
	case forms.FormScheduleA:
		return "Schedule A"
	case forms.FormScheduleB:
		return "Schedule B"
	case forms.FormScheduleC:
		return "Schedule C"
	case forms.FormScheduleD:
		return "Schedule D"
	case forms.FormSchedule1:
		return "Schedule 1"
	case forms.FormSchedule2:
		return "Schedule 2"
	case forms.FormSchedule3:
		return "Schedule 3"
	case forms.FormScheduleSE:
		return "Schedule SE"
	case forms.FormF8949:
		return "Form 8949"
	case forms.FormF8889:
		return "Form 8889"
	case forms.FormF8995:
		return "Form 8995"
	case forms.FormCA540:
		return "CA Form 540"
	case forms.FormScheduleCA:
		return "CA Schedule CA"
	case forms.FormF3514:
		return "CA Form 3514"
	case forms.FormF3853:
		return "CA Form 3853"
	case forms.FormCA540NR:
		return "CA Form 540NR"
	case forms.FormF2555:
		return "Form 2555"
	case forms.FormF1116:
		return "Form 1116"
	case forms.FormF8938:
		return "Form 8938"
	case forms.FormF8833:
		return "Form 8833"
	default:
		return string(id)
	}
}
