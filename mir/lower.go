package mir

import (
	"fmt"

	"github.com/yarlson/yarlang/ast"
)

// Lowerer lowers AST to MIR
type Lowerer struct {
	tmpCounter        int
	bbCounter         int
	strCounter        int // Counter for string constants
	module            *Module
	currentFn         *Function
	currentBB         *BasicBlock
	loopExitLabel     string // Label to jump to for break
	loopContinueLabel string // Label to jump to for continue
}

func NewLowerer() *Lowerer {
	return &Lowerer{
		module: &Module{Globals: []Global{}, Functions: []*Function{}},
	}
}

func (l *Lowerer) newTemp() string {
	l.tmpCounter++
	return fmt.Sprintf("t%d", l.tmpCounter)
}

func (l *Lowerer) newBB(name string) *BasicBlock {
	l.bbCounter++
	label := fmt.Sprintf("%s_%d", name, l.bbCounter)

	return &BasicBlock{Label: label, Instrs: []Instruction{}}
}

func (l *Lowerer) emit(instr Instruction) {
	if l.currentBB != nil {
		l.currentBB.Instrs = append(l.currentBB.Instrs, instr)
	}
}

func (l *Lowerer) LowerFile(file *ast.File) *Module {
	for _, item := range file.Items {
		if fn, ok := item.(*ast.FuncDecl); ok {
			l.lowerFunc(fn)
		}
	}

	return l.module
}

func (l *Lowerer) lowerFunc(fn *ast.FuncDecl) {
	mirFn := &Function{
		Name:   fn.Name,
		Params: []Param{},
		RetTy:  l.lowerType(fn.ReturnType),
		Blocks: []*BasicBlock{},
	}

	// Lower parameters
	for _, param := range fn.Params {
		mirFn.Params = append(mirFn.Params, Param{
			Name: param.Name,
			Type: l.lowerType(param.Type),
		})
	}

	l.currentFn = mirFn
	l.currentBB = l.newBB("entry")
	mirFn.Blocks = append(mirFn.Blocks, l.currentBB)

	// Lower body
	l.lowerBlock(fn.Body)

	// Add implicit return for void functions if not already present
	if l.currentBB != nil {
		hasTerminator := false
		if len(l.currentBB.Instrs) > 0 {
			lastInstr := l.currentBB.Instrs[len(l.currentBB.Instrs)-1]
			_, isRet := lastInstr.(*Ret)
			_, isBr := lastInstr.(*Br)
			_, isCondBr := lastInstr.(*CondBr)
			hasTerminator = isRet || isBr || isCondBr
		}

		// If the last instruction is not a terminator and the function is void, add implicit return
		if !hasTerminator {
			if voidType, ok := mirFn.RetTy.(*PrimitiveType); ok && voidType.Name == "void" {
				// Insert DeferRunAll before implicit return
				l.emit(&DeferRunAll{})
				l.emit(&Ret{Value: "", Type: &PrimitiveType{Name: "void"}})
			}
		}
	}

	l.module.Functions = append(l.module.Functions, mirFn)
	l.currentFn = nil
	l.currentBB = nil
}

func (l *Lowerer) lowerBlock(block *ast.Block) {
	for _, stmt := range block.Stmts {
		l.lowerStmt(stmt)
	}
}

