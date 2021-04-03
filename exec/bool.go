package exec

type Bool bool

func (b Bool) String() string {
	if b {
		return "true"
	}

	return "false"
}

func (b Bool) Number() float64 {
	if b {
		return 1.0
	}

	return 0.0
}

func (b Bool) Bool() bool {
	return bool(b)
}
