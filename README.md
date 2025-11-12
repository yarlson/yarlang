# YarLang

A dynamically-typed scripting language with Go-inspired syntax that compiles to native code via LLVM.

YarLang is a modern scripting language that combines the simplicity of dynamic typing with the performance of ahead-of-time compilation. It features Go-like syntax, automatic memory management via the Boehm garbage collector, and compiles directly to native executables through LLVM.

## Features

**Core Language Features:**

- Dynamic typing with runtime type checking
- First-class functions with support for recursion
- Lexical scoping with Python-style variable semantics
- Automatic memory management (Boehm GC)

**Data Types:**

- Numbers (double-precision floating-point)
- Strings
- Booleans (true/false)
- Nil
- Functions

**Operators:**

- Arithmetic: `+`, `-`, `*`, `/`, `%`
- Comparison: `==`, `!=`, `<`, `<=`, `>`, `>=`
- Logical: `&&`, `||`, `!`

**Control Flow:**

- If/else conditionals with else-if chains
- For loops with break and continue statements

**Functions:**

- User-defined functions with parameters
- Recursive function calls
- Return statements with values
- Multiple return values
- Built-in functions: `print`, `println`, `len`, `type`

## Module System

YarLang supports splitting code across multiple files using Go-style imports.

### Creating a Project

```bash
yar init
```

This creates a `yar.toml` file:

```toml
[package]
name = "myproject"
version = "0.1.0"
entry = "main.yar"
```

### Importing Modules

**Single import:**
```go
import "math"
```

**Import block:**
```go
import (
    "math"
    "strings" as str
    "io"
)
```

**Using imports:**
```go
import "math"

func main() {
    x := math.Sqrt(16)  // Qualified access
    println(x)
}
```

### Exports

Functions with capital letters are exported:

```go
// math.yar
func Sqrt(x) {     // Exported
    return x
}

func internal() {  // Private
    return 42
}
```

### Building Projects

```bash
yar build   # Build project
yar run     # Build and run
yar clean   # Remove build artifacts
yar check   # Check without building
```

### Module Resolution

Import paths are resolved in this order:

1. **Local files** - Current directory and parents up to `yar.toml`
2. **Standard library** - `~/.yar/stdlib/`
3. **Third-party** - `~/.yar/pkg/` (future)

Use `std/` prefix to force stdlib: `import "std/math"`

### Standard Library

Initial stdlib modules:

- `math` - Mathematical functions
- `strings` - String utilities
- `io` - Input/output
- `os` - Operating system interface

### Project Structure

```
myproject/
  yar.toml         # Project manifest
  main.yar         # Entry point
  math.yar         # Local module
  utils/           # Nested modules
    strings.yar
  build/           # Generated (gitignored)
```

## Prerequisites

To build and run YarLang, you need:

- **Go** (1.21 or later) - for building the compiler and LSP
- **LLVM** (14.0 or later) - for code generation
- **Boehm GC** - for garbage collection
- **Clang** - for compiling LLVM IR to native code

### Installing Prerequisites

**macOS (using Homebrew):**

```bash
brew install go llvm bdw-gc
```

**Linux (Ubuntu/Debian):**

```bash
sudo apt-get install golang llvm-14 libgc-dev clang
```

**Linux (Fedora/RHEL):**

```bash
sudo dnf install golang llvm gc-devel clang
```

## Building YarLang

1. Clone or download this repository
2. Build the runtime library:

```bash
cd runtime
make
cd ..
```

3. Build the compiler and LSP server:

```bash
# Build the YarLang compiler
go build -o yar ./cmd/yarlang

# Build the LSP server
go build -o yarlang-lsp ./cmd/yarlang-lsp
```

This creates `yar` and `yarlang-lsp` executables in the current directory.

## Quick Start

Create a file called `hello.yar`:

```yar
// My first YarLang program
name = "World"
println("Hello, " + name + "!")

// Calculate and print some numbers
for i = 1; i <= 5; i = i + 1 {
    println(i * i)
}
```

Compile and run it:

```bash
./yar hello.yar
./hello
```

The compiler will generate `hello.ll` (LLVM IR) and `hello` (native executable).

## Usage

```bash
# Compile a YarLang program
./yar <source-file.yar>

# Start the LSP server (for editor integration)
./yarlang-lsp
```

The compiler will:

1. Lex and parse the source file
2. Perform semantic analysis
3. Generate LLVM IR (saved as `<source-file>.ll`)
4. Compile to a native executable (saved as `<source-file>`)

Then you can run the generated executable:

```bash
./<source-file>
```

## Language Server Protocol (LSP)

YarLang includes a full-featured LSP server that provides:

- **Real-time diagnostics** - Syntax and semantic error detection
- **Code completion** - Context-aware suggestions for variables, functions, and keywords
- **Hover information** - Type and signature information on hover
- **Go-to-definition** - Navigate to symbol declarations

### Editor Integration

#### Zed Editor

1. Install the LSP binary:

   ```bash
   sudo cp yarlang-lsp /usr/local/bin/
   sudo chmod +x /usr/local/bin/yarlang-lsp
   ```

