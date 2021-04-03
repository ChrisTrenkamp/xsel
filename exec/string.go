package exec

import (
	"math"
	"strconv"
)

type String string

func (n String) String() string {
	return string(n)
}

func (n String) Number() float64 {
	ret, err := strconv.ParseFloat(string(n), 64)

	if err != nil {
		return math.NaN()
	}

	return ret
}

func (n String) Bool() bool {
	return len(n) > 0
}
