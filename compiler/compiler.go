package compiler

import (
	"fmt"
	"monkey/ast"
	"monkey/code"
	"monkey/object"
	"monkey/token"
	"slices"
	"strings"
)

type EmittedInstruction struct {
	Opcode code.Opcode
	Pos    int
}

type Compiler struct {
	instructions code.Instructions
	constants    []object.Object
	symbolTable  *SymbolTable
	lastInst     EmittedInstruction
	prevInst     EmittedInstruction
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

func NewWithState(s *SymbolTable, consts []object.Object) *Compiler {
	compiler := New()
	compiler.symbolTable = s
	compiler.constants = consts
	return compiler
}

func New() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    []object.Object{},
		symbolTable:  NewSymbolTable(),
		lastInst:     EmittedInstruction{},
		prevInst:     EmittedInstruction{},
	}
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}

func (c *Compiler) addConstant(con object.Object) int {
	c.constants = append(c.constants, con)
	return len(c.constants) - 1
}

func (c *Compiler) addInstruction(in code.Instructions) int {
	posInst := len(c.instructions)
	c.instructions = append(c.instructions, in...)
	return posInst
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	c.prevInst = c.lastInst
	c.lastInst = EmittedInstruction{Opcode: op, Pos: pos}
}

func (c *Compiler) deleteLastOpPop() {
	if c.lastInst.Opcode == code.OpPop {
		c.instructions = c.instructions[:c.lastInst.Pos]
		c.lastInst = c.prevInst
	}
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	inst := code.Make(op, operands...)
	pos := c.addInstruction(inst)

	c.setLastInstruction(op, pos)
	return pos
}

func (c *Compiler) Compile(node ast.Node) error {
	switch n := node.(type) {
	case *ast.Program:
		for _, st := range n.Statements {
			err := c.Compile(st)
			if err != nil {
				return err
			}
		}

	case *ast.LetStatement:
		err := c.Compile(n.Value)
		if err != nil {
			return err
		}
		symbol := c.symbolTable.Define(n.Name.Value)
		c.emit(code.OpSetGlobal, symbol.Idx)

	case *ast.BlockStatement:
		for _, elm := range n.Statements {
			err := c.Compile(elm)
			if err != nil {
				return err
			}
		}

	case *ast.ExpressionStatement:
		err := c.Compile(n.Expression)
		if err != nil {
			return err
		}
		c.emit(code.OpPop)

	case *ast.IndexExpression:
		err := c.Compile(n.Left)
		if err != nil {
			return err
		}

		err = c.Compile(n.Index)
		if err != nil {
			return err
		}

		c.emit(code.OpIndex)

	case *ast.IfExpression:
		err := c.Compile(n.Condition)
		if err != nil {
			return err
		}

		jmpNotTruthyIdx := c.emit(code.OpJumpNotTruthy, 9999)

		err = c.Compile(n.Consequence)
		if err != nil {
			return err
		}

		c.deleteLastOpPop()

		jmpIdx := c.emit(code.OpJump, 9999)

		code.PutUint16(c.instructions[jmpNotTruthyIdx+1:], uint16(len(c.instructions)))

		if n.Alternative != nil {
			err = c.Compile(n.Alternative)
			if err != nil {
				return err
			}

			c.deleteLastOpPop()
		} else {
			c.emit(code.OpNull)
		}

		code.PutUint16(c.instructions[jmpIdx+1:], uint16(len(c.instructions)))

	case *ast.PrefixExpression:
		err := c.Compile(n.Right)
		if err != nil {
			return err
		}

		switch n.Token.Type {
		case token.BANG:
			c.emit(code.OpBang)
		case token.MINUS:
			c.emit(code.OpMinus)
		default:
			return fmt.Errorf("unknown operator %s", n.Operator)
		}

	case *ast.InfixExpression:
		if n.Token.Type == token.LT {
			err := c.Compile(n.Right)
			if err != nil {
				return err
			}
			err = c.Compile(n.Left)
			if err != nil {
				return err
			}
			c.emit(code.OpGreaterThan)
			return nil
		}
		err := c.Compile(n.Left)
		if err != nil {
			return err
		}
		err = c.Compile(n.Right)
		if err != nil {
			return err
		}
		switch n.Token.Type {
		case token.PLUS:
			c.emit(code.OpAdd)
		case token.MINUS:
			c.emit(code.OpSub)
		case token.ASTERISK:
			c.emit(code.OpMul)
		case token.SLASH:
			c.emit(code.OpDiv)
		case token.EQ:
			c.emit(code.OpEqual)
		case token.NOT_EQ:
			c.emit(code.OpNotEqual)
		case token.GT:
			c.emit(code.OpGreaterThan)
		default:
			return fmt.Errorf("unknown operator %s", n.Operator)
		}

	case *ast.Identifier:
		sym, ok := c.symbolTable.Resolve(n.Value)
		if !ok {
			return fmt.Errorf("variable %s is undefined", n.Value)
		}

		c.emit(code.OpGetGlobal, sym.Idx)

	case *ast.IntegerLiteral:
		intObj := &object.Integer{Value: n.Value}
		c.emit(code.OpConstant, c.addConstant(intObj))

	case *ast.StringLiteral:
		strObj := &object.String{Value: n.Value}
		c.emit(code.OpConstant, c.addConstant(strObj))

	case *ast.Boolean:
		if n.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}

	case *ast.ArrayLiteral:
		for _, elem := range n.Elements {
			err := c.Compile(elem)
			if err != nil {
				return err
			}
		}
		c.emit(code.OpArray, len(n.Elements))

	case *ast.HashLiteral:
		keys := make([]ast.Expression, 0, len(n.Pairs))
		for e := range n.Pairs {
			keys = append(keys, e)
		}

		slices.SortFunc(keys, func(a ast.Expression, b ast.Expression) int {
			return strings.Compare(a.String(), b.String())
		})

		for _, key := range keys {
			err := c.Compile(key)
			if err != nil {
				return err
			}

			err = c.Compile(n.Pairs[key])
			if err != nil {
				return err
			}
		}

		c.emit(code.OpHash, len(n.Pairs))
	}

	return nil
}
