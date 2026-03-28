package taxmath

import (
	"fmt"
	"math"
)

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
// Federal 2024 brackets (IRS Rev. Proc. 2023-34)
// ---------------------------------------------------------------------------

var federalBrackets2024 = map[FilingStatus]BracketTable{
	Single: {
		{Min: 0, Max: 11600, Rate: 0.10},
		{Min: 11600, Max: 47150, Rate: 0.12},
		{Min: 47150, Max: 100525, Rate: 0.22},
		{Min: 100525, Max: 191950, Rate: 0.24},
		{Min: 191950, Max: 243725, Rate: 0.32},
		{Min: 243725, Max: 609350, Rate: 0.35},
		{Min: 609350, Max: math.MaxFloat64, Rate: 0.37},
	},
	MarriedFilingJoint: {
		{Min: 0, Max: 23200, Rate: 0.10},
		{Min: 23200, Max: 94300, Rate: 0.12},
		{Min: 94300, Max: 201050, Rate: 0.22},
		{Min: 201050, Max: 383900, Rate: 0.24},
		{Min: 383900, Max: 487450, Rate: 0.32},
		{Min: 487450, Max: 731200, Rate: 0.35},
		{Min: 731200, Max: math.MaxFloat64, Rate: 0.37},
	},
	MarriedFilingSep: {
		{Min: 0, Max: 11600, Rate: 0.10},
		{Min: 11600, Max: 47150, Rate: 0.12},
		{Min: 47150, Max: 100525, Rate: 0.22},
		{Min: 100525, Max: 191950, Rate: 0.24},
		{Min: 191950, Max: 243725, Rate: 0.32},
		{Min: 243725, Max: 365600, Rate: 0.35},
		{Min: 365600, Max: math.MaxFloat64, Rate: 0.37},
	},
	HeadOfHousehold: {
		{Min: 0, Max: 16550, Rate: 0.10},
		{Min: 16550, Max: 63100, Rate: 0.12},
		{Min: 63100, Max: 100525, Rate: 0.22},
		{Min: 100525, Max: 191950, Rate: 0.24},
		{Min: 191950, Max: 243700, Rate: 0.32},
		{Min: 243700, Max: 609350, Rate: 0.35},
		{Min: 609350, Max: math.MaxFloat64, Rate: 0.37},
	},
	QualifyingSurvivor: {
		{Min: 0, Max: 23200, Rate: 0.10},
		{Min: 23200, Max: 94300, Rate: 0.12},
		{Min: 94300, Max: 201050, Rate: 0.22},
		{Min: 201050, Max: 383900, Rate: 0.24},
		{Min: 383900, Max: 487450, Rate: 0.32},
		{Min: 487450, Max: 731200, Rate: 0.35},
		{Min: 731200, Max: math.MaxFloat64, Rate: 0.37},
	},
}

// ---------------------------------------------------------------------------
// Federal 2024 standard deductions
// ---------------------------------------------------------------------------

var federalStdDeduction2024 = map[FilingStatus]float64{
	Single:             14600,
	MarriedFilingJoint: 29200,
	MarriedFilingSep:   14600,
	HeadOfHousehold:    21900,
	QualifyingSurvivor: 29200,
}

// FederalAdditionalStdDeduction2024 holds extra deduction amounts for 65+/blind.
var FederalAdditionalStdDeduction2024 = map[FilingStatus]float64{
	Single:             1950,
	HeadOfHousehold:    1950,
	MarriedFilingJoint: 1550,
	MarriedFilingSep:   1550,
	QualifyingSurvivor: 1550,
}

// ---------------------------------------------------------------------------
// California 2024 brackets (FTB, indexed for inflation)
// ---------------------------------------------------------------------------

