package ast

import (
	"bytes"
	"strings"

	"xmonkey/token"
)

// Node is ast node, has many statements, through loop to visit all of them.
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement the element for a program
type Statement interface {
	Node
	statementNode()
}

// Expression self recursion
type Expression interface {
	Node
	expressionNode()
}

////////////////////////////////////////////////////////////////////////////////
// 总体结构
// input 会被解析成 1 个 program；
// 1 个 program 中有多个 statements；
// 每个 statement 有以下类型： let, return, expression, block statements {}
// 注意，expression 也是一种 statement，也即：一个  ExpressionStatement 中只包含一个 expression，这是 root，这个 expression 是自递归结构，
// 主要体现在 前缀操作 和 中缀操作，left/right 都是 expression

// 每个 expression 有多类型，比如：
// 简单类型：
// identifier, interger, boolean, string

// 组合类型：Left op Right，借此 expression 形成自递归
// prefix(! -), infix(+-*/, etc)；注意：有些函数名字中带有 prefix/infix，但是他们不代表全部，他们只是其中的一员而已；
// infix 还有很多其它类型，比如：函数调用 f(x, y), 数组下标操作 arr[3]

// 特殊类型
// 函数定义 fn(a, b) { let c = a + b; return c * c; }, group expression 参数列表 (1,2,3), 数组 [1,2,3],

////////////////////////////////////////////////////////////////////////////////
// the whole program

// Program stands for the whole input.
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

// BlockStatement for {}, only appears:
// 1. if-else: if condition { consequence } else { alternative }
// 2. fun definition: fn(a, b) { let c = a + b; return c * c }
// 和 program 是一样的，内部有多个 statements
type BlockStatement struct {
	Token      token.Token
	Statements []Statement
}

func (r *BlockStatement) statementNode()       {}
func (r *BlockStatement) TokenLiteral() string { return r.Token.RawString }
func (r *BlockStatement) String() string {
	var out bytes.Buffer

	for _, s := range r.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

// LetStatement generates AST for let statement
// statement1: let foobar = "hahaha";
// Name 是 变量 的 name，Value 是 变量 对应的 ast，在 parser 中没有求值，只是生成了 ast，在 eval 中才会求值，并存放到 env 中；
// Name  为 key， 对应的值 eval(Value, env)
// let a = 4; let b = a + 4;
// 在 parser 中，4 是 intergerLiteral，a，b 是 identifier，a + 4 是 infix expression
// 在 eval 中:
// 4 类型为 object.Integer 4；
// a 类型为 object.Identifier，是 key，在 env 中对应的 Object 是 4，类型是 Integer；
// b 类型为 object.Identifier, 是 key，在 env 中对应的 Object 是 eval(a + 4) 的值，
// a+4 这个 infix 被 eval 之后的值是 8，也即：b 在 env 中对应的 Object 是 8，类型为 Integer
type LetStatement struct {
	// the token.LET token
	Token token.Token
	Name  *Identifier
	Expr  Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.RawString }
func (ls *LetStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")

	if ls.Expr != nil {
		out.WriteString(ls.Expr.String())
	}

	out.WriteString(";")

	return out.String()
}

// ReturnStatement for return foo("a", "b");
type ReturnStatement struct {
	// the token.RETURN token
	Token token.Token
	Expr  Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.RawString }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")

	if rs.Expr != nil {
		out.WriteString(rs.Expr.String())
	}

	out.WriteString(";")

	return out.String()
}

// ExpressionStatement for one statement, which is the expression
// expression is self-recursion, Expr is the root, which may have Left, Op, Right, ect.
type ExpressionStatement struct {
	// the first token of the expression
	Token token.Token
	Expr  Expression
}

func (r *ExpressionStatement) statementNode()       {}
func (r *ExpressionStatement) TokenLiteral() string { return r.Token.RawString }
func (r *ExpressionStatement) String() string {
	if r.Expr != nil {
		return r.Expr.String()
	}

	return ""
}

////////////////////////////////////////////////////////////////////////////////

// Identifier 是 变量名，变量名是不可变的，这里只保存 变量名，不保存 变量值；
// 变量值 是在 eval(letStatement) 时计算出来的，保存到 env 中, key 是 这里的 变量名
// expression: foobar, 注意：没有引号，如果有引号，就表示 StringLiteral
type Identifier struct {
	// the token.IDENT token
	Token token.Token
	Name  string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.RawString }
func (i *Identifier) String() string {
	return i.Name
}

