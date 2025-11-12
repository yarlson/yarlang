package ast

import "testing"

func TestTypeNodes(t *testing.T) {
	// Type path: i32
	tp := &TypePath{Path: []string{"i32"}}
	if tp.String() != "i32" {
		t.Errorf("wrong string: %s", tp.String())
	}

	// Reference type: &mut T
	rt := &RefType{Mut: true, Elem: tp}
	if rt.String() != "&mut i32" {
		t.Errorf("wrong string: %s", rt.String())
	}

	// Slice type: []T
	st := &SliceType{Elem: tp}
	if st.String() != "[]i32" {
		t.Errorf("wrong string: %s", st.String())
	}
}

func TestDeclNodes(t *testing.T) {
	// Function: fn add(a i32, b i32) i32
	fn := &FuncDecl{
		Name: "add",
		Params: []Param{
			{Name: "a", Type: &TypePath{Path: []string{"i32"}}},
			{Name: "b", Type: &TypePath{Path: []string{"i32"}}},
		},
		ReturnType: &TypePath{Path: []string{"i32"}},
		Body:       &Block{Stmts: []Stmt{}},
	}

	if fn.Name != "add" {
		t.Errorf("wrong name: %s", fn.Name)
	}
}

func TestStmtNodes(t *testing.T) {
	// ShortDecl: x := 42
	sd := &ShortDecl{
		Name:  "x",
		Value: &IntLit{Value: "42"},
	}
	if sd.String() != "x := 42" {
		t.Errorf("wrong string: %s", sd.String())
	}

	// ConstStmt: const MAX: i32 = 100
	cs := &ConstStmt{
		Name:  "MAX",
		Type:  &TypePath{Path: []string{"i32"}},
		Value: &IntLit{Value: "100"},
	}

	expected := "const MAX: i32 = 100"
	if cs.String() != expected {
		t.Errorf("wrong string: expected=%s, got=%s", expected, cs.String())
	}
}
