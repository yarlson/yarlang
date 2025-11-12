// Conditionals Example
// Demonstrates if/else statements, else-if chains, nested conditionals,
// comparison operators, and logical operators

println("=== Simple If Statement ===")
x = 10
if x > 5 {
    println("x is greater than 5")
}

println("")
println("=== If/Else Statement ===")
y = 3
if y > 5 {
    println("y is greater than 5")
} else {
    println("y is not greater than 5")
}

println("")
println("=== If/Else-If/Else Chain ===")
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

println("")
println("=== Nested Conditionals ===")
age = 25
hasLicense = 1  // Using 1 for true, 0 for false

if age >= 16 {
    if hasLicense == 1 {
        println("Can drive")
    } else {
        println("Too young or no license")
    }
} else {
    println("Too young to drive")
}

println("")
println("=== Comparison Operators ===")
a = 10
b = 20

if a == b {
    println("a equals b")
}

if a != b {
    println("a is not equal to b")
}

if a < b {
    println("a is less than b")
}

if a <= b {
    println("a is less than or equal to b")
}

if b > a {
    println("b is greater than a")
}

if b >= a {
    println("b is greater than or equal to a")
}

println("")
println("=== Complex Conditions ===")
temperature = 75
if temperature > 70 {
    if temperature < 85 {
        println("Nice weather!")
    }
}

num = 42
if num > 0 {
    if num < 100 {
        println("Number is between 0 and 100")
    }
}

println("")
println("=== Multiple Paths ===")
value = 0
if value > 0 {
    println("Positive")
} else if value < 0 {
    println("Negative")
} else {
    println("Zero")
}

println("Done with conditionals!")
