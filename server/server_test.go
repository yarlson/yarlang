package server

import (
	"context"
	"testing"

	"go.lsp.dev/protocol"
)

func TestServerInitialize(t *testing.T) {
	srv := New()

	params := &protocol.InitializeParams{}

	result, err := srv.Initialize(context.Background(), params)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if result.ServerInfo.Name != "yarlang-lsp" {
		t.Errorf("Server name = %s, want yarlang-lsp", result.ServerInfo.Name)
	}

	// Check capabilities
	if result.Capabilities.CompletionProvider == nil {
		t.Error("Expected CompletionProvider capability")
	}

	if result.Capabilities.HoverProvider == nil {
		t.Error("Expected HoverProvider capability")
	}

	if result.Capabilities.DefinitionProvider == nil {
		t.Error("Expected DefinitionProvider capability")
	}
}

func TestServerDidOpen(t *testing.T) {
	srv := New()

	// Track diagnostics published
	var publishedDiags []protocol.Diagnostic

	srv.DiagnosticCallback = func(uri string, diags []protocol.Diagnostic) {
		publishedDiags = diags
	}
	_ = publishedDiags // Variable tracked for diagnostic callback verification

	params := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:     "file:///test.yar",
			Version: 1,
			Text:    "x = 42",
		},
	}

	err := srv.DidOpen(context.Background(), params)
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Check document is cached
	doc, ok := srv.documents["file:///test.yar"]
	if !ok {
		t.Fatal("Expected document to be cached")
	}

	if doc.Content != "x = 42" {
		t.Errorf("Document content = %s, want 'x = 42'", doc.Content)
	}

	if doc.AST == nil {
		t.Error("Expected AST to be parsed")
	}

	if doc.Symbols == nil {
		t.Error("Expected Symbols to be analyzed")
	}
}

func TestServerDidChange(t *testing.T) {
	srv := New()

	// Open document first
	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:     "file:///test.yar",
			Version: 1,
			Text:    "x = 1",
		},
	}
	if err := srv.DidOpen(context.Background(), openParams); err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Change document
	changeParams := &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{
				URI: "file:///test.yar",
			},
			Version: 2,
		},
		ContentChanges: []protocol.TextDocumentContentChangeEvent{
			{Text: "y = 2"},
		},
	}

	err := srv.DidChange(context.Background(), changeParams)
	if err != nil {
		t.Fatalf("DidChange failed: %v", err)
	}

	// Check document was updated
	doc := srv.documents["file:///test.yar"]
	if doc.Version != 2 {
		t.Errorf("Document version = %d, want 2", doc.Version)
	}

	if doc.Content != "y = 2" {
		t.Errorf("Document content = %s, want 'y = 2'", doc.Content)
	}
}

func TestServerDidChangeNotFound(t *testing.T) {
	srv := New()

	// Try to change a document that was never opened
	changeParams := &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{
				URI: "file:///notfound.yar",
			},
			Version: 1,
		},
		ContentChanges: []protocol.TextDocumentContentChangeEvent{
			{Text: "x = 1"},
		},
	}

	err := srv.DidChange(context.Background(), changeParams)
	if err == nil {
		t.Fatal("Expected error when changing non-existent document")
	}

	expectedErr := "document not found: file:///notfound.yar"
	if err.Error() != expectedErr {
		t.Errorf("Error message = %s, want %s", err.Error(), expectedErr)
	}
}

func TestServerDidClose(t *testing.T) {
	srv := New()

	// Open a document
	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:     "file:///test.yar",
			Version: 1,
			Text:    "x = 42",
		},
	}
	if err := srv.DidOpen(context.Background(), openParams); err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Verify document is cached
	if _, ok := srv.documents["file:///test.yar"]; !ok {
		t.Fatal("Expected document to be cached after open")
	}

	// Close the document
	closeParams := &protocol.DidCloseTextDocumentParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: "file:///test.yar",
		},
	}

	err := srv.DidClose(context.Background(), closeParams)
	if err != nil {
		t.Fatalf("DidClose failed: %v", err)
	}

	// Verify document is removed from cache
	if _, ok := srv.documents["file:///test.yar"]; ok {
		t.Error("Expected document to be removed after close")
	}
}

