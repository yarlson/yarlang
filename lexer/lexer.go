package lexer

// Lexer performs lexical analysis
type Lexer struct {
	input        string
	position     int  // current position
	readPosition int  // next position
	ch           byte // current char
	line         int
	column       int
}

// New creates a new Lexer
func New(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar()

	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // EOF
	} else {
		l.ch = l.input[l.readPosition]
	}

	l.position = l.readPosition
	l.readPosition++
	l.column++

	if l.ch == '\n' {
		l.line++
		l.column = 0
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}

	return l.input[l.readPosition]
}

// NextToken returns the next token
func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	tok.Line = l.line
	tok.Column = l.column

	switch l.ch {
	case 0:
		tok.Type = EOF
	case '\n':
		tok.Type = NEWLINE
		tok.Literal = "\n"
	case '(':
		tok.Type = LPAREN
		tok.Literal = "("
	case ')':
		tok.Type = RPAREN
		tok.Literal = ")"
	case '{':
		tok.Type = LBRACE
		tok.Literal = "{"
	case '}':
		tok.Type = RBRACE
		tok.Literal = "}"
	case '[':
		tok.Type = LBRACKET
		tok.Literal = "["
	case ']':
		tok.Type = RBRACKET
		tok.Literal = "]"
	case ',':
		tok.Type = COMMA
		tok.Literal = ","
	case '.':
		// Check if this is '..' (range operator)
		if l.peekChar() == '.' {
			ch := l.ch
			l.readChar()

			tok.Type = DOTDOT
			tok.Literal = string(ch) + string(l.ch)
		} else if isDigit(l.peekChar()) {
			// Check if this is a float starting with '.' (e.g., .5)
			return l.readNumber()
		} else {
			tok.Type = DOT
			tok.Literal = "."
		}
	case ';':
		tok.Type = SEMICOLON
		tok.Literal = ";"
	case ':':
		if l.peekChar() == ':' {
			ch := l.ch
			l.readChar()

			tok.Type = COLONCOLON
			tok.Literal = string(ch) + string(l.ch)
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()

			tok.Type = COLONASSIGN
			tok.Literal = string(ch) + string(l.ch)
		} else {
			tok.Type = COLON
			tok.Literal = ":"
		}
	case '?':
		tok.Type = QUESTION
		tok.Literal = "?"
	case '+':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()

			tok.Type = PLUS_EQ
			tok.Literal = string(ch) + string(l.ch)
		} else {
			tok.Type = PLUS
			tok.Literal = "+"
		}
	case '-':
		if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()

			tok.Type = ARROW
			tok.Literal = string(ch) + string(l.ch)
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()

			tok.Type = MINUS_EQ
			tok.Literal = string(ch) + string(l.ch)
		} else {
			tok.Type = MINUS
			tok.Literal = "-"
		}
	case '*':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()

			tok.Type = STAR_EQ
			tok.Literal = string(ch) + string(l.ch)
		} else {
			tok.Type = STAR
			tok.Literal = "*"
		}
	case '/':
		if l.peekChar() == '/' {
			tok.Type = COMMENT
			tok.Literal = l.readLineComment()
		} else if l.peekChar() == '*' {
			tok.Type = COMMENT
			tok.Literal = l.readBlockComment()
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()

			tok.Type = SLASH_EQ
			tok.Literal = string(ch) + string(l.ch)
		} else {
			tok.Type = SLASH
			tok.Literal = "/"
		}
	case '%':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()

			tok.Type = PERCENT_EQ
			tok.Literal = string(ch) + string(l.ch)
		} else {
			tok.Type = PERCENT
			tok.Literal = "%"
		}
	case '&':
		if l.peekChar() == '&' {
			ch := l.ch
			l.readChar()

			tok.Type = AND
			tok.Literal = string(ch) + string(l.ch)
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()

			tok.Type = AMP_EQ
			tok.Literal = string(ch) + string(l.ch)
		} else {
			tok.Type = AMP
			tok.Literal = "&"
		}
	case '|':
		if l.peekChar() == '|' {
			ch := l.ch
			l.readChar()

			tok.Type = OR
			tok.Literal = string(ch) + string(l.ch)
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()

			tok.Type = PIPE_EQ
			tok.Literal = string(ch) + string(l.ch)
		} else {
			tok.Type = PIPE
			tok.Literal = "|"
		}
	case '^':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()

			tok.Type = CARET_EQ
			tok.Literal = string(ch) + string(l.ch)
		} else {
			tok.Type = CARET
			tok.Literal = "^"
		}
	case '~':
		tok.Type = TILDE
		tok.Literal = "~"
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()

			tok.Type = NEQ
			tok.Literal = string(ch) + string(l.ch)
		} else {
			tok.Type = BANG
			tok.Literal = "!"
		}
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()

			tok.Type = LTE
			tok.Literal = string(ch) + string(l.ch)
		} else if l.peekChar() == '<' {
			ch := l.ch
			l.readChar()

			if l.peekChar() == '=' {
				lit := string(ch) + string(l.ch)
				l.readChar()

				tok.Type = SHL_EQ
				tok.Literal = lit + string(l.ch)
			} else {
				tok.Type = SHL
				tok.Literal = string(ch) + string(l.ch)
			}
		} else {
			tok.Type = LT
			tok.Literal = "<"
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()

			tok.Type = GTE
			tok.Literal = string(ch) + string(l.ch)
		} else if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()

			if l.peekChar() == '=' {
				lit := string(ch) + string(l.ch)
				l.readChar()

				tok.Type = SHR_EQ
				tok.Literal = lit + string(l.ch)
			} else {
				tok.Type = SHR
				tok.Literal = string(ch) + string(l.ch)
			}
		} else {
			tok.Type = GT
			tok.Literal = ">"
		}
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()

			tok.Type = EQ
			tok.Literal = string(ch) + string(l.ch)
		} else {
			tok.Type = ASSIGN
			tok.Literal = "="
		}
	case '"':
		tok.Type = STRING
		tok.Literal = l.readString()
	case '\'':
		tok.Type = CHAR
		tok.Literal = l.readChar2()
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = LookupIdent(tok.Literal)

			return tok
		} else if isDigit(l.ch) {
			return l.readNumber()
		} else {
			tok.Type = ILLEGAL
			tok.Literal = string(l.ch)
		}
	}

	l.readChar()

	return tok
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}

	return l.input[position:l.position]
}

