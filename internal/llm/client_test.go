package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// --- NewClient tests ---

func TestNewClientWithKey(t *testing.T) {
	c, err := NewClient("test-key-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.apiKey != "test-key-123" {
		t.Errorf("expected apiKey test-key-123, got %s", c.apiKey)
	}
	if c.model != DefaultModel {
		t.Errorf("expected model %s, got %s", DefaultModel, c.model)
	}
}

func TestNewClientFromEnv(t *testing.T) {
	t.Setenv("OPENROUTER_API_KEY", "env-key-456")
	c, err := NewClient("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.apiKey != "env-key-456" {
		t.Errorf("expected apiKey env-key-456, got %s", c.apiKey)
	}
}

func TestNewClientMissingKey(t *testing.T) {
	t.Setenv("OPENROUTER_API_KEY", "")
	_, err := NewClient("")
	if err == nil {
		t.Fatal("expected error when no API key provided")
	}
}

func TestSetModel(t *testing.T) {
	c, _ := NewClient("key")
	c.SetModel("openai/gpt-4o")
	if c.model != "openai/gpt-4o" {
		t.Errorf("expected model openai/gpt-4o, got %s", c.model)
	}
}

// --- Request serialization tests ---

func TestRequestSerialization(t *testing.T) {
	req := ChatRequest{
		Model: "anthropic/claude-sonnet-4",
		Messages: []Message{
			{Role: "system", Content: "You are helpful."},
			{Role: "user", Content: "Hello"},
		},
		MaxTokens:   512,
		Temperature: 0.5,
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if parsed["model"] != "anthropic/claude-sonnet-4" {
		t.Errorf("unexpected model: %v", parsed["model"])
	}
	msgs := parsed["messages"].([]any)
	if len(msgs) != 2 {
		t.Errorf("expected 2 messages, got %d", len(msgs))
	}
	if parsed["max_tokens"].(float64) != 512 {
		t.Errorf("unexpected max_tokens: %v", parsed["max_tokens"])
	}
	if parsed["temperature"].(float64) != 0.5 {
		t.Errorf("unexpected temperature: %v", parsed["temperature"])
	}
}

func TestRequestOmitsZeroOptionalFields(t *testing.T) {
	req := ChatRequest{
		Model:    "test",
		Messages: []Message{{Role: "user", Content: "hi"}},
	}
	data, _ := json.Marshal(req)
	var parsed map[string]any
	json.Unmarshal(data, &parsed)

	if _, ok := parsed["max_tokens"]; ok {
		t.Error("max_tokens should be omitted when zero")
	}
	// temperature 0 is a valid float, json omitempty won't omit it for float64
	// so we don't check that here
}

// --- Response parsing tests (mock HTTP server) ---

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := NewClient("test-key")
	if err != nil {
		t.Fatal(err)
	}
	c.baseURL = srv.URL
	return c
}

func TestChatSuccess(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("missing or wrong Authorization header")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("missing Content-Type header")
		}
		if r.Header.Get("X-Title") != "TaxPilot" {
			t.Errorf("missing X-Title header")
		}

		resp := ChatResponse{
			Choices: []Choice{
				{Message: Message{Role: "assistant", Content: "Hello there!"}},
			},
			Usage: Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	result, err := c.Chat(context.Background(), []Message{
		{Role: "user", Content: "Hi"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Hello there!" {
		t.Errorf("expected 'Hello there!', got %q", result)
	}
}

func TestChatNon200Status(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":{"message":"rate limited"}}`))
	})

	_, err := c.Chat(context.Background(), []Message{
		{Role: "user", Content: "Hi"},
	})
	if err == nil {
		t.Fatal("expected error for non-200 status")
	}
	if got := err.Error(); !contains(got, "status 429") {
		t.Errorf("error should mention status code, got: %s", got)
	}
}

