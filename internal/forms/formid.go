package forms

// FormID is a typed identifier for a tax form. Using a named type instead of
// raw strings enables compile-time checks for typos and missing registrations.
type FormID string

// --- Input Forms ---

const (
	FormW2      FormID = "w2"
	Form1099INT FormID = "1099int"
	Form1099DIV FormID = "1099div"
	Form1099NEC FormID = "1099nec"
	Form1099B   FormID = "1099b"
)

// --- Federal Forms ---

const (
	FormF1040      FormID = "1040"
	FormScheduleA  FormID = "schedule_a"
	FormScheduleB  FormID = "schedule_b"
	FormScheduleC  FormID = "schedule_c"
	FormScheduleD  FormID = "schedule_d"
	FormSchedule1  FormID = "schedule_1"
	FormSchedule2  FormID = "schedule_2"
	FormSchedule3  FormID = "schedule_3"
	FormScheduleSE FormID = "schedule_se"
	FormF8949      FormID = "form_8949"
	FormF8995      FormID = "form_8995"
	FormF8889      FormID = "form_8889"
	FormF2555      FormID = "form_2555"
	FormF1116      FormID = "form_1116"
	FormF8938      FormID = "form_8938"
	FormF8833      FormID = "form_8833"
)

// --- California State Forms ---

const (
	FormCA540      FormID = "ca_540"
	FormCA540NR    FormID = "ca_540nr"
	FormScheduleCA FormID = "ca_schedule_ca"
	FormF3514      FormID = "form_3514"
	FormF3853      FormID = "form_3853"
)

// AllFormIDs returns every known FormID. This is used for registration
// validation to ensure no form is forgotten.
func AllFormIDs() []FormID {
	return []FormID{
		// Input
		FormW2, Form1099INT, Form1099DIV, Form1099NEC, Form1099B,
		// Federal
		FormF1040, FormScheduleA, FormScheduleB, FormScheduleC, FormScheduleD,
		FormSchedule1, FormSchedule2, FormSchedule3, FormScheduleSE,
		FormF8949, FormF8995, FormF8889,
		FormF2555, FormF1116, FormF8938, FormF8833,
		// CA
		FormCA540, FormScheduleCA, FormF3514, FormF3853,
	}
}

// InputFormIDs returns the FormIDs that are input forms (W-2, 1099s).
// These use instance prefixes like "w2:1:wages".
func InputFormIDs() []FormID {
	return []FormID{FormW2, Form1099INT, Form1099DIV, Form1099NEC, Form1099B}
}

// InputFormPrefixes returns the string prefixes used for instance re-keying.
func InputFormPrefixes() []string {
	ids := InputFormIDs()
	prefixes := make([]string, len(ids))
	for i, id := range ids {
		prefixes[i] = string(id) + ":"
	}
	return prefixes
}
