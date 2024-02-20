//line lexer.rl:1
package goeval

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/tidwall/gjson"
)

//line lexer.go:18
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
	52, 55, 57, 59, 61, 62, 63, 66,
	68, 69, 71, 74, 76, 78, 81, 83,
	85, 88, 90, 92, 123, 126, 128, 130,
	132, 134, 136, 138, 140, 142, 144, 145,
	147, 149, 152, 154, 162, 164, 166, 168,
	175, 176, 178, 179, 180, 187, 189, 197,
	205, 213, 221, 229, 237, 245, 253, 261,
}

var _expression_trans_keys []byte = []byte{
	34, 92, 34, 92, 91, 34, 39, 96,
	34, 92, 93, 34, 92, 34, 92, 93,
	39, 92, 39, 92, 39, 92, 93, 92,
	96, 92, 96, 92, 93, 96, 34, 39,
	96, 34, 92, 93, 34, 92, 34, 92,
	93, 39, 92, 39, 92, 39, 92, 93,
	92, 96, 92, 96, 92, 93, 96, 39,
	92, 39, 92, 48, 57, 46, 91, 34,
	39, 96, 34, 92, 93, 34, 92, 34,
	92, 93, 39, 92, 39, 92, 39, 92,
	93, 92, 96, 92, 96, 92, 93, 96,
	92, 96, 92, 96, 32, 33, 34, 36,
	38, 39, 45, 46, 58, 60, 61, 62,
	63, 91, 93, 96, 102, 105, 110, 116,
	124, 9, 13, 37, 47, 48, 57, 65,
	90, 95, 122, 32, 9, 13, 61, 126,
	34, 92, 46, 91, 34, 92, 39, 92,
	92, 96, 34, 92, 39, 92, 92, 96,
	38, 39, 92, 48, 57, 46, 48, 57,
	48, 57, 36, 95, 48, 57, 65, 90,
	97, 122, 34, 92, 39, 92, 92, 96,
	95, 48, 57, 65, 90, 97, 122, 61,
	61, 126, 61, 63, 95, 48, 57, 65,
	90, 97, 122, 92, 96, 95, 97, 48,
	57, 65, 90, 98, 122, 95, 108, 48,
	57, 65, 90, 97, 122, 95, 115, 48,
	57, 65, 90, 97, 122, 95, 101, 48,
	57, 65, 90, 97, 122, 95, 110, 48,
	57, 65, 90, 97, 122, 95, 105, 48,
	57, 65, 90, 97, 122, 95, 108, 48,
	57, 65, 90, 97, 122, 95, 114, 48,
	57, 65, 90, 97, 122, 95, 117, 48,
	57, 65, 90, 97, 122, 124,
}

var _expression_single_lengths []byte = []byte{
	2, 2, 1, 3, 2, 1, 2, 3,
	2, 2, 3, 2, 2, 3, 3, 2,
	1, 2, 3, 2, 2, 3, 2, 2,
	3, 2, 2, 0, 1, 1, 3, 2,
	1, 2, 3, 2, 2, 3, 2, 2,
	3, 2, 2, 21, 1, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 1, 2,
	0, 1, 0, 2, 2, 2, 2, 1,
	1, 2, 1, 1, 1, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 1,
}

var _expression_range_lengths []byte = []byte{
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 1, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 5, 1, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 1, 3, 0, 0, 0, 3,
	0, 0, 0, 0, 3, 0, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 0,
}

var _expression_index_offsets []int16 = []int16{
	0, 3, 6, 8, 12, 15, 17, 20,
	24, 27, 30, 34, 37, 40, 44, 48,
	51, 53, 56, 60, 63, 66, 70, 73,
	76, 80, 83, 86, 88, 90, 92, 96,
	99, 101, 104, 108, 111, 114, 118, 121,
	124, 128, 131, 134, 161, 164, 167, 170,
	173, 176, 179, 182, 185, 188, 191, 193,
	196, 198, 201, 203, 209, 212, 215, 218,
	223, 225, 228, 230, 232, 237, 240, 246,
	252, 258, 264, 270, 276, 282, 288, 294,
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
	2, 36, 35, 37, 36, 35, 39, 38,
	40, 5, 41, 5, 42, 43, 44, 5,
	45, 46, 42, 47, 0, 48, 46, 42,
	45, 46, 49, 42, 45, 50, 43, 51,
	50, 43, 45, 50, 52, 43, 53, 45,
	44, 53, 54, 44, 53, 55, 45, 44,
	57, 2, 56, 57, 58, 56, 60, 61,
	62, 63, 65, 66, 67, 68, 64, 70,
	71, 72, 73, 64, 64, 75, 76, 77,
	78, 79, 80, 60, 64, 69, 74, 74,
	59, 60, 60, 81, 83, 84, 82, 2,
	3, 1, 86, 87, 85, 10, 11, 7,
	10, 15, 8, 18, 10, 9, 24, 25,
	21, 24, 29, 22, 32, 24, 23, 90,
	85, 2, 36, 35, 69, 82, 92, 69,
	91, 39, 91, 93, 94, 94, 94, 94,
	85, 45, 46, 42, 45, 50, 43, 53,
	45, 44, 94, 94, 94, 94, 96, 97,
	82, 98, 99, 82, 100, 82, 101, 82,
	74, 74, 74, 74, 0, 57, 2, 56,
	74, 103, 74, 74, 74, 102, 74, 104,
	74, 74, 74, 102, 74, 105, 74, 74,
	74, 102, 74, 106, 74, 74, 74, 102,
	74, 107, 74, 74, 74, 102, 74, 108,
	74, 74, 74, 102, 74, 109, 74, 74,
	74, 102, 74, 110, 74, 74, 74, 102,
	74, 105, 74, 74, 74, 102, 111, 85,
}

