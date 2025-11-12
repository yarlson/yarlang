package analysis

import (
	"testing"

	"github.com/yarlson/yarlang/ast"
	"github.com/yarlson/yarlang/lexer"
	"github.com/yarlson/yarlang/parser"
)

func parseProgram(t *testing.T, input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	return program
}

func TestAnalyzeGlobalVariables(t *testing.T) {
	input := `
x = 42
y = "hello"
`
	program := parseProgram(t, input)

	symbols, diagnostics := Analyze(program)

	if len(diagnostics) > 0 {
		t.Errorf("Unexpected diagnostics: %v", diagnostics)
	}

	// Should have global scope with x and y
	if len(symbols.scopes) == 0 {
		t.Fatal("Expected at least one scope")
	}

	globalScope := symbols.scopes[0]

	if sym := globalScope.Lookup("x"); sym == nil {
		t.Error("Expected to find symbol 'x'")
	} else {
		if sym.Kind != SymbolKindVariable {
			t.Errorf("Symbol x kind = %v, want Variable", sym.Kind)
		}
	}

	if sym := globalScope.Lookup("y"); sym == nil {
		t.Error("Expected to find symbol 'y'")
	}
}

func TestAnalyzeUndefinedVariable(t *testing.T) {
	input := `z = x + 1`

	program := parseProgram(t, input)

	_, diagnostics := Analyze(program)

	// Should have error for undefined x
	if len(diagnostics) == 0 {
		t.Fatal("Expected diagnostic for undefined variable")
	}

	found := false

	for _, diag := range diagnostics {
		if diag.Severity == SeverityError &&
			diag.Message == "undefined: x" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected 'undefined: x' error, got: %v", diagnostics)
	}
}

func TestAnalyzeFunctionDeclaration(t *testing.T) {
	input := `
func add(a, b) {
	return a + b
}
`
	program := parseProgram(t, input)

	symbols, diagnostics := Analyze(program)

	if len(diagnostics) > 0 {
		t.Errorf("Unexpected diagnostics: %v", diagnostics)
	}

	// Check global scope has function
	globalScope := symbols.scopes[0]
	if sym := globalScope.Lookup("add"); sym == nil {
		t.Fatal("Expected to find symbol 'add'")
	} else {
		if sym.Kind != SymbolKindFunction {
			t.Errorf("Symbol add kind = %v, want Function", sym.Kind)
		}
	}

	// Should have function scope with parameters
	if len(symbols.scopes) < 2 {
		t.Fatal("Expected at least 2 scopes (global + function)")
	}

	// Find function scope
	var fnScope *Scope

	for _, scope := range symbols.scopes {
		if scope.LookupLocal("a") != nil {
			fnScope = scope
			break
		}
	}

	if fnScope == nil {
		t.Fatal("Expected to find function scope with parameter 'a'")
	}

	// Check parameters
	if sym := fnScope.LookupLocal("a"); sym == nil {
		t.Error("Expected to find parameter 'a'")
	} else {
		if sym.Kind != SymbolKindParameter {
			t.Errorf("Symbol a kind = %v, want Parameter", sym.Kind)
		}
	}

	if sym := fnScope.LookupLocal("b"); sym == nil {
		t.Error("Expected to find parameter 'b'")
	}
}

func TestAnalyzeNestedScopes(t *testing.T) {
	input := `
x = 10
if x > 5 {
	y = 20
}
`
	program := parseProgram(t, input)

	symbols, diagnostics := Analyze(program)

	if len(diagnostics) > 0 {
		t.Errorf("Unexpected diagnostics: %v", diagnostics)
	}

	// Should have at least 2 scopes (global + if block)
	if len(symbols.scopes) < 2 {
		t.Fatalf("Expected at least 2 scopes, got %d", len(symbols.scopes))
	}

	// Global scope should have x
	globalScope := symbols.scopes[0]
	if globalScope.LookupLocal("x") == nil {
		t.Error("Expected to find 'x' in global scope")
	}

	// Global scope should not have y locally
	if globalScope.LookupLocal("y") != nil {
		t.Error("Did not expect to find 'y' in global scope locally")
	}

	// Should be able to find a scope with y
	foundY := false

	for _, scope := range symbols.scopes {
		if scope.LookupLocal("y") != nil {
			foundY = true
			break
		}
	}

	if !foundY {
		t.Error("Expected to find 'y' in some scope")
	}
}

