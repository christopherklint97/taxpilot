package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"taxpilot/internal/efile"
	"taxpilot/internal/efile/ca"
	"taxpilot/internal/efile/mef"
	"taxpilot/internal/interview"
	"taxpilot/internal/knowledge"
	"taxpilot/internal/llm"
	"taxpilot/internal/pdf"
	"taxpilot/internal/state"
	"taxpilot/internal/tui"
	"taxpilot/internal/tui/views"
)

// factory holds the shared state needed to build ViewFactory callbacks.
type factory struct {
	taxYear      int
	stateCode    string
	llmClient    *llm.Client
	explainer    *llm.Explainer
	rag          *knowledge.RAG
	priorNumeric map[string]float64
	priorString  map[string]string
}

// buildFactory creates a factory with optional LLM support.
func buildFactory(taxYear int, stateCode string, _ string) *factory {
	f := &factory{
		taxYear:   taxYear,
		stateCode: stateCode,
	}

	// Try to initialize LLM client
	client, err := llm.NewClient("")
	if err == nil {
		f.llmClient = client
		f.explainer = llm.NewExplainer(client)
		store := knowledge.NewStore()
		f.rag = knowledge.NewRAG(store, client)
	}

	return f
}

// ViewFactory builds the tui.ViewFactory with all callbacks wired.
func (f *factory) ViewFactory() tui.ViewFactory {
	vf := tui.ViewFactory{
		MakeInterview:   f.makeInterview,
		MakeSummary:     f.makeSummary,
		ImportPriorYear: f.importPriorYear,
		MakeEFile:       f.makeEFile,
		MakeReview:      f.makeReview,
		ExportPDF:       f.exportPDF,
		SubmitEFile:     f.submitEFile,
	}

	if f.llmClient != nil {
		vf.Explain = f.explain
		vf.ExplainWhy = f.explainWhy
		vf.ExplainCADiff = f.explainCADiff
	}

	return vf
}

func (f *factory) makeInterview(msg tui.StartInterviewMsg) (tea.Model, error) {
	registry := interview.SetupRegistry()

	var engine *interview.Engine
	var err error

	if msg.Continue {
		ret, loadErr := state.Load(state.DefaultStorePath())
		if loadErr != nil {
			return nil, fmt.Errorf("load saved state: %w", loadErr)
		}
		engine, err = interview.NewEngineWithInputs(registry, ret.TaxYear, ret.Inputs, ret.StrInputs)
		if err != nil {
			return nil, err
		}
		view := views.NewInterviewView(engine, ret.TaxYear, ret.State)
		return view, nil
	}

	if f.priorNumeric != nil {
		engine, err = interview.NewEngineWithPriorYear(registry, msg.TaxYear, f.priorNumeric, f.priorString, msg.StateCode)
	} else {
		engine, err = interview.NewEngine(registry, msg.TaxYear)
	}
	if err != nil {
		return nil, err
	}

	view := views.NewInterviewView(engine, msg.TaxYear, msg.StateCode)
	return view, nil
}

func (f *factory) makeSummary(msg tui.ShowSummaryMsg) tea.Model {
	// Save state as side effect
	ret := state.NewTaxReturn(msg.TaxYear, msg.State)
	ret.Inputs = msg.Results
	ret.StrInputs = msg.StrInputs
	ret.Computed = msg.Results
	ret.Complete = true
	ret.LastUpdated = time.Now().UTC().Format(time.RFC3339)
	_ = state.Save(state.DefaultStorePath(), ret)

	return views.NewSummaryView(msg.Results, msg.StrInputs, msg.TaxYear, msg.State)
}

func (f *factory) importPriorYear(msg tui.ImportPriorYearMsg) tea.Msg {
	merged, formNames, err := pdf.ParseMultipleFiles(msg.FilePaths)
	if err != nil {
		return tui.PriorYearImportedMsg{Err: err}
	}

	// Merge into any existing prior-year data (incremental imports).
	if f.priorNumeric == nil {
		f.priorNumeric = make(map[string]float64)
	}
	if f.priorString == nil {
		f.priorString = make(map[string]string)
	}
	for k, v := range merged.Fields {
		f.priorNumeric[k] = v
	}
	for k, v := range merged.StrFields {
		f.priorString[k] = v
	}

	return tui.PriorYearImportedMsg{
		NumericValues: f.priorNumeric,
		StringValues:  f.priorString,
		TaxYear:       merged.TaxYear,
		FormNames:     formNames,
	}
}

