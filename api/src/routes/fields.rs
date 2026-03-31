use std::collections::HashMap;
use std::sync::Arc;

use axum::{
    Json, Router,
    extract::{Path, State},
    http::StatusCode,
    routing::put,
};
use serde::{Deserialize, Serialize};

use crate::AppState;
use crate::domain::field::{FieldType, FieldValueType, field_key};
use crate::domain::solver::DependencyGraph;

pub fn router() -> Router<Arc<AppState>> {
    Router::new()
        .route("/fields/{key}", put(update_field))
        .route("/fields", put(batch_update_fields))
}

#[derive(Deserialize)]
pub struct UpdateFieldRequest {
    pub value_num: Option<f64>,
    pub value_str: Option<String>,
}

#[derive(Deserialize)]
pub struct BatchUpdateRequest {
    pub fields: Vec<BatchFieldEntry>,
}

#[derive(Deserialize)]
pub struct BatchFieldEntry {
    pub key: String,
    pub value_num: Option<f64>,
    pub value_str: Option<String>,
}

#[derive(Serialize)]
pub struct ChangedField {
    pub key: String,
    pub value_num: Option<f64>,
    pub value_str: Option<String>,
}

#[derive(Serialize)]
pub struct UpdateFieldResponse {
    pub changed_fields: Vec<ChangedField>,
}

/// Load field values for a return from the DB.
/// Public so other route modules (e.g. rollforward) can reuse it.
/// Returns (user_inputs, str_inputs, all_old_values) where:
/// - user_inputs: only user_input/pdf_import/prior_year source fields (for solver)
/// - str_inputs: string values for all user_input fields
/// - all_old_values: ALL field values including computed (for diffing)
pub fn load_field_values(
    conn: &rusqlite::Connection,
    return_id: &str,
) -> (HashMap<String, f64>, HashMap<String, String>, HashMap<String, f64>) {
    let mut user_inputs = HashMap::new();
    let mut str_inputs = HashMap::new();
    let mut all_old = HashMap::new();

    let mut stmt = conn
        .prepare(
            "SELECT field_key, value_num, value_str, source FROM field_values WHERE return_id = ?1",
        )
        .unwrap();
    let rows = stmt
        .query_map([return_id], |row| {
            let key: String = row.get(0)?;
            let num: Option<f64> = row.get(1)?;
            let str_val: Option<String> = row.get(2)?;
            let source: String = row.get(3)?;
            Ok((key, num, str_val, source))
        })
        .unwrap();

    for row in rows.flatten() {
        let (key, num, str_val, source) = row;
        // Track all old values for diffing
        if let Some(n) = num {
            all_old.insert(key.clone(), n);
        }
        // Only feed user_input/pdf_import/prior_year values to the solver
        if source != "computed" {
            if let Some(n) = num {
                user_inputs.insert(key.clone(), n);
            }
            if let Some(s) = str_val {
                str_inputs.insert(key.clone(), s);
            }
        }
    }

    (user_inputs, str_inputs, all_old)
}

/// Ensure all UserInput fields have default values in the DB for this return.
pub fn ensure_defaults(
    conn: &rusqlite::Connection,
    return_id: &str,
    registry: &crate::domain::registry::Registry,
) {
    for form in registry.all_forms() {
        for field in &form.fields {
            if field.field_type == FieldType::UserInput {
                let key = field_key(&form.id, &field.line);
                let exists: bool = conn
                    .query_row(
                        "SELECT COUNT(*) > 0 FROM field_values WHERE return_id = ?1 AND field_key = ?2",
                        rusqlite::params![return_id, key],
                        |row| row.get(0),
                    )
                    .unwrap_or(false);
                if !exists {
                    let (num, str_val) = match field.value_type {
                        FieldValueType::String => (Some(0.0), Some(String::new())),
                        _ => (Some(0.0), None),
                    };
                    conn.execute(
                        "INSERT OR IGNORE INTO field_values (return_id, field_key, value_num, value_str, source) \
                         VALUES (?1, ?2, ?3, ?4, 'user_input')",
                        rusqlite::params![return_id, key, num, str_val],
                    )
                    .unwrap();
                }
            }
        }
    }
}

