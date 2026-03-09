package forms

import "fmt"

// FormRegistration holds all known metadata about a single form.
type FormRegistration struct {
	ID          FormID
	Def         *FormDef // nil if not registered as a form definition
	HasPDF      bool     // has PDF field mappings
	HasMeFXML   bool     // has MeF XML builder (federal forms)
	HasCAXML    bool     // has CA FTB XML builder (CA forms)
}

// FormRegistry is a centralized registry that tracks all metadata per form.
// It consolidates information that was previously spread across separate
// registries (forms, PDF, MeF XML, CA XML) into a single queryable source.
type FormRegistry struct {
	registrations map[FormID]*FormRegistration
}

// NewFormRegistry creates a FormRegistry pre-populated with entries for every
// known FormID. Call the Set* methods to fill in metadata from each subsystem.
func NewFormRegistry() *FormRegistry {
	fr := &FormRegistry{
		registrations: make(map[FormID]*FormRegistration),
	}
	for _, id := range AllFormIDs() {
		fr.registrations[id] = &FormRegistration{ID: id}
	}
	return fr
}

// SetFormDefs populates form definitions from an existing Registry.
func (fr *FormRegistry) SetFormDefs(reg *Registry) {
	for _, form := range reg.AllForms() {
		if r, ok := fr.registrations[form.ID]; ok {
			r.Def = form
		}
	}
}

// SetPDFMapped marks the given FormIDs as having PDF mappings.
func (fr *FormRegistry) SetPDFMapped(ids map[FormID]bool) {
	for id := range ids {
		if r, ok := fr.registrations[id]; ok {
			r.HasPDF = true
		}
	}
}

// SetMeFCovered marks the given FormIDs as having MeF XML builders.
func (fr *FormRegistry) SetMeFCovered(ids map[FormID]bool) {
	for id := range ids {
		if r, ok := fr.registrations[id]; ok {
			r.HasMeFXML = true
		}
	}
}

// SetCACovered marks the given FormIDs as having CA FTB XML builders.
func (fr *FormRegistry) SetCACovered(ids map[FormID]bool) {
	for id := range ids {
		if r, ok := fr.registrations[id]; ok {
			r.HasCAXML = true
		}
	}
}

// Get returns the registration for the given FormID, or nil if unknown.
func (fr *FormRegistry) Get(id FormID) *FormRegistration {
	return fr.registrations[id]
}

// All returns all registrations.
func (fr *FormRegistry) All() []*FormRegistration {
	result := make([]*FormRegistration, 0, len(fr.registrations))
	for _, r := range fr.registrations {
		result = append(result, r)
	}
	return result
}

// HasPDF returns true if the form has PDF mappings.
func (fr *FormRegistry) HasPDF(id FormID) bool {
	if r, ok := fr.registrations[id]; ok {
		return r.HasPDF
	}
	return false
}

// HasMeF returns true if the form has a MeF XML builder.
func (fr *FormRegistry) HasMeF(id FormID) bool {
	if r, ok := fr.registrations[id]; ok {
		return r.HasMeFXML
	}
	return false
}

// HasCA returns true if the form has a CA FTB XML builder.
func (fr *FormRegistry) HasCA(id FormID) bool {
	if r, ok := fr.registrations[id]; ok {
		return r.HasCAXML
	}
	return false
}

// ValidateCoverage checks that forms have expected coverage based on their
// jurisdiction. Returns a list of issues found.
//
// Rules:
//   - Every non-input form should have a form definition
//   - Every non-input form should have a PDF mapping (known gaps allowed)
//   - Every federal form should have a MeF XML builder
//   - Every CA form should have a CA XML builder (flow-through forms allowed)
func (fr *FormRegistry) ValidateCoverage(opts CoverageOpts) []CoverageGap {
	var gaps []CoverageGap

	inputForms := make(map[FormID]bool)
	for _, id := range InputFormIDs() {
		inputForms[id] = true
	}

	for _, r := range fr.registrations {
		if inputForms[r.ID] {
			continue
		}

		if r.Def == nil {
			gaps = append(gaps, CoverageGap{
				FormID:  r.ID,
				Missing: "form_def",
				Message: fmt.Sprintf("form %s has no form definition", r.ID),
			})
			continue
		}

		if !r.HasPDF && !opts.PDFGaps[r.ID] {
			gaps = append(gaps, CoverageGap{
				FormID:  r.ID,
				Missing: "pdf",
				Message: fmt.Sprintf("form %s has no PDF mapping", r.ID),
			})
		}

		if r.Def.Jurisdiction == Federal && !r.HasMeFXML {
			gaps = append(gaps, CoverageGap{
				FormID:  r.ID,
				Missing: "mef_xml",
				Message: fmt.Sprintf("federal form %s has no MeF XML builder", r.ID),
			})
		}

		if r.Def.Jurisdiction == StateCA && !r.HasCAXML && !opts.CAFlowThrough[r.ID] {
			gaps = append(gaps, CoverageGap{
				FormID:  r.ID,
				Missing: "ca_xml",
				Message: fmt.Sprintf("CA form %s has no CA FTB XML builder", r.ID),
			})
		}
	}

	return gaps
}

// CoverageOpts allows callers to specify known gaps and exceptions for
// ValidateCoverage.
type CoverageOpts struct {
	PDFGaps        map[FormID]bool // forms known to lack PDF mappings
	CAFlowThrough  map[FormID]bool // CA forms that flow through Form 540
}

// CoverageGap describes a missing piece of coverage for a form.
type CoverageGap struct {
	FormID  FormID
	Missing string // "form_def", "pdf", "mef_xml", "ca_xml"
	Message string
}
