package efile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig_Missing(t *testing.T) {
	// Override home to a temp dir so config file doesn't exist
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config")
	}
	// Should be all defaults (zero values)
	if cfg.EFIN != "" {
		t.Errorf("expected empty EFIN, got %q", cfg.EFIN)
	}
	if cfg.CertPath != "" {
		t.Errorf("expected empty CertPath, got %q", cfg.CertPath)
	}
	if cfg.ATSMode {
		t.Error("expected ATSMode false")
	}
	if cfg.FTBProviderID != "" {
		t.Errorf("expected empty FTBProviderID, got %q", cfg.FTBProviderID)
	}
	if cfg.EFINApproved {
		t.Error("expected EFINApproved false")
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Create a cert file so validation can find it
	certPath := filepath.Join(tmpDir, "test.p12")
	if err := os.WriteFile(certPath, []byte("fake-cert"), 0600); err != nil {
		t.Fatal(err)
	}

	original := &Config{
		EFIN:          "123456",
		CertPath:      certPath,
		CertPassword:  "secret123", // should NOT be saved
		ATSMode:       true,
		FTBProviderID: "FTB999",
		PATSMode:      false,
		EFINApproved:  true,
		ATSPassed:     true,
		PATSPassed:    true,
		CertExpiry:    "2027-12-31",
	}

	if err := original.SaveConfig(); err != nil {
		t.Fatalf("SaveConfig() error: %v", err)
	}

	// Verify the file exists
	path := ConfigPath()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	// Verify CertPassword is NOT in the file
	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), "secret123") {
		t.Error("CertPassword was serialized to disk -- this is a security issue")
	}

	// Load and compare
	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}
	if loaded.EFIN != "123456" {
		t.Errorf("EFIN = %q, want %q", loaded.EFIN, "123456")
	}
	if loaded.CertPath != certPath {
		t.Errorf("CertPath = %q, want %q", loaded.CertPath, certPath)
	}
	if loaded.CertPassword != "" {
		t.Errorf("CertPassword should be empty after load, got %q", loaded.CertPassword)
	}
	if !loaded.ATSMode {
		t.Error("ATSMode should be true")
	}
	if loaded.FTBProviderID != "FTB999" {
		t.Errorf("FTBProviderID = %q, want %q", loaded.FTBProviderID, "FTB999")
	}
	if !loaded.EFINApproved {
		t.Error("EFINApproved should be true")
	}
	if !loaded.ATSPassed {
		t.Error("ATSPassed should be true")
	}
	if !loaded.PATSPassed {
		t.Error("PATSPassed should be true")
	}
	if loaded.CertExpiry != "2027-12-31" {
		t.Errorf("CertExpiry = %q, want %q", loaded.CertExpiry, "2027-12-31")
	}
}

func TestValidateConfig_Complete(t *testing.T) {
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.p12")
	if err := os.WriteFile(certPath, []byte("fake-cert"), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := &Config{
		EFIN:          "123456",
		CertPath:      certPath,
		ATSMode:       false,
		FTBProviderID: "FTB999",
		PATSMode:      false,
		EFINApproved:  true,
		ATSPassed:     true,
		PATSPassed:    true,
		CertExpiry:    "2027-12-31",
	}

	issues := cfg.ValidateConfig()
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %d: %v", len(issues), issues)
	}
}

func TestValidateConfig_MissingEFIN(t *testing.T) {
	cfg := &Config{
		CertPath:      "/some/cert.p12",
		FTBProviderID: "FTB999",
		EFINApproved:  true,
		ATSPassed:     true,
		PATSPassed:    true,
	}

	issues := cfg.ValidateConfig()
	found := false
	for _, issue := range issues {
		if strings.Contains(issue, "EFIN") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected EFIN-related issue, got: %v", issues)
	}
}

func TestValidateConfig_MissingCert(t *testing.T) {
	cfg := &Config{
		EFIN:          "123456",
		FTBProviderID: "FTB999",
		EFINApproved:  true,
		ATSPassed:     true,
		PATSPassed:    true,
	}

	issues := cfg.ValidateConfig()
	found := false
	for _, issue := range issues {
		if strings.Contains(issue, "certificate") || strings.Contains(issue, "cert_path") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected certificate-related issue, got: %v", issues)
	}
}

func TestCanFileFederal(t *testing.T) {
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.p12")
	if err := os.WriteFile(certPath, []byte("fake"), 0600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		cfg  Config
		want bool
	}{
		{
			name: "complete config",
			cfg: Config{
				EFIN:         "123456",
				CertPath:     certPath,
				EFINApproved: true,
				ATSPassed:    true,
				CertExpiry:   "2027-12-31",
			},
			want: true,
		},
		{
			name: "missing EFIN",
			cfg: Config{
				CertPath:     certPath,
				EFINApproved: true,
				ATSPassed:    true,
			},
			want: false,
		},
		{
			name: "missing cert",
			cfg: Config{
				EFIN:         "123456",
				EFINApproved: true,
				ATSPassed:    true,
			},
			want: false,
		},
		{
			name: "EFIN not approved",
			cfg: Config{
				EFIN:         "123456",
				CertPath:     certPath,
				EFINApproved: false,
				ATSPassed:    true,
			},
			want: false,
		},
		{
			name: "ATS not passed",
			cfg: Config{
				EFIN:         "123456",
				CertPath:     certPath,
				EFINApproved: true,
				ATSPassed:    false,
			},
			want: false,
		},
		{
			name: "expired cert",
			cfg: Config{
				EFIN:         "123456",
				CertPath:     certPath,
				EFINApproved: true,
				ATSPassed:    true,
				CertExpiry:   "2020-01-01",
			},
			want: false,
		},
		{
			name: "cert file not found",
			cfg: Config{
				EFIN:         "123456",
				CertPath:     "/nonexistent/cert.p12",
				EFINApproved: true,
				ATSPassed:    true,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.CanFileFederal()
			if got != tt.want {
				t.Errorf("CanFileFederal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanFileCA(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want bool
	}{
		{
			name: "complete config",
			cfg: Config{
				FTBProviderID: "FTB999",
				PATSPassed:    true,
			},
			want: true,
		},
		{
			name: "missing provider ID",
			cfg: Config{
				PATSPassed: true,
			},
			want: false,
		},
		{
			name: "PATS not passed",
			cfg: Config{
				FTBProviderID: "FTB999",
				PATSPassed:    false,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.CanFileCA()
			if got != tt.want {
				t.Errorf("CanFileCA() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatusSummary(t *testing.T) {
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.p12")
	if err := os.WriteFile(certPath, []byte("fake"), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := &Config{
		EFIN:          "123456",
		CertPath:      certPath,
		ATSMode:       false,
		FTBProviderID: "FTB999",
		PATSMode:      true,
		EFINApproved:  true,
		ATSPassed:     true,
		PATSPassed:    false,
		CertExpiry:    "2027-12-31",
	}

	summary := cfg.StatusSummary()

	// Should contain key information
	checks := []string{
		"123456",
		"READY",
		"NOT READY",
		"FTB999",
		"Federal",
		"California",
		"2027-12-31",
	}

	for _, check := range checks {
		if !strings.Contains(summary, check) {
			t.Errorf("StatusSummary() missing %q in output:\n%s", check, summary)
		}
	}
}

func TestCertPasswordNotSerialized(t *testing.T) {
	cfg := &Config{
		EFIN:         "123456",
		CertPassword: "super-secret",
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(data), "super-secret") {
		t.Error("CertPassword was included in JSON serialization")
	}
	if strings.Contains(string(data), "cert_password") {
		t.Error("cert_password key was included in JSON serialization")
	}
}
