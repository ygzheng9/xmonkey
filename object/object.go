package object

import "fmt"

type ObjectType string

const (
	INTEGER_OBJ      = "INTEGER"
	BOOLEAN_OBJ      = "BOOLEAN"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	ERROR_OBJ        = "ERROR"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Integer struct {
	Value int64
}

func (r *Integer) Type() ObjectType { return INTEGER_OBJ }
func (r *Integer) Inspect() string {
	return fmt.Sprintf("%d", r.Value)
}

type Boolean struct {
	Value bool
}

func (r *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (r *Boolean) Inspect() string {
	return fmt.Sprintf("%t", r.Value)
}

type NULL struct {
}

func (r *NULL) Type() ObjectType { return NULL_OBJ }
func (r *NULL) Inspect() string {
	return "null"
}

type ReturnValue struct {
	Value Object
}

func (r *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (r *ReturnValue) Inspect() string {
	return r.Value.Inspect()
}

type Error struct {
	Message string
}

func (r *Error) Type() ObjectType { return ERROR_OBJ }
func (r *Error) Inspect() string {
	return "ERROR: " + r.Message
}
