package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	wcrypto "github.com/white-blue-protocol/wblue/internal/crypto"
	"github.com/white-blue-protocol/wblue/internal/token"
	"github.com/white-blue-protocol/wblue/internal/types"
)

var (
	bcFrom           string
	bcName           string
	bcSymbol         string
	bcPoolRatio      uint8
	bcTeamRatio      uint8
	bcInitWhite      float64
	bcReleaseMonthly float64
	bcMultiSig       string
	bcURL            string
)

var bluecoinCmd = &cobra.Command{
	Use:   "bluecoin",
	Short: "Blue Coin operations",
}

var bluecoinDeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a new Blue Coin",
	RunE: func(cmd *cobra.Command, args []string) error {
		kp, err := loadWalletByAddress(bcFrom)
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

		params := token.DeployParams{
			Name:           bcName,
			Symbol:         bcSymbol,
			PoolRatio:      bcPoolRatio,
			TeamRatio:      bcTeamRatio,
			InitWhite:      uint64(bcInitWhite * 1_000_000),
			ReleaseMonthly: uint64(bcReleaseMonthly * 1_000_000),
			MultiSigAddr:   bcMultiSig,
		}
		if bcURL != "" {
			params.SourceURLs = []string{bcURL}
		}

		payload, _ := json.Marshal(params)

		tx := types.Transaction{
			Type:      types.TxDeployBlue,
			From:      kp.Address,
			To:        "",
			Amount:    0,
			Fee:       types.CalcFee(params.InitWhite),
			Nonce:     account.Nonce + 1,
			Payload:   payload,
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
			return err
		}
		defer submitResp.Body.Close()

		tokenID := token.GenerateTokenID(kp.Address, bcName, account.Nonce+1)

		fmt.Printf("Blue Coin deployed!\n")
		fmt.Printf("Token ID:  %s\n", tokenID)
		fmt.Printf("Name:      %s (%s)\n", bcName, bcSymbol)
		fmt.Printf("Pool:      %d%% | Team: %d%%\n", bcPoolRatio, bcTeamRatio)
		fmt.Printf("Init White: %.6f WC\n", bcInitWhite)
		fmt.Println("(Will be confirmed in next block)")
		return nil
	},
}

var bluecoinListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Blue Coins",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := http.Get(fmt.Sprintf("%s/api/v1/bluecoin", apiURL))
		if err != nil {
			return fmt.Errorf("node not running? %w", err)
		}
		defer resp.Body.Close()

		var configs []types.BlueCoinConfig
		json.NewDecoder(resp.Body).Decode(&configs)

		if len(configs) == 0 {
			fmt.Println("No Blue Coins deployed yet.")
			return nil
		}

		for _, c := range configs {
			fmt.Printf("%s  %s (%s)  Pool: %d%%  Creator: %s\n",
				c.TokenID, c.Name, c.Symbol, c.PoolRatio, c.Creator[:10]+"...")
		}
		return nil
	},
}

var bluecoinInfoCmd = &cobra.Command{
	Use:   "info [tokenId]",
	Short: "Show Blue Coin details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tokenID := args[0]

		resp, err := http.Get(fmt.Sprintf("%s/api/v1/bluecoin/%s", apiURL, tokenID))
		if err != nil {
			return fmt.Errorf("node not running? %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == 404 {
			return fmt.Errorf("token not found")
		}

		var config types.BlueCoinConfig
		json.NewDecoder(resp.Body).Decode(&config)

		fmt.Printf("Token ID:       %s\n", config.TokenID)
		fmt.Printf("Name:           %s (%s)\n", config.Name, config.Symbol)
		fmt.Printf("Creator:        %s\n", config.Creator)
		fmt.Printf("Total Supply:   %.6f\n", float64(config.TotalSupply)/1_000_000)
		fmt.Printf("Pool Ratio:     %d%%\n", config.PoolRatio)
		fmt.Printf("Team Ratio:     %d%%\n", config.TeamRatio)
		fmt.Printf("Init White:     %.6f WC\n", float64(config.InitWhite)/1_000_000)
		fmt.Printf("Monthly Release: %.6f\n", float64(config.ReleaseMonthly)/1_000_000)
		return nil
	},
}

func init() {
	bluecoinDeployCmd.Flags().StringVar(&bcFrom, "from", "", "Deployer address")
	bluecoinDeployCmd.Flags().StringVar(&bcName, "name", "", "Coin name")
	bluecoinDeployCmd.Flags().StringVar(&bcSymbol, "symbol", "", "Coin symbol")
	bluecoinDeployCmd.Flags().Uint8Var(&bcPoolRatio, "pool-ratio", 20, "Pool ratio (%)")
	bluecoinDeployCmd.Flags().Uint8Var(&bcTeamRatio, "team-ratio", 80, "Team ratio (%)")
	bluecoinDeployCmd.Flags().Float64Var(&bcInitWhite, "init-white", 0, "Initial white coins for pool")
	bluecoinDeployCmd.Flags().Float64Var(&bcReleaseMonthly, "release-monthly", 0, "Monthly team release amount")
	bluecoinDeployCmd.Flags().StringVar(&bcMultiSig, "multisig", "", "Team fund recipient address")
	bluecoinDeployCmd.Flags().StringVar(&bcURL, "url", "", "Project URL")

	bluecoinCmd.AddCommand(bluecoinDeployCmd)
	bluecoinCmd.AddCommand(bluecoinListCmd)
	bluecoinCmd.AddCommand(bluecoinInfoCmd)
	rootCmd.AddCommand(bluecoinCmd)
}
