// Code generated by goyacc -o parser.go parser.y. DO NOT EDIT.

//line parser.y:2
package goeval

import __yyfmt__ "fmt"

//line parser.y:2

//line parser.y:6
type yySymType struct {
	yys  int
	name string
	val  Value
	vals []any
}

const VALUE = 57346
const EQ = 57347
const NEQ = 57348
const GTE = 57349
const LTE = 57350
const RE = 57351
const NRE = 57352
const AND = 57353
const OR = 57354
const NC = 57355
const IN = 57356
const IDENTIFIER = 57357

var yyToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"VALUE",
	"EQ",
	"NEQ",
	"GTE",
	"LTE",
	"RE",
	"NRE",
	"AND",
	"OR",
	"NC",
	"IN",
	"IDENTIFIER",
	"'+'",
	"'-'",
	"'*'",
	"'/'",
	"'%'",
	"'<'",
	"'>'",
	"'?'",
	"':'",
	"'='",
	"'('",
	"')'",
	"','",
	"'['",
	"']'",
	"'!'",
}

var yyStatenames = [...]string{}

const yyEofCode = 1
const yyErrCode = 2
const yyInitialStackSize = 16

//line yacctab:1
var yyExca = [...]int8{
	-1, 1,
	1, -1,
	-2, 0,
}

const yyPrivate = 57344

const yyLast = 110

var yyAct = [...]int8{
	2, 4, 3, 74, 62, 32, 61, 72, 31, 33,
	75, 62, 41, 42, 43, 44, 73, 62, 45, 46,
	47, 48, 49, 50, 51, 52, 53, 54, 55, 56,
	57, 58, 59, 60, 6, 34, 33, 14, 64, 33,
	63, 66, 33, 65, 6, 8, 67, 9, 26, 27,
	28, 29, 30, 23, 24, 8, 11, 25, 10, 7,
	1, 5, 71, 62, 70, 0, 11, 38, 0, 7,
	76, 15, 16, 17, 18, 19, 20, 0, 40, 21,
	22, 39, 26, 27, 28, 29, 30, 23, 24, 35,
	0, 25, 12, 13, 12, 13, 0, 12, 13, 0,
	37, 0, 0, 36, 14, 0, 14, 69, 68, 14,
}

var yyPact = [...]int16{
	30, -1000, 86, 66, -1000, 30, -1000, 40, 9, 74,
	52, 30, 30, 30, 30, 30, 30, 30, 30, 30,
	30, 30, 30, 30, 30, 30, 30, 30, 30, 30,
	30, -1000, -24, -1000, 40, -1000, 40, 40, -1000, 40,
	40, 81, 14, 14, 83, 32, 32, 32, 32, 32,
	32, 32, 32, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-1000, -1000, 40, 35, -23, -11, -27, -17, -1000, 30,
	-1000, -1000, -1000, -1000, -1000, -1000, -1000,
}

var yyPgo = [...]int8{
	0, 60, 0, 1, 2, 58, 47, 5,
}

var yyR1 = [...]int8{
	0, 1, 3, 3, 3, 3, 3, 6, 6, 6,
	6, 5, 5, 5, 5, 7, 7, 7, 4, 4,
	4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	4, 4, 4, 4, 4, 4, 2, 2, 2, 2,
}

var yyR2 = [...]int8{
	0, 1, 1, 3, 4, 1, 1, 3, 2, 4,
	4, 1, 2, 4, 4, 0, 1, 3, 1, 2,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 1, 3, 3, 5,
}

var yyChk = [...]int16{
	-1000, -1, -2, -4, -3, 31, 4, 29, 15, -6,
	-5, 26, 11, 12, 23, 5, 6, 7, 8, 9,
	10, 13, 14, 21, 22, 25, 16, 17, 18, 19,
	20, -4, -7, -3, 26, 15, 29, 26, 15, 29,
	26, -2, -2, -2, -2, -4, -4, -4, -4, -4,
	-4, -4, -4, -4, -4, -4, -4, -4, -4, -4,
	-4, 30, 28, -7, -3, -7, -3, -7, 27, 24,
	-3, 27, 30, 27, 30, 27, -2,
}

