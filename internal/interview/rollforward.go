package interview

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"taxpilot/internal/forms"
	"taxpilot/internal/state"
	"taxpilot/pkg/taxmath"
)

// FieldChange describes a field whose computed value changed between tax years.
type FieldChange struct {
	Key           string
	Label         string
	FormName      string
	FieldType     forms.FieldType
	OldValue      float64
	NewValue      float64
	Delta         float64
	PercentChange float64
	Reason        string // human-readable explanation
}

// RollforwardField represents a single field in the rollforward view.
type RollforwardField struct {
	Key        string
	Line       string
	FormID     forms.FormID
	FormName   string
	Label      string
	FieldType  forms.FieldType
	Value      float64
	StrValue   string
	IsString   bool
	IsInteger  bool // whole numbers (counts), not currency
	Options    []string
	PriorValue float64
	PriorStr   string
	Changed    bool   // true if new year value != prior year value
	Flagged    bool   // true if needs user attention
	FlagReason string // why it's flagged
}

// Rollforward holds the state for rollforward mode.
type Rollforward struct {
	Registry      *forms.Registry
	Graph         *forms.DependencyGraph
	Inputs        map[string]float64
	StrInputs     map[string]string
	Computed      map[string]float64 // current solve result (new year)
	PriorComputed map[string]float64 // same inputs solved under old year
	PriorInputs   map[string]float64 // original prior-year input values (frozen)
	PriorStrIn    map[string]string  // original prior-year string inputs (frozen)
	Fields        []RollforwardField // ordered field list
	Changes       []FieldChange
	ParamChanges  []taxmath.ParameterChange
	TaxYear       int
	PriorYear     int
	StateCode     string
}

// NewRollforward creates a rollforward from a prior-year return.
func NewRollforward(registry *forms.Registry, taxYear int, prior *state.TaxReturn) (*Rollforward, error) {
	graph := forms.NewDependencyGraph(registry)
	if err := graph.Build(); err != nil {
		return nil, fmt.Errorf("build dependency graph: %w", err)
	}

	rf := &Rollforward{
		Registry:  registry,
		Graph:     graph,
		Inputs:    make(map[string]float64),
		StrInputs: make(map[string]string),
		TaxYear:   taxYear,
		PriorYear: prior.TaxYear,
		StateCode: prior.State,
	}

	// Copy ALL prior UserInput values
	for k, v := range prior.Inputs {
		rf.Inputs[k] = v
	}
	for k, v := range prior.StrInputs {
		rf.StrInputs[k] = v
	}

	// Ensure ALL UserInput fields have entries in the inputs maps.
	// The solver's MissingInputs checks the numeric map for ALL UserInput fields,
	// so even string-valued fields need a 0 entry in the numeric map.
	for _, form := range registry.AllForms() {
		for _, field := range form.Fields {
			if field.Type != forms.UserInput {
				continue
			}
			key := forms.FieldKey(form.ID, field.Line)

			// Ensure numeric entry exists (solver requires this for all UserInput)
			if _, ok := rf.Inputs[key]; !ok {
				if v, ok := prior.Computed[key]; ok {
					rf.Inputs[key] = v
				} else {
					rf.Inputs[key] = 0
				}
			}

			// Ensure string entry exists for string-valued fields
			if field.ValueType == forms.StringValue || len(field.Options) > 0 {
				if _, ok := rf.StrInputs[key]; !ok {
					rf.StrInputs[key] = ""
				}
			}
		}
	}

	// Freeze a snapshot of the original prior-year input values
	rf.PriorInputs = make(map[string]float64, len(rf.Inputs))
	for k, v := range rf.Inputs {
		rf.PriorInputs[k] = v
	}
	rf.PriorStrIn = make(map[string]string, len(rf.StrInputs))
	for k, v := range rf.StrInputs {
		rf.PriorStrIn[k] = v
	}

	// Solve with new year tables
	if err := rf.ReSolve(); err != nil {
		return nil, fmt.Errorf("initial solve: %w", err)
	}

	// Solve with prior year tables for comparison
	priorComputed, err := graph.Solve(rf.Inputs, rf.StrInputs, rf.PriorYear)
	if err != nil {
		// Non-fatal: we can still show rollforward without comparison
		rf.PriorComputed = make(map[string]float64)
	} else {
		rf.PriorComputed = priorComputed
	}

	// Detect filing status for parameter comparison
	status := taxmath.Single
	if fs, ok := rf.StrInputs["1040:filing_status"]; ok {
		status = taxmath.FilingStatus(fs)
	}
	rf.ParamChanges = taxmath.CompareYearParameters(rf.PriorYear, rf.TaxYear, status)

	// Build field list and analyze changes
	rf.buildFields()
	rf.analyzeChanges()

	return rf, nil
}

// ReSolve runs the solver with current inputs and updates Computed.
func (rf *Rollforward) ReSolve() error {
	result, err := rf.Graph.Solve(rf.Inputs, rf.StrInputs, rf.TaxYear)
	if err != nil {
		return err
	}
	rf.Computed = result
	return nil
}

// UpdateInput sets a numeric input and re-solves. Returns the set of field keys
// whose computed values changed.
func (rf *Rollforward) UpdateInput(key string, value float64) (changed []string, err error) {
	oldComputed := make(map[string]float64, len(rf.Computed))
	for k, v := range rf.Computed {
		oldComputed[k] = v
	}

	rf.Inputs[key] = value
	if err := rf.ReSolve(); err != nil {
		return nil, err
	}

	// Find what changed
	for k, newVal := range rf.Computed {
		if oldVal, ok := oldComputed[k]; !ok || oldVal != newVal {
			changed = append(changed, k)
		}
	}

	rf.refreshFieldValues()
	return changed, nil
}

