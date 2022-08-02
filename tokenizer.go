package main

import "fmt"

// I'll use a proper enum someday (never)
// Tokens
var TkNewLine = "NewLine"
var TkDot = "Dot"
var TkComma = "Dot"
var TkEqual = "Equal"
var TkFowardSlash = "FowardSlash"
var TkStar = "Star"
var TkPlus = "Plus"
var TkMinus = "Minus"
var TkEqualsEquals = "EqualsEquals"
var TkLessEquals = "LessEquals"
var TkGreaterEquals = "GreaterEquals"
var TkLessThan = "LessThan"
var TkGreaterThan = "GreaterThan"
var TkBang = "Bang"
var TkBangEquals = "BangEquals"
var TkKeywordIf = "KeywordIf"
var TkKeywordModule = "KeywordModule"
var TkKeywordClass = "KeywordClass"
var TkKeywordFunction = "KeywordFunction"
var TkKeywordEnd = "KeywordEnd"
var TkKeywordLoop = "KeywordLoop"
var TkIdentifier = "Identifier"
var TkString = "String"
var TkNumber = "Number"
var TkLeftSquareBracket = "LeftSquareBracket"
var TkRightSquareBracket = "RightSquareBracket"
var TkLeftParens = "LeftParens"
var TkRightParens = "RightParens"

tokenization.State = "SeenBang"
case ':':
	tokenization.State = "SeenColon"
case '=':
	tokenization.State = "SeenEquals"
case '>':
	tokenization.State = "SeenGreaterThan"
case '<':
	tokenization.State = "SeenLessThan"
case '/':
	tokenization.State = "SeenFowardSlash"
case '"':
	tokenization.State = "SeenQuote"
// States for the tokenizer
var StateInitial = "Initial"
var StateSeenBang = "SeenBang"
var StateSeenColon = "SeenColon"
var StateSeen

// TODO: Use proper enums
type Token struct {
	Flag   *string
	Value  string
	Line   int
	Column int
}

// TODO: Use proper enums
type Tokenization struct {
	File   *SourceFile
	State  *string
	Index  int
	Line   int
	Column int
}

type TokenizedFile struct {
	file   *SourceFile
	tokens []Token
}

func (tk *Tokenization) CheckCharAt(index int) rune {
	if index >= len(tk.File.Runes()) {
		return -1
	}
	return tk.File.Runes()[index]
}

func (tk *Tokenization) CurrentChar() rune {
	return tk.CheckCharAt(tk.Index)
}

func (tk *Tokenization) Next() {
	tk.Index++
}

func (tk *Tokenization) Error(message string) error {
	errorLine := fmt.Sprintf("tokenization error: %v at %v:%d:%d -- %+v", message, tk.File.filename, tk.Line, tk.Column, tk)
	// TODO: Add line context
	return fmt.Errorf(errorLine)
}

func (tk *Tokenization) NewToken(flag *string, value string) Token {
	return Token{flag, value, tk.Line, tk.Column - len(value)}
}

func (tk *Tokenization) IdentifierOrKeyword(identifier string) (token Token) {
	switch identifier {
	case "if":
		token = tk.NewToken(&TkKeywordIf, "")
	case "module":
		token = tk.NewToken(&TkKeywordModule, "")
	case "function":
		token = tk.NewToken(&TkKeywordFunction, "")
	case "class":
		token = tk.NewToken(&TkKeywordClass, "")
	case "end":
		token = tk.NewToken(&TkKeywordEnd, "")
	case "loop":
		token = tk.NewToken(&TkKeywordLoop, "")
	default:
		token = tk.NewToken(&TkIdentifier, identifier)
	}
	return
}

const MAX_TOKENIZER_ERROR = 10

