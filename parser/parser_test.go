package parser

import (
	"fmt"
	"testing"

	"xmonkey/ast"
	"xmonkey/lexer"
)

////////////////////////////////////////////////////////////////////////////////
// errors during the parsing
func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()

	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}

	t.FailNow()
}

////////////////////////////////////////////////////////////////////////////////
// 2 statements: let, return
// let statement, check the varibale/binding name
func TestLetStatements(t *testing.T) {
	input := `
let x = 5;
let y = 10;
let foobar = 838383;
`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}
	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d", len(program.Statements))
	}

	// table-driven testing
	tests := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"foobar"},
	}

	for i, tt := range tests {
		stmt := program.Statements[i]
		if !testLetStatement(t, stmt, tt.expectedIdentifier) {
			return
		}
	}
}

func testLetStatement(t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != "let" {
		t.Errorf("s.TokenLiteral no 'let'. got=%q", s.TokenLiteral())
		return false
	}

	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		t.Errorf("s not *ast.LetStatement. got=%T", s)
		return false
	}

	if letStmt.Name.Name != name {
		t.Errorf("letStmt.Name.Value not '%s'. got=%s", name, letStmt.Name.Name)
		return false
	}

	if letStmt.Name.TokenLiteral() != name {
		t.Errorf("s.Name not '%s'. got=%s", name, letStmt.Name)
		return false
	}

	return true
}

func TestLetStatements2(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{"let x = 5; ", "x", 5},
		{"let y = true; ", "y", true},
		{"let foobar = y;", "foobar", "y"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)

		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Errorf("does not contain 1 statement, got=%d", len(program.Statements))
		}

		stmt := program.Statements[0]
		if !testLetStatement(t, stmt, tt.expectedIdentifier) {
			t.Errorf("identifier is not %s, got=%s", tt.expectedIdentifier, stmt.TokenLiteral())
			return
		}

		val := stmt.(*ast.LetStatement).Expr

		if !testLiteralExpression(t, val, tt.expectedValue) {
			return
		}

	}
}

// return statements
func TestReturnStatements(t *testing.T) {
	input := `
return 5;
return 10;
return 993322;

`
	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}
	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d", len(program.Statements))
	}

	for _, stmt := range program.Statements {
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Errorf("stmt not *ast.returnstatement. got=%T", stmt)
			continue
		}

		if returnStmt.TokenLiteral() != "return" {
			t.Errorf("returnStmt.TokenLiteral not 'return', got %q", returnStmt.TokenLiteral())
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
//  3-expressions
// 最原始的测试方法，每次只测试一个 statement
// programe --> statements --> statements[0]
// statements[o] -> expressionStatement --> Identifier/IntergeLiteral/Boolean
func TestIdentifierExpression(t *testing.T) {
	input := "foobar"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has not enough statements. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement, got=%T", program.Statements[0])
	}

	ident, ok := stmt.Expr.(*ast.Identifier)
	if !ok {
		t.Fatalf("exp not *ast.Identifier. got=%T", stmt.Expr)
	}

	if ident.Name != "foobar" {
		t.Errorf("ident.Value not %s. got=%s", "foobar", ident.Name)
	}

	if ident.TokenLiteral() != "foobar" {
		t.Errorf("ident.TokenLiteral not %s. got=%s", "foobar", ident.TokenLiteral())
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "5"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has not enough statements. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement, got=%T", program.Statements[0])
	}

	literal, ok := stmt.Expr.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("exp not *ast.IntegerLiteral. got=%T", stmt.Expr)
	}

	if literal.Value != 5 {
		t.Errorf("literal.Value not %d. got=%d", 5, literal.Value)
	}

	if literal.TokenLiteral() != "5" {
		t.Errorf("literal.TokenLiteral not %s. got=%s", "5", literal.TokenLiteral())
	}
}

func TestBooleanExpression(t *testing.T) {
	input := `true;`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has not enough statements. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement, got=%T", program.Statements[0])
	}

	expr, ok := stmt.Expr.(*ast.Boolean)
	if !ok {
		t.Fatalf("expr not *ast.Boolean. got=%T", stmt.Expr)
	}

	if expr.Value != true {
		t.Errorf("expr.Value not %t. got=%t", true, expr.Value)
	}
}

