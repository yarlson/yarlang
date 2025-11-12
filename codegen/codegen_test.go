package codegen

import (
	"strings"
	"testing"

	"github.com/yarlson/yarlang/ast"
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
	input := `
func main() {
	x = 42
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

	// Verify IR contains main function with assignment
	ir := gen.EmitIR()
	if len(ir) == 0 {
		t.Fatal("generated IR is empty")
	}

	if !strings.Contains(ir, "alloca") {
		t.Fatal("expected alloca instruction for variable assignment")
	}
}

func TestCodegenBinaryExpr(t *testing.T) {
	input := `
func main() {
	x = 1 + 2
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

	// Verify IR contains main function with binary operation
	ir := gen.EmitIR()
	if len(ir) == 0 {
		t.Fatal("generated IR is empty")
	}

	if !strings.Contains(ir, "alloca") {
		t.Fatal("expected alloca instruction for variable assignment")
	}

	if !strings.Contains(ir, "yar_add") {
		t.Fatal("expected call to yar_add for binary addition")
	}
}

func TestGenerateExternalDeclarations(t *testing.T) {
	// Manually construct AST for a main function that calls math.Sqrt
	// We can't parse "math.Sqrt" yet since the parser doesn't handle qualified names
	// This will be supported when the full module system is integrated
	program := &ast.Program{
		Statements: []ast.Stmt{
			&ast.FuncDecl{
				Name:   "main",
				Params: []string{},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.AssignStmt{
							Targets: []string{"x"},
							Values: []ast.Expr{
								&ast.CallExpr{
									Function: &ast.Identifier{Name: "math.Sqrt"},
									Args: []ast.Expr{
										&ast.NumberLiteral{Value: 16},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	gen := New()
	gen.SetModuleName("main")

	// Register external function from math module
	gen.RegisterExternalFunction("math", "Sqrt", 1)

	err := gen.Generate(program)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	ir := gen.EmitIR()

	// Check for external declaration (using modern LLVM opaque pointer syntax)
	if !strings.Contains(ir, "declare ptr @math_Sqrt(ptr)") {
		t.Error("missing external declaration for math.Sqrt")
		t.Logf("Generated IR:\n%s", ir)
	}

	// Check for call to external function
	if !strings.Contains(ir, "call ptr @math_Sqrt") {
		t.Error("missing call to math.Sqrt")
		t.Logf("Generated IR:\n%s", ir)
	}
}
