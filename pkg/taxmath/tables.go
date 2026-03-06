package taxmath

import "math"

// FilingStatus represents the taxpayer's filing status.
type FilingStatus string

const (
	Single             FilingStatus = "single"
	MarriedFilingJoint FilingStatus = "mfj"
	MarriedFilingSep   FilingStatus = "mfs"
	HeadOfHousehold    FilingStatus = "hoh"
	QualifyingSurvivor FilingStatus = "qss"
)

// JurisdictionType identifies federal vs. state tax systems.
type JurisdictionType int

const (
	Federal JurisdictionType = iota
	StateCA
)

// ---------------------------------------------------------------------------
// Federal 2025 brackets (IRS Rev. Proc. 2024-40)
// ---------------------------------------------------------------------------

var federalBrackets2025 = map[FilingStatus]BracketTable{
	Single: {
		{Min: 0, Max: 11925, Rate: 0.10},
		{Min: 11925, Max: 48475, Rate: 0.12},
		{Min: 48475, Max: 103350, Rate: 0.22},
		{Min: 103350, Max: 197300, Rate: 0.24},
		{Min: 197300, Max: 250525, Rate: 0.32},
		{Min: 250525, Max: 626350, Rate: 0.35},
		{Min: 626350, Max: math.MaxFloat64, Rate: 0.37},
	},
	MarriedFilingJoint: {
		{Min: 0, Max: 23850, Rate: 0.10},
		{Min: 23850, Max: 96950, Rate: 0.12},
		{Min: 96950, Max: 206700, Rate: 0.22},
		{Min: 206700, Max: 394600, Rate: 0.24},
		{Min: 394600, Max: 501050, Rate: 0.32},
		{Min: 501050, Max: 751600, Rate: 0.35},
		{Min: 751600, Max: math.MaxFloat64, Rate: 0.37},
	},
	MarriedFilingSep: {
		{Min: 0, Max: 11925, Rate: 0.10},
		{Min: 11925, Max: 48475, Rate: 0.12},
		{Min: 48475, Max: 103350, Rate: 0.22},
		{Min: 103350, Max: 197300, Rate: 0.24},
		{Min: 197300, Max: 250525, Rate: 0.32},
		{Min: 250525, Max: 375800, Rate: 0.35},
		{Min: 375800, Max: math.MaxFloat64, Rate: 0.37},
	},
	HeadOfHousehold: {
		{Min: 0, Max: 17000, Rate: 0.10},
		{Min: 17000, Max: 64850, Rate: 0.12},
		{Min: 64850, Max: 103350, Rate: 0.22},
		{Min: 103350, Max: 197300, Rate: 0.24},
		{Min: 197300, Max: 250500, Rate: 0.32},
		{Min: 250500, Max: 626350, Rate: 0.35},
		{Min: 626350, Max: math.MaxFloat64, Rate: 0.37},
	},
	QualifyingSurvivor: {
		{Min: 0, Max: 23850, Rate: 0.10},
		{Min: 23850, Max: 96950, Rate: 0.12},
		{Min: 96950, Max: 206700, Rate: 0.22},
		{Min: 206700, Max: 394600, Rate: 0.24},
		{Min: 394600, Max: 501050, Rate: 0.32},
		{Min: 501050, Max: 751600, Rate: 0.35},
		{Min: 751600, Max: math.MaxFloat64, Rate: 0.37},
	},
}

// ---------------------------------------------------------------------------
// Federal 2025 standard deductions
// ---------------------------------------------------------------------------

var federalStdDeduction2025 = map[FilingStatus]float64{
	Single:             15000,
	MarriedFilingJoint: 30000,
	MarriedFilingSep:   15000,
	HeadOfHousehold:    22500,
	QualifyingSurvivor: 30000,
}

// FederalAdditionalStdDeduction2025 holds extra deduction amounts for 65+/blind.
// Single/HOH: $2,000 each; MFJ/MFS/QSS: $1,600 each.
var FederalAdditionalStdDeduction2025 = map[FilingStatus]float64{
	Single:             2000,
	HeadOfHousehold:    2000,
	MarriedFilingJoint: 1600,
	MarriedFilingSep:   1600,
	QualifyingSurvivor: 1600,
}

// ---------------------------------------------------------------------------
// California 2025 brackets (FTB, indexed for inflation)
// ---------------------------------------------------------------------------

