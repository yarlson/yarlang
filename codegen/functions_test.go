package codegen

import (
	"testing"

	"github.com/yarlson/yarlang/lexer"
	"github.com/yarlson/yarlang/parser"
)

func TestCodegenFunctionDecl(t *testing.T) {
	input := `
func add(a, b) {
	return a + b
}

result = add(5, 3)
println(result)
`

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

	// Verify function is declared
	ir := gen.EmitIR()
	if !contains(ir, "define") {
		t.Error("expected function definition in IR")
	}
}

func TestCodegenRecursiveFunction(t *testing.T) {
	input := `
func factorial(n) {
	if n <= 1 {
		return 1
	}
	prev = factorial(n - 1)
	return n * prev
}

result = factorial(5)
println(result)
`

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
}
