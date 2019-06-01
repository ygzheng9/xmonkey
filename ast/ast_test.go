package ast

import (
	"testing"

	"xmonkey/token"
)

func TestString(t *testing.T) {
	// let myVar = anotherVar;
	program := &Program{
		Statements: []Statement{
			&LetStatement{
				Token: token.Token{Type: token.LET, RawString: "let"},
				Name: &Identifier{
					Token: token.Token{Type: token.IDENT, RawString: "myVar"},
					Name:  "myVar",
				},
				Expr: &Identifier{
					Token: token.Token{Type: token.IDENT, RawString: "anotherVar"},
					Name:  "anotherVar",
				},
			},
		},
	}

	if program.String() != "let myVar = anotherVar;" {
		t.Errorf("program.String() wrong. got=%q", program.String())
	}
}
