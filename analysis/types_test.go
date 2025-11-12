package analysis

import (
	"testing"

	"github.com/yarlson/yarlang/ast"
)

func TestScopeDefineAndLookup(t *testing.T) {
	parentScope := NewScope(nil)
	childScope := NewScope(parentScope)

	// Define symbol in parent
	parentSym := &Symbol{
		Name: "x",
		Kind: SymbolKindVariable,
	}
	parentScope.Define("x", parentSym)

	// Define symbol in child
	childSym := &Symbol{
		Name: "y",
		Kind: SymbolKindVariable,
	}
	childScope.Define("y", childSym)

	// Lookup in parent scope
	if got := parentScope.Lookup("x"); got != parentSym {
		t.Error("Failed to lookup symbol in parent scope")
	}

	// Lookup in child scope - should find in parent
	if got := childScope.Lookup("x"); got != parentSym {
		t.Error("Failed to lookup parent symbol from child scope")
	}

	// Lookup child symbol
	if got := childScope.Lookup("y"); got != childSym {
		t.Error("Failed to lookup symbol in child scope")
	}

	// Lookup non-existent symbol
	if got := childScope.Lookup("z"); got != nil {
		t.Error("Expected nil for non-existent symbol")
	}
}

func TestScopeShadowing(t *testing.T) {
	parentScope := NewScope(nil)
	childScope := NewScope(parentScope)

	// Define x in parent
	parentSym := &Symbol{Name: "x", Kind: SymbolKindVariable}
	parentScope.Define("x", parentSym)

	// Define x in child (shadowing)
	childSym := &Symbol{Name: "x", Kind: SymbolKindParameter}
	childScope.Define("x", childSym)

	// Child lookup should return child's x
	if got := childScope.Lookup("x"); got != childSym {
		t.Error("Child scope should shadow parent symbol")
	}

	// Parent lookup should still return parent's x
	if got := parentScope.Lookup("x"); got != parentSym {
		t.Error("Parent scope should be unchanged by child shadowing")
	}
}

func TestSymbolTableFindSymbolAt(t *testing.T) {
	st := NewSymbolTable()

	// Create a symbol with declaration and references
	sym := &Symbol{
		Name: "foo",
		Kind: SymbolKindVariable,
		DeclRange: ast.Range{
			Start: ast.Position{Line: 1, Column: 5},
			End:   ast.Position{Line: 1, Column: 8},
		},
		References: []ast.Range{
			{
				Start: ast.Position{Line: 2, Column: 10},
				End:   ast.Position{Line: 2, Column: 13},
			},
			{
				Start: ast.Position{Line: 3, Column: 15},
				End:   ast.Position{Line: 3, Column: 18},
			},
		},
	}

	scope := NewScope(nil)
	scope.Define("foo", sym)
	st.AddScope(scope)

	// Test finding at declaration position
	declPos := ast.Position{Line: 1, Column: 6}

	found := st.FindSymbolAt(declPos)
	if found != sym {
		t.Error("Failed to find symbol at declaration position")
	}

	// Test finding at exact start of declaration
	startPos := ast.Position{Line: 1, Column: 5}

	found = st.FindSymbolAt(startPos)
	if found != sym {
		t.Error("Failed to find symbol at declaration start position")
	}

	// Test finding at exact end of declaration
	endPos := ast.Position{Line: 1, Column: 8}

	found = st.FindSymbolAt(endPos)
	if found != sym {
		t.Error("Failed to find symbol at declaration end position")
	}

	// Test finding at first reference
	refPos1 := ast.Position{Line: 2, Column: 11}

	found = st.FindSymbolAt(refPos1)
	if found != sym {
		t.Error("Failed to find symbol at first reference position")
	}

	// Test finding at second reference
	refPos2 := ast.Position{Line: 3, Column: 16}

	found = st.FindSymbolAt(refPos2)
	if found != sym {
		t.Error("Failed to find symbol at second reference position")
	}

	// Test not finding at wrong position
	wrongPos := ast.Position{Line: 5, Column: 1}

	found = st.FindSymbolAt(wrongPos)
	if found != nil {
		t.Error("Should not find symbol at wrong position")
	}

	// Test not finding at position just before declaration
	beforePos := ast.Position{Line: 1, Column: 4}

	found = st.FindSymbolAt(beforePos)
	if found != nil {
		t.Error("Should not find symbol just before declaration")
	}

	// Test not finding at position just after declaration
	afterPos := ast.Position{Line: 1, Column: 9}

	found = st.FindSymbolAt(afterPos)
	if found != nil {
		t.Error("Should not find symbol just after declaration")
	}
}

func TestSymbolTableFindSymbolAtMultipleScopes(t *testing.T) {
	st := NewSymbolTable()

	// Create symbols in different scopes
	sym1 := &Symbol{
		Name: "x",
		Kind: SymbolKindVariable,
		DeclRange: ast.Range{
			Start: ast.Position{Line: 1, Column: 0},
			End:   ast.Position{Line: 1, Column: 1},
		},
	}

	sym2 := &Symbol{
		Name: "y",
		Kind: SymbolKindFunction,
		DeclRange: ast.Range{
			Start: ast.Position{Line: 3, Column: 5},
			End:   ast.Position{Line: 3, Column: 6},
		},
	}

	scope1 := NewScope(nil)
	scope1.Define("x", sym1)
	st.AddScope(scope1)

	scope2 := NewScope(scope1)
	scope2.Define("y", sym2)
	st.AddScope(scope2)

	// Find in first scope
	found := st.FindSymbolAt(ast.Position{Line: 1, Column: 0})
	if found != sym1 {
		t.Error("Failed to find symbol in first scope")
	}

	// Find in second scope
	found = st.FindSymbolAt(ast.Position{Line: 3, Column: 5})
	if found != sym2 {
		t.Error("Failed to find symbol in second scope")
	}
}

