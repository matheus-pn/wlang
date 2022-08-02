package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/matheuziz/wlang/src/tokenizer"
)

type Value interface{}

type Expression struct {
	Literal   Value
	Operation string
	Operands  []Expression
	Position  *tokenizer.Token
}

type Statement struct {
	Value      tokenizer.Token
	Flag       string
	Expression *Expression
	Statements []Statement
}

type Parser struct {
	TokenizedFile *tokenizer.TokenizedFile
	Index         int
}

func (parser *Parser) CheckTokenAt(index int) tokenizer.Token {
	if index >= len(parser.TokenizedFile.Tokens) {
		return tokenizer.Token{Flag: &tokenizer.TkEof}
	}
	return parser.TokenizedFile.Tokens[index]
}

func (parser *Parser) CurrentToken() tokenizer.Token {
	return parser.CheckTokenAt(parser.Index)
}

func (parser *Parser) Peek() tokenizer.Token {
	return parser.CheckTokenAt(parser.Index + 1)
}

func Include(strs []*string, s *string) bool {
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
	for token.Flag == &tokenizer.TkNewLine {
		parser.Index++
		token = parser.CurrentToken()
	}
}

func (parser *Parser) WaitUntil(flag *string) bool {
	token := parser.CurrentToken()
	return token.Flag != flag && token.Flag != &tokenizer.TkEof
}

func (parser *Parser) Error(message string, errorToken tokenizer.Token) error {
	errorLine := fmt.Sprintf(
		"parser error: %v at %v:%d:%d",
		message, parser.TokenizedFile.File.Filename, errorToken.Line, errorToken.Column,
	)
	// TODO: Add line context
	return fmt.Errorf(errorLine)
}

func (parser *Parser) Expect(flags ...*string) (token tokenizer.Token, err error) {
	token = parser.CurrentToken()

	if !Include(flags, token.Flag) {
		strFlags := []string{}
		for _, f := range flags {
			strFlags = append(strFlags, *f)
		}
		err = parser.Error(
			fmt.Sprintf("expected one of %v got %v", strings.Join(strFlags, ", "), token.Flag),
			token,
		)
	}
	return
}

func (parser *Parser) ExpectConsumeWithWhitespace(flags ...*string) (token tokenizer.Token, err error) {
	token, err = parser.Expect(flags...)
	parser.Next()
	return
}

func (parser *Parser) ExpectConsume(flags ...*string) (token tokenizer.Token, err error) {
	token, err = parser.Expect(flags...)
	parser.NextWithoutWhitespace()
	return
}

func IsRHSOperator(operator *string) bool {
	switch operator {
	case &tokenizer.TkDot, &tokenizer.TkPlus, &tokenizer.TkMinus, &tokenizer.TkStar, &tokenizer.TkFowardSlash, &tokenizer.TkEqualsEquals, &tokenizer.TkBangEquals, &tokenizer.TkEqual:
		return true
	default:
		return false
	}
}

// https://en.cppreference.com/w/c/language/operator_precedence
func Assoc(operator *string) int {
	// 1 => left-to-right
	// 0 => right-to-left
	switch operator {
	case &tokenizer.TkDot, &tokenizer.TkPlus, &tokenizer.TkMinus, &tokenizer.TkStar, &tokenizer.TkFowardSlash, &tokenizer.TkEqualsEquals, &tokenizer.TkBangEquals:
		return 1
	case &tokenizer.TkEqual:
		return 0
	default:
		return 1
	}
}

// https://en.cppreference.com/w/c/language/operator_precedence
// inverted here
func Precedence(operator *string) int {
	switch operator {
	case &tokenizer.TkEqual:
		return 1
	case &tokenizer.TkEqualsEquals, &tokenizer.TkBangEquals:
		return 3
	case &tokenizer.TkPlus, &tokenizer.TkMinus:
		return 4
	case &tokenizer.TkFowardSlash, &tokenizer.TkStar:
		return 7
	case &tokenizer.TkDot:
		return 14
	default:
		return 0
	}
}

func (parser *Parser) RHSExpression(leftExpr Expression, operation tokenizer.Token, nextPrec int) (Expression, error) {
	parser.Next()
	rightExpr, err := parser.ParseExpression(nextPrec)
	if err != nil {
		return rightExpr, err
	}
	leftExpr = Expression{
		Operation: *operation.Flag,
		Operands:  []Expression{leftExpr, rightExpr},
		Position:  &operation,
	}
	return leftExpr, nil
}

