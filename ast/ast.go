package ast

import (
	"fmt"
	"strings"
)

// Node is the base interface for all AST nodes
type Node interface {
	String() string
}

// Expr represents an expression node
type Expr interface {
	Node
	exprNode()
}

// Stmt represents a statement node
type Stmt interface {
	Node
	stmtNode()
}

// Program is the root node of the AST
type Program struct {
	Statements []Stmt
}

func (p *Program) String() string {
	var out strings.Builder
	for _, s := range p.Statements {
		out.WriteString(s.String())
		out.WriteString("\n")
	}

	return out.String()
}

// Expressions

type NumberLiteral struct {
	Value float64
	Range Range
}

func (n *NumberLiteral) exprNode()      {}
func (n *NumberLiteral) String() string { return fmt.Sprintf("%g", n.Value) }

type StringLiteral struct {
	Value string
	Range Range
}

func (s *StringLiteral) exprNode()      {}
func (s *StringLiteral) String() string { return fmt.Sprintf(`"%s"`, s.Value) }

type BoolLiteral struct {
	Value bool
	Range Range
}

func (b *BoolLiteral) exprNode() {}
func (b *BoolLiteral) String() string {
	if b.Value {
		return "true"
	}

	return "false"
}

type NilLiteral struct {
	Range Range
}

func (n *NilLiteral) exprNode()      {}
func (n *NilLiteral) String() string { return "nil" }

type Identifier struct {
	Name  string
	Range Range
}

func (i *Identifier) exprNode()      {}
func (i *Identifier) String() string { return i.Name }

type BinaryExpr struct {
	Left     Expr
	Operator string
	Right    Expr
	Range    Range
}

func (b *BinaryExpr) exprNode() {}
func (b *BinaryExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", b.Left.String(), b.Operator, b.Right.String())
}

type UnaryExpr struct {
	Operator string
	Right    Expr
	Range    Range
}

func (u *UnaryExpr) exprNode() {}
func (u *UnaryExpr) String() string {
	return fmt.Sprintf("(%s%s)", u.Operator, u.Right.String())
}

type CallExpr struct {
	Function Expr // Identifier or other expression
	Args     []Expr
	Range    Range
}

func (c *CallExpr) exprNode() {}
func (c *CallExpr) String() string {
	args := make([]string, len(c.Args))
	for i, a := range c.Args {
		args[i] = a.String()
	}

	return fmt.Sprintf("%s(%s)", c.Function.String(), strings.Join(args, ", "))
}

// Statements

type ExprStmt struct {
	Expr  Expr
	Range Range
}

func (e *ExprStmt) stmtNode()      {}
func (e *ExprStmt) String() string { return e.Expr.String() }

type AssignStmt struct {
	Targets      []string // Variable names
	Values       []Expr   // Right-hand side expressions
	Range        Range
	TargetRanges []Range // Position of each target identifier
}

func (a *AssignStmt) stmtNode() {}
func (a *AssignStmt) String() string {
	vals := make([]string, len(a.Values))
	for i, v := range a.Values {
		vals[i] = v.String()
	}

	return fmt.Sprintf("%s = %s", strings.Join(a.Targets, ", "), strings.Join(vals, ", "))
}

type ReturnStmt struct {
	Values []Expr
	Range  Range
}

func (r *ReturnStmt) stmtNode() {}
func (r *ReturnStmt) String() string {
	if len(r.Values) == 0 {
		return "return"
	}

	vals := make([]string, len(r.Values))
	for i, v := range r.Values {
		vals[i] = v.String()
	}

	return fmt.Sprintf("return %s", strings.Join(vals, ", "))
}

type IfStmt struct {
	Condition Expr
	ThenBlock *BlockStmt
	ElseBlock *BlockStmt // Can be nil
	Range     Range
}

func (i *IfStmt) stmtNode() {}
func (i *IfStmt) String() string {
	var out strings.Builder
	out.WriteString(fmt.Sprintf("if %s %s", i.Condition.String(), i.ThenBlock.String()))

	if i.ElseBlock != nil {
		out.WriteString(fmt.Sprintf(" else %s", i.ElseBlock.String()))
	}

	return out.String()
}

type ForStmt struct {
	Init      Stmt // Can be nil
	Condition Expr // Can be nil (infinite loop)
	Post      Stmt // Can be nil
	Body      *BlockStmt
	Range     Range
}

func (f *ForStmt) stmtNode() {}
func (f *ForStmt) String() string {
	var parts []string
	if f.Init != nil {
		parts = append(parts, f.Init.String())
	}

	if f.Condition != nil {
		parts = append(parts, f.Condition.String())
	}

	if f.Post != nil {
		parts = append(parts, f.Post.String())
	}

	return fmt.Sprintf("for %s %s", strings.Join(parts, "; "), f.Body.String())
}

type BreakStmt struct {
	Range Range
}

func (b *BreakStmt) stmtNode()      {}
func (b *BreakStmt) String() string { return "break" }

type ContinueStmt struct {
	Range Range
}

func (c *ContinueStmt) stmtNode()      {}
func (c *ContinueStmt) String() string { return "continue" }

type BlockStmt struct {
	Statements []Stmt
	Range      Range
}

func (b *BlockStmt) stmtNode() {}
func (b *BlockStmt) String() string {
	var out strings.Builder
	out.WriteString("{ ")

	for _, s := range b.Statements {
		out.WriteString(s.String())
		out.WriteString("; ")
	}

	out.WriteString("}")

	return out.String()
}

type FuncDecl struct {
	Name        string
	Params      []string
	Body        *BlockStmt
	Range       Range   // Entire function range
	NameRange   Range   // Just the function name (for go-to-def)
	ParamRanges []Range // Position of each parameter
}

func (f *FuncDecl) stmtNode() {}
func (f *FuncDecl) String() string {
	return fmt.Sprintf("func %s(%s) %s", f.Name, strings.Join(f.Params, ", "), f.Body.String())
}
