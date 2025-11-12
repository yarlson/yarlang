package types

// Symbol represents a variable or function
type Symbol struct {
	Name string
	Type Type
	Mut  bool // Mutable?
}

// Scope represents a lexical scope
type Scope struct {
	symbols map[string]*Symbol
	parent  *Scope
}

func NewScope(parent *Scope) *Scope {
	return &Scope{
		symbols: make(map[string]*Symbol),
		parent:  parent,
	}
}

func (s *Scope) Define(name string, typ Type, mut bool) {
	s.symbols[name] = &Symbol{Name: name, Type: typ, Mut: mut}
}

func (s *Scope) Lookup(name string) (*Symbol, bool) {
	if sym, ok := s.symbols[name]; ok {
		return sym, true
	}

	if s.parent != nil {
		return s.parent.Lookup(name)
	}

	return nil, false
}

// Env is the type environment
type Env struct {
	currentScope *Scope
	typeVarID    int // Counter for type variables
}

func NewEnv() *Env {
	// Create root scope with builtins
	root := NewScope(nil)

	// Define primitive types
	builtins := map[string]Type{
		"i8":    &PrimitiveType{Name: "i8", Kind: Int8},
		"i16":   &PrimitiveType{Name: "i16", Kind: Int16},
		"i32":   &PrimitiveType{Name: "i32", Kind: Int32},
		"i64":   &PrimitiveType{Name: "i64", Kind: Int64},
		"isize": &PrimitiveType{Name: "isize", Kind: ISize},
		"u8":    &PrimitiveType{Name: "u8", Kind: UInt8},
		"u16":   &PrimitiveType{Name: "u16", Kind: UInt16},
		"u32":   &PrimitiveType{Name: "u32", Kind: UInt32},
		"u64":   &PrimitiveType{Name: "u64", Kind: UInt64},
		"usize": &PrimitiveType{Name: "usize", Kind: USize},
		"f32":   &PrimitiveType{Name: "f32", Kind: Float32},
		"f64":   &PrimitiveType{Name: "f64", Kind: Float64},
		"bool":  &PrimitiveType{Name: "bool", Kind: Bool},
		"char":  &PrimitiveType{Name: "char", Kind: Char},
		"void":  &PrimitiveType{Name: "void", Kind: Void},
	}

	for name, typ := range builtins {
		root.Define(name, typ, false)
	}

	// Define builtin functions
	voidType := &PrimitiveType{Name: "void", Kind: Void}
	stringType := &SliceType{Elem: &PrimitiveType{Name: "u8", Kind: UInt8}}

	// println(msg string) - accepts any type for now (variadic-like)
	env := &Env{currentScope: root, typeVarID: 0}
	anyType := env.NewTypeVar()
	root.Define("println", &FuncType{
		Params: []Type{anyType},
		Return: voidType,
	}, false)

	// panic(msg string)
	root.Define("panic", &FuncType{
		Params: []Type{stringType},
		Return: voidType,
	}, false)

	// len<T>(arr []T) usize
	usizeType := &PrimitiveType{Name: "usize", Kind: USize}
	sliceType := &SliceType{Elem: env.NewTypeVar()}
	root.Define("len", &FuncType{
		Params: []Type{sliceType},
		Return: usizeType,
	}, false)

	return env
}

func (e *Env) Define(name string, typ Type, mut bool) {
	e.currentScope.Define(name, typ, mut)
}

func (e *Env) Lookup(name string) (Type, bool, bool) {
	sym, ok := e.currentScope.Lookup(name)
	if !ok {
		return nil, false, false
	}

	return sym.Type, sym.Mut, true
}

func (e *Env) LookupSymbol(name string) (*Symbol, bool) {
	return e.currentScope.Lookup(name)
}

func (e *Env) PushScope() {
	e.currentScope = NewScope(e.currentScope)
}

func (e *Env) PopScope() {
	if e.currentScope.parent != nil {
		e.currentScope = e.currentScope.parent
	}
}

func (e *Env) NewTypeVar() *TypeVar {
	e.typeVarID++
	return &TypeVar{ID: e.typeVarID}
}
