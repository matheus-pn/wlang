package tokenizer

import (
	"testing"

	"github.com/matheuziz/wlang/src/sourcefile"
)

func TestTokenizeOperators(t *testing.T) {
	operators := ". = / * + - == <= >= < > ! !="
	source := sourcefile.SourceFile{Filename: "test", ByteSource: []byte(operators)}
	tokens, errs := Tokenize(&source)
	if len(errs) > 0 {
		t.Errorf("Tokenization success expected")
	}
	expectedTokenFlags := []*string{
		&TkDot, &TkEqual, &TkFowardSlash, &TkStar, &TkPlus,
		&TkMinus, &TkEqualsEquals, &TkLessEquals, &TkGreaterEquals,
		&TkLessThan, &TkGreaterThan, &TkBang, &TkBangEquals,
	}
	if len(tokens) != len(expectedTokenFlags) {
		t.Errorf("Expected %v tokens, got %v", len(expectedTokenFlags), len(tokens))
	}
	for i, flag := range expectedTokenFlags {
		t.Run(*flag, func(t2 *testing.T) {
			if tokens[i].Flag != flag {
				t2.Errorf("Expected Token %v to be %v", tokens[i].Flag, flag)
			}
		})
	}
}

func TestTokenizeKeywords(t *testing.T) {
	operators := "if module class end loop"
	source := sourcefile.SourceFile{Filename: "test", ByteSource: []byte(operators)}

	tokens, errs := Tokenize(&source)
	if len(errs) > 0 {
		t.Errorf("Tokenization success expected")
	}
	expectedTokenFlags := []*string{
		&TkKeywordIf, &TkKeywordModule, &TkKeywordClass, &TkKeywordEnd, &TkKeywordLoop,
	}
	if len(tokens) != len(expectedTokenFlags) {
		t.Errorf("Expected %v tokens, got %v", len(expectedTokenFlags), len(tokens))
	}
	for i, flag := range expectedTokenFlags {
		t.Run(*flag, func(t2 *testing.T) {
			if tokens[i].Flag != flag {
				t2.Errorf("Expected Token %v to be %v", tokens[i].Flag, flag)
			}
		})
	}
}

func TestTokenizeVariableSizeTokens(t *testing.T) {
	operators := "ideNtifier \"Thïs ìs á string\" 1337"
	source := sourcefile.SourceFile{Filename: "test", ByteSource: []byte(operators)}

	tokens, errs := Tokenize(&source)
	if len(errs) > 0 {
		t.Errorf("Tokenization success expected")
	}
	expectedTokenFlags := []*string{
		&TkIdentifier, &TkString, &TkNumber,
	}
	if len(tokens) != len(expectedTokenFlags) {
		t.Errorf("Expected %v tokens, got %v", len(expectedTokenFlags), len(tokens))
	}
	for i, flag := range expectedTokenFlags {
		t.Run(*flag, func(t2 *testing.T) {
			if tokens[i].Flag != flag {
				t2.Errorf("Expected Token %v to be %v", tokens[i].Flag, flag)
			}
		})
	}
}

func TestTokenizeMisc(t *testing.T) {
	operators := ", [ ( ] ) \n  \n// *&qwneqwióó \n"
	source := sourcefile.SourceFile{Filename: "test", ByteSource: []byte(operators)}

	tokens, errs := Tokenize(&source)
	if len(errs) > 0 {
		t.Errorf("Tokenization success expected")
	}
	expectedTokenFlags := []*string{
		&TkComma, &TkLeftSquareBracket, &TkLeftParens,
		&TkRightSquareBracket, &TkRightParens, &TkNewLine, &TkNewLine,
	}
	if len(tokens) != len(expectedTokenFlags) {
		t.Errorf("Expected %v tokens, got %v", len(expectedTokenFlags), len(tokens))
	}
	for i, flag := range expectedTokenFlags {
		t.Run(*flag, func(t2 *testing.T) {
			if tokens[i].Flag != flag {
				t2.Errorf("Expected Token %v to be %v", tokens[i].Flag, flag)
			}
		})
	}
}
