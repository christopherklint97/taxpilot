package knowledge

import (
	"testing"
)

func TestSeedStoreCount(t *testing.T) {
	store := SeedStore()
	fedCount := len(SeedFederalDocuments())
	caCount := len(SeedCADocuments())
	ircCount := len(SeedIRCSections())
	pubCount := len(SeedIRSPublications())
	ftbCount := len(SeedFTBPublications())
	expatCount := len(SeedExpatDocuments())
	expected := fedCount + caCount + ircCount + pubCount + ftbCount + expatCount

	if store.Count() != expected {
		t.Errorf("SeedStore() count = %d, want %d (fed=%d, ca=%d, irc=%d, pub=%d, ftb=%d, expat=%d)",
			store.Count(), expected, fedCount, caCount, ircCount, pubCount, ftbCount, expatCount)
	}
	// Verify minimum total (original 94 + 11 expat = 105)
	if store.Count() < 105 {
		t.Errorf("SeedStore() count = %d, want at least 105", store.Count())
	}
}

func TestAllDocumentsHaveRequiredFields(t *testing.T) {
	allDocs := collectAllDocuments()
	for _, doc := range allDocs {
		if doc.ID == "" {
			t.Error("found document with empty ID")
		}
		if doc.Title == "" {
			t.Errorf("document %q has empty Title", doc.ID)
		}
		if doc.Content == "" {
			t.Errorf("document %q has empty Content", doc.ID)
		}
		if doc.Source == "" {
			t.Errorf("document %q has empty Source", doc.ID)
		}
		if doc.Jurisdiction == "" {
			t.Errorf("document %q has empty Jurisdiction", doc.ID)
		}
		if doc.DocType == "" {
			t.Errorf("document %q has empty DocType", doc.ID)
		}
		if len(doc.Tags) == 0 {
			t.Errorf("document %q has no Tags", doc.ID)
		}
	}
}

func TestAllDocumentsUniqueIDs(t *testing.T) {
	allDocs := collectAllDocuments()
	seen := make(map[string]bool)
	for _, doc := range allDocs {
		if seen[doc.ID] {
			t.Errorf("duplicate document ID: %q", doc.ID)
		}
		seen[doc.ID] = true
	}
}

func TestIRCSectionsHaveProperDocType(t *testing.T) {
	for _, doc := range SeedIRCSections() {
		if doc.DocType != DocTypeIRCSection {
			t.Errorf("IRC section %q has DocType %q, want %q", doc.ID, doc.DocType, DocTypeIRCSection)
		}
		if doc.Jurisdiction != JurisdictionFederal {
			t.Errorf("IRC section %q has Jurisdiction %q, want %q", doc.ID, doc.Jurisdiction, JurisdictionFederal)
		}
		if doc.Section == "" {
			t.Errorf("IRC section %q has empty Section", doc.ID)
		}
	}
}

func TestIRSPublicationsHaveProperDocType(t *testing.T) {
	for _, doc := range SeedIRSPublications() {
		if doc.DocType != DocTypeIRSPublication {
			t.Errorf("IRS publication %q has DocType %q, want %q", doc.ID, doc.DocType, DocTypeIRSPublication)
		}
		if doc.Jurisdiction != JurisdictionFederal {
			t.Errorf("IRS publication %q has Jurisdiction %q, want %q", doc.ID, doc.Jurisdiction, JurisdictionFederal)
		}
	}
}

func TestFTBPublicationsHaveProperDocType(t *testing.T) {
	for _, doc := range SeedFTBPublications() {
		if doc.DocType != DocTypeFTBPublication {
			t.Errorf("FTB publication %q has DocType %q, want %q", doc.ID, doc.DocType, DocTypeFTBPublication)
		}
		if doc.Jurisdiction != JurisdictionCA {
			t.Errorf("FTB publication %q has Jurisdiction %q, want %q", doc.ID, doc.Jurisdiction, JurisdictionCA)
		}
	}
}

