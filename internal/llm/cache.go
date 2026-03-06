package llm

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Cache stores LLM responses to avoid redundant API calls.
// Uses an in-memory map backed by an optional file-based cache.
type Cache struct {
	mu       sync.RWMutex
	entries  map[string]string // hash -> response
	cacheDir string           // optional persistent cache directory
}

// NewCache creates a new cache. If cacheDir is empty, defaults to ~/.taxpilot/llm_cache/.
func NewCache(cacheDir string) *Cache {
	if cacheDir == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			cacheDir = filepath.Join(home, ".taxpilot", "llm_cache")
		}
	}
	return &Cache{
		entries:  make(map[string]string),
		cacheDir: cacheDir,
	}
}

// Get retrieves a cached response by key. Returns the response and true if found.
func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.entries[key]
	return val, ok
}

// Set stores a response in the cache.
func (c *Cache) Set(key string, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = value
}

// HashKey produces a SHA-256 hex digest of the serialized messages, suitable as a cache key.
func (c *Cache) HashKey(messages []Message) string {
	data, _ := json.Marshal(messages)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// Load reads the cache from disk (if cacheDir is set).
func (c *Cache) Load() error {
	if c.cacheDir == "" {
		return nil
	}
	path := filepath.Join(c.cacheDir, "cache.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // no cache file yet, that's fine
		}
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return json.Unmarshal(data, &c.entries)
}

// Save writes the cache to disk (if cacheDir is set).
func (c *Cache) Save() error {
	if c.cacheDir == "" {
		return nil
	}
	if err := os.MkdirAll(c.cacheDir, 0o755); err != nil {
		return err
	}
	c.mu.RLock()
	data, err := json.MarshalIndent(c.entries, "", "  ")
	c.mu.RUnlock()
	if err != nil {
		return err
	}
	path := filepath.Join(c.cacheDir, "cache.json")
	return os.WriteFile(path, data, 0o644)
}
