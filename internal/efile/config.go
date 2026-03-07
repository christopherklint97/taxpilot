package efile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Config holds e-file provider configuration.
type Config struct {
	// IRS MeF settings
	EFIN         string `json:"efin"`
	CertPath     string `json:"cert_path"`
	CertPassword string `json:"-"` // never serialized
	ATSMode      bool   `json:"ats_mode"`

	// CA FTB settings
	FTBProviderID string `json:"ftb_provider_id"`
	PATSMode      bool   `json:"pats_mode"`

	// Status tracking
	EFINApproved bool   `json:"efin_approved"`
	ATSPassed    bool   `json:"ats_passed"`
	PATSPassed   bool   `json:"pats_passed"`
	CertExpiry   string `json:"cert_expiry"`
}

// ConfigPath returns the path to the e-file config file.
func ConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".taxpilot", "efile.json")
	}
	return filepath.Join(home, ".taxpilot", "efile.json")
}

// LoadConfig reads e-file configuration from ~/.taxpilot/efile.json.
// If the file does not exist, a default (empty) Config is returned.
func LoadConfig() (*Config, error) {
	path := ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("reading efile config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing efile config: %w", err)
	}
	return &cfg, nil
}

// SaveConfig writes e-file configuration to ~/.taxpilot/efile.json.
func (c *Config) SaveConfig() error {
	path := ConfigPath()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling efile config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing efile config: %w", err)
	}
	return nil
}

// ValidateConfig checks that all required settings are present for e-filing.
// Returns a list of human-readable issues. An empty list means the config is valid.
func (c *Config) ValidateConfig() []string {
	var issues []string

	// Federal checks
	if c.EFIN == "" {
		issues = append(issues, "Missing EFIN (Electronic Filing Identification Number)")
	}
	if c.CertPath == "" {
		issues = append(issues, "Missing certificate path (cert_path)")
	} else if _, err := os.Stat(c.CertPath); err != nil {
		issues = append(issues, fmt.Sprintf("Certificate file not found: %s", c.CertPath))
	}
	if !c.EFINApproved {
		issues = append(issues, "EFIN not yet approved by IRS")
	}
	if !c.ATSPassed {
		issues = append(issues, "ATS certification not yet passed")
	}

	// Certificate expiry
	if c.CertExpiry != "" {
		expiry, err := time.Parse("2006-01-02", c.CertExpiry)
		if err != nil {
			issues = append(issues, fmt.Sprintf("Invalid cert_expiry date format: %s (expected YYYY-MM-DD)", c.CertExpiry))
		} else if time.Now().After(expiry) {
			issues = append(issues, fmt.Sprintf("Certificate expired on %s", c.CertExpiry))
		}
	}

	// CA checks
	if c.FTBProviderID == "" {
		issues = append(issues, "Missing FTB provider ID (ftb_provider_id)")
	}
	if !c.PATSPassed {
		issues = append(issues, "PATS certification not yet passed")
	}

	return issues
}

// CanFileFederal returns true if federal e-filing is fully configured.
func (c *Config) CanFileFederal() bool {
	if c.EFIN == "" || c.CertPath == "" {
		return false
	}
	if !c.EFINApproved || !c.ATSPassed {
		return false
	}
	// Check cert file exists
	if _, err := os.Stat(c.CertPath); err != nil {
		return false
	}
	// Check cert not expired
	if c.CertExpiry != "" {
		expiry, err := time.Parse("2006-01-02", c.CertExpiry)
		if err != nil {
			return false
		}
		if time.Now().After(expiry) {
			return false
		}
	}
	return true
}

// CanFileCA returns true if CA e-filing is fully configured.
func (c *Config) CanFileCA() bool {
	if c.FTBProviderID == "" {
		return false
	}
	if !c.PATSPassed {
		return false
	}
	return true
}

// StatusSummary returns a human-readable summary of e-file readiness.
func (c *Config) StatusSummary() string {
	var b strings.Builder

	b.WriteString("=== E-File Provider Status ===\n\n")

	// Federal
	b.WriteString("Federal (IRS MeF):\n")
	b.WriteString(fmt.Sprintf("  EFIN:             %s\n", valueOrMissing(c.EFIN)))
	b.WriteString(fmt.Sprintf("  EFIN Approved:    %s\n", yesNo(c.EFINApproved)))
	b.WriteString(fmt.Sprintf("  Certificate:      %s\n", valueOrMissing(c.CertPath)))
	if c.CertExpiry != "" {
		expired := ""
		if expiry, err := time.Parse("2006-01-02", c.CertExpiry); err == nil && time.Now().After(expiry) {
			expired = " (EXPIRED)"
		}
		b.WriteString(fmt.Sprintf("  Cert Expiry:      %s%s\n", c.CertExpiry, expired))
	}
	b.WriteString(fmt.Sprintf("  ATS Passed:       %s\n", yesNo(c.ATSPassed)))
	b.WriteString(fmt.Sprintf("  ATS Mode:         %s\n", yesNo(c.ATSMode)))
	if c.CanFileFederal() {
		b.WriteString("  Status:           READY\n")
	} else {
		b.WriteString("  Status:           NOT READY\n")
	}

	b.WriteString("\n")

	// CA
	b.WriteString("California (FTB):\n")
	b.WriteString(fmt.Sprintf("  FTB Provider ID:  %s\n", valueOrMissing(c.FTBProviderID)))
	b.WriteString(fmt.Sprintf("  PATS Passed:      %s\n", yesNo(c.PATSPassed)))
	b.WriteString(fmt.Sprintf("  PATS Mode:        %s\n", yesNo(c.PATSMode)))
	if c.CanFileCA() {
		b.WriteString("  Status:           READY\n")
	} else {
		b.WriteString("  Status:           NOT READY\n")
	}

	return b.String()
}

func valueOrMissing(s string) string {
	if s == "" {
		return "(not set)"
	}
	return s
}

func yesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
