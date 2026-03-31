use lopdf::{Document, Object, ObjectId};
use std::collections::HashMap;

use super::mappings::get_mappings;

/// Parse a PDF file and extract AcroForm field values.
/// Returns a map of TaxPilot field_keys to their string values,
/// using the reverse of our PDF field mappings.
pub fn extract_form_fields(pdf_bytes: &[u8]) -> Result<HashMap<String, String>, String> {
    let doc =
        Document::load_mem(pdf_bytes).map_err(|e| format!("Failed to parse PDF: {}", e))?;

    // Extract all PDF field values by recursively walking the AcroForm tree
    let pdf_fields = extract_all_fields(&doc);

    // Detect what form this is and build reverse mapping
    let form_id = detect_form_type_from_fields(&pdf_fields)
        .or_else(|| detect_form_type_from_metadata(&doc));

    let mut result = HashMap::new();

    if let Some(fid) = &form_id {
        let mappings = get_mappings(fid);
        for mapping in &mappings {
            if let Some(value) = pdf_fields.get(mapping.pdf_field) {
                if !value.is_empty() && value != "Off" && value != "0" {
                    result.insert(mapping.field_key.to_string(), value.clone());
                }
            }
        }
    }

    // If no mappings matched, fall back to raw field extraction
    // Store with the raw PDF field name (less useful but better than nothing)
    if result.is_empty() {
        for (name, value) in &pdf_fields {
            if !value.is_empty() && value != "Off" {
                result.insert(name.clone(), value.clone());
            }
        }
    }

    Ok(result)
}

/// Recursively extract all field name→value pairs from the PDF AcroForm.
fn extract_all_fields(doc: &Document) -> HashMap<String, String> {
    let mut fields = HashMap::new();

    let catalog = match doc.catalog() {
        Ok(c) => c.clone(),
        Err(_) => return fields,
    };

    let acroform = match catalog.get(b"AcroForm") {
        Ok(obj) => match obj {
            Object::Reference(r) => doc.get_object(*r).ok().cloned(),
            Object::Dictionary(_) => Some(obj.clone()),
            _ => None,
        },
        Err(_) => return fields,
    };

    let acroform_dict = match acroform.as_ref().and_then(|o| o.as_dict().ok()) {
        Some(d) => d.clone(),
        None => return fields,
    };

    let field_array = match acroform_dict.get(b"Fields") {
        Ok(f) => f.clone(),
        Err(_) => return fields,
    };

    let arr = match resolve_array(&field_array, doc) {
        Some(a) => a,
        None => return fields,
    };

    for field_ref in &arr {
        if let Ok(oid) = field_ref.as_reference() {
            walk_and_extract(doc, oid, "", &mut fields);
        }
    }

    fields
}

/// Recursively walk field tree, building full paths and extracting values.
fn walk_and_extract(
    doc: &Document,
    oid: ObjectId,
    parent_path: &str,
    fields: &mut HashMap<String, String>,
) {
    let obj = match doc.get_object(oid) {
        Ok(o) => o.clone(),
        Err(_) => return,
    };

    let dict = match obj.as_dict() {
        Ok(d) => d,
        Err(_) => return,
    };

    // Build path from T key
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

    // Check for Kids (container node)
    if let Ok(kids) = dict.get(b"Kids") {
        if let Some(kids_arr) = resolve_array(kids, doc) {
            for kid in &kids_arr {
                if let Ok(kid_oid) = kid.as_reference() {
                    walk_and_extract(doc, kid_oid, &current_path, fields);
                }
            }
            return;
        }
    }

    // Leaf field — extract value
    if current_path.is_empty() {
        return;
    }

    if let Ok(v) = dict.get(b"V") {
        let value = match v {
            Object::String(bytes, _) => String::from_utf8_lossy(bytes).to_string(),
            Object::Name(bytes) => String::from_utf8_lossy(bytes).to_string(),
            Object::Integer(n) => n.to_string(),
            Object::Real(n) => n.to_string(),
            _ => String::new(),
        };
        if !value.is_empty() {
            fields.insert(current_path, value);
        }
    }
}

fn resolve_array(obj: &Object, doc: &Document) -> Option<Vec<Object>> {
    match obj {
        Object::Array(arr) => Some(arr.clone()),
        Object::Reference(r) => doc.get_object(*r).ok().and_then(|o| resolve_array(o, doc)),
        _ => None,
    }
}

/// Detect form type from the extracted field names.
fn detect_form_type_from_fields(fields: &HashMap<String, String>) -> Option<String> {
    // IRS federal forms have fields like "topmostSubform[0].Page1[0].f1_XX[0]"
    // CA FTB forms have fields like "540-XXXX" or "540ca_form - XXXX"
    for key in fields.keys() {
        if key.starts_with("540ca_form") {
            return Some("ca_schedule_ca".to_string());
        }
        if key.starts_with("540-") {
            return Some("ca_540".to_string());
        }
        if key.starts_with("3514_Form") {
            return Some("form_3514".to_string());
        }
        if key.starts_with("3853 Form") {
            return Some("form_3853".to_string());
        }
    }
    // For federal forms, rely on metadata detection
    None
}

/// Detect form type from PDF metadata (Title, Subject).
fn detect_form_type_from_metadata(doc: &Document) -> Option<String> {
    let info_ref = doc.trailer.get(b"Info").ok()?;
    let info_id = info_ref.as_reference().ok()?;
    let info_obj = doc.get_object(info_id).ok()?;
    let info_dict = info_obj.as_dict().ok()?;

    let title = info_dict
        .get(b"Title")
        .ok()
        .and_then(|t| t.as_str().ok())
        .map(|b| String::from_utf8_lossy(b).to_lowercase())?;

    // Match against known form titles
    if title.contains("schedule se") { return Some("schedule_se".to_string()); }
    if title.contains("schedule a") { return Some("schedule_a".to_string()); }
    if title.contains("schedule b") { return Some("schedule_b".to_string()); }
    if title.contains("schedule c") { return Some("schedule_c".to_string()); }
    if title.contains("schedule d") { return Some("schedule_d".to_string()); }
    if title.contains("schedule 1") || title.contains("additional income") {
        return Some("schedule_1".to_string());
    }
    if title.contains("schedule 2") || title.contains("additional taxes") {
        return Some("schedule_2".to_string());
    }
    if title.contains("schedule 3") || title.contains("additional credits") {
        return Some("schedule_3".to_string());
    }
    if title.contains("8995") { return Some("form_8995".to_string()); }
    if title.contains("8889") { return Some("form_8889".to_string()); }
    if title.contains("8949") { return Some("form_8949".to_string()); }
    if title.contains("8938") { return Some("form_8938".to_string()); }
    if title.contains("8833") { return Some("form_8833".to_string()); }
    if title.contains("2555") { return Some("form_2555".to_string()); }
    if title.contains("1116") { return Some("form_1116".to_string()); }
    if title.contains("1040") { return Some("1040".to_string()); }
    if title.contains("w-2") || title.contains("w2") { return Some("w2".to_string()); }

    None
}

/// Public form type detection (used by upload route).
pub fn detect_form_type(pdf_bytes: &[u8]) -> Option<String> {
    let doc = Document::load_mem(pdf_bytes).ok()?;
    let fields = extract_all_fields(&doc);
    detect_form_type_from_fields(&fields)
        .or_else(|| detect_form_type_from_metadata(&doc))
}
