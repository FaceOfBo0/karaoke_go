package vm

import (
	"fmt"
	"karaoke/code"
	"karaoke/compiler"
	"karaoke/object"
)

const (
	MaxFrames   = 1024
	StackSize   = 2048
	GlobalsSize = 65536
)

var Null = &object.Null{}
var trueObj = &object.Boolean{Value: true}
var falseObj = &object.Boolean{Value: false}

type VM struct {
	frames    []*Frame
	framesPtr int
	constants []object.Object
	stack     []object.Object
	globals   []object.Object
	sp        int
}

func (vm *VM) currenFrame() *Frame {
	return vm.frames[vm.framesPtr-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.framesPtr] = f
	vm.framesPtr++
}

func (vm *VM) popFrame() *Frame {
	vm.framesPtr--
	return vm.frames[vm.framesPtr]
}

func NewWithGlobalsStore(bc *compiler.Bytecode, s []object.Object) *VM {
	vm := New(bc)
	vm.globals = s
	return vm
}

func New(bc *compiler.Bytecode) *VM {
	mainFrame := NewFrame(&object.CompiledFunction{Instructions: bc.Instructions}, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants: bc.Constants,
		globals:   make([]object.Object, GlobalsSize),
		stack:     make([]object.Object, StackSize),
		sp:        0,
		framesPtr: 1,
		frames:    frames,
	}
}

func (vm *VM) Run() error {
	var ip int
	var ins code.Instructions
	var op code.Opcode

	for vm.currenFrame().ip < len(vm.currenFrame().Instructions())-1 {
		vm.currenFrame().ip++

		ip = vm.currenFrame().ip
		ins = vm.currenFrame().Instructions()
		op = code.Opcode(ins[ip])

		switch op {

		case code.OpReturnValue:
			retVal := vm.stackPop()

			frame := vm.popFrame()
			vm.sp = frame.basePtr - 1

			err := vm.stackPush(retVal)
			if err != nil {
				return err
			}

		case code.OpReturn:
			frame := vm.popFrame()
			vm.sp = frame.basePtr - 1

			err := vm.stackPush(Null)
			if err != nil {
				return err
			}

		case code.OpCall:

			fnObj, ok := vm.stack[vm.sp-1].(*object.CompiledFunction)
			if !ok {
				return fmt.Errorf("wrong type for compiled function: %s", fnObj.Type())
			}

			funcFrame := NewFrame(fnObj, vm.sp)
			vm.pushFrame(funcFrame)
			vm.sp = funcFrame.basePtr + fnObj.NumLocals

		case code.OpIndex:
			idxObj := vm.stackPop()
			arrObj := vm.stackPop()

			switch arrObj := arrObj.(type) {
			case *object.Hash:
				idx, ok := idxObj.(object.Hashable)
				if !ok {
					return fmt.Errorf("unknown index type for hash: %s", idxObj.Type())
				}

				hashPair, ok := arrObj.Pairs[idx.HashKey()]
				if !ok {
					vm.stackPush(Null)
				} else {
					err := vm.stackPush(hashPair.Value)
					if err != nil {
						return err
					}
				}

			case *object.Array:
				idx, ok := idxObj.(*object.Integer)
				if !ok {
					return fmt.Errorf("unknown index type for array: %s", idxObj.Type())
				}

				max := int64(len(arrObj.Elements) - 1)
				if idx.Value < 0 || idx.Value > max {
					vm.stackPush(Null)
				} else {
					err := vm.stackPush(arrObj.Elements[idx.Value])
					if err != nil {
						return err
					}
				}
			}

		case code.OpHash:
			lenHash := int(code.ReadUint16(ins[ip+1:]))
			vm.currenFrame().ip += 2

			pairs := make(map[object.HashKey]object.HashPair, lenHash)
			for i := 0; i < lenHash; i++ {
				val := vm.stackPop()
				key := vm.stackPop()

				hashKey, ok := key.(object.Hashable)
				if !ok {
					return fmt.Errorf("unusable as hash key: %s", key.Type())
				}

				pairs[hashKey.HashKey()] = object.HashPair{Key: key, Value: val}
			}

			err := vm.stackPush(&object.Hash{Pairs: pairs})
			if err != nil {
				return err
			}

		case code.OpArray:
			lenArr := int(code.ReadUint16(ins[ip+1:]))
			vm.currenFrame().ip += 2

			array := make([]object.Object, lenArr)
			for i := lenArr - 1; i >= 0; i-- {
				array[i] = vm.stackPop()
			}

			err := vm.stackPush(&object.Array{Elements: array})
			if err != nil {
				return err
			}

		case code.OpSetGlobal:
			objIdx := code.ReadUint16(ins[ip+1:])
			vm.currenFrame().ip += 2

			vm.globals[objIdx] = vm.stackPop()

		case code.OpGetGlobal:
			objIdx := code.ReadUint16(ins[ip+1:])
			vm.currenFrame().ip += 2

			err := vm.stackPush(vm.globals[objIdx])
			if err != nil {
				return err
			}

		case code.OpSetLocal:
			objIdx := uint8(ins[ip+1])
			vm.currenFrame().ip += 1

			vm.stack[vm.currenFrame().basePtr+int(objIdx)] = vm.stackPop()

		case code.OpGetLocal:
			objIdx := uint8(ins[ip+1])
			vm.currenFrame().ip += 1

			err := vm.stackPush(vm.stack[vm.currenFrame().basePtr+int(objIdx)])
			if err != nil {
				return err
			}

		case code.OpNull:
			vm.stackPush(Null)

		case code.OpJumpNotTruthy:
			condObj := vm.stackPop()

			if !isTruthy(condObj) {
				vm.currenFrame().ip = int(code.ReadUint16(ins[ip+1:])) - 1
			} else {
				vm.currenFrame().ip += 2
			}

		case code.OpJump:
			jmpIdx := code.ReadUint16(ins[ip+1:])
			vm.currenFrame().ip = int(jmpIdx) - 1

		case code.OpConstant:
			constIdx := code.ReadUint16(ins[ip+1:])
			vm.currenFrame().ip += 2

			err := vm.stackPush(vm.constants[constIdx])
			if err != nil {
				return err
			}

		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			err := vm.execBinaryOp(op)
			if err != nil {
				return err
			}

		case code.OpPop:
			vm.stackPop()

		case code.OpTrue:
			err := vm.stackPush(trueObj)
			if err != nil {
				return err
			}

		case code.OpFalse:
			err := vm.stackPush(falseObj)
			if err != nil {
				return err
			}

		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan:
			err := vm.execComparison(op)
			if err != nil {
				return err
			}

		case code.OpBang:
			err := vm.execBangOp()
			if err != nil {
				return err
			}

		case code.OpMinus:
			argObj := vm.stackPop()

			if argObj.Type() != object.INTEGER_OBJ {
				return fmt.Errorf("unsupported type for negation: %s", argObj.Type())
			}

			value := argObj.(*object.Integer).Value
			vm.stackPush(&object.Integer{Value: -value})
		}
	}
	return nil
}