func TestServerDiagnosticPublishing(t *testing.T) {
	srv := New()

	// Track diagnostics
	var (
		capturedURI   string
		capturedDiags []protocol.Diagnostic
	)

	srv.DiagnosticCallback = func(uri string, diags []protocol.Diagnostic) {
		capturedURI = uri
		capturedDiags = diags
	}

	// Open document with valid YarLang code
	params := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:     "file:///test.yar",
			Version: 1,
			Text:    "x = 42\ny = x + 10",
		},
	}

	err := srv.DidOpen(context.Background(), params)
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Verify diagnostics were published
	if capturedURI != "file:///test.yar" {
		t.Errorf("Diagnostic URI = %s, want file:///test.yar", capturedURI)
	}

	// Diagnostics slice should exist (may be empty if no errors)
	if capturedDiags == nil {
		t.Error("Expected diagnostics to be published")
	}
}

func TestServerCompletion(t *testing.T) {
	srv := New()

	// Open document with some symbols
	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:     "file:///test.yar",
			Version: 1,
			Text: `
x = 42
func add(a, b) {
	return a + b
}
`,
		},
	}
	if err := srv.DidOpen(context.Background(), openParams); err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Request completion at line 4 (inside function)
	compParams := &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: "file:///test.yar",
			},
			Position: protocol.Position{Line: 3, Character: 10},
		},
	}

	result, err := srv.Completion(context.Background(), compParams)
	if err != nil {
		t.Fatalf("Completion failed: %v", err)
	}

	if result == nil || len(result.Items) == 0 {
		t.Fatal("Expected completion items")
	}

	// Should have parameters a, b, global x, function add, and keywords
	hasA := false
	hasX := false
	hasKeyword := false

	for _, item := range result.Items {
		if item.Label == "a" {
			hasA = true
		}

		if item.Label == "x" {
			hasX = true
		}

		if item.Label == "func" {
			hasKeyword = true
		}
	}

	if !hasA {
		t.Error("Expected parameter 'a' in completion")
	}

	if !hasX {
		t.Error("Expected variable 'x' in completion")
	}

	if !hasKeyword {
		t.Error("Expected keyword 'func' in completion")
	}
}

func TestServerCompletionGlobalScope(t *testing.T) {
	srv := New()

	// Open document with global scope symbols only
	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:     "file:///test.yar",
			Version: 1,
			Text: `
x = 42
y = 100
`,
		},
	}
	if err := srv.DidOpen(context.Background(), openParams); err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Request completion at line 3 (after declarations)
	compParams := &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: "file:///test.yar",
			},
			Position: protocol.Position{Line: 2, Character: 0},
		},
	}

	result, err := srv.Completion(context.Background(), compParams)
	if err != nil {
		t.Fatalf("Completion failed: %v", err)
	}

	if result == nil || len(result.Items) == 0 {
		t.Fatal("Expected completion items")
	}

	// Should have variables x, y, and keywords
	hasX := false
	hasY := false

	for _, item := range result.Items {
		if item.Label == "x" {
			hasX = true

			if item.Kind != protocol.CompletionItemKindVariable {
				t.Errorf("Expected x to be a variable, got %v", item.Kind)
			}
		}

		if item.Label == "y" {
			hasY = true
		}
	}

	if !hasX {
		t.Error("Expected variable 'x' in completion")
	}

	if !hasY {
		t.Error("Expected variable 'y' in completion")
	}
}

