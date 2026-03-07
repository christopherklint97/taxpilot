package interview

import (
	"strings"
	"testing"
)

func TestFBARRequired(t *testing.T) {
	tests := []struct {
		name     string
		maxValue float64
		want     bool
	}{
		{"above threshold", 15000, true},
		{"exactly threshold", 10000, false},
		{"below threshold", 5000, false},
		{"just above", 10001, true},
		{"zero", 0, false},
		{"large amount", 500000, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FBARRequired(tt.maxValue)
			if got != tt.want {
				t.Errorf("FBARRequired(%.0f) = %v, want %v", tt.maxValue, got, tt.want)
			}
		})
	}
}

func TestFBARDeadline(t *testing.T) {
	deadline := FBARDeadline(2025)
	if !strings.Contains(deadline, "April 15, 2026") {
		t.Errorf("expected April 15, 2026 in deadline, got %q", deadline)
	}
	if !strings.Contains(deadline, "October 15, 2026") {
		t.Errorf("expected October 15, 2026 extension in deadline, got %q", deadline)
	}
}

func TestFBARGuidanceMessage(t *testing.T) {
	msg := FBARGuidanceMessage()
	if !strings.Contains(msg, "FinCEN Form 114") {
		t.Error("guidance should mention FinCEN Form 114")
	}
	if !strings.Contains(msg, "$10,000") {
		t.Error("guidance should mention $10,000 threshold")
	}
	if !strings.Contains(msg, "bsaefiling.fincen.treas.gov") {
		t.Error("guidance should include BSA E-Filing URL")
	}
}

func TestFBARNotRequiredMessage(t *testing.T) {
	msg := FBARNotRequiredMessage()
	if !strings.Contains(msg, "do not need to file") {
		t.Error("not-required message should say filing not needed")
	}
}
