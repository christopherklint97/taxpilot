package pdf

import "taxpilot/internal/forms"

// PDFMappedFormIDs returns the set of FormIDs that have PDF field mappings.
// Used by registry sync tests to verify all forms have PDF coverage.
func PDFMappedFormIDs() map[forms.FormID]bool {
	filler := NewFiller("")
	filler.RegisterForm(Federal1040Mappings())
	filler.RegisterForm(ScheduleAMappings())
	filler.RegisterForm(ScheduleBMappings())
	filler.RegisterForm(ScheduleCMappings())
	filler.RegisterForm(ScheduleDMappings())
	filler.RegisterForm(Form8949Mappings())
	filler.RegisterForm(Schedule1Mappings())
	filler.RegisterForm(Schedule2Mappings())
	filler.RegisterForm(Schedule3Mappings())
	filler.RegisterForm(ScheduleSEMappings())
	filler.RegisterForm(Form8995Mappings())
	filler.RegisterForm(Form8889Mappings())
	filler.RegisterForm(CA540Mappings())
	filler.RegisterForm(ScheduleCAMappings())
	filler.RegisterForm(Form3514Mappings())
	filler.RegisterForm(Form3853Mappings())

	result := make(map[forms.FormID]bool)
	for id := range filler.configs {
		result[id] = true
	}
	return result
}
