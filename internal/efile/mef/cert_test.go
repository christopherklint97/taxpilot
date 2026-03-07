package mef

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	gopkcs12 "software.sslmate.com/src/go-pkcs12"
)

// generateTestCert creates a self-signed PKCS#12 certificate for testing.
// Returns the PKCS#12 data and the password used.
func generateTestCert(t *testing.T, notBefore, notAfter time.Time) ([]byte, string) {
	t.Helper()

	password := "test-password-123"

	// Generate RSA key pair
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}

	// Create self-signed certificate template
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "TaxPilot Test Certificate",
			Organization: []string{"TaxPilot Testing"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	// Self-sign the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("parse certificate: %v", err)
	}

	// Encode as PKCS#12
	p12Data, err := gopkcs12.Encode(rand.Reader, key, cert, nil, password)
	if err != nil {
		t.Fatalf("encode PKCS#12: %v", err)
	}

	return p12Data, password
}

// writeTestCert writes PKCS#12 data to a temp file and returns the path.
func writeTestCert(t *testing.T, p12Data []byte) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.p12")
	if err := os.WriteFile(path, p12Data, 0o600); err != nil {
		t.Fatalf("write test cert: %v", err)
	}
	return path
}

func TestLoadCertificate_Valid(t *testing.T) {
	notBefore := time.Now().Add(-24 * time.Hour)
	notAfter := time.Now().Add(365 * 24 * time.Hour)
	p12Data, password := generateTestCert(t, notBefore, notAfter)
	certPath := writeTestCert(t, p12Data)

	cc, err := LoadCertificate(certPath, password)
	if err != nil {
		t.Fatalf("LoadCertificate: %v", err)
	}
	if cc.CertPath != certPath {
		t.Errorf("CertPath: got %q, want %q", cc.CertPath, certPath)
	}
	if cc.certificate == nil {
		t.Error("certificate should not be nil")
	}
	if cc.privateKey == nil {
		t.Error("privateKey should not be nil")
	}
}

func TestLoadCertificate_FileNotFound(t *testing.T) {
	_, err := LoadCertificate("/nonexistent/path/cert.p12", "password")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "read certificate file") {
		t.Errorf("expected file read error, got: %v", err)
	}
}

func TestLoadCertificate_InvalidPassword(t *testing.T) {
	notBefore := time.Now().Add(-24 * time.Hour)
	notAfter := time.Now().Add(365 * 24 * time.Hour)
	p12Data, _ := generateTestCert(t, notBefore, notAfter)
	certPath := writeTestCert(t, p12Data)

	_, err := LoadCertificate(certPath, "wrong-password")
	if err == nil {
		t.Error("expected error for invalid password")
	}
	if !strings.Contains(err.Error(), "decode PKCS#12") {
		t.Errorf("expected decode error, got: %v", err)
	}
}

func TestValidateCertificate_Valid(t *testing.T) {
	notBefore := time.Now().Add(-24 * time.Hour)
	notAfter := time.Now().Add(365 * 24 * time.Hour)
	p12Data, password := generateTestCert(t, notBefore, notAfter)
	certPath := writeTestCert(t, p12Data)

	cc, err := LoadCertificate(certPath, password)
	if err != nil {
		t.Fatalf("LoadCertificate: %v", err)
	}

	if err := cc.ValidateCertificate(); err != nil {
		t.Errorf("ValidateCertificate: unexpected error: %v", err)
	}
}

func TestValidateCertificate_Expired(t *testing.T) {
	notBefore := time.Now().Add(-365 * 24 * time.Hour)
	notAfter := time.Now().Add(-24 * time.Hour) // expired yesterday
	p12Data, password := generateTestCert(t, notBefore, notAfter)
	certPath := writeTestCert(t, p12Data)

	cc, err := LoadCertificate(certPath, password)
	if err != nil {
		t.Fatalf("LoadCertificate: %v", err)
	}

	err = cc.ValidateCertificate()
	if err == nil {
		t.Error("expected error for expired certificate")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("expected expired error, got: %v", err)
	}
}

func TestValidateCertificate_NotYetValid(t *testing.T) {
	notBefore := time.Now().Add(24 * time.Hour) // valid tomorrow
	notAfter := time.Now().Add(365 * 24 * time.Hour)
	p12Data, password := generateTestCert(t, notBefore, notAfter)
	certPath := writeTestCert(t, p12Data)

	cc, err := LoadCertificate(certPath, password)
	if err != nil {
		t.Fatalf("LoadCertificate: %v", err)
	}

	err = cc.ValidateCertificate()
	if err == nil {
		t.Error("expected error for not-yet-valid certificate")
	}
	if !strings.Contains(err.Error(), "not yet valid") {
		t.Errorf("expected not-yet-valid error, got: %v", err)
	}
}

func TestValidateCertificate_NilCert(t *testing.T) {
	cc := &CertConfig{}
	err := cc.ValidateCertificate()
	if err == nil {
		t.Error("expected error for nil certificate")
	}
}