func (l *Lexer) readNumber() Token {
	tok := Token{Line: l.line, Column: l.column}
	position := l.position

	// Check for float starting with '.' (e.g., .5)
	if l.ch == '.' {
		tok.Type = FLOAT

		l.readChar()

		for isDigit(l.ch) || l.ch == '_' {
			l.readChar()
		}
		// Check for exponent
		if l.ch == 'e' || l.ch == 'E' {
			l.readChar()

			if l.ch == '+' || l.ch == '-' {
				l.readChar()
			}

			for isDigit(l.ch) || l.ch == '_' {
				l.readChar()
			}
		}

		tok.Literal = l.input[position:l.position]

		return tok
	}

	// Check for hex, binary, octal
	if l.ch == '0' && (l.peekChar() == 'x' || l.peekChar() == 'X') {
		l.readChar()
		l.readChar()

		for isHexDigit(l.ch) || l.ch == '_' {
			l.readChar()
		}

		tok.Type = INT
		tok.Literal = l.input[position:l.position]

		return tok
	}

	if l.ch == '0' && (l.peekChar() == 'b' || l.peekChar() == 'B') {
		l.readChar()
		l.readChar()

		for l.ch == '0' || l.ch == '1' || l.ch == '_' {
			l.readChar()
		}

		tok.Type = INT
		tok.Literal = l.input[position:l.position]

		return tok
	}

	if l.ch == '0' && (l.peekChar() == 'o' || l.peekChar() == 'O') {
		l.readChar()
		l.readChar()

		for isOctalDigit(l.ch) || l.ch == '_' {
			l.readChar()
		}

		tok.Type = INT
		tok.Literal = l.input[position:l.position]

		return tok
	}

	// Read integer part
	for isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}

	// Check for float with decimal point (e.g., 5. or 5.0)
	if l.ch == '.' {
		// Check if next char is a digit or if we're at the end/non-digit (for 5. format)
		nextCh := l.peekChar()
		if isDigit(nextCh) {
			// Standard float like 5.0
			l.readChar()

			for isDigit(l.ch) || l.ch == '_' {
				l.readChar()
			}

			tok.Type = FLOAT
		} else if !isLetter(nextCh) && nextCh != '.' {
			// Float ending with dot like 5.
			l.readChar()

			tok.Type = FLOAT
		} else {
			// Not a float, just an integer followed by something else
			tok.Type = INT
		}
	} else {
		tok.Type = INT
	}

	// Check for exponent
	if l.ch == 'e' || l.ch == 'E' {
		tok.Type = FLOAT

		l.readChar()

		if l.ch == '+' || l.ch == '-' {
			l.readChar()
		}

		for isDigit(l.ch) || l.ch == '_' {
			l.readChar()
		}
	}

	tok.Literal = l.input[position:l.position]

	return tok
}

func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()

		if l.ch == '"' || l.ch == 0 {
			break
		}

		if l.ch == '\\' {
			l.readChar() // skip escaped char
		}
	}

	return l.input[position:l.position]
}

func (l *Lexer) readChar2() string {
	position := l.position + 1
	for {
		l.readChar()

		if l.ch == '\'' || l.ch == 0 {
			break
		}

		if l.ch == '\\' {
			l.readChar() // skip escaped char
		}
	}

	return l.input[position:l.position]
}

func (l *Lexer) readLineComment() string {
	position := l.position
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}

	return l.input[position:l.position]
}

func (l *Lexer) readBlockComment() string {
	position := l.position
	l.readChar()
	l.readChar()

	for l.ch != 0 {
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar()
			l.readChar()

			break
		}

		l.readChar()
	}

	return l.input[position:l.position]
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isHexDigit(ch byte) bool {
	return '0' <= ch && ch <= '9' || 'a' <= ch && ch <= 'f' || 'A' <= ch && ch <= 'F'
}

func isOctalDigit(ch byte) bool {
	return '0' <= ch && ch <= '7'
}
