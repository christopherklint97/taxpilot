package taxmath

import "math"

// RoundToNearest rounds to the nearest dollar (IRS/FTB standard rounding).
// Values at exactly .50 round up.
func RoundToNearest(amount float64) float64 {
	return math.Round(amount)
}

// RoundDown truncates to the dollar (used for some specific lines).
func RoundDown(amount float64) float64 {
	return math.Floor(amount)
}
