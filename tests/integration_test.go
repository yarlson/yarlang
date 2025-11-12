package tests

import (
	"os"
	"os/exec"
	"testing"
)

func TestCompileHello(t *testing.T) {
	// Build hello example
	cmd := exec.Command("../yar", "build", "../examples/hello.yar")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Build failed: %v\n%s", err, output)
	}

	// Check executable exists
	if _, err := os.Stat("../examples/hello"); err != nil {
		t.Fatal("Executable not created")
	}

	// Clean up
	_ = os.Remove("../examples/hello")
	_ = os.Remove("../examples/hello.ll")
}

func TestTypeCheckErrors(t *testing.T) {
	// Create invalid source
	source := `fn main() {
		let x: i32 = "string"
	}`

	tmpFile := "/tmp/test_invalid.yar"
	if err := os.WriteFile(tmpFile, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.Remove(tmpFile) }()

	// Should fail type check
	cmd := exec.Command("../yar", "check", tmpFile)
	if err := cmd.Run(); err == nil {
		t.Error("Expected type error, but check passed")
	}
}
