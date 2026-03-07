package security

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	vault, err := NewVault("test-passphrase-123")
	if err != nil {
		t.Fatalf("NewVault: %v", err)
	}

	plaintext := []byte("SSN: 123-45-6789, AGI: $85,000")
	ciphertext, err := vault.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	// Ciphertext should differ from plaintext.
	if bytes.Equal(plaintext, ciphertext) {
		t.Fatal("ciphertext should not equal plaintext")
	}

	decrypted, err := vault.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Fatalf("round-trip mismatch: got %q, want %q", decrypted, plaintext)
	}
}

func TestWrongPassphraseFailsDecrypt(t *testing.T) {
	vault1, err := NewVault("correct-passphrase")
	if err != nil {
		t.Fatalf("NewVault: %v", err)
	}

	plaintext := []byte("sensitive tax data")
	ciphertext, err := vault1.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	// Create a second vault with the same salt but wrong passphrase.
	vault2 := &Vault{
		key:  deriveKey("wrong-passphrase", vault1.salt),
		salt: vault1.salt,
	}

	_, err = vault2.Decrypt(ciphertext)
	if err == nil {
		t.Fatal("expected decryption to fail with wrong passphrase")
	}
}

func TestTamperedCiphertextFails(t *testing.T) {
	vault, err := NewVault("my-passphrase")
	if err != nil {
		t.Fatalf("NewVault: %v", err)
	}

	plaintext := []byte("important financial data")
	ciphertext, err := vault.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	// Tamper with the ciphertext (flip a byte near the end).
	tampered := make([]byte, len(ciphertext))
	copy(tampered, ciphertext)
	tampered[len(tampered)-1] ^= 0xFF

	_, err = vault.Decrypt(tampered)
	if err == nil {
		t.Fatal("expected decryption to fail with tampered ciphertext")
	}
}

func TestEncryptJSONDecryptJSONRoundTrip(t *testing.T) {
	vault, err := NewVault("json-test-pass")
	if err != nil {
		t.Fatalf("NewVault: %v", err)
	}

	type TaxData struct {
		SSN    string  `json:"ssn"`
		AGI    float64 `json:"agi"`
		Refund float64 `json:"refund"`
	}

	original := TaxData{
		SSN:    "123-45-6789",
		AGI:    85000.50,
		Refund: 2345.67,
	}

	ciphertext, err := vault.EncryptJSON(original)
	if err != nil {
		t.Fatalf("EncryptJSON: %v", err)
	}

	var decoded TaxData
	if err := vault.DecryptJSON(ciphertext, &decoded); err != nil {
		t.Fatalf("DecryptJSON: %v", err)
	}

	if decoded.SSN != original.SSN || decoded.AGI != original.AGI || decoded.Refund != original.Refund {
		t.Fatalf("JSON round-trip mismatch: got %+v, want %+v", decoded, original)
	}
}

func TestSaltSaveLoadRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	saltPath := filepath.Join(tmpDir, "vault.salt")

	vault1, err := NewVault("salt-test-pass")
	if err != nil {
		t.Fatalf("NewVault: %v", err)
	}

	if err := vault1.SaveSalt(saltPath); err != nil {
		t.Fatalf("SaveSalt: %v", err)
	}

	// Verify the salt file has correct permissions.
	info, err := os.Stat(saltPath)
	if err != nil {
		t.Fatalf("stat salt file: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("salt file permissions: got %o, want 600", perm)
	}

	// Encrypt with original vault.
	plaintext := []byte("test data for salt round-trip")
	ciphertext, err := vault1.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	// Load vault from saved salt with same passphrase.
	vault2, err := LoadVault("salt-test-pass", saltPath)
	if err != nil {
		t.Fatalf("LoadVault: %v", err)
	}

	decrypted, err := vault2.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt with loaded vault: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Fatalf("salt round-trip mismatch: got %q, want %q", decrypted, plaintext)
	}
}

func TestEncryptDecryptFile(t *testing.T) {
	tmpDir := t.TempDir()

	vault, err := NewVault("file-test-pass")
	if err != nil {
		t.Fatalf("NewVault: %v", err)
	}

	// Write a plaintext file.
	inPath := filepath.Join(tmpDir, "state.json")
	content := []byte(`{"tax_year":2025,"ssn":"123-45-6789"}`)
	if err := os.WriteFile(inPath, content, 0o644); err != nil {
		t.Fatalf("write input file: %v", err)
	}

	encPath := filepath.Join(tmpDir, "state.enc")
	if err := vault.EncryptFile(inPath, encPath); err != nil {
		t.Fatalf("EncryptFile: %v", err)
	}

	// Encrypted file should differ from original.
	encData, _ := os.ReadFile(encPath)
	if bytes.Equal(content, encData) {
		t.Fatal("encrypted file should not equal plaintext")
	}

	outPath := filepath.Join(tmpDir, "state_decrypted.json")
	if err := vault.DecryptFile(encPath, outPath); err != nil {
		t.Fatalf("DecryptFile: %v", err)
	}

	decrypted, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read decrypted file: %v", err)
	}

	if !bytes.Equal(content, decrypted) {
		t.Fatalf("file round-trip mismatch: got %q, want %q", decrypted, content)
	}
}

func TestNewVaultEmptyPassphrase(t *testing.T) {
	_, err := NewVault("")
	if err == nil {
		t.Fatal("expected error for empty passphrase")
	}
}

func TestLoadVaultInvalidSalt(t *testing.T) {
	tmpDir := t.TempDir()
	saltPath := filepath.Join(tmpDir, "bad.salt")

	// Write a salt file with wrong length.
	if err := os.WriteFile(saltPath, []byte("tooshort"), 0o600); err != nil {
		t.Fatalf("write bad salt: %v", err)
	}

	_, err := LoadVault("some-pass", saltPath)
	if err == nil {
		t.Fatal("expected error for invalid salt length")
	}
}

func TestDecryptTooShortCiphertext(t *testing.T) {
	vault, err := NewVault("short-test")
	if err != nil {
		t.Fatalf("NewVault: %v", err)
	}

	_, err = vault.Decrypt([]byte("tiny"))
	if err == nil {
		t.Fatal("expected error for too-short ciphertext")
	}
}
