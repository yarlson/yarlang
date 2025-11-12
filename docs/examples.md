# YarLang Examples Guide

Detailed explanations of all example programs included with YarLang.

## Table of Contents

1. [hello.yar](#helloyar) - Basic output
2. [arithmetic.yar](#arithmeticyar) - Arithmetic operations
3. [conditionals.yar](#conditionsyar) - Conditional logic
4. [loops.yar](#loopsyar) - Loop constructs
5. [functions.yar](#functionsyar) - Function definitions and calls
6. [fibonacci.yar](#fibonacciyar) - Recursive Fibonacci

---

## hello.yar

**Location:** `examples/hello.yar`

**Features Demonstrated:**

- Variable assignment
- Basic output with `println()`

### Code

```yar
x = 42
println(x)
```

### Explanation

This is the simplest YarLang program. It demonstrates:

1. **Variable Assignment:** `x = 42` creates a variable named `x` and assigns it the numeric value 42
2. **Output:** `println(x)` prints the value of `x` followed by a newline

### Expected Output

```
42
```

### Key Concepts

- Variables are created through assignment (no declaration needed)
- Numbers are double-precision floating-point
- `println()` is a built-in function that prints a value with a newline

---

## arithmetic.yar

**Location:** `examples/arithmetic.yar`

**Features Demonstrated:**

- Variable assignment
- Arithmetic operators (+, -, \*, /)
- Multiple println() calls

### Code

```yar
a = 10
b = 5
sum = a + b
diff = a - b
prod = a * b
quot = a / b

println(sum)
println(diff)
println(prod)
println(quot)
```

### Explanation

This program demonstrates basic arithmetic operations:

1. **Variables:** Creates two variables `a` (10) and `b` (5)
2. **Addition:** `sum = a + b` computes 10 + 5 = 15
3. **Subtraction:** `diff = a - b` computes 10 - 5 = 5
4. **Multiplication:** `prod = a * b` computes 10 \* 5 = 50
5. **Division:** `quot = a / b` computes 10 / 5 = 2
6. **Output:** Each result is printed on a separate line

### Expected Output

```
15
5
50
2
```

### Key Concepts

- Arithmetic operators work on numeric values
- Division produces floating-point results (2.0, displayed as 2)
- Variables can be used in expressions
- Results of expressions can be stored in new variables

---

## conditionals.yar

**Location:** `examples/conditionals.yar`

**Features Demonstrated:**

- If statements
- If/else statements
- Else-if chains
- Nested conditionals
- Comparison operators (==, !=, <, <=, >, >=)
- Complex conditions

### Code Highlights

The full code is extensive (107 lines). Here are key sections:

**Simple If Statement:**

```yar
x = 10
if x > 5 {
    println("x is greater than 5")
}
```

**If/Else Statement:**

```yar
y = 3
if y > 5 {
    println("y is greater than 5")
} else {
    println("y is not greater than 5")
}
```

**Else-If Chain:**

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

**Nested Conditionals:**

```yar
age = 25
hasLicense = 1  // Using 1 for true

if age >= 16 {
    if hasLicense == 1 {
        println("Can drive")
    } else {
        println("Too young or no license")
    }
} else {
    println("Too young to drive")
}
```

### Explanation

This comprehensive example demonstrates all conditional logic features:

1. **Simple If:** Executes code only if condition is true
2. **If/Else:** Provides alternative path when condition is false
3. **Else-If Chains:** Tests multiple conditions sequentially
4. **Nested Conditionals:** If statements inside other if statements
5. **All Comparison Operators:** ==, !=, <, <=, >, >=
6. **Complex Conditions:** Multiple conditions in sequence

### Expected Output

```
=== Simple If Statement ===
x is greater than 5

=== If/Else Statement ===
y is not greater than 5

=== If/Else-If/Else Chain ===
Grade: B

=== Nested Conditionals ===
Can drive

=== Comparison Operators ===
a is not equal to b
a is less than b
a is less than or equal to b
b is greater than a
b is greater than or equal to a

=== Complex Conditions ===
Nice weather!
Number is between 0 and 100

=== Multiple Paths ===
Zero
Done with conditionals!
```

### Key Concepts

- Else-if chains test conditions in order; first true condition executes
- Nested conditionals allow testing multiple criteria
- All comparison operators return boolean values
- Block structure with curly braces defines scope

---

## loops.yar

**Location:** `examples/loops.yar`

**Features Demonstrated:**

- Basic for loops
- Loop counters with different increments
- Break statement
- Continue statement
- Nested loops
- Loops with conditionals
- Finding values in loops
- Accumulator patterns

### Code Highlights

The full code is 100 lines. Here are key sections:

**Basic For Loop:**

```yar
for i = 0; i < 5; i = i + 1 {
    println(i)
}
```

**Counting by Twos:**

```yar
for j = 0; j <= 10; j = j + 2 {
    println(j)
}
```

**For Loop with Break:**

```yar
for k = 0; k < 10; k = k + 1 {
    println(k)
    if k == 5 {
        println("Breaking at 5")
        break
    }
}
```

**For Loop with Continue:**

```yar
for m = 0; m < 10; m = m + 1 {
    if m == 3 {
        continue
    }
    if m == 7 {
        continue
    }
    println(m)
}
```

**Nested Loops:**

```yar
for outer = 0; outer < 3; outer = outer + 1 {
    for inner = 0; inner < 3; inner = inner + 1 {
        println(outer * 10 + inner)
    }
}
```

**Sum with Loop (Accumulator Pattern):**

```yar
sum = 0
for counter = 1; counter <= 10; counter = counter + 1 {
    sum = sum + counter
}
println("Sum of 1 to 10:")
println(sum)
```

### Explanation

This example demonstrates all loop-related features:

1. **Basic For Loop:** Standard C-style loop with counter
2. **Different Increments:** Counting by 2s, backwards, etc.
3. **Break Statement:** Exit loop early based on condition
4. **Continue Statement:** Skip rest of iteration, continue with next
5. **Nested Loops:** Loops inside loops for multi-dimensional iteration
6. **Conditionals in Loops:** Combining loops with if statements
7. **Finding Values:** Search pattern with break
8. **Accumulator Pattern:** Building up a sum over iterations

### Expected Output

```
=== Basic For Loop ===
0
1
2
3
4

=== Counting by Twos ===
0
2
4
6
8
10

=== For Loop with Break ===
0
1
2
3
4
5
Breaking at 5

=== For Loop with Continue ===
0
1
2
4
5
6
8
9

=== Nested Loops ===
0
1
2
10
11
12
20
21
22

=== Loop with Conditionals ===
Small number
Small number
Small number
Small number
Small number
Medium number
Medium number
Medium number
Medium number
Medium number
Large number
Large number
Large number
Large number

=== Finding a Number ===
Found target!

=== Sum with Loop ===
Sum of 1 to 10:
55

=== Complex Break/Continue Logic ===
1
2
4
5
7
8
10
11

Done with loops!
```

### Key Concepts

- For loops have three parts: initialization, condition, post-iteration
- Break exits the innermost loop immediately
- Continue skips to the next iteration
- Nested loops allow multi-dimensional iteration
- Accumulator pattern: initialize outside loop, update inside loop

---

## functions.yar

**Location:** `examples/functions.yar`

**Features Demonstrated:**

- Function definition with `func` keyword
- Functions with no parameters
- Functions with parameters
- Return statements
- Multiple return paths
- Functions calling other functions
- Recursive functions
- Nested function calls

### Code Highlights

The full code is 174 lines. Here are key sections:

**Simple Function (No Parameters):**

```yar
func greet() {
    println("Hello from a function!")
}

greet()
```

**Function with Parameters:**

```yar
func add(a, b) {
    return a + b
}

result1 = add(5, 3)
println("5 + 3 = ")
println(result1)
```

**Function with Multiple Operations:**

```yar
func calculate(x, y) {
    sum = x + y
    product = x * y
    if sum > product {
        return sum
    }
    return product
}

val1 = calculate(5, 3)   // Returns 15 (product)
val2 = calculate(10, 20) // Returns 200 (product)
```

**Multiple Functions Calling Each Other:**

```yar
func double(n) {
    return n * 2
}

func triple(n) {
    return n * 3
}

func process(x) {
    doubled = double(x)
    tripled = triple(x)
    return doubled + tripled
}

val3 = process(5)  // 10 + 15 = 25
```

**Recursive Function (Factorial):**

```yar
func factorial(n) {
    if n <= 1 {
        return 1
    }
    return n * factorial(n - 1)
}

fact5 = factorial(5)  // 120
fact7 = factorial(7)  // 5040
```

**Recursive Function (Countdown):**

```yar
func countdown(n) {
    if n <= 0 {
        println("Blastoff!")
        return 0
    }
    println(n)
    return countdown(n - 1)
}

countdown(5)  // Prints: 5, 4, 3, 2, 1, Blastoff!
```

**Nested Function Calls:**

```yar
func addOne(n) {
    return n + 1
}

func addTwo(n) {
    return addOne(addOne(n))
}

func addFour(n) {
    return addTwo(addTwo(n))
}

result = addFour(10)  // 14
```

**Computing Maximum:**

```yar
func max(a, b) {
    if a > b {
        return a
    }
    return b
}

func maxOfThree(a, b, c) {
    return max(a, max(b, c))
}

maxVal = maxOfThree(15, 42, 27)  // 42
```

### Explanation

This comprehensive example demonstrates all function features:

1. **Basic Definition:** Functions defined with `func` keyword
2. **Parameters:** Functions can accept zero or more parameters
3. **Return Values:** Functions return values with `return` statement
4. **Multiple Return Paths:** Different return statements based on conditions
5. **Function Calls:** Calling user-defined functions
6. **Composition:** Functions calling other user-defined functions
7. **Recursion:** Functions calling themselves
8. **Nested Calls:** Using return values as arguments to other functions

### Expected Output

```
=== Simple Function ===
Hello from a function!

=== Function with Parameters ===
5 + 3 =
8
10 + 20 =
30

=== Function with Multiple Operations ===
calculate(5, 3) =
15
calculate(10, 20) =
200

=== Multiple Functions Calling Each Other ===
process(5) =
25

=== Recursive Function (Factorial) ===
factorial(5) =
120
factorial(7) =
5040

=== Recursive Function (Countdown) ===
5
4
3
2
1
Blastoff!

=== Function with Conditionals ===
classify(150) =
3
classify(75) =
2
classify(25) =
1

=== Nested Function Calls ===
addFour(10) =
14

=== Function Computing Maximum ===
maxOfThree(15, 42, 27) =
42

Done with functions!
```

### Key Concepts

- Functions encapsulate reusable code
- Parameters allow passing data into functions
- Return statements pass data back to the caller
- Functions can call other functions (composition)
- Recursion requires a base case to avoid infinite loops
- Functions create their own local scope

---

## fibonacci.yar

**Location:** `examples/fibonacci.yar`

**Features Demonstrated:**

- Recursive function definition
- Conditional logic in recursion
- Multiple return paths
- For loop integration
- Function calls in expressions

### Code

```yar
// Fibonacci Example
// Demonstrates recursive function definition, conditional logic,
// multiple return paths, and arithmetic operations

// Recursive fibonacci function
func fibonacci(n) {
    if n <= 1 {
        return n
    }
    return fibonacci(n - 1) + fibonacci(n - 2)
}

// Main execution
println("Computing Fibonacci sequence:")

// Compute and print fibonacci numbers
for i = 0; i <= 10; i = i + 1 {
    result = fibonacci(i)
    println(result)
}

println("Done!")
```

### Explanation

This classic example computes the Fibonacci sequence using recursion:

**The Fibonacci Sequence:**

Each number is the sum of the two preceding numbers:

- F(0) = 0
- F(1) = 1
- F(n) = F(n-1) + F(n-2) for n > 1

Sequence: 0, 1, 1, 2, 3, 5, 8, 13, 21, 34, 55...

**How It Works:**

1. **Base Cases:** If n <= 1, return n directly
   - fibonacci(0) returns 0
   - fibonacci(1) returns 1

2. **Recursive Case:** For n > 1, return the sum of the two previous numbers
   - fibonacci(2) = fibonacci(1) + fibonacci(0) = 1 + 0 = 1
   - fibonacci(3) = fibonacci(2) + fibonacci(1) = 1 + 1 = 2
   - fibonacci(4) = fibonacci(3) + fibonacci(2) = 2 + 1 = 3
   - And so on...

3. **Main Loop:** Computes and prints F(0) through F(10)

### Expected Output

```
Computing Fibonacci sequence:
0
1
1
2
3
5
8
13
21
34
55
Done!
```

### Key Concepts

- **Recursion:** Function calls itself with different arguments
- **Base Case:** Essential to stop recursion (n <= 1)
- **Recursive Case:** Breaks problem into smaller subproblems
- **Call Tree:** Each call spawns two more calls (exponential growth)
- **Expression Recursion:** fibonacci() called twice in return expression

### Performance Note

This implementation has exponential time complexity O(2^n) because it recalculates the same values multiple times. For example:

- fibonacci(4) calls fibonacci(3) and fibonacci(2)
- fibonacci(3) calls fibonacci(2) and fibonacci(1)
- So fibonacci(2) is calculated twice

This makes it slow for large values of n, but it's a clear demonstration of recursive thinking.

---

## Running the Examples

To run any example:

```bash
# Compile the example
./yar examples/fibonacci.yar

# Run the generated executable
./examples/fibonacci
```

Or use the test script to run all examples:

```bash
./test-examples.sh
```

## Learning Path

Recommended order for learning YarLang:

1. **hello.yar** - Start here for basic syntax
2. **arithmetic.yar** - Learn about operators and expressions
3. **conditionals.yar** - Understand control flow
4. **loops.yar** - Master iteration
5. **functions.yar** - Learn function concepts
6. **fibonacci.yar** - Apply everything with recursion

## Modifying Examples

Feel free to modify these examples to experiment:

**Try These Modifications:**

- **hello.yar:** Change the value and type of variables
- **arithmetic.yar:** Add modulo (%) operation, test with different numbers
- **conditionals.yar:** Add more else-if branches, try different conditions
- **loops.yar:** Change loop bounds, add more complex break/continue logic
- **functions.yar:** Create your own functions, combine them in new ways
- **fibonacci.yar:** Try computing larger Fibonacci numbers (be patient!)

## Creating Your Own Programs

To create a new YarLang program:

1. Create a file with `.yar` extension
2. Write your code using YarLang syntax
3. Compile with `./yar yourfile.yar`
4. Run with `./yourfile`

**Example New Program (sum.yar):**

```yar
// Compute sum of squares
func sumOfSquares(n) {
    sum = 0
    for i = 1; i <= n; i = i + 1 {
        sum = sum + (i * i)
    }
    return sum
}

result = sumOfSquares(10)
println("Sum of squares from 1 to 10:")
println(result)  // 385
```

## Getting Help

- See [language-guide.md](language-guide.md) for complete syntax reference
- See [README.md](../README.md) for build and usage instructions
- Check error messages for syntax errors or runtime issues
- Use `println()` liberally to debug your programs

Happy coding with YarLang!
