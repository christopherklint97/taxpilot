package interview

import (
	"testing"

	"taxpilot/internal/efile/ca"
	"taxpilot/internal/efile/mef"
	"taxpilot/internal/forms"
	"taxpilot/internal/pdf"
)

// TestPDFMappingsCoverAllComputedForms verifies that every non-input form
// has PDF field mappings registered. Input forms (W-2, 1099s) are excluded
// because they are source documents, not filled by TaxPilot.
func TestPDFMappingsCoverAllComputedForms(t *testing.T) {
	pdfMapped := pdf.PDFMappedFormIDs()
	inputForms := make(map[forms.FormID]bool)
	for _, id := range forms.InputFormIDs() {
		inputForms[id] = true
	}

	// Forms that don't have PDF export yet (expat forms)
	// These are known gaps that should be addressed.
	knownGaps := map[forms.FormID]bool{
		forms.FormF2555: true,
		forms.FormF1116: true,
		forms.FormF8938: true,
		forms.FormF8833: true,
	}

	for _, id := range forms.AllFormIDs() {
		if inputForms[id] {
			continue // input forms don't need PDF mappings
		}
		if knownGaps[id] {
			continue // tracked as known gaps
		}
		if !pdfMapped[id] {
			t.Errorf("form %s has no PDF mapping registered — add it to pdf/mappings.go and pdf/registry.go", id)
		}
	}
}

// TestMeFXMLCoversAllFederalForms verifies that every federal form has a
// MeF XML builder.
func TestMeFXMLCoversAllFederalForms(t *testing.T) {
	mefCovered := mef.MeFCoveredFormIDs()
	reg := SetupRegistry()
	inputForms := make(map[forms.FormID]bool)
	for _, id := range forms.InputFormIDs() {
		inputForms[id] = true
	}

	for _, id := range forms.AllFormIDs() {
		if inputForms[id] && id != forms.FormW2 {
			continue // 1099s are source documents, not submitted via MeF (W-2 is)
		}
		formDef, ok := reg.Get(id)
		if !ok {
			continue
		}
		if formDef.Jurisdiction != forms.Federal {
			continue // MeF only covers federal forms
		}
		if !mefCovered[id] {
			t.Errorf("federal form %s has no MeF XML builder — add it to mef/xml.go and mef/registry.go", id)
		}
	}
}

// TestCAXMLCoversAllCAForms verifies that every CA form has a CA FTB XML builder.
func TestCAXMLCoversAllCAForms(t *testing.T) {
	caCovered := ca.CACoveredFormIDs()
	reg := SetupRegistry()

	// Forms that are CA-specific but don't have their own XML section
	// (they flow through Form 540 totals).
	flowThroughForms := map[forms.FormID]bool{
		forms.FormF3514: true, // CalEITC flows into Form 540
		forms.FormF3853: true, // Health coverage flows into Form 540
	}

	for _, id := range forms.AllFormIDs() {
		formDef, ok := reg.Get(id)
		if !ok {
			continue
		}
		if formDef.Jurisdiction != forms.StateCA {
			continue
		}
		if flowThroughForms[id] {
			continue
		}
		if !caCovered[id] {
			t.Errorf("CA form %s has no CA FTB XML builder — add it to ca/xml.go and ca/registry.go", id)
		}
	}
}
