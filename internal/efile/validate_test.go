package efile

import (
	"testing"
)

// baseValidInputs returns a minimal valid set of inputs that pass all federal rules.
func baseValidInputs() (map[string]float64, map[string]string) {
	results := map[string]float64{
		"1040:9":      50000,
		"1040:11":     50000,
		"1040:15":     40000,
		"1040:24":     6000,
		"1040:25a":    5000,
		"1040:25b":    500,
		"1040:25d":    5500,
		"1040:34":     0,
		"1040:37":     500,
		"w2:1:wages":  50000,
		"schedule_c:7": 0,
		"schedule_c:28": 0,
		"schedule_c:31": 0,
		"schedule_a:1":  0,
		"schedule_a:5d": 0,
		"schedule_a:12": 0,
	}
	strInputs := map[string]string{
		"1040:ssn":           "123-45-6789",
		"1040:filing_status": "single",
		"1040:first_name":    "John",
		"1040:last_name":     "Doe",
	}
	return results, strInputs
}

func findResult(report ValidationReport, code string) *ValidationResult {
	for i, r := range report.Results {
		if r.Code == code {
			return &report.Results[i]
		}
	}
	return nil
}

func TestValidateReturn_AllValid(t *testing.T) {
	results, strInputs := baseValidInputs()
	report := ValidateReturn(results, strInputs, 2025)
	if !report.IsValid {
		t.Errorf("expected valid report, got %d results", len(report.Results))
		for _, r := range report.Results {
			t.Errorf("  %s (%d): %s", r.Code, r.Severity, r.Message)
		}
	}
}

func TestValidateReturn_FederalRules(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		severity Severity
		modify   func(results map[string]float64, strInputs map[string]string)
	}{
		{
			name:     "R0001 SSN missing",
			code:     "R0001",
			severity: SeverityError,
			modify: func(r map[string]float64, s map[string]string) {
				delete(s, "1040:ssn")
			},
		},
		{
			name:     "R0001 SSN empty",
			code:     "R0001",
			severity: SeverityError,
			modify: func(r map[string]float64, s map[string]string) {
				s["1040:ssn"] = ""
			},
		},
		{
			name:     "R0002 filing status missing",
			code:     "R0002",
			severity: SeverityError,
			modify: func(r map[string]float64, s map[string]string) {
				delete(s, "1040:filing_status")
			},
		},
		{
			name:     "R0002 filing status invalid",
			code:     "R0002",
			severity: SeverityError,
			modify: func(r map[string]float64, s map[string]string) {
				s["1040:filing_status"] = "married"
			},
		},
		{
			name:     "R0003 first name missing",
			code:     "R0003",
			severity: SeverityError,
			modify: func(r map[string]float64, s map[string]string) {
				delete(s, "1040:first_name")
			},
		},
		{
			name:     "R0004 last name missing",
			code:     "R0004",
			severity: SeverityError,
			modify: func(r map[string]float64, s map[string]string) {
				delete(s, "1040:last_name")
			},
		},
		{
			name:     "R0005 negative total income",
			code:     "R0005",
			severity: SeverityError,
			modify: func(r map[string]float64, s map[string]string) {
				r["1040:9"] = -1000
			},
		},
		{
			name:     "R0006 taxable income exceeds total income",
			code:     "R0006",
			severity: SeverityError,
			modify: func(r map[string]float64, s map[string]string) {
				r["1040:9"] = 50000
				r["1040:15"] = 60000
			},
		},
		{
			name:     "R0007 negative total tax",
			code:     "R0007",
			severity: SeverityError,
			modify: func(r map[string]float64, s map[string]string) {
				r["1040:24"] = -100
			},
		},
		{
			name:     "R0008 withholding mismatch",
			code:     "R0008",
			severity: SeverityError,
			modify: func(r map[string]float64, s map[string]string) {
				r["1040:25a"] = 5000
				r["1040:25b"] = 500
				r["1040:25d"] = 7000 // should be 5500
			},
		},
		{
			name:     "R0009 both refund and owed positive",
			code:     "R0009",
			severity: SeverityError,
			modify: func(r map[string]float64, s map[string]string) {
				r["1040:34"] = 1000
				r["1040:37"] = 500
			},
		},
		{
			name:     "R0010 no income source",
			code:     "R0010",
			severity: SeverityError,
			modify: func(r map[string]float64, s map[string]string) {
				delete(r, "w2:1:wages")
				r["schedule_c:31"] = 0
			},
		},
		{
			name:     "W0001 high charitable donations",
			code:     "W0001",
			severity: SeverityWarning,
			modify: func(r map[string]float64, s map[string]string) {
				r["1040:11"] = 100000
				r["schedule_a:12"] = 70000 // 70% of AGI
			},
		},
		{
			name:     "W0002 high medical expenses",
			code:     "W0002",
			severity: SeverityWarning,
			modify: func(r map[string]float64, s map[string]string) {
				r["1040:11"] = 100000
				r["schedule_a:1"] = 25000 // 25% of AGI
			},
		},
		{
			name:     "W0003 high business expense ratio",
			code:     "W0003",
			severity: SeverityWarning,
			modify: func(r map[string]float64, s map[string]string) {
				r["schedule_c:7"] = 100000
				r["schedule_c:28"] = 85000 // 85%
				r["schedule_c:31"] = 15000
			},
		},
		{
			name:     "W0004 SALT at cap single",
			code:     "W0004",
			severity: SeverityInfo,
			modify: func(r map[string]float64, s map[string]string) {
				r["schedule_a:5d"] = 10000
				s["1040:filing_status"] = "single"
			},
		},
		{
			name:     "W0004 SALT at cap MFS",
			code:     "W0004",
			severity: SeverityInfo,
			modify: func(r map[string]float64, s map[string]string) {
				r["schedule_a:5d"] = 5000
				s["1040:filing_status"] = "mfs"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, strInputs := baseValidInputs()
			tt.modify(results, strInputs)
			report := ValidateReturn(results, strInputs, 2025)
			res := findResult(report, tt.code)
			if res == nil {
				t.Fatalf("expected rule %s to trigger, but it did not", tt.code)
			}
			if res.Severity != tt.severity {
				t.Errorf("expected severity %d, got %d", tt.severity, res.Severity)
			}
		})
	}
}

