package forms

import (
	"fmt"
	"strings"
	"sync"
)

// FormConstructor is a function that returns a FormDef. Form packages
// call RegisterForm in their init() functions to add constructors to
// the global auto-registration list.
type FormConstructor func() *FormDef

var (
	autoFormsMu sync.Mutex
	autoForms   []FormConstructor
)

// RegisterForm adds a FormDef constructor to the global auto-registration
// list. It is intended to be called from init() functions in form packages.
func RegisterForm(fn FormConstructor) {
	autoFormsMu.Lock()
	defer autoFormsMu.Unlock()
	autoForms = append(autoForms, fn)
}

// AutoRegisteredForms returns FormDefs from all constructors that have
// been registered via RegisterForm. Each constructor is called once.
func AutoRegisteredForms() []*FormDef {
	autoFormsMu.Lock()
	defer autoFormsMu.Unlock()
	result := make([]*FormDef, 0, len(autoForms))
	for _, fn := range autoForms {
		result = append(result, fn())
	}
	return result
}

// NewRegistryFromAutoForms creates a Registry pre-populated with every
// form that was auto-registered via init() + RegisterForm.
func NewRegistryFromAutoForms() *Registry {
	reg := NewRegistry()
	for _, form := range AutoRegisteredForms() {
		reg.Register(form)
	}
	return reg
}

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

	field := form.FieldByLine(line)
	if field == nil {
		return nil, nil, fmt.Errorf("field %q not found in form %q", line, formID)
	}
	return form, field, nil
}

// AllForms returns all registered form definitions.
func (r *Registry) AllForms() []*FormDef {
	result := make([]*FormDef, 0, len(r.forms))
	for _, f := range r.forms {
		result = append(result, f)
	}
	return result
}

// ValidateFieldDefs checks all registered forms for common errors.
func (r *Registry) ValidateFieldDefs() []error {
	var errs []error
	for _, form := range r.AllForms() {
		for _, field := range form.Fields {
			if field.Type == UserInput && field.Compute != nil {
				errs = append(errs, fmt.Errorf("form %s field %s: UserInput should not have Compute", form.ID, field.Line))
			}
		}
	}
	return errs
}

// ValidateFederalRefs checks that all FederalRef fields reference forms with
// Jurisdiction == Federal. Returns a list of errors for any violations.
func (r *Registry) ValidateFederalRefs() []error {
	var errs []error
	for _, form := range r.AllForms() {
		for _, field := range form.Fields {
			if field.Type != FederalRef {
				continue
			}
			for _, dep := range field.DependsOn {
				parts := strings.SplitN(dep, ":", 2)
				if len(parts) != 2 {
					errs = append(errs, fmt.Errorf("form %s field %s: FederalRef dependency %q has invalid format", form.ID, field.Line, dep))
					continue
				}
				depFormID := FormID(parts[0])
				depForm, ok := r.forms[depFormID]
				if !ok {
					// The referenced form may not be registered (e.g., input forms).
					// Only validate forms that are present in the registry.
					continue
				}
				if depForm.Jurisdiction != Federal {
					errs = append(errs, fmt.Errorf("form %s field %s: FederalRef dependency %q references non-federal form %s (jurisdiction %d)", form.ID, field.Line, dep, depFormID, depForm.Jurisdiction))
				}
			}
		}
	}
	return errs
}