var caBrackets2024 = map[FilingStatus]BracketTable{
	Single: {
		{Min: 0, Max: 10412, Rate: 0.01},
		{Min: 10412, Max: 24684, Rate: 0.02},
		{Min: 24684, Max: 38959, Rate: 0.04},
		{Min: 38959, Max: 54081, Rate: 0.06},
		{Min: 54081, Max: 68350, Rate: 0.08},
		{Min: 68350, Max: 349137, Rate: 0.093},
		{Min: 349137, Max: 418961, Rate: 0.103},
		{Min: 418961, Max: 698271, Rate: 0.113},
		{Min: 698271, Max: math.MaxFloat64, Rate: 0.123},
	},
	MarriedFilingJoint: {
		{Min: 0, Max: 20824, Rate: 0.01},
		{Min: 20824, Max: 49368, Rate: 0.02},
		{Min: 49368, Max: 77918, Rate: 0.04},
		{Min: 77918, Max: 108162, Rate: 0.06},
		{Min: 108162, Max: 136700, Rate: 0.08},
		{Min: 136700, Max: 698274, Rate: 0.093},
		{Min: 698274, Max: 837922, Rate: 0.103},
		{Min: 837922, Max: 1396542, Rate: 0.113},
		{Min: 1396542, Max: math.MaxFloat64, Rate: 0.123},
	},
	MarriedFilingSep: {
		{Min: 0, Max: 10412, Rate: 0.01},
		{Min: 10412, Max: 24684, Rate: 0.02},
		{Min: 24684, Max: 38959, Rate: 0.04},
		{Min: 38959, Max: 54081, Rate: 0.06},
		{Min: 54081, Max: 68350, Rate: 0.08},
		{Min: 68350, Max: 349137, Rate: 0.093},
		{Min: 349137, Max: 418961, Rate: 0.103},
		{Min: 418961, Max: 698271, Rate: 0.113},
		{Min: 698271, Max: math.MaxFloat64, Rate: 0.123},
	},
	HeadOfHousehold: {
		{Min: 0, Max: 20824, Rate: 0.01},
		{Min: 20824, Max: 49368, Rate: 0.02},
		{Min: 49368, Max: 77918, Rate: 0.04},
		{Min: 77918, Max: 108162, Rate: 0.06},
		{Min: 108162, Max: 136700, Rate: 0.08},
		{Min: 136700, Max: 698274, Rate: 0.093},
		{Min: 698274, Max: 837922, Rate: 0.103},
		{Min: 837922, Max: 1396542, Rate: 0.113},
		{Min: 1396542, Max: math.MaxFloat64, Rate: 0.123},
	},
	QualifyingSurvivor: {
		{Min: 0, Max: 20824, Rate: 0.01},
		{Min: 20824, Max: 49368, Rate: 0.02},
		{Min: 49368, Max: 77918, Rate: 0.04},
		{Min: 77918, Max: 108162, Rate: 0.06},
		{Min: 108162, Max: 136700, Rate: 0.08},
		{Min: 136700, Max: 698274, Rate: 0.093},
		{Min: 698274, Max: 837922, Rate: 0.103},
		{Min: 837922, Max: 1396542, Rate: 0.113},
		{Min: 1396542, Max: math.MaxFloat64, Rate: 0.123},
	},
}

// ---------------------------------------------------------------------------
// California 2024 standard deductions
// ---------------------------------------------------------------------------

var caStdDeduction2024 = map[FilingStatus]float64{
	Single:             5540,
	MarriedFilingJoint: 11080,
	MarriedFilingSep:   5540,
	HeadOfHousehold:    11080,
	QualifyingSurvivor: 11080,
}

// ---------------------------------------------------------------------------
// California 2024 Exemption Credits
// ---------------------------------------------------------------------------

var caExemptionCredit2024 = map[FilingStatus]float64{
	Single:             140,
	MarriedFilingJoint: 280,
	MarriedFilingSep:   140,
	HeadOfHousehold:    140,
	QualifyingSurvivor: 140,
}

const caExemptionCreditDependent2024 = 421.0

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
// Federal 2026 brackets (projected — ~2.8% CPI adjustment from 2025)
// ---------------------------------------------------------------------------

