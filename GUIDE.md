# Getting Started with YarLang

Welcome to **YarLang**, a statically typed systems language that borrows Go's approachable surface syntax and Rust-inspired ownership rules. This guide walks you from "what even is YarLang?" all the way to structuring real programs, using the `yar` toolchain, and understanding the language features currently implemented.

---

## 1. Toolchain and Project Setup

### 1.1 Prerequisites

| Requirement          | Why                                         | Installation (macOS/Homebrew) |
| -------------------- | ------------------------------------------- | ----------------------------- |
| Go 1.21+             | Build the YarLang compiler (written in Go). | `brew install go`             |
| LLVM 14+ & Clang     | Emit and link LLVM IR.                      | `brew install llvm`           |
| A POSIX shell & make | Run the examples/scripts in this guide.     | Comes with macOS/Linux        |

> On Linux use your distro packages (`apt`, `dnf`, etc.) to install Go, LLVM, and Clang.

### 1.2 Build the `yar` CLI

From the repository root (`/Users/yar/home/yarlang/yarlang`):

```bash
cd cmd/yar
go build -o ../../yar
cd ../..
```

This leaves an executable named `yar` in the project root. Confirm it works:

```bash
./yar help   # shows usage information
```

### 1.3 Standard Layout

```
yarlang/
├─ yar                  # CLI you just built
├─ examples/            # Sample *.yar programs
├─ runtime/runtime.c    # Minimal C runtime (println, panic)
├─ cmd/                 # CLI entry point (Go)
├─ lexer/, parser/, ... # Compiler stages (Go)
└─ docs/                # Specs and planning docs
```

Most user work happens in standalone `.yar` files that you pass to `./yar build` or `./yar run`.

---

## 2. The `yar` Workflow

| Command                | Description                                                                                                         |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------- |
| `./yar build file.yar` | Type checks, lowers to LLVM IR, invokes `clang`, and produces an executable next to `file.yar` (without extension). |
| `./yar run file.yar`   | Equivalent to `build` followed by executing the produced binary.                                                    |
| `./yar check file.yar` | Runs parser + type checker only. Great for fast iteration.                                                          |

Each build also saves an `.ll` file (LLVM IR) alongside the source, which is useful when debugging codegen issues.

### 2.1 Your First Program

Create `hello.yar`:

```
fn main() {
    println("Hello, YarLang!")
}
```

Build and run:

```bash
./yar run hello.yar
```

Output:

```
Hello, YarLang!
```

---

## 3. Language Fundamentals

YarLang code is organized into `fn` functions. Execution starts in `fn main()`.

### 3.1 Types at a Glance

Primitives (subset implemented today):

- Signed integers: `i8`, `i16`, `i32`, `i64`, `isize`
- Unsigned integers: `u8`, `u16`, `u32`, `u64`, `usize`
- Floats: `f32`, `f64`
- Boolean: `bool`
- Character: `char`
- Void / unit: `void`

Composite/derived types currently parsed and partially checked:

- References: `&T` (shared), `&mut T` (exclusive)
- Raw pointers: `*T`
- Arrays `[T; N]` and slices `[]T`
- Tuples `(T1, T2, ...)`

Strings are currently represented as `[]u8` (UTF-8 byte slices). String literals in source (e.g. `"hello"`) automatically lower to global byte arrays plus a pointer.

### 3.2 Variables and Bindings

- `let name: Type = expr` — immutable binding with explicit type.
- `let mut name: Type = expr` — mutable binding (mutation support is under construction; reassignments on immutable bindings will currently error in the checker).
- `name := expr` — short immutable binding with type inference.

Examples:

```
fn main() {
    let answer: i32 = 42
    let name := "Yar"
    let mut counter: i32 = 0
    counter = counter + 1
}
```

> Tip: Because values move by default, assigning `let y = x` transfers ownership. Borrow with `&x` if you need to keep using `x`.

### 3.3 Expressions

All the usual arithmetic and comparison operators exist: `+`, `-`, `*`, `/`, `%`, `==`, `!=`, `<`, `<=`, `>`, `>=`, `&&`, `||`.

Example:

```
let delta := (high - low) / 2
let is_small := delta < 10 && delta > 0
```

### 3.4 Control Flow

#### `if / else`

```
fn describe(n i32) {
    if n < 0 {
        println("negative")
    } else if n == 0 {
        println("zero")
    } else {
        println("positive")
    }
}
```

#### `while`

```
fn countdown(start i32) {
    let mut current := start
    while current >= 0 {
        println(current)
        current = current - 1
    }
    println("Lift off!")
}
```

#### Range-based `for`

The currently implemented `for` form iterates over a numeric range (`start..end`) and binds each value to a loop variable:

