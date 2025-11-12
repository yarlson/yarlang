package parser

import (
	"fmt"

	"github.com/yarlson/yarlang/ast"
	"github.com/yarlson/yarlang/lexer"
)

// Parser parses tokens into AST
type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  lexer.Token
	peekToken lexer.Token
}

// New creates a new Parser
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	// Read two tokens to initialize curToken and peekToken
	p.nextToken()
	p.nextToken()

	return p
}

// Errors returns parser errors
func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) error(msg string) {
	p.errors = append(p.errors, fmt.Sprintf("line %d: %s", p.curToken.Line, msg))
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()

	// Skip comments and newlines (handle ASI later)
	for p.peekToken.Type == lexer.COMMENT {
		p.peekToken = p.l.NextToken()
	}
}

func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}

	p.error(fmt.Sprintf("expected next token to be %v, got %v instead", t, p.peekToken.Type))

	return false
}

// ===== Type Parsing =====

func (p *Parser) parseType() ast.Type {
	switch p.curToken.Type {
	case lexer.AMP:
		return p.parseRefType()
	case lexer.STAR:
		return p.parsePtrType()
	case lexer.LBRACKET:
		return p.parseArrayOrSliceType()
	case lexer.LPAREN:
		return p.parseTupleType()
	case lexer.IDENT:
		return p.parseTypePath()
	case lexer.VOID:
		return &ast.VoidType{}
	default:
		p.error(fmt.Sprintf("unexpected token in type: %v", p.curToken.Type))
		return nil
	}
}

func (p *Parser) parseRefType() ast.Type {
	// & or &mut
	p.nextToken() // consume &

	mut := false
	if p.curTokenIs(lexer.MUT) {
		mut = true

		p.nextToken()
	}

	elem := p.parseType()

	return &ast.RefType{Mut: mut, Elem: elem}
}

func (p *Parser) parsePtrType() ast.Type {
	p.nextToken() // consume *
	elem := p.parseType()

	return &ast.PtrType{Elem: elem}
}

func (p *Parser) parseArrayOrSliceType() ast.Type {
	p.nextToken() // consume [

	if p.curTokenIs(lexer.RBRACKET) {
		// []T - slice
		p.nextToken() // consume ]
		elem := p.parseType()

		return &ast.SliceType{Elem: elem}
	}

	// [T; N] - array
	elem := p.parseType()
	if !p.expectPeek(lexer.SEMICOLON) {
		return nil
	}

	p.nextToken()

	len := p.parseExpression(LOWEST)
	if !p.expectPeek(lexer.RBRACKET) {
		return nil
	}
	// curToken is now RBRACKET - leave it there (last token of type)
	return &ast.ArrayType{Elem: elem, Len: len}
}

func (p *Parser) parseTupleType() ast.Type {
	p.nextToken() // consume (

	if p.curTokenIs(lexer.RPAREN) {
		// () - unit
		return &ast.TupleType{Elems: []ast.Type{}}
	}

	elems := []ast.Type{}
	elems = append(elems, p.parseType())

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // consume previous
		p.nextToken() // consume ,

		if p.curTokenIs(lexer.RPAREN) {
			break
		}

		elems = append(elems, p.parseType())
	}

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}
	// curToken is now RPAREN - leave it there (last token of type)

	return &ast.TupleType{Elems: elems}
}

func (p *Parser) parseTypePath() ast.Type {
	path := []string{p.curToken.Literal}

	for p.peekTokenIs(lexer.COLONCOLON) {
		p.nextToken() // consume previous
		p.nextToken() // consume ::

		if !p.curTokenIs(lexer.IDENT) {
			p.error("expected identifier after ::")
			return nil
		}

		path = append(path, p.curToken.Literal)
	}

	// Check for generic args
	var args []ast.Type

	if p.peekTokenIs(lexer.LT) {
		p.nextToken() // consume type name
		p.nextToken() // consume <

		args = append(args, p.parseType())
		for p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // consume previous
			p.nextToken() // consume ,
			args = append(args, p.parseType())
		}

		if !p.expectPeek(lexer.GT) {
			return nil
		}
		// curToken is now GT - leave it there (last token of type)
	}

	return &ast.TypePath{Path: path, Args: args}
}

// ===== Expression Parsing =====

const (
	_ int = iota
	LOWEST
	ASSIGN      // = += etc
	OR          // ||
	AND         // &&
	BIT_OR      // |
	BIT_XOR     // ^
	BIT_AND     // &
	EQUALS      // == !=
	LESSGREATER // > < <= >=
	SHIFT       // << >>
	SUM         // + -
	PRODUCT     // * / %
	PREFIX      // -X !X &X *X
	POSTFIX     // X() X[] X. X?
)

