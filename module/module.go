package module

import "github.com/yarlson/yarlang/ast"

// Module represents a compiled YarLang module
type Module struct {
	Path         string         // Absolute file path
	Name         string         // Module name (filename without extension)
	Namespace    string         // Namespace for imports (name or alias)
	AST          *ast.Program   // Parsed AST
	Imports      []*ImportRef   // Modules this one imports
	ContentHash  string         // SHA-256 hash of source content
	IRPath       string         // Path to compiled IR file
}

// ImportRef represents an import relationship
type ImportRef struct {
	Path      string // Import path as written (e.g., "math", "./utils")
	Alias     string // Optional alias
	Resolved  string // Absolute path to resolved module
	Module    *Module // Resolved module (nil until loaded)
}
