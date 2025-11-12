package semantic

import (
	"fmt"
	"strings"

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
		// Skip validation for qualified names (e.g., math.Add) - they're cross-module
		if !strings.Contains(e.Name, ".") && !a.currentScope.resolve(e.Name) {
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

// ModuleInfo holds information about a module for cross-module analysis
type ModuleInfo struct {
	Name    string
	AST     *ast.Program
	Exports map[string]*Symbol // Exported symbols only
}

// Symbol represents an exported symbol
type Symbol struct {
	Name       string
	Kind       SymbolKind
	Exported   bool
	Definition ast.Node
}

type SymbolKind int

const (
	SymbolFunction SymbolKind = iota
	SymbolVariable
)

// CrossModuleAnalyzer performs semantic analysis across multiple modules
type CrossModuleAnalyzer struct {
	modules map[string]*ModuleInfo
	scopes  map[string]*Scope // Module name -> root scope
}

// NewCrossModuleAnalyzer creates a cross-module analyzer
func NewCrossModuleAnalyzer(modules map[string]*ModuleInfo) *CrossModuleAnalyzer {
	return &CrossModuleAnalyzer{
		modules: modules,
		scopes:  make(map[string]*Scope),
	}
}

// Analyze analyzes a module and all its imports
func (a *CrossModuleAnalyzer) Analyze(moduleName string) error {
	module, ok := a.modules[moduleName]
	if !ok {
		return fmt.Errorf("module %q not found", moduleName)
	}

	// First pass: collect exports from all modules
	for name, mod := range a.modules {
		if err := a.collectExports(name, mod); err != nil {
			return err
		}
	}

	// Second pass: analyze the target module with imported symbols
	return a.analyzeModule(moduleName, module)
}

func (a *CrossModuleAnalyzer) collectExports(name string, mod *ModuleInfo) error {
	exports := make(map[string]*Symbol)

	// Walk AST and find exported functions (capital letter)
	for _, stmt := range mod.AST.Statements {
		if funcDecl, ok := stmt.(*ast.FuncDecl); ok {
			// Exported if starts with capital letter
			if len(funcDecl.Name) > 0 && isExported(funcDecl.Name) {
				exports[funcDecl.Name] = &Symbol{
					Name:       funcDecl.Name,
					Kind:       SymbolFunction,
					Exported:   true,
					Definition: funcDecl,
				}
			}
		}
	}

	mod.Exports = exports

	return nil
}

func (a *CrossModuleAnalyzer) analyzeModule(name string, mod *ModuleInfo) error {
	// Create scope for this module
	scope := NewScope(nil)
	a.scopes[name] = scope

	// Process imports
	imports := make(map[string]*ModuleInfo) // alias/name -> module

	for _, stmt := range mod.AST.Statements {
		if imp, ok := stmt.(*ast.ImportStmt); ok {
			// Find the imported module
			importedMod, ok := a.modules[imp.Path]
			if !ok {
				return fmt.Errorf("imported module %q not found", imp.Path)
			}

			// Use alias if provided, otherwise use module name
			namespace := imp.Path
			if imp.Alias != "" {
				namespace = imp.Alias
			}

			imports[namespace] = importedMod
		} else if block, ok := stmt.(*ast.ImportBlock); ok {
			for _, imp := range block.Imports {
				// Find the imported module
				importedMod, ok := a.modules[imp.Path]
				if !ok {
					return fmt.Errorf("imported module %q not found", imp.Path)
				}

				// Use alias if provided, otherwise use module name
				namespace := imp.Path
				if imp.Alias != "" {
					namespace = imp.Alias
				}

				imports[namespace] = importedMod
			}
		}
	}

	// Walk AST and check symbol references
	return a.checkReferences(mod.AST, scope, imports)
}

func (a *CrossModuleAnalyzer) checkReferences(program *ast.Program, scope *Scope, imports map[string]*ModuleInfo) error {
	for _, stmt := range program.Statements {
		if err := a.checkStmt(stmt, scope, imports); err != nil {
			return err
		}
	}

	return nil
}

func (a *CrossModuleAnalyzer) checkStmt(stmt ast.Stmt, scope *Scope, imports map[string]*ModuleInfo) error {
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		return a.checkExpr(s.Expr, scope, imports)
	case *ast.AssignStmt:
		for _, val := range s.Values {
			if err := a.checkExpr(val, scope, imports); err != nil {
				return err
			}
		}
	case *ast.FuncDecl:
		// Add function to scope
		scope.define(s.Name)

		// Check function body
		funcScope := NewScope(scope)
		for _, param := range s.Params {
			funcScope.define(param)
		}

		for _, stmt := range s.Body.Statements {
			if err := a.checkStmt(stmt, funcScope, imports); err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *CrossModuleAnalyzer) checkExpr(expr ast.Expr, scope *Scope, imports map[string]*ModuleInfo) error {
	switch e := expr.(type) {
	case *ast.CallExpr:
		// Check for qualified call: module.Function()
		if ident, ok := e.Function.(*ast.Identifier); ok {
			// Check if it's a qualified name (contains ".")
			if containsDot(ident.Name) {
				parts := splitFirst(ident.Name, ".")
				moduleName := parts[0]
				symbolName := parts[1]

				// Check if module is imported
				importedMod, ok := imports[moduleName]
				if !ok {
					return fmt.Errorf("module %q not imported", moduleName)
				}

				// Check if symbol is exported
				symbol, ok := importedMod.Exports[symbolName]
				if !ok {
					return fmt.Errorf("%q is not exported from module %q", symbolName, moduleName)
				}

				if !symbol.Exported {
					return fmt.Errorf("%q is not exported from module %q (lowercase)", symbolName, moduleName)
				}
			}
		}

		// Check arguments
		for _, arg := range e.Args {
			if err := a.checkExpr(arg, scope, imports); err != nil {
				return err
			}
		}
	}

	return nil
}

func isExported(name string) bool {
	if len(name) == 0 {
		return false
	}
	// Check if first character is uppercase
	firstChar := rune(name[0])

	return firstChar >= 'A' && firstChar <= 'Z'
}

// NewScope creates a new scope with a parent
func NewScope(parent *Scope) *Scope {
	return &Scope{
		parent:  parent,
		symbols: make(map[string]bool),
	}
}

// containsDot checks if a string contains a dot
func containsDot(s string) bool {
	for _, c := range s {
		if c == '.' {
			return true
		}
	}

	return false
}

// splitFirst splits a string at the first occurrence of sep
func splitFirst(s, sep string) [2]string {
	for i, c := range s {
		if string(c) == sep {
			return [2]string{s[:i], s[i+1:]}
		}
	}

	return [2]string{s, ""}
}
