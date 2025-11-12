// Functions Example
// Demonstrates function definitions, function calls, parameters,
// return values, multiple functions calling each other, and recursion

println("=== Simple Function ===")

// Function with no parameters
func greet() {
    println("Hello from a function!")
}

greet()
println("")
println("=== Function with Parameters ===")

// Function that takes parameters
func add(a, b) {
    return a + b
}

result1 = add(5, 3)
println("5 + 3 = ")
println(result1)

result2 = add(10, 20)
println("10 + 20 = ")
println(result2)

println("")
println("=== Function with Multiple Operations ===")

// Function with more complex logic
func calculate(x, y) {
    sum = x + y
    product = x * y
    if sum > product {
        return sum
    }
    return product
}

val1 = calculate(5, 3)
println("calculate(5, 3) = ")
println(val1)

val2 = calculate(10, 20)
println("calculate(10, 20) = ")
println(val2)

println("")
println("=== Multiple Functions Calling Each Other ===")

// Function that doubles a number
func double(n) {
    return n * 2
}

// Function that triples a number
func triple(n) {
    return n * 3
}

// Function that uses other functions
func process(x) {
    doubled = double(x)
    tripled = triple(x)
    return doubled + tripled
}

val3 = process(5)
println("process(5) = ")
println(val3)

println("")
println("=== Recursive Function (Factorial) ===")

// Recursive factorial function
func factorial(n) {
    if n <= 1 {
        return 1
    }
    return n * factorial(n - 1)
}

fact5 = factorial(5)
println("factorial(5) = ")
println(fact5)

fact7 = factorial(7)
println("factorial(7) = ")
println(fact7)

println("")
println("=== Recursive Function (Countdown) ===")

// Recursive countdown function
func countdown(n) {
    if n <= 0 {
        println("Blastoff!")
        return 0
    }
    println(n)
    return countdown(n - 1)
}

countdown(5)

println("")
println("=== Function with Conditionals ===")

// Function that classifies numbers
func classify(num) {
    if num > 100 {
        return 3  // Large
    } else if num > 50 {
        return 2  // Medium
    } else if num > 0 {
        return 1  // Small
    }
    return 0  // Zero or negative
}

class1 = classify(150)
println("classify(150) = ")
println(class1)

class2 = classify(75)
println("classify(75) = ")
println(class2)

class3 = classify(25)
println("classify(25) = ")
println(class3)

println("")
println("=== Nested Function Calls ===")

func addOne(n) {
    return n + 1
}

func addTwo(n) {
    return addOne(addOne(n))
}

func addFour(n) {
    return addTwo(addTwo(n))
}

result = addFour(10)
println("addFour(10) = ")
println(result)

println("")
println("=== Function Computing Maximum ===")

func max(a, b) {
    if a > b {
        return a
    }
    return b
}

func maxOfThree(a, b, c) {
    return max(a, max(b, c))
}

maxVal = maxOfThree(15, 42, 27)
println("maxOfThree(15, 42, 27) = ")
println(maxVal)

println("Done with functions!")