func TestAnalyzeForLoopScope(t *testing.T) {
	input := `
for i = 0; i < 10; i = i + 1 {
	x = i * 2
}
`
	program := parseProgram(t, input)

	symbols, diagnostics := Analyze(program)

	if len(diagnostics) > 0 {
		t.Errorf("Unexpected diagnostics: %v", diagnostics)
	}

	// Should have at least 2 scopes (global + for)
	if len(symbols.scopes) < 2 {
		t.Fatalf("Expected at least 2 scopes, got %d", len(symbols.scopes))
	}

	// Find for scope with 'i'
	var forScope *Scope

	for _, scope := range symbols.scopes {
		if scope.LookupLocal("i") != nil {
			forScope = scope
			break
		}
	}

	if forScope == nil {
		t.Fatal("Expected to find for scope with 'i'")
	}

	// For scope should have both i and x
	if forScope.Lookup("i") == nil {
		t.Error("Expected to find 'i' in for scope")
	}

	if forScope.Lookup("x") == nil {
		t.Error("Expected to find 'x' in for scope")
	}
}

func TestAnalyzeAssignmentCountMismatch(t *testing.T) {
	input := `x, y = 1`

	program := parseProgram(t, input)

	_, diagnostics := Analyze(program)

	// Should have error for assignment count mismatch
	if len(diagnostics) == 0 {
		t.Fatal("Expected diagnostic for assignment count mismatch")
	}

	found := false

	for _, diag := range diagnostics {
		if diag.Severity == SeverityError &&
			diag.Message == "assignment count mismatch: 2 = 1" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected 'assignment count mismatch' error, got: %v", diagnostics)
	}
}

func TestAnalyzeDuplicateFunctionDeclaration(t *testing.T) {
	input := `
func foo() {
	return 1
}
func foo() {
	return 2
}
`
	program := parseProgram(t, input)

	_, diagnostics := Analyze(program)

	// Should have error for duplicate function
	if len(diagnostics) == 0 {
		t.Fatal("Expected diagnostic for duplicate function declaration")
	}

	found := false

	for _, diag := range diagnostics {
		if diag.Severity == SeverityError &&
			diag.Message == "function foo already declared" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected 'function foo already declared' error, got: %v", diagnostics)
	}
}

func TestAnalyzeVariableReferences(t *testing.T) {
	input := `
x = 10
y = x + x
`
	program := parseProgram(t, input)

	symbols, diagnostics := Analyze(program)

	if len(diagnostics) > 0 {
		t.Errorf("Unexpected diagnostics: %v", diagnostics)
	}

	globalScope := symbols.scopes[0]

	xSym := globalScope.Lookup("x")
	if xSym == nil {
		t.Fatal("Expected to find symbol 'x'")
	}

	// x should be referenced twice in the second assignment
	if len(xSym.References) != 2 {
		t.Errorf("Expected 2 references to 'x', got %d", len(xSym.References))
	}
}

func TestAnalyzeComplexProgram(t *testing.T) {
	input := `
x = 10

func factorial(n) {
	if n <= 1 {
		return 1
	}
	return n * factorial(n - 1)
}

result = factorial(x)
`
	program := parseProgram(t, input)

	symbols, diagnostics := Analyze(program)

	if len(diagnostics) > 0 {
		t.Errorf("Unexpected diagnostics: %v", diagnostics)
	}

	// Check all expected symbols exist
	globalScope := symbols.scopes[0]

	if globalScope.Lookup("x") == nil {
		t.Error("Expected to find symbol 'x'")
	}

	if globalScope.Lookup("factorial") == nil {
		t.Error("Expected to find symbol 'factorial'")
	}

	if globalScope.Lookup("result") == nil {
		t.Error("Expected to find symbol 'result'")
	}

	// Find function scope and check parameter
	var fnScope *Scope

	for _, scope := range symbols.scopes {
		if scope.LookupLocal("n") != nil {
			fnScope = scope
			break
		}
	}

	if fnScope == nil {
		t.Fatal("Expected to find function scope with parameter 'n'")
	}

	// Should be able to see factorial from within the function (recursion)
	if fnScope.Lookup("factorial") == nil {
		t.Error("Expected to find 'factorial' from function scope (for recursion)")
	}
}

