package parser

import (
	"fmt"
	"strconv"

	"github.com/yarlson/yarlang/ast"
	"github.com/yarlson/yarlang/lexer"
)

// Precedence levels
const (
	_ int = iota
	LOWEST
	OR          // ||
	AND         // &&
	EQUALS      // == !=
	LESSGREATER // < > <= >=
	SUM         // + -
	PRODUCT     // * / %
	PREFIX      // ! -
	CALL        // function()
)

var precedences = map[lexer.TokenType]int{
	lexer.OR:      OR,
	lexer.AND:     AND,
	lexer.EQ:      EQUALS,
	lexer.NEQ:     EQUALS,
	lexer.LT:      LESSGREATER,
	lexer.GT:      LESSGREATER,
	lexer.LTE:     LESSGREATER,
	lexer.GTE:     LESSGREATER,
	lexer.PLUS:    SUM,
	lexer.MINUS:   SUM,
	lexer.STAR:    PRODUCT,
	lexer.SLASH:   PRODUCT,
	lexer.PERCENT: PRODUCT,
	lexer.LPAREN:  CALL,
}

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  lexer.Token
	peekToken lexer.Token

	prefixParseFns map[lexer.TokenType]prefixParseFn
	infixParseFns  map[lexer.TokenType]infixParseFn
}

type (
	prefixParseFn func() ast.Expr
	infixParseFn  func(ast.Expr) ast.Expr
)

// Position helper functions

func (p *Parser) tokenStartPos() ast.Position {
	return ast.Position{
		Line:   p.curToken.Line,
		Column: p.curToken.Column,
		Offset: -1, // We don't track byte offsets yet
	}
}

func (p *Parser) tokenEndPos() ast.Position {
	return ast.Position{
		Line:   p.curToken.Line,
		Column: p.curToken.Column + len(p.curToken.Literal),
		Offset: -1,
	}
}

func (p *Parser) tokenRange() ast.Range {
	return ast.Range{
		Start: p.tokenStartPos(),
		End:   p.tokenEndPos(),
	}
}

func (p *Parser) makeRange(start, end ast.Position) ast.Range {
	return ast.Range{Start: start, End: end}
}

// getExprRange extracts the Range from any expression node
func getExprRange(expr ast.Expr) ast.Range {
	switch e := expr.(type) {
	case *ast.Identifier:
		return e.Range
	case *ast.NumberLiteral:
		return e.Range
	case *ast.StringLiteral:
		return e.Range
	case *ast.BoolLiteral:
		return e.Range
	case *ast.NilLiteral:
		return e.Range
	case *ast.BinaryExpr:
		return e.Range
	case *ast.UnaryExpr:
		return e.Range
	case *ast.CallExpr:
		return e.Range
	default:
		return ast.Range{}
	}
}

// getStmtRange extracts the Range from any statement node
func getStmtRange(stmt ast.Stmt) ast.Range {
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		return s.Range
	case *ast.AssignStmt:
		return s.Range
	case *ast.ReturnStmt:
		return s.Range
	case *ast.IfStmt:
		return s.Range
	case *ast.ForStmt:
		return s.Range
	case *ast.BreakStmt:
		return s.Range
	case *ast.ContinueStmt:
		return s.Range
	case *ast.BlockStmt:
		return s.Range
	case *ast.FuncDecl:
		return s.Range
	default:
		return ast.Range{}
	}
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	// Register prefix parse functions
	p.prefixParseFns = make(map[lexer.TokenType]prefixParseFn)
	p.registerPrefix(lexer.IDENT, p.parseIdentifier)
	p.registerPrefix(lexer.NUMBER, p.parseNumberLiteral)
	p.registerPrefix(lexer.STRING, p.parseStringLiteral)
	p.registerPrefix(lexer.TRUE, p.parseBoolLiteral)
	p.registerPrefix(lexer.FALSE, p.parseBoolLiteral)
	p.registerPrefix(lexer.NIL, p.parseNilLiteral)
	p.registerPrefix(lexer.NOT, p.parseUnaryExpr)
	p.registerPrefix(lexer.MINUS, p.parseUnaryExpr)
	p.registerPrefix(lexer.LPAREN, p.parseGroupedExpr)

	// Register infix parse functions
	p.infixParseFns = make(map[lexer.TokenType]infixParseFn)
	p.registerInfix(lexer.PLUS, p.parseBinaryExpr)
	p.registerInfix(lexer.MINUS, p.parseBinaryExpr)
	p.registerInfix(lexer.STAR, p.parseBinaryExpr)
	p.registerInfix(lexer.SLASH, p.parseBinaryExpr)
	p.registerInfix(lexer.PERCENT, p.parseBinaryExpr)
	p.registerInfix(lexer.EQ, p.parseBinaryExpr)
	p.registerInfix(lexer.NEQ, p.parseBinaryExpr)
	p.registerInfix(lexer.LT, p.parseBinaryExpr)
	p.registerInfix(lexer.GT, p.parseBinaryExpr)
	p.registerInfix(lexer.LTE, p.parseBinaryExpr)
	p.registerInfix(lexer.GTE, p.parseBinaryExpr)
	p.registerInfix(lexer.AND, p.parseBinaryExpr)
	p.registerInfix(lexer.OR, p.parseBinaryExpr)
	p.registerInfix(lexer.LPAREN, p.parseCallExpr)

	// Read two tokens to initialize curToken and peekToken
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead at line %d",
		t, p.peekToken.Type, p.peekToken.Line)
	p.errors = append(p.errors, msg)
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

	p.peekError(t)

	return false
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

