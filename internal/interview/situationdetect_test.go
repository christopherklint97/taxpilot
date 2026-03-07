package interview

import (
	"testing"
)

func TestScreeningDefaultCount(t *testing.T) {
	// We expect 6 active screening questions (rental income is commented out)
	expected := 6
	if got := len(DefaultScreeningQuestions); got != expected {
		t.Errorf("expected %d screening questions, got %d", expected, got)
	}
}

func TestScreeningQuestionsHaveRequiredFields(t *testing.T) {
	for _, sq := range DefaultScreeningQuestions {
		if sq.ID == "" {
			t.Error("screening question has empty ID")
		}
		if sq.Question == "" {
			t.Errorf("screening question %q has empty Question", sq.ID)
		}
		if sq.HelpText == "" {
			t.Errorf("screening question %q has empty HelpText", sq.ID)
		}
		if sq.OnYes.ID == "" {
			t.Errorf("screening question %q has empty OnYes.ID", sq.ID)
		}
		if len(sq.OnYes.FormsNeeded) == 0 {
			t.Errorf("screening question %q has no FormsNeeded", sq.ID)
		}
	}
}

func TestScreeningEvaluateNone(t *testing.T) {
	answers := map[string]bool{
		"has_self_employment": false,
		"has_capital_gains":   false,
	}
	situations := EvaluateScreening(answers)
	if len(situations) != 0 {
		t.Errorf("expected 0 situations, got %d", len(situations))
	}
}

func TestScreeningEvaluateSelfEmployment(t *testing.T) {
	answers := map[string]bool{
		"has_self_employment": true,
	}
	situations := EvaluateScreening(answers)
	if len(situations) != 1 {
		t.Fatalf("expected 1 situation, got %d", len(situations))
	}
	if situations[0].ID != "self_employed" {
		t.Errorf("expected situation ID 'self_employed', got %q", situations[0].ID)
	}
	if len(situations[0].FormsNeeded) != 2 {
		t.Errorf("expected 2 forms needed, got %d", len(situations[0].FormsNeeded))
	}
}

func TestScreeningEvaluateMultiple(t *testing.T) {
	answers := map[string]bool{
		"has_self_employment":    true,
		"has_capital_gains":      true,
		"has_interest_income":    false,
		"has_dividend_income":    false,
		"has_hsa":               true,
		"has_itemized_deductions": false,
	}
	situations := EvaluateScreening(answers)
	if len(situations) != 3 {
		t.Fatalf("expected 3 situations, got %d", len(situations))
	}

	ids := make(map[string]bool)
	for _, s := range situations {
		ids[s.ID] = true
	}
	for _, expected := range []string{"self_employed", "capital_gains", "hsa"} {
		if !ids[expected] {
			t.Errorf("expected situation %q to be present", expected)
		}
	}
}

func TestScreeningEvaluateAllYes(t *testing.T) {
	answers := make(map[string]bool)
	for _, sq := range DefaultScreeningQuestions {
		answers[sq.ID] = true
	}
	situations := EvaluateScreening(answers)
	if len(situations) != len(DefaultScreeningQuestions) {
		t.Errorf("expected %d situations for all-yes, got %d", len(DefaultScreeningQuestions), len(situations))
	}
}

func TestScreeningHSAHasCANote(t *testing.T) {
	for _, sq := range DefaultScreeningQuestions {
		if sq.ID == "has_hsa" {
			if sq.CANote == "" {
				t.Error("HSA screening question should have a CA note")
			}
			return
		}
	}
	t.Error("has_hsa screening question not found")
}

func TestAutoDetectSituationsEmpty(t *testing.T) {
	prior := PriorYearData{
		FormsPresent:  []string{},
		NumericValues: map[string]float64{},
	}
	detected := AutoDetectSituations(prior)
	if len(detected) != 0 {
		t.Errorf("expected 0 detected situations for empty prior data, got %d", len(detected))
	}
}

func TestAutoDetectSituationsFromForms(t *testing.T) {
	prior := PriorYearData{
		FormsPresent: []string{"schedule_c", "schedule_d", "1099int", "form_8889"},
	}
	detected := AutoDetectSituations(prior)

	expected := map[string]bool{
		"has_self_employment": true,
		"has_capital_gains":   true,
		"has_interest_income": true,
		"has_hsa":             true,
	}
	for key, val := range expected {
		if detected[key] != val {
			t.Errorf("expected %q=%v, got %v", key, val, detected[key])
		}
	}

	// These should NOT be detected
	for _, key := range []string{"has_dividend_income", "has_itemized_deductions"} {
		if detected[key] {
			t.Errorf("did not expect %q to be detected", key)
		}
	}
}

func TestAutoDetectSituationsFromNumericValues(t *testing.T) {
	prior := PriorYearData{
		FormsPresent: []string{},
		NumericValues: map[string]float64{
			"form_8889:2":  3500,
			"schedule_a:17": 25000,
			"schedule_c:31": 50000,
		},
	}
	detected := AutoDetectSituations(prior)

	for _, key := range []string{"has_hsa", "has_itemized_deductions", "has_self_employment"} {
		if !detected[key] {
			t.Errorf("expected %q to be detected from numeric values", key)
		}
	}
}

func TestAutoDetectSituationsDividends(t *testing.T) {
	prior := PriorYearData{
		FormsPresent: []string{"1099div"},
	}
	detected := AutoDetectSituations(prior)
	if !detected["has_dividend_income"] {
		t.Error("expected has_dividend_income to be detected")
	}
}

func TestAutoDetectSituationsScheduleA(t *testing.T) {
	prior := PriorYearData{
		FormsPresent: []string{"schedule_a"},
	}
	detected := AutoDetectSituations(prior)
	if !detected["has_itemized_deductions"] {
		t.Error("expected has_itemized_deductions to be detected")
	}
}
