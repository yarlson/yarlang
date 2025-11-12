package codegen

import (
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
	"github.com/yarlson/yarlang/mir"
)

// Codegen generates LLVM IR from MIR
type Codegen struct {
	mod       *ir.Module
	currentFn *ir.Func
	locals    map[string]*ir.InstAlloca
	values    map[string]value.Value // Track all SSA values
	blocks    map[string]*ir.Block   // Map from label to LLVM block
	globals   map[string]*ir.Global  // Map from global name to LLVM global
}

func NewCodegen() *Codegen {
	return &Codegen{
		mod:     ir.NewModule(),
		locals:  make(map[string]*ir.InstAlloca),
		values:  make(map[string]value.Value),
		blocks:  make(map[string]*ir.Block),
		globals: make(map[string]*ir.Global),
	}
}

func (cg *Codegen) GenModule(mirMod *mir.Module) *ir.Module {
	// Generate global constants first
	for _, global := range mirMod.Globals {
		cg.genGlobal(global)
	}

	// Then generate functions
	for _, fn := range mirMod.Functions {
		cg.genFunction(fn)
	}

	return cg.mod
}

func (cg *Codegen) genGlobal(global mir.Global) {
	switch g := global.(type) {
	case *mir.GlobalString:
		// Create a global string constant
		// String content + null terminator
		content := g.Value + "\x00"

		// Create the constant string
		strConst := constant.NewCharArrayFromString(content)

		// Create global variable
		global := cg.mod.NewGlobalDef(g.Name, strConst)
		global.Linkage = enum.LinkagePrivate
		global.UnnamedAddr = enum.UnnamedAddrUnnamedAddr

		// Store in globals map
		cg.globals[g.Name] = global

		// Also store in values map for easy lookup
		cg.values[g.Name] = global
	}
}

func (cg *Codegen) genFunction(mirFn *mir.Function) {
	// Convert MIR types to LLVM types
	params := make([]*ir.Param, len(mirFn.Params))
	for i, p := range mirFn.Params {
		params[i] = ir.NewParam(p.Name, cg.toLLVMType(p.Type))
		// Track function parameters as values
		cg.values[p.Name] = params[i]
	}

	retTy := cg.toLLVMType(mirFn.RetTy)
	fn := cg.mod.NewFunc(mirFn.Name, retTy, params...)
	cg.currentFn = fn

	// Create all LLVM blocks first (so we can reference them in branches)
	for _, bb := range mirFn.Blocks {
		llvmBlock := fn.NewBlock(bb.Label)
		cg.blocks[bb.Label] = llvmBlock
	}

	// Materialize parameters on the stack when MIR expects loads/stores by name
	if len(mirFn.Blocks) > 0 && len(mirFn.Params) > 0 {
		entryBlock := cg.blocks[mirFn.Blocks[0].Label]
		for _, param := range mirFn.Params {
			llvmTy := cg.toLLVMType(param.Type)
			alloca := entryBlock.NewAlloca(llvmTy)
			alloca.SetName(param.Name + ".addr")
			entryBlock.NewStore(cg.values[param.Name], alloca)
			cg.locals[param.Name] = alloca
		}
	}

	// Generate instructions for each block
	for _, bb := range mirFn.Blocks {
		llvmBlock := cg.blocks[bb.Label]
		cg.genBasicBlock(bb, llvmBlock)
	}

	cg.currentFn = nil
	cg.locals = make(map[string]*ir.InstAlloca)
	cg.values = make(map[string]value.Value)
	cg.blocks = make(map[string]*ir.Block)
}

