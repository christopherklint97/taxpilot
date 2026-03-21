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
		QuestionGroup: forms.GroupExpat,
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
				Options: forms.YesNoOptions,
			},
			// Self-employed abroad?
			{
				Line:    "self_employed_abroad",
				Type:    forms.UserInput,
				ValueType: forms.StringValue,
				Label:   "Self-employed abroad",
				Prompt:  "Are you self-employed in your foreign country?",
				Options: forms.YesNoOptions,
			},

			// --- Part II: Qualifying Test ---

			// Which qualifying test
			{
				Line:    "qualifying_test",
				Type:    forms.UserInput,
				ValueType: forms.StringValue,
				Label:   "Qualifying test for FEIE",
				Prompt:  "Which qualifying test do you meet for the Foreign Earned Income Exclusion?",
				Options: forms.QualifyingTestOptions,
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
				Options: forms.YesNoOptions,
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
				DependsOn: []string{forms.F2555QualifyingTest, forms.F2555BFRTFullYear, forms.F2555PPTDaysPresent},
				Compute: func(dv forms.DepValues) float64 {
					test := dv.GetString(forms.F2555QualifyingTest)
					if test == forms.QualifyingTestBFRT {
						fullYear := dv.GetString(forms.F2555BFRTFullYear)
						if fullYear == forms.OptionYes {
							return 365
						}
						// Partial year BFRT: use PPT days as fallback
						days := dv.Get(forms.F2555PPTDaysPresent)
						if days > 0 {
							return days
						}
						return 365 // default to full year if not specified
					}
					// Physical presence test
					return dv.Get(forms.F2555PPTDaysPresent)
				},
			},
			// Prorated exclusion
			{
				Line:      "prorated_exclusion",
				Type:      forms.Computed,
				Label:     "Prorated exclusion amount",
				DependsOn: []string{forms.F2555ExclusionLimit, forms.F2555QualifyingDays},
				Compute: func(dv forms.DepValues) float64 {
					limit := dv.Get(forms.F2555ExclusionLimit)
					days := int(dv.Get(forms.F2555QualifyingDays))
					return taxmath.ProrateExclusion(limit, days, 365)
				},
			},
			// Foreign income exclusion (the actual amount excluded)
			{
				Line:      "foreign_income_exclusion",
				Type:      forms.Computed,
				Label:     "Foreign earned income exclusion",
				DependsOn: []string{forms.F2555ForeignEarnedIncome, "form_2555:prorated_exclusion"},
				Compute: func(dv forms.DepValues) float64 {
					income := dv.Get(forms.F2555ForeignEarnedIncome)
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
				DependsOn: []string{forms.F2555HousingExpenses, "form_2555:housing_max", "form_2555:housing_base_amount"},
				Compute: func(dv forms.DepValues) float64 {
					expenses := dv.Get(forms.F2555HousingExpenses)
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
				DependsOn: []string{"form_2555:housing_qualifying_amount", forms.F2555EmployerHousing},
				Compute: func(dv forms.DepValues) float64 {
					qualifying := dv.Get("form_2555:housing_qualifying_amount")
					employerProvided := dv.Get(forms.F2555EmployerHousing)
					return math.Min(qualifying, employerProvided)
				},
			},
			// Housing deduction (for self-employed — excess over housing exclusion)
			{
				Line:      "housing_deduction",
				Type:      forms.Computed,
				Label:     "Foreign housing deduction (self-employed)",
				DependsOn: []string{"form_2555:housing_qualifying_amount", forms.F2555HousingExclusion, forms.F2555SelfEmployedAbroad},
				Compute: func(dv forms.DepValues) float64 {
					selfEmployed := dv.GetString(forms.F2555SelfEmployedAbroad)
					if selfEmployed != forms.OptionYes {
						return 0
					}
					qualifying := dv.Get("form_2555:housing_qualifying_amount")
					exclusion := dv.Get(forms.F2555HousingExclusion)
					return math.Max(0, qualifying-exclusion)
				},
			},

			// --- Total exclusion ---

			// Total exclusion (FEIE + housing exclusion)
			{
				Line:      "total_exclusion",
				Type:      forms.Computed,
				Label:     "Total foreign earned income and housing exclusion",
				DependsOn: []string{forms.F2555ForeignIncomeExcl, forms.F2555HousingExclusion},
				Compute: func(dv forms.DepValues) float64 {
					return dv.Get(forms.F2555ForeignIncomeExcl) +
						dv.Get(forms.F2555HousingExclusion)
				},
			},
		},
	}
}