var _expression_trans_targs []byte = []byte{
	43, 0, 43, 1, 46, 43, 3, 4,
	8, 11, 5, 6, 43, 7, 48, 9,
	10, 49, 12, 13, 50, 15, 19, 22,
	16, 17, 43, 18, 51, 20, 21, 52,
	23, 24, 53, 25, 26, 55, 43, 58,
	29, 30, 31, 35, 38, 32, 33, 43,
	34, 60, 36, 37, 61, 39, 40, 62,
	41, 42, 69, 43, 44, 45, 46, 47,
	43, 54, 55, 56, 59, 57, 64, 65,
	66, 67, 68, 69, 70, 74, 75, 77,
	79, 43, 43, 43, 43, 43, 2, 14,
	43, 43, 43, 43, 27, 28, 63, 43,
	43, 43, 43, 43, 43, 43, 43, 71,
	72, 73, 68, 68, 76, 68, 78, 43,
}

var _expression_trans_actions []byte = []byte{
	59, 0, 7, 0, 61, 57, 0, 0,
	0, 0, 0, 0, 31, 0, 82, 0,
	0, 82, 0, 0, 82, 0, 0, 0,
	0, 0, 29, 0, 79, 0, 0, 79,
	0, 0, 79, 0, 0, 61, 55, 0,
	0, 0, 0, 0, 0, 0, 0, 27,
	0, 76, 0, 0, 76, 0, 0, 76,
	0, 0, 61, 35, 0, 0, 85, 85,
	33, 0, 85, 0, 85, 5, 0, 0,
	0, 0, 73, 85, 0, 0, 0, 0,
	0, 49, 51, 11, 19, 53, 0, 0,
	47, 45, 21, 37, 0, 0, 0, 43,
	41, 15, 9, 17, 13, 25, 39, 0,
	0, 0, 64, 70, 0, 67, 0, 23,
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
}

var _expression_eof_trans []int16 = []int16{
	1, 1, 6, 6, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 6, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 39, 6, 6, 6, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 0, 82, 83, 1, 86,
	89, 89, 89, 90, 90, 90, 86, 1,
	83, 92, 92, 86, 96, 96, 96, 97,
	83, 83, 83, 83, 1, 1, 103, 103,
	103, 103, 103, 103, 103, 103, 103, 86,
}

const expression_start int = 43
const expression_first_final int = 43
const expression_error int = -1

const expression_en_main int = 43

//line lexer.rl:20

type lexer struct {
	data        []byte
	p, pe, cs   int
	ts, te, act int
	answer      Value
	kv          map[string]any
	err         error
	tokens      []string
	fns         map[string]Func
	once        sync.Once
	json        gjson.Result
}

func (lex *lexer) token() string {
	return string(lex.data[lex.ts:lex.te])
}

func newLexer(data []byte, kv map[string]any, fns map[string]Func) *lexer {
	lex := &lexer{
		data: data,
		pe:   len(data),
		kv:   kv,
		fns:  fns,
	}

//line lexer.go:270
	{
		lex.cs = expression_start
		lex.ts = 0
		lex.te = 0
		lex.act = 0
	}

//line lexer.rl:47
	return lex
}