////////////////////////////////////////////////////////////////////////////////
// Literal means can not be changed once created.
// Created means eval will create object.Object for this ast node.

// IntegerLiteral for 125
// after lexer, we got string "125", during parsing(parseIntegerLiteral), we cast it to integer 125.
// the casting can be done during parsing, or just leave it to eval.
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (r *IntegerLiteral) expressionNode()      {}
func (r *IntegerLiteral) TokenLiteral() string { return r.Token.RawString }
func (r *IntegerLiteral) String() string {
	return r.Token.RawString
}

// Boolean for true/false
type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.RawString }
func (b *Boolean) String() string       { return b.Token.RawString }

// StringLiteral for "foobar", Attention: the double quotes, if no double quotes, it's Identifier
type StringLiteral struct {
	Token token.Token
	Value string
}

func (r *StringLiteral) expressionNode()      {}
func (r *StringLiteral) TokenLiteral() string { return r.Token.RawString }
func (r *StringLiteral) String() string       { return r.Token.RawString }

// ArrayLiteral for [1, 3, 3+4]
type ArrayLiteral struct {
	Token    token.Token
	Elements []Expression
}

func (r *ArrayLiteral) expressionNode()      {}
func (r *ArrayLiteral) TokenLiteral() string { return r.Token.RawString }
func (r *ArrayLiteral) String() string {
	var out bytes.Buffer

	elements := []string{}
	for _, el := range r.Elements {
		elements = append(elements, el.String())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ","))
	out.WriteString("]")

	return out.String()
}

// FunctionLiteral parses function definition, fn(a, b) { c = a + b; c; }
// Notice: no name for the function
type FunctionLiteral struct {
	// token.FUNCTION is always the same (fn)
	Token        token.Token
	FormalParams []*Identifier
	Body         *BlockStatement
}

func (r *FunctionLiteral) expressionNode()      {}
func (r *FunctionLiteral) TokenLiteral() string { return r.Token.RawString }
func (r *FunctionLiteral) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range r.FormalParams {
		params = append(params, p.String())
	}

	out.WriteString(r.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(r.Body.String())

	return out.String()
}

type HashLiteral struct {
	Token token.Token
	Pairs map[Expression]Expression
}

func (r *HashLiteral) expressionNode()      {}
func (r *HashLiteral) TokenLiteral() string { return r.Token.RawString }
func (r *HashLiteral) String() string {
	var out bytes.Buffer

	pairs := []string{}
	for key, value := range r.Pairs {
		pairs = append(pairs, key.String()+":"+value.String())
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}

////////////////////////////////////////////////////////////////////////////////
// expression4: prefix

// only for ! -
type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (r *PrefixExpression) expressionNode()      {}
func (r *PrefixExpression) TokenLiteral() string { return r.Token.RawString }
func (r *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(r.Operator)
	out.WriteString(r.Right.String())
	out.WriteString(")")

	return out.String()
}

////////////////////////////////////////////////////////////////////////////////
// expression: infix

// InfixExpression eg:2 + 3 * 4 - 5
// There would be many infix operators, here is only for simple ones, such as: +-*/ == !=
// It happens to have Infix in the name.
type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (r *InfixExpression) expressionNode()      {}
func (r *InfixExpression) TokenLiteral() string { return r.Token.RawString }
func (r *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(r.Left.String())
	out.WriteString(r.Operator)
	out.WriteString(r.Right.String())
	out.WriteString(")")

	return out.String()
}

// CallExpression is also the infix operator
// expression: fn call
type CallExpression struct {
	// Token is fixed to (
	Token        token.Token
	CallableName Expression
	ActualParams []Expression
}

func (r *CallExpression) expressionNode()      {}
func (r *CallExpression) TokenLiteral() string { return r.Token.RawString }
func (r *CallExpression) String() string {
	var out bytes.Buffer

	args := []string{}
	for _, p := range r.ActualParams {
		args = append(args, p.String())
	}

	out.WriteString(r.CallableName.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ","))
	out.WriteString(")")

	return out.String()
}

// IndexExpression is another infix operator
// arr[2]
type IndexExpression struct {
	Token token.Token
	Left  Expression
	Index Expression
}

func (r *IndexExpression) expressionNode()      {}
func (r *IndexExpression) TokenLiteral() string { return r.Token.RawString }
func (r *IndexExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(r.Left.String())
	out.WriteString("[")
	out.WriteString(r.Index.String())
	out.WriteString("])")

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

func (r *IfExpression) expressionNode()      {}
func (r *IfExpression) TokenLiteral() string { return r.Token.RawString }
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
