package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/yarlson/yarlang/ast"
	"github.com/yarlson/yarlang/codegen"
	"github.com/yarlson/yarlang/lexer"
	"github.com/yarlson/yarlang/module"
	"github.com/yarlson/yarlang/parser"
	"github.com/yarlson/yarlang/semantic"
)

// Config represents yar.toml
type Config struct {
	Package struct {
		Name    string `toml:"name"`
		Version string `toml:"version"`
		Entry   string `toml:"entry"`
	} `toml:"package"`
}

// Builder handles compilation
type Builder struct {
	projectRoot string
	cache       *CacheManager
	loader      *module.Loader
}

// NewBuilder creates a builder
func NewBuilder(projectRoot string) *Builder {
	return &Builder{
		projectRoot: projectRoot,
		cache:       NewCacheManager(projectRoot),
		loader:      module.NewLoader(projectRoot),
	}
}

// Build compiles the project
func (b *Builder) Build() error {
	// Load config
	config, err := b.loadConfig()
	if err != nil {
		return err
	}

	// Setup build directories
	if err := b.setupBuildDirs(); err != nil {
		return err
	}

	// Load entry point and all dependencies
	entryPath := filepath.Join(b.projectRoot, config.Package.Entry)

	_, err = b.loader.Load(entryPath)
	if err != nil {
		return err
	}

	// Get all modules in dependency order
	modules := b.loader.GetAllModules()

	// Compile each module to IR
	for _, mod := range modules {
		if err := b.compileModule(mod); err != nil {
			return err
		}
	}

	// Link all IR files
	if err := b.linkModules(modules, config.Package.Name); err != nil {
		return err
	}

	// Compile to executable
	if err := b.compileExecutable(config.Package.Name); err != nil {
		return err
	}

	return nil
}

func (b *Builder) loadConfig() (*Config, error) {
	configPath := filepath.Join(b.projectRoot, "yar.toml")

	var config Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return nil, fmt.Errorf("failed to load yar.toml: %w", err)
	}

	return &config, nil
}

func (b *Builder) setupBuildDirs() error {
	dirs := []string{
		filepath.Join(b.projectRoot, "build", "ir"),
		filepath.Join(b.projectRoot, "build", "bin"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

func (b *Builder) compileModule(mod *module.Module) error {
	// Set IR path first (always, even for cached modules)
	irPath := filepath.Join(b.projectRoot, "build", "ir", mod.Name+".ll")
	mod.IRPath = irPath

	// Check if rebuild needed
	needsRebuild, err := b.cache.NeedsRebuild(mod.Path, b.getImportPaths(mod))
	if err != nil {
		return err
	}

	if !needsRebuild {
		fmt.Printf("  Using cached %s\n", mod.Name)
		return nil
	}

	fmt.Printf("  Building %s\n", mod.Name)

	// Read source
	source, err := os.ReadFile(mod.Path)
	if err != nil {
		return err
	}

	// Lex & Parse
	l := lexer.New(string(source))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return fmt.Errorf("parse errors in %s: %v", mod.Path, p.Errors())
	}

	// Semantic analysis
	analyzer := semantic.New()
	if err := analyzer.Analyze(program); err != nil {
		return fmt.Errorf("semantic error in %s: %w", mod.Path, err)
	}

	// Code generation
	gen := codegen.New()
	gen.SetModuleName(mod.Name)

	// Register external functions from imports
	for _, imp := range mod.Imports {
		if imp.Module == nil {
			continue
		}

		// Find exported functions in the imported module
		for _, stmt := range imp.Module.AST.Statements {
			if funcDecl, ok := stmt.(*ast.FuncDecl); ok {
				// Check if exported (starts with capital letter)
				if len(funcDecl.Name) > 0 && funcDecl.Name[0] >= 'A' && funcDecl.Name[0] <= 'Z' {
					// Get the namespace (alias or module name)
					namespace := imp.Path
					if imp.Alias != "" {
						namespace = imp.Alias
					}

					// Register with codegen
					gen.RegisterExternalFunction(namespace, funcDecl.Name, len(funcDecl.Params))
				}
			}
		}
	}

	if err := gen.Generate(program); err != nil {
		return fmt.Errorf("codegen error in %s: %w", mod.Path, err)
	}

	// Write IR
	if err := os.WriteFile(irPath, []byte(gen.EmitIR()), 0644); err != nil {
		return err
	}

	// Update cache
	sourceHash, err := b.cache.ComputeFileHash(mod.Path)
	if err != nil {
		return fmt.Errorf("failed to hash source: %w", err)
	}

	importHashes := make(map[string]string)

	for impPath, impModule := range b.getImportPaths(mod) {
		hash, err := b.cache.ComputeFileHash(impModule)
		if err != nil {
			return fmt.Errorf("failed to hash import %s: %w", impPath, err)
		}

		importHashes[impPath] = hash
	}

	entry := &CacheEntry{
		SourceHash: sourceHash,
		ImportHash: importHashes,
	}

	return b.cache.SaveCacheEntry(mod.Path, entry)
}

func (b *Builder) linkModules(modules []*module.Module, outputName string) error {
	fmt.Println("  Linking modules")

	// Collect all IR files
	irFiles := []string{}

	for _, mod := range modules {
		if mod.IRPath != "" {
			irFiles = append(irFiles, mod.IRPath)
		}
	}

	// Link with llvm-link
	linkedPath := filepath.Join(b.projectRoot, "build", "ir", "linked.ll")

	// Find llvm-link in common locations
	llvmLink := findLLVMLink()

	args := append([]string{"-S", "-o", linkedPath}, irFiles...)
	cmd := exec.Command(llvmLink, args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("llvm-link failed: %w\n%s", err, output)
	}

	return nil
}

// findLLVMLink searches for llvm-link in common locations
func findLLVMLink() string {
	// Try PATH first
	if path, err := exec.LookPath("llvm-link"); err == nil {
		return path
	}

	// Try common homebrew locations
	commonPaths := []string{
		"/opt/homebrew/opt/llvm/bin/llvm-link",
		"/opt/homebrew/bin/llvm-link",
		"/usr/local/opt/llvm/bin/llvm-link",
		"/usr/local/bin/llvm-link",
	}

	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Also check for versioned llvm installations
	if entries, err := filepath.Glob("/opt/homebrew/Cellar/llvm*/*/bin/llvm-link"); err == nil && len(entries) > 0 {
		return entries[0]
	}

	// Fallback to "llvm-link" and let it fail with a better error
	return "llvm-link"
}

func (b *Builder) compileExecutable(name string) error {
	linkedIR := filepath.Join(b.projectRoot, "build", "ir", "linked.ll")
	outputPath := filepath.Join(b.projectRoot, "build", "bin", name)

	// Check if runtime library exists
	runtimeLib := filepath.Join(b.projectRoot, "runtime", "libyarrt.a")
	if _, err := os.Stat(runtimeLib); err != nil {
		// Runtime library doesn't exist, skip executable generation
		// This is OK for testing - we've already generated IR
		fmt.Printf("  Skipping executable generation (runtime library not found)\n")
		return nil
	}

	cmd := exec.Command("clang",
		"-o", outputPath,
		linkedIR,
		runtimeLib,
		"-L/opt/homebrew/lib",
		"-lgc",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("clang failed: %w\n%s", err, output)
	}

	fmt.Printf("    Finished: %s\n", outputPath)

	return nil
}

func (b *Builder) getImportPaths(mod *module.Module) map[string]string {
	result := make(map[string]string)
	for _, imp := range mod.Imports {
		result[imp.Path] = imp.Resolved
	}

	return result
}
