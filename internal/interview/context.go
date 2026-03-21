package interview

import (
	"fmt"
	"strings"

	"taxpilot/internal/forms"
)

// ContextSummary provides a human-readable summary of what's known so far,
// used as context for LLM-generated explanations.
type ContextSummary struct {
	FilingStatus     string
	HasPriorYear     bool
	PriorYearSummary string            // "Last year: Single, $75K wages, $8K federal tax, CA resident"
	AnsweredSoFar    map[string]string  // key -> formatted value
	CurrentForm      string
	StateCode        string
}

// BuildContextSummary creates a ContextSummary from the engine's current state.
func BuildContextSummary(e *Engine) ContextSummary {
	cs := ContextSummary{
		AnsweredSoFar: make(map[string]string),
		StateCode:     "",
	}

	// Extract filing status from string inputs
	if fs, ok := e.strInputs[forms.F1040FilingStatus]; ok {
		cs.FilingStatus = fs
	}

	// Determine current form context
	if q := e.Current(); q != nil {
		cs.CurrentForm = q.FormName
	}

	// Build answered-so-far map with human-readable labels
	for _, q := range e.questions {
		if q.IsString || len(q.Options) > 0 {
			if sv, ok := e.strInputs[q.Key]; ok {
				cs.AnsweredSoFar[q.Key] = sv
			}
		} else {
			if nv, ok := e.inputs[q.Key]; ok {
				cs.AnsweredSoFar[q.Key] = formatCurrency(nv)
			}
		}
	}

	// Build prior-year summary if available
	if len(e.priorYear) > 0 || len(e.priorYearStr) > 0 {
		cs.HasPriorYear = true
		cs.PriorYearSummary = buildPriorYearSummary(e)
	}

	return cs
}

// buildPriorYearSummary creates a concise summary string from prior-year data.
func buildPriorYearSummary(e *Engine) string {
	var parts []string

	// Filing status
	if fs, ok := e.priorYearStr[forms.F1040FilingStatus]; ok && fs != "" {
		parts = append(parts, formatFilingStatus(fs))
	}

	// Wages
	if wages, ok := e.priorYear["w2:1:wages"]; ok && wages > 0 {
		parts = append(parts, fmt.Sprintf("%s wages", formatCompact(wages)))
	}

	// Federal tax
	if fedTax, ok := e.priorYear["1040:total_tax"]; ok && fedTax > 0 {
		parts = append(parts, fmt.Sprintf("%s federal tax", formatCompact(fedTax)))
	}

	// State tax
	if stateTax, ok := e.priorYear["540:total_tax"]; ok && stateTax > 0 {
		parts = append(parts, fmt.Sprintf("%s CA tax", formatCompact(stateTax)))
	}

	if len(parts) == 0 {
		return "Prior year data available"
	}
	return strings.Join(parts, ", ")
}

// formatCompact formats a number as "$75,000" or "$8,114" (no cents).
func formatCompact(amount float64) string {
	whole := int64(amount + 0.5)
	s := fmt.Sprintf("%d", whole)
	if len(s) > 3 {
		var groups []string
		for len(s) > 3 {
			groups = append([]string{s[len(s)-3:]}, groups...)
			s = s[:len(s)-3]
		}
		groups = append([]string{s}, groups...)
		s = strings.Join(groups, ",")
	}
	return "$" + s
}

// formatFilingStatus converts a filing status code to a readable label.
func formatFilingStatus(code string) string {
	switch code {
	case forms.FilingSingle:
		return "Single"
	case forms.FilingMFJ:
		return "Married Filing Jointly"
	case forms.FilingMFS:
		return "Married Filing Separately"
	case forms.FilingHOH:
		return "Head of Household"
	case forms.FilingQSS:
		return "Qualifying Surviving Spouse"
	default:
		return code
	}
}

// FormatForLLM returns the context as a string suitable for inclusion in an LLM prompt.
func (cs ContextSummary) FormatForLLM() string {
	var lines []string
	lines = append(lines, "Taxpayer context:")

	if cs.FilingStatus != "" {
		lines = append(lines, fmt.Sprintf("- Filing status: %s", formatFilingStatus(cs.FilingStatus)))
	}

	if cs.StateCode != "" {
		stateName := cs.StateCode
		if cs.StateCode == forms.StateCodeCA {
			stateName = "California"
		}
		lines = append(lines, fmt.Sprintf("- State: %s", stateName))
	}

	if cs.HasPriorYear && cs.PriorYearSummary != "" {
		lines = append(lines, fmt.Sprintf("- Prior year: %s", cs.PriorYearSummary))
	}

	// Include a subset of current answers for context
	if len(cs.AnsweredSoFar) > 0 {
		var answerParts []string
		// Show key answers in a readable format
		friendlyNames := map[string]string{
			"1040:first_name":       "First name",
			"1040:last_name":        "Last name",
			"w2:1:employer_name":    "Employer",
			"w2:1:wages":            "Wages",
			"w2:1:federal_tax_withheld": "Federal withholding",
			"w2:1:state_tax_withheld":   "State withholding",
		}
		for key, friendly := range friendlyNames {
			if val, ok := cs.AnsweredSoFar[key]; ok {
				answerParts = append(answerParts, fmt.Sprintf("%s: %s", friendly, val))
			}
		}
		if len(answerParts) > 0 {
			lines = append(lines, fmt.Sprintf("- Current answers: %s", strings.Join(answerParts, ", ")))
		}
	}

	return strings.Join(lines, "\n")
}
