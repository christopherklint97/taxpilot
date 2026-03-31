use std::collections::HashMap;
use std::sync::LazyLock;

// ---------------------------------------------------------------------------
// Filing Status
// ---------------------------------------------------------------------------

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub enum FilingStatus {
    Single,
    MarriedFilingJoint,
    MarriedFilingSep,
    HeadOfHousehold,
    QualifyingSurvivor,
}

impl FilingStatus {
    pub fn from_str_code(s: &str) -> Option<FilingStatus> {
        match s {
            "single" => Some(FilingStatus::Single),
            "mfj" => Some(FilingStatus::MarriedFilingJoint),
            "mfs" => Some(FilingStatus::MarriedFilingSep),
            "hoh" => Some(FilingStatus::HeadOfHousehold),
            "qss" => Some(FilingStatus::QualifyingSurvivor),
            _ => None,
        }
    }

    pub fn code(&self) -> &'static str {
        match self {
            FilingStatus::Single => "single",
            FilingStatus::MarriedFilingJoint => "mfj",
            FilingStatus::MarriedFilingSep => "mfs",
            FilingStatus::HeadOfHousehold => "hoh",
            FilingStatus::QualifyingSurvivor => "qss",
        }
    }

    pub fn all() -> &'static [FilingStatus] {
        &[
            FilingStatus::Single,
            FilingStatus::MarriedFilingJoint,
            FilingStatus::MarriedFilingSep,
            FilingStatus::HeadOfHousehold,
            FilingStatus::QualifyingSurvivor,
        ]
    }
}

impl std::fmt::Display for FilingStatus {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.code())
    }
}

// ---------------------------------------------------------------------------
// Jurisdiction
// ---------------------------------------------------------------------------

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub enum JurisdictionType {
    Federal,
    StateCA,
}

// ---------------------------------------------------------------------------
// Bracket
// ---------------------------------------------------------------------------

#[derive(Debug, Clone)]
pub struct Bracket {
    pub min: f64,
    pub max: f64, // f64::MAX for the top bracket
    pub rate: f64, // e.g., 0.10 for 10%
}

pub type BracketTable = Vec<Bracket>;

/// Applies progressive bracket rates to taxable income.
pub fn compute_bracket_tax(income: f64, brackets: &[Bracket]) -> f64 {
    if income <= 0.0 {
        return 0.0;
    }
    let mut tax = 0.0;
    for b in brackets {
        if income <= b.min {
            break;
        }
        let top = income.min(b.max);
        let taxable = top - b.min;
        tax += taxable * b.rate;
    }
    tax
}

// ---------------------------------------------------------------------------
// Bracket tables (all years, all jurisdictions)
// ---------------------------------------------------------------------------

macro_rules! brackets {
    ($( { $min:expr, $max:expr, $rate:expr } ),* $(,)?) => {
        vec![ $( Bracket { min: $min, max: $max, rate: $rate } ),* ]
    };
}

// Federal 2024
static FEDERAL_BRACKETS_2024: LazyLock<HashMap<FilingStatus, BracketTable>> = LazyLock::new(|| {
    let mut m = HashMap::new();
    m.insert(FilingStatus::Single, brackets![
        { 0.0, 11600.0, 0.10 },
        { 11600.0, 47150.0, 0.12 },
        { 47150.0, 100525.0, 0.22 },
        { 100525.0, 191950.0, 0.24 },
        { 191950.0, 243725.0, 0.32 },
        { 243725.0, 609350.0, 0.35 },
        { 609350.0, f64::MAX, 0.37 },
    ]);
    m.insert(FilingStatus::MarriedFilingJoint, brackets![
        { 0.0, 23200.0, 0.10 },
        { 23200.0, 94300.0, 0.12 },
        { 94300.0, 201050.0, 0.22 },
        { 201050.0, 383900.0, 0.24 },
        { 383900.0, 487450.0, 0.32 },
        { 487450.0, 731200.0, 0.35 },
        { 731200.0, f64::MAX, 0.37 },
    ]);
    m.insert(FilingStatus::MarriedFilingSep, brackets![
        { 0.0, 11600.0, 0.10 },
        { 11600.0, 47150.0, 0.12 },
        { 47150.0, 100525.0, 0.22 },
        { 100525.0, 191950.0, 0.24 },
        { 191950.0, 243725.0, 0.32 },
        { 243725.0, 365600.0, 0.35 },
        { 365600.0, f64::MAX, 0.37 },
    ]);
    m.insert(FilingStatus::HeadOfHousehold, brackets![
        { 0.0, 16550.0, 0.10 },
        { 16550.0, 63100.0, 0.12 },
        { 63100.0, 100525.0, 0.22 },
        { 100525.0, 191950.0, 0.24 },
        { 191950.0, 243700.0, 0.32 },
        { 243700.0, 609350.0, 0.35 },
        { 609350.0, f64::MAX, 0.37 },
    ]);
    m.insert(FilingStatus::QualifyingSurvivor, brackets![
        { 0.0, 23200.0, 0.10 },
        { 23200.0, 94300.0, 0.12 },
        { 94300.0, 201050.0, 0.22 },
        { 201050.0, 383900.0, 0.24 },
        { 383900.0, 487450.0, 0.32 },
        { 487450.0, 731200.0, 0.35 },
        { 731200.0, f64::MAX, 0.37 },
    ]);
    m
});

