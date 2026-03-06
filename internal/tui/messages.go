package tui

// StartInterviewMsg tells the App to switch from welcome to interview view.
type StartInterviewMsg struct {
	TaxYear   int
	StateCode string
	Continue  bool // if true, load saved state
}

// ShowSummaryMsg tells the App to switch from interview to summary view.
type ShowSummaryMsg struct {
	Results    map[string]float64
	StrInputs  map[string]string
	TaxYear    int
	State      string
}

// ImportPriorYearMsg tells the App to import a prior-year return.
type ImportPriorYearMsg struct {
	FilePath string
}

// PriorYearImportedMsg signals that prior-year import is complete.
type PriorYearImportedMsg struct {
	NumericValues map[string]float64
	StringValues  map[string]string
	TaxYear       int
	Err           error
}

// RequestExplanationMsg triggers a RAG-powered explanation.
type RequestExplanationMsg struct {
	FieldKey string
	Label    string
	FormName string
}

// ExplanationResponseMsg carries the LLM's explanation back to the view.
type ExplanationResponseMsg struct {
	Explanation string
	Err         error
}
