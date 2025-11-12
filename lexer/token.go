package lexer

// TokenType represents the type of a token
type TokenType int

const (
	// Special tokens
	ILLEGAL TokenType = iota
	EOF
	COMMENT

	// Literals
	IDENT  // x, foo
	INT    // 123, 0xFF, 0b1010, 0o755
	FLOAT  // 1.0, 3.14e-2
	CHAR   // 'a', '\n'
	STRING // "hello"

	// Keywords
	AS       // as
	BREAK    // break
	CONST    // const
	CONTINUE // continue
	DEFER    // defer
	ELSE     // else
	ENUM     // enum
	EXTERN   // extern
	FALSE    // false
	FN       // fn
	FOR      // for
	IF       // if
	IMPL     // impl
	LET      // let
	MODULE   // module
	MUT      // mut
	NIL      // nil
	PUB      // pub
	RETURN   // return
	STRUCT   // struct
	TRAIT    // trait
	TRUE     // true
	TYPE     // type
	UNSAFE   // unsafe
	USE      // use
	VOID     // void
	WHILE    // while

	// Operators
	ASSIGN     // =
	PLUS       // +
	MINUS      // -
	STAR       // *
	SLASH      // /
	PERCENT    // %
	AMP        // &
	PIPE       // |
	CARET      // ^
	TILDE      // ~
	BANG       // !
	LT         // <
	GT         // >
	LTE        // <=
	GTE        // >=
	EQ         // ==
	NEQ        // !=
	AND        // &&
	OR         // ||
	SHL        // <<
	SHR        // >>
	PLUS_EQ    // +=
	MINUS_EQ   // -=
	STAR_EQ    // *=
	SLASH_EQ   // /=
	PERCENT_EQ // %=
	AMP_EQ     // &=
	PIPE_EQ    // |=
	CARET_EQ   // ^=
	SHL_EQ     // <<=
	SHR_EQ     // >>=
	QUESTION   // ?

	// Delimiters
	COLONASSIGN // :=
	LPAREN      // (
	RPAREN      // )
	LBRACE      // {
	RBRACE      // }
	LBRACKET    // [
	RBRACKET    // ]
	COMMA       // ,
	DOT         // .
	DOTDOT      // .. (for ranges)
	SEMICOLON   // ;
	COLON       // :
	COLONCOLON  // ::
	ARROW       // ->
	NEWLINE     // \n (for ASI)
)

var keywords = map[string]TokenType{
	"as":       AS,
	"break":    BREAK,
	"const":    CONST,
	"continue": CONTINUE,
	"defer":    DEFER,
	"else":     ELSE,
	"enum":     ENUM,
	"extern":   EXTERN,
	"false":    FALSE,
	"fn":       FN,
	"for":      FOR,
	"if":       IF,
	"impl":     IMPL,
	"let":      LET,
	"module":   MODULE,
	"mut":      MUT,
	"nil":      NIL,
	"pub":      PUB,
	"return":   RETURN,
	"struct":   STRUCT,
	"trait":    TRAIT,
	"true":     TRUE,
	"type":     TYPE,
	"unsafe":   UNSAFE,
	"use":      USE,
	"void":     VOID,
	"while":    WHILE,
}

// Token represents a lexical token
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// LookupIdent returns the TokenType for an identifier (keyword or IDENT)
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}

	return IDENT
}

// String returns the string representation of a TokenType
func (t TokenType) String() string {
	names := [...]string{
		ILLEGAL:     "ILLEGAL",
		EOF:         "EOF",
		COMMENT:     "COMMENT",
		IDENT:       "IDENT",
		INT:         "INT",
		FLOAT:       "FLOAT",
		CHAR:        "CHAR",
		STRING:      "STRING",
		AS:          "AS",
		BREAK:       "BREAK",
		CONST:       "CONST",
		CONTINUE:    "CONTINUE",
		DEFER:       "DEFER",
		ELSE:        "ELSE",
		ENUM:        "ENUM",
		EXTERN:      "EXTERN",
		FALSE:       "FALSE",
		FN:          "FN",
		FOR:         "FOR",
		IF:          "IF",
		IMPL:        "IMPL",
		LET:         "LET",
		MODULE:      "MODULE",
		MUT:         "MUT",
		NIL:         "NIL",
		PUB:         "PUB",
		RETURN:      "RETURN",
		STRUCT:      "STRUCT",
		TRAIT:       "TRAIT",
		TRUE:        "TRUE",
		TYPE:        "TYPE",
		UNSAFE:      "UNSAFE",
		USE:         "USE",
		VOID:        "VOID",
		WHILE:       "WHILE",
		ASSIGN:      "ASSIGN",
		PLUS:        "PLUS",
		MINUS:       "MINUS",
		STAR:        "STAR",
		SLASH:       "SLASH",
		PERCENT:     "PERCENT",
		AMP:         "AMP",
		PIPE:        "PIPE",
		CARET:       "CARET",
		TILDE:       "TILDE",
		BANG:        "BANG",
		LT:          "LT",
		GT:          "GT",
		LTE:         "LTE",
		GTE:         "GTE",
		EQ:          "EQ",
		NEQ:         "NEQ",
		AND:         "AND",
		OR:          "OR",
		SHL:         "SHL",
		SHR:         "SHR",
		PLUS_EQ:     "PLUS_EQ",
		MINUS_EQ:    "MINUS_EQ",
		STAR_EQ:     "STAR_EQ",
		SLASH_EQ:    "SLASH_EQ",
		PERCENT_EQ:  "PERCENT_EQ",
		AMP_EQ:      "AMP_EQ",
		PIPE_EQ:     "PIPE_EQ",
		CARET_EQ:    "CARET_EQ",
		SHL_EQ:      "SHL_EQ",
		SHR_EQ:      "SHR_EQ",
		QUESTION:    "QUESTION",
		LPAREN:      "LPAREN",
		RPAREN:      "RPAREN",
		LBRACE:      "LBRACE",
		RBRACE:      "RBRACE",
		LBRACKET:    "LBRACKET",
		RBRACKET:    "RBRACKET",
		COMMA:       "COMMA",
		DOT:         "DOT",
		DOTDOT:      "DOTDOT",
		SEMICOLON:   "SEMICOLON",
		COLON:       "COLON",
		COLONCOLON:  "COLONCOLON",
		COLONASSIGN: "COLONASSIGN",
		ARROW:       "ARROW",
		NEWLINE:     "NEWLINE",
	}
	if int(t) < len(names) && names[t] != "" {
		return names[t]
	}

	return "UNKNOWN"
}
