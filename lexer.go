//line lexer.rl:1
package goeval

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/tidwall/gjson"
)

//line lexer.go:16
var _expression_actions []byte = []byte{
	0, 1, 0, 1, 1, 1, 2, 1, 12,
	1, 13, 1, 14, 1, 15, 1, 16,
	1, 17, 1, 18, 1, 19, 1, 20,
	1, 21, 1, 22, 1, 23, 1, 24,
	1, 25, 1, 26, 1, 27, 1, 28,
	1, 29, 1, 30, 1, 31, 1, 32,
	1, 33, 1, 34, 1, 35, 1, 36,
	1, 37, 1, 38, 2, 2, 3, 2,
	2, 4, 2, 2, 5, 2, 2, 6,
	2, 2, 7, 2, 2, 8, 2, 2,
	9, 2, 2, 10, 2, 2, 11,
}

var _expression_key_offsets []int16 = []int16{
	0, 2, 4, 5, 8, 10, 11, 13,
	16, 18, 20, 23, 25, 27, 30, 33,
	35, 36, 38, 41, 43, 45, 48, 50,
	52, 55, 57, 59, 60, 61, 64, 66,
	67, 69, 72, 74, 76, 79, 81, 83,
	86, 88, 90, 92, 127, 130, 132, 134,
	136, 138, 140, 142, 144, 146, 148, 149,
	151, 159, 161, 163, 165, 172, 175, 177,
	178, 180, 181, 182, 189, 191, 199, 207,
	215, 223, 231, 239, 247, 255, 263, 271,
	279, 287, 295, 303, 311, 319, 326, 334,
	342, 350, 358, 366, 374, 382, 390, 399,
	407, 415, 423, 431, 439, 447, 455, 463,
	471, 479, 487, 495, 503, 510, 518, 526,
	534,
}

var _expression_trans_keys []byte = []byte{
	34, 92, 34, 92, 91, 34, 39, 96,
	34, 92, 93, 34, 92, 34, 92, 93,
	39, 92, 39, 92, 39, 92, 93, 92,
	96, 92, 96, 92, 93, 96, 34, 39,
	96, 34, 92, 93, 34, 92, 34, 92,
	93, 39, 92, 39, 92, 39, 92, 93,
	92, 96, 92, 96, 92, 93, 96, 39,
	92, 39, 92, 46, 91, 34, 39, 96,
	34, 92, 93, 34, 92, 34, 92, 93,
	39, 92, 39, 92, 39, 92, 93, 92,
	96, 92, 96, 92, 93, 96, 48, 57,
	92, 96, 92, 96, 32, 33, 34, 36,
	38, 39, 46, 58, 60, 61, 62, 63,
	91, 93, 96, 98, 99, 101, 102, 105,
	110, 115, 116, 119, 124, 9, 13, 37,
	47, 48, 57, 65, 90, 95, 122, 32,
	9, 13, 61, 126, 34, 92, 46, 91,
	34, 92, 39, 92, 92, 96, 34, 92,
	39, 92, 92, 96, 38, 39, 92, 36,
	95, 48, 57, 65, 90, 97, 122, 34,
	92, 39, 92, 92, 96, 95, 48, 57,
	65, 90, 97, 122, 46, 48, 57, 48,
	57, 61, 61, 126, 61, 63, 95, 48,
	57, 65, 90, 97, 122, 92, 96, 95,
	101, 48, 57, 65, 90, 97, 122, 95,
	116, 48, 57, 65, 90, 97, 122, 95,
	119, 48, 57, 65, 90, 97, 122, 95,
	101, 48, 57, 65, 90, 97, 122, 95,
	101, 48, 57, 65, 90, 97, 122, 95,
	110, 48, 57, 65, 90, 97, 122, 95,
	111, 48, 57, 65, 90, 97, 122, 95,
	110, 48, 57, 65, 90, 97, 122, 95,
	116, 48, 57, 65, 90, 97, 122, 95,
	97, 48, 57, 65, 90, 98, 122, 95,
	105, 48, 57, 65, 90, 97, 122, 95,
	110, 48, 57, 65, 90, 97, 122, 95,
	115, 48, 57, 65, 90, 97, 122, 95,
	110, 48, 57, 65, 90, 97, 122, 95,
	100, 48, 57, 65, 90, 97, 122, 95,
	115, 48, 57, 65, 90, 97, 122, 95,
	48, 57, 65, 90, 97, 122, 95, 119,
	48, 57, 65, 90, 97, 122, 95, 105,
	48, 57, 65, 90, 97, 122, 95, 116,
	48, 57, 65, 90, 97, 122, 95, 104,
	48, 57, 65, 90, 97, 122, 95, 97,
	48, 57, 65, 90, 98, 122, 95, 108,
	48, 57, 65, 90, 97, 122, 95, 115,
	48, 57, 65, 90, 97, 122, 95, 101,
	48, 57, 65, 90, 97, 122, 95, 105,
	111, 48, 57, 65, 90, 97, 122, 95,
	108, 48, 57, 65, 90, 97, 122, 95,
	116, 48, 57, 65, 90, 97, 122, 95,
	116, 48, 57, 65, 90, 97, 122, 95,
	97, 48, 57, 65, 90, 98, 122, 95,
	114, 48, 57, 65, 90, 97, 122, 95,
	116, 48, 57, 65, 90, 97, 122, 95,
	114, 48, 57, 65, 90, 97, 122, 95,
	117, 48, 57, 65, 90, 97, 122, 95,
	105, 48, 57, 65, 90, 97, 122, 95,
	116, 48, 57, 65, 90, 97, 122, 95,
	104, 48, 57, 65, 90, 97, 122, 95,
	105, 48, 57, 65, 90, 97, 122, 95,
	110, 48, 57, 65, 90, 97, 122, 95,
	48, 57, 65, 90, 97, 122, 95, 108,
	48, 57, 65, 90, 97, 122, 95, 97,
	48, 57, 65, 90, 98, 122, 95, 115,
	48, 57, 65, 90, 97, 122, 124,
}

