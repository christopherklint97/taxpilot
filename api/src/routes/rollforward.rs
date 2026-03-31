use std::sync::Arc;

use axum::{
    Json, Router,
    extract::{Path, State},
    http::StatusCode,
    routing::{get, post},
};
use serde::{Deserialize, Serialize};

use crate::AppState;
use crate::routes::fields::solve_and_diff;

pub fn router() -> Router<Arc<AppState>> {
    Router::new()
        .route("/rollforward", post(rollforward))
        .route("/prior-year", get(get_prior_year_values))
}

#[derive(Deserialize)]
struct RollforwardRequest {
    source_return_id: String,
    target_tax_year: Option<i32>,
}

#[derive(Serialize)]
struct RollforwardResponse {
    return_id: String,
    tax_year: i32,
    fields_carried: usize,
    prior_year_values: usize,
}

#[derive(Serialize)]
struct PriorYearValue {
    field_key: String,
    source_year: i32,
    value_num: Option<f64>,
    value_str: Option<String>,
}

async fn rollforward(
    State(state): State<Arc<AppState>>,
    Path(id): Path<String>,
    Json(body): Json<RollforwardRequest>,
) -> Result<(StatusCode, Json<RollforwardResponse>), StatusCode> {
    let conn = state.db.conn();

    // Look up the source return
    let (source_year, source_state_code, source_filing_status): (i32, String, Option<String>) =
        conn.query_row(
            "SELECT tax_year, state_code, filing_status FROM tax_returns WHERE id = ?1",
            [&body.source_return_id],
            |row| Ok((row.get(0)?, row.get(1)?, row.get(2)?)),
        )
        .map_err(|_| StatusCode::NOT_FOUND)?;

    // Verify the path id matches the source return (the route is nested under /returns/{id})
    if id != body.source_return_id {
        return Err(StatusCode::BAD_REQUEST);
    }

    let target_year = body.target_tax_year.unwrap_or(source_year + 1);

    // Create the new return
    let new_id = uuid::Uuid::new_v4().to_string();
    conn.execute(
        "INSERT INTO tax_returns (id, tax_year, state_code, filing_status) VALUES (?1, ?2, ?3, ?4)",
        rusqlite::params![new_id, target_year, source_state_code, source_filing_status],
    )
    .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    // Load ALL field values from the source return
    let mut source_fields: Vec<(String, Option<f64>, Option<String>, String)> = Vec::new();
    {
        let mut stmt = conn
            .prepare(
                "SELECT field_key, value_num, value_str, source \
                 FROM field_values WHERE return_id = ?1",
            )
            .unwrap();
        let rows = stmt
            .query_map([&body.source_return_id], |row| {
                Ok((
                    row.get::<_, String>(0)?,
                    row.get::<_, Option<f64>>(1)?,
                    row.get::<_, Option<String>>(2)?,
                    row.get::<_, String>(3)?,
                ))
            })
            .unwrap();
        for row in rows.flatten() {
            source_fields.push(row);
        }
    }

    // Store ALL source values as prior_year_values (for delta comparison)
    let mut prior_year_count = 0;
    for (key, num, str_val, _source) in &source_fields {
        conn.execute(
            "INSERT OR REPLACE INTO prior_year_values (return_id, source_year, field_key, value_num, value_str) \
             VALUES (?1, ?2, ?3, ?4, ?5)",
            rusqlite::params![new_id, source_year, key, num, str_val],
        )
        .unwrap();
        prior_year_count += 1;
    }

    // Copy user_input fields to the new return's field_values (as source='prior_year')
    let mut fields_carried = 0;
    for (key, num, str_val, source) in &source_fields {
        if source == "user_input" || source == "pdf_import" || source == "prior_year" {
            conn.execute(
                "INSERT OR IGNORE INTO field_values (return_id, field_key, value_num, value_str, source) \
                 VALUES (?1, ?2, ?3, ?4, 'prior_year')",
                rusqlite::params![new_id, key, num, str_val],
            )
            .unwrap();
            fields_carried += 1;
        }
    }

    // Drop the connection lock before calling solve_and_diff (which acquires it)
    drop(conn);

    // Run the solver on the new return to compute all dependent fields
    let _ = solve_and_diff(&state, &new_id, target_year);

    Ok((
        StatusCode::CREATED,
        Json(RollforwardResponse {
            return_id: new_id,
            tax_year: target_year,
            fields_carried,
            prior_year_values: prior_year_count,
        }),
    ))
}

async fn get_prior_year_values(
    State(state): State<Arc<AppState>>,
    Path(id): Path<String>,
) -> Result<Json<Vec<PriorYearValue>>, StatusCode> {
    let conn = state.db.conn();

    // Verify the return exists
    conn.query_row(
        "SELECT 1 FROM tax_returns WHERE id = ?1",
        [&id],
        |_| Ok(()),
    )
    .map_err(|_| StatusCode::NOT_FOUND)?;

    let mut stmt = conn
        .prepare(
            "SELECT field_key, source_year, value_num, value_str \
             FROM prior_year_values WHERE return_id = ?1 ORDER BY field_key",
        )
        .unwrap();
    let values = stmt
        .query_map([&id], |row| {
            Ok(PriorYearValue {
                field_key: row.get(0)?,
                source_year: row.get(1)?,
                value_num: row.get(2)?,
                value_str: row.get(3)?,
            })
        })
        .unwrap()
        .filter_map(|r| r.ok())
        .collect();

    Ok(Json(values))
}
