package forms

import (
	"fmt"
	"strings"
)

// Registry holds all registered forms and provides lookup.
type Registry struct {
	forms map[FormID]*FormDef
}

// NewRegistry creates an empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		forms: make(map[FormID]*FormDef),
	}
}

// Register adds a form definition to the registry.
func (r *Registry) Register(form *FormDef) {
	r.forms[form.ID] = form
}

// Get returns the form with the given ID, or false if not found.
func (r *Registry) Get(formID FormID) (*FormDef, bool) {
	f, ok := r.forms[formID]
	return f, ok
}

// GetField looks up a field by its fully qualified key ("form_id:line").
// Returns the parent form and field definition, or an error if not found.
func (r *Registry) GetField(key string) (*FormDef, *FieldDef, error) {
	parts := strings.SplitN(key, ":", 2)
	if len(parts) != 2 {
		return nil, nil, fmt.Errorf("invalid field key %q: expected form_id:line", key)
	}
	formID, line := FormID(parts[0]), parts[1]

	form, ok := r.forms[formID]
	if !ok {
		return nil, nil, fmt.Errorf("form %q not found", formID)
	}

	for i := range form.Fields {
		if form.Fields[i].Line == line {
			return form, &form.Fields[i], nil
		}
	}
	return nil, nil, fmt.Errorf("field %q not found in form %q", line, formID)
}

// AllForms returns all registered form definitions.
func (r *Registry) AllForms() []*FormDef {
	result := make([]*FormDef, 0, len(r.forms))
	for _, f := range r.forms {
		result = append(result, f)
	}
	return result
}
