package evaluator

import (
	"fmt"

	"github.com/ajwerner/monkey/ast"
	"github.com/ajwerner/monkey/object"
)

const TRUE = object.Bool(true)
const FALSE = object.Bool(false)

var NULL = object.Null{}

func Eval(node ast.Node, env *object.Environment) object.Object {

	switch node := node.(type) {

	// Statements
	case *ast.Program:
		return evalProgram(node, env)

	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)

	case *ast.BlockStatement:
		return evalBlockStatement(node, env)

	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)

	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return object.ReturnValue{Value: val}
		// Expressions
	case *ast.IntegerLiteral:
		return object.Integer(node.Value)
	case *ast.FloatLiteral:
		return object.Float(node.Value)
	case *ast.StringLiteral:
		return object.String(node.Value)
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Env: env, Body: body}
	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return (*object.Array)(&elements)
	case *ast.Bool:
		return object.Bool(node.Value)
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(function, args)
	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index)
	case *ast.HashLiteral:
		return evalHashLiteral(node, env)
	}

	return nil
}

func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object

	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {

	case *object.Function:
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)

	case *object.Builtin:
		return fn.Fn(args...)

	default:
		return newError("not a function: %s", fn.Type())
	}
}

func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}

	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(object.ReturnValue); ok {
		return returnValue.Value
	}

	return obj
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}
	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	return newError("identifier not found: " + node.Value)
}

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range program.Statements {
		result = Eval(statement, env)

		switch result := result.(type) {
		case object.ReturnValue:
			return result.Value
		case object.Error:
			return result
		}
	}

	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = Eval(statement, env)
		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE || rt == object.ERROR {
				return result
			}
		}
	}

	return result
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch v := right.(type) {
	case object.Bool:
		return !v
	case object.Null:
		return TRUE
	default:
		return FALSE
	}
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	lt, rt := left.Type(), right.Type()
	if lt == object.INTEGER && rt == object.FLOAT {
		left, lt = object.Float(left.(object.Integer)), object.FLOAT
	} else if lt == object.FLOAT && rt == object.INTEGER {
		right, rt = object.Float(right.(object.Integer)), object.FLOAT
	}
	switch {
	case lt == object.INTEGER && rt == object.INTEGER:
		return evalIntegerInfixExpression(operator, left.(object.Integer), right.(object.Integer))
	case lt == object.FLOAT && rt == object.FLOAT:
		return evalFloatInfixExpression(operator, left.(object.Float), right.(object.Float))
	case lt == object.STRING && rt == object.STRING:
		return evalStringInfixExpression(operator, left.(object.String), right.(object.String))
	case operator == "==":
		return object.Bool(left == right)
	case operator == "!=":
		return object.Bool(left != right)
	case lt != rt:
		return newError("type mismatch: %s %s %s",
			lt, operator, rt)
	default:
		return newError("unknown operator: %s %s %s",
			lt, operator, rt)

	}
}

func evalStringInfixExpression(operator string, left, right object.String) object.Object {
	if operator != "+" {
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
	return left + right
}

func evalIntegerInfixExpression(operator string, left, right object.Integer) object.Object {
	switch operator {
	case "+":
		return left + right
	case "-":
		return left - right
	case "*":
		return left * right
	case "/":
		return left / right
	case "<":
		return object.Bool(left < right)
	case ">":
		return object.Bool(left > right)
	case "==":
		return object.Bool(left == right)
	case "!=":
		return object.Bool(left != right)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalFloatInfixExpression(operator string, left, right object.Float) object.Object {
	switch operator {
	case "+":
		return left + right
	case "-":
		return left - right
	case "*":
		return left * right
	case "/":
		return left / right
	case "<":
		return object.Bool(left < right)
	case ">":
		return object.Bool(left > right)
	case "==":
		return object.Bool(left == right)
	case "!=":
		return object.Bool(left != right)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}
	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	}
	if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	}
	return NULL
}

func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY && index.Type() == object.INTEGER:
		return evalArrayIndexExpression(left, index)
	case left.Type() == object.HASH:
		return evalHashIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalHashIndexExpression(hash, index object.Object) object.Object {
	hashObject := hash.(object.Hash)
	if !object.Hashable(index) {
		return newError("unusable as hash key: %v", index.Type())
	}

	got, ok := hashObject[index]
	if !ok {
		return NULL
	}
	return got

}

func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(object.Integer)
	max := object.Integer(len(*arrayObject) - 1)

	if idx < 0 || idx > max {
		return NULL
	}

	return (*arrayObject)[idx]
}

func evalHashLiteral(node *ast.HashLiteral, env *object.Environment) (o object.Object) {
	defer func() {
		if r := recover(); r != nil {
			o = newError("unhashable key: %v", r)
		}
	}()
	m := make(object.Hash, len(node.Pairs))
	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}
		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}
		m[key] = value
	}
	return m
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER {
		return newError("unknown operator: -%s", right.Type())
	}
	return -1 * right.(object.Integer)
}

func newError(format string, a ...interface{}) object.Error {
	return object.Error{Err: fmt.Errorf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR
	}
	return false
}
