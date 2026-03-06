package knowledge

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"taxpilot/internal/llm"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"", nil},
		{"a", nil}, // too short
		{"standard deduction", []string{"standard", "deduction"}},
		{"IRC §199A QBI", []string{"irc", "199a", "qbi"}},
		{"The SALT deduction is capped", []string{"salt", "deduction", "capped"}}, // "the" and "is" are stop words
		{"W-2 Box 1 wages", []string{"box", "wages"}},                             // "1" too short, "w" too short after split
		{"Hello, World! 123", []string{"hello", "world", "123"}},
	}

	for _, tt := range tests {
		got := tokenize(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("tokenize(%q) = %v (len %d), want %v (len %d)", tt.input, got, len(got), tt.want, len(tt.want))
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("tokenize(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}

func TestScoring(t *testing.T) {
	store := NewStore()
	doc1 := Document{
		ID: "test1", Title: "Standard Deduction Amounts",
		Content: "The standard deduction for single filers is $15,000.",
		Source:  "IRC §63", Tags: []string{"standard deduction", "single"},
	}
	doc2 := Document{
		ID: "test2", Title: "Capital Gains Rates",
		Content: "Long-term capital gains are taxed at preferential rates.",
		Source:  "IRC §1(h)", Tags: []string{"capital gains", "investment"},
	}
	store.Add(doc1)
	store.Add(doc2)

	queryTokens := tokenize("standard deduction")

	s1 := score(doc1, queryTokens, 2, store.index)
	s2 := score(doc2, queryTokens, 2, store.index)

	if s1 <= 0 {
		t.Errorf("doc1 score for 'standard deduction' should be > 0, got %f", s1)
	}
	if s2 != 0 {
		t.Errorf("doc2 score for 'standard deduction' should be 0, got %f", s2)
	}
	if s1 <= s2 {
		t.Errorf("doc1 score (%f) should be greater than doc2 score (%f) for 'standard deduction'", s1, s2)
	}
}

func TestSearchFederal(t *testing.T) {
	store := SeedStore()
	results := store.Search("standard deduction", JurisdictionFederal, 5)

	if len(results) == 0 {
		t.Fatal("expected results for 'standard deduction' in federal jurisdiction")
	}

	// The top result should be about taxable income / standard deduction (IRC §63)
	found := false
	for _, r := range results {
		if strings.Contains(r.Document.Title, "Taxable Income") || strings.Contains(r.Document.Title, "Standard Deduction") || strings.Contains(r.Document.ID, "irc_63") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected IRC §63 (taxable income/standard deduction) in results, got: %v", resultIDs(results))
	}
}

func TestSearchCA(t *testing.T) {
	store := SeedStore()
	results := store.Search("mental health tax surcharge", JurisdictionCA, 5)

	if len(results) == 0 {
		t.Fatal("expected results for 'mental health tax surcharge' in CA jurisdiction")
	}

	found := false
	for _, r := range results {
		if strings.Contains(r.Document.ID, "ca_mental_health") || strings.Contains(r.Document.Title, "Mental Health") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected mental health tax document in results, got: %v", resultIDs(results))
	}
}

func TestSearchJurisdictionFilter(t *testing.T) {
	store := SeedStore()

	// Search for "standard deduction" in CA only — should not return federal IRC §63
	caResults := store.Search("standard deduction", JurisdictionCA, 10)
	for _, r := range caResults {
		if r.Document.Jurisdiction != JurisdictionCA {
			t.Errorf("CA search returned non-CA document: %s (jurisdiction: %s)", r.Document.ID, r.Document.Jurisdiction)
		}
	}

	// Search for "mental health" in federal only — should return nothing
	fedResults := store.Search("mental health surcharge", JurisdictionFederal, 10)
	for _, r := range fedResults {
		if r.Document.Jurisdiction != JurisdictionFederal {
			t.Errorf("federal search returned non-federal document: %s (jurisdiction: %s)", r.Document.ID, r.Document.Jurisdiction)
		}
	}
}

func TestSeedDocuments(t *testing.T) {
	fedDocs := SeedFederalDocuments()
	caDocs := SeedCADocuments()

	if len(fedDocs) < 25 {
		t.Errorf("expected at least 25 federal seed documents, got %d", len(fedDocs))
	}
	if len(caDocs) < 15 {
		t.Errorf("expected at least 15 CA seed documents, got %d", len(caDocs))
	}

	store := SeedStore()
	if store.Count() != len(fedDocs)+len(caDocs) {
		t.Errorf("store count = %d, want %d", store.Count(), len(fedDocs)+len(caDocs))
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()

	// Create and save
	store := NewStore()
	store.Add(Document{
		ID: "test1", Title: "Test Document One",
		Content: "This is test content about income tax.",
		Source:  "Test §1", Jurisdiction: JurisdictionFederal,
		DocType: DocTypeIRCSection, Section: "1",
		Tags: []string{"test", "income"},
	})
	store.Add(Document{
		ID: "test2", Title: "Test CA Document",
		Content: "California conformity rules.",
		Source:  "CA R&TC §100", Jurisdiction: JurisdictionCA,
		DocType: DocTypeCARTCSection, Section: "100",
		Tags: []string{"conformity", "california"},
	})

	if err := store.Save(dir); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	// Verify files exist
	fedFile := filepath.Join(dir, "federal.json")
	caFile := filepath.Join(dir, "ca.json")
	if _, err := os.Stat(fedFile); os.IsNotExist(err) {
		t.Error("federal.json not created")
	}
	if _, err := os.Stat(caFile); os.IsNotExist(err) {
		t.Error("ca.json not created")
	}

	// Load into new store
	loaded, err := NewStoreFromDir(dir)
	if err != nil {
		t.Fatalf("NewStoreFromDir error: %v", err)
	}
	if loaded.Count() != 2 {
		t.Errorf("loaded store count = %d, want 2", loaded.Count())
	}

	// Verify search works on loaded store
	results := loaded.Search("income tax", JurisdictionFederal, 5)
	if len(results) == 0 {
		t.Error("expected search results from loaded store")
	}
}

func TestFormatContext(t *testing.T) {
	results := []SearchResult{
		{
			Document: Document{
				Title:   "Test Title",
				Source:  "IRC §100",
				Content: "This is the content.",
			},
			Score: 5.0,
		},
	}

	formatted := FormatContext(results)
	if !strings.Contains(formatted, "Test Title") {
		t.Error("formatted context should contain document title")
	}
	if !strings.Contains(formatted, "IRC §100") {
		t.Error("formatted context should contain source")
	}
	if !strings.Contains(formatted, "This is the content.") {
		t.Error("formatted context should contain content")
	}
	if !strings.Contains(formatted, "Reference 1") {
		t.Error("formatted context should contain reference number")
	}

	// Empty results
	empty := FormatContext(nil)
	if empty != "" {
		t.Errorf("FormatContext(nil) = %q, want empty string", empty)
	}
}

func TestRAGQueryForField(t *testing.T) {
	// This test verifies the RAG wiring without actually calling an LLM.
	// We verify that NewRAG works and that the store search is used.
	store := SeedStore()

	// We can't create a real client without an API key, so test the store
	// search portion that feeds into RAG.
	results := store.Search("wages compensation W-2", JurisdictionFederal, 3)
	if len(results) == 0 {
		t.Fatal("expected results for wages/W-2 search")
	}

	// Verify the context formatting works
	ctx := FormatContext(results)
	if ctx == "" {
		t.Error("expected non-empty context from search results")
	}

	// Test that NewRAG doesn't panic with nil client
	// (it should still construct, just fail on actual queries)
	rag := NewRAG(store, nil)
	if rag == nil {
		t.Error("NewRAG should not return nil")
	}

	// If API key is available, test full RAG query
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		t.Skip("OPENROUTER_API_KEY not set, skipping live RAG test")
	}

	client, err := llm.NewClient(apiKey)
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	rag = NewRAG(store, client)
	answer, err := rag.QueryForField(context.Background(), "1040:wages", "Wages, salaries, tips", JurisdictionFederal)
	if err != nil {
		t.Fatalf("QueryForField error: %v", err)
	}
	if answer == "" {
		t.Error("expected non-empty answer from RAG query")
	}
}

// resultIDs returns the IDs from search results for diagnostic output.
func resultIDs(results []SearchResult) []string {
	ids := make([]string, len(results))
	for i, r := range results {
		ids[i] = r.Document.ID
	}
	return ids
}
