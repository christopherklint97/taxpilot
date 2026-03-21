package main

import (
	"os"
	"path/filepath"
	"testing"

	"taxpilot/internal/state"
	"taxpilot/internal/tui"
)

func TestMakeInterview_New(t *testing.T) {
	f := buildFactory(2025, "CA", "", "")
	msg := tui.StartInterviewMsg{TaxYear: 2025, StateCode: "CA"}
	view, err := f.makeInterview(msg)
	if err != nil {
		t.Fatalf("makeInterview (new): %v", err)
	}
	if view == nil {
		t.Fatal("makeInterview returned nil view")
	}
}

func TestMakeInterview_WithPriorYear(t *testing.T) {
	f := buildFactory(2025, "CA", "", "")
	f.priorNumeric = map[string]float64{"1040:11": 75000}
	f.priorString = map[string]string{"1040:first_name": "Jane"}
	msg := tui.StartInterviewMsg{TaxYear: 2025, StateCode: "CA"}
	view, err := f.makeInterview(msg)
	if err != nil {
		t.Fatalf("makeInterview (prior year): %v", err)
	}
	if view == nil {
		t.Fatal("makeInterview returned nil view")
	}
}

func TestMakeInterview_Continue(t *testing.T) {
	// Save state first so continue can load it
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "state.json")
	ret := state.NewTaxReturn(2025, "CA")
	ret.Inputs = map[string]float64{"1040:filing_status": 1}
	ret.StrInputs = map[string]string{"1040:first_name": "Test"}
	if err := state.Save(storePath, ret); err != nil {
		t.Fatalf("save state: %v", err)
	}

	// Override DefaultStorePath by setting env
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	// Create .taxpilot dir for DefaultStorePath
	os.MkdirAll(filepath.Join(tmpDir, ".taxpilot"), 0o755)
	// Copy state file to default location
	defaultPath := state.DefaultStorePath()
	data, _ := os.ReadFile(storePath)
	os.WriteFile(defaultPath, data, 0o644)
	defer os.Setenv("HOME", origHome)

	f := buildFactory(2025, "CA", "", "")
	msg := tui.StartInterviewMsg{TaxYear: 2025, StateCode: "CA", Continue: true}
	view, err := f.makeInterview(msg)
	if err != nil {
		t.Fatalf("makeInterview (continue): %v", err)
	}
	if view == nil {
		t.Fatal("makeInterview returned nil view")
	}
}

func TestMakeSummary_SavesState(t *testing.T) {
	// Set HOME to temp dir so state saves there
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	os.MkdirAll(filepath.Join(tmpDir, ".taxpilot"), 0o755)
	defer os.Setenv("HOME", origHome)

	f := buildFactory(2025, "CA", "", "")
	msg := tui.ShowSummaryMsg{
		Results:   map[string]float64{"1040:11": 50000},
		StrInputs: map[string]string{"1040:first_name": "Alice"},
		TaxYear:   2025,
		State:     "CA",
	}

	view := f.makeSummary(msg)
	if view == nil {
		t.Fatal("makeSummary returned nil")
	}

	// Verify state was saved
	defaultPath := state.DefaultStorePath()
	ret, err := state.Load(defaultPath)
	if err != nil {
		t.Fatalf("load saved state: %v", err)
	}
	if !ret.Complete {
		t.Error("expected saved return to be marked complete")
	}
	if ret.TaxYear != 2025 {
		t.Errorf("expected tax year 2025, got %d", ret.TaxYear)
	}
}

func TestExplainCallbacks_NilLLM(t *testing.T) {
	// When no API key, LLM callbacks should be nil
	f := buildFactory(2025, "CA", "", "")
	vf := f.ViewFactory()

	// Without an API key, Explain should be nil (gracefully handled by App)
	// The factory only sets Explain if llmClient != nil
	if f.llmClient != nil {
		t.Skip("LLM client initialized (OPENROUTER_API_KEY is set)")
	}
	if vf.Explain != nil {
		t.Error("expected Explain to be nil without LLM client")
	}
	if vf.ExplainWhy != nil {
		t.Error("expected ExplainWhy to be nil without LLM client")
	}
	if vf.ExplainCADiff != nil {
		t.Error("expected ExplainCADiff to be nil without LLM client")
	}
}

func TestExportPDF(t *testing.T) {
	tmpDir := t.TempDir()
	f := buildFactory(2025, "CA", "", "")
	msg := tui.ExportPDFMsg{
		Results:   map[string]float64{"1040:11": 50000, "1040:15": 40000},
		StrInputs: map[string]string{"1040:first_name": "Test"},
		TaxYear:   2025,
		OutputDir: tmpDir,
	}

	result := f.exportPDF(msg)
	res, ok := result.(tui.ExportPDFResultMsg)
	if !ok {
		t.Fatalf("expected ExportPDFResultMsg, got %T", result)
	}
	if res.Err != nil {
		t.Fatalf("exportPDF error: %v", res.Err)
	}
	if len(res.Paths) == 0 {
		t.Error("expected at least one exported file")
	}
}

func TestViewFactory_AllFieldsSet(t *testing.T) {
	f := buildFactory(2025, "CA", "", "")
	vf := f.ViewFactory()

	if vf.MakeInterview == nil {
		t.Error("MakeInterview is nil")
	}
	if vf.MakeSummary == nil {
		t.Error("MakeSummary is nil")
	}
	if vf.ImportPriorYear == nil {
		t.Error("ImportPriorYear is nil")
	}
	if vf.MakeEFile == nil {
		t.Error("MakeEFile is nil")
	}
	if vf.MakeReview == nil {
		t.Error("MakeReview is nil")
	}
	if vf.ExportPDF == nil {
		t.Error("ExportPDF is nil")
	}
	if vf.SubmitEFile == nil {
		t.Error("SubmitEFile is nil")
	}
}
