package parser

import (
	"strings"
	"testing"

	"github.com/yarlson/yarlang/lexer"
)

func TestParseTypes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Basic types
		{"i32", "i32"},
		{"f64", "f64"},
		{"bool", "bool"},
		{"void", "void"},

		// Reference types
		{"&T", "&T"},
		{"&mut T", "&mut T"},
		{"&i32", "&i32"},

		// Pointer types
		{"*T", "*T"},
		{"*i32", "*i32"},
		{"*f64", "*f64"},

		// Slice types
		{"[]i32", "[]i32"},
		{"[]f64", "[]f64"},
		{"[]bool", "[]bool"},

		// Array types
		{"[i32; 10]", "[i32; 10]"},
		{"[f64; 5]", "[f64; 5]"},

		// Tuple types
		{"(i32, f64)", "(i32, f64)"},
		{"()", "()"},
		{"(i32, f64, bool)", "(i32, f64, bool)"},

		// Generic types
		{"Vec<T>", "Vec<T>"},
		{"Result<i32, Error>", "Result<i32, Error>"},
		{"HashMap<String, i32>", "HashMap<String, i32>"},

		// Namespaced paths
		{"std::io::File", "std::io::File"},
		{"std::collections::HashMap", "std::collections::HashMap"},
		{"my::pkg::Type", "my::pkg::Type"},

		// Nested types
		{"&[]i32", "&[]i32"},
		{"Vec<Vec<i32> >", "Vec<Vec<i32>>"},
		{"&Result<Vec<i32>, Error>", "&Result<Vec<i32>, Error>"},
		{"*[]i32", "*[]i32"},
		{"&[i32; 10]", "&[i32; 10]"},

		// Complex nested types
		{"Vec<(i32, f64)>", "Vec<(i32, f64)>"},
		{"Result<Vec<i32>, std::io::Error>", "Result<Vec<i32>, std::io::Error>"},
		{"&mut Vec<String>", "&mut Vec<String>"},
		{"HashMap<String, Vec<i32> >", "HashMap<String, Vec<i32>>"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		typ := p.parseType()

		if len(p.Errors()) != 0 {
			t.Fatalf("parser has %d errors for input %q: %v", len(p.Errors()), tt.input, p.Errors())
		}

		// Ensure we have a valid ast.Type
		var _ = typ

		if typ.String() != tt.expected {
			t.Errorf("wrong type for input %q. expected=%q, got=%q", tt.input, tt.expected, typ.String())
		}
	}
}

func TestParseBinaryExpr(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1 + 2", "(1 + 2)"},
		{"1 + 2 * 3", "(1 + (2 * 3))"},
		{"(1 + 2) * 3", "((1 + 2) * 3)"},
		{"a == b", "(a == b)"},
		{"a < b && c > d", "((a < b) && (c > d))"},
		{"a | b ^ c & d", "(a | (b ^ (c & d)))"},
		{"1 << 2 + 3", "(1 << (2 + 3))"},
		// NOTE: Assignment is a statement, not an expression in YarLang
		// {"a = b = c", "(a = (b = c))"},
		// {"x += y += z", "(x += (y += z))"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		expr := p.parseExpression(LOWEST)

		if len(p.Errors()) != 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		if expr.String() != tt.expected {
			t.Errorf("wrong expr. expected=%q, got=%q", tt.expected, expr.String())
		}
	}
}

func TestParseUnaryExpr(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"-5", "(-5)"},
		{"!true", "(!true)"},
		{"~x", "(~x)"},
		{"&x", "(&x)"},
		{"*p", "(*p)"},
		{"-a + b", "((-a) + b)"},
		{"&mut x", "(&mut x)"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		expr := p.parseExpression(LOWEST)

		if len(p.Errors()) != 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		if expr.String() != tt.expected {
			t.Errorf("wrong expr. expected=%q, got=%q", tt.expected, expr.String())
		}
	}
}

func TestParsePostfixExpr(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"f()", "f()"},
		{"f(a)", "f(a)"},
		{"f(a, b)", "f(a, b)"},
		{"arr[0]", "arr[0]"},
		{"p.x", "p.x"},
		{"result?", "result?"},
		{"f()?", "f()?"},
		{"p.x.y", "p.x.y"},
		{"arr[i][j]", "arr[i][j]"},
		{"f().g()", "f().g()"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		expr := p.parseExpression(LOWEST)

		if len(p.Errors()) != 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		if expr.String() != tt.expected {
			t.Errorf("wrong expr. expected=%q, got=%q", tt.expected, expr.String())
		}
	}
}