var _expression_single_lengths []byte = []byte{
	2, 2, 1, 3, 2, 1, 2, 3,
	2, 2, 3, 2, 2, 3, 3, 2,
	1, 2, 3, 2, 2, 3, 2, 2,
	3, 2, 2, 1, 1, 3, 2, 1,
	2, 3, 2, 2, 3, 2, 2, 3,
	0, 2, 2, 25, 1, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 1, 2,
	2, 2, 2, 2, 1, 1, 0, 1,
	2, 1, 1, 1, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 1, 2, 2,
	2, 2, 2, 2, 2, 2, 3, 2,
	2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 1, 2, 2, 2,
	1,
}

var _expression_range_lengths []byte = []byte{
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	1, 0, 0, 5, 1, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	3, 0, 0, 0, 3, 1, 1, 0,
	0, 0, 0, 3, 0, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3,
	0,
}

var _expression_index_offsets []int16 = []int16{
	0, 3, 6, 8, 12, 15, 17, 20,
	24, 27, 30, 34, 37, 40, 44, 48,
	51, 53, 56, 60, 63, 66, 70, 73,
	76, 80, 83, 86, 88, 90, 94, 97,
	99, 102, 106, 109, 112, 116, 119, 122,
	126, 128, 131, 134, 165, 168, 171, 174,
	177, 180, 183, 186, 189, 192, 195, 197,
	200, 206, 209, 212, 215, 220, 223, 225,
	227, 230, 232, 234, 239, 242, 248, 254,
	260, 266, 272, 278, 284, 290, 296, 302,
	308, 314, 320, 326, 332, 338, 343, 349,
	355, 361, 367, 373, 379, 385, 391, 398,
	404, 410, 416, 422, 428, 434, 440, 446,
	452, 458, 464, 470, 476, 481, 487, 493,
	499,
}

