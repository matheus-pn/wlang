package main

import (
	"encoding/json"
	"fmt"

	"github.com/matheuziz/wlang/src/sourcefile"
	"github.com/matheuziz/wlang/src/tokenizer"
)

func main() {
	source, err := sourcefile.OpenSource("test-assets/expr.w")
	if err != nil {
		return
	}

	tokens, errs := tokenizer.Tokenize(source)
	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Println(err)
		}
		return
	}

	tokenizedFile := tokenizer.TokenizedFile{File: source, Tokens: tokens}

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
