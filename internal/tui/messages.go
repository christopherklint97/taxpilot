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

// RequestWhyAskedMsg triggers a "why am I being asked this?" explanation.
type RequestWhyAskedMsg struct {
	FieldKey     string
	Label        string
	FilingStatus string
	AnsweredKeys map[string]string // key -> value for context
}

// WhyAskedResponseMsg carries the explanation back.
type WhyAskedResponseMsg struct {
	Explanation string
	Err         error
}

// RequestCADiffMsg triggers a CA vs federal difference explanation.
type RequestCADiffMsg struct {
	FieldKey string
	Label    string
}

// CADiffResponseMsg carries the CA difference explanation back.
type CADiffResponseMsg struct {
	Explanation string
	Err         error
}

// ShowReviewMsg tells the App to switch to the review view.
type ShowReviewMsg struct {
	Results      map[string]float64
	StrInputs    map[string]string
	PriorResults map[string]float64
	TaxYear      int
	State        string
}

// StartEFileMsg tells the App to switch to the e-file view.
type StartEFileMsg struct {
	Results     map[string]float64
	StrInputs   map[string]string
	TaxYear     int
	State       string
	FederalOnly bool
	StateOnly   bool
}
