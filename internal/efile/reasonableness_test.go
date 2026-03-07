package efile

import (
	"testing"
)

// baseReasonablenessInputs returns inputs that pass all reasonableness checks.
func baseReasonablenessInputs() (map[string]float64, map[string]string) {
	results := map[string]float64{
		"1040:9":      50000, // total income
		"1040:11":     50000, // AGI (no adjustments)
		"1040:12":     14600, // standard deduction
		"1040:15":     35400, // taxable income = AGI - deduction
		"1040:24":     4000,  // total tax
		"1040:25a":    4500,  // withholding
		"1040:25b":    0,
		"1040:25d":    4500,
		"1040:33":     4500,  // total payments
		"1040:34":     500,   // refund = payments - tax
		"1040:37":     0,
		"w2:1:wages":  50000,
		"schedule_c:7":  0,
		"schedule_c:28": 0,
		"schedule_c:31": 0,
	}
	strInputs := map[string]string{
		"1040:ssn":           "123-45-6789",
		"1040:filing_status": "single",
		"1040:first_name":    "John",
		"1040:last_name":     "Doe",
	}
	return results, strInputs
}

func TestReasonablenessCheck_AllClean(t *testing.T) {
	results, strInputs := baseReasonablenessInputs()
	report := ReasonablenessCheck(results, strInputs, 2025, false)
	if !report.IsValid {
		t.Errorf("expected valid report, got %d results", len(report.Results))
		for _, r := range report.Results {
			t.Errorf("  %s (%d): %s", r.Code, r.Severity, r.Message)
		}
	}
	if len(report.Results) != 0 {
		t.Errorf("expected zero results for clean inputs, got %d", len(report.Results))
		for _, r := range report.Results {
			t.Errorf("  %s: %s", r.Code, r.Message)
		}
	}
}

func TestReasonablenessCheck_NeverError(t *testing.T) {
	// Even with bad data, reasonableness checks should never produce SeverityError
	results := map[string]float64{
		"1040:9":          100000,
		"1040:11":         50000, // AGI doesn't match — triggers RC001
		"1040:12":         14600,
		"1040:15":         99999, // doesn't match — triggers RC002
		"1040:24":         50000,
		"1040:33":         10000,
		"1040:34":         9999, // doesn't match — triggers RC003
		"1040:37":         0,
		"schedule_c:7":    10000,
		"schedule_c:30":   5000,  // 50% home office — triggers RC005
		"schedule_c:31":   5000,
		"schedule_2:4":    0,      // no SE tax — triggers RC006
		"form_8889:2":     10000,  // over limit — triggers RC007
		"schedule_d:21":   -5000,  // over limit — triggers RC008
		"schedule_3:8":    60000,  // > total tax — triggers RC009
		"w2:1:wages":      50000,
	}
	strInputs := map[string]string{
		"1040:filing_status": "single",
	}
	report := ReasonablenessCheck(results, strInputs, 2025, false)
	for _, r := range report.Results {
		if r.Severity == SeverityError {
			t.Errorf("reasonableness check %s produced SeverityError: %s", r.Code, r.Message)
		}
	}
	// Must always be valid since no errors
	if !report.IsValid {
		t.Error("expected IsValid=true since reasonableness checks never produce errors")
	}
}

