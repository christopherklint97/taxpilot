package state

// StateFormSet defines the interface for a state's form collection.
type StateFormSet interface {
	// Code returns the two-letter state abbreviation (e.g., "CA").
	Code() string

	// Name returns the full state name (e.g., "California").
	Name() string

	// RequiredForms returns the form IDs that are always needed for a state filing.
	RequiredForms() []string

	// ConditionalForms returns a map of form ID to a human-readable description
	// of the condition under which the form is required.
	ConditionalForms() map[string]string
}