var yyDef = [...]int8{
	0, -2, 1, 36, 18, 0, 2, 15, 11, 5,
	6, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 19, 0, 16, 15, 8, 0, 15, 12, 0,
	15, 0, 37, 38, 0, 20, 21, 22, 23, 24,
	25, 26, 27, 28, 29, 30, 31, 32, 33, 34,
	35, 3, 0, 0, 0, 0, 0, 0, 7, 0,
	17, 4, 9, 10, 13, 14, 39,
}

var yyTok1 = [...]int8{
	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 31, 3, 3, 3, 20, 3, 3,
	26, 27, 18, 16, 28, 17, 3, 19, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 24, 3,
	21, 25, 22, 23, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 29, 3, 30,
}

var yyTok2 = [...]int8{
	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15,
}

var yyTok3 = [...]int8{
	0,
}

var yyErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

//line yaccpar:1

/*	parser for yacc output	*/

var (
	yyDebug        = 0
	yyErrorVerbose = false
)

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

type yyParser interface {
	Parse(yyLexer) int
	Lookahead() int
}

type yyParserImpl struct {
	lval  yySymType
	stack [yyInitialStackSize]yySymType
	char  int
}

func (p *yyParserImpl) Lookahead() int {
	return p.char
}

func yyNewParser() yyParser {
	return &yyParserImpl{}
}

const yyFlag = -1000

