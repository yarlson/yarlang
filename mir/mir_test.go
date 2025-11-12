package mir

import (
	"strings"
	"testing"

	"github.com/yarlson/yarlang/lexer"
	"github.com/yarlson/yarlang/parser"
)

func TestMIRNodes(t *testing.T) {
	// Test basic ops
	alloca := &Alloca{Name: "x", Type: &PrimitiveType{Name: "i32"}}
	if alloca.String() != "%x = alloca i32" {
		t.Errorf("wrong string: %s", alloca.String())
	}

	// Test binary op (operands stored WITHOUT % prefix)
	add := &BinOp{Dest: "t1", Op: Add, Left: "a", Right: "b", Type: &PrimitiveType{Name: "i32"}}
	if add.Op != Add {
		t.Error("wrong op")
	}

	expected := "%t1 = add i32 %a, %b"
	if add.String() != expected {
		t.Errorf("wrong string: got %q, want %q", add.String(), expected)
	}
}

func TestCallExprLowering(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name: "simple function call with int args",
			input: `
fn add(a i32, b i32) i32 {
	return a + b
}

fn main() {
	let x = add(1, 2)
}`,
			contains: []string{
				"call i32 @add",
				"1", "2",
			},
		},
		{
			name: "void function call with string literal",
			input: `
fn println(msg ptr) {
}

fn main() {
	println("hello")
}`,
			contains: []string{
				"call void @println",
				`@.str.1`, // String literals are now lowered to global references
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			file := p.ParseFile()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			lowerer := NewLowerer()
			module := lowerer.LowerFile(file)

			// Find the main function
			var mainFunc *Function

			for _, fn := range module.Functions {
				if fn.Name == "main" {
					mainFunc = fn
					break
				}
			}

			if mainFunc == nil {
				t.Fatal("main function not found")
			}

			// Check that we have at least one block
			if len(mainFunc.Blocks) == 0 {
				t.Fatal("main function has no blocks")
			}

			// Convert all instructions to string
			var output string

			for _, block := range mainFunc.Blocks {
				for _, instr := range block.Instrs {
					output += instr.String() + "\n"
				}
			}

			// Check for expected substrings
			for _, substr := range tt.contains {
				if !strings.Contains(output, substr) {
					t.Errorf("output missing expected substring %q\nGot:\n%s", substr, output)
				}
			}
		})
	}
}

func TestVoidFunctionImplicitReturn(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name: "void function with no explicit return gets implicit ret void",
			input: `
fn main() {
	println("hello")
}`,
			contains: []string{
				"call void @println",
				"ret void",
			},
		},
		{
			name: "void function with explicit return keeps it",
			input: `
fn main() {
	println("hello")
	return
}`,
			contains: []string{
				"call void @println",
				"ret void",
			},
		},
		{
			name: "non-void function with explicit return unchanged",
			input: `
fn add(a i32, b i32) i32 {
	return a + b
}`,
			contains: []string{
				"ret i32",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			file := p.ParseFile()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			lowerer := NewLowerer()
			module := lowerer.LowerFile(file)

			// Find the function (main or add)
			var testFunc *Function

			for _, fn := range module.Functions {
				if fn.Name == "main" || fn.Name == "add" {
					testFunc = fn
					break
				}
			}

			if testFunc == nil {
				t.Fatal("test function not found")
			}

			// Check that we have at least one block
			if len(testFunc.Blocks) == 0 {
				t.Fatal("test function has no blocks")
			}

			// Check that the last instruction is a terminator (Ret or Br)
			lastBlock := testFunc.Blocks[len(testFunc.Blocks)-1]
			if len(lastBlock.Instrs) == 0 {
				t.Fatal("last block has no instructions")
			}

			lastInstr := lastBlock.Instrs[len(lastBlock.Instrs)-1]
			if _, ok := lastInstr.(*Ret); !ok {
				if _, ok := lastInstr.(*Br); !ok {
					if _, ok := lastInstr.(*CondBr); !ok {
						t.Errorf("last instruction is not a terminator: %T: %s", lastInstr, lastInstr.String())
					}
				}
			}

			// Convert all instructions to string
			var output string

			for _, block := range testFunc.Blocks {
				for _, instr := range block.Instrs {
					output += instr.String() + "\n"
				}
			}

			// Check for expected substrings
			for _, substr := range tt.contains {
				if !strings.Contains(output, substr) {
					t.Errorf("output missing expected substring %q\nGot:\n%s", substr, output)
				}
			}
		})
	}
}

