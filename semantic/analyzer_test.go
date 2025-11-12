package semantic

import (
	"testing"

	"github.com/yarlson/yarlang/lexer"
	"github.com/yarlson/yarlang/parser"
)

func TestUndefinedVariableError(t *testing.T) {
	input := `
x = y + 1
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	analyzer := New()

	err := analyzer.Analyze(program)
	if err == nil {
		t.Fatal("expected error for undefined variable, got nil")
	}

	if err.Error() != "undefined variable: y" {
		t.Errorf("wrong error message. got=%q", err.Error())
	}
}

func TestValidProgram(t *testing.T) {
	input := `
x = 42
y = x + 1
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	analyzer := New()

	err := analyzer.Analyze(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFunctionScope(t *testing.T) {
	input := `
func foo(x) {
	y = x + 1
	return y
}

z = y  // y should be undefined here
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	analyzer := New()

	err := analyzer.Analyze(program)
	if err == nil {
		t.Fatal("expected error for undefined variable, got nil")
	}

	if err.Error() != "undefined variable: y" {
		t.Errorf("wrong error message. got=%q", err.Error())
	}
}

func TestBuiltInFunctions(t *testing.T) {
	input := `
print("hello")
println("world")
x = len("test")
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	analyzer := New()

	err := analyzer.Analyze(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
