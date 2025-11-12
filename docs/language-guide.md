# YarLang Language Guide

Complete reference for the YarLang programming language syntax and semantics.

## Table of Contents

1. [Introduction](#introduction)
2. [Basic Syntax](#basic-syntax)
3. [Data Types](#data-types)
4. [Variables](#variables)
5. [Operators](#operators)
6. [Control Flow](#control-flow)
7. [Loops](#loops)
8. [Functions](#functions)
9. [Built-in Functions](#built-in-functions)
10. [Examples](#examples)

## Introduction

### What is YarLang?

YarLang is a dynamically-typed scripting language with Go-inspired syntax that compiles to native code via LLVM. It combines the ease of scripting languages with the performance of compiled languages.

### Design Philosophy

**Dynamic Typing:** Variables can hold values of any type, and types are checked at runtime. This provides flexibility and rapid development.

**Go-Style Syntax:** YarLang uses Go-like syntax for familiarity, including:

- C-style operators
- Curly braces for blocks
- Simple function declaration syntax

**Compiled to Native:** Unlike interpreted scripting languages, YarLang compiles to native executables through LLVM, providing fast execution speed.

**Automatic Memory Management:** The Boehm garbage collector handles memory allocation and deallocation automatically.

## Basic Syntax

### Comments

Single-line comments start with `//`:

```yar
// This is a comment
x = 42  // This is also a comment
```

Multi-line comments are not currently supported.

### Statements and Expressions

YarLang distinguishes between statements and expressions:

**Statements** perform actions:

- Variable assignments: `x = 10`
- Function definitions: `func greet() { }`
- Control flow: `if`, `for`, `return`, `break`, `continue`

**Expressions** produce values:

- Literals: `42`, `"hello"`, `true`
- Variables: `x`
- Operators: `a + b`, `x > 5`
- Function calls: `fibonacci(10)`

### Whitespace

Whitespace (spaces, tabs, newlines) is used to separate tokens but is otherwise ignored. YarLang does not use significant indentation.

```yar
// These are equivalent:
x = 10
y = 20

x=10 y=20

x = 10 y = 20
```

However, proper indentation is recommended for readability.

## Data Types

YarLang has five core data types:

### Numbers

All numbers are double-precision floating-point (64-bit):

```yar
x = 42
y = 3.14159
z = -17.5
large = 1000000.0
```

Integers are represented as floating-point values internally. Mathematical operations may produce floating-point results:

```yar
a = 7 / 2   // Result is 3.5, not 3
b = 10 / 5  // Result is 2.0
```

### Strings

Strings are sequences of characters enclosed in double quotes:

```yar
name = "Alice"
message = "Hello, World!"
empty = ""
```

**String Concatenation:**

The `+` operator concatenates strings:

```yar
first = "Hello"
last = "World"
greeting = first + ", " + last + "!"  // "Hello, World!"
```

**Escape Sequences:**

Standard escape sequences are supported:

- `\n` - newline
- `\t` - tab
- `\\` - backslash
- `\"` - double quote

```yar
message = "Line 1\nLine 2"
path = "C:\\Users\\Name"
quote = "She said \"Hello\""
```

### Booleans

Boolean values represent true or false:

```yar
isReady = true
isDone = false
```

**Truthiness:**

In conditional contexts, values are evaluated for truthiness:

- `false` and `nil` are falsy
- All other values are truthy (including 0, "", etc.)

```yar
if true {
    println("This runs")
}

if false {
    println("This doesn't run")
}
```

### Nil

`nil` represents the absence of a value:

```yar
x = nil
```

### Functions

Functions are first-class values that can be defined and called:

```yar
func greet(name) {
    println("Hello, " + name)
}
```

See [Functions](#functions) for details.

## Variables

### Assignment

Variables are created through assignment:

```yar
x = 10
name = "Alice"
isValid = true
```

No declaration keyword is needed. Variables are dynamically typed and can hold any value type.

### Reassignment

Variables can be reassigned to different values (even of different types):

```yar
x = 10        // x is a number
x = "hello"   // now x is a string
x = true      // now x is a boolean
```

### Scoping Rules

YarLang uses **lexical scoping** with Python-style variable semantics:

**Global Scope:**

Variables assigned at the top level are global:

```yar
x = 10  // Global variable

func test() {
    println(x)  // Can access global x
}
```

**Local Scope:**

The first assignment to a variable in a function creates a local variable:

```yar
x = 10  // Global

func test() {
    x = 20  // Creates LOCAL x, doesn't modify global
    println(x)  // Prints 20
}

test()
println(x)  // Prints 10 (global unchanged)
```

**Function Parameters:**

Function parameters are always local to the function:

```yar
func add(a, b) {
    result = a + b  // Local variable
    return result
}
```

**Block Scope:**

Variables in control flow blocks follow the same rules:

```yar
x = 10

if true {
    x = 20     // Modifies outer x
    y = 30     // Creates new local y
}

println(x)  // 20
println(y)  // Error: y not defined outside block
```

## Operators

### Arithmetic Operators

| Operator | Description      | Example | Result |
| -------- | ---------------- | ------- | ------ |
| `+`      | Addition         | `5 + 3` | `8`    |
| `-`      | Subtraction      | `5 - 3` | `2`    |
| `*`      | Multiplication   | `5 * 3` | `15`   |
| `/`      | Division         | `5 / 2` | `2.5`  |
| `%`      | Modulo           | `5 % 2` | `1`    |
| `-`      | Negation (unary) | `-5`    | `-5`   |

**Examples:**

```yar
a = 10 + 5      // 15
b = 10 - 5      // 5
c = 10 * 5      // 50
d = 10 / 4      // 2.5
e = 10 % 3      // 1
f = -10         // -10
```

**String Concatenation:**

The `+` operator also concatenates strings:

```yar
greeting = "Hello, " + "World!"  // "Hello, World!"
```

### Comparison Operators

| Operator | Description           | Example  |
| -------- | --------------------- | -------- |
| `==`     | Equal to              | `5 == 5` |
| `!=`     | Not equal to          | `5 != 3` |
| `<`      | Less than             | `3 < 5`  |
| `<=`     | Less than or equal    | `5 <= 5` |
| `>`      | Greater than          | `5 > 3`  |
| `>=`     | Greater than or equal | `5 >= 5` |

All comparison operators return boolean values (`true` or `false`).

**Examples:**

```yar
a = 10
b = 20

a == b   // false
a != b   // true
a < b    // true
a <= b   // true
b > a    // true
b >= a   // true
```

**Comparing Different Types:**

Comparing values of different types returns `false` for equality:

```yar
10 == "10"     // false
true == 1      // false
nil == false   // false
```

### Logical Operators

| Operator | Description         | Example           |
| -------- | ------------------- | ----------------- |
| `&&`     | Logical AND         | `true && false`   |
| `\|\|`   | Logical OR          | `true \|\| false` |
| `!`      | Logical NOT (unary) | `!true`           |

**Examples:**

```yar
a = true
b = false

a && b    // false
a || b    // true
!a        // false
!b        // true
```

**Short-Circuit Evaluation:**

Logical operators use short-circuit evaluation:

- `&&` returns false if the left operand is falsy (doesn't evaluate right)
- `||` returns true if the left operand is truthy (doesn't evaluate right)

```yar
false && dangerous()  // dangerous() not called
true || expensive()   // expensive() not called
```

### Operator Precedence

From highest to lowest precedence:

1. Unary operators: `!`, `-` (unary)
2. Multiplicative: `*`, `/`, `%`
3. Additive: `+`, `-`
4. Comparison: `<`, `<=`, `>`, `>=`
5. Equality: `==`, `!=`
6. Logical AND: `&&`
7. Logical OR: `||`

**Examples:**

```yar
2 + 3 * 4        // 14 (not 20)
10 - 5 - 2       // 3 (left-to-right)
5 < 10 && 10 < 20  // true
```

Use parentheses to override precedence:

```yar
(2 + 3) * 4      // 20
```

## Control Flow

### If Statements

Execute code conditionally based on a boolean expression:

```yar
if condition {
    // code runs if condition is true
}
```

**Example:**

```yar
x = 10
if x > 5 {
    println("x is greater than 5")
}
```

### If/Else Statements

Provide an alternative path when the condition is false:

```yar
if condition {
    // runs if condition is true
} else {
    // runs if condition is false
}
```

**Example:**

```yar
age = 18
if age >= 18 {
    println("Adult")
} else {
    println("Minor")
}
```

### Else-If Chains

Test multiple conditions in sequence:

```yar
if condition1 {
    // runs if condition1 is true
} else if condition2 {
    // runs if condition1 is false and condition2 is true
} else if condition3 {
    // runs if condition1 and condition2 are false and condition3 is true
} else {
    // runs if all conditions are false
}
```

**Example:**

```yar
score = 85
if score >= 90 {
    println("Grade: A")
} else if score >= 80 {
    println("Grade: B")
} else if score >= 70 {
    println("Grade: C")
} else if score >= 60 {
    println("Grade: D")
} else {
    println("Grade: F")
}
```

### Nested Conditionals

If statements can be nested:

```yar
age = 25
hasLicense = true

if age >= 16 {
    if hasLicense {
        println("Can drive")
    } else {
        println("Need license")
    }
} else {
    println("Too young")
}
```

## Loops

### For Loops

YarLang supports C-style for loops with initialization, condition, and post-iteration:

```yar
for initialization; condition; post {
    // loop body
}
```

**Components:**

- **Initialization:** Executed once before the loop starts (typically sets a counter)
- **Condition:** Checked before each iteration; loop continues while true
- **Post:** Executed after each iteration (typically increments counter)

**Example:**

```yar
// Print numbers 0 through 4
for i = 0; i < 5; i = i + 1 {
    println(i)
}
```

**Counting Backwards:**

```yar
for i = 10; i >= 0; i = i - 1 {
    println(i)
}
```

**Counting by Steps:**

```yar
// Count by twos
for i = 0; i <= 10; i = i + 2 {
    println(i)  // 0, 2, 4, 6, 8, 10
}
```

### Nested Loops

Loops can be nested inside other loops:

```yar
for i = 0; i < 3; i = i + 1 {
    for j = 0; j < 3; j = j + 1 {
        println(i * 10 + j)
    }
}
```

### Break Statement

Exit a loop immediately:

```yar
for i = 0; i < 10; i = i + 1 {
    if i == 5 {
        break  // Exit loop when i is 5
    }
    println(i)  // Prints 0, 1, 2, 3, 4
}
```

**Finding a Value:**

```yar
target = 7
found = false

for i = 0; i < 100; i = i + 1 {
    if i == target {
        println("Found!")
        found = true
        break
    }
}
```

### Continue Statement

Skip the rest of the current iteration and continue with the next:

```yar
for i = 0; i < 10; i = i + 1 {
    if i == 5 {
        continue  // Skip printing 5
    }
    println(i)  // Prints 0, 1, 2, 3, 4, 6, 7, 8, 9
}
```

**Skipping Even Numbers:**

```yar
for i = 0; i < 10; i = i + 1 {
    if i % 2 == 0 {
        continue  // Skip even numbers
    }
    println(i)  // Prints 1, 3, 5, 7, 9
}
```

## Functions

### Function Definition

Define functions using the `func` keyword:

```yar
func functionName(param1, param2) {
    // function body
}
```

**Example:**

```yar
func greet(name) {
    println("Hello, " + name + "!")
}
```

### Function Parameters

Functions can accept zero or more parameters:

```yar
// No parameters
func sayHello() {
    println("Hello!")
}

// One parameter
func square(x) {
    return x * x
}

// Multiple parameters
func add(a, b) {
    return a + b
}
```

### Return Statements

Use `return` to return a value from a function:

```yar
func double(x) {
    return x * 2
}

result = double(5)  // result is 10
```

**Multiple Return Paths:**

Functions can have multiple return statements:

```yar
func abs(x) {
    if x < 0 {
        return -x
    }
    return x
}
```

**Returning Nothing:**

Functions without an explicit return value return `nil`:

```yar
func printMessage(msg) {
    println(msg)
    // Implicitly returns nil
}
```

### Calling Functions

Call functions by name with arguments in parentheses:

```yar
greet("Alice")
result = add(5, 3)
value = square(4)
```

### Recursion

Functions can call themselves recursively:

```yar
// Recursive factorial
func factorial(n) {
    if n <= 1 {
        return 1
    }
    return n * factorial(n - 1)
}

result = factorial(5)  // 120
```

**Fibonacci Example:**

```yar
func fibonacci(n) {
    if n <= 1 {
        return n
    }
    return fibonacci(n - 1) + fibonacci(n - 2)
}

println(fibonacci(10))  // 55
```

### Functions Calling Other Functions

Functions can call other user-defined functions:

```yar
func double(x) {
    return x * 2
}

func quadruple(x) {
    return double(double(x))
}

result = quadruple(5)  // 20
```

### Function Scope

Functions have their own local scope:

```yar
x = 10  // Global

func test() {
    x = 20  // Local to test()
    println(x)  // 20
}

test()
println(x)  // 10 (global unchanged)
```

## Built-in Functions

YarLang provides several built-in functions:

### print(value)

Print a value to stdout without a trailing newline:

```yar
print("Hello")
print(" ")
print("World")
// Output: Hello World
```

### println(value)

Print a value to stdout with a trailing newline:

```yar
println("Hello")
println("World")
// Output:
// Hello
// World
```

**Printing Numbers:**

```yar
println(42)
println(3.14159)
```

**Printing Booleans:**

```yar
println(true)   // prints: true
println(false)  // prints: false
```

### len(value)

Return the length of a string:

```yar
str = "Hello"
length = len(str)  // 5
println(length)
```

**Note:** Currently only works with strings. Returns an error for other types.

### type(value)

Return the type name of a value as a string:

```yar
println(type(42))        // "number"
println(type("hello"))   // "string"
println(type(true))      // "boolean"
println(type(nil))       // "nil"
```

**Use for Debugging:**

```yar
x = 10
println("x is a " + type(x))  // "x is a number"
```

## Examples

YarLang includes several example programs in the `examples/` directory:

### Basic Examples

- **hello.yar** - Simple variable assignment and printing
- **arithmetic.yar** - Basic arithmetic operations

### Feature Demonstrations

- **conditionals.yar** - If/else statements, else-if chains, nested conditionals
- **loops.yar** - For loops, break, continue, nested loops
- **functions.yar** - Function definitions, parameters, return values, recursion
- **fibonacci.yar** - Classic recursive Fibonacci implementation

See [examples.md](examples.md) for detailed explanations and output for each example.

## Language Limitations

Current limitations (may be addressed in future versions):

1. **No arrays or maps** - Only scalar types are supported
2. **No structs/objects** - No composite data structures
3. **No string interpolation** - Must use concatenation
4. **No file I/O** - No built-in file operations
5. **No standard library** - Limited built-in functions
6. **Single return values** - Syntax exists for multiple returns but runtime pending
7. **Limited error messages** - Error reporting could be more detailed
8. **No closures** - Functions cannot capture variables from outer scope

## Best Practices

### Code Style

- Use descriptive variable names
- Indent blocks with 4 spaces or 1 tab
- Put opening braces on the same line as the statement
- Use blank lines to separate logical sections
- Comment non-obvious code

```yar
// Good
func calculateArea(width, height) {
    area = width * height
    return area
}

// Also acceptable
func calculateArea(width, height)
{
    area = width * height
    return area
}
```

### Performance

- Avoid deeply recursive functions (no tail-call optimization)
- Minimize string concatenation in loops
- Use appropriate loop bounds

### Debugging

- Use `println()` to inspect values
- Use `type()` to check value types
- Test functions independently
- Build complex logic incrementally

## Summary

YarLang is a simple, dynamically-typed language with:

- Five data types (numbers, strings, booleans, nil, functions)
- Standard operators (arithmetic, comparison, logical)
- Control flow (if/else, for loops)
- First-class functions with recursion
- Automatic memory management
- Compilation to native code

For more details and examples, see:

- [README.md](../README.md) - Project overview and setup
- [examples.md](examples.md) - Detailed example walkthroughs
