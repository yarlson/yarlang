package codegen

import (
	"fmt"
	"os"

	"github.com/yarlson/yarlang/ast"
	"tinygo.org/x/go-llvm"
)

type CodeGen struct {
	module  llvm.Module
	builder llvm.Builder
	context llvm.Context

	// Runtime function declarations
	runtimeFuncs map[string]llvm.Value
	// Runtime function types (cached to avoid GlobalValueType() issues)
	runtimeFuncTypes map[string]llvm.Type

	// Variable storage (maps variable name to LLVM value)
	variables map[string]llvm.Value
	// Variable types (cached to avoid Type().ElementType() issues)
	variableTypes map[string]llvm.Type

	// Current function being generated
	currentFunc llvm.Value

	// Loop tracking for break/continue
	currentLoopExit llvm.BasicBlock // For break
	currentLoopPost llvm.BasicBlock // For continue
}

func New() *CodeGen {
	context := llvm.GlobalContext()
	module := context.NewModule("yarlang")
	builder := context.NewBuilder()

	gen := &CodeGen{
		module:           module,
		builder:          builder,
		context:          context,
		runtimeFuncs:     make(map[string]llvm.Value),
		runtimeFuncTypes: make(map[string]llvm.Type),
		variables:        make(map[string]llvm.Value),
		variableTypes:    make(map[string]llvm.Type),
	}

	gen.declareRuntimeFunctions()

	return gen
}

func (g *CodeGen) declareRuntimeFunctions() {
	// Declare runtime types
	valueType := g.context.StructType([]llvm.Type{
		g.context.Int8Type(),                      // type tag
		llvm.PointerType(g.context.Int8Type(), 0), // data pointer
	}, false)

	valuePtr := llvm.PointerType(valueType, 0)

	// Helper to declare a function and cache its type
	addFunc := func(name string, retType llvm.Type, paramTypes []llvm.Type) {
		fnType := llvm.FunctionType(retType, paramTypes, false)
		g.runtimeFuncTypes[name] = fnType
		g.runtimeFuncs[name] = llvm.AddFunction(g.module, name, fnType)
	}

	// Declare runtime functions
	addFunc("yar_number", valuePtr, []llvm.Type{g.context.DoubleType()})
	addFunc("yar_string", valuePtr, []llvm.Type{llvm.PointerType(g.context.Int8Type(), 0)})
	addFunc("yar_bool", valuePtr, []llvm.Type{g.context.Int1Type()})
	addFunc("yar_nil", valuePtr, []llvm.Type{})
	addFunc("yar_add", valuePtr, []llvm.Type{valuePtr, valuePtr})
	addFunc("yar_subtract", valuePtr, []llvm.Type{valuePtr, valuePtr})
	addFunc("yar_multiply", valuePtr, []llvm.Type{valuePtr, valuePtr})
	addFunc("yar_divide", valuePtr, []llvm.Type{valuePtr, valuePtr})
	addFunc("yar_modulo", valuePtr, []llvm.Type{valuePtr, valuePtr})
	addFunc("yar_eq", valuePtr, []llvm.Type{valuePtr, valuePtr})
	addFunc("yar_neq", valuePtr, []llvm.Type{valuePtr, valuePtr})
	addFunc("yar_lt", valuePtr, []llvm.Type{valuePtr, valuePtr})
	addFunc("yar_gt", valuePtr, []llvm.Type{valuePtr, valuePtr})
	addFunc("yar_lte", valuePtr, []llvm.Type{valuePtr, valuePtr})
	addFunc("yar_gte", valuePtr, []llvm.Type{valuePtr, valuePtr})
	addFunc("yar_and", valuePtr, []llvm.Type{valuePtr, valuePtr})
	addFunc("yar_or", valuePtr, []llvm.Type{valuePtr, valuePtr})
	addFunc("yar_not", valuePtr, []llvm.Type{valuePtr})
	addFunc("yar_negate", valuePtr, []llvm.Type{valuePtr})
	addFunc("yar_print", g.context.VoidType(), []llvm.Type{valuePtr})
	addFunc("yar_println", g.context.VoidType(), []llvm.Type{valuePtr})
	addFunc("yar_len", valuePtr, []llvm.Type{valuePtr})
	addFunc("yar_type", valuePtr, []llvm.Type{valuePtr})
	addFunc("yar_is_truthy", g.context.Int1Type(), []llvm.Type{valuePtr})
}

