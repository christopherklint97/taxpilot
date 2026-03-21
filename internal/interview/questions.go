package interview

import "taxpilot/internal/forms"

// ContextualPrompt holds an enhanced prompt for a form field.
type ContextualPrompt struct {
	Prompt   string // the question to show the user
	HelpText string // additional context shown below the prompt
	CANote   string // CA-specific note (shown when filing in CA)
	IRCRef   string // IRC section reference (e.g., "IRC §61(a)")
	CARef    string // CA Revenue & Taxation Code reference (e.g., "R&TC §17071")
}

// contextualPrompts maps field keys to enhanced prompts.
// These are used instead of the raw FieldDef.Prompt values.
var contextualPrompts = map[string]ContextualPrompt{
	forms.F1040FilingStatus: {
		Prompt:   "What is your filing status for 2025?",
		HelpText: "Your filing status affects your tax brackets, standard deduction, and eligibility for certain credits.",
		CANote:   "California uses the same filing status as your federal return.",
		IRCRef:   "IRC §1",
		CARef:    "R&TC §17042",
	},
	forms.F1040FirstName: {
		Prompt:   "What is your first name?",
		HelpText: "As shown on your Social Security card.",
	},
	forms.F1040LastName: {
		Prompt:   "What is your last name?",
		HelpText: "As shown on your Social Security card.",
	},
	forms.F1040SSN: {
		Prompt:   "What is your Social Security number?",
		HelpText: "Format: XXX-XX-XXXX. This is required for filing and is kept secure.",
	},
	// --- Foreign Wages (Form 1040) ---
	forms.F1040ForeignWages: {
		Prompt:   "How much did you earn in wages from foreign employers in 2025 (in USD)?",
		HelpText: "Enter wages paid by foreign (non-US) employers that are NOT reported on a US W-2. Convert to USD using the yearly average exchange rate. Enter 0 if you had no foreign employment income. You can use the calc command with currency conversion (e.g., 'calc 500000 SEK').",
		CANote:   "California taxes all worldwide wages, including foreign-source wages.",
		IRCRef:   "IRC §61(a)",
	},
	forms.F1040ForeignEmployer: {
		Prompt:   "Who was your foreign employer?",
		HelpText: "Enter the name and country of your foreign employer(s) (e.g., \"Ericsson AB, Sweden\"). This is reported on Form 1040 line 1a alongside any W-2 wages.",
	},

	// --- W-2 (US employers only) ---
	// Keys use 2-part format (form_id:line) to match engine-generated keys.
	"w2:employer_name": {
		Prompt:   "Who is your US employer?",
		HelpText: "Enter the employer name exactly as shown on your W-2 (Box c). W-2 forms are issued by US employers only. If your employer is foreign, skip this form and enter your wages in the foreign wages section.",
	},
	"w2:employer_ein": {
		Prompt:   "What is the employer's EIN?",
		HelpText: "The 9-digit Employer Identification Number from your W-2 (Box b). Format: XX-XXXXXXX. If your employer is a foreign entity without a US EIN, enter FOREIGN.",
	},
	"w2:wages": {
		Prompt:   "What were your total wages from this employer?",
		HelpText: "This is Box 1 on your W-2: Wages, tips, and other compensation. This is your gross pay minus pre-tax deductions (401k, health insurance, etc.).",
		CANote:   "If your CA wages (Box 16) differ from federal wages (Box 1), we'll ask about that separately.",
		IRCRef:   "IRC §61(a)",
		CARef:    "R&TC §17071",
	},
	"w2:federal_tax_withheld": {
		Prompt:   "How much federal income tax was withheld?",
		HelpText: "W-2 Box 2. This is the amount your employer sent to the IRS on your behalf throughout the year.",
		IRCRef:   "IRC §3402",
	},
	"w2:ss_wages": {
		Prompt:   "What were your Social Security wages?",
		HelpText: "W-2 Box 3. Usually the same as Box 1, but may differ if you have pre-tax deductions that are subject to Social Security tax.",
		IRCRef:   "IRC §3121",
	},
	"w2:ss_tax_withheld": {
		Prompt:   "How much Social Security tax was withheld?",
		HelpText: "W-2 Box 4. Should be 6.2% of Box 3 (capped at the Social Security wage base of $176,100 for 2025).",
	},
	"w2:medicare_wages": {
		Prompt:   "What were your Medicare wages?",
		HelpText: "W-2 Box 5. Usually the same as Box 1. There is no cap on Medicare wages.",
		IRCRef:   "IRC §3101",
	},
	"w2:medicare_tax_withheld": {
		Prompt:   "How much Medicare tax was withheld?",
		HelpText: "W-2 Box 6. Should be 1.45% of Box 5 (plus 0.9% Additional Medicare Tax on wages over $200,000).",
	},
	"w2:state_wages": {
		Prompt:   "What were your state wages?",
		HelpText: "W-2 Box 16. This is your California taxable wages. Usually the same as Box 1 (federal wages).",
		CANote:   "If different from federal wages, this is typically due to items taxed differently by California.",
		CARef:    "R&TC §17071",
	},
	"w2:state_tax_withheld": {
		Prompt:   "How much California state tax was withheld?",
		HelpText: "W-2 Box 17. This is the amount your employer sent to the California FTB on your behalf.",
	},

	// --- Schedule A fields ---
	"schedule_a:1": {
		Prompt:   "What were your total medical and dental expenses?",
		HelpText: "Include insurance premiums you paid (not employer-paid), doctor visits, prescriptions, etc. Only the amount exceeding 7.5% of your AGI is deductible.",
		CANote:   "California uses the same 7.5% AGI threshold for medical expense deductions.",
		IRCRef:   "IRC §213(a)",
		CARef:    "R&TC §17201",
	},
	"schedule_a:5a": {
		Prompt:   "How much did you pay in state and local income taxes?",
		HelpText: "Include state income tax payments, estimated tax payments, and any prior-year state tax paid this year. The federal SALT deduction is capped at $10,000 ($5,000 if MFS).",
		CANote:   "California does NOT allow a deduction for state income taxes on Schedule CA. This amount will be added back.",
		IRCRef:   "IRC §164",
		CARef:    "R&TC §17220 (not allowed)",
	},
	"schedule_a:5b": {
		Prompt:   "How much did you pay in personal property taxes?",
		HelpText: "Taxes on personal property like vehicles, based on the value of the property. Only the ad valorem (value-based) portion is deductible.",
	},
	"schedule_a:5c": {
		Prompt:   "How much did you pay in real estate taxes?",
		HelpText: "Property taxes paid on your home and other real estate you own. Subject to the $10,000 SALT cap.",
		IRCRef:   "IRC §164(a)(1)",
		CARef:    "R&TC §17220",
	},
	"schedule_a:8a": {
		Prompt:   "How much home mortgage interest did you pay?",
		HelpText: "From Form 1098. Deductible on mortgages up to $750,000 ($375,000 if MFS). Includes points paid.",
		CANote:   "California generally conforms to federal mortgage interest deduction rules.",
		IRCRef:   "IRC §163(h)",
		CARef:    "R&TC §17201",
	},
	"schedule_a:12": {
		Prompt:   "How much did you give to charity (cash or check)?",
		HelpText: "Contributions to qualified charitable organizations by cash, check, or electronic payment. Keep receipts for all donations.",
		IRCRef:   "IRC §170",
		CARef:    "R&TC §17201",
	},
	"schedule_a:13": {
		Prompt:   "Any charitable contributions other than cash?",
		HelpText: "Donated clothing, household items, vehicles, stocks, etc. Items must be in good condition. Enter fair market value.",
	},
	"schedule_a:14": {
		Prompt:   "Any charitable contribution carryover from last year?",
		HelpText: "If your charitable deductions exceeded the AGI limitation in a prior year, you may carry the excess forward. Enter 0 if none.",
	},

	// --- 1099-INT fields ---
	"1099int:payer_name": {
		Prompt:   "Who is the US payer for your 1099-INT?",
		HelpText: "Enter the name of the US bank or institution that issued a 1099-INT. Only US financial institutions issue 1099-INT forms. If all your interest is from foreign banks, skip this form — foreign interest is entered separately on Schedule B.",
	},
	"1099int:payer_tin": {
		Prompt:   "What is the payer's TIN?",
		HelpText: "The 9-digit Taxpayer Identification Number from your 1099-INT. Format: XX-XXXXXXX.",
	},
	"1099int:interest_income": {
		Prompt:   "What was your interest income?",
		HelpText: "1099-INT Box 1. This is your total taxable interest earned during the year.",
		IRCRef:   "IRC §61(a)(4)",
	},
	"1099int:early_withdrawal_penalty": {
		Prompt:   "Was there an early withdrawal penalty?",
		HelpText: "1099-INT Box 2. Penalty for early withdrawal of a time deposit (e.g., CD). Enter 0 if none.",
	},
	"1099int:us_savings_bond_interest": {
		Prompt:   "Any interest on U.S. Savings Bonds or Treasury obligations?",
		HelpText: "1099-INT Box 3. This interest is taxable federally but exempt from state tax in most states.",
		CANote:   "California does not tax interest on U.S. government obligations. This will be subtracted on Schedule CA.",
		IRCRef:   "IRC §103",
		CARef:    "R&TC §17133",
	},
	"1099int:federal_tax_withheld": {
		Prompt:   "Was any federal tax withheld on interest?",
		HelpText: "1099-INT Box 4. Usually $0 unless backup withholding applied.",
	},
	"1099int:tax_exempt_interest": {
		Prompt:   "Any tax-exempt interest?",
		HelpText: "1099-INT Box 8. Interest from municipal bonds. Federally exempt but may be subject to state tax.",
		CANote:   "Only interest from California municipal bonds is exempt from CA tax. Out-of-state muni interest is taxable in CA.",
		IRCRef:   "IRC §103",
		CARef:    "R&TC §17133.5",
	},
	"1099int:private_activity_bond_interest": {
		Prompt:   "Any specified private activity bond interest?",
		HelpText: "1099-INT Box 9. This may be subject to the Alternative Minimum Tax (AMT). Enter 0 if none.",
	},

	// --- 1099-DIV fields (US payers only) ---
	"1099div:payer_name": {
		Prompt:   "Who is the US payer for your 1099-DIV?",
		HelpText: "Enter the name of the US brokerage or fund company that issued a 1099-DIV. Only US financial institutions issue 1099-DIV forms. If all your dividends are from foreign sources, skip this form — foreign dividends are reported directly on Schedule B and Form 1040.",
	},
	"1099div:payer_tin": {
		Prompt:   "What is the payer's TIN?",
		HelpText: "The 9-digit Taxpayer Identification Number from your 1099-DIV. Format: XX-XXXXXXX.",
	},
	"1099div:ordinary_dividends": {
		Prompt:   "What were your total ordinary dividends?",
		HelpText: "1099-DIV Box 1a. This includes both qualified and non-qualified dividends.",
		IRCRef:   "IRC §61(a)(7)",
	},
	"1099div:qualified_dividends": {
		Prompt:   "How much of that was qualified dividends?",
		HelpText: "1099-DIV Box 1b. Qualified dividends are taxed at lower capital gains rates. This is a subset of Box 1a.",
		CANote:   "California taxes qualified dividends as ordinary income \u2014 there is no preferential rate.",
		IRCRef:   "IRC §1(h)(11)",
		CARef:    "R&TC §17041 (taxed as ordinary)",
	},
	"1099div:total_capital_gain": {
		Prompt:   "Any capital gain distributions?",
		HelpText: "1099-DIV Box 2a. Long-term capital gain distributions from mutual funds. Enter 0 if none.",
	},
	"1099div:section_1250_gain": {
		Prompt:   "Any unrecaptured Section 1250 gain?",
		HelpText: "1099-DIV Box 2b. Gain from depreciation on real property in a fund. Usually 0. Enter 0 if none.",
	},
	"1099div:section_199a_dividends": {
		Prompt:   "Any Section 199A dividends?",
		HelpText: "1099-DIV Box 5. REIT dividends that may qualify for the 20% QBI deduction. Enter 0 if none.",
		CANote:   "California does not allow the Section 199A (QBI) deduction.",
		IRCRef:   "IRC §199A",
		CARef:    "R&TC (not allowed)",
	},
	"1099div:federal_tax_withheld": {
		Prompt:   "Was any federal tax withheld on dividends?",
		HelpText: "1099-DIV Box 4. Usually $0 unless backup withholding applied.",
	},
	"1099div:exempt_interest_dividends": {
		Prompt:   "Any exempt-interest dividends?",
		HelpText: "1099-DIV Box 12. Tax-exempt dividends from a mutual fund holding municipal bonds. Enter 0 if none.",
		CANote:   "Only the portion from California municipal bonds is exempt from CA tax.",
	},
	"1099div:private_activity_bond_dividends": {
		Prompt:   "Any specified private activity bond interest dividends?",
		HelpText: "1099-DIV Box 13. May be subject to AMT. Enter 0 if none.",
	},

	// --- Form 8889 (HSA) fields ---
	"form_8889:1": {
		Prompt:   "What type of high deductible health plan (HDHP) coverage do you have?",
		HelpText: "Self-only if you cover just yourself. Family if your HDHP covers you and at least one other person.",
		CANote:   "California does NOT conform to federal HSA treatment. HSA contributions are not deductible for CA purposes.",
	},
	"form_8889:2": {
		Prompt:   "How much did you contribute to your HSA for 2025?",
		HelpText: "Enter your personal HSA contributions (not through payroll). 2025 limits: $4,300 self-only, $8,550 family.",
		CANote:   "This deduction will be added back on Schedule CA \u2014 CA does not allow HSA deductions.",
		IRCRef:   "IRC §223",
		CARef:    "R&TC §17215 (not allowed)",
	},
	"form_8889:3": {
		Prompt:   "How much did your employer contribute to your HSA?",
		HelpText: "W-2 Box 12, code W. Includes both employer contributions and pre-tax payroll deductions. These reduce your available contribution limit.",
	},
	"form_8889:5": {
		Prompt:   "Are you age 55 or older? Enter catch-up contribution ($1,000 max, or 0):",
		HelpText: "If you're 55 or older by the end of 2025, you can contribute an additional $1,000.",
	},
	"form_8889:14a": {
		Prompt:   "How much did you receive in HSA distributions in 2025?",
		HelpText: "Total distributions from your HSA during the year. Form 1099-SA, Box 1.",
	},
	"form_8889:14c": {
		Prompt:   "How much of your HSA distributions were for qualified medical expenses?",
		HelpText: "Qualified expenses include doctor visits, prescriptions, dental, and vision care. Non-qualified distributions are subject to income tax plus a 20% penalty.",
	},

	// --- Schedule 3 fields ---
	"schedule_3:10": {
		Prompt:   "How much did you pay in federal estimated taxes for 2025?",
		HelpText: "Total of all quarterly estimated tax payments (1040-ES) you sent to the IRS for 2025. Enter 0 if none.",
		CANote:   "CA estimated payments are entered separately. This is federal only.",
		IRCRef:   "IRC §6654",
	},

	// --- 1099-NEC fields (US payers only) ---
	"1099nec:payer_name": {
		Prompt:   "Who is the US payer for your 1099-NEC?",
		HelpText: "Enter the name of the US client or company that issued a 1099-NEC. Only US entities issue 1099-NEC forms. If all your contractor income is from foreign clients, skip this form — foreign self-employment income is entered separately.",
	},
	"1099nec:payer_tin": {
		Prompt:   "What is the payer's TIN?",
		HelpText: "The 9-digit Taxpayer Identification Number from your 1099-NEC. Format: XX-XXXXXXX.",
	},
	"1099nec:nonemployee_compensation": {
		Prompt:   "How much nonemployee compensation did you receive?",
		HelpText: "1099-NEC Box 1. This is income for work you performed as an independent contractor. It is subject to self-employment tax.",
		CANote:   "California generally conforms to federal treatment of self-employment income.",
		IRCRef:   "IRC §61(a)(1)",
		CARef:    "R&TC §17071",
	},
	"1099nec:federal_tax_withheld": {
		Prompt:   "Was any federal tax withheld on this 1099-NEC?",
		HelpText: "1099-NEC Box 4. Usually $0 unless backup withholding applied.",
	},

	// --- Schedule C fields ---
	"schedule_c:business_name": {
		Prompt:   "What is your business name?",
		HelpText: "Enter your business name, or your own name if you are a sole proprietor without a separate business name.",
	},
	"schedule_c:business_code": {
		Prompt:   "What is your principal business activity code?",
		HelpText: "Enter the 6-digit NAICS code that best describes your business activity. Common codes: 541990 (other professional services), 541611 (management consulting), 541511 (custom computer programming).",
	},
	"schedule_c:8": {
		Prompt:   "How much did you spend on advertising?",
		HelpText: "Include costs for online ads, print ads, business cards, website hosting, and other promotional materials.",
		IRCRef:   "IRC §162(a)",
	},
	"schedule_c:10": {
		Prompt:   "What were your car and truck expenses for business?",
		HelpText: "Business-use portion only. You can use actual expenses or the standard mileage rate (70 cents/mile for 2025). Keep a mileage log.",
		IRCRef:   "IRC §162(a), §274(d)",
	},
	"schedule_c:17": {
		Prompt:   "How much did you pay for legal and professional services?",
		HelpText: "Include fees paid to attorneys, accountants, tax preparers, and other professionals for business-related services.",
		IRCRef:   "IRC §162(a)",
	},
	"schedule_c:18": {
		Prompt:   "What were your office expenses?",
		HelpText: "Include office supplies, postage, software subscriptions, and other general office costs. Does not include home office deduction (Form 8829).",
		IRCRef:   "IRC §162(a)",
	},
	"schedule_c:22": {
		Prompt:   "How much did you spend on supplies?",
		HelpText: "Materials and supplies consumed and used during the year in your business. Does not include items that are inventory.",
		IRCRef:   "IRC §162(a)",
	},
	"schedule_c:25": {
		Prompt:   "What were your business utility expenses?",
		HelpText: "Business portion of utilities (electricity, gas, water, internet, phone). If you work from home, enter only the business-use percentage.",
		IRCRef:   "IRC §162(a)",
	},
	"schedule_c:27": {
		Prompt:   "Any other business expenses not listed above?",
		HelpText: "Enter the total of any other ordinary and necessary business expenses not covered by the specific expense lines (e.g., professional development, subscriptions, tools).",
		IRCRef:   "IRC §162(a)",
	},

	// --- 1099-B fields (US brokers only) ---
	"1099b:description": {
		Prompt:   "Describe the security you sold (from a US broker):",
		HelpText: "Enter a short description like \"100 sh AAPL\" or \"VTSAX mutual fund.\" Only US brokers issue 1099-B forms. If you sold securities through a foreign brokerage, skip this form — those sales are reported directly on Form 8949.",
	},
	"1099b:date_acquired": {
		Prompt:   "When did you acquire this security?",
		HelpText: "Enter the date in MM/DD/YYYY format, or VARIOUS if acquired over multiple dates. Holdings over 1 year qualify for lower long-term capital gains rates.",
	},
	"1099b:date_sold": {
		Prompt:   "When did you sell this security?",
		HelpText: "Enter the date in MM/DD/YYYY format from your 1099-B.",
	},
	"1099b:proceeds": {
		Prompt:   "What were the proceeds from this sale?",
		HelpText: "1099-B Box 1d. The total amount you received from the sale, before commissions.",
		IRCRef:   "IRC §1001",
		CARef:    "R&TC §18031",
	},
	"1099b:cost_basis": {
		Prompt:   "What was the cost basis?",
		HelpText: "1099-B Box 1e. Your original purchase price plus commissions. If basis was not reported to the IRS, check your brokerage statements.",
		IRCRef:   "IRC §1001",
		CARef:    "R&TC §18031",
	},
	"1099b:wash_sale_loss": {
		Prompt:   "Any wash sale loss disallowed?",
		HelpText: "1099-B Box 1g. If you bought substantially identical securities within 30 days of selling at a loss, the loss is disallowed. Enter 0 if none.",
	},
	"1099b:federal_tax_withheld": {
		Prompt:   "Was any federal tax withheld?",
		HelpText: "1099-B Box 4. Usually $0 unless backup withholding applied.",
	},
	"1099b:term": {
		Prompt:   "Was this a short-term or long-term holding?",
		HelpText: "Short-term: held 1 year or less (taxed as ordinary income). Long-term: held more than 1 year (taxed at preferential rates of 0%, 15%, or 20%).",
		CANote:   "California taxes capital gains as ordinary income \u2014 there is no preferential long-term rate.",
	},
	"1099b:basis_reported": {
		Prompt:   "Was the cost basis reported to the IRS?",
		HelpText: "Check Box 12 on your 1099-B. Most brokers report basis for stocks purchased after 2011.",
	},

	// --- Form 3514 (CalEITC) fields ---
	"form_3514:3": {
		Prompt:   "How many qualifying children do you have for the California Earned Income Tax Credit?",
		HelpText: "A qualifying child must be under age 18 (or under 24 if a student), live with you in California for more than half the year, and have a valid SSN or ITIN.",
		CANote:   "The CalEITC provides a refundable credit for low-income workers with earned income up to $30,950.",
		CARef:    "R&TC \u00a718051",
	},
	"form_3514:6_yctc": {
		Prompt:   "Do you have a qualifying child under age 6?",
		HelpText: "If yes, you may qualify for the Young Child Tax Credit (YCTC), an additional $1,117 credit.",
		CANote:   "The YCTC is available to CalEITC-eligible filers with a child under age 6 at the end of the tax year.",
		CARef:    "R&TC \u00a717052.1",
	},

	// --- Form 3853 (Health Coverage) fields ---
	"form_3853:1": {
		Prompt:   "Did you have qualifying health coverage for all 12 months of 2025?",
		HelpText: "Qualifying coverage includes employer plans, Covered California, Medicare, Medi-Cal, TRICARE, and most other minimum essential coverage.",
		CANote:   "California requires all residents to have health coverage or pay a penalty (Individual Shared Responsibility).",
		CARef:    "R&TC \u00a761015",
	},
	"form_3853:2": {
		Prompt:   "How many months were you without qualifying health coverage?",
		HelpText: "Count only full months without coverage. If you had coverage for any part of a month, that month counts as covered.",
		CANote:   "The penalty is calculated per month without coverage.",
	},
	"form_3853:3": {
		Prompt:   "Did you have an exemption from the health coverage requirement?",
		HelpText: "Exemptions include religious conscience, coverage gap of less than 3 months, affordability hardship, and certain other circumstances.",
		CANote:   "You can apply for exemptions through Covered California or claim them on Form 3853.",
		CARef:    "R&TC \u00a761030",
	},

	// --- Form 2555 (Foreign Earned Income Exclusion) fields ---
	"form_2555:foreign_country": {
		Prompt:   "In which country do you live and work?",
		HelpText: "Enter the country where you have established your tax home (principal place of business or employment).",
		CANote:   "California does NOT allow the FEIE — your excluded income will be added back for CA tax purposes.",
		IRCRef:   "IRC §911(d)(3)",
	},
	"form_2555:foreign_earned_income": {
		Prompt:   "What was your total foreign earned income in 2025 (in USD)?",
		HelpText: "Include wages, salary, professional fees, and self-employment income earned while living abroad. Convert to USD using the yearly average exchange rate.",
		CANote:   "California taxes all worldwide income — this full amount will be taxed by CA even if excluded federally.",
		IRCRef:   "IRC §911(b)",
	},
	"form_2555:qualifying_test": {
		Prompt:   "Which qualifying test do you meet for the FEIE?",
		HelpText: "Bona Fide Residence: You established residence in a foreign country for a full tax year. Physical Presence: You were in a foreign country for 330+ days in a 12-month period.",
		IRCRef:   "IRC §911(d)(1)",
	},
	"form_2555:bfrt_full_year": {
		Prompt:   "Were you a bona fide resident of a foreign country for the entire 2025 tax year?",
		HelpText: "If yes, you qualify for the full exclusion. If you established or ended your foreign residence during 2025, the exclusion will be prorated.",
		IRCRef:   "IRC §911(d)(1)(A)",
	},
	"form_2555:ppt_days_present": {
		Prompt:   "How many full days were you physically present in a foreign country?",
		HelpText: "Count only complete days (midnight to midnight) in a foreign country during a qualifying 12-month period. You need at least 330 days to qualify.",
		IRCRef:   "IRC §911(d)(1)(B)",
	},
	"form_2555:housing_expenses": {
		Prompt:   "What were your total housing expenses abroad in 2025 (in USD)?",
		HelpText: "Include rent, utilities (not phone), insurance, parking, furniture rental, and repairs. Do not include mortgage payments, purchased furniture, or domestic labor.",
		IRCRef:   "IRC §911(c)(3)",
	},
	"form_2555:foreign_tax_paid": {
		Prompt:   "How much foreign income tax did you pay in 2025 (in USD)?",
		HelpText: "Enter the total income taxes paid to your country of residence. You may be able to claim a Foreign Tax Credit on the income not excluded by the FEIE.",
		IRCRef:   "IRC §901",
	},

	// --- Form 1116 (Foreign Tax Credit) fields ---
	"form_1116:foreign_source_income": {
		Prompt:   "What was your foreign source taxable income (not excluded by FEIE) in USD?",
		HelpText: "This is income from foreign sources that was NOT excluded by the FEIE. You cannot claim both the FEIE and FTC on the same income.",
		IRCRef:   "IRC §901(a)",
	},
	"form_1116:foreign_tax_paid_income": {
		Prompt:   "How much foreign income tax did you pay on non-excluded income (in USD)?",
		HelpText: "Enter foreign taxes paid on income that was NOT excluded by the FEIE. Taxes on FEIE-excluded income cannot be credited.",
		IRCRef:   "IRC §911(d)(6)",
	},
	"form_1116:foreign_country": {
		Prompt:   "Which country did you pay foreign taxes to?",
		HelpText: "Enter the primary country where you paid foreign income taxes.",
	},

	// --- Form 8938 (FATCA) fields ---
	"form_8938:lives_abroad": {
		Prompt:   "Do you meet the bona fide residence or physical presence test for living abroad?",
		HelpText: "If yes, higher FATCA reporting thresholds apply ($200,000/$300,000 single vs $50,000/$75,000 for US residents).",
		IRCRef:   "IRC §6038D(b)",
	},
	"form_8938:max_value_accounts": {
		Prompt:   "What was the maximum aggregate value of all your foreign financial accounts at any time during 2025 (in USD)?",
		HelpText: "Include bank accounts, brokerage accounts, and pension accounts. Use the highest combined value at any point during the year.",
		IRCRef:   "IRC §6038D",
	},
	"form_8938:yearend_value_accounts": {
		Prompt:   "What was the total value of all your foreign financial accounts on December 31, 2025 (in USD)?",
		HelpText: "Report the combined balance of all foreign accounts as of year-end.",
	},

	// --- Form 8833 (Treaty Disclosure) fields ---
	"form_8833:treaty_country": {
		Prompt:   "Which country's tax treaty are you relying on?",
		HelpText: "Enter the country whose tax treaty with the US you are using to take a treaty-based return position.",
		IRCRef:   "IRC §6114",
	},
	"form_8833:treaty_article": {
		Prompt:   "Which article of the tax treaty applies?",
		HelpText: "For example, 'Article 18 — Pensions' for the US-Sweden treaty. Failure to disclose carries a $1,000 penalty per position.",
		IRCRef:   "IRC §6114/7701(b)",
	},

	// --- Schedule B Part I (Foreign Interest) ---
	forms.SchedBForeignInterest: {
		Prompt:   "How much interest income did you receive from foreign banks or institutions in 2025 (in USD)?",
		HelpText: "Report interest from foreign bank accounts, pension funds, and other non-US financial institutions. This income is NOT reported on a 1099-INT. Convert to USD using the yearly average exchange rate. Enter 0 if none. You can use the calc command with currency conversion (e.g., 'calc 5000 SEK').",
		CANote:   "California taxes all worldwide interest income, including foreign-source interest.",
		IRCRef:   "IRC §61(a)(4)",
	},
	forms.SchedBForeignInterestPayer: {
		Prompt:   "Who paid the foreign interest? (e.g., \"Nordea Bank, Sweden\")",
		HelpText: "List the foreign bank(s) or institution(s) that paid you interest. This is reported in Part I of Schedule B alongside any 1099-INT payers. Enter 'none' or leave blank if you entered 0 for foreign interest.",
	},

	// --- Schedule B Part III (Foreign Accounts) ---
	forms.SchedBLine7a: {
		Prompt:   "Did you have a financial interest in or signature authority over any foreign financial account?",
		HelpText: "This includes bank accounts, securities accounts, and other financial accounts in a foreign country. If yes and the aggregate value exceeded $10,000, you must file an FBAR (FinCEN 114).",
		IRCRef:   "31 USC §5314",
	},
	forms.SchedBLine7b: {
		Prompt:   "In which countries are your foreign financial accounts located?",
		HelpText: "List all countries where you have foreign financial accounts.",
	},
}

// GetContextualPrompt returns the enhanced prompt for a field key, falling back
// to the original prompt if no contextual prompt is defined.
func GetContextualPrompt(fieldKey string, originalPrompt string, stateCode string) ContextualPrompt {
	if cp, ok := contextualPrompts[fieldKey]; ok {
		result := ContextualPrompt{
			Prompt:   cp.Prompt,
			HelpText: cp.HelpText,
			IRCRef:   cp.IRCRef,
		}
		// Only include CANote and CARef when filing in California
		if stateCode == forms.StateCodeCA {
			result.CANote = cp.CANote
			result.CARef = cp.CARef
		}
		return result
	}

	// Fall back to the original prompt with no extra help text
	return ContextualPrompt{
		Prompt: originalPrompt,
	}
}
