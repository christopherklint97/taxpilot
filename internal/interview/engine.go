package interview

import (
	"fmt"
	"strconv"
	"strings"

	"taxpilot/internal/forms"
	"taxpilot/internal/forms/federal"
	"taxpilot/internal/forms/inputs"
	"taxpilot/internal/forms/state/ca"
)

// PriorYearDefault represents a prior-year value being offered as a default.
type PriorYearDefault struct {
	FieldKey   string
	Label      string
	PriorValue string  // formatted for display
	RawValue   float64 // numeric value (0 for strings)
	StrValue   string  // string value
}

// Engine drives the interactive Q&A flow by walking the dependency graph
// and collecting missing UserInput values.
type Engine struct {
	registry  *forms.Registry
	graph     *forms.DependencyGraph
	inputs    map[string]float64
	strInputs map[string]string
	taxYear   int
	// ordered list of questions to ask
	questions []Question
	current   int
	// prior-year data for pre-fill defaults
	priorYear    map[string]float64 // prior-year numeric values
	priorYearStr map[string]string  // prior-year string values
}

// Question represents a single question to ask the user.
type Question struct {
	Key      string   // field key like "1040:filing_status"
	Label    string   // human-readable label
	Prompt   string   // question text to show user
	Options  []string // for enum fields (filing status, etc.)
	IsString bool     // true for string inputs (names, SSN, EIN)
	FormName string   // which form this belongs to
}

// stringFields are field lines that should be treated as string inputs
// rather than numeric.
var stringFields = map[string]bool{
	"first_name":    true,
	"last_name":     true,
	"ssn":           true,
	"employer_name": true,
	"employer_ein":  true,
}

// NewEngine creates a new Engine, registers all forms, builds the dependency
// graph, and determines the ordered list of questions to ask.
func NewEngine(registry *forms.Registry, taxYear int) (*Engine, error) {
	e := &Engine{
		registry:  registry,
		inputs:    make(map[string]float64),
		strInputs: make(map[string]string),
		taxYear:   taxYear,
	}

	// Build dependency graph
	graph := forms.NewDependencyGraph(registry)
	if err := graph.Build(); err != nil {
		return nil, fmt.Errorf("build dependency graph: %w", err)
	}
	e.graph = graph

	// Collect all UserInput fields
	e.buildQuestions()

	return e, nil
}

// NewEngineWithInputs creates an Engine pre-populated with existing inputs,
// useful for resuming a saved session.
func NewEngineWithInputs(registry *forms.Registry, taxYear int, numInputs map[string]float64, strInputs map[string]string) (*Engine, error) {
	e, err := NewEngine(registry, taxYear)
	if err != nil {
		return nil, err
	}

	// Pre-populate inputs
	for k, v := range numInputs {
		e.inputs[k] = v
	}
	for k, v := range strInputs {
		e.strInputs[k] = v
	}

	// Skip already-answered questions
	e.skipAnswered()

	return e, nil
}

// NewEngineWithPriorYear creates an Engine with prior-year defaults.
// Questions that have prior-year values will show "Last year: $X. Same? [Y/n]"
func NewEngineWithPriorYear(registry *forms.Registry, taxYear int,
	priorNumeric map[string]float64, priorStr map[string]string) (*Engine, error) {
	e, err := NewEngine(registry, taxYear)
	if err != nil {
		return nil, err
	}

	e.priorYear = make(map[string]float64)
	for k, v := range priorNumeric {
		e.priorYear[k] = v
	}
	e.priorYearStr = make(map[string]string)
	for k, v := range priorStr {
		e.priorYearStr[k] = v
	}

	return e, nil
}

// GetPriorYearDefault returns the prior-year default for the current question, if any.
func (e *Engine) GetPriorYearDefault() *PriorYearDefault {
	if e.current >= len(e.questions) {
		return nil
	}
	q := &e.questions[e.current]

	// Check string prior-year values first (for string/enum fields)
	if q.IsString || len(q.Options) > 0 {
		if sv, ok := e.priorYearStr[q.Key]; ok && sv != "" {
			return &PriorYearDefault{
				FieldKey:   q.Key,
				Label:      q.Label,
				PriorValue: sv,
				RawValue:   0,
				StrValue:   sv,
			}
		}
	}

	// Check numeric prior-year values
	if nv, ok := e.priorYear[q.Key]; ok {
		return &PriorYearDefault{
			FieldKey:   q.Key,
			Label:      q.Label,
			PriorValue: formatCurrency(nv),
			RawValue:   nv,
			StrValue:   "",
		}
	}

	return nil
}

