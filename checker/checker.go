package checker

import (
	"fmt"
	"strconv"

	"github.com/yarlson/yarlang/ast"
	"github.com/yarlson/yarlang/types"
)

type BorrowState int

const (
	NotBorrowed BorrowState = iota
	SharedBorrow
	MutBorrow
)

// Checker performs semantic analysis
type Checker struct {
	env     *types.Env
	errors  []string
	moved   map[*types.Symbol]bool        // Track moved variables by symbol pointer (scope-aware)
	borrows map[*types.Symbol]BorrowState // Track borrow state
}

func NewChecker() *Checker {
	return &Checker{
		env:     types.NewEnv(),
		errors:  []string{},
		moved:   make(map[*types.Symbol]bool),
		borrows: make(map[*types.Symbol]BorrowState),
	}
}

func (c *Checker) error(msg string) {
	c.errors = append(c.errors, msg)
}

func (c *Checker) CheckFile(file *ast.File) error {
	// Check all declarations
	for _, decl := range file.Items {
		c.checkDecl(decl)
	}

	if len(c.errors) > 0 {
		return fmt.Errorf("type errors: %v", c.errors)
	}

	return nil
}

func (c *Checker) checkDecl(decl ast.Decl) {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		c.checkFuncDecl(d)
	case *ast.StructDecl:
		c.checkStructDecl(d)
	case *ast.EnumDecl:
		c.checkEnumDecl(d)
	// ... other decls
	default:
		c.error(fmt.Sprintf("unknown declaration type: %T", decl))
	}
}

func (c *Checker) checkFuncDecl(fn *ast.FuncDecl) {
	// Build function type
	paramTypes := []types.Type{}
	for _, param := range fn.Params {
		paramTypes = append(paramTypes, c.resolveType(param.Type))
	}

	var returnType types.Type = &types.PrimitiveType{Name: "void", Kind: types.Void}
	if fn.ReturnType != nil {
		returnType = c.resolveType(fn.ReturnType)
	}

	funcType := &types.FuncType{
		Params: paramTypes,
		Return: returnType,
	}

	// Register function in environment
	c.env.Define(fn.Name, funcType, false)

	// Push new scope for function body
	c.env.PushScope()
	defer c.env.PopScope()

	// Add parameters to scope
	for _, param := range fn.Params {
		typ := c.resolveType(param.Type)
		c.env.Define(param.Name, typ, param.Mut)
	}

	// Check body
	c.checkBlock(fn.Body)
}

func (c *Checker) checkBlock(block *ast.Block) {
	for _, stmt := range block.Stmts {
		c.checkStmt(stmt)
	}
}

func (c *Checker) checkStmt(stmt ast.Stmt) types.Type {
	switch s := stmt.(type) {
	case *ast.LetStmt:
		return c.checkLetStmt(s)
	case *ast.AssignStmt:
		return c.checkAssignStmt(s)
	case *ast.ReturnStmt:
		return c.checkReturnStmt(s)
	case *ast.ExprStmt:
		return c.checkExpr(s.Expr)
	case *ast.IfStmt:
		return c.checkIfStmt(s)
	// ... other stmts
	default:
		c.error(fmt.Sprintf("unknown statement type: %T", stmt))
		return nil
	}
}

func (c *Checker) checkLetStmt(let *ast.LetStmt) types.Type {
	// Check value expression
	valueType := c.checkExpr(let.Value)

	// If value is an identifier of Move type, mark it as moved
	if ident, ok := let.Value.(*ast.Ident); ok {
		if !types.IsCopy(valueType) {
			// Look up the symbol and mark it as moved
			sym, ok := c.env.LookupSymbol(ident.Name)
			if ok {
				c.moved[sym] = true
			}
		}
	}

	// If type annotation present, check compatibility
	if let.Type != nil {
		declaredType := c.resolveType(let.Type)
		if !types.TypesEqual(valueType, declaredType) {
			c.error(fmt.Sprintf("type mismatch: expected %s, got %s",
				declaredType.String(), valueType.String()))
		}
	}

	// Define variable
	finalType := valueType
	if let.Type != nil {
		finalType = c.resolveType(let.Type)
	}

	c.env.Define(let.Name, finalType, let.Mut)

	return nil
}