/// Run the solver and return changed computed fields.
pub fn solve_and_diff(
    state: &AppState,
    return_id: &str,
    tax_year: i32,
) -> Result<Vec<ChangedField>, String> {
    let conn = state.db.conn();

    // Ensure defaults exist
    ensure_defaults(&conn, return_id, &state.registry);

    // Load current values (user inputs for solver, all values for diffing)
    let (inputs, str_inputs, old_computed) = load_field_values(&conn, return_id);

    // Build and solve
    let mut graph = DependencyGraph::new(&state.registry);
    graph.build()?;
    let results = graph.solve(&inputs, &str_inputs, tax_year)?;

    // Diff: find computed fields that changed
    let mut changed = Vec::new();
    for form in state.registry.all_forms() {
        for field in &form.fields {
            if field.field_type == FieldType::UserInput {
                continue;
            }
            let key = field_key(&form.id, &field.line);
            let new_val = results.get(&key).copied();
            let old_val = old_computed.get(&key).copied();

            let val_changed = match (new_val, old_val) {
                (Some(n), Some(o)) => (n - o).abs() > 1e-10,
                (Some(n), None) => n.abs() > 1e-10,
                (None, Some(o)) => o.abs() > 1e-10,
                (None, None) => false,
            };

            if val_changed || old_val.is_none() {
                let num = new_val.or(Some(0.0));
                // Upsert computed value
                conn.execute(
                    "INSERT INTO field_values (return_id, field_key, value_num, value_str, source) \
                     VALUES (?1, ?2, ?3, NULL, 'computed') \
                     ON CONFLICT(return_id, field_key) DO UPDATE SET value_num = ?3, source = 'computed', \
                     updated_at = datetime('now')",
                    rusqlite::params![return_id, key, num],
                )
                .unwrap();

                changed.push(ChangedField {
                    key,
                    value_num: num,
                    value_str: None,
                });
            }
        }
    }

    // Update return's updated_at
    conn.execute(
        "UPDATE tax_returns SET updated_at = datetime('now') WHERE id = ?1",
        [return_id],
    )
    .unwrap();

    Ok(changed)
}

/// GET the tax_year for a return.
pub fn get_tax_year(conn: &rusqlite::Connection, return_id: &str) -> Result<i32, StatusCode> {
    conn.query_row(
        "SELECT tax_year FROM tax_returns WHERE id = ?1",
        [return_id],
        |row| row.get(0),
    )
    .map_err(|_| StatusCode::NOT_FOUND)
}

async fn update_field(
    State(state): State<Arc<AppState>>,
    Path((id, key)): Path<(String, String)>,
    Json(body): Json<UpdateFieldRequest>,
) -> Result<Json<UpdateFieldResponse>, StatusCode> {
    let tax_year = {
        let conn = state.db.conn();
        get_tax_year(&conn, &id)?
    };

    // Upsert the field value
    {
        let conn = state.db.conn();
        conn.execute(
            "INSERT INTO field_values (return_id, field_key, value_num, value_str, source) \
             VALUES (?1, ?2, ?3, ?4, 'user_input') \
             ON CONFLICT(return_id, field_key) DO UPDATE SET \
             value_num = ?3, value_str = ?4, source = 'user_input', updated_at = datetime('now')",
            rusqlite::params![id, key, body.value_num, body.value_str],
        )
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

        // If updating filing_status, also update the returns table
        if key == "1040:filing_status" {
            if let Some(ref fs) = body.value_str {
                conn.execute(
                    "UPDATE tax_returns SET filing_status = ?1, updated_at = datetime('now') WHERE id = ?2",
                    rusqlite::params![fs, id],
                )
                .unwrap();
            }
        }
    }

    // Re-solve
    let changed = solve_and_diff(&state, &id, tax_year)
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    Ok(Json(UpdateFieldResponse {
        changed_fields: changed,
    }))
}

async fn batch_update_fields(
    State(state): State<Arc<AppState>>,
    Path(id): Path<String>,
    Json(body): Json<BatchUpdateRequest>,
) -> Result<Json<UpdateFieldResponse>, StatusCode> {
    let tax_year = {
        let conn = state.db.conn();
        get_tax_year(&conn, &id)?
    };

    // Upsert all fields
    {
        let conn = state.db.conn();
        for entry in &body.fields {
            conn.execute(
                "INSERT INTO field_values (return_id, field_key, value_num, value_str, source) \
                 VALUES (?1, ?2, ?3, ?4, 'user_input') \
                 ON CONFLICT(return_id, field_key) DO UPDATE SET \
                 value_num = ?3, value_str = ?4, source = 'user_input', updated_at = datetime('now')",
                rusqlite::params![id, entry.key, entry.value_num, entry.value_str],
            )
            .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

            if entry.key == "1040:filing_status" {
                if let Some(ref fs) = entry.value_str {
                    conn.execute(
                        "UPDATE tax_returns SET filing_status = ?1, updated_at = datetime('now') WHERE id = ?2",
                        rusqlite::params![fs, id],
                    )
                    .unwrap();
                }
            }
        }
    }

    // Re-solve
    let changed = solve_and_diff(&state, &id, tax_year)
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    Ok(Json(UpdateFieldResponse {
        changed_fields: changed,
    }))
}
