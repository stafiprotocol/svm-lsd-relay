package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"
)

var (
	appName = "svm-lsd-relay"
)

const (
	flagLogLevel     = "log_level"
	flagConfigPath   = "config"
	flagStakeManager = "stake_manager"
	flagRpcEndPoint  = "rpc_endpoint"
	flagKeystorePath = "keystore_path"
	flagExportTx     = "export"

	defaultKeystorePath = "./keys/solana_keys.json"
	defaultConfigPath   = "./config.toml"
)

// NewRootCmd returns the root command.
func NewRootCmd() *cobra.Command {
	// RootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:   appName,
		Short: "svm-lsd-relay",
	}

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, segments []string) error {
		return nil
	}

	rootCmd.AddCommand(
		keysCmd(),
		stakeManagerCmd(),
		startCmd(),
		versionCmd(),
		buildUpgradeProgramTx(),
		buildTransferTokenCmd(),
	)

	return rootCmd
}

func keysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Manage keystore",
	}

	cmd.AddCommand(
		vaultImportCmd(),
		vaultGenCmd(),
		vaultExportCmd(),
		vaultListCmd(),
	)
	return cmd
}

func stakeManagerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stake-manager",
		Short: "Stake manager operation",
	}

	cmd.AddCommand(
		stakeManagerInitCmd(),
		stakeManagerSetCmd(),
		stakeManagerDetailCmd(),
		stakeManagerTransferAdminCmd(),
		createTokenMetadataCmd(),
	)
	return cmd
}

func Execute() {

	rootCmd := NewRootCmd()
	rootCmd.SilenceUsage = true
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	ctx := context.Background()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
