package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
	wcrypto "github.com/white-blue-protocol/wblue/internal/crypto"
	"github.com/white-blue-protocol/wblue/internal/node"
)

var (
	dataDir   string
	validator string
)

var rootCmd = &cobra.Command{
	Use:   "wblue",
	Short: "White & Blue Protocol node",
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the node",
	RunE: func(cmd *cobra.Command, args []string) error {
		if validator == "" {
			kp, err := wcrypto.GenerateKeyPair()
			if err != nil {
				return err
			}
			validator = kp.Address

			walletDir := filepath.Join(dataDir, "wallets")
			os.MkdirAll(walletDir, 0755)
			walletFile := filepath.Join(walletDir, kp.Address+".json")
			data := fmt.Sprintf(`{"privateKey":"%s","publicKey":"%s","address":"%s"}`, kp.PrivateKey, kp.PublicKey, kp.Address)
			os.WriteFile(walletFile, []byte(data), 0600)

			fmt.Printf("Generated validator address: %s\n", kp.Address)
			fmt.Printf("Private key: %s\n", kp.PrivateKey)
			fmt.Println("(Saved to wallet file)")
			fmt.Println()
		}

		n, err := node.NewNode(dataDir, validator)
		if err != nil {
			return err
		}

		if err := n.Start(); err != nil {
			return err
		}

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		fmt.Println("\nShutting down...")
		n.Stop()
		return nil
	},
}

func init() {
	home, _ := os.UserHomeDir()
	defaultDataDir := filepath.Join(home, ".wblue", "data")

	rootCmd.PersistentFlags().StringVar(&dataDir, "data-dir", defaultDataDir, "Data directory")
	startCmd.Flags().StringVar(&validator, "validator", "", "Validator address (auto-generated if empty)")
	rootCmd.AddCommand(startCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
