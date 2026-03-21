package interview

import "testing"

func TestIRCRefsPopulated(t *testing.T) {
	tests := []struct {
		fieldKey  string
		expectIRC string
	}{
		{"1040:filing_status", "IRC \u00a71"},
		{"w2:wages", "IRC \u00a761(a)"},
		{"w2:federal_tax_withheld", "IRC \u00a73402"},
		{"w2:ss_wages", "IRC \u00a73121"},
		{"w2:medicare_wages", "IRC \u00a73101"},
		{"schedule_a:1", "IRC \u00a7213(a)"},
		{"schedule_a:5a", "IRC \u00a7164"},
		{"schedule_a:8a", "IRC \u00a7163(h)"},
		{"schedule_a:12", "IRC \u00a7170"},
		{"1099int:interest_income", "IRC \u00a761(a)(4)"},
		{"1099int:us_savings_bond_interest", "IRC \u00a7103"},
		{"1099int:tax_exempt_interest", "IRC \u00a7103"},
		{"1099div:ordinary_dividends", "IRC \u00a761(a)(7)"},
		{"1099div:qualified_dividends", "IRC \u00a71(h)(11)"},
		{"1099div:section_199a_dividends", "IRC \u00a7199A"},
		{"form_8889:2", "IRC \u00a7223"},
		{"1099b:proceeds", "IRC \u00a71001"},
		{"1099b:cost_basis", "IRC \u00a71001"},
		{"schedule_3:10", "IRC \u00a76654"},
	}

	for _, tt := range tests {
		t.Run(tt.fieldKey, func(t *testing.T) {
			cp := GetContextualPrompt(tt.fieldKey, "fallback", "")
			if cp.IRCRef != tt.expectIRC {
				t.Errorf("GetContextualPrompt(%q).IRCRef = %q, want %q", tt.fieldKey, cp.IRCRef, tt.expectIRC)
			}
		})
	}
}

func TestCARefOnlyWithCA(t *testing.T) {
	// Fields with CARef defined
	fieldsWithCARef := []struct {
		fieldKey  string
		expectRef string
	}{
		{"1040:filing_status", "R&TC \u00a717042"},
		{"w2:wages", "R&TC \u00a717071"},
		{"schedule_a:1", "R&TC \u00a717201"},
		{"schedule_a:5a", "R&TC \u00a717220 (not allowed)"},
		{"form_8889:2", "R&TC \u00a717215 (not allowed)"},
		{"1099div:qualified_dividends", "R&TC \u00a717041 (taxed as ordinary)"},
		{"1099b:proceeds", "R&TC \u00a718031"},
	}

	for _, tt := range fieldsWithCARef {
		t.Run(tt.fieldKey+"_with_CA", func(t *testing.T) {
			cp := GetContextualPrompt(tt.fieldKey, "fallback", "CA")
			if cp.CARef != tt.expectRef {
				t.Errorf("GetContextualPrompt(%q, CA).CARef = %q, want %q", tt.fieldKey, cp.CARef, tt.expectRef)
			}
		})

		t.Run(tt.fieldKey+"_without_CA", func(t *testing.T) {
			cp := GetContextualPrompt(tt.fieldKey, "fallback", "")
			if cp.CARef != "" {
				t.Errorf("GetContextualPrompt(%q, \"\").CARef = %q, want empty", tt.fieldKey, cp.CARef)
			}
		})

		t.Run(tt.fieldKey+"_other_state", func(t *testing.T) {
			cp := GetContextualPrompt(tt.fieldKey, "fallback", "NY")
			if cp.CARef != "" {
				t.Errorf("GetContextualPrompt(%q, NY).CARef = %q, want empty", tt.fieldKey, cp.CARef)
			}
		})
	}
}

func TestContextualPromptFallback(t *testing.T) {
	cp := GetContextualPrompt("nonexistent:field", "original prompt text", "CA")
	if cp.Prompt != "original prompt text" {
		t.Errorf("Fallback Prompt = %q, want %q", cp.Prompt, "original prompt text")
	}
	if cp.HelpText != "" {
		t.Errorf("Fallback HelpText = %q, want empty", cp.HelpText)
	}
	if cp.CANote != "" {
		t.Errorf("Fallback CANote = %q, want empty", cp.CANote)
	}
	if cp.IRCRef != "" {
		t.Errorf("Fallback IRCRef = %q, want empty", cp.IRCRef)
	}
	if cp.CARef != "" {
		t.Errorf("Fallback CARef = %q, want empty", cp.CARef)
	}
}

func TestIRCRefAlwaysIncluded(t *testing.T) {
	// IRC ref should be included regardless of state code
	cp := GetContextualPrompt("w2:wages", "fallback", "")
	if cp.IRCRef == "" {
		t.Error("IRCRef should be populated even without state code")
	}

	cp = GetContextualPrompt("w2:wages", "fallback", "CA")
	if cp.IRCRef == "" {
		t.Error("IRCRef should be populated with CA state code")
	}

	cp = GetContextualPrompt("w2:wages", "fallback", "NY")
	if cp.IRCRef == "" {
		t.Error("IRCRef should be populated with NY state code")
	}
}