// Federal 2025
static FEDERAL_BRACKETS_2025: LazyLock<HashMap<FilingStatus, BracketTable>> = LazyLock::new(|| {
    let mut m = HashMap::new();
    m.insert(FilingStatus::Single, brackets![
        { 0.0, 11925.0, 0.10 },
        { 11925.0, 48475.0, 0.12 },
        { 48475.0, 103350.0, 0.22 },
        { 103350.0, 197300.0, 0.24 },
        { 197300.0, 250525.0, 0.32 },
        { 250525.0, 626350.0, 0.35 },
        { 626350.0, f64::MAX, 0.37 },
    ]);
    m.insert(FilingStatus::MarriedFilingJoint, brackets![
        { 0.0, 23850.0, 0.10 },
        { 23850.0, 96950.0, 0.12 },
        { 96950.0, 206700.0, 0.22 },
        { 206700.0, 394600.0, 0.24 },
        { 394600.0, 501050.0, 0.32 },
        { 501050.0, 751600.0, 0.35 },
        { 751600.0, f64::MAX, 0.37 },
    ]);
    m.insert(FilingStatus::MarriedFilingSep, brackets![
        { 0.0, 11925.0, 0.10 },
        { 11925.0, 48475.0, 0.12 },
        { 48475.0, 103350.0, 0.22 },
        { 103350.0, 197300.0, 0.24 },
        { 197300.0, 250525.0, 0.32 },
        { 250525.0, 375800.0, 0.35 },
        { 375800.0, f64::MAX, 0.37 },
    ]);
    m.insert(FilingStatus::HeadOfHousehold, brackets![
        { 0.0, 17000.0, 0.10 },
        { 17000.0, 64850.0, 0.12 },
        { 64850.0, 103350.0, 0.22 },
        { 103350.0, 197300.0, 0.24 },
        { 197300.0, 250500.0, 0.32 },
        { 250500.0, 626350.0, 0.35 },
        { 626350.0, f64::MAX, 0.37 },
    ]);
    m.insert(FilingStatus::QualifyingSurvivor, brackets![
        { 0.0, 23850.0, 0.10 },
        { 23850.0, 96950.0, 0.12 },
        { 96950.0, 206700.0, 0.22 },
        { 206700.0, 394600.0, 0.24 },
        { 394600.0, 501050.0, 0.32 },
        { 501050.0, 751600.0, 0.35 },
        { 751600.0, f64::MAX, 0.37 },
    ]);
    m
});

// Federal 2026
static FEDERAL_BRACKETS_2026: LazyLock<HashMap<FilingStatus, BracketTable>> = LazyLock::new(|| {
    let mut m = HashMap::new();
    m.insert(FilingStatus::Single, brackets![
        { 0.0, 12250.0, 0.10 },
        { 12250.0, 49825.0, 0.12 },
        { 49825.0, 106250.0, 0.22 },
        { 106250.0, 202850.0, 0.24 },
        { 202850.0, 257550.0, 0.32 },
        { 257550.0, 643900.0, 0.35 },
        { 643900.0, f64::MAX, 0.37 },
    ]);
    m.insert(FilingStatus::MarriedFilingJoint, brackets![
        { 0.0, 24500.0, 0.10 },
        { 24500.0, 99700.0, 0.12 },
        { 99700.0, 212500.0, 0.22 },
        { 212500.0, 405650.0, 0.24 },
        { 405650.0, 515100.0, 0.32 },
        { 515100.0, 772650.0, 0.35 },
        { 772650.0, f64::MAX, 0.37 },
    ]);
    m.insert(FilingStatus::MarriedFilingSep, brackets![
        { 0.0, 12250.0, 0.10 },
        { 12250.0, 49825.0, 0.12 },
        { 49825.0, 106250.0, 0.22 },
        { 106250.0, 202850.0, 0.24 },
        { 202850.0, 257550.0, 0.32 },
        { 257550.0, 386325.0, 0.35 },
        { 386325.0, f64::MAX, 0.37 },
    ]);
    m.insert(FilingStatus::HeadOfHousehold, brackets![
        { 0.0, 17475.0, 0.10 },
        { 17475.0, 66675.0, 0.12 },
        { 66675.0, 106250.0, 0.22 },
        { 106250.0, 202850.0, 0.24 },
        { 202850.0, 257500.0, 0.32 },
        { 257500.0, 643900.0, 0.35 },
        { 643900.0, f64::MAX, 0.37 },
    ]);
    m.insert(FilingStatus::QualifyingSurvivor, brackets![
        { 0.0, 24500.0, 0.10 },
        { 24500.0, 99700.0, 0.12 },
        { 99700.0, 212500.0, 0.22 },
        { 212500.0, 405650.0, 0.24 },
        { 405650.0, 515100.0, 0.32 },
        { 515100.0, 772650.0, 0.35 },
        { 772650.0, f64::MAX, 0.37 },
    ]);
    m
});

