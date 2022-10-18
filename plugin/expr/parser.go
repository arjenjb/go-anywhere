// nolint: govet
package expr

import (
	"fmt"
	"github.com/alecthomas/participle/v2/lexer"
	"math"
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2"
)

type Operator int

const (
	// In order of precedence

	// Exponent
	OpExp Operator = iota

	// Factors
	OpMul
	OpDiv
	OpMod
	OpDivFloor

	// Terms
	OpAdd
	OpSub

	// Bit shift
	OpShl
	OpShr

	// Bitwise and
	OpAnd

	// Bitwise xor
	OpXor

	// Bitwise or
	OpOr

	// Comparison
	OpGt
	OpGe
	OpLt
	OpLe
	OpEq
	OpNe
)

var operatorMap = map[string]Operator{
	"+":   OpAdd,
	"-":   OpSub,
	"*":   OpMul,
	"/":   OpDiv,
	"%":   OpMod,
	"//":  OpDivFloor,
	"|":   OpOr,
	"&":   OpAnd,
	"xor": OpXor,
	"^":   OpExp,
	">>":  OpShr,
	"<<":  OpShl,
	">":   OpGt,
	">=":  OpGe,
	"<":   OpLt,
	"<=":  OpLe,
	"==":  OpEq,
	"!=":  OpNe,
}

func (o *Operator) Capture(s []string) error {
	*o = operatorMap[s[0]]
	return nil
}

//func (si *Sign) Capture(s []string) error {
//	if s[0] == "-" {
//		sign = -1
//	} else {
//		sign = 1
//	}
//
//	return inl
//}

// E --> T {( "+" | "-" ) T}
// T --> F {( "*" | "/" ) F}
// F --> P ["^" F]
// P --> v | "(" E ")" | "-" T

type Base struct {
	Hex     *string  `   @Hexadecimal`
	Octal   *string  ` | @Octal`
	Binary  *string  ` | @Binary`
	Decimal *float64 ` | @((Int? ".")? Int)`
}

type Exponent struct {
	Sign   string `("e" | "E") @("-" | "+")?`
	Number int    `@Int`
}

type Number struct {
	Negative bool      `@"-"?`
	Base     Base      `@@`
	Exponent *Exponent `@@?`
}

type Value struct {
	Number        *Number     `@@`
	Subexpression *Expression `| "(" @@ ")"`
}

type Factor struct {
	Base     *Value `@@`
	Exponent *Value `(WS? "^" WS? @@ )?`
}

type OpFactor struct {
	Operator Operator `WS? @("*" | "/" | "//" | "%") WS?`
	Factor   *Factor  `@@`
}

type Term struct {
	Left  *Factor     `@@`
	Right []*OpFactor `@@*`
}

type OpTerm struct {
	Operator Operator `WS? @("+" | "-") WS?`
	Term     *Term    `@@`
}

type Shift struct {
	Left  *Term     `@@`
	Right []*OpTerm `@@*`
}

type OpShift struct {
	Operator Operator `WS? @("<<" | ">>") WS?`
	Shift    *Shift   `@@`
}

type BitAnd struct {
	Left  *Shift     `@@`
	Right []*OpShift `@@*`
}

type OpBitAnd struct {
	Operator Operator `WS? @("&") WS?`
	BitAnd   *BitAnd  `@@`
}

type BitOr struct {
	Left  *BitAnd     `@@`
	Right []*OpBitAnd `@@*`
}

type OpBitOr struct {
	Operator Operator `WS? @("|") WS?`
	BitOr    *BitOr   `@@`
}

type Expression struct {
	Left  *BitOr     `@@`
	Right []*OpBitOr `@@*`
}

// Display

func (o Operator) String() string {
	switch o {
	case OpMul:
		return "*"
	case OpDiv:
		return "/"
	case OpSub:
		return "-"
	case OpAdd:
		return "+"
	}
	panic("unsupported operator")
}

func (v *Value) String() string {
	if v.Number != nil {
		return fmt.Sprintf("%g", v.Number.Eval())
	}
	return "(" + v.Subexpression.String() + ")"
}

func (f *Factor) String() string {
	out := f.Base.String()
	if f.Exponent != nil {
		out += " ^ " + f.Exponent.String()
	}
	return out
}

func (o *OpFactor) String() string {
	return fmt.Sprintf("%s %s", o.Operator, o.Factor)
}

func (t *Term) String() string {
	out := []string{t.Left.String()}
	for _, r := range t.Right {
		out = append(out, r.String())
	}
	return strings.Join(out, " ")
}