var federalBrackets2026 = map[FilingStatus]BracketTable{
	Single: {
		{Min: 0, Max: 12250, Rate: 0.10},
		{Min: 12250, Max: 49825, Rate: 0.12},
		{Min: 49825, Max: 106250, Rate: 0.22},
		{Min: 106250, Max: 202850, Rate: 0.24},
		{Min: 202850, Max: 257550, Rate: 0.32},
		{Min: 257550, Max: 643900, Rate: 0.35},
		{Min: 643900, Max: math.MaxFloat64, Rate: 0.37},
	},
	MarriedFilingJoint: {
		{Min: 0, Max: 24500, Rate: 0.10},
		{Min: 24500, Max: 99700, Rate: 0.12},
		{Min: 99700, Max: 212500, Rate: 0.22},
		{Min: 212500, Max: 405650, Rate: 0.24},
		{Min: 405650, Max: 515100, Rate: 0.32},
		{Min: 515100, Max: 772650, Rate: 0.35},
		{Min: 772650, Max: math.MaxFloat64, Rate: 0.37},
	},
	MarriedFilingSep: {
		{Min: 0, Max: 12250, Rate: 0.10},
		{Min: 12250, Max: 49825, Rate: 0.12},
		{Min: 49825, Max: 106250, Rate: 0.22},
		{Min: 106250, Max: 202850, Rate: 0.24},
		{Min: 202850, Max: 257550, Rate: 0.32},
		{Min: 257550, Max: 386325, Rate: 0.35},
		{Min: 386325, Max: math.MaxFloat64, Rate: 0.37},
	},
	HeadOfHousehold: {
		{Min: 0, Max: 17475, Rate: 0.10},
		{Min: 17475, Max: 66675, Rate: 0.12},
		{Min: 66675, Max: 106250, Rate: 0.22},
		{Min: 106250, Max: 202850, Rate: 0.24},
		{Min: 202850, Max: 257500, Rate: 0.32},
		{Min: 257500, Max: 643900, Rate: 0.35},
		{Min: 643900, Max: math.MaxFloat64, Rate: 0.37},
	},
	QualifyingSurvivor: {
		{Min: 0, Max: 24500, Rate: 0.10},
		{Min: 24500, Max: 99700, Rate: 0.12},
		{Min: 99700, Max: 212500, Rate: 0.22},
		{Min: 212500, Max: 405650, Rate: 0.24},
		{Min: 405650, Max: 515100, Rate: 0.32},
		{Min: 515100, Max: 772650, Rate: 0.35},
		{Min: 772650, Max: math.MaxFloat64, Rate: 0.37},
	},
}

// ---------------------------------------------------------------------------
// Federal 2026 standard deductions (projected)
// ---------------------------------------------------------------------------

var federalStdDeduction2026 = map[FilingStatus]float64{
	Single:             15400,
	MarriedFilingJoint: 30800,
	MarriedFilingSep:   15400,
	HeadOfHousehold:    23100,
	QualifyingSurvivor: 30800,
}

// FederalAdditionalStdDeduction2026 holds extra deduction amounts for 65+/blind.
var FederalAdditionalStdDeduction2026 = map[FilingStatus]float64{
	Single:             2050,
	HeadOfHousehold:    2050,
	MarriedFilingJoint: 1650,
	MarriedFilingSep:   1650,
	QualifyingSurvivor: 1650,
}

// ---------------------------------------------------------------------------
// California 2026 brackets (projected — ~2.8% CPI adjustment from 2025)
// ---------------------------------------------------------------------------

