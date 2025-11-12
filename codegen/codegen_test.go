package codegen

import (
	"testing"

	"github.com/yarlson/yarlang/mir"
)

func TestCodegenSimpleFunction(t *testing.T) {
	// Create MIR for: fn add(a i32, b i32) i32 { return a + b }
	mirFn := &mir.Function{
		Name: "add",
		Params: []mir.Param{
			{Name: "a", Type: &mir.PrimitiveType{Name: "i32"}},
			{Name: "b", Type: &mir.PrimitiveType{Name: "i32"}},
		},
		RetTy: &mir.PrimitiveType{Name: "i32"},
		Blocks: []*mir.BasicBlock{
			{
				Label: "entry",
				Instrs: []mir.Instruction{
					// Simplified: just return first param for test
					&mir.Ret{Value: "a", Type: &mir.PrimitiveType{Name: "i32"}},
				},
			},
		},
	}

	cg := NewCodegen()
	llvmMod := cg.GenModule(&mir.Module{Functions: []*mir.Function{mirFn}})

	if llvmMod == nil {
		t.Fatal("expected LLVM module, got nil")
	}

	// Verify function was generated
	if len(llvmMod.Funcs) != 1 {
		t.Fatalf("expected 1 function, got %d", len(llvmMod.Funcs))
	}
}

func TestCodegenLoadsFunctionParameter(t *testing.T) {
	i32 := &mir.PrimitiveType{Name: "i32"}
	mirFn := &mir.Function{
		Name: "id",
		Params: []mir.Param{{Name: "n", Type: i32}},
		RetTy:  i32,
		Blocks: []*mir.BasicBlock{
			{
				Label: "entry",
				Instrs: []mir.Instruction{
					&mir.Load{Dest: "tmp", Source: "n", Type: i32},
					&mir.Ret{Value: "tmp", Type: i32},
				},
			},
		},
	}

	cg := NewCodegen()
	llvmMod := cg.GenModule(&mir.Module{Functions: []*mir.Function{mirFn}})

	if llvmMod == nil {
		t.Fatal("expected LLVM module, got nil")
	}

	moduleIR := llvmMod.String()
	t.Logf("Generated IR:\n%s", moduleIR)

	if containsString(moduleIR, "load i32, <nil>") {
		t.Fatalf("expected loads from parameters to reference stack slots, got IR:\n%s", moduleIR)
	}

	if !containsString(moduleIR, "store i32 %n") {
		t.Fatalf("expected parameter value to be stored to a stack slot, got IR:\n%s", moduleIR)
	}
}

func TestCodegenStoreWithConstant(t *testing.T) {
	// Create MIR for: fn test() i32 { %x = alloca i32; store i32 42, i32* %x; %y = load i32, i32* %x; ret i32 %y }
	mirFn := &mir.Function{
		Name:   "test",
		Params: []mir.Param{},
		RetTy:  &mir.PrimitiveType{Name: "i32"},
		Blocks: []*mir.BasicBlock{
			{
				Label: "entry",
				Instrs: []mir.Instruction{
					&mir.Alloca{Name: "x", Type: &mir.PrimitiveType{Name: "i32"}},
					&mir.Store{Value: "42", Dest: "x", Type: &mir.PrimitiveType{Name: "i32"}},
					&mir.Load{Dest: "y", Source: "x", Type: &mir.PrimitiveType{Name: "i32"}},
					&mir.Ret{Value: "y", Type: &mir.PrimitiveType{Name: "i32"}},
				},
			},
		},
	}

	cg := NewCodegen()
	llvmMod := cg.GenModule(&mir.Module{Functions: []*mir.Function{mirFn}})

	if llvmMod == nil {
		t.Fatal("expected LLVM module, got nil")
	}

	// Verify function was generated
	if len(llvmMod.Funcs) != 1 {
		t.Fatalf("expected 1 function, got %d", len(llvmMod.Funcs))
	}

	// Print the full module IR for debugging
	moduleIR := llvmMod.String()
	t.Logf("Generated IR:\n%s", moduleIR)

	// Check that the store instruction uses 42, not 0
	if !containsString(moduleIR, "store i32 42") {
		t.Errorf("expected 'store i32 42' in generated IR, got:\n%s", moduleIR)
	}

	// Make sure it doesn't use the wrong value (0)
	if containsString(moduleIR, "store i32 0") {
		t.Errorf("expected 'store i32 42', but found 'store i32 0' in generated IR:\n%s", moduleIR)
	}
}

