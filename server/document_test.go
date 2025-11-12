package server

import (
	"testing"
)

func TestDocumentParse(t *testing.T) {
	doc := &Document{
		URI:     "file:///test.yar",
		Version: 1,
		Content: "x = 42",
	}

	doc.Parse()

	if doc.AST == nil {
		t.Fatal("Expected AST to be populated")
	}

	if len(doc.AST.Statements) != 1 {
		t.Errorf("Expected 1 statement, got %d", len(doc.AST.Statements))
	}
}

func TestDocumentAnalyze(t *testing.T) {
	doc := &Document{
		URI:     "file:///test.yar",
		Version: 1,
		Content: "z = x + 1",
	}

	doc.Parse()
	doc.Analyze()

	if doc.Symbols == nil {
		t.Fatal("Expected Symbols to be populated")
	}

	// Should have diagnostic for undefined x
	hasError := false

	for _, diag := range doc.Diagnostics {
		if diag.Message == "undefined: x" {
			hasError = true
			break
		}
	}

	if !hasError {
		t.Error("Expected diagnostic for undefined variable")
	}
}

func TestDocumentUpdate(t *testing.T) {
	doc := &Document{
		URI:     "file:///test.yar",
		Version: 1,
		Content: "x = 1",
	}

	doc.Update("y = 2", 2)

	if doc.Version != 2 {
		t.Errorf("Version = %d, want 2", doc.Version)
	}

	if doc.Content != "y = 2" {
		t.Errorf("Content = %s, want 'y = 2'", doc.Content)
	}

	if doc.AST == nil {
		t.Error("Expected AST to be reparsed after update")
	}
}