////////////////////////////////////////////////////////////////////////////////
// 统一的测试 assertions，在 prefix/infix 中使用
func testLiteralExpression(t *testing.T, exp ast.Expression, expected interface{}) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case string:
		return testIdentifier(t, exp, v)
	case bool:
		return testBooleanLiteral(t, exp, v)

	}

	t.Errorf("type of exp not handled. got=%T", exp)
	return false
}

// 使用到的 3 个具体的 assertions: integer/identifier/boolean
func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) bool {
	integ, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("il not *ast.IntegerLiteral. got=%T, %s", il, il.String())
		return false
	}

	if integ.Value != value {
		t.Errorf("integ.Value not %d. got=%d", value, integ.Value)
		return false
	}

	if integ.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf("integ.TokenLiteral not %d. got=%s", value, integ.TokenLiteral())
		return false
	}

	return true
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("exp not *ast.Identifier. got=%T %s", exp, exp.String())
		return false
	}

	if ident.Name != value {
		t.Errorf("ident.Value not %s. got=%s", value, ident.Name)
		return false
	}

	if ident.TokenLiteral() != value {
		t.Errorf("ident.TokenLiteral not %s. got=%s", value, ident.TokenLiteral())
		return false
	}

	return true
}

func testBooleanLiteral(t *testing.T, exp ast.Expression, value bool) bool {
	bo, ok := exp.(*ast.Boolean)
	if !ok {
		t.Errorf("exp no *ast.Boolean. got=%T", exp)
		return false
	}

	if bo.Value != value {
		t.Errorf("bo.Value not %t. got=%t", value, bo.Value)
		return false
	}

	if bo.TokenLiteral() != fmt.Sprintf("%t", value) {
		t.Errorf("bo.TokenLiteral not %t. got=%s", value, bo.TokenLiteral())
		return false
	}

	return true
}

////////////////////////////////////////////////////////////////////////////////
// testLiteralExpression 的一个应用
func testInfixExpression(t *testing.T, exp ast.Expression,
	left interface{}, operator string, right interface{}) bool {
	opExp, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Errorf("exp is not ast.OperatorExpression. got=%T(%s)", exp, exp)
		return false

	}

	if !testLiteralExpression(t, opExp.Left, left) {
		return false
	}

	if opExp.Operator != operator {
		t.Errorf("exp.Operator is not '%s'. got=%q", operator, opExp.Operator)
		return false
	}

	if !testLiteralExpression(t, opExp.Right, right) {
		return false
	}

	return true
}

func TestInfixExpression2(t *testing.T) {
	input := "5 + 10;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0]
	s, ok := stmt.(*ast.ExpressionStatement)
	if !ok {
		t.Errorf("not expressionStatement")
	}

	if !testInfixExpression(t, s.Expr, 5, "+", 10) {
		t.Errorf("failed.")
	}
}

////////////////////////////////////////////////////////////////////////////////
// prefix expression
func TestParsingPrefixExpression(t *testing.T) {
	prefixTests := []struct {
		input    string
		operator string
		value    interface{}
	}{
		{"!5", "!", 5},
		{"-15;", "-", 15},
		{"!false", "!", false},
	}

	for _, tt := range prefixTests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d", 1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement, got=%T", program.Statements[0])
		}

		exp, ok := stmt.Expr.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("exp not *ast.PrefixExpression. got=%T", stmt.Expr)
		}

		if exp.Operator != tt.operator {
			t.Fatalf("exp.Operator is not '%s'. got=%s", tt.operator, exp.Operator)
		}

		if !testLiteralExpression(t, exp.Right, tt.value) {
			return
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// infix expression
func TestParsingInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  interface{}
		operator   string
		rightValue interface{}
	}{
		{"5 + 5; ", 5, "+", 5},
		{"5 - 5; ", 5, "-", 5},
		{"5 * 5; ", 5, "*", 5},
		{"5 / 5; ", 5, "/", 5},
		{"5 > 5; ", 5, ">", 5},
		{"5 < 5; ", 5, "<", 5},
		{"5 == 5; ", 5, "==", 5},
		{"5 != 5; ", 5, "!=", 5},
		{"true != false; ", true, "!=", false},
	}

	for _, tt := range infixTests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d", 1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement, got=%T", program.Statements[0])
		}

		exp, ok := stmt.Expr.(*ast.InfixExpression)
		if !ok {
			t.Fatalf("exp not *ast.PrefixExpression. got=%T", stmt.Expr)
		}

		if !testLiteralExpression(t, exp.Left, tt.leftValue) {
			return
		}

		if exp.Operator != tt.operator {
			t.Fatalf("exp.Operator is not '%s'. got=%s", tt.operator, exp.Operator)
		}

		if !testLiteralExpression(t, exp.Right, tt.rightValue) {
			return
		}
	}
}

