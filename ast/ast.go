package ast

import (
	"fmt"
	"strings"
)

// Node is the base interface for all AST nodes
type Node interface {
	String() string
}

// ===== Types =====

// Type represents a type expression
type Type interface {
	Node
	typeNode()
}

// TypePath represents a type path like i32, Vec<T>, std::io::File
type TypePath struct {
	Path []string // ["std", "io", "File"]
	Args []Type   // Generic arguments
}

func (t *TypePath) typeNode() {}
func (t *TypePath) String() string {
	s := strings.Join(t.Path, "::")
	if len(t.Args) > 0 {
		args := make([]string, len(t.Args))
		for i, a := range t.Args {
			args[i] = a.String()
		}

		s += "<" + strings.Join(args, ", ") + ">"
	}

	return s
}

// RefType represents &T or &mut T
type RefType struct {
	Mut  bool
	Elem Type
}

func (r *RefType) typeNode() {}
func (r *RefType) String() string {
	if r.Mut {
		return "&mut " + r.Elem.String()
	}

	return "&" + r.Elem.String()
}

// PtrType represents *T (unsafe raw pointer)
type PtrType struct {
	Elem Type
}

func (p *PtrType) typeNode() {}
func (p *PtrType) String() string {
	return "*" + p.Elem.String()
}

// SliceType represents []T
type SliceType struct {
	Elem Type
}

func (s *SliceType) typeNode() {}
func (s *SliceType) String() string {
	return "[]" + s.Elem.String()
}

// ArrayType represents [T; N]
type ArrayType struct {
	Elem Type
	Len  Expr
}

func (a *ArrayType) typeNode() {}
func (a *ArrayType) String() string {
	return fmt.Sprintf("[%s; %s]", a.Elem.String(), a.Len.String())
}

// TupleType represents (T1, T2, ...)
type TupleType struct {
	Elems []Type
}

func (t *TupleType) typeNode() {}
func (t *TupleType) String() string {
	elems := make([]string, len(t.Elems))
	for i, e := range t.Elems {
		elems[i] = e.String()
	}

	return "(" + strings.Join(elems, ", ") + ")"
}

// VoidType represents void
type VoidType struct{}

func (v *VoidType) typeNode() {}
func (v *VoidType) String() string {
	return "void"
}

// ===== Expressions =====

// Expr represents an expression
type Expr interface {
	Node
	exprNode()
}

// Ident represents an identifier
type Ident struct {
	Name string
}

func (i *Ident) exprNode() {}
func (i *Ident) String() string {
	return i.Name
}

// IntLit represents an integer literal
type IntLit struct {
	Value string // "123", "0xFF", etc.
}

func (i *IntLit) exprNode() {}
func (i *IntLit) String() string {
	return i.Value
}

// FloatLit represents a float literal
type FloatLit struct {
	Value string
}

func (f *FloatLit) exprNode() {}
func (f *FloatLit) String() string {
	return f.Value
}

// CharLit represents a char literal
type CharLit struct {
	Value string
}

func (c *CharLit) exprNode() {}
func (c *CharLit) String() string {
	return "'" + c.Value + "'"
}

// StringLit represents a string literal
type StringLit struct {
	Value string
}

func (s *StringLit) exprNode() {}
func (s *StringLit) String() string {
	return `"` + s.Value + `"`
}

// BoolLit represents true/false
type BoolLit struct {
	Value bool
}

func (b *BoolLit) exprNode() {}
func (b *BoolLit) String() string {
	if b.Value {
		return "true"
	}

	return "false"
}

// NilLit represents nil
type NilLit struct{}

func (n *NilLit) exprNode() {}
func (n *NilLit) String() string {
	return "nil"
}

// BinaryExpr represents binary operations
type BinaryExpr struct {
	Left  Expr
	Op    string
	Right Expr
}

func (b *BinaryExpr) exprNode() {}
func (b *BinaryExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", b.Left.String(), b.Op, b.Right.String())
}

// UnaryExpr represents unary operations
type UnaryExpr struct {
	Op   string
	Expr Expr
}

