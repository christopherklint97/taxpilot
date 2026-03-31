use std::sync::Arc;

use axum::{
    Json, Router,
    extract::{Path, State},
    http::{StatusCode, header},
    routing::{get, post},
    response::IntoResponse,
};
use serde::Serialize;

use crate::AppState;
use crate::efile;
use crate::routes::fields;

pub fn router() -> Router<Arc<AppState>> {
    Router::new()
        .route("/validate", get(validate))
        .route("/efile/mef", post(generate_mef))
        .route("/efile/ca", post(generate_ca))
}

#[derive(Serialize)]
struct ValidateResponse {
    is_valid: bool,
    results: Vec<efile::validate::ValidationResult>,
}

/// GET /api/returns/{id}/validate
///
/// Runs the solver, then runs all validation and reasonableness checks.
async fn validate(
    State(state): State<Arc<AppState>>,
    Path(id): Path<String>,
) -> Result<Json<ValidateResponse>, StatusCode> {
    let conn = state.db.conn();

    let tax_year = fields::get_tax_year(&conn, &id)?;

    let state_code: String = conn
        .query_row(
            "SELECT state_code FROM tax_returns WHERE id = ?1",
            [&id],
            |row| row.get(0),
        )
        .map_err(|_| StatusCode::NOT_FOUND)?;

    // Ensure defaults and solve
    fields::ensure_defaults(&conn, &id, &state.registry);
    let (inputs, str_inputs, _) = fields::load_field_values(&conn, &id);

    let mut graph = crate::domain::solver::DependencyGraph::new(&state.registry);
    graph.build().map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;
    let solved = graph
        .solve(&inputs, &str_inputs, tax_year)
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    // Merge user inputs into solved values for validation
    let mut all_values = solved;
    for (k, v) in &inputs {
        all_values.entry(k.clone()).or_insert(*v);
    }

    let report = efile::validate::full_validation(&all_values, &str_inputs, tax_year, &state_code);

    Ok(Json(ValidateResponse {
        is_valid: report.is_valid,
        results: report.results,
    }))
}

/// POST /api/returns/{id}/efile/mef
///
/// Generates MeF XML for federal e-filing.
async fn generate_mef(
    State(state): State<Arc<AppState>>,
    Path(id): Path<String>,
) -> Result<impl IntoResponse, StatusCode> {
    let conn = state.db.conn();

    let tax_year = fields::get_tax_year(&conn, &id)?;

    fields::ensure_defaults(&conn, &id, &state.registry);
    let (inputs, str_inputs, _) = fields::load_field_values(&conn, &id);

    let mut graph = crate::domain::solver::DependencyGraph::new(&state.registry);
    graph.build().map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;
    let solved = graph
        .solve(&inputs, &str_inputs, tax_year)
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    let mut all_values = solved;
    for (k, v) in &inputs {
        all_values.entry(k.clone()).or_insert(*v);
    }

    let xml = efile::mef::generate_mef_xml(&all_values, &str_inputs, tax_year)
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    Ok((
        StatusCode::OK,
        [
            (header::CONTENT_TYPE, "application/xml"),
            (
                header::CONTENT_DISPOSITION,
                "attachment; filename=\"federal_return.xml\"",
            ),
        ],
        xml,
    ))
}

/// POST /api/returns/{id}/efile/ca
///
/// Generates CA FTB XML for state e-filing.
async fn generate_ca(
    State(state): State<Arc<AppState>>,
    Path(id): Path<String>,
) -> Result<impl IntoResponse, StatusCode> {
    let conn = state.db.conn();

    let tax_year = fields::get_tax_year(&conn, &id)?;

    fields::ensure_defaults(&conn, &id, &state.registry);
    let (inputs, str_inputs, _) = fields::load_field_values(&conn, &id);

    let mut graph = crate::domain::solver::DependencyGraph::new(&state.registry);
    graph.build().map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;
    let solved = graph
        .solve(&inputs, &str_inputs, tax_year)
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    let mut all_values = solved;
    for (k, v) in &inputs {
        all_values.entry(k.clone()).or_insert(*v);
    }

    let xml = efile::ca::generate_ca_xml(&all_values, &str_inputs, tax_year)
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    Ok((
        StatusCode::OK,
        [
            (header::CONTENT_TYPE, "application/xml"),
            (
                header::CONTENT_DISPOSITION,
                "attachment; filename=\"ca_return.xml\"",
            ),
        ],
        xml,
    ))
}
