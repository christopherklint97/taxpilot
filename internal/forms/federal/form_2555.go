package federal

import (
	"math"

	"taxpilot/internal/forms"
	"taxpilot/pkg/taxmath"
)

func init() { forms.RegisterForm(Form2555) }

// Form2555 returns the FormDef for Form 2555 — Foreign Earned Income.
// This computes the Foreign Earned Income Exclusion (FEIE) and foreign
// housing exclusion/deduction for US citizens and residents living abroad.
//
// The total exclusion flows to Schedule 1 line 8d → Form 1040 line 8.
// When FEIE is claimed, Form 1040 line 16 must use tax stacking
// (ComputeTaxWithStacking) to prevent bracket manipulation.
//
// CA does NOT conform to the FEIE — the entire exclusion must be
// added back on Schedule CA.
func Form2555() *forms.FormDef {
	return &forms.FormDef{
		ID:           forms.FormF2555,
		Name:         "Form 2555 — Foreign Earned Income",
		Jurisdiction: forms.Federal,
		TaxYears:      []int{2025},
		QuestionGroup: "expat",
		QuestionOrder: 4,
		Fields: []forms.FieldDef{
			// --- Part I: General Information ---

			// Foreign country of residence
			{
				Line:   "foreign_country",
				Type:   forms.UserInput,
				ValueType: forms.StringValue,
				Label:  "Foreign country of residence",
				Prompt: "What country do you live in?",
			},
			// Foreign address
			{
				Line:   "foreign_address",
				Type:   forms.UserInput,
				ValueType: forms.StringValue,
				Label:  "Foreign address",
				Prompt: "What is your foreign address?",
			},
			// Foreign employer name
			{
				Line:   "employer_name_2555",
				Type:   forms.UserInput,
				ValueType: forms.StringValue,
				Label:  "Employer name (foreign)",
				Prompt: "What is your foreign employer's name?",
			},
			// Is employer foreign?
			{
				Line:    "employer_foreign",
				Type:    forms.UserInput,
				ValueType: forms.StringValue,
				Label:   "Is your employer a foreign entity?",
				Prompt:  "Is your employer a foreign (non-US) entity?",
				Options: []string{"yes", "no"},
			},
			// Self-employed abroad?
			{
				Line:    "self_employed_abroad",
				Type:    forms.UserInput,
				ValueType: forms.StringValue,
				Label:   "Self-employed abroad",
				Prompt:  "Are you self-employed in your foreign country?",
				Options: []string{"yes", "no"},
			},

			// --- Part II: Qualifying Test ---

			// Which qualifying test
			{
				Line:    "qualifying_test",
				Type:    forms.UserInput,
				ValueType: forms.StringValue,
				Label:   "Qualifying test for FEIE",
				Prompt:  "Which qualifying test do you meet for the Foreign Earned Income Exclusion?",
				Options: []string{"bona_fide_residence", "physical_presence"},
			},

			// --- Bona Fide Residence Test fields ---

			// BFRT start date
			{
				Line:   "bfrt_start_date",
				Type:   forms.UserInput,
				ValueType: forms.StringValue,
				Label:  "Bona fide residence start date",
				Prompt: "When did your bona fide residence in a foreign country begin? (MM/DD/YYYY)",
			},
			// BFRT end date
			{
				Line:   "bfrt_end_date",
				Type:   forms.UserInput,
				ValueType: forms.StringValue,
				Label:  "Bona fide residence end date",
				Prompt: "When did (or will) your bona fide residence end? (MM/DD/YYYY or 'continuing')",
			},
			// BFRT full tax year?
			{
				Line:    "bfrt_full_year",
				Type:    forms.UserInput,
				ValueType: forms.StringValue,
				Label:   "Bona fide resident for full tax year",
				Prompt:  "Were you a bona fide resident of a foreign country for the entire tax year?",
				Options: []string{"yes", "no"},
			},

			// --- Physical Presence Test fields ---

			// PPT days in foreign country
			{
				Line:   "ppt_days_present",
				Type:   forms.UserInput,
				Label:  "Days physically present in foreign country",
				Prompt: "How many days were you physically present in a foreign country during the 12-month qualifying period?",
			},
			// PPT period start
			{
				Line:   "ppt_period_start",
				Type:   forms.UserInput,
				ValueType: forms.StringValue,
				Label:  "Physical presence period start",
				Prompt: "What is the start date of your 12-month qualifying period? (MM/DD/YYYY)",
			},
			// PPT period end
			{
				Line:   "ppt_period_end",
				Type:   forms.UserInput,
				ValueType: forms.StringValue,
				Label:  "Physical presence period end",
				Prompt: "What is the end date of your 12-month qualifying period? (MM/DD/YYYY)",
			},

			// --- Part III: Foreign Earned Income ---

			// Foreign earned income
			{
				Line:   "foreign_earned_income",
				Type:   forms.UserInput,
				Label:  "Total foreign earned income (in USD)",
				Prompt: "What was your total foreign earned income for 2025 (converted to USD)?",
			},
			// Currency code
			{
				Line:   "currency_code",
				Type:   forms.UserInput,
				ValueType: forms.StringValue,
				Label:  "Foreign currency code",
				Prompt: "What currency were you paid in? (e.g., SEK, EUR, GBP)",
			},
			// Exchange rate
			{
				Line:   "exchange_rate",
				Type:   forms.UserInput,
				Label:  "Average exchange rate (foreign currency per USD)",
				Prompt: "What average exchange rate did you use to convert to USD? (IRS yearly average recommended)",
			},
			// Foreign taxes paid
			{
				Line:   "foreign_tax_paid",
				Type:   forms.UserInput,
				Label:  "Foreign income taxes paid (in USD)",
				Prompt: "How much foreign income tax did you pay in 2025 (converted to USD)?",
			},

			// --- Part IV: Housing ---

			// Employer-provided housing
			{
				Line:   "employer_provided_housing",
				Type:   forms.UserInput,
				Label:  "Employer-provided housing amounts",
				Prompt: "How much did your employer provide for housing (included in income)?",
			},
			// Housing expenses
			{
				Line:   "housing_expenses",
				Type:   forms.UserInput,
				Label:  "Qualifying foreign housing expenses",
				Prompt: "What were your total qualifying foreign housing expenses (rent, utilities, insurance, etc.)?",
			},

			// --- Computed Fields ---

			// FEIE exclusion limit for tax year
			{
				Line:      "exclusion_limit",
				Type:      forms.Computed,
				Label:     "Foreign earned income exclusion limit",
				DependsOn: []string{},
				Compute: func(dv forms.DepValues) float64 {
					return taxmath.FEIELimit(dv.TaxYear())
				},
			},
			// Qualifying days (365 for BFRT full year, PPT days otherwise)
			{
				Line:      "qualifying_days",
				Type:      forms.Computed,
				Label:     "Qualifying days",
				DependsOn: []string{"form_2555:qualifying_test", "form_2555:bfrt_full_year", "form_2555:ppt_days_present"},
				Compute: func(dv forms.DepValues) float64 {
					test := dv.GetString("form_2555:qualifying_test")
					if test == "bona_fide_residence" {
						fullYear := dv.GetString("form_2555:bfrt_full_year")
						if fullYear == "yes" {
							return 365
						}
						// Partial year BFRT: use PPT days as fallback
						days := dv.Get("form_2555:ppt_days_present")
						if days > 0 {
							return days
						}
						return 365 // default to full year if not specified
					}
					// Physical presence test
					return dv.Get("form_2555:ppt_days_present")
				},
			},
			// Prorated exclusion
			{
				Line:      "prorated_exclusion",
				Type:      forms.Computed,
				Label:     "Prorated exclusion amount",
				DependsOn: []string{"form_2555:exclusion_limit", "form_2555:qualifying_days"},
				Compute: func(dv forms.DepValues) float64 {
					limit := dv.Get("form_2555:exclusion_limit")
					days := int(dv.Get("form_2555:qualifying_days"))
					return taxmath.ProrateExclusion(limit, days, 365)
				},
			},
			// Foreign income exclusion (the actual amount excluded)
			{
				Line:      "foreign_income_exclusion",
				Type:      forms.Computed,
				Label:     "Foreign earned income exclusion",
				DependsOn: []string{"form_2555:foreign_earned_income", "form_2555:prorated_exclusion"},
				Compute: func(dv forms.DepValues) float64 {
					income := dv.Get("form_2555:foreign_earned_income")
					limit := dv.Get("form_2555:prorated_exclusion")
					return math.Min(income, limit)
				},
			},

			// --- Housing computation ---

			// Housing base amount (16% of FEIE limit)
			{
				Line:      "housing_base_amount",
				Type:      forms.Computed,
				Label:     "Housing base amount (16% of exclusion limit)",
				DependsOn: []string{},
				Compute: func(dv forms.DepValues) float64 {
					return taxmath.HousingBaseAmount(dv.TaxYear())
				},
			},
			// Housing max (30% of FEIE limit)
			{
				Line:      "housing_max",
				Type:      forms.Computed,
				Label:     "Maximum housing amount (30% of exclusion limit)",
				DependsOn: []string{},
				Compute: func(dv forms.DepValues) float64 {
					return taxmath.HousingMaxAmount(dv.TaxYear())
				},
			},
			// Housing qualifying amount
			{
				Line:      "housing_qualifying_amount",
				Type:      forms.Computed,
				Label:     "Qualifying housing amount",
				DependsOn: []string{"form_2555:housing_expenses", "form_2555:housing_max", "form_2555:housing_base_amount"},
				Compute: func(dv forms.DepValues) float64 {
					expenses := dv.Get("form_2555:housing_expenses")
					maxAmt := dv.Get("form_2555:housing_max")
					baseAmt := dv.Get("form_2555:housing_base_amount")
					limited := math.Min(expenses, maxAmt)
					return math.Max(0, limited-baseAmt)
				},
			},
			// Housing exclusion (for employees — limited by employer-provided amount)
			{
				Line:      "housing_exclusion",
				Type:      forms.Computed,
				Label:     "Foreign housing exclusion",
				DependsOn: []string{"form_2555:housing_qualifying_amount", "form_2555:employer_provided_housing"},
				Compute: func(dv forms.DepValues) float64 {
					qualifying := dv.Get("form_2555:housing_qualifying_amount")
					employerProvided := dv.Get("form_2555:employer_provided_housing")
					return math.Min(qualifying, employerProvided)
				},
			},
			// Housing deduction (for self-employed — excess over housing exclusion)
			{
				Line:      "housing_deduction",
				Type:      forms.Computed,
				Label:     "Foreign housing deduction (self-employed)",
				DependsOn: []string{"form_2555:housing_qualifying_amount", "form_2555:housing_exclusion", "form_2555:self_employed_abroad"},
				Compute: func(dv forms.DepValues) float64 {
					selfEmployed := dv.GetString("form_2555:self_employed_abroad")
					if selfEmployed != "yes" {
						return 0
					}
					qualifying := dv.Get("form_2555:housing_qualifying_amount")
					exclusion := dv.Get("form_2555:housing_exclusion")
					return math.Max(0, qualifying-exclusion)
				},
			},

			// --- Total exclusion ---

			// Total exclusion (FEIE + housing exclusion)
			{
				Line:      "total_exclusion",
				Type:      forms.Computed,
				Label:     "Total foreign earned income and housing exclusion",
				DependsOn: []string{"form_2555:foreign_income_exclusion", "form_2555:housing_exclusion"},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get("form_2555:foreign_income_exclusion") +
						dv.Get("form_2555:housing_exclusion")
				},
			},
		},
	}
}
