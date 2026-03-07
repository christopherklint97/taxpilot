package knowledge

import (
	"math"
	"sort"
)

// VectorStore adds vector similarity search on top of the keyword store.
type VectorStore struct {
	store      *Store
	vectors    map[int][]float64 // document index -> TF-IDF vector
	vocabulary []string          // ordered list of all terms
	vocabIndex map[string]int    // term -> index in vocabulary
	idf        []float64         // IDF weight per vocabulary term
}

// NewVectorStore creates a VectorStore from an existing Store.
func NewVectorStore(store *Store) *VectorStore {
	return &VectorStore{
		store:      store,
		vectors:    make(map[int][]float64),
		vocabIndex: make(map[string]int),
	}
}

// Build computes TF-IDF vectors for all documents.
func (vs *VectorStore) Build() {
	// Step 1: Collect all unique terms across all documents.
	vocabSet := make(map[string]bool)
	docTokensList := make([][]string, len(vs.store.documents))
	for i, doc := range vs.store.documents {
		tokens := docTokens(doc)
		docTokensList[i] = tokens
		for _, tok := range tokens {
			vocabSet[tok] = true
		}
	}

	// Build ordered vocabulary.
	vs.vocabulary = make([]string, 0, len(vocabSet))
	for term := range vocabSet {
		vs.vocabulary = append(vs.vocabulary, term)
	}
	sort.Strings(vs.vocabulary)

	vs.vocabIndex = make(map[string]int, len(vs.vocabulary))
	for i, term := range vs.vocabulary {
		vs.vocabIndex[term] = i
	}

	// Step 2: Compute IDF for each term.
	n := float64(len(vs.store.documents))
	vs.idf = make([]float64, len(vs.vocabulary))

	// Count document frequency for each term.
	df := make([]int, len(vs.vocabulary))
	for _, tokens := range docTokensList {
		seen := make(map[string]bool)
		for _, tok := range tokens {
			if !seen[tok] {
				if idx, ok := vs.vocabIndex[tok]; ok {
					df[idx]++
				}
				seen[tok] = true
			}
		}
	}

	for i, d := range df {
		if d > 0 {
			vs.idf[i] = math.Log(n / float64(d))
		}
	}

	// Step 3: Build TF-IDF vector for each document and normalize.
	vs.vectors = make(map[int][]float64, len(vs.store.documents))
	for i, tokens := range docTokensList {
		vec := vs.tfidfVector(tokens)
		normalize(vec)
		vs.vectors[i] = vec
	}
}

// SemanticSearch finds documents by cosine similarity of TF-IDF vectors.
func (vs *VectorStore) SemanticSearch(query string, jurisdiction Jurisdiction, maxResults int) []SearchResult {
	queryTokens := tokenize(query)
	if len(queryTokens) == 0 {
		return nil
	}

	queryVec := vs.tfidfVector(queryTokens)
	normalize(queryVec)

	var results []SearchResult
	for idx, docVec := range vs.vectors {
		doc := vs.store.documents[idx]
		if jurisdiction != "" && doc.Jurisdiction != jurisdiction {
			continue
		}
		sim := cosineSimilarity(queryVec, docVec)
		if sim > 0 {
			results = append(results, SearchResult{Document: doc, Score: sim})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if maxResults > 0 && len(results) > maxResults {
		results = results[:maxResults]
	}
	return results
}

// HybridSearch combines keyword (BM25-style) and vector search results.
// alpha controls the blend: 0.0 = all keyword, 1.0 = all vector.
func (vs *VectorStore) HybridSearch(query string, jurisdiction Jurisdiction, maxResults int, alpha float64) []SearchResult {
	// Clamp alpha.
	if alpha < 0 {
		alpha = 0
	}
	if alpha > 1 {
		alpha = 1
	}

	// Run both searches with a generous limit to get good coverage for merging.
	fetchLimit := maxResults * 3
	if fetchLimit < 20 {
		fetchLimit = 20
	}

	keywordResults := vs.store.Search(query, jurisdiction, fetchLimit)
	vectorResults := vs.SemanticSearch(query, jurisdiction, fetchLimit)

	if len(keywordResults) == 0 && len(vectorResults) == 0 {
		return nil
	}

	// Normalize scores to [0, 1].
	keywordMax := maxScore(keywordResults)
	vectorMax := maxScore(vectorResults)

	// Build combined score map keyed by document ID.
	type combinedEntry struct {
		doc          Document
		keywordScore float64
		vectorScore  float64
	}
	merged := make(map[string]*combinedEntry)

	for _, r := range keywordResults {
		normalized := 0.0
		if keywordMax > 0 {
			normalized = r.Score / keywordMax
		}
		if e, ok := merged[r.Document.ID]; ok {
			e.keywordScore = normalized
		} else {
			merged[r.Document.ID] = &combinedEntry{
				doc:          r.Document,
				keywordScore: normalized,
			}
		}
	}

	for _, r := range vectorResults {
		normalized := 0.0
		if vectorMax > 0 {
			normalized = r.Score / vectorMax
		}
		if e, ok := merged[r.Document.ID]; ok {
			e.vectorScore = normalized
		} else {
			merged[r.Document.ID] = &combinedEntry{
				doc:         r.Document,
				vectorScore: normalized,
			}
		}
	}

	// Compute final blended scores.
	var results []SearchResult
	for _, e := range merged {
		finalScore := (1-alpha)*e.keywordScore + alpha*e.vectorScore
		if finalScore > 0 {
			results = append(results, SearchResult{Document: e.doc, Score: finalScore})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if maxResults > 0 && len(results) > maxResults {
		results = results[:maxResults]
	}
	return results
}

// tfidfVector builds a TF-IDF vector for a set of tokens using the store's vocabulary and IDF weights.
func (vs *VectorStore) tfidfVector(tokens []string) []float64 {
	vec := make([]float64, len(vs.vocabulary))

	// Count term frequencies.
	tf := make(map[string]int)
	for _, tok := range tokens {
		tf[tok]++
	}

	for term, count := range tf {
		if idx, ok := vs.vocabIndex[term]; ok {
			vec[idx] = float64(count) * vs.idf[idx]
		}
	}

	return vec
}

// cosineSimilarity computes the cosine similarity between two vectors.
// Returns 0.0 if either vector has zero magnitude.
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dot, magA, magB float64
	for i := range a {
		dot += a[i] * b[i]
		magA += a[i] * a[i]
		magB += b[i] * b[i]
	}

	if magA == 0 || magB == 0 {
		return 0
	}

	return dot / (math.Sqrt(magA) * math.Sqrt(magB))
}

// normalize scales a vector to unit length in place.
func normalize(vec []float64) {
	var mag float64
	for _, v := range vec {
		mag += v * v
	}
	if mag == 0 {
		return
	}
	mag = math.Sqrt(mag)
	for i := range vec {
		vec[i] /= mag
	}
}

// maxScore returns the maximum score from a list of search results.
func maxScore(results []SearchResult) float64 {
	m := 0.0
	for _, r := range results {
		if r.Score > m {
			m = r.Score
		}
	}
	return m
}
