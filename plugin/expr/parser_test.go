package expr

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func evaluate(input string) (result float64, err error) {
	expr := &Expression{}
	if err = parser.ParseString("", input, expr); err != nil {
		return
	}
	result = expr.Eval()
	return
}

func test(t *testing.T, input string, output float64) {
	t.Helper()
	result, err := evaluate(input)
	assert.NoError(t, err)
	assert.Equal(t, result, output)
}

func TestExponent(t *testing.T) {
	test(t, "1E3", float64(1000))
	test(t, "1e3", float64(1000))
	test(t, "1e+3", float64(1000))
}

func TestExponentNeg(t *testing.T) {
	test(t, "2.3E-3", 0.0023)
	test(t, "2.3e-3", 0.0023)
}

func TestAddition(t *testing.T) {
	test(t, "1 + 2", float64(3))
	test(t, "1+2", float64(3))
	test(t, "2+0b100", float64(6))
}

func TestMultiplication(t *testing.T) {
	test(t, "5 * 5", 25)
	test(t, "5 * 3.5", 17.5)
}

func TestModulo(t *testing.T) {
	test(t, "10 % 3", 1)
}

func TestDivFloor(t *testing.T) {
	test(t, "10 // 3", 3)
}

func TestDiv(t *testing.T) {
	test(t, "10 / 3", 10.0/3)
}

func TestBinary(t *testing.T) {
	result, err := evaluate("0b10101")
	assert.NoError(t, err)
	assert.Equal(t, float64(21), result)
}

func TestOctal(t *testing.T) {
	result, err := evaluate("0o777")
	assert.NoError(t, err)
	assert.Equal(t, float64(511), result)
}

func TestHexadecimal(t *testing.T) {
	result, err := evaluate("0xdeadbeef")
	assert.NoError(t, err)
	assert.Equal(t, float64(3735928559), result)
}

func TestAddDifferentBases(t *testing.T) {
	result, err := evaluate("0xff + 0b10")
	assert.NoError(t, err)
	assert.Equal(t, float64(257), result)
}

func TestBitwiseOr(t *testing.T) {
	test(t, "3 | 5", 7)
}

func TestBitwiseAnd(t *testing.T) {
	test(t, "3 & 5", 1)
	test(t, "0b01011 & 0b11001", 0b01001)
}

func TestBitShifting(t *testing.T) {
	test(t, "0b1 << 3", 0b1000)
	test(t, "(0b1000 >> 2) + 1", 0b0011)
	test(t, "0b1000 >> 2 + 1", 0b0001)
	test(t, "1 << 5 >> 2", 8)
}

func TestNegativeNumbers(t *testing.T) {
	test(t, "4 + -3", 1)
	test(t, "-4 + -3", -7)
	test(t, "-4 - -3", -1)
	test(t, "-4 * -3", 12)
	test(t, "-0xff", -255)
}

func TestParenthesis(t *testing.T) {
	test(t, "(32)", 32)
	test(t, "(((((32)))))", 32)
	test(t, "2*(((((32)))))", 64)
}
