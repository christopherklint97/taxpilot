package knowledge

// SeedExpatDocuments returns knowledge documents for Americans living abroad.
func SeedExpatDocuments() []Document {
	return []Document{
		{
			ID: "irc_911", Title: "Foreign Earned Income Exclusion (FEIE)",
			Content: "IRC §911 allows qualifying US citizens and residents living abroad to exclude up to $130,000 (2025) of foreign earned income from US taxable income. To qualify, you must have a tax home in a foreign country and meet either the Bona Fide Residence Test (BFR — established residence in a foreign country for an uninterrupted period that includes an entire tax year) or the Physical Presence Test (PPT — present in a foreign country for at least 330 full days during a 12-month period). The exclusion is claimed on Form 2555. If you don't qualify for the full year, the exclusion is prorated based on qualifying days. The FEIE also includes a housing exclusion/deduction for housing expenses above a base amount ($20,800 for 2025).",
			Source: "IRC §911", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "911", Tags: []string{"FEIE", "foreign earned income", "exclusion", "expat", "abroad", "Form 2555", "bona fide residence", "physical presence"},
		},
		{
			ID: "irc_901", Title: "Foreign Tax Credit (FTC)",
			Content: "IRC §901 allows a credit for income taxes paid or accrued to foreign countries. The credit is limited to the US tax attributable to foreign source income: FTC limit = US tax × (foreign source taxable income / worldwide taxable income). Excess credits can be carried back 1 year or forward 10 years. You cannot claim both the FEIE and FTC on the same income — if you exclude income under §911, you cannot also claim a credit for taxes paid on that excluded income. The FTC is claimed on Form 1116 and flows to Schedule 3, line 1 of Form 1040.",
			Source: "IRC §901/903", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "901", Tags: []string{"foreign tax credit", "FTC", "Form 1116", "credit", "limitation", "carryforward", "foreign taxes paid"},
		},
		{
			ID: "irc_6038d", Title: "FATCA — Statement of Foreign Financial Assets",
			Content: "IRC §6038D requires US persons to report specified foreign financial assets on Form 8938 if the total value exceeds the applicable threshold. For taxpayers living abroad: $200,000 at year-end or $300,000 at any time (single), $400,000/$600,000 (MFJ). For US residents: $50,000/$75,000 (single), $100,000/$150,000 (MFJ). Specified foreign assets include bank accounts, brokerage accounts, foreign stocks/securities held directly, interests in foreign entities, and foreign pension plans. Failure to file carries a $10,000 penalty per year, plus potential additional penalties up to $50,000 for continued non-filing after notice.",
			Source: "IRC §6038D", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "6038D", Tags: []string{"FATCA", "Form 8938", "foreign assets", "reporting", "threshold", "penalty", "bank accounts"},
		},
		{
			ID: "pub_54", Title: "Tax Guide for US Citizens Abroad",
			Content: "IRS Publication 54 is the comprehensive guide for Americans living and working overseas. Key topics: (1) Filing requirements — US citizens must file a return regardless of where they live if income exceeds filing thresholds. (2) Automatic extension — expats get an automatic 2-month extension to June 15 (with further extension available to October 15). (3) Foreign earned income exclusion (Form 2555) — up to $130,000 for 2025. (4) Foreign tax credit (Form 1116) — credit for taxes paid to foreign governments. (5) Tax treaties may modify rules. (6) Foreign financial account reporting (FBAR/FATCA). (7) Self-employment tax still applies unless covered by a Totalization Agreement.",
			Source: "IRS Publication 54", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSPublication,
			Section: "Pub54", Tags: []string{"Publication 54", "expat", "abroad", "citizens", "filing", "extension", "FEIE", "FTC"},
		},
		{
			ID: "pub_514", Title: "Foreign Tax Credit for Individuals",
			Content: "IRS Publication 514 explains the foreign tax credit in detail. The credit applies to income taxes (or taxes in lieu of income taxes) paid or accrued to any foreign country or US possession. You can choose to take the credit or deduct foreign taxes (but not both). The credit is generally more beneficial. The limitation formula prevents the FTC from reducing US tax on US-source income. Separate limitation categories exist for different types of income (general, passive, etc.). If your foreign tax exceeds the limitation, the excess can be carried back 1 year or forward 10 years.",
			Source: "IRS Publication 514", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSPublication,
			Section: "Pub514", Tags: []string{"Publication 514", "foreign tax credit", "credit", "limitation", "carryforward", "categories"},
		},
		{
			ID: "fbar_fincen", Title: "FBAR (FinCEN Form 114) Requirements",
			Content: "The Report of Foreign Bank and Financial Accounts (FBAR) must be filed by any US person who has a financial interest in or signature authority over foreign financial accounts with an aggregate value exceeding $10,000 at any time during the calendar year. The FBAR is filed electronically with FinCEN (NOT the IRS) through the BSA E-Filing System at bsaefiling.fincen.treas.gov. The deadline is April 15 with an automatic extension to October 15 (no request needed). Penalties for non-filing can be severe: up to $10,000 per violation for non-willful failures, and up to the greater of $100,000 or 50% of account balances for willful violations. FBAR is separate from FATCA (Form 8938) — both may be required.",
			Source: "31 USC §5314 / 31 CFR §1010.350", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSPublication,
			Section: "FBAR", Tags: []string{"FBAR", "FinCEN", "foreign accounts", "bank accounts", "reporting", "penalty", "$10,000"},
		},
		{
			ID: "us_sweden_treaty", Title: "US-Sweden Tax Treaty Overview",
			Content: "The US-Sweden Income Tax Convention provides rules for avoiding double taxation. Key articles: Article 15 (Employment Income) — employment income is generally taxable where earned, with exceptions for short-term assignments. Article 18 (Pensions) — Swedish pensions paid to US residents may be taxable only in the US, but Swedish social security benefits may be taxable in Sweden. Article 23 (Elimination of Double Taxation) — the US allows a credit for Swedish taxes. Article 24 (Non-Discrimination). A treaty-based position must be disclosed on Form 8833. The US-Sweden Totalization Agreement coordinates Social Security between the two countries, potentially exempting expats from one country's social security contributions.",
			Source: "US-Sweden Income Tax Convention", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSPublication,
			Section: "Treaty-Sweden", Tags: []string{"Sweden", "treaty", "tax treaty", "double taxation", "pension", "Totalization", "Form 8833"},
		},
		{
			ID: "swedish_pension_treaty", Title: "Swedish Pension Treaty Treatment",
			Content: "Under the US-Sweden tax treaty (Article 18), pensions paid for services rendered in Sweden to a US resident are generally taxable only in the US. Swedish social insurance pensions (allmän pension) may be treated differently — they can be taxable in Sweden under Article 18(2). US citizens claiming treaty benefits on pension income must disclose the position on Form 8833. For IRA/401k purposes, Swedish pension contributions may not be deductible on the US return unless the treaty applies. Swedish PPM funds and tjänstepension should be analyzed individually for US tax treatment.",
			Source: "US-Sweden Treaty, Article 18", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSPublication,
			Section: "Treaty-Sweden-Pension", Tags: []string{"pension", "Sweden", "treaty", "allmän pension", "tjänstepension", "Article 18", "retirement"},
		},
		{
			ID: "ca_feie_nonconformity", Title: "California FEIE Non-Conformity",
			Content: "California does NOT conform to the federal Foreign Earned Income Exclusion (IRC §911). If you claimed the FEIE on your federal return, you must add the entire excluded amount back on Schedule CA (540). This means your California AGI will be higher than your federal AGI by the amount of the exclusion, often resulting in significant California tax liability even if your federal tax is zero. For example, excluding $120,000 on the federal return results in $120,000 being added back for California, potentially generating $7,000+ in CA tax. California also does not allow the foreign housing exclusion or deduction.",
			Source: "CA R&TC §17024.5", Jurisdiction: JurisdictionCA, DocType: DocTypeCARTCSection,
			Section: "17024.5_feie", Tags: []string{"FEIE", "California", "non-conformity", "add back", "Schedule CA", "foreign income", "exclusion"},
		},
		{
			ID: "currency_conversion", Title: "Currency Conversion Rules for Foreign Income",
			Content: "Foreign income must be reported in US dollars on the tax return. The IRS generally requires using the exchange rate prevailing when the income was received, paid, or accrued. For wages paid throughout the year, you may use the yearly average exchange rate published by the IRS. The IRS publishes yearly average exchange rates at irs.gov/individuals/international-taxpayers/yearly-average-currency-exchange-rates. For large one-time transactions, use the spot rate on the date of the transaction. Foreign taxes paid should also be converted to USD, using the rate on the date the taxes were paid or accrued.",
			Source: "IRS Publication 54, Ch. 1", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSPublication,
			Section: "Currency", Tags: []string{"currency", "exchange rate", "conversion", "foreign income", "USD", "IRS rate"},
		},
		{
			ID: "expat_deadlines", Title: "Expat Filing Deadlines and Extensions",
			Content: "US citizens and residents living abroad get an automatic 2-month extension to file (June 15 instead of April 15), but must still pay any tax owed by April 15 to avoid interest. You can request a further extension to October 15 using Form 4868. The FBAR deadline is April 15 with an automatic extension to October 15. Form 8938 (FATCA) is due with your tax return. If you need even more time, Form 2350 provides a special extension for expats who expect to qualify for the FEIE or foreign housing exclusion but haven't yet met the time requirements.",
			Source: "IRS Publication 54, Ch. 1", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRSPublication,
			Section: "Deadlines-Expat", Tags: []string{"deadline", "extension", "June 15", "expat", "abroad", "Form 4868", "Form 2350", "filing date"},
		},
	}
}