func TestServerCompletionNoDocument(t *testing.T) {
	srv := New()

	// Request completion without opening document
	compParams := &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: "file:///notfound.yar",
			},
			Position: protocol.Position{Line: 0, Character: 0},
		},
	}

	result, err := srv.Completion(context.Background(), compParams)
	if err != nil {
		t.Fatalf("Completion failed: %v", err)
	}

	// Should return empty list, not nil
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if len(result.Items) != 0 {
		t.Errorf("Expected empty completion items, got %d items", len(result.Items))
	}
}

func TestServerCompletionEmptyDocument(t *testing.T) {
	srv := New()

	// Open empty document
	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:     "file:///empty.yar",
			Version: 1,
			Text:    "",
		},
	}
	if err := srv.DidOpen(context.Background(), openParams); err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Request completion
	compParams := &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: "file:///empty.yar",
			},
			Position: protocol.Position{Line: 0, Character: 0},
		},
	}

	result, err := srv.Completion(context.Background(), compParams)
	if err != nil {
		t.Fatalf("Completion failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Should have at least keywords
	hasKeyword := false

	for _, item := range result.Items {
		if item.Kind == protocol.CompletionItemKindKeyword {
			hasKeyword = true
			break
		}
	}

	if !hasKeyword {
		t.Error("Expected at least keywords in completion for empty document")
	}
}

func TestServerHover(t *testing.T) {
	srv := New()

	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:     "file:///test.yar",
			Version: 1,
			Text: `
x = 42
func add(a, b) {
	return a + b
}
result = add(1, 2)
`,
		},
	}
	if err := srv.DidOpen(context.Background(), openParams); err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Hover over 'add' in function call
	hoverParams := &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: "file:///test.yar",
			},
			Position: protocol.Position{Line: 5, Character: 10}, // position of 'add'
		},
	}

	result, err := srv.Hover(context.Background(), hoverParams)
	if err != nil {
		t.Fatalf("Hover failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected hover result")
	}

	// Should contain function information
	content := result.Contents.Value
	if content == "" {
		t.Error("Expected hover content")
	}
}

func TestServerDefinition(t *testing.T) {
	srv := New()

	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:     "file:///test.yar",
			Version: 1,
			Text: `
x = 42
y = x + 1
`,
		},
	}
	if err := srv.DidOpen(context.Background(), openParams); err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Go to definition of 'x' in line 3
	defParams := &protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: "file:///test.yar",
			},
			Position: protocol.Position{Line: 2, Character: 5}, // position of x in "y = x + 1"
		},
	}

	result, err := srv.Definition(context.Background(), defParams)
	if err != nil {
		t.Fatalf("Definition failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected definition location")
	}

	// Should point to line 2 (where x is declared)
	loc := result[0]
	if loc.Range.Start.Line != 1 { // 0-based, so line 2 is index 1
		t.Errorf("Definition line = %d, want 1", loc.Range.Start.Line)
	}
}

func TestServerHoverNoSymbol(t *testing.T) {
	srv := New()

	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:     "file:///test.yar",
			Version: 1,
			Text:    "x = 42",
		},
	}
	if err := srv.DidOpen(context.Background(), openParams); err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Hover over the '=' sign (not a symbol)
	hoverParams := &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: "file:///test.yar",
			},
			Position: protocol.Position{Line: 0, Character: 2}, // '=' sign
		},
	}

	result, err := srv.Hover(context.Background(), hoverParams)
	if err != nil {
		t.Fatalf("Hover failed: %v", err)
	}

	// Should return nil when no symbol found
	if result != nil {
		t.Error("Expected nil result when no symbol at position")
	}
}

func TestServerHoverVariable(t *testing.T) {
	srv := New()

	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:     "file:///test.yar",
			Version: 1,
			Text:    "myVar = 42",
		},
	}
	if err := srv.DidOpen(context.Background(), openParams); err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Hover over 'myVar'
	hoverParams := &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: "file:///test.yar",
			},
			Position: protocol.Position{Line: 0, Character: 2}, // 'm' in myVar
		},
	}

	result, err := srv.Hover(context.Background(), hoverParams)
	if err != nil {
		t.Fatalf("Hover failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected hover result for variable")
	}

	content := result.Contents.Value
	if content == "" {
		t.Error("Expected hover content for variable")
	}

	// Should mention it's a variable
	if !contains(content, "Variable") && !contains(content, "myVar") {
		t.Errorf("Expected variable info in hover content, got: %s", content)
	}
}

