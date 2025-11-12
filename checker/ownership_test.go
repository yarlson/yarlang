package checker

import (
	"strings"
	"testing"

	"github.com/yarlson/yarlang/lexer"
	"github.com/yarlson/yarlang/parser"
)

func TestMoveSemantics(t *testing.T) {
	tests := []struct {
		input     string
		shouldErr bool
		errMsg    string
	}{
		{
			// Move primitive (Copy type) - OK
			`fn main() { let x = 5; let y = x; let z = x }`,
			false,
			"",
		},
		{
			// Move struct (Move type) - ERROR on second use
			`struct Point { x: i32, y: i32 }
fn main() { let s = Point{x: 1, y: 2}; let a = s; let b = s }`,
			true,
			"use of moved value",
		},
		{
			// Borrow - OK
			`struct Point { x: i32, y: i32 }
fn main() { let s = Point{x: 1, y: 2}; let a = &s; let b = &s }`,
			false,
			"",
		},
		{
			// Same variable name in different scopes - OK (scope-aware move tracking)
			// This tests that the move tracking is scope-aware and doesn't confuse
			// variables with the same name in different scopes
			`struct Point { x: i32, y: i32 }
fn main() {
	let s = Point{x: 1, y: 2};
	let a = s;
	let s = Point{x: 3, y: 4};
	let b = s;
}`,
			false,
			"",
		},
		{
			// Move semantics within same scope should still error
			`struct Point { x: i32, y: i32 }
fn main() {
	let s = Point{x: 1, y: 2};
	let a = s;
	let b = s;
}`,
			true,
			"use of moved value",
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		file := p.ParseFile()

		if len(p.Errors()) != 0 {
			t.Fatalf("parser errors: %v", p.Errors())
		}

		c := NewChecker()
		err := c.CheckFile(file)

		if tt.shouldErr {
			if err == nil {
				t.Errorf("expected error for: %s", tt.input)
			} else if !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error for: %s\nerror: %v", tt.input, err)
			}
		}
	}
}

func TestBorrowChecking(t *testing.T) {
	tests := []struct {
		input     string
		shouldErr bool
		errMsg    string
	}{
		{
			// Multiple shared borrows - OK
			`fn main() { let x = 5; let a = &x; let b = &x }`,
			false,
			"",
		},
		{
			// Exclusive borrow + shared borrow - ERROR
			`fn main() { let mut x = 5; let a = &mut x; let b = &x }`,
			true,
			"cannot borrow",
		},
		{
			// Multiple exclusive borrows - ERROR
			`fn main() { let mut x = 5; let a = &mut x; let b = &mut x }`,
			true,
			"cannot borrow",
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		file := p.ParseFile()

		c := NewChecker()
		err := c.CheckFile(file)

		if tt.shouldErr {
			if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("expected error for: %s", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}
	}
}
