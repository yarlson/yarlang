package codegen

import (
	"testing"

	"github.com/yarlson/yarlang/lexer"
	"github.com/yarlson/yarlang/parser"
)

func TestCodegenIfStmt(t *testing.T) {
	input := `
x = 10
if x > 5 {
	println(x)
}
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

	// Verify IR contains basic blocks for if statement
	ir := gen.EmitIR()
	if !contains(ir, "br i1") {
		t.Error("expected conditional branch in IR")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || (len(s) >= len(substr) &&
			(s[:len(substr)] == substr || contains(s[1:], substr))))
}

func TestCodegenIfElseStmt(t *testing.T) {
	input := `
x = 3
if x > 5 {
	println(10)
} else {
	println(20)
}
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

	// Verify IR contains else block
	ir := gen.EmitIR()
	if !contains(ir, "else") {
		t.Error("expected 'else' block in IR")
	}
}

func TestCodegenIfElseIfStmt(t *testing.T) {
	input := `
x = 7
if x > 10 {
	println(10)
} else if x > 5 {
	println(20)
} else {
	println(30)
}
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

	// Verify IR is generated successfully and contains conditional branches
	ir := gen.EmitIR()
	if !contains(ir, "br i1") {
		t.Error("expected conditional branches in IR for else-if chaining")
	}

	// Test passes if we can generate IR without errors
	// The presence of conditional branches indicates proper else-if handling
}

func TestCodegenForLoop(t *testing.T) {
	input := `
sum = 0
for i = 0; i < 5; i = i + 1 {
	sum = sum + i
}
println(sum)
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

	// Verify IR contains loop blocks
	ir := gen.EmitIR()
	if !contains(ir, "br label") {
		t.Error("expected loop branch in IR")
	}
}

func TestCodegenBreak(t *testing.T) {
	input := `
for i = 0; i < 10; i = i + 1 {
	if i > 3 {
		break
	}
	println(i)
}
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
