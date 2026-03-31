/// Maps a field_key (e.g., "1040:filing_status") to the PDF AcroForm field name.
pub struct FieldMapping {
    pub field_key: &'static str,
    pub pdf_field: &'static str,
    pub format: FieldFormat,
}

pub enum FieldFormat {
    Currency,
    String,
    Integer,
    Ssn,
    Ein,
    Checkbox,
}

/// Returns all PDF field mappings for a given form.
pub fn get_mappings(form_id: &str) -> Vec<FieldMapping> {
    match form_id {
        "1040" => f1040_mappings(),
        _ => vec![],
    }
}

fn f1040_mappings() -> Vec<FieldMapping> {
    // These are the actual IRS PDF AcroForm field IDs
    // They'll be populated as we verify against real PDF templates
    vec![
        FieldMapping {
            field_key: "1040:filing_status",
            pdf_field: "topmostSubform[0].Page1[0].FilingStatus[0].c1_01[0]",
            format: FieldFormat::Checkbox,
        },
        FieldMapping {
            field_key: "1040:first_name",
            pdf_field: "topmostSubform[0].Page1[0].f1_02[0]",
            format: FieldFormat::String,
        },
        FieldMapping {
            field_key: "1040:last_name",
            pdf_field: "topmostSubform[0].Page1[0].f1_03[0]",
            format: FieldFormat::String,
        },
        FieldMapping {
            field_key: "1040:ssn",
            pdf_field: "topmostSubform[0].Page1[0].f1_04[0]",
            format: FieldFormat::Ssn,
        },
        FieldMapping {
            field_key: "1040:1a",
            pdf_field: "topmostSubform[0].Page1[0].f1_07[0]",
            format: FieldFormat::Currency,
        },
        FieldMapping {
            field_key: "1040:11",
            pdf_field: "topmostSubform[0].Page1[0].f1_25[0]",
            format: FieldFormat::Currency,
        },
    ]
}
