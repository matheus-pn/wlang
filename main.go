package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type SourceFile struct {
	filename   string
	byteSource []byte
}

func OpenSource(filename string) (*SourceFile, error) {
	input, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return &SourceFile{filename, input}, nil
}

func (file *SourceFile) Text() string {
	return string(file.byteSource)
}

func (file *SourceFile) Runes() []rune {
	return []rune(string(file.byteSource))
}

type Value interface{}

type Expression struct {
	Literal   Value
	Operation string
	Operands  []Expression
	Position  *Token
}

type Statement struct {
	Value      Token
	Flag       string
	Expression *Expression
	Statements []Statement
}

type Parser struct {
	TokenizedFile *TokenizedFile
	Index         int
}

func (parser *Parser) CheckTokenAt(index int) Token {
	if index >= len(parser.TokenizedFile.tokens) {
		return Token{Flag: "EOF"}
	}
	return parser.TokenizedFile.tokens[index]
}

func (parser *Parser) CurrentToken() Token {
	return parser.CheckTokenAt(parser.Index)
}

func (parser *Parser) Peek() Token {
	return parser.CheckTokenAt(parser.Index + 1)
}

func Include(strs []string, s string) bool {
	for _, str := range strs {
		if str == s {
			return true
		}
	}
	return false
}

func (parser *Parser) Next() {
	parser.Index++
}

func (parser *Parser) NextWithoutWhitespace() {
	parser.Index++
	parser.SkipWhitespace()
}

func (parser *Parser) SkipWhitespace() {
	token := parser.CurrentToken()
	for token.Flag == "NewLine" {
		parser.Index++
		token = parser.CurrentToken()
	}
}

func (parser *Parser) WaitUntil(flag string) bool {
	token := parser.CurrentToken()
	return token.Flag != flag && token.Flag != "EOF"
}

func (parser *Parser) Error(message string, errorToken Token) error {
	errorLine := fmt.Sprintf(
		"parser error: %v at %v:%d:%d",
		message, parser.TokenizedFile.file.filename, errorToken.Line, errorToken.Column,
	)
	// TODO: Add line context
	return fmt.Errorf(errorLine)
}

func (parser *Parser) Expect(flags ...string) (token Token, err error) {
	token = parser.CurrentToken()

	if !Include(flags, token.Flag) {
		err = parser.Error(
			fmt.Sprintf("expected one of %v got %v", strings.Join(flags, ", "), token.Flag),
			token,
		)
	}
	return
}

func (parser *Parser) ExpectConsumeWithWhitespace(flags ...string) (token Token, err error) {
	token, err = parser.Expect(flags...)
	parser.Next()
	return
}

func (parser *Parser) ExpectConsume(flags ...string) (token Token, err error) {
	token, err = parser.Expect(flags...)
	parser.NextWithoutWhitespace()
	return
}

func IsRHSOperator(operator string) bool {
	switch operator {
	case "Dot", "Plus", "Minus", "Star", "FowardSlash", "EqualsEquals", "BangEquals", "Equal":
		return true
	default:
		return false
	}
}

// https://en.cppreference.com/w/c/language/operator_precedence
func Assoc(operator string) int {
	// 1 => left-to-right
	// 0 => right-to-left
	switch operator {
	case "Dot", "Plus", "Minus", "Star", "FowardSlash", "EqualsEquals", "BangEquals":
		return 1
	case "Equal":
		return 0
	default:
		return 1
	}
}

// https://en.cppreference.com/w/c/language/operator_precedence
// inverted here
func Precedence(operator string) int {
	switch operator {
	case "Equal":
		return 1
	case "EqualsEquals", "BangEquals":
		return 3
	case "Plus", "Minus":
		return 4
	case "FowardSlash", "Star":
		return 7
	case "Dot":
		return 14
	default:
		return 0
	}
}

func RHSExpression(parser *Parser, leftExpr Expression, operation Token, nextPrec int) (Expression, error) {
	parser.Next()
	rightExpr, err := ParseExpression(parser, nextPrec)
	if err != nil {
		return rightExpr, err
	}
	leftExpr = Expression{
		Operation: operation.Flag,
		Operands:  []Expression{leftExpr, rightExpr},
		Position:  &operation,
	}
	return leftExpr, nil
}

