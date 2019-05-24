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
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()

		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}

		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
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
func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

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
var nestLevel = 0

func leading() string {
	return strings.Repeat("    ", nestLevel)
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	fmt.Printf("ExpressionStatement. curToken=%q\n", p.curToken)

	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.ExpressionValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SIMICOLON) {
		p.nextToken()
	}

	return stmt

}

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

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

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

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	nestLevel++
	fmt.Printf("%sPrefix >> curToken=%q\n", leading(), p.curToken)

	exp := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()
	// 这里是递归
	exp.Right = p.parseExpression(PREFIX)

	fmt.Printf("%sPrefix << curToken=%q\n",
		leading(), p.curToken)
	nestLevel--

	return exp
}

////////////////////////////////////////////////////////////////////////////////
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	nestLevel++
	fmt.Printf("%sInfix >> curToken=%q\n", leading(), p.curToken)

	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedure := p.curPrecedence()
	p.nextToken()

	// 这里是递归
	expression.Right = p.parseExpression(precedure)

	fmt.Printf("%sInfix << curToken=%q\n", leading(), p.curToken)
	nestLevel--

	return expression

}
