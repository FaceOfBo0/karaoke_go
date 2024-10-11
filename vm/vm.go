package vm

import (
	"fmt"
	"monkey/code"
	"monkey/compiler"
	"monkey/object"
)

const (
	StackSize   = 2048
	GlobalsSize = 65536
)

var Null = &object.Null{}
var trueObj = &object.Boolean{Value: true}
var falseObj = &object.Boolean{Value: false}

type VM struct {
	instructions code.Instructions
	constants    []object.Object
	stack        []object.Object
	globals      []object.Object
	sp           int
}

func NewWithGlobalsStore(bc *compiler.Bytecode, s []object.Object) *VM {
	vm := New(bc)
	vm.globals = s
	return vm
}

func New(bc *compiler.Bytecode) *VM {
	return &VM{
		instructions: bc.Instructions,
		constants:    bc.Constants,
		globals:      make([]object.Object, GlobalsSize),
		stack:        make([]object.Object, StackSize),
		sp:           0,
	}
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])

		switch op {

		case code.OpSetGlobal:
			objIdx := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			vm.globals[objIdx] = vm.stackPop()

		case code.OpGetGlobal:
			objIdx := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			err := vm.stackPush(vm.globals[objIdx])
			if err != nil {
				return err
			}

		case code.OpNull:
			vm.stackPush(Null)

		case code.OpJumpNotTruthy:
			condObj := vm.stackPop()

			if !isTruthy(condObj) {
				ip = int(code.ReadUint16(vm.instructions[ip+1:])) - 1
			} else {
				ip += 2
			}

		case code.OpJump:
			jmpIdx := code.ReadUint16(vm.instructions[ip+1:])
			ip = int(jmpIdx) - 1

		case code.OpConstant:
			constIdx := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

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
			err := vm.execComp(op)
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

func (vm *VM) execComp(op code.Opcode) error {
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