func (u *UnaryExpr) exprNode() {}
func (u *UnaryExpr) String() string {
	// Add space for &mut operator
	if u.Op == "&mut" {
		return fmt.Sprintf("(%s %s)", u.Op, u.Expr.String())
	}

	return fmt.Sprintf("(%s%s)", u.Op, u.Expr.String())
}

// CallExpr represents function calls
type CallExpr struct {
	Callee Expr
	Args   []Expr
}

func (c *CallExpr) exprNode() {}
func (c *CallExpr) String() string {
	args := make([]string, len(c.Args))
	for i, a := range c.Args {
		args[i] = a.String()
	}

	return fmt.Sprintf("%s(%s)", c.Callee.String(), strings.Join(args, ", "))
}

// IndexExpr represents array/slice indexing
type IndexExpr struct {
	Expr  Expr
	Index Expr
}

func (i *IndexExpr) exprNode() {}
func (i *IndexExpr) String() string {
	return fmt.Sprintf("%s[%s]", i.Expr.String(), i.Index.String())
}

// FieldExpr represents field access
type FieldExpr struct {
	Expr  Expr
	Field string
}

func (f *FieldExpr) exprNode() {}
func (f *FieldExpr) String() string {
	return fmt.Sprintf("%s.%s", f.Expr.String(), f.Field)
}

// PropagateExpr represents ? operator
type PropagateExpr struct {
	Expr Expr
}

func (p *PropagateExpr) exprNode() {}
func (p *PropagateExpr) String() string {
	return p.Expr.String() + "?"
}

// StructExpr represents struct literal
type StructExpr struct {
	Type  Type
	Inits []FieldInit
}

type FieldInit struct {
	Name string
	Val  Expr
}

func (s *StructExpr) exprNode() {}
func (s *StructExpr) String() string {
	inits := make([]string, len(s.Inits))
	for i, init := range s.Inits {
		inits[i] = fmt.Sprintf("%s: %s", init.Name, init.Val.String())
	}

	return fmt.Sprintf("%s{ %s }", s.Type.String(), strings.Join(inits, ", "))
}

// ArrayExpr represents array literal
type ArrayExpr struct {
	Elems []Expr
}

func (a *ArrayExpr) exprNode() {}
func (a *ArrayExpr) String() string {
	elems := make([]string, len(a.Elems))
	for i, e := range a.Elems {
		elems[i] = e.String()
	}

	return "[" + strings.Join(elems, ", ") + "]"
}

// TupleExpr represents tuple literal
type TupleExpr struct {
	Elems []Expr
}

func (t *TupleExpr) exprNode() {}
func (t *TupleExpr) String() string {
	elems := make([]string, len(t.Elems))
	for i, e := range t.Elems {
		elems[i] = e.String()
	}

	return "(" + strings.Join(elems, ", ") + ")"
}

// ===== Statements =====

// Stmt represents a statement
type Stmt interface {
	Node
	stmtNode()
}

// LetStmt represents let binding
type LetStmt struct {
	Mut   bool
	Name  string
	Type  Type // nil if inferred
	Value Expr
}

func (l *LetStmt) stmtNode() {}
func (l *LetStmt) String() string {
	mut := ""
	if l.Mut {
		mut = "mut "
	}

	typ := ""
	if l.Type != nil {
		typ = ": " + l.Type.String()
	}

	return fmt.Sprintf("let %s%s%s = %s", mut, l.Name, typ, l.Value.String())
}

// AssignStmt represents assignment
type AssignStmt struct {
	Target Expr
	Op     string // "=" or "+=", etc.
	Value  Expr
}

func (a *AssignStmt) stmtNode() {}
func (a *AssignStmt) String() string {
	return fmt.Sprintf("%s %s %s", a.Target.String(), a.Op, a.Value.String())
}

// ExprStmt represents expression statement
type ExprStmt struct {
	Expr Expr
}

func (e *ExprStmt) stmtNode() {}
func (e *ExprStmt) String() string {
	return e.Expr.String()
}

// ReturnStmt represents return
type ReturnStmt struct {
	Value Expr // nil for bare return
}

