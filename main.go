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
		return Token{Flag: &TkEof}
	}
	return parser.TokenizedFile.tokens[index]
}

func (parser *Parser) CurrentToken() Token {
	return parser.CheckTokenAt(parser.Index)
}

func (parser *Parser) Peek() Token {
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
	for token.Flag == &TkNewLine {
		parser.Index++
		token = parser.CurrentToken()
	}
}

func (parser *Parser) WaitUntil(flag *string) bool {
	token := parser.CurrentToken()
	return token.Flag != flag && token.Flag != &TkEof
}

func (parser *Parser) Error(message string, errorToken Token) error {
	errorLine := fmt.Sprintf(
		"parser error: %v at %v:%d:%d",
		message, parser.TokenizedFile.file.filename, errorToken.Line, errorToken.Column,
	)
	// TODO: Add line context
	return fmt.Errorf(errorLine)
}

func (parser *Parser) Expect(flags ...*string) (token Token, err error) {
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

func (parser *Parser) ExpectConsumeWithWhitespace(flags ...*string) (token Token, err error) {
	token, err = parser.Expect(flags...)
	parser.Next()
	return
}

func (parser *Parser) ExpectConsume(flags ...*string) (token Token, err error) {
	token, err = parser.Expect(flags...)
	parser.NextWithoutWhitespace()
	return
}

func IsRHSOperator(operator *string) bool {
	switch operator {
	case &TkDot, &TkPlus, &TkMinus, &TkStar, &TkFowardSlash, &TkEqualsEquals, &TkBangEquals, &TkEqual:
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
	case &TkDot, &TkPlus, &TkMinus, &TkStar, &TkFowardSlash, &TkEqualsEquals, &TkBangEquals:
		return 1
	case &TkEqual:
		return 0
	default:
		return 1
	}
}

// https://en.cppreference.com/w/c/language/operator_precedence
// inverted here
func Precedence(operator *string) int {
	switch operator {
	case &TkEqual:
		return 1
	case &TkEqualsEquals, &TkBangEquals:
		return 3
	case &TkPlus, &TkMinus:
		return 4
	case &TkFowardSlash, &TkStar:
		return 7
	case &TkDot:
		return 14
	default:
		return 0
	}
}

func (parser *Parser) RHSExpression(leftExpr Expression, operation Token, nextPrec int) (Expression, error) {
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
	if leftToken.Flag == &TkNumber || leftToken.Flag == &TkString {
		leftExpr, err = parser.ParseLiteralExpression()
		if err != nil {
			return leftExpr, err
		}
		// Parenthesised expression
	} else if leftToken.Flag == &TkIdentifier {
		parser.Next()
		leftExpr = Expression{
			Literal:   leftToken.Value,
			Operation: "Variable",
			Position:  &leftToken,
		}
	} else if leftToken.Flag == &TkLeftParens {
		parser.Next()
		leftExpr, err = parser.ParseExpression(0)
		if err != nil {
			return leftExpr, err
		}
		_, err = parser.ExpectConsume(&TkRightParens)
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
		if token.Flag == &TkNewLine ||
			token.Flag == &TkEof ||
			token.Flag == &TkRightParens {
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
	token, err := parser.ExpectConsumeWithWhitespace(&TkNumber, &TkString)
	if err != nil {
		exprErr = err
		return
	}

	switch token.Flag {
	case &TkNumber:
		// TODO: Add floating point and hex
		number, err := strconv.ParseInt(token.Value, 10, 64)
		if err != nil {
			exprErr = err
			return
		}

		expr = Expression{Operation: "NumberLiteral", Literal: number, Position: &token}
	case &TkString:
		text := token.Value[1 : len(token.Value)-1]
		expr = Expression{Operation: "StringLiteral", Literal: text, Position: &token}
	}
	return
}

func ParseFunctionBody(parser *Parser, scope *Statement) error {
	token := parser.CurrentToken()
	switch token.Flag {
	case &TkKeywordIf:
		// TODO: parse if
		parser.NextWithoutWhitespace()
		for parser.WaitUntil(&TkKeywordEnd) {
			ParseFunctionBody(parser, scope)
		}
		parser.NextWithoutWhitespace()
	case &TkKeywordLoop:
		// TODO: parse loop
		parser.NextWithoutWhitespace()
		for parser.WaitUntil(&TkKeywordEnd) {
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
	_, err := parser.ExpectConsume(&TkKeywordFunction)
	if err != nil {
		errs = append(errs, err)
	}
	token, err := parser.ExpectConsume(&TkIdentifier)
	if err != nil {
		errs = append(errs, err)
	}
	function := &Statement{Flag: "Function", Value: token}
	ParseAttributesList(parser, function)
	for parser.WaitUntil(&TkKeywordEnd) {
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
	for parser.WaitUntil(&TkKeywordEnd) {
		errs := parser.ParseFunction(class)
		errors = append(errors, errs...)
	}
	parser.NextWithoutWhitespace()
	return
}

func ParseAttributesList(parser *Parser, class *Statement) (errors []error) {
	// attribute list is optional
	flag := parser.CurrentToken().Flag
	if flag != &TkIdentifier && flag != &TkLeftParens {
		return
	}

	seenParens := false
	if flag == &TkLeftParens {
		seenParens = true
		parser.NextWithoutWhitespace()
	}
	flag = parser.CurrentToken().Flag
	if flag == &TkRightParens {
		return
	}

	for {
		token, err := parser.ExpectConsume(&TkIdentifier)
		if err != nil {
			errors = append(errors, err)
		}
		attribute := Statement{Flag: "Attribute", Value: token}

		token = parser.CurrentToken()

		if token.Flag == &TkEqual {
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

		if token.Flag == &TkComma {
			parser.NextWithoutWhitespace()
		} else {
			// fmt.Println(token)
			break
		}
	}
	if seenParens {
		_, err := parser.ExpectConsume(&TkRightParens)
		if err != nil {
			errors = append(errors, err)
		}
	}
	return
}

func (parser *Parser) ParseClassInheritance(root *Statement) error {
	// inheritance is optional
	if parser.CurrentToken().Flag != &TkLessThan {
		return nil
	}
	parser.ExpectConsume(&TkLessThan)
	token, err := parser.ExpectConsume(&TkIdentifier)
	if err != nil {
		return err
	}

	inherits := Statement{Flag: "Inherits", Value: token}
	root.Statements = append(root.Statements, inherits)
	return nil
}

func (parser *Parser) ParseClass(root *Statement) (errors []error) {
	token, err := parser.ExpectConsume(&TkIdentifier)
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
	for parser.WaitUntil(&TkKeywordEnd) {
		errs := parser.ParseStatement(class)
		errors = append(errors, errs...)
	}
	parser.NextWithoutWhitespace()
	return
}

func (parser *Parser) ParseModule(root *Statement) (errors []error) {
	token, err := parser.ExpectConsume(&TkIdentifier)
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
	case &TkKeywordModule:
		classErrors := parser.ParseModule(root)
		errors = append(errors, classErrors...)
	case &TkKeywordClass:
		classErrors := parser.ParseClass(root)
		errors = append(errors, classErrors...)
	case &TkKeywordFunction:
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

func Parse(file *TokenizedFile) (root Statement, errors []error) {
	parser := &Parser{TokenizedFile: file}
	root = Statement{Flag: "Module", Value: Token{Flag: &TkIdentifier, Value: "Main"}}
	parser.SkipWhitespace()
	for parser.CurrentToken().Flag != &TkEof {
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
