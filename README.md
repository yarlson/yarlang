# YarLang v0.1.0

A statically-typed systems programming language with ownership and borrowing, combining Go's clean syntax with Rust's memory safety guarantees.

> New to YarLang? Start with the hands-on [GUIDE.md](GUIDE.md) for a step-by-step walkthrough from installation to building small programs.

## Design Philosophy

**"Go-look, Rust-core"** - Simple, familiar surface syntax with real ownership semantics, move-by-default, and compile-time memory safety without garbage collection.

## Features

### Type System

- **Static typing** with Hindley-Milner type inference
- **Ownership and borrowing** - values move by default, compile-time borrow checking
- **Primitives**: `i8`, `i16`, `i32`, `i64`, `isize`, `u8`, `u16`, `u32`, `u64`, `usize`, `f32`, `f64`, `bool`, `char`, `void`
- **References**: `&T` (shared), `&mut T` (exclusive)
- **Generics**: `Vec<T>`, `Result<T, E>`, `Option<T>`
- **Arrays**: `[T; N]` (fixed size)
- **Slices**: `[]T` (borrowed view)
- **Tuples**: `(T1, T2, ...)`

### Language Features

- **Result<T,E>** - Explicit error handling with `?` operator for propagation
- **Option<T>** - Null safety with Some/None variants
- **defer** - RAII-style cleanup (LIFO execution on scope exit)
- **Semicolons optional** - Automatic semicolon insertion (ASI)
- **No garbage collection** - Compile-time memory management via ownership

### Syntax Style

- **Go-style function parameters**: `fn add(a i32, b i32) i32`
- **Rust-style bindings**: `let x: i32 = 5`, `let mut y = 10`
- **Short declarations**: `x := 42` (type inferred, immutable)

## Quick Start

### Hello World

Create `hello.yar`:

```
fn main() {
    println("Hello, YarLang!")
}
```

Build and run:

```bash
./yar build hello.yar
./hello
```

Output:

```
Hello, YarLang!
```

### More Examples

**Fibonacci:**

```
fn fib(n i32) i32 {
    if n <= 1 {
        return n
    }
    return fib(n - 1) + fib(n - 2)
}

fn main() {
    let result: i32 = fib(10)
    println("fib(10) = ")
    println(result)
}
```

**Ownership and Borrowing:**

```
fn main() {
    let x: i32 = 42
    let y: i32 = x        // Error: x moved to y

    let a: i32 = 10
    let b: &i32 = &a      // OK: borrow a
    let c: i32 = a        // Error: a is borrowed
}
```

**Error Handling:**

```
fn divide(a i32, b i32) Result<i32, []u8> {
    if b == 0 {
        return Err("division by zero")
    }
    return Ok(a / b)
}

fn main() {
    let result := divide(10, 2)?  // Propagates error or unwraps
    println("Result: 5")
}
```

## Installation

### Prerequisites

- **Go** 1.21+ (for building the compiler)
- **LLVM** 14.0+ (for code generation)
- **Clang** (for linking LLVM IR to native code)

### Installing Dependencies

**macOS:**

```bash
brew install go llvm
```

**Ubuntu/Debian:**

```bash
sudo apt-get install golang llvm-14 clang
```

**Fedora/RHEL:**

```bash
sudo dnf install golang llvm clang
```

### Building YarLang

```bash
# Clone the repository
git clone https://github.com/yarlson/yarlang.git
cd yarlang

# Build the compiler
cd cmd/yar
go build -o ../../yar
cd ../..
```

This creates the `yar` executable in the project root.

## Usage

```bash
# Type check a program
./yar check <file.yar>

# Build an executable
./yar build <file.yar>

# Build and run
./yar run <file.yar>
```

## Language Guide

### Variables and Types

```
// Immutable by default
let x: i32 = 42
let y := 100          // Type inferred

// Mutable variables
let mut counter: i32 = 0
counter = counter + 1

// Short declaration (immutable)
name := "YarLang"
```

### Functions

```
// Go-style syntax
fn add(a i32, b i32) i32 {
    return a + b
}

// Void functions (no return type)
fn greet(name []u8) {
    println("Hello, " + name)
}

// With generics (future)
fn identity<T>(x T) T {
    return x
}
```

### Control Flow

```
// If expressions
if x > 0 {
    println("positive")
} else if x < 0 {
    println("negative")
} else {
    println("zero")
}

// While loops
let mut i := 0
while i < 10 {
    println("tick")
    i = i + 1
}

// For loops (range-based)
for i in 0..10 {
    println("iteration")
}

// Break and continue
while true {
    if condition {
        break
    }
    if other {
        continue
    }
}
```

### Structs and Enums

```
// Struct definition
struct Point {
    x: f64,
    y: f64,
}

// Enum definition
enum Result<T, E> {
    Ok(T),
    Err(E),
}

// Implementation blocks
impl Point {
    fn len(&self) f64 {
        return (self.x * self.x + self.y * self.y).sqrt()
    }
}
```