// 符号优先级
func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"-a * b", "((-a)*b)"},
		{"!-a", "(!(-a))"},
		{"a + b + c", "((a+b)+c)"},
		{"a + b - c", "((a+b)-c)"},
		{"a * b * c", "((a*b)*c)"},
		{"a * b / c", "((a*b)/c)"},
		{"a + b / c", "(a+(b/c))"},
		{"a + b * c + d / e -f", "(((a+(b*c))+(d/e))-f)"},
		{"true", "true"},
		{" 3 > 5 == false", "((3>5)==false)"},
		{"1 * (2 + 3) * 4", "((1*(2+3))*4)"},
		{"!(true == false);", "(!(true==false))"},
		{"a + add(b *c) +d", "((a+add((b*c)))+d)"},
		{"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8))",
			"add(a,b,1,(2*3),(4+5),add(6,(7*8)))"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		actual := program.String()

		if actual != tt.expected {
			t.Errorf("expected %q, got=%q", tt.expected, actual)
			return
		}
	}
}

func TestSingleParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// {"a + b + c;", "((a+b)+c)"},
		{"a + b * c; ", "(a+(b*c))"},
		// {"a + b", "(a+b)"},
		// {"a + b + c + d + e; ", "((((a+b)+c)+d)+e)"},
		// {"a + b * c +d; ", "((a+(b*c))+d)"},
		// {"a + b * !c", "(a+(b*(!c)))"},
		{"a * [1,2,3,4][b*c] * d",
			"((a*([1,2,3,4][(b*c)]))*d)"},
		{"add(a*b[2], b[1], 2 * [1,2][1])",
			"add((a*(b[2])),(b[1]),(2*([1,2][1])))"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		actual := program.String()

		if actual != tt.expected {
			t.Errorf("expected %q, got=%q", tt.expected, actual)
			return
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// test if-then-else
func TestIfExpression(t *testing.T) {
	input := "if (x < y) {x}"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Body does not contain %d statements. got=%d\n",
			1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	exp, ok := stmt.Expr.(*ast.IfExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.IfExpression. got=%T", stmt.Expr)
	}

	if !testInfixExpression(t, exp.Condition, "x", "<", "y") {
		return
	}

	if len(exp.Consequence.Statements) != 1 {
		t.Errorf("consequence is not 1 statement. got=%d\n", len(exp.Consequence.Statements))
	}

	consequence, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not ast.ExpressionStatement. got=%T", exp.Consequence.Statements[0])
	}

	if !testIdentifier(t, consequence.Expr, "x") {
		return
	}

	if exp.Alternative != nil {
		t.Errorf("exp.Alternative.Statements was not nil. got=%+v", exp.Alternative)
	}
}

func TestIfElseExpression(t *testing.T) {
	input := "if (x < y) {x} else {y}"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Body does not contain %d statements. got=%d\n",
			1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	exp, ok := stmt.Expr.(*ast.IfExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.IfExpression. got=%T", stmt.Expr)
	}

	if !testInfixExpression(t, exp.Condition, "x", "<", "y") {
		return
	}

	if len(exp.Consequence.Statements) != 1 {
		t.Errorf("consequence is not 1 statement. got=%d\n", len(exp.Consequence.Statements))
	}

	consequence, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not ast.ExpressionStatement. got=%T", exp.Consequence.Statements[0])
	}

	if !testIdentifier(t, consequence.Expr, "x") {
		t.Fatalf("consequence.ExpressionValue is not 'x', got=%s", consequence.Expr)
		return
	}

	if exp.Alternative == nil {
		t.Error("exp.Alternative.Statements was nil.")
	}

	alt, ok := exp.Alternative.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("alt.Statements[0] is not ast.ExpressionStatement. got=%T", exp.Alternative.Statements[0])
	}

	if !testIdentifier(t, alt.Expr, "y") {
		t.Fatalf("alt.ExpressionValue is not 'y', got=%s", alt.Expr)
		return
	}
}

