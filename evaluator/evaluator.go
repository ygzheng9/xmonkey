package evaluator

import (
	"fmt"
	"xmonkey/ast"
	"xmonkey/object"
)

var (
	// only need one instance
	NULL  = &object.NULL{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

// Eval always needs env
// return signature is Object, which is interface, however the actual returned value is always the pointer of struct
func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node, env)

	case *ast.BlockStatement:
		return evalBlockStatement(node, env)

	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}

	case *ast.LetStatement:
		// eval the expression value
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}

		// save the identifier to env
		env.Set(node.Name.Name, val)

	case *ast.ExpressionStatement:
		return Eval(node.ExpressionValue, env)

	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)

	case *ast.InfixExpression:
		// ast.InfixExpression is from registerPrefix, here is +-*/ == !=
		// not include function call, which is also infix op but with different parsefn
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}

		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)

	case *ast.CallExpression:
		// function call, which is also infix op
		fun := Eval(node.Function, env)
		if isError(fun) {
			return fun
		}

		// eval for each actual args
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(fun, args)

	case *ast.Identifier:
		// lookup from env
		return evalIdentifier(node, env)

	// below can only appear on the right side of assignment =
	case *ast.IfExpression:
		return evalIfExpression(node, env)

	case *ast.IntegerLiteral:
		// returned is struct pointer, which implements the object.Object interface
		return &object.Integer{Value: node.Value}

	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)

	case *ast.FunctionLiteral:
		// FunctionLiteral is the same as IntegerLiteral and Boolean, can only on the right side of assignment =
		// will be saved to env as the object, and the corresponding name is from let statement
		params := node.Parameters
		body := node.Body

		// when define fn, there is no name for the fn, so no need to save to env.
		// However, need bind the env to fn, which will used during the call (closure)
		// no eval here, only return executable object.
		return &object.Function{FormalParams: params, Env: env, Body: body}
	}

	return nil
}

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	for _, stmt := range program.Statements {
		result = Eval(stmt, env)

		// will return for the first return or error
		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object

	for _, stmt := range block.Statements {
		result = Eval(stmt, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

func evalStatements(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	// for multi-statements, return the last statement's eval
	for _, stmt := range stmts {
		result = Eval(stmt, env)

		// will return for the first return object
		if returnValue, ok := result.(*object.ReturnValue); ok {
			return returnValue.Value
		}
	}

	return result
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}

	return FALSE
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	// only defined two actual prefix op: ! -, see parser.go
	switch operator {
	case "!":
		return evalBangOperator(right)
	case "-":
		return evalMinusPrefixOperator(right)
	default:
		// return NULL
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalBangOperator(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

// TODO: can not handle -12abd
func evalMinusPrefixOperator(right object.Object) object.Object {
	if right == nil {
		return NULL
	}

	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())

	}

	i, ok := right.(*object.Integer)
	if !ok {
		return NULL
	}

	return &object.Integer{Value: -i.Value}

}

func evalInfixExpression(op string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfix(op, left, right)
	case op == "==":
		return nativeBoolToBooleanObject(left == right)
	case op == "!=":
		return nativeBoolToBooleanObject(left != right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), op, right.Type())
	default:
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

func evalIntegerInfix(op string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch op {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

func evalIfExpression(expr *ast.IfExpression, env *object.Environment) object.Object {
	cond := Eval(expr.Condition, env)
	if isError(cond) {
		return cond
	}

	if isTruthy(cond) {
		return Eval(expr.Consequence, env)
	} else if expr.Alternative != nil {
		return Eval(expr.Alternative, env)
	}
	return NULL
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

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}

	return false
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	// get value from env
	val, ok := env.Get(node.Name)
	if !ok {
		return newError("identifier not found: %s", node.Name)
	}

	return val
}

// eval for each expr in the exps
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
	fun, ok := fn.(*object.Function)
	if !ok {
		return newError("not a function: %s", fn.Type())
	}

	extendedEnv := extendFunctionEnv(fun, args)
	evaluated := Eval(fun.Body, extendedEnv)
	return unwrapReturnValue(evaluated)
}

func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	// create call env(new env) based on fn define env (old env)
	env := object.NewEnclosedEnv(fn.Env)

	// setup the new env(call env), name is from fn definition's params' name, value is evaled args' values
	for paramIdx, param := range fn.FormalParams {
		env.Set(param.Name, args[paramIdx])
	}

	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}

	return obj
}
