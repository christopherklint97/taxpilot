package interview

// CAFederalDifference describes a specific CA vs federal difference with plain-English explanation.
type CAFederalDifference struct {
	Area       string   // e.g., "HSA Deduction"
	IRCSection string   // e.g., "IRC §223"
	RTCSection string   // e.g., "R&TC §17215"
	Federal    string   // plain-English federal treatment
	California string   // plain-English CA treatment
	Impact     string   // what this means for the taxpayer
	FieldKeys  []string // which interview fields this affects
}

// caFederalDifferences contains all known CA vs federal differences.
var caFederalDifferences = []CAFederalDifference{
	{
		Area:       "Wages",
		IRCSection: "IRC §61(a)",
		RTCSection: "R&TC §17071",
		Federal:    "Wages, salaries, tips, and other compensation are included in gross income.",
		California: "California generally conforms to federal wage treatment. State wages (W-2 Box 16) usually match federal wages (Box 1), but may differ for items like HSA contributions that CA does not exclude.",
		Impact:     "If you have pre-tax HSA contributions, your CA wages may be higher than federal wages.",
		FieldKeys:  []string{"w2:1:wages", "w2:1:state_wages"},
	},
	{
		Area:       "SALT Deduction",
		IRCSection: "IRC §164",
		RTCSection: "R&TC §17220",
		Federal:    "State and local taxes (income, sales, property) are deductible as an itemized deduction, subject to a $10,000 cap ($5,000 if MFS).",
		California: "California does not allow a deduction for state income taxes paid. The SALT deduction claimed on the federal Schedule A must be added back on Schedule CA.",
		Impact:     "Your CA itemized deductions will be lower than federal because state income taxes cannot be deducted.",
		FieldKeys:  []string{"schedule_a:5a"},
	},
	{
		Area:       "HSA Deduction",
		IRCSection: "IRC §223",
		RTCSection: "R&TC §17215",
		Federal:    "Contributions to a Health Savings Account are deductible above the line. Earnings grow tax-free, and qualified medical distributions are tax-free.",
		California: "California does not conform to federal HSA treatment. Contributions are not deductible, earnings are taxable, and employer contributions are included in CA income.",
		Impact:     "Your HSA deduction will be added back on Schedule CA, increasing your CA taxable income. HSA earnings are also taxable for CA purposes.",
		FieldKeys:  []string{"form_8889:2", "form_8889:3", "form_8889:14a", "form_8889:14c"},
	},
	{
		Area:       "QBI Deduction (Section 199A)",
		IRCSection: "IRC §199A",
		RTCSection: "R&TC (not allowed)",
		Federal:    "Qualified business income from pass-through entities (sole proprietorships, partnerships, S corps) may qualify for a 20% deduction, subject to income limits.",
		California: "California does not allow the Section 199A qualified business income deduction. The full amount of QBI is taxable for CA purposes.",
		Impact:     "If you claim the QBI deduction federally, it will be added back on Schedule CA, increasing your CA taxable income.",
		FieldKeys:  []string{"1099div:1:section_199a_dividends"},
	},
	{
		Area:       "Qualified Dividends",
		IRCSection: "IRC §1(h)(11)",
		RTCSection: "R&TC §17041",
		Federal:    "Qualified dividends are taxed at preferential long-term capital gains rates (0%, 15%, or 20% depending on income).",
		California: "California taxes qualified dividends as ordinary income. There is no preferential rate for qualified dividends.",
		Impact:     "Your effective tax rate on qualified dividends will be higher for CA purposes since they are taxed at your ordinary income rate.",
		FieldKeys:  []string{"1099div:1:qualified_dividends"},
	},
	{
		Area:       "Capital Gains",
		IRCSection: "IRC §1001",
		RTCSection: "R&TC §18031",
		Federal:    "Long-term capital gains (assets held over 1 year) are taxed at preferential rates of 0%, 15%, or 20%.",
		California: "California taxes all capital gains as ordinary income. There is no preferential long-term capital gains rate.",
		Impact:     "Your CA tax on long-term capital gains will likely be higher than federal since CA applies ordinary income rates.",
		FieldKeys:  []string{"1099b:1:proceeds", "1099b:1:cost_basis", "1099b:1:term"},
	},
	{
		Area:       "U.S. Government Bond Interest",
		IRCSection: "IRC §103",
		RTCSection: "R&TC §17133",
		Federal:    "Interest on U.S. Treasury bonds, notes, and savings bonds is included in federal gross income.",
		California: "California does not tax interest on U.S. government obligations. This interest is subtracted on Schedule CA.",
		Impact:     "Your CA taxable income will be lower if you have U.S. government bond interest, since it is exempt from CA tax.",
		FieldKeys:  []string{"1099int:1:us_savings_bond_interest"},
	},
	{
		Area:       "Municipal Bond Interest",
		IRCSection: "IRC §103",
		RTCSection: "R&TC §17133.5",
		Federal:    "Interest from state and local municipal bonds is exempt from federal income tax.",
		California: "Only interest from California municipal bonds is exempt from CA tax. Interest from out-of-state municipal bonds is taxable in California.",
		Impact:     "If you hold out-of-state muni bonds, you will owe CA tax on that interest even though it is federally tax-exempt.",
		FieldKeys:  []string{"1099int:1:tax_exempt_interest", "1099div:1:exempt_interest_dividends"},
	},
	{
		Area:       "Standard Deduction",
		IRCSection: "IRC §63",
		RTCSection: "R&TC §17073.5",
		Federal:    "The federal standard deduction for 2025 is $15,000 (single), $30,000 (MFJ), or $22,500 (HOH).",
		California: "California's standard deduction is much lower: $5,540 (single) or $11,080 (MFJ) for 2025. Itemizing may be more beneficial for CA even if you take the standard deduction federally.",
		Impact:     "You may want to itemize on your CA return even if you take the federal standard deduction, since the CA standard deduction is significantly lower.",
		FieldKeys:  []string{},
	},
	{
		Area:       "Social Security Benefits",
		IRCSection: "IRC §86",
		RTCSection: "R&TC §17087",
		Federal:    "Up to 85% of Social Security benefits may be taxable federally, depending on your provisional income.",
		California: "California fully exempts Social Security benefits from state income tax, regardless of income level.",
		Impact:     "If you receive Social Security, your CA taxable income will be lower since CA does not tax these benefits at all.",
		FieldKeys:  []string{},
	},
	{
		Area:       "Foreign Earned Income Exclusion",
		IRCSection: "IRC §911",
		RTCSection: "R&TC §17024.5 (not adopted)",
		Federal:    "Up to $130,000 of foreign earned income can be excluded from federal tax using Form 2555. To qualify, you must have a tax home in a foreign country and meet either the Bona Fide Residence Test or the Physical Presence Test (330+ days abroad).",
		California: "California does NOT conform to the FEIE. All foreign earned income is fully taxable by California regardless of the federal exclusion. The entire FEIE amount is added back on Schedule CA.",
		Impact:     "Your CA tax will be significantly higher than your federal tax because the full FEIE exclusion (up to $130,000) is added back to CA income. This can result in $10,000-$17,000+ in additional CA tax.",
		FieldKeys:  []string{"form_2555:total_exclusion", "form_2555:foreign_earned_income"},
	},
	{
		Area:       "Foreign Housing Exclusion/Deduction",
		IRCSection: "IRC §911(c)",
		RTCSection: "R&TC §17024.5 (not adopted)",
		Federal:    "Foreign housing expenses above a base amount may be excluded (employees) or deducted (self-employed) from federal tax via Form 2555.",
		California: "California does not allow the foreign housing exclusion or deduction. The full amount is added back on Schedule CA.",
		Impact:     "Any housing exclusion or deduction claimed federally increases your CA taxable income.",
		FieldKeys:  []string{"form_2555:housing_exclusion", "form_2555:housing_deduction"},
	},
	{
		Area:       "Foreign Tax Credit",
		IRCSection: "IRC §901",
		RTCSection: "R&TC §18001",
		Federal:    "A credit is allowed for income taxes paid to foreign governments, limited by the ratio of foreign source income to worldwide income.",
		California: "California allows a credit for taxes paid to other states and foreign countries on income that is also taxed by CA. This is claimed on the CA return directly.",
		Impact:     "Since CA taxes your worldwide income (including FEIE-excluded income), you may be able to claim a larger CA foreign tax credit than the federal credit.",
		FieldKeys:  []string{"form_1116:foreign_tax_paid_income", "form_1116:22"},
	},
	{
		Area:       "Mental Health Services Tax",
		IRCSection: "",
		RTCSection: "R&TC §17043",
		Federal:    "No equivalent federal provision.",
		California: "California imposes an additional 1% tax on taxable income over $1,000,000 to fund mental health services (Proposition 63).",
		Impact:     "If your CA taxable income exceeds $1,000,000, you will owe an additional 1% mental health services tax on the excess amount.",
		FieldKeys:  []string{},
	},
}

// fieldKeyToDifference maps field keys to their CA difference index for fast lookup.
var fieldKeyToDifference map[string]int

func init() {
	fieldKeyToDifference = make(map[string]int)
	for i, diff := range caFederalDifferences {
		for _, key := range diff.FieldKeys {
			fieldKeyToDifference[key] = i
		}
	}
}

// GetCADifference returns the CA vs federal difference for a given field key,
// or nil if no difference applies to that field.
func GetCADifference(fieldKey string) *CAFederalDifference {
	if idx, ok := fieldKeyToDifference[fieldKey]; ok {
		diff := caFederalDifferences[idx]
		return &diff
	}
	return nil
}

// AllCADifferences returns all known CA vs federal differences.
func AllCADifferences() []CAFederalDifference {
	result := make([]CAFederalDifference, len(caFederalDifferences))
	copy(result, caFederalDifferences)
	return result
}
