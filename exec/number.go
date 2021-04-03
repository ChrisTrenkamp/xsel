package exec

import (
	"math"
	"strconv"
)

type Number float64

func (n Number) String() string {
	if math.IsInf(float64(n), 1) {
		return "Infinity"
	}

	if math.IsInf(float64(n), -1) {
		return "-Infinity"
	}

	return strconv.FormatFloat(float64(n), 'f', -1, 64)
}

func (n Number) Number() float64 {
	return float64(n)
}

func (n Number) Bool() bool {
	return n != 0
}
