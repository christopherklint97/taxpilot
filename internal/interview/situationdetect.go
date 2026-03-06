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

// DefaultScreeningQuestions returns the screening questions for the MVP.
// These are asked after personal info but before form-specific questions.
// Defined for Phase 5 when we add more forms. For now, documents the pattern.
var DefaultScreeningQuestions = []ScreeningQuestion{
	// {
	//     ID:       "has_self_employment",
	//     Question: "Did you have any self-employment or freelance income in 2025?",
	//     HelpText: "This includes 1099-NEC income, gig work, freelancing, or running a business.",
	//     OnYes: Situation{
	//         ID:          "self_employed",
	//         Label:       "Self-Employment Income",
	//         Description: "You had freelance or business income",
	//         FormsNeeded: []string{"schedule_c", "schedule_se"},
	//         Screening:   "has_self_employment",
	//     },
	// },
	// {
	//     ID:       "has_capital_gains",
	//     Question: "Did you sell any stocks, bonds, mutual funds, or other investments in 2025?",
	//     HelpText: "This includes sales reported on a 1099-B from your brokerage.",
	//     CANote:   "California taxes capital gains as ordinary income (no preferential rate).",
	//     OnYes: Situation{
	//         ID:          "capital_gains",
	//         Label:       "Capital Gains/Losses",
	//         Description: "You sold investments during the year",
	//         FormsNeeded: []string{"schedule_d", "f8949"},
	//         Screening:   "has_capital_gains",
	//     },
	// },
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