func (f *factory) makeEFile(msg tui.StartEFileMsg) tea.Model {
	return views.NewEFileView(
		msg.Results, msg.StrInputs, msg.TaxYear, msg.State,
		msg.FederalOnly, msg.StateOnly,
	)
}

func (f *factory) makeReview(msg tui.ShowReviewMsg) tea.Model {
	return views.NewReviewView(msg)
}

func (f *factory) explain(msg tui.RequestExplanationMsg) tea.Msg {
	jurisdiction := knowledge.JurisdictionFederal
	if len(msg.FormName) >= 2 && msg.FormName[:2] == "ca" {
		jurisdiction = knowledge.JurisdictionCA
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := f.rag.QueryForField(ctx, msg.FieldKey, msg.Label, jurisdiction)
	return tui.ExplanationResponseMsg{
		Explanation: result,
		Err:         err,
	}
}

func (f *factory) explainWhy(msg tui.RequestWhyAskedMsg) tea.Msg {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := f.explainer.ExplainWhyAsked(ctx, msg.FieldKey, msg.Label, msg.FilingStatus, msg.AnsweredKeys)
	return tui.WhyAskedResponseMsg{
		Explanation: result,
		Err:         err,
	}
}

func (f *factory) explainCADiff(msg tui.RequestCADiffMsg) tea.Msg {
	diff := interview.GetCADifference(msg.FieldKey)
	if diff == nil {
		return tui.CADiffResponseMsg{
			Explanation: "No known CA vs federal difference for this field.",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := f.explainer.ExplainCADifference(ctx, diff.Area, diff.Federal, diff.California)
	return tui.CADiffResponseMsg{
		Explanation: result,
		Err:         err,
	}
}

func (f *factory) exportPDF(msg tui.ExportPDFMsg) tea.Msg {
	outputDir := msg.OutputDir
	if outputDir == "" {
		home, _ := os.UserHomeDir()
		outputDir = filepath.Join(home, ".taxpilot", "export")
	}

	paths, err := pdf.ExportReturn(outputDir, msg.Results, msg.StrInputs, msg.TaxYear)
	return tui.ExportPDFResultMsg{
		Paths: paths,
		Err:   err,
	}
}

func (f *factory) submitEFile(msg tui.EFileSubmitMsg) tea.Msg {
	var fedResult *tui.EFileSubmissionResult
	var caResult *tui.EFileSubmissionResult

	auth, _ := msg.Auth.(*efile.EFileAuth)

	// Submit federal
	if !msg.StateOnly && auth != nil {
		fedXML, err := mef.GenerateReturn(msg.Results, msg.StrInputs, msg.TaxYear)
		if err != nil {
			return tui.EFileResultMsg{Err: fmt.Errorf("generate federal XML: %w", err)}
		}

		client := mef.NewTestClient(true)
		result, err := client.SendSubmission(fedXML, auth.SelfSelectPIN)
		if err != nil {
			return tui.EFileResultMsg{Err: fmt.Errorf("federal submission: %w", err)}
		}

		fedResult = &tui.EFileSubmissionResult{
			SubmissionID: result.SubmissionID,
			Status:       result.Status.String(),
			Message:      result.Message,
		}
	}

	// Submit CA
	if msg.State == "CA" && !msg.FederalOnly && auth != nil {
		caXML, err := ca.GenerateReturn(msg.Results, msg.StrInputs, msg.TaxYear)
		if err != nil {
			return tui.EFileResultMsg{Err: fmt.Errorf("generate CA XML: %w", err)}
		}

		client := ca.NewTestClient(true)
		priorCAagi := 0.0
		result, err := client.SendSubmission(caXML, auth.CASelfSelectPIN, priorCAagi)
		if err != nil {
			return tui.EFileResultMsg{Err: fmt.Errorf("CA submission: %w", err)}
		}

		caResult = &tui.EFileSubmissionResult{
			SubmissionID: result.SubmissionID,
			Status:       result.Status.String(),
			Message:      result.Message,
		}
	}

	return tui.EFileResultMsg{
		FederalResult: fedResult,
		CAResult:      caResult,
	}
}
