package ca

// ConformityDifference documents an area where California tax treatment
// differs from the federal treatment. These differences drive Schedule CA
// adjustments and CA-specific interview questions.
type ConformityDifference struct {
	Area           string // category of the difference
	Federal        string // federal tax treatment
	CA             string // California tax treatment
	ScheduleCALine string // corresponding Schedule CA line, if applicable
}

// CAConformityDifferences lists key areas where CA differs from federal tax treatment.
// This is used by the interview engine to ask CA-specific questions and to
// populate Schedule CA adjustments.
var CAConformityDifferences = []ConformityDifference{
	{
		Area:           "Social Security Benefits",
		Federal:        "Partially taxable (up to 85%)",
		CA:             "Not taxable (fully exempt)",
		ScheduleCALine: "6a",
	},
	{
		Area:           "SALT Deduction",
		Federal:        "Deductible up to $10,000 ($5,000 MFS)",
		CA:             "No deduction for state/local income taxes paid",
		ScheduleCALine: "5a",
	},
	{
		Area:           "Standard Deduction",
		Federal:        "$15,000 single / $30,000 MFJ (2025)",
		CA:             "$5,706 single / $11,412 MFJ (2025)",
		ScheduleCALine: "",
	},
	{
		Area:           "QBI Deduction (Section 199A)",
		Federal:        "Up to 20% deduction for qualified business income",
		CA:             "Not allowed — add back on Schedule CA",
		ScheduleCALine: "13",
	},
	{
		Area:           "Municipal Bond Interest",
		Federal:        "Tax-exempt for all states",
		CA:             "Only CA-issued bonds are exempt; out-of-state bonds are taxable",
		ScheduleCALine: "2a",
	},
	{
		Area:           "Health Savings Account (HSA)",
		Federal:        "Contributions deductible; earnings tax-free",
		CA:             "No deduction; earnings taxable",
		ScheduleCALine: "13",
	},
	{
		Area:           "529 Plan Distributions",
		Federal:        "Up to $10,000/year for K-12 tuition is tax-free",
		CA:             "K-12 distributions are taxable (higher-ed distributions are tax-free)",
		ScheduleCALine: "8",
	},
	{
		Area:           "Moving Expenses",
		Federal:        "Deductible only for active-duty military",
		CA:             "Deductible for all taxpayers meeting distance/time tests",
		ScheduleCALine: "14",
	},
	{
		Area:           "Gambling Losses",
		Federal:        "Deductible up to gambling winnings (itemized)",
		CA:             "Same as federal",
		ScheduleCALine: "",
	},
	{
		Area:           "Foreign Earned Income Exclusion",
		Federal:        "Up to $130,000 of foreign earned income excluded (2025) via Form 2555",
		CA:             "Not allowed — CA taxes worldwide income regardless of FEIE",
		ScheduleCALine: "8d",
	},
	{
		Area:           "Foreign Housing Exclusion/Deduction",
		Federal:        "Additional exclusion for qualifying housing expenses abroad (Form 2555)",
		CA:             "Not allowed — add back on Schedule CA",
		ScheduleCALine: "8d",
	},
	{
		Area:           "Foreign Tax Credit",
		Federal:        "Credit for taxes paid to foreign governments (Form 1116)",
		CA:             "CA allows a credit for taxes paid to other states and foreign countries",
		ScheduleCALine: "",
	},
	{
		Area:           "Mental Health Services Tax",
		Federal:        "N/A",
		CA:             "Additional 1% on taxable income over $1,000,000",
		ScheduleCALine: "",
	},
}