func (l *Lowerer) lowerStmt(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.ReturnStmt:
		// Insert DeferRunAll before return
		l.emit(&DeferRunAll{})
		if s.Value != nil {
			val := l.lowerExpr(s.Value)
			l.emit(&Ret{Value: val, Type: &PrimitiveType{Name: "i32"}})
		} else {
			l.emit(&Ret{Type: &PrimitiveType{Name: "void"}})
		}
	case *ast.LetStmt:
		// Allocate on stack
		l.emit(&Alloca{Name: s.Name, Type: &PrimitiveType{Name: "i32"}})
		val := l.lowerExpr(s.Value)
		l.emit(&Store{Value: val, Dest: s.Name, Type: &PrimitiveType{Name: "i32"}})
	case *ast.AssignStmt:
		// Handle assignment to existing variable
		val := l.lowerExpr(s.Value)
		if ident, ok := s.Target.(*ast.Ident); ok {
			l.emit(&Store{Value: val, Dest: ident.Name, Type: &PrimitiveType{Name: "i32"}})
		}
	case *ast.IfStmt:
		l.lowerIfStmt(s)
	case *ast.WhileStmt:
		l.lowerWhileStmt(s)
	case *ast.ForStmt:
		l.lowerForStmt(s)
	case *ast.BreakStmt:
		// Break jumps to the loop exit label
		if l.loopExitLabel == "" {
			panic("break statement outside of loop")
		}
		l.emit(&Br{Label: l.loopExitLabel})
	case *ast.ContinueStmt:
		// Continue jumps to the loop continue label
		if l.loopContinueLabel == "" {
			panic("continue statement outside of loop")
		}
		l.emit(&Br{Label: l.loopContinueLabel})
	case *ast.DeferStmt:
		// Lower the deferred expression (typically a call)
		l.lowerDeferStmt(s)
	case *ast.ExprStmt:
		// Expression statements (like println("hello"))
		l.lowerExpr(s.Expr)
		// Add more statements as needed
	}
}

func (l *Lowerer) lowerExpr(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.BinaryExpr:
		left := l.lowerExpr(e.Left)
		right := l.lowerExpr(e.Right)
		result := l.newTemp()
		op := l.binOpKind(e.Op)
		l.emit(&BinOp{Dest: result, Op: op, Left: left, Right: right, Type: &PrimitiveType{Name: "i32"}})

		return result
	case *ast.Ident:
		// Load from stack
		result := l.newTemp()
		l.emit(&Load{Dest: result, Source: e.Name, Type: &PrimitiveType{Name: "i32"}})

		return result
	case *ast.IntLit:
		return e.Value // Immediate value
	case *ast.StringLit:
		// Create a global string constant and return reference to it
		l.strCounter++
		globalName := fmt.Sprintf(".str.%d", l.strCounter)
		l.module.Globals = append(l.module.Globals, &GlobalString{
			Name:  globalName,
			Value: e.Value,
		})
		return "@" + globalName
	case *ast.CallExpr:
		return l.lowerCallExpr(e)
	case *ast.PropagateExpr:
		return l.lowerPropagateExpr(e)
	// Add more expressions as needed
	default:
		return "undef"
	}
}

func (l *Lowerer) lowerCallExpr(call *ast.CallExpr) string {
	// Get function name from callee
	var calleeName string
	if ident, ok := call.Callee.(*ast.Ident); ok {
		calleeName = ident.Name
	} else {
		// Handle more complex callees later (method calls, etc.)
		return "undef"
	}

	// Lower each argument
	args := make([]string, len(call.Args))
	for i, arg := range call.Args {
		args[i] = l.lowerExpr(arg)
	}

	// Determine return type by looking up the function
	// For now, use a simple heuristic: println is void, others return i32
	var (
		retTy Type
		dest  string
	)

	if calleeName == "println" {
		retTy = &PrimitiveType{Name: "void"}
		dest = "" // void calls don't have a destination
	} else {
		// Look up the function to get its return type
		retTy = l.getFunctionReturnType(calleeName)
		dest = l.newTemp()
	}

	l.emit(&Call{
		Dest:   dest,
		Callee: calleeName,
		Args:   args,
		RetTy:  retTy,
	})

	return dest
}

// getFunctionReturnType looks up the return type of a function in the module
func (l *Lowerer) getFunctionReturnType(name string) Type {
	for _, fn := range l.module.Functions {
		if fn.Name == name {
			return fn.RetTy
		}
	}
	// Default to i32 if not found
	return &PrimitiveType{Name: "i32"}
}

