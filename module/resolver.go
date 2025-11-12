package module

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yarlson/yarlang/ast"
	"github.com/yarlson/yarlang/lexer"
	"github.com/yarlson/yarlang/parser"
)

// Resolver handles module path resolution
type Resolver struct {
	projectRoot string // Path to directory containing yar.toml
	stdlibPath  string // Path to stdlib (~/.yar/stdlib)
}

// NewResolver creates a new module resolver
func NewResolver(projectRoot string) *Resolver {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to project-relative cache if home directory cannot be determined
		homeDir = filepath.Join(projectRoot, ".cache")
	}

	stdlibPath := filepath.Join(homeDir, ".yar", "stdlib")

	return &Resolver{
		projectRoot: projectRoot,
		stdlibPath:  stdlibPath,
	}
}

// Resolve resolves an import path to an absolute file path
func (r *Resolver) Resolve(importPath, fromFile string) (string, error) {
	// Handle explicit stdlib: std/math
	if strings.HasPrefix(importPath, "std/") {
		moduleName := strings.TrimPrefix(importPath, "std/")
		return r.resolveStdlib(moduleName)
	}

	// Handle relative paths: ./math, ../utils/math
	if strings.HasPrefix(importPath, "./") || strings.HasPrefix(importPath, "../") {
		return r.resolveRelative(importPath, fromFile)
	}

	// Context-aware search: local -> stdlib -> third-party
	// 1. Try local (current dir and walk up to project root)
	if resolved, err := r.resolveLocal(importPath, fromFile); err == nil {
		return resolved, nil
	}

	// 2. Try stdlib
	if resolved, err := r.resolveStdlib(importPath); err == nil {
		return resolved, nil
	}

	// 3. Not found
	return "", fmt.Errorf("module %q not found", importPath)
}

func (r *Resolver) resolveLocal(importPath, fromFile string) (string, error) {
	dir := filepath.Dir(fromFile)

	// Walk up to project root
	for {
		// Try in current directory
		candidate := filepath.Join(dir, importPath+".yar")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}

		// Try as subdirectory with same-named file
		candidate = filepath.Join(dir, importPath, importPath+".yar")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}

		// Check if we've reached project root
		if _, err := os.Stat(filepath.Join(dir, "yar.toml")); err == nil {
			break // Stop at project root
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached filesystem root
		}

		dir = parent
	}

	return "", fmt.Errorf("module %q not found locally", importPath)
}

func (r *Resolver) resolveRelative(importPath, fromFile string) (string, error) {
	dir := filepath.Dir(fromFile)
	absPath := filepath.Join(dir, importPath+".yar")

	if _, err := os.Stat(absPath); err != nil {
		return "", fmt.Errorf("module %q not found at %s", importPath, absPath)
	}

	return absPath, nil
}

func (r *Resolver) resolveStdlib(moduleName string) (string, error) {
	stdlibFile := filepath.Join(r.stdlibPath, moduleName+".yar")

	if _, err := os.Stat(stdlibFile); err != nil {
		return "", fmt.Errorf("stdlib module %q not found", moduleName)
	}

	return stdlibFile, nil
}

// FindProjectRoot walks up from dir until it finds yar.toml
func FindProjectRoot(dir string) (string, error) {
	current := dir

	for {
		yarToml := filepath.Join(current, "yar.toml")
		if _, err := os.Stat(yarToml); err == nil {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("no yar.toml found in %s or parent directories", dir)
		}

		current = parent
	}
}

// Loader handles loading and parsing modules with cycle detection
type Loader struct {
	resolver *Resolver
	cache    map[string]*Module // Path -> Module
	loading  map[string]bool    // Track modules currently being loaded (for cycle detection)
}

// NewLoader creates a new module loader
func NewLoader(projectRoot string) *Loader {
	return &Loader{
		resolver: NewResolver(projectRoot),
		cache:    make(map[string]*Module),
		loading:  make(map[string]bool),
	}
}

// Load loads a module and all its dependencies
func (l *Loader) Load(path string) (*Module, error) {
	// Normalize path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	// Check cache
	if mod, ok := l.cache[absPath]; ok {
		return mod, nil
	}

	// Check for cycle
	if l.loading[absPath] {
		return nil, fmt.Errorf("import cycle detected involving %s", absPath)
	}

	// Mark as loading
	l.loading[absPath] = true
	defer delete(l.loading, absPath)

	// Read source
	source, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", absPath, err)
	}

	// Parse the source file to get the AST
	program, imports, err := l.parseModule(string(source))
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", absPath, err)
	}

	// Create module
	mod := &Module{
		Path:    absPath,
		Name:    moduleNameFromPath(absPath),
		AST:     program,
		Imports: []*ImportRef{},
	}

	// Load imports
	for _, impInfo := range imports {
		resolved, err := l.resolver.Resolve(impInfo.Path, absPath)
		if err != nil {
			return nil, err
		}

		// Recursively load imported module (this is where cycle detection happens)
		importedMod, err := l.Load(resolved)
		if err != nil {
			return nil, err
		}

		mod.Imports = append(mod.Imports, &ImportRef{
			Path:     impInfo.Path,
			Alias:    impInfo.Alias,
			Resolved: resolved,
			Module:   importedMod,
		})
	}

	l.cache[absPath] = mod

	return mod, nil
}

// ImportInfo holds information about an import statement
type ImportInfo struct {
	Path  string
	Alias string
}

// parseModule parses a module source file and returns its AST and imports
func (l *Loader) parseModule(source string) (*ast.Program, []ImportInfo, error) {
	lex := lexer.New(source)
	p := parser.New(lex)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return nil, nil, fmt.Errorf("parse errors: %v", p.Errors())
	}

	// Extract import information from AST
	var imports []ImportInfo

	for _, stmt := range program.Statements {
		if impStmt, ok := stmt.(*ast.ImportStmt); ok {
			imports = append(imports, ImportInfo{
				Path:  impStmt.Path,
				Alias: impStmt.Alias,
			})
		}

		if impBlock, ok := stmt.(*ast.ImportBlock); ok {
			for _, impStmt := range impBlock.Imports {
				imports = append(imports, ImportInfo{
					Path:  impStmt.Path,
					Alias: impStmt.Alias,
				})
			}
		}
	}

	return program, imports, nil
}

func moduleNameFromPath(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// extractImports is a simple regex-based import extractor for testing
// In production, this will use the actual parser
func extractImports(source string) []string {
	var imports []string

	lines := strings.Split(source, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "import ") {
			// Extract path from: import "path"
			start := strings.Index(line, `"`)

			end := strings.LastIndex(line, `"`)
			if start != -1 && end != -1 && start < end {
				imports = append(imports, line[start+1:end])
			}
		}
	}

	return imports
}

// GetAllModules returns all loaded modules
func (l *Loader) GetAllModules() []*Module {
	modules := make([]*Module, 0, len(l.cache))
	for _, mod := range l.cache {
		modules = append(modules, mod)
	}

	return modules
}
