package compiler

import (
	"fmt"

	"github.com/ajwerner/monkey/ast"
	"github.com/ajwerner/monkey/code"
	"github.com/ajwerner/monkey/object"
)

type Compiler struct {
	instructions code.Instructions
	constants    []object.Object
}

func New() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    []object.Object{},
	}
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(code.OpPop)

	case *ast.InfixExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Right)
		if err != nil {
			return err
		}
		switch node.Operator {
		case "+":
			c.emit(code.OpAdd)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.IntegerLiteral:
		integer := object.Integer(node.Value)
		c.emit(code.OpConstant, c.addConstant(integer))
		// TODO: What now?!
	}

	return nil
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)
	return pos
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}

func (c *Compiler) addInstruction(ins []byte) (pos int) {
	pos = len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return pos
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}
