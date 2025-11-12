package analysis

import (
	"fmt"

	"github.com/yarlson/yarlang/ast"
)

// Analyzer performs semantic analysis on an AST
type Analyzer struct {
	program      *ast.Program
	symbols      *SymbolTable
	diagnostics  []Diagnostic
	currentScope *Scope
}

// Analyze performs semantic analysis on a program
func Analyze(program *ast.Program) (*SymbolTable, []Diagnostic) {
	a := &Analyzer{
		program:     program,
		symbols:     NewSymbolTable(),
		diagnostics: []Diagnostic{},
	}

	a.analyzeProgram()

	return a.symbols, a.diagnostics
}

func (a *Analyzer) analyzeProgram() {
	// Create global scope
	globalScope := NewScope(nil)
	a.symbols.AddScope(globalScope)
	a.currentScope = globalScope

	// Add built-in functions to global scope
	a.addBuiltins(globalScope)

	// Analyze all statements
	for _, stmt := range a.program.Statements {
		a.analyzeStmt(stmt)
	}
}

func (a *Analyzer) addBuiltins(scope *Scope) {
	// Add built-in functions
	builtins := []string{
		"println",
		"print",
		"len",
		"panic",
	}

	for _, name := range builtins {
		scope.Define(name, &Symbol{
			Name:      name,
			Kind:      SymbolKindFunction,
			Type:      "builtin",
			DeclRange: ast.Range{}, // No source location for builtins
		})
	}
}

func (a *Analyzer) analyzeStmt(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.FuncDecl:
		a.analyzeFuncDecl(s)
	case *ast.AssignStmt:
		a.analyzeAssignStmt(s)
	case *ast.IfStmt:
		a.analyzeIfStmt(s)
	case *ast.ForStmt:
		a.analyzeForStmt(s)
	case *ast.ReturnStmt:
		a.analyzeReturnStmt(s)
	case *ast.ExprStmt:
		a.analyzeExpr(s.Expr)
	case *ast.BlockStmt:
		a.analyzeBlockStmt(s)
	case *ast.BreakStmt, *ast.ContinueStmt:
		// Nothing to analyze
	}
}

func (a *Analyzer) analyzeFuncDecl(fn *ast.FuncDecl) {
	// Add function to current scope
	symbol := &Symbol{
		Name:      fn.Name,
		Kind:      SymbolKindFunction,
		DeclRange: fn.NameRange,
		Type:      "function",
		Node:      fn,
	}

	if existing := a.currentScope.LookupLocal(fn.Name); existing != nil {
		a.diagnostics = append(a.diagnostics, Diagnostic{
			Range:    fn.NameRange,
			Severity: SeverityError,
			Message:  fmt.Sprintf("function %s already declared", fn.Name),
		})
	} else {
		a.currentScope.Define(fn.Name, symbol)
	}

	// Create function scope
	fnScope := NewScope(a.currentScope)
	fnScope.node = fn
	a.symbols.AddScope(fnScope)

	// Save current scope and switch to function scope
	outerScope := a.currentScope
	a.currentScope = fnScope

	// Add parameters to function scope
	for i, param := range fn.Params {
		paramSym := &Symbol{
			Name:      param,
			Kind:      SymbolKindParameter,
			DeclRange: fn.ParamRanges[i],
			Type:      "dynamic",
		}
		fnScope.Define(param, paramSym)
	}

	// Analyze function body statements directly in function scope
	// Don't create another block scope for the function body
	for _, stmt := range fn.Body.Statements {
		a.analyzeStmt(stmt)
	}

	// Restore outer scope
	a.currentScope = outerScope
}

func (a *Analyzer) analyzeAssignStmt(assign *ast.AssignStmt) {
	// Analyze right-hand side first
	for _, expr := range assign.Values {
		a.analyzeExpr(expr)
	}

	// Check arity
	if len(assign.Targets) != len(assign.Values) {
		a.diagnostics = append(a.diagnostics, Diagnostic{
			Range:    assign.Range,
			Severity: SeverityError,
			Message: fmt.Sprintf("assignment count mismatch: %d = %d",
				len(assign.Targets), len(assign.Values)),
		})
	}

	// Left side: declare or assign
	for i, target := range assign.Targets {
		targetRange := assign.TargetRanges[i]

		if existing := a.currentScope.Lookup(target); existing != nil {
			// Assignment to existing variable
			existing.References = append(existing.References, targetRange)
		} else {
			// New variable declaration in current scope
			symbol := &Symbol{
				Name:      target,
				Kind:      SymbolKindVariable,
				DeclRange: targetRange,
				Type:      "dynamic",
			}
			a.currentScope.Define(target, symbol)
		}
	}
}

func (a *Analyzer) analyzeIfStmt(ifStmt *ast.IfStmt) {
	// Analyze condition
	a.analyzeExpr(ifStmt.Condition)

	// Analyze then block (creates new scope)
	a.analyzeBlockStmt(ifStmt.ThenBlock)

	// Analyze else block if present
	if ifStmt.ElseBlock != nil {
		a.analyzeBlockStmt(ifStmt.ElseBlock)
	}
}

func (a *Analyzer) analyzeForStmt(forStmt *ast.ForStmt) {
	// Create scope for for loop
	forScope := NewScope(a.currentScope)
	forScope.node = forStmt
	a.symbols.AddScope(forScope)

	outerScope := a.currentScope
	a.currentScope = forScope

	// Analyze init statement
	if forStmt.Init != nil {
		a.analyzeStmt(forStmt.Init)
	}

	// Analyze condition
	if forStmt.Condition != nil {
		a.analyzeExpr(forStmt.Condition)
	}

	// Analyze post statement
	if forStmt.Post != nil {
		a.analyzeStmt(forStmt.Post)
	}

	// Analyze body statements directly in for scope
	// Don't create another block scope for the for loop body
	for _, stmt := range forStmt.Body.Statements {
		a.analyzeStmt(stmt)
	}

	a.currentScope = outerScope
}

func (a *Analyzer) analyzeReturnStmt(ret *ast.ReturnStmt) {
	for _, expr := range ret.Values {
		a.analyzeExpr(expr)
	}
}

func (a *Analyzer) analyzeBlockStmt(block *ast.BlockStmt) {
	// Block creates a new scope
	blockScope := NewScope(a.currentScope)
	blockScope.node = block
	a.symbols.AddScope(blockScope)

	outerScope := a.currentScope
	a.currentScope = blockScope

	for _, stmt := range block.Statements {
		a.analyzeStmt(stmt)
	}

	a.currentScope = outerScope
}

func (a *Analyzer) analyzeExpr(expr ast.Expr) {
	switch e := expr.(type) {
	case *ast.Identifier:
		// Check if identifier is defined
		if sym := a.currentScope.Lookup(e.Name); sym != nil {
			sym.References = append(sym.References, e.Range)
		} else {
			a.diagnostics = append(a.diagnostics, Diagnostic{
				Range:    e.Range,
				Severity: SeverityError,
				Message:  fmt.Sprintf("undefined: %s", e.Name),
			})
		}
	case *ast.CallExpr:
		a.analyzeExpr(e.Function)

		for _, arg := range e.Args {
			a.analyzeExpr(arg)
		}
	case *ast.BinaryExpr:
		a.analyzeExpr(e.Left)
		a.analyzeExpr(e.Right)
	case *ast.UnaryExpr:
		a.analyzeExpr(e.Right)
	// Literals don't need analysis
	case *ast.NumberLiteral, *ast.StringLiteral, *ast.BoolLiteral, *ast.NilLiteral:
		// Nothing to do
	}
}