func (l *Lowerer) lowerType(astType ast.Type) Type {
	if astType == nil {
		// nil return type means void (no explicit return type)
		return &PrimitiveType{Name: "void"}
	}

	switch t := astType.(type) {
	case *ast.TypePath:
		if len(t.Path) == 1 {
			return &PrimitiveType{Name: t.Path[0]}
		}

		return &PrimitiveType{Name: "i32"} // Default
	case *ast.VoidType:
		return &PrimitiveType{Name: "void"}
	case *ast.PtrType:
		elem := l.lowerType(t.Elem)
		return &PtrType{Elem: elem}
	default:
		return &PrimitiveType{Name: "i32"}
	}
}

func (l *Lowerer) binOpKind(op string) OpKind {
	switch op {
	case "+":
		return Add
	case "-":
		return Sub
	case "*":
		return Mul
	case "/":
		return Div
	case "%":
		return Mod
	case "&":
		return And
	case "|":
		return Or
	case "^":
		return Xor
	case "<<":
		return Shl
	case ">>":
		return Shr
	case "==":
		return Eq
	case "!=":
		return Ne
	case "<":
		return Lt
	case "<=":
		return Le
	case ">":
		return Gt
	case ">=":
		return Ge
	default:
		return Add
	}
}

func (l *Lowerer) lowerIfStmt(stmt *ast.IfStmt) {
	// Lower condition expression
	cond := l.lowerExpr(stmt.Cond)

	// Create basic blocks
	thenBlock := l.newBB("then")
	var elseBlock *BasicBlock
	mergeBlock := l.newBB("merge")

	if stmt.Else != nil {
		elseBlock = l.newBB("else")
		// Emit conditional branch: if cond then thenBlock else elseBlock
		l.emit(&CondBr{Cond: cond, TrueLabel: thenBlock.Label, FalseLabel: elseBlock.Label})
	} else {
		// Emit conditional branch: if cond then thenBlock else mergeBlock
		l.emit(&CondBr{Cond: cond, TrueLabel: thenBlock.Label, FalseLabel: mergeBlock.Label})
	}

	// Add blocks to function
	l.currentFn.Blocks = append(l.currentFn.Blocks, thenBlock)

	// Lower then branch
	l.currentBB = thenBlock
	l.lowerBlock(stmt.Then)
	// Jump to merge block (only if no terminator already)
	if len(l.currentBB.Instrs) == 0 || !isTerminator(l.currentBB.Instrs[len(l.currentBB.Instrs)-1]) {
		l.emit(&Br{Label: mergeBlock.Label})
	}

	// Lower else branch if it exists
	if stmt.Else != nil {
		l.currentFn.Blocks = append(l.currentFn.Blocks, elseBlock)
		l.currentBB = elseBlock

		// Check if Else is a block or another if statement
		if elseBlock, ok := stmt.Else.(*ast.Block); ok {
			l.lowerBlock(elseBlock)
		} else {
			l.lowerStmt(stmt.Else)
		}

		// Jump to merge block (only if no terminator already)
		if len(l.currentBB.Instrs) == 0 || !isTerminator(l.currentBB.Instrs[len(l.currentBB.Instrs)-1]) {
			l.emit(&Br{Label: mergeBlock.Label})
		}
	}

	// Continue in merge block
	l.currentFn.Blocks = append(l.currentFn.Blocks, mergeBlock)
	l.currentBB = mergeBlock
}

func (l *Lowerer) lowerWhileStmt(stmt *ast.WhileStmt) {
	// Create basic blocks
	condBlock := l.newBB("cond")
	bodyBlock := l.newBB("body")
	exitBlock := l.newBB("exit")

	// Jump from current block to condition block
	l.emit(&Br{Label: condBlock.Label})
	l.currentFn.Blocks = append(l.currentFn.Blocks, condBlock)

	// Lower condition in condition block
	l.currentBB = condBlock
	cond := l.lowerExpr(stmt.Cond)
	l.emit(&CondBr{Cond: cond, TrueLabel: bodyBlock.Label, FalseLabel: exitBlock.Label})

	// Lower body with loop context
	l.currentFn.Blocks = append(l.currentFn.Blocks, bodyBlock)
	l.currentBB = bodyBlock

	// Save previous loop context and set current loop context
	prevExitLabel := l.loopExitLabel
	prevContinueLabel := l.loopContinueLabel
	l.loopExitLabel = exitBlock.Label
	l.loopContinueLabel = condBlock.Label

	l.lowerBlock(stmt.Body)

	// Restore previous loop context
	l.loopExitLabel = prevExitLabel
	l.loopContinueLabel = prevContinueLabel

	// Jump back to condition block (only if no terminator already)
	if len(l.currentBB.Instrs) == 0 || !isTerminator(l.currentBB.Instrs[len(l.currentBB.Instrs)-1]) {
		l.emit(&Br{Label: condBlock.Label})
	}

	// Continue in exit block
	l.currentFn.Blocks = append(l.currentFn.Blocks, exitBlock)
	l.currentBB = exitBlock
}