func Tokenize(src *SourceFile) (tokens []Token, errors []error) {
	tokenization := Tokenization{File: src, State: &StateInitial, Line: 1, Column: 1}
	var currentPhrase []rune

	for ; ; tokenization.Next() {
		if len(errors) > MAX_TOKENIZER_ERROR {
			return
		}

		letter := tokenization.CurrentChar()

	retry:
		switch tokenization.State {
		case &StateInitial:
			switch letter {
			// ignore empty space and EOF
			case ' ', '\t', -1:
				break
			case '\n':
				tokens = append(tokens, tokenization.NewToken(&TkNewLine, ""))
			case '.':
				tokens = append(tokens, tokenization.NewToken(&TkDot, ""))
			case ',':
				tokens = append(tokens, tokenization.NewToken(&TkComma, ""))
			case '+':
				tokens = append(tokens, tokenization.NewToken(&TkPlus, ""))
			case '-':
				tokens = append(tokens, tokenization.NewToken(&TkMinus, ""))
			case '*':
				tokens = append(tokens, tokenization.NewToken(&TkStar, ""))
			case '[':
				tokens = append(tokens, tokenization.NewToken(&TkLeftSquareBracket, ""))
			case ']':
				tokens = append(tokens, tokenization.NewToken(&TkRightSquareBracket, ""))
			case '(':
				tokens = append(tokens, tokenization.NewToken(&TkLeftParens, ""))
			case ')':
				tokens = append(tokens, tokenization.NewToken(&TkRightParens, ""))
			case '!':
				tokenization.State = "SeenBang"
			case ':':
				tokenization.State = "SeenColon"
			case '=':
				tokenization.State = "SeenEquals"
			case '>':
				tokenization.State = "SeenGreaterThan"
			case '<':
				tokenization.State = "SeenLessThan"
			case '/':
				tokenization.State = "SeenFowardSlash"
			case '"':
				tokenization.State = "SeenQuote"
				currentPhrase = append(currentPhrase, letter)
			default:
				if letter >= 'a' && letter <= 'z' || letter >= 'A' && letter <= 'Z' || letter == '_' {
					tokenization.State = "SeenIdentifier"
					currentPhrase = append(currentPhrase, letter)
					break
				}

				if letter >= '0' && letter <= '9' {
					tokenization.State = "SeenNumber"
					currentPhrase = append(currentPhrase, letter)
					break
				}

				errors = append(errors, tokenization.Error(fmt.Sprintf("Unepected Rune %U %q", letter, letter)))
			}

		case "SeenFowardSlash":
			switch letter {
			case '/':
				tokenization.State = "InsideInlineComment"
				currentPhrase = append(currentPhrase, letter)
			default:
				tokens = append(tokens, tokenization.NewToken("FowardSlash", ""))
				tokenization.State = "Initial"
				goto retry
			}

		case "SeenColon":
			switch letter {
			case '=':
				tokenization.State = "Initial"
				tokens = append(tokens, tokenization.NewToken("ColonEquals", ""))
			default:
				tokens = append(tokens, tokenization.NewToken("Colon", ""))
				tokenization.State = "Initial"
				goto retry
			}

		case "SeenBang":
			switch letter {
			case '=':
				tokenization.State = "Initial"
				tokens = append(tokens, tokenization.NewToken("BangEquals", ""))
			default:
				tokens = append(tokens, tokenization.NewToken("Bang", ""))
				tokenization.State = "Initial"
				goto retry
			}

		case "SeenEquals":
			switch letter {
			case '=':
				tokenization.State = "Initial"
				tokens = append(tokens, tokenization.NewToken("EqualsEquals", ""))
			default:
				tokens = append(tokens, tokenization.NewToken("Equal", ""))
				tokenization.State = "Initial"
				goto retry
			}

		case "SeenGreaterThan":
			switch letter {
			case '=':
				tokenization.State = "Initial"
				tokens = append(tokens, tokenization.NewToken("GreaterEquals", ""))
			default:
				tokens = append(tokens, tokenization.NewToken("GreaterThan", ""))
				tokenization.State = "Initial"
				goto retry
			}

		case "SeenLessThan":
			switch letter {
			case '=':
				tokenization.State = "Initial"
				tokens = append(tokens, tokenization.NewToken("LessEquals", ""))
			default:
				tokens = append(tokens, tokenization.NewToken("LessThan", ""))
				tokenization.State = "Initial"
				goto retry
			}

		case "SeenIdentifier":
			switch letter {
			default:
				if letter >= 'a' && letter <= 'z' || letter >= 'A' && letter <= 'Z' || letter == '_' || letter >= '0' && letter <= '9' {
					currentPhrase = append(currentPhrase, letter)
					break
				}

				tokens = append(tokens, tokenization.IdentifierOrKeyword(string(currentPhrase)))
				currentPhrase = nil
				tokenization.State = "Initial"
				// Process the char that ended the identifier
				goto retry
			}

		case "SeenNumber":
			switch letter {
			default:
				if letter >= '0' && letter <= '9' {
					currentPhrase = append(currentPhrase, letter)
					break
				}

				tokens = append(tokens, tokenization.NewToken("Number", string(currentPhrase)))
				currentPhrase = nil
				tokenization.State = "Initial"
				goto retry
			}

		case "SeenQuote":
			switch letter {
			case '"':
				currentPhrase = append(currentPhrase, letter)
				tokens = append(tokens, tokenization.NewToken("String", string(currentPhrase)))
				currentPhrase = nil
				tokenization.State = "Initial"

			default:
				currentPhrase = append(currentPhrase, letter)
			}

		case "InsideInlineComment":
			switch letter {
			case '\n':
				// tokens = append(tokens, tokenization.NewToken("Comment", string(currentPhrase)))
				// tokens = append(tokens, tokenization.NewToken("NewLine", ""))
				currentPhrase = nil
				tokenization.State = "Initial"
			default:
				currentPhrase = append(currentPhrase, letter)
			}

		default:
			errors = append(errors, tokenization.Error("Invalid tokenization state"))
		}

		tokenization.Column++
		if letter == '\n' {
			tokenization.Line++
			tokenization.Column = 1
		}

		if tokenization.CurrentChar() == -1 {
			break
		}
	}
	return
}
