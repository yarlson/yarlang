// Loops Example
// Demonstrates for loops with counters, break, continue,
// nested loops, and loops with conditionals

println("=== Basic For Loop ===")
for i = 0; i < 5; i = i + 1 {
    println(i)
}

println("")
println("=== Counting by Twos ===")
for j = 0; j <= 10; j = j + 2 {
    println(j)
}
println("")
println("=== For Loop with Break ===")
for k = 0; k < 10; k = k + 1 {
    println(k)
    if k == 5 {
        println("Breaking at 5")
        break
    }
}

println("")
println("=== For Loop with Continue ===")
for m = 0; m < 10; m = m + 1 {
    if m == 3 {
        continue
    }
    if m == 7 {
        continue
    }
    println(m)
}

println("")
println("=== Nested Loops ===")
for outer = 0; outer < 3; outer = outer + 1 {
    for inner = 0; inner < 3; inner = inner + 1 {
        println(outer * 10 + inner)
    }
}

println("")
println("=== Loop with Conditionals ===")
for n = 0; n < 20; n = n + 1 {
    if n > 15 {
        println("Large number")
    } else if n > 10 {
        println("Medium number")
    } else if n > 5 {
        println("Small number")
    }
}

println("")
println("=== Finding a Number ===")
target = 7
found = 0
for current = 0; current <= 10; current = current + 1 {
    if current == target {
        println("Found target!")
        found = 1
        break
    }
}
if found == 0 {
    println("Target not found")
}

println("")
println("=== Sum with Loop ===")
sum = 0
for counter = 1; counter <= 10; counter = counter + 1 {
    sum = sum + counter
}
println("Sum of 1 to 10:")
println(sum)

println("")
println("=== Complex Break/Continue Logic ===")
for idx = 0; idx < 15; idx = idx + 1 {
    // Skip multiples of 3
    remainder = idx - (idx / 3) * 3
    if remainder == 0 {
        continue
    }

    // Stop at 12
    if idx > 12 {
        break
    }

    println(idx)
}

println("Done with loops!")
