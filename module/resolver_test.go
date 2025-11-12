package module

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveLocalModule(t *testing.T) {
	// Setup test directory structure
	tmpDir := t.TempDir()

	// Create yar.toml (project root marker)
	yarToml := filepath.Join(tmpDir, "yar.toml")
	os.WriteFile(yarToml, []byte("[package]\nname=\"test\"\n"), 0644)

	// Create math.yar
	mathFile := filepath.Join(tmpDir, "math.yar")
	os.WriteFile(mathFile, []byte("func Add(x, y) { return x + y }"), 0644)

	// Create main.yar
	mainFile := filepath.Join(tmpDir, "main.yar")
	os.WriteFile(mainFile, []byte(`import "math"`), 0644)

	// Resolve from main.yar context
	resolver := NewResolver(tmpDir)

	resolved, err := resolver.Resolve("math", mainFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved != mathFile {
		t.Errorf("wrong path. expected=%q, got=%q", mathFile, resolved)
	}
}

func TestResolveNonExistent(t *testing.T) {
	tmpDir := t.TempDir()

	mainFile := filepath.Join(tmpDir, "main.yar")
	os.WriteFile(mainFile, []byte(`import "nonexistent"`), 0644)

	resolver := NewResolver(tmpDir)

	_, err := resolver.Resolve("nonexistent", mainFile)
	if err == nil {
		t.Fatal("expected error for non-existent module")
	}
}

func TestDetectImportCycle(t *testing.T) {
	tmpDir := t.TempDir()

	// Create yar.toml
	yarToml := filepath.Join(tmpDir, "yar.toml")
	os.WriteFile(yarToml, []byte("[package]\nname=\"test\"\n"), 0644)

	// Create a.yar that imports b
	aFile := filepath.Join(tmpDir, "a.yar")
	os.WriteFile(aFile, []byte(`import "b"`), 0644)

	// Create b.yar that imports a (cycle!)
	bFile := filepath.Join(tmpDir, "b.yar")
	os.WriteFile(bFile, []byte(`import "a"`), 0644)

	loader := NewLoader(tmpDir)

	_, err := loader.Load(aFile)
	if err == nil {
		t.Fatal("expected cycle detection error")
	}

	if !strings.Contains(err.Error(), "import cycle") {
		t.Errorf("expected 'import cycle' in error, got: %v", err)
	}
}