func TestChatAPIErrorInBody(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		resp := ChatResponse{
			Error: &APIError{Message: "invalid model", Type: "invalid_request", Code: 400},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	_, err := c.Chat(context.Background(), []Message{
		{Role: "user", Content: "Hi"},
	})
	if err == nil {
		t.Fatal("expected error for API error in body")
	}
	if got := err.Error(); !contains(got, "invalid model") {
		t.Errorf("error should contain API message, got: %s", got)
	}
}

func TestChatEmptyChoices(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		resp := ChatResponse{Choices: []Choice{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	_, err := c.Chat(context.Background(), []Message{
		{Role: "user", Content: "Hi"},
	})
	if err == nil {
		t.Fatal("expected error for empty choices")
	}
	if got := err.Error(); !contains(got, "no choices") {
		t.Errorf("error should mention no choices, got: %s", got)
	}
}

func TestChatInvalidJSON(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{not valid json`))
	})

	_, err := c.Chat(context.Background(), []Message{
		{Role: "user", Content: "Hi"},
	})
	if err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
}

// --- Cache tests ---

func TestCacheGetSet(t *testing.T) {
	c := NewCache("")
	_, ok := c.Get("missing")
	if ok {
		t.Error("expected cache miss")
	}

	c.Set("key1", "value1")
	val, ok := c.Get("key1")
	if !ok {
		t.Error("expected cache hit")
	}
	if val != "value1" {
		t.Errorf("expected value1, got %s", val)
	}
}

func TestCacheHashKey(t *testing.T) {
	c := NewCache("")
	msgs1 := []Message{{Role: "user", Content: "hello"}}
	msgs2 := []Message{{Role: "user", Content: "hello"}}
	msgs3 := []Message{{Role: "user", Content: "world"}}

	h1 := c.HashKey(msgs1)
	h2 := c.HashKey(msgs2)
	h3 := c.HashKey(msgs3)

	if h1 != h2 {
		t.Error("identical messages should produce same hash")
	}
	if h1 == h3 {
		t.Error("different messages should produce different hashes")
	}
	if len(h1) != 64 {
		t.Errorf("expected 64-char hex SHA-256, got length %d", len(h1))
	}
}

func TestCachePersistence(t *testing.T) {
	dir := t.TempDir()
	c1 := NewCache(dir)
	c1.Set("persist-key", "persist-value")
	if err := c1.Save(); err != nil {
		t.Fatalf("save error: %v", err)
	}

	// Verify file exists
	path := filepath.Join(dir, "cache.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("cache file not created: %v", err)
	}

	// Load into new cache
	c2 := NewCache(dir)
	if err := c2.Load(); err != nil {
		t.Fatalf("load error: %v", err)
	}
	val, ok := c2.Get("persist-key")
	if !ok {
		t.Error("expected cache hit after load")
	}
	if val != "persist-value" {
		t.Errorf("expected persist-value, got %s", val)
	}
}

func TestCacheLoadMissingFile(t *testing.T) {
	dir := t.TempDir()
	c := NewCache(dir)
	// Should not error on missing file
	if err := c.Load(); err != nil {
		t.Fatalf("load should not fail for missing file: %v", err)
	}
}

// --- Explainer message building tests ---

func TestExplainerExplainField(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ChatRequest
		json.NewDecoder(r.Body).Decode(&req)

		// Verify system prompt is included
		if req.Messages[0].Role != "system" {
			t.Error("first message should be system")
		}
		// Verify user prompt contains field info
		userMsg := req.Messages[1].Content
		if !contains(userMsg, "wages") || !contains(userMsg, "Form 1040") {
			t.Errorf("user message should contain field info, got: %s", userMsg)
		}

		resp := ChatResponse{
			Choices: []Choice{
				{Message: Message{Role: "assistant", Content: "This is your total wages."}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client, _ := NewClient("test-key")
	client.baseURL = srv.URL
	explainer := NewExplainerWithCache(client, NewCache(t.TempDir()))

	result, err := explainer.ExplainField(context.Background(), "wages", "Wages, salaries, tips", "Form 1040", "75000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "This is your total wages." {
		t.Errorf("unexpected result: %s", result)
	}
}

func TestExplainerCachesResponses(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		resp := ChatResponse{
			Choices: []Choice{
				{Message: Message{Role: "assistant", Content: "Cached answer"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client, _ := NewClient("test-key")
	client.baseURL = srv.URL
	explainer := NewExplainerWithCache(client, NewCache(t.TempDir()))

	ctx := context.Background()
	// First call hits API
	_, _ = explainer.ExplainField(ctx, "wages", "Wages", "1040", "")
	// Second identical call should use cache
	_, _ = explainer.ExplainField(ctx, "wages", "Wages", "1040", "")

	if callCount != 1 {
		t.Errorf("expected 1 API call (cached), got %d", callCount)
	}
}

func TestExplainerCADifference(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ChatRequest
		json.NewDecoder(r.Body).Decode(&req)

		// System prompt should include CA adjustments context
		sysMsg := req.Messages[0].Content
		if !contains(sysMsg, "California") {
			t.Error("system message should include CA context")
		}

		resp := ChatResponse{
			Choices: []Choice{
				{Message: Message{Role: "assistant", Content: "CA difference explained."}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client, _ := NewClient("test-key")
	client.baseURL = srv.URL
	explainer := NewExplainerWithCache(client, NewCache(t.TempDir()))

	result, err := explainer.ExplainCADifference(context.Background(), "SALT", "Deductible up to $10K", "Not deductible")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "CA difference explained." {
		t.Errorf("unexpected result: %s", result)
	}
}

func TestExplainerWhyAsked(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ChatRequest
		json.NewDecoder(r.Body).Decode(&req)

		userMsg := req.Messages[1].Content
		if !contains(userMsg, "Single") || !contains(userMsg, "filing_status") {
			t.Errorf("user message should include context, got: %s", userMsg)
		}

		resp := ChatResponse{
			Choices: []Choice{
				{Message: Message{Role: "assistant", Content: "This is needed because..."}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client, _ := NewClient("test-key")
	client.baseURL = srv.URL
	explainer := NewExplainerWithCache(client, NewCache(t.TempDir()))

	answered := map[string]string{"filing_status": "Single", "wages": "75000"}
	result, err := explainer.ExplainWhyAsked(context.Background(), "itemized_deductions", "Itemized Deductions", "Single", answered)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "This is needed because..." {
		t.Errorf("unexpected result: %s", result)
	}
}

// helper
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