### Error Handling

```
// Result type
fn parse_int(s []u8) Result<i32, []u8> {
    // ... implementation
    return Ok(42)
}

// Using ? operator
fn process() Result<i32, []u8> {
    let value := parse_int("123")?  // Propagates error
    return Ok(value * 2)
}
```

### Ownership and Borrowing

```
// Move semantics (default)
let x := vec![1, 2, 3]
let y := x      // x moved to y, x is invalid

// Borrowing (shared)
let a := 42
let b: &i32 = &a    // Borrow a

// Mutable borrowing (exclusive)
let mut x := 10
let y: &mut i32 = &mut x
*y = 20
```

### Defer

```
fn process_file(path []u8) Result<(), []u8> {
    let f := File::open(path)?
    defer f.close()  // Runs on scope exit (LIFO)

    // Use file...
    return Ok(())
}
```

### Modules

```
// Declare module
module myproject::utils

// Import modules
use std::io::File
use core::fmt::println

// Public items
pub fn exported() {
    // ...
}

fn private() {
    // ...
}
```

## Project Structure

```
yarlang/
├── cmd/yar/          # Compiler CLI
├── lexer/            # Tokenization
├── parser/           # Syntax analysis
├── ast/              # Abstract syntax tree
├── types/            # Type system
├── checker/          # Type checking and borrow checking
├── mir/              # Mid-level IR (SSA-based)
├── codegen/          # LLVM code generation
├── runtime/          # Minimal C runtime (println, panic)
├── stdlib/           # Standard library (Result, Option)
├── examples/         # Example programs
├── tests/            # Integration tests
└── docs/             # Language specification
```

## Compilation Pipeline

```
Source (.yar)
    ↓
Lexer → Tokens
    ↓
Parser → AST
    ↓
Type Checker → Typed AST
    ↓
MIR Lowerer → SSA IR
    ↓
LLVM Codegen → LLVM IR (.ll)
    ↓
Clang → Native Executable
```

## Testing

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./lexer -v
go test ./parser -v
go test ./checker -v
go test ./mir -v
go test ./codegen -v

# Run integration tests
go test ./tests -v

# Run test script
./test_all.sh
```

## Implementation Status

### ✅ Completed

- Lexer with full v0.1.0 token support
- Parser with ASI (automatic semicolon insertion)
- Type system with generics
- Type inference (Hindley-Milner basics)
- Ownership and move semantics
- Borrow checking (basic)
- MIR lowering (SSA-style IR)
- LLVM code generation
- Control flow (if/while/for/break/continue)
- Function calls with type checking
- String constants
- Defer statements (MIR-level)
- ? operator (MIR-level)
- Result<T,E> and Option<T> types
- C runtime integration

### 🚧 In Progress

- Full borrow checker (lifetimes)
- Pattern matching
- Trait system
- Complete standard library

### 📋 Planned

- Slice operations
- Method calls and field access
- Array literals and indexing
- Runtime defer support
- Runtime Result/Option support
- Error messages with source locations
- Optimization passes

## Documentation

- **[GUIDE.md](GUIDE.md)** - Beginner-friendly setup + language tutorial
- **[Language Specification](docs/yarlang-v0.1.0.md)** - Complete v0.1.0 spec
- **[Implementation Plan](docs/plans/)** - Phased development roadmap
- **[Examples](examples/)** - Sample programs

## Built-in Functions

```
fn println(msg: []u8) -> void    // Print string/byte slice with newline
fn println(value: i32) -> void   // Print integers
fn println(value: bool) -> void  // Print booleans
fn panic(msg: []u8) -> void      // Panic with message
fn len<T>(xs: []T) -> usize      // Length of slice (strings today)
```

## Current Limitations (v0.1.0)

- No pattern matching (deferred to v0.2.0)
- No async/await (library-based concurrency)
- No macros
- Simplified defer (no full runtime stack)
- Simplified ? operator (no full Result runtime)

## Contributing

We welcome contributions! Areas that need help:

1. Standard library implementation
2. Error message improvements
3. Documentation and examples
4. Test coverage
5. Performance optimizations

Please ensure:

- All tests pass (`go test ./...`)
- Code follows Go conventions (`go fmt`, `golangci-lint`)
- Commits use conventional commit format

## License

MIT License - see LICENSE file for details

## Resources

- [YarLang v0.1.0 Specification](docs/yarlang-v0.1.0.md)
- [LLVM Documentation](https://llvm.org/docs/)
- [Go Programming Language](https://golang.org/)
- [The Rust Programming Language](https://www.rust-lang.org/)

---

**Status**: Active development - v0.1.0 core features complete, runtime and stdlib in progress.

Created as a learning project exploring compiler design, type systems, ownership semantics, and LLVM code generation.
