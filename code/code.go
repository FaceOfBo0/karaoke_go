package code

import (
	"encoding/binary"
	"fmt"
)

type Instructions []byte

type Opcode byte

type Definition struct {
	Name          string
	OperandWidths []int
}

const (
	OpConstant Opcode = iota
)

var definitions = map[Opcode]*Definition{
	OpConstant: {"OpConstant", []int{2}},
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
			binary.BigEndian.PutUint16(inst[offset:], uint16(el))
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