func yyTokname(c int) string {
	if c >= 1 && c-1 < len(yyToknames) {
		if yyToknames[c-1] != "" {
			return yyToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func yyStatname(s int) string {
	if s >= 0 && s < len(yyStatenames) {
		if yyStatenames[s] != "" {
			return yyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func yyErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !yyErrorVerbose {
		return "syntax error"
	}

	for _, e := range yyErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + yyTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := int(yyPact[state])
	for tok := TOKSTART; tok-1 < len(yyToknames); tok++ {
		if n := base + tok; n >= 0 && n < yyLast && int(yyChk[int(yyAct[n])]) == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if yyDef[state] == -2 {
		i := 0
		for yyExca[i] != -1 || int(yyExca[i+1]) != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; yyExca[i] >= 0; i += 2 {
			tok := int(yyExca[i])
			if tok < TOKSTART || yyExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if yyExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += yyTokname(tok)
	}
	return res
}

func yylex1(lex yyLexer, lval *yySymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = int(yyTok1[0])
		goto out
	}
	if char < len(yyTok1) {
		token = int(yyTok1[char])
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			token = int(yyTok2[char-yyPrivate])
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		token = int(yyTok3[i+0])
		if token == char {
			token = int(yyTok3[i+1])
			goto out
		}
	}

out:
	if token == 0 {
		token = int(yyTok2[1]) /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", yyTokname(token), uint(char))
	}
	return char, token
}

func yyParse(yylex yyLexer) int {
	return yyNewParser().Parse(yylex)
}

func (yyrcvr *yyParserImpl) Parse(yylex yyLexer) int {
	var yyn int
	var yyVAL yySymType
	var yyDollar []yySymType
	_ = yyDollar // silence set and not used
	yyS := yyrcvr.stack[:]

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yyrcvr.char = -1
	yytoken := -1 // yyrcvr.char translated into internal numbering
	defer func() {
		// Make sure we report no lookahead when not parsing.
		yystate = -1
		yyrcvr.char = -1
		yytoken = -1
	}()
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yytoken), yyStatname(yystate))
	}

	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	yyn = int(yyPact[yystate])
	if yyn <= yyFlag {
		goto yydefault /* simple state */
	}
	if yyrcvr.char < 0 {
		yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
	}
	yyn += yytoken
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = int(yyAct[yyn])
	if int(yyChk[yyn]) == yytoken { /* valid shift */
		yyrcvr.char = -1
		yytoken = -1
		yyVAL = yyrcvr.lval
		yystate = yyn
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	}

yydefault:
	/* default state action */
	yyn = int(yyDef[yystate])
	if yyn == -2 {
		if yyrcvr.char < 0 {
			yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && int(yyExca[xi+1]) == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = int(yyExca[xi+0])
			if yyn < 0 || yyn == yytoken {
				break
			}
		}
		yyn = int(yyExca[xi+1])
		if yyn < 0 {
			goto ret0
		}
	}
	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			yylex.Error(yyErrorMessage(yystate, yytoken))
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf(" saw %s\n", yyTokname(yytoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				yyn = int(yyPact[yyS[yyp].yys]) + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = int(yyAct[yyn]) /* simulate a shift of "error" */
					if int(yyChk[yystate]) == yyErrCode {
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yytoken))
			}
			if yytoken == yyEofCode {
				goto ret1
			}
			yyrcvr.char = -1
			yytoken = -1
			goto yynewstate /* try again in the same state */
		}
	}

	/* reduction by production yyn */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", yyn, yyStatname(yystate))
	}

	yynt := yyn
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= int(yyR2[yyn])
	// yyp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if yyp+1 >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = int(yyR1[yyn])
	yyg := int(yyPgo[yyn])
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = int(yyAct[yyg])
	} else {
		yystate = int(yyAct[yyj])
		if int(yyChk[yystate]) != -yyn {
			yystate = int(yyAct[yyg])
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 1:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:24
		{
			yylex.(*lexer).answer = yyDollar[1].val
			yyVAL.val = yyDollar[1].val
			return 0
		}
	case 2:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:32
		{
			yyVAL.val = yyDollar[1].val
		}
	case 3:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:36
		{
			yyVAL.val = NewValue("", yyDollar[2].vals)
		}
	case 4:
		yyDollar = yyS[yypt-4 : yypt+1]
//line parser.y:40
		{
			fn := yylex.(*lexer).fns[yyDollar[1].name]
			if fn == nil {
				panic(__yyfmt__.Errorf("unknown function %s", yyDollar[1].name))
			}
			res, err := fn(yyDollar[3].vals...)
			if err != nil {
				panic(err)
			}
			yyVAL.val = NewValue("", res)
		}
	case 7:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:57
		{
			yyVAL.val = yyDollar[2].val
		}
	case 8:
		yyDollar = yyS[yypt-2 : yypt+1]
//line parser.y:61
		{
			val := SelectValue(yyDollar[1].val.val, yyDollar[2].name)
			name := yyDollar[1].val.name + "." + yyDollar[2].name
			yyVAL.val = NewValue(name, val)
		}
	case 9:
		yyDollar = yyS[yypt-4 : yypt+1]
//line parser.y:67
		{
			val := SelectValue(yyDollar[1].val.val, yyDollar[3].val.String())
			name := __yyfmt__.Sprintf("%s[%#v]", yyDollar[1].val.name, yyDollar[3].val.val)
			yyVAL.val = NewValue(name, val)
		}
	case 10:
		yyDollar = yyS[yypt-4 : yypt+1]
//line parser.y:73
		{
			funcValue := toFunc(yyDollar[1].val.val)
			val, err := funcValue(yyDollar[3].vals...)
			if err != nil {
				panic(err)
			}
			name := yyDollar[1].val.name + "("
			for _, v := range yyDollar[3].vals {
				name += __yyfmt__.Sprintf("%#v, ", v)
			}
			if name[len(name)-2] == ',' {
				name = name[:len(name)-2]
			}
			name += ")"
			yyVAL.val = NewValue(name, val)
		}
	case 11:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:92
		{
			val := yylex.(*lexer).kv[yyDollar[1].name]
			yyVAL.val = NewValue(yyDollar[1].name, val)
		}
	case 12:
		yyDollar = yyS[yypt-2 : yypt+1]
//line parser.y:97
		{
			val := SelectValue(yyDollar[1].val.val, yyDollar[2].name)
			name := yyDollar[1].val.name + "." + yyDollar[2].name
			yyVAL.val = NewValue(name, val)
		}
	case 13:
		yyDollar = yyS[yypt-4 : yypt+1]
//line parser.y:103
		{
			val := SelectValue(yyDollar[1].val.val, yyDollar[3].val.String())
			name := __yyfmt__.Sprintf("%s[%#v]", yyDollar[1].val.name, yyDollar[3].val.val)
			yyVAL.val = NewValue(name, val)
		}
	case 14:
		yyDollar = yyS[yypt-4 : yypt+1]
//line parser.y:109
		{
			funcValue := toFunc(yyDollar[1].val.val)
			val, err := funcValue(yyDollar[3].vals...)
			if err != nil {
				panic(err)
			}
			name := yyDollar[1].val.name + "("
			for _, v := range yyDollar[3].vals {
				name += __yyfmt__.Sprintf("%#v, ", v)
			}
			if name[len(name)-2] == ',' {
				name = name[:len(name)-2]
			}
			name += ")"
			yyVAL.val = NewValue(name, val)
		}
	case 15:
		yyDollar = yyS[yypt-0 : yypt+1]
//line parser.y:127
		{
			yyVAL.vals = []any{}
		}
	case 16:
		yyDollar = yyS[yypt-1 : yypt+1]
//line parser.y:131
		{
			yyVAL.vals = []any{yyDollar[1].val.val}
		}
	case 17:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:135
		{
			yyVAL.vals = append(yyVAL.vals, yyDollar[3].val.val)
		}
	case 19:
		yyDollar = yyS[yypt-2 : yypt+1]
//line parser.y:141
		{
			yyVAL.val = yyDollar[2].val.Not()
		}
	case 20:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:144
		{
			yyVAL.val = yyDollar[1].val.Eq(yyDollar[3].val)
		}
	case 21:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:147
		{
			yyVAL.val = yyDollar[1].val.Neq(yyDollar[3].val)
		}
	case 22:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:150
		{
			yyVAL.val = yyDollar[1].val.Gte(yyDollar[3].val)
		}
	case 23:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:153
		{
			yyVAL.val = yyDollar[1].val.Lte(yyDollar[3].val)
		}
	case 24:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:156
		{
			yyVAL.val = yyDollar[1].val.Re(yyDollar[3].val)
		}
	case 25:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:159
		{
			yyVAL.val = yyDollar[1].val.Nre(yyDollar[3].val)
		}
	case 26:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:162
		{
			yyVAL.val = yyDollar[1].val.Nc(yyDollar[3].val)
		}
	case 27:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:165
		{
			yyVAL.val = yyDollar[1].val.In(yyDollar[3].val)
		}
	case 28:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:168
		{
			yyVAL.val = yyDollar[1].val.Lt(yyDollar[3].val)
		}
	case 29:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:171
		{
			yyVAL.val = yyDollar[1].val.Gt(yyDollar[3].val)
		}
	case 30:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:174
		{
			yyVAL.val = yyDollar[1].val.Match(yyDollar[3].val)
		}
	case 31:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:177
		{
			yyVAL.val = yyDollar[1].val.Add(yyDollar[3].val)
		}
	case 32:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:180
		{
			yyVAL.val = yyDollar[1].val.Sub(yyDollar[3].val)
		}
	case 33:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:183
		{
			yyVAL.val = yyDollar[1].val.Multi(yyDollar[3].val)
		}
	case 34:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:186
		{
			yyVAL.val = yyDollar[1].val.Div(yyDollar[3].val)
		}
	case 35:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:189
		{
			yyVAL.val = yyDollar[1].val.Mod(yyDollar[3].val)
		}
	case 37:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:196
		{
			yyVAL.val = yyDollar[1].val.And(yyDollar[3].val)
		}
	case 38:
		yyDollar = yyS[yypt-3 : yypt+1]
//line parser.y:199
		{
			yyVAL.val = yyDollar[1].val.Or(yyDollar[3].val)
		}
	case 39:
		yyDollar = yyS[yypt-5 : yypt+1]
//line parser.y:202
		{
			yyVAL.val = yyDollar[1].val.Ternary(yyDollar[3].val, yyDollar[5].val)
		}
	}
	goto yystack /* stack new state and value */
}
