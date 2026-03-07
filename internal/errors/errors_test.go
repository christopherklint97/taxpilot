package errors

import (
	"fmt"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Error() method tests
// ---------------------------------------------------------------------------

func TestUnsupportedError_Error(t *testing.T) {
	e := &UnsupportedError{
		Situation:  "Non-resident alien filing",
		Reason:     "TaxPilot only supports resident filers",
		Suggestion: "Consult a CPA or use Form 1040-NR",
	}
	got := e.Error()
	if !strings.Contains(got, "Non-resident alien filing") {
		t.Errorf("expected situation in error message, got: %s", got)
	}
	if !strings.Contains(got, "TaxPilot only supports resident filers") {
		t.Errorf("expected reason in error message, got: %s", got)
	}
}

func TestIncompleteError_Error(t *testing.T) {
	t.Run("with message", func(t *testing.T) {
		e := &IncompleteError{
			MissingFields: []string{"1040:ssn"},
			Message:       "Custom message",
		}
		if e.Error() != "Custom message" {
			t.Errorf("expected custom message, got: %s", e.Error())
		}
	})

	t.Run("without message", func(t *testing.T) {
		e := &IncompleteError{
			MissingFields: []string{"1040:ssn", "1040:first_name"},
		}
		got := e.Error()
		if !strings.Contains(got, "2 required field(s)") {
			t.Errorf("expected field count in error, got: %s", got)
		}
		if !strings.Contains(got, "1040:ssn") {
			t.Errorf("expected field name in error, got: %s", got)
		}
	})
}

func TestConformityError_Error(t *testing.T) {
	e := ConformityError{
		FederalField: "form_8889:13",
		CAField:      "ca_schedule_ca:13_col_b",
		Situation:    "HSA deduction",
		Message:      "California does not conform to federal HSA deduction.",
	}
	got := e.Error()
	if !strings.Contains(got, "HSA deduction") {
		t.Errorf("expected situation in error, got: %s", got)
	}
	if !strings.Contains(got, "California does not conform") {
		t.Errorf("expected message in error, got: %s", got)
	}
}

func TestCPAReferralError_Error(t *testing.T) {
	e := &CPAReferralError{
		Situation:  "Partnership income",
		Reason:     "K-1 allocation rules are complex",
		Complexity: "very_high",
	}
	got := e.Error()
	if !strings.Contains(got, "very_high") {
		t.Errorf("expected complexity in error, got: %s", got)
	}
	if !strings.Contains(got, "Partnership income") {
		t.Errorf("expected situation in error, got: %s", got)
	}
}

// ---------------------------------------------------------------------------
// CheckUnsupported tests
// ---------------------------------------------------------------------------

func TestCheckUnsupported(t *testing.T) {
	tests := []struct {
		name      string
		results   map[string]float64
		strInputs map[string]string
		wantCount int
		wantSit   string // substring expected in at least one error
	}{
		{
			name:      "clean single filer",
			results:   map[string]float64{"1040:11": 75000},
			strInputs: map[string]string{"1040:filing_status": "single", "1040:state": "CA"},
			wantCount: 0,
		},
		{
			name:      "MFS with itemized deductions",
			results:   map[string]float64{"schedule_a:17": 20000},
			strInputs: map[string]string{"1040:filing_status": "mfs"},
			wantCount: 1,
			wantSit:   "Married Filing Separately",
		},
		{
			name:      "foreign earned income",
			results:   map[string]float64{"schedule_1:8_foreign_income": 50000},
			strInputs: map[string]string{"1040:filing_status": "single"},
			wantCount: 1,
			wantSit:   "Foreign earned income",
		},
		{
			name:      "high income AMT risk",
			results:   map[string]float64{"1040:11": 600000, "schedule_a:5d": 10000},
			strInputs: map[string]string{"1040:filing_status": "single"},
			wantCount: 1,
			wantSit:   "Alternative Minimum Tax",
		},
		{
			name:      "unsupported state",
			results:   map[string]float64{},
			strInputs: map[string]string{"1040:state": "NY"},
			wantCount: 1,
			wantSit:   "NY",
		},
		{
			name:      "CA state is supported",
			results:   map[string]float64{},
			strInputs: map[string]string{"1040:state": "CA"},
			wantCount: 0,
		},
		{
			name: "multiple unsupported situations",
			results: map[string]float64{
				"schedule_1:8_foreign_income": 30000,
				"1040:11":                     700000,
				"schedule_a:5d":               10000,
			},
			strInputs: map[string]string{"1040:filing_status": "single", "1040:state": "TX"},
			wantCount: 3,
		},
		{
			name:      "nil inputs",
			results:   nil,
			strInputs: nil,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := CheckUnsupported(tt.results, tt.strInputs)
			if len(errs) != tt.wantCount {
				t.Errorf("got %d errors, want %d; errors: %v", len(errs), tt.wantCount, errs)
			}
			if tt.wantSit != "" {
				found := false
				for _, e := range errs {
					if strings.Contains(e.Error(), tt.wantSit) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error containing %q, got: %v", tt.wantSit, errs)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// CheckIncomplete tests
// ---------------------------------------------------------------------------

func TestCheckIncomplete(t *testing.T) {
	tests := []struct {
		name         string
		results      map[string]float64
		strInputs    map[string]string
		wantNil      bool
		wantMissing  int
		wantContains string
	}{
		{
			name: "complete return",
			results: map[string]float64{
				"1040:9": 75000, "1040:11": 75000,
				"1040:15": 60000, "1040:24": 8000,
			},
			strInputs: map[string]string{
				"1040:ssn": "123-45-6789", "1040:first_name": "John",
				"1040:last_name": "Doe", "1040:filing_status": "single",
			},
			wantNil: true,
		},
		{
			name:    "all missing",
			results: map[string]float64{},
			strInputs: map[string]string{},
			wantNil:     false,
			wantMissing: 8, // 4 string + 4 numeric
		},
		{
			name: "missing SSN only",
			results: map[string]float64{
				"1040:9": 75000, "1040:11": 75000,
				"1040:15": 60000, "1040:24": 8000,
			},
			strInputs: map[string]string{
				"1040:first_name": "John", "1040:last_name": "Doe",
				"1040:filing_status": "single",
			},
			wantNil:      false,
			wantMissing:  1,
			wantContains: "1040:ssn",
		},
		{
			name:        "nil inputs",
			results:     nil,
			strInputs:   nil,
			wantNil:     false,
			wantMissing: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckIncomplete(tt.results, tt.strInputs)
			if tt.wantNil {
				if err != nil {
					t.Errorf("expected nil, got: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if len(err.MissingFields) != tt.wantMissing {
				t.Errorf("got %d missing fields, want %d; fields: %v",
					len(err.MissingFields), tt.wantMissing, err.MissingFields)
			}
			if tt.wantContains != "" {
				found := false
				for _, f := range err.MissingFields {
					if f == tt.wantContains {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected %q in missing fields, got: %v", tt.wantContains, err.MissingFields)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// CheckCAConformity tests
// ---------------------------------------------------------------------------

func TestCheckCAConformity(t *testing.T) {
	tests := []struct {
		name      string
		results   map[string]float64
		wantCount int
		wantSit   string
	}{
		{
			name:      "no conformity issues",
			results:   map[string]float64{"1040:11": 75000},
			wantCount: 0,
		},
		{
			name:      "HSA deduction",
			results:   map[string]float64{"form_8889:13": 3850},
			wantCount: 1,
			wantSit:   "HSA deduction",
		},
		{
			name:      "QBI deduction",
			results:   map[string]float64{"form_8995:15": 12000},
			wantCount: 1,
			wantSit:   "QBI deduction",
		},
		{
			name:      "Social Security benefits",
			results:   map[string]float64{"1040:6b": 18000},
			wantCount: 1,
			wantSit:   "Social Security",
		},
		{
			name:      "tax-exempt interest",
			results:   map[string]float64{"1040:2a": 5000},
			wantCount: 1,
			wantSit:   "Tax-exempt interest",
		},
		{
			name: "HSA and QBI together",
			results: map[string]float64{
				"form_8889:13": 3850,
				"form_8995:15": 12000,
			},
			wantCount: 2,
		},
		{
			name: "all conformity issues",
			results: map[string]float64{
				"form_8889:13": 3850,
				"form_8995:15": 12000,
				"1040:6b":      18000,
				"1040:2a":      5000,
			},
			wantCount: 4,
		},
		{
			name:      "nil results",
			results:   nil,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := CheckCAConformity(tt.results)
			if len(errs) != tt.wantCount {
				t.Errorf("got %d conformity errors, want %d; errors: %v",
					len(errs), tt.wantCount, errs)
			}
			if tt.wantSit != "" {
				found := false
				for _, e := range errs {
					if strings.Contains(e.Situation, tt.wantSit) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected conformity error for %q", tt.wantSit)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// CheckComplexity tests
// ---------------------------------------------------------------------------

func TestCheckComplexity(t *testing.T) {
	tests := []struct {
		name           string
		results        map[string]float64
		strInputs      map[string]string
		wantNil        bool
		wantComplexity string
		wantSit        string
	}{
		{
			name:      "simple return",
			results:   map[string]float64{"1040:11": 75000},
			strInputs: map[string]string{"1040:filing_status": "single"},
			wantNil:   true,
		},
		{
			name:           "K-1 partnership income",
			results:        map[string]float64{"k1:1:income": 50000},
			strInputs:      map[string]string{},
			wantNil:        false,
			wantComplexity: "very_high",
			wantSit:        "Partnership",
		},
		{
			name:           "partnership via schedule 1",
			results:        map[string]float64{"schedule_1:5_partnership": 30000},
			strInputs:      map[string]string{},
			wantNil:        false,
			wantComplexity: "very_high",
			wantSit:        "Partnership",
		},
		{
			name:           "estate/trust income",
			results:        map[string]float64{"schedule_1:5_estate_trust": 10000},
			strInputs:      map[string]string{},
			wantNil:        false,
			wantComplexity: "high",
			wantSit:        "Estate",
		},
		{
			name:           "foreign tax credit over 300",
			results:        map[string]float64{"schedule_3:1_foreign_tax": 500},
			strInputs:      map[string]string{},
			wantNil:        false,
			wantComplexity: "high",
			wantSit:        "Foreign tax credit",
		},
		{
			name:      "foreign tax credit under 300",
			results:   map[string]float64{"schedule_3:1_foreign_tax": 200},
			strInputs: map[string]string{},
			wantNil:   true,
		},
		{
			name:           "high income with itemized deductions",
			results:        map[string]float64{"1040:11": 600000, "schedule_a:17": 35000},
			strInputs:      map[string]string{},
			wantNil:        false,
			wantComplexity: "high",
			wantSit:        "Alternative Minimum Tax",
		},
		{
			name:      "high income WITHOUT itemized — no AMT trigger",
			results:   map[string]float64{"1040:11": 600000},
			strInputs: map[string]string{},
			wantNil:   true,
		},
		{
			name:      "nil inputs",
			results:   nil,
			strInputs: nil,
			wantNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckComplexity(tt.results, tt.strInputs)
			if tt.wantNil {
				if err != nil {
					t.Errorf("expected nil, got: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected CPAReferralError, got nil")
			}
			if err.Complexity != tt.wantComplexity {
				t.Errorf("got complexity %q, want %q", err.Complexity, tt.wantComplexity)
			}
			if tt.wantSit != "" && !strings.Contains(err.Situation, tt.wantSit) {
				t.Errorf("expected situation containing %q, got %q", tt.wantSit, err.Situation)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// FormatForUser tests
// ---------------------------------------------------------------------------

func TestFormatForUser(t *testing.T) {
	t.Run("UnsupportedError", func(t *testing.T) {
		e := &UnsupportedError{
			Situation:  "Foreign income",
			Reason:     "Form 2555 not supported",
			Suggestion: "Use a CPA",
		}
		got := FormatForUser(e)
		if !strings.Contains(got, "NOT SUPPORTED") {
			t.Errorf("expected NOT SUPPORTED header, got: %s", got)
		}
		if !strings.Contains(got, "What to do: Use a CPA") {
			t.Errorf("expected suggestion, got: %s", got)
		}
	})

	t.Run("IncompleteError", func(t *testing.T) {
		e := &IncompleteError{
			MissingFields: []string{"1040:ssn", "1040:first_name"},
			Message:       "Return incomplete",
		}
		got := FormatForUser(e)
		if !strings.Contains(got, "MISSING INFORMATION") {
			t.Errorf("expected MISSING INFORMATION header, got: %s", got)
		}
		if !strings.Contains(got, "Social Security Number") {
			t.Errorf("expected friendly field name, got: %s", got)
		}
	})

	t.Run("ConformityError value type", func(t *testing.T) {
		e := ConformityError{
			Situation: "HSA deduction",
			Message:   "CA does not conform",
		}
		got := FormatForUser(e)
		if !strings.Contains(got, "CALIFORNIA ADJUSTMENT") {
			t.Errorf("expected CALIFORNIA ADJUSTMENT header, got: %s", got)
		}
	})

	t.Run("ConformityError pointer type", func(t *testing.T) {
		e := &ConformityError{
			Situation: "QBI deduction",
			Message:   "CA does not allow QBI",
		}
		got := FormatForUser(e)
		if !strings.Contains(got, "CALIFORNIA ADJUSTMENT") {
			t.Errorf("expected CALIFORNIA ADJUSTMENT header, got: %s", got)
		}
	})

	t.Run("CPAReferralError", func(t *testing.T) {
		e := &CPAReferralError{
			Situation:  "K-1 income",
			Reason:     "Complex allocation",
			Complexity: "very_high",
		}
		got := FormatForUser(e)
		if !strings.Contains(got, "PROFESSIONAL HELP RECOMMENDED") {
			t.Errorf("expected PROFESSIONAL HELP header, got: %s", got)
		}
		if !strings.Contains(got, "CPA or enrolled agent") {
			t.Errorf("expected CPA recommendation, got: %s", got)
		}
	})

	t.Run("generic error", func(t *testing.T) {
		e := fmt.Errorf("something went wrong")
		got := FormatForUser(e)
		if got != "something went wrong" {
			t.Errorf("expected raw message, got: %s", got)
		}
	})
}

// ---------------------------------------------------------------------------
// FormatAllIssues tests
// ---------------------------------------------------------------------------

func TestFormatAllIssues(t *testing.T) {
	t.Run("no issues", func(t *testing.T) {
		got := FormatAllIssues(nil)
		if !strings.Contains(got, "No issues found") {
			t.Errorf("expected no issues message, got: %s", got)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		got := FormatAllIssues([]error{})
		if !strings.Contains(got, "No issues found") {
			t.Errorf("expected no issues message, got: %s", got)
		}
	})

	t.Run("mixed errors", func(t *testing.T) {
		errs := []error{
			&UnsupportedError{
				Situation: "Foreign income",
				Reason:    "Not supported",
			},
			&IncompleteError{
				MissingFields: []string{"1040:ssn"},
				Message:       "Missing SSN",
			},
			ConformityError{
				Situation: "HSA",
				Message:   "Add-back required",
			},
			&CPAReferralError{
				Situation:  "K-1",
				Reason:     "Complex",
				Complexity: "very_high",
			},
			fmt.Errorf("generic error"),
		}

		got := FormatAllIssues(errs)
		if !strings.Contains(got, "5 issue(s)") {
			t.Errorf("expected 5 issues count, got: %s", got)
		}
		if !strings.Contains(got, "Unsupported Situations") {
			t.Errorf("expected unsupported section, got: %s", got)
		}
		if !strings.Contains(got, "Missing Information") {
			t.Errorf("expected missing info section, got: %s", got)
		}
		if !strings.Contains(got, "California Adjustments") {
			t.Errorf("expected conformity section, got: %s", got)
		}
		if !strings.Contains(got, "Professional Help Recommended") {
			t.Errorf("expected referral section, got: %s", got)
		}
		if !strings.Contains(got, "Other Issues") {
			t.Errorf("expected other section, got: %s", got)
		}
	})

	t.Run("only conformity errors", func(t *testing.T) {
		errs := []error{
			ConformityError{Situation: "HSA", Message: "Add-back"},
			ConformityError{Situation: "QBI", Message: "Add-back"},
		}
		got := FormatAllIssues(errs)
		if !strings.Contains(got, "2 issue(s)") {
			t.Errorf("expected 2 issues, got: %s", got)
		}
		if strings.Contains(got, "Unsupported") {
			t.Errorf("should not have unsupported section, got: %s", got)
		}
	})
}

// ---------------------------------------------------------------------------
// friendlyFieldName tests
// ---------------------------------------------------------------------------

func TestFriendlyFieldName(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{"1040:ssn", "Social Security Number"},
		{"1040:first_name", "First Name"},
		{"1040:11", "Adjusted Gross Income (Form 1040, Line 11)"},
		{"unknown:field", "unknown:field"}, // falls through to raw key
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := friendlyFieldName(tt.key)
			if got != tt.want {
				t.Errorf("friendlyFieldName(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}