var precedences = map[lexer.TokenType]int{
	// NOTE: Assignment operators are NOT in precedence table
	// They are handled at statement level in parseAssignOrExprStmt
	// lexer.ASSIGN:     ASSIGN,
	// lexer.PLUS_EQ:    ASSIGN,
	// lexer.MINUS_EQ:   ASSIGN,
	// lexer.STAR_EQ:    ASSIGN,
	// lexer.SLASH_EQ:   ASSIGN,
	// lexer.PERCENT_EQ: ASSIGN,
	// lexer.AMP_EQ:     ASSIGN,
	// lexer.PIPE_EQ:    ASSIGN,
	// lexer.CARET_EQ:   ASSIGN,
	// lexer.SHL_EQ:     ASSIGN,
	// lexer.SHR_EQ:     ASSIGN,
	lexer.OR:       OR,
	lexer.AND:      AND,
	lexer.PIPE:     BIT_OR,
	lexer.CARET:    BIT_XOR,
	lexer.AMP:      BIT_AND,
	lexer.EQ:       EQUALS,
	lexer.NEQ:      EQUALS,
	lexer.LT:       LESSGREATER,
	lexer.GT:       LESSGREATER,
	lexer.LTE:      LESSGREATER,
	lexer.GTE:      LESSGREATER,
	lexer.SHL:      SHIFT,
	lexer.SHR:      SHIFT,
	lexer.PLUS:     SUM,
	lexer.MINUS:    SUM,
	lexer.DOTDOT:   SUM, // Same precedence as + -
	lexer.STAR:     PRODUCT,
	lexer.SLASH:    PRODUCT,
	lexer.PERCENT:  PRODUCT,
	lexer.LPAREN:   POSTFIX,
	lexer.LBRACKET: POSTFIX,
	lexer.DOT:      POSTFIX,
	lexer.QUESTION: POSTFIX,
}

func (p *Parser) peekPrecedence() int {
	if prec, ok := precedences[p.peekToken.Type]; ok {
		return prec
	}

	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if prec, ok := precedences[p.curToken.Type]; ok {
		return prec
	}

	return LOWEST
}

func (p *Parser) parseExpression(precedence int) ast.Expr {
	// Parse prefix expression
	prefix := p.parsePrefixExpression()
	if prefix == nil {
		return nil
	}

	// Parse infix expressions with precedence
	for !p.peekTokenIs(lexer.SEMICOLON) && !p.peekTokenIs(lexer.NEWLINE) && precedence < p.peekPrecedence() {
		infix := p.parseInfixExpression(prefix)
		if infix == nil {
			return prefix
		}

		prefix = infix
	}

	return prefix
}

func (p *Parser) parsePrefixExpression() ast.Expr {
	switch p.curToken.Type {
	case lexer.IDENT:
		// Check if it's a struct literal
		if p.peekTokenIs(lexer.LBRACE) {
			return p.parseStructLiteral()
		}

		return &ast.Ident{Name: p.curToken.Literal}
	case lexer.LBRACKET:
		return p.parseArrayLiteral()
	case lexer.INT:
		return &ast.IntLit{Value: p.curToken.Literal}
	case lexer.FLOAT:
		return &ast.FloatLit{Value: p.curToken.Literal}
	case lexer.STRING:
		return &ast.StringLit{Value: p.curToken.Literal}
	case lexer.CHAR:
		return &ast.CharLit{Value: p.curToken.Literal}
	case lexer.TRUE:
		return &ast.BoolLit{Value: true}
	case lexer.FALSE:
		return &ast.BoolLit{Value: false}
	case lexer.NIL:
		return &ast.NilLit{}
	case lexer.LPAREN:
		return p.parseGroupedExpression()
	case lexer.AMP:
		// Check for &mut
		if p.peekTokenIs(lexer.MUT) {
			p.nextToken() // consume &
			p.nextToken() // consume mut
			expr := p.parseExpression(PREFIX)

			return &ast.UnaryExpr{Op: "&mut", Expr: expr}
		}

		return p.parseUnaryExpression()
	case lexer.MINUS, lexer.BANG, lexer.TILDE, lexer.STAR:
		return p.parseUnaryExpression()
	default:
		p.error(fmt.Sprintf("no prefix parse function for %v", p.curToken.Type))
		return nil
	}
}

func (p *Parser) parseInfixExpression(left ast.Expr) ast.Expr {
	switch p.peekToken.Type {
	case lexer.LPAREN:
		return p.parseCallExpression(left)
	case lexer.LBRACKET:
		return p.parseIndexExpression(left)
	case lexer.DOT:
		return p.parseFieldExpression(left)
	case lexer.QUESTION:
		return p.parsePropagateExpression(left)
	default:
		// Binary operator
		p.nextToken() // move to operator
		op := p.curToken.Literal
		precedence := p.curPrecedence()
		p.nextToken() // move to right operand

		// Right-associative for assignment operators
		var right ast.Expr
		if precedence == ASSIGN {
			right = p.parseExpression(precedence - 1)
		} else {
			right = p.parseExpression(precedence)
		}

		return &ast.BinaryExpr{Left: left, Op: op, Right: right}
	}
}

