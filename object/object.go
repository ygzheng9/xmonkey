package object

import (
	"bytes"
	"fmt"
	"strings"

	"xmonkey/ast"
)

type ObjectType string

const (
	INTEGER_OBJ      = "INTEGER"
	BOOLEAN_OBJ      = "BOOLEAN"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	ERROR_OBJ        = "ERROR"
	FUNCTION_OBJ     = "FUNCTION"

	STRING_OBJ = "STRING"

	BUILTIN_OBJ = "BUILTIN"

	ARRAY_OBJ = "ARRAY"
)

// Object 是 eval 的返回值，是一个 interface，具体的返回值都是 struct pointer
type Object interface {
	Type() ObjectType
	Inspect() string
}

// Integer
type Integer struct {
	Value int64
}

func (r *Integer) Type() ObjectType { return INTEGER_OBJ }
func (r *Integer) Inspect() string  { return fmt.Sprintf("%d", r.Value) }

// Boolean
type Boolean struct {
	Value bool
}

func (r *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (r *Boolean) Inspect() string  { return fmt.Sprintf("%t", r.Value) }

// NULL
type NULL struct{}

func (r *NULL) Type() ObjectType { return NULL_OBJ }
func (r *NULL) Inspect() string {
	return "null"
}

// ReturnValue
type ReturnValue struct {
	Value Object
}

func (r *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (r *ReturnValue) Inspect() string {
	return r.Value.Inspect()
}

// Error
type Error struct {
	Message string
}

func (r *Error) Type() ObjectType { return ERROR_OBJ }
func (r *Error) Inspect() string  { return "ERROR: " + r.Message }

// Function is function definition, Body will evaled only when call, not definition
// Parameters are formal params, the name will be used for set up the call env.
// Env will be passed to call env as the outer.
type Function struct {
	FormalParams   []*ast.Identifier
	Body           *ast.BlockStatement
	EnvWhenDefined *Environment
}

func (r *Function) Type() ObjectType { return FUNCTION_OBJ }
func (r *Function) Inspect() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range r.FormalParams {
		params = append(params, p.String())
	}

	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ","))
	out.WriteString(") {\n")
	out.WriteString(r.Body.String())
	out.WriteString("\n}")

	return out.String()
}

type String struct {
	Value string
}

func (r *String) Type() ObjectType { return STRING_OBJ }
func (r *String) Inspect() string  { return r.Value }

type Array struct {
	Elements []Object
}

func (r *Array) Type() ObjectType { return ARRAY_OBJ }
func (r *Array) Inspect() string {
	var out bytes.Buffer

	elements := []string{}
	for _, e := range r.Elements {
		elements = append(elements, e.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ","))
	out.WriteString("]")

	return out.String()
}

////////////////////////////////////////////////////////////////////////////////
// BuiltinFunc 内置函数
type BuiltinFunc func(args ...Object) Object

type Builtin struct {
	Fn BuiltinFunc
}

func (r *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (r *Builtin) Inspect() string  { return "built-in function" }
