package types

import "testing"

func TestTypeEnv(t *testing.T) {
	env := NewEnv()

	// Define a variable
	i32 := &PrimitiveType{Name: "i32", Kind: Int32}
	env.Define("x", i32, false)

	// Look up variable
	typ, mut, ok := env.Lookup("x")
	if !ok {
		t.Fatal("expected to find x")
	}

	if !TypesEqual(typ, i32) {
		t.Errorf("expected i32, got %s", typ.String())
	}

	if mut {
		t.Error("expected x to be immutable")
	}

	// Test scopes
	env.PushScope()
	env.Define("y", i32, true)

	_, mut2, ok2 := env.Lookup("y")
	if !ok2 {
		t.Fatal("expected to find y")
	}

	if !mut2 {
		t.Error("expected y to be mutable")
	}

	env.PopScope()

	// y should not be visible
	_, _, ok3 := env.Lookup("y")
	if ok3 {
		t.Error("expected y to be out of scope")
	}
}