func (r *ReturnStmt) stmtNode() {}
func (r *ReturnStmt) String() string {
	if r.Value != nil {
		return "return " + r.Value.String()
	}

	return "return"
}

// IfStmt represents if/else
type IfStmt struct {
	Cond Expr
	Then *Block
	Else Stmt // nil, *Block, or *IfStmt
}

func (i *IfStmt) stmtNode() {}
func (i *IfStmt) String() string {
	s := fmt.Sprintf("if %s %s", i.Cond.String(), i.Then.String())
	if i.Else != nil {
		s += " else " + i.Else.String()
	}

	return s
}

// WhileStmt represents while loop
type WhileStmt struct {
	Cond Expr
	Body *Block
}

func (w *WhileStmt) stmtNode() {}
func (w *WhileStmt) String() string {
	return fmt.Sprintf("while %s %s", w.Cond.String(), w.Body.String())
}

// ForStmt represents for loop
type ForStmt struct {
	Key  string // empty if not used
	Val  string
	Iter Expr
	Body *Block
}

func (f *ForStmt) stmtNode() {}
func (f *ForStmt) String() string {
	if f.Key != "" {
		return fmt.Sprintf("for %s, %s in %s %s", f.Key, f.Val, f.Iter.String(), f.Body.String())
	}

	return fmt.Sprintf("for %s in %s %s", f.Val, f.Iter.String(), f.Body.String())
}

// BreakStmt represents break
type BreakStmt struct{}

func (b *BreakStmt) stmtNode() {}
func (b *BreakStmt) String() string {
	return "break"
}

// ContinueStmt represents continue
type ContinueStmt struct{}

func (c *ContinueStmt) stmtNode() {}
func (c *ContinueStmt) String() string {
	return "continue"
}

// DeferStmt represents defer
type DeferStmt struct {
	Expr Expr
}

func (d *DeferStmt) stmtNode() {}
func (d *DeferStmt) String() string {
	return "defer " + d.Expr.String()
}

// ShortDecl represents := declaration
type ShortDecl struct {
	Name  string
	Value Expr
}

func (s *ShortDecl) stmtNode() {}
func (s *ShortDecl) String() string {
	return fmt.Sprintf("%s := %s", s.Name, s.Value.String())
}

// ConstStmt represents block-level const statement
type ConstStmt struct {
	Name  string
	Type  Type
	Value Expr
}

func (c *ConstStmt) stmtNode() {}
func (c *ConstStmt) String() string {
	return fmt.Sprintf("const %s: %s = %s", c.Name, c.Type.String(), c.Value.String())
}

// UnsafeBlock represents unsafe { }
type UnsafeBlock struct {
	Body *Block
}

func (u *UnsafeBlock) stmtNode() {}
func (u *UnsafeBlock) String() string {
	return "unsafe " + u.Body.String()
}

// Block represents a block of statements
type Block struct {
	Stmts []Stmt
}

func (b *Block) stmtNode() {}
func (b *Block) String() string {
	stmts := make([]string, len(b.Stmts))
	for i, s := range b.Stmts {
		stmts[i] = s.String()
	}

	return "{ " + strings.Join(stmts, "; ") + " }"
}

// ===== Declarations =====

// Decl represents a top-level declaration
type Decl interface {
	Node
	declNode()
}

// UseDecl represents use/import
type UseDecl struct {
	Path  []string
	Alias string // empty if no alias
}

func (u *UseDecl) declNode() {}
func (u *UseDecl) String() string {
	path := strings.Join(u.Path, "::")
	if u.Alias != "" {
		return fmt.Sprintf("use %s as %s", path, u.Alias)
	}

	return "use " + path
}

// ConstDecl represents const declaration
type ConstDecl struct {
	Name  string
	Type  Type
	Value Expr
}

func (c *ConstDecl) declNode() {}
func (c *ConstDecl) String() string {
	return fmt.Sprintf("const %s: %s = %s", c.Name, c.Type.String(), c.Value.String())
}