func (p *Parser) parseGroupedExpression() ast.Expr {
	p.nextToken() // consume (

	expr := p.parseExpression(LOWEST)

	// Check for tuple
	if p.peekTokenIs(lexer.COMMA) {
		elems := []ast.Expr{expr}

		for p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // consume current
			p.nextToken() // consume comma

			if p.curTokenIs(lexer.RPAREN) {
				break // trailing comma
			}

			elems = append(elems, p.parseExpression(LOWEST))
		}

		if !p.expectPeek(lexer.RPAREN) {
			return nil
		}

		return &ast.TupleExpr{Elems: elems}
	}

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	return expr
}

func (p *Parser) parseUnaryExpression() ast.Expr {
	op := p.curToken.Literal
	p.nextToken()
	expr := p.parseExpression(PREFIX)

	return &ast.UnaryExpr{Op: op, Expr: expr}
}

func (p *Parser) parseCallExpression(callee ast.Expr) ast.Expr {
	p.nextToken() // consume (

	args := []ast.Expr{}

	if !p.peekTokenIs(lexer.RPAREN) {
		p.nextToken() // move to first argument

		args = append(args, p.parseExpression(LOWEST))
		for p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // consume current
			p.nextToken() // consume comma
			args = append(args, p.parseExpression(LOWEST))
		}

		if !p.expectPeek(lexer.RPAREN) {
			return nil
		}
	} else {
		p.nextToken() // consume )
	}

	return &ast.CallExpr{Callee: callee, Args: args}
}

func (p *Parser) parseIndexExpression(expr ast.Expr) ast.Expr {
	p.nextToken() // consume [
	p.nextToken() // move to index expression

	index := p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.RBRACKET) {
		return nil
	}

	return &ast.IndexExpr{Expr: expr, Index: index}
}

func (p *Parser) parseFieldExpression(expr ast.Expr) ast.Expr {
	p.nextToken() // consume .

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	field := p.curToken.Literal

	return &ast.FieldExpr{Expr: expr, Field: field}
}

func (p *Parser) parsePropagateExpression(expr ast.Expr) ast.Expr {
	p.nextToken() // consume ?

	return &ast.PropagateExpr{Expr: expr}
}

func (p *Parser) parseArrayLiteral() ast.Expr {
	p.nextToken() // consume [

	elems := []ast.Expr{}
	if !p.curTokenIs(lexer.RBRACKET) {
		elems = append(elems, p.parseExpression(LOWEST))
		for p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // move to comma
			p.nextToken() // move past comma

			if p.curTokenIs(lexer.RBRACKET) {
				break // trailing comma
			}

			elems = append(elems, p.parseExpression(LOWEST))
		}
		// After last element, need to move to ]
		if !p.expectPeek(lexer.RBRACKET) {
			return nil
		}
	}
	// Empty array case: curToken is already ]

	return &ast.ArrayExpr{Elems: elems}
}

func (p *Parser) parseStructLiteral() ast.Expr {
	// Parse type path
	typePath := p.parseTypePath()

	p.nextToken() // consume type name
	p.nextToken() // consume {

	inits := []ast.FieldInit{}

	if !p.curTokenIs(lexer.RBRACE) {
		// Parse field: value
		if !p.curTokenIs(lexer.IDENT) {
			p.error("expected field name")
			return nil
		}

		fieldName := p.curToken.Literal

		if !p.expectPeek(lexer.COLON) {
			return nil
		}

		p.nextToken() // consume :

		value := p.parseExpression(LOWEST)
		inits = append(inits, ast.FieldInit{Name: fieldName, Val: value})

		for p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // consume current
			p.nextToken() // consume comma

			if p.curTokenIs(lexer.RBRACE) {
				break // trailing comma
			}

			fieldName := p.curToken.Literal
			if !p.expectPeek(lexer.COLON) {
				return nil
			}

			p.nextToken() // consume :
			value := p.parseExpression(LOWEST)
			inits = append(inits, ast.FieldInit{Name: fieldName, Val: value})
		}
	}

	if !p.expectPeek(lexer.RBRACE) {
		return nil
	}

	return &ast.StructExpr{Type: typePath, Inits: inits}
}

// ===== Statement Parsing =====

// parseStatement parses a statement
func (p *Parser) parseStatement() ast.Stmt {
	switch p.curToken.Type {
	case lexer.LET:
		return p.parseLetStmt()
	case lexer.RETURN:
		return p.parseReturnStmt()
	case lexer.IF:
		return p.parseIfStmt()
	case lexer.WHILE:
		return p.parseWhileStmt()
	case lexer.FOR:
		return p.parseForStmt()
	case lexer.BREAK:
		return p.parseBreakStmt()
	case lexer.CONTINUE:
		return p.parseContinueStmt()
	case lexer.DEFER:
		return p.parseDeferStmt()
	case lexer.UNSAFE:
		return p.parseUnsafeBlock()
	case lexer.LBRACE:
		return p.parseBlock()
	default:
		// Try assignment or expression statement
		return p.parseAssignOrExprStmt()
	}
}

