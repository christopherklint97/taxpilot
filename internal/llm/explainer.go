package llm

import (
	"context"
	"fmt"
	"strings"
)

// Explainer generates plain-English explanations for tax questions.
type Explainer struct {
	client *Client
	cache  *Cache
}

// NewExplainer creates an Explainer backed by the given client.
// A response cache is initialised automatically.
func NewExplainer(client *Client) *Explainer {
	cache := NewCache("")
	_ = cache.Load() // best-effort load from disk
	return &Explainer{
		client: client,
		cache:  cache,
	}
}

// NewExplainerWithCache creates an Explainer with a caller-provided cache.
// Useful for testing or when a custom cache directory is desired.
func NewExplainerWithCache(client *Client, cache *Cache) *Explainer {
	return &Explainer{
		client: client,
		cache:  cache,
	}
}

// interviewSystemPrompt is loaded once and reused.
const interviewSystemPrompt = `You are a helpful tax preparation assistant for TaxPilot, a US federal and California state tax filing tool for tax year 2025.

Your role is to explain tax concepts in plain English. You help users understand:
- What each question on their tax forms means
- Why they need to provide certain information
- How California tax treatment may differ from federal
- What their options are and how each choice affects their return

Rules:
- Be concise (2-3 sentences for simple fields, 4-5 for complex ones)
- Use plain English, not tax jargon
- Never suggest specific tax strategies or amounts
- Never perform calculations — the software handles all math
- If unsure, say so and suggest consulting a tax professional
- Always clarify when California treatment differs from federal`

const explainerSystemPrompt = `You are a tax code reference assistant. When asked about a tax topic, provide:
1. A brief plain-English explanation (1-2 sentences)
2. The relevant IRC section or CA Revenue & Taxation Code reference
3. Any important limits, thresholds, or phase-outs for 2025

Keep responses under 100 words. Be factual and precise.`

const caAdjustmentsContext = `When explaining California tax adjustments, note these key differences from federal:

- Social Security benefits: Not taxable in CA (exempt from state tax)
- SALT deduction: CA does not allow deduction for state taxes paid
- Standard deduction: CA uses $5,706 (single) / $11,412 (MFJ) — much lower than federal
- QBI deduction (Section 199A): CA does not conform — must add back on Schedule CA
- Tax brackets: CA has 9 brackets (1%-12.3%) plus 1% Mental Health Services surcharge on income over $1M
- Municipal bond interest: Only CA-issued bonds are exempt from CA tax
- HSA: CA does not conform to federal HSA tax treatment
- 529 plans: CA contributions are not deductible (unlike some states)

When a CA difference is relevant, always explain both the federal and CA treatment.`

// ExplainField generates a plain-English explanation of what a form field means
// and why the user needs to provide it.
func (e *Explainer) ExplainField(ctx context.Context, fieldKey, label, formName string, priorValue string) (string, error) {
	userContent := fmt.Sprintf(
		"Explain this tax form field to the user.\nForm: %s\nField key: %s\nLabel: %s",
		formName, fieldKey, label,
	)
	if priorValue != "" {
		userContent += fmt.Sprintf("\nLast year's value: %s", priorValue)
	}

	messages := []Message{
		{Role: "system", Content: interviewSystemPrompt},
		{Role: "user", Content: userContent},
	}

	return e.cachedChat(ctx, messages)
}

// ExplainCADifference explains why California treats something differently from federal.
func (e *Explainer) ExplainCADifference(ctx context.Context, area, federalTreatment, caTreatment string) (string, error) {
	userContent := fmt.Sprintf(
		"Explain this California vs. federal tax difference to the user.\nArea: %s\nFederal treatment: %s\nCalifornia treatment: %s",
		area, federalTreatment, caTreatment,
	)

	messages := []Message{
		{Role: "system", Content: explainerSystemPrompt + "\n\n" + caAdjustmentsContext},
		{Role: "user", Content: userContent},
	}

	return e.cachedChat(ctx, messages)
}

// ExplainWhyAsked explains why a particular question is being asked in context.
func (e *Explainer) ExplainWhyAsked(ctx context.Context, fieldKey, label string, filingStatus string, answeredSoFar map[string]string) (string, error) {
	var sb strings.Builder
	fmt.Fprintf(&sb,
		"The user is being asked this question during their tax interview.\nField key: %s\nLabel: %s\nFiling status: %s\n",
		fieldKey, label, filingStatus,
	)

	if len(answeredSoFar) > 0 {
		sb.WriteString("Context — answers so far:\n")
		for k, v := range answeredSoFar {
			fmt.Fprintf(&sb, "  %s = %s\n", k, v)
		}
	}

	sb.WriteString("\nExplain briefly why this question is needed to complete their return.")

	messages := []Message{
		{Role: "system", Content: interviewSystemPrompt},
		{Role: "user", Content: sb.String()},
	}

	return e.cachedChat(ctx, messages)
}

// AskAboutField answers a free-form user question in the context of the
// current interview field and previously answered questions.
func (e *Explainer) AskAboutField(ctx context.Context, userQuestion, fieldKey, label, formName, filingStatus string, answeredSoFar map[string]string) (string, error) {
	var sb strings.Builder
	fmt.Fprintf(&sb,
		"The user is filling out their tax return and is currently on this question:\nForm: %s\nField: %s\nQuestion: %s\nFiling status: %s\n",
		formName, fieldKey, label, filingStatus,
	)

	if len(answeredSoFar) > 0 {
		sb.WriteString("\nContext — answers so far:\n")
		count := 0
		for k, v := range answeredSoFar {
			if v == "" || v == "0" {
				continue
			}
			fmt.Fprintf(&sb, "  %s = %s\n", k, v)
			count++
			if count >= 100 {
				sb.WriteString("  ... (truncated)\n")
				break
			}
		}
	}

	fmt.Fprintf(&sb, "\nThe user asks: %s", userQuestion)

	messages := []Message{
		{Role: "system", Content: interviewSystemPrompt},
		{Role: "user", Content: sb.String()},
	}

	// Don't cache free-form questions — each is unique
	return e.client.Chat(ctx, messages)
}

