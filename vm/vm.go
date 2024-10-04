package vm

import (
	"fmt"
	"monkey/code"
	"monkey/compiler"
	"monkey/object"
)

const StackSize = 2048

type VM struct {
	instructions code.Instructions
	constants    []object.Object
	stack        []object.Object
	sp           int
}

func New(bc *compiler.Bytecode) *VM {
	return &VM{
		instructions: bc.Instructions,
		constants:    bc.Constants,
		stack:        make([]object.Object, StackSize),
		sp:           0,
	}
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])

		switch op {
		case code.OpConstant:
			constIdx := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			err := vm.stackPush(vm.constants[constIdx])
			if err != nil {
				return err
			}

		case code.OpAdd:
			rightObj := vm.stackPop().(*object.Integer)
			leftObj := vm.stackPop().(*object.Integer)
			vm.stackPush(&object.Integer{Value: leftObj.Value + rightObj.Value})

		case code.OpSub:
			rightObj := vm.stackPop().(*object.Integer)
			leftObj := vm.stackPop().(*object.Integer)
			vm.stackPush(&object.Integer{Value: leftObj.Value - rightObj.Value})

		case code.OpMult:
			rightObj := vm.stackPop().(*object.Integer)
			leftObj := vm.stackPop().(*object.Integer)
			vm.stackPush(&object.Integer{Value: leftObj.Value * rightObj.Value})

		case code.OpDiv:
			rightObj := vm.stackPop().(*object.Integer)
			leftObj := vm.stackPop().(*object.Integer)
			vm.stackPush(&object.Integer{Value: leftObj.Value / rightObj.Value})
		case code.OpPop:
			vm.stackPop()
		}
	}
	return nil
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
