package checker

import (
	"testing"

	"github.com/yarlson/yarlang/lexer"
	"github.com/yarlson/yarlang/parser"
)

func TestGenerics(t *testing.T) {
	input := `
struct Vec<T> {
	data: *T,
	len: usize,
}

fn main() {
	let v: Vec<i32> = Vec{data: nil, len: 0}
}
`

	l := lexer.New(input)
	p := parser.New(l)
	file := p.ParseFile()

	c := NewChecker()

	err := c.CheckFile(file)
	if err != nil {
		t.Fatalf("checker error: %v", err)
	}
}
