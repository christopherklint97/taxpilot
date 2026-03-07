package interview

// Situation represents a detected tax situation that requires additional forms.
type Situation struct {
	ID          string   // e.g., "self_employed", "capital_gains"
	Label       string   // "Self-Employment Income"
	Description string   // "You had freelance or business income"
	FormsNeeded []string // form IDs that need to be added
	Screening   string   // the screening question that triggered this
}

// ScreeningQuestion is a yes/no question asked early to determine which forms are needed.
type ScreeningQuestion struct {
	ID       string
	Question string
	HelpText string
	CANote   string // CA-specific note
	OnYes    Situation
}

// DefaultScreeningQuestions returns the screening questions for the interview.
// These are asked after personal info but before form-specific questions.
var DefaultScreeningQuestions = []ScreeningQuestion{
	{
		ID:       "has_self_employment",
		Question: "Did you have any self-employment or freelance income in 2025?",
		HelpText: "This includes 1099-NEC income, gig work, freelancing, or running a business.",
		OnYes: Situation{
			ID:          "self_employed",
			Label:       "Self-Employment Income",
			Description: "You had freelance or business income",
			FormsNeeded: []string{"schedule_c", "schedule_se"},
			Screening:   "has_self_employment",
		},
	},
	{
		ID:       "has_capital_gains",
		Question: "Did you sell any stocks, bonds, mutual funds, or other investments in 2025?",
		HelpText: "This includes sales reported on a 1099-B from your brokerage.",
		CANote:   "California taxes capital gains as ordinary income (no preferential rate).",
		OnYes: Situation{
			ID:          "capital_gains",
			Label:       "Capital Gains/Losses",
			Description: "You sold investments during the year",
			FormsNeeded: []string{"schedule_d", "f8949"},
			Screening:   "has_capital_gains",
		},
	},
	// Rental income — Schedule E not yet implemented
	// {
	//     ID:       "has_rental_income",
	//     Question: "Did you receive any rental income from real estate in 2025?",
	//     HelpText: "This includes income from renting out a house, apartment, room, or other property.",
	//     OnYes: Situation{
	//         ID:          "rental_income",
	//         Label:       "Rental Income",
	//         Description: "You received rental income from real estate",
	//         FormsNeeded: []string{"schedule_e"},
	//         Screening:   "has_rental_income",
	//     },
	// },
	{
		ID:       "has_interest_income",
		Question: "Did you receive any interest income in 2025?",
		HelpText: "This includes interest from bank accounts, CDs, bonds, or other sources reported on a 1099-INT.",
		OnYes: Situation{
			ID:          "interest_income",
			Label:       "Interest Income",
			Description: "You received interest income",
			FormsNeeded: []string{"1099int"},
			Screening:   "has_interest_income",
		},
	},
	{
		ID:       "has_dividend_income",
		Question: "Did you receive any dividend income in 2025?",
		HelpText: "This includes dividends from stocks, mutual funds, or other investments reported on a 1099-DIV.",
		OnYes: Situation{
			ID:          "dividend_income",
			Label:       "Dividend Income",
			Description: "You received dividend income",
			FormsNeeded: []string{"1099div"},
			Screening:   "has_dividend_income",
		},
	},
	{
		ID:       "has_hsa",
		Question: "Do you have a Health Savings Account (HSA)?",
		HelpText: "If you contributed to or received distributions from an HSA, you need Form 8889.",
		CANote:   "California does not conform to federal HSA treatment — contributions are not deductible for CA.",
		OnYes: Situation{
			ID:          "hsa",
			Label:       "Health Savings Account",
			Description: "You have a Health Savings Account (HSA)",
			FormsNeeded: []string{"form_8889"},
			Screening:   "has_hsa",
		},
	},
	{
		ID:       "has_itemized_deductions",
		Question: "Do you want to itemize deductions instead of taking the standard deduction?",
		HelpText: "The 2025 standard deduction is $15,000 for single filers and $30,000 for married filing jointly. Itemize only if your deductions exceed these amounts.",
		CANote:   "California has its own standard deduction amounts ($5,540 single / $11,080 MFJ for 2025). You may benefit from itemizing on one return but not the other.",
		OnYes: Situation{
			ID:          "itemized_deductions",
			Label:       "Itemized Deductions",
			Description: "You want to itemize deductions on Schedule A",
			FormsNeeded: []string{"schedule_a"},
			Screening:   "has_itemized_deductions",
		},
	},
}

// EvaluateScreening checks screening answers and returns the situations that apply.
// answers maps screening question ID to true (yes) or false (no).
func EvaluateScreening(answers map[string]bool) []Situation {
	var situations []Situation
	for _, sq := range DefaultScreeningQuestions {
		if answers[sq.ID] {
			situations = append(situations, sq.OnYes)
		}
	}
	return situations
}

// PriorYearData holds data from a prior-year return used for auto-detection.
type PriorYearData struct {
	// FormsPresent lists form IDs that were in the prior-year return.
	FormsPresent []string
	// NumericValues holds numeric field values from the prior year.
	NumericValues map[string]float64
}

// AutoDetectSituations analyzes prior-year return data and returns
// screening question IDs that should default to "yes" based on what
// forms were present in the prior year.
func AutoDetectSituations(prior PriorYearData) map[string]bool {
	detected := make(map[string]bool)

	formSet := make(map[string]bool, len(prior.FormsPresent))
	for _, f := range prior.FormsPresent {
		formSet[f] = true
	}

	// Self-employment: had Schedule C or Schedule SE
	if formSet["schedule_c"] || formSet["schedule_se"] {
		detected["has_self_employment"] = true
	}

	// Capital gains: had Schedule D or Form 8949
	if formSet["schedule_d"] || formSet["f8949"] {
		detected["has_capital_gains"] = true
	}

	// Interest income: had 1099-INT or Schedule B with interest
	if formSet["1099int"] || formSet["schedule_b"] {
		detected["has_interest_income"] = true
	}

	// Dividend income: had 1099-DIV
	if formSet["1099div"] {
		detected["has_dividend_income"] = true
	}

	// HSA: had Form 8889
	if formSet["form_8889"] {
		detected["has_hsa"] = true
	}

	// Itemized deductions: had Schedule A
	if formSet["schedule_a"] {
		detected["has_itemized_deductions"] = true
	}

	// Also check numeric values for additional signals
	if prior.NumericValues != nil {
		// If they had HSA contributions last year
		if prior.NumericValues["form_8889:2"] > 0 {
			detected["has_hsa"] = true
		}
		// If they had Schedule A total deductions
		if prior.NumericValues["schedule_a:17"] > 0 {
			detected["has_itemized_deductions"] = true
		}
		// If they had self-employment income
		if prior.NumericValues["schedule_c:31"] > 0 || prior.NumericValues["schedule_se:4"] > 0 {
			detected["has_self_employment"] = true
		}
	}

	return detected
}
