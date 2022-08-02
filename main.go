package main

import (
	"encoding/json"
	"fmt"
	"os"
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
