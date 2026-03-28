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
	"taxpilot/internal/forms"
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
func buildFactory(taxYear int, stateCode string, _ string, modelOverride string) *factory {
	f := &factory{
		taxYear:   taxYear,
		stateCode: stateCode,
	}

	// Try to initialize LLM client
	client, err := llm.NewClient("")
	if err == nil {
		if modelOverride != "" {
			client.SetModel(modelOverride)
		}
		client.SetZDR(true)
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
		MakeRollforward: f.makeRollforward,
	}

	if f.llmClient != nil {
		vf.Explain = f.explain
		vf.ExplainWhy = f.explainWhy
		vf.ExplainCADiff = f.explainCADiff
		vf.AskAI = f.askAI
	}

	return vf
}

func (f *factory) makeInterview(msg tui.StartInterviewMsg) (tea.Model, error) {
	registry := interview.SetupRegistry()

	// Debug: log factory prior-year state
	debugPriorYear("makeInterview entry", f.priorNumeric, f.priorString)

	// Resolve prior-year data: saved TaxPilot state first, then --import fallback
	priorNumeric, priorStr := f.resolvePriorYear(msg.TaxYear)

	debugPriorYear("after resolvePriorYear", priorNumeric, priorStr)

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
		if priorNumeric != nil {
			engine.SetPriorYear(priorNumeric, priorStr, ret.State)
			n, s := engine.PriorYearCount()
			debugLog("SetPriorYear called: engine now has %d numeric, %d string", n, s)
		} else {
			debugLog("priorNumeric is nil — NOT calling SetPriorYear")
		}
		view := views.NewInterviewView(engine, ret.TaxYear, ret.State)
		return view, nil
	}

	if priorNumeric != nil {
		engine, err = interview.NewEngineWithPriorYear(registry, msg.TaxYear, priorNumeric, priorStr, msg.StateCode)
	} else {
		engine, err = interview.NewEngine(registry, msg.TaxYear)
	}
	if err != nil {
		return nil, err
	}

	view := views.NewInterviewView(engine, msg.TaxYear, msg.StateCode)
	return view, nil
}

// resolvePriorYear returns prior-year data with priority:
// 1. Saved TaxPilot state from prior year (~/.taxpilot/state_YYYY.json)
// 2. Imported PDF data (--import flag)
func (f *factory) resolvePriorYear(taxYear int) (map[string]float64, map[string]string) {
	// First priority: saved TaxPilot state from prior year
	priorRet, err := state.LoadPriorYear(taxYear)
	if err == nil {
		ctx := state.ExtractPriorYearContext(priorRet)
		debugLog("resolvePriorYear: loaded state_%d.json — AllValues=%d, AllStrValues=%d, Inputs=%d, StrInputs=%d, Computed=%d",
			taxYear-1, len(ctx.AllValues), len(ctx.AllStrValues),
			len(priorRet.Inputs), len(priorRet.StrInputs), len(priorRet.Computed))
		if len(ctx.AllValues) > 0 || len(ctx.AllStrValues) > 0 {
			return ctx.AllValues, ctx.AllStrValues
		}
		debugLog("resolvePriorYear: state file found but AllValues and AllStrValues are empty")
	} else {
		debugLog("resolvePriorYear: LoadPriorYear(%d) error: %v", taxYear, err)
	}

	// Second priority: imported PDF data
	if f.priorNumeric != nil {
		debugLog("resolvePriorYear: using f.priorNumeric (%d values)", len(f.priorNumeric))
		return f.priorNumeric, f.priorString
	}

	debugLog("resolvePriorYear: no prior-year data found")
	return nil, nil
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
	debugLog("importPriorYear: merged.Fields=%d, merged.StrFields=%d, merged.TaxYear=%d",
		len(merged.Fields), len(merged.StrFields), merged.TaxYear)
	for k, v := range merged.Fields {
		f.priorNumeric[k] = v
	}
	for k, v := range merged.StrFields {
		f.priorString[k] = v
	}
	debugLog("importPriorYear: after merge f.priorNumeric=%d, f.priorString=%d",
		len(f.priorNumeric), len(f.priorString))

	// Persist imported prior-year data so it survives across sessions.
	// This lets resolvePriorYear → LoadPriorYear find it on --continue.
	if merged.TaxYear > 0 {
		priorRet := state.NewTaxReturn(merged.TaxYear, f.stateCode)
		priorRet.Inputs = f.priorNumeric
		priorRet.StrInputs = f.priorString
		priorRet.Complete = true
		savePath := state.YearStorePath(merged.TaxYear)
		debugLog("importPriorYear: persisting to %s (inputs=%d, str_inputs=%d)",
			savePath, len(priorRet.Inputs), len(priorRet.StrInputs))
		if err := state.Save(savePath, priorRet); err != nil {
			debugLog("importPriorYear: save error: %v", err)
		}
	}

	return tui.PriorYearImportedMsg{
		NumericValues: f.priorNumeric,
		StringValues:  f.priorString,
		TaxYear:       merged.TaxYear,
		FormNames:     formNames,
	}
}