func (p *Parser) parseLetStmt() *ast.LetStmt {
	stmt := &ast.LetStmt{}

	p.nextToken() // consume let

	// Check for mut
	if p.curTokenIs(lexer.MUT) {
		stmt.Mut = true

		p.nextToken()
	}

	// Parse name
	if !p.curTokenIs(lexer.IDENT) {
		p.error("expected identifier after let")
		return nil
	}

	stmt.Name = p.curToken.Literal

	// Check for type annotation
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // consume name
		p.nextToken() // consume :
		stmt.Type = p.parseType()
	}

	// Expect =
	if !p.expectPeek(lexer.ASSIGN) {
		return nil
	}

	p.nextToken() // consume =

	// Parse value
	stmt.Value = p.parseExpression(LOWEST)

	// Skip optional semicolon or newline
	if p.peekTokenIs(lexer.SEMICOLON) || p.peekTokenIs(lexer.NEWLINE) {
		p.nextToken()
	}

	return stmt
}

// Placeholder stubs for other statement types
func (p *Parser) parseReturnStmt() *ast.ReturnStmt {
	stmt := &ast.ReturnStmt{}

	p.nextToken() // consume return

	// Check for value
	if !p.curTokenIs(lexer.SEMICOLON) && !p.curTokenIs(lexer.NEWLINE) &&
		!p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		stmt.Value = p.parseExpression(LOWEST)
	}

	// Skip optional semicolon or newline
	if p.peekTokenIs(lexer.SEMICOLON) || p.peekTokenIs(lexer.NEWLINE) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseIfStmt() *ast.IfStmt {
	stmt := &ast.IfStmt{}

	p.nextToken() // consume if

	// Parse condition
	stmt.Cond = p.parseExpression(LOWEST)

	// Parse then block
	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	stmt.Then = p.parseBlock()

	// Check for else
	if p.peekTokenIs(lexer.ELSE) {
		p.nextToken() // consume }
		p.nextToken() // consume else

		if p.curTokenIs(lexer.IF) {
			// else if
			stmt.Else = p.parseIfStmt()
		} else if p.curTokenIs(lexer.LBRACE) {
			// else block
			stmt.Else = p.parseBlock()
		} else {
			p.error("expected if or { after else")
			return nil
		}
	}

	return stmt
}

func (p *Parser) parseWhileStmt() *ast.WhileStmt {
	stmt := &ast.WhileStmt{}

	p.nextToken() // consume while

	// Parse condition
	stmt.Cond = p.parseExpression(LOWEST)

	// Parse body
	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlock()

	return stmt
}

func (p *Parser) parseForStmt() *ast.ForStmt {
	stmt := &ast.ForStmt{}

	p.nextToken() // consume for

	// Parse key (optional) and val
	if !p.curTokenIs(lexer.IDENT) {
		p.error("expected identifier after for")
		return nil
	}

	firstName := p.curToken.Literal

	// Check for comma (key, val)
	if p.peekTokenIs(lexer.COMMA) {
		stmt.Key = firstName

		p.nextToken() // consume first name
		p.nextToken() // consume comma

		if !p.curTokenIs(lexer.IDENT) {
			p.error("expected identifier after comma")
			return nil
		}

		stmt.Val = p.curToken.Literal
	} else {
		stmt.Val = firstName
	}

	// Expect in
	if !p.expectPeek(lexer.IDENT) || p.curToken.Literal != "in" {
		p.error("expected 'in' after for variable")
		return nil
	}

	p.nextToken() // consume in

	// Parse iterator expression
	stmt.Iter = p.parseExpression(LOWEST)

	// Parse body
	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlock()

	return stmt
}

func (p *Parser) parseBreakStmt() *ast.BreakStmt {
	p.nextToken() // consume break

	// Skip optional semicolon or newline
	if p.peekTokenIs(lexer.SEMICOLON) || p.peekTokenIs(lexer.NEWLINE) {
		p.nextToken()
	}

	return &ast.BreakStmt{}
}

func (p *Parser) parseContinueStmt() *ast.ContinueStmt {
	p.nextToken() // consume continue

	// Skip optional semicolon or newline
	if p.peekTokenIs(lexer.SEMICOLON) || p.peekTokenIs(lexer.NEWLINE) {
		p.nextToken()
	}

	return &ast.ContinueStmt{}
}

func (p *Parser) parseDeferStmt() *ast.DeferStmt {
	stmt := &ast.DeferStmt{}

	p.nextToken() // consume defer

	// Parse expression (should be a call)
	stmt.Expr = p.parseExpression(LOWEST)

	// Skip optional semicolon or newline
	if p.peekTokenIs(lexer.SEMICOLON) || p.peekTokenIs(lexer.NEWLINE) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseUnsafeBlock() *ast.UnsafeBlock {
	stmt := &ast.UnsafeBlock{}

	p.nextToken() // consume unsafe

	// Expect block
	if !p.curTokenIs(lexer.LBRACE) {
		p.error("expected { after unsafe")
		return nil
	}

	stmt.Body = p.parseBlock()

	return stmt
}

func (p *Parser) parseBlock() *ast.Block {
	block := &ast.Block{Stmts: []ast.Stmt{}}

	p.nextToken() // consume {

	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		// Skip newlines and semicolons
		if p.curTokenIs(lexer.NEWLINE) || p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
			continue
		}

		stmt := p.parseStatement()
		block.Stmts = append(block.Stmts, stmt)

		p.nextToken()
	}

	return block
}