var caBrackets2025 = map[FilingStatus]BracketTable{
	Single: {
		{Min: 0, Max: 10756, Rate: 0.01},
		{Min: 10756, Max: 25499, Rate: 0.02},
		{Min: 25499, Max: 40245, Rate: 0.04},
		{Min: 40245, Max: 55866, Rate: 0.06},
		{Min: 55866, Max: 70612, Rate: 0.08},
		{Min: 70612, Max: 360659, Rate: 0.093},
		{Min: 360659, Max: 432791, Rate: 0.103},
		{Min: 432791, Max: 721319, Rate: 0.113},
		{Min: 721319, Max: math.MaxFloat64, Rate: 0.123},
	},
	MarriedFilingJoint: {
		{Min: 0, Max: 21512, Rate: 0.01},
		{Min: 21512, Max: 50998, Rate: 0.02},
		{Min: 50998, Max: 80490, Rate: 0.04},
		{Min: 80490, Max: 111732, Rate: 0.06},
		{Min: 111732, Max: 141224, Rate: 0.08},
		{Min: 141224, Max: 721318, Rate: 0.093},
		{Min: 721318, Max: 865582, Rate: 0.103},
		{Min: 865582, Max: 1442638, Rate: 0.113},
		{Min: 1442638, Max: math.MaxFloat64, Rate: 0.123},
	},
	MarriedFilingSep: {
		{Min: 0, Max: 10756, Rate: 0.01},
		{Min: 10756, Max: 25499, Rate: 0.02},
		{Min: 25499, Max: 40245, Rate: 0.04},
		{Min: 40245, Max: 55866, Rate: 0.06},
		{Min: 55866, Max: 70612, Rate: 0.08},
		{Min: 70612, Max: 360659, Rate: 0.093},
		{Min: 360659, Max: 432791, Rate: 0.103},
		{Min: 432791, Max: 721319, Rate: 0.113},
		{Min: 721319, Max: math.MaxFloat64, Rate: 0.123},
	},
	HeadOfHousehold: {
		{Min: 0, Max: 21512, Rate: 0.01},
		{Min: 21512, Max: 50998, Rate: 0.02},
		{Min: 50998, Max: 80490, Rate: 0.04},
		{Min: 80490, Max: 111732, Rate: 0.06},
		{Min: 111732, Max: 141224, Rate: 0.08},
		{Min: 141224, Max: 721318, Rate: 0.093},
		{Min: 721318, Max: 865582, Rate: 0.103},
		{Min: 865582, Max: 1442638, Rate: 0.113},
		{Min: 1442638, Max: math.MaxFloat64, Rate: 0.123},
	},
	QualifyingSurvivor: {
		{Min: 0, Max: 21512, Rate: 0.01},
		{Min: 21512, Max: 50998, Rate: 0.02},
		{Min: 50998, Max: 80490, Rate: 0.04},
		{Min: 80490, Max: 111732, Rate: 0.06},
		{Min: 111732, Max: 141224, Rate: 0.08},
		{Min: 141224, Max: 721318, Rate: 0.093},
		{Min: 721318, Max: 865582, Rate: 0.103},
		{Min: 865582, Max: 1442638, Rate: 0.113},
		{Min: 1442638, Max: math.MaxFloat64, Rate: 0.123},
	},
}

// ---------------------------------------------------------------------------
// California 2025 standard deductions
// ---------------------------------------------------------------------------

var caStdDeduction2025 = map[FilingStatus]float64{
	Single:             5706,
	MarriedFilingJoint: 11412,
	MarriedFilingSep:   5706,
	HeadOfHousehold:    11412,
	QualifyingSurvivor: 11412,
}

// ---------------------------------------------------------------------------
// California 2025 Mental Health Services Tax
// ---------------------------------------------------------------------------

const caMentalHealthThreshold = 1_000_000
const caMentalHealthRate = 0.01

// ---------------------------------------------------------------------------
// California 2025 Exemption Credits
// ---------------------------------------------------------------------------

var caExemptionCredit2025 = map[FilingStatus]float64{
	Single:             144,
	MarriedFilingJoint: 288,
	MarriedFilingSep:   144,
	HeadOfHousehold:    144,
	QualifyingSurvivor: 144,
}

const caExemptionCreditDependent2025 = 433.0

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

// GetBrackets returns the bracket table for a given year, jurisdiction, and filing status.
// Currently only 2025 is supported; other years return nil.
func GetBrackets(year int, jurisdiction JurisdictionType, status FilingStatus) BracketTable {
	if year != 2025 {
		return nil
	}
	switch jurisdiction {
	case Federal:
		return federalBrackets2025[status]
	case StateCA:
		return caBrackets2025[status]
	}
	return nil
}

// GetStandardDeduction returns the standard deduction for a given year, jurisdiction, and filing status.
func GetStandardDeduction(year int, jurisdiction JurisdictionType, status FilingStatus) float64 {
	if year != 2025 {
		return 0
	}
	switch jurisdiction {
	case Federal:
		return federalStdDeduction2025[status]
	case StateCA:
		return caStdDeduction2025[status]
	}
	return 0
}

// ComputeTax is the main entry point — computes tax for the given parameters.
// For CA, this includes the mental health services surcharge.
func ComputeTax(taxableIncome float64, status FilingStatus, year int, jurisdiction JurisdictionType) float64 {
	if taxableIncome <= 0 {
		return 0
	}
	brackets := GetBrackets(year, jurisdiction, status)
	if brackets == nil {
		return 0
	}
	tax := ComputeBracketTax(taxableIncome, brackets)
	if jurisdiction == StateCA {
		tax += GetCAMentalHealthTax(taxableIncome)
	}
	return tax
}

// GetCAMentalHealthTax computes the 1% surcharge on taxable income over $1M.
func GetCAMentalHealthTax(taxableIncome float64) float64 {
	if taxableIncome <= caMentalHealthThreshold {
		return 0
	}
	return (taxableIncome - caMentalHealthThreshold) * caMentalHealthRate
}

// GetCAExemptionCredit returns the exemption credit amount for the given status and number of dependents.
func GetCAExemptionCredit(year int, status FilingStatus, numDependents int) float64 {
	if year != 2025 {
		return 0
	}
	credit := caExemptionCredit2025[status]
	if numDependents > 0 {
		credit += float64(numDependents) * caExemptionCreditDependent2025
	}
	return credit
}
