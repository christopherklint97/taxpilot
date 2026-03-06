package forms

import (
	"fmt"
	"strings"
)

// FieldType represents how a field's value is determined.
type FieldType int

const (
	UserInput  FieldType = iota // needs taxpayer input
	Computed                    // calculated from other fields
	Lookup                      // from tax tables
	PriorYear                   // carried from last year
	FederalRef                  // state form referencing federal field
)

// Jurisdiction represents the taxing authority for a form.
type Jurisdiction int

const (
	Federal Jurisdiction = iota
	StateCA
)

// DepValues is passed to Compute functions for accessing dependency values.
type DepValues struct {
	values    map[string]float64
	strValues map[string]string
	taxYear   int
}

// NewDepValues creates a DepValues from the given maps and tax year.
func NewDepValues(values map[string]float64, strValues map[string]string, taxYear int) DepValues {
	if values == nil {
		values = make(map[string]float64)
	}
	if strValues == nil {
		strValues = make(map[string]string)
	}
	return DepValues{
		values:    values,
		strValues: strValues,
		taxYear:   taxYear,
	}
}

// Get returns the float64 value for the given key, or 0 if not found.
func (d DepValues) Get(key string) float64 {
	return d.values[key]
}

// GetString returns the string value for the given key, or "" if not found.
func (d DepValues) GetString(key string) string {
	return d.strValues[key]
}

// SumAll sums all float64 values whose keys match the given wildcard pattern.
// The pattern supports '*' as a wildcard that matches any substring within a
// single segment. For example, "w2:*:wages" matches "w2:employer1:wages" and
// "w2:employer2:wages".
func (d DepValues) SumAll(pattern string) float64 {
	var sum float64
	for k, v := range d.values {
		if matchWildcard(pattern, k) {
			sum += v
		}
	}
	return sum
}

// TaxYear returns the tax year for this computation context.
func (d DepValues) TaxYear() int {
	return d.taxYear
}

// matchWildcard checks if s matches a pattern containing '*' wildcards.
// Each '*' matches zero or more characters.
func matchWildcard(pattern, s string) bool {
	// Split pattern on '*' and check segments appear in order
	parts := strings.Split(pattern, "*")
	if len(parts) == 1 {
		return pattern == s
	}

	idx := 0
	for i, part := range parts {
		if part == "" {
			continue
		}
		pos := strings.Index(s[idx:], part)
		if pos < 0 {
			return false
		}
		if i == 0 && pos != 0 {
			// First part must match at start if it's non-empty
			return false
		}
		idx += pos + len(part)
	}
	// If the last part is non-empty, the string must end with it
	if last := parts[len(parts)-1]; last != "" {
		return strings.HasSuffix(s, last)
	}
	return true
}

// FieldDef defines a single field on a tax form.
type FieldDef struct {
	Line       string
	Type       FieldType
	Label      string
	Prompt     string                // human-readable question (for UserInput)
	DependsOn  []string              // field keys this depends on (form_id:line format)
	Options    []string              // for enum-type UserInput fields
	Compute    func(DepValues) float64 // for Computed/FederalRef/Lookup fields
	ComputeStr func(DepValues) string  // for string-valued computed fields
}

// FormDef defines a complete tax form.
type FormDef struct {
	ID           string
	Name         string
	Jurisdiction Jurisdiction
	TaxYears     []int
	Fields       []FieldDef
}

// FieldKey returns the fully qualified key for a field: "form_id:line".
func FieldKey(formID, line string) string {
	return fmt.Sprintf("%s:%s", formID, line)
}