func TestParseLiterals(t *testing.T) {
	tests := []struct {
		input    string
		contains []string // strings that should be in output
	}{
		{"[1, 2, 3]", []string{"1", "2", "3", "[", "]"}},
		{"[]", []string{"[", "]"}},
		{"(1, 2)", []string{"1", "2"}},
		{"Point{x: 1, y: 2}", []string{"Point", "x", "y", "1", "2"}},
		{"Point{x: 1.0, y: 2.0}", []string{"Point", "x", "y", "1", "2"}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		expr := p.parseExpression(LOWEST)

		if len(p.Errors()) != 0 {
			t.Fatalf("parser errors for %q: %v", tt.input, p.Errors())
		}

		// Check that all expected strings are present
		result := expr.String()
		for _, expected := range tt.contains {
			if !strings.Contains(result, expected) {
				t.Errorf("wrong expr for %q. expected to contain %q, got=%q", tt.input, expected, result)
			}
		}
	}
}

func TestParseLetStmt(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"let x = 5", "let x = 5"},
		{"let mut x = 5", "let mut x = 5"},
		{"let x: i32 = 5", "let x: i32 = 5"},
		{"let mut x: i32 = 5", "let mut x: i32 = 5"},
		{"let x = 1 + 2", "let x = (1 + 2)"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		stmt := p.parseStatement()

		if len(p.Errors()) != 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		if stmt.String() != tt.expected {
			t.Errorf("wrong stmt. expected=%q, got=%q", tt.expected, stmt.String())
		}
	}
}

func TestParseAssignStmt(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"x = 5", "x = 5"},
		{"x = 1 + 2", "x = (1 + 2)"},
		{"x += 5", "x += 5"},
		{"x -= 5", "x -= 5"},
		{"x *= 5", "x *= 5"},
		{"arr[i] = 10", "arr[i] = 10"},
		{"p.x = 1", "p.x = 1"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		stmt := p.parseStatement()

		if len(p.Errors()) != 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		if stmt.String() != tt.expected {
			t.Errorf("wrong stmt. expected=%q, got=%q", tt.expected, stmt.String())
		}
	}
}

func TestParseShortDecl(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"x := 5", "x := 5"},
		{"x := 1 + 2", "x := (1 + 2)"},
		{"result := f()", "result := f()"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		stmt := p.parseStatement()

		if len(p.Errors()) != 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		if stmt.String() != tt.expected {
			t.Errorf("wrong stmt. expected=%q, got=%q", tt.expected, stmt.String())
		}
	}
}