func (p *Parser) registerPrefix(tokenType lexer.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType lexer.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// ParseProgram parses the entire program
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{
		Statements: []ast.Stmt{},
	}

	for !p.curTokenIs(lexer.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}

		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Stmt {
	switch p.curToken.Type {
	case lexer.RETURN:
		return p.parseReturnStmt()
	case lexer.IF:
		return p.parseIfStmt()
	case lexer.FOR:
		return p.parseForStmt()
	case lexer.BREAK:
		r := p.tokenRange()
		return &ast.BreakStmt{Range: r}
	case lexer.CONTINUE:
		r := p.tokenRange()
		return &ast.ContinueStmt{Range: r}
	case lexer.FUNC:
		return p.parseFuncDecl()
	case lexer.LBRACE:
		return p.parseBlockStmt()
	default:
		return p.parseExprOrAssignStmt()
	}
}

func (p *Parser) parseExprOrAssignStmt() ast.Stmt {
	// Try to parse as assignment or expression statement
	expr := p.parseExpression(LOWEST)

	// Check if next token is assignment or comma (multiple assignment)
	if p.peekTokenIs(lexer.ASSIGN) || p.peekTokenIs(lexer.COMMA) {
		return p.parseAssignStmt(expr)
	}

	exprRange := getExprRange(expr)

	return &ast.ExprStmt{
		Expr:  expr,
		Range: exprRange,
	}
}

func (p *Parser) parseAssignStmt(firstExpr ast.Expr) ast.Stmt {
	targets := []string{}
	targetRanges := []ast.Range{}

	// First target
	ident, ok := firstExpr.(*ast.Identifier)
	if !ok {
		p.errors = append(p.errors, "assignment target must be identifier")
		return nil
	}

	targets = append(targets, ident.Name)
	targetRanges = append(targetRanges, ident.Range)

	start := ident.Range.Start

	// Additional targets if comma-separated
	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // consume comma
		p.nextToken() // move to identifier

		if !p.curTokenIs(lexer.IDENT) {
			p.errors = append(p.errors, "expected identifier after comma in assignment")
			return nil
		}

		targets = append(targets, p.curToken.Literal)
		targetRanges = append(targetRanges, p.tokenRange())
	}

	if !p.expectPeek(lexer.ASSIGN) {
		return nil
	}

	p.nextToken() // move to first value

	values := []ast.Expr{}
	values = append(values, p.parseExpression(LOWEST))

	// Additional values if comma-separated
	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // consume comma
		p.nextToken() // move to expression
		values = append(values, p.parseExpression(LOWEST))
	}

	// End position is end of last value
	var end ast.Position

	if len(values) > 0 {
		lastExprRange := getExprRange(values[len(values)-1])
		end = lastExprRange.End
	} else {
		end = p.tokenEndPos()
	}

	return &ast.AssignStmt{
		Targets:      targets,
		Values:       values,
		Range:        p.makeRange(start, end),
		TargetRanges: targetRanges,
	}
}

