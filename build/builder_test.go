package build

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildSingleModule(t *testing.T) {
	tmpDir := t.TempDir()

	// Create project structure
	yarToml := filepath.Join(tmpDir, "yar.toml")
	os.WriteFile(yarToml, []byte(`[package]
name = "test"
version = "0.1.0"
entry = "main.yar"
`), 0644)

	mainFile := filepath.Join(tmpDir, "main.yar")
	os.WriteFile(mainFile, []byte(`func main() { println("hello") }`), 0644)

	// Build
	builder := NewBuilder(tmpDir)

	err := builder.Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	// Check outputs
	buildDir := filepath.Join(tmpDir, "build")
	irFile := filepath.Join(buildDir, "ir", "main.ll")

	if _, err := os.Stat(irFile); err != nil {
		t.Errorf("IR file not created: %v", err)
	}
}

func TestIncrementalBuild(t *testing.T) {
	tmpDir := t.TempDir()

	// Create project
	yarToml := filepath.Join(tmpDir, "yar.toml")
	os.WriteFile(yarToml, []byte(`[package]
name = "test"
version = "0.1.0"
entry = "main.yar"
`), 0644)

	mainFile := filepath.Join(tmpDir, "main.yar")
	os.WriteFile(mainFile, []byte(`func main() { println("v1") }`), 0644)

	builder := NewBuilder(tmpDir)

	// First build
	err := builder.Build()
	if err != nil {
		t.Fatalf("first build failed: %v", err)
	}

	irFile := filepath.Join(tmpDir, "build", "ir", "main.ll")
	info1, _ := os.Stat(irFile)

	// Second build without changes - should use cache
	err = builder.Build()
	if err != nil {
		t.Fatalf("second build failed: %v", err)
	}

	info2, _ := os.Stat(irFile)

	// File should not be regenerated (same mod time)
	if !info1.ModTime().Equal(info2.ModTime()) {
		t.Error("IR file was regenerated unnecessarily")
	}

	// Modify source
	os.WriteFile(mainFile, []byte(`func main() { println("v2") }`), 0644)

	// Third build - should regenerate
	err = builder.Build()
	if err != nil {
		t.Fatalf("third build failed: %v", err)
	}

	info3, _ := os.Stat(irFile)

	if info2.ModTime().Equal(info3.ModTime()) {
		t.Error("IR file was not regenerated after source change")
	}
}
