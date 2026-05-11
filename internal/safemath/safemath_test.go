package safemath

import (
	"math"
	"testing"
)

func TestSafeAddNormal(t *testing.T) {
	r, err := SafeAdd(100, 200)
	if err != nil {
		t.Fatal(err)
	}
	if r != 300 {
		t.Fatalf("expected 300, got %d", r)
	}
}

func TestSafeAddOverflow(t *testing.T) {
	_, err := SafeAdd(math.MaxUint64, 1)
	if err == nil {
		t.Fatal("expected overflow error")
	}
}

func TestSafeAddBoundary(t *testing.T) {
	r, err := SafeAdd(math.MaxUint64, 0)
	if err != nil {
		t.Fatal(err)
	}
	if r != math.MaxUint64 {
		t.Fatalf("expected MaxUint64, got %d", r)
	}
}

func TestSafeSubNormal(t *testing.T) {
	r, err := SafeSub(300, 100)
	if err != nil {
		t.Fatal(err)
	}
	if r != 200 {
		t.Fatalf("expected 200, got %d", r)
	}
}

func TestSafeSubUnderflow(t *testing.T) {
	_, err := SafeSub(0, 1)
	if err == nil {
		t.Fatal("expected underflow error")
	}
}

func TestSafeSubZero(t *testing.T) {
	r, err := SafeSub(100, 100)
	if err != nil {
		t.Fatal(err)
	}
	if r != 0 {
		t.Fatalf("expected 0, got %d", r)
	}
}

func TestSafeMulNormal(t *testing.T) {
	r, err := SafeMul(1000, 2000)
	if err != nil {
		t.Fatal(err)
	}
	if r != 2_000_000 {
		t.Fatalf("expected 2000000, got %d", r)
	}
}

func TestSafeMulOverflow(t *testing.T) {
	_, err := SafeMul(math.MaxUint64, 2)
	if err == nil {
		t.Fatal("expected overflow error")
	}
}

func TestSafeMulZero(t *testing.T) {
	r, err := SafeMul(0, math.MaxUint64)
	if err != nil {
		t.Fatal(err)
	}
	if r != 0 {
		t.Fatalf("expected 0, got %d", r)
	}
}

func TestSafeMulOne(t *testing.T) {
	r, err := SafeMul(math.MaxUint64, 1)
	if err != nil {
		t.Fatal(err)
	}
	if r != math.MaxUint64 {
		t.Fatalf("expected MaxUint64, got %d", r)
	}
}

func TestSafeAddFeeOverflowAttack(t *testing.T) {
	amount := uint64(math.MaxUint64)
	fee := uint64(1)
	_, err := SafeAdd(amount, fee)
	if err == nil {
		t.Fatal("attacker can bypass balance check via amount+fee overflow")
	}
}

func TestMustAddPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	MustAdd(math.MaxUint64, 1)
}

func TestMustSubPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	MustSub(0, 1)
}
