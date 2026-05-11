package amm

import (
	"math/big"
	"testing"
)

func TestGetAmountOutBasic(t *testing.T) {
	amountIn := big.NewInt(100_000_000)
	reserveIn := big.NewInt(1_000_000_000)
	reserveOut := big.NewInt(500_000_000_000)

	out, fee := GetAmountOut(amountIn, reserveIn, reserveOut)

	if out.Sign() <= 0 {
		t.Fatal("amountOut should be positive")
	}
	if fee.Sign() <= 0 {
		t.Fatal("fee should be positive")
	}
	if out.Cmp(reserveOut) >= 0 {
		t.Fatal("amountOut should be less than reserveOut")
	}
}

func TestGetAmountOutFeeCalculation(t *testing.T) {
	amountIn := big.NewInt(10_000)
	reserveIn := big.NewInt(1_000_000)
	reserveOut := big.NewInt(1_000_000)

	_, fee := GetAmountOut(amountIn, reserveIn, reserveOut)

	expectedFee := big.NewInt(10)
	if fee.Cmp(expectedFee) != 0 {
		t.Fatalf("fee should be %s (0.1%%), got %s", expectedFee, fee)
	}
}

func TestGetAmountOutMinFee(t *testing.T) {
	amountIn := big.NewInt(500)
	reserveIn := big.NewInt(1_000_000)
	reserveOut := big.NewInt(1_000_000)

	_, fee := GetAmountOut(amountIn, reserveIn, reserveOut)

	if fee.Cmp(big.NewInt(1)) != 0 {
		t.Fatalf("fee should be minimum 1 for small amounts, got %s", fee)
	}
}

func TestGetAmountOutZeroInput(t *testing.T) {
	out, fee := GetAmountOut(big.NewInt(0), big.NewInt(1000), big.NewInt(1000))
	if out.Sign() != 0 || fee.Sign() != 0 {
		t.Fatal("zero amountIn should return zero out and fee")
	}
}

func TestGetAmountOutNegativeInput(t *testing.T) {
	out, fee := GetAmountOut(big.NewInt(-1), big.NewInt(1000), big.NewInt(1000))
	if out.Sign() != 0 || fee.Sign() != 0 {
		t.Fatal("negative amountIn should return zero")
	}
}

func TestGetAmountOutZeroReserve(t *testing.T) {
	out, _ := GetAmountOut(big.NewInt(100), big.NewInt(0), big.NewInt(1000))
	if out.Sign() != 0 {
		t.Fatal("zero reserveIn should return zero")
	}

	out2, _ := GetAmountOut(big.NewInt(100), big.NewInt(1000), big.NewInt(0))
	if out2.Sign() != 0 {
		t.Fatal("zero reserveOut should return zero")
	}
}

func TestGetAmountOutConstantProduct(t *testing.T) {
	reserveIn := big.NewInt(1_000_000_000)
	reserveOut := big.NewInt(1_000_000_000)
	amountIn := big.NewInt(100_000_000)

	out, fee := GetAmountOut(amountIn, reserveIn, reserveOut)

	newReserveIn := new(big.Int).Add(reserveIn, new(big.Int).Sub(amountIn, fee))
	newReserveOut := new(big.Int).Sub(reserveOut, out)

	oldK := new(big.Int).Mul(reserveIn, reserveOut)
	newK := new(big.Int).Mul(newReserveIn, newReserveOut)

	if newK.Cmp(oldK) < 0 {
		t.Fatalf("K invariant violated: oldK=%s newK=%s", oldK, newK)
	}
}

func TestGetAmountOutLargeValues(t *testing.T) {
	amountIn := new(big.Int).SetUint64(^uint64(0) / 2)
	reserveIn := new(big.Int).SetUint64(^uint64(0) / 2)
	reserveOut := new(big.Int).SetUint64(^uint64(0) / 2)

	out, fee := GetAmountOut(amountIn, reserveIn, reserveOut)
	if out.Sign() <= 0 {
		t.Fatal("large values should still produce output")
	}
	if fee.Sign() <= 0 {
		t.Fatal("large values should still produce fee")
	}
}

func TestBlueBurnRateConstant(t *testing.T) {
	if BlueBurnRate != 2 {
		t.Fatalf("burn rate should be 2%%, got %d%%", BlueBurnRate)
	}
}
