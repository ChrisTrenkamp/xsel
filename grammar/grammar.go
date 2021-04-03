package grammar

import (
	"fmt"

	"github.com/ChrisTrenkamp/xsel/grammar/lexer"
	"github.com/ChrisTrenkamp/xsel/grammar/parser"
	"github.com/ChrisTrenkamp/xsel/grammar/parser/bsr"
)

type Grammar struct {
	BSR *bsr.BSR
	lex *lexer.Lexer
}

func (g *Grammar) Next(bsr *bsr.BSR) *Grammar {
	return &Grammar{
		BSR: bsr,
		lex: g.lex,
	}
}

func (g *Grammar) GetString() string {
	return g.lex.GetString(g.BSR.LeftExtent(), g.BSR.RightExtent()-1)
}

// Creates an XPath query.
func Build(xpath string) (Grammar, error) {
	lex := lexer.New([]rune(xpath))
	parse, err := parser.Parse(lex)

	if err != nil {
		errStr := ""

		for _, e := range err {
			errStr += e.String()
		}

		return Grammar{}, fmt.Errorf(errStr)
	}

	roots := parse.GetRoots()

	if len(roots) == 0 {
		return Grammar{}, fmt.Errorf("could not build expression tree")
	}

	return Grammar{&roots[0], lex}, nil
}