// CA 2024
static CA_BRACKETS_2024: LazyLock<HashMap<FilingStatus, BracketTable>> = LazyLock::new(|| {
    let mut m = HashMap::new();
    m.insert(FilingStatus::Single, brackets![
        { 0.0, 10412.0, 0.01 },
        { 10412.0, 24684.0, 0.02 },
        { 24684.0, 38959.0, 0.04 },
        { 38959.0, 54081.0, 0.06 },
        { 54081.0, 68350.0, 0.08 },
        { 68350.0, 349137.0, 0.093 },
        { 349137.0, 418961.0, 0.103 },
        { 418961.0, 698271.0, 0.113 },
        { 698271.0, f64::MAX, 0.123 },
    ]);
    m.insert(FilingStatus::MarriedFilingJoint, brackets![
        { 0.0, 20824.0, 0.01 },
        { 20824.0, 49368.0, 0.02 },
        { 49368.0, 77918.0, 0.04 },
        { 77918.0, 108162.0, 0.06 },
        { 108162.0, 136700.0, 0.08 },
        { 136700.0, 698274.0, 0.093 },
        { 698274.0, 837922.0, 0.103 },
        { 837922.0, 1396542.0, 0.113 },
        { 1396542.0, f64::MAX, 0.123 },
    ]);
    m.insert(FilingStatus::MarriedFilingSep, brackets![
        { 0.0, 10412.0, 0.01 },
        { 10412.0, 24684.0, 0.02 },
        { 24684.0, 38959.0, 0.04 },
        { 38959.0, 54081.0, 0.06 },
        { 54081.0, 68350.0, 0.08 },
        { 68350.0, 349137.0, 0.093 },
        { 349137.0, 418961.0, 0.103 },
        { 418961.0, 698271.0, 0.113 },
        { 698271.0, f64::MAX, 0.123 },
    ]);
    m.insert(FilingStatus::HeadOfHousehold, brackets![
        { 0.0, 20824.0, 0.01 },
        { 20824.0, 49368.0, 0.02 },
        { 49368.0, 77918.0, 0.04 },
        { 77918.0, 108162.0, 0.06 },
        { 108162.0, 136700.0, 0.08 },
        { 136700.0, 698274.0, 0.093 },
        { 698274.0, 837922.0, 0.103 },
        { 837922.0, 1396542.0, 0.113 },
        { 1396542.0, f64::MAX, 0.123 },
    ]);
    m.insert(FilingStatus::QualifyingSurvivor, brackets![
        { 0.0, 20824.0, 0.01 },
        { 20824.0, 49368.0, 0.02 },
        { 49368.0, 77918.0, 0.04 },
        { 77918.0, 108162.0, 0.06 },
        { 108162.0, 136700.0, 0.08 },
        { 136700.0, 698274.0, 0.093 },
        { 698274.0, 837922.0, 0.103 },
        { 837922.0, 1396542.0, 0.113 },
        { 1396542.0, f64::MAX, 0.123 },
    ]);
    m
});

