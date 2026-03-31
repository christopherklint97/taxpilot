use std::collections::HashMap;
use std::sync::Arc;

use axum::{
    Json, Router,
    extract::{Multipart, Path, State},
    http::{StatusCode, header},
    response::IntoResponse,
    routing::{get, post},
};
use serde::Serialize;

use crate::AppState;
use crate::domain::field::{FieldType, field_key};
use crate::domain::solver::DependencyGraph;

pub fn router() -> Router<Arc<AppState>> {
    Router::new()
        .route("/pdf/upload", post(upload_pdf))
        .route("/pdf/filled/{form_id}", get(get_filled_pdf))
}

#[derive(Serialize)]
struct UploadResponse {
    id: i64,
    form_id: Option<String>,
    fields_extracted: usize,
}

async fn upload_pdf(
    State(state): State<Arc<AppState>>,
    Path(id): Path<String>,
    mut multipart: Multipart,
) -> Result<Json<UploadResponse>, StatusCode> {
    // Read the uploaded file
    let field = multipart
        .next_field()
        .await
        .map_err(|_| StatusCode::BAD_REQUEST)?
        .ok_or(StatusCode::BAD_REQUEST)?;

    let file_name = field
        .file_name()
        .unwrap_or("upload.pdf")
        .to_string();
    let data = field
        .bytes()
        .await
        .map_err(|_| StatusCode::BAD_REQUEST)?;

    // Try to detect form type
    let form_id = crate::pdf::parser::detect_form_type(&data);

    // Try to extract fields
    let extracted = crate::pdf::parser::extract_form_fields(&data).unwrap_or_default();
    let fields_extracted = extracted.len();

    // Store the PDF document record
    let conn = state.db.conn();

    // Get tax_year for storage path
    let tax_year: i32 = conn
        .query_row(
            "SELECT tax_year FROM tax_returns WHERE id = ?1",
            [&id],
            |row| row.get(0),
        )
        .map_err(|_| StatusCode::NOT_FOUND)?;

    // Store PDF bytes in data directory
    let pdf_dir = format!("data/pdfs/{}", id);
    std::fs::create_dir_all(&pdf_dir).ok();
    let file_path = format!("{}/{}", pdf_dir, file_name);
    std::fs::write(&file_path, &data).map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    // Record in DB
    conn.execute(
        "INSERT INTO pdf_documents (return_id, form_id, tax_year, doc_type, file_path, file_name) \
         VALUES (?1, ?2, ?3, 'uploaded', ?4, ?5)",
        rusqlite::params![id, form_id, tax_year, file_path, file_name],
    )
    .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    let doc_id: i64 = conn.last_insert_rowid();

    // Store extracted field values
    for (pdf_field, value) in &extracted {
        let num_val: Option<f64> = value.parse().ok();
        conn.execute(
            "INSERT OR REPLACE INTO field_values (return_id, field_key, value_num, value_str, source) \
             VALUES (?1, ?2, ?3, ?4, 'pdf_import')",
            rusqlite::params![id, pdf_field, num_val, value],
        )
        .ok();
    }

    Ok(Json(UploadResponse {
        id: doc_id,
        form_id,
        fields_extracted,
    }))
}

async fn get_filled_pdf(
    State(state): State<Arc<AppState>>,
    Path((id, form_id)): Path<(String, String)>,
) -> Result<impl IntoResponse, StatusCode> {
    let conn = state.db.conn();

    // Get tax year
    let tax_year: i32 = conn
        .query_row(
            "SELECT tax_year FROM tax_returns WHERE id = ?1",
            [&id],
            |row| row.get(0),
        )
        .map_err(|_| StatusCode::NOT_FOUND)?;

    // Load all field values (user inputs only for solver)
    let mut inputs = HashMap::new();
    let mut str_inputs = HashMap::new();

    {
        let mut stmt = conn
            .prepare(
                "SELECT field_key, value_num, value_str, source FROM field_values WHERE return_id = ?1",
            )
            .unwrap();
        let rows = stmt
            .query_map([&id], |row| {
                let key: String = row.get(0)?;
                let num: Option<f64> = row.get(1)?;
                let str_val: Option<String> = row.get(2)?;
                let source: String = row.get(3)?;
                Ok((key, num, str_val, source))
            })
            .unwrap();

        for row in rows.flatten() {
            let (key, num, str_val, source) = row;
            if source != "computed" {
                if let Some(n) = num {
                    inputs.insert(key.clone(), n);
                }
                if let Some(s) = str_val {
                    str_inputs.insert(key.clone(), s);
                }
            }
        }
    }

    // Ensure defaults for all UserInput fields
    for form in state.registry.all_forms() {
        for field in &form.fields {
            if field.field_type == FieldType::UserInput {
                let key = field_key(&form.id, &field.line);
                inputs.entry(key).or_insert(0.0);
            }
        }
    }

    let mut graph = DependencyGraph::new(&state.registry);
    graph
        .build()
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;
    let results = graph
        .solve(&inputs, &str_inputs, tax_year)
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    // Try to load template PDF
    let template_path = format!("data/tax_years/{}/federal/{}.pdf", tax_year, form_id);
    let template_bytes = std::fs::read(&template_path);

    match template_bytes {
        Ok(template) => {
            // Fill the template
            let filled =
                crate::pdf::filler::fill_pdf(&template, &form_id, &results, &str_inputs)
                    .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

            Ok((
                [
                    (header::CONTENT_TYPE, "application/pdf".to_string()),
                    (
                        header::CONTENT_DISPOSITION,
                        format!("attachment; filename=\"{}_filled.pdf\"", form_id),
                    ),
                ],
                filled,
            ))
        }
        Err(_) => {
            // No template available - generate a JSON summary instead
            let summary: HashMap<&str, f64> = crate::pdf::mappings::get_mappings(&form_id)
                .iter()
                .filter_map(|m| results.get(m.field_key).map(|v| (m.field_key, *v)))
                .collect();

            let json_bytes =
                serde_json::to_vec(&summary).map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

            Ok((
                [
                    (header::CONTENT_TYPE, "application/json".to_string()),
                    (
                        header::CONTENT_DISPOSITION,
                        "inline".to_string(),
                    ),
                ],
                json_bytes,
            ))
        }
    }
}