func TestServerHoverParameter(t *testing.T) {
	srv := New()

	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:     "file:///test.yar",
			Version: 1,
			Text: `
func test(param1, param2) {
	x = param1 + param2
}
`,
		},
	}
	if err := srv.DidOpen(context.Background(), openParams); err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Hover over 'param1' in function body
	hoverParams := &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: "file:///test.yar",
			},
			Position: protocol.Position{Line: 2, Character: 7}, // 'param1' reference
		},
	}

	result, err := srv.Hover(context.Background(), hoverParams)
	if err != nil {
		t.Fatalf("Hover failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected hover result for parameter")
	}

	content := result.Contents.Value
	if content == "" {
		t.Error("Expected hover content for parameter")
	}

	// Should mention it's a parameter or show param1
	if !contains(content, "Parameter") && !contains(content, "param") && !contains(content, "param1") {
		t.Errorf("Expected parameter info in hover content, got: %s", content)
	}
}

func TestServerDefinitionNoSymbol(t *testing.T) {
	srv := New()

	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:     "file:///test.yar",
			Version: 1,
			Text:    "x = 42",
		},
	}
	if err := srv.DidOpen(context.Background(), openParams); err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Try definition on empty space
	defParams := &protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: "file:///test.yar",
			},
			Position: protocol.Position{Line: 0, Character: 2}, // '=' sign
		},
	}

	result, err := srv.Definition(context.Background(), defParams)
	if err != nil {
		t.Fatalf("Definition failed: %v", err)
	}

	// Should return nil when no symbol found
	if result != nil {
		t.Error("Expected nil result when no symbol at position")
	}
}

func TestServerDefinitionFromReference(t *testing.T) {
	srv := New()

	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:     "file:///test.yar",
			Version: 1,
			Text: `
x = 10
y = 20
z = x + y
result = x * 2
`,
		},
	}
	if err := srv.DidOpen(context.Background(), openParams); err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Go to definition of 'x' from last line
	defParams := &protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: "file:///test.yar",
			},
			Position: protocol.Position{Line: 4, Character: 10}, // 'x' in "result = x * 2"
		},
	}

	result, err := srv.Definition(context.Background(), defParams)
	if err != nil {
		t.Fatalf("Definition failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected definition location")
	}

	// Should point to line 1 where x is declared (0-based)
	loc := result[0]
	if loc.Range.Start.Line != 1 {
		t.Errorf("Definition line = %d, want 1", loc.Range.Start.Line)
	}
}

func TestServerDefinitionFunction(t *testing.T) {
	srv := New()

	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:     "file:///test.yar",
			Version: 1,
			Text: `
func multiply(a, b) {
	return a * b
}

result = multiply(5, 3)
`,
		},
	}
	if err := srv.DidOpen(context.Background(), openParams); err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Go to definition of 'multiply' from call site
	defParams := &protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: "file:///test.yar",
			},
			Position: protocol.Position{Line: 5, Character: 10}, // 'multiply' in call
		},
	}

	result, err := srv.Definition(context.Background(), defParams)
	if err != nil {
		t.Fatalf("Definition failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected definition location for function")
	}

	// Should point to function declaration line (line 1, 0-based)
	loc := result[0]
	if loc.Range.Start.Line != 1 {
		t.Errorf("Definition line = %d, want 1", loc.Range.Start.Line)
	}
}