var caBrackets2026 = map[FilingStatus]BracketTable{
	Single: {
		{Min: 0, Max: 11057, Rate: 0.01},
		{Min: 11057, Max: 26213, Rate: 0.02},
		{Min: 26213, Max: 41372, Rate: 0.04},
		{Min: 41372, Max: 57430, Rate: 0.06},
		{Min: 57430, Max: 72589, Rate: 0.08},
		{Min: 72589, Max: 370758, Rate: 0.093},
		{Min: 370758, Max: 444909, Rate: 0.103},
		{Min: 444909, Max: 741516, Rate: 0.113},
		{Min: 741516, Max: math.MaxFloat64, Rate: 0.123},
	},
	MarriedFilingJoint: {
		{Min: 0, Max: 22114, Rate: 0.01},
		{Min: 22114, Max: 52426, Rate: 0.02},
		{Min: 52426, Max: 82744, Rate: 0.04},
		{Min: 82744, Max: 114860, Rate: 0.06},
		{Min: 114860, Max: 145178, Rate: 0.08},
		{Min: 145178, Max: 741515, Rate: 0.093},
		{Min: 741515, Max: 889818, Rate: 0.103},
		{Min: 889818, Max: 1483072, Rate: 0.113},
		{Min: 1483072, Max: math.MaxFloat64, Rate: 0.123},
	},
	MarriedFilingSep: {
		{Min: 0, Max: 11057, Rate: 0.01},
		{Min: 11057, Max: 26213, Rate: 0.02},
		{Min: 26213, Max: 41372, Rate: 0.04},
		{Min: 41372, Max: 57430, Rate: 0.06},
		{Min: 57430, Max: 72589, Rate: 0.08},
		{Min: 72589, Max: 370758, Rate: 0.093},
		{Min: 370758, Max: 444909, Rate: 0.103},
		{Min: 444909, Max: 741516, Rate: 0.113},
		{Min: 741516, Max: math.MaxFloat64, Rate: 0.123},
	},
	HeadOfHousehold: {
		{Min: 0, Max: 22114, Rate: 0.01},
		{Min: 22114, Max: 52426, Rate: 0.02},
		{Min: 52426, Max: 82744, Rate: 0.04},
		{Min: 82744, Max: 114860, Rate: 0.06},
		{Min: 114860, Max: 145178, Rate: 0.08},
		{Min: 145178, Max: 741515, Rate: 0.093},
		{Min: 741515, Max: 889818, Rate: 0.103},
		{Min: 889818, Max: 1483072, Rate: 0.113},
		{Min: 1483072, Max: math.MaxFloat64, Rate: 0.123},
	},
	QualifyingSurvivor: {
		{Min: 0, Max: 22114, Rate: 0.01},
		{Min: 22114, Max: 52426, Rate: 0.02},
		{Min: 52426, Max: 82744, Rate: 0.04},
		{Min: 82744, Max: 114860, Rate: 0.06},
		{Min: 114860, Max: 145178, Rate: 0.08},
		{Min: 145178, Max: 741515, Rate: 0.093},
		{Min: 741515, Max: 889818, Rate: 0.103},
		{Min: 889818, Max: 1483072, Rate: 0.113},
		{Min: 1483072, Max: math.MaxFloat64, Rate: 0.123},
	},
}

// ---------------------------------------------------------------------------
// California 2026 standard deductions (projected)
// ---------------------------------------------------------------------------

var caStdDeduction2026 = map[FilingStatus]float64{
	Single:             5866,
	MarriedFilingJoint: 11732,
	MarriedFilingSep:   5866,
	HeadOfHousehold:    11732,
	QualifyingSurvivor: 11732,
}

// ---------------------------------------------------------------------------
// California 2026 Exemption Credits (projected)
// ---------------------------------------------------------------------------

var caExemptionCredit2026 = map[FilingStatus]float64{
	Single:             148,
	MarriedFilingJoint: 296,
	MarriedFilingSep:   148,
	HeadOfHousehold:    148,
	QualifyingSurvivor: 148,
}

const caExemptionCreditDependent2026 = 445.0

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

