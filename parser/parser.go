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
	PREFIX
	CALL
)

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l         *lexer.Lexer
	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn

	errors []string
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)

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
		stmt := p.parseStatement()

		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}

		// 读取下一个 token，也即消耗 input
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
// statement1: let
func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	// expectPeek 如果返回 true，会调用 p.nextToken()，消耗 input 向前推进，curToken 已经变了
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	// TODO: will replace later on
	for !p.curTokenIs(token.SIMICOLON) {
		p.nextToken()
	}

	return stmt
}

// statement-2: return
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	// TODO: skip expressions for now
	for !p.curTokenIs(token.SIMICOLON) {
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

	// 以 LOWEST 为开始，表示还是解析 expression
	stmt.ExpressionValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SIMICOLON) {
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
//
// 和 parsePrefixExpression  和 parseInfixExpression 形成递归
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
	for !p.peekTokenIs(token.SIMICOLON) && precedence < p.peekPrecedence() {
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

// prefixFn-1
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// prefixFn-2
func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

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

// prefix-4
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

////////////////////////////////////////////////////////////////////////////////
// infix-1
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
