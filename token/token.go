package token

type TokenType string

// All input are string, no matter it is 125, "125", foo, "foo"
// based on the pattern(grammar rule), we can treat them as int, string, identifier, string, ect.
// pattern means start with what, end with what. ect.

// Token stream is the lexer output, and will feed to parser to build the AST
// Type and Literal are all string.
// input = 125, and after lexer the token is string "125", after parser will cast to int 125
type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL = "ILLEGA"
	EOF     = "EOF"

	IDENT = "IDENT"
	INT   = "INT"

	// Operaters
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"

	LT = "<"
	GT = ">"

	EQ     = "=="
	NOT_EQ = "!="

	// Delimiters
	COMMA     = ","
	SEMICOLON = ";"

	LPAREN = "("
	RPAREN = ")"
	LBRACE = "{"
	RBRACE = "}"

	STRING = "STRING"

	LBRACKET = "["
	RBRACKET = "]"

	// Keywords
	FUNCTION = "FUNCTION"
	LET      = "LET"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
)

// keywords mean something predefined(a subset of identifier),
// true/false is also keywords
var keywords = map[string]TokenType{
	"true":   TRUE,
	"false":  FALSE,
	"let":    LET,
	"return": RETURN,
	"fn":     FUNCTION,
	"if":     IF,
	"else":   ELSE,
}

// LookupIdent first find in keyword list, if not exist, then it should be identifier
// ATTENTION: keywords and identifiers are all strings which have no double quotes in the input.
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
