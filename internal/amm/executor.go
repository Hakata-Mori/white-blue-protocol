package amm

import (
	"fmt"
	"math/big"

	"github.com/white-blue-protocol/wblue/internal/storage"
)

func ExecuteSwap(db *storage.DB, from string, tokenID string, amountIn uint64, direction string) (uint64, error) {
	pool, err := db.GetPool(tokenID)
	if err != nil {
		return 0, fmt.Errorf("pool not found for token %s", tokenID)
	}

	var amountOut *big.Int
	var fee *big.Int

	switch direction {
	case "white-to-blue":
		amountOut, fee = GetAmountOut(
			new(big.Int).SetUint64(amountIn),
			new(big.Int).SetUint64(pool.WhiteReserve),
			new(big.Int).SetUint64(pool.BlueReserve),
		)
		pool.WhiteReserve += amountIn - fee.Uint64()
		pool.BlueReserve -= amountOut.Uint64()
		pool.TotalFeeBurned += fee.Uint64()

		account := db.GetOrCreateAccount(from)
		account.WhiteBalance -= amountIn
		account.BlueBalances[tokenID] += amountOut.Uint64()
		account.Nonce++
		db.SaveAccount(account)

	case "blue-to-white":
		amountOut, fee = GetAmountOut(
			new(big.Int).SetUint64(amountIn),
			new(big.Int).SetUint64(pool.BlueReserve),
			new(big.Int).SetUint64(pool.WhiteReserve),
		)
		pool.BlueReserve += amountIn - fee.Uint64()
		pool.WhiteReserve -= amountOut.Uint64()
		pool.TotalFeeBurned += fee.Uint64()

		account := db.GetOrCreateAccount(from)
		account.BlueBalances[tokenID] -= amountIn
		account.WhiteBalance += amountOut.Uint64()
		account.Nonce++
		db.SaveAccount(account)

	default:
		return 0, fmt.Errorf("invalid direction: %s", direction)
	}

	newK := new(big.Int).Mul(
		new(big.Int).SetUint64(pool.WhiteReserve),
		new(big.Int).SetUint64(pool.BlueReserve),
	)
	pool.K = newK.String()

	if err := db.SavePool(pool); err != nil {
		return 0, err
	}

	return amountOut.Uint64(), nil
}