2. Copy the language configuration:

   ```bash
   mkdir -p ~/.config/zed/languages/yarlang
   cp editors/zed/languages/yarlang/config.toml ~/.config/zed/languages/yarlang/
   ```

3. Add to `~/.config/zed/settings.json`:

   ```json
   {
     "lsp": {
       "yarlang-lsp": {
         "binary": {
           "path": "/usr/local/bin/yarlang-lsp"
         },
         "settings": {}
       }
     },
     "languages": {
       "YarLang": {
         "language_servers": ["yarlang-lsp"]
       }
     }
   }
   ```

4. Restart Zed and open a `.yar` file

See [`server/README.md`](server/README.md) for complete LSP documentation, troubleshooting, and support for other editors (VS Code, Neovim, Emacs, Sublime Text).

## Project Structure

```
yarlang/
├── cmd/
│   ├── yarlang/      # YarLang compiler binary
│   └── yarlang-lsp/  # LSP server binary
├── lexer/            # Tokenization and lexical analysis
├── parser/           # Parser and AST construction
├── ast/              # Abstract Syntax Tree definitions with position tracking
├── semantic/         # Semantic analysis (type checking, etc.)
├── analysis/         # LSP semantic analysis (symbols, scopes)
├── codegen/          # LLVM code generation
├── server/           # LSP server implementation
├── runtime/          # C runtime library (value.h, value.c)
├── editors/          # Editor configurations (Zed, etc.)
├── examples/         # Example YarLang programs
├── testdata/         # Test YarLang files for LSP
└── docs/             # Design documents and specifications
```

## Example Programs

The `examples/` directory contains several programs demonstrating YarLang features:

- `hello.yar` - Simple variable and print statement
- `arithmetic.yar` - Basic arithmetic operations
- `conditionals.yar` - If/else statements and comparisons
- `loops.yar` - For loops with break and continue
- `functions.yar` - Function definitions, recursion, and parameters
- `fibonacci.yar` - Recursive Fibonacci implementation

### Example Code

```yar
// Variables
x = 42
name = "YarLang"
active = true

// Functions
func add(a, b) {
    return a + b
}

result = add(10, 20)

// Control flow
if result > 25 {
    x = 100
} else {
    x = 0
}

// Loops
for i = 0; i < 10; i = i + 1 {
    result = result + i
}

// Multiple return values
func divmod(a, b) {
    return a / b, a % b
}

quotient, remainder = divmod(17, 5)
```

## Testing

Run the test suite:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./lexer
go test ./parser
go test ./analysis
go test ./server
go test ./codegen
```

Test the example programs:

```bash
# Compile and test each example
for file in examples/*.yar; do
    echo "Testing $file"
    ./yar "$file"
done
```

## Implementation Details

**Compilation Pipeline:**

1. **Lexer** - Tokenizes source code into tokens
2. **Parser** - Builds abstract syntax tree (AST)
3. **Semantic Analyzer** - Resolves scopes and checks for errors
4. **Code Generator** - Emits LLVM IR
5. **LLVM/Clang** - Compiles IR to native executable

**Runtime:**

- Written in C for performance
- Uses Boehm GC for automatic memory management
- Dynamic typing with tagged Value\* pointers
- All values are heap-allocated and GC-managed

**Memory Management:**

- All YarLang values are represented as `Value*` pointers
- Memory is automatically managed by the Boehm GC
- No manual memory management required

## Documentation

- **Language Guide**: [docs/language-guide.md](docs/language-guide.md) - Comprehensive syntax reference
- **Examples Guide**: [docs/examples.md](docs/examples.md) - Annotated example programs
- **LSP Server**: [server/README.md](server/README.md) - Complete LSP documentation

## Current Status

**Active Development** - YarLang is under active development with the following features implemented:

- ✅ Full lexer and parser with comprehensive tests
- ✅ Semantic analysis with scope resolution
- ✅ LLVM code generation to native executables
- ✅ C runtime with dynamic typing and garbage collection
- ✅ All core language features (functions, loops, conditionals, etc.)
- ✅ LSP server with diagnostics, completion, hover, and go-to-definition
- ✅ Editor integration (Zed)

## Future Enhancements

Potential future features:

- Arrays and maps
- Structs/objects
- String interpolation
- Standard library (file I/O, networking, etc.)
- REPL (read-eval-print loop)
- Better error messages with source location
- Optimization passes
- Additional editor support (VS Code extension)

## Contributing

Contributions are welcome! Please ensure:

- All tests pass (`go test ./...`)
- Code follows Go conventions (`go fmt`, `go vet`)
- New features include tests
- Commits follow conventional commit format

## License

MIT

## Resources

- [LLVM Documentation](https://llvm.org/docs/)
- [Language Server Protocol Specification](https://microsoft.github.io/language-server-protocol/)
- [Boehm GC](https://www.hboehm.info/gc/)
- [Go Programming Language](https://golang.org/)

---

Created as a learning project to explore compiler design, language server protocols, LLVM code generation, and language implementation.
