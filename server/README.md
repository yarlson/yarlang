# YarLang LSP

Language Server Protocol implementation for YarLang - a dynamically-typed scripting language with Go-inspired syntax.

## Features

The YarLang LSP provides real-time language intelligence for YarLang development:

- **Diagnostics** - Real-time syntax and semantic error detection
  - Undefined variable warnings
  - Undefined function errors
  - Use-before-definition detection
  - Parser syntax errors

- **Code Completion** - Context-aware suggestions for:
  - Variables and functions in scope
  - Function parameters
  - Language keywords (func, if, else, for, return, break, continue, nil, true, false)

- **Hover Information** - Rich type information on hover
  - Variable types and values
  - Function signatures
  - Parameter information

- **Go-to-Definition** - Navigate to symbol declarations
  - Jump to variable declarations
  - Jump to function definitions
  - Works across function scopes

## Architecture

```
Editor (Zed/VSCode) → LSP Server → Parser → AST → Semantic Analyzer → LSP Features
```

The LSP is built on top of YarLang's existing lexer and parser, adding position tracking and semantic analysis capabilities. The architecture consists of:

- **Lexer** - Tokenizes YarLang source with position information
- **Parser** - Builds an Abstract Syntax Tree (AST) with position tracking
- **Semantic Analyzer** - Performs symbol resolution, scope analysis, and type inference
- **LSP Server** - Implements the Language Server Protocol to communicate with editors

## Project Structure

```
yarlang/
├── cmd/yarlang-lsp/      # LSP server binary
├── ast/                  # AST node definitions with positions
├── lexer/                # Tokenization with position tracking
├── parser/               # Syntax analysis
├── analysis/             # Semantic analysis (symbol tables, scopes)
├── server/               # LSP protocol handlers
├── editors/              # Editor configurations
│   └── zed/             # Zed editor setup
├── testdata/            # Test YarLang files
└── docs/plans/          # Design documents
```

## Installation

### Prerequisites

- Go 1.21 or later
- Zed editor (or any LSP-compatible editor)

### Build from Source

```bash
# Clone the repository
git clone https://github.com/yarlson/yarlang.git
cd yarlang

# Switch to LSP branch/worktree
cd .worktrees/lsp  # if using worktrees

# Install dependencies
go mod download

# Build LSP server
cd cmd/yarlang-lsp
go build -o yarlang-lsp

# Install to system path (optional)
sudo cp yarlang-lsp /usr/local/bin/
```

Alternatively, build from the project root:

```bash
# Build the LSP server binary
cd /Users/yar/home/yarlang/.worktrees/lsp
go build -o yarlang-lsp ./cmd/yarlang-lsp
```

### Configure Zed Editor

See [editors/zed/README.md](editors/zed/README.md) for detailed setup instructions.

Quick setup:

1. Install the LSP server binary to your PATH (see Build from Source above)

2. Copy language configuration:

   ```bash
   mkdir -p ~/.config/zed/languages/yarlang
   cp editors/zed/languages/yarlang/config.toml ~/.config/zed/languages/yarlang/
   ```

3. Add LSP configuration to `~/.config/zed/settings.json`:

   ```json
   {
     "lsp": {
       "yarlang-lsp": {
         "binary": {
           "path": "/usr/local/bin/yarlang-lsp",
           "arguments": []
         }
       }
     },
     "languages": {
       "YarLang": {
         "language_servers": ["yarlang-lsp"],
         "tab_size": 4,
         "hard_tabs": false
       }
     }
   }
   ```

4. Restart Zed

### Configure Other Editors

The YarLang LSP follows the standard Language Server Protocol specification and should work with any LSP-compatible editor:

- **VS Code**: Install using a custom extension or configure in `settings.json`
- **Neovim**: Configure using `nvim-lspconfig`
- **Emacs**: Configure using `lsp-mode` or `eglot`
- **Sublime Text**: Configure using LSP package

## Testing

### Run Unit Tests

```bash
# Test all packages
go test ./...

# Test specific packages
go test -v ./ast
go test -v ./analysis
go test -v ./server
go test -v ./parser
go test -v ./lexer

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -v -run TestAnalyzeUndefinedVariable ./analysis
```

### Manual Testing with Example Files

Open the provided test files in your editor to verify LSP functionality:

```bash
# Open test files in Zed
zed testdata/lsp_demo.yar
zed testdata/test.yar
```

The `testdata/lsp_demo.yar` file contains comprehensive examples demonstrating:

- Variable and function declarations
- Diagnostic errors (undefined variables, undefined functions)
- Hover information examples
- Go-to-definition test cases
- Code completion scenarios

## Development

### Project Setup

```bash
# Install dependencies
go mod download

# Verify all packages build
go build ./...
```

### Running Tests During Development

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with race detection
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Debugging the LSP Server