func (g *CodeGen) Generate(program *ast.Program) error {
	// Create main function
	mainType := llvm.FunctionType(g.context.Int32Type(), []llvm.Type{}, false)
	mainFunc := llvm.AddFunction(g.module, "main", mainType)
	entry := g.context.AddBasicBlock(mainFunc, "entry")
	g.builder.SetInsertPointAtEnd(entry)

	g.currentFunc = mainFunc

	// Generate code for statements
	for _, stmt := range program.Statements {
		if err := g.generateStmt(stmt); err != nil {
			return err
		}
	}

	// Return 0 from main
	g.builder.CreateRet(llvm.ConstInt(g.context.Int32Type(), 0, false))

	return nil
}

func (g *CodeGen) generateStmt(stmt ast.Stmt) error {
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		_, err := g.generateExpr(s.Expr)
		return err
	case *ast.AssignStmt:
		return g.generateAssign(s)
	case *ast.ReturnStmt:
		return g.generateReturn(s)
	case *ast.IfStmt:
		return g.generateIf(s)
	case *ast.ForStmt:
		return g.generateFor(s)
	case *ast.BlockStmt:
		return g.generateBlock(s)
	case *ast.FuncDecl:
		return g.generateFuncDecl(s)
	case *ast.BreakStmt:
		if g.currentLoopExit.IsNil() {
			return fmt.Errorf("break statement outside loop")
		}

		g.builder.CreateBr(g.currentLoopExit)

		return nil
	case *ast.ContinueStmt:
		if g.currentLoopPost.IsNil() {
			return fmt.Errorf("continue statement outside loop")
		}

		g.builder.CreateBr(g.currentLoopPost)

		return nil
	default:
		return fmt.Errorf("unsupported statement type: %T", stmt)
	}
}

func (g *CodeGen) generateExpr(expr ast.Expr) (llvm.Value, error) {
	switch e := expr.(type) {
	case *ast.NumberLiteral:
		// Create Value* via runtime
		numVal := llvm.ConstFloat(g.context.DoubleType(), e.Value)
		fn := g.runtimeFuncs["yar_number"]
		fnType := g.runtimeFuncTypes["yar_number"]

		return g.builder.CreateCall(fnType, fn, []llvm.Value{numVal}, ""), nil

	case *ast.StringLiteral:
		// Create global string constant
		strVal := g.builder.CreateGlobalStringPtr(e.Value, "str")
		fn := g.runtimeFuncs["yar_string"]
		fnType := g.runtimeFuncTypes["yar_string"]

		return g.builder.CreateCall(fnType, fn, []llvm.Value{strVal}, ""), nil

	case *ast.BoolLiteral:
		boolVal := llvm.ConstInt(g.context.Int1Type(), 0, false)
		if e.Value {
			boolVal = llvm.ConstInt(g.context.Int1Type(), 1, false)
		}

		fn := g.runtimeFuncs["yar_bool"]
		fnType := g.runtimeFuncTypes["yar_bool"]

		return g.builder.CreateCall(fnType, fn, []llvm.Value{boolVal}, ""), nil

	case *ast.NilLiteral:
		fn := g.runtimeFuncs["yar_nil"]
		fnType := g.runtimeFuncTypes["yar_nil"]

		return g.builder.CreateCall(fnType, fn, []llvm.Value{}, ""), nil

	case *ast.Identifier:
		if val, ok := g.variables[e.Name]; ok {
			// Get the cached type
			valType := g.variableTypes[e.Name]
			return g.builder.CreateLoad(valType, val, e.Name), nil
		}

		return llvm.Value{}, fmt.Errorf("undefined variable: %s", e.Name)

	case *ast.BinaryExpr:
		return g.generateBinaryExpr(e)

	case *ast.UnaryExpr:
		return g.generateUnaryExpr(e)

	case *ast.CallExpr:
		return g.generateCall(e)

	default:
		return llvm.Value{}, fmt.Errorf("unsupported expression type: %T", expr)
	}
}

