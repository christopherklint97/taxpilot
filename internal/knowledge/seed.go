package knowledge

// SeedFederalDocuments returns essential federal tax knowledge documents.
func SeedFederalDocuments() []Document {
	return []Document{
		{
			ID: "irc_1", Title: "Tax Rates and Brackets",
			Content: "IRC §1 imposes federal income tax on taxable income. For 2025, the tax brackets for single filers are: 10% on income up to $11,925; 12% on $11,926-$48,475; 22% on $48,476-$103,350; 24% on $103,351-$197,300; 32% on $197,301-$250,525; 35% on $250,526-$626,350; 37% on income over $626,350. Married filing jointly brackets are roughly double the single thresholds.",
			Source: "IRC §1", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "1", Tags: []string{"tax rates", "brackets", "tax table", "marginal rate", "progressive tax"},
		},
		{
			ID: "irc_2", Title: "Filing Status Definitions",
			Content: "IRC §2 defines the filing statuses that determine tax rates and standard deduction amounts. The five statuses are: Single (unmarried individuals), Married Filing Jointly (MFJ, spouses filing one return), Married Filing Separately (MFS, spouses filing separate returns), Head of Household (HOH, unmarried with qualifying dependent), and Qualifying Surviving Spouse (QSS, within two years of spouse's death with dependent child). Filing status is determined as of December 31 of the tax year.",
			Source: "IRC §2", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "2", Tags: []string{"filing status", "single", "married", "head of household", "qualifying surviving spouse", "MFJ", "MFS"},
		},
		{
			ID: "irc_61", Title: "Gross Income Defined",
			Content: "IRC §61 defines gross income as all income from whatever source derived, including compensation for services (wages, salaries, tips), business income, gains from property dealings, interest, rents, royalties, dividends, alimony (for pre-2019 agreements), annuities, life insurance proceeds, pensions, and income from discharge of indebtedness. This is the broadest definition of income in the tax code.",
			Source: "IRC §61", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "61", Tags: []string{"income", "gross income", "wages", "compensation", "definition", "all income"},
		},
		{
			ID: "irc_62", Title: "Adjusted Gross Income (AGI) Defined",
			Content: "IRC §62 defines adjusted gross income (AGI) as gross income minus specific 'above-the-line' deductions. These include: educator expenses (up to $300), IRA contributions, student loan interest (up to $2,500), HSA contributions, self-employment tax deduction (50%), self-employed health insurance, alimony paid (pre-2019 agreements), and moving expenses for military. AGI is a critical threshold for many other tax provisions.",
			Source: "IRC §62", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "62", Tags: []string{"AGI", "adjusted gross income", "above the line", "deductions", "IRA", "student loan"},
		},
		{
			ID: "irc_63", Title: "Taxable Income Defined",
			Content: "IRC §63 defines taxable income as gross income minus deductions. For taxpayers who do not itemize, taxable income equals adjusted gross income minus the standard deduction. The standard deduction for 2025 is $15,000 for single filers, $30,000 for married filing jointly, $15,000 for married filing separately, and $22,500 for head of household. Taxpayers age 65+ or blind receive additional standard deduction amounts.",
			Source: "IRC §63", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "63", Tags: []string{"taxable income", "standard deduction", "deductions", "AGI", "itemize"},
		},
		{
			ID: "irc_151", Title: "Personal Exemptions",
			Content: "IRC §151 provides for personal exemptions, but the Tax Cuts and Jobs Act (TCJA) reduced the personal exemption amount to $0 for tax years 2018 through 2025. The personal exemption is expected to return in 2026 unless Congress extends TCJA provisions. For 2025, taxpayers cannot claim personal exemptions on their federal return, though the concept still affects certain calculations.",
			Source: "IRC §151", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "151", Tags: []string{"personal exemption", "exemption", "TCJA", "dependents"},
		},
		{
			ID: "irc_162", Title: "Business Expenses",
			Content: "IRC §162 allows deductions for ordinary and necessary expenses paid in carrying on a trade or business. This includes salaries, wages, supplies, rent, travel, and other costs directly related to business operations. For employees, most unreimbursed employee business expenses are not deductible (suspended by TCJA through 2025). Self-employed individuals report business expenses on Schedule C.",
			Source: "IRC §162", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "162", Tags: []string{"business expenses", "deduction", "ordinary", "necessary", "Schedule C", "self-employment"},
		},
		{
			ID: "irc_163", Title: "Interest Deduction",
			Content: "IRC §163 allows a deduction for interest paid or accrued during the tax year. For individuals, the most common application is the mortgage interest deduction. Under TCJA, mortgage interest is deductible on acquisition debt up to $750,000 ($375,000 MFS). Home equity loan interest is only deductible if used to buy, build, or improve the home. Investment interest expense is deductible up to net investment income.",
			Source: "IRC §163", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "163", Tags: []string{"interest", "mortgage", "deduction", "home equity", "investment interest", "itemized"},
		},
		{
			ID: "irc_164", Title: "State and Local Tax (SALT) Deduction",
			Content: "IRC §164 allows a deduction for state and local taxes paid, including income taxes (or general sales taxes as an alternative), real property taxes, and personal property taxes. Under TCJA, the total SALT deduction is capped at $10,000 ($5,000 for MFS) for tax years 2018-2025. This cap significantly impacts taxpayers in high-tax states like California and New York.",
			Source: "IRC §164", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "164", Tags: []string{"SALT", "state and local taxes", "property tax", "income tax deduction", "cap", "itemized"},
		},
		{
			ID: "irc_170", Title: "Charitable Contributions",
			Content: "IRC §170 allows a deduction for charitable contributions made to qualified organizations. Cash contributions are generally deductible up to 60% of AGI. Contributions of appreciated property are deductible at fair market value, limited to 30% of AGI. Excess contributions can be carried forward for five years. Donations must be substantiated: receipts required for $250+ contributions, qualified appraisals for property donations over $5,000.",
			Source: "IRC §170", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "170", Tags: []string{"charitable", "donation", "contribution", "deduction", "itemized", "nonprofit"},
		},
		{
			ID: "irc_199a", Title: "Qualified Business Income (QBI) Deduction",
			Content: "IRC §199A allows eligible taxpayers to deduct up to 20% of qualified business income from pass-through entities (sole proprietorships, partnerships, S corporations). The deduction is limited for high-income taxpayers based on W-2 wages paid and qualified property held by the business. For 2025, the phase-out begins at $191,950 (single) and $383,900 (MFJ). Specified service trades (law, health, consulting, etc.) face additional limitations. This deduction is taken below the line but is available even if you don't itemize.",
			Source: "IRC §199A", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "199A", Tags: []string{"QBI", "qualified business income", "pass-through", "deduction", "199A", "self-employment"},
		},
		{
			ID: "irc_401k", Title: "401(k) Retirement Plans",
			Content: "IRC §401(k) governs employer-sponsored retirement savings plans. For 2025, the employee contribution limit is $23,500 ($31,000 if age 50+, $34,750 if age 60-63 under SECURE 2.0 catch-up). Traditional 401(k) contributions reduce taxable income. Employer matches are not counted toward the employee limit but are subject to the overall $70,000 annual addition limit. Roth 401(k) contributions are made after-tax but grow tax-free. Early withdrawals before age 59.5 generally incur a 10% penalty plus income tax.",
			Source: "IRC §401(k)", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "401k", Tags: []string{"401k", "retirement", "contribution", "employer", "Roth", "traditional", "catch-up"},
		},
		{
			ID: "irc_408", Title: "Individual Retirement Accounts (IRAs)",
			Content: "IRC §408 governs traditional IRAs. For 2025, the contribution limit is $7,000 ($8,000 if age 50+). Contributions may be tax-deductible depending on income and whether you have an employer retirement plan. If covered by a workplace plan, the deduction phases out at $79,000-$89,000 AGI (single) or $126,000-$146,000 (MFJ). Roth IRA contributions (IRC §408A) are not deductible but qualified distributions are tax-free. Roth income phase-out: $150,000-$165,000 (single), $236,000-$246,000 (MFJ).",
			Source: "IRC §408", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "408", Tags: []string{"IRA", "traditional IRA", "Roth IRA", "retirement", "contribution", "deduction", "phase-out"},
		},
		{
			ID: "irc_3101", Title: "FICA Taxes (Social Security and Medicare)",
			Content: "IRC §3101 and §3111 impose Federal Insurance Contributions Act (FICA) taxes on wages. The employee share is 6.2% for Social Security (on wages up to $176,100 for 2025) and 1.45% for Medicare (no wage cap). An Additional Medicare Tax of 0.9% applies to wages over $200,000 (single) or $250,000 (MFJ). FICA taxes are withheld by the employer and reported in Box 4 (SS) and Box 6 (Medicare) of Form W-2.",
			Source: "IRC §3101/§3111", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "3101", Tags: []string{"FICA", "Social Security", "Medicare", "payroll tax", "withholding", "W-2"},
		},
		{
			ID: "irc_32", Title: "Earned Income Tax Credit (EITC)",
			Content: "IRC §32 provides the Earned Income Tax Credit, a refundable credit for low-to-moderate income workers. For 2025, the maximum credit ranges from $649 (no children) to $7,830 (3+ children). The credit phases in and then phases out based on earned income and AGI. Investment income must be $11,600 or less. Filing status, number of qualifying children, and income level determine the credit amount. The EITC is fully refundable.",
			Source: "IRC §32", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "32", Tags: []string{"EITC", "earned income credit", "refundable", "credit", "low income", "children"},
		},
		{
			ID: "irc_24", Title: "Child Tax Credit",
			Content: "IRC §24 provides a credit of up to $2,000 per qualifying child under age 17. For 2025, the credit begins to phase out at $200,000 AGI (single) and $400,000 (MFJ). Up to $1,700 of the credit is refundable as the Additional Child Tax Credit. A qualifying child must have a valid SSN, be claimed as a dependent, and meet relationship, age, residency, and support tests. A $500 credit is available for other dependents.",
			Source: "IRC §24", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "24", Tags: []string{"child tax credit", "CTC", "dependent", "refundable", "additional child tax credit"},
		},
		{
			ID: "pub_w2", Title: "Understanding Form W-2",
			Content: "Form W-2 reports wages and tax withholding from an employer. Key boxes: Box 1 — wages, tips, other compensation (federal taxable); Box 2 — federal income tax withheld; Box 3 — Social Security wages; Box 4 — Social Security tax withheld; Box 5 — Medicare wages; Box 6 — Medicare tax withheld; Box 12 — coded items (D=401k, DD=health insurance cost, W=HSA); Box 13 — checkboxes for statutory employee, retirement plan, third-party sick pay; Box 16 — state wages; Box 17 — state income tax withheld.",
			Source: "IRS Form W-2 Instructions", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSInstruction,
			Section: "W-2", Tags: []string{"W-2", "wages", "withholding", "employer", "Box 1", "Box 2", "income"},
		},
		{
			ID: "pub_withholding", Title: "Federal Income Tax Withholding",
			Content: "Federal income tax withholding is the amount your employer deducts from your paycheck and sends to the IRS on your behalf. The amount withheld depends on your Form W-4 selections (filing status, dependents, additional withholding). When you file your return, your total withholding (Box 2 of all W-2s) is compared to your total tax liability. If withholding exceeds your tax, you receive a refund. If it's less, you owe the difference. Underpayment may result in penalties.",
			Source: "IRS Pub 505", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSPublication,
			Section: "withholding", Tags: []string{"withholding", "W-4", "refund", "tax owed", "paycheck", "employer"},
		},
		{
			ID: "pub_filing_requirements", Title: "Who Must File a Return",
			Content: "Most US citizens and residents must file a federal return if their gross income exceeds the filing threshold. For 2025, single filers under 65 must file if gross income exceeds $15,000 (standard deduction amount). MFJ under 65: $30,000. HOH under 65: $22,500. You must also file if you had self-employment income of $400+, owe special taxes, or received advance premium tax credits. Even if not required, you should file to claim refundable credits like EITC.",
			Source: "IRS Pub 17, Ch. 1", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSPublication,
			Section: "filing", Tags: []string{"filing requirement", "who must file", "threshold", "gross income", "self-employment"},
		},
		{
			ID: "irc_86", Title: "Social Security Benefits Taxation",
			Content: "IRC §86 determines how much of Social Security benefits are taxable at the federal level. If your combined income (AGI + nontaxable interest + half of SS benefits) exceeds $25,000 (single) or $32,000 (MFJ), up to 50% of benefits may be taxable. If combined income exceeds $34,000 (single) or $44,000 (MFJ), up to 85% of benefits may be taxable. The calculation uses a two-tier formula on Form 1040 or the Social Security Benefits Worksheet.",
			Source: "IRC §86", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "86", Tags: []string{"Social Security", "benefits", "taxable", "combined income", "retirement"},
		},
		{
			ID: "irc_223", Title: "Health Savings Accounts (HSAs)",
			Content: "IRC §223 allows tax-deductible contributions to Health Savings Accounts for individuals with high-deductible health plans (HDHPs). For 2025, contribution limits are $4,300 (self-only) and $8,550 (family). Individuals 55+ can contribute an additional $1,000. Contributions are deductible above the line (reduce AGI). Earnings grow tax-free, and distributions for qualified medical expenses are tax-free. Non-medical distributions before age 65 are subject to income tax plus a 20% penalty.",
			Source: "IRC §223", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "223", Tags: []string{"HSA", "health savings account", "HDHP", "medical", "deduction", "contribution"},
		},
		{
			ID: "pub_1040_overview", Title: "Form 1040 Overview",
			Content: "Form 1040 is the main individual income tax return. It follows a flow: (1) Report income from all sources (wages, interest, dividends, capital gains, etc.) to get total income. (2) Subtract adjustments to income to get AGI (line 11). (3) Subtract the greater of standard or itemized deductions, plus QBI deduction, to get taxable income (line 15). (4) Calculate tax using tax tables or brackets. (5) Apply credits to reduce tax. (6) Compare tax to withholding/payments to determine refund or amount owed.",
			Source: "IRS Form 1040 Instructions", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSInstruction,
			Section: "1040", Tags: []string{"Form 1040", "overview", "income", "AGI", "deductions", "credits", "tax", "refund"},
		},
		{
			ID: "irc_1401", Title: "Self-Employment Tax",
			Content: "IRC §1401 imposes self-employment tax on net earnings from self-employment. The rate is 15.3% (12.4% Social Security on earnings up to $176,100 for 2025, plus 2.9% Medicare with no cap). An additional 0.9% Medicare tax applies to self-employment earnings over $200,000 (single) or $250,000 (MFJ). Self-employed individuals may deduct 50% of self-employment tax as an above-the-line deduction on Form 1040. Calculated on Schedule SE.",
			Source: "IRC §1401", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "1401", Tags: []string{"self-employment tax", "Schedule SE", "FICA", "Social Security", "Medicare", "sole proprietor"},
		},
		{
			ID: "irc_6012", Title: "Estimated Tax Payments",
			Content: "IRC §6654 requires estimated tax payments if you expect to owe $1,000 or more when you file. Estimated taxes are paid quarterly (April 15, June 15, September 15, January 15 of the following year). To avoid penalties, you must pay at least 90% of current year tax or 100% of prior year tax (110% if prior year AGI exceeded $150,000). Self-employed individuals, landlords, and investors commonly need to make estimated payments.",
			Source: "IRC §6654", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "6654", Tags: []string{"estimated tax", "quarterly", "penalty", "safe harbor", "self-employment"},
		},
		{
			ID: "pub_schedule_a", Title: "Itemized Deductions (Schedule A)",
			Content: "Schedule A is used to claim itemized deductions instead of the standard deduction. Common itemized deductions include: medical expenses exceeding 7.5% of AGI; state and local taxes (SALT, capped at $10,000); mortgage interest; charitable contributions; and casualty/theft losses from federally declared disasters. You should itemize only if your total itemized deductions exceed the standard deduction. Most taxpayers take the standard deduction since TCJA roughly doubled it.",
			Source: "IRS Schedule A Instructions", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSInstruction,
			Section: "Schedule A", Tags: []string{"itemized deductions", "Schedule A", "medical", "SALT", "mortgage", "charitable", "standard deduction"},
		},
		{
			ID: "irc_1_capital_gains", Title: "Capital Gains Tax Rates",
			Content: "Long-term capital gains (assets held over one year) are taxed at preferential rates: 0% if taxable income is below $48,350 (single) or $96,700 (MFJ) for 2025; 15% for income up to $533,400 (single) or $600,050 (MFJ); 20% above those thresholds. Short-term capital gains (assets held one year or less) are taxed as ordinary income. An additional 3.8% Net Investment Income Tax (NIIT) applies when MAGI exceeds $200,000 (single) or $250,000 (MFJ).",
			Source: "IRC §1(h)", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "1h", Tags: []string{"capital gains", "long-term", "short-term", "NIIT", "investment", "stock", "sale"},
		},
	}
}

