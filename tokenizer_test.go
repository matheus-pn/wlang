package main

import (
	"testing"
)

func TestTokenizeOperators(t *testing.T) {
	operators := ". = / * + - == <= >= < > ! !="
	source := SourceFile{filename: "test", byteSource: []byte(operators)}
	tokens, errs := Tokenize(&source)
	if len(errs) > 0 {
		t.Errorf("Tokenization success expected")
	}
	expectedTokenFlags := []string{
		"Dot", "Equal", "FowardSlash", "Star", "Plus",
		"Minus", "EqualsEquals", "LessEquals", "GreaterEquals",
		"LessThan", "GreaterThan", "Bang", "BangEquals",
	}
	if len(tokens) != len(expectedTokenFlags) {
		t.Errorf("Expected %v tokens, got %v", len(expectedTokenFlags), len(tokens))
	}
	for i, flag := range expectedTokenFlags {
		t.Run(flag, func(t2 *testing.T) {
			if tokens[i].Flag != flag {
				t2.Errorf("Expected Token %v to be %v", tokens[i].Flag, flag)
			}
		})
	}
}

func TestTokenizeKeywords(t *testing.T) {
	operators := "if module class end loop"
	source := SourceFile{filename: "test", byteSource: []byte(operators)}
	tokens, errs := Tokenize(&source)
	if len(errs) > 0 {
		t.Errorf("Tokenization success expected")
	}
	expectedTokenFlags := []string{
		"KeywordIf", "KeywordModule", "KeywordClass", "KeywordEnd", "KeywordLoop",
	}
	if len(tokens) != len(expectedTokenFlags) {
		t.Errorf("Expected %v tokens, got %v", len(expectedTokenFlags), len(tokens))
	}
	for i, flag := range expectedTokenFlags {
		t.Run(flag, func(t2 *testing.T) {
			if tokens[i].Flag != flag {
				t2.Errorf("Expected Token %v to be %v", tokens[i].Flag, flag)
			}
		})
	}
}

func TestTokenizeVariableSizeTokens(t *testing.T) {
	operators := "ideNtifier \"Thïs ìs á string\" 1337"
	source := SourceFile{filename: "test", byteSource: []byte(operators)}
	tokens, errs := Tokenize(&source)
	if len(errs) > 0 {
		t.Errorf("Tokenization success expected")
	}
	expectedTokenFlags := []string{
		"Identifier", "String", "Number",
	}
	if len(tokens) != len(expectedTokenFlags) {
		t.Errorf("Expected %v tokens, got %v", len(expectedTokenFlags), len(tokens))
	}
	for i, flag := range expectedTokenFlags {
		t.Run(flag, func(t2 *testing.T) {
			if tokens[i].Flag != flag {
				t2.Errorf("Expected Token %v to be %v", tokens[i].Flag, flag)
			}
		})
	}
}

func TestTokenizeMisc(t *testing.T) {
	operators := ", [ ( ] ) \n  \n// *&qwneqwióó \n"
	source := SourceFile{filename: "test", byteSource: []byte(operators)}
	tokens, errs := Tokenize(&source)
	if len(errs) > 0 {
		t.Errorf("Tokenization success expected")
	}
	expectedTokenFlags := []string{
		"Comma", "LeftSquareBracket", "LeftParens",
		"RightSquareBracket", "RightParens", "NewLine", "NewLine",
	}
	if len(tokens) != len(expectedTokenFlags) {
		t.Errorf("Expected %v tokens, got %v", len(expectedTokenFlags), len(tokens))
	}
	for i, flag := range expectedTokenFlags {
		t.Run(flag, func(t2 *testing.T) {
			if tokens[i].Flag != flag {
				t2.Errorf("Expected Token %v to be %v", tokens[i].Flag, flag)
			}
		})
	}
}
