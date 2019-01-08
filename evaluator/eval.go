package evaluator

import (
	"github.com/ajwerner/monkey/ast"
	"github.com/ajwerner/monkey/object"
)

func Eval(node ast.Node) object.Object {

	switch node := node.(type) {

	// Statements
	case *ast.Program:
		return evalStatements(node.Statements)

	case *ast.ExpressionStatement:
		return Eval(node.Expression)

		// Expressions
	case *ast.IntegerLiteral:
		return object.Integer(node.Value)
	case *ast.Boolean:
		return object.Boolean(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right)
		return evalPrefixExpression(node.Operator, right)
	}

	return nil
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	default:
		return object.Null{}
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch v := right.(type) {
	case object.Boolean:
		return !v
	case object.Null:
		return object.Boolean(true)
	default:
		return object.Boolean(false)
	}
}

func evalStatements(stmts []ast.Statement) object.Object {
	var result object.Object

	for _, statement := range stmts {
		result = Eval(statement)
	}

	return result
}
