package parser

import (
	"fmt"
	"strconv"
	"strings"

	"xmonkey/ast"
	"xmonkey/lexer"
	"xmonkey/token"
)

const (
	_ int = iota
	LOWEST
	EQUALS
	LESSGREATER
	SUM
	PRODUCT

	// PREFIX means ! -
	PREFIX

	// CALL means function call (
	CALL

	// arr[index]
	INDEX
)

// used in infix
var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,

	// ( is for function call, which is the heighest priority
	token.LPAREN: CALL,

	token.LBRACKET: INDEX,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// Parser will parse input token stream and generate ast
type Parser struct {
	l         *lexer.Lexer
	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn

	errors []string
}

// New creates a new parser
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)

	// !false
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	// -5
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)

	// 1 * ( 2 + 3) + 4
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)

	// if (2 > 1) { 2 + 3 } else { 4 + 5 }
	p.registerPrefix(token.IF, p.parseIfExpression)

	// fn(a, b) { let c = a + b; c }
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)

	// "abc"
	p.registerPrefix(token.STRING, p.parseStringLiteral)

	// arraylist [1, 2 + 2, 3 * 3]
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	// 都是 infix，尽管类型多，但是构造的 ast.node 类型是一样的 InfixExpression，
	// 在 eval 顶层是一个入口， 然后根据 op 不同，再做不同的 case 处理
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)

	// ( 是 fun call 的 infix op，处理逻辑和 +- 等上面的不同
	// 返回 CallExpression, 直接在 eval 顶层直接处理
	p.registerInfix(token.LPAREN, p.parseCallExpression)

	// arr[2]
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)

	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		// 这里向前取了一个 token
		p.nextToken()
		return true
	}

	p.peekError(t)
	return false
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expect next token to be %s. got %s instead", t, p.peekToken.Type)

	p.errors = append(p.errors, msg)
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

////////////////////////////////////////////////////////////////////////////////

// ParseProgram 处理 input 的入口，内部有循环，一直读取到 token.EOF
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	// 这里是循环，一直到遇到了 EOF，也即：处理完 input
	for p.curToken.Type != token.EOF {
		// 解析 statement，在解析过程中，会不断读取 token，并根据当前 token 做不同的逻辑处理
		stmt := p.parseStatement()

		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}

		// 读取下一个 token，即消耗 input
		p.nextToken()
	}

	return program
}

// 在 ParseProgram 的循环中被调用，每次处理一个 statement
func (p *Parser) parseStatement() ast.Statement {
	// statement 只有 3 种类型
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

////////////////////////////////////////////////////////////////////////////////
// 这里是 statement，不是 expression；
// 和 parseProgram 逻辑是一样的，内部是 statements
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	// curToken 固定是 {, 但是不是通过 registerPrefix 设置的;
	// {}  只出现在 if 和 fn definition 中，所以在 parseIfExpression 和 parseFunctionLiteral 中 hard code 了,  也即：
	// 在调用 parseBlockStatement 之前，都 expectPeek 确保是 { 了
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()

		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}

		p.nextToken()
	}

	return block
}

////////////////////////////////////////////////////////////////////////////////
// statement1: let foobar = 123;
// let is fixed chars;
// foobar is identifier, = is skipped, 123 is integer, ; is ignored(skipped)
func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	// expectPeek 如果返回 true，会调用 p.nextToken()，消耗 input 向前推进，curToken 已经变了
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Name: p.curToken.Literal}

	// if expectPeek return true, it will call nextToken, means move curToken to =
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	// move curToken to the right of =
	p.nextToken()

	stmt.Expr = p.parseExpression(LOWEST)

	// skip ; if any at the end
	for p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// statement-2: return
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	stmt.Expr = p.parseExpression(LOWEST)

	for p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

////////////////////////////////////////////////////////////////////////////////
// statement-3: expressionStatement
var nestLevel = 0