func (c *Checker) checkAssignStmt(assign *ast.AssignStmt) types.Type {
	// Check target is mutable
	if ident, ok := assign.Target.(*ast.Ident); ok {
		typ, mut, ok := c.env.Lookup(ident.Name)
		if !ok {
			c.error(fmt.Sprintf("undefined variable: %s", ident.Name))
			return nil
		}

		if !mut {
			c.error(fmt.Sprintf("cannot assign to immutable variable: %s", ident.Name))
		}

		// Check value type matches
		valueType := c.checkExpr(assign.Value)
		if !types.TypesEqual(typ, valueType) {
			c.error(fmt.Sprintf("type mismatch: expected %s, got %s",
				typ.String(), valueType.String()))
		}
	}

	return nil
}

func (c *Checker) checkReturnStmt(ret *ast.ReturnStmt) types.Type {
	if ret.Value != nil {
		return c.checkExpr(ret.Value)
	}

	return &types.PrimitiveType{Name: "void", Kind: types.Void}
}

func (c *Checker) checkIfStmt(ifStmt *ast.IfStmt) types.Type {
	// Check condition is bool
	condType := c.checkExpr(ifStmt.Cond)

	boolType := &types.PrimitiveType{Name: "bool", Kind: types.Bool}
	if !types.TypesEqual(condType, boolType) {
		c.error(fmt.Sprintf("if condition must be bool, got %s", condType.String()))
	}

	// Check then block
	c.checkBlock(ifStmt.Then)

	// Check else block if present
	if ifStmt.Else != nil {
		c.checkStmt(ifStmt.Else)
	}

	return nil
}

func (c *Checker) checkExpr(expr ast.Expr) types.Type {
	switch e := expr.(type) {
	case *ast.IntLit:
		return &types.PrimitiveType{Name: "i32", Kind: types.Int32}
	case *ast.FloatLit:
		return &types.PrimitiveType{Name: "f64", Kind: types.Float64}
	case *ast.BoolLit:
		return &types.PrimitiveType{Name: "bool", Kind: types.Bool}
	case *ast.StringLit:
		// String is []u8
		u8 := &types.PrimitiveType{Name: "u8", Kind: types.UInt8}
		return &types.SliceType{Elem: u8}
	case *ast.NilLit:
		// nil can be any pointer type, return a type var for now
		return c.env.NewTypeVar()
	case *ast.Ident:
		// Look up the symbol
		sym, ok := c.env.LookupSymbol(e.Name)
		if !ok {
			c.error(fmt.Sprintf("undefined variable: %s", e.Name))
			return c.env.NewTypeVar()
		}

		// Check if moved
		if c.moved[sym] {
			c.error(fmt.Sprintf("use of moved value: %s", e.Name))
			return c.env.NewTypeVar()
		}

		// If type is not Copy and not borrowed, mark as moved
		// (simplified: only track in assignments)
		return sym.Type
	case *ast.BinaryExpr:
		return c.checkBinaryExpr(e)
	case *ast.UnaryExpr:
		return c.checkUnaryExpr(e)
	case *ast.CallExpr:
		return c.checkCallExpr(e)
	case *ast.StructExpr:
		return c.checkStructExpr(e)
	// ... other exprs
	default:
		c.error(fmt.Sprintf("unknown expression type: %T", expr))
		return c.env.NewTypeVar()
	}
}

func (c *Checker) checkBinaryExpr(bin *ast.BinaryExpr) types.Type {
	leftType := c.checkExpr(bin.Left)
	rightType := c.checkExpr(bin.Right)

	// Check types match
	if !types.TypesEqual(leftType, rightType) {
		c.error(fmt.Sprintf("type mismatch in binary expression: %s and %s",
			leftType.String(), rightType.String()))
	}

	// Arithmetic operators return same type
	if bin.Op == "+" || bin.Op == "-" || bin.Op == "*" || bin.Op == "/" || bin.Op == "%" {
		return leftType
	}

	// Comparison operators return bool
	if bin.Op == "==" || bin.Op == "!=" || bin.Op == "<" || bin.Op == ">" ||
		bin.Op == "<=" || bin.Op == ">=" {
		return &types.PrimitiveType{Name: "bool", Kind: types.Bool}
	}

	// Logical operators return bool
	if bin.Op == "&&" || bin.Op == "||" {
		return &types.PrimitiveType{Name: "bool", Kind: types.Bool}
	}

	return leftType
}