func TestSymbolTableFindSymbolAtWithNoReferences(t *testing.T) {
	st := NewSymbolTable()

	// Create a symbol with no references
	sym := &Symbol{
		Name: "unused",
		Kind: SymbolKindVariable,
		DeclRange: ast.Range{
			Start: ast.Position{Line: 1, Column: 0},
			End:   ast.Position{Line: 1, Column: 6},
		},
		References: []ast.Range{}, // Empty references
	}

	scope := NewScope(nil)
	scope.Define("unused", sym)
	st.AddScope(scope)

	// Should find at declaration
	found := st.FindSymbolAt(ast.Position{Line: 1, Column: 3})
	if found != sym {
		t.Error("Failed to find symbol with no references")
	}

	// Should not find elsewhere
	found = st.FindSymbolAt(ast.Position{Line: 2, Column: 0})
	if found != nil {
		t.Error("Should not find symbol outside declaration")
	}
}

func TestScopeAtSimple(t *testing.T) {
	st := NewSymbolTable()

	// Create a scope with a block statement
	block := &ast.BlockStmt{
		Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 0},
			End:   ast.Position{Line: 10, Column: 0},
		},
	}

	scope := NewScope(nil)
	scope.node = block
	st.AddScope(scope)

	// Find scope at position inside block
	found := st.ScopeAt(ast.Position{Line: 5, Column: 5})
	if found != scope {
		t.Error("Failed to find scope at position inside block")
	}

	// Should not find scope outside block
	found = st.ScopeAt(ast.Position{Line: 15, Column: 0})
	if found != nil {
		t.Error("Should not find scope outside block range")
	}
}

func TestScopeAtNested(t *testing.T) {
	st := NewSymbolTable()

	// Create parent scope
	parentBlock := &ast.BlockStmt{
		Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 0},
			End:   ast.Position{Line: 20, Column: 0},
		},
	}
	parentScope := NewScope(nil)
	parentScope.node = parentBlock
	st.AddScope(parentScope)

	// Create child scope (nested)
	childBlock := &ast.BlockStmt{
		Range: ast.Range{
			Start: ast.Position{Line: 5, Column: 2},
			End:   ast.Position{Line: 10, Column: 2},
		},
	}
	childScope := NewScope(parentScope)
	childScope.node = childBlock
	st.AddScope(childScope)

	// Position in child scope should return child (innermost)
	found := st.ScopeAt(ast.Position{Line: 7, Column: 3})
	if found != childScope {
		t.Error("Should find innermost (child) scope")
	}

	// Position in parent but outside child should return parent
	found = st.ScopeAt(ast.Position{Line: 3, Column: 1})
	if found != parentScope {
		t.Error("Should find parent scope when outside child")
	}

	// Position outside both should return nil
	found = st.ScopeAt(ast.Position{Line: 25, Column: 0})
	if found != nil {
		t.Error("Should not find scope outside all ranges")
	}
}

func TestScopeAtWithFuncDecl(t *testing.T) {
	st := NewSymbolTable()

	// Create scope with function declaration
	funcDecl := &ast.FuncDecl{
		Name: "test",
		Range: ast.Range{
			Start: ast.Position{Line: 1, Column: 0},
			End:   ast.Position{Line: 10, Column: 0},
		},
	}

	scope := NewScope(nil)
	scope.node = funcDecl
	st.AddScope(scope)

	// Find scope at position inside function
	found := st.ScopeAt(ast.Position{Line: 5, Column: 5})
	if found != scope {
		t.Error("Failed to find scope for function declaration")
	}
}

func TestScopeAtWithForStmt(t *testing.T) {
	st := NewSymbolTable()

	// Create scope with for statement
	forStmt := &ast.ForStmt{
		Range: ast.Range{
			Start: ast.Position{Line: 5, Column: 0},
			End:   ast.Position{Line: 10, Column: 0},
		},
	}

	scope := NewScope(nil)
	scope.node = forStmt
	st.AddScope(scope)

	// Find scope at position inside for loop
	found := st.ScopeAt(ast.Position{Line: 7, Column: 2})
	if found != scope {
		t.Error("Failed to find scope for for statement")
	}
}

func TestScopeAtWithIfStmt(t *testing.T) {
	st := NewSymbolTable()

	// Create scope with if statement
	ifStmt := &ast.IfStmt{
		Range: ast.Range{
			Start: ast.Position{Line: 3, Column: 0},
			End:   ast.Position{Line: 8, Column: 0},
		},
	}

	scope := NewScope(nil)
	scope.node = ifStmt
	st.AddScope(scope)

	// Find scope at position inside if statement
	found := st.ScopeAt(ast.Position{Line: 5, Column: 1})
	if found != scope {
		t.Error("Failed to find scope for if statement")
	}
}

func TestScopeAtWithNoNode(t *testing.T) {
	st := NewSymbolTable()

	// Create scope without node (global scope)
	scope := NewScope(nil)
	scope.node = nil
	st.AddScope(scope)

	// Should not find scope without node
	found := st.ScopeAt(ast.Position{Line: 1, Column: 0})
	if found != nil {
		t.Error("Should not find scope without associated node")
	}
}