// AcceptDefault accepts the prior-year default for the current question.
func (e *Engine) AcceptDefault() error {
	d := e.GetPriorYearDefault()
	if d == nil {
		return fmt.Errorf("no prior-year default available for current question")
	}

	q := &e.questions[e.current]

	if len(q.Options) > 0 {
		// Enum field: store string and resolve numeric index
		e.strInputs[q.Key] = d.StrValue
		for i, opt := range q.Options {
			if opt == d.StrValue {
				e.inputs[q.Key] = float64(i + 1)
				break
			}
		}
	} else if q.IsString {
		e.strInputs[q.Key] = d.StrValue
		e.inputs[q.Key] = 0
	} else {
		e.inputs[q.Key] = d.RawValue
	}

	e.current++
	// Skip any already-answered questions
	for e.current < len(e.questions) && e.isAnswered(&e.questions[e.current]) {
		e.current++
	}

	return nil
}

// formatCurrency formats a float as "$1,234.00" for display within the engine.
func formatCurrency(amount float64) string {
	negative := amount < 0
	if negative {
		amount = -amount
	}

	whole := int64(amount)
	cents := int64((amount-float64(whole))*100 + 0.5)
	if cents >= 100 {
		whole++
		cents -= 100
	}

	s := fmt.Sprintf("%d", whole)
	if len(s) > 3 {
		var parts []string
		for len(s) > 3 {
			parts = append([]string{s[len(s)-3:]}, parts...)
			s = s[:len(s)-3]
		}
		parts = append([]string{s}, parts...)
		s = strings.Join(parts, ",")
	}

	result := fmt.Sprintf("$%s.%02d", s, cents)
	if negative {
		result = "-" + result
	}
	return result
}

// buildQuestions collects all UserInput fields and orders them logically.
func (e *Engine) buildQuestions() {
	var filingStatus []Question
	var personalInfo []Question
	var employerInfo []Question
	var w2Financial []Question
	var remaining []Question

	for _, form := range e.registry.AllForms() {
		for _, field := range form.Fields {
			if field.Type != forms.UserInput {
				continue
			}

			key := forms.FieldKey(form.ID, field.Line)
			isStr := stringFields[field.Line] || len(field.Options) > 0

			q := Question{
				Key:      key,
				Label:    field.Label,
				Prompt:   field.Prompt,
				Options:  field.Options,
				IsString: isStr,
				FormName: form.Name,
			}

			// Categorize for ordering
			switch {
			case field.Line == "filing_status":
				filingStatus = append(filingStatus, q)
			case field.Line == "first_name" || field.Line == "last_name" || field.Line == "ssn":
				personalInfo = append(personalInfo, q)
			case field.Line == "employer_name" || field.Line == "employer_ein":
				employerInfo = append(employerInfo, q)
			case form.ID == "w2":
				w2Financial = append(w2Financial, q)
			default:
				remaining = append(remaining, q)
			}
		}
	}

	// Order personal info: first_name, last_name, ssn
	personalInfo = sortByLineOrder(personalInfo, []string{"first_name", "last_name", "ssn"})
	// Order employer info: employer_name, employer_ein
	employerInfo = sortByLineOrder(employerInfo, []string{"employer_name", "employer_ein"})

	e.questions = nil
	e.questions = append(e.questions, filingStatus...)
	e.questions = append(e.questions, personalInfo...)
	e.questions = append(e.questions, employerInfo...)
	e.questions = append(e.questions, w2Financial...)
	e.questions = append(e.questions, remaining...)
	e.current = 0
}

// sortByLineOrder sorts questions by the order of their field lines in the
// provided order slice.
func sortByLineOrder(qs []Question, order []string) []Question {
	if len(qs) == 0 {
		return qs
	}
	result := make([]Question, 0, len(qs))
	idx := make(map[string]Question)
	for _, q := range qs {
		// Extract line from key (form:line)
		parts := strings.SplitN(q.Key, ":", 2)
		if len(parts) == 2 {
			idx[parts[1]] = q
		}
	}
	for _, line := range order {
		if q, ok := idx[line]; ok {
			result = append(result, q)
			delete(idx, line)
		}
	}
	// Append any remaining
	for _, q := range idx {
		result = append(result, q)
	}
	return result
}

// skipAnswered advances past questions that already have answers.
func (e *Engine) skipAnswered() {
	e.current = 0
	for e.current < len(e.questions) {
		q := &e.questions[e.current]
		if e.isAnswered(q) {
			e.current++
		} else {
			break
		}
	}
}