func (g *CodeGen) generateBinaryExpr(expr *ast.BinaryExpr) (llvm.Value, error) {
	left, err := g.generateExpr(expr.Left)
	if err != nil {
		return llvm.Value{}, err
	}

	right, err := g.generateExpr(expr.Right)
	if err != nil {
		return llvm.Value{}, err
	}

	var runtimeFunc string

	switch expr.Operator {
	case "+":
		runtimeFunc = "yar_add"
	case "-":
		runtimeFunc = "yar_subtract"
	case "*":
		runtimeFunc = "yar_multiply"
	case "/":
		runtimeFunc = "yar_divide"
	case "%":
		runtimeFunc = "yar_modulo"
	case "==":
		runtimeFunc = "yar_eq"
	case "!=":
		runtimeFunc = "yar_neq"
	case "<":
		runtimeFunc = "yar_lt"
	case ">":
		runtimeFunc = "yar_gt"
	case "<=":
		runtimeFunc = "yar_lte"
	case ">=":
		runtimeFunc = "yar_gte"
	case "&&":
		runtimeFunc = "yar_and"
	case "||":
		runtimeFunc = "yar_or"
	default:
		return llvm.Value{}, fmt.Errorf("unsupported binary operator: %s", expr.Operator)
	}

	fn := g.runtimeFuncs[runtimeFunc]
	fnType := g.runtimeFuncTypes[runtimeFunc]

	return g.builder.CreateCall(fnType, fn, []llvm.Value{left, right}, ""), nil
}

func (g *CodeGen) generateUnaryExpr(expr *ast.UnaryExpr) (llvm.Value, error) {
	right, err := g.generateExpr(expr.Right)
	if err != nil {
		return llvm.Value{}, err
	}

	var runtimeFunc string

	switch expr.Operator {
	case "!":
		runtimeFunc = "yar_not"
	case "-":
		runtimeFunc = "yar_negate"
	default:
		return llvm.Value{}, fmt.Errorf("unsupported unary operator: %s", expr.Operator)
	}

	fn := g.runtimeFuncs[runtimeFunc]
	fnType := g.runtimeFuncTypes[runtimeFunc]

	return g.builder.CreateCall(fnType, fn, []llvm.Value{right}, ""), nil
}

func (g *CodeGen) generateCall(expr *ast.CallExpr) (llvm.Value, error) {
	ident, ok := expr.Function.(*ast.Identifier)
	if !ok {
		return llvm.Value{}, fmt.Errorf("function calls only support identifiers for now")
	}

	// Generate arguments
	args := []llvm.Value{}

	for _, arg := range expr.Args {
		val, err := g.generateExpr(arg)
		if err != nil {
			return llvm.Value{}, err
		}

		args = append(args, val)
	}

	// Check if it's a runtime built-in function
	runtimeFuncName := "yar_" + ident.Name
	if fn, ok := g.runtimeFuncs[runtimeFuncName]; ok {
		fnType := g.runtimeFuncTypes[runtimeFuncName]
		return g.builder.CreateCall(fnType, fn, args, ""), nil
	}

	// Check if it's a user-defined function
	userFunc := g.module.NamedFunction(ident.Name)
	if userFunc.IsNil() {
		return llvm.Value{}, fmt.Errorf("undefined function: %s", ident.Name)
	}

	return g.builder.CreateCall(userFunc.GlobalValueType(), userFunc, args, ""), nil
}

func (g *CodeGen) generateAssign(stmt *ast.AssignStmt) error {
	// Simple case: single assignment
	if len(stmt.Targets) != 1 || len(stmt.Values) != 1 {
		return fmt.Errorf("multiple assignment not yet fully implemented")
	}

	target := stmt.Targets[0]

	value, err := g.generateExpr(stmt.Values[0])
	if err != nil {
		return err
	}

	// Allocate or get variable
	var ptr llvm.Value
	if existing, ok := g.variables[target]; ok {
		ptr = existing
	} else {
		// Allocate new variable
		valueType := value.Type()
		ptr = g.builder.CreateAlloca(valueType, target)
		g.variables[target] = ptr
		g.variableTypes[target] = valueType
	}

	g.builder.CreateStore(value, ptr)

	return nil
}