func (p *Parser) parseAssignOrExprStmt() ast.Stmt {
	// Check if it's a short declaration (identifier followed by :=)
	if p.curTokenIs(lexer.IDENT) && p.peekTokenIs(lexer.COLONASSIGN) {
		name := p.curToken.Literal
		p.nextToken() // consume name
		p.nextToken() // consume :=

		value := p.parseExpression(LOWEST)

		// Skip optional semicolon or newline
		if p.peekTokenIs(lexer.SEMICOLON) || p.peekTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		return &ast.ShortDecl{Name: name, Value: value}
	}

	// Parse left side as expression
	expr := p.parseExpression(LOWEST)

	// Check if next token is assignment operator
	if p.peekTokenIs(lexer.ASSIGN) || p.peekTokenIs(lexer.PLUS_EQ) ||
		p.peekTokenIs(lexer.MINUS_EQ) || p.peekTokenIs(lexer.STAR_EQ) ||
		p.peekTokenIs(lexer.SLASH_EQ) || p.peekTokenIs(lexer.PERCENT_EQ) ||
		p.peekTokenIs(lexer.AMP_EQ) || p.peekTokenIs(lexer.PIPE_EQ) ||
		p.peekTokenIs(lexer.CARET_EQ) || p.peekTokenIs(lexer.SHL_EQ) ||
		p.peekTokenIs(lexer.SHR_EQ) {
		// Assignment statement
		p.nextToken() // move to operator
		op := p.curToken.Literal
		p.nextToken() // move to value

		value := p.parseExpression(LOWEST)

		// Skip optional semicolon or newline
		if p.peekTokenIs(lexer.SEMICOLON) || p.peekTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		return &ast.AssignStmt{Target: expr, Op: op, Value: value}
	}

	// Expression statement
	// Skip optional semicolon or newline
	if p.peekTokenIs(lexer.SEMICOLON) || p.peekTokenIs(lexer.NEWLINE) {
		p.nextToken()
	}

	return &ast.ExprStmt{Expr: expr}
}

// ===== Declaration Parsing =====

// parseDeclaration parses a top-level declaration
func (p *Parser) parseDeclaration() ast.Decl {
	// Check for pub
	pub := false
	if p.curTokenIs(lexer.PUB) {
		pub = true

		p.nextToken()
	}

	switch p.curToken.Type {
	case lexer.FN:
		return p.parseFuncDecl(pub)
	case lexer.STRUCT:
		return p.parseStructDecl(pub)
	case lexer.ENUM:
		return p.parseEnumDecl(pub)
	case lexer.TRAIT:
		return p.parseTraitDecl(pub)
	case lexer.IMPL:
		return p.parseImplBlock()
	case lexer.TYPE:
		return p.parseTypeAlias()
	case lexer.CONST:
		return p.parseConstDecl()
	case lexer.USE:
		return p.parseUseDecl()
	default:
		p.error(fmt.Sprintf("unexpected token in declaration: %v", p.curToken.Type))
		return nil
	}
}

func (p *Parser) parseFuncDecl(pub bool) *ast.FuncDecl {
	decl := &ast.FuncDecl{Pub: pub}

	p.nextToken() // consume fn

	// Parse name
	if !p.curTokenIs(lexer.IDENT) {
		p.error("expected function name")
		return nil
	}

	decl.Name = p.curToken.Literal

	// Check for generic parameters
	if p.peekTokenIs(lexer.LT) {
		p.nextToken() // consume name
		p.nextToken() // consume <

		for !p.curTokenIs(lexer.GT) && !p.curTokenIs(lexer.EOF) {
			if !p.curTokenIs(lexer.IDENT) {
				p.error("expected type parameter name")
				return nil
			}

			decl.TParams = append(decl.TParams, p.curToken.Literal)

			if p.peekTokenIs(lexer.COMMA) {
				p.nextToken() // consume param
				p.nextToken() // consume comma
			} else {
				break
			}
		}

		if !p.expectPeek(lexer.GT) {
			return nil
		}
	}

	// Parse parameters
	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	p.nextToken() // consume (

	for !p.curTokenIs(lexer.RPAREN) && !p.curTokenIs(lexer.EOF) {
		param := ast.Param{}

		// Check for mut
		if p.curTokenIs(lexer.MUT) {
			param.Mut = true

			p.nextToken()
		}

		// Handle &self / &mut self (for impl methods)
		if p.curTokenIs(lexer.AMP) {
			param.Name = "&self"

			if p.peekTokenIs(lexer.MUT) {
				p.nextToken()

				param.Name = "&mut self"
			}

			if p.peekTokenIs(lexer.IDENT) && p.peekToken.Literal == "self" {
				p.nextToken()
			}
			// self has implicit type
			param.Type = nil
		} else {
			// Parse name
			if !p.curTokenIs(lexer.IDENT) {
				p.error("expected parameter name")
				return nil
			}

			param.Name = p.curToken.Literal
			p.nextToken()

			// Parse type
			param.Type = p.parseType()
		}

		decl.Params = append(decl.Params, param)

		if p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // consume type/self
			p.nextToken() // consume comma
		} else if p.peekTokenIs(lexer.RPAREN) {
			p.nextToken() // consume type/self, move to )
			break
		} else {
			p.error("expected comma or ) after parameter")
			return nil
		}
	}

	// curToken should now be RPAREN
	if !p.curTokenIs(lexer.RPAREN) {
		p.error("expected )")
		return nil
	}

	// Check for return type
	if p.peekTokenIs(lexer.IDENT) || p.peekTokenIs(lexer.AMP) ||
		p.peekTokenIs(lexer.STAR) || p.peekTokenIs(lexer.LBRACKET) ||
		p.peekTokenIs(lexer.LPAREN) || p.peekTokenIs(lexer.VOID) {
		p.nextToken() // consume )
		decl.ReturnType = p.parseType()
	}

	// Parse body
	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	decl.Body = p.parseBlock()

	return decl
}

