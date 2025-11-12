package types

import "testing"

func TestPrimitiveTypes(t *testing.T) {
	// Test primitive type creation
	i32 := &PrimitiveType{Name: "i32", Kind: Int32}
	if i32.String() != "i32" {
		t.Errorf("expected i32, got %s", i32.String())
	}

	// Test type equality
	i32_2 := &PrimitiveType{Name: "i32", Kind: Int32}
	if !TypesEqual(i32, i32_2) {
		t.Error("expected i32 types to be equal")
	}

	// Test type inequality
	i64 := &PrimitiveType{Name: "i64", Kind: Int64}
	if TypesEqual(i32, i64) {
		t.Error("expected i32 and i64 to be different")
	}
}