func TestServerHoverNoDocument(t *testing.T) {
	srv := New()

	// Try hover without opening document
	hoverParams := &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: "file:///notfound.yar",
			},
			Position: protocol.Position{Line: 0, Character: 0},
		},
	}

	result, err := srv.Hover(context.Background(), hoverParams)
	if err != nil {
		t.Fatalf("Hover failed: %v", err)
	}

	if result != nil {
		t.Error("Expected nil result for non-existent document")
	}
}

func TestServerDefinitionNoDocument(t *testing.T) {
	srv := New()

	// Try definition without opening document
	defParams := &protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: "file:///notfound.yar",
			},
			Position: protocol.Position{Line: 0, Character: 0},
		},
	}

	result, err := srv.Definition(context.Background(), defParams)
	if err != nil {
		t.Fatalf("Definition failed: %v", err)
	}

	if result != nil {
		t.Error("Expected nil result for non-existent document")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}

func TestServerCompletionNestedScope(t *testing.T) {
	srv := New()

	// Open document with nested scopes
	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:     "file:///test.yar",
			Version: 1,
			Text: `
x = 1
func outer(a) {
	y = 2
	func inner(b) {
		z = 3
		return z + b + y + a + x
	}
	return inner(5)
}
`,
		},
	}
	if err := srv.DidOpen(context.Background(), openParams); err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Request completion inside inner function at line 6
	compParams := &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: "file:///test.yar",
			},
			Position: protocol.Position{Line: 6, Character: 10},
		},
	}

	result, err := srv.Completion(context.Background(), compParams)
	if err != nil {
		t.Fatalf("Completion failed: %v", err)
	}

	if result == nil || len(result.Items) == 0 {
		t.Fatal("Expected completion items")
	}

	// Should have all symbols: z, b (inner params), y (outer var), a (outer params), x (global), and both functions
	hasZ := false
	hasB := false
	hasY := false
	hasA := false
	hasX := false

	for _, item := range result.Items {
		switch item.Label {
		case "z":
			hasZ = true
		case "b":
			hasB = true
		case "y":
			hasY = true
		case "a":
			hasA = true
		case "x":
			hasX = true
		}
	}

	if !hasZ {
		t.Error("Expected variable 'z' in completion (innermost scope)")
	}

	if !hasB {
		t.Error("Expected parameter 'b' in completion (inner function)")
	}

	if !hasY {
		t.Error("Expected variable 'y' in completion (outer function)")
	}

	if !hasA {
		t.Error("Expected parameter 'a' in completion (outer function)")
	}

	if !hasX {
		t.Error("Expected variable 'x' in completion (global scope)")
	}
}

func TestServerCompletionFunctionKinds(t *testing.T) {
	srv := New()

	// Open document with different symbol kinds
	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:     "file:///test.yar",
			Version: 1,
			Text: `
func myFunc(param1, param2) {
	localVar = 10
	return param1 + localVar
}
`,
		},
	}
	if err := srv.DidOpen(context.Background(), openParams); err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Request completion inside function
	compParams := &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: "file:///test.yar",
			},
			Position: protocol.Position{Line: 3, Character: 10},
		},
	}

	result, err := srv.Completion(context.Background(), compParams)
	if err != nil {
		t.Fatalf("Completion failed: %v", err)
	}

	// Check that items have correct kinds
	for _, item := range result.Items {
		switch item.Label {
		case "myFunc":
			if item.Kind != protocol.CompletionItemKindFunction {
				t.Errorf("Expected myFunc to be CompletionItemKindFunction, got %v", item.Kind)
			}
		case "param1", "param2", "localVar":
			if item.Kind != protocol.CompletionItemKindVariable {
				t.Errorf("Expected %s to be CompletionItemKindVariable, got %v", item.Label, item.Kind)
			}
		case "func", "if", "else", "for", "return":
			if item.Kind != protocol.CompletionItemKindKeyword {
				t.Errorf("Expected %s to be CompletionItemKindKeyword, got %v", item.Label, item.Kind)
			}
		}
	}
}
