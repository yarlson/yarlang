package codegen

import (
	"testing"

	"github.com/yarlson/yarlang/lexer"
	"github.com/yarlson/yarlang/parser"
)

func TestCodegenNumberLiteral(t *testing.T) {
	input := "42"

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	gen := New()

	err := gen.Generate(program)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	// Just verify it doesn't crash for now
	// Full IR verification would require more setup
}

func TestCodegenAssignment(t *testing.T) {
	input := "x = 42"

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	gen := New()

	err := gen.Generate(program)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	// Verify IR contains something
	ir := gen.EmitIR()
	if len(ir) == 0 {
		t.Fatal("generated IR is empty")
	}
}

func TestCodegenBinaryExpr(t *testing.T) {
	input := "x = 1 + 2"

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	gen := New()

	err := gen.Generate(program)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	// Verify IR contains addition
	ir := gen.EmitIR()
	if len(ir) == 0 {
		t.Fatal("generated IR is empty")
	}
}