func (g *CodeGen) generateReturn(stmt *ast.ReturnStmt) error {
	if len(stmt.Values) == 0 {
		// Return nil
		nilFn := g.runtimeFuncs["yar_nil"]
		nilType := g.runtimeFuncTypes["yar_nil"]
		nilValue := g.builder.CreateCall(nilType, nilFn, []llvm.Value{}, "")
		g.builder.CreateRet(nilValue)

		return nil
	}

	if len(stmt.Values) == 1 {
		// Single return value
		val, err := g.generateExpr(stmt.Values[0])
		if err != nil {
			return err
		}

		g.builder.CreateRet(val)

		return nil
	}

	// Multiple return values - TODO: implement multi-value return
	return fmt.Errorf("multiple return values not yet fully implemented")
}

func (g *CodeGen) generateIf(stmt *ast.IfStmt) error {
	// Generate condition expression
	condValue, err := g.generateExpr(stmt.Condition)
	if err != nil {
		return err
	}

	// Convert condition to boolean (check if truthy)
	isTruthyFn := g.runtimeFuncs["yar_is_truthy"]
	isTruthyType := g.runtimeFuncTypes["yar_is_truthy"]
	boolValue := g.builder.CreateCall(isTruthyType, isTruthyFn, []llvm.Value{condValue}, "")

	// Create basic blocks
	thenBlock := g.context.AddBasicBlock(g.currentFunc, "then")

	var elseBlock llvm.BasicBlock

	mergeBlock := g.context.AddBasicBlock(g.currentFunc, "merge")

	if stmt.ElseBlock != nil {
		elseBlock = g.context.AddBasicBlock(g.currentFunc, "else")
		g.builder.CreateCondBr(boolValue, thenBlock, elseBlock)
	} else {
		g.builder.CreateCondBr(boolValue, thenBlock, mergeBlock)
	}

	// Generate then block
	g.builder.SetInsertPointAtEnd(thenBlock)

	if err := g.generateBlock(stmt.ThenBlock); err != nil {
		return err
	}
	// Only add branch if block doesn't end with return/break/continue
	// Use GetInsertBlock() to get the current block after nested statements
	currentBlock := g.builder.GetInsertBlock()
	if !g.blockHasTerminator(currentBlock) {
		g.builder.CreateBr(mergeBlock)
	}

	// Generate else block if it exists
	if stmt.ElseBlock != nil {
		g.builder.SetInsertPointAtEnd(elseBlock)

		if err := g.generateBlock(stmt.ElseBlock); err != nil {
			return err
		}
		// Use GetInsertBlock() to get the current block after nested statements
		currentBlock := g.builder.GetInsertBlock()
		if !g.blockHasTerminator(currentBlock) {
			g.builder.CreateBr(mergeBlock)
		}
	}

	// Continue in merge block
	g.builder.SetInsertPointAtEnd(mergeBlock)

	return nil
}

// Helper to check if block has a terminator
func (g *CodeGen) blockHasTerminator(block llvm.BasicBlock) bool {
	lastInst := block.LastInstruction()
	if lastInst.IsNil() {
		return false
	}
	// Check if the last instruction is a terminator (ret, br, switch, unreachable, etc.)
	return !lastInst.IsAReturnInst().IsNil() ||
		!lastInst.IsABranchInst().IsNil() ||
		!lastInst.IsASwitchInst().IsNil() ||
		!lastInst.IsAUnreachableInst().IsNil() ||
		!lastInst.IsAInvokeInst().IsNil()
}

