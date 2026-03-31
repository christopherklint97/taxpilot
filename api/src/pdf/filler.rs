use lopdf::{Document, Object, ObjectId};
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
    if mappings.is_empty() {
        // No mappings: return template as-is so it at least renders
        return Ok(template_bytes.to_vec());
    }

    // Build a lookup from PDF field name -> value to set
    let mut fill_map: HashMap<String, String> = HashMap::new();
    for mapping in &mappings {
        if let Some(value) = format_field_value(mapping, values, str_values) {
            fill_map.insert(mapping.pdf_field.to_string(), value);
        }
    }

    if fill_map.is_empty() {
        return Ok(template_bytes.to_vec());
    }

    // Collect all field object IDs by walking the AcroForm tree recursively
    let field_ids = collect_field_ids(&doc);

    for (oid, full_path) in &field_ids {
        if let Some(value) = fill_map.get(full_path.as_str()) {
            set_field(*oid, value, &mut doc);
        }
    }

    // Remove NeedAppearances if set, to force viewers to regenerate
    set_need_appearances(&mut doc);

    let mut buf = Vec::new();
    doc.save_to(&mut buf)
        .map_err(|e| format!("Failed to save filled PDF: {}", e))?;
    Ok(buf)
}

/// Recursively collect all field ObjectIds and their full hierarchical T paths.
fn collect_field_ids(doc: &Document) -> Vec<(ObjectId, String)> {
    let mut result = Vec::new();

    let catalog = match doc.catalog() {
        Ok(c) => c.clone(),
        Err(_) => return result,
    };

    let acroform = match catalog.get(b"AcroForm") {
        Ok(obj) => match obj {
            Object::Reference(r) => doc.get_object(*r).ok().cloned(),
            Object::Dictionary(_) => Some(obj.clone()),
            _ => None,
        },
        Err(_) => return result,
    };

    let acroform_dict = match acroform.as_ref().and_then(|o| o.as_dict().ok()) {
        Some(d) => d.clone(),
        None => return result,
    };

    let fields = match acroform_dict.get(b"Fields") {
        Ok(f) => f.clone(),
        Err(_) => return result,
    };

    let arr = match resolve_array(&fields, doc) {
        Some(a) => a,
        None => return result,
    };

    for field_ref in &arr {
        if let Ok(oid) = field_ref.as_reference() {
            walk_field(doc, oid, "", &mut result);
        }
    }

    result
}

/// Recursively walk a field node, building the full T-name path.
fn walk_field(doc: &Document, oid: ObjectId, parent_path: &str, result: &mut Vec<(ObjectId, String)>) {
    let obj = match doc.get_object(oid) {
        Ok(o) => o.clone(),
        Err(_) => return,
    };

    let dict = match obj.as_dict() {
        Ok(d) => d,
        Err(_) => return,
    };

    // Build this node's path from parent + T key
    let t_name = dict
        .get(b"T")
        .ok()
        .and_then(|t| t.as_str().ok())
        .map(|b| String::from_utf8_lossy(b).to_string());

    let current_path = match &t_name {
        Some(name) if !parent_path.is_empty() => format!("{}.{}", parent_path, name),
        Some(name) => name.clone(),
        None => parent_path.to_string(),
    };

    // Check for Kids array (container node)
    if let Ok(kids) = dict.get(b"Kids") {
        if let Some(kids_arr) = resolve_array(kids, doc) {
            for kid in &kids_arr {
                if let Ok(kid_oid) = kid.as_reference() {
                    walk_field(doc, kid_oid, &current_path, result);
                }
            }
            return; // Container nodes are not leaf fields
        }
    }

    // Leaf field — has FT (field type) or no Kids
    if !current_path.is_empty() {
        result.push((oid, current_path));
    }
}

fn resolve_array<'a>(obj: &Object, doc: &'a Document) -> Option<Vec<Object>> {
    match obj {
        Object::Array(arr) => Some(arr.clone()),
        Object::Reference(r) => {
            doc.get_object(*r).ok().and_then(|o| resolve_array(o, doc))
        }
        _ => None,
    }
}

/// Set a field's value by modifying its dictionary in place.
fn set_field(oid: ObjectId, value: &str, doc: &mut Document) {
    if let Ok(Object::Dictionary(dict)) = doc.get_object(oid) {
        let mut new_dict = dict.clone();

        // Determine field type
        let ft = new_dict
            .get(b"FT")
            .ok()
            .and_then(|o| o.as_name().ok())
            .map(|n| String::from_utf8_lossy(n).to_string())
            .unwrap_or_default();

        if ft == "Btn" {
            // Checkbox/radio: set V to the value as a Name
            new_dict.set(b"V", Object::Name(value.as_bytes().to_vec()));
            // Also set AS (appearance state) for visual rendering
            new_dict.set(b"AS", Object::Name(value.as_bytes().to_vec()));
        } else {
            // Text field: set V as a string
            new_dict.set(
                b"V",
                Object::String(value.as_bytes().to_vec(), lopdf::StringFormat::Literal),
            );
        }

        // Remove AP (appearance) to force regeneration by the viewer
        new_dict.remove(b"AP");

        let _ = doc.set_object(oid, Object::Dictionary(new_dict));
    }
}

/// Set NeedAppearances=true in the AcroForm dictionary so viewers regenerate field appearances.
fn set_need_appearances(doc: &mut Document) {
    if let Ok(catalog) = doc.catalog() {
        if let Ok(acroform_ref) = catalog.get(b"AcroForm") {
            if let Ok(oid) = acroform_ref.as_reference() {
                if let Ok(Object::Dictionary(d)) = doc.get_object(oid) {
                    let mut new_d = d.clone();
                    new_d.set(b"NeedAppearances", Object::Boolean(true));
                    let _ = doc.set_object(oid, Object::Dictionary(new_d));
                }
            }
        }
    }
}

fn format_field_value(
    mapping: &super::mappings::FieldMapping,
    values: &HashMap<String, f64>,
    str_values: &HashMap<String, String>,
) -> Option<String> {
    match mapping.format {
        FieldFormat::Currency => {
            let v = values.get(mapping.field_key)?;
            if *v == 0.0 { return None; }
            Some(format!("{:.0}", v))
        }
        FieldFormat::String => str_values.get(mapping.field_key).cloned().filter(|s| !s.is_empty()),
        FieldFormat::Ssn => str_values.get(mapping.field_key).map(|s| {
            let digits: String = s.chars().filter(|c| c.is_ascii_digit()).collect();
            if digits.len() == 9 {
                format!("{}-{}-{}", &digits[..3], &digits[3..5], &digits[5..])
            } else {
                s.clone()
            }
        }).filter(|s| !s.is_empty()),
        FieldFormat::Ein => str_values.get(mapping.field_key).cloned().filter(|s| !s.is_empty()),
        FieldFormat::Integer => {
            let v = values.get(mapping.field_key)?;
            if *v == 0.0 { return None; }
            Some(format!("{:.0}", v))
        }
        FieldFormat::Checkbox => {
            // Return the checkbox "on" value (usually "Yes" or "1")
            str_values.get(mapping.field_key).cloned().filter(|s| !s.is_empty())
        }
        FieldFormat::FilingStatus(idx) => {
            // Filing status is a radio group — only fill if the user's status matches this index
            let status = str_values.get(mapping.field_key)?;
            let selected = match status.as_str() {
                "single" => 0,
                "mfj" => 1,
                "mfs" => 2,
                "hoh" => 3,
                "qss" => 4,
                _ => return None,
            };
            if selected == idx { Some("1".to_string()) } else { None }
        }
    }
}
