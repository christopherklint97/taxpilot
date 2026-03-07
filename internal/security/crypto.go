package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters (OWASP recommended).
const (
	argonTime    = 3
	argonMemory  = 64 * 1024 // 64 MB
	argonThreads = 4
	argonKeyLen  = 32 // AES-256
	saltLen      = 16
)

// Vault handles encryption/decryption of sensitive data at rest.
type Vault struct {
	key  []byte // derived AES-256 key
	salt []byte // random salt for key derivation
}

// deriveKey derives an AES-256 key from a passphrase and salt using Argon2id.
func deriveKey(passphrase string, salt []byte) []byte {
	return argon2.IDKey([]byte(passphrase), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
}

// NewVault creates a Vault from a user passphrase.
// Derives an AES-256 key using Argon2id with a random salt.
func NewVault(passphrase string) (*Vault, error) {
	if passphrase == "" {
		return nil, errors.New("passphrase must not be empty")
	}

	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("generate salt: %w", err)
	}

	key := deriveKey(passphrase, salt)
	return &Vault{key: key, salt: salt}, nil
}

// LoadVault loads a Vault from an existing salt file and passphrase.
// The salt is stored in ~/.taxpilot/vault.salt (or a custom path).
func LoadVault(passphrase string, saltPath string) (*Vault, error) {
	if passphrase == "" {
		return nil, errors.New("passphrase must not be empty")
	}

	salt, err := os.ReadFile(saltPath)
	if err != nil {
		return nil, fmt.Errorf("read salt file %s: %w", saltPath, err)
	}

	if len(salt) != saltLen {
		return nil, fmt.Errorf("invalid salt length: got %d, want %d", len(salt), saltLen)
	}

	key := deriveKey(passphrase, salt)
	return &Vault{key: key, salt: salt}, nil
}

// SaveSalt saves the salt to disk for later key derivation.
func (v *Vault) SaveSalt(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}
	if err := os.WriteFile(path, v.salt, 0o600); err != nil {
		return fmt.Errorf("write salt to %s: %w", path, err)
	}
	return nil
}

// Encrypt encrypts plaintext data using AES-256-GCM.
// Returns: nonce + ciphertext (nonce prepended).
func (v *Vault) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(v.key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	// Seal appends ciphertext to nonce, so result is nonce + ciphertext + tag.
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts data encrypted by Encrypt.
func (v *Vault) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(v.key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ct := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	return plaintext, nil
}

// EncryptFile reads a file, encrypts it, and writes to outPath.
func (v *Vault) EncryptFile(inPath, outPath string) error {
	plaintext, err := os.ReadFile(inPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", inPath, err)
	}

	ciphertext, err := v.Encrypt(plaintext)
	if err != nil {
		return err
	}

	dir := filepath.Dir(outPath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(outPath, ciphertext, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", outPath, err)
	}
	return nil
}

// DecryptFile reads an encrypted file and writes decrypted to outPath.
func (v *Vault) DecryptFile(inPath, outPath string) error {
	ciphertext, err := os.ReadFile(inPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", inPath, err)
	}

	plaintext, err := v.Decrypt(ciphertext)
	if err != nil {
		return err
	}

	dir := filepath.Dir(outPath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(outPath, plaintext, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", outPath, err)
	}
	return nil
}

// EncryptJSON marshals data to JSON, then encrypts.
func (v *Vault) EncryptJSON(data interface{}) ([]byte, error) {
	plaintext, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal JSON: %w", err)
	}
	return v.Encrypt(plaintext)
}

// DecryptJSON decrypts, then unmarshals from JSON.
func (v *Vault) DecryptJSON(ciphertext []byte, target interface{}) error {
	plaintext, err := v.Decrypt(ciphertext)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(plaintext, target); err != nil {
		return fmt.Errorf("unmarshal JSON: %w", err)
	}
	return nil
}