func (c *Checker) checkUnaryExpr(un *ast.UnaryExpr) types.Type {
	exprType := c.checkExpr(un.Expr)

	if un.Op == "&" {
		// Shared borrow
		if ident, ok := un.Expr.(*ast.Ident); ok {
			// Look up the symbol
			sym, ok := c.env.LookupSymbol(ident.Name)
			if ok {
				// Check not exclusively borrowed
				if c.borrows[sym] == MutBorrow {
					c.error(fmt.Sprintf("cannot borrow %s as shared because it is also borrowed as mutable", ident.Name))
				}

				c.borrows[sym] = SharedBorrow
			}
		}

		return &types.RefType{Mut: false, Elem: exprType}
	}

	if un.Op == "&mut" {
		// Exclusive borrow
		if ident, ok := un.Expr.(*ast.Ident); ok {
			// Look up the symbol
			sym, ok := c.env.LookupSymbol(ident.Name)
			if ok {
				// Check not borrowed at all
				if c.borrows[sym] != NotBorrowed {
					c.error(fmt.Sprintf("cannot borrow %s as mutable because it is already borrowed", ident.Name))
				}

				c.borrows[sym] = MutBorrow
			}
		}

		return &types.RefType{Mut: true, Elem: exprType}
	}

	if un.Op == "*" {
		// Dereference
		if refType, ok := exprType.(*types.RefType); ok {
			return refType.Elem
		}

		if ptrType, ok := exprType.(*types.PtrType); ok {
			return ptrType.Elem
		}

		c.error("cannot dereference non-pointer type")
	}

	return exprType
}

func (c *Checker) checkCallExpr(call *ast.CallExpr) types.Type {
	// Extract function name from callee
	var funcName string
	switch callee := call.Callee.(type) {
	case *ast.Ident:
		funcName = callee.Name
	case *ast.FieldExpr:
		// For module paths like std::io::println, just use the field name
		funcName = callee.Field
	default:
		c.error(fmt.Sprintf("invalid function call: %T", call.Callee))
		return c.env.NewTypeVar()
	}

	// Look up function
	funcType, _, ok := c.env.Lookup(funcName)
	if !ok {
		c.error(fmt.Sprintf("undefined function: %s", funcName))
		return c.env.NewTypeVar()
	}

	// Check if it's actually a function type
	fn, ok := funcType.(*types.FuncType)
	if !ok {
		c.error(fmt.Sprintf("%s is not a function", funcName))
		return c.env.NewTypeVar()
	}

	// Check argument count
	if len(call.Args) != len(fn.Params) {
		c.error(fmt.Sprintf("function %s expects %d arguments, got %d",
			funcName, len(fn.Params), len(call.Args)))
		// Still check arguments to find other errors
	}

	// Check argument types
	minArgs := len(call.Args)
	if len(fn.Params) < minArgs {
		minArgs = len(fn.Params)
	}

	for i := 0; i < minArgs; i++ {
		argType := c.checkExpr(call.Args[i])
		expectedType := fn.Params[i]

		// Skip type checking if either is a type variable (for generic/builtin functions)
		if _, isTypeVar := expectedType.(*types.TypeVar); isTypeVar {
			continue
		}

		if !types.TypesEqual(argType, expectedType) {
			c.error(fmt.Sprintf("argument %d to %s: expected %s, got %s",
				i+1, funcName, expectedType.String(), argType.String()))
		}
	}

	// Check remaining arguments if there are extra
	for i := minArgs; i < len(call.Args); i++ {
		c.checkExpr(call.Args[i])
	}

	// Return function's return type
	return fn.Return
}

func (c *Checker) checkStructExpr(s *ast.StructExpr) types.Type {
	// Resolve the struct type
	structType := c.resolveType(s.Type)

	// Check each field initialization
	for _, init := range s.Inits {
		c.checkExpr(init.Val)
	}

	return structType
}