var _expression_indicies []byte = []byte{
	2, 3, 1, 4, 3, 1, 6, 5,
	7, 8, 9, 5, 10, 11, 7, 12,
	0, 13, 11, 7, 10, 11, 14, 7,
	10, 15, 8, 16, 15, 8, 10, 15,
	17, 8, 18, 10, 9, 18, 19, 9,
	18, 20, 10, 9, 21, 22, 23, 5,
	24, 25, 21, 26, 0, 27, 25, 21,
	24, 25, 28, 21, 24, 29, 22, 30,
	29, 22, 24, 29, 31, 22, 32, 24,
	23, 32, 33, 23, 32, 34, 24, 23,
	2, 36, 35, 37, 36, 35, 38, 5,
	39, 5, 40, 41, 42, 5, 43, 44,
	40, 45, 0, 46, 44, 40, 43, 44,
	47, 40, 43, 48, 41, 49, 48, 41,
	43, 48, 50, 41, 51, 43, 42, 51,
	52, 42, 51, 53, 43, 42, 55, 54,
	57, 2, 56, 57, 58, 56, 60, 61,
	62, 63, 65, 66, 67, 64, 69, 70,
	71, 72, 64, 64, 74, 75, 76, 77,
	78, 79, 80, 81, 82, 83, 84, 60,
	64, 68, 73, 73, 59, 60, 60, 85,
	87, 88, 86, 2, 3, 1, 90, 91,
	89, 10, 11, 7, 10, 15, 8, 18,
	10, 9, 24, 25, 21, 24, 29, 22,
	32, 24, 23, 94, 89, 2, 36, 35,
	95, 96, 96, 96, 96, 89, 43, 44,
	40, 43, 48, 41, 51, 43, 42, 96,
	96, 96, 96, 98, 100, 68, 99, 55,
	99, 101, 86, 102, 103, 86, 104, 86,
	105, 86, 73, 73, 73, 73, 0, 57,
	2, 56, 73, 107, 73, 73, 73, 106,
	73, 108, 73, 73, 73, 106, 73, 109,
	73, 73, 73, 106, 73, 110, 73, 73,
	73, 106, 73, 79, 73, 73, 73, 106,
	73, 111, 73, 73, 73, 106, 73, 112,
	73, 73, 73, 106, 73, 113, 73, 73,
	73, 106, 73, 114, 73, 73, 73, 106,
	73, 115, 73, 73, 73, 106, 73, 116,
	73, 73, 73, 106, 73, 117, 73, 73,
	73, 106, 73, 111, 73, 73, 73, 106,
	73, 118, 73, 73, 73, 106, 73, 119,
	73, 73, 73, 106, 73, 120, 73, 73,
	73, 106, 121, 73, 73, 73, 106, 73,
	122, 73, 73, 73, 106, 73, 123, 73,
	73, 73, 106, 73, 124, 73, 73, 73,
	106, 73, 111, 73, 73, 73, 106, 73,
	125, 73, 73, 73, 106, 73, 126, 73,
	73, 73, 106, 73, 127, 73, 73, 73,
	106, 73, 128, 73, 73, 73, 106, 73,
	129, 130, 73, 73, 73, 106, 73, 131,
	73, 73, 73, 106, 73, 111, 73, 73,
	73, 106, 73, 132, 73, 73, 73, 106,
	73, 133, 73, 73, 73, 106, 73, 134,
	73, 73, 73, 106, 73, 119, 73, 73,
	73, 106, 73, 135, 73, 73, 73, 106,
	73, 127, 73, 73, 73, 106, 73, 136,
	73, 73, 73, 106, 73, 137, 73, 73,
	73, 106, 73, 138, 73, 73, 73, 106,
	73, 139, 73, 73, 73, 106, 73, 140,
	73, 73, 73, 106, 141, 73, 73, 73,
	106, 73, 142, 73, 73, 73, 106, 73,
	143, 73, 73, 73, 106, 73, 130, 73,
	73, 73, 106, 144, 89,
}