// CA 2025
static CA_BRACKETS_2025: LazyLock<HashMap<FilingStatus, BracketTable>> = LazyLock::new(|| {
    let mut m = HashMap::new();
    m.insert(FilingStatus::Single, brackets![
        { 0.0, 10756.0, 0.01 },
        { 10756.0, 25499.0, 0.02 },
        { 25499.0, 40245.0, 0.04 },
        { 40245.0, 55866.0, 0.06 },
        { 55866.0, 70612.0, 0.08 },
        { 70612.0, 360659.0, 0.093 },
        { 360659.0, 432791.0, 0.103 },
        { 432791.0, 721319.0, 0.113 },
        { 721319.0, f64::MAX, 0.123 },
    ]);
    m.insert(FilingStatus::MarriedFilingJoint, brackets![
        { 0.0, 21512.0, 0.01 },
        { 21512.0, 50998.0, 0.02 },
        { 50998.0, 80490.0, 0.04 },
        { 80490.0, 111732.0, 0.06 },
        { 111732.0, 141224.0, 0.08 },
        { 141224.0, 721318.0, 0.093 },
        { 721318.0, 865582.0, 0.103 },
        { 865582.0, 1442638.0, 0.113 },
        { 1442638.0, f64::MAX, 0.123 },
    ]);
    m.insert(FilingStatus::MarriedFilingSep, brackets![
        { 0.0, 10756.0, 0.01 },
        { 10756.0, 25499.0, 0.02 },
        { 25499.0, 40245.0, 0.04 },
        { 40245.0, 55866.0, 0.06 },
        { 55866.0, 70612.0, 0.08 },
        { 70612.0, 360659.0, 0.093 },
        { 360659.0, 432791.0, 0.103 },
        { 432791.0, 721319.0, 0.113 },
        { 721319.0, f64::MAX, 0.123 },
    ]);
    m.insert(FilingStatus::HeadOfHousehold, brackets![
        { 0.0, 21512.0, 0.01 },
        { 21512.0, 50998.0, 0.02 },
        { 50998.0, 80490.0, 0.04 },
        { 80490.0, 111732.0, 0.06 },
        { 111732.0, 141224.0, 0.08 },
        { 141224.0, 721318.0, 0.093 },
        { 721318.0, 865582.0, 0.103 },
        { 865582.0, 1442638.0, 0.113 },
        { 1442638.0, f64::MAX, 0.123 },
    ]);
    m.insert(FilingStatus::QualifyingSurvivor, brackets![
        { 0.0, 21512.0, 0.01 },
        { 21512.0, 50998.0, 0.02 },
        { 50998.0, 80490.0, 0.04 },
        { 80490.0, 111732.0, 0.06 },
        { 111732.0, 141224.0, 0.08 },
        { 141224.0, 721318.0, 0.093 },
        { 721318.0, 865582.0, 0.103 },
        { 865582.0, 1442638.0, 0.113 },
        { 1442638.0, f64::MAX, 0.123 },
    ]);
    m
});

// CA 2026
static CA_BRACKETS_2026: LazyLock<HashMap<FilingStatus, BracketTable>> = LazyLock::new(|| {
    let mut m = HashMap::new();
    m.insert(FilingStatus::Single, brackets![
        { 0.0, 11057.0, 0.01 },
        { 11057.0, 26213.0, 0.02 },
        { 26213.0, 41372.0, 0.04 },
        { 41372.0, 57430.0, 0.06 },
        { 57430.0, 72589.0, 0.08 },
        { 72589.0, 370758.0, 0.093 },
        { 370758.0, 444909.0, 0.103 },
        { 444909.0, 741516.0, 0.113 },
        { 741516.0, f64::MAX, 0.123 },
    ]);
    m.insert(FilingStatus::MarriedFilingJoint, brackets![
        { 0.0, 22114.0, 0.01 },
        { 22114.0, 52426.0, 0.02 },
        { 52426.0, 82744.0, 0.04 },
        { 82744.0, 114860.0, 0.06 },
        { 114860.0, 145178.0, 0.08 },
        { 145178.0, 741515.0, 0.093 },
        { 741515.0, 889818.0, 0.103 },
        { 889818.0, 1483072.0, 0.113 },
        { 1483072.0, f64::MAX, 0.123 },
    ]);
    m.insert(FilingStatus::MarriedFilingSep, brackets![
        { 0.0, 11057.0, 0.01 },
        { 11057.0, 26213.0, 0.02 },
        { 26213.0, 41372.0, 0.04 },
        { 41372.0, 57430.0, 0.06 },
        { 57430.0, 72589.0, 0.08 },
        { 72589.0, 370758.0, 0.093 },
        { 370758.0, 444909.0, 0.103 },
        { 444909.0, 741516.0, 0.113 },
        { 741516.0, f64::MAX, 0.123 },
    ]);
    m.insert(FilingStatus::HeadOfHousehold, brackets![
        { 0.0, 22114.0, 0.01 },
        { 22114.0, 52426.0, 0.02 },
        { 52426.0, 82744.0, 0.04 },
        { 82744.0, 114860.0, 0.06 },
        { 114860.0, 145178.0, 0.08 },
        { 145178.0, 741515.0, 0.093 },
        { 741515.0, 889818.0, 0.103 },
        { 889818.0, 1483072.0, 0.113 },
        { 1483072.0, f64::MAX, 0.123 },
    ]);
    m.insert(FilingStatus::QualifyingSurvivor, brackets![
        { 0.0, 22114.0, 0.01 },
        { 22114.0, 52426.0, 0.02 },
        { 52426.0, 82744.0, 0.04 },
        { 82744.0, 114860.0, 0.06 },
        { 114860.0, 145178.0, 0.08 },
        { 145178.0, 741515.0, 0.093 },
        { 741515.0, 889818.0, 0.103 },
        { 889818.0, 1483072.0, 0.113 },
        { 1483072.0, f64::MAX, 0.123 },
    ]);
    m
});

