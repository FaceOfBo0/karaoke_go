package compiler

import (
	"fmt"
	"karaoke/ast"
	"karaoke/code"
	"karaoke/lexer"
	"karaoke/object"
	"karaoke/parser"
	"testing"
)

type CompilerTestCase struct {
	input         string
	expectedConst []interface{}
	expectedInsts []code.Instructions
}

func TestFuncsWithArguments(t *testing.T) {
	tests := []CompilerTestCase{
		{
			input: `
			let oneArg = fn(a) { a };
			oneArg(24);
			`,
			expectedConst: []interface{}{
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpReturnValue),
				},
				24,
			},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpCall, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
			let manyArg = fn(a, b, c) { a; b; c; };
			manyArg(24, 25, 26);
			`,
			expectedConst: []interface{}{
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpPop),
					code.Make(code.OpGetLocal, 1),
					code.Make(code.OpPop),
					code.Make(code.OpGetLocal, 2),
					code.Make(code.OpReturnValue),
				},
				24,
				25,
				26,
			},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpCall, 3),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
			let add = fn(a,b) { a + b };
			add(2,5);
			`,
			expectedConst: []interface{}{
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpGetLocal, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
				2,
				5,
			},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpCall, 2),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestLocalBindings(t *testing.T) {
	tests := []CompilerTestCase{
		{
			input: `
			let num = 55;
			fn() { num }`,
			expectedConst: []interface{}{
				55,
				[]code.Instructions{
					code.Make(code.OpGetGlobal, 0),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
			fn() {
				let num = 55;
				num }`,
			expectedConst: []interface{}{
				55,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
			fn() {
			let a = 55;
			let b = 77;
			a + b }`,
			expectedConst: []interface{}{
				55,
				77,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpSetLocal, 1),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpGetLocal, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 2),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestFunctionCalls(t *testing.T) {
	tests := []CompilerTestCase{
		{
			input: `fn() { 24 }();`,
			expectedConst: []interface{}{
				24,
				[]code.Instructions{
					code.Make(code.OpConstant, 0), // The literal "24"
					code.Make(code.OpReturnValue),
				},
			},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 1), // The compiled function
				code.Make(code.OpCall),
				code.Make(code.OpPop),
			},
		},
		{
			input: "fn () { 5 + 10 }();",
			expectedConst: []interface{}{
				5,
				10,
				&object.CompiledFunction{
					Instructions: concatInstructions([]code.Instructions{
						code.Make(code.OpConstant, 0),
						code.Make(code.OpConstant, 1),
						code.Make(code.OpAdd),
						code.Make(code.OpReturnValue),
					}),
				},
			},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 2),
				code.Make(code.OpCall),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
			let noArg = fn() { 24 };
			noArg();
			`,
			expectedConst: []interface{}{
				24,
				[]code.Instructions{
					code.Make(code.OpConstant, 0), // The literal "24"
					code.Make(code.OpReturnValue),
				},
			},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 1), // The compiled function
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpCall),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestFunctionsWithoutReturnValue(t *testing.T) {
	tests := []CompilerTestCase{
		{
			input: `fn() { }`,
			expectedConst: []interface{}{
				[]code.Instructions{
					code.Make(code.OpReturn),
				},
			},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestFunctions(t *testing.T) {
	tests := []CompilerTestCase{
		{
			input: `fn() { 1; 2 }`,
			expectedConst: []interface{}{
				1,
				2,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpPop),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 2),
				code.Make(code.OpPop),
			},
		},
		{
			input: "fn () { 5 + 10 }",
			expectedConst: []interface{}{
				5,
				10,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 2),
				code.Make(code.OpPop),
			},
		},
		{
			input: "fn () { return 5 + 10 }",
			expectedConst: []interface{}{
				5,
				10,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 2),
				code.Make(code.OpPop),
			},
		},
		{
			input: "fn () { return 15 - 5 }",
			expectedConst: []interface{}{
				15,
				5,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpSub),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 2),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestIndexExpressions(t *testing.T) {
	tests := []CompilerTestCase{
		{
			input:         "[1, 2, 3][1 + 1]",
			expectedConst: []interface{}{1, 2, 3, 1, 1},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpAdd),
				code.Make(code.OpIndex),
				code.Make(code.OpPop),
			},
		},
		{
			input:         "{1: 2}[2 - 1]",
			expectedConst: []interface{}{1, 2, 2, 1},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpHash, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpSub),
				code.Make(code.OpIndex),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestHashLiterals(t *testing.T) {
	tests := []CompilerTestCase{
		{
			input:         `{1 + 1: 2 * 2, 3 + 3: 4 * 4}`,
			expectedConst: []interface{}{1, 1, 2, 2, 3, 3, 4, 4},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpMul),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpAdd),
				code.Make(code.OpConstant, 6),
				code.Make(code.OpConstant, 7),
				code.Make(code.OpMul),
				code.Make(code.OpHash, 2),
				code.Make(code.OpPop),
			},
		},
		{
			input:         "{1: 2 + 3, 4: 5 * 6}",
			expectedConst: []interface{}{1, 2, 3, 4, 5, 6},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpAdd),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpMul),
				code.Make(code.OpHash, 2),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestArrayLiterls(t *testing.T) {
	tests := []CompilerTestCase{
		{
			input:         `[]`,
			expectedConst: []interface{}{},
			expectedInsts: []code.Instructions{
				code.Make(code.OpArray, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:         `[1, 2, 3, 4]`,
			expectedConst: []interface{}{1, 2, 3, 4},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpArray, 4),
				code.Make(code.OpPop),
			},
		},
		{
			input:         `[1 + 2, 3 - 4, 5 * 6]`,
			expectedConst: []interface{}{1, 2, 3, 4, 5, 6},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpSub),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpMul),
				code.Make(code.OpArray, 3),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestStringExpressions(t *testing.T) {
	tests := []CompilerTestCase{
		{
			input:         `"monkey"`,
			expectedConst: []interface{}{"monkey"},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:         `"mon" + "key"`,
			expectedConst: []interface{}{"mon", "key"},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestLetStatements(t *testing.T) {
	tests := []CompilerTestCase{
		{
			input: `
			let one = 1;
			let two = 2;
			`,
			expectedConst: []interface{}{1, 2},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSetGlobal, 1),
			},
		},

		{
			input: `
			let one = 1;
			one;
			`,
			expectedConst: []interface{}{1},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
			let one = 1;
			let two = one;
			two;
			`,
			expectedConst: []interface{}{1},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input:         "let x = 66; let y = 33; let z = x + y",
			expectedConst: []interface{}{66, 33},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpSetGlobal, 2),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []CompilerTestCase{
		{
			input:         "if (true) { 10 }; 3333;",
			expectedConst: []interface{}{10, 3333},
			expectedInsts: []code.Instructions{
				// 0000
				code.Make(code.OpTrue),
				// 0001
				code.Make(code.OpJumpNotTruthy, 10),
				// 0004
				code.Make(code.OpConstant, 0),
				// 0007
				code.Make(code.OpJump, 11),
				// 0010
				code.Make(code.OpNull),
				// 0011
				code.Make(code.OpPop),
				// 0012
				code.Make(code.OpConstant, 1),
				// 0015
				code.Make(code.OpPop),
			},
		},
		{
			input:         "if (true) { 10 } else { 20 }; 3333;",
			expectedConst: []interface{}{10, 20, 3333},
			expectedInsts: []code.Instructions{
				// 0000
				code.Make(code.OpTrue),
				// 0001
				code.Make(code.OpJumpNotTruthy, 10),
				// 0004
				code.Make(code.OpConstant, 0),
				// 0007
				code.Make(code.OpJump, 13),
				// 0010
				code.Make(code.OpConstant, 1),
				// 0013
				code.Make(code.OpPop),
				// 0014
				code.Make(code.OpConstant, 2),
				// 0017
				code.Make(code.OpPop),
			},
		},
		{
			input:         "if (true) { 10 };",
			expectedConst: []interface{}{10},
			expectedInsts: []code.Instructions{
				// 0000
				code.Make(code.OpTrue),
				// 0001
				code.Make(code.OpJumpNotTruthy, 10),
				// 0004
				code.Make(code.OpConstant, 0),
				// 0007
				code.Make(code.OpJump, 11),
				// 0010
				code.Make(code.OpNull),
				// 0011
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestIntegerArithemtic(t *testing.T) {
	tests := []CompilerTestCase{
		{
			input:         "4 + 7",
			expectedConst: []interface{}{4, 7},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpPop),
			},
		},
		{
			input:         "3 - 9",
			expectedConst: []interface{}{3, 9},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSub),
				code.Make(code.OpPop),
			},
		},
		{
			input:         "12 * 5",
			expectedConst: []interface{}{12, 5},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpMul),
				code.Make(code.OpPop),
			},
		},
		{
			input:         "8 / 2",
			expectedConst: []interface{}{8, 2},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpDiv),
				code.Make(code.OpPop),
			},
		},
		{
			input:         "5; 3",
			expectedConst: []interface{}{5, 3},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input:         "-3",
			expectedConst: []interface{}{3},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpMinus),
				code.Make(code.OpPop),
			},
		},
		{
			input:         "-4 + 7",
			expectedConst: []interface{}{4, 7},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpMinus),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpPop),
			},
		},
		{
			input:         "-(4 + 7)",
			expectedConst: []interface{}{4, 7},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpMinus),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []CompilerTestCase{
		{
			input:         "true",
			expectedConst: []interface{}{},
			expectedInsts: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpPop),
			},
		},
		{
			input:         "false",
			expectedConst: []interface{}{},
			expectedInsts: []code.Instructions{
				code.Make(code.OpFalse),
				code.Make(code.OpPop),
			},
		},
		{
			input:         "false == true",
			expectedConst: []interface{}{},
			expectedInsts: []code.Instructions{
				code.Make(code.OpFalse),
				code.Make(code.OpTrue),
				code.Make(code.OpEqual),
				code.Make(code.OpPop),
			},
		},
		{
			input:         "3 == 1",
			expectedConst: []interface{}{3, 1},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpEqual),
				code.Make(code.OpPop),
			},
		},
		{
			input:         "3 != 1",
			expectedConst: []interface{}{3, 1},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpNotEqual),
				code.Make(code.OpPop),
			},
		},
		{
			input:         "true != true",
			expectedConst: []interface{}{},
			expectedInsts: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpTrue),
				code.Make(code.OpNotEqual),
				code.Make(code.OpPop),
			},
		},
		{
			input:         "3 > 1",
			expectedConst: []interface{}{3, 1},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpGreaterThan),
				code.Make(code.OpPop),
			},
		},
		{
			input:         "3 < 1",
			expectedConst: []interface{}{1, 3},
			expectedInsts: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpGreaterThan),
				code.Make(code.OpPop),
			},
		},
		{
			input:         "!true",
			expectedConst: []interface{}{},
			expectedInsts: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpBang),
				code.Make(code.OpPop),
			},
		},
		{
			input:         "!false",
			expectedConst: []interface{}{},
			expectedInsts: []code.Instructions{
				code.Make(code.OpFalse),
				code.Make(code.OpBang),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestCompilerScopes(t *testing.T) {
	compiler := New()
	if compiler.scopeIdx != 0 {
		t.Errorf("scopeIdx wrong. got=%d, want=%d", compiler.scopeIdx, 0)
	}
	globalSymTable := compiler.symbolTable

	compiler.emit(code.OpMul)

	compiler.enterScope()
	if compiler.scopeIdx != 1 {
		t.Errorf("scopeIdx wrong. got=%d, want=%d", compiler.scopeIdx, 1)
	}

	compiler.emit(code.OpSub)

	if len(compiler.scopes[compiler.scopeIdx].instructions) != 1 {
		t.Errorf("instructions length wrong. got=%d",
			len(compiler.scopes[compiler.scopeIdx].instructions))
	}

	last := compiler.scopes[compiler.scopeIdx].lastInst
	if last.Opcode != code.OpSub {
		t.Errorf("lastInstruction.Opcode wrong. got=%d, want=%d",
			last.Opcode, code.OpSub)
	}

	if compiler.symbolTable.Outer != globalSymTable {
		t.Errorf("compiler did not enclose symbolTable")
	}

	compiler.leaveScope()
	if compiler.scopeIdx != 0 {
		t.Errorf("scopeIdx wrong. got=%d, want=%d", compiler.scopeIdx, 0)
	}

	if compiler.symbolTable != globalSymTable {
		t.Errorf("compiler did not restore global symbol table")
	}
	if compiler.symbolTable.Outer != nil {
		t.Errorf("compiler modified global symbol table incorrectly")
	}

	compiler.emit(code.OpAdd)

	if len(compiler.scopes[compiler.scopeIdx].instructions) != 2 {
		t.Errorf("instructions length wrong. got=%d",
			len(compiler.scopes[compiler.scopeIdx].instructions))
	}

	last = compiler.scopes[compiler.scopeIdx].lastInst
	if last.Opcode != code.OpAdd {
		t.Errorf("lastInstruction.Opcode wrong. got=%d, want=%d",
			last.Opcode, code.OpAdd)
	}

	previous := compiler.scopes[compiler.scopeIdx].prevInst
	if previous.Opcode != code.OpMul {
		t.Errorf("previousInstruction.Opcode wrong. got=%d, want=%d",
			previous.Opcode, code.OpMul)
	}
}

func runCompilerTests(t *testing.T, tests []CompilerTestCase) {
	t.Helper()

	for _, tt := range tests {
		program := parse(tt.input)

		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		bytecode := compiler.Bytecode()

		err = testInstructions(tt.expectedInsts, bytecode.Instructions)
		if err != nil {
			t.Fatalf("testInstructions failed: %s", err)
		}

		err = testConstants(tt.expectedConst, bytecode.Constants)
		if err != nil {
			t.Fatalf("testConstants failed: %s", err)
		}

	}
}

func concatInstructions(insts []code.Instructions) code.Instructions {
	out := code.Instructions{}
	for _, el := range insts {
		out = append(out, el...)
	}
	return out
}

func testInstructions(expected []code.Instructions, actual code.Instructions) error {
	concatted := concatInstructions(expected)

	if len(actual) != len(concatted) {
		return fmt.Errorf("wrong instructions length.\nwant=%q\ngot =%q",
			concatted, actual)
	}

	for i, ins := range concatted {
		if actual[i] != ins {
			return fmt.Errorf("wrong instruction at %d.\nwant=%q\ngot =%q", i, concatted, actual)
		}
	}

	return nil
}

func testIntegerObject(expected int64, actual object.Object) error {
	result, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("object is not an Integer. got =%T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. want=%d, got =%d", expected, result.Value)
	}
	return nil
}

func testStringObject(expected string, actual object.Object) error {
	result, ok := actual.(*object.String)
	if !ok {
		return fmt.Errorf("object is not an String. got =%T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. want=%s, got =%s", expected, result.Value)
	}
	return nil
}

func testConstants(expected []interface{}, actual []object.Object) error {

	if len(expected) != len(actual) {
		return fmt.Errorf("wrong number of constants. want=%d, got =%d", len(expected), len(actual))
	}

	for i, constant := range expected {
		switch constant := constant.(type) {
		case string:
			err := testStringObject(string(constant), actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - testStringObject failed: %s", i, err)
			}

		case int:
			err := testIntegerObject(int64(constant), actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - testIntegerObject failed: %s", i, err)
			}

		case []code.Instructions:
			fn, ok := actual[i].(*object.CompiledFunction)
			if !ok {
				return fmt.Errorf("constant %d - not a function: %T", i, actual[i])
			}

			err := testInstructions(constant, fn.Instructions)

			if err != nil {
				return fmt.Errorf("constant %d - testInstructions failed: %s", i, err)
			}

		}
	}

	return nil
}

func parse(input string) *ast.Program {
	lx := lexer.New(input)
	ps := parser.New(lx)
	return ps.ParseProgram()

}