func (lex *lexer) Lex(out *yySymType) int {
	eof := lex.pe
	tok := 0

//line lexer.go:287
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

//line lexer.go:307
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
//line lexer.rl:103
				lex.act = 1
			case 4:
//line lexer.rl:55
				lex.act = 3
			case 5:
//line lexer.rl:133
				lex.act = 4
			case 6:
//line lexer.rl:147
				lex.act = 14
			case 7:
//line lexer.rl:74
				lex.act = 15
			case 8:
//line lexer.rl:89
				lex.act = 17
			case 9:
//line lexer.rl:60
				lex.act = 18
			case 10:
//line lexer.rl:79
				lex.act = 19
			case 11:
//line lexer.rl:156
				lex.act = 22
			case 12:
//line lexer.rl:103
				lex.te = (lex.p) + 1
				{
					tok = VALUE
					val := string(lex.token())
					val = strings.ReplaceAll(val, "\\\\", "\\")
					if val[0] == '"' {
						val = strings.ReplaceAll(val, "\\\"", "\"")
					} else if val[0] == '\'' {
						val = strings.ReplaceAll(val, "\\'", "'")
					} else if val[0] == '`' {
						val = strings.ReplaceAll(val, "\\`", "`")
					}
					out.val = NewValue("", val[1:len(val)-1])
					(lex.p)++
					goto _out

				}
			case 13:
//line lexer.rl:136
				lex.te = (lex.p) + 1
				{
					tok = EQ
					(lex.p)++
					goto _out
				}
			case 14:
//line lexer.rl:137
				lex.te = (lex.p) + 1
				{
					tok = NEQ
					(lex.p)++
					goto _out
				}
			case 15:
//line lexer.rl:138
				lex.te = (lex.p) + 1
				{
					tok = GTE
					(lex.p)++
					goto _out
				}
			case 16:
//line lexer.rl:139
				lex.te = (lex.p) + 1
				{
					tok = LTE
					(lex.p)++
					goto _out
				}
			case 17:
//line lexer.rl:140
				lex.te = (lex.p) + 1
				{
					tok = RE
					(lex.p)++
					goto _out
				}
			case 18:
//line lexer.rl:141
				lex.te = (lex.p) + 1
				{
					tok = NRE
					(lex.p)++
					goto _out
				}
			case 19:
//line lexer.rl:142
				lex.te = (lex.p) + 1
				{
					tok = AND
					(lex.p)++
					goto _out
				}
			case 20:
//line lexer.rl:143
				lex.te = (lex.p) + 1
				{
					tok = OR
					(lex.p)++
					goto _out
				}
			case 21:
//line lexer.rl:144
				lex.te = (lex.p) + 1
				{
					tok = NC
					(lex.p)++
					goto _out
				}
			case 22:
//line lexer.rl:89
				lex.te = (lex.p) + 1
				{
					tok = IDENTIFIER
					out.name = string(lex.data[lex.ts+5 : lex.te-2])
					(lex.p)++
					goto _out

				}
			case 23:
//line lexer.rl:60
				lex.te = (lex.p) + 1
				{
					tok = VALUE
					path := string(lex.data[lex.ts+3 : lex.te-2])
					lex.once.Do(func() {
						bs, err := json.Marshal(lex.kv)
						if err != nil {
							panic(fmt.Errorf("parameter json marshal failed, %s", err))
						}
						lex.json = gjson.ParseBytes(bs)
					})
					res := lex.json.Get(path)
					out.val = NewValue(path, res)
					(lex.p)++
					goto _out

				}
			case 24:
//line lexer.rl:79
				lex.te = (lex.p) + 1
				{
					tok = IDENTIFIER
					out.name = string(lex.data[lex.ts+4 : lex.te-2])
					(lex.p)++
					goto _out

				}
			case 25:
//line lexer.rl:155
				lex.te = (lex.p) + 1
				{
					tok = int(lex.data[lex.ts])
					(lex.p)++
					goto _out
				}
			case 26:
//line lexer.rl:156
				lex.te = (lex.p) + 1
				{
					panic(errors.New("unexpected character: " + string(lex.data[lex.ts])))
				}
			case 27:
//line lexer.rl:94
				lex.te = (lex.p)
				(lex.p)--
				{
					tok = VALUE
					n, err := strconv.ParseFloat(lex.token(), 64)
					if err != nil {
						panic(err)
					}
					out.val = NewValue("", n)
					(lex.p)++
					goto _out

				}
			case 28:
//line lexer.rl:74
				lex.te = (lex.p)
				(lex.p)--
				{
					tok = IDENTIFIER
					out.name = lex.token()
					(lex.p)++
					goto _out

				}
			case 29:
//line lexer.rl:84
				lex.te = (lex.p)
				(lex.p)--
				{
					tok = IDENTIFIER
					out.name = string(lex.data[lex.ts+1 : lex.te])
					(lex.p)++
					goto _out

				}
			case 30:
//line lexer.rl:89
				lex.te = (lex.p)
				(lex.p)--
				{
					tok = IDENTIFIER
					out.name = string(lex.data[lex.ts+5 : lex.te-2])
					(lex.p)++
					goto _out

				}
			case 31:
//line lexer.rl:60
				lex.te = (lex.p)
				(lex.p)--
				{
					tok = VALUE
					path := string(lex.data[lex.ts+3 : lex.te-2])
					lex.once.Do(func() {
						bs, err := json.Marshal(lex.kv)
						if err != nil {
							panic(fmt.Errorf("parameter json marshal failed, %s", err))
						}
						lex.json = gjson.ParseBytes(bs)
					})
					res := lex.json.Get(path)
					out.val = NewValue(path, res)
					(lex.p)++
					goto _out

				}
			case 32:
//line lexer.rl:79
				lex.te = (lex.p)
				(lex.p)--
				{
					tok = IDENTIFIER
					out.name = string(lex.data[lex.ts+4 : lex.te-2])
					(lex.p)++
					goto _out

				}
			case 33:
//line lexer.rl:154
				lex.te = (lex.p)
				(lex.p)--

			case 34:
//line lexer.rl:155
				lex.te = (lex.p)
				(lex.p)--
				{
					tok = int(lex.data[lex.ts])
					(lex.p)++
					goto _out
				}
			case 35:
//line lexer.rl:156
				lex.te = (lex.p)
				(lex.p)--
				{
					panic(errors.New("unexpected character: " + string(lex.data[lex.ts])))
				}
			case 36:
//line lexer.rl:94
				(lex.p) = (lex.te) - 1
				{
					tok = VALUE
					n, err := strconv.ParseFloat(lex.token(), 64)
					if err != nil {
						panic(err)
					}
					out.val = NewValue("", n)
					(lex.p)++
					goto _out

				}
			case 37:
//line lexer.rl:156
				(lex.p) = (lex.te) - 1
				{
					panic(errors.New("unexpected character: " + string(lex.data[lex.ts])))
				}
			case 38:
//line NONE:1
				switch lex.act {
				case 1:
					{
						(lex.p) = (lex.te) - 1

						tok = VALUE
						val := string(lex.token())
						val = strings.ReplaceAll(val, "\\\\", "\\")
						if val[0] == '"' {
							val = strings.ReplaceAll(val, "\\\"", "\"")
						} else if val[0] == '\'' {
							val = strings.ReplaceAll(val, "\\'", "'")
						} else if val[0] == '`' {
							val = strings.ReplaceAll(val, "\\`", "`")
						}
						out.val = NewValue("", val[1:len(val)-1])
						(lex.p)++
						goto _out

					}
				case 3:
					{
						(lex.p) = (lex.te) - 1

						tok = VALUE
						out.val = NewValue("", lex.token() == "true")
						(lex.p)++
						goto _out

					}
				case 4:
					{
						(lex.p) = (lex.te) - 1
						tok = VALUE
						out.val = nilValue
						(lex.p)++
						goto _out
					}
				case 14:
					{
						(lex.p) = (lex.te) - 1
						tok = IN
						(lex.p)++
						goto _out
					}
				case 15:
					{
						(lex.p) = (lex.te) - 1

						tok = IDENTIFIER
						out.name = lex.token()
						(lex.p)++
						goto _out

					}
				case 17:
					{
						(lex.p) = (lex.te) - 1

						tok = IDENTIFIER
						out.name = string(lex.data[lex.ts+5 : lex.te-2])
						(lex.p)++
						goto _out

					}
				case 18:
					{
						(lex.p) = (lex.te) - 1

						tok = VALUE
						path := string(lex.data[lex.ts+3 : lex.te-2])
						lex.once.Do(func() {
							bs, err := json.Marshal(lex.kv)
							if err != nil {
								panic(fmt.Errorf("parameter json marshal failed, %s", err))
							}
							lex.json = gjson.ParseBytes(bs)
						})
						res := lex.json.Get(path)
						out.val = NewValue(path, res)
						(lex.p)++
						goto _out

					}
				case 19:
					{
						(lex.p) = (lex.te) - 1

						tok = IDENTIFIER
						out.name = string(lex.data[lex.ts+4 : lex.te-2])
						(lex.p)++
						goto _out

					}
				case 22:
					{
						(lex.p) = (lex.te) - 1
						panic(errors.New("unexpected character: " + string(lex.data[lex.ts])))
					}
				}

//line lexer.go:704
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

//line lexer.go:718
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

//line lexer.rl:160

	if tok != 0 {
		lex.tokens = append(lex.tokens, lex.token())
	}

	return tok
}

func (lex *lexer) Error(e string) {
	lex.err = errors.New(e)
}
