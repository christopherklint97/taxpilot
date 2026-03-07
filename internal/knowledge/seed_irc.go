package knowledge

// SeedIRCSections returns additional IRC section documents for the knowledge base.
func SeedIRCSections() []Document {
	return []Document{
		{
			ID: "irc_103", Title: "Tax-Exempt Interest",
			Content: "IRC §103 excludes from gross income interest on state and local government bonds (municipal bonds). This includes bonds issued by states, cities, counties, and other political subdivisions for public purposes. However, interest on private activity bonds may be subject to the Alternative Minimum Tax (AMT). Taxpayers must still report tax-exempt interest on Form 1040, line 2a, even though it is not taxable. For 2025, tax-exempt interest may affect the taxation of Social Security benefits and other income-based calculations.",
			Source: "IRC §103", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "103", Tags: []string{"tax-exempt interest", "municipal bonds", "muni bonds", "government bonds", "tax-free", "AMT"},
		},
		{
			ID: "irc_121", Title: "Home Sale Exclusion",
			Content: "IRC §121 allows taxpayers to exclude gain from the sale of a principal residence: up to $250,000 for single filers and $500,000 for married filing jointly. To qualify, the taxpayer must have owned and used the home as their principal residence for at least 2 of the 5 years before the sale. The exclusion can generally be used only once every 2 years. Partial exclusions may be available for sales due to health, employment changes, or unforeseen circumstances. Any gain exceeding the exclusion amount is taxed as a capital gain.",
			Source: "IRC §121", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "121", Tags: []string{"home sale", "exclusion", "principal residence", "250000", "500000", "capital gain", "real estate"},
		},
		{
			ID: "irc_125", Title: "Cafeteria Plans",
			Content: "IRC §125 allows employers to offer cafeteria plans that let employees choose between taxable cash compensation and qualified pre-tax benefits. Eligible benefits include health insurance premiums, flexible spending accounts (FSAs), dependent care assistance, and adoption assistance. FSA contribution limits for 2025 are $3,300 for health care FSAs and $5,000 for dependent care FSAs. Amounts elected under a cafeteria plan are excluded from gross income and not subject to FICA or federal income tax withholding.",
			Source: "IRC §125", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "125", Tags: []string{"cafeteria plan", "FSA", "flexible spending", "pre-tax", "benefits", "employer", "health insurance"},
		},
		{
			ID: "irc_127", Title: "Educational Assistance Programs",
			Content: "IRC §127 allows employers to provide up to $5,250 per year in educational assistance benefits tax-free to employees. Qualifying expenses include tuition, fees, books, supplies, and equipment for undergraduate or graduate courses. The education does not need to be job-related. Amounts exceeding $5,250 are taxable as wages unless they qualify for exclusion under another provision such as IRC §132 (working condition fringe benefit). Employer-paid student loan repayments also qualify under §127 through 2025.",
			Source: "IRC §127", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "127", Tags: []string{"educational assistance", "tuition", "employer", "5250", "education", "student loan repayment"},
		},
		{
			ID: "irc_132", Title: "Certain Fringe Benefits",
			Content: "IRC §132 excludes certain employer-provided fringe benefits from gross income. Key categories include: no-additional-cost services, qualified employee discounts, working condition fringe benefits, de minimis fringe benefits, and qualified transportation fringe benefits. For 2025, the qualified transportation exclusion is $325/month for transit passes and parking. Qualified bicycle commuting reimbursements are suspended through 2025 under TCJA. Employer-provided meals and on-premises athletic facilities may also qualify for exclusion.",
			Source: "IRC §132", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "132", Tags: []string{"fringe benefits", "transportation", "parking", "transit", "de minimis", "employee discount", "employer"},
		},
		{
			ID: "irc_213", Title: "Medical and Dental Expense Deduction",
			Content: "IRC §213 allows an itemized deduction for unreimbursed medical and dental expenses that exceed 7.5% of adjusted gross income (AGI). Qualifying expenses include payments for diagnosis, treatment, prevention of disease, health insurance premiums (not pre-tax), prescription drugs, dental care, vision care, and long-term care costs. Cosmetic surgery is generally not deductible. Medical expenses must be paid during the tax year for the taxpayer, spouse, or dependents. The deduction is claimed on Schedule A, line 1.",
			Source: "IRC §213", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "213", Tags: []string{"medical expenses", "dental", "deduction", "7.5 percent", "AGI floor", "itemized", "health insurance", "Schedule A"},
		},
		{
			ID: "irc_219", Title: "Traditional IRA Deduction",
			Content: "IRC §219 governs the deduction for contributions to traditional Individual Retirement Accounts. For 2025, the maximum deductible contribution is $7,000 ($8,000 if age 50+). The deduction may be limited if the taxpayer or spouse is covered by an employer retirement plan: for active participants, the deduction phases out at $79,000-$89,000 AGI (single) and $126,000-$146,000 (MFJ). If only the spouse is covered, the phase-out is $236,000-$246,000 (MFJ). The IRA deduction is an above-the-line deduction on Form 1040, Schedule 1.",
			Source: "IRC §219", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "219", Tags: []string{"IRA", "traditional IRA", "deduction", "retirement", "contribution", "phase-out", "above the line"},
		},
		{
			ID: "irc_221", Title: "Student Loan Interest Deduction",
			Content: "IRC §221 allows an above-the-line deduction of up to $2,500 for interest paid on qualified student loans. The deduction phases out for 2025 at MAGI of $80,000-$95,000 (single) and $165,000-$195,000 (MFJ). Married filing separately taxpayers cannot claim this deduction. The loan must have been taken out solely to pay qualified higher education expenses for the taxpayer, spouse, or dependent. This deduction is claimed on Form 1040, Schedule 1, and reduces AGI.",
			Source: "IRC §221", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "221", Tags: []string{"student loan", "interest", "deduction", "2500", "education", "above the line", "phase-out"},
		},
		{
			ID: "irc_453", Title: "Installment Sales",
			Content: "IRC §453 allows taxpayers who sell property and receive payments over multiple years to report gain proportionally as payments are received (installment method). The gain recognized each year equals the payment received multiplied by the gross profit ratio (total gain divided by total contract price). Installment sales are reported on Form 6252. Depreciation recapture under §1245 and §1250 must be recognized in the year of sale regardless of when payments are received. The installment method is not available for dealer sales of personal property or sales of publicly traded stock.",
			Source: "IRC §453", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "453", Tags: []string{"installment sale", "deferred gain", "Form 6252", "payment", "gross profit ratio", "property sale"},
		},
		{
			ID: "irc_469", Title: "Passive Activity Loss Rules",
			Content: "IRC §469 limits the deduction of losses from passive activities to income from passive activities. Passive activities include trade or business activities in which the taxpayer does not materially participate, and rental activities (regardless of participation, with limited exceptions). Disallowed losses are suspended and carried forward to offset future passive income or are fully deductible when the activity is disposed of. A special $25,000 rental loss allowance exists for active participants with AGI under $100,000, phasing out completely at $150,000 AGI.",
			Source: "IRC §469", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "469", Tags: []string{"passive activity", "rental", "loss limitation", "material participation", "suspended loss", "25000 allowance"},
		},
		{
			ID: "irc_529", Title: "Qualified Tuition Programs (529 Plans)",
			Content: "IRC §529 provides tax-favored treatment for qualified tuition programs (529 plans). Contributions are not federally deductible, but earnings grow tax-free and distributions for qualified education expenses are excluded from income. Qualified expenses include tuition, fees, books, room and board, computers, and up to $10,000/year for K-12 tuition. Up to $35,000 in unused 529 funds can be rolled over to a beneficiary's Roth IRA (subject to annual Roth limits and a 15-year account age requirement, effective 2024). Non-qualified withdrawals are subject to income tax plus a 10% penalty on earnings.",
			Source: "IRC §529", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "529", Tags: []string{"529 plan", "education", "tuition", "college", "savings", "tax-free", "Roth IRA rollover"},
		},
		{
			ID: "irc_1001", Title: "Gain or Loss from Property Dispositions",
			Content: "IRC §1001 determines gain or loss from the sale or other disposition of property. Gain equals the amount realized (cash plus fair market value of property received) minus the adjusted basis of the property. Adjusted basis is generally original cost plus improvements minus depreciation and other reductions. Losses on personal-use property are not deductible. The character of the gain or loss (ordinary vs. capital, short-term vs. long-term) is determined by the type of property and holding period under other code sections.",
			Source: "IRC §1001", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "1001", Tags: []string{"gain", "loss", "basis", "amount realized", "adjusted basis", "property", "disposition", "sale"},
		},
		{
			ID: "irc_1014", Title: "Stepped-Up Basis at Death",
			Content: "IRC §1014 provides that property acquired from a decedent receives a basis equal to its fair market value at the date of death (stepped-up basis). This eliminates any unrealized capital gain that accrued during the decedent's lifetime. For example, if a decedent purchased stock for $10,000 and it was worth $100,000 at death, the heir's basis is $100,000. If the heir sells immediately, no gain is recognized. Community property in community property states (including California) receives a full step-up for both halves upon one spouse's death.",
			Source: "IRC §1014", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "1014", Tags: []string{"stepped-up basis", "death", "inheritance", "fair market value", "estate", "community property", "heir"},
		},
		{
			ID: "irc_1031", Title: "Like-Kind Exchanges",
			Content: "IRC §1031 allows taxpayers to defer gain on the exchange of like-kind property. Since the Tax Cuts and Jobs Act (TCJA), like-kind exchanges are limited to real property (real estate) only; exchanges of personal property, vehicles, equipment, and artwork no longer qualify. The replacement property must be identified within 45 days and received within 180 days of the transfer. Boot (non-like-kind property or cash received) is taxable to the extent of gain realized. Qualified intermediaries are typically used to facilitate exchanges.",
			Source: "IRC §1031", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "1031", Tags: []string{"like-kind exchange", "1031 exchange", "real property", "deferred gain", "real estate", "boot", "TCJA"},
		},
		{
			ID: "irc_1202", Title: "Qualified Small Business Stock (QSBS) Exclusion",
			Content: "IRC §1202 allows taxpayers to exclude up to 100% of gain from the sale of qualified small business stock (QSBS) held for more than 5 years. The maximum exclusion is the greater of $10 million or 10 times the adjusted basis in the stock. The stock must be in a domestic C corporation with gross assets of $50 million or less at the time of issuance and must be acquired at original issuance. The corporation must use at least 80% of its assets in active qualified trades or businesses. Certain industries (finance, law, engineering, hospitality) are excluded.",
			Source: "IRC §1202", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "1202", Tags: []string{"QSBS", "small business stock", "exclusion", "capital gain", "C corporation", "startup", "10 million"},
		},
		{
			ID: "irc_1211", Title: "Capital Loss Limitation",
			Content: "IRC §1211 limits the deduction of capital losses for individual taxpayers. Net capital losses (capital losses exceeding capital gains) can offset up to $3,000 of ordinary income per year ($1,500 for married filing separately). Capital losses first offset capital gains of the same character (short-term offsets short-term, long-term offsets long-term), then net losses offset gains of the other character. Unused capital losses are carried forward indefinitely to future tax years. Capital losses cannot be carried back for individual taxpayers.",
			Source: "IRC §1211", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "1211", Tags: []string{"capital loss", "limitation", "3000", "carryforward", "offset", "ordinary income", "net capital loss"},
		},
		{
			ID: "irc_6662", Title: "Accuracy-Related Penalties",
			Content: "IRC §6662 imposes a 20% penalty on the portion of an underpayment attributable to negligence, disregard of rules, substantial understatement of income tax, or substantial or gross valuation misstatements. A substantial understatement exists when the understatement exceeds the greater of 10% of the correct tax or $5,000. The penalty can be avoided by showing reasonable cause and good faith, or by adequate disclosure of positions with a reasonable basis. Taxpayers should keep thorough records and rely on professional advice to mitigate penalty risk.",
			Source: "IRC §6662", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "6662", Tags: []string{"penalty", "accuracy-related", "understatement", "negligence", "20 percent", "reasonable cause"},
		},
		{
			ID: "irc_7702b", Title: "Long-Term Care Insurance",
			Content: "IRC §7702B defines qualified long-term care insurance contracts and their tax treatment. Premiums paid for qualified long-term care insurance are deductible as medical expenses on Schedule A, subject to age-based limits. For 2025, the maximum deductible premium ranges from $480 (age 40 or under) to $6,010 (age 70+). Benefits received under a qualified policy are generally excluded from income up to a per diem limit of $420/day for 2025. Employer-provided coverage under a cafeteria plan does not qualify for pre-tax treatment.",
			Source: "IRC §7702B", Jurisdiction: JurisdictionFederal, DocType: DocTypeIRCSection,
			Section: "7702B", Tags: []string{"long-term care", "insurance", "medical expense", "deduction", "premium", "age-based limit", "per diem"},
		},
	}
}
