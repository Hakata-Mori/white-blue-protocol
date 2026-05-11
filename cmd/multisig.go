package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/white-blue-protocol/wblue/internal/types"
)

var multisigCmd = &cobra.Command{
	Use:   "multisig",
	Short: "Multi-signature wallet operations",
}

var multisigRegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new multisig wallet",
	RunE: func(cmd *cobra.Command, args []string) error {
		fromAddr, _ := cmd.Flags().GetString("from")
		ownersStr, _ := cmd.Flags().GetString("owners")
		threshold, _ := cmd.Flags().GetUint8("threshold")

		if fromAddr == "" || ownersStr == "" || threshold == 0 {
			return fmt.Errorf("--from, --owners, and --threshold are required")
		}

		owners := strings.Split(ownersStr, ",")
		if len(owners) < 2 {
			return fmt.Errorf("at least 2 owners required")
		}

		kp, err := loadWalletByAddress(fromAddr)
		if err != nil {
			return err
		}

		nonce, err := getAccountNonce(kp.Address)
		if err != nil {
			return err
		}

		payload, _ := json.Marshal(struct {
			Owners    []string `json:"owners"`
			Threshold uint8    `json:"threshold"`
		}{Owners: owners, Threshold: threshold})

		tx := types.Transaction{
			Type:      types.TxMultiSigRegister,
			From:      kp.Address,
			Amount:    0,
			Nonce:     nonce + 1,
			PublicKey: kp.PublicKey,
			Payload:   payload,
			Timestamp: time.Now().Unix(),
		}

		if err := signAndSubmit(&tx, kp.PrivateKey); err != nil {
			return err
		}

		msAddr := types.DeriveMultiSigAddress(owners, threshold)
		fmt.Printf("Multisig registered: %s\n", msAddr)
		fmt.Printf("Owners: %s\n", ownersStr)
		fmt.Printf("Threshold: %d-of-%d\n", threshold, len(owners))
		return nil
	},
}

var multisigProposeCmd = &cobra.Command{
	Use:   "propose",
	Short: "Propose a transfer from multisig wallet",
	RunE: func(cmd *cobra.Command, args []string) error {
		fromAddr, _ := cmd.Flags().GetString("from")
		msAddr, _ := cmd.Flags().GetString("multisig")
		toAddr, _ := cmd.Flags().GetString("to")
		amount, _ := cmd.Flags().GetFloat64("amount")

		if fromAddr == "" || msAddr == "" || toAddr == "" || amount <= 0 {
			return fmt.Errorf("--from, --multisig, --to, and --amount are required")
		}

		kp, err := loadWalletByAddress(fromAddr)
		if err != nil {
			return err
		}

		nonce, err := getAccountNonce(kp.Address)
		if err != nil {
			return err
		}

		innerTx := types.Transaction{
			Type:   types.TxTransferWhite,
			From:   msAddr,
			To:     toAddr,
			Amount: uint64(amount * 1_000_000),
		}

		payload, _ := json.Marshal(struct {
			MultiSigAddr string            `json:"multiSigAddr"`
			TxPayload    types.Transaction `json:"txPayload"`
		}{MultiSigAddr: msAddr, TxPayload: innerTx})

		tx := types.Transaction{
			Type:      types.TxMultiSigPropose,
			From:      kp.Address,
			Amount:    0,
			Nonce:     nonce + 1,
			PublicKey: kp.PublicKey,
			Payload:   payload,
			Timestamp: time.Now().Unix(),
		}

		if err := signAndSubmit(&tx, kp.PrivateKey); err != nil {
			return err
		}

		fmt.Printf("Proposal submitted: %s\n", tx.Hash[:16])
		fmt.Printf("Transfer %.6f WC from %s to %s\n", amount, msAddr[:14]+"...", toAddr[:14]+"...")
		return nil
	},
}

var multisigApproveCmd = &cobra.Command{
	Use:   "approve",
	Short: "Approve a multisig proposal",
	RunE: func(cmd *cobra.Command, args []string) error {
		fromAddr, _ := cmd.Flags().GetString("from")
		proposalID, _ := cmd.Flags().GetString("proposal")

		if fromAddr == "" || proposalID == "" {
			return fmt.Errorf("--from and --proposal are required")
		}

		kp, err := loadWalletByAddress(fromAddr)
		if err != nil {
			return err
		}

		nonce, err := getAccountNonce(kp.Address)
		if err != nil {
			return err
		}

		payload, _ := json.Marshal(struct {
			ProposalID string `json:"proposalId"`
		}{ProposalID: proposalID})

		tx := types.Transaction{
			Type:      types.TxMultiSigApprove,
			From:      kp.Address,
			Amount:    0,
			Nonce:     nonce + 1,
			PublicKey: kp.PublicKey,
			Payload:   payload,
			Timestamp: time.Now().Unix(),
		}

		if err := signAndSubmit(&tx, kp.PrivateKey); err != nil {
			return err
		}

		fmt.Printf("Approval submitted for proposal %s\n", proposalID[:16])
		return nil
	},
}

var multisigInfoCmd = &cobra.Command{
	Use:   "info [address]",
	Short: "Show multisig account info",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := http.Get(fmt.Sprintf("%s/api/v1/multisig/%s", apiURL, args[0]))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return fmt.Errorf("multisig not found")
		}

		var ms types.MultiSigAccount
		json.NewDecoder(resp.Body).Decode(&ms)

		fmt.Printf("Address:   %s\n", ms.Address)
		fmt.Printf("Threshold: %d-of-%d\n", ms.Threshold, len(ms.Owners))
		fmt.Println("Owners:")
		for _, o := range ms.Owners {
			fmt.Printf("  %s\n", o)
		}
		return nil
	},
}

func init() {
	multisigRegisterCmd.Flags().String("from", "", "Sender address")
	multisigRegisterCmd.Flags().String("owners", "", "Comma-separated owner addresses")
	multisigRegisterCmd.Flags().Uint8("threshold", 0, "Required approvals")

	multisigProposeCmd.Flags().String("from", "", "Proposer address (must be owner)")
	multisigProposeCmd.Flags().String("multisig", "", "Multisig wallet address")
	multisigProposeCmd.Flags().String("to", "", "Transfer recipient")
	multisigProposeCmd.Flags().Float64("amount", 0, "Transfer amount (WC)")

	multisigApproveCmd.Flags().String("from", "", "Approver address (must be owner)")
	multisigApproveCmd.Flags().String("proposal", "", "Proposal ID to approve")

	multisigCmd.AddCommand(multisigRegisterCmd)
	multisigCmd.AddCommand(multisigProposeCmd)
	multisigCmd.AddCommand(multisigApproveCmd)
	multisigCmd.AddCommand(multisigInfoCmd)
	rootCmd.AddCommand(multisigCmd)
}
