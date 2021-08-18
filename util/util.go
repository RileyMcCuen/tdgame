package util

func Check(e error) {
	if e != nil {
		panic(e.Error())
	}
}

func CheckNil(i interface{}) {
	if i == nil {
		panic("object is nil and should not be")
	}
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Square(a int) int {
	return a * a
}
