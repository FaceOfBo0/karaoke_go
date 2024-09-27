package code

import (
	"testing"
)

func TestMake(t *testing.T) {

	tests := []struct {
		op       Opcode
		operands []int
		expected []byte
	}{
		{OpConstant, []int{65534}, []byte{byte(OpConstant), 255, 254}},
	}

	for _, tt := range tests {
		instruction := Make(tt.op, tt.operands...)

		if len(instruction) != len(tt.expected) {
			t.Errorf("instruction has wrong length. want=%d, got=%d",
				len(tt.expected), len(instruction))
		}

		for i, el := range tt.expected {
			if instruction[i] != el {
				t.Errorf("wrong byte at pos %d. want=%d, got=%d",
					i, el, instruction[i])
			}
		}

	}

}