var _expression_trans_targs []byte = []byte{
	43, 0, 43, 1, 46, 43, 3, 4,
	8, 11, 5, 6, 43, 7, 48, 9,
	10, 49, 12, 13, 50, 15, 19, 22,
	16, 17, 43, 18, 51, 20, 21, 52,
	23, 24, 53, 25, 26, 55, 28, 29,
	30, 34, 37, 31, 32, 43, 33, 57,
	35, 36, 58, 38, 39, 59, 43, 62,
	41, 42, 68, 43, 44, 45, 46, 47,
	43, 54, 55, 56, 61, 63, 64, 65,
	66, 67, 68, 69, 75, 82, 90, 74,
	94, 97, 101, 103, 112, 43, 43, 43,
	43, 43, 2, 14, 43, 43, 43, 27,
	60, 43, 43, 43, 40, 43, 43, 43,
	43, 43, 43, 70, 71, 72, 73, 67,
	76, 77, 78, 79, 80, 81, 83, 84,
	85, 86, 87, 88, 89, 91, 92, 93,
	67, 95, 96, 67, 98, 99, 100, 102,
	104, 105, 106, 107, 108, 109, 110, 111,
	43,
}

var _expression_trans_actions []byte = []byte{
	59, 0, 7, 0, 61, 57, 0, 0,
	0, 0, 0, 0, 31, 0, 82, 0,
	0, 82, 0, 0, 82, 0, 0, 0,
	0, 0, 29, 0, 79, 0, 0, 79,
	0, 0, 79, 0, 0, 61, 0, 0,
	0, 0, 0, 0, 0, 27, 0, 76,
	0, 0, 76, 0, 0, 76, 55, 0,
	0, 0, 61, 35, 0, 0, 85, 85,
	33, 0, 85, 85, 5, 0, 0, 0,
	0, 73, 85, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 49, 51, 11,
	19, 53, 0, 0, 47, 45, 21, 0,
	0, 43, 41, 37, 0, 15, 9, 17,
	13, 25, 39, 0, 0, 0, 0, 70,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	64, 0, 0, 67, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	23,
}

var _expression_to_state_actions []byte = []byte{
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 1, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0,
}

var _expression_from_state_actions []byte = []byte{
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 3, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0,
}

var _expression_eof_trans []int16 = []int16{
	1, 1, 6, 6, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 6, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 6, 6, 6, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	55, 1, 1, 0, 86, 87, 1, 90,
	93, 93, 93, 94, 94, 94, 90, 1,
	90, 98, 98, 98, 99, 100, 100, 87,
	87, 87, 87, 1, 1, 107, 107, 107,
	107, 107, 107, 107, 107, 107, 107, 107,
	107, 107, 107, 107, 107, 107, 107, 107,
	107, 107, 107, 107, 107, 107, 107, 107,
	107, 107, 107, 107, 107, 107, 107, 107,
	107, 107, 107, 107, 107, 107, 107, 107,
	90,
}

const expression_start int = 43
const expression_first_final int = 43
const expression_error int = -1

const expression_en_main int = 43

//line lexer.rl:18

type lexer struct {
	data          string
	p, pe, cs     int
	ts, te, act   int
	answer        Value
	kv            map[string]any
	err           error
	tokens        []string
	collectTokens bool
	fns           map[string]Func
	useDecimal    bool
	jsonParsed    bool
	json          gjson.Result
}

var lexerPool = sync.Pool{
	New: func() any {
		return new(lexer)
	},
}

func (lex *lexer) token() string {
	return lex.data[lex.ts:lex.te]
}