func (l *Lowerer) lowerForStmt(stmt *ast.ForStmt) {
	// For v0.4, handle simplified `for i in 0..n` range form
	// Range is represented as BinaryExpr with ".." operator
	rangeExpr, ok := stmt.Iter.(*ast.BinaryExpr)
	if !ok || rangeExpr.Op != ".." {
		// For now, only support range expressions (start..end)
		return
	}

	// Create iterator variable
	iterVar := stmt.Val
	l.emit(&Alloca{Name: iterVar, Type: &PrimitiveType{Name: "i32"}})

	// Initialize iterator to start value
	start := l.lowerExpr(rangeExpr.Left)
	l.emit(&Store{Value: start, Dest: iterVar, Type: &PrimitiveType{Name: "i32"}})

	// Lower end value once (may be expression)
	endVal := l.lowerExpr(rangeExpr.Right)

	// Create basic blocks
	condBlock := l.newBB("cond")
	bodyBlock := l.newBB("body")
	exitBlock := l.newBB("exit")

	// Jump from current block to condition block
	l.emit(&Br{Label: condBlock.Label})
	l.currentFn.Blocks = append(l.currentFn.Blocks, condBlock)

	// Lower condition in condition block: i < end
	l.currentBB = condBlock
	iterVal := l.newTemp()
	l.emit(&Load{Dest: iterVal, Source: iterVar, Type: &PrimitiveType{Name: "i32"}})

	condResult := l.newTemp()
	l.emit(&BinOp{Dest: condResult, Op: Lt, Left: iterVal, Right: endVal, Type: &PrimitiveType{Name: "i32"}})
	l.emit(&CondBr{Cond: condResult, TrueLabel: bodyBlock.Label, FalseLabel: exitBlock.Label})

	// Lower body with loop context
	l.currentFn.Blocks = append(l.currentFn.Blocks, bodyBlock)
	l.currentBB = bodyBlock

	// Save previous loop context and set current loop context
	prevExitLabel := l.loopExitLabel
	prevContinueLabel := l.loopContinueLabel
	l.loopExitLabel = exitBlock.Label
	l.loopContinueLabel = condBlock.Label

	l.lowerBlock(stmt.Body)

	// Restore previous loop context
	l.loopExitLabel = prevExitLabel
	l.loopContinueLabel = prevContinueLabel

	// Increment iterator: i = i + 1
	iterVal2 := l.newTemp()
	l.emit(&Load{Dest: iterVal2, Source: iterVar, Type: &PrimitiveType{Name: "i32"}})
	incResult := l.newTemp()
	l.emit(&BinOp{Dest: incResult, Op: Add, Left: iterVal2, Right: "1", Type: &PrimitiveType{Name: "i32"}})
	l.emit(&Store{Value: incResult, Dest: iterVar, Type: &PrimitiveType{Name: "i32"}})

	// Jump back to condition block
	l.emit(&Br{Label: condBlock.Label})

	// Continue in exit block
	l.currentFn.Blocks = append(l.currentFn.Blocks, exitBlock)
	l.currentBB = exitBlock
}

// isTerminator checks if an instruction is a terminator (Ret, Br, CondBr)
func isTerminator(instr Instruction) bool {
	switch instr.(type) {
	case *Ret, *Br, *CondBr:
		return true
	default:
		return false
	}
}

