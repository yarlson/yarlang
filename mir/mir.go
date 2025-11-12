package mir

import "fmt"

// Instruction represents a MIR instruction
type Instruction interface {
	String() string
	isInstr()
}

// Type represents a MIR type
type Type interface {
	String() string
	isType()
}

// PrimitiveType represents primitive types
type PrimitiveType struct {
	Name string // i8, i16, i32, i64, isize, u8, u16, u32, u64, usize, f32, f64, bool, char, void
}

func (p *PrimitiveType) isType() {}
func (p *PrimitiveType) String() string {
	return p.Name
}

// PtrType represents pointer types
type PtrType struct {
	Elem Type
}

func (p *PtrType) isType() {}
func (p *PtrType) String() string {
	return fmt.Sprintf("*%s", p.Elem.String())
}

// StructType represents struct types
type StructType struct {
	Name   string
	Fields []Type
}

func (s *StructType) isType() {}
func (s *StructType) String() string {
	return fmt.Sprintf("%%struct.%s", s.Name)
}

// OpKind represents operation kinds
type OpKind int

const (
	Add OpKind = iota
	Sub
	Mul
	Div
	Mod
	And
	Or
	Xor
	Shl
	Shr
	Eq
	Ne
	Lt
	Le
	Gt
	Ge
)

// opNames maps operation kinds to their string representations
var opNames = map[OpKind]string{
	Add: "add", Sub: "sub", Mul: "mul", Div: "div", Mod: "mod",
	And: "and", Or: "or", Xor: "xor", Shl: "shl", Shr: "shr",
	Eq: "eq", Ne: "ne", Lt: "lt", Le: "le", Gt: "gt", Ge: "ge",
}

// Alloca allocates stack space
type Alloca struct {
	Name string
	Type Type
}

func (a *Alloca) isInstr() {}
func (a *Alloca) String() string {
	return fmt.Sprintf("%%%s = alloca %s", a.Name, a.Type.String())
}

// Load loads from memory
type Load struct {
	Dest   string
	Source string
	Type   Type
}

func (l *Load) isInstr() {}
func (l *Load) String() string {
	return fmt.Sprintf("%%%s = load %s, %s* %%%s", l.Dest, l.Type.String(), l.Type.String(), l.Source)
}

// Store stores to memory
type Store struct {
	Value string
	Dest  string
	Type  Type
}

func (s *Store) isInstr() {}
func (s *Store) String() string {
	return fmt.Sprintf("store %s %%%s, %s* %%%s", s.Type.String(), s.Value, s.Type.String(), s.Dest)
}

// BinOp represents binary operations
type BinOp struct {
	Dest  string
	Op    OpKind
	Left  string
	Right string
	Type  Type
}

func (b *BinOp) isInstr() {}
func (b *BinOp) String() string {
	return fmt.Sprintf("%%%s = %s %s %%%s, %%%s", b.Dest, opNames[b.Op], b.Type.String(), b.Left, b.Right)
}

// Call represents function call
type Call struct {
	Dest   string   // destination register (empty for void calls)
	Callee string   // function name
	Args   []string // argument values (registers or immediates)
	RetTy  Type     // return type
}

func (c *Call) isInstr() {}
func (c *Call) String() string {
	args := ""

	for i, arg := range c.Args {
		if i > 0 {
			args += ", "
		}
		// Check if arg is a string literal (starts with quote)
		if len(arg) > 0 && arg[0] == '"' {
			args += arg
		} else if len(arg) > 0 && (arg[0] >= '0' && arg[0] <= '9' || arg[0] == '-') {
			// Number (immediate)
			args += arg
		} else {
			// Register
			args += "%" + arg
		}
	}

	if c.Dest == "" {
		// Void call
		return fmt.Sprintf("call %s @%s(%s)", c.RetTy.String(), c.Callee, args)
	}

	return fmt.Sprintf("%%%s = call %s @%s(%s)", c.Dest, c.RetTy.String(), c.Callee, args)
}

// Ret represents return
type Ret struct {
	Value string // empty for void return
	Type  Type
}

func (r *Ret) isInstr() {}
func (r *Ret) String() string {
	if r.Value == "" {
		return "ret void"
	}

	return fmt.Sprintf("ret %s %%%s", r.Type.String(), r.Value)
}

// Br represents unconditional branch
type Br struct {
	Label string
}

func (b *Br) isInstr() {}
func (b *Br) String() string {
	return fmt.Sprintf("br label %%bb_%s", b.Label)
}

// CondBr represents conditional branch
type CondBr struct {
	Cond       string
	TrueLabel  string
	FalseLabel string
}

func (c *CondBr) isInstr() {}
func (c *CondBr) String() string {
	return fmt.Sprintf("br i1 %%%s, label %%bb_%s, label %%bb_%s", c.Cond, c.TrueLabel, c.FalseLabel)
}

// DeferPush pushes a deferred call onto the defer stack
type DeferPush struct {
	Call *Call // the deferred call
}

func (d *DeferPush) isInstr() {}
func (d *DeferPush) String() string {
	return fmt.Sprintf("defer_push %s", d.Call.String())
}

// DeferRunAll runs all deferred calls in LIFO order
type DeferRunAll struct{}

func (d *DeferRunAll) isInstr() {}
func (d *DeferRunAll) String() string {
	return "defer_run_all"
}

// BasicBlock represents a basic block
type BasicBlock struct {
	Label  string
	Instrs []Instruction
}

func (bb *BasicBlock) String() string {
	s := fmt.Sprintf("bb_%s:\n", bb.Label)
	for _, instr := range bb.Instrs {
		s += "  " + instr.String() + "\n"
	}

	return s
}

// Function represents a MIR function
type Function struct {
	Name   string
	Params []Param
	RetTy  Type
	Blocks []*BasicBlock
}

type Param struct {
	Name string
	Type Type
}

func (f *Function) String() string {
	return fmt.Sprintf("define %s @%s(...) { ... }", f.RetTy.String(), f.Name)
}

// Global represents a global variable or constant
type Global interface {
	GlobalName() string
	isGlobal()
}

// GlobalString represents a global string constant
type GlobalString struct {
	Name  string // e.g., ".str.0"
	Value string // the string content (without quotes)
}

func (g *GlobalString) isGlobal() {}
func (g *GlobalString) GlobalName() string {
	return g.Name
}

// Module represents a MIR module
type Module struct {
	Globals   []Global
	Functions []*Function
}
