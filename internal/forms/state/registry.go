package state

// StateRegistry holds all registered state form sets.
type StateRegistry struct {
	states map[string]StateFormSet
}

// NewStateRegistry creates an empty state registry.
func NewStateRegistry() *StateRegistry {
	return &StateRegistry{
		states: make(map[string]StateFormSet),
	}
}

// Register adds a state form set to the registry.
func (r *StateRegistry) Register(sfs StateFormSet) {
	r.states[sfs.Code()] = sfs
}

// Get returns the state form set for the given code, or nil.
func (r *StateRegistry) Get(code string) StateFormSet {
	return r.states[code]
}

// All returns all registered state form sets.
func (r *StateRegistry) All() []StateFormSet {
	result := make([]StateFormSet, 0, len(r.states))
	for _, sfs := range r.states {
		result = append(result, sfs)
	}
	return result
}

// Codes returns all registered state codes.
func (r *StateRegistry) Codes() []string {
	result := make([]string, 0, len(r.states))
	for code := range r.states {
		result = append(result, code)
	}
	return result
}
