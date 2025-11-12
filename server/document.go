package server

import (
	"github.com/yarlson/yarlang/analysis"
	"github.com/yarlson/yarlang/ast"
	"github.com/yarlson/yarlang/lexer"
	"github.com/yarlson/yarlang/parser"
)

// Document represents an open text document
type Document struct {
	URI         string
	Version     int
	Content     string
	AST         *ast.Program
	Symbols     *analysis.SymbolTable
	Diagnostics []analysis.Diagnostic
}

// Parse parses the document content and updates AST
func (d *Document) Parse() {
	l := lexer.New(d.Content)
	p := parser.New(l)
	d.AST = p.ParseProgram()

	// Collect parse errors as diagnostics
	d.Diagnostics = []analysis.Diagnostic{}
	for _, err := range p.Errors() {
		// TODO: Extract position information from error message
		d.Diagnostics = append(d.Diagnostics, analysis.Diagnostic{
			Severity: analysis.SeverityError,
			Message:  err,
		})
	}
}

// Analyze performs semantic analysis on the parsed AST
func (d *Document) Analyze() {
	if d.AST == nil {
		return
	}

	symbols, errors := analysis.Analyze(d.AST)
	d.Symbols = symbols

	// Add semantic errors to diagnostics
	d.Diagnostics = append(d.Diagnostics, errors...)
}

// Update updates the document content, version, and re-analyzes
func (d *Document) Update(content string, version int) {
	d.Content = content
	d.Version = version
	d.Parse()
	d.Analyze()
}
