use std::collections::HashMap;
use std::fmt;

pub use crate::domain::wildcard::{build_corresponding_key, match_wildcard};

// ---------------------------------------------------------------------------
// Enums
// ---------------------------------------------------------------------------

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub enum FieldType {
    UserInput,
    Computed,
    Lookup,
    PriorYear,
    FederalRef,
}

impl fmt::Display for FieldType {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            FieldType::UserInput => write!(f, "UserInput"),
            FieldType::Computed => write!(f, "Computed"),
            FieldType::Lookup => write!(f, "Lookup"),
            FieldType::PriorYear => write!(f, "PriorYear"),
            FieldType::FederalRef => write!(f, "FederalRef"),
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub enum FieldValueType {
    Numeric,
    String,
    Integer,
}

impl Default for FieldValueType {
    fn default() -> Self {
        FieldValueType::Numeric
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub enum Jurisdiction {
    Federal,
    StateCA,
}

impl fmt::Display for Jurisdiction {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Jurisdiction::Federal => write!(f, "Federal"),
            Jurisdiction::StateCA => write!(f, "CA"),
        }
    }
}

// ---------------------------------------------------------------------------
// DepValues -- dependency values passed to compute closures
// ---------------------------------------------------------------------------

#[derive(Debug, Clone)]
pub struct DepValues {
    pub num: HashMap<String, f64>,
    pub str_vals: HashMap<String, String>,
    pub tax_year: i32,
}

impl DepValues {
    pub fn new(
        num: HashMap<String, f64>,
        str_vals: HashMap<String, String>,
        tax_year: i32,
    ) -> Self {
        Self {
            num,
            str_vals,
            tax_year,
        }
    }

    /// Get a numeric value, returning 0.0 if not found.
    pub fn get(&self, key: &str) -> f64 {
        self.num.get(key).copied().unwrap_or(0.0)
    }

    /// Get a numeric value, returning an error if not found.
    pub fn get_strict(&self, key: &str) -> Result<f64, String> {
        self.num
            .get(key)
            .copied()
            .ok_or_else(|| format!("dependency key {:?} not found in DepValues", key))
    }

    /// Get a string value, returning empty string if not found.
    pub fn get_string(&self, key: &str) -> String {
        self.str_vals.get(key).cloned().unwrap_or_default()
    }

    /// Return all keys in the numeric map.
    pub fn keys(&self) -> Vec<&String> {
        self.num.keys().collect()
    }

    /// Sum all numeric values whose keys match the given wildcard pattern.
    pub fn sum_all(&self, pattern: &str) -> f64 {
        self.num
            .iter()
            .filter(|(k, _)| match_wildcard(pattern, k))
            .map(|(_, v)| *v)
            .sum()
    }

    /// Sum all numeric values whose keys match value_pattern, but only for
    /// instances where the corresponding filter_pattern key has filter_value
    /// as its string value.
    ///
    /// Example: sum_all_where("1099b:*:proceeds", "1099b:*:term", "short")
    /// sums proceeds for all 1099b instances where term == "short".
    pub fn sum_all_where(
        &self,
        value_pattern: &str,
        filter_pattern: &str,
        filter_value: &str,
    ) -> f64 {
        let mut sum = 0.0;
        for (k, v) in &self.num {
            if !match_wildcard(value_pattern, k) {
                continue;
            }
            if let Some(filter_key) =
                build_corresponding_key(value_pattern, filter_pattern, k)
            {
                if let Some(sv) = self.str_vals.get(&filter_key) {
                    if sv == filter_value {
                        sum += v;
                    }
                }
            }
        }
        sum
    }

    /// Return the tax year.
    pub fn tax_year(&self) -> i32 {
        self.tax_year
    }
}

// ---------------------------------------------------------------------------
// FieldDef
// ---------------------------------------------------------------------------

pub struct FieldDef {
    pub line: String,
    pub field_type: FieldType,
    pub value_type: FieldValueType,
    pub label: String,
    pub prompt: String,
    pub depends_on: Vec<String>,
    pub options: Vec<String>,
    pub compute: Option<Box<dyn Fn(&DepValues) -> f64 + Send + Sync>>,
    pub compute_str: Option<Box<dyn Fn(&DepValues) -> String + Send + Sync>>,
}

impl fmt::Debug for FieldDef {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.debug_struct("FieldDef")
            .field("line", &self.line)
            .field("field_type", &self.field_type)
            .field("value_type", &self.value_type)
            .field("label", &self.label)
            .field("depends_on", &self.depends_on)
            .finish()
    }
}

impl FieldDef {
    pub fn new_user_input(line: &str, label: &str, prompt: &str) -> Self {
        Self {
            line: line.to_string(),
            field_type: FieldType::UserInput,
            value_type: FieldValueType::Numeric,
            label: label.to_string(),
            prompt: prompt.to_string(),
            depends_on: Vec::new(),
            options: Vec::new(),
            compute: None,
            compute_str: None,
        }
    }

    pub fn new_computed(
        line: &str,
        label: &str,
        deps: Vec<String>,
        compute: Box<dyn Fn(&DepValues) -> f64 + Send + Sync>,
    ) -> Self {
        Self {
            line: line.to_string(),
            field_type: FieldType::Computed,
            value_type: FieldValueType::Numeric,
            label: label.to_string(),
            prompt: String::new(),
            depends_on: deps,
            options: Vec::new(),
            compute: Some(compute),
            compute_str: None,
        }
    }
}

// ---------------------------------------------------------------------------
// FormDef
// ---------------------------------------------------------------------------

pub struct FormDef {
    pub id: String,
    pub name: String,
    pub jurisdiction: Jurisdiction,
    pub tax_years: Vec<i32>,
    pub fields: Vec<FieldDef>,
    pub question_group: String,
    pub question_order: i32,
}

impl fmt::Debug for FormDef {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.debug_struct("FormDef")
            .field("id", &self.id)
            .field("name", &self.name)
            .field("jurisdiction", &self.jurisdiction)
            .field("fields_count", &self.fields.len())
            .finish()
    }
}

impl FormDef {
    /// Returns the FieldDef for the given line, or None if not found.
    pub fn field_by_line(&self, line: &str) -> Option<&FieldDef> {
        self.fields.iter().find(|f| f.line == line)
    }
}

// ---------------------------------------------------------------------------
// Key helpers
// ---------------------------------------------------------------------------

/// Build a field key from form ID and line, e.g. field_key("1040", "line1") => "1040:line1"
pub fn field_key(form_id: &str, line: &str) -> String {
    format!("{}:{}", form_id, line)
}

/// Shorthand alias for field_key.
pub fn fk(form_id: &str, line: &str) -> String {
    field_key(form_id, line)
}
