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

// GetStrict returns the float64 value for the given key, or an error if the key doesn't exist.
func (d DepValues) GetStrict(key string) (float64, error) {
	v, ok := d.values[key]
	if !ok {
		return 0, fmt.Errorf("dependency key %q not found in DepValues", key)
	}
	return v, nil
}

// Keys returns all available numeric keys in the DepValues.
func (d DepValues) Keys() []string {
	keys := make([]string, 0, len(d.values))
	for k := range d.values {
		keys = append(keys, k)
	}
	return keys
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

// SumAllWhere sums all float64 values whose keys match valuePattern, but only
// for instances where the corresponding filterPattern key has the given
// filterValue as its string value. This enables filtering wildcard sums by
// a categorical field (e.g., summing proceeds only for short-term transactions).
//
// Example: SumAllWhere("1099b:*:proceeds", "1099b:*:term", "short")
// sums proceeds for all 1099b instances where term == "short".
func (d DepValues) SumAllWhere(valuePattern, filterPattern, filterValue string) float64 {
	var sum float64
	for k, v := range d.values {
		if !matchWildcard(valuePattern, k) {
			continue
		}
		// Extract the instance segment from the matched key and build
		// the corresponding filter key.
		filterKey := buildCorrespondingKey(valuePattern, filterPattern, k)
		if filterKey == "" {
			continue
		}
		if sv, ok := d.strValues[filterKey]; ok && sv == filterValue {
			sum += v
		}
	}
	return sum
}

// buildCorrespondingKey takes a valuePattern, a filterPattern, and a concrete
// key that matched valuePattern, and returns the concrete key that corresponds
// to filterPattern with the same wildcard segment.
//
// Example: buildCorrespondingKey("1099b:*:proceeds", "1099b:*:term", "1099b:1:proceeds")
// returns "1099b:1:term"
func buildCorrespondingKey(valuePattern, filterPattern, concreteKey string) string {
	// Find the wildcard segment by splitting both pattern and concrete key
	vParts := strings.Split(valuePattern, "*")
	if len(vParts) != 2 {
		return "" // only single-wildcard patterns supported
	}
	prefix := vParts[0]
	suffix := vParts[1]

	if !strings.HasPrefix(concreteKey, prefix) || !strings.HasSuffix(concreteKey, suffix) {
		return ""
	}

	// Extract the wildcard segment
	wildcardSeg := concreteKey[len(prefix) : len(concreteKey)-len(suffix)]

	// Build the filter key by replacing * with the wildcard segment
	return strings.Replace(filterPattern, "*", wildcardSeg, 1)
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

// FieldValueType indicates the data type of a field's value.
type FieldValueType int

const (
	NumericValue FieldValueType = iota // default: numeric/currency
	StringValue                        // text input (names, SSN, EIN, etc.)
	IntegerValue                       // whole number (counts, not currency)
)

// FieldDef defines a single field on a tax form.
type FieldDef struct {
	Line       string
	Type       FieldType
	ValueType  FieldValueType // NumericValue or StringValue
	Label      string
	Prompt     string                  // human-readable question (for UserInput)
	DependsOn  []string                // field keys this depends on (form_id:line format)
	Options    []string                // for enum-type UserInput fields
	Compute    func(DepValues) float64 // for Computed/FederalRef/Lookup fields
	ComputeStr func(DepValues) string  // for string-valued computed fields
}

// FormDef defines a complete tax form.
type FormDef struct {
	ID            FormID
	Name          string
	Jurisdiction  Jurisdiction
	TaxYears      []int
	Fields        []FieldDef
	QuestionGroup string // e.g., "personal", "income_w2", "expat", "ca"
	QuestionOrder int    // sort order within group (lower = earlier)

	fieldIndex map[string]int // lazily built: line -> index in Fields
}

// FieldByLine returns the FieldDef for the given line, or nil if not found.
// The first call builds an internal index for O(1) lookups on subsequent calls.
func (f *FormDef) FieldByLine(line string) *FieldDef {
	if f.fieldIndex == nil {
		f.fieldIndex = make(map[string]int, len(f.Fields))
		for i := range f.Fields {
			f.fieldIndex[f.Fields[i].Line] = i
		}
	}
	idx, ok := f.fieldIndex[line]
	if !ok {
		return nil
	}
	return &f.Fields[idx]
}

// FieldKey returns the fully qualified key for a field: "form_id:line".
func FieldKey(formID FormID, line string) string {
	return fmt.Sprintf("%s:%s", string(formID), line)
}

// FK is a convenience for building field keys with a typed FormID and line.
// It is identical to FieldKey but shorter for use in DependsOn lists and map lookups.
func FK(formID FormID, line string) string {
	return string(formID) + ":" + line
}
