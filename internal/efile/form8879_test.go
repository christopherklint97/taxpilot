package efile

import (
	"testing"
)

func TestNewEFileAuth(t *testing.T) {
	results := map[string]float64{
		"1040:11":   75000,
		"1040:24":   8114,
		"1040:34":   1386,
		"1040:37":   0,
		"ca_540:17": 75000,
		"ca_540:40": 2860,
		"ca_540:91": 340,
		"ca_540:93": 0,
	}
	strInputs := map[string]string{
		"1040:first_name":    "Jane",
		"1040:last_name":     "Doe",
		"1040:ssn":           "123-45-6789",
		"1040:filing_status": "single",
	}

	auth := NewEFileAuth(results, strInputs, 2025)

	if auth.TaxpayerName != "Jane Doe" {
		t.Errorf("TaxpayerName: got %q, want %q", auth.TaxpayerName, "Jane Doe")
	}
	if auth.AGI != 75000 {
		t.Errorf("AGI: got %.2f, want 75000", auth.AGI)
	}
	if auth.TotalTax != 8114 {
		t.Errorf("TotalTax: got %.2f, want 8114", auth.TotalTax)
	}
	if auth.FederalRefund != 1386 {
		t.Errorf("FederalRefund: got %.2f, want 1386", auth.FederalRefund)
	}
	if auth.CAAgi != 75000 {
		t.Errorf("CAAgi: got %.2f, want 75000", auth.CAAgi)
	}
}

func TestGeneratePIN(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		pin, err := GeneratePIN()
		if err != nil {
			t.Fatalf("GeneratePIN() error: %v", err)
		}
		if len(pin) != 5 {
			t.Errorf("PIN length: got %d, want 5", len(pin))
		}
		if pin == "00000" {
			t.Error("PIN should never be 00000")
		}
		for _, c := range pin {
			if c < '0' || c > '9' {
				t.Errorf("PIN contains non-digit: %q", pin)
			}
		}
		seen[pin] = true
	}
	// With 100 attempts, we should see multiple unique PINs
	if len(seen) < 10 {
		t.Errorf("expected diverse PINs, only got %d unique", len(seen))
	}
}

func TestSetFederalPIN(t *testing.T) {
	auth := &EFileAuth{SSN: "123-45-6789"}

	// Valid PIN
	if err := auth.SetFederalPIN("12345"); err != nil {
		t.Errorf("SetFederalPIN(12345): unexpected error: %v", err)
	}
	if auth.SelfSelectPIN != "12345" {
		t.Errorf("PIN: got %q, want %q", auth.SelfSelectPIN, "12345")
	}
	if auth.SignatureDate.IsZero() {
		t.Error("SignatureDate should be set")
	}

	// Invalid: too short
	if err := auth.SetFederalPIN("123"); err == nil {
		t.Error("expected error for 3-digit PIN")
	}

	// Invalid: all zeros
	if err := auth.SetFederalPIN("00000"); err == nil {
		t.Error("expected error for all-zero PIN")
	}

	// Invalid: non-digits
	if err := auth.SetFederalPIN("1234a"); err == nil {
		t.Error("expected error for non-digit PIN")
	}
}

func TestSetCAPIN(t *testing.T) {
	auth := &EFileAuth{SSN: "123-45-6789"}

	if err := auth.SetCAPIN("54321"); err != nil {
		t.Errorf("SetCAPIN(54321): unexpected error: %v", err)
	}
	if auth.CASelfSelectPIN != "54321" {
		t.Errorf("CA PIN: got %q, want %q", auth.CASelfSelectPIN, "54321")
	}
	if auth.CASignatureDate.IsZero() {
		t.Error("CASignatureDate should be set")
	}
}

func TestIsReady(t *testing.T) {
	auth := &EFileAuth{SSN: "123-45-6789"}

	if auth.IsReadyFederal() {
		t.Error("should not be ready before setting PIN")
	}
	if auth.IsReadyCA() {
		t.Error("CA should not be ready before setting PIN")
	}

	auth.SetFederalPIN("12345")
	if !auth.IsReadyFederal() {
		t.Error("should be ready after setting federal PIN")
	}
	if auth.IsReadyCA() {
		t.Error("CA should still not be ready")
	}

	auth.SetCAPIN("54321")
	if !auth.IsReadyCA() {
		t.Error("CA should be ready after setting CA PIN")
	}
}
