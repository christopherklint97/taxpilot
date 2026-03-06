package knowledge

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

// Store holds the knowledge base documents and provides search.
type Store struct {
	documents []Document
	index     map[string][]int // keyword -> document indices (inverted index)
	dataDir   string
}

// NewStore creates a new empty Store.
func NewStore() *Store {
	return &Store{
		documents: nil,
		index:     make(map[string][]int),
	}
}

// NewStoreFromDir loads documents from JSON files in the given directory.
func NewStoreFromDir(dataDir string) (*Store, error) {
	s := &Store{
		index:   make(map[string][]int),
		dataDir: dataDir,
	}
	if err := s.Load(dataDir); err != nil {
		return nil, err
	}
	return s, nil
}

// Add adds a document to the store and updates the index.
func (s *Store) Add(doc Document) {
	idx := len(s.documents)
	s.documents = append(s.documents, doc)
	// Index all tokens from the document
	tokens := docTokens(doc)
	seen := make(map[string]bool)
	for _, tok := range tokens {
		if !seen[tok] {
			s.index[tok] = append(s.index[tok], idx)
			seen[tok] = true
		}
	}
}

// Search finds documents matching the query, scoped by jurisdiction.
// Returns up to maxResults results sorted by relevance score.
func (s *Store) Search(query string, jurisdiction Jurisdiction, maxResults int) []SearchResult {
	queryTokens := tokenize(query)
	if len(queryTokens) == 0 {
		return nil
	}

	// Collect candidate document indices that share at least one token
	candidates := make(map[int]bool)
	for _, tok := range queryTokens {
		for _, idx := range s.index[tok] {
			candidates[idx] = true
		}
	}

	var results []SearchResult
	for idx := range candidates {
		doc := s.documents[idx]
		// Filter by jurisdiction
		if jurisdiction != "" && doc.Jurisdiction != jurisdiction {
			continue
		}
		sc := score(doc, queryTokens, len(s.documents), s.index)
		if sc > 0 {
			results = append(results, SearchResult{Document: doc, Score: sc})
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

// SearchAll searches across all jurisdictions.
func (s *Store) SearchAll(query string, maxResults int) []SearchResult {
	return s.Search(query, "", maxResults)
}

// Save persists the store to disk as JSON files, one per jurisdiction.
func (s *Store) Save(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	// Group by jurisdiction
	byJurisdiction := make(map[Jurisdiction][]Document)
	for _, doc := range s.documents {
		byJurisdiction[doc.Jurisdiction] = append(byJurisdiction[doc.Jurisdiction], doc)
	}

	for jur, docs := range byJurisdiction {
		data, err := json.MarshalIndent(docs, "", "  ")
		if err != nil {
			return err
		}
		path := filepath.Join(dir, string(jur)+".json")
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// Load reads documents from a directory of JSON files.
func (s *Store) Load(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var docs []Document
		if err := json.Unmarshal(data, &docs); err != nil {
			return err
		}
		for _, doc := range docs {
			s.Add(doc)
		}
	}
	return nil
}

// Count returns the number of documents in the store.
func (s *Store) Count() int {
	return len(s.documents)
}

// buildIndex rebuilds the inverted index from all documents.
func (s *Store) buildIndex() {
	s.index = make(map[string][]int)
	for idx, doc := range s.documents {
		tokens := docTokens(doc)
		seen := make(map[string]bool)
		for _, tok := range tokens {
			if !seen[tok] {
				s.index[tok] = append(s.index[tok], idx)
				seen[tok] = true
			}
		}
	}
}

// docTokens returns all tokens from a document's searchable fields.
func docTokens(doc Document) []string {
	var all []string
	all = append(all, tokenize(doc.Title)...)
	all = append(all, tokenize(doc.Content)...)
	all = append(all, tokenize(doc.Source)...)
	for _, tag := range doc.Tags {
		all = append(all, tokenize(tag)...)
	}
	return all
}

// tokenize splits text into lowercase search tokens, filtering out stop words and short tokens.
func tokenize(text string) []string {
	text = strings.ToLower(text)
	// Split on non-alphanumeric characters
	fields := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	var tokens []string
	for _, f := range fields {
		if len(f) < 2 {
			continue
		}
		if stopWords[f] {
			continue
		}
		tokens = append(tokens, f)
	}
	return tokens
}

// score computes a TF-IDF-like relevance score for a document against query tokens.
// Weights: Title 3x, Source 2x, Tags 2x, Content 1x.
// Normalized by document content length to avoid biasing toward long documents.
func score(doc Document, queryTokens []string, totalDocs int, index map[string][]int) float64 {
	if len(queryTokens) == 0 {
		return 0
	}

	titleTokens := tokenSet(tokenize(doc.Title))
	sourceTokens := tokenSet(tokenize(doc.Source))
	contentTokens := tokenize(doc.Content)
	contentSet := tokenSet(contentTokens)

	var tagTokens map[string]bool
	{
		var allTagTokens []string
		for _, tag := range doc.Tags {
			allTagTokens = append(allTagTokens, tokenize(tag)...)
		}
		tagTokens = tokenSet(allTagTokens)
	}

	var totalScore float64
	for _, qt := range queryTokens {
		var fieldScore float64

		// Title match (3x weight)
		if titleTokens[qt] {
			fieldScore += 3.0
		}
		// Source match (2x weight)
		if sourceTokens[qt] {
			fieldScore += 2.0
		}
		// Tag match (2x weight)
		if tagTokens[qt] {
			fieldScore += 2.0
		}
		// Content match (1x weight)
		if contentSet[qt] {
			fieldScore += 1.0
		}

		// Apply IDF weighting: rarer terms across the corpus get higher scores
		df := len(index[qt])
		if df > 0 && totalDocs > 0 {
			idf := math.Log(float64(totalDocs)/float64(df)) + 1.0
			fieldScore *= idf
		}

		totalScore += fieldScore
	}

	// Normalize by document length to avoid biasing toward long documents.
	// Use log to dampen the effect of very long documents.
	docLen := float64(len(contentTokens))
	if docLen < 1 {
		docLen = 1
	}
	totalScore /= math.Log(docLen + 1)

	return totalScore
}

// tokenSet converts a slice of tokens to a set (map) for O(1) lookup.
func tokenSet(tokens []string) map[string]bool {
	m := make(map[string]bool, len(tokens))
	for _, t := range tokens {
		m[t] = true
	}
	return m
}

// stopWords are common English words excluded from indexing and search.
var stopWords = map[string]bool{
	"the": true, "is": true, "at": true, "which": true, "on": true,
	"an": true, "and": true, "or": true, "of": true, "to": true,
	"in": true, "for": true, "with": true, "by": true, "from": true,
	"as": true, "it": true, "that": true, "this": true, "be": true,
	"are": true, "was": true, "were": true, "been": true, "have": true,
	"has": true, "had": true, "do": true, "does": true, "did": true,
	"but": true, "not": true, "if": true, "its": true, "may": true,
	"can": true, "will": true, "so": true, "no": true, "than": true,
	"other": true, "also": true, "into": true, "any": true, "all": true,
	"each": true, "such": true, "who": true, "they": true, "their": true,
	"there": true, "would": true, "should": true, "could": true,
}