```
fn sum_first_ten() i32 {
    let mut total: i32 = 0
    for i in 0..10 {
        total = total + i
    }
    return total
}
```

`break` and `continue` are available inside loops.

### 3.5 Functions and Recursion

Function syntax mirrors Go: `fn name(params) return_type { ... }`. Return type is optional (defaults to `void`). Values move by default unless borrowed.

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

This example exercises recursion, integer arithmetic, and both string + integer `println` calls. Build it via `./yar run examples/fibonacci.yar`.

### 3.6 Ownership, Moves, and Borrows

YarLang aims for Rust-like ownership but is still stabilizing. Today’s rules to remember:

- Bindings move on assignment: once you bind `let y = x`, you cannot use `x` unless you borrowed it (`&x`) beforehand.
- Shared borrows (`&x`) let you read without taking ownership; mutable borrows (`&mut x`) grant exclusive write access until the borrow ends.
- The checker tracks simple borrow states and flags invalid mixes (e.g. mutable + shared at the same time).

```
fn main() {
    let a: i32 = 10
    let b: &i32 = &a   // borrow
    println(b)         // OK because println accepts references to arrays/strings and integers directly
    // println(a)      // still OK: `a` was borrowed immutably
}
```

### 3.7 Error Handling with `panic`

For now there is no `Result` or `Option` runtime. Use `panic("message")` to abort execution via the runtime:

```
fn checked_div(a i32, b i32) i32 {
    if b == 0 {
        panic("division by zero")
    }
    return a / b
}
```

### 3.8 Built-in Functions

| Function          | Signature                                       | Notes                                                                           |
| ----------------- | ----------------------------------------------- | ------------------------------------------------------------------------------- |
| `println(value)`  | Overloaded for `[]u8` (strings), `i32`, `bool`. | Lowered to runtime helpers embedded in the CLI. More types will arrive later.   |
| `panic(msg []u8)` | Immediately terminates the program.             | Provided by the C runtime.                                                      |
| `len([]T) usize`  | Length of a byte slice (strings only for now).  | Type inference treats it generically but runtime currently handles byte slices. |

Example mixing `len` and string literals:

```
fn report(name []u8) {
    println("name length:")
    println(len(name))
}
```

---

## 4. Working with Strings and Printing

As mentioned, string literals become byte slices stored in read-only global memory. When you pass them to `println`, the compiler automatically emits the LLVM `getelementptr` dance to produce an `i8*`. For dynamic data:

```
fn greet(name []u8) {
    println("Hello, ")
    println(name)
}
```

Printing numbers or booleans uses the specialized runtime shims we added (`println_i32`, `println_bool`). If you need to print other types (e.g., custom structs), convert them manually for now.

---

## 5. Putting It All Together: Mini Project

Below is a simple program that exercises user input simulation, branching, loops, and helper functions.

```
fn clamp(value i32, min i32, max i32) i32 {
    if value < min {
        return min
    }
    if value > max {
        return max
    }
    return value
}

fn main() {
    let mut readings: i32 = 0
    let mut total: i32 = 0

    for i in 0..5 {
        let sample := clamp(i * 7 - 10, 0, 30)
        total = total + sample
        readings = readings + 1
    }

    println("processed readings:")
    println(readings)

    println("average:")
    println(total / readings)
}
```

Steps to try it out:

1. Save as `examples/sensors.yar`.
2. `./yar build examples/sensors.yar`
3. Run the produced `examples/sensors` executable.

Inspect the generated LLVM IR (`examples/sensors.ll`) if you’re curious how loops and helper calls look after lowering.

---

## 6. Troubleshooting & Tips

- **`load i32, <nil>` in LLVM IR** — indicates a bug where the compiler attempted to load from an unallocated pointer. Version `v0.4` fixes this for function parameters by spilling them to the stack automatically.
- **`panic: use of moved value`** — the checker caught ownership misuse. Borrow (`&value`) instead of moving, or restructure your code.
- **Inspect IR** when diagnosing codegen issues: search for `%t` temporaries to see how MIR instructions turned into LLVM instructions.

---

## 7. Roadmap Snapshot

The rewrite plan in `docs/plans/2025-11-12-yarlang-v04-rewrite.md` outlines upcoming features such as trait-based generics, richer borrow checking, and a standard library with `Result`, `Option`, and `Vec`. Keep an eye on the `docs/plans/` directory for progress updates.

---

## 8. Next Steps

1. Try modifying `examples/fibonacci.yar` to memoize results using a slice.
2. Explore `tree-sitter-yarlang/test/corpus` for more language samples recognized by the parser.
3. Contribute documentation or tests—there’s plenty of surface area as the v0.4 rewrite evolves.

Happy hacking in YarLang!
