package compiler

import (
	"fmt"
	"monkey/ast"
	"monkey/code"
	"monkey/object"
	"monkey/token"
)

type EmittedInstruction struct {
	Opcode code.Opcode
	Pos    int
}

type Compiler struct {
	instructions code.Instructions
	constants    []object.Object
	lastInst     EmittedInstruction
	prevInst     EmittedInstruction
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

func New() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    []object.Object{},
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

	case *ast.IfExpression:
		err := c.Compile(n.Condition)
		if err != nil {
			return err
		}

		jmpIdx := c.emit(code.OpJumpNotTruthy, 9999)

		err = c.Compile(n.Consequence)
		if err != nil {
			return err
		}

		if c.lastInst.Opcode == code.OpPop {
			c.instructions = c.instructions[:c.lastInst.Pos]
			c.lastInst = c.prevInst
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

	case *ast.IntegerLiteral:
		intObj := &object.Integer{Value: n.Value}
		c.emit(code.OpConstant, c.addConstant(intObj))

	case *ast.Boolean:
		if n.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}

	}

	return nil
}