// lowerDeferStmt lowers a defer statement to DeferPush instruction
func (l *Lowerer) lowerDeferStmt(stmt *ast.DeferStmt) {
	// The deferred expression must be a call expression
	callExpr, ok := stmt.Expr.(*ast.CallExpr)
	if !ok {
		// For now, only support defer with call expressions
		return
	}

	// Get function name from callee
	var calleeName string
	if ident, ok := callExpr.Callee.(*ast.Ident); ok {
		calleeName = ident.Name
	} else {
		// Handle more complex callees later
		return
	}

	// Lower each argument
	args := make([]string, len(callExpr.Args))
	for i, arg := range callExpr.Args {
		args[i] = l.lowerExpr(arg)
	}

	// Determine return type
	retTy := l.getFunctionReturnType(calleeName)

	// Create the Call instruction (but don't emit it directly)
	call := &Call{
		Dest:   "", // Deferred calls are always void (result is discarded)
		Callee: calleeName,
		Args:   args,
		RetTy:  retTy,
	}

	// Emit DeferPush with the call
	l.emit(&DeferPush{Call: call})
}

// lowerPropagateExpr lowers ? operator for Result<T,E> error propagation
// Following v0.4 spec section 11:
//   t = X
//   if is_err(t) { return Err(extract_err(t)) }
//   v = extract_ok(t)
func (l *Lowerer) lowerPropagateExpr(expr *ast.PropagateExpr) string {
	// Lower the inner expression to get the Result<T,E> value
	// This evaluates X and gets the result in current block
	resultVal := l.lowerExpr(expr.Expr)

	// Create basic blocks for control flow
	checkBlock := l.newBB("check")
	errorBlock := l.newBB("error")
	okBlock := l.newBB("ok")

	// In the current block, jump to check block
	l.emit(&Br{Label: checkBlock.Label})

	// Add check block to function
	l.currentFn.Blocks = append(l.currentFn.Blocks, checkBlock)
	l.currentBB = checkBlock

	// TODO: For full implementation, need runtime support for:
	// - is_err() function to check if Result is Err variant
	// - extract_err() to get error value
	// - extract_ok() to get ok value
	// - Proper Result<T,E> enum type representation
	//
	// For v0.4, we generate the control flow structure with stub condition
	// The actual runtime functions will be implemented when enum support is added

	// Generate is_err check (stub: always false for now)
	// In full implementation: %is_err = call i1 @is_err(%result_type %t)
	isErrTemp := l.newTemp()
	// For now, create a stub boolean check that always evaluates to false
	// This shows the structure without breaking codegen
	l.emit(&BinOp{
		Dest:  isErrTemp,
		Op:    Eq,
		Left:  resultVal,
		Right: resultVal,
		Type:  &PrimitiveType{Name: "i1"},
	})

	// Conditional branch based on error check
	l.emit(&CondBr{
		Cond:       isErrTemp,
		TrueLabel:  errorBlock.Label,
		FalseLabel: okBlock.Label,
	})

	// Error block: extract error and return early
	l.currentFn.Blocks = append(l.currentFn.Blocks, errorBlock)
	l.currentBB = errorBlock

	// TODO: Full implementation would be:
	// %err = call %error_type @extract_err(%result_type %t)
	// %result = call %result_type @make_err(%error_type %err)
	// ret %result_type %result

	// For now, emit stub early return
	// Insert DeferRunAll before return
	l.emit(&DeferRunAll{})
	l.emit(&Ret{
		Value: resultVal,
		Type:  &PrimitiveType{Name: "i32"},
	})

	// Ok block: extract ok value and continue
	l.currentFn.Blocks = append(l.currentFn.Blocks, okBlock)
	l.currentBB = okBlock

	// TODO: Full implementation would be:
	// %ok_val = call %T @extract_ok(%result_type %t)
	// return %ok_val

	// For now, just return the result value as-is
	// This represents the extracted ok value
	okVal := resultVal

	return okVal
}