func (o *OpTerm) String() string {
	return fmt.Sprintf("%s %s", o.Operator, o.Term)
}

func (s Shift) String() string {
	out := []string{s.Left.String()}
	for _, r := range s.Right {
		out = append(out, r.String())
	}
	return strings.Join(out, " ")
}

func (s OpShift) String() string {
	return fmt.Sprintf("%s %s", s.Operator, s.Shift)
}

func (a BitAnd) String() string {
	out := []string{a.Left.String()}
	for _, r := range a.Right {
		out = append(out, r.String())
	}
	return strings.Join(out, " ")
}

func (a OpBitAnd) String() string {
	return fmt.Sprintf("%s %s", a.Operator, a.BitAnd)
}

func (o BitOr) String() string {
	out := []string{o.Left.String()}
	for _, r := range o.Right {
		out = append(out, r.String())
	}
	return strings.Join(out, " ")
}

func (o OpBitOr) String() string {
	return fmt.Sprintf("%s %s", o.Operator, o.BitOr)
}

func (e *Expression) String() string {
	out := []string{e.Left.String()}
	for _, r := range e.Right {
		out = append(out, r.String())
	}
	return strings.Join(out, " ")
}

// Evaluation

func (n Number) Eval() float64 {
	v := n.Base.Eval()
	if n.Negative {
		v = v * -1
	}

	if n.Exponent != nil {
		if n.Exponent.Sign == "-" {
			return float64(v * math.Pow(10, -float64(n.Exponent.Number)))
		} else {
			return v * math.Pow(10, float64(n.Exponent.Number))
		}
	} else {
		return v
	}
}

func (b Base) Eval() float64 {
	if b.Hex != nil {
		i, _ := strconv.ParseInt(*b.Hex, 0, 64)
		return float64(i)
	} else if b.Octal != nil {
		i, _ := strconv.ParseInt(*b.Octal, 0, 64)
		return float64(i)
	} else if b.Binary != nil {
		i, _ := strconv.ParseInt(*b.Binary, 0, 64)
		return float64(i)
	} else {
		return float64(*b.Decimal)
	}
}

func (o Operator) Eval(l, r float64) float64 {
	switch o {
	case OpMul:
		return l * r
	case OpDiv:
		return l / r
	case OpAdd:
		return l + r
	case OpSub:
		return l - r
	case OpMod:
		return math.Mod(l, r)
	case OpDivFloor:
		return math.Floor(l / r)
	case OpAnd:
		return float64(int(l) & int(r))
	case OpOr:
		return float64(int(l) | int(r))
	case OpShl:
		return float64(int(l) << int(r))
	case OpShr:
		return float64(int(l) >> int(r))
	}
	panic("unsupported operator")
}

func (v *Value) Eval() float64 {
	switch {
	case v.Number != nil:
		return v.Number.Eval()
	default:
		return v.Subexpression.Eval()
	}
}

func (f *Factor) Eval() float64 {
	b := f.Base.Eval()
	if f.Exponent != nil {
		return math.Pow(b, f.Exponent.Eval())
	}
	return b
}

func (t *Term) Eval() float64 {
	n := t.Left.Eval()
	for _, r := range t.Right {
		n = r.Operator.Eval(n, r.Factor.Eval())
	}
	return n
}

func (a *BitAnd) Eval() float64 {
	n := a.Left.Eval()
	for _, r := range a.Right {
		n = r.Operator.Eval(n, r.Shift.Eval())
	}
	return n
}

func (s *Shift) Eval() float64 {
	n := s.Left.Eval()
	for _, r := range s.Right {
		n = r.Operator.Eval(n, r.Term.Eval())
	}
	return n
}
func (o BitOr) Eval() float64 {
	n := o.Left.Eval()
	for _, r := range o.Right {
		n = r.Operator.Eval(n, r.BitAnd.Eval())
	}
	return n
}

func (e *Expression) Eval() float64 {
	l := e.Left.Eval()
	for _, r := range e.Right {
		l = r.Operator.Eval(l, r.BitOr.Eval())
	}
	return l
}

var l = lexer.MustSimple([]lexer.SimpleRule{
	{`Hexadecimal`, `0x[A-Fa-f0-9]+`},
	{`Octal`, `0o[0-7]+`},
	{`Binary`, `0b[01]+`},
	{`Int`, `\d+`},
	{`Operators`, `<<|>>|!=|<=|>=|==|//|[-+*/%,.()=<>|&^]`},
	{`Exponent`, `e|E`},
	{"WS", ` +`},
})

var parser = participle.MustBuild(&Expression{}, participle.Lexer(l))
