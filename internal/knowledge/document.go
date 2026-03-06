package knowledge

// Jurisdiction tags a document as federal or state.
type Jurisdiction string

const (
	JurisdictionFederal Jurisdiction = "federal"
	JurisdictionCA      Jurisdiction = "ca"
)

// DocumentType categorizes knowledge base documents.
type DocumentType string

const (
	DocTypeIRCSection     DocumentType = "irc_section"
	DocTypeIRSPublication DocumentType = "irs_publication"
	DocTypeIRSInstruction DocumentType = "irs_instruction"
	DocTypeCARTCSection   DocumentType = "ca_rtc_section"
	DocTypeFTBPublication DocumentType = "ftb_publication"
	DocTypeFTBInstruction DocumentType = "ftb_instruction"
)

// Document represents a single chunk of tax knowledge.
type Document struct {
	ID           string       `json:"id"`
	Title        string       `json:"title"`
	Content      string       `json:"content"`
	Source       string       `json:"source"`       // e.g., "IRC §61", "Pub 17, Ch. 5"
	Jurisdiction Jurisdiction `json:"jurisdiction"`
	DocType      DocumentType `json:"doc_type"`
	Section      string       `json:"section"`      // e.g., "61", "162"
	Tags         []string     `json:"tags"`         // searchable keywords
	TaxYear      int          `json:"tax_year"`
}

// SearchResult holds a document and its relevance score.
type SearchResult struct {
	Document Document
	Score    float64 // higher is more relevant
}
