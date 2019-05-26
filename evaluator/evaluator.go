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
		// eval the expression value, the result would be Integer, Function, Boolean
		// the result is concrete struct pointer of object.Object interface
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}

		// save the identifier to env
		env.Set(node.Name.Name, val)

	case *ast.ExpressionStatement:
		return Eval(node.ExpressionValue, env)

	case *ast.PrefixExpression:
		// there could be many prefix op, however, here is only for only for ! -,
		// the ast.node is returned by parsePrefixExpression in parser.go
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
		// eval will eventually goto the evalIdentifier,
		// which will return object.Function(the concrete struct pointer of Object interface ) from store
		// or builtin func in the builtins map
		fun := Eval(node.CallableName, env)
		if isError(fun) {
			return fun
		}

		// eval for each actual args
		actualParams := evalExpressions(node.ActualParams, env)
		if len(actualParams) == 1 && isError(actualParams[0]) {
			return actualParams[0]
		}

		// fun will have 2 types: object.Function or object.Builtin
		return applyFunction(fun, actualParams)

	case *ast.Identifier:
		// lookup from env
		return evalIdentifier(node, env)

	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}

		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}

		return evalIndexExpresson(left, index)

	// below can only appear on the right side of assignment =
	case *ast.IfExpression:
		return evalIfExpression(node, env)

	case *ast.IntegerLiteral:
		// returned is struct pointer, which implements the object.Object interface
		return &object.Integer{Value: node.Value}

	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)

	case *ast.StringLiteral:
		return &object.String{Value: node.Value}

	case *ast.FunctionLiteral:
		// FunctionLiteral is the same as IntegerLiteral and Boolean, can only on the right side of assignment =
		// will be saved to env as the object when eval letStatement
		params := node.FormalParams
		body := node.Body

		// when define fn, there is no name for the fn, so no need to save to env.
		// However, need bind the env to fn, which will used during the call (closure)
		// no eval here, only return executable object.
		return &object.Function{FormalParams: params, EnvWhenDefined: env, Body: body}

	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}

		return &object.Array{Elements: elements}

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
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(op, left, right)
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
	if val, ok := env.Get(node.Name); ok {
		return val
	}

	if builtin, ok := builtins[node.Name]; ok {
		return builtin
	}

	return newError("identifier not found: %s", node.Name)
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
	switch fun := fn.(type) {
	case *object.Function:
		// let foo = fn(a,b) { a + b}
		// when eval this letStatement, will build the env: foo is the key, the value is &object.Function(see eval case for FunctionLiteral)
		// when eval foo(2, 3), foo is the identifier, which eval in env (see eval case for CallExpression), and
		// the result is the Function saved by letStatement
		extendedEnv := createCallEnv(fun, args)
		evaluated := Eval(fun.Body, extendedEnv)
		return unwrapReturnValue(evaluated)

	case *object.Builtin:
		// returned from evalIdentifier
		// Fn is func in golang, and will not be evaled, in the definition of Fn, there is no closure.
		// so there is no env here, all infos should passed through the args, which will be evaled in the env
		return fun.Fn(args...)

	default:
		return newError("not a function: %s", fn.Type())
	}

}

func createCallEnv(fn *object.Function, args []object.Object) *object.Environment {
	// create call env(new env) based on fn define env (old env)
	env := object.NewEnclosedEnv(fn.EnvWhenDefined)

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

func evalStringInfixExpression(op string, left, right object.Object) object.Object {
	if op != "+" {
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}

	l := left.(*object.String).Value
	r := right.(*object.String).Value
	return &object.String{Value: l + r}
}

////////////////////////////////////////////////////////////////////////////////
func evalIndexExpresson(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)

	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	if idx < 0 || idx > max {
		return NULL
	}

	return arrayObject.Elements[idx]
}
