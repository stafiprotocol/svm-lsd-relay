package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"svm-lsd-relay/pkg/config"
	"svm-lsd-relay/pkg/utils"

	"github.com/davecgh/go-spew/spew"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/mr-tron/base58"
	"github.com/spf13/cobra"
	"golang.org/x/time/rate"
)

func buildUpgradeProgramTx() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "build-upgrade-tx",
		Short: "",

		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := cmd.Flags().GetString(flagConfigPath)
			if err != nil {
				return err
			}
			fmt.Printf("config path: %s\n", configPath)

			cfg, err := config.LoadConfig[config.ConfigBuildUpgradeTx](configPath)
			if err != nil {
				return err
			}

			rpcClient := rpc.NewWithCustomRPCClient(rpc.NewWithLimiter(
				cfg.RpcEndpoint,
				rate.Every(time.Second), // time frame
				5,                       // limit of requests per time frame
			))

			feePayer := solana.MustPublicKeyFromBase58(cfg.FeePayerAccount)
			lsdProgramID := solana.MustPublicKeyFromBase58(cfg.LsdProgramID)
			bufferAccount := solana.MustPublicKeyFromBase58(cfg.BufferAccount)
			upgradeAuthority := solana.MustPublicKeyFromBase58(cfg.UpgradeAuthority)
			spillAddress, _, err := solana.FindProgramAddress([][]byte{
				bufferAccount[:],
				[]byte("spill"),
			}, solana.BPFLoaderUpgradeableProgramID)
			if err != nil {
				return err
			}
			upgradeInstruction, err := utils.NewUpgradeInstruction(
				lsdProgramID,
				bufferAccount,
				upgradeAuthority,
				spillAddress,
			)
			if err != nil {
				return err
			}

			bts, _ := json.MarshalIndent(cfg, "", "  ")
			fmt.Printf("Config: \n%s\n", string(bts))
		Out:
			for {
				fmt.Println("\ncheck account info, then press (y/n) to continue:")
				var input string
				fmt.Scanln(&input)
				switch input {
				case "y":
					break Out
				case "n":
					return nil
				default:
					fmt.Println("press `y` or `n`")
					continue
				}
			}
			instructions := []solana.Instruction{upgradeInstruction}

			latestBlockHashRes, err := rpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentConfirmed)
			if err != nil {
				fmt.Printf("get recent block hash error, err: %v\n", err)
			}
			recentBlockHash := latestBlockHashRes.Value.Blockhash
			tx, err := solana.NewTransaction(
				instructions,
				recentBlockHash,
				solana.TransactionPayer(feePayer))
			if err != nil {
				return fmt.Errorf("NewTransaction failed, err: %s, tx: %s", err.Error(), tx.String())
			}

			tx.Message.SetVersion(solana.MessageVersionLegacy)
			bytes, err := tx.Message.MarshalBinary()
			if err != nil {
				return fmt.Errorf("fail to marshal tx.Message: %w", err)
			}
			spew.Dump(bytes)
			fmt.Println("tx:")
			fmt.Println(base58.Encode(bytes))
			return nil
		},
	}
	cmd.Flags().String(flagConfigPath, defaultConfigPath, "Config file path")
	return cmd
}