// Placeholder stubs for other declaration types
func (p *Parser) parseStructDecl(pub bool) *ast.StructDecl {
	decl := &ast.StructDecl{Pub: pub}

	p.nextToken() // consume struct

	// Parse name
	if !p.curTokenIs(lexer.IDENT) {
		p.error("expected struct name")
		return nil
	}

	decl.Name = p.curToken.Literal

	// Check for generic parameters
	if p.peekTokenIs(lexer.LT) {
		p.nextToken() // consume name
		p.nextToken() // consume <

		for !p.curTokenIs(lexer.GT) && !p.curTokenIs(lexer.EOF) {
			if !p.curTokenIs(lexer.IDENT) {
				p.error("expected type parameter name")
				return nil
			}

			decl.TParams = append(decl.TParams, p.curToken.Literal)

			if p.peekTokenIs(lexer.COMMA) {
				p.nextToken() // consume param
				p.nextToken() // consume comma
			} else {
				break
			}
		}

		if !p.expectPeek(lexer.GT) {
			return nil
		}
	}

	// Parse fields
	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	p.nextToken() // consume {

	// Skip newlines after {
	for p.curTokenIs(lexer.NEWLINE) {
		p.nextToken()
	}

	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		// Parse field name
		if !p.curTokenIs(lexer.IDENT) {
			p.error("expected field name")
			return nil
		}

		fieldName := p.curToken.Literal

		// Expect :
		if !p.expectPeek(lexer.COLON) {
			return nil
		}

		p.nextToken() // consume :

		// Parse type
		fieldType := p.parseType()

		decl.Fields = append(decl.Fields, ast.Field{Name: fieldName, Type: fieldType})

		// Check for comma or }
		if p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // consume type
			p.nextToken() // consume comma
			// Skip newlines after comma
			for p.curTokenIs(lexer.NEWLINE) {
				p.nextToken()
			}
		} else if p.peekTokenIs(lexer.RBRACE) {
			p.nextToken() // consume type
			break
		} else {
			p.error("expected comma or } after field")
			return nil
		}
	}

	// After loop, curToken is either at RBRACE (from loop condition) or at the last field type (from break)
	// Need to consume the } if we haven't yet
	if !p.curTokenIs(lexer.RBRACE) {
		if !p.expectPeek(lexer.RBRACE) {
			return nil
		}
	}

	return decl
}