func (p *Parser) parseExpression(precedence int) ast.Expr {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}

	leftExp := prefix()

	for !p.peekTokenIs(lexer.EOF) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) noPrefixParseFnError(t lexer.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found at line %d", t, p.curToken.Line)
	p.errors = append(p.errors, msg)
}

// Parse functions for expressions

func (p *Parser) parseIdentifier() ast.Expr {
	return &ast.Identifier{
		Name:  p.curToken.Literal,
		Range: p.tokenRange(),
	}
}

func (p *Parser) parseNumberLiteral() ast.Expr {
	r := p.tokenRange()

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as number", p.curToken.Literal)
		p.errors = append(p.errors, msg)

		return nil
	}

	return &ast.NumberLiteral{Value: value, Range: r}
}

func (p *Parser) parseStringLiteral() ast.Expr {
	// For strings, the token literal doesn't include quotes, but the range should
	r := ast.Range{
		Start: p.tokenStartPos(),
		End: ast.Position{
			Line:   p.curToken.Line,
			Column: p.curToken.Column + len(p.curToken.Literal) + 2, // +2 for opening and closing quotes
			Offset: -1,
		},
	}

	return &ast.StringLiteral{
		Value: p.curToken.Literal,
		Range: r,
	}
}

func (p *Parser) parseBoolLiteral() ast.Expr {
	return &ast.BoolLiteral{
		Value: p.curTokenIs(lexer.TRUE),
		Range: p.tokenRange(),
	}
}

func (p *Parser) parseNilLiteral() ast.Expr {
	return &ast.NilLiteral{
		Range: p.tokenRange(),
	}
}

func (p *Parser) parseUnaryExpr() ast.Expr {
	start := p.tokenStartPos()
	operator := p.curToken.Literal

	p.nextToken()

	right := p.parseExpression(PREFIX)
	if right == nil {
		return nil
	}

	rightRange := getExprRange(right)

	return &ast.UnaryExpr{
		Operator: operator,
		Right:    right,
		Range:    p.makeRange(start, rightRange.End),
	}
}

func (p *Parser) parseBinaryExpr(left ast.Expr) ast.Expr {
	leftRange := getExprRange(left)
	operator := p.curToken.Literal
	precedence := p.curPrecedence()

	p.nextToken()

	right := p.parseExpression(precedence)
	if right == nil {
		return nil
	}

	rightRange := getExprRange(right)

	return &ast.BinaryExpr{
		Left:     left,
		Operator: operator,
		Right:    right,
		Range:    p.makeRange(leftRange.Start, rightRange.End),
	}
}

func (p *Parser) parseGroupedExpr() ast.Expr {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseCallExpr(function ast.Expr) ast.Expr {
	funcRange := getExprRange(function)
	start := funcRange.Start

	args := []ast.Expr{}

	p.nextToken() // move past (

	if p.curTokenIs(lexer.RPAREN) {
		end := p.tokenEndPos()

		return &ast.CallExpr{
			Function: function,
			Args:     args,
			Range:    p.makeRange(start, end),
		}
	}

	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // consume comma
		p.nextToken() // move to expression
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	end := p.tokenEndPos()

	return &ast.CallExpr{
		Function: function,
		Args:     args,
		Range:    p.makeRange(start, end),
	}
}

// Parse functions for statements (stubs for now)

func (p *Parser) parseReturnStmt() ast.Stmt {
	start := p.tokenStartPos()

	p.nextToken()

	values := []ast.Expr{}

	if !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		values = append(values, p.parseExpression(LOWEST))

		for p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // consume comma
			p.nextToken() // move to expression
			values = append(values, p.parseExpression(LOWEST))
		}
	}

	// Determine end position
	var end ast.Position

	if len(values) > 0 {
		lastExprRange := getExprRange(values[len(values)-1])
		end = lastExprRange.End
	} else {
		// "return" with no values - end is after keyword
		end = start
		end.Column += 6 // len("return")
	}

	return &ast.ReturnStmt{
		Values: values,
		Range:  p.makeRange(start, end),
	}
}

