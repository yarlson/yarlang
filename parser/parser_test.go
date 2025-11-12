package parser

import (
	"github.com/yarlson/yarlang/ast"
	"github.com/yarlson/yarlang/lexer"
	"testing"
)

func TestParseNumberLiteral(t *testing.T) {
	input := "42"
	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("statement is not ExprStmt. got=%T", program.Statements[0])
	}

	literal, ok := stmt.Expr.(*ast.NumberLiteral)
	if !ok {
		t.Fatalf("expr is not NumberLiteral. got=%T", stmt.Expr)
	}

	if literal.Value != 42 {
		t.Errorf("literal.Value wrong. got=%f", literal.Value)
	}
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))

	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}

	t.FailNow()
}

func TestParseBinaryExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1 + 2", "(1 + 2)"},
		{"1 - 2", "(1 - 2)"},
		{"1 * 2", "(1 * 2)"},
		{"1 / 2", "(1 / 2)"},
		{"1 + 2 * 3", "(1 + (2 * 3))"},
		{"(1 + 2) * 3", "((1 + 2) * 3)"},
		{"1 < 2", "(1 < 2)"},
		{"1 == 2", "(1 == 2)"},
		{"true && false", "(true && false)"},
		{"true || false", "(true || false)"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExprStmt)
		if !ok {
			t.Fatalf("statement is not ExprStmt. got=%T", program.Statements[0])
		}

		if stmt.Expr.String() != tt.expected {
			t.Errorf("expression wrong. expected=%q, got=%q", tt.expected, stmt.Expr.String())
		}
	}
}

func TestParseAssignment(t *testing.T) {
	input := "x = 42"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("statement is not AssignStmt. got=%T", program.Statements[0])
	}

	if len(stmt.Targets) != 1 || stmt.Targets[0] != "x" {
		t.Errorf("targets wrong. got=%v", stmt.Targets)
	}

	if len(stmt.Values) != 1 {
		t.Errorf("values count wrong. got=%d", len(stmt.Values))
	}
}