func TestDaysUntilExpiry(t *testing.T) {
	tests := []struct {
		name     string
		expiry   time.Time
		wantMin  int
		wantMax  int
	}{
		{
			name:    "expires in 90 days",
			expiry:  time.Now().Add(90 * 24 * time.Hour),
			wantMin: 89,
			wantMax: 90,
		},
		{
			name:    "expires in 1 day",
			expiry:  time.Now().Add(24 * time.Hour),
			wantMin: 0,
			wantMax: 1,
		},
		{
			name:    "already expired",
			expiry:  time.Now().Add(-24 * time.Hour),
			wantMin: 0,
			wantMax: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := &CertConfig{
				ExpiresAt:   tt.expiry,
				certificate: &x509.Certificate{NotAfter: tt.expiry},
			}
			got := cc.DaysUntilExpiry()
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("DaysUntilExpiry: got %d, want between %d and %d",
					got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestDaysUntilExpiry_NilCert(t *testing.T) {
	cc := &CertConfig{}
	if got := cc.DaysUntilExpiry(); got != 0 {
		t.Errorf("DaysUntilExpiry with nil cert: got %d, want 0", got)
	}
}

func TestWarnIfExpiringSoon(t *testing.T) {
	tests := []struct {
		name      string
		expiry    time.Time
		wantEmpty bool
		wantMsg   string
	}{
		{
			name:      "expires in 90 days - no warning",
			expiry:    time.Now().Add(90 * 24 * time.Hour),
			wantEmpty: true,
		},
		{
			name:    "expires in 15 days - warning",
			expiry:  time.Now().Add(15 * 24 * time.Hour),
			wantMsg: "expires in",
		},
		{
			name:    "expires in 5 days - warning",
			expiry:  time.Now().Add(5 * 24 * time.Hour),
			wantMsg: "expires in",
		},
		{
			name:    "already expired",
			expiry:  time.Now().Add(-24 * time.Hour),
			wantMsg: "expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := &CertConfig{
				ExpiresAt:   tt.expiry,
				certificate: &x509.Certificate{NotAfter: tt.expiry},
			}
			got := cc.WarnIfExpiringSoon()
			if tt.wantEmpty && got != "" {
				t.Errorf("WarnIfExpiringSoon: got %q, want empty", got)
			}
			if !tt.wantEmpty && !strings.Contains(got, tt.wantMsg) {
				t.Errorf("WarnIfExpiringSoon: got %q, want to contain %q", got, tt.wantMsg)
			}
		})
	}
}

func TestCertificateInfo(t *testing.T) {
	notBefore := time.Now().Add(-24 * time.Hour)
	notAfter := time.Now().Add(365 * 24 * time.Hour)
	p12Data, password := generateTestCert(t, notBefore, notAfter)
	certPath := writeTestCert(t, p12Data)

	cc, err := LoadCertificate(certPath, password)
	if err != nil {
		t.Fatalf("LoadCertificate: %v", err)
	}

	info := cc.CertificateInfo()
	if !strings.Contains(info, "TaxPilot Test Certificate") {
		t.Errorf("CertificateInfo should contain subject CN, got: %s", info)
	}
	if !strings.Contains(info, "Subject:") {
		t.Errorf("CertificateInfo should contain 'Subject:', got: %s", info)
	}
	if !strings.Contains(info, "Days until expiry:") {
		t.Errorf("CertificateInfo should contain expiry info, got: %s", info)
	}
}

func TestCertificateInfo_NilCert(t *testing.T) {
	cc := &CertConfig{}
	info := cc.CertificateInfo()
	if info != "No certificate loaded" {
		t.Errorf("CertificateInfo with nil cert: got %q", info)
	}
}

func TestTLSConfig(t *testing.T) {
	notBefore := time.Now().Add(-24 * time.Hour)
	notAfter := time.Now().Add(365 * 24 * time.Hour)
	p12Data, password := generateTestCert(t, notBefore, notAfter)
	certPath := writeTestCert(t, p12Data)

	cc, err := LoadCertificate(certPath, password)
	if err != nil {
		t.Fatalf("LoadCertificate: %v", err)
	}

	tlsConf, err := cc.TLSConfig()
	if err != nil {
		t.Fatalf("TLSConfig: %v", err)
	}
	if len(tlsConf.Certificates) != 1 {
		t.Errorf("TLSConfig: got %d certificates, want 1", len(tlsConf.Certificates))
	}
	if tlsConf.MinVersion != 0x0303 { // TLS 1.2
		t.Errorf("TLSConfig MinVersion: got %x, want 0x0303 (TLS 1.2)", tlsConf.MinVersion)
	}
}

func TestTLSConfig_NilCert(t *testing.T) {
	cc := &CertConfig{}
	_, err := cc.TLSConfig()
	if err == nil {
		t.Error("expected error for nil certificate")
	}
}