func TestCodegenStoreWithVariable(t *testing.T) {
	// Create MIR for: fn test() i32 { %x = alloca i32; store i32 100, i32* %x; %y = alloca i32; %z = load i32, i32* %x; store i32 %z, i32* %y; %result = load i32, i32* %y; ret i32 %result }
	// This tests storing a value from another variable (not just a constant)
	mirFn := &mir.Function{
		Name:   "test",
		Params: []mir.Param{},
		RetTy:  &mir.PrimitiveType{Name: "i32"},
		Blocks: []*mir.BasicBlock{
			{
				Label: "entry",
				Instrs: []mir.Instruction{
					&mir.Alloca{Name: "x", Type: &mir.PrimitiveType{Name: "i32"}},
					&mir.Store{Value: "100", Dest: "x", Type: &mir.PrimitiveType{Name: "i32"}},
					&mir.Alloca{Name: "y", Type: &mir.PrimitiveType{Name: "i32"}},
					&mir.Load{Dest: "z", Source: "x", Type: &mir.PrimitiveType{Name: "i32"}},
					&mir.Store{Value: "z", Dest: "y", Type: &mir.PrimitiveType{Name: "i32"}},
					&mir.Load{Dest: "result", Source: "y", Type: &mir.PrimitiveType{Name: "i32"}},
					&mir.Ret{Value: "result", Type: &mir.PrimitiveType{Name: "i32"}},
				},
			},
		},
	}

	cg := NewCodegen()
	llvmMod := cg.GenModule(&mir.Module{Functions: []*mir.Function{mirFn}})

	if llvmMod == nil {
		t.Fatal("expected LLVM module, got nil")
	}

	moduleIR := llvmMod.String()
	t.Logf("Generated IR:\n%s", moduleIR)

	// Check that the first store uses constant 100
	if !containsString(moduleIR, "store i32 100") {
		t.Errorf("expected 'store i32 100' in generated IR")
	}

	// Check that we have a load for z from x
	if !containsString(moduleIR, "%z = load i32") {
		t.Errorf("expected load of %%z in generated IR")
	}

	// Check that the second store uses the loaded value z, not a constant 0
	// After loading %z, we should store it to %y
	// The store should reference %z as the value
	if containsString(moduleIR, "store i32 0, i32* %y") {
		t.Errorf("should not store constant 0 to %%y, should use value from %%z")
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findInString(s, substr))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestCodegenBinOp(t *testing.T) {
	// Create MIR for: fn test() i32 { %x = add i32 10, 20; ret i32 %x }
	mirFn := &mir.Function{
		Name:   "test",
		Params: []mir.Param{},
		RetTy:  &mir.PrimitiveType{Name: "i32"},
		Blocks: []*mir.BasicBlock{
			{
				Label: "entry",
				Instrs: []mir.Instruction{
					&mir.BinOp{Dest: "x", Op: mir.Add, Left: "10", Right: "20", Type: &mir.PrimitiveType{Name: "i32"}},
					&mir.Ret{Value: "x", Type: &mir.PrimitiveType{Name: "i32"}},
				},
			},
		},
	}

	cg := NewCodegen()
	llvmMod := cg.GenModule(&mir.Module{Functions: []*mir.Function{mirFn}})

	if llvmMod == nil {
		t.Fatal("expected LLVM module, got nil")
	}

	moduleIR := llvmMod.String()
	t.Logf("Generated IR:\n%s", moduleIR)

	// Check that the add instruction is generated correctly
	if !containsString(moduleIR, "%x = add i32 10, 20") {
		t.Errorf("expected '%%x = add i32 10, 20' in generated IR")
	}

	// Check that return uses the result
	if !containsString(moduleIR, "ret i32 %x") {
		t.Errorf("expected 'ret i32 %%x' in generated IR")
	}
}

func TestCodegenBinOpComparison(t *testing.T) {
	// Create MIR for: fn test() bool { %x = eq i32 42, 42; ret bool %x }
	mirFn := &mir.Function{
		Name:   "test",
		Params: []mir.Param{},
		RetTy:  &mir.PrimitiveType{Name: "bool"},
		Blocks: []*mir.BasicBlock{
			{
				Label: "entry",
				Instrs: []mir.Instruction{
					&mir.BinOp{Dest: "x", Op: mir.Eq, Left: "42", Right: "42", Type: &mir.PrimitiveType{Name: "i32"}},
					&mir.Ret{Value: "x", Type: &mir.PrimitiveType{Name: "bool"}},
				},
			},
		},
	}

	cg := NewCodegen()
	llvmMod := cg.GenModule(&mir.Module{Functions: []*mir.Function{mirFn}})

	if llvmMod == nil {
		t.Fatal("expected LLVM module, got nil")
	}

	moduleIR := llvmMod.String()
	t.Logf("Generated IR:\n%s", moduleIR)

	// Check that the comparison instruction is generated correctly
	if !containsString(moduleIR, "%x = icmp eq i32 42, 42") {
		t.Errorf("expected '%%x = icmp eq i32 42, 42' in generated IR")
	}
}

func TestCodegenCall(t *testing.T) {
	// Create MIR for:
	// fn helper() i32 { ret i32 42 }
	// fn test() i32 { %x = call i32 @helper(); ret i32 %x }
	helperFn := &mir.Function{
		Name:   "helper",
		Params: []mir.Param{},
		RetTy:  &mir.PrimitiveType{Name: "i32"},
		Blocks: []*mir.BasicBlock{
			{
				Label: "entry",
				Instrs: []mir.Instruction{
					&mir.Ret{Value: "42", Type: &mir.PrimitiveType{Name: "i32"}},
				},
			},
		},
	}

	testFn := &mir.Function{
		Name:   "test",
		Params: []mir.Param{},
		RetTy:  &mir.PrimitiveType{Name: "i32"},
		Blocks: []*mir.BasicBlock{
			{
				Label: "entry",
				Instrs: []mir.Instruction{
					&mir.Call{Dest: "x", Callee: "helper", Args: []string{}, RetTy: &mir.PrimitiveType{Name: "i32"}},
					&mir.Ret{Value: "x", Type: &mir.PrimitiveType{Name: "i32"}},
				},
			},
		},
	}

	cg := NewCodegen()
	llvmMod := cg.GenModule(&mir.Module{Functions: []*mir.Function{helperFn, testFn}})

	if llvmMod == nil {
		t.Fatal("expected LLVM module, got nil")
	}

	moduleIR := llvmMod.String()
	t.Logf("Generated IR:\n%s", moduleIR)

	// Check that the call instruction is generated correctly
	if !containsString(moduleIR, "%x = call i32 @helper()") {
		t.Errorf("expected '%%x = call i32 @helper()' in generated IR")
	}
}

func TestCodegenBranch(t *testing.T) {
	// Create MIR for: fn test() i32 { br label %end; end: ret i32 42 }
	mirFn := &mir.Function{
		Name:   "test",
		Params: []mir.Param{},
		RetTy:  &mir.PrimitiveType{Name: "i32"},
		Blocks: []*mir.BasicBlock{
			{
				Label: "entry",
				Instrs: []mir.Instruction{
					&mir.Br{Label: "end"},
				},
			},
			{
				Label: "end",
				Instrs: []mir.Instruction{
					&mir.Ret{Value: "42", Type: &mir.PrimitiveType{Name: "i32"}},
				},
			},
		},
	}

	cg := NewCodegen()
	llvmMod := cg.GenModule(&mir.Module{Functions: []*mir.Function{mirFn}})

	if llvmMod == nil {
		t.Fatal("expected LLVM module, got nil")
	}

	moduleIR := llvmMod.String()
	t.Logf("Generated IR:\n%s", moduleIR)

	// Check that the branch instruction is generated correctly
	if !containsString(moduleIR, "br label %end") {
		t.Errorf("expected 'br label %%end' in generated IR")
	}

	// Check that the end block exists
	if !containsString(moduleIR, "end:") {
		t.Errorf("expected 'end:' label in generated IR")
	}
}

func TestCodegenCondBranch(t *testing.T) {
	// Create MIR for: fn test(cond bool) i32 { br i1 %cond, label %then, label %else; then: ret i32 1; else: ret i32 0 }
	mirFn := &mir.Function{
		Name: "test",
		Params: []mir.Param{
			{Name: "cond", Type: &mir.PrimitiveType{Name: "bool"}},
		},
		RetTy: &mir.PrimitiveType{Name: "i32"},
		Blocks: []*mir.BasicBlock{
			{
				Label: "entry",
				Instrs: []mir.Instruction{
					&mir.CondBr{Cond: "cond", TrueLabel: "then", FalseLabel: "else"},
				},
			},
			{
				Label: "then",
				Instrs: []mir.Instruction{
					&mir.Ret{Value: "1", Type: &mir.PrimitiveType{Name: "i32"}},
				},
			},
			{
				Label: "else",
				Instrs: []mir.Instruction{
					&mir.Ret{Value: "0", Type: &mir.PrimitiveType{Name: "i32"}},
				},
			},
		},
	}

	cg := NewCodegen()
	llvmMod := cg.GenModule(&mir.Module{Functions: []*mir.Function{mirFn}})

	if llvmMod == nil {
		t.Fatal("expected LLVM module, got nil")
	}

	moduleIR := llvmMod.String()
	t.Logf("Generated IR:\n%s", moduleIR)

	// Check that the conditional branch instruction is generated correctly
	if !containsString(moduleIR, "br i1 %cond, label %then, label %else") {
		t.Errorf("expected 'br i1 %%cond, label %%then, label %%else' in generated IR")
	}

	// Check that both blocks exist
	if !containsString(moduleIR, "then:") {
		t.Errorf("expected 'then:' label in generated IR")
	}
	if !containsString(moduleIR, "else:") {
		t.Errorf("expected 'else:' label in generated IR")
	}
}

func TestCodegenComprehensive(t *testing.T) {
	// Create MIR that exercises all instruction types:
	// fn compute(a i32, b i32) i32 {
	//   %x = add i32 %a, %b
	//   %cmp = lt i32 %x, 100
	//   br i1 %cmp, label %small, label %large
	// small:
	//   %y = mul i32 %x, 2
	//   br label %end
	// large:
	//   %z = sub i32 %x, 50
	//   br label %end
	// end:
	//   %result = alloca i32
	//   store i32 %y, i32* %result (or %z depending on path)
	//   %final = load i32, i32* %result
	//   ret i32 %final
	// }
	// Simplified version that's easier to verify:
	mirFn := &mir.Function{
		Name: "compute",
		Params: []mir.Param{
			{Name: "a", Type: &mir.PrimitiveType{Name: "i32"}},
			{Name: "b", Type: &mir.PrimitiveType{Name: "i32"}},
		},
		RetTy: &mir.PrimitiveType{Name: "i32"},
		Blocks: []*mir.BasicBlock{
			{
				Label: "entry",
				Instrs: []mir.Instruction{
					&mir.BinOp{Dest: "x", Op: mir.Add, Left: "a", Right: "b", Type: &mir.PrimitiveType{Name: "i32"}},
					&mir.BinOp{Dest: "cmp", Op: mir.Lt, Left: "x", Right: "100", Type: &mir.PrimitiveType{Name: "i32"}},
					&mir.CondBr{Cond: "cmp", TrueLabel: "small", FalseLabel: "large"},
				},
			},
			{
				Label: "small",
				Instrs: []mir.Instruction{
					&mir.BinOp{Dest: "y", Op: mir.Mul, Left: "x", Right: "2", Type: &mir.PrimitiveType{Name: "i32"}},
					&mir.Ret{Value: "y", Type: &mir.PrimitiveType{Name: "i32"}},
				},
			},
			{
				Label: "large",
				Instrs: []mir.Instruction{
					&mir.BinOp{Dest: "z", Op: mir.Sub, Left: "x", Right: "50", Type: &mir.PrimitiveType{Name: "i32"}},
					&mir.Ret{Value: "z", Type: &mir.PrimitiveType{Name: "i32"}},
				},
			},
		},
	}

	cg := NewCodegen()
	llvmMod := cg.GenModule(&mir.Module{Functions: []*mir.Function{mirFn}})

	if llvmMod == nil {
		t.Fatal("expected LLVM module, got nil")
	}

	moduleIR := llvmMod.String()
	t.Logf("Generated IR:\n%s", moduleIR)

	// Verify all instruction types are present
	tests := []string{
		"%x = add i32 %a, %b",                     // BinOp: Add
		"%cmp = icmp slt i32 %x, 100",             // BinOp: Lt (comparison)
		"br i1 %cmp, label %small, label %large", // CondBr
		"%y = mul i32 %x, 2",                      // BinOp: Mul
		"%z = sub i32 %x, 50",                     // BinOp: Sub
		"ret i32 %y",                              // Ret
		"ret i32 %z",                              // Ret
	}

	for _, expected := range tests {
		if !containsString(moduleIR, expected) {
			t.Errorf("expected '%s' in generated IR", expected)
		}
	}
}

func TestCodegenStringConstant(t *testing.T) {
	// Test that string constants are generated as global variables
	// MIR module with a global string
	mirMod := &mir.Module{
		Globals: []mir.Global{
			&mir.GlobalString{Name: ".str.0", Value: "hello"},
		},
		Functions: []*mir.Function{},
	}

	cg := NewCodegen()
	llvmMod := cg.GenModule(mirMod)

	if llvmMod == nil {
		t.Fatal("expected LLVM module, got nil")
	}

	moduleIR := llvmMod.String()
	t.Logf("Generated IR:\n%s", moduleIR)

	// Check that global string constant was created
	// LLVM format: @.str.0 = private unnamed_addr constant [6 x i8] c"hello\00"
	if !containsString(moduleIR, "@.str.0") {
		t.Errorf("expected global string constant @.str.0 in generated IR")
	}
	if !containsString(moduleIR, "hello") {
		t.Errorf("expected string content 'hello' in generated IR")
	}
}

func TestCodegenStringInCall(t *testing.T) {
	// Test that string constants can be passed to function calls
	// MIR for: fn main() { call void @println(i8* @.str.0) }
	mirMod := &mir.Module{
		Globals: []mir.Global{
			&mir.GlobalString{Name: ".str.0", Value: "hello"},
		},
		Functions: []*mir.Function{
			{
				Name:   "main",
				Params: []mir.Param{},
				RetTy:  &mir.PrimitiveType{Name: "void"},
				Blocks: []*mir.BasicBlock{
					{
						Label: "entry",
						Instrs: []mir.Instruction{
							&mir.Call{
								Dest:   "",
								Callee: "println",
								Args:   []string{"@.str.0"},
								RetTy:  &mir.PrimitiveType{Name: "void"},
							},
							&mir.Ret{Type: &mir.PrimitiveType{Name: "void"}},
						},
					},
				},
			},
		},
	}

	cg := NewCodegen()
	llvmMod := cg.GenModule(mirMod)

	if llvmMod == nil {
		t.Fatal("expected LLVM module, got nil")
	}

	moduleIR := llvmMod.String()
	t.Logf("Generated IR:\n%s", moduleIR)

	// Check that global string constant was created
	if !containsString(moduleIR, "@.str.0") {
		t.Errorf("expected global string constant @.str.0 in generated IR")
	}

	// Check that the call uses the string reference
	// The exact format depends on how we handle string arguments
	if !containsString(moduleIR, "call void @println") {
		t.Errorf("expected 'call void @println' in generated IR")
	}
}

func TestCodegenPrintlnInt(t *testing.T) {
	i32 := &mir.PrimitiveType{Name: "i32"}
	void := &mir.PrimitiveType{Name: "void"}
	mirFn := &mir.Function{
		Name:   "main",
		Params: []mir.Param{},
		RetTy:  void,
		Blocks: []*mir.BasicBlock{
			{
				Label: "entry",
				Instrs: []mir.Instruction{
					&mir.Alloca{Name: "result", Type: i32},
					&mir.Store{Value: "55", Dest: "result", Type: i32},
					&mir.Load{Dest: "tmp", Source: "result", Type: i32},
					&mir.Call{Dest: "", Callee: "println", Args: []string{"tmp"}, RetTy: void},
					&mir.Ret{Type: void},
				},
			},
		},
	}

	cg := NewCodegen()
	llvmMod := cg.GenModule(&mir.Module{Functions: []*mir.Function{mirFn}})

	moduleIR := llvmMod.String()
	if !containsString(moduleIR, "call void @println_i32") {
		t.Fatalf("expected int println call to lower to @println_i32, got:\n%s", moduleIR)
	}

	if containsString(moduleIR, "declare void @println(i8*") {
		t.Fatalf("println string declaration should not be emitted for pure int usage:\n%s", moduleIR)
	}
}

func TestCodegenPrintlnBool(t *testing.T) {
	i32 := &mir.PrimitiveType{Name: "i32"}
	void := &mir.PrimitiveType{Name: "void"}
	mirFn := &mir.Function{
		Name:   "main",
		Params: []mir.Param{},
		RetTy:  void,
		Blocks: []*mir.BasicBlock{
			{
				Label: "entry",
				Instrs: []mir.Instruction{
					&mir.BinOp{Dest: "cmp", Op: mir.Eq, Left: "1", Right: "1", Type: i32},
					&mir.Call{Dest: "", Callee: "println", Args: []string{"cmp"}, RetTy: void},
					&mir.Ret{Type: void},
				},
			},
		},
	}

	cg := NewCodegen()
	llvmMod := cg.GenModule(&mir.Module{Functions: []*mir.Function{mirFn}})
	moduleIR := llvmMod.String()

	if !containsString(moduleIR, "call void @println_bool(i1 %cmp)") {
		t.Fatalf("expected bool println call to lower to @println_bool, got:\n%s", moduleIR)
	}

	if containsString(moduleIR, "call void @println(i1") {
		t.Fatalf("bool println should not call the string variant, got:\n%s", moduleIR)
	}

	if containsString(moduleIR, "declare void @println_i32") {
		t.Fatalf("bool-only module should not declare println_i32, got:\n%s", moduleIR)
	}
}
