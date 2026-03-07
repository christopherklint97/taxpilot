package efile

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

// EFileAuth holds e-file signature authorization data (Form 8879).
type EFileAuth struct {
	// Federal
	TaxpayerName  string
	SSN           string
	TaxYear       int
	FilingStatus  string
	AGI           float64
	TotalTax      float64
	FederalRefund float64
	FederalOwed   float64
	SelfSelectPIN string // 5-digit self-select PIN
	PriorYearAGI  float64
	PriorYearPIN  string
	SignatureDate time.Time

	// CA-specific
	CAAgi           float64
	CATotalTax      float64
	CARefund        float64
	CAOwed          float64
	CASelfSelectPIN string
	CASignatureDate time.Time
}

// NewEFileAuth creates an EFileAuth from solver results.
func NewEFileAuth(results map[string]float64, strInputs map[string]string, taxYear int) *EFileAuth {
	return &EFileAuth{
		TaxpayerName:  strInputs["1040:first_name"] + " " + strInputs["1040:last_name"],
		SSN:           strInputs["1040:ssn"],
		TaxYear:       taxYear,
		FilingStatus:  strInputs["1040:filing_status"],
		AGI:           results["1040:11"],
		TotalTax:      results["1040:24"],
		FederalRefund: results["1040:34"],
		FederalOwed:   results["1040:37"],
		CAAgi:         results["ca_540:17"],
		CATotalTax:    results["ca_540:40"],
		CARefund:      results["ca_540:91"],
		CAOwed:        results["ca_540:93"],
	}
}

// GeneratePIN generates a random 5-digit self-select PIN.
// The PIN must be 00001-99999 (no all-zeros).
func GeneratePIN() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(99999))
	if err != nil {
		return "", fmt.Errorf("generating PIN: %w", err)
	}
	pin := n.Int64() + 1 // 1-99999
	return fmt.Sprintf("%05d", pin), nil
}

// SetFederalPIN sets the federal self-select PIN and signature date.
func (a *EFileAuth) SetFederalPIN(pin string) error {
	if len(pin) != 5 {
		return fmt.Errorf("PIN must be exactly 5 digits, got %d", len(pin))
	}
	for _, c := range pin {
		if c < '0' || c > '9' {
			return fmt.Errorf("PIN must contain only digits")
		}
	}
	if pin == "00000" {
		return fmt.Errorf("PIN cannot be all zeros")
	}
	a.SelfSelectPIN = pin
	a.SignatureDate = time.Now()
	return nil
}

// SetCAPIN sets the CA FTB self-select PIN and signature date.
func (a *EFileAuth) SetCAPIN(pin string) error {
	if len(pin) != 5 {
		return fmt.Errorf("PIN must be exactly 5 digits, got %d", len(pin))
	}
	for _, c := range pin {
		if c < '0' || c > '9' {
			return fmt.Errorf("PIN must contain only digits")
		}
	}
	if pin == "00000" {
		return fmt.Errorf("PIN cannot be all zeros")
	}
	a.CASelfSelectPIN = pin
	a.CASignatureDate = time.Now()
	return nil
}

// IsReadyFederal returns true if federal e-file authorization is complete.
func (a *EFileAuth) IsReadyFederal() bool {
	return a.SelfSelectPIN != "" && a.SSN != "" && !a.SignatureDate.IsZero()
}

// IsReadyCA returns true if CA e-file authorization is complete.
func (a *EFileAuth) IsReadyCA() bool {
	return a.CASelfSelectPIN != "" && a.SSN != "" && !a.CASignatureDate.IsZero()
}
