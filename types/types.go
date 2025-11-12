package types

import "fmt"

// Type represents a YarLang type
type Type interface {
	String() string
	isType()
}

// TypeKind represents primitive type kinds
type TypeKind int

const (
	Int8 TypeKind = iota
	Int16
	Int32
	Int64
	ISize
	UInt8
	UInt16
	UInt32
	UInt64
	USize
	Float32
	Float64
	Bool
	Char
	Void
)

// PrimitiveType represents built-in types
type PrimitiveType struct {
	Name string
	Kind TypeKind
}

func (p *PrimitiveType) isType()        {}
func (p *PrimitiveType) String() string { return p.Name }

// RefType represents &T and &mut T
type RefType struct {
	Mut  bool
	Elem Type
}

func (r *RefType) isType() {}
func (r *RefType) String() string {
	if r.Mut {
		return fmt.Sprintf("&mut %s", r.Elem.String())
	}

	return fmt.Sprintf("&%s", r.Elem.String())
}

// PtrType represents *T
type PtrType struct {
	Elem Type
}

func (p *PtrType) isType() {}
func (p *PtrType) String() string {
	return fmt.Sprintf("*%s", p.Elem.String())
}

// SliceType represents []T
type SliceType struct {
	Elem Type
}

func (s *SliceType) isType() {}
func (s *SliceType) String() string {
	return fmt.Sprintf("[]%s", s.Elem.String())
}

// ArrayType represents [T; N]
type ArrayType struct {
	Elem Type
	Len  int
}

func (a *ArrayType) isType() {}
func (a *ArrayType) String() string {
	return fmt.Sprintf("[%s; %d]", a.Elem.String(), a.Len)
}

// TupleType represents (T1, T2, ...)
type TupleType struct {
	Elems []Type
}

func (t *TupleType) isType() {}
func (t *TupleType) String() string {
	s := "("

	for i, e := range t.Elems {
		if i > 0 {
			s += ", "
		}

		s += e.String()
	}

	return s + ")"
}

// StructType represents user-defined structs
type StructType struct {
	Name    string
	Fields  map[string]Type
	TParams []string // Generic type parameters
}

func (s *StructType) isType() {}
func (s *StructType) String() string {
	if len(s.TParams) > 0 {
		return fmt.Sprintf("%s<%v>", s.Name, s.TParams)
	}

	return s.Name
}

// EnumType represents user-defined enums
type EnumType struct {
	Name     string
	Variants map[string][]Type // Variant name -> payload types
	TParams  []string
}

func (e *EnumType) isType() {}
func (e *EnumType) String() string {
	if len(e.TParams) > 0 {
		return fmt.Sprintf("%s<%v>", e.Name, e.TParams)
	}

	return e.Name
}

// FuncType represents function types
type FuncType struct {
	Params []Type
	Return Type
}

func (f *FuncType) isType() {}
func (f *FuncType) String() string {
	return fmt.Sprintf("fn(%v) %s", f.Params, f.Return.String())
}

// TypeVar represents a type variable for inference
type TypeVar struct {
	ID int
}

func (t *TypeVar) isType() {}
func (t *TypeVar) String() string {
	return fmt.Sprintf("?T%d", t.ID)
}

// TypesEqual checks if two types are equal
func TypesEqual(t1, t2 Type) bool {
	switch t1 := t1.(type) {
	case *PrimitiveType:
		t2, ok := t2.(*PrimitiveType)
		return ok && t1.Kind == t2.Kind
	case *RefType:
		t2, ok := t2.(*RefType)
		return ok && t1.Mut == t2.Mut && TypesEqual(t1.Elem, t2.Elem)
	case *PtrType:
		t2, ok := t2.(*PtrType)
		return ok && TypesEqual(t1.Elem, t2.Elem)
	case *SliceType:
		t2, ok := t2.(*SliceType)
		return ok && TypesEqual(t1.Elem, t2.Elem)
	case *ArrayType:
		t2, ok := t2.(*ArrayType)
		return ok && t1.Len == t2.Len && TypesEqual(t1.Elem, t2.Elem)
	case *TupleType:
		t2, ok := t2.(*TupleType)
		if !ok || len(t1.Elems) != len(t2.Elems) {
			return false
		}

		for i := range t1.Elems {
			if !TypesEqual(t1.Elems[i], t2.Elems[i]) {
				return false
			}
		}

		return true
	case *StructType:
		t2, ok := t2.(*StructType)
		return ok && t1.Name == t2.Name
	case *EnumType:
		t2, ok := t2.(*EnumType)
		return ok && t1.Name == t2.Name
	case *FuncType:
		t2, ok := t2.(*FuncType)
		if !ok || len(t1.Params) != len(t2.Params) {
			return false
		}

		for i := range t1.Params {
			if !TypesEqual(t1.Params[i], t2.Params[i]) {
				return false
			}
		}

		return TypesEqual(t1.Return, t2.Return)
	default:
		return false
	}
}

// IsCopy returns true if type is Copy (doesn't need move semantics)
func IsCopy(t Type) bool {
	switch t := t.(type) {
	case *PrimitiveType:
		return true // All primitives are Copy
	case *RefType:
		return true // References are Copy
	case *PtrType:
		return true // Raw pointers are Copy
	case *TupleType:
		// Tuple is Copy if all elements are Copy
		for _, elem := range t.Elems {
			if !IsCopy(elem) {
				return false
			}
		}

		return true
	default:
		return false // Structs, enums, arrays, slices are Move by default
	}
}
