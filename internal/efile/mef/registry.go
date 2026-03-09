package mef

import "taxpilot/internal/forms"

// MeFCoveredFormIDs returns the set of federal FormIDs that have MeF XML builders.
// Used by registry sync tests to verify all federal forms have XML coverage.
func MeFCoveredFormIDs() map[forms.FormID]bool {
	return map[forms.FormID]bool{
		forms.FormF1040:      true,
		forms.FormScheduleA:  true,
		forms.FormScheduleB:  true,
		forms.FormScheduleC:  true,
		forms.FormScheduleD:  true,
		forms.FormSchedule1:  true,
		forms.FormSchedule2:  true,
		forms.FormSchedule3:  true,
		forms.FormScheduleSE: true,
		forms.FormF8889:      true,
		forms.FormF8949:      true,
		forms.FormF8995:      true,
		forms.FormF2555:      true,
		forms.FormF1116:      true,
		forms.FormF8938:      true,
		forms.FormF8833:      true,
		forms.FormW2:         true, // W-2 XML builder exists
	}
}