func (p *Parser) parseEnumDecl(pub bool) *ast.EnumDecl {
	decl := &ast.EnumDecl{Pub: pub}

	p.nextToken() // consume enum

	// Parse name
	if !p.curTokenIs(lexer.IDENT) {
		p.error("expected enum name")
		return nil
	}

	decl.Name = p.curToken.Literal

	// Check for generic parameters
	if p.peekTokenIs(lexer.LT) {
		p.nextToken() // consume name
		p.nextToken() // consume <

		for !p.curTokenIs(lexer.GT) && !p.curTokenIs(lexer.EOF) {
			if !p.curTokenIs(lexer.IDENT) {
				p.error("expected type parameter name")
				return nil
			}

			decl.TParams = append(decl.TParams, p.curToken.Literal)

			if p.peekTokenIs(lexer.COMMA) {
				p.nextToken() // consume param
				p.nextToken() // consume comma
			} else {
				break
			}
		}

		if !p.expectPeek(lexer.GT) {
			return nil
		}
	}

	// Parse variants
	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	p.nextToken() // consume {

	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		// Parse variant name
		if !p.curTokenIs(lexer.IDENT) {
			p.error("expected variant name")
			return nil
		}

		variant := ast.Variant{Name: p.curToken.Literal}

		// Check for payload
		if p.peekTokenIs(lexer.LPAREN) {
			p.nextToken() // consume name
			p.nextToken() // consume (

			for !p.curTokenIs(lexer.RPAREN) && !p.curTokenIs(lexer.EOF) {
				variantType := p.parseType()
				variant.Types = append(variant.Types, variantType)

				if p.peekTokenIs(lexer.COMMA) {
					p.nextToken() // consume type
					p.nextToken() // consume comma
				} else {
					break
				}
			}

			if !p.expectPeek(lexer.RPAREN) {
				return nil
			}
		}

		decl.Variants = append(decl.Variants, variant)

		// Check for comma or }
		if p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // consume variant/paren
			p.nextToken() // consume comma
		} else if p.peekTokenIs(lexer.RBRACE) {
			p.nextToken() // consume variant/paren
			break
		} else {
			p.error("expected comma or } after variant")
			return nil
		}
	}

	// After loop, curToken is either at RBRACE (from loop condition) or at the last variant/paren (from break)
	// Need to consume the } if we haven't yet
	if !p.curTokenIs(lexer.RBRACE) {
		if !p.expectPeek(lexer.RBRACE) {
			return nil
		}
	}

	return decl
}

func (p *Parser) parseTraitDecl(pub bool) *ast.TraitDecl {
	decl := &ast.TraitDecl{Pub: pub}

	p.nextToken() // consume trait

	// Parse name
	if !p.curTokenIs(lexer.IDENT) {
		p.error("expected trait name")
		return nil
	}

	decl.Name = p.curToken.Literal

	// Check for generic parameters
	if p.peekTokenIs(lexer.LT) {
		p.nextToken() // consume name
		p.nextToken() // consume <

		for !p.curTokenIs(lexer.GT) && !p.curTokenIs(lexer.EOF) {
			if !p.curTokenIs(lexer.IDENT) {
				p.error("expected type parameter name")
				return nil
			}

			decl.TParams = append(decl.TParams, p.curToken.Literal)

			if p.peekTokenIs(lexer.COMMA) {
				p.nextToken() // consume param
				p.nextToken() // consume comma
			} else {
				break
			}
		}

		if !p.expectPeek(lexer.GT) {
			return nil
		}
	}

	// Parse method signatures
	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	p.nextToken() // consume {

	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		// Skip newlines/semicolons
		if p.curTokenIs(lexer.NEWLINE) || p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
			continue
		}

		// Parse fn signature
		if !p.curTokenIs(lexer.FN) {
			p.error("expected fn in trait")
			return nil
		}

		p.nextToken() // consume fn

		sig := ast.FnSig{}

		// Parse name
		if !p.curTokenIs(lexer.IDENT) {
			p.error("expected method name")
			return nil
		}

		sig.Name = p.curToken.Literal

		// Parse parameters (same as function)
		if !p.expectPeek(lexer.LPAREN) {
			return nil
		}

		p.nextToken() // consume (

		for !p.curTokenIs(lexer.RPAREN) && !p.curTokenIs(lexer.EOF) {
			param := ast.Param{}

			// Check for mut
			if p.curTokenIs(lexer.MUT) {
				param.Mut = true

				p.nextToken()
			}

			// Parse name
			if !p.curTokenIs(lexer.IDENT) && !p.curTokenIs(lexer.AMP) {
				p.error("expected parameter name")
				return nil
			}

			// Handle &self / &mut self
			if p.curTokenIs(lexer.AMP) {
				param.Name = "&self"

				if p.peekTokenIs(lexer.MUT) {
					p.nextToken()

					param.Name = "&mut self"
				}

				if p.peekTokenIs(lexer.IDENT) && p.peekToken.Literal == "self" {
					p.nextToken()
				}
				// self has implicit type
				param.Type = nil
			} else {
				param.Name = p.curToken.Literal
				p.nextToken()
				param.Type = p.parseType()
			}

			sig.Params = append(sig.Params, param)

			if p.peekTokenIs(lexer.COMMA) {
				p.nextToken() // consume type/self
				p.nextToken() // consume comma
			} else if p.peekTokenIs(lexer.RPAREN) {
				p.nextToken() // consume type/self, move to )
				break
			} else {
				p.error("expected comma or ) after parameter")
				return nil
			}
		}

		// curToken should now be RPAREN
		if !p.curTokenIs(lexer.RPAREN) {
			p.error("expected )")
			return nil
		}

		// Parse return type
		if p.peekTokenIs(lexer.IDENT) || p.peekTokenIs(lexer.AMP) ||
			p.peekTokenIs(lexer.STAR) || p.peekTokenIs(lexer.LBRACKET) ||
			p.peekTokenIs(lexer.LPAREN) || p.peekTokenIs(lexer.VOID) {
			p.nextToken() // consume )
			sig.Return = p.parseType()
		}

		decl.Sigs = append(decl.Sigs, sig)

		// Expect semicolon
		if p.peekTokenIs(lexer.SEMICOLON) {
			p.nextToken() // consume return type
			p.nextToken() // consume semicolon
		} else {
			p.nextToken() // consume return type/paren
		}
	}

	// curToken should be at } from loop termination
	if !p.curTokenIs(lexer.RBRACE) {
		p.error("expected } at end of trait")
		return nil
	}

	return decl
}