func TestReasonablenessCheck_CrossChecks(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		severity Severity
		modify   func(results map[string]float64, strInputs map[string]string)
	}{
		{
			name:     "RC001 AGI mismatch with adjustments",
			code:     "RC001",
			severity: SeverityWarning,
			modify: func(r map[string]float64, s map[string]string) {
				r["1040:9"] = 60000
				r["schedule_1:26"] = 5000
				r["1040:11"] = 50000 // should be 55000
			},
		},
		{
			name:     "RC001 AGI mismatch without adjustments",
			code:     "RC001",
			severity: SeverityWarning,
			modify: func(r map[string]float64, s map[string]string) {
				r["1040:9"] = 60000
				r["1040:11"] = 50000 // should be 60000
			},
		},
		{
			name:     "RC002 taxable income mismatch",
			code:     "RC002",
			severity: SeverityWarning,
			modify: func(r map[string]float64, s map[string]string) {
				r["1040:11"] = 50000
				r["1040:12"] = 14600
				r["1040:15"] = 30000 // should be 35400
			},
		},
		{
			name:     "RC002 taxable income mismatch with itemized",
			code:     "RC002",
			severity: SeverityWarning,
			modify: func(r map[string]float64, s map[string]string) {
				r["1040:11"] = 50000
				r["1040:12"] = 14600
				r["schedule_a:17"] = 20000
				r["1040:15"] = 25000 // should be 30000
			},
		},
		{
			name:     "RC003 refund mismatch",
			code:     "RC003",
			severity: SeverityWarning,
			modify: func(r map[string]float64, s map[string]string) {
				r["1040:33"] = 6000
				r["1040:24"] = 4000
				r["1040:34"] = 500 // should be 2000
				r["1040:37"] = 0
			},
		},
		{
			name:     "RC004 amount owed mismatch",
			code:     "RC004",
			severity: SeverityWarning,
			modify: func(r map[string]float64, s map[string]string) {
				r["1040:33"] = 3000
				r["1040:24"] = 4000
				r["1040:34"] = 0
				r["1040:37"] = 500 // should be 1000
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, strInputs := baseReasonablenessInputs()
			tt.modify(results, strInputs)
			report := ReasonablenessCheck(results, strInputs, 2025, false)
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

func TestReasonablenessCheck_CrossChecks_NotTriggered(t *testing.T) {
	tests := []struct {
		name   string
		code   string
		modify func(results map[string]float64, strInputs map[string]string)
	}{
		{
			name: "RC001 AGI matches with adjustments",
			code: "RC001",
			modify: func(r map[string]float64, s map[string]string) {
				r["1040:9"] = 60000
				r["schedule_1:26"] = 5000
				r["1040:11"] = 55000
			},
		},
		{
			name: "RC001 AGI matches within tolerance",
			code: "RC001",
			modify: func(r map[string]float64, s map[string]string) {
				r["1040:9"] = 50000
				r["1040:11"] = 50000.50 // within $1
			},
		},
		{
			name: "RC002 taxable income matches",
			code: "RC002",
			modify: func(r map[string]float64, s map[string]string) {
				r["1040:11"] = 50000
				r["1040:12"] = 14600
				r["1040:15"] = 35400
			},
		},
		{
			name: "RC003 refund matches",
			code: "RC003",
			modify: func(r map[string]float64, s map[string]string) {
				r["1040:33"] = 6000
				r["1040:24"] = 4000
				r["1040:34"] = 2000
				r["1040:37"] = 0
			},
		},
		{
			name: "RC004 amount owed matches",
			code: "RC004",
			modify: func(r map[string]float64, s map[string]string) {
				r["1040:33"] = 3000
				r["1040:24"] = 4000
				r["1040:34"] = 0
				r["1040:37"] = 1000
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, strInputs := baseReasonablenessInputs()
			tt.modify(results, strInputs)
			report := ReasonablenessCheck(results, strInputs, 2025, false)
			if findResult(report, tt.code) != nil {
				t.Errorf("expected rule %s NOT to trigger", tt.code)
			}
		})
	}
}

func TestReasonablenessCheck_AuditTriggers(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		severity Severity
		modify   func(results map[string]float64, strInputs map[string]string)
	}{
		{
			name:     "RC005 home office > 30% of business income",
			code:     "RC005",
			severity: SeverityWarning,
			modify: func(r map[string]float64, s map[string]string) {
				r["schedule_c:7"] = 100000
				r["schedule_c:30"] = 35000 // 35%
			},
		},
		{
			name:     "RC006 SE income but no SE tax",
			code:     "RC006",
			severity: SeverityWarning,
			modify: func(r map[string]float64, s map[string]string) {
				r["schedule_c:31"] = 5000
				r["schedule_2:4"] = 0
			},
		},
		{
			name:     "RC007 HSA over single limit",
			code:     "RC007",
			severity: SeverityWarning,
			modify: func(r map[string]float64, s map[string]string) {
				r["form_8889:2"] = 5000 // over $4300
				s["1040:filing_status"] = "single"
			},
		},
		{
			name:     "RC007 HSA over family limit",
			code:     "RC007",
			severity: SeverityWarning,
			modify: func(r map[string]float64, s map[string]string) {
				r["form_8889:2"] = 9000 // over $8550
				s["1040:filing_status"] = "mfj"
			},
		},
		{
			name:     "RC008 capital losses exceed single/mfj limit",
			code:     "RC008",
			severity: SeverityWarning,
			modify: func(r map[string]float64, s map[string]string) {
				r["schedule_d:21"] = -5000 // exceeds $3000
				s["1040:filing_status"] = "single"
			},
		},
		{
			name:     "RC008 capital losses exceed mfs limit",
			code:     "RC008",
			severity: SeverityWarning,
			modify: func(r map[string]float64, s map[string]string) {
				r["schedule_d:21"] = -2000 // exceeds $1500 for mfs
				s["1040:filing_status"] = "mfs"
			},
		},
		{
			name:     "RC009 estimated payments exceed total tax",
			code:     "RC009",
			severity: SeverityInfo,
			modify: func(r map[string]float64, s map[string]string) {
				r["schedule_3:8"] = 10000
				r["1040:24"] = 4000
			},
		},
		{
			name:     "RC010 effective rate > 37%",
			code:     "RC010",
			severity: SeverityWarning,
			modify: func(r map[string]float64, s map[string]string) {
				r["1040:11"] = 100000
				r["1040:24"] = 40000 // 40%
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, strInputs := baseReasonablenessInputs()
			tt.modify(results, strInputs)
			report := ReasonablenessCheck(results, strInputs, 2025, false)
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

func TestReasonablenessCheck_AuditTriggers_NotTriggered(t *testing.T) {
	tests := []struct {
		name   string
		code   string
		modify func(results map[string]float64, strInputs map[string]string)
	}{
		{
			name: "RC005 home office <= 30%",
			code: "RC005",
			modify: func(r map[string]float64, s map[string]string) {
				r["schedule_c:7"] = 100000
				r["schedule_c:30"] = 25000 // 25%
			},
		},
		{
			name: "RC006 SE income with SE tax",
			code: "RC006",
			modify: func(r map[string]float64, s map[string]string) {
				r["schedule_c:31"] = 5000
				r["schedule_2:4"] = 707
			},
		},
		{
			name: "RC006 SE income <= $400",
			code: "RC006",
			modify: func(r map[string]float64, s map[string]string) {
				r["schedule_c:31"] = 400
				r["schedule_2:4"] = 0
			},
		},
		{
			name: "RC007 HSA within single limit",
			code: "RC007",
			modify: func(r map[string]float64, s map[string]string) {
				r["form_8889:2"] = 4000
				s["1040:filing_status"] = "single"
			},
		},
		{
			name: "RC007 HSA within family limit",
			code: "RC007",
			modify: func(r map[string]float64, s map[string]string) {
				r["form_8889:2"] = 8000
				s["1040:filing_status"] = "mfj"
			},
		},
		{
			name: "RC008 capital losses within limit",
			code: "RC008",
			modify: func(r map[string]float64, s map[string]string) {
				r["schedule_d:21"] = -2500 // within $3000
			},
		},
		{
			name: "RC008 capital losses within mfs limit",
			code: "RC008",
			modify: func(r map[string]float64, s map[string]string) {
				r["schedule_d:21"] = -1200 // within $1500
				s["1040:filing_status"] = "mfs"
			},
		},
		{
			name: "RC009 estimated payments <= total tax",
			code: "RC009",
			modify: func(r map[string]float64, s map[string]string) {
				r["schedule_3:8"] = 3000
				r["1040:24"] = 4000
			},
		},
		{
			name: "RC010 effective rate <= 37%",
			code: "RC010",
			modify: func(r map[string]float64, s map[string]string) {
				r["1040:11"] = 100000
				r["1040:24"] = 35000 // 35%
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, strInputs := baseReasonablenessInputs()
			tt.modify(results, strInputs)
			report := ReasonablenessCheck(results, strInputs, 2025, false)
			if findResult(report, tt.code) != nil {
				t.Errorf("expected rule %s NOT to trigger", tt.code)
			}
		})
	}
}

// --- CA reasonableness tests ---

func baseReasonablenessCAInputs() (map[string]float64, map[string]string) {
	results, strInputs := baseReasonablenessInputs()
	results["ca_540:17"] = 50000  // CA AGI same as federal
	results["ca_540:19"] = 35400  // CA taxable income
	results["ca_540:31"] = 2000   // CA tax
	return results, strInputs
}

func TestReasonablenessCheck_CA_AllClean(t *testing.T) {
	results, strInputs := baseReasonablenessCAInputs()
	report := ReasonablenessCheck(results, strInputs, 2025, true)
	if len(report.Results) != 0 {
		t.Errorf("expected zero results for clean CA inputs, got %d", len(report.Results))
		for _, r := range report.Results {
			t.Errorf("  %s: %s", r.Code, r.Message)
		}
	}
}

func TestReasonablenessCheck_CA_Rules(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		severity Severity
		modify   func(results map[string]float64, strInputs map[string]string)
	}{
		{
			name:     "RC011 CA AGI differs > 20% from federal",
			code:     "RC011",
			severity: SeverityInfo,
			modify: func(r map[string]float64, s map[string]string) {
				r["1040:11"] = 50000
				r["ca_540:17"] = 70000 // 40% difference, > $5000
			},
		},
		{
			name:     "RC012 CA tax exceeds max rate",
			code:     "RC012",
			severity: SeverityWarning,
			modify: func(r map[string]float64, s map[string]string) {
				r["ca_540:19"] = 100000
				r["ca_540:31"] = 15000 // 15% > 13.3%
			},
		},
		{
			name:     "RC012 CA tax with mental health surcharge",
			code:     "RC012",
			severity: SeverityWarning,
			modify: func(r map[string]float64, s map[string]string) {
				r["ca_540:19"] = 2000000
				// Max: 13.3% * 2M + 1% * 1M = 266000 + 10000 = 276000
				r["ca_540:31"] = 280000
			},
		},
		{
			name:     "RC013 HSA without CA add-back",
			code:     "RC013",
			severity: SeverityInfo,
			modify: func(r map[string]float64, s map[string]string) {
				r["form_8889:2"] = 3000
				// no ca_schedule_ca:15_col_c
			},
		},
		{
			name:     "RC014 QBI deduction may carry to CA",
			code:     "RC014",
			severity: SeverityInfo,
			modify: func(r map[string]float64, s map[string]string) {
				r["form_8995:15"] = 5000
				r["1040:12"] = 20000
				r["ca_540:18"] = 20000 // same as federal, suggests QBI included
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, strInputs := baseReasonablenessCAInputs()
			tt.modify(results, strInputs)
			report := ReasonablenessCheck(results, strInputs, 2025, true)
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

func TestReasonablenessCheck_CA_NotTriggered(t *testing.T) {
	tests := []struct {
		name   string
		code   string
		modify func(results map[string]float64, strInputs map[string]string)
	}{
		{
			name: "RC011 CA AGI close to federal",
			code: "RC011",
			modify: func(r map[string]float64, s map[string]string) {
				r["1040:11"] = 50000
				r["ca_540:17"] = 55000 // 10% diff
			},
		},
		{
			name: "RC011 CA AGI differs > 20% but diff <= $5000",
			code: "RC011",
			modify: func(r map[string]float64, s map[string]string) {
				r["1040:11"] = 10000
				r["ca_540:17"] = 14000 // 40% diff but only $4000
			},
		},
		{
			name: "RC012 CA tax within max rate",
			code: "RC012",
			modify: func(r map[string]float64, s map[string]string) {
				r["ca_540:19"] = 100000
				r["ca_540:31"] = 10000 // 10%
			},
		},
		{
			name: "RC013 HSA with CA add-back",
			code: "RC013",
			modify: func(r map[string]float64, s map[string]string) {
				r["form_8889:2"] = 3000
				r["ca_schedule_ca:15_col_c"] = 3000
			},
		},
		{
			name: "RC014 QBI deduction not carried to CA",
			code: "RC014",
			modify: func(r map[string]float64, s map[string]string) {
				r["form_8995:15"] = 5000
				r["1040:12"] = 20000
				r["ca_540:18"] = 15000 // less than federal, QBI not included
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, strInputs := baseReasonablenessCAInputs()
			tt.modify(results, strInputs)
			report := ReasonablenessCheck(results, strInputs, 2025, true)
			if findResult(report, tt.code) != nil {
				t.Errorf("expected rule %s NOT to trigger", tt.code)
			}
		})
	}
}

func TestReasonablenessCheck_CA_SkippedWhenNotIncluded(t *testing.T) {
	results, strInputs := baseReasonablenessCAInputs()
	// Set up conditions that would trigger RC011
	results["1040:11"] = 50000
	results["ca_540:17"] = 100000

	report := ReasonablenessCheck(results, strInputs, 2025, false)
	if findResult(report, "RC011") != nil {
		t.Error("CA rules should not run when includeCA is false")
	}
}

func TestReasonablenessCheck_NilMaps(t *testing.T) {
	report := ReasonablenessCheck(nil, nil, 2025, false)
	if !report.IsValid {
		t.Error("expected valid report with nil inputs (no errors possible)")
	}
	if len(report.Results) != 0 {
		t.Errorf("expected no results with nil inputs, got %d", len(report.Results))
	}
}

func TestReasonablenessCheck_RC002_NegativeTaxableIncome(t *testing.T) {
	// When AGI < deduction, taxable income should be 0, not negative
	results, strInputs := baseReasonablenessInputs()
	results["1040:11"] = 10000
	results["1040:12"] = 14600
	results["1040:15"] = 0 // correct: max(0, 10000-14600)
	report := ReasonablenessCheck(results, strInputs, 2025, false)
	if findResult(report, "RC002") != nil {
		t.Error("RC002 should not trigger when taxable income is correctly floored at 0")
	}
}