func (p *Parser) parseIfStmt() ast.Stmt {
	start := p.tokenStartPos() // position of 'if'

	p.nextToken() // move past 'if'

	condition := p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	thenBlock := p.parseBlockStmt()

	var (
		elseBlock *ast.BlockStmt
		end       ast.Position
	)

	end = thenBlock.Range.End

	if p.peekTokenIs(lexer.ELSE) {
		p.nextToken() // consume 'else'

		if p.peekTokenIs(lexer.IF) {
			// else if - wrap in block
			p.nextToken()
			elseIf := p.parseIfStmt()
			elseIfRange := getStmtRange(elseIf)
			end = elseIfRange.End
			elseBlock = &ast.BlockStmt{
				Statements: []ast.Stmt{elseIf},
				Range:      elseIfRange,
			}
		} else if p.peekTokenIs(lexer.LBRACE) {
			p.nextToken()
			elseBlock = p.parseBlockStmt()
			end = elseBlock.Range.End
		}
	}

	return &ast.IfStmt{
		Condition: condition,
		ThenBlock: thenBlock,
		ElseBlock: elseBlock,
		Range:     p.makeRange(start, end),
	}
}

func (p *Parser) parseForStmt() ast.Stmt {
	start := p.tokenStartPos() // position of 'for'

	p.nextToken() // move past 'for'

	// for { } - infinite loop
	if p.curTokenIs(lexer.LBRACE) {
		body := p.parseBlockStmt()

		return &ast.ForStmt{
			Body:  body,
			Range: p.makeRange(start, body.Range.End),
		}
	}

	// Parse the first part
	firstStmt := p.parseStatement()

	// Check if we have a semicolon (3-part for loop)
	if p.peekTokenIs(lexer.SEMICOLON) {
		// 3-part for loop: for init; cond; post { }
		p.nextToken() // consume semicolon
		p.nextToken() // move to condition

		var condition ast.Expr
		if !p.curTokenIs(lexer.SEMICOLON) {
			condition = p.parseExpression(LOWEST)
		}

		if !p.expectPeek(lexer.SEMICOLON) {
			return nil
		}

		p.nextToken() // move to post statement

		var post ast.Stmt
		if !p.curTokenIs(lexer.LBRACE) {
			post = p.parseStatement()
		}

		if !p.expectPeek(lexer.LBRACE) {
			return nil
		}

		body := p.parseBlockStmt()

		return &ast.ForStmt{
			Init:      firstStmt,
			Condition: condition,
			Post:      post,
			Body:      body,
			Range:     p.makeRange(start, body.Range.End),
		}
	}

	// While-style: for condition { }
	var condition ast.Expr
	if exprStmt, ok := firstStmt.(*ast.ExprStmt); ok {
		condition = exprStmt.Expr
	}

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	body := p.parseBlockStmt()

	return &ast.ForStmt{
		Condition: condition,
		Body:      body,
		Range:     p.makeRange(start, body.Range.End),
	}
}

func (p *Parser) parseBlockStmt() *ast.BlockStmt {
	start := p.tokenStartPos() // position of {

	block := &ast.BlockStmt{Statements: []ast.Stmt{}}

	p.nextToken() // move past {

	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}

		p.nextToken()
	}

	end := p.tokenEndPos() // position after }

	block.Range = p.makeRange(start, end)

	return block
}

func (p *Parser) parseFuncDecl() ast.Stmt {
	start := p.tokenStartPos() // position of 'func'

	p.nextToken() // move past 'func'

	if !p.curTokenIs(lexer.IDENT) {
		p.errors = append(p.errors, "expected function name")
		return nil
	}

	name := p.curToken.Literal
	nameRange := p.tokenRange()

	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	params := []string{}
	paramRanges := []ast.Range{}

	p.nextToken() // move past (

	if !p.curTokenIs(lexer.RPAREN) {
		if !p.curTokenIs(lexer.IDENT) {
			p.errors = append(p.errors, "expected parameter name")
			return nil
		}

		params = append(params, p.curToken.Literal)
		paramRanges = append(paramRanges, p.tokenRange())

		for p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // consume comma
			p.nextToken() // move to identifier

			if !p.curTokenIs(lexer.IDENT) {
				p.errors = append(p.errors, "expected parameter name")
				return nil
			}

			params = append(params, p.curToken.Literal)
			paramRanges = append(paramRanges, p.tokenRange())
		}

		if !p.expectPeek(lexer.RPAREN) {
			return nil
		}
	}

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	body := p.parseBlockStmt()

	end := body.Range.End

	return &ast.FuncDecl{
		Name:        name,
		Params:      params,
		Body:        body,
		Range:       p.makeRange(start, end),
		NameRange:   nameRange,
		ParamRanges: paramRanges,
	}
}
