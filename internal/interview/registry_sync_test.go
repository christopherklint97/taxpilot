package interview

import (
	"testing"

	"taxpilot/internal/forms"
)

// TestSetupRegistryCoversAllFormIDs verifies that every FormID in
// AllFormIDs() is registered by SetupRegistry(). This catches the common
// bug where a new form is added to the constants but not registered.
func TestSetupRegistryCoversAllFormIDs(t *testing.T) {
	reg := SetupRegistry()

	for _, id := range forms.AllFormIDs() {
		if _, ok := reg.Get(id); !ok {
			t.Errorf("FormID %s is in AllFormIDs() but not registered in SetupRegistry()", id)
		}
	}
}

// TestSetupRegistryNoExtraForms verifies that every form registered by
// SetupRegistry() has a corresponding FormID constant in AllFormIDs().
func TestSetupRegistryNoExtraForms(t *testing.T) {
	reg := SetupRegistry()
	known := make(map[forms.FormID]bool)
	for _, id := range forms.AllFormIDs() {
		known[id] = true
	}

	for _, form := range reg.AllForms() {
		if !known[form.ID] {
			t.Errorf("form %s is registered in SetupRegistry() but has no FormID constant", form.ID)
		}
	}
}
