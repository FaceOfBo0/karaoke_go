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
		def, _ := Lookup(inst[offset])
		/* 	if err != nil {
			fmt.Errorf("definition not found: %q\n", err)
		} */
		operands, bytesRead := ReadOperands(def, inst[offset+1:])
		output += fmt.Sprintf("%04d %s %s\n", offset, def.Name, strings.Trim(fmt.Sprint(operands), "[]"))
		offset += bytesRead + 1
	}

	return output[:len(output)-1]
}

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

func ReadOperands(def *Definition, inst []byte) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0
	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(binary.BigEndian.Uint16(inst[offset:]))
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