Enable LSP logging in Zed:

- Menu: View → Debug → Language Server Logs

Run the server manually with logging:

```bash
# Run server with stderr logging
./yarlang-lsp 2>lsp-debug.log

# Or run with explicit logging (if implemented)
./yarlang-lsp --log-file=lsp.log --verbose
```

Test the server with a specific file:

```bash
# Open a test file in Zed
zed testdata/lsp_demo.yar

# Check for errors in the LSP logs
```

### Code Organization

- **ast/** - AST node types with position information
  - Defines Position and Range types
  - All nodes implement the Node interface with GetRange()

- **lexer/** - Tokenization
  - Token types and definitions
  - Position-aware tokenization

- **parser/** - Syntax analysis
  - Recursive descent parser
  - Expression and statement parsing
  - Error recovery

- **analysis/** - Semantic analysis
  - Symbol table construction
  - Scope management
  - Type inference
  - Diagnostic generation

- **server/** - LSP protocol implementation
  - Document management
  - Protocol handlers (completion, hover, definition)
  - Diagnostic publishing

## YarLang Language Overview

YarLang is a dynamically-typed scripting language with familiar syntax:

```yar
// Variables - dynamically typed
x = 42
name = "Alice"
isActive = true
nothing = nil

// Functions
func add(a, b) {
    return a + b
}

func fibonacci(n) {
    if n <= 1 {
        return n
    }
    return fibonacci(n - 1) + fibonacci(n - 2)
}

// Control flow
if x > 0 {
    println("positive")
} else {
    println("non-positive")
}

// Loops
for i = 0; i < 10; i = i + 1 {
    println(i)
}

// Multiple return values
func divmod(a, b) {
    return a / b, a % b
}

quotient, remainder = divmod(10, 3)
```

### Language Features

- Dynamic typing
- First-class functions
- Multiple return values
- Go-like syntax
- Lexical scoping
- Control flow (if/else, for loops)
- Basic operators (+, -, \*, /, %, ==, !=, <, >, <=, >=, &&, ||, !)

See [docs/plans/2025-11-11-yarlang-design.md](docs/plans/2025-11-11-yarlang-design.md) for complete language specification.

## Troubleshooting

### LSP Server Not Starting

1. Verify the binary is in your PATH:

   ```bash
   which yarlang-lsp
   ```

2. Check file permissions:

   ```bash
   chmod +x /usr/local/bin/yarlang-lsp
   ```

3. Test the binary manually:
   ```bash
   /usr/local/bin/yarlang-lsp
   ```
   The server should start and wait for LSP messages on stdin.

### No Diagnostics Showing

1. Check that the file has `.yar` extension
2. Verify Zed recognizes the file as YarLang (check status bar)
3. Check LSP logs for errors: View → Debug → Language Server Logs
4. Ensure the language configuration is properly installed

### Code Completion Not Working

1. Verify the LSP server is running (check logs)
2. Try triggering completion manually (Ctrl+Space)
3. Check that you're in a valid context (after an identifier start)
4. Verify symbols are defined before the cursor position

### Go-to-Definition Not Working

1. Ensure the symbol is defined in the current file
2. Check that the definition appears before the usage (YarLang requires forward declaration)
3. Verify position tracking is working (hover should show information)

### Reporting Issues

When reporting issues, please include:

1. YarLang LSP version: `yarlang-lsp --version` (if implemented)
2. Editor and version (e.g., Zed 0.x.x)
3. Operating system
4. Sample YarLang code that reproduces the issue
5. LSP logs (View → Debug → Language Server Logs)

## Contributing

Contributions are welcome! To contribute:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes with tests
4. Run tests: `go test ./...`
5. Commit your changes (`git commit -m 'feat: add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Development Guidelines

- Write tests for all new features
- Follow Go conventions and formatting (`gofmt`)
- Update documentation for API changes
- Add examples for new language features
- Ensure all tests pass before submitting PR

## License

MIT License - see LICENSE file for details.

## References

- [LSP Specification](https://microsoft.github.io/language-server-protocol/) - Official Language Server Protocol documentation
- [YarLang Language Design](docs/plans/2025-11-11-yarlang-design.md) - Complete language specification
- [LSP Implementation Design](docs/plans/2025-11-11-lsp-design.md) - LSP architecture and design decisions
- [LSP Implementation Plan](docs/plans/2025-11-11-yarlang-lsp-implementation.md) - Step-by-step implementation guide

## Acknowledgments

Built with:

- [go-lsp](https://pkg.go.dev/go.lsp.dev/protocol) - Go LSP protocol implementation
- [LLVM Go bindings](https://tinygo.org/x/go-llvm) - For code generation

---

**Status**: Active development. The core LSP features (diagnostics, completion, hover, go-to-definition) are implemented and functional. Additional features and language enhancements are planned.