// ---------------------------------------------------------------------------
// Standard deduction tables
// ---------------------------------------------------------------------------

fn federal_std_deduction(year: i32) -> HashMap<FilingStatus, f64> {
    match year {
        2024 => HashMap::from([
            (FilingStatus::Single, 14600.0),
            (FilingStatus::MarriedFilingJoint, 29200.0),
            (FilingStatus::MarriedFilingSep, 14600.0),
            (FilingStatus::HeadOfHousehold, 21900.0),
            (FilingStatus::QualifyingSurvivor, 29200.0),
        ]),
        2025 => HashMap::from([
            (FilingStatus::Single, 15000.0),
            (FilingStatus::MarriedFilingJoint, 30000.0),
            (FilingStatus::MarriedFilingSep, 15000.0),
            (FilingStatus::HeadOfHousehold, 22500.0),
            (FilingStatus::QualifyingSurvivor, 30000.0),
        ]),
        2026 => HashMap::from([
            (FilingStatus::Single, 15400.0),
            (FilingStatus::MarriedFilingJoint, 30800.0),
            (FilingStatus::MarriedFilingSep, 15400.0),
            (FilingStatus::HeadOfHousehold, 23100.0),
            (FilingStatus::QualifyingSurvivor, 30800.0),
        ]),
        _ => HashMap::new(),
    }
}

fn ca_std_deduction(year: i32) -> HashMap<FilingStatus, f64> {
    match year {
        2024 => HashMap::from([
            (FilingStatus::Single, 5540.0),
            (FilingStatus::MarriedFilingJoint, 11080.0),
            (FilingStatus::MarriedFilingSep, 5540.0),
            (FilingStatus::HeadOfHousehold, 11080.0),
            (FilingStatus::QualifyingSurvivor, 11080.0),
        ]),
        2025 => HashMap::from([
            (FilingStatus::Single, 5706.0),
            (FilingStatus::MarriedFilingJoint, 11412.0),
            (FilingStatus::MarriedFilingSep, 5706.0),
            (FilingStatus::HeadOfHousehold, 11412.0),
            (FilingStatus::QualifyingSurvivor, 11412.0),
        ]),
        2026 => HashMap::from([
            (FilingStatus::Single, 5866.0),
            (FilingStatus::MarriedFilingJoint, 11732.0),
            (FilingStatus::MarriedFilingSep, 5866.0),
            (FilingStatus::HeadOfHousehold, 11732.0),
            (FilingStatus::QualifyingSurvivor, 11732.0),
        ]),
        _ => HashMap::new(),
    }
}

// ---------------------------------------------------------------------------
// CA exemption credit tables
// ---------------------------------------------------------------------------

fn ca_exemption_credit_base(year: i32) -> HashMap<FilingStatus, f64> {
    match year {
        2024 => HashMap::from([
            (FilingStatus::Single, 140.0),
            (FilingStatus::MarriedFilingJoint, 280.0),
            (FilingStatus::MarriedFilingSep, 140.0),
            (FilingStatus::HeadOfHousehold, 140.0),
            (FilingStatus::QualifyingSurvivor, 140.0),
        ]),
        2025 => HashMap::from([
            (FilingStatus::Single, 144.0),
            (FilingStatus::MarriedFilingJoint, 288.0),
            (FilingStatus::MarriedFilingSep, 144.0),
            (FilingStatus::HeadOfHousehold, 144.0),
            (FilingStatus::QualifyingSurvivor, 144.0),
        ]),
        2026 => HashMap::from([
            (FilingStatus::Single, 148.0),
            (FilingStatus::MarriedFilingJoint, 296.0),
            (FilingStatus::MarriedFilingSep, 148.0),
            (FilingStatus::HeadOfHousehold, 148.0),
            (FilingStatus::QualifyingSurvivor, 148.0),
        ]),
        _ => HashMap::new(),
    }
}

fn ca_exemption_credit_dependent(year: i32) -> f64 {
    match year {
        2024 => 421.0,
        2025 => 433.0,
        2026 => 445.0,
        _ => 0.0,
    }
}

// ---------------------------------------------------------------------------
// CA Mental Health Services Tax
// ---------------------------------------------------------------------------

const CA_MENTAL_HEALTH_THRESHOLD: f64 = 1_000_000.0;
const CA_MENTAL_HEALTH_RATE: f64 = 0.01;