func TestValidateReturn_R0008_WithinTolerance(t *testing.T) {
	results, strInputs := baseValidInputs()
	// Off by $0.50 — should be OK
	results["1040:25a"] = 5000
	results["1040:25b"] = 500
	results["1040:25d"] = 5500.50
	report := ValidateReturn(results, strInputs, 2025)
	if findResult(report, "R0008") != nil {
		t.Error("R0008 should not trigger when difference is within $1")
	}
}

func TestValidateReturn_R0010_ScheduleC_Sufficient(t *testing.T) {
	results, strInputs := baseValidInputs()
	delete(results, "w2:1:wages")
	results["schedule_c:31"] = 50000
	report := ValidateReturn(results, strInputs, 2025)
	if findResult(report, "R0010") != nil {
		t.Error("R0010 should not trigger when Schedule C has positive net profit")
	}
}

func TestValidateReturn_R0010_MultipleW2s(t *testing.T) {
	results, strInputs := baseValidInputs()
	delete(results, "w2:1:wages")
	results["w2:2:wages"] = 30000
	report := ValidateReturn(results, strInputs, 2025)
	if findResult(report, "R0010") != nil {
		t.Error("R0010 should not trigger when any w2:*:wages > 0")
	}
}

func TestValidateReturn_W0001_NotTriggered(t *testing.T) {
	results, strInputs := baseValidInputs()
	results["1040:11"] = 100000
	results["schedule_a:12"] = 50000 // exactly 50%, under 60%
	report := ValidateReturn(results, strInputs, 2025)
	if findResult(report, "W0001") != nil {
		t.Error("W0001 should not trigger when charitable is 50% of AGI")
	}
}

func TestValidateReturn_ValidFilingStatuses(t *testing.T) {
	for _, status := range []string{"single", "mfj", "mfs", "hoh", "qss"} {
		t.Run(status, func(t *testing.T) {
			results, strInputs := baseValidInputs()
			strInputs["1040:filing_status"] = status
			report := ValidateReturn(results, strInputs, 2025)
			if findResult(report, "R0002") != nil {
				t.Errorf("R0002 should not trigger for valid status %q", status)
			}
		})
	}
}

// --- CA validation tests ---

func baseValidCAInputs() (map[string]float64, map[string]string) {
	results, strInputs := baseValidInputs()
	results["ca_540:17"] = 50000
	results["ca_540:40"] = 3000
	results["ca_540:75"] = 0
	results["ca_540:81"] = 200
	return results, strInputs
}

func TestValidateCAReturn_AllValid(t *testing.T) {
	results, strInputs := baseValidCAInputs()
	report := ValidateCAReturn(results, strInputs, 2025)
	if !report.IsValid {
		t.Errorf("expected valid CA report, got %d results", len(report.Results))
		for _, r := range report.Results {
			t.Errorf("  %s (%d): %s", r.Code, r.Severity, r.Message)
		}
	}
}

