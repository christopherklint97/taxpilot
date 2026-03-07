package mef

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	gopkcs12 "software.sslmate.com/src/go-pkcs12"
)

// CertConfig holds the configuration for IRS MeF Strong Authentication.
type CertConfig struct {
	CertPath    string            // path to PKCS#12 (.p12/.pfx) file
	Password    string            // certificate password (should come from secure store)
	EFIN        string            // Electronic Filing Identification Number
	ExpiresAt   time.Time         // certificate expiration date
	certificate *x509.Certificate // parsed certificate
	privateKey  interface{}       // parsed private key (RSA or ECDSA)
}

// LoadCertificate loads and validates a PKCS#12 certificate for MeF authentication.
// It reads the .p12/.pfx file, parses it with the given password, and validates
// the certificate type and expiration.
func LoadCertificate(certPath, password string) (*CertConfig, error) {
	data, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("mef: read certificate file %s: %w", certPath, err)
	}

	privateKey, cert, err := gopkcs12.Decode(data, password)
	if err != nil {
		return nil, fmt.Errorf("mef: decode PKCS#12 certificate: %w", err)
	}

	// Validate the private key type — MeF requires RSA or ECDSA.
	switch privateKey.(type) {
	case *rsa.PrivateKey, *ecdsa.PrivateKey:
		// OK
	default:
		return nil, fmt.Errorf("mef: unsupported key type %T — must be RSA or ECDSA", privateKey)
	}

	cc := &CertConfig{
		CertPath:    certPath,
		Password:    password,
		ExpiresAt:   cert.NotAfter,
		certificate: cert,
		privateKey:  privateKey,
	}

	return cc, nil
}

// ValidateCertificate checks that the certificate is valid and not expired.
func (cc *CertConfig) ValidateCertificate() error {
	if cc.certificate == nil {
		return fmt.Errorf("mef: no certificate loaded")
	}

	now := time.Now()
	if now.Before(cc.certificate.NotBefore) {
		return fmt.Errorf("mef: certificate is not yet valid (valid from %s)",
			cc.certificate.NotBefore.Format("2006-01-02"))
	}
	if now.After(cc.certificate.NotAfter) {
		return fmt.Errorf("mef: certificate expired on %s",
			cc.certificate.NotAfter.Format("2006-01-02"))
	}

	return nil
}

// TLSConfig returns a tls.Config configured with the Strong Authentication certificate.
// Used for mutual TLS (mTLS) with the IRS MeF SOAP endpoint.
func (cc *CertConfig) TLSConfig() (*tls.Config, error) {
	if cc.certificate == nil || cc.privateKey == nil {
		return nil, fmt.Errorf("mef: certificate or private key not loaded")
	}

	tlsCert := tls.Certificate{
		Certificate: [][]byte{cc.certificate.Raw},
		PrivateKey:  cc.privateKey,
		Leaf:        cc.certificate,
	}

	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// CertificateInfo returns human-readable info about the certificate.
func (cc *CertConfig) CertificateInfo() string {
	if cc.certificate == nil {
		return "No certificate loaded"
	}

	return fmt.Sprintf("Subject: %s\nIssuer: %s\nValid: %s to %s\nSerial: %s\nDays until expiry: %d",
		cc.certificate.Subject.CommonName,
		cc.certificate.Issuer.CommonName,
		cc.certificate.NotBefore.Format("2006-01-02"),
		cc.certificate.NotAfter.Format("2006-01-02"),
		cc.certificate.SerialNumber.String(),
		cc.DaysUntilExpiry(),
	)
}

// DaysUntilExpiry returns the number of days until the certificate expires.
// Returns 0 if the certificate is already expired.
func (cc *CertConfig) DaysUntilExpiry() int {
	if cc.certificate == nil {
		return 0
	}

	days := int(time.Until(cc.ExpiresAt).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

// WarnIfExpiringSoon returns a warning message if the cert expires within 30 days.
// Returns an empty string if the certificate has more than 30 days remaining.
func (cc *CertConfig) WarnIfExpiringSoon() string {
	days := cc.DaysUntilExpiry()
	if days <= 0 {
		return fmt.Sprintf("WARNING: Certificate has expired")
	}
	if days <= 30 {
		return fmt.Sprintf("WARNING: Certificate expires in %d days — renew soon", days)
	}
	return ""
}
