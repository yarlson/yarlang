package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/yarlson/yarlang/checker"
	"github.com/yarlson/yarlang/codegen"
	"github.com/yarlson/yarlang/lexer"
	"github.com/yarlson/yarlang/mir"
	"github.com/yarlson/yarlang/parser"
)

func handleBuild(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: no input file specified")
		os.Exit(1)
	}

	inputFile := args[0]
	outputFile := strings.TrimSuffix(inputFile, filepath.Ext(inputFile))

	// Read source
	source, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Lex
	l := lexer.New(string(source))
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		fmt.Println("Parser errors:")

		for _, err := range p.Errors() {
			fmt.Printf("  %s\n", err)
		}

		os.Exit(1)
	}

	// Type check
	c := checker.NewChecker()
	if err := c.CheckFile(file); err != nil {
		fmt.Printf("Type error: %v\n", err)
		os.Exit(1)
	}

	// Lower to MIR
	lower := mir.NewLowerer()
	mirMod := lower.LowerFile(file)

	// Generate LLVM IR
	cg := codegen.NewCodegen()
	llvmMod := cg.GenModule(mirMod)

	// Write LLVM IR to file
	llFile := outputFile + ".ll"
	if err := os.WriteFile(llFile, []byte(llvmMod.String()), 0644); err != nil {
		fmt.Printf("Error writing LLVM IR: %v\n", err)
		os.Exit(1)
	}

	// Get runtime path (relative to executable or source)
	runtimePath := "runtime/runtime.c"

	// Compile with clang, linking the runtime
	cmd := exec.Command("clang", "-O2", llFile, runtimePath, "-o", outputFile)
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("Error compiling: %v\n%s\n", err, output)
		os.Exit(1)
	}

	fmt.Printf("Built: %s\n", outputFile)
}

func handleRun(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: no input file specified")
		os.Exit(1)
	}

	// Build first
	handleBuild(args)

	// Run the executable
	inputFile := args[0]
	execFile := strings.TrimSuffix(inputFile, filepath.Ext(inputFile))

	cmd := exec.Command("./" + execFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running: %v\n", err)
		os.Exit(1)
	}
}

func handleCheck(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: no input file specified")
		os.Exit(1)
	}

	inputFile := args[0]

	// Read source
	source, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Lex and parse
	l := lexer.New(string(source))
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		fmt.Println("Parser errors:")

		for _, err := range p.Errors() {
			fmt.Printf("  %s\n", err)
		}

		os.Exit(1)
	}

	// Type check
	c := checker.NewChecker()
	if err := c.CheckFile(file); err != nil {
		fmt.Printf("Type error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ %s type-checks successfully\n", inputFile)
}
