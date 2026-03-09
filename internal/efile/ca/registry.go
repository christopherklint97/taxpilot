package ca

import "taxpilot/internal/forms"

// CACoveredFormIDs returns the set of CA FormIDs that have CA FTB XML builders.
// Used by registry sync tests to verify all CA forms have XML coverage.
func CACoveredFormIDs() map[forms.FormID]bool {
	return map[forms.FormID]bool{
		forms.FormCA540:      true,
		forms.FormScheduleCA: true,
	}
}
