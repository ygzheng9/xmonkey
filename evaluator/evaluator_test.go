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
		{`"Hello" - "World"`, "unknown operator: STRING - STRING"},
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

func TestStringLiteral(t *testing.T) {
	input := `"Hello world!`

	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("Object is not string, got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "Hello world!" {
		t.Errorf("String has wrong value, got=%q", str.Value)
	}

}

func TestStringConcatenation(t *testing.T) {
	input := `"Hello" + " " + "world!"`

	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("Object is not string, got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "Hello world!" {
		t.Errorf("String has wrong value, got=%q", str.Value)
	}

}

func TestBuiltinFunc(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{"len(1)", "argument to len not supported, got INTEGER"},
		{`len("one", "two")`, "wrong number of arguments. got=2, want=1"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		case string:
			errObj, ok := evaluated.(*object.Error)
			if !ok {
				t.Errorf("object is not Error. got=%T (%+v)", evaluated, evaluated)
				continue
			}

			if errObj.Message != expected {
				t.Errorf("wrong error message, got=%q, want=%q", errObj.Message, expected)
			}
		}
	}

}

func TestArrayLiterals(t *testing.T) {
	input := "[1, 2*2, 3 + 3]"

	evaluated := testEval(input)
	result, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("object is not array. got=%T (%+v)", evaluated, evaluated)
	}

	if len(result.Elements) != 3 {
		t.Fatalf("array has wrong num of elements. got=%d", len(result.Elements))
	}

	testIntegerObject(t, result.Elements[0], 1)
	testIntegerObject(t, result.Elements[1], 4)
	testIntegerObject(t, result.Elements[2], 6)
}

func TestArrayIndexExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"[1,2,3][0]", 1},
		{"[1,2,3][1]", 2},
		{"[1,2,3][2]", 3},
		{"let myArray = [1,2,3]; let b = myArray[0];  myArray[1]; ", 2},
		{"[1,2,3][3]", nil},
		{"first([1,2,3])", 1},
		{"last([1,2,3])", 3},
		{"let a=3; let b = 4; let f = fn(x, y) { x*y + x + y};  last([1, 3*3, f(a,b)])", 19},
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

func TestArrayRest(t *testing.T) {
	input := "let a = [1, 3, 2 * 5]; let b = a;  rest(b)"

	evaluated := testEval(input)
	result, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("object is not array. got=%T (%+v)", evaluated, evaluated)
	}

	if len(result.Elements) != 2 {
		t.Fatalf("array has wrong num of elements. got=%d", len(result.Elements))
	}

	testIntegerObject(t, result.Elements[0], 3)
	testIntegerObject(t, result.Elements[1], 10)
}

func TestArrayPush(t *testing.T) {
	input := "let a = [3, 2 * 5]; let b = a;  let c = push(b, 99); c"

	evaluated := testEval(input)
	result, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("object is not array. got=%T (%+v)", evaluated, evaluated)
	}

	if len(result.Elements) != 3 {
		t.Fatalf("array has wrong num of elements. got=%d", len(result.Elements))
	}

	testIntegerObject(t, result.Elements[0], 3)
	testIntegerObject(t, result.Elements[1], 10)
	testIntegerObject(t, result.Elements[2], 99)
}
