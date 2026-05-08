package amm

import (
	"fmt"
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

func SpotPrice(whiteReserve, blueReserve uint64) string {
	if blueReserve == 0 {
		return "0"
	}
	w := new(big.Float).SetUint64(whiteReserve)
	b := new(big.Float).SetUint64(blueReserve)
	price := new(big.Float).Quo(w, b)
	return fmt.Sprintf("%.6f", price)
}
