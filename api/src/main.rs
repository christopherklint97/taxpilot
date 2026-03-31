use std::sync::Arc;

use axum::{Json, Router, routing::get};
use serde::Serialize;
use tokio::net::TcpListener;
use tower_http::cors::CorsLayer;
use tracing_subscriber::EnvFilter;

mod db;
pub mod domain;
pub mod efile;
pub mod forms;
pub mod llm;
pub mod pdf;
mod routes;

#[cfg(test)]
mod tests;

pub struct AppState {
    pub db: db::Database,
    pub registry: domain::registry::Registry,
    pub llm: llm::LlmClient,
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        .with_env_filter(EnvFilter::from_default_env().add_directive("info".parse().unwrap()))
        .init();

    let database_url =
        std::env::var("DATABASE_URL").unwrap_or_else(|_| "taxpilot.db".to_string());
    let db = db::Database::new(&database_url).expect("Failed to initialize database");
    let registry = forms::register_all_forms();

    tracing::info!(
        "Registered {} forms with {} total fields",
        registry.all_forms().len(),
        registry
            .all_forms()
            .iter()
            .map(|f| f.fields.len())
            .sum::<usize>()
    );

    let llm = llm::LlmClient::new();
    if llm.is_configured() {
        tracing::info!("LLM configured with model {}", llm.model());
    } else {
        tracing::warn!("OPENROUTER_API_KEY not set — explain endpoint will return fallback messages");
    }

    let state = Arc::new(AppState { db, registry, llm });

    // Merge all sub-routers that nest under /api/returns/{id}
    let return_detail_router = routes::fields::router()
        .merge(routes::pdf::router())
        .merge(routes::rollforward::router())
        .merge(routes::efile::router());

    let app = Router::new()
        .route("/api/health", get(health))
        .merge(routes::explain::router())
        .nest("/api/returns", routes::returns::router())
        .nest("/api/returns/{id}", return_detail_router)
        .nest("/api/forms", routes::forms_meta::router())
        .with_state(state)
        .layer(CorsLayer::permissive());

    let port = std::env::var("PORT").unwrap_or_else(|_| "4100".to_string());
    let addr = format!("0.0.0.0:{port}");
    tracing::info!("TaxPilot API listening on {addr}");

    let listener = TcpListener::bind(&addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}

#[derive(Serialize)]
struct HealthResponse {
    status: &'static str,
    version: &'static str,
}

async fn health() -> Json<HealthResponse> {
    Json(HealthResponse {
        status: "ok",
        version: env!("CARGO_PKG_VERSION"),
    })
}