func (parser *Parser) ParseExpression(minPrec int) (Expression, error) {
	var leftExpr Expression
	var err error

	leftToken := parser.CurrentToken()
	// Literal expression
	if leftToken.Flag == &tokenizer.TkNumber || leftToken.Flag == &tokenizer.TkString {
		leftExpr, err = parser.ParseLiteralExpression()
		if err != nil {
			return leftExpr, err
		}
		// Parenthesised expression
	} else if leftToken.Flag == &tokenizer.TkIdentifier {
		parser.Next()
		leftExpr = Expression{
			Literal:   leftToken.Value,
			Operation: "Variable",
			Position:  &leftToken,
		}
	} else if leftToken.Flag == &tokenizer.TkLeftParens {
		parser.Next()
		leftExpr, err = parser.ParseExpression(0)
		if err != nil {
			return leftExpr, err
		}
		_, err = parser.ExpectConsume(&tokenizer.TkRightParens)
		if err != nil {
			return leftExpr, err
		}
	} else {
		return leftExpr, parser.Error(
			"Unexpected token: "+*leftToken.Flag+" on lhs of expression", leftToken,
		)
	}

	for {
		token := parser.CurrentToken()
		// Binary expression
		if token.Flag == &tokenizer.TkNewLine ||
			token.Flag == &tokenizer.TkEof ||
			token.Flag == &tokenizer.TkRightParens {
			return leftExpr, nil
		}

		prec := Precedence(token.Flag)
		if minPrec > prec {
			break
		}

		nextMinPrec := prec + Assoc(token.Flag)
		if IsRHSOperator(token.Flag) {
			leftExpr, err = parser.RHSExpression(leftExpr, token, nextMinPrec)
			if err != nil {
				return leftExpr, err
			}
		} else {
			return leftExpr, parser.Error(
				"Unexpected token: "+*token.Flag+" on rhs of expression", token,
			)
		}
	}
	return leftExpr, nil
}

func (parser *Parser) ParseLiteralExpression() (expr Expression, exprErr error) {
	token, err := parser.ExpectConsumeWithWhitespace(&tokenizer.TkNumber, &tokenizer.TkString)
	if err != nil {
		exprErr = err
		return
	}

	switch token.Flag {
	case &tokenizer.TkNumber:
		// TODO: Add floating point and hex
		number, err := strconv.ParseInt(token.Value, 10, 64)
		if err != nil {
			exprErr = err
			return
		}

		expr = Expression{Operation: "NumberLiteral", Literal: number, Position: &token}
	case &tokenizer.TkString:
		text := token.Value[1 : len(token.Value)-1]
		expr = Expression{Operation: "StringLiteral", Literal: text, Position: &token}
	}
	return
}

