package knowledge

// SeedIRSPublications returns IRS publication document chunks for the knowledge base.
func SeedIRSPublications() []Document {
	return []Document{
		// === IRS Publication 17: Your Federal Income Tax ===
		{
			ID: "pub17_filing_requirements", Title: "Pub 17: Filing Requirements and Deadlines",
			Content: "Most U.S. citizens and permanent residents must file a federal income tax return if their gross income exceeds the filing threshold for their filing status and age. The standard filing deadline is April 15 (or the next business day if it falls on a weekend or holiday). An automatic 6-month extension to October 15 is available by filing Form 4868, but this extends the filing deadline only — taxes owed are still due by April 15. Failure to file timely may result in a failure-to-file penalty of 5% per month (up to 25%), while failure to pay results in a 0.5% per month penalty.",
			Source: "IRS Pub 17, Ch. 1", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSPublication,
			Section: "Pub17_filing", Tags: []string{"filing requirements", "deadline", "April 15", "extension", "Form 4868", "penalty", "who must file"},
		},
		{
			ID: "pub17_filing_status", Title: "Pub 17: Choosing the Right Filing Status",
			Content: "Your filing status determines your tax bracket thresholds, standard deduction amount, and eligibility for certain credits. The five statuses are: Single, Married Filing Jointly (MFJ), Married Filing Separately (MFS), Head of Household (HOH), and Qualifying Surviving Spouse (QSS). Head of Household requires being unmarried (or considered unmarried), paying more than half the cost of maintaining a home, and having a qualifying person live with you for more than half the year. If you qualify for more than one status, choose the one that results in the lowest tax.",
			Source: "IRS Pub 17, Ch. 2", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSPublication,
			Section: "Pub17_status", Tags: []string{"filing status", "single", "married", "head of household", "qualifying surviving spouse", "MFJ", "MFS"},
		},
		{
			ID: "pub17_dependents", Title: "Pub 17: Dependents and Exemptions",
			Content: "A dependent is either a qualifying child or qualifying relative. A qualifying child must meet tests for relationship, age (under 19, or under 24 if full-time student), residency (lived with you more than half the year), and support (child did not provide more than half their own support). A qualifying relative must have gross income below $5,150 (2025), receive more than half their support from you, and meet the relationship or member-of-household test. Each dependent must have a valid SSN or ITIN. While personal exemptions are $0 under TCJA, dependent status affects eligibility for credits like the Child Tax Credit.",
			Source: "IRS Pub 17, Ch. 3", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSPublication,
			Section: "Pub17_dependents", Tags: []string{"dependent", "qualifying child", "qualifying relative", "exemption", "support test", "relationship test", "SSN"},
		},
		{
			ID: "pub17_taxable_income", Title: "Pub 17: Taxable vs. Non-Taxable Income",
			Content: "Taxable income includes wages, salaries, tips, business income, capital gains, interest, dividends, rental income, alimony received (pre-2019 agreements), unemployment compensation, and gambling winnings. Non-taxable income includes gifts, inheritances, life insurance proceeds paid due to death, municipal bond interest, qualified Roth IRA distributions, child support, workers' compensation, and welfare benefits. Some income is partially taxable, such as Social Security benefits (up to 85% may be taxable depending on income). Scholarships used for tuition are generally tax-free, but amounts for room and board are taxable.",
			Source: "IRS Pub 17, Ch. 5-7", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSPublication,
			Section: "Pub17_income", Tags: []string{"taxable income", "non-taxable", "wages", "interest", "dividends", "gifts", "inheritance", "Social Security"},
		},
		{
			ID: "pub17_adjustments", Title: "Pub 17: Adjustments to Income (Above-the-Line Deductions)",
			Content: "Adjustments to income are deductions taken on Schedule 1 of Form 1040 that reduce your gross income to arrive at AGI. Key adjustments include: educator expenses (up to $300), HSA contributions, self-employment tax deduction (50% of SE tax), self-employed health insurance premiums, IRA deduction, student loan interest (up to $2,500), and alimony paid (pre-2019 agreements). These deductions are available regardless of whether you itemize or take the standard deduction. AGI is a critical figure because many other tax benefits phase out based on AGI levels.",
			Source: "IRS Pub 17, Ch. 10", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSPublication,
			Section: "Pub17_adjustments", Tags: []string{"adjustments", "above the line", "AGI", "Schedule 1", "educator", "HSA", "IRA", "student loan"},
		},
		{
			ID: "pub17_deductions", Title: "Pub 17: Standard vs. Itemized Deductions",
			Content: "After calculating AGI, you subtract either the standard deduction or itemized deductions (whichever is greater) to determine taxable income. The 2025 standard deduction is $15,000 (single), $30,000 (MFJ), $22,500 (HOH), and $15,000 (MFS). Additional amounts apply for age 65+ or blindness. Itemized deductions on Schedule A include: medical expenses exceeding 7.5% of AGI, state and local taxes (SALT, capped at $10,000), home mortgage interest, charitable contributions, and casualty/theft losses from federally declared disasters. Since TCJA roughly doubled the standard deduction, about 90% of taxpayers now use it.",
			Source: "IRS Pub 17, Ch. 11-12", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSPublication,
			Section: "Pub17_deductions", Tags: []string{"standard deduction", "itemized deductions", "Schedule A", "SALT", "mortgage interest", "charitable"},
		},
		{
			ID: "pub17_credits", Title: "Pub 17: Tax Credits Overview",
			Content: "Tax credits directly reduce your tax liability, dollar for dollar. Nonrefundable credits reduce tax to zero but not below: they include the Child and Dependent Care Credit (up to $3,000 for one, $6,000 for two+ dependents), Education Credits (American Opportunity up to $2,500 per student, Lifetime Learning up to $2,000), and the Retirement Savings Contributions Credit. Refundable credits can produce a refund even if you owe no tax: they include the Earned Income Tax Credit (EITC), the refundable portion of the Child Tax Credit, and the Premium Tax Credit for health insurance. Credits are generally more valuable than deductions of the same amount.",
			Source: "IRS Pub 17, Ch. 13-14", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSPublication,
			Section: "Pub17_credits", Tags: []string{"tax credits", "refundable", "nonrefundable", "child care", "education", "EITC", "American Opportunity", "Lifetime Learning"},
		},
		{
			ID: "pub17_withholding_estimated", Title: "Pub 17: Withholding and Estimated Taxes",
			Content: "Federal income tax is primarily collected through withholding from wages (per Form W-4 elections) and estimated tax payments. When you file your return, total tax is compared to amounts already paid (withholding plus estimated payments). If you paid more, you receive a refund; if less, you owe the balance. Self-employed individuals, investors, and retirees without sufficient withholding must make quarterly estimated tax payments using Form 1040-ES. Underpayment of estimated tax may result in a penalty calculated on Form 2210.",
			Source: "IRS Pub 17, Ch. 4", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSPublication,
			Section: "Pub17_withholding", Tags: []string{"withholding", "estimated tax", "W-4", "refund", "balance due", "Form 1040-ES", "quarterly"},
		},
		{
			ID: "pub17_income_types", Title: "Pub 17: Types of Investment Income",
			Content: "Investment income includes interest, dividends, and capital gains. Ordinary interest (bank accounts, CDs) is fully taxable. Qualified dividends from domestic corporations and certain foreign corporations are taxed at preferential capital gains rates (0%, 15%, or 20%). Non-qualified (ordinary) dividends are taxed at regular rates. Interest and dividends are reported on Forms 1099-INT and 1099-DIV respectively. Capital gains and losses from sales of investment property are reported on Schedule D with details on Form 8949.",
			Source: "IRS Pub 17, Ch. 7-8", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSPublication,
			Section: "Pub17_investment", Tags: []string{"interest", "dividends", "capital gains", "investment income", "1099-INT", "1099-DIV", "Schedule D", "qualified dividends"},
		},
		{
			ID: "pub17_record_keeping", Title: "Pub 17: Record Keeping for Tax Returns",
			Content: "Taxpayers should keep records that support income, deductions, and credits claimed on their return. Generally, records should be retained for 3 years from the filing date (matching the standard IRS audit statute of limitations). Keep records for 6 years if you underreported income by more than 25%, and indefinitely if you filed a fraudulent return or did not file. Important records include W-2s, 1099s, receipts for deductible expenses, home purchase/improvement records, investment purchase records, and prior year tax returns.",
			Source: "IRS Pub 17, Ch. 1", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSPublication,
			Section: "Pub17_records", Tags: []string{"record keeping", "retention", "audit", "statute of limitations", "receipts", "documentation"},
		},

		// === IRS Publication 334: Tax Guide for Small Business ===
		{
			ID: "pub334_self_employed", Title: "Pub 334: Who Is Self-Employed",
			Content: "You are self-employed if you carry on a trade or business as a sole proprietor, independent contractor, member of a partnership, or are otherwise in business for yourself. The IRS uses factors like behavioral control, financial control, and relationship type to distinguish employees from independent contractors. Self-employed individuals report business income and expenses on Schedule C (or Schedule C-EZ) and must pay self-employment tax on net earnings of $400 or more. Gig economy workers (rideshare drivers, freelancers, etc.) are generally considered self-employed.",
			Source: "IRS Pub 334, Ch. 1", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSPublication,
			Section: "Pub334_self_employed", Tags: []string{"self-employed", "independent contractor", "sole proprietor", "Schedule C", "gig economy", "freelancer"},
		},
		{
			ID: "pub334_business_income", Title: "Pub 334: Business Income Reporting",
			Content: "Business income includes all income received from your trade or business, including payments for services, sales of products, barter income, and recoveries of previously deducted amounts. Report gross receipts on Schedule C, line 1. If you received payments of $600 or more from a single payer, they should issue Form 1099-NEC (for services) or 1099-K (for payment card/third-party network transactions exceeding thresholds). Even if you do not receive a 1099, you must report all business income. The method of accounting (cash or accrual) determines when income is recognized.",
			Source: "IRS Pub 334, Ch. 5", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSPublication,
			Section: "Pub334_income", Tags: []string{"business income", "Schedule C", "1099-NEC", "1099-K", "gross receipts", "reporting", "cash method", "accrual"},
		},
		{
			ID: "pub334_business_expenses", Title: "Pub 334: Business Expenses (Ordinary and Necessary)",
			Content: "Business expenses must be both ordinary (common and accepted in your trade) and necessary (helpful and appropriate for your business) to be deductible on Schedule C. Common deductible expenses include advertising, car and truck expenses (standard mileage rate of $0.70/mile for 2025 or actual expenses), insurance, legal and professional fees, office supplies, rent, utilities, and wages paid to employees. Startup costs up to $5,000 can be deducted in the first year, with the remainder amortized over 180 months. Capital expenditures must generally be depreciated rather than expensed immediately.",
			Source: "IRS Pub 334, Ch. 8", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSPublication,
			Section: "Pub334_expenses", Tags: []string{"business expenses", "ordinary and necessary", "Schedule C", "mileage", "deduction", "advertising", "rent", "startup costs"},
		},
		{
			ID: "pub334_se_tax", Title: "Pub 334: Self-Employment Tax Basics",
			Content: "Self-employment tax is the Social Security and Medicare tax for self-employed individuals, calculated on Schedule SE. The SE tax rate is 15.3% (12.4% Social Security + 2.9% Medicare) on 92.35% of net self-employment earnings. For 2025, Social Security tax applies to the first $176,100 of combined wages and SE earnings. The Additional Medicare Tax of 0.9% applies to SE earnings over $200,000 (single) or $250,000 (MFJ). You can deduct 50% of your SE tax as an above-the-line adjustment on Schedule 1. SE tax is owed even if no income tax is due.",
			Source: "IRS Pub 334, Ch. 12", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSPublication,
			Section: "Pub334_se_tax", Tags: []string{"self-employment tax", "Schedule SE", "15.3 percent", "Social Security", "Medicare", "deduction"},
		},
		{
			ID: "pub334_record_keeping", Title: "Pub 334: Record Keeping Requirements for Business",
			Content: "Self-employed individuals must maintain adequate records to substantiate income and deductions. Records should include a daily summary of business receipts, documentation of all expenses (receipts, canceled checks, bank statements), an asset record for depreciation, and records of employment taxes if you have employees. Use a consistent accounting method and keep records for at least 3 years after filing. Good records help prepare accurate tax returns, monitor business performance, and support deductions if audited.",
			Source: "IRS Pub 334, Ch. 1", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSPublication,
			Section: "Pub334_records", Tags: []string{"record keeping", "business records", "receipts", "documentation", "audit", "self-employed"},
		},
		{
			ID: "pub334_home_office", Title: "Pub 334: Home Office Deduction",
			Content: "If you use part of your home regularly and exclusively for business, you may deduct home office expenses. There are two methods: the simplified method allows $5 per square foot up to 300 sq ft ($1,500 maximum), and the regular method calculates actual expenses (mortgage interest, rent, utilities, insurance, depreciation) based on the percentage of home used for business. The home office must be your principal place of business, a place where you regularly meet clients, or a separate structure used for business. Employees working from home generally cannot deduct home office expenses under TCJA.",
			Source: "IRS Pub 334, Ch. 8", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSPublication,
			Section: "Pub334_home_office", Tags: []string{"home office", "deduction", "simplified method", "regular method", "exclusive use", "business use"},
		},

		// === IRS Publication 505: Tax Withholding and Estimated Tax ===
		{
			ID: "pub505_withholding", Title: "Pub 505: How Withholding Works",
			Content: "Federal income tax withholding is based on your Form W-4 selections. The 2020 redesigned W-4 uses filing status, multiple jobs adjustments, dependent credits, other income, and additional withholding rather than the old allowance system. Your employer uses IRS withholding tables to calculate the amount to withhold from each paycheck. You can adjust withholding at any time by submitting a new W-4. If you have non-wage income (interest, dividends, capital gains), you can request additional withholding on W-4 line 4(c) to avoid owing at filing time.",
			Source: "IRS Pub 505, Ch. 1", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSPublication,
			Section: "Pub505_withholding", Tags: []string{"withholding", "W-4", "employer", "paycheck", "wage", "adjustment"},
		},
		{
			ID: "pub505_estimated_when", Title: "Pub 505: When Estimated Tax Is Required",
			Content: "You must make estimated tax payments if you expect to owe at least $1,000 in tax after subtracting withholding and credits, AND you expect your withholding plus credits to be less than the smaller of 90% of current year tax or 100% of prior year tax (110% if prior year AGI exceeded $150,000). Estimated tax payments are typically required for self-employed individuals, freelancers, landlords, investors with significant gains, and retirees without sufficient pension withholding. Payments are due quarterly: April 15, June 15, September 15, and January 15.",
			Source: "IRS Pub 505, Ch. 2", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSPublication,
			Section: "Pub505_when", Tags: []string{"estimated tax", "quarterly", "1000 threshold", "self-employed", "due dates", "required"},
		},
		{
			ID: "pub505_estimated_how", Title: "Pub 505: How to Figure Estimated Tax",
			Content: "To calculate estimated tax, use Form 1040-ES worksheet: estimate your expected AGI, deductions, and credits for the year, then compute your expected tax liability. Subtract expected withholding to determine the estimated tax you need to pay. Divide the annual amount by four for equal quarterly payments, or use the annualized income installment method (Schedule AI of Form 2210) if your income is uneven throughout the year. You can pay online via IRS Direct Pay, EFTPS, or by mailing a check with a 1040-ES voucher.",
			Source: "IRS Pub 505, Ch. 2", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSPublication,
			Section: "Pub505_how", Tags: []string{"estimated tax", "Form 1040-ES", "calculation", "quarterly payments", "annualized", "Direct Pay", "EFTPS"},
		},
		{
			ID: "pub505_underpayment", Title: "Pub 505: Underpayment Penalty and Safe Harbor Rules",
			Content: "If you do not pay enough tax through withholding and estimated payments, you may owe an underpayment penalty calculated on Form 2210. The penalty is essentially interest on the underpaid amount for the period it was underpaid, using the federal short-term rate plus 3 percentage points. Safe harbor rules let you avoid the penalty: pay at least 90% of current year tax, or 100% of prior year tax (110% if prior year AGI was over $150,000). The penalty may be waived for casualty, disaster, retirement (age 62+), or other unusual circumstances. No penalty applies if total tax owed minus withholding/credits is under $1,000.",
			Source: "IRS Pub 505, Ch. 4", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSPublication,
			Section: "Pub505_penalty", Tags: []string{"underpayment penalty", "safe harbor", "Form 2210", "90 percent", "100 percent", "110 percent", "waiver"},
		},
		{
			ID: "pub505_special_withholding", Title: "Pub 505: Special Withholding Situations",
			Content: "Certain types of income have special withholding rules. Pension and annuity payments are subject to withholding unless you elect out using Form W-4P. Social Security benefits may have voluntary withholding via Form W-4V at rates of 7%, 10%, 12%, or 22%. Backup withholding at 24% applies to interest, dividends, and other payments if you fail to provide a correct TIN to the payer. Supplemental wages (bonuses, commissions) may be withheld at a flat 22% rate (37% for amounts over $1 million) or aggregated with regular wages.",
			Source: "IRS Pub 505, Ch. 1", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSPublication,
			Section: "Pub505_special", Tags: []string{"withholding", "pension", "Social Security", "backup withholding", "supplemental wages", "bonus", "W-4P"},
		},
	}
}
