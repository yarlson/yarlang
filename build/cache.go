package build

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

// CacheEntry stores metadata about a compiled module
type CacheEntry struct {
	SourceHash string            `json:"source_hash"`
	ImportHash map[string]string `json:"import_hash"` // Import path -> hash
}

// CacheManager handles build cache
type CacheManager struct {
	cacheDir string
}

// NewCacheManager creates a cache manager
func NewCacheManager(projectRoot string) *CacheManager {
	cacheDir := filepath.Join(projectRoot, "build", "ir")
	return &CacheManager{cacheDir: cacheDir}
}

// GetCacheEntry loads cache metadata for a module
func (c *CacheManager) GetCacheEntry(modulePath string) (*CacheEntry, error) {
	hashFile := c.hashFilePath(modulePath)

	data, err := os.ReadFile(hashFile)
	if err != nil {
		return nil, err
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}

	return &entry, nil
}

// SaveCacheEntry saves cache metadata
func (c *CacheManager) SaveCacheEntry(modulePath string, entry *CacheEntry) error {
	hashFile := c.hashFilePath(modulePath)

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(hashFile, data, 0644)
}

// ComputeFileHash computes SHA-256 hash of file content
func (c *CacheManager) ComputeFileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// NeedsRebuild checks if a module needs recompilation
func (c *CacheManager) NeedsRebuild(modulePath string, imports map[string]string) (bool, error) {
	// Check if cache entry exists
	entry, err := c.GetCacheEntry(modulePath)
	if err != nil {
		return true, nil // No cache, need rebuild
	}

	// Check source hash
	currentHash, err := c.ComputeFileHash(modulePath)
	if err != nil {
		return false, err
	}

	if currentHash != entry.SourceHash {
		return true, nil // Source changed
	}

	// Check import hashes
	for impPath, impModule := range imports {
		cachedHash, ok := entry.ImportHash[impPath]
		if !ok {
			return true, nil // New import
		}

		currentImpHash, err := c.ComputeFileHash(impModule)
		if err != nil {
			return false, err
		}

		if currentImpHash != cachedHash {
			return true, nil // Import changed
		}
	}

	return false, nil // No rebuild needed
}

func (c *CacheManager) hashFilePath(modulePath string) string {
	base := filepath.Base(modulePath)
	name := base[:len(base)-len(filepath.Ext(base))]

	return filepath.Join(c.cacheDir, name+".ll.hash")
}
