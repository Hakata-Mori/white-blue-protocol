package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
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

func init() {
	chainCmd.AddCommand(chainStatusCmd)
	rootCmd.AddCommand(chainCmd)
}
