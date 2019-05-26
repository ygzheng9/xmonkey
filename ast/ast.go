package ast

import (
	"bytes"
	"strings"

	"xmonkey/token"
)

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

////////////////////////////////////////////////////////////////////////////////
// 总体结构
// input 会被解析成 1 个 program；
// 1 个 program 中有多个 statements；
// 每个 statement 只有三种类型： let, return, expression
// 注意，expression 也是一种 statement；
// 每个 expression 有 5 种类型：literal, integer, boolean, prefix, infix

////////////////////////////////////////////////////////////////////////////////
// the whole program
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}

	return ""
}

func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

////////////////////////////////////////////////////////////////////////////////
// statement: block statement
// ATTENTION: why statement not expression
// 和 program 是一样的，内部有多个 statements
type BlockStatement struct {
	Token      token.Token
	Statements []Statement
}

func (r *BlockStatement) statementNode()       {}
func (r *BlockStatement) TokenLiteral() string { return r.Token.Literal }
func (r *BlockStatement) String() string {
	var out bytes.Buffer

	for _, s := range r.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

////////////////////////////////////////////////////////////////////////////////
// statement1: let foobar = "hahaha";
type LetStatement struct {
	// the token.LET token
	Token token.Token
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode() {}
func (ls *LetStatement) TokenLiteral() string {
	return ls.Token.Literal
}

func (ls *LetStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")

	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}

	out.WriteString(";")

	return out.String()
}

////////////////////////////////////////////////////////////////////////////////
// statement2: return bar("a", "b");
type ReturnStatement struct {
	// the token.RETURN token
	Token       token.Token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode() {}
func (rs *ReturnStatement) TokenLiteral() string {
	return rs.Token.Literal
}

func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")

	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}

	out.WriteString(";")

	return out.String()
}

////////////////////////////////////////////////////////////////////////////////
// statement3: expression
type ExpressionStatement struct {
	// the first token of the expression
	Token           token.Token
	ExpressionValue Expression
}

func (r *ExpressionStatement) statementNode() {}
func (r *ExpressionStatement) TokenLiteral() string {
	return r.Token.Literal
}

func (r *ExpressionStatement) String() string {
	if r.ExpressionValue != nil {
		return r.ExpressionValue.String()
	}

	return ""
}

////////////////////////////////////////////////////////////////////////////////
// expression1: literal foobar
type Identifier struct {
	// the token.IDENT token
	Token token.Token
	Name  string
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) TokenLiteral() string {
	return i.Token.Literal
}

func (i *Identifier) String() string {
	return i.Name
}

////////////////////////////////////////////////////////////////////////////////
// expression2: integer 5
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (r *IntegerLiteral) expressionNode() {}
func (r *IntegerLiteral) TokenLiteral() string {
	return r.Token.Literal
}

func (r *IntegerLiteral) String() string {
	return r.Token.Literal
}

////////////////////////////////////////////////////////////////////////////////
// expression3: boolean
type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }
func (b *Boolean) String() string       { return b.Token.Literal }

////////////////////////////////////////////////////////////////////////////////
// expression4: prefix
type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (r *PrefixExpression) expressionNode() {}
func (r *PrefixExpression) TokenLiteral() string {
	return r.Token.Literal
}

func (r *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(r.Operator)
	out.WriteString(r.Right.String())
	out.WriteString(")")

	return out.String()
}

////////////////////////////////////////////////////////////////////////////////
// expression5: infix
type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (r *InfixExpression) expressionNode() {}
func (r *InfixExpression) TokenLiteral() string {
	return r.Token.Literal
}

func (r *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(r.Left.String())
	out.WriteString(r.Operator)
	out.WriteString(r.Right.String())
	out.WriteString(")")

	return out.String()
}

////////////////////////////////////////////////////////////////////////////////
// expression: if-then-else
type IfExpression struct {
	Token       token.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (r *IfExpression) expressionNode() {}
func (r *IfExpression) TokenLiteral() string {
	return r.Token.Literal
}

func (r *IfExpression) String() string {
	var out bytes.Buffer

	out.WriteString("if")
	out.WriteString(r.Condition.String())
	out.WriteString(" ")
	out.WriteString(r.Consequence.String())

	if r.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(r.Alternative.String())
	}

	return out.String()
}

////////////////////////////////////////////////////////////////////////////////
// expression: fn(a, b) { c = a + b; c; }
// Notice: no name for the function
type FunctionLiteral struct {
	// token.FUNCTION is always the same (fn)
	Token      token.Token
	Parameters []*Identifier
	Body       *BlockStatement
}

func (r *FunctionLiteral) expressionNode()      {}
func (r *FunctionLiteral) TokenLiteral() string { return r.Token.Literal }
func (r *FunctionLiteral) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range r.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(r.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(r.Body.String())

	return out.String()
}

// expression: fn call
type CallExpression struct {
	Token     token.Token
	Function  Expression
	Arguments []Expression
}

func (r *CallExpression) expressionNode()      {}
func (r *CallExpression) TokenLiteral() string { return r.Token.Literal }
func (r *CallExpression) String() string {
	var out bytes.Buffer

	args := []string{}
	for _, p := range r.Arguments {
		args = append(args, p.String())
	}

	out.WriteString(r.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ","))
	out.WriteString(")")

	return out.String()
}