func (f *factory) makeRollforward(msg tui.StartRollforwardMsg) (tea.Model, error) {
	// Load prior-year return: try saved state first, fall back to --import data
	priorRet, err := state.LoadPriorYear(msg.TaxYear)
	if err != nil && f.priorNumeric != nil {
		// Build a TaxReturn from imported PDF data
		debugLog("makeRollforward: LoadPriorYear failed, using --import data (%d numeric, %d string)",
			len(f.priorNumeric), len(f.priorString))
		priorRet = state.NewTaxReturn(msg.TaxYear-1, msg.StateCode)
		priorRet.Inputs = f.priorNumeric
		priorRet.StrInputs = f.priorString
		priorRet.Complete = true
		err = nil
	}
	if err != nil {
		return nil, fmt.Errorf("no prior year return found for %d — use --import to load PDFs: %w",
			msg.TaxYear, err)
	}

	// If state code not set on prior return, use the one from the flag
	if priorRet.State == "" {
		priorRet.State = msg.StateCode
	}

	registry := interview.SetupRegistry()

	rf, err := interview.NewRollforward(registry, msg.TaxYear, priorRet)
	if err != nil {
		return nil, fmt.Errorf("rollforward: %w", err)
	}

	debugLog("makeRollforward: %d -> %d, %d fields, %d flagged, %d changes",
		rf.PriorYear, rf.TaxYear, len(rf.Fields), rf.CountFlagged(), len(rf.Changes))

	view := views.NewRollforwardView(rf)
	return view, nil
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

// readFirstChunk reads the first chunk from a streaming channel and returns
// it as an AIStreamChunkMsg. The channel is wrapped in an interface{} channel
// so the tui package doesn't need to import llm.
func readFirstChunk(ch <-chan llm.StreamChunk) tui.AIStreamChunkMsg {
	chunk, ok := <-ch
	if !ok {
		return tui.AIStreamChunkMsg{Done: true}
	}
	// Wrap the llm channel into an interface{} channel for the view
	wrapped := make(chan interface{}, cap(ch)+1)
	go func() {
		defer close(wrapped)
		for c := range ch {
			wrapped <- c
		}
	}()
	return tui.AIStreamChunkMsg{
		Text: chunk.Text,
		Err:  chunk.Err,
		Done: chunk.Done,
		Ch:   wrapped,
	}
}

func (f *factory) explain(msg tui.RequestExplanationMsg) tea.Msg {
	jurisdiction := knowledge.JurisdictionFederal
	if len(msg.FormName) >= 2 && msg.FormName[:2] == "ca" {
		jurisdiction = knowledge.JurisdictionCA
	}

	ctx := context.Background()
	ch, err := f.rag.QueryForFieldStream(ctx, msg.FieldKey, msg.Label, jurisdiction)
	if err != nil {
		return tui.ExplanationResponseMsg{Err: err}
	}
	return readFirstChunk(ch)
}

func (f *factory) explainWhy(msg tui.RequestWhyAskedMsg) tea.Msg {
	ctx := context.Background()
	ch, err := f.explainer.ExplainWhyAskedStream(ctx, msg.FieldKey, msg.Label, msg.FilingStatus, msg.AnsweredKeys)
	if err != nil {
		return tui.WhyAskedResponseMsg{Err: err}
	}
	return readFirstChunk(ch)
}

func (f *factory) askAI(msg tui.RequestAIPromptMsg) tea.Msg {
	ctx := context.Background()
	ch, err := f.explainer.AskAboutFieldStream(ctx, msg.UserQuestion, msg.FieldKey, msg.Label, msg.FormName, msg.FilingStatus, msg.AnsweredKeys)
	if err != nil {
		return tui.AIPromptResponseMsg{Err: err}
	}
	return readFirstChunk(ch)
}

func (f *factory) explainCADiff(msg tui.RequestCADiffMsg) tea.Msg {
	diff := interview.GetCADifference(msg.FieldKey)
	if diff == nil {
		return tui.CADiffResponseMsg{
			Explanation: "No known CA vs federal difference for this field.",
		}
	}

	ctx := context.Background()
	ch, err := f.explainer.ExplainCADifferenceStream(ctx, diff.Area, diff.Federal, diff.California)
	if err != nil {
		return tui.CADiffResponseMsg{Err: err}
	}
	return readFirstChunk(ch)
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
	if msg.State == forms.StateCodeCA && !msg.FederalOnly && auth != nil {
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

// debugLog writes a line to ~/.taxpilot/debug.log for troubleshooting.
func debugLog(format string, args ...interface{}) {
	home, _ := os.UserHomeDir()
	logPath := filepath.Join(home, ".taxpilot", "debug.log")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	line := fmt.Sprintf(format, args...)
	fmt.Fprintf(f, "[%s] %s\n", time.Now().Format("15:04:05"), line)
}

// debugPriorYear logs the state of prior-year data maps.
func debugPriorYear(label string, numeric map[string]float64, str map[string]string) {
	if numeric == nil && str == nil {
		debugLog("%s: numeric=nil, str=nil", label)
		return
	}
	numLen := 0
	strLen := 0
	if numeric != nil {
		numLen = len(numeric)
	}
	if str != nil {
		strLen = len(str)
	}
	debugLog("%s: numeric=%d values, str=%d values", label, numLen, strLen)
	// Log first few keys as sample
	count := 0
	for k, v := range numeric {
		if count >= 5 {
			break
		}
		debugLog("  numeric[%s] = %v", k, v)
		count++
	}
	count = 0
	for k, v := range str {
		if count >= 5 {
			break
		}
		debugLog("  str[%s] = %q", k, v)
		count++
	}
}
