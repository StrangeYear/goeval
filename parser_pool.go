package goeval

import "sync"

var parserPool = sync.Pool{
	New: func() any {
		return new(yyParserImpl)
	},
}

func parseWithPool(lex yyLexer) int {
	parser := parserPool.Get().(*yyParserImpl)
	defer func() {
		parser.lval = yySymType{}
		parser.stack = [yyInitialStackSize]yySymType{}
		parser.char = 0
		parserPool.Put(parser)
	}()
	return parser.Parse(lex)
}
