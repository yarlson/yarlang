package mir

import (
	"testing"

	"github.com/yarlson/yarlang/checker"
	"github.com/yarlson/yarlang/lexer"
	"github.com/yarlson/yarlang/parser"
)

func TestLowerFunction(t *testing.T) {
	input := `fn add(a i32, b i32) i32 {
		return a + b
	}`

	l := lexer.New(input)
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	c := checker.NewChecker()

	err := c.CheckFile(file)
	if err != nil {
		t.Fatalf("checker error: %v", err)
	}

	lower := NewLowerer()
	mod := lower.LowerFile(file)

	if len(mod.Functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(mod.Functions))
	}

	fn := mod.Functions[0]
	if fn.Name != "add" {
		t.Errorf("wrong function name: %s", fn.Name)
	}

	if len(fn.Params) != 2 {
		t.Errorf("expected 2 params, got %d", len(fn.Params))
	}
}

func TestLowerStringLiteral(t *testing.T) {
	input := `fn main() {
		println("hello")
	}`

	l := lexer.New(input)
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	c := checker.NewChecker()

	err := c.CheckFile(file)
	if err != nil {
		t.Fatalf("checker error: %v", err)
	}

	lower := NewLowerer()
	mod := lower.LowerFile(file)

	// Check that a global string constant was created
	if len(mod.Globals) != 1 {
		t.Fatalf("expected 1 global, got %d", len(mod.Globals))
	}

	globalStr, ok := mod.Globals[0].(*GlobalString)
	if !ok {
		t.Fatalf("expected GlobalString, got %T", mod.Globals[0])
	}

	if globalStr.Name != ".str.1" {
		t.Errorf("expected global name .str.1, got %s", globalStr.Name)
	}

	if globalStr.Value != "hello" {
		t.Errorf("expected global value 'hello', got %s", globalStr.Value)
	}

	// Check that the function contains a call with the global reference
	if len(mod.Functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(mod.Functions))
	}

	fn := mod.Functions[0]
	if fn.Name != "main" {
		t.Errorf("wrong function name: %s", fn.Name)
	}

	// Find the call instruction
	var callInstr *Call
	for _, block := range fn.Blocks {
		for _, instr := range block.Instrs {
			if call, ok := instr.(*Call); ok {
				callInstr = call
				break
			}
		}
	}

	if callInstr == nil {
		t.Fatal("expected Call instruction")
	}

	if callInstr.Callee != "println" {
		t.Errorf("expected callee 'println', got %s", callInstr.Callee)
	}

	if len(callInstr.Args) != 1 {
		t.Fatalf("expected 1 argument, got %d", len(callInstr.Args))
	}

	if callInstr.Args[0] != "@.str.1" {
		t.Errorf("expected argument '@.str.1', got %s", callInstr.Args[0])
	}
}