func leading() string {
	return strings.Repeat("    ", nestLevel)
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	fmt.Printf("ExpressionStatement. curToken=%q\n", p.curToken)

	// 一个 statement 只会执行一次，并且 Token 等于第一个 token
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	// 以 LOWEST 为开始，表示开始解析 expression
	stmt.Expr = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt

}

////////////////////////////////////////////////////////////////////////////////
// 处理 expression 的入口，核心逻辑是：对每个 token 绑定 prefixfn 或 infixfn
// 执行一次 prefixFn，并且如果后续操作符优先级高的话，循环调用 infixFn 递归；

// 当前是 token 时，prefixfn 的目的是把 token 转换成 expression；这里碰到的 token 有两类：
// 1.直接可以转成 expr，比如：identifier/integer/boolean；
// 2.有 prefix 前缀的（比如：!, -），parsePrefixExpression 在读取了前缀后，将会再次调用 parseExpression 变成可直接转换成 expr

// infixfn 的目的是处理中缀运算符，比如：当 input 为 a + b + c 时，第一次进入 infixfn 时，
// curToken=a, precedence=LOWEST,进入infixfn 之后，会 next，把 curToken 变成第一个+

// parseExpression is the start point for build ast for expression.
// 和 prefixFn  和 infixFn 形成递归
// 处理一个新的 expressionStatement，传入的是 LOWEST；
// 后续 prefixFn 和 infixFn 传入的是 左边操作符的 precedence；
func (p *Parser) parseExpression(precedence int) ast.Expression {
	nestLevel++
	fmt.Printf("%sExpr >> curToken=%q, precedence=%d\n",
		leading(), p.curToken, precedence)

	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}

	// 递归
	// prefix 不仅仅是 prefixExpression，而是包括所有通过 registerPrefixFn 注册的函数，
	// 包括：Identifier/Integer/Boolean/Prefix
	leftExp := prefix()

	// show log info
	if precedence < p.peekPrecedence() {
		fmt.Printf("%sLoop-Got curToken=%q, precedence=%d, peekToken=%q, peekPrecedence=%d\n",
			leading(), p.curToken, precedence, p.peekToken, p.peekPrecedence())
	} else {
		fmt.Printf("%sLoop-Skip curToken=%q, precedence=%d, peekToken=%q, peekPrecedence=%d\n",
			leading(), p.curToken, precedence, p.peekToken, p.peekPrecedence())
	}

	// local variable to track the loop times
	// nestLevel is global to track the depth
	loopIndex := 0

	// 这里是循环，循环里还有递归
	// precedence is from parameter, and will not change during the loop, however,
	// p.peekPrecedence() will change, cause p.nextToken() is called during the loop.
	// after call infix(leftExp) will create another new call stack, which
	// will have different precedence(set in parseInfix).
	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		loopIndex++
		nestLevel++
		fmt.Printf("%s%d >> curToken=%q, precedence=%d, peekToken=%q, peekPrecedence=%d\n",
			leading(), loopIndex, p.curToken, precedence, p.peekToken, p.peekPrecedence())

		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		// 这里 next 之后，curToken 变成了 operator
		p.nextToken()

		// 外边的循环，这里的迭代赋值
		leftExp = infix(leftExp)

		fmt.Printf("%s%d << curToken=%q, precedence=%d, peekToekn=%q, peekPrecedence=%d\n",
			leading(), loopIndex, p.curToken, precedence, p.peekToken, p.peekPrecedence())
		nestLevel--
	}

	fmt.Printf("%sExpr << curToken=%q, precedence=%d\n",
		leading(), p.curToken, precedence)
	nestLevel--

	return leftExp
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

////////////////////////////////////////////////////////////////////////////////
// prefixFn 有很多，不仅仅是这一个，只是这个碰巧名字中含有 prefix, 处理 ! -  单目前缀运算符
func (p *Parser) parsePrefixExpression() ast.Expression {
	nestLevel++
	fmt.Printf("%sPrefix >> curToken=%q\n", leading(), p.curToken)

	exp := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	// 前缀要处理的是 !a ， 而  curToken=!，所以需要 next
	// prefix-1/2/3 不需要 next
	p.nextToken()

	// 这里是递归，两个函数之间
	// PREFIX 的优先级是最高的，也即，后续不可能有更高的优先级
	exp.Right = p.parseExpression(PREFIX)

	fmt.Printf("%sPrefix << curToken=%q\n",
		leading(), p.curToken)
	nestLevel--

	return exp
}