func (lex *lexer) position(pos int) (int, int) {
	if pos < 0 {
		pos = 0
	}
	if pos > len(lex.data) {
		pos = len(lex.data)
	}
	line := 1
	column := 1
	for i := 0; i < pos; i++ {
		if lex.data[i] == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	return line, column
}

func (lex *lexer) errorAt(pos int, format string, args ...any) error {
	line, column := lex.position(pos)
	return fmt.Errorf("%s at line %d, column %d", fmt.Sprintf(format, args...), line, column)
}

func newLexer(data string, kv map[string]any, fns map[string]Func, collectTokens bool, useDecimal bool) *lexer {
	lex := lexerPool.Get().(*lexer)
	*lex = lexer{
		data:          data,
		pe:            len(data),
		kv:            kv,
		fns:           fns,
		collectTokens: collectTokens,
		useDecimal:    useDecimal,
	}

//line lexer.go:408
	{
		lex.cs = expression_start
		lex.ts = 0
		lex.te = 0
		lex.act = 0
	}

//line lexer.rl:81
	return lex
}

func (lex *lexer) newValue(name string, val any) Value {
	return newValue(name, val, lex.useDecimal)
}

func (lex *lexer) release() {
	*lex = lexer{}
	lexerPool.Put(lex)
}

type lexeme struct {
	val        Value
	name       string
	jsonPath   string
	isJSONPath bool
}

func (lex *lexer) jsonPathValue(path string) Value {
	if val, ok := SelectPath(lex.kv, path); ok {
		return lex.newValue(path, val)
	}
	if !lex.jsonParsed {
		bs, err := json.Marshal(lex.kv)
		if err != nil {
			panic(fmt.Errorf("parameter json marshal failed, %s", err))
		}
		lex.json = gjson.ParseBytes(bs)
		lex.jsonParsed = true
	}
	return lex.newValue(path, lex.json.Get(path))
}

func (lex *lexer) next() (int, lexeme) {
	eof := lex.pe
	tok := 0
	var item lexeme

//line lexer.go:457
	{
		var _klen int
		var _trans int
		var _acts int
		var _nacts uint
		var _keys int
		if (lex.p) == (lex.pe) {
			goto _test_eof
		}
	_resume:
		_acts = int(_expression_from_state_actions[lex.cs])
		_nacts = uint(_expression_actions[_acts])
		_acts++
		for ; _nacts > 0; _nacts-- {
			_acts++
			switch _expression_actions[_acts-1] {
			case 1:
//line NONE:1
				lex.ts = (lex.p)

//line lexer.go:477
			}
		}

		_keys = int(_expression_key_offsets[lex.cs])
		_trans = int(_expression_index_offsets[lex.cs])

		_klen = int(_expression_single_lengths[lex.cs])
		if _klen > 0 {
			_lower := int(_keys)
			var _mid int
			_upper := int(_keys + _klen - 1)
			for {
				if _upper < _lower {
					break
				}

				_mid = _lower + ((_upper - _lower) >> 1)
				switch {
				case lex.data[(lex.p)] < _expression_trans_keys[_mid]:
					_upper = _mid - 1
				case lex.data[(lex.p)] > _expression_trans_keys[_mid]:
					_lower = _mid + 1
				default:
					_trans += int(_mid - int(_keys))
					goto _match
				}
			}
			_keys += _klen
			_trans += _klen
		}

		_klen = int(_expression_range_lengths[lex.cs])
		if _klen > 0 {
			_lower := int(_keys)
			var _mid int
			_upper := int(_keys + (_klen << 1) - 2)
			for {
				if _upper < _lower {
					break
				}

				_mid = _lower + (((_upper - _lower) >> 1) & ^1)
				switch {
				case lex.data[(lex.p)] < _expression_trans_keys[_mid]:
					_upper = _mid - 2
				case lex.data[(lex.p)] > _expression_trans_keys[_mid+1]:
					_lower = _mid + 2
				default:
					_trans += int((_mid - int(_keys)) >> 1)
					goto _match
				}
			}
			_trans += _klen
		}

	_match:
		_trans = int(_expression_indicies[_trans])
	_eof_trans:
		lex.cs = int(_expression_trans_targs[_trans])

		if _expression_trans_actions[_trans] == 0 {
			goto _again
		}

		_acts = int(_expression_trans_actions[_trans])
		_nacts = uint(_expression_actions[_acts])
		_acts++
		for ; _nacts > 0; _nacts-- {
			_acts++
			switch _expression_actions[_acts-1] {
			case 2:
//line NONE:1
				lex.te = (lex.p) + 1

			case 3:
//line lexer.rl:177
				lex.act = 1
			case 4:
//line lexer.rl:121
				lex.act = 3
			case 5:
//line lexer.rl:209
				lex.act = 4
			case 6:
//line lexer.rl:137
				lex.act = 14
			case 7:
//line lexer.rl:132
				lex.act = 15
			case 8:
//line lexer.rl:167
				lex.act = 17
			case 9:
//line lexer.rl:126
				lex.act = 18
			case 10:
//line lexer.rl:157
				lex.act = 19
			case 11:
//line lexer.rl:232
				lex.act = 22
			case 12:
//line lexer.rl:177
				lex.te = (lex.p) + 1
				{
					tok = VALUE
					val := lex.token()
					if strings.IndexByte(val, '\\') >= 0 {
						val = strings.ReplaceAll(val, "\\\\", "\\")
						if val[0] == '"' {
							val = strings.ReplaceAll(val, "\\\"", "\"")
						} else if val[0] == '\'' {
							val = strings.ReplaceAll(val, "\\'", "'")
						} else if val[0] == '`' {
							val = strings.ReplaceAll(val, "\\`", "`")
						}
					}
					item.val = lex.newValue("", val[1:len(val)-1])
					(lex.p)++
					goto _out

				}
			case 13:
//line lexer.rl:212
				lex.te = (lex.p) + 1
				{
					tok = EQ
					(lex.p)++
					goto _out
				}
			case 14:
//line lexer.rl:213
				lex.te = (lex.p) + 1
				{
					tok = NEQ
					(lex.p)++
					goto _out
				}
			case 15:
//line lexer.rl:214
				lex.te = (lex.p) + 1
				{
					tok = GTE
					(lex.p)++
					goto _out
				}
			case 16:
//line lexer.rl:215
				lex.te = (lex.p) + 1
				{
					tok = LTE
					(lex.p)++
					goto _out
				}
			case 17:
//line lexer.rl:216
				lex.te = (lex.p) + 1
				{
					tok = RE
					(lex.p)++
					goto _out
				}
			case 18:
//line lexer.rl:217
				lex.te = (lex.p) + 1
				{
					tok = NRE
					(lex.p)++
					goto _out
				}
			case 19:
//line lexer.rl:218
				lex.te = (lex.p) + 1
				{
					tok = AND
					(lex.p)++
					goto _out
				}
			case 20:
//line lexer.rl:219
				lex.te = (lex.p) + 1
				{
					tok = OR
					(lex.p)++
					goto _out
				}
			case 21:
//line lexer.rl:220
				lex.te = (lex.p) + 1
				{
					tok = NC
					(lex.p)++
					goto _out
				}
			case 22:
//line lexer.rl:167
				lex.te = (lex.p) + 1
				{
					tok = IDENTIFIER
					item.name = lex.data[lex.ts+5 : lex.te-2]
					(lex.p)++
					goto _out

				}
			case 23:
//line lexer.rl:126
				lex.te = (lex.p) + 1
				{
					tok = VALUE
					item.jsonPath = lex.data[lex.ts+3 : lex.te-2]
					item.isJSONPath = true
					(lex.p)++
					goto _out

				}
			case 24:
//line lexer.rl:157
				lex.te = (lex.p) + 1
				{
					tok = IDENTIFIER
					item.name = lex.data[lex.ts+4 : lex.te-2]
					(lex.p)++
					goto _out

				}
			case 25:
//line lexer.rl:231
				lex.te = (lex.p) + 1
				{
					tok = int(lex.data[lex.ts])
					(lex.p)++
					goto _out
				}
			case 26:
//line lexer.rl:232
				lex.te = (lex.p) + 1
				{
					panic(lex.errorAt(lex.ts, "unexpected character %q", lex.data[lex.ts]))
				}
			case 27:
//line lexer.rl:172
				lex.te = (lex.p)
				(lex.p)--
				{
					tok = VALUE
					item.val = newNumberLiteral("", lex.token(), lex.useDecimal)
					(lex.p)++
					goto _out

				}
			case 28:
//line lexer.rl:132
				lex.te = (lex.p)
				(lex.p)--
				{
					tok = IDENTIFIER
					item.name = lex.token()
					(lex.p)++
					goto _out

				}
			case 29:
//line lexer.rl:162
				lex.te = (lex.p)
				(lex.p)--
				{
					tok = IDENTIFIER
					item.name = lex.data[lex.ts+1 : lex.te]
					(lex.p)++
					goto _out

				}
			case 30:
//line lexer.rl:167
				lex.te = (lex.p)
				(lex.p)--
				{
					tok = IDENTIFIER
					item.name = lex.data[lex.ts+5 : lex.te-2]
					(lex.p)++
					goto _out

				}
			case 31:
//line lexer.rl:126
				lex.te = (lex.p)
				(lex.p)--
				{
					tok = VALUE
					item.jsonPath = lex.data[lex.ts+3 : lex.te-2]
					item.isJSONPath = true
					(lex.p)++
					goto _out

				}
			case 32:
//line lexer.rl:157
				lex.te = (lex.p)
				(lex.p)--
				{
					tok = IDENTIFIER
					item.name = lex.data[lex.ts+4 : lex.te-2]
					(lex.p)++
					goto _out

				}
			case 33:
//line lexer.rl:230
				lex.te = (lex.p)
				(lex.p)--

			case 34:
//line lexer.rl:231
				lex.te = (lex.p)
				(lex.p)--
				{
					tok = int(lex.data[lex.ts])
					(lex.p)++
					goto _out
				}
			case 35:
//line lexer.rl:232
				lex.te = (lex.p)
				(lex.p)--
				{
					panic(lex.errorAt(lex.ts, "unexpected character %q", lex.data[lex.ts]))
				}
			case 36:
//line lexer.rl:172
				(lex.p) = (lex.te) - 1
				{
					tok = VALUE
					item.val = newNumberLiteral("", lex.token(), lex.useDecimal)
					(lex.p)++
					goto _out

				}
			case 37:
//line lexer.rl:232
				(lex.p) = (lex.te) - 1
				{
					panic(lex.errorAt(lex.ts, "unexpected character %q", lex.data[lex.ts]))
				}
			case 38:
//line NONE:1
				switch lex.act {
				case 1:
					{
						(lex.p) = (lex.te) - 1

						tok = VALUE
						val := lex.token()
						if strings.IndexByte(val, '\\') >= 0 {
							val = strings.ReplaceAll(val, "\\\\", "\\")
							if val[0] == '"' {
								val = strings.ReplaceAll(val, "\\\"", "\"")
							} else if val[0] == '\'' {
								val = strings.ReplaceAll(val, "\\'", "'")
							} else if val[0] == '`' {
								val = strings.ReplaceAll(val, "\\`", "`")
							}
						}
						item.val = lex.newValue("", val[1:len(val)-1])
						(lex.p)++
						goto _out

					}
				case 3:
					{
						(lex.p) = (lex.te) - 1

						tok = VALUE
						item.val = lex.newValue("", lex.token() == "true")
						(lex.p)++
						goto _out

					}
				case 4:
					{
						(lex.p) = (lex.te) - 1
						tok = VALUE
						item.val = nilValue
						(lex.p)++
						goto _out
					}
				case 14:
					{
						(lex.p) = (lex.te) - 1

						item.name = lex.token()
						switch item.name {
						case "contains":
							tok = CONTAINS
						case "starts_with":
							tok = STARTS_WITH
						case "ends_with":
							tok = ENDS_WITH
						case "between":
							tok = BETWEEN
						case "within_last":
							tok = WITHIN_LAST
						case "in":
							tok = IN
						case "not":
							tok = NOT
						}
						(lex.p)++
						goto _out

					}
				case 15:
					{
						(lex.p) = (lex.te) - 1

						tok = IDENTIFIER
						item.name = lex.token()
						(lex.p)++
						goto _out

					}
				case 17:
					{
						(lex.p) = (lex.te) - 1

						tok = IDENTIFIER
						item.name = lex.data[lex.ts+5 : lex.te-2]
						(lex.p)++
						goto _out

					}
				case 18:
					{
						(lex.p) = (lex.te) - 1

						tok = VALUE
						item.jsonPath = lex.data[lex.ts+3 : lex.te-2]
						item.isJSONPath = true
						(lex.p)++
						goto _out

					}
				case 19:
					{
						(lex.p) = (lex.te) - 1

						tok = IDENTIFIER
						item.name = lex.data[lex.ts+4 : lex.te-2]
						(lex.p)++
						goto _out

					}
				case 22:
					{
						(lex.p) = (lex.te) - 1
						panic(lex.errorAt(lex.ts, "unexpected character %q", lex.data[lex.ts]))
					}
				}

//line lexer.go:865
			}
		}

	_again:
		_acts = int(_expression_to_state_actions[lex.cs])
		_nacts = uint(_expression_actions[_acts])
		_acts++
		for ; _nacts > 0; _nacts-- {
			_acts++
			switch _expression_actions[_acts-1] {
			case 0:
//line NONE:1
				lex.ts = 0

//line lexer.go:879
			}
		}

		(lex.p)++
		if (lex.p) != (lex.pe) {
			goto _resume
		}
	_test_eof:
		{
		}
		if (lex.p) == eof {
			if _expression_eof_trans[lex.cs] > 0 {
				_trans = int(_expression_eof_trans[lex.cs] - 1)
				goto _eof_trans
			}
		}

	_out:
		{
		}
	}

//line lexer.rl:236

	if tok != 0 && lex.collectTokens {
		lex.tokens = append(lex.tokens, lex.token())
	}

	return tok, item
}

func (lex *lexer) Lex(out *yySymType) int {
	tok, item := lex.next()
	switch tok {
	case VALUE:
		if item.isJSONPath {
			out.val = lex.jsonPathValue(item.jsonPath)
		} else {
			out.val = item.val
		}
	case IDENTIFIER:
		out.name = item.name
	case CONTAINS, STARTS_WITH, ENDS_WITH, BETWEEN, WITHIN_LAST, IN, NOT:
		out.name = item.name
	}
	return tok
}

type compileLexerAdapter struct {
	*lexer
	answer exprNode
}

func (lex *compileLexerAdapter) Lex(out *compileSymType) int {
	tok, item := lex.next()
	switch tok {
	case VALUE:
		if item.isJSONPath {
			out.node = jsonPathNode{path: item.jsonPath}
		} else {
			out.node = literalNode{value: item.val}
		}
		return C_VALUE
	case IDENTIFIER:
		out.name = item.name
		return C_IDENTIFIER
	case CONTAINS:
		out.name = item.name
		return C_CONTAINS
	case STARTS_WITH:
		out.name = item.name
		return C_STARTS_WITH
	case ENDS_WITH:
		out.name = item.name
		return C_ENDS_WITH
	case BETWEEN:
		out.name = item.name
		return C_BETWEEN
	case WITHIN_LAST:
		out.name = item.name
		return C_WITHIN_LAST
	case NOT:
		out.name = item.name
		return C_NOT
	case IN:
		out.name = item.name
		return C_IN
	case EQ:
		return C_EQ
	case NEQ:
		return C_NEQ
	case GTE:
		return C_GTE
	case LTE:
		return C_LTE
	case RE:
		return C_RE
	case NRE:
		return C_NRE
	case AND:
		return C_AND
	case OR:
		return C_OR
	case NC:
		return C_NC
	default:
		return tok
	}
}

func (lex *lexer) Error(e string) {
	lex.err = lex.errorAt(lex.p, "%s", e)
}