// isAnswered checks whether a question has already been answered.
func (e *Engine) isAnswered(q *Question) bool {
	if q.IsString {
		_, ok := e.strInputs[q.Key]
		return ok
	}
	_, ok := e.inputs[q.Key]
	return ok
}

// Questions returns the full list of questions.
func (e *Engine) Questions() []Question {
	return e.questions
}

// Current returns the current question, or nil if all questions are done.
func (e *Engine) Current() *Question {
	if e.current >= len(e.questions) {
		return nil
	}
	return &e.questions[e.current]
}

// HasNext returns true if there are more questions to answer.
func (e *Engine) HasNext() bool {
	return e.current < len(e.questions)
}

// Answer parses and stores the user's answer, then advances to the next question.
func (e *Engine) Answer(value string) error {
	if e.current >= len(e.questions) {
		return fmt.Errorf("no more questions")
	}

	q := &e.questions[e.current]
	value = strings.TrimSpace(value)

	if value == "" {
		return fmt.Errorf("please enter a value")
	}

	if len(q.Options) > 0 {
		// Handle enum fields -- accept number or text
		resolved, err := resolveOption(value, q.Options)
		if err != nil {
			return err
		}
		e.strInputs[q.Key] = resolved
		// Also store a numeric value (1-based index) for the solver
		for i, opt := range q.Options {
			if opt == resolved {
				e.inputs[q.Key] = float64(i + 1)
				break
			}
		}
	} else if q.IsString {
		e.strInputs[q.Key] = value
		// Store 0 as numeric placeholder so MissingInputs is satisfied
		e.inputs[q.Key] = 0
	} else {
		// Numeric field -- strip $ and commas
		cleaned := strings.ReplaceAll(value, ",", "")
		cleaned = strings.ReplaceAll(cleaned, "$", "")
		num, err := strconv.ParseFloat(cleaned, 64)
		if err != nil {
			return fmt.Errorf("please enter a valid number (got %q)", value)
		}
		e.inputs[q.Key] = num
	}

	e.current++
	// Skip any already-answered questions (in case of resume)
	for e.current < len(e.questions) && e.isAnswered(&e.questions[e.current]) {
		e.current++
	}

	return nil
}

// resolveOption matches user input against the options list.
// Accepts a 1-based number or a case-insensitive prefix/exact match.
func resolveOption(input string, options []string) (string, error) {
	// Try as number first
	if n, err := strconv.Atoi(input); err == nil {
		if n >= 1 && n <= len(options) {
			return options[n-1], nil
		}
		return "", fmt.Errorf("please enter a number between 1 and %d", len(options))
	}

	// Try case-insensitive match
	lower := strings.ToLower(input)
	for _, opt := range options {
		if strings.ToLower(opt) == lower {
			return opt, nil
		}
	}

	// Try prefix match
	var matches []string
	for _, opt := range options {
		if strings.HasPrefix(strings.ToLower(opt), lower) {
			matches = append(matches, opt)
		}
	}
	if len(matches) == 1 {
		return matches[0], nil
	}

	return "", fmt.Errorf("invalid option %q — choose from: %s", input, strings.Join(options, ", "))
}

// Back moves to the previous question. Returns false if already at the first question.
func (e *Engine) Back() bool {
	if e.current <= 0 {
		return false
	}
	e.current--
	return true
}

// Solve runs the dependency graph solver with all collected inputs.
func (e *Engine) Solve() (map[string]float64, error) {
	return e.graph.Solve(e.inputs, e.strInputs, e.taxYear)
}

// Inputs returns the numeric inputs collected so far.
func (e *Engine) Inputs() map[string]float64 {
	result := make(map[string]float64, len(e.inputs))
	for k, v := range e.inputs {
		result[k] = v
	}
	return result
}

// StrInputs returns the string inputs collected so far.
func (e *Engine) StrInputs() map[string]string {
	result := make(map[string]string, len(e.strInputs))
	for k, v := range e.strInputs {
		result[k] = v
	}
	return result
}

// CurrentFieldKey returns the field key for the current question,
// or an empty string if all questions are done.
func (e *Engine) CurrentFieldKey() string {
	if e.current >= len(e.questions) {
		return ""
	}
	return e.questions[e.current].Key
}

// Progress returns the current question index (0-based) and total question count.
func (e *Engine) Progress() (current int, total int) {
	return e.current, len(e.questions)
}

// SetupRegistry creates a Registry with all known forms registered.
func SetupRegistry() *forms.Registry {
	reg := forms.NewRegistry()
	reg.Register(inputs.W2())
	reg.Register(federal.F1040())
	reg.Register(ca.F540())
	reg.Register(ca.ScheduleCA())
	return reg
}
