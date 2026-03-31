/// Documents an area where California tax treatment differs from federal treatment.
/// These differences drive Schedule CA adjustments and CA-specific interview questions.
#[derive(Debug, Clone)]
pub struct ConformityDifference {
    pub area: &'static str,
    pub federal: &'static str,
    pub ca: &'static str,
    pub schedule_ca_line: &'static str,
}

/// Key areas where CA differs from federal tax treatment.
/// Used by the interview engine to ask CA-specific questions and to
/// populate Schedule CA adjustments.
pub static CA_CONFORMITY_DIFFERENCES: &[ConformityDifference] = &[
    ConformityDifference {
        area: "Social Security Benefits",
        federal: "Partially taxable (up to 85%)",
        ca: "Not taxable (fully exempt)",
        schedule_ca_line: "6a",
    },
    ConformityDifference {
        area: "SALT Deduction",
        federal: "Deductible up to $10,000 ($5,000 MFS)",
        ca: "No deduction for state/local income taxes paid",
        schedule_ca_line: "5a",
    },
    ConformityDifference {
        area: "Standard Deduction",
        federal: "$15,000 single / $30,000 MFJ (2025)",
        ca: "$5,706 single / $11,412 MFJ (2025)",
        schedule_ca_line: "",
    },
    ConformityDifference {
        area: "QBI Deduction (Section 199A)",
        federal: "Up to 20% deduction for qualified business income",
        ca: "Not allowed -- add back on Schedule CA",
        schedule_ca_line: "13",
    },
    ConformityDifference {
        area: "Municipal Bond Interest",
        federal: "Tax-exempt for all states",
        ca: "Only CA-issued bonds are exempt; out-of-state bonds are taxable",
        schedule_ca_line: "2a",
    },
    ConformityDifference {
        area: "Health Savings Account (HSA)",
        federal: "Contributions deductible; earnings tax-free",
        ca: "No deduction; earnings taxable",
        schedule_ca_line: "13",
    },
    ConformityDifference {
        area: "529 Plan Distributions",
        federal: "Up to $10,000/year for K-12 tuition is tax-free",
        ca: "K-12 distributions are taxable (higher-ed distributions are tax-free)",
        schedule_ca_line: "8",
    },
    ConformityDifference {
        area: "Moving Expenses",
        federal: "Deductible only for active-duty military",
        ca: "Deductible for all taxpayers meeting distance/time tests",
        schedule_ca_line: "14",
    },
    ConformityDifference {
        area: "Gambling Losses",
        federal: "Deductible up to gambling winnings (itemized)",
        ca: "Same as federal",
        schedule_ca_line: "",
    },
    ConformityDifference {
        area: "Foreign Earned Income Exclusion",
        federal: "Up to $130,000 of foreign earned income excluded (2025) via Form 2555",
        ca: "Not allowed -- CA taxes worldwide income regardless of FEIE",
        schedule_ca_line: "8d",
    },
    ConformityDifference {
        area: "Foreign Housing Exclusion/Deduction",
        federal: "Additional exclusion for qualifying housing expenses abroad (Form 2555)",
        ca: "Not allowed -- add back on Schedule CA",
        schedule_ca_line: "8d",
    },
    ConformityDifference {
        area: "Foreign Tax Credit",
        federal: "Credit for taxes paid to foreign governments (Form 1116)",
        ca: "CA allows a credit for taxes paid to other states and foreign countries",
        schedule_ca_line: "",
    },
    ConformityDifference {
        area: "Mental Health Services Tax",
        federal: "N/A",
        ca: "Additional 1% on taxable income over $1,000,000",
        schedule_ca_line: "",
    },
];