// GetBrackets returns the bracket table for a given year, jurisdiction, and filing status.
// Currently only 2025 is supported; other years return nil.
func GetBrackets(year int, jurisdiction JurisdictionType, status FilingStatus) BracketTable {
	switch year {
	case 2024:
		switch jurisdiction {
		case Federal:
			return federalBrackets2024[status]
		case StateCA:
			return caBrackets2024[status]
		}
	case 2025:
		switch jurisdiction {
		case Federal:
			return federalBrackets2025[status]
		case StateCA:
			return caBrackets2025[status]
		}
	case 2026:
		switch jurisdiction {
		case Federal:
			return federalBrackets2026[status]
		case StateCA:
			return caBrackets2026[status]
		}
	}
	return nil
}

// GetStandardDeduction returns the standard deduction for a given year, jurisdiction, and filing status.
func GetStandardDeduction(year int, jurisdiction JurisdictionType, status FilingStatus) float64 {
	switch year {
	case 2024:
		switch jurisdiction {
		case Federal:
			return federalStdDeduction2024[status]
		case StateCA:
			return caStdDeduction2024[status]
		}
	case 2025:
		switch jurisdiction {
		case Federal:
			return federalStdDeduction2025[status]
		case StateCA:
			return caStdDeduction2025[status]
		}
	case 2026:
		switch jurisdiction {
		case Federal:
			return federalStdDeduction2026[status]
		case StateCA:
			return caStdDeduction2026[status]
		}
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
	switch year {
	case 2024:
		credit := caExemptionCredit2024[status]
		if numDependents > 0 {
			credit += float64(numDependents) * caExemptionCreditDependent2024
		}
		return credit
	case 2025:
		credit := caExemptionCredit2025[status]
		if numDependents > 0 {
			credit += float64(numDependents) * caExemptionCreditDependent2025
		}
		return credit
	case 2026:
		credit := caExemptionCredit2026[status]
		if numDependents > 0 {
			credit += float64(numDependents) * caExemptionCreditDependent2026
		}
		return credit
	}
	return 0
}

// ParameterChange describes a change between two tax years.
type ParameterChange struct {
	Name     string
	OldValue float64
	NewValue float64
	Delta    float64
	Category string // "bracket", "deduction", "credit", "limit"
}

// CompareYearParameters returns the differences between two tax years for a given filing status.
func CompareYearParameters(priorYear, newYear int, status FilingStatus) []ParameterChange {
	var changes []ParameterChange

	// Standard deductions
	oldFedStd := GetStandardDeduction(priorYear, Federal, status)
	newFedStd := GetStandardDeduction(newYear, Federal, status)
	if oldFedStd != newFedStd {
		changes = append(changes, ParameterChange{
			Name: "Federal standard deduction", OldValue: oldFedStd, NewValue: newFedStd,
			Delta: newFedStd - oldFedStd, Category: "deduction",
		})
	}

	oldCAStd := GetStandardDeduction(priorYear, StateCA, status)
	newCAStd := GetStandardDeduction(newYear, StateCA, status)
	if oldCAStd != newCAStd {
		changes = append(changes, ParameterChange{
			Name: "CA standard deduction", OldValue: oldCAStd, NewValue: newCAStd,
			Delta: newCAStd - oldCAStd, Category: "deduction",
		})
	}

	// Compare bracket thresholds
	oldFedBrackets := GetBrackets(priorYear, Federal, status)
	newFedBrackets := GetBrackets(newYear, Federal, status)
	if len(oldFedBrackets) == len(newFedBrackets) {
		for i := range oldFedBrackets {
			if oldFedBrackets[i].Max != newFedBrackets[i].Max && oldFedBrackets[i].Max != math.MaxFloat64 {
				changes = append(changes, ParameterChange{
					Name:     fmt.Sprintf("Federal %v%% bracket ceiling", oldFedBrackets[i].Rate*100),
					OldValue: oldFedBrackets[i].Max, NewValue: newFedBrackets[i].Max,
					Delta: newFedBrackets[i].Max - oldFedBrackets[i].Max, Category: "bracket",
				})
			}
		}
	}

	return changes
}
