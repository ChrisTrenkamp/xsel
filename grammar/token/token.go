
// Package token is generated by GoGLL. Do not edit
package token

import(
    "fmt"
)

// Token is returned by the lexer for every scanned lexical token
type Token struct {
    typ        Type
    lext, rext int
    input      []rune
}

/*
New returns a new token.
lext is the left extent and rext the right extent of the token in the input.
input is the input slice scanned by the lexer.
*/
func New(t Type, lext, rext int, input []rune) *Token {
    return &Token{
        typ:   t,
        lext:  lext,
        rext:  rext,
        input: input,
    }
}

// GetLineColumn returns the line and column of the left extent of t
func (t *Token) GetLineColumn() (line, col int) {
    line, col = 1, 1
    for j := 0; j < t.lext; j++ {
        switch t.input[j] {
        case '\n':
            line++
            col = 1
        case '\t':
            col += 4
        default:
            col++
        }
    }
    return
}

// GetInput returns the input from which t was parsed.
func (t *Token) GetInput() []rune {
    return t.input
}

// Lext returns the left extent of t in the input stream of runes
func (t *Token) Lext() int {
    return t.lext
}

// Literal returns the literal runes of t scanned by the lexer
func (t *Token) Literal() []rune {
    return t.input[t.lext:t.rext]
}

// LiteralString returns string(t.Literal())
func (t *Token) LiteralString() string {
    return string(t.Literal())
}

// LiteralStripEscape returns the literal runes of t scanned by the lexer
func (t *Token) LiteralStripEscape() []rune {
	lit := t.Literal()
	strip := make([]rune, 0, len(lit))
	for i := 0; i < len(lit); i++ {
		if lit[i] == '\\' {
			i++
			switch lit[i] {
			case 't':
				strip = append(strip, '\t')
			case 'r':
				strip = append(strip, '\r')
			case 'n':
				strip = append(strip, '\r')
			default:
				strip = append(strip, lit[i])
			}
		} else {
			strip = append(strip, lit[i])
		}
	}
	return strip
}

// LiteralStringStripEscape returns string(t.LiteralStripEscape())
func (t *Token) LiteralStringStripEscape() string {
	return string(t.LiteralStripEscape())
}

// Rext returns the right extent of t in the input stream of runes
func (t *Token) Rext() int {
    return t.rext
}

func (t *Token) String() string {
    return fmt.Sprintf("%s (%d,%d) %s",
        t.TypeID(), t.lext, t.rext, t.LiteralString())
}

// Suppress returns true iff t is suppressed by the lexer
func (t *Token) Suppress() bool {
	return Suppress[t.typ]
}

// Type returns the token Type of t
func (t *Token) Type() Type {
    return t.typ
}

// TypeID returns the token Type ID of t. 
// This may be different from the literal of token t.
func (t *Token) TypeID() string {
    return t.Type().ID()
}

// Type is the token type
type Type int

func (t Type) String() string {
    return TypeToString[t]
}

// ID returns the token type ID of token Type t
func (t Type) ID() string {
    return TypeToID[t]
}


const(
    Error  Type = iota  // Error 
    EOF  // $ 
    T_0  // != 
    T_1  // ( 
    T_2  // ) 
    T_3  // * 
    T_4  // + 
    T_5  // , 
    T_6  // - 
    T_7  // . 
    T_8  // .. 
    T_9  // / 
    T_10  // // 
    T_11  // : 
    T_12  // :: 
    T_13  // < 
    T_14  // <= 
    T_15  // = 
    T_16  // > 
    T_17  // >= 
    T_18  // @ 
    T_19  // [ 
    T_20  // ] 
    T_21  // ancestor 
    T_22  // ancestor-or-self 
    T_23  // and 
    T_24  // attribute 
    T_25  // child 
    T_26  // comment 
    T_27  // descendant 
    T_28  // descendant-or-self 
    T_29  // digits 
    T_30  // div 
    T_31  // doublequote 
    T_32  // following 
    T_33  // following-sibling 
    T_34  // mod 
    T_35  // namespace 
    T_36  // ncname 
    T_37  // node 
    T_38  // or 
    T_39  // parent 
    T_40  // preceding 
    T_41  // preceding-sibling 
    T_42  // processing-instruction 
    T_43  // self 
    T_44  // singlequote 
    T_45  // text 
    T_46  // variableReference 
    T_47  // | 
)

