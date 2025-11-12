package semantic

import (
	"fmt"

	"github.com/yarlson/yarlang/ast"
)

type Scope struct {
	parent  *Scope
	symbols map[string]bool
}

func newScope(parent *Scope) *Scope {
	return &Scope{
		parent:  parent,
		symbols: make(map[string]bool),
	}
}

func (s *Scope) define(name string) {
	s.symbols[name] = true
}

func (s *Scope) resolve(name string) bool {
	if _, ok := s.symbols[name]; ok {
		return true
	}

	if s.parent != nil {
		return s.parent.resolve(name)
	}

	return false
}

type Analyzer struct {
	currentScope *Scope
	errors       []string
}

func New() *Analyzer {
	global := newScope(nil)

	// Define built-in functions
	global.define("print")
	global.define("println")
	global.define("len")
	global.define("type")

	return &Analyzer{
		currentScope: global,
		errors:       []string{},
	}
}

func (a *Analyzer) Analyze(program *ast.Program) error {
	for _, stmt := range program.Statements {
		a.analyzeStmt(stmt)
	}

	if len(a.errors) > 0 {
		return fmt.Errorf("%s", a.errors[0])
	}

	return nil
}

func (a *Analyzer) enterScope() {
	a.currentScope = newScope(a.currentScope)
}

func (a *Analyzer) exitScope() {
	a.currentScope = a.currentScope.parent
}

func (a *Analyzer) analyzeStmt(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		a.analyzeExpr(s.Expr)
	case *ast.AssignStmt:
		// Analyze right-hand side first
		for _, expr := range s.Values {
			a.analyzeExpr(expr)
		}
		// Define variables in current scope
		for _, target := range s.Targets {
			a.currentScope.define(target)
		}
	case *ast.ReturnStmt:
		for _, expr := range s.Values {
			a.analyzeExpr(expr)
		}
	case *ast.IfStmt:
		a.analyzeExpr(s.Condition)
		a.enterScope()
		a.analyzeBlockStmt(s.ThenBlock)
		a.exitScope()

		if s.ElseBlock != nil {
			a.enterScope()
			a.analyzeBlockStmt(s.ElseBlock)
			a.exitScope()
		}
	case *ast.ForStmt:
		a.enterScope()

		if s.Init != nil {
			a.analyzeStmt(s.Init)
		}

		if s.Condition != nil {
			a.analyzeExpr(s.Condition)
		}

		if s.Post != nil {
			a.analyzeStmt(s.Post)
		}

		a.analyzeBlockStmt(s.Body)
		a.exitScope()
	case *ast.BlockStmt:
		a.analyzeBlockStmt(s)
	case *ast.FuncDecl:
		// Define function in current scope
		a.currentScope.define(s.Name)

		// Create new scope for function body
		a.enterScope()

		// Define parameters
		for _, param := range s.Params {
			a.currentScope.define(param)
		}

		// Analyze body
		a.analyzeBlockStmt(s.Body)

		a.exitScope()
	case *ast.BreakStmt, *ast.ContinueStmt:
		// TODO: verify these are inside loops
	}
}

func (a *Analyzer) analyzeBlockStmt(block *ast.BlockStmt) {
	for _, stmt := range block.Statements {
		a.analyzeStmt(stmt)
	}
}

func (a *Analyzer) analyzeExpr(expr ast.Expr) {
	switch e := expr.(type) {
	case *ast.Identifier:
		if !a.currentScope.resolve(e.Name) {
			a.errors = append(a.errors, fmt.Sprintf("undefined variable: %s", e.Name))
		}
	case *ast.BinaryExpr:
		a.analyzeExpr(e.Left)
		a.analyzeExpr(e.Right)
	case *ast.UnaryExpr:
		a.analyzeExpr(e.Right)
	case *ast.CallExpr:
		a.analyzeExpr(e.Function)

		for _, arg := range e.Args {
			a.analyzeExpr(arg)
		}
	case *ast.NumberLiteral, *ast.StringLiteral, *ast.BoolLiteral, *ast.NilLiteral:
		// Literals are always valid
	}
}