func TestNewDocumentsAreSearchable(t *testing.T) {
	store := SeedStore()

	tests := []struct {
		query        string
		jurisdiction Jurisdiction
		expectID     string
		desc         string
	}{
		{
			query:        "home sale exclusion principal residence",
			jurisdiction: JurisdictionFederal,
			expectID:     "irc_121",
			desc:         "IRC 121 home sale exclusion",
		},
		{
			query:        "passive activity rental loss",
			jurisdiction: JurisdictionFederal,
			expectID:     "irc_469",
			desc:         "IRC 469 passive activity losses",
		},
		{
			query:        "self-employed business expenses Schedule C",
			jurisdiction: JurisdictionFederal,
			expectID:     "pub334_business_expenses",
			desc:         "Pub 334 business expenses",
		},
		{
			query:        "estimated tax quarterly underpayment penalty",
			jurisdiction: JurisdictionFederal,
			expectID:     "pub505_underpayment",
			desc:         "Pub 505 underpayment penalty",
		},
		{
			query:        "California resident domicile worldwide income",
			jurisdiction: JurisdictionCA,
			expectID:     "ftb1031_resident",
			desc:         "FTB 1031 resident definition",
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			results := store.Search(tt.query, tt.jurisdiction, 10)
			if len(results) == 0 {
				t.Fatalf("no results for query %q", tt.query)
			}
			found := false
			for _, r := range results {
				if r.Document.ID == tt.expectID {
					found = true
					break
				}
			}
			if !found {
				ids := make([]string, len(results))
				for i, r := range results {
					ids[i] = r.Document.ID
				}
				t.Errorf("expected %q in results for %q, got: %v", tt.expectID, tt.query, ids)
			}
		})
	}
}

func TestExpatDocumentsSearchable(t *testing.T) {
	store := SeedStore()

	tests := []struct {
		query    string
		expectID string
		desc     string
	}{
		{"foreign earned income exclusion expat abroad", "irc_911", "FEIE search"},
		{"foreign tax credit taxes paid", "irc_901", "FTC search"},
		{"FBAR foreign bank accounts FinCEN", "fbar_fincen", "FBAR search"},
		{"California FEIE non-conformity add back", "ca_feie_nonconformity", "CA FEIE search"},
		{"Sweden tax treaty pension", "us_sweden_treaty", "Sweden treaty search"},
		{"FATCA form 8938 foreign assets", "irc_6038d", "FATCA search"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			results := store.Search(tt.query, "", 10)
			if len(results) == 0 {
				t.Fatalf("no results for query %q", tt.query)
			}
			found := false
			for _, r := range results {
				if r.Document.ID == tt.expectID {
					found = true
					break
				}
			}
			if !found {
				ids := make([]string, len(results))
				for i, r := range results {
					ids[i] = r.Document.ID
				}
				t.Errorf("expected %q in results for %q, got: %v", tt.expectID, tt.query, ids)
			}
		})
	}
}

func TestExpatDocumentCount(t *testing.T) {
	docs := SeedExpatDocuments()
	if len(docs) < 11 {
		t.Errorf("SeedExpatDocuments() returned %d documents, want at least 11", len(docs))
	}
}

func TestIRCSectionCount(t *testing.T) {
	docs := SeedIRCSections()
	if len(docs) < 18 {
		t.Errorf("SeedIRCSections() returned %d documents, want at least 18", len(docs))
	}
}

func TestIRSPublicationCount(t *testing.T) {
	docs := SeedIRSPublications()
	if len(docs) < 21 {
		t.Errorf("SeedIRSPublications() returned %d documents, want at least 21", len(docs))
	}
}

func TestFTBPublicationCount(t *testing.T) {
	docs := SeedFTBPublications()
	if len(docs) < 14 {
		t.Errorf("SeedFTBPublications() returned %d documents, want at least 14", len(docs))
	}
}

// collectAllDocuments gathers documents from all seed functions.
func collectAllDocuments() []Document {
	var all []Document
	all = append(all, SeedFederalDocuments()...)
	all = append(all, SeedCADocuments()...)
	all = append(all, SeedIRCSections()...)
	all = append(all, SeedIRSPublications()...)
	all = append(all, SeedFTBPublications()...)
	all = append(all, SeedExpatDocuments()...)
	return all
}
