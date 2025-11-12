package ast

import "testing"

func TestNumberLiteralString(t *testing.T) {
	lit := &NumberLiteral{Value: 42.0}
	if lit.String() != "42" {
		t.Errorf("NumberLiteral.String() wrong. got=%q", lit.String())
	}
}

func TestBinaryExprString(t *testing.T) {
	expr := &BinaryExpr{
		Left:     &NumberLiteral{Value: 1},
		Operator: "+",
		Right:    &NumberLiteral{Value: 2},
	}
	if expr.String() != "(1 + 2)" {
		t.Errorf("BinaryExpr.String() wrong. got=%q", expr.String())
	}
}

func TestIdentifierWithRange(t *testing.T) {
	ident := &Identifier{
		Name: "foo",
		Range: Range{
			Start: Position{Line: 1, Column: 5},
			End:   Position{Line: 1, Column: 8},
		},
	}

	if ident.Name != "foo" {
		t.Errorf("Name = %s, want foo", ident.Name)
	}

	if ident.Range.Start.Line != 1 || ident.Range.Start.Column != 5 {
		t.Errorf("Start position incorrect")
	}

	if ident.Range.End.Line != 1 || ident.Range.End.Column != 8 {
		t.Errorf("End position incorrect")
	}
}

func TestExpressionNodesHaveRange(t *testing.T) {
	// Test that all expression nodes have Range field
	_ = &NumberLiteral{Value: 42, Range: Range{}}
	_ = &StringLiteral{Value: "test", Range: Range{}}
	_ = &BoolLiteral{Value: true, Range: Range{}}
	_ = &NilLiteral{Range: Range{}}
	_ = &BinaryExpr{
		Left:     &Identifier{Name: "a"},
		Operator: "+",
		Right:    &Identifier{Name: "b"},
		Range:    Range{},
	}
	_ = &UnaryExpr{
		Operator: "!",
		Right:    &Identifier{Name: "x"},
		Range:    Range{},
	}
	_ = &CallExpr{
		Function: &Identifier{Name: "f"},
		Args:     []Expr{},
		Range:    Range{},
	}
}

func TestStatementNodesHaveRange(t *testing.T) {
	// Test that all statement nodes have Range field
	_ = &ExprStmt{Expr: &NilLiteral{}, Range: Range{}}
	_ = &AssignStmt{
		Targets:      []string{"x"},
		Values:       []Expr{&NumberLiteral{Value: 1}},
		Range:        Range{},
		TargetRanges: []Range{{}},
	}
	_ = &ReturnStmt{Values: []Expr{}, Range: Range{}}
	_ = &IfStmt{
		Condition: &BoolLiteral{Value: true},
		ThenBlock: &BlockStmt{},
		Range:     Range{},
	}
	_ = &ForStmt{Body: &BlockStmt{}, Range: Range{}}
	_ = &BreakStmt{Range: Range{}}
	_ = &ContinueStmt{Range: Range{}}
	_ = &BlockStmt{Statements: []Stmt{}, Range: Range{}}
	_ = &FuncDecl{
		Name:        "foo",
		Params:      []string{"a"},
		Body:        &BlockStmt{},
		Range:       Range{},
		NameRange:   Range{},
		ParamRanges: []Range{{}},
	}
}
