package parser

import (
	"strings"
	"testing"

	"github.com/yarlson/yarlang/lexer"
)

func TestParseFile(t *testing.T) {
	input := `module demo

use std::io::File

struct Point {
	x: f64,
	y: f64,
}

impl Point {
	fn new(x f64, y f64) Point {
		return Point{x: x, y: y}
	}

	fn len(&self) f64 {
		return 0.0
	}
}

fn main() {
	let x = 5
	println("hello")
}
`

	l := lexer.New(input)
	p := New(l)
	file := p.ParseFile()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	// Verify module
	if len(file.Module) != 1 || file.Module[0] != "demo" {
		t.Errorf("wrong module: %v", file.Module)
	}

	// Verify we got declarations
	if len(file.Items) < 4 {
		t.Errorf("expected at least 4 declarations, got %d", len(file.Items))
	}

	// Verify string output contains key elements
	result := file.String()

	keywords := []string{"use", "struct", "Point", "impl", "fn", "main"}
	for _, kw := range keywords {
		if !strings.Contains(result, kw) {
			t.Errorf("expected file to contain %q", kw)
		}
	}
}
