package evaluator

import (
	"fmt"
	"xmonkey/ast"
	"xmonkey/object"
)

var (
	NULL  = &object.NULL{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		// return evalStatements(node.Statements)
		return evalProgram(node)

	case *ast.BlockStatement:
		// return evalStatements(node.Statements)
		return evalBlockStatement(node)

	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}

	case *ast.ExpressionStatement:
		return Eval(node.ExpressionValue)

	case *ast.IfExpression:
		return evalIfExpression(node)

	case *ast.PrefixExpression:
		right := Eval(node.Right)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)

	case *ast.InfixExpression:
		left := Eval(node.Left)
		if isError(left) {
			return left
		}

		right := Eval(node.Right)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)

	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}

	case *ast.Boolean:
		// return &object.Boolean{Value: node.Value}
		return nativeBoolToBooleanObject(node.Value)
	}

	return nil
}

func evalProgram(program *ast.Program) object.Object {
	var result object.Object

	for _, stmt := range program.Statements {
		result = Eval(stmt)

		// if returnValue, ok := result.(*object.ReturnValue); ok {
		// 	return returnValue.Value
		// }

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func evalBlockStatement(block *ast.BlockStatement) object.Object {
	var result object.Object

	for _, stmt := range block.Statements {
		result = Eval(stmt)

		// if result != nil && result.Type() == object.RETURN_VALUE_OBJ {
		// 	return result
		// }

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

func evalStatements(stmts []ast.Statement) object.Object {
	var result object.Object

	// for multi-statements, return the last statement's eval
	for _, stmt := range stmts {
		result = Eval(stmt)

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
		// return NULL
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
		// return NULL
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
		// return NULL
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

func evalIfExpression(expr *ast.IfExpression) object.Object {
	cond := Eval(expr.Condition)
	if isError(cond) {
		return cond
	}

	if isTruthy(cond) {
		return Eval(expr.Consequence)
	} else if expr.Alternative != nil {
		return Eval(expr.Alternative)
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
