package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/white-blue-protocol/wblue/internal/config"
	wcrypto "github.com/white-blue-protocol/wblue/internal/crypto"
	"github.com/white-blue-protocol/wblue/internal/keystore"
	"github.com/white-blue-protocol/wblue/internal/node"
	"github.com/white-blue-protocol/wblue/internal/types"
	"github.com/white-blue-protocol/wblue/internal/version"
)

var (
	dataDir     string
	validator   string
	apiPort     int
	apiURL      string
	noValidator bool
	p2pPort     int
	seeds       []string
	noP2P       bool
	enableMDNS  bool
	valPassword string
	devMode     bool
	chainID     string
)

var rootCmd = &cobra.Command{
	Use:   "wblue",
	Short: "White & Blue Protocol node",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.String())
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the node",
	RunE: func(cmd *cobra.Command, args []string) error {
		if devMode {
			types.SetDevMode()
			fmt.Println("*** DEV MODE: accelerated timing ***")
		}

		fileCfg := config.LoadOrDefault(dataDir)

		effectiveChainID := fileCfg.ChainID
		if cmd.Flags().Changed("chain-id") {
			effectiveChainID = chainID
		}
		if effectiveChainID == "" {
			effectiveChainID = config.DefaultConfig.ChainID
		}

		isValidator := !noValidator

		var validatorKey, validatorPub string

		if isValidator {
			if validator == "" {
				kp, err := wcrypto.GenerateKeyPair()
				if err != nil {
					return err
				}
				validator = kp.Address
				validatorKey = kp.PrivateKey
				validatorPub = kp.PublicKey

				password := valPassword
				if password == "" {
					password = os.Getenv("WBLUE_VALIDATOR_PASSWORD")
				}
				if password == "" {
					var err error
					password, err = readPassword("Set password for validator wallet: ")
					if err != nil {
						return err
					}
				}

				walletDir := filepath.Join(dataDir, "wallets")
				os.MkdirAll(walletDir, 0755)
				walletFile := filepath.Join(walletDir, kp.Address+".json")

				ks, err := keystore.Encrypt(kp.PrivateKey, kp.PublicKey, kp.Address, password)
				if err != nil {
					return err
				}
				if err := keystore.Save(ks, walletFile); err != nil {
					return err
				}

				fmt.Printf("Generated validator address: %s\n", kp.Address)
				fmt.Println("(Saved to encrypted wallet file)")
				fmt.Println()
			} else {
				kp, err := loadWalletByAddress(validator)
				if err != nil {
					return fmt.Errorf("load validator wallet: %w", err)
				}
				validatorKey = kp.PrivateKey
				validatorPub = kp.PublicKey
			}
		}

		cfg := node.Config{
			DataDir:      dataDir,
			Validator:    validator,
			ValidatorKey: validatorKey,
			ValidatorPub: validatorPub,
			APIPort:      apiPort,
			IsValidator:  isValidator,
			P2PEnabled:   !noP2P,
			P2PPort:      p2pPort,
			P2PSeeds:     seeds,
			P2PMDNS:      enableMDNS,
			ChainID:      effectiveChainID,
		}

		n, err := node.NewNode(cfg)
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
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "http://localhost:8080", "Node API URL for CLI commands")
	startCmd.Flags().StringVar(&validator, "validator", "", "Validator address (auto-generated if empty)")
	startCmd.Flags().IntVar(&apiPort, "api-port", 8080, "HTTP API port")
	startCmd.Flags().BoolVar(&noValidator, "no-validator", false, "Run as full node without block production")
	startCmd.Flags().IntVar(&p2pPort, "p2p-port", 30303, "P2P listen port")
	startCmd.Flags().StringArrayVar(&seeds, "seeds", nil, "Seed node multiaddrs")
	startCmd.Flags().BoolVar(&noP2P, "no-p2p", false, "Disable P2P networking")
	startCmd.Flags().BoolVar(&enableMDNS, "mdns", true, "Enable mDNS discovery")
	startCmd.Flags().StringVar(&valPassword, "password", "", "Validator wallet password (or use WBLUE_VALIDATOR_PASSWORD env)")
	startCmd.Flags().BoolVar(&devMode, "dev", false, "Dev mode: accelerated block timing for testing")
	startCmd.Flags().StringVar(&chainID, "chain-id", "", "Chain ID (default from config or wblue-mainnet-1)")
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(versionCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