func (c *Checker) checkStructDecl(s *ast.StructDecl) {
	// Track if we pushed a scope for type parameters
	scopePushed := false

	// If struct has type parameters, push a new scope and define them
	if len(s.TParams) > 0 {
		c.env.PushScope()

		scopePushed = true

		// Define type parameters as type variables in scope
		for _, tparam := range s.TParams {
			typeVar := c.env.NewTypeVar()
			c.env.Define(tparam, typeVar, false)
		}
	}

	// Register struct type
	fields := make(map[string]types.Type)
	for _, field := range s.Fields {
		fields[field.Name] = c.resolveType(field.Type)
	}

	structType := &types.StructType{
		Name:    s.Name,
		Fields:  fields,
		TParams: s.TParams,
	}

	// Pop scope if we pushed one, then define struct in parent scope
	if scopePushed {
		c.env.PopScope()
	}

	// Define struct type in current (parent) scope
	c.env.Define(s.Name, structType, false)
}

func (c *Checker) checkEnumDecl(e *ast.EnumDecl) {
	// Register enum type
	variants := make(map[string][]types.Type)

	for _, variant := range e.Variants {
		variantTypes := []types.Type{}
		for _, vtype := range variant.Types {
			variantTypes = append(variantTypes, c.resolveType(vtype))
		}

		variants[variant.Name] = variantTypes
	}

	enumType := &types.EnumType{
		Name:     e.Name,
		Variants: variants,
		TParams:  e.TParams,
	}

	c.env.Define(e.Name, enumType, false)
}

func (c *Checker) resolveType(astType ast.Type) types.Type {
	switch t := astType.(type) {
	case *ast.TypePath:
		// Handle generic instantiation
		if len(t.Args) > 0 {
			// Generic instantiation
			baseType, _, ok := c.env.Lookup(t.Path[len(t.Path)-1])
			if !ok {
				c.error(fmt.Sprintf("undefined type: %s", t.Path[len(t.Path)-1]))
				return c.env.NewTypeVar()
			}

			// Resolve generic arguments (for validation)
			for _, arg := range t.Args {
				_ = c.resolveType(arg)
			}

			// Create instantiated type (simplified - just store base type)
			// In a full implementation, we would create a new type with substituted type parameters
			return baseType
		}

		// Look up type by name
		if len(t.Path) == 1 {
			typ, _, ok := c.env.Lookup(t.Path[0])
			if !ok {
				c.error(fmt.Sprintf("undefined type: %s", t.Path[0]))
				return c.env.NewTypeVar()
			}

			return typ
		}
		// For now, just use last component
		typ, _, ok := c.env.Lookup(t.Path[len(t.Path)-1])
		if !ok {
			c.error(fmt.Sprintf("undefined type: %s", t.Path[len(t.Path)-1]))
			return c.env.NewTypeVar()
		}

		return typ
	case *ast.RefType:
		elem := c.resolveType(t.Elem)
		return &types.RefType{Mut: t.Mut, Elem: elem}
	case *ast.PtrType:
		elem := c.resolveType(t.Elem)
		return &types.PtrType{Elem: elem}
	case *ast.SliceType:
		elem := c.resolveType(t.Elem)
		return &types.SliceType{Elem: elem}
	case *ast.ArrayType:
		elem := c.resolveType(t.Elem)

		// Parse array length from AST
		if intLit, ok := t.Len.(*ast.IntLit); ok {
			// strconv.ParseInt with base 0 automatically handles:
			// - Decimal: "10"
			// - Hex: "0x10"
			// - Octal: "0o10" or "010"
			// - Binary: "0b10"
			length, err := strconv.ParseInt(intLit.Value, 0, 64)
			if err != nil {
				c.error(fmt.Sprintf("invalid array length: %s", intLit.Value))
				return &types.ArrayType{Elem: elem, Len: 0}
			}

			// Validate length is positive
			if length <= 0 {
				c.error(fmt.Sprintf("array length must be positive, got %d", length))
				return &types.ArrayType{Elem: elem, Len: 0}
			}

			return &types.ArrayType{Elem: elem, Len: int(length)}
		}

		// If no length specified or not an IntLit, default to 0 (error case)
		c.error("array length must be a constant integer")
		return &types.ArrayType{Elem: elem, Len: 0}
	case *ast.TupleType:
		elems := []types.Type{}
		for _, e := range t.Elems {
			elems = append(elems, c.resolveType(e))
		}

		return &types.TupleType{Elems: elems}
	case *ast.VoidType:
		return &types.PrimitiveType{Name: "void", Kind: types.Void}
	default:
		c.error(fmt.Sprintf("unknown type: %T", astType))
		return c.env.NewTypeVar()
	}
}
