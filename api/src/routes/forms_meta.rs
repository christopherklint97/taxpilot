use std::sync::Arc;

use axum::{
    Json, Router,
    extract::{Path, State},
    http::StatusCode,
    routing::get,
};
use serde::Serialize;

use crate::AppState;
use crate::domain::field::{FieldType, FieldValueType};

pub fn router() -> Router<Arc<AppState>> {
    Router::new()
        .route("/", get(list_forms))
        .route("/{form_id}", get(get_form))
}

#[derive(Serialize)]
struct FormMeta {
    id: String,
    name: String,
    jurisdiction: String,
    question_group: String,
    question_order: i32,
    field_count: usize,
}

#[derive(Serialize)]
struct FormDetail {
    id: String,
    name: String,
    jurisdiction: String,
    question_group: String,
    question_order: i32,
    fields: Vec<FieldMeta>,
}

#[derive(Serialize)]
struct FieldMeta {
    line: String,
    field_key: String,
    field_type: String,
    value_type: String,
    label: String,
    prompt: Option<String>,
    depends_on: Vec<String>,
    options: Vec<String>,
}

fn jurisdiction_str(j: &crate::domain::field::Jurisdiction) -> &'static str {
    match j {
        crate::domain::field::Jurisdiction::Federal => "federal",
        crate::domain::field::Jurisdiction::StateCA => "state_ca",
    }
}

fn field_type_str(ft: &FieldType) -> &'static str {
    match ft {
        FieldType::UserInput => "user_input",
        FieldType::Computed => "computed",
        FieldType::Lookup => "lookup",
        FieldType::PriorYear => "prior_year",
        FieldType::FederalRef => "federal_ref",
    }
}

fn value_type_str(vt: &FieldValueType) -> &'static str {
    match vt {
        FieldValueType::Numeric => "numeric",
        FieldValueType::String => "string",
        FieldValueType::Integer => "integer",
    }
}

async fn list_forms(State(state): State<Arc<AppState>>) -> Json<Vec<FormMeta>> {
    let mut forms: Vec<FormMeta> = state
        .registry
        .all_forms()
        .iter()
        .map(|f| FormMeta {
            id: f.id.clone(),
            name: f.name.clone(),
            jurisdiction: jurisdiction_str(&f.jurisdiction).to_string(),
            question_group: f.question_group.clone(),
            question_order: f.question_order,
            field_count: f.fields.len(),
        })
        .collect();

    // Also include input form metadata
    for f in crate::forms::all_input_forms() {
        forms.push(FormMeta {
            id: f.id.clone(),
            name: f.name.clone(),
            jurisdiction: jurisdiction_str(&f.jurisdiction).to_string(),
            question_group: f.question_group.clone(),
            question_order: f.question_order,
            field_count: f.fields.len(),
        });
    }

    forms.sort_by(|a, b| {
        a.question_group
            .cmp(&b.question_group)
            .then(a.question_order.cmp(&b.question_order))
    });
    Json(forms)
}

async fn get_form(
    State(state): State<Arc<AppState>>,
    Path(form_id): Path<String>,
) -> Result<Json<FormDetail>, StatusCode> {
    // Helper to convert a FormDef reference to FormDetail
    let to_detail = |form: &crate::domain::field::FormDef| -> FormDetail {
        let fields = form
            .fields
            .iter()
            .map(|f| FieldMeta {
                line: f.line.clone(),
                field_key: format!("{}:{}", form.id, f.line),
                field_type: field_type_str(&f.field_type).to_string(),
                value_type: value_type_str(&f.value_type).to_string(),
                label: f.label.clone(),
                prompt: if f.prompt.is_empty() {
                    None
                } else {
                    Some(f.prompt.clone())
                },
                depends_on: f.depends_on.clone(),
                options: f.options.clone(),
            })
            .collect();
        FormDetail {
            id: form.id.clone(),
            name: form.name.clone(),
            jurisdiction: jurisdiction_str(&form.jurisdiction).to_string(),
            question_group: form.question_group.clone(),
            question_order: form.question_order,
            fields,
        }
    };

    // Check registry first
    if let Some(form) = state.registry.get(&form_id) {
        return Ok(Json(to_detail(form)));
    }

    // Then check input forms
    let input_forms = crate::forms::all_input_forms();
    if let Some(form) = input_forms.iter().find(|f| f.id == form_id) {
        return Ok(Json(to_detail(form)));
    }

    Err(StatusCode::NOT_FOUND)
}
