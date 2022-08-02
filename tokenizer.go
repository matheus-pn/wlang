package main

import "fmt"

// I'll use a proper enum someday (never)
// TODO: Change to interned strings
// I want to have a nice stringy name for debbuging
// while only having to compare a pointer

// Tokens
var TkEof = "Eof"
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
var TkColonEquals = "ColonEquals"
var TkColon = "Colon"

// States for the tokenizer
var StateInitial = "Initial"
var StateSeenBang = "SeenBang"
var StateSeenColon = "SeenColon"
var StateSeenEquals = "SeenEquals"
var StateSeenGreaterThan = "SeenGreaterThan"
var StateSeenLessThan = "SenLessThan"
var StateSeenFowardSlash = "SeenFowardSlash"
var StateSeenQuote = "SeenQuote"
var StateSeenIdentifier = "SeenIdentifier"
var StateSeenNumber = "SeenNumber"
var StateInsideInlineComment = "InsideInlineComment"

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
				tokenization.State = &StateSeenBang
			case ':':
				tokenization.State = &StateSeenColon
			case '=':
				tokenization.State = &StateSeenEquals
			case '>':
				tokenization.State = &StateSeenGreaterThan
			case '<':
				tokenization.State = &StateSeenLessThan
			case '/':
				tokenization.State = &StateSeenFowardSlash
			case '"':
				tokenization.State = &StateSeenQuote
				currentPhrase = append(currentPhrase, letter)
			default:
				if letter >= 'a' && letter <= 'z' || letter >= 'A' && letter <= 'Z' || letter == '_' {
					tokenization.State = &StateSeenIdentifier
					currentPhrase = append(currentPhrase, letter)
					break
				}

				if letter >= '0' && letter <= '9' {
					tokenization.State = &StateSeenNumber
					currentPhrase = append(currentPhrase, letter)
					break
				}

				errors = append(errors, tokenization.Error(fmt.Sprintf("Unepected Rune %U %q", letter, letter)))
			}

		case &StateSeenFowardSlash:
			switch letter {
			case '/':
				tokenization.State = &StateInsideInlineComment
				currentPhrase = append(currentPhrase, letter)
			default:
				tokens = append(tokens, tokenization.NewToken(&TkFowardSlash, ""))
				tokenization.State = &StateInitial
				goto retry
			}

		case &StateSeenColon:
			switch letter {
			case '=':
				tokenization.State = &StateInitial
				tokens = append(tokens, tokenization.NewToken(&TkColonEquals, ""))
			default:
				tokens = append(tokens, tokenization.NewToken(&TkColon, ""))
				tokenization.State = &StateInitial
				goto retry
			}

		case &StateSeenBang:
			switch letter {
			case '=':
				tokenization.State = &StateInitial
				tokens = append(tokens, tokenization.NewToken(&TkBangEquals, ""))
			default:
				tokens = append(tokens, tokenization.NewToken(&TkBang, ""))
				tokenization.State = &StateInitial
				goto retry
			}

		case &StateSeenEquals:
			switch letter {
			case '=':
				tokenization.State = &StateInitial
				tokens = append(tokens, tokenization.NewToken(&TkEqualsEquals, ""))
			default:
				tokens = append(tokens, tokenization.NewToken(&TkEqual, ""))
				tokenization.State = &StateInitial
				goto retry
			}

		case &StateSeenGreaterThan:
			switch letter {
			case '=':
				tokenization.State = &StateInitial
				tokens = append(tokens, tokenization.NewToken(&TkGreaterEquals, ""))
			default:
				tokens = append(tokens, tokenization.NewToken(&TkGreaterThan, ""))
				tokenization.State = &StateInitial
				goto retry
			}

		case &StateSeenLessThan:
			switch letter {
			case '=':
				tokenization.State = &StateInitial
				tokens = append(tokens, tokenization.NewToken(&TkLessEquals, ""))
			default:
				tokens = append(tokens, tokenization.NewToken(&TkLessThan, ""))
				tokenization.State = &StateInitial
				goto retry
			}

		case &StateSeenIdentifier:
			switch letter {
			default:
				if letter >= 'a' && letter <= 'z' || letter >= 'A' && letter <= 'Z' || letter == '_' || letter >= '0' && letter <= '9' {
					currentPhrase = append(currentPhrase, letter)
					break
				}

				tokens = append(tokens, tokenization.IdentifierOrKeyword(string(currentPhrase)))
				currentPhrase = nil
				tokenization.State = &StateInitial
				// Process the char that ended the identifier
				goto retry
			}

		case &StateSeenNumber:
			switch letter {
			default:
				if letter >= '0' && letter <= '9' {
					currentPhrase = append(currentPhrase, letter)
					break
				}

				tokens = append(tokens, tokenization.NewToken(&TkNumber, string(currentPhrase)))
				currentPhrase = nil
				tokenization.State = &StateInitial
				goto retry
			}

		case &StateSeenQuote:
			switch letter {
			case '"':
				currentPhrase = append(currentPhrase, letter)
				tokens = append(tokens, tokenization.NewToken(&TkString, string(currentPhrase)))
				currentPhrase = nil
				tokenization.State = &StateInitial

			default:
				currentPhrase = append(currentPhrase, letter)
			}

		case &StateInsideInlineComment:
			switch letter {
			case '\n':
				// tokens = append(tokens, tokenization.NewToken("Comment", string(currentPhrase)))
				// tokens = append(tokens, tokenization.NewToken("NewLine", ""))
				currentPhrase = nil
				tokenization.State = &StateInitial
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
