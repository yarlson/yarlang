package lexer

import "testing"

func TestTokenString(t *testing.T) {
	tests := []struct {
		tok      TokenType
		expected string
	}{
		{IDENT, "IDENT"},
		{NUMBER, "NUMBER"},
		{PLUS, "PLUS"},
		{FUNC, "FUNC"},
	}

	for _, tt := range tests {
		if tt.tok.String() != tt.expected {
			t.Errorf("TokenType.String() = %v, want %v", tt.tok.String(), tt.expected)
		}
	}
}