func TestIfStmtLowering(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		contains     []string
		blockCount   int
		blockLabels  []string
	}{
		{
			name: "simple if without else",
			input: `
fn main() {
	let x = 5
	if x > 0 {
		let y = 1
	}
}`,
			contains: []string{
				"alloca i32",
				"gt i32",
				"br i1",
				"label %bb_then",
				"label %bb_merge",
			},
			blockCount: 3, // entry, then, merge
			blockLabels: []string{"entry", "then", "merge"},
		},
		{
			name: "if with else",
			input: `
fn main() {
	let x = 5
	if x > 0 {
		let y = 1
	} else {
		let y = 2
	}
}`,
			contains: []string{
				"alloca i32",
				"gt i32",
				"br i1",
				"label %bb_then",
				"label %bb_else",
				"label %bb_merge",
			},
			blockCount: 4, // entry, then, else, merge
			blockLabels: []string{"entry", "then", "else", "merge"},
		},
		{
			name: "nested if",
			input: `
fn main() {
	let x = 5
	if x > 0 {
		if x < 10 {
			let y = 1
		}
	}
}`,
			contains: []string{
				"alloca i32",
				"gt i32",
				"lt i32",
				"br i1",
			},
			blockCount: 5, // entry, then1, then2, merge2, merge1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			file := p.ParseFile()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			lowerer := NewLowerer()
			module := lowerer.LowerFile(file)

			var mainFunc *Function
			for _, fn := range module.Functions {
				if fn.Name == "main" {
					mainFunc = fn
					break
				}
			}

			if mainFunc == nil {
				t.Fatal("main function not found")
			}

			// Check block count
			if tt.blockCount > 0 && len(mainFunc.Blocks) != tt.blockCount {
				t.Errorf("expected %d blocks, got %d", tt.blockCount, len(mainFunc.Blocks))
			}

			// Check block labels if specified
			if len(tt.blockLabels) > 0 {
				for i, expectedLabel := range tt.blockLabels {
					if i >= len(mainFunc.Blocks) {
						t.Errorf("expected block %d with label containing %q, but only %d blocks exist", i, expectedLabel, len(mainFunc.Blocks))
						continue
					}
					if !strings.Contains(mainFunc.Blocks[i].Label, expectedLabel) {
						t.Errorf("expected block %d label to contain %q, got %q", i, expectedLabel, mainFunc.Blocks[i].Label)
					}
				}
			}

			// Convert all blocks to string
			var output string
			for _, block := range mainFunc.Blocks {
				output += block.String()
			}

			// Check for expected substrings
			for _, substr := range tt.contains {
				if !strings.Contains(output, substr) {
					t.Errorf("output missing expected substring %q\nGot:\n%s", substr, output)
				}
			}
		})
	}
}

