package checker

import (
	"fmt"
	"testing"

	"github.com/yarlson/yarlang/lexer"
	"github.com/yarlson/yarlang/parser"
	"github.com/yarlson/yarlang/types"
)

func TestCheckLet(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "explicit type annotation",
			input: `
fn main() {
	let x: i32 = 5
}
`,
			wantErr: false,
		},
		{
			name: "type inference",
			input: `
fn main() {
	let y = 10
}
`,
			wantErr: false,
		},
		{
			name: "both explicit and inferred types",
			input: `
fn main() {
	let x: i32 = 5
	let y = 10
	let z: i32 = x + y
}
`,
			wantErr: false,
		},
		{
			name: "type mismatch should error",
			input: `
fn main() {
	let x: bool = 5
}
`,
			wantErr: true,
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

			c := NewChecker()
			err := c.CheckFile(file)

			if (err != nil) != tt.wantErr {
				t.Errorf("CheckFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestArrayLength(t *testing.T) {
	tests := []struct {
		name         string
		arrayType    string
		wantErr      bool
		wantLength   int
	}{
		{
			name:       "decimal array length",
			arrayType:  "[i32; 5]",
			wantErr:    false,
			wantLength: 5,
		},
		{
			name:       "hex array length",
			arrayType:  "[i32; 0x10]",
			wantErr:    false,
			wantLength: 16,
		},
		{
			name:       "binary array length",
			arrayType:  "[i32; 0b100]",
			wantErr:    false,
			wantLength: 4,
		},
		{
			name:       "octal array length",
			arrayType:  "[i32; 0o10]",
			wantErr:    false,
			wantLength: 8,
		},
		{
			name:       "large array length",
			arrayType:  "[i32; 1000]",
			wantErr:    false,
			wantLength: 1000,
		},
		{
			name:       "zero array length should error",
			arrayType:  "[i32; 0]",
			wantErr:    true,
			wantLength: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create input that uses the array type in a struct field
			// This lets us test type resolution without assignment issues
			input := fmt.Sprintf(`
struct Test {
	field: %s,
}
`, tt.arrayType)

			l := lexer.New(input)
			p := parser.New(l)
			file := p.ParseFile()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser errors: %v", p.Errors())
			}

			c := NewChecker()
			err := c.CheckFile(file)

			if (err != nil) != tt.wantErr {
				t.Errorf("CheckFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			// If we need to check the actual length, look up the struct type
			if !tt.wantErr && err == nil {
				typ, _, ok := c.env.Lookup("Test")
				if !ok {
					t.Fatalf("struct 'Test' not found in environment")
				}

				structType, ok := typ.(*types.StructType)
				if !ok {
					t.Fatalf("expected StructType, got %T", typ)
				}

				fieldType, ok := structType.Fields["field"]
				if !ok {
					t.Fatalf("field 'field' not found in struct")
				}

				arrType, ok := fieldType.(*types.ArrayType)
				if !ok {
					t.Fatalf("expected ArrayType, got %T", fieldType)
				}

				if arrType.Len != tt.wantLength {
					t.Errorf("array length = %d, want %d", arrType.Len, tt.wantLength)
				}
			}
		})
	}
}

func TestFunctionCallTypeChecking(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "correct function call",
			input: `
fn add(a i32, b i32) i32 {
	return a + b
}

fn main() {
	let x = add(1, 2)
}
`,
			wantErr: false,
		},
		{
			name: "correct return type usage",
			input: `
fn add(a i32, b i32) i32 {
	return a + b
}

fn main() {
	let x: i32 = add(1, 2)
}
`,
			wantErr: false,
		},
		{
			name: "argument count mismatch - too few",
			input: `
fn add(a i32, b i32) i32 {
	return a + b
}

fn main() {
	let x = add(1)
}
`,
			wantErr: true,
		},
		{
			name: "argument count mismatch - too many",
			input: `
fn add(a i32, b i32) i32 {
	return a + b
}

fn main() {
	let x = add(1, 2, 3)
}
`,
			wantErr: true,
		},
		{
			name: "argument type mismatch",
			input: `
fn add(a i32, b i32) i32 {
	return a + b
}

fn main() {
	let x = add(true, 2)
}
`,
			wantErr: true,
		},
		{
			name: "undefined function",
			input: `
fn main() {
	let x = undefined_func(1, 2)
}
`,
			wantErr: true,
		},
		{
			name: "void function call",
			input: `
fn do_something() {
	let x = 5
}

fn main() {
	do_something()
}
`,
			wantErr: false,
		},
		{
			name: "builtin println",
			input: `
fn main() {
	println("hello")
}
`,
			wantErr: false,
		},
		{
			name: "builtin panic",
			input: `
fn main() {
	panic("error")
}
`,
			wantErr: false,
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

			c := NewChecker()
			err := c.CheckFile(file)

			if (err != nil) != tt.wantErr {
				t.Errorf("CheckFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
