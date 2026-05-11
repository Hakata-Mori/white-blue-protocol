package amm

import (
	"math/big"
)

func GetAmountOut(amountIn, reserveIn, reserveOut *big.Int) (amountOut *big.Int, fee *big.Int) {
	if amountIn.Sign() <= 0 || reserveIn.Sign() <= 0 || reserveOut.Sign() <= 0 {
		return big.NewInt(0), big.NewInt(0)
	}

	fee = new(big.Int).Div(amountIn, big.NewInt(1000))
	if fee.Sign() == 0 {
		fee = big.NewInt(1)
	}

	amountInAfterFee := new(big.Int).Sub(amountIn, fee)

	numerator := new(big.Int).Mul(amountInAfterFee, reserveOut)
	denominator := new(big.Int).Add(reserveIn, amountInAfterFee)

	amountOut = new(big.Int).Div(numerator, denominator)
	return amountOut, fee
}
