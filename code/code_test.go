package code

import (
	"testing"
)

func TestReadOperands(t *testing.T) {

	tests := []struct {
		op        Opcode
		operands  []int
		bytesRead int
	}{
		{OpConstant, []int{65535}, 2},
		{OpPop, []int{}, 0},
		{OpAdd, []int{}, 0},
		{OpSetLocal, []int{43}, 1},
	}

	for _, tt := range tests {
		instruction := Make(tt.op, tt.operands...)
		def, err := Lookup(byte(tt.op))
		if err != nil {
			t.Fatalf("definition not found: %q\n", err)
		}

		operandsRead, n := ReadOperands(def, instruction[1:])
		if n != tt.bytesRead {
			t.Fatalf("n wrong. want=%d, got=%d", tt.bytesRead, n)
		}

		for i, want := range tt.operands {
			if operandsRead[i] != want {
				t.Errorf("operand wrong. want=%d, got=%d", want, operandsRead[i])
			}
		}
	}
}

func TestInstructionsString(t *testing.T) {
	instructions := []Instructions{
		Make(OpAdd),
		Make(OpConstant, 2),
		Make(OpMinus),
		Make(OpConstant, 65535),
		Make(OpSub),
		Make(OpTrue),
		Make(OpNotEqual),
		Make(OpGetLocal, 23),
		Make(OpSetLocal, 35),
	}

	expected := `0000 OpAdd
0001 OpConstant 2
0004 OpMinus
0005 OpConstant 65535
0008 OpSub
0009 OpTrue
0010 OpNotEqual
0011 OpGetLocal 23
0013 OpSetLocal 35
`

	concatted := Instructions{}
	for _, elm := range instructions {
		concatted = append(concatted, elm...)
	}
	if concatted.String() != expected {
		t.Errorf("instructions wrongly formatted. \nwant=%q\ngot =%q",
			expected, concatted.String())
	}
}

func TestMake(t *testing.T) {

	tests := []struct {
		op       Opcode
		operands []int
		expected []byte
	}{
		{OpConstant, []int{65534}, []byte{byte(OpConstant), 255, 254}},
		{OpAdd, []int{}, []byte{byte(OpAdd)}},
		{OpEqual, []int{}, []byte{byte(OpEqual)}},
		{OpGetLocal, []int{137}, []byte{byte(OpGetLocal), 137}},
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