func TestParseFunctionDecl(t *testing.T) {
	input := `
func add(a, b) {
	return a + b
}
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
	}

	fn, ok := program.Statements[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("statement is not FuncDecl. got=%T", program.Statements[0])
	}

	if fn.Name != "add" {
		t.Errorf("function name wrong. got=%q", fn.Name)
	}

	if len(fn.Params) != 2 {
		t.Errorf("params count wrong. got=%d", len(fn.Params))
	}

	if fn.Params[0] != "a" || fn.Params[1] != "b" {
		t.Errorf("params wrong. got=%v", fn.Params)
	}
}

func TestParseIfElseStatements(t *testing.T) {
	tests := []struct {
		input string
		desc  string
	}{
		{
			input: `if x > 0 { return x }`,
			desc:  "if without else",
		},
		{
			input: `if x > 0 { return x } else { return 0 }`,
			desc:  "if with else",
		},
		{
			input: `if x > 0 { return 1 } else if x < 0 { return -1 } else { return 0 }`,
			desc:  "if with else if and else",
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.IfStmt)
			if !ok {
				t.Fatalf("statement is not IfStmt. got=%T", program.Statements[0])
			}

			if stmt.Condition == nil {
				t.Error("condition is nil")
			}

			if stmt.ThenBlock == nil {
				t.Error("then block is nil")
			}
		})
	}
}

func TestParseForLoops(t *testing.T) {
	tests := []struct {
		input string
		desc  string
	}{
		{
			input: `for { break }`,
			desc:  "infinite loop",
		},
		{
			input: `for x < 10 { x = x + 1 }`,
			desc:  "while-style loop",
		},
		{
			input: `for i = 0; i < 10; i = i + 1 { print(i) }`,
			desc:  "3-part C-style loop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.ForStmt)
			if !ok {
				t.Fatalf("statement is not ForStmt. got=%T", program.Statements[0])
			}

			if stmt.Body == nil {
				t.Error("body is nil")
			}
		})
	}
}

func TestParseUnaryExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"!true", "(!true)"},
		{"!false", "(!false)"},
		{"-42", "(-42)"},
		{"-x", "(-x)"},
		{"!!true", "(!(!true))"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExprStmt)
		if !ok {
			t.Fatalf("statement is not ExprStmt. got=%T", program.Statements[0])
		}

		if stmt.Expr.String() != tt.expected {
			t.Errorf("expression wrong. expected=%q, got=%q", tt.expected, stmt.Expr.String())
		}
	}
}

func TestParseCallExpressions(t *testing.T) {
	tests := []struct {
		input    string
		funcName string
		argCount int
	}{
		{"foo()", "foo", 0},
		{"bar(x)", "bar", 1},
		{"add(1, 2)", "add", 2},
		{"max(a, b, c)", "max", 3},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExprStmt)
		if !ok {
			t.Fatalf("statement is not ExprStmt. got=%T", program.Statements[0])
		}

		call, ok := stmt.Expr.(*ast.CallExpr)
		if !ok {
			t.Fatalf("expr is not CallExpr. got=%T", stmt.Expr)
		}

		ident, ok := call.Function.(*ast.Identifier)
		if !ok {
			t.Fatalf("function is not Identifier. got=%T", call.Function)
		}

		if ident.Name != tt.funcName {
			t.Errorf("function name wrong. expected=%q, got=%q", tt.funcName, ident.Name)
		}

		if len(call.Args) != tt.argCount {
			t.Errorf("arg count wrong. expected=%d, got=%d", tt.argCount, len(call.Args))
		}
	}
}

func TestParseMultipleAssignment(t *testing.T) {
	input := "x, y = 1, 2"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has wrong number of statements. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("statement is not AssignStmt. got=%T", program.Statements[0])
	}

	if len(stmt.Targets) != 2 {
		t.Errorf("targets count wrong. expected=2, got=%d", len(stmt.Targets))
	}

	if stmt.Targets[0] != "x" || stmt.Targets[1] != "y" {
		t.Errorf("targets wrong. got=%v", stmt.Targets)
	}

	if len(stmt.Values) != 2 {
		t.Errorf("values count wrong. expected=2, got=%d", len(stmt.Values))
	}
}

func TestParseCompleteProgram(t *testing.T) {
	input := `
func factorial(n) {
	if n <= 1 {
		return 1
	} else {
		return n * factorial(n - 1)
	}
}

func main() {
	result = 1
	for i = 1; i < 10; i = i + 1 {
		result = result * i
	}
	return result
}
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 2 {
		t.Fatalf("program has wrong number of statements. expected=2, got=%d", len(program.Statements))
	}

	// Check first function
	fn1, ok := program.Statements[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("first statement is not FuncDecl. got=%T", program.Statements[0])
	}

	if fn1.Name != "factorial" {
		t.Errorf("first function name wrong. expected=factorial, got=%q", fn1.Name)
	}

	if len(fn1.Params) != 1 {
		t.Errorf("first function params count wrong. expected=1, got=%d", len(fn1.Params))
	}

	// Check second function
	fn2, ok := program.Statements[1].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("second statement is not FuncDecl. got=%T", program.Statements[1])
	}

	if fn2.Name != "main" {
		t.Errorf("second function name wrong. expected=main, got=%q", fn2.Name)
	}

	if len(fn2.Body.Statements) != 3 {
		t.Errorf("main function body statement count wrong. expected=3, got=%d", len(fn2.Body.Statements))
	}

	// Check for loop in main
	forStmt, ok := fn2.Body.Statements[1].(*ast.ForStmt)
	if !ok {
		t.Fatalf("second statement in main is not ForStmt. got=%T", fn2.Body.Statements[1])
	}

	if forStmt.Init == nil {
		t.Error("for loop init is nil")
	}

	if forStmt.Condition == nil {
		t.Error("for loop condition is nil")
	}

	if forStmt.Post == nil {
		t.Error("for loop post is nil")
	}
}

func TestPositionHelpers(t *testing.T) {
	input := "foo"
	l := lexer.New(input)
	p := New(l)

	start := p.tokenStartPos()
	if start.Line != 1 || start.Column != 1 {
		t.Errorf("tokenStartPos() = %+v, want Line=1 Column=1", start)
	}

	end := p.tokenEndPos()
	if end.Line != 1 || end.Column != 4 {
		t.Errorf("tokenEndPos() = %+v, want Line=1 Column=4", end)
	}

	r := p.tokenRange()
	if r.Start.Line != 1 || r.Start.Column != 1 || r.End.Column != 4 {
		t.Errorf("tokenRange() = %+v, want Start(1,1) End(1,4)", r)
	}
}