// prefixFn-1
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Name: p.curToken.Literal}
}

// prefixFn-2
func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	// here cast the string to int.
	// cast 的依据是 token.Type, token.Type 是 lexer 中根据 pattern 读取 input 中的 string 解析出来的
	// the casting could also be left to eval，如果那样的话，eval 的依据是 这里返回的 node 的类型 (&ast.IntegerLiteral)
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value

	return lit
}

// prefix-3
func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	arr := &ast.ArrayLiteral{Token: p.curToken}

	arr.Elements = p.parseExpressionList(token.RBRACKET)

	return arr
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()

		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

// prefix-4 处理小括号
func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

// expression: fn(a, b) { a + b; }
// this is fun definition, not call
// Notice: no name for the fn, only can use let to assign
// ast.Expression is interface. the actual return value is struct pointer
func (p *Parser) parseFunctionLiteral() ast.Expression {
	// actual return value is pointer, which impl the interface
	// all the ast.FunctionLiteral is the same token, no name for fn
	// curToken is fixed to fn, see registerInfix
	fn := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	fn.FormalParams = p.parseFormalParams()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	fn.Body = p.parseBlockStatement()

	return fn
}

// 函数定义时的 形参，只能是 identifier, 只需要 name，不需要 eval;
// 形参的 name 在  callExpression 的 eval 时使用，作为 实参 的 name，保存在 callEnv 中
func (p *Parser) parseFormalParams() []*ast.Identifier {
	// ast.Identifier is struct, so use the pointer.
	identifiers := []*ast.Identifier{}

	//  没有参数
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	// 第一个参数
	p.nextToken()

	ident := &ast.Identifier{Token: p.curToken, Name: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		//  后续参数
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Name: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	// 参数的右括号
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}

////////////////////////////////////////////////////////////////////////////////
// 中缀运算符有很多个，可以对应不同的处理逻辑（区别在于返回不同类型的 Expression），比如：函数调用的 callExpression
// 这个处理函数中，名字碰巧有 infix
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	nestLevel++
	fmt.Printf("%sInfix >> curToken=%q\n", leading(), p.curToken)

	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	// 要处理的是 a + b * c， 假设 curToken=+（第一次进入，在 for 中有 next，从 a 变成了+）,
	// 则 precedure 为 + 的 precedence, next 之后，curToken=b,  注意，
	// 这里只是把 b 变成为当前 token，并没有把 b 解析成 expression, 只有在 prefix-1/2/3 中才会被转化成 expression
	precedure := p.curPrecedence()
	p.nextToken()

	// 这里是递归，两个函数之间
	// 把 + 的优先级传入，后续会和 * 的优先级比较
	expression.Right = p.parseExpression(precedure)

	fmt.Printf("%sInfix << curToken=%q\n", leading(), p.curToken)
	nestLevel--

	return expression
}

// 函数调用，这是 (  infix  op 的处理逻辑，fn 是 left
func (p *Parser) parseCallExpression(fn ast.Expression) ast.Expression {
	expr := &ast.CallExpression{Token: p.curToken, CallableName: fn}

	expr.ActualParams = p.parseExpressionList(token.RPAREN)

	return expr
}

// 数组下标操作
func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()

	exp.Index = p.parseExpression(LOWEST)

	// skip the right ]
	if !p.expectPeek(token.RBRACKET) {
		return nil
	}
	return exp
}

////////////////////////////////////////////////////////////////////////////////
// compound expression: if ( condition ) { consequence } else  { alternative }
func (p *Parser) parseIfExpression() ast.Expression {
	expr := &ast.IfExpression{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	expr.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	expr.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		expr.Alternative = p.parseBlockStatement()
	}

	return expr
}