var TypeToString = []string{ 
    "Error",
    "EOF",
    "T_0",
    "T_1",
    "T_2",
    "T_3",
    "T_4",
    "T_5",
    "T_6",
    "T_7",
    "T_8",
    "T_9",
    "T_10",
    "T_11",
    "T_12",
    "T_13",
    "T_14",
    "T_15",
    "T_16",
    "T_17",
    "T_18",
    "T_19",
    "T_20",
    "T_21",
    "T_22",
    "T_23",
    "T_24",
    "T_25",
    "T_26",
    "T_27",
    "T_28",
    "T_29",
    "T_30",
    "T_31",
    "T_32",
    "T_33",
    "T_34",
    "T_35",
    "T_36",
    "T_37",
    "T_38",
    "T_39",
    "T_40",
    "T_41",
    "T_42",
    "T_43",
    "T_44",
    "T_45",
    "T_46",
    "T_47",
}

var StringToType = map[string] Type { 
    "Error" : Error, 
    "EOF" : EOF, 
    "T_0" : T_0, 
    "T_1" : T_1, 
    "T_2" : T_2, 
    "T_3" : T_3, 
    "T_4" : T_4, 
    "T_5" : T_5, 
    "T_6" : T_6, 
    "T_7" : T_7, 
    "T_8" : T_8, 
    "T_9" : T_9, 
    "T_10" : T_10, 
    "T_11" : T_11, 
    "T_12" : T_12, 
    "T_13" : T_13, 
    "T_14" : T_14, 
    "T_15" : T_15, 
    "T_16" : T_16, 
    "T_17" : T_17, 
    "T_18" : T_18, 
    "T_19" : T_19, 
    "T_20" : T_20, 
    "T_21" : T_21, 
    "T_22" : T_22, 
    "T_23" : T_23, 
    "T_24" : T_24, 
    "T_25" : T_25, 
    "T_26" : T_26, 
    "T_27" : T_27, 
    "T_28" : T_28, 
    "T_29" : T_29, 
    "T_30" : T_30, 
    "T_31" : T_31, 
    "T_32" : T_32, 
    "T_33" : T_33, 
    "T_34" : T_34, 
    "T_35" : T_35, 
    "T_36" : T_36, 
    "T_37" : T_37, 
    "T_38" : T_38, 
    "T_39" : T_39, 
    "T_40" : T_40, 
    "T_41" : T_41, 
    "T_42" : T_42, 
    "T_43" : T_43, 
    "T_44" : T_44, 
    "T_45" : T_45, 
    "T_46" : T_46, 
    "T_47" : T_47, 
}

var TypeToID = []string { 
    "Error", 
    "$", 
    "!=", 
    "(", 
    ")", 
    "*", 
    "+", 
    ",", 
    "-", 
    ".", 
    "..", 
    "/", 
    "//", 
    ":", 
    "::", 
    "<", 
    "<=", 
    "=", 
    ">", 
    ">=", 
    "@", 
    "[", 
    "]", 
    "ancestor", 
    "ancestor-or-self", 
    "and", 
    "attribute", 
    "child", 
    "comment", 
    "descendant", 
    "descendant-or-self", 
    "digits", 
    "div", 
    "doublequote", 
    "following", 
    "following-sibling", 
    "mod", 
    "namespace", 
    "ncname", 
    "node", 
    "or", 
    "parent", 
    "preceding", 
    "preceding-sibling", 
    "processing-instruction", 
    "self", 
    "singlequote", 
    "text", 
    "variableReference", 
    "|", 
}

var IDToType = map[string]Type { 
    "Error": 0, 
    "$": 1, 
    "!=": 2, 
    "(": 3, 
    ")": 4, 
    "*": 5, 
    "+": 6, 
    ",": 7, 
    "-": 8, 
    ".": 9, 
    "..": 10, 
    "/": 11, 
    "//": 12, 
    ":": 13, 
    "::": 14, 
    "<": 15, 
    "<=": 16, 
    "=": 17, 
    ">": 18, 
    ">=": 19, 
    "@": 20, 
    "[": 21, 
    "]": 22, 
    "ancestor": 23, 
    "ancestor-or-self": 24, 
    "and": 25, 
    "attribute": 26, 
    "child": 27, 
    "comment": 28, 
    "descendant": 29, 
    "descendant-or-self": 30, 
    "digits": 31, 
    "div": 32, 
    "doublequote": 33, 
    "following": 34, 
    "following-sibling": 35, 
    "mod": 36, 
    "namespace": 37, 
    "ncname": 38, 
    "node": 39, 
    "or": 40, 
    "parent": 41, 
    "preceding": 42, 
    "preceding-sibling": 43, 
    "processing-instruction": 44, 
    "self": 45, 
    "singlequote": 46, 
    "text": 47, 
    "variableReference": 48, 
    "|": 49, 
}

var Suppress = []bool { 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
    false, 
}

