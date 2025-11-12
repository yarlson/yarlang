package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yarlson/yarlang/build"
)

func TestSimpleMultiModule(t *testing.T) {
	projectDir := filepath.Join("testdata", "modules", "simple")

	// Build
	builder := build.NewBuilder(projectDir)

	err := builder.Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	// Run executable
	exePath := filepath.Join(projectDir, "build", "bin", "simple")
	cmd := exec.Command(exePath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("execution failed: %v\n%s", err, output)
	}

	// Check output
	expected, _ := os.ReadFile(filepath.Join(projectDir, "expected_output.txt"))

	actualLines := strings.Split(strings.TrimSpace(string(output)), "\n")
	expectedLines := strings.Split(strings.TrimSpace(string(expected)), "\n")

	if len(actualLines) != len(expectedLines) {
		t.Fatalf("output line count mismatch.\nExpected:\n%s\nGot:\n%s",
			expected, output)
	}

	for i, line := range expectedLines {
		if strings.TrimSpace(actualLines[i]) != strings.TrimSpace(line) {
			t.Errorf("line %d mismatch.\nExpected: %q\nGot: %q",
				i+1, line, actualLines[i])
		}
	}
}