/// Computes the 1% surcharge on taxable income over $1M.
pub fn get_ca_mental_health_tax(taxable_income: f64) -> f64 {
    if taxable_income <= CA_MENTAL_HEALTH_THRESHOLD {
        return 0.0;
    }
    (taxable_income - CA_MENTAL_HEALTH_THRESHOLD) * CA_MENTAL_HEALTH_RATE
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

/// Returns the bracket table for a given year, jurisdiction, and filing status.
pub fn get_brackets(
    year: i32,
    jurisdiction: JurisdictionType,
    status: FilingStatus,
) -> Option<&'static BracketTable> {
    let map: &HashMap<FilingStatus, BracketTable> = match (year, jurisdiction) {
        (2024, JurisdictionType::Federal) => &FEDERAL_BRACKETS_2024,
        (2025, JurisdictionType::Federal) => &FEDERAL_BRACKETS_2025,
        (2026, JurisdictionType::Federal) => &FEDERAL_BRACKETS_2026,
        (2024, JurisdictionType::StateCA) => &CA_BRACKETS_2024,
        (2025, JurisdictionType::StateCA) => &CA_BRACKETS_2025,
        (2026, JurisdictionType::StateCA) => &CA_BRACKETS_2026,
        _ => return None,
    };
    map.get(&status)
}

/// Returns the standard deduction for a given year, jurisdiction, and filing status.
pub fn get_standard_deduction(
    year: i32,
    jurisdiction: JurisdictionType,
    status: FilingStatus,
) -> f64 {
    let map = match jurisdiction {
        JurisdictionType::Federal => federal_std_deduction(year),
        JurisdictionType::StateCA => ca_std_deduction(year),
    };
    map.get(&status).copied().unwrap_or(0.0)
}

/// Computes tax for the given parameters. For CA, includes mental health surcharge.
pub fn compute_tax(
    taxable_income: f64,
    status: FilingStatus,
    year: i32,
    jurisdiction: JurisdictionType,
) -> f64 {
    if taxable_income <= 0.0 {
        return 0.0;
    }
    let brackets = match get_brackets(year, jurisdiction, status) {
        Some(b) => b,
        None => return 0.0,
    };
    let mut tax = compute_bracket_tax(taxable_income, brackets);
    if jurisdiction == JurisdictionType::StateCA {
        tax += get_ca_mental_health_tax(taxable_income);
    }
    tax
}

/// Returns the exemption credit amount for the given status and number of dependents.
pub fn get_ca_exemption_credit(year: i32, status: FilingStatus, num_dependents: i32) -> f64 {
    let base_map = ca_exemption_credit_base(year);
    let credit = base_map.get(&status).copied().unwrap_or(0.0);
    if num_dependents > 0 {
        credit + (num_dependents as f64) * ca_exemption_credit_dependent(year)
    } else {
        credit
    }
}

// ---------------------------------------------------------------------------
// TaxYearConfig
// ---------------------------------------------------------------------------

#[derive(Debug, Clone)]
pub struct TaxYearConfig {
    pub year: i32,

    // HSA contribution limits
    pub hsa_limit_self_only: f64,
    pub hsa_limit_family: f64,

    // Capital loss deduction limits (negative values)
    pub capital_loss_limit: f64,
    pub capital_loss_limit_mfs: f64,

    // Effective tax rate threshold
    pub max_effective_tax_rate: f64,

    // SALT deduction cap
    pub salt_cap: f64,
    pub salt_cap_mfs: f64,

    // FEIE
    pub feie_exclusion_limit: f64,
    pub physical_presence_min_days: i32,

    // FBAR
    pub fbar_threshold: f64,

    // FATCA thresholds
    pub fatca_abroad_single_year_end: f64,
    pub fatca_abroad_single_any_time: f64,
    pub fatca_abroad_mfj_year_end: f64,
    pub fatca_abroad_mfj_any_time: f64,
    pub fatca_us_single_year_end: f64,
    pub fatca_us_single_any_time: f64,
    pub fatca_us_mfj_year_end: f64,
    pub fatca_us_mfj_any_time: f64,

    // CA-specific
    pub ca_max_marginal_rate: f64,
    pub ca_mental_health_rate: f64,
    pub ca_mental_health_threshold: f64,
}