func (cg *Codegen) genBasicBlock(mirBB *mir.BasicBlock, llvmBB *ir.Block) {
	for _, instr := range mirBB.Instrs {
		switch i := instr.(type) {
		case *mir.Alloca:
			alloca := llvmBB.NewAlloca(cg.toLLVMType(i.Type))
			alloca.SetName(i.Name)
			cg.locals[i.Name] = alloca
			cg.values[i.Name] = alloca
		case *mir.Load:
			src := cg.locals[i.Source]
			load := llvmBB.NewLoad(cg.toLLVMType(i.Type), src)
			load.SetName(i.Dest)
			cg.values[i.Dest] = load
		case *mir.Store:
			// Get the value to store
			val := cg.getValue(i.Value, i.Type, llvmBB)
			dest := cg.locals[i.Dest]
			llvmBB.NewStore(val, dest)
		case *mir.BinOp:
			// Get operands
			left := cg.getValue(i.Left, i.Type, llvmBB)
			right := cg.getValue(i.Right, i.Type, llvmBB)

			var result value.Value
			// Handle comparison operations (return i1/bool)
			if i.Op >= mir.Eq && i.Op <= mir.Ge {
				result = llvmBB.NewICmp(cg.opToICmpPred(i.Op), left, right)
			} else {
				// Handle arithmetic operations
				switch i.Op {
				case mir.Add:
					result = llvmBB.NewAdd(left, right)
				case mir.Sub:
					result = llvmBB.NewSub(left, right)
				case mir.Mul:
					result = llvmBB.NewMul(left, right)
				case mir.Div:
					result = llvmBB.NewSDiv(left, right)
				case mir.Mod:
					result = llvmBB.NewSRem(left, right)
				case mir.And:
					result = llvmBB.NewAnd(left, right)
				case mir.Or:
					result = llvmBB.NewOr(left, right)
				case mir.Xor:
					result = llvmBB.NewXor(left, right)
				case mir.Shl:
					result = llvmBB.NewShl(left, right)
				case mir.Shr:
					result = llvmBB.NewAShr(left, right)
				}
			}
			if result != nil {
				result.(interface{ SetName(string) }).SetName(i.Dest)
				cg.values[i.Dest] = result
			}
		case *mir.Call:
			args, argTypes := cg.buildCallArgs(i, llvmBB)
			if cg.genBuiltinCall(i, llvmBB, args) {
				continue
			}

			callee := cg.getFunctionByName(i.Callee)

			// If function not found, create external declaration using inferred arg types
			if callee == nil {
				retTy := cg.toLLVMType(i.RetTy)
				params := make([]*ir.Param, len(argTypes))
				for idx, argTy := range argTypes {
					params[idx] = ir.NewParam("", argTy)
				}
				callee = cg.mod.NewFunc(i.Callee, retTy, params...)
			}

			call := llvmBB.NewCall(callee, args...)

			if i.Dest != "" {
				call.SetName(i.Dest)
				cg.values[i.Dest] = call
			}
		case *mir.Br:
			// Unconditional branch
			targetBlock := cg.blocks[i.Label]
			llvmBB.NewBr(targetBlock)
		case *mir.CondBr:
			// Conditional branch
			cond := cg.getValue(i.Cond, &mir.PrimitiveType{Name: "bool"}, llvmBB)
			trueBlock := cg.blocks[i.TrueLabel]
			falseBlock := cg.blocks[i.FalseLabel]
			llvmBB.NewCondBr(cond, trueBlock, falseBlock)
		case *mir.Ret:
			if i.Value == "" {
				llvmBB.NewRet(nil)
			} else {
				// Check if it's a tracked value (from Load, etc.)
				if val, ok := cg.values[i.Value]; ok {
					llvmBB.NewRet(val)
				} else {
					// Otherwise, try to load from local
					if alloca, ok := cg.locals[i.Value]; ok {
						val := llvmBB.NewLoad(cg.toLLVMType(i.Type), alloca)
						llvmBB.NewRet(val)
					} else {
						// Must be a constant
						val := cg.parseConstant(i.Value, i.Type)
						llvmBB.NewRet(val)
					}
				}
			}
		case *mir.DeferPush:
			// TODO: proper defer runtime support needed
			// For v0.4, simplified implementation - defer is not yet fully functional
			// This would need a defer stack and runtime support
			// For now, generate a comment as a placeholder
		case *mir.DeferRunAll:
			// TODO: proper defer runtime support needed
			// For v0.4, simplified implementation - defer is not yet fully functional
			// This would need to iterate the defer stack in LIFO order
			// For now, generate a comment as a placeholder
		}
	}
}

func (cg *Codegen) buildCallArgs(call *mir.Call, block *ir.Block) ([]value.Value, []types.Type) {
	args := make([]value.Value, len(call.Args))
	argTypes := make([]types.Type, len(call.Args))
	for idx, arg := range call.Args {
		val := cg.getValue(arg, &mir.PrimitiveType{Name: "i32"}, block)
		args[idx] = val
		argTypes[idx] = val.Type()
	}
	return args, argTypes
}

func (cg *Codegen) genBuiltinCall(call *mir.Call, block *ir.Block, args []value.Value) bool {
	switch call.Callee {
	case "println":
		return cg.lowerPrintln(block, args)
	default:
		return false
	}
}

func (cg *Codegen) lowerPrintln(block *ir.Block, args []value.Value) bool {
	if len(args) != 1 {
		return false
	}

	arg := args[0]
	switch t := arg.Type().(type) {
	case *types.PointerType:
		if !isI8Pointer(t) {
			return false
		}
		fn := cg.getOrCreateFunction("println", types.Void, []types.Type{t})
		block.NewCall(fn, arg)
		return true
	case *types.IntType:
		name := "println_i32"
		if t.BitSize == 1 {
			name = "println_bool"
		}
		fn := cg.getOrCreateFunction(name, types.Void, []types.Type{t})
		block.NewCall(fn, arg)
		return true
	default:
		return false
	}
}

