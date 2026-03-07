package taxmath

// FEIE (Foreign Earned Income Exclusion) limits by tax year.
var feieLimits = map[int]float64{
	2025: 130000,
}

// FEIELimit returns the foreign earned income exclusion limit for the given tax year.
func FEIELimit(taxYear int) float64 {
	if limit, ok := feieLimits[taxYear]; ok {
		return limit
	}
	return 0
}

// HousingBaseAmount returns the base housing amount (16% of FEIE limit).
// Housing expenses below this amount are not excludable.
func HousingBaseAmount(taxYear int) float64 {
	return FEIELimit(taxYear) * 0.16
}

// HousingMaxAmount returns the maximum housing expenses that can be considered
// for the exclusion/deduction (30% of FEIE limit for default locations).
func HousingMaxAmount(taxYear int) float64 {
	return FEIELimit(taxYear) * 0.30
}

// ProrateExclusion prorates the FEIE limit based on qualifying days.
// For the physical presence test, qualifyingDays is the number of days
// physically present in a foreign country. For BFRT with full year, use 365.
func ProrateExclusion(limit float64, qualifyingDays, totalDays int) float64 {
	if totalDays <= 0 || qualifyingDays <= 0 {
		return 0
	}
	if qualifyingDays >= totalDays {
		return limit
	}
	return limit * float64(qualifyingDays) / float64(totalDays)
}

// ComputeTaxWithStacking computes tax using the "stacking" method required
// when FEIE is claimed. The tax on the remaining taxable income is computed
// at the rate that would apply if the excluded income were still included.
//
// This prevents taxpayers from benefiting from lower brackets on their
// remaining income after excluding a large amount via FEIE.
//
// Formula: tax(taxableIncome + excludedIncome) - tax(excludedIncome)
func ComputeTaxWithStacking(taxableIncome, excludedIncome float64, fs FilingStatus, year int, jurisdiction JurisdictionType) float64 {
	if taxableIncome <= 0 {
		return 0
	}
	if excludedIncome <= 0 {
		return ComputeTax(taxableIncome, fs, year, jurisdiction)
	}

	// Tax on total income (as if exclusion didn't exist)
	taxOnTotal := ComputeTax(taxableIncome+excludedIncome, fs, year, jurisdiction)
	// Tax on just the excluded amount
	taxOnExcluded := ComputeTax(excludedIncome, fs, year, jurisdiction)

	// The difference is the tax on the remaining income at stacked rates
	stacked := taxOnTotal - taxOnExcluded
	if stacked < 0 {
		return 0
	}
	return stacked
}