func (g *CodeGen) generateFor(stmt *ast.ForStmt) error {
	// Create basic blocks
	initBlock := g.context.AddBasicBlock(g.currentFunc, "for.init")
	condBlock := g.context.AddBasicBlock(g.currentFunc, "for.cond")
	bodyBlock := g.context.AddBasicBlock(g.currentFunc, "for.body")
	postBlock := g.context.AddBasicBlock(g.currentFunc, "for.post")
	endBlock := g.context.AddBasicBlock(g.currentFunc, "for.end")

	// Save loop exit for break/continue
	oldLoopExit := g.currentLoopExit
	oldLoopPost := g.currentLoopPost
	g.currentLoopExit = endBlock
	g.currentLoopPost = postBlock

	defer func() {
		g.currentLoopExit = oldLoopExit
		g.currentLoopPost = oldLoopPost
	}()

	// Generate init statement
	if stmt.Init != nil {
		g.builder.CreateBr(initBlock)
		g.builder.SetInsertPointAtEnd(initBlock)

		if err := g.generateStmt(stmt.Init); err != nil {
			return err
		}

		g.builder.CreateBr(condBlock)
	} else {
		g.builder.CreateBr(condBlock)
	}

	// Generate condition
	g.builder.SetInsertPointAtEnd(condBlock)

	if stmt.Condition != nil {
		condValue, err := g.generateExpr(stmt.Condition)
		if err != nil {
			return err
		}

		// Convert to boolean
		isTruthyFn := g.runtimeFuncs["yar_is_truthy"]
		isTruthyType := g.runtimeFuncTypes["yar_is_truthy"]
		boolValue := g.builder.CreateCall(isTruthyType, isTruthyFn, []llvm.Value{condValue}, "")

		g.builder.CreateCondBr(boolValue, bodyBlock, endBlock)
	} else {
		// Infinite loop (for { })
		g.builder.CreateBr(bodyBlock)
	}

	// Generate body
	g.builder.SetInsertPointAtEnd(bodyBlock)

	if err := g.generateBlock(stmt.Body); err != nil {
		return err
	}
	// Use GetInsertBlock() to get the current block after nested statements
	currentBlock := g.builder.GetInsertBlock()
	if !g.blockHasTerminator(currentBlock) {
		g.builder.CreateBr(postBlock)
	}

	// Generate post statement
	g.builder.SetInsertPointAtEnd(postBlock)

	if stmt.Post != nil {
		if err := g.generateStmt(stmt.Post); err != nil {
			return err
		}
	}

	g.builder.CreateBr(condBlock)

	// Continue after loop
	g.builder.SetInsertPointAtEnd(endBlock)

	return nil
}

func (g *CodeGen) generateBlock(stmt *ast.BlockStmt) error {
	for _, s := range stmt.Statements {
		if err := g.generateStmt(s); err != nil {
			return err
		}
	}

	return nil
}

func (g *CodeGen) generateFuncDecl(stmt *ast.FuncDecl) error {
	// Create function type: Value* func(Value*, Value*, ...)
	valueType := g.context.StructType([]llvm.Type{
		g.context.Int8Type(),
		llvm.PointerType(g.context.Int8Type(), 0),
	}, false)
	valuePtr := llvm.PointerType(valueType, 0)

	paramTypes := make([]llvm.Type, len(stmt.Params))
	for i := range paramTypes {
		paramTypes[i] = valuePtr
	}

	funcType := llvm.FunctionType(valuePtr, paramTypes, false)
	function := llvm.AddFunction(g.module, stmt.Name, funcType)

	// Create entry block
	entry := g.context.AddBasicBlock(function, "entry")
	g.builder.SetInsertPointAtEnd(entry)

	// Save old function context
	oldFunc := g.currentFunc
	oldVars := g.variables
	oldVarTypes := g.variableTypes
	g.currentFunc = function
	g.variables = make(map[string]llvm.Value)
	g.variableTypes = make(map[string]llvm.Type)

	// Allocate space for parameters
	for i, param := range stmt.Params {
		paramValue := function.Param(i)
		ptr := g.builder.CreateAlloca(valuePtr, param)
		g.builder.CreateStore(paramValue, ptr)
		g.variables[param] = ptr
		g.variableTypes[param] = valuePtr
	}

	// Generate function body
	if err := g.generateBlock(stmt.Body); err != nil {
		return err
	}

	// If no explicit return, return nil
	currentBlock := g.builder.GetInsertBlock()
	if !g.blockHasTerminator(currentBlock) {
		nilFn := g.runtimeFuncs["yar_nil"]
		nilType := g.runtimeFuncTypes["yar_nil"]
		nilValue := g.builder.CreateCall(nilType, nilFn, []llvm.Value{}, "")
		g.builder.CreateRet(nilValue)
	}

	// Restore context
	g.currentFunc = oldFunc
	g.variables = oldVars
	g.variableTypes = oldVarTypes

	// Switch back to main function
	if !oldFunc.IsNil() {
		// Find the last basic block and continue there
		lastBlock := oldFunc.LastBasicBlock()
		g.builder.SetInsertPointAtEnd(lastBlock)
	}

	return nil
}

func (g *CodeGen) EmitIR() string {
	return g.module.String()
}

func (g *CodeGen) WriteToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer func() {
		if err := file.Close(); err != nil {
			// Log the error but don't override the return value
			// since WriteBitcodeToFile error is more important
			_ = err
		}
	}()

	return llvm.WriteBitcodeToFile(g.module, file)
}
