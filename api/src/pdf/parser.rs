use lopdf::Document;
use std::collections::HashMap;

/// Parse a PDF file and extract AcroForm field values.
/// Returns a map of PDF field names to their string values.
pub fn extract_form_fields(pdf_bytes: &[u8]) -> Result<HashMap<String, String>, String> {
    let doc =
        Document::load_mem(pdf_bytes).map_err(|e| format!("Failed to parse PDF: {}", e))?;

    let mut fields = HashMap::new();

    // Try to extract AcroForm fields from the catalog
    if let Ok(catalog_id) = doc.catalog() {
        if let Ok(acroform_ref) = catalog_id.get(b"AcroForm") {
            if let Ok(acroform) = acroform_ref.as_dict() {
                if let Ok(field_array) = acroform.get(b"Fields") {
                    if let Ok(arr) = field_array.as_array() {
                        for field_ref in arr {
                            if let Ok(field_id) = field_ref.as_reference() {
                                if let Ok(field_obj) = doc.get_object(field_id) {
                                    if let Ok(dict) = field_obj.as_dict() {
                                        extract_field_from_dict(dict, &mut fields);
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }

    Ok(fields)
}

/// Extract a single field name and value from an AcroForm field dictionary.
fn extract_field_from_dict(
    dict: &lopdf::Dictionary,
    fields: &mut HashMap<String, String>,
) {
    // Get field name (T key)
    let name = match dict.get(b"T") {
        Ok(t) => match t.as_str() {
            Ok(s) => std::str::from_utf8(s).unwrap_or("").to_string(),
            Err(_) => return,
        },
        Err(_) => return,
    };

    if name.is_empty() {
        return;
    }

    // Get field value (V key)
    if let Ok(v) = dict.get(b"V") {
        let value = match v {
            lopdf::Object::String(bytes, _) => {
                String::from_utf8_lossy(bytes).to_string()
            }
            lopdf::Object::Name(bytes) => {
                String::from_utf8_lossy(bytes).to_string()
            }
            lopdf::Object::Integer(n) => n.to_string(),
            lopdf::Object::Real(n) => n.to_string(),
            _ => String::new(),
        };
        if !value.is_empty() {
            fields.insert(name, value);
        }
    }
}

/// Attempt to detect which form this PDF is (1040, Schedule A, etc.)
pub fn detect_form_type(pdf_bytes: &[u8]) -> Option<String> {
    let doc = Document::load_mem(pdf_bytes).ok()?;

    // Look for form-identifying metadata in the document info dictionary
    if let Ok(info_ref) = doc.trailer.get(b"Info") {
        if let Ok(info_id) = info_ref.as_reference() {
            if let Ok(info_obj) = doc.get_object(info_id) {
                if let Ok(info_dict) = info_obj.as_dict() {
                    // Check Title field for form identification
                    if let Ok(title) = info_dict.get(b"Title") {
                        if let Ok(title_bytes) = title.as_str() {
                            let title_str =
                                String::from_utf8_lossy(title_bytes).to_lowercase();
                            if title_str.contains("1040") {
                                return Some("1040".to_string());
                            }
                            if title_str.contains("schedule a") {
                                return Some("schedule_a".to_string());
                            }
                            if title_str.contains("schedule b") {
                                return Some("schedule_b".to_string());
                            }
                            if title_str.contains("schedule c") {
                                return Some("schedule_c".to_string());
                            }
                            if title_str.contains("schedule d") {
                                return Some("schedule_d".to_string());
                            }
                            if title_str.contains("w-2") || title_str.contains("w2") {
                                return Some("w2".to_string());
                            }
                        }
                    }
                }
            }
        }
    }

    // Could also check AcroForm field names for form-specific patterns
    None
}
