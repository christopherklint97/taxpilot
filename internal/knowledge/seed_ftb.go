package knowledge

// SeedFTBPublications returns FTB publication document chunks for the knowledge base.
func SeedFTBPublications() []Document {
	return []Document{
		// === FTB Publication 1001: Supplemental Guidelines to California Adjustments ===
		{
			ID: "ftb1001_overview", Title: "FTB Pub 1001: Income Adjustments Overview",
			Content: "FTB Publication 1001 provides guidelines for adjustments California taxpayers must make when their federal and California income differ. California uses federal AGI as the starting point on Form 540, then applies adjustments via Schedule CA (540). Adjustments fall into two categories: additions (income items California taxes but the federal government does not, or deductions California does not allow) and subtractions (income items California exempts or deductions California allows beyond federal). Every California resident filing a return should review whether Schedule CA adjustments apply.",
			Source: "FTB Pub 1001", Jurisdiction: JurisdictionCA, DocType: DocTypeFTBPublication,
			Section: "Pub1001_overview", Tags: []string{"FTB", "adjustments", "additions", "subtractions", "Schedule CA", "California", "income"},
		},
		{
			ID: "ftb1001_fed_ca_diff", Title: "FTB Pub 1001: Federal/California Differences Summary",
			Content: "Key differences between federal and California tax law include: California does not tax Social Security benefits (subtract on Schedule CA); California does not allow the QBI deduction under IRC §199A (add back on Schedule CA); California does not conform to federal HSA treatment (add back contributions and earnings); California has its own standard deduction ($5,706 single, $11,412 MFJ for 2025); California does not conform to the $10,000 SALT cap (no cap on property taxes, but no deduction for CA income tax); California maintains personal exemption credits while federal exemption is $0 under TCJA.",
			Source: "FTB Pub 1001", Jurisdiction: JurisdictionCA, DocType: DocTypeFTBPublication,
			Section: "Pub1001_differences", Tags: []string{"differences", "conformity", "Social Security", "QBI", "HSA", "SALT", "standard deduction", "exemption"},
		},
		{
			ID: "ftb1001_hsa", Title: "FTB Pub 1001: HSA Non-Conformity Details",
			Content: "California does not conform to the federal tax treatment of Health Savings Accounts (HSAs). For California purposes: HSA contributions deducted on the federal return must be added back on Schedule CA; employer HSA contributions excluded from federal wages must be added to California income; earnings (interest, dividends, capital gains) within the HSA are taxable by California in the year earned; and distributions, even for qualified medical expenses, may be taxable for California. Taxpayers should track HSA activity separately for California reporting and may need to file Form 3805P if early withdrawal penalties apply.",
			Source: "FTB Pub 1001", Jurisdiction: JurisdictionCA, DocType: DocTypeFTBPublication,
			Section: "Pub1001_hsa", Tags: []string{"HSA", "non-conformity", "add back", "contributions", "earnings", "California", "Schedule CA"},
		},
		{
			ID: "ftb1001_qbi", Title: "FTB Pub 1001: QBI Non-Conformity Details",
			Content: "California does not allow the federal Qualified Business Income (QBI) deduction under IRC §199A. If you claimed a QBI deduction on your federal return (Form 1040, line 13), you must add the entire amount back on Schedule CA (540), Part I, line 15. This applies to all pass-through business income including sole proprietorships (Schedule C), partnerships (Schedule E), S corporations, and qualified REIT dividends. The add-back increases California taxable income relative to federal, often resulting in a higher effective state tax rate for business owners.",
			Source: "FTB Pub 1001", Jurisdiction: JurisdictionCA, DocType: DocTypeFTBPublication,
			Section: "Pub1001_qbi", Tags: []string{"QBI", "199A", "non-conformity", "add back", "pass-through", "Schedule CA", "business income"},
		},
		{
			ID: "ftb1001_schedule_ca", Title: "FTB Pub 1001: Schedule CA Filing Requirements",
			Content: "You must file Schedule CA (540) if your California income differs from your federal income in any way. Common situations requiring Schedule CA include: receiving Social Security benefits, claiming the QBI deduction federally, making HSA contributions, having out-of-state municipal bond interest, claiming educator expenses (California does not conform), or having differences in itemized deductions such as the SALT cap adjustment. Schedule CA has two main parts: Part I adjusts income items, and Part II adjusts itemized deductions. If you take the California standard deduction, you only need Part I.",
			Source: "FTB Pub 1001", Jurisdiction: JurisdictionCA, DocType: DocTypeFTBPublication,
			Section: "Pub1001_schedule_ca", Tags: []string{"Schedule CA", "filing", "requirements", "when to file", "adjustments", "Part I", "Part II"},
		},
		{
			ID: "ftb1001_itemized", Title: "FTB Pub 1001: Itemized Deduction Adjustments",
			Content: "California itemized deductions differ from federal in several ways. The federal $10,000 SALT cap does not apply to California — you can deduct unlimited real property taxes and other local taxes, but you cannot deduct California state income tax or SDI on your California return. Mortgage interest rules generally conform to federal limits. Medical expenses follow the federal 7.5% AGI floor. Charitable contributions generally conform. California does not allow deductions for foreign taxes paid (add back on Schedule CA). Miscellaneous itemized deductions subject to the 2% AGI floor, eliminated federally by TCJA, remain available in California.",
			Source: "FTB Pub 1001", Jurisdiction: JurisdictionCA, DocType: DocTypeFTBPublication,
			Section: "Pub1001_itemized", Tags: []string{"itemized deductions", "SALT", "property tax", "mortgage interest", "charitable", "miscellaneous", "2 percent floor"},
		},
		{
			ID: "ftb1001_wages", Title: "FTB Pub 1001: Wage and Income Differences",
			Content: "Most wage income is the same for federal and California purposes, but differences can arise. California does not conform to the exclusion for employer-provided HSA contributions (add to CA wages). California conforms to the exclusion for qualified educational assistance up to $5,250 and most fringe benefits under IRC §132. California wages are reported in W-2 Box 16 and may differ from federal wages in Box 1 due to these non-conformity items. State disability insurance (SDI) withheld is shown in Box 14 and is not deductible on the California return.",
			Source: "FTB Pub 1001", Jurisdiction: JurisdictionCA, DocType: DocTypeFTBPublication,
			Section: "Pub1001_wages", Tags: []string{"wages", "W-2", "Box 16", "SDI", "HSA", "income differences", "California"},
		},
		{
			ID: "ftb1001_moving", Title: "FTB Pub 1001: Moving Expenses",
			Content: "The federal deduction for moving expenses is suspended under TCJA through 2025, except for active-duty military members. California, however, does not conform to this suspension and continues to allow the moving expense deduction for all taxpayers who meet the distance and time tests (new workplace must be at least 50 miles farther from old home, and you must work full-time for at least 39 weeks in the 12 months after the move). Qualifying moving expenses are subtracted on Schedule CA as a California adjustment.",
			Source: "FTB Pub 1001", Jurisdiction: JurisdictionCA, DocType: DocTypeFTBPublication,
			Section: "Pub1001_moving", Tags: []string{"moving expenses", "non-conformity", "TCJA", "subtraction", "Schedule CA", "distance test"},
		},

		// === FTB Publication 1005: Pension and Annuity Guidelines ===
		{
			ID: "ftb1005_retirement", Title: "FTB Pub 1005: California Treatment of Retirement Income",
			Content: "California generally follows federal rules for taxing pension, annuity, and retirement plan distributions. Traditional IRA distributions, 401(k) withdrawals, and pension payments are taxable by California. Roth IRA qualified distributions are tax-free for both federal and California purposes. California taxes the same portion of Social Security that the federal government does not — but since California fully exempts Social Security benefits, this is a subtraction rather than a conformity. Railroad retirement benefits Tier 1 are treated like Social Security (exempt). Early distribution penalties generally conform to federal rules.",
			Source: "FTB Pub 1005", Jurisdiction: JurisdictionCA, DocType: DocTypeFTBPublication,
			Section: "Pub1005_retirement", Tags: []string{"pension", "annuity", "retirement", "IRA", "401k", "Roth", "distribution", "California"},
		},
		{
			ID: "ftb1005_fed_diff", Title: "FTB Pub 1005: CA/Federal Differences for Pensions",
			Content: "Key California/federal differences for retirement income: California fully exempts Social Security benefits while federal may tax up to 85%. California conforms to federal IRA contribution and deduction limits. California does not conform to the federal exclusion of HSA distributions, so HSA withdrawals may be taxable at the state level even if tax-free federally. For government pensions, California follows the same rules as federal for taxing public employee retirement (CalPERS, CalSTRS distributions are fully taxable). Out-of-state government pensions received by California residents are also taxable by California under the reciprocity rules, but the source state generally cannot tax them if you are a California resident.",
			Source: "FTB Pub 1005", Jurisdiction: JurisdictionCA, DocType: DocTypeFTBPublication,
			Section: "Pub1005_differences", Tags: []string{"pension", "Social Security", "IRA", "HSA", "CalPERS", "CalSTRS", "federal differences"},
		},
		{
			ID: "ftb1005_early_withdrawal", Title: "FTB Pub 1005: Early Withdrawal Penalties",
			Content: "California generally conforms to the federal 10% early withdrawal penalty on distributions from retirement accounts taken before age 59 1/2. Exceptions that apply for both federal and California include: death, disability, substantially equal periodic payments (72(t) distributions), medical expenses exceeding 7.5% of AGI, qualified first-time homebuyer distributions (IRA only, up to $10,000), and qualified higher education expenses. California also conforms to the SECURE Act provisions raising the required minimum distribution (RMD) age to 73. The penalty is reported on California Form 3805P.",
			Source: "FTB Pub 1005", Jurisdiction: JurisdictionCA, DocType: DocTypeFTBPublication,
			Section: "Pub1005_penalty", Tags: []string{"early withdrawal", "penalty", "10 percent", "retirement", "exceptions", "Form 3805P", "RMD"},
		},

		// === FTB Publication 1031: Guidelines for Determining Resident Status ===
		{
			ID: "ftb1031_resident", Title: "FTB Pub 1031: California Resident Definition",
			Content: "A California resident is any individual who is in California for other than a temporary or transitory purpose, or any individual domiciled in California who is outside the state for a temporary or transitory purpose. Domicile is your true, fixed, and permanent home — the place you intend to return to whenever absent. Factors used to determine domicile include: where you maintain your largest home, where your spouse and children live, where you are registered to vote, where your vehicles are registered, where your bank accounts are, and where you hold professional licenses. California residents are taxed on worldwide income regardless of source.",
			Source: "FTB Pub 1031", Jurisdiction: JurisdictionCA, DocType: DocTypeFTBPublication,
			Section: "Pub1031_resident", Tags: []string{"resident", "domicile", "California", "worldwide income", "temporary", "transitory", "factors"},
		},
		{
			ID: "ftb1031_part_year", Title: "FTB Pub 1031: Part-Year and Nonresident Rules",
			Content: "Part-year residents are individuals who moved into or out of California during the tax year. They file Form 540NR and are taxed on all income received while a California resident, plus California-source income received while a nonresident. Nonresidents are taxed only on income from California sources (wages for work performed in CA, CA rental income, CA business income, gain from CA real property sales). Nonresidents also file Form 540NR. The tax is calculated on total income (for rate purposes) but only the California-source portion is actually taxed.",
			Source: "FTB Pub 1031", Jurisdiction: JurisdictionCA, DocType: DocTypeFTBPublication,
			Section: "Pub1031_part_year", Tags: []string{"part-year resident", "nonresident", "Form 540NR", "California-source", "income allocation", "moving"},
		},
		{
			ID: "ftb1031_safe_harbor", Title: "FTB Pub 1031: Safe Harbor Rules for Residency",
			Content: "California provides safe harbor rules for determining residency in certain situations. If you are domiciled in California but are outside the state for at least 546 consecutive days under an employment contract, you may be considered a nonresident for that period (safe harbor). You must not spend more than 45 days in California during any taxable year during the contract period. Days spent in California for medical treatment, natural disasters, or certain family emergencies may be excluded from the 45-day count. The safe harbor does not apply to individuals whose income is primarily from intangible property (investment income).",
			Source: "FTB Pub 1031", Jurisdiction: JurisdictionCA, DocType: DocTypeFTBPublication,
			Section: "Pub1031_safe_harbor", Tags: []string{"safe harbor", "546 days", "residency", "nonresident", "employment contract", "45 days", "domicile"},
		},
	}
}
