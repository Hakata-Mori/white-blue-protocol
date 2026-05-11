package safemath

import "fmt"

func SafeAdd(a, b uint64) (uint64, error) {
	if a > ^uint64(0)-b {
		return 0, fmt.Errorf("uint64 overflow: %d + %d", a, b)
	}
	return a + b, nil
}

func SafeSub(a, b uint64) (uint64, error) {
	if a < b {
		return 0, fmt.Errorf("uint64 underflow: %d - %d", a, b)
	}
	return a - b, nil
}

func SafeMul(a, b uint64) (uint64, error) {
	if a == 0 || b == 0 {
		return 0, nil
	}
	if a > ^uint64(0)/b {
		return 0, fmt.Errorf("uint64 overflow: %d * %d", a, b)
	}
	return a * b, nil
}

func MustAdd(a, b uint64) uint64 {
	r, err := SafeAdd(a, b)
	if err != nil {
		panic(err)
	}
	return r
}

func MustSub(a, b uint64) uint64 {
	r, err := SafeSub(a, b)
	if err != nil {
		panic(err)
	}
	return r
}