func TestWhileStmtLowering(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		contains    []string
		blockCount  int
		blockLabels []string
	}{
		{
			name: "simple while loop",
			input: `
fn main() {
	let x = 0
	while x < 10 {
		x = x + 1
	}
}`,
			contains: []string{
				"alloca i32",
				"lt i32",
				"br i1",
				"label %bb_cond",
				"label %bb_body",
				"label %bb_exit",
				"br label %bb_cond", // loop back
			},
			blockCount: 4, // entry, cond, body, exit
			blockLabels: []string{"entry", "cond", "body", "exit"},
		},
		{
			name: "nested while loops",
			input: `
fn main() {
	let x = 0
	while x < 10 {
		let y = 0
		while y < 5 {
			y = y + 1
		}
		x = x + 1
	}
}`,
			contains: []string{
				"alloca i32",
				"lt i32",
				"br i1",
				"label %bb_cond",
				"label %bb_body",
			},
			blockCount: 7, // entry, cond1, body1, cond2, body2, exit2, exit1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			file := p.ParseFile()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			lowerer := NewLowerer()
			module := lowerer.LowerFile(file)

			var mainFunc *Function
			for _, fn := range module.Functions {
				if fn.Name == "main" {
					mainFunc = fn
					break
				}
			}

			if mainFunc == nil {
				t.Fatal("main function not found")
			}

			// Check block count
			if tt.blockCount > 0 && len(mainFunc.Blocks) != tt.blockCount {
				t.Errorf("expected %d blocks, got %d", tt.blockCount, len(mainFunc.Blocks))
			}

			// Check block labels if specified
			if len(tt.blockLabels) > 0 {
				for i, expectedLabel := range tt.blockLabels {
					if i >= len(mainFunc.Blocks) {
						t.Errorf("expected block %d with label containing %q, but only %d blocks exist", i, expectedLabel, len(mainFunc.Blocks))
						continue
					}
					if !strings.Contains(mainFunc.Blocks[i].Label, expectedLabel) {
						t.Errorf("expected block %d label to contain %q, got %q", i, expectedLabel, mainFunc.Blocks[i].Label)
					}
				}
			}

			// Convert all blocks to string
			var output string
			for _, block := range mainFunc.Blocks {
				output += block.String()
			}

			// Check for expected substrings
			for _, substr := range tt.contains {
				if !strings.Contains(output, substr) {
					t.Errorf("output missing expected substring %q\nGot:\n%s", substr, output)
				}
			}
		})
	}
}

func TestForStmtLowering(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		contains    []string
		blockCount  int
		blockLabels []string
	}{
		{
			name: "simple for loop with range",
			input: `
fn main() {
	for i in 0..10 {
		let x = i
	}
}`,
			contains: []string{
				"alloca i32",
				"store i32",
				"lt i32",
				"br i1",
				"label %bb_cond",
				"label %bb_body",
				"label %bb_exit",
				"add i32", // iterator increment
			},
			blockCount: 4, // entry, cond, body, exit
			blockLabels: []string{"entry", "cond", "body", "exit"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			file := p.ParseFile()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			lowerer := NewLowerer()
			module := lowerer.LowerFile(file)

			var mainFunc *Function
			for _, fn := range module.Functions {
				if fn.Name == "main" {
					mainFunc = fn
					break
				}
			}

			if mainFunc == nil {
				t.Fatal("main function not found")
			}

			// Check block count
			if tt.blockCount > 0 && len(mainFunc.Blocks) != tt.blockCount {
				t.Errorf("expected %d blocks, got %d", tt.blockCount, len(mainFunc.Blocks))
			}

			// Check block labels if specified
			if len(tt.blockLabels) > 0 {
				for i, expectedLabel := range tt.blockLabels {
					if i >= len(mainFunc.Blocks) {
						t.Errorf("expected block %d with label containing %q, but only %d blocks exist", i, expectedLabel, len(mainFunc.Blocks))
						continue
					}
					if !strings.Contains(mainFunc.Blocks[i].Label, expectedLabel) {
						t.Errorf("expected block %d label to contain %q, got %q", i, expectedLabel, mainFunc.Blocks[i].Label)
					}
				}
			}

			// Convert all blocks to string
			var output string
			for _, block := range mainFunc.Blocks {
				output += block.String()
			}

			// Check for expected substrings
			for _, substr := range tt.contains {
				if !strings.Contains(output, substr) {
					t.Errorf("output missing expected substring %q\nGot:\n%s", substr, output)
				}
			}
		})
	}
}