func TestFuncDef(t *testing.T) {
	input := `fn(x, y) { x + y; }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Body does not contain %d statements, got=%d", 1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("statement[0] is not ast.ExpressionStatement, got=%T", program.Statements[0])
	}

	fn, ok := stmt.Expr.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.FunctionLiteral, got=%T", stmt.Expr)
	}

	if len(fn.FormalParams) != 2 {
		t.Fatalf("fn parameters is not 2, got=%d", len(fn.FormalParams))
	}

	testLiteralExpression(t, fn.FormalParams[0], "x")
	testLiteralExpression(t, fn.FormalParams[1], "y")

	if len(fn.Body.Statements) != 1 {
		t.Fatalf("fn.Body.Statements has not 1 statement, got=%d", len(fn.Body.Statements))
	}

	bodyStmt, ok := fn.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("fn.Body is not ast.ExpressionStatement, got=%T", fn.Body.Statements[0])
	}

	testInfixExpression(t, bodyStmt.Expr, "x", "+", "y")
}

func TestFuncParameters(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{input: "fn() {}", expected: []string{}},
		{input: "fn(x) {};", expected: []string{"x"}},
		{input: "fn(x, y) {}; ", expected: []string{"x", "y"}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Errorf("%s is not expressionStatement, got=%T", tt.input, program.Statements[0])
		}

		fn, ok := stmt.Expr.(*ast.FunctionLiteral)
		if !ok {
			t.Errorf("%s is not FunctionLiteral, got=%T", tt.input, stmt.Expr)
		}

		if len(fn.FormalParams) != len(tt.expected) {
			t.Errorf("parameter lenght wrong. want=%d, got=%d", len(tt.expected), len(fn.FormalParams))
		}

		for i, ident := range tt.expected {
			testLiteralExpression(t, fn.FormalParams[i], ident)
		}

	}
}

func TestCallExpression(t *testing.T) {
	// input := `add(1, a, 2 + 3 * a, !true); `
	input := `add(1, 2 + 3, 4 * 5); `

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Errorf("program.Statements does not contains %d statements. got=%d", 1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Errorf("statements[0] is not ast.ExpressionStatement, got=%T", program.Statements[0])
	}

	exp, ok := stmt.Expr.(*ast.CallExpression)
	if !ok {
		t.Errorf("stmt.Expression is not ast.CallExpression, got=%T", stmt.Expr)
	}

	if !testIdentifier(t, exp.CallableName, "add") {
		return
	}

	if len(exp.ActualParams) != 3 {
		t.Errorf("args lenght wrong. want=%d, got=%d", 3, len(exp.ActualParams))
	}

	testLiteralExpression(t, exp.ActualParams[0], 1)
	testInfixExpression(t, exp.ActualParams[1], 2, "+", 3)
	testInfixExpression(t, exp.ActualParams[2], 4, "*", 5)
}

func TestStringLiteralExpression(t *testing.T) {
	input := `"hello world";`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	literal, ok := stmt.Expr.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("exp no *ast.StringLiteral. got=%T (%+v)", stmt.Expr, stmt.Expr)
	}

	if literal.Value != "hello world" {
		t.Fatalf("literal.Value not %q, got=%q", "hello world", literal.Value)
	}

}

func TestParsingArrayLiterals(t *testing.T) {
	input := `[1, 2 * 2, 3 + 3 ]; `

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	arr, ok := stmt.Expr.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("exp no *ast.ArrayLiteral. got=%T (%+v)", stmt.Expr, stmt.Expr)
	}

	if len(arr.Elements) != 3 {
		t.Fatalf("len(arr.Elements) not 3. got=%d", len(arr.Elements))
	}

	testIntegerLiteral(t, arr.Elements[0], 1)
	testInfixExpression(t, arr.Elements[1], 2, "*", 2)
	testInfixExpression(t, arr.Elements[2], 3, "+", 3)
}

func TestParsingIndexExpression(t *testing.T) {
	input := `myArray[1 + 1]; `

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	indExpr, ok := stmt.Expr.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("exp no *ast.IndexExpression. got=%T (%+v)", stmt.Expr, stmt.Expr)
	}

	if !testIdentifier(t, indExpr.Left, "myArray") {
		return
	}

	if !testInfixExpression(t, indExpr.Index, 1, "+", 1) {
		return
	}

}
