package evaluator

import (
	"testing"
	"xmonkey/lexer"
	"xmonkey/object"
	"xmonkey/parser"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 +5+5+5-10", 10},
		{"(5+10*2+15/3)*2+-10", 50},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		testIntegerObject(t, evaluated, tt.expected)
	}
}

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	env := object.NewEnvironment()

	return Eval(program, env)
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not integer, got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%d, want=%d", result.Value, expected)
		return false

	}

	return true
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2;", true},
		{"1 > 2", false},
		{"1 ==1;", true},
		{"1 !=1;", false},
		{"1 ==2;", false},
		{"true == false", false},
		{"true == true", true},
		{"false == false", true},
		{"(1<2) == true;", true},
		{"(1<2) == false", false},
		{"(1>2) == true;", false},
		{"(1>2) == false", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. got=%t, want=%t", result.Value, expected)
		return false
	}

	return false
}

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestIfExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10 };", 10},
		{"if (false) { 10} ", nil},
		{"if (1) {10}", 10},
		{"if (!1) {10}", nil},
		{"if (1<2) {10}", 10},
		{"if (1>2) {10}", nil},
		{"if (1<2) {10} else {5}", 10},
		{"if (1>2) {10} else {5}", 5},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNull(t, evaluated)
		}
	}
}

func testNull(t *testing.T, obj object.Object) bool {
	if obj != NULL {
		t.Errorf("object is not NULL, got=%T (%+v)", obj, obj)
		return false
	}
	return true
}

func TestReturnStatement(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"return 10;", 10},
		{"return (3*2+5)", 11},
		{"return 10; 9", 10},
		{`if (10>1) { return 10} return 1;`, 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"5 + true; 5", "type mismatch: INTEGER + BOOLEAN"},
		{"-true; ", "unknown operator: -BOOLEAN"},
		{"true + false", "unknown operator: BOOLEAN + BOOLEAN"},
		{"if (10 > 1) { true - false}", "unknown operator: BOOLEAN - BOOLEAN"},
		{"foobar", "identifier not found: foobar"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned. got=%T (%+v)", evaluated, evaluated)
			continue
		}

		if errObj.Message != tt.expected {
			t.Errorf("wrong error message. expected=%q, got=%q", tt.expected, errObj.Message)
		}
	}
}

func TestLetStatement(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; let c = a + b + 10; c;", 20},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestFunctionDef(t *testing.T) {
	input := `fn(x) { x + 2; };`

	evaluated := testEval(input)
	fn, ok := evaluated.(*object.Function)
	if !ok {
		t.Fatalf("object is not Function def. got=%T (%+v)", evaluated, evaluated)
	}

	if len(fn.FormalParams) != 1 {
		t.Fatalf("parameters len is wrong. got=%d, want=1", len(fn.FormalParams))
	}

	if fn.FormalParams[0].String() != "x" {
		t.Fatalf("param1 is wrong. got=%s, want=x", fn.FormalParams[0])
	}

	expectedBody := "(x+2)"

	if fn.Body.String() != expectedBody {
		t.Fatalf("body is not %q. got=%q", expectedBody, fn.Body.String())
	}
}

func TestFunctionApplicaiton(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let f = fn(x) { x }; f(5); ", 5},
		{"let f = fn(x) { return x; }; f(10); ", 10},
		{"let double=fn(x) { 2 * x }; double(3); ", 6},
		{"let add=fn(x, y) {x + y}; add(1+2, 3*4); ", 15},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestClosure(t *testing.T) {
	input := `
let adder = fn(x) {
    fn(y) { x + y}
}

let addTwo = adder(2)

addTwo(5);

`

	testIntegerObject(t, testEval(input), 7)
}

func TestCompound(t *testing.T) {
	input := `
let add = fn(a, b) { a + b }
let sub = fn(a, b) { a - b }
let apply = fn(a, b, f) { f(a, b) }

apply(2, 5, add);

`
	testIntegerObject(t, testEval(input), 7)
}
