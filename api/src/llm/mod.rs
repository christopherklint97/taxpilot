use serde::{Deserialize, Serialize};

const DEFAULT_MODEL: &str = "anthropic/claude-sonnet-4-6";
const OPENROUTER_URL: &str = "https://openrouter.ai/api/v1/chat/completions";

#[derive(Clone)]
pub struct LlmClient {
    api_key: Option<String>,
    model: String,
    http: reqwest::Client,
}

#[derive(Serialize)]
struct ChatRequest {
    model: String,
    messages: Vec<Message>,
    max_tokens: u32,
    temperature: f64,
    #[serde(skip_serializing_if = "Option::is_none")]
    provider: Option<ProviderPreferences>,
}

#[derive(Serialize)]
struct ProviderPreferences {
    #[serde(skip_serializing_if = "std::ops::Not::not")]
    zdr: bool,
}

#[derive(Serialize, Deserialize, Clone)]
pub struct Message {
    pub role: String,
    pub content: String,
}

#[derive(Deserialize)]
struct ChatResponse {
    choices: Vec<Choice>,
}

#[derive(Deserialize)]
struct Choice {
    message: Message,
}

impl LlmClient {
    pub fn new() -> Self {
        let api_key = std::env::var("OPENROUTER_API_KEY").ok();
        let model =
            std::env::var("TAXPILOT_MODEL").unwrap_or_else(|_| DEFAULT_MODEL.to_string());
        Self {
            api_key,
            model,
            http: reqwest::Client::new(),
        }
    }

    pub fn is_configured(&self) -> bool {
        self.api_key.is_some()
    }

    pub fn model(&self) -> &str {
        &self.model
    }

    pub async fn explain(
        &self,
        field_key: &str,
        field_label: &str,
        context: &str,
    ) -> Result<String, String> {
        let api_key = self
            .api_key
            .as_ref()
            .ok_or_else(|| {
                "LLM not configured: set OPENROUTER_API_KEY environment variable".to_string()
            })?;

        let system_prompt = include_str!("../../../data/prompts/explainer_system.txt");

        let user_prompt = format!(
            "Explain this tax form field:\n\
             Field: {field_label}\n\
             Key: {field_key}\n\
             {context}"
        );

        let request = ChatRequest {
            model: self.model.clone(),
            messages: vec![
                Message {
                    role: "system".to_string(),
                    content: system_prompt.to_string(),
                },
                Message {
                    role: "user".to_string(),
                    content: user_prompt,
                },
            ],
            max_tokens: 300,
            temperature: 0.3,
            provider: Some(ProviderPreferences { zdr: true }),
        };

        let response = self
            .http
            .post(OPENROUTER_URL)
            .header("Authorization", format!("Bearer {api_key}"))
            .header("HTTP-Referer", "https://taxpilot.local")
            .header("X-Title", "TaxPilot")
            .json(&request)
            .send()
            .await
            .map_err(|e| format!("LLM request failed: {e}"))?;

        if !response.status().is_success() {
            let status = response.status();
            let body = response.text().await.unwrap_or_default();
            return Err(format!("LLM API error {status}: {body}"));
        }

        let chat_response: ChatResponse = response
            .json()
            .await
            .map_err(|e| format!("Failed to parse LLM response: {e}"))?;

        chat_response
            .choices
            .first()
            .map(|c| c.message.content.clone())
            .ok_or_else(|| "No response from LLM".to_string())
    }
}