// TypeAlias represents type alias
type TypeAlias struct {
	Name string
	Type Type
}

func (t *TypeAlias) declNode() {}
func (t *TypeAlias) String() string {
	return fmt.Sprintf("type %s = %s", t.Name, t.Type.String())
}

// StructDecl represents struct definition
type StructDecl struct {
	Pub     bool
	Name    string
	TParams []string // Generic type parameters
	Fields  []Field
}

type Field struct {
	Name string
	Type Type
}

func (s *StructDecl) declNode() {}
func (s *StructDecl) String() string {
	pub := ""
	if s.Pub {
		pub = "pub "
	}

	fields := make([]string, len(s.Fields))
	for i, f := range s.Fields {
		fields[i] = fmt.Sprintf("%s: %s", f.Name, f.Type.String())
	}

	tparams := ""
	if len(s.TParams) > 0 {
		tparams = "<" + strings.Join(s.TParams, ", ") + ">"
	}

	return fmt.Sprintf("%sstruct %s%s { %s }", pub, s.Name, tparams, strings.Join(fields, ", "))
}

// EnumDecl represents enum definition
type EnumDecl struct {
	Pub      bool
	Name     string
	TParams  []string
	Variants []Variant
}

type Variant struct {
	Name  string
	Types []Type // nil if no payload
}

func (e *EnumDecl) declNode() {}
func (e *EnumDecl) String() string {
	pub := ""
	if e.Pub {
		pub = "pub "
	}

	tparams := ""
	if len(e.TParams) > 0 {
		tparams = "<" + strings.Join(e.TParams, ", ") + ">"
	}

	return fmt.Sprintf("%senum %s%s { ... }", pub, e.Name, tparams)
}

// TraitDecl represents trait definition
type TraitDecl struct {
	Pub     bool
	Name    string
	TParams []string
	Sigs    []FnSig
}

type FnSig struct {
	Name   string
	Params []Param
	Return Type
}

func (t *TraitDecl) declNode() {}
func (t *TraitDecl) String() string {
	pub := ""
	if t.Pub {
		pub = "pub "
	}

	tparams := ""
	if len(t.TParams) > 0 {
		tparams = "<" + strings.Join(t.TParams, ", ") + ">"
	}

	return fmt.Sprintf("%strait %s%s { ... }", pub, t.Name, tparams)
}

// ImplBlock represents impl block
type ImplBlock struct {
	Trait *TypePath // nil if inherent impl
	For   Type
	Fns   []*FuncDecl
}

func (i *ImplBlock) declNode() {}
func (i *ImplBlock) String() string {
	if i.Trait != nil {
		return fmt.Sprintf("impl %s for %s { ... }", i.Trait.String(), i.For.String())
	}

	return fmt.Sprintf("impl %s { ... }", i.For.String())
}

// FuncDecl represents function declaration
type FuncDecl struct {
	Pub        bool
	Name       string
	TParams    []string
	Params     []Param
	ReturnType Type
	Body       *Block
}

type Param struct {
	Mut  bool
	Name string
	Type Type
}

func (f *FuncDecl) declNode() {}
func (f *FuncDecl) String() string {
	pub := ""
	if f.Pub {
		pub = "pub "
	}

	params := make([]string, len(f.Params))
	for i, p := range f.Params {
		mut := ""
		if p.Mut {
			mut = "mut "
		}

		params[i] = fmt.Sprintf("%s%s %s", mut, p.Name, p.Type.String())
	}

	tparams := ""
	if len(f.TParams) > 0 {
		tparams = "<" + strings.Join(f.TParams, ", ") + ">"
	}

	ret := "void"
	if f.ReturnType != nil {
		ret = f.ReturnType.String()
	}

	return fmt.Sprintf("%sfn %s%s(%s) %s", pub, f.Name, tparams, strings.Join(params, ", "), ret)
}

// File represents a source file
type File struct {
	Module []string // module path
	Items  []Decl
}

func (f *File) String() string {
	items := make([]string, len(f.Items))
	for i, it := range f.Items {
		items[i] = it.String()
	}

	return strings.Join(items, "\n")
}
