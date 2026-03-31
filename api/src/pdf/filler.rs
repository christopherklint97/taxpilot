use lopdf::Document;
use std::collections::HashMap;

use super::mappings::{FieldFormat, get_mappings};

/// Fill a PDF template with field values and return the filled PDF bytes.
pub fn fill_pdf(
    template_bytes: &[u8],
    form_id: &str,
    values: &HashMap<String, f64>,
    str_values: &HashMap<String, String>,
) -> Result<Vec<u8>, String> {
    let mut doc = Document::load_mem(template_bytes)
        .map_err(|e| format!("Failed to load PDF template: {}", e))?;

    let mappings = get_mappings(form_id);

    for mapping in &mappings {
        let formatted_value = format_field_value(mapping, values, str_values);
        if let Some(value) = formatted_value {
            set_field_value(&mut doc, mapping.pdf_field, &value);
        }
    }

    let mut buf = Vec::new();
    doc.save_to(&mut buf)
        .map_err(|e| format!("Failed to save filled PDF: {}", e))?;
    Ok(buf)
}

fn format_field_value(
    mapping: &super::mappings::FieldMapping,
    values: &HashMap<String, f64>,
    str_values: &HashMap<String, String>,
) -> Option<String> {
    match mapping.format {
        FieldFormat::Currency => values.get(mapping.field_key).map(|v| format!("{:.0}", v)),
        FieldFormat::String => str_values.get(mapping.field_key).cloned(),
        FieldFormat::Ssn => str_values.get(mapping.field_key).map(|s| {
            let digits: String = s.chars().filter(|c| c.is_ascii_digit()).collect();
            if digits.len() == 9 {
                format!("{}-{}-{}", &digits[..3], &digits[3..5], &digits[5..])
            } else {
                s.clone()
            }
        }),
        FieldFormat::Ein => str_values.get(mapping.field_key).cloned(),
        FieldFormat::Integer => values.get(mapping.field_key).map(|v| format!("{:.0}", v)),
        FieldFormat::Checkbox => str_values.get(mapping.field_key).cloned(),
    }
}

fn set_field_value(doc: &mut Document, field_name: &str, value: &str) {
    // Walk the AcroForm fields and find the one matching field_name.
    // lopdf doesn't have a direct "set form field" API, so we need to
    // iterate through the AcroForm/Fields array and match on the T (title) key.
    //
    // This is a simplified implementation -- full AcroForm field setting
    // will be refined when we have real PDF templates to test against.

    let catalog_id = match doc.catalog() {
        Ok(c) => c.clone(),
        Err(_) => return,
    };

    let acroform_ref = match catalog_id.get(b"AcroForm") {
        Ok(r) => r.clone(),
        Err(_) => return,
    };

    let acroform = match acroform_ref.as_dict() {
        Ok(d) => d.clone(),
        Err(_) => return,
    };

    let field_array = match acroform.get(b"Fields") {
        Ok(f) => f.clone(),
        Err(_) => return,
    };

    let arr = match field_array.as_array() {
        Ok(a) => a.clone(),
        Err(_) => return,
    };

    for field_ref in &arr {
        if let Ok(field_oid) = field_ref.as_reference() {
            if let Ok(field_obj) = doc.get_object(field_oid) {
                if let Ok(dict) = field_obj.as_dict() {
                    if let Ok(t) = dict.get(b"T") {
                        if let Ok(name_bytes) = t.as_str() {
                            let name = String::from_utf8_lossy(name_bytes);
                            if name == field_name {
                                // Set the V (value) key on this field
                                let mut new_dict = dict.clone();
                                new_dict.set(
                                    b"V",
                                    lopdf::Object::String(
                                        value.as_bytes().to_vec(),
                                        lopdf::StringFormat::Literal,
                                    ),
                                );
                                let _ = doc.set_object(
                                    field_oid,
                                    lopdf::Object::Dictionary(new_dict),
                                );
                                return;
                            }
                        }
                    }
                }
            }
        }
    }
}
