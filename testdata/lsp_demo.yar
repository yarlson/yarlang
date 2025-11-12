// LSP Feature Demonstration File
// This file demonstrates all LSP features that should work in Zed

// ============================================
// SECTION 1: Variable Declarations
// ============================================
// These should show up in document symbols
x = 42
name = "Alice"
isActive = true
nothing = nil

// ============================================
// SECTION 2: Function Definitions
// ============================================
// Hover over function names to see signatures
// Go-to-definition should work on function calls

func add(a, b) {
	return a + b
}

func multiply(x, y) {
	result = x * y
	return result
}

func greet(name) {
	message = "Hello, " + name
	return message
}

// ============================================
// SECTION 3: Function Calls with Completion
// ============================================
// Type 'add(' to see parameter hints
sum = add(10, 20)

// Type 'multiply(' to see parameter hints
product = multiply(5, 6)

// Type 'greet(' to see parameter hints
greeting = greet(name)

// ============================================
// SECTION 4: Expressions and Type Checking
// ============================================
// These should work without errors
calculation = x + 10
fullName = "Mr. " + name
comparison = x > 20
logic = isActive && true

// ============================================
// SECTION 5: Diagnostics - These SHOULD show errors
// ============================================

// ERROR: Undefined variable
badVar = undefined_variable + 1

// ERROR: Undefined function
badFunc = nonexistent_function(42)

// ERROR: Using variable before definition
early = laterVar
laterVar = 100

// ============================================
// SECTION 6: Hover Information
// ============================================
// Hover over these to see type information:
// - 'x' should show: int (value: 42)
// - 'name' should show: string (value: "Alice")
// - 'add' should show: func(a, b) -> any
// - 'sum' should show: any (result of add)

testHover = x
testHover2 = name
testHover3 = add

// ============================================
// SECTION 7: Go to Definition
// ============================================
// Cmd/Ctrl+Click on these should jump to definition:
useX = x              // Should jump to line 8
useName = name        // Should jump to line 9
useAdd = add(1, 2)    // Should jump to line 18
useMultiply = multiply(3, 4)  // Should jump to line 23

// ============================================
// SECTION 8: Code Completion
// ============================================
// Type these prefixes and trigger completion (Ctrl+Space):
// - Type 'x' -> should suggest 'x'
// - Type 'add' -> should suggest 'add'
// - Type 'mul' -> should suggest 'multiply'
// - Type 'gre' -> should suggest 'greet', 'greeting'

// Try completion here (uncomment and type):
// test =

// ============================================
// SECTION 9: Document Symbols
// ============================================
// Open symbol palette (Cmd+Shift+O) to see:
// Variables: x, name, isActive, nothing, sum, product, greeting, etc.
// Functions: add, multiply, greet

// ============================================
// SECTION 10: Complex Expressions
// ============================================
func fibonacci(n) {
	if n <= 1 {
		return n, nil
	}
	a, _ = fibonacci(n - 1)
	b, _ = fibonacci(n - 2)
	return a + b, nil
}

// Test fibonacci
fib10, err = fibonacci(10)
if err != nil {
	// Handle error
	errorMsg = "Failed"
} else {
	// Success
	successMsg = "Success"
}
