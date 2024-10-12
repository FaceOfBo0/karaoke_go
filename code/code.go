package code

import (
	"encoding/binary"
	"fmt"
	"strings"
)

type Instructions []byte

func (inst Instructions) String() string {
	offset := 0
	output := ""
	for offset < len(inst) {
		def, err := Lookup(inst[offset])
		if err != nil {
			fmt.Printf("definition not found: %q\n", err)
			continue
		}
		operands, bytesRead := ReadOperands(def, inst[offset+1:])
		if len(operands) > 0 {
			output += fmt.Sprintf("%04d %s %s\n", offset, def.Name, strings.Trim(fmt.Sprint(operands), "[]"))
		} else {
			output += fmt.Sprintf("%04d %s\n", offset, def.Name)
		}

		offset += bytesRead + 1
	}

	return output
}

type Opcode byte

type Definition struct {
	Name          string
	OperandWidths []int
}

const (
	OpConstant Opcode = iota
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpPop
	OpTrue
	OpFalse
	OpEqual
	OpNotEqual
	OpGreaterThan
	OpMinus
	OpBang
	OpJumpNotTruthy
	OpJump
	OpNull
	OpGetGlobal
	OpSetGlobal
	OpArray
	OpHash
)

var definitions = map[Opcode]*Definition{
	OpConstant:      {"OpConstant", []int{2}},
	OpJumpNotTruthy: {"OpJumpNotTruthy", []int{2}},
	OpJump:          {"OpJump", []int{2}},
	OpGetGlobal:     {"OpGetGlobal", []int{2}},
	OpSetGlobal:     {"OpSetGlobal", []int{2}},
	OpArray:         {"OpArray", []int{2}},
	OpHash:          {"OpHash", []int{2}},
	OpAdd:           {"OpAdd", []int{}},
	OpSub:           {"OpSub", []int{}},
	OpMul:           {"OpMul", []int{}},
	OpDiv:           {"OpDiv", []int{}},
	OpPop:           {"OpPop", []int{}},
	OpTrue:          {"OpTrue", []int{}},
	OpFalse:         {"OpFalse", []int{}},
	OpEqual:         {"OpEqual", []int{}},
	OpNotEqual:      {"OpNotEqual", []int{}},
	OpGreaterThan:   {"OpGreaterThan", []int{}},
	OpMinus:         {"OpMinus", []int{}},
	OpBang:          {"OpBang", []int{}},
	OpNull:          {"OpNull", []int{}},
}

func ReadUint16(inst []byte) uint16 {
	return binary.BigEndian.Uint16(inst)
}

func PutUint16(inst []byte, val uint16) {
	binary.BigEndian.PutUint16(inst, val)
}

func ReadOperands(def *Definition, inst []byte) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0
	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(inst[offset:]))
		}

		offset += width
	}
	return operands, offset
}

func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instLen := 1
	for _, el := range def.OperandWidths {
		instLen += el
	}

	inst := make([]byte, instLen)
	inst[0] = byte(op)

	offset := 1
	for i, el := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 2:
			PutUint16(inst[offset:], uint16(el))
		}
		offset += width
	}

	return inst
}

func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}

	return def, nil
}
