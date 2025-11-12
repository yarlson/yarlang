package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/yarlson/yarlang/codegen"
	"github.com/yarlson/yarlang/lexer"
	"github.com/yarlson/yarlang/parser"
	"github.com/yarlson/yarlang/semantic"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	var err error

	switch command {
	case "init":
		err = initCommand(args)
	case "build":
		err = buildCommand(args)
	case "run":
		err = runCommand(args)
	case "compile":
		err = compileCommand(args)
	case "clean":
		err = cleanCommand(args)
	case "check":
		err = checkCommand(args)
	default:
		// Legacy: if first arg is .yar file, compile it
		if filepath.Ext(command) == ".yar" {
			err = legacyCompile(command)
		} else {
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
			printUsage()
			os.Exit(1)
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("YarLang Compiler")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  yar init              Create a new project")
	fmt.Println("  yar build             Build the project")
	fmt.Println("  yar run               Build and run the project")
	fmt.Println("  yar compile <file>    Compile a single file")
	fmt.Println("  yar clean             Remove build artifacts")
	fmt.Println("  yar check             Check code without building")
	fmt.Println()
	fmt.Println("Legacy:")
	fmt.Println("  yar <file.yar>        Compile a single file")
}

func buildCommand(args []string) error {
	fmt.Println("Build command not yet implemented (requires Phase 5)")
	return nil
}

func runCommand(args []string) error {
	// Build first
	if err := buildCommand(args); err != nil {
		return err
	}

	// TODO: Run the executable
	fmt.Println("Running...")

	return nil
}

func compileCommand(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: yar compile <file.yar>")
	}

	return legacyCompile(args[0])
}

func cleanCommand(args []string) error {
	buildDir := "build"
	if err := os.RemoveAll(buildDir); err != nil {
		return err
	}

	fmt.Println("   Removing build/")

	return nil
}

func checkCommand(args []string) error {
	fmt.Println("   Checking project...")
	// TODO: Run parser and semantic analysis without codegen
	return nil
}

func legacyCompile(sourceFile string) error {
	// Original single-file compilation logic
	// Read source file
	source, err := os.ReadFile(sourceFile)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Lex
	l := lexer.New(string(source))

	// Parse
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Fprintf(os.Stderr, "Parser errors:\n")

		for _, err := range p.Errors() {
			fmt.Fprintf(os.Stderr, "  %s\n", err)
		}

		return fmt.Errorf("parser errors occurred")
	}

	// Semantic analysis
	analyzer := semantic.New()
	if err := analyzer.Analyze(program); err != nil {
		return fmt.Errorf("semantic error: %w", err)
	}

	// Code generation
	gen := codegen.New()
	if err := gen.Generate(program); err != nil {
		return fmt.Errorf("codegen error: %w", err)
	}

	// Emit LLVM IR to file
	irFile := strings.TrimSuffix(sourceFile, filepath.Ext(sourceFile)) + ".ll"
	if err := os.WriteFile(irFile, []byte(gen.EmitIR()), 0644); err != nil {
		return fmt.Errorf("error writing IR: %w", err)
	}

	fmt.Printf("Generated LLVM IR: %s\n", irFile)

	// Compile IR to executable
	outputFile := strings.TrimSuffix(sourceFile, filepath.Ext(sourceFile))
	cmd := exec.Command("clang",
		"-o", outputFile,
		irFile,
		"runtime/libyarrt.a",
		"-L/opt/homebrew/lib",
		"-lgc",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error compiling: %w\n%s", err, output)
	}

	fmt.Printf("Compiled executable: %s\n", outputFile)

	return nil
}
