use std::collections::HashMap;

use crate::domain::field::{FieldDef, FieldType, FormDef, Jurisdiction};

/// Registry holds all registered forms and provides lookup.
pub struct Registry {
    forms: HashMap<String, FormDef>,
}

impl Registry {
    /// Creates an empty Registry.
    pub fn new() -> Self {
        Self {
            forms: HashMap::new(),
        }
    }

    /// Adds a form definition to the registry.
    pub fn register(&mut self, form: FormDef) {
        self.forms.insert(form.id.clone(), form);
    }

    /// Returns the form with the given ID, or None if not found.
    pub fn get(&self, form_id: &str) -> Option<&FormDef> {
        self.forms.get(form_id)
    }

    /// Looks up a field by its fully qualified key ("form_id:line").
    /// Returns the parent form and field definition, or an error if not found.
    pub fn get_field(&self, key: &str) -> Result<(&FormDef, &FieldDef), String> {
        let parts: Vec<&str> = key.splitn(2, ':').collect();
        if parts.len() != 2 {
            return Err(format!(
                "invalid field key {:?}: expected form_id:line",
                key
            ));
        }
        let (form_id, line) = (parts[0], parts[1]);

        let form = self
            .forms
            .get(form_id)
            .ok_or_else(|| format!("form {:?} not found", form_id))?;

        let field = form
            .field_by_line(line)
            .ok_or_else(|| format!("field {:?} not found in form {:?}", line, form_id))?;

        Ok((form, field))
    }

    /// Returns all registered form definitions.
    pub fn all_forms(&self) -> Vec<&FormDef> {
        self.forms.values().collect()
    }

    /// Checks all registered forms for common errors.
    /// Returns a list of error messages.
    pub fn validate_field_defs(&self) -> Vec<String> {
        let mut errs = Vec::new();
        for form in self.forms.values() {
            for field in &form.fields {
                if field.field_type == FieldType::UserInput && field.compute.is_some() {
                    errs.push(format!(
                        "form {} field {}: UserInput should not have Compute",
                        form.id, field.line
                    ));
                }
            }
        }
        errs
    }

    /// Checks that all FederalRef fields reference forms with Jurisdiction::Federal.
    pub fn validate_federal_refs(&self) -> Vec<String> {
        let mut errs = Vec::new();
        for form in self.forms.values() {
            for field in &form.fields {
                if field.field_type != FieldType::FederalRef {
                    continue;
                }
                for dep in &field.depends_on {
                    let parts: Vec<&str> = dep.splitn(2, ':').collect();
                    if parts.len() != 2 {
                        errs.push(format!(
                            "form {} field {}: FederalRef dependency {:?} has invalid format",
                            form.id, field.line, dep
                        ));
                        continue;
                    }
                    let dep_form_id = parts[0];
                    if let Some(dep_form) = self.forms.get(dep_form_id) {
                        if dep_form.jurisdiction != Jurisdiction::Federal {
                            errs.push(format!(
                                "form {} field {}: FederalRef dependency {:?} references non-federal form {} (jurisdiction {:?})",
                                form.id, field.line, dep, dep_form_id, dep_form.jurisdiction
                            ));
                        }
                    }
                    // If the referenced form is not registered, skip validation
                }
            }
        }
        errs
    }
}

impl Default for Registry {
    fn default() -> Self {
        Self::new()
    }
}