func (p *Parser) parseImplBlock() *ast.ImplBlock {
	impl := &ast.ImplBlock{}

	p.nextToken() // consume impl

	// Parse trait or type
	firstPath := p.parseTypePath()

	// Check for "for" (trait impl)
	if p.peekTokenIs(lexer.FOR) {
		// Trait impl: impl Trait for Type
		impl.Trait = firstPath.(*ast.TypePath)

		p.nextToken() // consume trait
		p.nextToken() // consume for
		impl.For = p.parseType()
	} else {
		// Inherent impl: impl Type
		impl.For = firstPath
	}

	// Parse methods
	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	p.nextToken() // consume {

	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		// Skip newlines/semicolons
		if p.curTokenIs(lexer.NEWLINE) || p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
			continue
		}

		// Parse function (can be pub or not)
		pub := false
		if p.curTokenIs(lexer.PUB) {
			pub = true

			p.nextToken()
		}

		if !p.curTokenIs(lexer.FN) {
			p.error("expected fn in impl block")
			return nil
		}

		fn := p.parseFuncDecl(pub)
		if fn != nil {
			impl.Fns = append(impl.Fns, fn)
		}

		p.nextToken()
	}

	// curToken should be at } from loop termination
	if !p.curTokenIs(lexer.RBRACE) {
		p.error("expected } at end of impl block")
		return nil
	}

	return impl
}

func (p *Parser) parseTypeAlias() *ast.TypeAlias {
	alias := &ast.TypeAlias{}

	p.nextToken() // consume type

	// Parse name
	if !p.curTokenIs(lexer.IDENT) {
		p.error("expected type name")
		return nil
	}

	alias.Name = p.curToken.Literal

	// Expect =
	if !p.expectPeek(lexer.ASSIGN) {
		return nil
	}

	p.nextToken() // consume =

	// Parse type
	alias.Type = p.parseType()

	return alias
}

func (p *Parser) parseConstDecl() *ast.ConstDecl {
	decl := &ast.ConstDecl{}

	p.nextToken() // consume const

	// Parse name
	if !p.curTokenIs(lexer.IDENT) {
		p.error("expected const name")
		return nil
	}

	decl.Name = p.curToken.Literal

	// Expect :
	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	p.nextToken() // consume :

	// Parse type
	decl.Type = p.parseType()

	// Expect =
	if !p.expectPeek(lexer.ASSIGN) {
		return nil
	}

	p.nextToken() // consume =

	// Parse value
	decl.Value = p.parseExpression(LOWEST)

	return decl
}

func (p *Parser) parseUseDecl() *ast.UseDecl {
	decl := &ast.UseDecl{}

	p.nextToken() // consume use

	// Parse path
	for p.curTokenIs(lexer.IDENT) {
		decl.Path = append(decl.Path, p.curToken.Literal)

		if p.peekTokenIs(lexer.COLONCOLON) {
			p.nextToken() // consume ident
			p.nextToken() // consume ::
		} else {
			break
		}
	}

	// Check for alias
	if p.peekTokenIs(lexer.AS) {
		p.nextToken() // consume last ident
		p.nextToken() // consume as

		if !p.curTokenIs(lexer.IDENT) {
			p.error("expected alias name")
			return nil
		}

		decl.Alias = p.curToken.Literal
	}

	return decl
}

// ===== File Parsing =====

// ParseFile parses a complete YarLang source file
func (p *Parser) ParseFile() *ast.File {
	file := &ast.File{Items: []ast.Decl{}}

	// Skip initial newlines
	for p.curTokenIs(lexer.NEWLINE) {
		p.nextToken()
	}

	// Check for module declaration
	if p.curTokenIs(lexer.MODULE) {
		p.nextToken() // consume module

		for p.curTokenIs(lexer.IDENT) {
			file.Module = append(file.Module, p.curToken.Literal)

			if p.peekTokenIs(lexer.COLONCOLON) {
				p.nextToken() // consume ident
				p.nextToken() // consume ::
			} else {
				break
			}
		}

		p.nextToken() // consume last ident

		// Skip newlines after module
		for p.curTokenIs(lexer.NEWLINE) || p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
		}
	}

	// Parse declarations
	for !p.curTokenIs(lexer.EOF) {
		// Skip newlines and semicolons
		if p.curTokenIs(lexer.NEWLINE) || p.curTokenIs(lexer.SEMICOLON) {
			p.nextToken()
			continue
		}

		decl := p.parseDeclaration()
		if decl != nil {
			file.Items = append(file.Items, decl)
		}

		p.nextToken()
	}

	return file
}