func (cg *Codegen) getFunctionByName(name string) *ir.Func {
	for _, fn := range cg.mod.Funcs {
		if fn.Name() == name {
			return fn
		}
	}
	return nil
}

func (cg *Codegen) getOrCreateFunction(name string, retTy types.Type, paramTypes []types.Type) *ir.Func {
	if fn := cg.getFunctionByName(name); fn != nil {
		return fn
	}
	params := make([]*ir.Param, len(paramTypes))
	for idx, ty := range paramTypes {
		params[idx] = ir.NewParam("", ty)
	}
	return cg.mod.NewFunc(name, retTy, params...)
}

func isI8Pointer(ty types.Type) bool {
	ptr, ok := ty.(*types.PointerType)
	if !ok {
		return false
	}
	if elem, ok := ptr.ElemType.(*types.IntType); ok {
		return elem.BitSize == 8
	}
	return false
}

// opToICmpPred converts MIR comparison operations to LLVM icmp predicates
func (cg *Codegen) opToICmpPred(op mir.OpKind) enum.IPred {
	switch op {
	case mir.Eq:
		return enum.IPredEQ
	case mir.Ne:
		return enum.IPredNE
	case mir.Lt:
		return enum.IPredSLT
	case mir.Le:
		return enum.IPredSLE
	case mir.Gt:
		return enum.IPredSGT
	case mir.Ge:
		return enum.IPredSGE
	default:
		return enum.IPredEQ
	}
}

// getValue gets an LLVM value from a MIR value string
// Handles both constants (like "42") and local variables (like "x")
func (cg *Codegen) getValue(valueStr string, ty mir.Type, block *ir.Block) value.Value {
	// Check if it's a global reference (starts with @)
	if len(valueStr) > 0 && valueStr[0] == '@' {
		globalName := valueStr[1:] // Remove @ prefix
		if global, ok := cg.globals[globalName]; ok {
			// For string constants, we need to get a pointer to the first element
			// Use getelementptr to convert [N x i8]* to i8*
			globalType := global.ContentType
			if arrayType, ok := globalType.(*types.ArrayType); ok {
				// Create indices for getelementptr: (i32 0, i32 0)
				zero := constant.NewInt(types.I32, 0)
				indices := []value.Value{zero, zero}

				// Get pointer to first element
				gep := block.NewGetElementPtr(arrayType, global, indices...)
				return gep
			}
			return global
		}
	}

	// First check if it's an SSA value (from Load, BinOp, etc.)
	if val, ok := cg.values[valueStr]; ok {
		return val
	}

	// Check if it's a local variable (alloca) that needs to be loaded
	if alloca, ok := cg.locals[valueStr]; ok {
		// Load from the local variable
		return block.NewLoad(cg.toLLVMType(ty), alloca)
	}

	// Otherwise, treat it as a constant
	return cg.parseConstant(valueStr, ty)
}

// parseConstant parses a constant value from a string
func (cg *Codegen) parseConstant(value string, ty mir.Type) constant.Constant {
	llvmType := cg.toLLVMType(ty)

	// Parse integer constants
	if intType, ok := llvmType.(*types.IntType); ok {
		var intVal int64
		// Simple parsing - in a real implementation, handle errors
		if len(value) > 0 {
			negative := false
			str := value
			if value[0] == '-' {
				negative = true
				str = value[1:]
			}
			for _, ch := range str {
				if ch >= '0' && ch <= '9' {
					intVal = intVal*10 + int64(ch-'0')
				}
			}
			if negative {
				intVal = -intVal
			}
		}
		return constant.NewInt(intType, intVal)
	}

	// Default to zero for unhandled types
	return constant.NewInt(types.I32, 0)
}

func (cg *Codegen) toLLVMType(mirType mir.Type) types.Type {
	switch t := mirType.(type) {
	case *mir.PrimitiveType:
		switch t.Name {
		case "i8":
			return types.I8
		case "i16":
			return types.I16
		case "i32":
			return types.I32
		case "i64":
			return types.I64
		case "u8":
			return types.I8
		case "u16":
			return types.I16
		case "u32":
			return types.I32
		case "u64":
			return types.I64
		case "f32":
			return types.Float
		case "f64":
			return types.Double
		case "bool":
			return types.I1
		case "void":
			return types.Void
		default:
			return types.I32
		}
	case *mir.PtrType:
		elem := cg.toLLVMType(t.Elem)
		return types.NewPointer(elem)
	default:
		return types.I32
	}
}
