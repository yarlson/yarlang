package lexer

import "testing"

func TestTokenTypes(t *testing.T) {
	tests := []struct {
		input    string
		expected []TokenType
	}{
		{"fn", []TokenType{FN}},
		{"let", []TokenType{LET}},
		{"mut", []TokenType{MUT}},
		{"struct", []TokenType{STRUCT}},
		{"enum", []TokenType{ENUM}},
		{"trait", []TokenType{TRAIT}},
		{"impl", []TokenType{IMPL}},
	}

	for _, tt := range tests {
		l := New(tt.input)
		for i, expected := range tt.expected {
			tok := l.NextToken()
			if tok.Type != expected {
				t.Errorf("test[%d] - wrong token type. expected=%v, got=%v",
					i, expected, tok.Type)
			}
		}
	}
}

func TestV04Example(t *testing.T) {
	input := `fn add(a i32, b i32) i32 {
    return a + b
}

struct Point {
    x: f64,
    y: f64,
}

let mut x: i32 = 42
x += 10

let p := &Point{ x: 1.0, y: 2.0 }
`

	l := New(input)

	// Just verify it tokenizes without panics
	for {
		tok := l.NextToken()
		if tok.Type == EOF {
			break
		}

		if tok.Type == ILLEGAL {
			t.Fatalf("Illegal token: %s at line %d", tok.Literal, tok.Line)
		}
	}
}

func TestFloatParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
		literal  string
	}{
		{"1.0", FLOAT, "1.0"},
		{".5", FLOAT, ".5"},
		{"5.", FLOAT, "5."},
		{"1e9", FLOAT, "1e9"},
		{"3.14e-2", FLOAT, "3.14e-2"},
		{"0.5", FLOAT, "0.5"},
		{".123", FLOAT, ".123"},
		{"99.", FLOAT, "99."},
		{"2.5e+10", FLOAT, "2.5e+10"},
	}

	for _, tt := range tests {
		l := New(tt.input)

		tok := l.NextToken()
		if tok.Type != tt.expected {
			t.Errorf("input %q - wrong token type. expected=%v, got=%v",
				tt.input, tt.expected, tok.Type)
		}

		if tok.Literal != tt.literal {
			t.Errorf("input %q - wrong literal. expected=%q, got=%q",
				tt.input, tt.literal, tok.Literal)
		}
	}
}