func TestBreakStmtLowering(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
		wantErr  bool
	}{
		{
			name: "break in while loop",
			input: `
fn main() {
	let x = 0
	while x < 10 {
		if x == 5 {
			break
		}
		x = x + 1
	}
}`,
			contains: []string{
				"label %bb_cond",
				"label %bb_body",
				"label %bb_exit",
				"br label %bb_exit", // break jumps to exit
			},
		},
		{
			name: "break in for loop",
			input: `
fn main() {
	for i in 0..10 {
		if i == 5 {
			break
		}
	}
}`,
			contains: []string{
				"label %bb_cond",
				"label %bb_body",
				"label %bb_exit",
				"br label %bb_exit", // break jumps to exit
			},
		},
		{
			name: "nested loops with break in inner loop",
			input: `
fn main() {
	let x = 0
	while x < 10 {
		let y = 0
		while y < 5 {
			if y == 3 {
				break
			}
			y = y + 1
		}
		x = x + 1
	}
}`,
			contains: []string{
				"label %bb_cond",
				"label %bb_body",
				"label %bb_exit",
				"br label %bb_exit", // break in inner loop jumps to inner exit
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			file := p.ParseFile()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			lowerer := NewLowerer()
			module := lowerer.LowerFile(file)

			var mainFunc *Function
			for _, fn := range module.Functions {
				if fn.Name == "main" {
					mainFunc = fn
					break
				}
			}

			if mainFunc == nil {
				t.Fatal("main function not found")
			}

			// Convert all blocks to string
			var output string
			for _, block := range mainFunc.Blocks {
				output += block.String()
			}

			// Check for expected substrings
			for _, substr := range tt.contains {
				if !strings.Contains(output, substr) {
					t.Errorf("output missing expected substring %q\nGot:\n%s", substr, output)
				}
			}
		})
	}
}

func TestContinueStmtLowering(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name: "continue in while loop",
			input: `
fn main() {
	let x = 0
	while x < 10 {
		x = x + 1
		if x == 5 {
			continue
		}
		let y = 1
	}
}`,
			contains: []string{
				"label %bb_cond",
				"label %bb_body",
				"label %bb_exit",
				"br label %bb_cond", // continue jumps to condition
			},
		},
		{
			name: "continue in for loop",
			input: `
fn main() {
	for i in 0..10 {
		if i == 5 {
			continue
		}
		let x = i
	}
}`,
			contains: []string{
				"label %bb_cond",
				"label %bb_body",
				"label %bb_exit",
				"br label %bb_cond", // continue jumps to condition
			},
		},
		{
			name: "nested loops with continue in inner loop",
			input: `
fn main() {
	let x = 0
	while x < 10 {
		let y = 0
		while y < 5 {
			y = y + 1
			if y == 3 {
				continue
			}
		}
		x = x + 1
	}
}`,
			contains: []string{
				"label %bb_cond",
				"label %bb_body",
				"br label %bb_cond", // continue in inner loop jumps to inner condition
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			file := p.ParseFile()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			lowerer := NewLowerer()
			module := lowerer.LowerFile(file)

			var mainFunc *Function
			for _, fn := range module.Functions {
				if fn.Name == "main" {
					mainFunc = fn
					break
				}
			}

			if mainFunc == nil {
				t.Fatal("main function not found")
			}

			// Convert all blocks to string
			var output string
			for _, block := range mainFunc.Blocks {
				output += block.String()
			}

			// Check for expected substrings
			for _, substr := range tt.contains {
				if !strings.Contains(output, substr) {
					t.Errorf("output missing expected substring %q\nGot:\n%s", substr, output)
				}
			}
		})
	}
}

