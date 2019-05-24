package ast

import (
	"bytes"

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
// statement: let foobar = "hahaha";
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
// statement: return bar("a", "b");
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
// expression
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
// expression: foobar
type Identifier struct {
	// the token.IDENT token
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) TokenLiteral() string {
	return i.Token.Literal
}

func (i *Identifier) String() string {
	return i.Value
}

////////////////////////////////////////////////////////////////////////////////
// expression: 5
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
// prefix expression
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

//
////////////////////////////////////////////////////////////////////////////////
// infix expression
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
// literal: boolean
type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }
func (b *Boolean) String() string       { return b.Token.Literal }