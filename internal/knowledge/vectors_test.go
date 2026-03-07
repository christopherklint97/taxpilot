package knowledge

import (
	"math"
	"testing"
)

func TestVectorStoreBuild(t *testing.T) {
	store := SeedStore()
	vs := NewVectorStore(store)
	vs.Build()

	// Should have vectors for all documents.
	if len(vs.vectors) != store.Count() {
		t.Errorf("expected %d vectors, got %d", store.Count(), len(vs.vectors))
	}

	// Vocabulary should be non-empty.
	if len(vs.vocabulary) == 0 {
		t.Error("expected non-empty vocabulary")
	}

	// vocabIndex should match vocabulary length.
	if len(vs.vocabIndex) != len(vs.vocabulary) {
		t.Errorf("vocabIndex size %d != vocabulary size %d", len(vs.vocabIndex), len(vs.vocabulary))
	}

	// IDF should have same length as vocabulary.
	if len(vs.idf) != len(vs.vocabulary) {
		t.Errorf("IDF size %d != vocabulary size %d", len(vs.idf), len(vs.vocabulary))
	}

	// Each vector should have same length as vocabulary.
	for idx, vec := range vs.vectors {
		if len(vec) != len(vs.vocabulary) {
			t.Errorf("vector %d length %d != vocabulary size %d", idx, len(vec), len(vs.vocabulary))
		}
	}

	// Vectors should be normalized (unit length).
	for idx, vec := range vs.vectors {
		var mag float64
		for _, v := range vec {
			mag += v * v
		}
		mag = math.Sqrt(mag)
		if mag > 0 && math.Abs(mag-1.0) > 1e-9 {
			t.Errorf("vector %d not normalized: magnitude = %f", idx, mag)
		}
	}
}

func TestSemanticSearch(t *testing.T) {
	store := SeedStore()
	vs := NewVectorStore(store)
	vs.Build()

	results := vs.SemanticSearch("HSA health savings", "", 5)
	if len(results) == 0 {
		t.Fatal("expected results for 'HSA health savings'")
	}

	// The top result should be HSA-related.
	found := false
	for _, r := range results {
		if r.Document.ID == "irc_223" || r.Document.ID == "ca_hsa" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected HSA-related document in results, got: %v", resultIDs(results))
	}

	// Scores should be positive and sorted descending.
	for i := 1; i < len(results); i++ {
		if results[i].Score > results[i-1].Score {
			t.Errorf("results not sorted: score[%d]=%f > score[%d]=%f", i, results[i].Score, i-1, results[i-1].Score)
		}
	}
}

func TestSemanticSearchJurisdiction(t *testing.T) {
	store := SeedStore()
	vs := NewVectorStore(store)
	vs.Build()

	// Search CA only.
	caResults := vs.SemanticSearch("standard deduction", JurisdictionCA, 10)
	for _, r := range caResults {
		if r.Document.Jurisdiction != JurisdictionCA {
			t.Errorf("CA search returned non-CA document: %s (jurisdiction: %s)", r.Document.ID, r.Document.Jurisdiction)
		}
	}

	// Search federal only.
	fedResults := vs.SemanticSearch("mental health surcharge", JurisdictionFederal, 10)
	for _, r := range fedResults {
		if r.Document.Jurisdiction != JurisdictionFederal {
			t.Errorf("federal search returned non-federal document: %s (jurisdiction: %s)", r.Document.ID, r.Document.Jurisdiction)
		}
	}
}

func TestHybridSearch(t *testing.T) {
	store := SeedStore()
	vs := NewVectorStore(store)
	vs.Build()

	results := vs.HybridSearch("capital gains investment", "", 5, 0.5)
	if len(results) == 0 {
		t.Fatal("expected results for hybrid search 'capital gains investment'")
	}

	// Should find capital gains document.
	found := false
	for _, r := range results {
		if r.Document.ID == "irc_1_capital_gains" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected capital gains document in hybrid results, got: %v", resultIDs(results))
	}
}

func TestHybridSearchAlpha(t *testing.T) {
	store := SeedStore()
	vs := NewVectorStore(store)
	vs.Build()

	query := "standard deduction"
	jurisdiction := Jurisdiction("")
	maxResults := 10

	// alpha=0 should give same ranking as keyword search.
	keywordOnly := vs.HybridSearch(query, jurisdiction, maxResults, 0.0)
	pureKeyword := store.Search(query, jurisdiction, maxResults)

	if len(keywordOnly) == 0 {
		t.Fatal("expected results for alpha=0 hybrid search")
	}

	// Top result from alpha=0 should match top keyword result.
	if len(pureKeyword) > 0 && len(keywordOnly) > 0 {
		if keywordOnly[0].Document.ID != pureKeyword[0].Document.ID {
			t.Errorf("alpha=0 top result %s != keyword top result %s",
				keywordOnly[0].Document.ID, pureKeyword[0].Document.ID)
		}
	}

	// alpha=1 should give same ranking as vector search.
	vectorOnly := vs.HybridSearch(query, jurisdiction, maxResults, 1.0)
	pureVector := vs.SemanticSearch(query, jurisdiction, maxResults)

	if len(vectorOnly) == 0 {
		t.Fatal("expected results for alpha=1 hybrid search")
	}

	// Top result from alpha=1 should match top vector result.
	if len(pureVector) > 0 && len(vectorOnly) > 0 {
		if vectorOnly[0].Document.ID != pureVector[0].Document.ID {
			t.Errorf("alpha=1 top result %s != vector top result %s",
				vectorOnly[0].Document.ID, pureVector[0].Document.ID)
		}
	}
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name string
		a, b []float64
		want float64
	}{
		{
			name: "identical vectors",
			a:    []float64{1, 2, 3},
			b:    []float64{1, 2, 3},
			want: 1.0,
		},
		{
			name: "orthogonal vectors",
			a:    []float64{1, 0, 0},
			b:    []float64{0, 1, 0},
			want: 0.0,
		},
		{
			name: "opposite vectors",
			a:    []float64{1, 0},
			b:    []float64{-1, 0},
			want: -1.0,
		},
		{
			name: "zero vector a",
			a:    []float64{0, 0, 0},
			b:    []float64{1, 2, 3},
			want: 0.0,
		},
		{
			name: "zero vector b",
			a:    []float64{1, 2, 3},
			b:    []float64{0, 0, 0},
			want: 0.0,
		},
		{
			name: "known angle",
			a:    []float64{1, 0},
			b:    []float64{1, 1},
			want: 1.0 / math.Sqrt(2),
		},
		{
			name: "different lengths",
			a:    []float64{1, 2},
			b:    []float64{1, 2, 3},
			want: 0.0, // mismatched lengths return 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cosineSimilarity(tt.a, tt.b)
			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("cosineSimilarity(%v, %v) = %f, want %f", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestEmptyQuery(t *testing.T) {
	store := SeedStore()
	vs := NewVectorStore(store)
	vs.Build()

	// Semantic search with empty query.
	results := vs.SemanticSearch("", "", 5)
	if len(results) != 0 {
		t.Errorf("expected no results for empty semantic search, got %d", len(results))
	}

	// Hybrid search with empty query.
	hybrid := vs.HybridSearch("", "", 5, 0.5)
	if len(hybrid) != 0 {
		t.Errorf("expected no results for empty hybrid search, got %d", len(hybrid))
	}
}
