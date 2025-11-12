package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func initCommand(args []string) error {
	// Get current directory name as default package name
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	defaultName := filepath.Base(cwd)

	// Interactive prompts
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Creating new YarLang project...")

	// Package name
	fmt.Printf("  Package name (%s): ", defaultName)

	name, _ := reader.ReadString('\n')

	name = strings.TrimSpace(name)
	if name == "" {
		name = defaultName
	}

	// Entry point
	fmt.Print("  Entry point (main.yar): ")

	entry, _ := reader.ReadString('\n')

	entry = strings.TrimSpace(entry)
	if entry == "" {
		entry = "main.yar"
	}

	// Create yar.toml
	yarToml := fmt.Sprintf(`[package]
name = "%s"
version = "0.1.0"
entry = "%s"
`, name, entry)

	if err := os.WriteFile("yar.toml", []byte(yarToml), 0644); err != nil {
		return fmt.Errorf("failed to create yar.toml: %w", err)
	}

	fmt.Println("Created yar.toml")

	// Create entry file if it doesn't exist
	if _, err := os.Stat(entry); os.IsNotExist(err) {
		mainContent := `func main() {
    println("Hello, YarLang!")
}
`
		if err := os.WriteFile(entry, []byte(mainContent), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", entry, err)
		}

		fmt.Printf("Created %s\n", entry)
	}

	// Create .gitignore
	gitignore := `# Build artifacts
build/

# IDE
.vscode/
.idea/

# OS
.DS_Store
`
	if _, err := os.Stat(".gitignore"); os.IsNotExist(err) {
		if err := os.WriteFile(".gitignore", []byte(gitignore), 0644); err != nil {
			return fmt.Errorf("failed to create .gitignore: %w", err)
		}

		fmt.Println("Created .gitignore")
	}

	return nil
}