// UpdateStrInput sets a string input and re-solves.
func (rf *Rollforward) UpdateStrInput(key string, value string) (changed []string, err error) {
	oldComputed := make(map[string]float64, len(rf.Computed))
	for k, v := range rf.Computed {
		oldComputed[k] = v
	}

	rf.StrInputs[key] = value
	if err := rf.ReSolve(); err != nil {
		return nil, err
	}

	for k, newVal := range rf.Computed {
		if oldVal, ok := oldComputed[k]; !ok || oldVal != newVal {
			changed = append(changed, k)
		}
	}

	rf.refreshFieldValues()
	return changed, nil
}

// buildFields creates the ordered field list from registry.
func (rf *Rollforward) buildFields() {
	rf.Fields = nil

	// Collect forms sorted by jurisdiction then name
	allForms := rf.Registry.AllForms()
	sort.Slice(allForms, func(i, j int) bool {
		if allForms[i].Jurisdiction != allForms[j].Jurisdiction {
			return allForms[i].Jurisdiction < allForms[j].Jurisdiction
		}
		if allForms[i].QuestionOrder != allForms[j].QuestionOrder {
			return allForms[i].QuestionOrder < allForms[j].QuestionOrder
		}
		return string(allForms[i].ID) < string(allForms[j].ID)
	})

	for _, form := range allForms {
		for _, field := range form.Fields {
			key := forms.FieldKey(form.ID, field.Line)
			isStr := field.ValueType == forms.StringValue || len(field.Options) > 0

			rff := RollforwardField{
				Key:       key,
				Line:      field.Line,
				FormID:    form.ID,
				FormName:  form.Name,
				Label:     field.Label,
				FieldType: field.Type,
				IsString:  isStr,
				IsInteger: field.ValueType == forms.IntegerValue,
				Options:   field.Options,
			}

			rf.Fields = append(rf.Fields, rff)
		}
	}

	rf.refreshFieldValues()
}

// refreshFieldValues updates field values from the current Computed/Inputs maps.
func (rf *Rollforward) refreshFieldValues() {
	for i := range rf.Fields {
		f := &rf.Fields[i]
		if f.IsString {
			f.StrValue = rf.StrInputs[f.Key]
			f.PriorStr = rf.PriorStrIn[f.Key] // frozen prior-year value
		} else {
			f.Value = rf.Computed[f.Key]
			f.PriorValue = rf.PriorComputed[f.Key]
			f.Changed = f.Value != f.PriorValue
		}
	}
}

// analyzeChanges compares new-year vs prior-year computed values.
func (rf *Rollforward) analyzeChanges() {
	rf.Changes = nil

	// Build a reason map from parameter changes
	reasonMap := make(map[string]string)
	for _, pc := range rf.ParamChanges {
		reasonMap[pc.Category] = fmt.Sprintf("%s changed from $%.0f to $%.0f", pc.Name, pc.OldValue, pc.NewValue)
	}

	for i := range rf.Fields {
		f := &rf.Fields[i]

		if f.IsString {
			continue
		}
		if f.FieldType == forms.UserInput {
			continue // user inputs are the same, not "changed"
		}

		if !f.Changed {
			continue
		}

		delta := f.Value - f.PriorValue
		pctChange := 0.0
		if f.PriorValue != 0 {
			pctChange = delta / math.Abs(f.PriorValue)
		} else if f.Value != 0 {
			pctChange = 1.0
		}

		// Determine reason
		reason := ""
		key := f.Key
		if strings.Contains(key, "1040:12") || strings.Contains(key, "ca_540:18") {
			if r, ok := reasonMap["deduction"]; ok {
				reason = r
			}
		} else if strings.Contains(key, "1040:16") || strings.Contains(key, "ca_540:31") {
			if r, ok := reasonMap["bracket"]; ok {
				reason = r
			} else {
				reason = "Tax bracket thresholds adjusted for inflation"
			}
		}
		if reason == "" {
			reason = "Tax year parameter changes"
		}

		change := FieldChange{
			Key:           key,
			Label:         f.Label,
			FormName:      f.FormName,
			FieldType:     f.FieldType,
			OldValue:      f.PriorValue,
			NewValue:      f.Value,
			Delta:         delta,
			PercentChange: pctChange,
			Reason:        reason,
		}
		rf.Changes = append(rf.Changes, change)

		f.Flagged = true
		f.FlagReason = reason
	}

	// Also flag UserInput fields that are zero (might need real values)
	for i := range rf.Fields {
		f := &rf.Fields[i]
		if f.FieldType != forms.UserInput {
			continue
		}
		if !f.IsString && f.Value == 0 && f.Key != "1040:filing_status" {
			// Don't flag zero values for fields that are legitimately zero
			// Only flag if the field was also zero in prior year (likely unfilled)
			if rf.PriorComputed[f.Key] == 0 {
				continue
			}
			f.Flagged = true
			f.FlagReason = "Value is zero — may need updating"
		}
	}
}

// CountFlagged returns the number of flagged fields.
func (rf *Rollforward) CountFlagged() int {
	count := 0
	for _, f := range rf.Fields {
		if f.Flagged {
			count++
		}
	}
	return count
}

// SaveState persists the current rollforward state.
func (rf *Rollforward) SaveState() error {
	ret := state.NewTaxReturn(rf.TaxYear, rf.StateCode)
	ret.Inputs = rf.Inputs
	ret.StrInputs = rf.StrInputs
	ret.Computed = rf.Computed
	ret.Complete = true
	return state.Save(state.DefaultStorePath(), ret)
}