func ParseExpression(parser *Parser, minPrec int) (Expression, error) {
	var leftExpr Expression
	var err error

	leftToken := parser.CurrentToken()
	// Literal expression
	if leftToken.Flag == "Number" || leftToken.Flag == "String" {
		leftExpr, err = ParseLiteralExpression(parser)
		if err != nil {
			return leftExpr, err
		}
		// Parenthesised expression
	} else if leftToken.Flag == "Identifier" {
		parser.Next()
		leftExpr = Expression{
			Literal:   leftToken.Value,
			Operation: "Variable",
			Position:  &leftToken,
		}
	} else if leftToken.Flag == "LeftParens" {
		parser.Next()
		leftExpr, err = ParseExpression(parser, 0)
		if err != nil {
			return leftExpr, err
		}
		_, err = parser.ExpectConsume("RightParens")
		if err != nil {
			return leftExpr, err
		}
	} else {
		return leftExpr, parser.Error(
			"Unexpected token: "+leftToken.Flag+" on lhs of expression", leftToken,
		)
	}

	for {
		token := parser.CurrentToken()
		// Binary expression
		if token.Flag == "NewLine" ||
			token.Flag == "EOF" ||
			token.Flag == "RightParens" {
			return leftExpr, nil
		}

		prec := Precedence(token.Flag)
		if minPrec > prec {
			break
		}

		nextMinPrec := prec + Assoc(token.Flag)
		if IsRHSOperator(token.Flag) {
			leftExpr, err = RHSExpression(parser, leftExpr, token, nextMinPrec)
			if err != nil {
				return leftExpr, err
			}
		} else {
			return leftExpr, parser.Error(
				"Unexpected token: "+token.Flag+" on rhs of expression", token,
			)
		}
	}
	return leftExpr, nil
}

func ParseLiteralExpression(parser *Parser) (expr Expression, exprErr error) {
	token, err := parser.ExpectConsumeWithWhitespace("Number", "String")
	if err != nil {
		exprErr = err
		return
	}

	switch token.Flag {
	case "Number":
		// TODO: Add floating point and hex
		number, err := strconv.ParseInt(token.Value, 10, 64)
		if err != nil {
			exprErr = err
			return
		}

		expr = Expression{Operation: "NumberLiteral", Literal: number, Position: &token}
	case "String":
		text := token.Value[1 : len(token.Value)-1]
		expr = Expression{Operation: "StringLiteral", Literal: text, Position: &token}
	}
	return
}

func ParseFunctionBody(parser *Parser, scope *Statement) error {
	token := parser.CurrentToken()
	switch token.Flag {
	case "KeywordIf":
		// TODO: parse if
		parser.NextWithoutWhitespace()
		for parser.WaitUntil("KeywordEnd") {
			ParseFunctionBody(parser, scope)
		}
		parser.NextWithoutWhitespace()
	case "KeywordLoop":
		// TODO: parse loop
		parser.NextWithoutWhitespace()
		for parser.WaitUntil("KeywordEnd") {
			ParseFunctionBody(parser, scope)
		}
		parser.NextWithoutWhitespace()
	default:
		expr, err := ParseExpression(parser, 0)
		parser.Next()
		if err != nil {
			return err
		}

		scope.Statements = append(scope.Statements, Statement{
			Flag:       "Expression",
			Expression: &expr,
		})
	}
	return nil
}

func ParseFunction(parser *Parser, scope *Statement) (errs []error) {
	_, err := parser.ExpectConsume("KeywordFunction")
	if err != nil {
		errs = append(errs, err)
	}
	token, err := parser.ExpectConsume("Identifier")
	if err != nil {
		errs = append(errs, err)
	}
	function := &Statement{Flag: "Function", Value: token}
	ParseAttributesList(parser, function)
	for parser.WaitUntil("KeywordEnd") {
		// TODO: Parse statements inside function
		err := ParseFunctionBody(parser, function)
		if err != nil {
			errs = append(errs, err)
		}
		// parser.NextWithoutWhitespace()
	}
	scope.Statements = append(scope.Statements, *function)
	parser.NextWithoutWhitespace()
	return
}

func ParseClassBody(parser *Parser, class *Statement) (errors []error) {
	for parser.WaitUntil("KeywordEnd") {
		errs := ParseFunction(parser, class)
		errors = append(errors, errs...)
	}
	parser.NextWithoutWhitespace()
	return
}

