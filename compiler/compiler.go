package compiler

import (
	"fmt"
	"monkey/ast"
	"monkey/code"
	"monkey/object"
	"monkey/token"
)

type Compiler struct {
	instructions code.Instructions
	constants    []object.Object
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

func New() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    []object.Object{},
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

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	inst := code.Make(op, operands...)
	pos := c.addInstruction(inst)
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

	case *ast.ExpressionStatement:
		err := c.Compile(n.Expression)
		if err != nil {
			return err
		}

	case *ast.InfixExpression:
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
			c.emit(code.OpMult)
		default:
			return fmt.Errorf("unknown operator %s", n.Operator)
		}

	case *ast.IntegerLiteral:
		intObj := &object.Integer{Value: n.Value}
		c.emit(code.OpConstant, c.addConstant(intObj))
	}
	return nil
}