// SeedCADocuments returns essential California tax knowledge documents.
func SeedCADocuments() []Document {
	return []Document{
		{
			ID: "ca_conformity", Title: "California IRC Conformity",
			Content: "California generally conforms to the Internal Revenue Code as of January 1, 2015, with specific modifications enacted through subsequent legislation. Key non-conformity areas include: California does not allow the qualified business income deduction under IRC §199A; California does not tax Social Security benefits; California has its own standard deduction amounts ($5,706 single, $11,412 MFJ for 2025); California does not conform to federal HSA treatment; California does not conform to TCJA's increased standard deduction or suspended personal exemption.",
			Source: "CA R&TC §17024.5", Jurisdiction: JurisdictionCA, DocType: DocTypeCARTCSection,
			Section: "17024.5", Tags: []string{"conformity", "IRC", "differences", "non-conformity", "federal"},
		},
		{
			ID: "ca_rates", Title: "California Income Tax Rates and Brackets",
			Content: "California has nine income tax brackets ranging from 1% to 12.3%. For 2025 single filers: 1% on income up to $10,756; 2% on $10,757-$25,499; 4% on $25,500-$40,245; 6% on $40,246-$55,866; 8% on $55,867-$70,612; 9.3% on $70,613-$360,659; 10.3% on $360,660-$432,787; 11.3% on $432,788-$721,314; 12.3% on income over $721,314. The Mental Health Services Tax adds an additional 1% surcharge on taxable income exceeding $1,000,000.",
			Source: "CA R&TC §17041", Jurisdiction: JurisdictionCA, DocType: DocTypeCARTCSection,
			Section: "17041", Tags: []string{"tax rates", "brackets", "California", "marginal rate", "progressive"},
		},
		{
			ID: "ca_mental_health", Title: "California Mental Health Services Tax",
			Content: "California imposes an additional 1% Mental Health Services Tax (Proposition 63) on taxable income exceeding $1,000,000. This applies regardless of filing status — the $1M threshold is the same for single, MFJ, MFS, and HOH filers. This surcharge is calculated on Form 540, line 36 and added to the base tax calculated using the regular bracket rates. High-income taxpayers should factor this into estimated tax calculations.",
			Source: "CA R&TC §17043", Jurisdiction: JurisdictionCA, DocType: DocTypeCARTCSection,
			Section: "17043", Tags: []string{"mental health tax", "surcharge", "Proposition 63", "millionaire", "1 percent", "high income"},
		},
		{
			ID: "ca_standard_deduction", Title: "California Standard Deduction",
			Content: "California's standard deduction for 2025 is $5,706 for single or married filing separately filers, and $11,412 for married filing jointly, qualifying surviving spouse, or head of household filers. These amounts are significantly lower than the federal standard deduction ($15,000 single / $30,000 MFJ). California did not conform to TCJA's increased standard deduction. Blind or elderly taxpayers may receive additional amounts.",
			Source: "CA R&TC §17073.5", Jurisdiction: JurisdictionCA, DocType: DocTypeCARTCSection,
			Section: "17073.5", Tags: []string{"standard deduction", "California", "deduction", "single", "married"},
		},
		{
			ID: "ca_exemption_credits", Title: "California Exemption Credits",
			Content: "Unlike the federal system which suspended personal exemptions, California provides exemption credits. For 2025: each personal exemption credit is $144 (single/MFS) or $288 (MFJ/QSS/HOH). Each dependent exemption credit is $433. These credits directly reduce tax liability. The credits phase out for high-income taxpayers: the phase-out begins at $212,288 (single) and $424,581 (MFJ).",
			Source: "CA R&TC §17054", Jurisdiction: JurisdictionCA, DocType: DocTypeCARTCSection,
			Section: "17054", Tags: []string{"exemption credit", "personal exemption", "dependent", "credit", "phase-out"},
		},
		{
			ID: "ca_social_security", Title: "California Social Security Benefits Exemption",
			Content: "California does not tax Social Security benefits. Unlike the federal government, which may tax up to 85% of Social Security benefits based on income, California fully exempts all Social Security benefits from state income tax. If you receive Social Security and file a California return, you subtract the full amount of your benefits on Schedule CA (540), Part I, Section B. This is one of the most significant California/federal differences for retirees.",
			Source: "CA R&TC §17087", Jurisdiction: JurisdictionCA, DocType: DocTypeCARTCSection,
			Section: "17087", Tags: []string{"Social Security", "exempt", "benefits", "retirement", "subtraction"},
		},
		{
			ID: "ca_schedule_ca", Title: "Schedule CA (540) — California Adjustments",
			Content: "Schedule CA (540) reconciles differences between federal and California tax law. It has two main sections: Part I (Income Adjustments) adjusts federal income items — additions increase California income, subtractions decrease it. Part II (Adjustments to Federal Itemized Deductions) adjusts itemized deductions where California differs from federal. Common adjustments include: adding back the QBI deduction (§199A), subtracting Social Security benefits, adjusting SALT deduction (CA does not cap state tax deduction since it doesn't allow it at all), and HSA adjustments.",
			Source: "FTB Schedule CA (540) Instructions", Jurisdiction: JurisdictionCA, DocType: DocTypeFTBInstruction,
			Section: "Schedule CA", Tags: []string{"Schedule CA", "adjustments", "additions", "subtractions", "conformity", "differences"},
		},
		{
			ID: "ca_salt", Title: "California SALT Deduction Treatment",
			Content: "California does not allow a deduction for state income taxes or state disability insurance (SDI) paid, since you cannot deduct California taxes on your California return. However, California does allow deductions for real property taxes and other local taxes without the federal $10,000 SALT cap. On Schedule CA, you add back the state income tax deduction claimed on federal Schedule A, and the federal $10,000 SALT cap does not apply for California purposes.",
			Source: "CA R&TC §17220", Jurisdiction: JurisdictionCA, DocType: DocTypeCARTCSection,
			Section: "17220", Tags: []string{"SALT", "state tax", "property tax", "deduction", "cap", "itemized"},
		},
		{
			ID: "ca_qbi", Title: "California QBI Deduction Non-Conformity",
			Content: "California does not conform to the federal Qualified Business Income (QBI) deduction under IRC §199A. If you claimed a QBI deduction on your federal return, you must add it back as a California adjustment on Schedule CA (540). This means pass-through business income is effectively taxed at the full California rate with no 20% deduction available. This can result in a significantly higher California tax liability for self-employed individuals and business owners.",
			Source: "CA R&TC §17024.5(a)", Jurisdiction: JurisdictionCA, DocType: DocTypeCARTCSection,
			Section: "17024.5_qbi", Tags: []string{"QBI", "qualified business income", "199A", "non-conformity", "add back", "Schedule CA"},
		},
		{
			ID: "ca_eitc", Title: "California Earned Income Tax Credit (CalEITC)",
			Content: "California offers its own Earned Income Tax Credit (CalEITC) in addition to the federal EITC. For 2025, CalEITC is available to filers with earned income up to approximately $30,950. The credit amount varies based on income and number of qualifying children. CalEITC is claimed on Form 3514. Unlike the federal EITC, CalEITC also allows ITIN filers to claim the credit. The Young Child Tax Credit (YCTC) provides up to $1,117 additional credit for filers with a child under age 6.",
			Source: "CA R&TC §17052.1", Jurisdiction: JurisdictionCA, DocType: DocTypeCARTCSection,
			Section: "17052.1", Tags: []string{"CalEITC", "earned income credit", "credit", "refundable", "ITIN", "young child"},
		},
		{
			ID: "ca_hsa", Title: "California HSA Non-Conformity",
			Content: "California does not conform to federal Health Savings Account (HSA) tax treatment. HSA contributions deducted on the federal return must be added back on the California return via Schedule CA. Additionally, HSA earnings are taxable by California, and California-qualified HSA distributions are not tax-free for state purposes. California treats HSAs similarly to regular investment accounts. This requires additional reporting on Schedule CA and potentially Form 3805P for early distribution penalties.",
			Source: "FTB Pub 1001", Jurisdiction: JurisdictionCA, DocType: DocTypeFTBPublication,
			Section: "HSA", Tags: []string{"HSA", "health savings account", "non-conformity", "add back", "California"},
		},
		{
			ID: "ca_540_overview", Title: "Form 540 Overview — California Resident Income Tax",
			Content: "Form 540 is the California resident income tax return. The filing flow: (1) Start with federal AGI from Form 1040 line 11. (2) Apply California adjustments from Schedule CA to get California AGI. (3) Subtract CA standard deduction or itemized deductions. (4) Calculate tax using CA brackets. (5) Add Mental Health Services Tax if applicable. (6) Apply exemption credits and other credits. (7) Compare to withholding (W-2 Box 17) and estimated payments. Filing is required if CA gross income exceeds $22,557 (single) or $45,114 (MFJ) for 2025.",
			Source: "FTB Form 540 Instructions", Jurisdiction: JurisdictionCA, DocType: DocTypeFTBInstruction,
			Section: "540", Tags: []string{"Form 540", "California", "resident", "filing", "overview", "AGI"},
		},
		{
			ID: "ca_muni_bonds", Title: "California Municipal Bond Interest",
			Content: "For California tax purposes, interest from bonds issued by California state or local governments is exempt from California income tax. However, interest from bonds issued by other states (e.g., New York, Texas municipal bonds) that is exempt from federal tax must be added to California income on Schedule CA. Conversely, if you hold US government bonds (Treasury bonds, savings bonds), that interest is taxable on the federal return but exempt from California tax.",
			Source: "CA R&TC §17133", Jurisdiction: JurisdictionCA, DocType: DocTypeCARTCSection,
			Section: "17133", Tags: []string{"municipal bonds", "interest", "exempt", "out-of-state", "Treasury bonds"},
		},
		{
			ID: "ca_withholding", Title: "California Income Tax Withholding",
			Content: "California state income tax is withheld from wages and reported in Box 17 of Form W-2 (with state wages in Box 16). California also requires withholding on non-wage payments over certain thresholds (7% default rate on payments to non-residents). SDI (State Disability Insurance) is reported in Box 14 or Box 19. The withholding amount is compared to your total CA tax on Form 540 to determine your refund or balance due.",
			Source: "FTB DE 4 Instructions", Jurisdiction: JurisdictionCA, DocType: DocTypeFTBInstruction,
			Section: "withholding", Tags: []string{"withholding", "California", "W-2", "Box 17", "SDI", "state tax"},
		},
		{
			ID: "ca_529", Title: "California 529 Plan Treatment",
			Content: "California does not provide a state income tax deduction for contributions to 529 education savings plans. While the federal tax code allows tax-free growth and tax-free withdrawals for qualified education expenses, California conforms to this treatment for distributions but offers no deduction for contributions (unlike many other states). ScholarShare 529 is California's plan, but contributions to any 529 plan receive the same California tax treatment.",
			Source: "FTB Pub 1005", Jurisdiction: JurisdictionCA, DocType: DocTypeFTBPublication,
			Section: "529", Tags: []string{"529", "education", "savings", "college", "ScholarShare", "no deduction"},
		},
	}
}

// SeedStore returns a Store pre-populated with essential federal and CA documents.
func SeedStore() *Store {
	s := NewStore()
	for _, doc := range SeedFederalDocuments() {
		s.Add(doc)
	}
	for _, doc := range SeedCADocuments() {
		s.Add(doc)
	}
	for _, doc := range SeedIRCSections() {
		s.Add(doc)
	}
	for _, doc := range SeedIRSPublications() {
		s.Add(doc)
	}
	for _, doc := range SeedFTBPublications() {
		s.Add(doc)
	}
	return s
}