func (vm *VM) execBinaryIntOp(operand code.Opcode, left object.Object, right object.Object) error {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value
	var result int64
	switch operand {
	case code.OpAdd:
		result = leftVal + rightVal
	case code.OpSub:
		result = leftVal - rightVal
	case code.OpMul:
		result = leftVal * rightVal
	case code.OpDiv:
		result = leftVal / rightVal
	default:
		return fmt.Errorf("unknown Integer operation: %d", operand)
	}

	return vm.stackPush(&object.Integer{Value: result})

}

func (vm *VM) execBinaryStrOp(op code.Opcode, left, right object.Object) error {
	if op != code.OpAdd {
		return fmt.Errorf("unknown string operator: %d", op)
	}

	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	return vm.stackPush(&object.String{Value: leftVal + rightVal})
}

func (vm *VM) execBinaryOp(operand code.Opcode) error {
	rightObj := vm.stackPop()
	leftObj := vm.stackPop()
	rightType := rightObj.Type()
	leftType := leftObj.Type()

	if rightType == object.INTEGER_OBJ && leftType == object.INTEGER_OBJ {
		return vm.execBinaryIntOp(operand, leftObj, rightObj)
	} else if rightType == object.STRING_OBJ && leftType == object.STRING_OBJ {
		return vm.execBinaryStrOp(operand, leftObj, rightObj)
	}

	return fmt.Errorf("unsupported types for binary operation: %s %s",
		leftType, rightType)
}

func (vm *VM) execBangOp() error {
	operand := vm.stackPop()
	switch operand {
	case Null:
		return vm.stackPush(trueObj)
	case falseObj:
		return vm.stackPush(trueObj)
	default:
		return vm.stackPush(falseObj)
	}
}

func (vm *VM) execComparison(op code.Opcode) error {
	rightObj := vm.stackPop()
	leftObj := vm.stackPop()

	if leftObj.Type() == object.INTEGER_OBJ && rightObj.Type() == object.INTEGER_OBJ {
		leftVal := leftObj.(*object.Integer).Value
		rightVal := rightObj.(*object.Integer).Value

		switch op {
		case code.OpEqual:
			return vm.stackPush(nativeBoolToBoolObj(leftVal == rightVal))

		case code.OpNotEqual:
			return vm.stackPush(nativeBoolToBoolObj(leftVal != rightVal))

		case code.OpGreaterThan:
			return vm.stackPush(nativeBoolToBoolObj(leftVal > rightVal))
		default:
			return fmt.Errorf("unknown operator: %d", op)
		}
	}

	switch op {
	case code.OpEqual:
		return vm.stackPush(nativeBoolToBoolObj(leftObj == rightObj))
	case code.OpNotEqual:
		return vm.stackPush(nativeBoolToBoolObj(leftObj != rightObj))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)", op, leftObj.Type(), rightObj.Type())
	}
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Null:
		return false
	default:
		return true

	}
}

func nativeBoolToBoolObj(input bool) *object.Boolean {
	if input {
		return trueObj
	}
	return falseObj
}

func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) stackPop() object.Object {
	elem := vm.stack[vm.sp-1]
	vm.sp--
	return elem
}

func (vm *VM) stackPush(elem object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}
	vm.stack[vm.sp] = elem
	vm.sp++

	return nil
}