func TestAnalyzeBlockStatementScope(t *testing.T) {
	input := `
x = 1
{
	y = 2
	x = 3
}
`
	program := parseProgram(t, input)

	symbols, diagnostics := Analyze(program)

	if len(diagnostics) > 0 {
		t.Errorf("Unexpected diagnostics: %v", diagnostics)
	}

	// Should have global scope and block scope
	if len(symbols.scopes) < 2 {
		t.Fatalf("Expected at least 2 scopes, got %d", len(symbols.scopes))
	}

	globalScope := symbols.scopes[0]

	// x should be in global scope
	xSym := globalScope.LookupLocal("x")
	if xSym == nil {
		t.Fatal("Expected to find 'x' in global scope")
	}

	// The second assignment to x should be a reference
	if len(xSym.References) != 1 {
		t.Errorf("Expected 1 reference to 'x', got %d", len(xSym.References))
	}
}

func TestAnalyzeUndefinedInExpression(t *testing.T) {
	input := `
x = 10
y = x + z + w
`
	program := parseProgram(t, input)

	_, diagnostics := Analyze(program)

	// Should have errors for undefined z and w
	if len(diagnostics) < 2 {
		t.Fatalf("Expected at least 2 diagnostics, got %d", len(diagnostics))
	}

	errors := make(map[string]bool)

	for _, diag := range diagnostics {
		if diag.Severity == SeverityError {
			errors[diag.Message] = true
		}
	}

	if !errors["undefined: z"] {
		t.Error("Expected 'undefined: z' error")
	}

	if !errors["undefined: w"] {
		t.Error("Expected 'undefined: w' error")
	}
}

func TestAnalyzeElseBlock(t *testing.T) {
	input := `
x = 10
if x > 5 {
	y = 1
} else {
	z = 2
}
`
	program := parseProgram(t, input)

	symbols, diagnostics := Analyze(program)

	if len(diagnostics) > 0 {
		t.Errorf("Unexpected diagnostics: %v", diagnostics)
	}

	// Should have multiple scopes (global + then block + else block)
	if len(symbols.scopes) < 3 {
		t.Fatalf("Expected at least 3 scopes, got %d", len(symbols.scopes))
	}

	// Find scopes with y and z
	foundY := false
	foundZ := false

	for _, scope := range symbols.scopes {
		if scope.LookupLocal("y") != nil {
			foundY = true
		}

		if scope.LookupLocal("z") != nil {
			foundZ = true
		}
	}

	if !foundY {
		t.Error("Expected to find 'y' in then block scope")
	}

	if !foundZ {
		t.Error("Expected to find 'z' in else block scope")
	}
}

func TestAnalyzeFunctionCallExpr(t *testing.T) {
	input := `
func greet(name) {
	return "Hello, " + name
}

x = greet("World")
`
	program := parseProgram(t, input)

	symbols, diagnostics := Analyze(program)

	if len(diagnostics) > 0 {
		t.Errorf("Unexpected diagnostics: %v", diagnostics)
	}

	// Check that greet function has references
	globalScope := symbols.scopes[0]

	greetSym := globalScope.Lookup("greet")
	if greetSym == nil {
		t.Fatal("Expected to find symbol 'greet'")
	}

	// greet should be referenced once in the call
	if len(greetSym.References) != 1 {
		t.Errorf("Expected 1 reference to 'greet', got %d", len(greetSym.References))
	}
}