func ParseFunctionBody(parser *Parser, scope *Statement) error {
	token := parser.CurrentToken()
	switch token.Flag {
	case &tokenizer.TkKeywordIf:
		// TODO: parse if
		parser.NextWithoutWhitespace()
		for parser.WaitUntil(&tokenizer.TkKeywordEnd) {
			ParseFunctionBody(parser, scope)
		}
		parser.NextWithoutWhitespace()
	case &tokenizer.TkKeywordLoop:
		// TODO: parse loop
		parser.NextWithoutWhitespace()
		for parser.WaitUntil(&tokenizer.TkKeywordEnd) {
			ParseFunctionBody(parser, scope)
		}
		parser.NextWithoutWhitespace()
	default:
		expr, err := parser.ParseExpression(0)
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

func (parser *Parser) ParseFunction(scope *Statement) (errs []error) {
	_, err := parser.ExpectConsume(&tokenizer.TkKeywordFunction)
	if err != nil {
		errs = append(errs, err)
	}
	token, err := parser.ExpectConsume(&tokenizer.TkIdentifier)
	if err != nil {
		errs = append(errs, err)
	}
	function := &Statement{Flag: "Function", Value: token}
	ParseAttributesList(parser, function)
	for parser.WaitUntil(&tokenizer.TkKeywordEnd) {
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
	for parser.WaitUntil(&tokenizer.TkKeywordEnd) {
		errs := parser.ParseFunction(class)
		errors = append(errors, errs...)
	}
	parser.NextWithoutWhitespace()
	return
}

func ParseAttributesList(parser *Parser, class *Statement) (errors []error) {
	// attribute list is optional
	flag := parser.CurrentToken().Flag
	if flag != &tokenizer.TkIdentifier && flag != &tokenizer.TkLeftParens {
		return
	}

	seenParens := false
	if flag == &tokenizer.TkLeftParens {
		seenParens = true
		parser.NextWithoutWhitespace()
	}
	flag = parser.CurrentToken().Flag
	if flag == &tokenizer.TkRightParens {
		return
	}

	for {
		token, err := parser.ExpectConsume(&tokenizer.TkIdentifier)
		if err != nil {
			errors = append(errors, err)
		}
		attribute := Statement{Flag: "Attribute", Value: token}

		token = parser.CurrentToken()

		if token.Flag == &tokenizer.TkEqual {
			parser.NextWithoutWhitespace()
			// TODO: Add parse expression to default value of attribute
			expr, err := parser.ParseLiteralExpression()
			attribute.Expression = &expr
			token = parser.CurrentToken()
			if err != nil {
				errors = append(errors, err)
			}
		}

		class.Statements = append(class.Statements, attribute)

		if token.Flag == &tokenizer.TkComma {
			parser.NextWithoutWhitespace()
		} else {
			// fmt.Println(token)
			break
		}
	}
	if seenParens {
		_, err := parser.ExpectConsume(&tokenizer.TkRightParens)
		if err != nil {
			errors = append(errors, err)
		}
	}
	return
}

func (parser *Parser) ParseClassInheritance(root *Statement) error {
	// inheritance is optional
	if parser.CurrentToken().Flag != &tokenizer.TkLessThan {
		return nil
	}
	parser.ExpectConsume(&tokenizer.TkLessThan)
	token, err := parser.ExpectConsume(&tokenizer.TkIdentifier)
	if err != nil {
		return err
	}

	inherits := Statement{Flag: "Inherits", Value: token}
	root.Statements = append(root.Statements, inherits)
	return nil
}

func (parser *Parser) ParseClass(root *Statement) (errors []error) {
	token, err := parser.ExpectConsume(&tokenizer.TkIdentifier)
	class := Statement{Flag: "Class", Value: token}
	if err != nil {
		errors = append(errors, err)
	}
	err = parser.ParseClassInheritance(&class)
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

func (parser *Parser) ParseModuleBody(class *Statement) (errors []error) {
	for parser.WaitUntil(&tokenizer.TkKeywordEnd) {
		errs := parser.ParseStatement(class)
		errors = append(errors, errs...)
	}
	parser.NextWithoutWhitespace()
	return
}

func (parser *Parser) ParseModule(root *Statement) (errors []error) {
	token, err := parser.ExpectConsume(&tokenizer.TkIdentifier)
	module := Statement{Flag: "Module", Value: token}
	if err != nil {
		errors = append(errors, err)
	}
	errs := parser.ParseModuleBody(&module)
	errors = append(errors, errs...)
	root.Statements = append(root.Statements, module)
	return
}

func (parser *Parser) ParseStatement(root *Statement) (errors []error) {
	token := parser.CurrentToken()

	switch token.Flag {
	case &tokenizer.TkKeywordModule:
		classErrors := parser.ParseModule(root)
		errors = append(errors, classErrors...)
	case &tokenizer.TkKeywordClass:
		classErrors := parser.ParseClass(root)
		errors = append(errors, classErrors...)
	case &tokenizer.TkKeywordFunction:
		funcErrors := parser.ParseFunction(root)
		errors = append(errors, funcErrors...)
	default:
		statement := Statement{Flag: "error-statement", Value: token}
		errors = append(errors, parser.Error("expected statement, found "+*token.Flag, token))
		root.Statements = append(root.Statements, statement)
	}
	return
}

const MAX_PARSER_ERROR = 5

func Parse(file *tokenizer.TokenizedFile) (root Statement, errors []error) {
	parser := &Parser{TokenizedFile: file}
	root = Statement{Flag: "Module", Value: tokenizer.Token{Flag: &tokenizer.TkIdentifier, Value: "Main"}}
	parser.SkipWhitespace()
	for parser.CurrentToken().Flag != &tokenizer.TkEof {
		if len(errors) > MAX_PARSER_ERROR {
			break
		}

		declErrors := parser.ParseStatement(&root)
		errors = append(errors, declErrors...)
	}
	if len(errors) > MAX_PARSER_ERROR {
		errors = errors[:MAX_PARSER_ERROR]
	}
	return
}
