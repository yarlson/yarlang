package analysis

import "github.com/yarlson/yarlang/ast"

// SymbolKind represents the kind of symbol
type SymbolKind int

const (
	SymbolKindVariable SymbolKind = iota
	SymbolKindFunction
	SymbolKindParameter
)

func (k SymbolKind) String() string {
	switch k {
	case SymbolKindVariable:
		return "variable"
	case SymbolKindFunction:
		return "function"
	case SymbolKindParameter:
		return "parameter"
	default:
		return "unknown"
	}
}

// Symbol represents a declared symbol
type Symbol struct {
	Name       string
	Kind       SymbolKind
	DeclRange  ast.Range   // Where it's declared
	Type       string      // Inferred type or "dynamic"
	References []ast.Range // All uses of this symbol
	Node       ast.Node    // Declaration AST node
}

// Scope represents a lexical scope
type Scope struct {
	parent  *Scope
	symbols map[string]*Symbol
	node    ast.Node // Associated AST node (for range queries)
}

// NewScope creates a new scope with optional parent
func NewScope(parent *Scope) *Scope {
	return &Scope{
		parent:  parent,
		symbols: make(map[string]*Symbol),
	}
}

// Define adds a symbol to this scope
func (s *Scope) Define(name string, symbol *Symbol) {
	s.symbols[name] = symbol
}

// Lookup finds a symbol in this scope or parent scopes
func (s *Scope) Lookup(name string) *Symbol {
	// Check current scope
	if sym, ok := s.symbols[name]; ok {
		return sym
	}

	// Check parent scopes
	if s.parent != nil {
		return s.parent.Lookup(name)
	}

	return nil
}

// LookupLocal finds a symbol only in this scope (not parents)
func (s *Scope) LookupLocal(name string) *Symbol {
	if sym, ok := s.symbols[name]; ok {
		return sym
	}

	return nil
}

// AllSymbols returns all symbols visible from this scope
func (s *Scope) AllSymbols() []*Symbol {
	symbols := []*Symbol{}

	// Collect from current scope
	for _, sym := range s.symbols {
		symbols = append(symbols, sym)
	}

	// Collect from parent (avoiding duplicates from shadowing)
	if s.parent != nil {
		for _, parentSym := range s.parent.AllSymbols() {
			if s.LookupLocal(parentSym.Name) == nil {
				symbols = append(symbols, parentSym)
			}
		}
	}

	return symbols
}

// DiagnosticSeverity represents the severity of a diagnostic message
type DiagnosticSeverity int

const (
	SeverityError DiagnosticSeverity = iota
	SeverityWarning
	SeverityInfo
	SeverityHint
)

// Diagnostic represents an error or warning
type Diagnostic struct {
	Range    ast.Range
	Severity DiagnosticSeverity
	Message  string
}

// SymbolTable holds all scopes and provides lookup functions
type SymbolTable struct {
	scopes []*Scope
}

// NewSymbolTable creates a new symbol table
func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		scopes: []*Scope{},
	}
}

// AddScope adds a scope to the symbol table
func (st *SymbolTable) AddScope(scope *Scope) {
	st.scopes = append(st.scopes, scope)
}

// Scopes returns a copy of all scopes in the symbol table
func (st *SymbolTable) Scopes() []*Scope {
	result := make([]*Scope, len(st.scopes))
	copy(result, st.scopes)

	return result
}

// FindSymbolAt finds a symbol whose DeclRange or References contains the position
func (st *SymbolTable) FindSymbolAt(pos ast.Position) *Symbol {
	for _, scope := range st.scopes {
		for _, symbol := range scope.symbols {
			// Check declaration range
			if symbol.DeclRange.Contains(pos) {
				return symbol
			}
			// Check reference ranges
			for _, ref := range symbol.References {
				if ref.Contains(pos) {
					return symbol
				}
			}
		}
	}

	return nil
}

// ScopeAt finds the innermost scope containing the position
func (st *SymbolTable) ScopeAt(pos ast.Position) *Scope {
	var (
		result      *Scope
		resultDepth int
	)

	for _, scope := range st.scopes {
		if scope.node == nil {
			continue
		}

		nodeRange := getNodeRange(scope.node)
		if nodeRange.Contains(pos) {
			// Calculate depth
			depth := 0
			for s := scope; s != nil; s = s.parent {
				depth++
			}

			// Keep the deepest (most nested) scope
			if result == nil || depth > resultDepth {
				result = scope
				resultDepth = depth
			}
		}
	}

	return result
}

// getNodeRange extracts the range from an AST node
func getNodeRange(node ast.Node) ast.Range {
	switch n := node.(type) {
	case *ast.FuncDecl:
		return n.Range
	case *ast.BlockStmt:
		return n.Range
	case *ast.ForStmt:
		return n.Range
	case *ast.IfStmt:
		return n.Range
	default:
		return ast.Range{}
	}
}
