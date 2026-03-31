use crate::domain::field::{DepValues, FieldDef, FieldType, FieldValueType};

/// A computed field that simply references (copies) another field's value.
pub fn ref_field(line: &str, label: &str, dep: &str) -> FieldDef {
    let dep_key = dep.to_string();
    let dep_clone = dep_key.clone();
    FieldDef {
        line: line.to_string(),
        field_type: FieldType::Computed,
        value_type: FieldValueType::Numeric,
        label: label.to_string(),
        prompt: String::new(),
        depends_on: vec![dep_key],
        options: Vec::new(),
        compute: Some(Box::new(move |dv: &DepValues| dv.get(&dep_clone))),
        compute_str: None,
    }
}

/// A computed field that sums all its dependencies.
pub fn sum_field(line: &str, label: &str, deps: Vec<&str>) -> FieldDef {
    let dep_strings: Vec<String> = deps.iter().map(|s| s.to_string()).collect();
    let deps_clone = dep_strings.clone();
    FieldDef {
        line: line.to_string(),
        field_type: FieldType::Computed,
        value_type: FieldValueType::Numeric,
        label: label.to_string(),
        prompt: String::new(),
        depends_on: dep_strings,
        options: Vec::new(),
        compute: Some(Box::new(move |dv: &DepValues| {
            deps_clone.iter().map(|k| dv.get(k)).sum()
        })),
        compute_str: None,
    }
}

/// A computed field that returns a - b.
pub fn diff_field(line: &str, label: &str, a: &str, b: &str) -> FieldDef {
    let a_key = a.to_string();
    let b_key = b.to_string();
    let a_clone = a_key.clone();
    let b_clone = b_key.clone();
    FieldDef {
        line: line.to_string(),
        field_type: FieldType::Computed,
        value_type: FieldValueType::Numeric,
        label: label.to_string(),
        prompt: String::new(),
        depends_on: vec![a_key, b_key],
        options: Vec::new(),
        compute: Some(Box::new(move |dv: &DepValues| {
            dv.get(&a_clone) - dv.get(&b_clone)
        })),
        compute_str: None,
    }
}

/// A computed field that returns max(a - b, 0).
pub fn max_zero_field(line: &str, label: &str, a: &str, b: &str) -> FieldDef {
    let a_key = a.to_string();
    let b_key = b.to_string();
    let a_clone = a_key.clone();
    let b_clone = b_key.clone();
    FieldDef {
        line: line.to_string(),
        field_type: FieldType::Computed,
        value_type: FieldValueType::Numeric,
        label: label.to_string(),
        prompt: String::new(),
        depends_on: vec![a_key, b_key],
        options: Vec::new(),
        compute: Some(Box::new(move |dv: &DepValues| {
            (dv.get(&a_clone) - dv.get(&b_clone)).max(0.0)
        })),
        compute_str: None,
    }
}

/// A computed field that negates the dependency value.
pub fn neg_field(line: &str, label: &str, dep: &str) -> FieldDef {
    let dep_key = dep.to_string();
    let dep_clone = dep_key.clone();
    FieldDef {
        line: line.to_string(),
        field_type: FieldType::Computed,
        value_type: FieldValueType::Numeric,
        label: label.to_string(),
        prompt: String::new(),
        depends_on: vec![dep_key],
        options: Vec::new(),
        compute: Some(Box::new(move |dv: &DepValues| -dv.get(&dep_clone))),
        compute_str: None,
    }
}

/// A computed field that sums all values matching a wildcard pattern.
/// The dependency is the pattern itself (contains `*`).
pub fn wildcard_sum_field(line: &str, label: &str, pattern: &str) -> FieldDef {
    let pat = pattern.to_string();
    let pat_clone = pat.clone();
    FieldDef {
        line: line.to_string(),
        field_type: FieldType::Computed,
        value_type: FieldValueType::Numeric,
        label: label.to_string(),
        prompt: String::new(),
        depends_on: vec![pat],
        options: Vec::new(),
        compute: Some(Box::new(move |dv: &DepValues| dv.sum_all(&pat_clone))),
        compute_str: None,
    }
}

/// A computed field that always returns 0.
pub fn zero_field(line: &str, label: &str) -> FieldDef {
    FieldDef {
        line: line.to_string(),
        field_type: FieldType::Computed,
        value_type: FieldValueType::Numeric,
        label: label.to_string(),
        prompt: String::new(),
        depends_on: Vec::new(),
        options: Vec::new(),
        compute: Some(Box::new(|_: &DepValues| 0.0)),
        compute_str: None,
    }
}

/// A computed field that copies a string value from another field.
pub fn str_ref_field(line: &str, label: &str, dep: &str) -> FieldDef {
    let dep_key = dep.to_string();
    let dep_clone = dep_key.clone();
    FieldDef {
        line: line.to_string(),
        field_type: FieldType::Computed,
        value_type: FieldValueType::String,
        label: label.to_string(),
        prompt: String::new(),
        depends_on: vec![dep_key],
        options: Vec::new(),
        compute: None,
        compute_str: Some(Box::new(move |dv: &DepValues| dv.get_string(&dep_clone))),
    }
}

/// A UserInput field (numeric).
pub fn input_field(line: &str, label: &str, prompt: &str) -> FieldDef {
    FieldDef {
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

/// A UserInput field (string type).
pub fn string_input_field(line: &str, label: &str, prompt: &str) -> FieldDef {
    FieldDef {
        line: line.to_string(),
        field_type: FieldType::UserInput,
        value_type: FieldValueType::String,
        label: label.to_string(),
        prompt: prompt.to_string(),
        depends_on: Vec::new(),
        options: Vec::new(),
        compute: None,
        compute_str: None,
    }
}

/// A UserInput field with predefined option choices (enum).
pub fn enum_field(line: &str, label: &str, prompt: &str, options: Vec<&str>) -> FieldDef {
    FieldDef {
        line: line.to_string(),
        field_type: FieldType::UserInput,
        value_type: FieldValueType::String,
        label: label.to_string(),
        prompt: prompt.to_string(),
        depends_on: Vec::new(),
        options: options.iter().map(|s| s.to_string()).collect(),
        compute: None,
        compute_str: None,
    }
}
