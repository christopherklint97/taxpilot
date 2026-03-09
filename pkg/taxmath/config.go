package taxmath

// TaxYearConfig holds all year-specific tax constants in one place.
// This centralizes values that were previously scattered across multiple files.
type TaxYearConfig struct {
	Year int

	// HSA contribution limits
	HSALimitSelfOnly float64
	HSALimitFamily   float64

	// Capital loss deduction limits (negative values)
	CapitalLossLimit    float64
	CapitalLossLimitMFS float64

	// Effective tax rate threshold (highest bracket rate)
	MaxEffectiveTaxRate float64

	// SALT deduction cap
	SALTCap    float64
	SALTCapMFS float64

	// FEIE
	FEIEExclusionLimit     float64
	PhysicalPresenceMinDays int

	// FBAR
	FBARThreshold float64

	// FATCA thresholds
	FATCAAbroadSingleYearEnd float64
	FATCAAbroadSingleAnyTime float64
	FATCAAbroadMFJYearEnd    float64
	FATCAAbroadMFJAnyTime    float64
	FATCAUSSingleYearEnd     float64
	FATCAUSSingleAnyTime     float64
	FATCAUSMFJYearEnd        float64
	FATCAUSMFJAnyTime        float64

	// CA-specific
	CAMaxMarginalRate      float64
	CAMentalHealthRate     float64
	CAMentalHealthThreshold float64
}

// configs stores all known tax year configurations.
var configs = map[int]*TaxYearConfig{
	2025: {
		Year: 2025,

		HSALimitSelfOnly: 4300,
		HSALimitFamily:   8550,

		CapitalLossLimit:    -3000,
		CapitalLossLimitMFS: -1500,

		MaxEffectiveTaxRate: 0.37,

		SALTCap:    10000,
		SALTCapMFS: 5000,

		FEIEExclusionLimit:      130000,
		PhysicalPresenceMinDays: 330,

		FBARThreshold: 10000,

		FATCAAbroadSingleYearEnd: 200000,
		FATCAAbroadSingleAnyTime: 300000,
		FATCAAbroadMFJYearEnd:    400000,
		FATCAAbroadMFJAnyTime:    600000,
		FATCAUSSingleYearEnd:     50000,
		FATCAUSSingleAnyTime:     75000,
		FATCAUSMFJYearEnd:        100000,
		FATCAUSMFJAnyTime:        150000,

		CAMaxMarginalRate:       0.133,
		CAMentalHealthRate:      0.01,
		CAMentalHealthThreshold: 1_000_000,
	},
}

// GetConfig returns the TaxYearConfig for the given year, or nil if not found.
func GetConfig(year int) *TaxYearConfig {
	return configs[year]
}

// GetConfigOrDefault returns the TaxYearConfig for the given year.
// If the year is not found, returns the most recent known config.
func GetConfigOrDefault(year int) *TaxYearConfig {
	if c := configs[year]; c != nil {
		return c
	}
	// Fall back to 2025
	return configs[2025]
}
