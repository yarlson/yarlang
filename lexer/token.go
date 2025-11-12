package lexer

import (
	"fmt"

	"github.com/yarlson/yarlang/ast"
)

// TokenType represents the type of a token
type TokenType int

const (
	// Special tokens
	ILLEGAL TokenType = iota
	EOF
	COMMENT

	// Literals
	IDENT  // x, foo, bar
	NUMBER // 123, 3.14
	STRING // "hello"

	// Operators
	ASSIGN  // =
	PLUS    // +
	MINUS   // -
	STAR    // *
	SLASH   // /
	PERCENT // %

	EQ  // ==
	NEQ // !=
	LT  // <
	GT  // >
	LTE // <=
	GTE // >=

	AND // &&
	OR  // ||
	NOT // !

	// Delimiters
	LPAREN    // (
	RPAREN    // )
	LBRACE    // {
	RBRACE    // }
	COMMA     // ,
	SEMICOLON // ;
	NEWLINE   // \n

	// Keywords
	FUNC     // func
	RETURN   // return
	IF       // if
	ELSE     // else
	FOR      // for
	BREAK    // break
	CONTINUE // continue
	IMPORT   // import
	AS       // as (for import aliases)
	NIL      // nil
	TRUE     // true
	FALSE    // false
)

var tokenNames = map[TokenType]string{
	ILLEGAL:   "ILLEGAL",
	EOF:       "EOF",
	COMMENT:   "COMMENT",
	IDENT:     "IDENT",
	NUMBER:    "NUMBER",
	STRING:    "STRING",
	ASSIGN:    "ASSIGN",
	PLUS:      "PLUS",
	MINUS:     "MINUS",
	STAR:      "STAR",
	SLASH:     "SLASH",
	PERCENT:   "PERCENT",
	EQ:        "EQ",
	NEQ:       "NEQ",
	LT:        "LT",
	GT:        "GT",
	LTE:       "LTE",
	GTE:       "GTE",
	AND:       "AND",
	OR:        "OR",
	NOT:       "NOT",
	LPAREN:    "LPAREN",
	RPAREN:    "RPAREN",
	LBRACE:    "LBRACE",
	RBRACE:    "RBRACE",
	COMMA:     "COMMA",
	SEMICOLON: "SEMICOLON",
	NEWLINE:   "NEWLINE",
	FUNC:      "FUNC",
	RETURN:    "RETURN",
	IF:        "IF",
	ELSE:      "ELSE",
	FOR:       "FOR",
	BREAK:     "BREAK",
	CONTINUE:  "CONTINUE",
	IMPORT:    "IMPORT",
	AS:        "AS",
	NIL:       "NIL",
	TRUE:      "TRUE",
	FALSE:     "FALSE",
}

func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}

	return fmt.Sprintf("TokenType(%d)", t)
}

// Token represents a lexical token
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

func (t Token) String() string {
	return fmt.Sprintf("%s(%s) at %d:%d", t.Type, t.Literal, t.Line, t.Column)
}

// Position returns the position of this token as an ast.Position
func (t Token) Position() ast.Position {
	return ast.Position{
		Line:   t.Line,
		Column: t.Column,
		Offset: -1,
	}
}

// Keywords maps keyword strings to their token types
var Keywords = map[string]TokenType{
	"func":     FUNC,
	"return":   RETURN,
	"if":       IF,
	"else":     ELSE,
	"for":      FOR,
	"break":    BREAK,
	"continue": CONTINUE,
	"import":   IMPORT,
	"as":       AS,
	"nil":      NIL,
	"true":     TRUE,
	"false":    FALSE,
}

// LookupIdent checks if an identifier is a keyword
func LookupIdent(ident string) TokenType {
	if tok, ok := Keywords[ident]; ok {
		return tok
	}

	return IDENT
}
