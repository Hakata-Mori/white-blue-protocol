package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	wcrypto "github.com/white-blue-protocol/wblue/internal/crypto"
	"github.com/white-blue-protocol/wblue/internal/types"
)

var walletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "Wallet operations",
}

var walletCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new wallet",
	RunE: func(cmd *cobra.Command, args []string) error {
		kp, err := wcrypto.GenerateKeyPair()
		if err != nil {
			return err
		}

		walletDir := filepath.Join(dataDir, "wallets")
		if err := os.MkdirAll(walletDir, 0755); err != nil {
			return err
		}

		walletFile := filepath.Join(walletDir, kp.Address+".json")
		data, _ := json.MarshalIndent(kp, "", "  ")
		if err := os.WriteFile(walletFile, data, 0600); err != nil {
			return err
		}

		fmt.Printf("Address:     %s\n", kp.Address)
		fmt.Printf("Public Key:  %s\n", kp.PublicKey)
		fmt.Printf("Private Key: %s\n", kp.PrivateKey)
		fmt.Printf("Saved to:    %s\n", walletFile)
		return nil
	},
}

var walletListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all wallets",
	RunE: func(cmd *cobra.Command, args []string) error {
		walletDir := filepath.Join(dataDir, "wallets")
		entries, err := os.ReadDir(walletDir)
		if err != nil {
			fmt.Println("No wallets found.")
			return nil
		}

		for _, entry := range entries {
			if filepath.Ext(entry.Name()) == ".json" {
				data, err := os.ReadFile(filepath.Join(walletDir, entry.Name()))
				if err != nil {
					continue
				}
				var kp types.KeyPair
				json.Unmarshal(data, &kp)
				fmt.Printf("%s\n", kp.Address)
			}
		}
		return nil
	},
}

var walletInfoCmd = &cobra.Command{
	Use:   "info [address]",
	Short: "Show wallet balance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		address := args[0]

		resp, err := http.Get(fmt.Sprintf("http://localhost:8080/api/v1/wallet/%s", address))
		if err != nil {
			return fmt.Errorf("node not running? %w", err)
		}
		defer resp.Body.Close()

		var account types.Account
		if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
			return err
		}

		fmt.Printf("Address: %s\n", account.Address)
		fmt.Printf("White Balance: %.6f WC\n", float64(account.WhiteBalance)/1_000_000)
		fmt.Printf("Nonce: %d\n", account.Nonce)

		if len(account.BlueBalances) > 0 {
			fmt.Println("Blue Balances:")
			for tokenID, balance := range account.BlueBalances {
				fmt.Printf("  %s: %.6f\n", tokenID, float64(balance)/1_000_000)
			}
		}
		return nil
	},
}

func init() {
	walletCmd.AddCommand(walletCreateCmd)
	walletCmd.AddCommand(walletListCmd)
	walletCmd.AddCommand(walletInfoCmd)
	rootCmd.AddCommand(walletCmd)
}