static CONFIGS: LazyLock<HashMap<i32, TaxYearConfig>> = LazyLock::new(|| {
    let mut m = HashMap::new();
    m.insert(
        2024,
        TaxYearConfig {
            year: 2024,
            hsa_limit_self_only: 4150.0,
            hsa_limit_family: 8300.0,
            capital_loss_limit: -3000.0,
            capital_loss_limit_mfs: -1500.0,
            max_effective_tax_rate: 0.37,
            salt_cap: 10000.0,
            salt_cap_mfs: 5000.0,
            feie_exclusion_limit: 126500.0,
            physical_presence_min_days: 330,
            fbar_threshold: 10000.0,
            fatca_abroad_single_year_end: 200000.0,
            fatca_abroad_single_any_time: 300000.0,
            fatca_abroad_mfj_year_end: 400000.0,
            fatca_abroad_mfj_any_time: 600000.0,
            fatca_us_single_year_end: 50000.0,
            fatca_us_single_any_time: 75000.0,
            fatca_us_mfj_year_end: 100000.0,
            fatca_us_mfj_any_time: 150000.0,
            ca_max_marginal_rate: 0.133,
            ca_mental_health_rate: 0.01,
            ca_mental_health_threshold: 1_000_000.0,
        },
    );
    m.insert(
        2025,
        TaxYearConfig {
            year: 2025,
            hsa_limit_self_only: 4300.0,
            hsa_limit_family: 8550.0,
            capital_loss_limit: -3000.0,
            capital_loss_limit_mfs: -1500.0,
            max_effective_tax_rate: 0.37,
            salt_cap: 10000.0,
            salt_cap_mfs: 5000.0,
            feie_exclusion_limit: 130000.0,
            physical_presence_min_days: 330,
            fbar_threshold: 10000.0,
            fatca_abroad_single_year_end: 200000.0,
            fatca_abroad_single_any_time: 300000.0,
            fatca_abroad_mfj_year_end: 400000.0,
            fatca_abroad_mfj_any_time: 600000.0,
            fatca_us_single_year_end: 50000.0,
            fatca_us_single_any_time: 75000.0,
            fatca_us_mfj_year_end: 100000.0,
            fatca_us_mfj_any_time: 150000.0,
            ca_max_marginal_rate: 0.133,
            ca_mental_health_rate: 0.01,
            ca_mental_health_threshold: 1_000_000.0,
        },
    );
    m.insert(
        2026,
        TaxYearConfig {
            year: 2026,
            hsa_limit_self_only: 4400.0,
            hsa_limit_family: 8750.0,
            capital_loss_limit: -3000.0,
            capital_loss_limit_mfs: -1500.0,
            max_effective_tax_rate: 0.37,
            salt_cap: 10000.0,
            salt_cap_mfs: 5000.0,
            feie_exclusion_limit: 133600.0,
            physical_presence_min_days: 330,
            fbar_threshold: 10000.0,
            fatca_abroad_single_year_end: 200000.0,
            fatca_abroad_single_any_time: 300000.0,
            fatca_abroad_mfj_year_end: 400000.0,
            fatca_abroad_mfj_any_time: 600000.0,
            fatca_us_single_year_end: 50000.0,
            fatca_us_single_any_time: 75000.0,
            fatca_us_mfj_year_end: 100000.0,
            fatca_us_mfj_any_time: 150000.0,
            ca_max_marginal_rate: 0.133,
            ca_mental_health_rate: 0.01,
            ca_mental_health_threshold: 1_000_000.0,
        },
    );
    m
});

/// Returns the TaxYearConfig for the given year, or None if not found.
pub fn get_config(year: i32) -> Option<&'static TaxYearConfig> {
    CONFIGS.get(&year)
}

/// Returns the TaxYearConfig for the given year, falling back to closest known year.
pub fn get_config_or_default(year: i32) -> &'static TaxYearConfig {
    if let Some(c) = CONFIGS.get(&year) {
        return c;
    }
    // Fall back to the closest known year
    CONFIGS
        .values()
        .min_by_key(|c| (c.year - year).unsigned_abs())
        .expect("at least one config should exist")
}

// ---------------------------------------------------------------------------
// Rounding
// ---------------------------------------------------------------------------

/// Rounds to the nearest dollar (IRS/FTB standard rounding).
pub fn round_to_nearest(amount: f64) -> f64 {
    amount.round()
}

/// Truncates to the dollar (used for some specific lines).
pub fn round_down(amount: f64) -> f64 {
    amount.floor()
}

// ---------------------------------------------------------------------------
// Expat functions
// ---------------------------------------------------------------------------

/// Returns the foreign earned income exclusion limit for the given tax year.
pub fn feie_limit(tax_year: i32) -> f64 {
    get_config(tax_year)
        .map(|c| c.feie_exclusion_limit)
        .unwrap_or(0.0)
}

/// Returns the base housing amount (16% of FEIE limit).
pub fn housing_base_amount(tax_year: i32) -> f64 {
    feie_limit(tax_year) * 0.16
}

/// Returns the maximum housing expenses (30% of FEIE limit for default locations).
pub fn housing_max_amount(tax_year: i32) -> f64 {
    feie_limit(tax_year) * 0.30
}

/// Prorates the FEIE limit based on qualifying days.
pub fn prorate_exclusion(limit: f64, qualifying_days: i32, total_days: i32) -> f64 {
    if total_days <= 0 || qualifying_days <= 0 {
        return 0.0;
    }
    if qualifying_days >= total_days {
        return limit;
    }
    limit * (qualifying_days as f64) / (total_days as f64)
}