func TestDeferStmtLowering(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name: "basic defer with function call",
			input: `
fn cleanup() {
}

fn main() {
	defer cleanup()
}`,
			contains: []string{
				"defer_push call void @cleanup()",
				"defer_run_all",
				"ret void",
			},
		},
		{
			name: "defer with return statement",
			input: `
fn cleanup() {
}

fn main() {
	defer cleanup()
	return
}`,
			contains: []string{
				"defer_push call void @cleanup()",
				"defer_run_all",
				"ret void",
			},
		},
		{
			name: "multiple defers in LIFO order",
			input: `
fn first() {
}

fn second() {
}

fn main() {
	defer first()
	defer second()
}`,
			contains: []string{
				"defer_push call void @first()",
				"defer_push call void @second()",
				"defer_run_all",
			},
		},
		{
			name: "defer in conditional block",
			input: `
fn cleanup() {
}

fn main() {
	let x = 5
	if x > 0 {
		defer cleanup()
	}
}`,
			contains: []string{
				"defer_push call void @cleanup()",
				"defer_run_all",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			file := p.ParseFile()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			lowerer := NewLowerer()
			module := lowerer.LowerFile(file)

			var mainFunc *Function
			for _, fn := range module.Functions {
				if fn.Name == "main" {
					mainFunc = fn
					break
				}
			}

			if mainFunc == nil {
				t.Fatal("main function not found")
			}

			// Convert all blocks to string
			var output string
			for _, block := range mainFunc.Blocks {
				output += block.String()
			}

			// Check for expected substrings
			for _, substr := range tt.contains {
				if !strings.Contains(output, substr) {
					t.Errorf("output missing expected substring %q\nGot:\n%s", substr, output)
				}
			}
		})
	}
}

func TestPropagateExprLowering(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		contains    []string
		blockCount  int
		blockLabels []string
	}{
		{
			name: "basic ? operator generates control flow",
			input: `
fn may_fail() i32 {
	return 42
}

fn main() {
	let x = may_fail()?
}`,
			contains: []string{
				"call i32 @may_fail",
				"label %bb_check",
				"label %bb_error",
				"label %bb_ok",
				"br label %bb_check",
			},
			blockCount: 4, // entry, check, error, ok
			blockLabels: []string{"entry", "check", "error", "ok"},
		},
		{
			name: "? operator with early return",
			input: `
fn may_fail() i32 {
	return 1
}

fn caller() i32 {
	let result = may_fail()?
	return result
}`,
			contains: []string{
				"call i32 @may_fail",
				"label %bb_check",
				"label %bb_error",
				"label %bb_ok",
				"ret i32",
			},
			blockCount: 4, // entry, check, error, ok
		},
		{
			name: "? operator extracts value on success path",
			input: `
fn get_value() i32 {
	return 10
}

fn use_value() {
	let val = get_value()?
	let doubled = val + val
}`,
			contains: []string{
				"call i32 @get_value",
				"label %bb_check",
				"label %bb_error",
				"label %bb_ok",
				"add i32",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			file := p.ParseFile()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			lowerer := NewLowerer()
			module := lowerer.LowerFile(file)

			// Find the function that uses ? operator
			var testFunc *Function
			for _, fn := range module.Functions {
				if fn.Name == "main" || fn.Name == "caller" || fn.Name == "use_value" {
					testFunc = fn
					break
				}
			}

			if testFunc == nil {
				t.Fatal("test function not found")
			}

			// Check block count
			if tt.blockCount > 0 && len(testFunc.Blocks) != tt.blockCount {
				t.Errorf("expected %d blocks, got %d", tt.blockCount, len(testFunc.Blocks))
			}

			// Check block labels if specified
			if len(tt.blockLabels) > 0 {
				for i, expectedLabel := range tt.blockLabels {
					if i >= len(testFunc.Blocks) {
						t.Errorf("expected block %d with label containing %q, but only %d blocks exist", i, expectedLabel, len(testFunc.Blocks))
						continue
					}
					if !strings.Contains(testFunc.Blocks[i].Label, expectedLabel) {
						t.Errorf("expected block %d label to contain %q, got %q", i, expectedLabel, testFunc.Blocks[i].Label)
					}
				}
			}

			// Convert all blocks to string
			var output string
			for _, block := range testFunc.Blocks {
				output += block.String()
			}

			// Check for expected substrings
			for _, substr := range tt.contains {
				if !strings.Contains(output, substr) {
					t.Errorf("output missing expected substring %q\nGot:\n%s", substr, output)
				}
			}
		})
	}
}
