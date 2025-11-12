package semantic

import (
	"strings"
	"testing"

	"github.com/yarlson/yarlang/ast"
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

func TestCrossModuleSymbolResolution(t *testing.T) {
	// Module: math.yar
	mathSource := `func Sqrt(x) {
	return x
}

func internal() {
	return 42
}
`

	// Parse math module
	mathAST := parseSource(t, mathSource)

	// Create main module AST manually since parser doesn't support module.Function syntax yet
	mainAST := &ast.Program{
		Statements: []ast.Stmt{
			&ast.ImportStmt{
				Path: "math",
			},
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

	// Create module info
	modules := map[string]*ModuleInfo{
		"math": {
			Name: "math",
			AST:  mathAST,
		},
		"main": {
			Name: "main",
			AST:  mainAST,
		},
	}

	analyzer := NewCrossModuleAnalyzer(modules)

	err := analyzer.Analyze("main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnexportedSymbolError(t *testing.T) {
	mathSource := `func internal() {
	return 42
}
`

	mathAST := parseSource(t, mathSource)

	// Create main module AST manually
	mainAST := &ast.Program{
		Statements: []ast.Stmt{
			&ast.ImportStmt{
				Path: "math",
			},
			&ast.FuncDecl{
				Name:   "main",
				Params: []string{},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.AssignStmt{
							Targets: []string{"x"},
							Values: []ast.Expr{
								&ast.CallExpr{
									Function: &ast.Identifier{Name: "math.internal"},
									Args:     []ast.Expr{},
								},
							},
						},
					},
				},
			},
		},
	}

	modules := map[string]*ModuleInfo{
		"math": {Name: "math", AST: mathAST},
		"main": {Name: "main", AST: mainAST},
	}

	analyzer := NewCrossModuleAnalyzer(modules)

	err := analyzer.Analyze("main")
	if err == nil {
		t.Fatal("expected error for unexported symbol")
	}

	if !strings.Contains(err.Error(), "not exported") {
		t.Errorf("expected 'not exported' error, got: %v", err)
	}
}

func parseSource(t *testing.T, source string) *ast.Program {
	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	return program
}
