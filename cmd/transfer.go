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
	transferFrom   string
	transferTo     string
	transferAmount float64
	transferToken  string
)

var transferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "Transfer coins",
}

var transferWhiteCmd = &cobra.Command{
	Use:   "white",
	Short: "Transfer White Coins",
	RunE: func(cmd *cobra.Command, args []string) error {
		return doTransfer(types.TxTransferWhite, "")
	},
}

var transferBlueCmd = &cobra.Command{
	Use:   "blue",
	Short: "Transfer Blue Coins",
	RunE: func(cmd *cobra.Command, args []string) error {
		if transferToken == "" {
			return fmt.Errorf("--token is required")
		}
		return doTransfer(types.TxTransferBlue, transferToken)
	},
}

func doTransfer(txType types.TxType, tokenID string) error {
	kp, err := loadWalletByAddress(transferFrom)
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

	amount := uint64(transferAmount * 1_000_000)

	tx := types.Transaction{
		Type:      txType,
		From:      kp.Address,
		To:        transferTo,
		Amount:    amount,
		TokenID:   tokenID,
		Fee:       types.CalcFee(amount),
		Nonce:     account.Nonce + 1,
		PublicKey: kp.PublicKey,
		Timestamp: time.Now().Unix(),
	}

	txData, _ := json.Marshal(tx)
	tx.Hash = wcrypto.SHA256Hex(txData)

	sig, err := wcrypto.Sign(kp.PrivateKey, txData)
	if err != nil {
		return fmt.Errorf("sign: %w", err)
	}
	tx.Signature = sig

	txJSON, _ := json.Marshal(tx)
	submitResp, err := http.Post(fmt.Sprintf("%s/api/v1/tx/submit", apiURL), "application/json", bytes.NewReader(txJSON))
	if err != nil {
		return fmt.Errorf("submit tx: %w", err)
	}
	defer submitResp.Body.Close()

	fmt.Printf("Transaction submitted: %s\n", tx.Hash[:16]+"...")
	fmt.Printf("From: %s\n", tx.From)
	fmt.Printf("To: %s\n", tx.To)
	fmt.Printf("Amount: %.6f\n", transferAmount)
	fmt.Println("(Will be confirmed in next block)")
	return nil
}

func init() {
	transferCmd.PersistentFlags().StringVar(&transferFrom, "from", "", "Sender address")
	transferCmd.PersistentFlags().StringVar(&transferTo, "to", "", "Recipient address")
	transferCmd.PersistentFlags().Float64Var(&transferAmount, "amount", 0, "Amount to transfer")
	transferBlueCmd.Flags().StringVar(&transferToken, "token", "", "Blue coin token ID")

	transferCmd.AddCommand(transferWhiteCmd)
	transferCmd.AddCommand(transferBlueCmd)
	rootCmd.AddCommand(transferCmd)
}
