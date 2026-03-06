package taxmath

import "math"

// Bracket represents a single tax bracket.
type Bracket struct {
	Min  float64
	Max  float64 // use math.MaxFloat64 for the top bracket
	Rate float64 // e.g., 0.10 for 10%
}

// BracketTable is a set of brackets for a filing status.
type BracketTable []Bracket

// ComputeBracketTax applies progressive bracket rates to taxable income.
// Income is taxed at each bracket's rate only for the portion within that bracket.
func ComputeBracketTax(income float64, brackets BracketTable) float64 {
	if income <= 0 {
		return 0
	}
	tax := 0.0
	for _, b := range brackets {
		if income <= b.Min {
			break
		}
		top := math.Min(income, b.Max)
		taxable := top - b.Min
		tax += taxable * b.Rate
	}
	return tax
}
