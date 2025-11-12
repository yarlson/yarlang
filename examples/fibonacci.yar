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