// ExplainFieldStream streams a field explanation. Returns cached result as a
// closed single-element channel, or a live stream channel.
func (e *Explainer) ExplainFieldStream(ctx context.Context, fieldKey, label, formName string, priorValue string) (<-chan StreamChunk, error) {
	userContent := fmt.Sprintf(
		"Explain this tax form field to the user.\nForm: %s\nField key: %s\nLabel: %s",
		formName, fieldKey, label,
	)
	if priorValue != "" {
		userContent += fmt.Sprintf("\nLast year's value: %s", priorValue)
	}

	messages := []Message{
		{Role: "system", Content: interviewSystemPrompt},
		{Role: "user", Content: userContent},
	}

	return e.cachedStream(ctx, messages)
}

// ExplainWhyAskedStream streams a "why asked" explanation.
func (e *Explainer) ExplainWhyAskedStream(ctx context.Context, fieldKey, label string, filingStatus string, answeredSoFar map[string]string) (<-chan StreamChunk, error) {
	var sb strings.Builder
	fmt.Fprintf(&sb,
		"The user is being asked this question during their tax interview.\nField key: %s\nLabel: %s\nFiling status: %s\n",
		fieldKey, label, filingStatus,
	)

	if len(answeredSoFar) > 0 {
		sb.WriteString("Context — answers so far:\n")
		for k, v := range answeredSoFar {
			fmt.Fprintf(&sb, "  %s = %s\n", k, v)
		}
	}

	sb.WriteString("\nExplain briefly why this question is needed to complete their return.")

	messages := []Message{
		{Role: "system", Content: interviewSystemPrompt},
		{Role: "user", Content: sb.String()},
	}

	return e.cachedStream(ctx, messages)
}

// ExplainCADifferenceStream streams a CA vs federal difference explanation.
func (e *Explainer) ExplainCADifferenceStream(ctx context.Context, area, federalTreatment, caTreatment string) (<-chan StreamChunk, error) {
	userContent := fmt.Sprintf(
		"Explain this California vs. federal tax difference to the user.\nArea: %s\nFederal treatment: %s\nCalifornia treatment: %s",
		area, federalTreatment, caTreatment,
	)

	messages := []Message{
		{Role: "system", Content: explainerSystemPrompt + "\n\n" + caAdjustmentsContext},
		{Role: "user", Content: userContent},
	}

	return e.cachedStream(ctx, messages)
}

// AskAboutFieldStream streams a free-form AI answer.
func (e *Explainer) AskAboutFieldStream(ctx context.Context, userQuestion, fieldKey, label, formName, filingStatus string, answeredSoFar map[string]string) (<-chan StreamChunk, error) {
	var sb strings.Builder
	fmt.Fprintf(&sb,
		"The user is filling out their tax return and is currently on this question:\nForm: %s\nField: %s\nQuestion: %s\nFiling status: %s\n",
		formName, fieldKey, label, filingStatus,
	)

	if len(answeredSoFar) > 0 {
		sb.WriteString("\nContext — answers so far:\n")
		count := 0
		for k, v := range answeredSoFar {
			if v == "" || v == "0" {
				continue
			}
			fmt.Fprintf(&sb, "  %s = %s\n", k, v)
			count++
			if count >= 100 {
				sb.WriteString("  ... (truncated)\n")
				break
			}
		}
	}

	fmt.Fprintf(&sb, "\nThe user asks: %s", userQuestion)

	messages := []Message{
		{Role: "system", Content: interviewSystemPrompt},
		{Role: "user", Content: sb.String()},
	}

	// Don't cache free-form questions
	return e.client.ChatStream(ctx, messages)
}

// cachedStream checks the cache and returns a fake channel for cache hits,
// or a real streaming channel for cache misses.
func (e *Explainer) cachedStream(ctx context.Context, messages []Message) (<-chan StreamChunk, error) {
	key := e.cache.HashKey(messages)

	if cached, ok := e.cache.Get(key); ok {
		ch := make(chan StreamChunk, 2)
		ch <- StreamChunk{Text: cached}
		ch <- StreamChunk{Done: true}
		close(ch)
		return ch, nil
	}

	liveCh, err := e.client.ChatStream(ctx, messages)
	if err != nil {
		return nil, err
	}

	// Wrap the live channel to capture the full response for caching.
	cacheCh := make(chan StreamChunk, 8)
	go func() {
		defer close(cacheCh)
		var full strings.Builder
		for chunk := range liveCh {
			cacheCh <- chunk
			if chunk.Text != "" {
				full.WriteString(chunk.Text)
			}
			if chunk.Done || chunk.Err != nil {
				if chunk.Err == nil && full.Len() > 0 {
					e.cache.Set(key, full.String())
					_ = e.cache.Save()
				}
				return
			}
		}
	}()

	return cacheCh, nil
}

// cachedChat checks the cache before calling the LLM, then caches the result.
func (e *Explainer) cachedChat(ctx context.Context, messages []Message) (string, error) {
	key := e.cache.HashKey(messages)

	if cached, ok := e.cache.Get(key); ok {
		return cached, nil
	}

	resp, err := e.client.Chat(ctx, messages)
	if err != nil {
		return "", err
	}

	e.cache.Set(key, resp)
	_ = e.cache.Save() // best-effort persist

	return resp, nil
}
