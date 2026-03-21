package knowledge

import (
	"context"
	"fmt"
	"strings"

	"taxpilot/internal/llm"
)

// RAG provides retrieval-augmented generation for tax questions.
type RAG struct {
	store  *Store
	client *llm.Client
}

// NewRAG creates a new RAG instance with the given store and LLM client.
func NewRAG(store *Store, client *llm.Client) *RAG {
	return &RAG{
		store:  store,
		client: client,
	}
}

// Query searches the knowledge base and generates an LLM-enhanced answer.
func (r *RAG) Query(ctx context.Context, question string, jurisdiction Jurisdiction) (string, error) {
	results := r.store.Search(question, jurisdiction, 5)
	return r.ExplainWithContext(ctx, question, results)
}

// QueryForField searches the knowledge base for context relevant to a specific
// form field and returns an LLM-generated explanation.
func (r *RAG) QueryForField(ctx context.Context, fieldKey, label string, jurisdiction Jurisdiction) (string, error) {
	// Build a search query from field key and label
	query := label + " " + fieldKey
	results := r.store.Search(query, jurisdiction, 3)

	// Also search across all jurisdictions if the field might have
	// cross-jurisdiction relevance (e.g., a CA field referencing federal concepts)
	if jurisdiction == JurisdictionCA {
		federalResults := r.store.Search(query, JurisdictionFederal, 2)
		results = append(results, federalResults...)
	}

	// Cap at 5 total results
	if len(results) > 5 {
		results = results[:5]
	}

	if len(results) == 0 {
		// No context found — ask the LLM without context
		messages := []llm.Message{
			{Role: "system", Content: ragSystemPrompt},
			{Role: "user", Content: fmt.Sprintf("Explain the tax form field '%s' (key: %s) in plain English. What does it mean and why is it needed on the tax return?", label, fieldKey)},
		}
		return r.client.Chat(ctx, messages)
	}

	userContent := fmt.Sprintf(
		"The user needs help understanding a tax form field.\nField key: %s\nField label: %s\n\nUsing the provided tax code references, explain what this field means and why it matters.",
		fieldKey, label,
	)

	return r.explainWithDocs(ctx, userContent, results)
}

// ExplainWithContext retrieves relevant knowledge and asks the LLM to explain.
func (r *RAG) ExplainWithContext(ctx context.Context, question string, docs []SearchResult) (string, error) {
	if len(docs) == 0 {
		// No context — plain LLM query
		messages := []llm.Message{
			{Role: "system", Content: ragSystemPrompt},
			{Role: "user", Content: question},
		}
		return r.client.Chat(ctx, messages)
	}

	return r.explainWithDocs(ctx, question, docs)
}

// explainWithDocs sends a question to the LLM with document context.
func (r *RAG) explainWithDocs(ctx context.Context, question string, docs []SearchResult) (string, error) {
	contextText := FormatContext(docs)

	systemPrompt := ragSystemPrompt + "\n\n" + ragContextPrefix + contextText

	messages := []llm.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: question},
	}

	return r.client.Chat(ctx, messages)
}

// FormatContext formats search results as context text for an LLM prompt.
func FormatContext(results []SearchResult) string {
	if len(results) == 0 {
		return ""
	}

	var sb strings.Builder
	for i, r := range results {
		fmt.Fprintf(&sb, "--- Reference %d: %s (%s) ---\n", i+1, r.Document.Title, r.Document.Source)
		sb.WriteString(r.Document.Content)
		sb.WriteString("\n\n")
	}
	return sb.String()
}

const ragSystemPrompt = `You are a tax preparation assistant for TaxPilot, a US federal and California state tax filing tool for tax year 2025.

Your role is to explain tax concepts in plain English using the provided tax code references. Follow these rules:
- Be concise (3-5 sentences)
- Use plain English, avoid unnecessary jargon
- Cite the relevant tax code section when applicable
- Never suggest specific tax strategies or dollar amounts
- Never perform calculations
- When California treatment differs from federal, explain both
- If the provided references don't cover the question, say so honestly`

// QueryForFieldStream is like QueryForField but returns a streaming channel.
func (r *RAG) QueryForFieldStream(ctx context.Context, fieldKey, label string, jurisdiction Jurisdiction) (<-chan llm.StreamChunk, error) {
	query := label + " " + fieldKey
	results := r.store.Search(query, jurisdiction, 3)

	if jurisdiction == JurisdictionCA {
		federalResults := r.store.Search(query, JurisdictionFederal, 2)
		results = append(results, federalResults...)
	}

	if len(results) > 5 {
		results = results[:5]
	}

	if len(results) == 0 {
		messages := []llm.Message{
			{Role: "system", Content: ragSystemPrompt},
			{Role: "user", Content: fmt.Sprintf("Explain the tax form field '%s' (key: %s) in plain English. What does it mean and why is it needed on the tax return?", label, fieldKey)},
		}
		return r.client.ChatStream(ctx, messages)
	}

	contextText := FormatContext(results)
	systemPrompt := ragSystemPrompt + "\n\n" + ragContextPrefix + contextText

	userContent := fmt.Sprintf(
		"The user needs help understanding a tax form field.\nField key: %s\nField label: %s\n\nUsing the provided tax code references, explain what this field means and why it matters.",
		fieldKey, label,
	)

	messages := []llm.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userContent},
	}
	return r.client.ChatStream(ctx, messages)
}

const ragContextPrefix = `Use the following tax code references to inform your answer:

`
