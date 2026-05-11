package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/white-blue-protocol/wblue/internal/types"
)

var chainCmd = &cobra.Command{
	Use:   "chain",
	Short: "Chain queries",
}

var chainStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show chain status",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := http.Get(fmt.Sprintf("%s/api/v1/chain/status", apiURL))
		if err != nil {
			return fmt.Errorf("node not running? %w", err)
		}
		defer resp.Body.Close()

		var status struct {
			Height      uint64 `json:"height"`
			TotalMinted uint64 `json:"totalMinted"`
			ChainID     string `json:"chainId"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
			return err
		}

		fmt.Printf("Chain ID:      %s\n", status.ChainID)
		fmt.Printf("Height:        %d\n", status.Height)
		fmt.Printf("Total Minted:  %.6f WC\n", float64(status.TotalMinted)/1_000_000)
		fmt.Printf("Max Supply:    1,000,000,000 WC\n")
		return nil
	},
}

var chainTxCmd = &cobra.Command{
	Use:   "tx [hash]",
	Short: "Query transaction status",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		txHash := args[0]

		resp, err := http.Get(fmt.Sprintf("%s/api/v1/tx/%s", apiURL, txHash))
		if err != nil {
			return fmt.Errorf("node not running? %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == 404 {
			fmt.Println("Transaction not found")
			return nil
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return err
		}

		if status, ok := result["status"].(string); ok && status == "pending" {
			fmt.Println("Status: pending (in mempool)")
			return nil
		}

		var receipt types.TxReceipt
		data, _ := json.Marshal(result)
		json.Unmarshal(data, &receipt)

		fmt.Printf("Tx Hash:     %s\n", receipt.TxHash)
		fmt.Printf("Status:      %s\n", receipt.Status)
		fmt.Printf("Block:       %d\n", receipt.BlockHeight)
		fmt.Printf("Block Hash:  %s\n", receipt.BlockHash)
		if receipt.Error != "" {
			fmt.Printf("Error:       %s\n", receipt.Error)
		}
		return nil
	},
}

func init() {
	chainCmd.AddCommand(chainStatusCmd)
	chainCmd.AddCommand(chainTxCmd)
	rootCmd.AddCommand(chainCmd)
}
