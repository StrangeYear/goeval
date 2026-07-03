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

var compileParserPool = sync.Pool{
	New: func() any {
		return new(compileParserImpl)
	},
}

func compileParseWithPool(lex compileLexer) int {
	parser := compileParserPool.Get().(*compileParserImpl)
	defer func() {
		parser.lval = compileSymType{}
		parser.stack = [compileInitialStackSize]compileSymType{}
		parser.char = 0
		compileParserPool.Put(parser)
	}()
	return parser.Parse(lex)
}