/// Computes tax using the "stacking" method required when FEIE is claimed.
///
/// Formula: tax(taxableIncome + excludedIncome) - tax(excludedIncome)
pub fn compute_tax_with_stacking(
    taxable_income: f64,
    excluded_income: f64,
    status: FilingStatus,
    year: i32,
    jurisdiction: JurisdictionType,
) -> f64 {
    if taxable_income <= 0.0 {
        return 0.0;
    }
    if excluded_income <= 0.0 {
        return compute_tax(taxable_income, status, year, jurisdiction);
    }

    let tax_on_total =
        compute_tax(taxable_income + excluded_income, status, year, jurisdiction);
    let tax_on_excluded = compute_tax(excluded_income, status, year, jurisdiction);

    let stacked = tax_on_total - tax_on_excluded;
    if stacked < 0.0 {
        0.0
    } else {
        stacked
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_compute_bracket_tax_single_2025() {
        // $50,000 single 2025: 10% on first $11,925 + 12% on $11,925-$48,475 + 22% on $48,475-$50,000
        let brackets = get_brackets(2025, JurisdictionType::Federal, FilingStatus::Single).unwrap();
        let tax = compute_bracket_tax(50000.0, brackets);
        let expected = 11925.0 * 0.10 + (48475.0 - 11925.0) * 0.12 + (50000.0 - 48475.0) * 0.22;
        assert!((tax - expected).abs() < 0.01);
    }

    #[test]
    fn test_compute_bracket_tax_zero() {
        let brackets = get_brackets(2025, JurisdictionType::Federal, FilingStatus::Single).unwrap();
        assert_eq!(compute_bracket_tax(0.0, brackets), 0.0);
        assert_eq!(compute_bracket_tax(-1000.0, brackets), 0.0);
    }

    #[test]
    fn test_standard_deduction_2025() {
        assert_eq!(
            get_standard_deduction(2025, JurisdictionType::Federal, FilingStatus::Single),
            15000.0
        );
        assert_eq!(
            get_standard_deduction(2025, JurisdictionType::Federal, FilingStatus::MarriedFilingJoint),
            30000.0
        );
    }

    #[test]
    fn test_ca_mental_health_tax() {
        assert_eq!(get_ca_mental_health_tax(500_000.0), 0.0);
        assert_eq!(get_ca_mental_health_tax(1_000_000.0), 0.0);
        assert!((get_ca_mental_health_tax(1_500_000.0) - 5000.0).abs() < 0.01);
    }

    #[test]
    fn test_ca_exemption_credit_2025() {
        assert_eq!(
            get_ca_exemption_credit(2025, FilingStatus::Single, 0),
            144.0
        );
        assert_eq!(
            get_ca_exemption_credit(2025, FilingStatus::MarriedFilingJoint, 2),
            288.0 + 2.0 * 433.0
        );
    }

    #[test]
    fn test_feie_limit() {
        assert_eq!(feie_limit(2024), 126500.0);
        assert_eq!(feie_limit(2025), 130000.0);
        assert_eq!(feie_limit(2026), 133600.0);
    }

    #[test]
    fn test_prorate_exclusion() {
        assert_eq!(prorate_exclusion(130000.0, 365, 365), 130000.0);
        assert!((prorate_exclusion(130000.0, 330, 365) - 130000.0 * 330.0 / 365.0).abs() < 0.01);
        assert_eq!(prorate_exclusion(130000.0, 0, 365), 0.0);
    }

    #[test]
    fn test_tax_with_stacking() {
        let tax_normal = compute_tax(50000.0, FilingStatus::Single, 2025, JurisdictionType::Federal);
        let tax_stacked = compute_tax_with_stacking(
            50000.0,
            80000.0,
            FilingStatus::Single,
            2025,
            JurisdictionType::Federal,
        );
        // Stacked tax should be higher since remaining income is taxed at higher marginal rates
        assert!(tax_stacked > tax_normal);
    }

    #[test]
    fn test_filing_status_roundtrip() {
        for status in FilingStatus::all() {
            let code = status.code();
            let parsed = FilingStatus::from_str_code(code).unwrap();
            assert_eq!(*status, parsed);
        }
    }

    #[test]
    fn test_get_config() {
        let c = get_config(2025).unwrap();
        assert_eq!(c.year, 2025);
        assert_eq!(c.feie_exclusion_limit, 130000.0);
        assert_eq!(c.hsa_limit_self_only, 4300.0);
    }

    #[test]
    fn test_rounding() {
        assert_eq!(round_to_nearest(1234.5), 1235.0);
        assert_eq!(round_to_nearest(1234.4), 1234.0);
        assert_eq!(round_down(1234.9), 1234.0);
    }
}