func ParseAttributesList(parser *Parser, class *Statement) (errors []error) {
	// attribute list is optional
	flag := parser.CurrentToken().Flag
	if flag != "Identifier" && flag != "LeftParens" {
		return
	}

	seenParens := false
	if flag == "LeftParens" {
		seenParens = true
		parser.NextWithoutWhitespace()
	}
	flag = parser.CurrentToken().Flag
	if flag == "RightParens" {
		return
	}

	for {
		token, err := parser.ExpectConsume("Identifier")
		if err != nil {
			errors = append(errors, err)
		}
		attribute := Statement{Flag: "Attribute", Value: token}

		token = parser.CurrentToken()

		if token.Flag == "Equal" {
			parser.NextWithoutWhitespace()
			// TODO: Add parse expression to default value of attribute
			expr, err := ParseLiteralExpression(parser)
			attribute.Expression = &expr
			token = parser.CurrentToken()
			if err != nil {
				errors = append(errors, err)
			}
		}

		class.Statements = append(class.Statements, attribute)

		if token.Flag == "Comma" {
			parser.NextWithoutWhitespace()
		} else {
			// fmt.Println(token)
			break
		}
	}
	if seenParens {
		_, err := parser.ExpectConsume("RightParens")
		if err != nil {
			errors = append(errors, err)
		}
	}
	return
}

func ParseClassInheritance(parser *Parser, root *Statement) error {
	// inheritance is optional
	if parser.CurrentToken().Flag != "LessThan" {
		return nil
	}
	parser.ExpectConsume("LessThan")
	token, err := parser.ExpectConsume("Identifier")
	if err != nil {
		return err
	}

	inherits := Statement{Flag: "Inherits", Value: token}
	root.Statements = append(root.Statements, inherits)
	return nil
}

func ParseClass(parser *Parser, root *Statement) (errors []error) {
	token, err := parser.ExpectConsume("Identifier")
	class := Statement{Flag: "Class", Value: token}
	if err != nil {
		errors = append(errors, err)
	}
	err = ParseClassInheritance(parser, &class)
	if err != nil {
		errors = append(errors, err)
	}
	errs := ParseAttributesList(parser, &class)
	errors = append(errors, errs...)
	errs = ParseClassBody(parser, &class)
	errors = append(errors, errs...)
	root.Statements = append(root.Statements, class)
	return
}

func ParseModuleBody(parser *Parser, class *Statement) (errors []error) {
	for parser.WaitUntil("KeywordEnd") {
		errs := ParseStatement(parser, class)
		errors = append(errors, errs...)
	}
	parser.NextWithoutWhitespace()
	return
}

func ParseModule(parser *Parser, root *Statement) (errors []error) {
	token, err := parser.ExpectConsume("Identifier")
	module := Statement{Flag: "Module", Value: token}
	if err != nil {
		errors = append(errors, err)
	}
	errs := ParseModuleBody(parser, &module)
	errors = append(errors, errs...)
	root.Statements = append(root.Statements, module)
	return
}

func ParseStatement(parser *Parser, root *Statement) (errors []error) {
	token := parser.CurrentToken()

	switch token.Flag {
	case "KeywordModule":
		classErrors := ParseModule(parser, root)
		errors = append(errors, classErrors...)
	case "KeywordClass":
		classErrors := ParseClass(parser, root)
		errors = append(errors, classErrors...)
	case "KeywordFunction":
		funcErrors := ParseFunction(parser, root)
		errors = append(errors, funcErrors...)
	default:
		statement := Statement{Flag: "error-statement", Value: token}
		errors = append(errors, parser.Error("expected statement, found "+token.Flag, token))
		root.Statements = append(root.Statements, statement)
	}
	return
}

const MAX_PARSER_ERROR = 5

func Parse(file *TokenizedFile) (root Statement, errors []error) {
	parser := &Parser{TokenizedFile: file}
	root = Statement{Flag: "Module", Value: Token{Flag: "Identifier", Value: "Main"}}
	parser.SkipWhitespace()
	for parser.CurrentToken().Flag != "EOF" {
		if len(errors) > MAX_PARSER_ERROR {
			break
		}

		declErrors := ParseStatement(parser, &root)
		errors = append(errors, declErrors...)
	}
	if len(errors) > MAX_PARSER_ERROR {
		errors = errors[:MAX_PARSER_ERROR]
	}
	return
}

func main() {
	source, err := OpenSource("expr.w")
	if err != nil {
		return
	}

	tokens, errs := Tokenize(source)
	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Println(err)
		}
		return
	}

	tokenizedFile := TokenizedFile{source, tokens}

	// parser := &Parser{TokenizedFile: &tokenizedFile}
	// expr, _ := ParseExpression(parser, 0)
	// e, _ := json.Marshal(expr)
	// fmt.Println(string(e))
	// fmt.Println(tokens)
	tree, errs := Parse(&tokenizedFile)
	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Println(err)
			// err.Error()
		}
	}
	e, _ := json.Marshal(tree)
	fmt.Println(string(e))
}