func TestParseIfStmt(t *testing.T) {
	tests := []struct {
		input    string
		contains []string
	}{
		{"if x > 0 { y = 1 }", []string{"if", "x", ">", "0", "y", "=", "1"}},
		{"if x > 0 { y = 1 } else { y = 0 }", []string{"if", "else", "y", "=", "1", "y", "=", "0"}},
		{"if x > 0 { y = 1 } else if x < 0 { y = -1 }", []string{"if", "else", "if", "y", "=", "1", "y", "=", "-1"}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		stmt := p.parseStatement()

		if len(p.Errors()) != 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		result := stmt.String()
		for _, str := range tt.contains {
			if !strings.Contains(result, str) {
				t.Errorf("expected %q to contain %q", result, str)
			}
		}
	}
}

func TestParseLoops(t *testing.T) {
	tests := []struct {
		input    string
		contains []string
	}{
		{"while x > 0 { x = x - 1 }", []string{"while", "x", ">", "0"}},
		{"for i in 0..10 { println(i) }", []string{"for", "i", "in"}},
		{"for i, val in arr[0] { val }", []string{"for", "i", "val", "in"}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		stmt := p.parseStatement()

		if len(p.Errors()) != 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		result := stmt.String()
		for _, str := range tt.contains {
			if !strings.Contains(result, str) {
				t.Errorf("expected %q to contain %q", result, str)
			}
		}
	}
}

func TestParseSimpleStmts(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"return", "return"},
		{"return 5", "return 5"},
		{"return 1 + 2", "return (1 + 2)"},
		{"break", "break"},
		{"continue", "continue"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		stmt := p.parseStatement()

		if len(p.Errors()) != 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		if stmt.String() != tt.expected {
			t.Errorf("wrong stmt. expected=%q, got=%q", tt.expected, stmt.String())
		}
	}
}

func TestParseDeferAndUnsafe(t *testing.T) {
	tests := []struct {
		input    string
		contains []string
	}{
		{"defer f()", []string{"defer", "f()"}},
		{"defer x.close()", []string{"defer", "close()"}},
		{"unsafe { *p = 5 }", []string{"unsafe", "*p", "=", "5"}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		stmt := p.parseStatement()

		if len(p.Errors()) != 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		result := stmt.String()
		for _, str := range tt.contains {
			if !strings.Contains(result, str) {
				t.Errorf("expected %q to contain %q", result, str)
			}
		}
	}
}

func TestParseFuncDecl(t *testing.T) {
	tests := []struct {
		input    string
		contains []string
	}{
		{"fn add(a i32, b i32) i32 { return a + b }", []string{"fn", "add", "a", "i32", "b", "i32"}},
		{"pub fn mul(x i32, y i32) i32 { return x * y }", []string{"pub", "fn", "mul"}},
		{"fn generic<T>(x T) T { return x }", []string{"fn", "generic", "<T>"}},
		{"fn void_fn() { println(\"hi\") }", []string{"fn", "void_fn"}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		decl := p.parseDeclaration()

		if len(p.Errors()) != 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		result := decl.String()
		for _, str := range tt.contains {
			if !strings.Contains(result, str) {
				t.Errorf("expected %q to contain %q", result, str)
			}
		}
	}
}

func TestParseStructDecl(t *testing.T) {
	tests := []struct {
		input    string
		contains []string
	}{
		{"struct Point { x: f64, y: f64 }", []string{"struct", "Point", "x", "f64", "y"}},
		{"pub struct Vec<T> { data: *T, len: usize }", []string{"pub", "struct", "Vec", "<T>", "data", "len", "usize"}},
		{"struct Empty {}", []string{"struct", "Empty"}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		decl := p.parseDeclaration()

		if len(p.Errors()) != 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		result := decl.String()
		for _, str := range tt.contains {
			if !strings.Contains(result, str) {
				t.Errorf("expected %q to contain %q", result, str)
			}
		}
	}
}

func TestParseEnumDecl(t *testing.T) {
	tests := []struct {
		input    string
		contains []string
	}{
		{"enum Color { Red, Green, Blue }", []string{"enum", "Color"}},
		{"enum Option<T> { Some(T), None }", []string{"enum", "Option", "<T>"}},
		{"enum Result<T, E> { Ok(T), Err(E) }", []string{"enum", "Result", "<T, E>"}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		decl := p.parseDeclaration()

		if len(p.Errors()) != 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		result := decl.String()
		for _, str := range tt.contains {
			if !strings.Contains(result, str) {
				t.Errorf("expected %q to contain %q", result, str)
			}
		}
	}
}

func TestParseTraitDecl(t *testing.T) {
	tests := []struct {
		input    string
		contains []string
	}{
		{"trait Display { fn display(&self) void; }", []string{"trait", "Display"}},
		{"pub trait Add<T> { fn add(&self, other T) T; }", []string{"pub", "trait", "Add", "<T>"}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		decl := p.parseDeclaration()

		if len(p.Errors()) != 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		result := decl.String()
		for _, str := range tt.contains {
			if !strings.Contains(result, str) {
				t.Errorf("expected %q to contain %q", result, str)
			}
		}
	}
}

func TestParseImplBlock(t *testing.T) {
	tests := []struct {
		input    string
		contains []string
	}{
		{"impl Point { fn new(x f64, y f64) Point { return Point{x: x, y: y} } }", []string{"impl", "Point"}},
		{"impl Display for Point { fn display(&self) void { println(\"Point\") } }", []string{"impl", "Display", "for", "Point"}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		decl := p.parseDeclaration()

		if len(p.Errors()) != 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		result := decl.String()
		for _, str := range tt.contains {
			if !strings.Contains(result, str) {
				t.Errorf("expected %q to contain %q", result, str)
			}
		}
	}
}

func TestParseSimpleDecls(t *testing.T) {
	tests := []struct {
		input    string
		contains []string
	}{
		{"type MyInt = i32", []string{"type", "MyInt", "=", "i32"}},
		{"const MAX: i32 = 100", []string{"const", "MAX", "i32", "100"}},
		{"use std::io::File", []string{"use", "std", "io", "File"}},
		{"use std::collections::Vec as Vector", []string{"use", "Vec", "as", "Vector"}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		decl := p.parseDeclaration()

		if len(p.Errors()) != 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		result := decl.String()
		for _, str := range tt.contains {
			if !strings.Contains(result, str) {
				t.Errorf("expected %q to contain %q", result, str)
			}
		}
	}
}
