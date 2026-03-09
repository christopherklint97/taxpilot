package forms

import "testing"

func TestNewFormRegistryHasAllFormIDs(t *testing.T) {
	fr := NewFormRegistry()
	for _, id := range AllFormIDs() {
		if fr.Get(id) == nil {
			t.Errorf("FormRegistry missing entry for %s", id)
		}
	}
}

func TestFormRegistryGetUnknown(t *testing.T) {
	fr := NewFormRegistry()
	if fr.Get("nonexistent") != nil {
		t.Error("expected nil for unknown FormID")
	}
}

func TestFormRegistrySetFormDefs(t *testing.T) {
	fr := NewFormRegistry()
	reg := NewRegistry()
	reg.Register(&FormDef{ID: FormF1040, Name: "Form 1040", Jurisdiction: Federal})

	fr.SetFormDefs(reg)

	r := fr.Get(FormF1040)
	if r == nil || r.Def == nil {
		t.Fatal("expected FormDef to be set for 1040")
	}
	if r.Def.Name != "Form 1040" {
		t.Errorf("expected name 'Form 1040', got %q", r.Def.Name)
	}
}

func TestFormRegistrySetPDFMapped(t *testing.T) {
	fr := NewFormRegistry()
	fr.SetPDFMapped(map[FormID]bool{FormF1040: true, FormScheduleA: true})

	if !fr.HasPDF(FormF1040) {
		t.Error("expected HasPDF(1040) = true")
	}
	if !fr.HasPDF(FormScheduleA) {
		t.Error("expected HasPDF(schedule_a) = true")
	}
	if fr.HasPDF(FormScheduleB) {
		t.Error("expected HasPDF(schedule_b) = false")
	}
}

func TestFormRegistrySetMeFCovered(t *testing.T) {
	fr := NewFormRegistry()
	fr.SetMeFCovered(map[FormID]bool{FormF1040: true})

	if !fr.HasMeF(FormF1040) {
		t.Error("expected HasMeF(1040) = true")
	}
	if fr.HasMeF(FormCA540) {
		t.Error("expected HasMeF(ca_540) = false")
	}
}

func TestFormRegistrySetCACovered(t *testing.T) {
	fr := NewFormRegistry()
	fr.SetCACovered(map[FormID]bool{FormCA540: true})

	if !fr.HasCA(FormCA540) {
		t.Error("expected HasCA(ca_540) = true")
	}
	if fr.HasCA(FormF1040) {
		t.Error("expected HasCA(1040) = false")
	}
}

func TestFormRegistryHasPDFUnknownForm(t *testing.T) {
	fr := NewFormRegistry()
	if fr.HasPDF("nonexistent") {
		t.Error("expected HasPDF for unknown form to be false")
	}
}

func TestFormRegistryHasMeFUnknownForm(t *testing.T) {
	fr := NewFormRegistry()
	if fr.HasMeF("nonexistent") {
		t.Error("expected HasMeF for unknown form to be false")
	}
}

func TestFormRegistryHasCAUnknownForm(t *testing.T) {
	fr := NewFormRegistry()
	if fr.HasCA("nonexistent") {
		t.Error("expected HasCA for unknown form to be false")
	}
}

func TestFormRegistryAll(t *testing.T) {
	fr := NewFormRegistry()
	all := fr.All()
	if len(all) != len(AllFormIDs()) {
		t.Errorf("expected %d registrations, got %d", len(AllFormIDs()), len(all))
	}
}

func TestFormRegistryValidateCoverage(t *testing.T) {
	fr := NewFormRegistry()

	// Set up a minimal registry with one federal and one CA form
	reg := NewRegistry()
	reg.Register(&FormDef{ID: FormF1040, Name: "Form 1040", Jurisdiction: Federal})
	reg.Register(&FormDef{ID: FormCA540, Name: "Form 540", Jurisdiction: StateCA})
	fr.SetFormDefs(reg)

	// Give 1040 full coverage
	fr.SetPDFMapped(map[FormID]bool{FormF1040: true})
	fr.SetMeFCovered(map[FormID]bool{FormF1040: true})

	// Give CA 540 partial coverage (no CA XML)
	fr.SetPDFMapped(map[FormID]bool{FormCA540: true})

	opts := CoverageOpts{
		PDFGaps:       make(map[FormID]bool),
		CAFlowThrough: make(map[FormID]bool),
	}

	gaps := fr.ValidateCoverage(opts)

	// Should find gaps for: all other non-input forms missing defs,
	// and CA 540 missing CA XML
	var foundCA540XML bool
	for _, g := range gaps {
		if g.FormID == FormCA540 && g.Missing == "ca_xml" {
			foundCA540XML = true
		}
	}
	if !foundCA540XML {
		t.Error("expected gap for CA 540 missing CA XML builder")
	}

	// 1040 should have no gaps
	for _, g := range gaps {
		if g.FormID == FormF1040 {
			t.Errorf("unexpected gap for 1040: %s", g.Message)
		}
	}
}

func TestFormRegistryValidateCoverageWithExceptions(t *testing.T) {
	fr := NewFormRegistry()

	reg := NewRegistry()
	reg.Register(&FormDef{ID: FormF2555, Name: "Form 2555", Jurisdiction: Federal})
	reg.Register(&FormDef{ID: FormF3514, Name: "Form 3514", Jurisdiction: StateCA})
	fr.SetFormDefs(reg)

	fr.SetMeFCovered(map[FormID]bool{FormF2555: true})

	opts := CoverageOpts{
		PDFGaps:       map[FormID]bool{FormF2555: true},   // known gap
		CAFlowThrough: map[FormID]bool{FormF3514: true},   // flow-through
	}

	gaps := fr.ValidateCoverage(opts)

	// 2555 should not have PDF gap (it's in known gaps)
	for _, g := range gaps {
		if g.FormID == FormF2555 && g.Missing == "pdf" {
			t.Error("form 2555 PDF gap should be suppressed by PDFGaps exception")
		}
	}

	// 3514 should not have CA XML gap (it's flow-through)
	for _, g := range gaps {
		if g.FormID == FormF3514 && g.Missing == "ca_xml" {
			t.Error("form 3514 CA XML gap should be suppressed by CAFlowThrough exception")
		}
	}
}