func TestValidateCAReturn_Rules(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		severity Severity
		modify   func(results map[string]float64, strInputs map[string]string)
	}{
		{
			name:     "R1001 CA AGI missing",
			code:     "R1001",
			severity: SeverityError,
			modify: func(r map[string]float64, s map[string]string) {
				delete(r, "ca_540:17")
			},
		},
		{
			name:     "R1002 negative CA tax",
			code:     "R1002",
			severity: SeverityError,
			modify: func(r map[string]float64, s map[string]string) {
				r["ca_540:40"] = -500
			},
		},
		{
			name:     "R1003 both CA refund and owed positive",
			code:     "R1003",
			severity: SeverityError,
			modify: func(r map[string]float64, s map[string]string) {
				r["ca_540:75"] = 1000
				r["ca_540:81"] = 500
			},
		},
		{
			name:     "W1001 CA AGI differs from federal by > 100k",
			code:     "W1001",
			severity: SeverityWarning,
			modify: func(r map[string]float64, s map[string]string) {
				r["1040:11"] = 50000
				r["ca_540:17"] = 200000
			},
		},
		{
			name:     "W1002 CA effective rate > 13.3%",
			code:     "W1002",
			severity: SeverityWarning,
			modify: func(r map[string]float64, s map[string]string) {
				r["ca_540:17"] = 100000
				r["ca_540:40"] = 14000 // 14%
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, strInputs := baseValidCAInputs()
			tt.modify(results, strInputs)
			report := ValidateCAReturn(results, strInputs, 2025)
			res := findResult(report, tt.code)
			if res == nil {
				t.Fatalf("expected rule %s to trigger, but it did not", tt.code)
			}
			if res.Severity != tt.severity {
				t.Errorf("expected severity %d, got %d", tt.severity, res.Severity)
			}
		})
	}
}

func TestValidateCAReturn_W1001_NotTriggered(t *testing.T) {
	results, strInputs := baseValidCAInputs()
	results["1040:11"] = 50000
	results["ca_540:17"] = 120000 // diff is 70k, under 100k
	report := ValidateCAReturn(results, strInputs, 2025)
	if findResult(report, "W1001") != nil {
		t.Error("W1001 should not trigger when AGI difference is under $100,000")
	}
}

func TestValidateCAReturn_W1002_NotTriggered(t *testing.T) {
	results, strInputs := baseValidCAInputs()
	results["ca_540:17"] = 100000
	results["ca_540:40"] = 10000 // 10%, under 13.3%
	report := ValidateCAReturn(results, strInputs, 2025)
	if findResult(report, "W1002") != nil {
		t.Error("W1002 should not trigger when effective rate is under 13.3%")
	}
}

// --- ValidateFull tests ---

func TestValidateFull_FederalOnly(t *testing.T) {
	results, strInputs := baseValidInputs()
	report := ValidateFull(results, strInputs, 2025, false)
	if !report.IsValid {
		t.Error("expected valid report for federal-only validation")
	}
}

func TestValidateFull_WithCA(t *testing.T) {
	results, strInputs := baseValidCAInputs()
	report := ValidateFull(results, strInputs, 2025, true)
	if !report.IsValid {
		t.Errorf("expected valid full report, got %d results", len(report.Results))
		for _, r := range report.Results {
			t.Errorf("  %s (%d): %s", r.Code, r.Severity, r.Message)
		}
	}
}

func TestValidateFull_MergesErrors(t *testing.T) {
	results, strInputs := baseValidCAInputs()
	// Trigger federal error
	delete(strInputs, "1040:ssn")
	// Trigger CA error
	delete(results, "ca_540:17")

	report := ValidateFull(results, strInputs, 2025, true)
	if report.IsValid {
		t.Error("expected invalid report when both federal and CA have errors")
	}
	if findResult(report, "R0001") == nil {
		t.Error("expected federal R0001 in merged results")
	}
	if findResult(report, "R1001") == nil {
		t.Error("expected CA R1001 in merged results")
	}
}

func TestValidateFull_WarningsDoNotBlockValidity(t *testing.T) {
	results, strInputs := baseValidCAInputs()
	// Trigger a warning only
	results["1040:11"] = 100000
	results["schedule_a:1"] = 25000 // W0002: medical > 20% AGI
	report := ValidateFull(results, strInputs, 2025, true)
	if !report.IsValid {
		t.Error("warnings should not block validity")
	}
	if findResult(report, "W0002") == nil {
		t.Error("expected W0002 warning in results")
	}
}

func TestValidateReturn_NilMaps(t *testing.T) {
	report := ValidateReturn(nil, nil, 2025)
	if report.IsValid {
		t.Error("expected invalid report with nil inputs")
	}
	// Should have R0001, R0002, R0003, R0004, R0010 at minimum
	if findResult(report, "R0001") == nil {
		t.Error("expected R0001 with nil strInputs")
	}
	if findResult(report, "R0010") == nil {
		t.Error("expected R0010 with nil results")
	}
}
