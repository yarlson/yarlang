package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "build":
		handleBuild(os.Args[2:])
	case "run":
		handleRun(os.Args[2:])
	case "check":
		handleCheck(os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("YarLang Compiler v0.1.0")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  yar build <file>    Compile YarLang source to executable")
	fmt.Println("  yar run <file>      Compile and run YarLang source")
	fmt.Println("  yar check <file>    Type-check without compiling")
}
