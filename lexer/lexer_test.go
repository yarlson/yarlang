package lexer

import "testing"

func TestLexerBasic(t *testing.T) {
	input := `x = 42`

	l := New(input)

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{IDENT, "x"},
		{ASSIGN, "="},
		{NUMBER, "42"},
		{EOF, ""},
	}

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestLexerComprehensive(t *testing.T) {
	input := `
func add(a, b) {
	return a + b
}

x = 42
name = "Alice"

if x > 0 {
	println("positive")
} else {
	println("zero or negative")
}

for i = 0; i < 10; i = i + 1 {
	print(i)
}

// This is a comment
result = 3.14 * 2
check = true && false || !true
`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{FUNC, "func"},
		{IDENT, "add"},
		{LPAREN, "("},
		{IDENT, "a"},
		{COMMA, ","},
		{IDENT, "b"},
		{RPAREN, ")"},
		{LBRACE, "{"},
		{RETURN, "return"},
		{IDENT, "a"},
		{PLUS, "+"},
		{IDENT, "b"},
		{RBRACE, "}"},
		{IDENT, "x"},
		{ASSIGN, "="},
		{NUMBER, "42"},
		{IDENT, "name"},
		{ASSIGN, "="},
		{STRING, "Alice"},
		{IF, "if"},
		{IDENT, "x"},
		{GT, ">"},
		{NUMBER, "0"},
		{LBRACE, "{"},
		{IDENT, "println"},
		{LPAREN, "("},
		{STRING, "positive"},
		{RPAREN, ")"},
		{RBRACE, "}"},
		{ELSE, "else"},
		{LBRACE, "{"},
		{IDENT, "println"},
		{LPAREN, "("},
		{STRING, "zero or negative"},
		{RPAREN, ")"},
		{RBRACE, "}"},
		{FOR, "for"},
		{IDENT, "i"},
		{ASSIGN, "="},
		{NUMBER, "0"},
		{SEMICOLON, ";"},
		{IDENT, "i"},
		{LT, "<"},
		{NUMBER, "10"},
		{SEMICOLON, ";"},
		{IDENT, "i"},
		{ASSIGN, "="},
		{IDENT, "i"},
		{PLUS, "+"},
		{NUMBER, "1"},
		{LBRACE, "{"},
		{IDENT, "print"},
		{LPAREN, "("},
		{IDENT, "i"},
		{RPAREN, ")"},
		{RBRACE, "}"},
		{IDENT, "result"},
		{ASSIGN, "="},
		{NUMBER, "3.14"},
		{STAR, "*"},
		{NUMBER, "2"},
		{IDENT, "check"},
		{ASSIGN, "="},
		{TRUE, "true"},
		{AND, "&&"},
		{FALSE, "false"},
		{OR, "||"},
		{NOT, "!"},
		{TRUE, "true"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q (literal=%q)",
				i, tt.expectedType, tok.Type, tok.Literal)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}
