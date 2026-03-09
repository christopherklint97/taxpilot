package mef

import (
	"bytes"
	"testing"
)

// TestXMLDeterminism_SelfEmployed verifies that the self-employed scenario
// (with Schedules C, SE, 1, 2) produces byte-identical XML across runs.
func TestXMLDeterminism_SelfEmployed(t *testing.T) {
	results, strInputs := selfEmployedScenario()

	first, err := GenerateReturn(results, strInputs, 2025)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	for i := 0; i < 20; i++ {
		got, err := GenerateReturn(results, strInputs, 2025)
		if err != nil {
			t.Fatalf("call %d failed: %v", i, err)
		}
		if !bytes.Equal(first, got) {
			t.Fatalf("call %d produced different XML output (non-deterministic)\nfirst length: %d\ngot length: %d",
				i, len(first), len(got))
		}
	}
}

// TestXMLDeterminism_MultipleW2s verifies deterministic ordering of W-2
// elements when multiple W-2 instances exist.
func TestXMLDeterminism_MultipleW2s(t *testing.T) {
	results := map[string]float64{
		"w2:1:wages":                50000,
		"w2:1:federal_tax_withheld": 5000,
		"w2:2:wages":                30000,
		"w2:2:federal_tax_withheld": 3000,
		"w2:3:wages":                20000,
		"w2:3:federal_tax_withheld": 2000,

		"1040:1a":  100000,
		"1040:1z":  100000,
		"1040:9":   100000,
		"1040:11":  100000,
		"1040:14":  15000,
		"1040:15":  85000,
		"1040:16":  14000,
		"1040:22":  14000,
		"1040:24":  14000,
		"1040:25a": 10000,
		"1040:25d": 10000,
		"1040:33":  10000,
		"1040:37":  4000,
	}
	strInputs := map[string]string{
		"1040:filing_status": "single",
		"1040:first_name":    "Multi",
		"1040:last_name":     "Worker",
		"1040:ssn":           "111-22-3333",
		"w2:1:employer_name": "Alpha Corp",
		"w2:1:employer_ein":  "11-1111111",
		"w2:2:employer_name": "Beta LLC",
		"w2:2:employer_ein":  "22-2222222",
		"w2:3:employer_name": "Gamma Inc",
		"w2:3:employer_ein":  "33-3333333",
	}

	first, err := GenerateReturn(results, strInputs, 2025)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	for i := 0; i < 20; i++ {
		got, err := GenerateReturn(results, strInputs, 2025)
		if err != nil {
			t.Fatalf("call %d failed: %v", i, err)
		}
		if !bytes.Equal(first, got) {
			t.Fatalf("call %d produced different XML for multiple W-2s (non-deterministic)", i)
		}
	}
}

// expatScenario returns solver results for an expat filer with Forms 2555,
// 1116, 8938, and 8833.
func expatScenario() (map[string]float64, map[string]string) {
	results := map[string]float64{
		"w2:1:wages":                0,
		"w2:1:federal_tax_withheld": 0,

		// Form 2555
		"form_2555:qualifying_days":       365,
		"form_2555:foreign_earned_income":  120000,
		"form_2555:exclusion_limit":        126500,
		"form_2555:foreign_income_excl":    120000,
		"form_2555:housing_expenses":       30000,
		"form_2555:housing_base":           20232,
		"form_2555:housing_max":            37960,
		"form_2555:housing_exclusion":      9768,
		"form_2555:housing_deduction":      0,
		"form_2555:total_exclusion":        120000,
		"form_2555:prorated_exclusion":     120000,

		// Form 1116
		"form_1116:line_7":      5000,
		"form_1116:line_15":     1500,
		"form_1116:line_21":     1200,
		"form_1116:line_22":     1200,
		"form_1116:carryforward": 300,

		// Form 8938
		"form_8938:max_value_accounts":      250000,
		"form_8938:year_end_accounts":       200000,
		"form_8938:total_max_value":         250000,
		"form_8938:total_year_end_value":    200000,
		"form_8938:filing_required":         1,

		// Form 8833
		"form_8833:treaty_amount":   5000,
		"form_8833:treaty_claimed":  1,

		// Schedule 1
		"schedule_1:8d":  120000,
		"schedule_1:10":  120000,

		// Schedule 3
		"schedule_3:1":  1200,
		"schedule_3:8":  1200,

		// 1040
		"1040:1a":  0,
		"1040:1z":  0,
		"1040:9":   120000,
		"1040:10":  120000,
		"1040:11":  0,
		"1040:14":  15000,
		"1040:15":  0,
		"1040:16":  0,
		"1040:20":  1200,
		"1040:22":  0,
		"1040:24":  0,
		"1040:25d": 0,
		"1040:33":  0,
		"1040:34":  0,
		"1040:37":  0,
	}
	strInputs := map[string]string{
		"1040:filing_status":          "single",
		"1040:first_name":             "Expat",
		"1040:last_name":              "Abroad",
		"1040:ssn":                    "555-66-7777",
		"form_2555:foreign_country":   "Germany",
		"form_2555:qualifying_test":   "bona_fide_residence",
		"form_1116:foreign_country":   "Germany",
		"form_1116:category":          "general",
		"form_8938:lives_abroad":      "yes",
		"form_8833:treaty_country":    "Germany",
		"form_8833:treaty_article":    "Article 15",
		"form_8833:irc_provision":     "IRC 894(a)",
	}
	return results, strInputs
}

// TestXMLDeterminism_ExpatForms verifies that expat forms (2555, 1116, 8938,
// 8833) produce byte-identical XML across runs.
func TestXMLDeterminism_ExpatForms(t *testing.T) {
	results, strInputs := expatScenario()

	first, err := GenerateReturn(results, strInputs, 2025)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	for i := 0; i < 20; i++ {
		got, err := GenerateReturn(results, strInputs, 2025)
		if err != nil {
			t.Fatalf("call %d failed: %v", i, err)
		}
		if !bytes.Equal(first, got) {
			t.Fatalf("call %d produced different XML for expat scenario (non-deterministic)", i)
		}
	}
}

// TestXMLDeterminism_CrossTaxYear verifies that the same inputs with
// different tax years produce deterministic (but different) outputs.
func TestXMLDeterminism_CrossTaxYear(t *testing.T) {
	results, strInputs := simpleW2Scenario()

	for _, year := range []int{2024, 2025, 2026} {
		first, err := GenerateReturn(results, strInputs, year)
		if err != nil {
			t.Fatalf("year %d first call failed: %v", year, err)
		}
		for i := 0; i < 10; i++ {
			got, err := GenerateReturn(results, strInputs, year)
			if err != nil {
				t.Fatalf("year %d call %d failed: %v", year, i, err)
			}
			if !bytes.Equal(first, got) {
				t.Fatalf("year %d call %d produced different XML (non-deterministic)", year, i)
			}
		}
	}
}
