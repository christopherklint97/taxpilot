use std::sync::Arc;

use axum::{Json, Router, extract::State, http::StatusCode, routing::post};
use serde::{Deserialize, Serialize};

use crate::AppState;

pub fn router() -> Router<Arc<AppState>> {
    Router::new().route("/api/explain", post(explain))
}

#[derive(Deserialize)]
struct ExplainRequest {
    field_key: String,
    context: Option<String>,
}

#[derive(Serialize)]
struct ExplainResponse {
    explanation: String,
    model: String,
    configured: bool,
}

async fn explain(
    State(state): State<Arc<AppState>>,
    Json(body): Json<ExplainRequest>,
) -> Result<Json<ExplainResponse>, StatusCode> {
    // Look up field label from registry
    let label = state
        .registry
        .get_field(&body.field_key)
        .map(|(_, field)| field.label.clone())
        .unwrap_or_else(|_| body.field_key.clone());

    let context = body.context.unwrap_or_default();

    if !state.llm.is_configured() {
        return Ok(Json(ExplainResponse {
            explanation: "LLM not configured. Set OPENROUTER_API_KEY to enable AI explanations."
                .to_string(),
            model: String::new(),
            configured: false,
        }));
    }

    match state.llm.explain(&body.field_key, &label, &context).await {
        Ok(explanation) => Ok(Json(ExplainResponse {
            explanation,
            model: state.llm.model().to_string(),
            configured: true,
        })),
        Err(e) => {
            tracing::warn!("LLM explain failed: {e}");
            Ok(Json(ExplainResponse {
                explanation: format!("Unable to generate explanation: {e}"),
                model: String::new(),
                configured: true,
            }))
        }
    }
}