func TestParseIdentifierWithPosition(t *testing.T) {
	input := "foobar"
	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	exprStmt, ok := program.Statements[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("Expected ExprStmt, got %T", program.Statements[0])
	}

	ident, ok := exprStmt.Expr.(*ast.Identifier)
	if !ok {
		t.Fatalf("Expected Identifier, got %T", exprStmt.Expr)
	}

	if ident.Name != "foobar" {
		t.Errorf("Name = %s, want foobar", ident.Name)
	}

	if ident.Range.Start.Line != 1 || ident.Range.Start.Column != 1 {
		t.Errorf("Start = %+v, want Line=1 Column=1", ident.Range.Start)
	}

	if ident.Range.End.Column != 7 {
		t.Errorf("End column = %d, want 7", ident.Range.End.Column)
	}
}

func TestParseLiteralsWithPosition(t *testing.T) {
	tests := []struct {
		input      string
		checkRange func(t *testing.T, expr ast.Expr)
	}{
		{
			input: "42",
			checkRange: func(t *testing.T, expr ast.Expr) {
				num := expr.(*ast.NumberLiteral)
				if num.Range.Start.Column != 1 || num.Range.End.Column != 3 {
					t.Errorf("NumberLiteral range = %+v, want columns 1-3", num.Range)
				}
			},
		},
		{
			input: `"hello"`,
			checkRange: func(t *testing.T, expr ast.Expr) {
				str := expr.(*ast.StringLiteral)
				if str.Range.Start.Column != 1 || str.Range.End.Column != 8 {
					t.Errorf("StringLiteral range = %+v, want columns 1-8", str.Range)
				}
			},
		},
		{
			input: "true",
			checkRange: func(t *testing.T, expr ast.Expr) {
				b := expr.(*ast.BoolLiteral)
				if b.Range.Start.Column != 1 || b.Range.End.Column != 5 {
					t.Errorf("BoolLiteral range = %+v, want columns 1-5", b.Range)
				}
			},
		},
		{
			input: "nil",
			checkRange: func(t *testing.T, expr ast.Expr) {
				n := expr.(*ast.NilLiteral)
				if n.Range.Start.Column != 1 || n.Range.End.Column != 4 {
					t.Errorf("NilLiteral range = %+v, want columns 1-4", n.Range)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			exprStmt := program.Statements[0].(*ast.ExprStmt)
			tt.checkRange(t, exprStmt.Expr)
		})
	}
}

func TestParseExpressionsWithPosition(t *testing.T) {
	input := "1 + 2"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Statements[0].(*ast.ExprStmt)
	binary := exprStmt.Expr.(*ast.BinaryExpr)

	// Binary expression should span from first operand to last
	if binary.Range.Start.Column != 1 {
		t.Errorf("BinaryExpr start column = %d, want 1", binary.Range.Start.Column)
	}

	if binary.Range.End.Column != 6 {
		t.Errorf("BinaryExpr end column = %d, want 6", binary.Range.End.Column)
	}
}

func TestParseUnaryExprWithPosition(t *testing.T) {
	input := "!true"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Statements[0].(*ast.ExprStmt)
	unary := exprStmt.Expr.(*ast.UnaryExpr)

	if unary.Range.Start.Column != 1 {
		t.Errorf("UnaryExpr start column = %d, want 1", unary.Range.Start.Column)
	}

	if unary.Range.End.Column != 6 {
		t.Errorf("UnaryExpr end column = %d, want 6", unary.Range.End.Column)
	}
}

func TestParseCallExprWithPosition(t *testing.T) {
	input := "add(1, 2)"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Statements[0].(*ast.ExprStmt)
	call := exprStmt.Expr.(*ast.CallExpr)

	if call.Range.Start.Column != 1 {
		t.Errorf("CallExpr start column = %d, want 1", call.Range.Start.Column)
	}

	if call.Range.End.Column != 10 {
		t.Errorf("CallExpr end column = %d, want 10", call.Range.End.Column)
	}
}

func TestParseSimpleStatementsWithPosition(t *testing.T) {
	tests := []struct {
		input      string
		checkRange func(t *testing.T, stmt ast.Stmt)
	}{
		{
			input: "return 42",
			checkRange: func(t *testing.T, stmt ast.Stmt) {
				ret := stmt.(*ast.ReturnStmt)
				if ret.Range.Start.Column != 1 {
					t.Errorf("ReturnStmt start = %d, want 1", ret.Range.Start.Column)
				}
			},
		},
		{
			input: "break",
			checkRange: func(t *testing.T, stmt ast.Stmt) {
				brk := stmt.(*ast.BreakStmt)
				if brk.Range.Start.Column != 1 || brk.Range.End.Column != 6 {
					t.Errorf("BreakStmt range = %+v", brk.Range)
				}
			},
		},
		{
			input: "continue",
			checkRange: func(t *testing.T, stmt ast.Stmt) {
				cont := stmt.(*ast.ContinueStmt)
				if cont.Range.Start.Column != 1 || cont.Range.End.Column != 9 {
					t.Errorf("ContinueStmt range = %+v", cont.Range)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)
			tt.checkRange(t, program.Statements[0])
		})
	}
}

func TestParseAssignStmtWithPosition(t *testing.T) {
	input := "x, y = 1, 2"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	assign := program.Statements[0].(*ast.AssignStmt)

	if assign.Range.Start.Column != 1 {
		t.Errorf("AssignStmt start = %d, want 1", assign.Range.Start.Column)
	}

	if len(assign.TargetRanges) != 2 {
		t.Fatalf("Expected 2 target ranges, got %d", len(assign.TargetRanges))
	}

	// First target "x" at column 1
	if assign.TargetRanges[0].Start.Column != 1 {
		t.Errorf("Target 0 start = %d, want 1", assign.TargetRanges[0].Start.Column)
	}

	// Second target "y" at column 4
	if assign.TargetRanges[1].Start.Column != 4 {
		t.Errorf("Target 1 start = %d, want 4", assign.TargetRanges[1].Start.Column)
	}
}

func TestParseComplexStatementsWithPosition(t *testing.T) {
	input := `
func add(a, b) {
	return a + b
}
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn := program.Statements[0].(*ast.FuncDecl)

	if fn.Name != "add" {
		t.Fatalf("Function name = %s, want add", fn.Name)
	}

	// Check function has valid range
	if fn.Range.Start.Line == 0 {
		t.Error("FuncDecl range not set")
	}

	// Check name range is set
	if fn.NameRange.Start.Line == 0 {
		t.Error("FuncDecl NameRange not set")
	}

	// Check param ranges
	if len(fn.ParamRanges) != 2 {
		t.Errorf("Expected 2 param ranges, got %d", len(fn.ParamRanges))
	}
}

func TestParseImportStatement(t *testing.T) {
	input := `import "math"`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements should have 1 statement. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ImportStmt)
	if !ok {
		t.Fatalf("statement is not *ast.ImportStmt. got=%T",
			program.Statements[0])
	}

	if stmt.Path != "math" {
		t.Errorf("import path wrong. expected=%q, got=%q",
			"math", stmt.Path)
	}

	if stmt.Alias != "" {
		t.Errorf("import alias should be empty. got=%q", stmt.Alias)
	}
}

func TestParseImportWithAlias(t *testing.T) {
	input := `import "math" as m`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ImportStmt)
	if !ok {
		t.Fatalf("not *ast.ImportStmt. got=%T", program.Statements[0])
	}

	if stmt.Path != "math" {
		t.Errorf("path wrong. expected=%q, got=%q", "math", stmt.Path)
	}

	if stmt.Alias != "m" {
		t.Errorf("alias wrong. expected=%q, got=%q", "m", stmt.Alias)
	}
}

func TestParseImportBlock(t *testing.T) {
	input := `import (
    "math"
    "strings" as str
    "io"
)`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement. got=%d", len(program.Statements))
	}

	block, ok := program.Statements[0].(*ast.ImportBlock)
	if !ok {
		t.Fatalf("not *ast.ImportBlock. got=%T", program.Statements[0])
	}

	if len(block.Imports) != 3 {
		t.Fatalf("expected 3 imports. got=%d", len(block.Imports))
	}

	tests := []struct {
		expectedPath  string
		expectedAlias string
	}{
		{"math", ""},
		{"strings", "str"},
		{"io", ""},
	}

	for i, tt := range tests {
		imp := block.Imports[i]
		if imp.Path != tt.expectedPath {
			t.Errorf("import[%d] path wrong. expected=%q, got=%q",
				i, tt.expectedPath, imp.Path)
		}

		if imp.Alias != tt.expectedAlias {
			t.Errorf("import[%d] alias wrong. expected=%q, got=%q",
				i, tt.expectedAlias, imp.Alias)
		}
	}
}
