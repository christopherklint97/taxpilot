use std::sync::Arc;

use axum::{
    Json, Router,
    extract::{Path, State},
    http::StatusCode,
    routing::get,
};
use serde::{Deserialize, Serialize};

use crate::AppState;

pub fn router() -> Router<Arc<AppState>> {
    Router::new()
        .route("/", get(list_returns).post(create_return))
        .route("/{id}", get(get_return).delete(delete_return))
}

#[derive(Serialize)]
pub struct TaxReturnSummary {
    pub id: String,
    pub tax_year: i32,
    pub state_code: String,
    pub filing_status: Option<String>,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Serialize)]
pub struct FieldValue {
    pub field_key: String,
    pub value_num: Option<f64>,
    pub value_str: Option<String>,
    pub source: String,
}

#[derive(Serialize)]
pub struct TaxReturnDetail {
    #[serde(flatten)]
    pub summary: TaxReturnSummary,
    pub fields: Vec<FieldValue>,
}

#[derive(Deserialize)]
pub struct CreateReturn {
    pub tax_year: i32,
    pub state_code: Option<String>,
}

fn row_to_summary(row: &rusqlite::Row) -> rusqlite::Result<TaxReturnSummary> {
    Ok(TaxReturnSummary {
        id: row.get(0)?,
        tax_year: row.get(1)?,
        state_code: row.get(2)?,
        filing_status: row.get(3)?,
        created_at: row.get(4)?,
        updated_at: row.get(5)?,
    })
}

async fn list_returns(State(state): State<Arc<AppState>>) -> Json<Vec<TaxReturnSummary>> {
    let conn = state.db.conn();
    let mut stmt = conn
        .prepare(
            "SELECT id, tax_year, state_code, filing_status, created_at, updated_at \
             FROM tax_returns ORDER BY updated_at DESC",
        )
        .unwrap();
    let returns = stmt
        .query_map([], |row| row_to_summary(row))
        .unwrap()
        .filter_map(|r| r.ok())
        .collect();
    Json(returns)
}

async fn create_return(
    State(state): State<Arc<AppState>>,
    Json(body): Json<CreateReturn>,
) -> (StatusCode, Json<TaxReturnSummary>) {
    let id = uuid::Uuid::new_v4().to_string();
    let state_code = body.state_code.unwrap_or_else(|| "CA".to_string());
    let conn = state.db.conn();
    conn.execute(
        "INSERT INTO tax_returns (id, tax_year, state_code) VALUES (?1, ?2, ?3)",
        rusqlite::params![id, body.tax_year, state_code],
    )
    .unwrap();

    let ret = conn
        .query_row(
            "SELECT id, tax_year, state_code, filing_status, created_at, updated_at \
             FROM tax_returns WHERE id = ?1",
            [&id],
            |row| row_to_summary(row),
        )
        .unwrap();

    (StatusCode::CREATED, Json(ret))
}

async fn get_return(
    State(state): State<Arc<AppState>>,
    Path(id): Path<String>,
) -> Result<Json<TaxReturnDetail>, StatusCode> {
    let conn = state.db.conn();

    let summary = conn
        .query_row(
            "SELECT id, tax_year, state_code, filing_status, created_at, updated_at \
             FROM tax_returns WHERE id = ?1",
            [&id],
            |row| row_to_summary(row),
        )
        .map_err(|_| StatusCode::NOT_FOUND)?;

    let mut stmt = conn
        .prepare(
            "SELECT field_key, value_num, value_str, source \
             FROM field_values WHERE return_id = ?1 ORDER BY field_key",
        )
        .unwrap();
    let fields = stmt
        .query_map([&id], |row| {
            Ok(FieldValue {
                field_key: row.get(0)?,
                value_num: row.get(1)?,
                value_str: row.get(2)?,
                source: row.get(3)?,
            })
        })
        .unwrap()
        .filter_map(|r| r.ok())
        .collect();

    Ok(Json(TaxReturnDetail { summary, fields }))
}

async fn delete_return(
    State(state): State<Arc<AppState>>,
    Path(id): Path<String>,
) -> StatusCode {
    let conn = state.db.conn();
    let rows = conn
        .execute("DELETE FROM tax_returns WHERE id = ?1", [&id])
        .unwrap();
    if rows > 0 {
        StatusCode::NO_CONTENT
    } else {
        StatusCode::NOT_FOUND
    }
}
