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
		fmt.Fprintf(os.Stderr, "Usage: %s <source.yar>\n", os.Args[0])
		os.Exit(1)
	}

	sourceFile := os.Args[1]

	// Read source file
	source, err := os.ReadFile(sourceFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
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

		os.Exit(1)
	}

	// Semantic analysis
	analyzer := semantic.New()
	if err := analyzer.Analyze(program); err != nil {
		fmt.Fprintf(os.Stderr, "Semantic error: %v\n", err)
		os.Exit(1)
	}

	// Code generation
	gen := codegen.New()
	if err := gen.Generate(program); err != nil {
		fmt.Fprintf(os.Stderr, "Codegen error: %v\n", err)
		os.Exit(1)
	}

	// Emit LLVM IR to file
	irFile := strings.TrimSuffix(sourceFile, filepath.Ext(sourceFile)) + ".ll"
	if err := os.WriteFile(irFile, []byte(gen.EmitIR()), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing IR: %v\n", err)
		os.Exit(1)
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
		fmt.Fprintf(os.Stderr, "Error compiling: %v\n%s\n", err, output)
		os.Exit(1)
	}

	fmt.Printf("Compiled executable: %s\n", outputFile)
}
