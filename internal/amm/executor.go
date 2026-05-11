package amm

import (
	"fmt"
	"math/big"

	"github.com/white-blue-protocol/wblue/internal/storage"
	bolt "go.etcd.io/bbolt"
)

const BlueBurnRate = 2

func ExecuteSwap(db *storage.DB, from string, tokenID string, amountIn uint64, direction string, minAmountOut uint64) (uint64, error) {
	var result uint64
	err := db.Update(func(btx *bolt.Tx) error {
		var e error
		result, e = ExecuteSwapInTx(btx, from, tokenID, amountIn, direction, minAmountOut)
		return e
	})
	return result, err
}

func ExecuteSwapInTx(btx *bolt.Tx, from string, tokenID string, amountIn uint64, direction string, minAmountOut uint64) (uint64, error) {
	pool, err := storage.GetPoolInTx(btx, tokenID)
	if err != nil {
		return 0, fmt.Errorf("pool not found for token %s", tokenID)
	}

	oldK, ok := new(big.Int).SetString(pool.K, 10)
	if !ok {
		return 0, fmt.Errorf("invalid pool K value: %s", pool.K)
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
		if !amountOut.IsUint64() {
			return 0, fmt.Errorf("amountOut exceeds uint64 range")
		}
		if !fee.IsUint64() {
			return 0, fmt.Errorf("fee exceeds uint64 range")
		}
		pool.WhiteReserve += amountIn - fee.Uint64()
		pool.BlueReserve -= amountOut.Uint64()
		pool.TotalFeeBurned += fee.Uint64()

		burnAmount := amountOut.Uint64() * BlueBurnRate / 100
		userReceives := amountOut.Uint64() - burnAmount

		account := storage.GetOrCreateAccountInTx(btx, from)
		account.WhiteBalance -= amountIn
		account.BlueBalances[tokenID] += userReceives
		if err := storage.SaveAccountInTx(btx, account); err != nil {
			return 0, err
		}

		if burnAmount > 0 {
			blueState, err := storage.GetBlueCoinStateInTx(btx, tokenID)
			if err == nil {
				blueState.Burned += burnAmount
				storage.SaveBlueCoinStateInTx(btx, blueState)
			}
		}

		amountOut = new(big.Int).SetUint64(userReceives)

	case "blue-to-white":
		amountOut, fee = GetAmountOut(
			new(big.Int).SetUint64(amountIn),
			new(big.Int).SetUint64(pool.BlueReserve),
			new(big.Int).SetUint64(pool.WhiteReserve),
		)
		if !amountOut.IsUint64() {
			return 0, fmt.Errorf("amountOut exceeds uint64 range")
		}
		if !fee.IsUint64() {
			return 0, fmt.Errorf("fee exceeds uint64 range")
		}
		pool.BlueReserve += amountIn - fee.Uint64()
		pool.WhiteReserve -= amountOut.Uint64()
		pool.TotalFeeBurned += fee.Uint64()

		burnAmount := amountIn * BlueBurnRate / 100
		effectiveIn := amountIn - burnAmount

		account := storage.GetOrCreateAccountInTx(btx, from)
		account.BlueBalances[tokenID] -= amountIn
		account.WhiteBalance += amountOut.Uint64()
		if err := storage.SaveAccountInTx(btx, account); err != nil {
			return 0, err
		}

		if burnAmount > 0 {
			blueState, err := storage.GetBlueCoinStateInTx(btx, tokenID)
			if err == nil {
				blueState.Burned += burnAmount
				storage.SaveBlueCoinStateInTx(btx, blueState)
			}
		}
		_ = effectiveIn

	default:
		return 0, fmt.Errorf("invalid direction: %s", direction)
	}

	newK := new(big.Int).Mul(
		new(big.Int).SetUint64(pool.WhiteReserve),
		new(big.Int).SetUint64(pool.BlueReserve),
	)
	if newK.Cmp(oldK) < 0 {
		return 0, fmt.Errorf("constant product invariant violated: newK=%s < oldK=%s", newK, oldK)
	}
	pool.K = newK.String()

	if err := storage.SavePoolInTx(btx, pool); err != nil {
		return 0, err
	}

	out := amountOut.Uint64()
	if minAmountOut > 0 && out < minAmountOut {
		return 0, fmt.Errorf("slippage exceeded: got %d, min %d", out, minAmountOut)
	}

	return out, nil
}
