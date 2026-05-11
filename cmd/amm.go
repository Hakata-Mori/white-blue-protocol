package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	wcrypto "github.com/white-blue-protocol/wblue/internal/crypto"
	"github.com/white-blue-protocol/wblue/internal/types"
)

var (
	swapFrom      string
	swapToken     string
	swapDirection string
	swapAmountIn  float64
	swapMinOut    float64
)

var ammCmd = &cobra.Command{
	Use:   "amm",
	Short: "AMM pool operations",
}

var ammSwapCmd = &cobra.Command{
	Use:   "swap",
	Short: "Swap coins via AMM",
	RunE: func(cmd *cobra.Command, args []string) error {
		kp, err := loadWalletByAddress(swapFrom)
		if err != nil {
			return err
		}

		resp, err := http.Get(fmt.Sprintf("%s/api/v1/wallet/%s", apiURL, kp.Address))
		if err != nil {
			return fmt.Errorf("node not running? %w", err)
		}
		defer resp.Body.Close()

		var account types.Account
		json.NewDecoder(resp.Body).Decode(&account)

		var txType types.TxType
		switch swapDirection {
		case "white-to-blue":
			txType = types.TxSwapWhiteBlue
		case "blue-to-white":
			txType = types.TxSwapBlueWhite
		default:
			return fmt.Errorf("direction must be 'white-to-blue' or 'blue-to-white'")
		}

		amount := uint64(swapAmountIn * 1_000_000)

		tx := types.Transaction{
			Type:         txType,
			From:         kp.Address,
			To:           "",
			Amount:       amount,
			TokenID:      swapToken,
			Fee:          0,
			Nonce:        account.Nonce + 1,
			PublicKey:    kp.PublicKey,
			MinAmountOut: uint64(swapMinOut * 1_000_000),
			Timestamp:    time.Now().Unix(),
		}

		txData, _ := json.Marshal(tx)
		tx.Hash = wcrypto.SHA256Hex(txData)

		sig, err := wcrypto.Sign(kp.PrivateKey, txData)
		if err != nil {
			return err
		}
		tx.Signature = sig

		txJSON, _ := json.Marshal(tx)
		submitResp, err := http.Post(fmt.Sprintf("%s/api/v1/tx/submit", apiURL), "application/json", bytes.NewReader(txJSON))
		if err != nil {
			return err
		}
		defer submitResp.Body.Close()

		fmt.Printf("Swap submitted: %s\n", swapDirection)
		fmt.Printf("Token: %s\n", swapToken)
		fmt.Printf("Amount In: %.6f\n", swapAmountIn)
		fmt.Println("(Will be confirmed in next block)")
		return nil
	},
}

var ammPoolInfoCmd = &cobra.Command{
	Use:   "pool-info [tokenId]",
	Short: "Show pool reserves and price",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tokenID := args[0]

		resp, err := http.Get(fmt.Sprintf("%s/api/v1/pool/%s", apiURL, tokenID))
		if err != nil {
			return fmt.Errorf("node not running? %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == 404 {
			return fmt.Errorf("pool not found")
		}

		var pool types.AMMPool
		json.NewDecoder(resp.Body).Decode(&pool)

		whiteRes := float64(pool.WhiteReserve) / 1_000_000
		blueRes := float64(pool.BlueReserve) / 1_000_000
		price := float64(0)
		if pool.BlueReserve > 0 {
			price = float64(pool.WhiteReserve) / float64(pool.BlueReserve)
		}

		fmt.Printf("Token:         %s\n", pool.TokenID)
		fmt.Printf("White Reserve: %.6f WC\n", whiteRes)
		fmt.Printf("Blue Reserve:  %.6f\n", blueRes)
		fmt.Printf("Price:         1 Blue = %.6f WC\n", price)
		fmt.Printf("Fee Burned:    %.6f WC\n", float64(pool.TotalFeeBurned)/1_000_000)
		return nil
	},
}

var ammPriceCmd = &cobra.Command{
	Use:   "price [tokenId]",
	Short: "Show current price",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tokenID := args[0]

		resp, err := http.Get(fmt.Sprintf("%s/api/v1/pool/%s", apiURL, tokenID))
		if err != nil {
			return fmt.Errorf("node not running? %w", err)
		}
		defer resp.Body.Close()

		var pool types.AMMPool
		json.NewDecoder(resp.Body).Decode(&pool)

		price := float64(0)
		if pool.BlueReserve > 0 {
			price = float64(pool.WhiteReserve) / float64(pool.BlueReserve)
		}
		fmt.Printf("1 Blue = %.6f WC\n", price)
		return nil
	},
}

func init() {
	ammSwapCmd.Flags().StringVar(&swapFrom, "from", "", "Sender address")
	ammSwapCmd.Flags().StringVar(&swapToken, "token", "", "Blue coin token ID")
	ammSwapCmd.Flags().StringVar(&swapDirection, "direction", "", "white-to-blue or blue-to-white")
	ammSwapCmd.Flags().Float64Var(&swapAmountIn, "amount-in", 0, "Amount to swap")
	ammSwapCmd.Flags().Float64Var(&swapMinOut, "min-out", 0, "Minimum output amount (slippage protection)")

	ammCmd.AddCommand(ammSwapCmd)
	ammCmd.AddCommand(ammPoolInfoCmd)
	ammCmd.AddCommand(ammPriceCmd)
	rootCmd.AddCommand(ammCmd)
}
