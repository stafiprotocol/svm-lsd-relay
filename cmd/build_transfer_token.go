package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"svm-lsd-relay/pkg/config"
	"svm-lsd-relay/pkg/utils"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/mr-tron/base58"
	"github.com/spf13/cobra"
	"golang.org/x/time/rate"
)

func buildTransferTokenCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "build-transfer-token-tx",
		Short: "build transfer token tx for multisig",

		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := cmd.Flags().GetString(flagConfigPath)
			if err != nil {
				return err
			}
			fmt.Printf("config path: %s\n", configPath)

			cfg, err := config.LoadConfig[config.ConfigBuildTransferTokenTx](configPath)
			if err != nil {
				return err
			}

			rpcClient := rpc.NewWithCustomRPCClient(rpc.NewWithLimiter(
				cfg.RpcEndpoint,
				rate.Every(time.Second), // time frame
				5,                       // limit of requests per time frame
			))

			feePayer := solana.MustPublicKeyFromBase58(cfg.FeePayerAccount)

			owner := solana.MustPublicKeyFromBase58(cfg.FromAddress)
			mint := solana.MustPublicKeyFromBase58(cfg.Mint)
			mintAccountRes, err := rpcClient.GetAccountInfo(context.TODO(), mint)
			if err != nil {
				return fmt.Errorf("couldn't get mint account data: %w", err)
			}

			tokenMint := new(token.Mint)
			if err := tokenMint.Decode(mintAccountRes.Value.Data.GetBinary()); err != nil {
				return fmt.Errorf("unable to retrieve mint account information: %w", err)
			}
			amount := uint64(cfg.Amount * math.Pow(10, float64(tokenMint.Decimals)))

			from, _ := utils.MustFindATA_withCorrectProgram(owner, mint, mintAccountRes.Value)
			to, _ := utils.MustFindATA_withCorrectProgram(solana.MustPublicKeyFromBase58(cfg.ToAddress), mint, mintAccountRes.Value)

			// fmt.Println("owner", owner)
			// fmt.Println("from", from)
			// fmt.Println("to", to)
			// fmt.Println("amount", amount)

			token.SetProgramID(mintAccountRes.Value.Owner)
			transferInstruction := token.NewTransferInstruction(amount, from, to, owner, nil).Build()
			// spew.Dump(transferInstruction)

			bts, _ := json.MarshalIndent(cfg, "", "  ")
			fmt.Printf("Config: \n%s\n", string(bts))

			latestBlockHashRes, err := rpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentConfirmed)
			if err != nil {
				fmt.Printf("get recent block hash error, err: %v\n", err)
			}
			recentBlockHash := latestBlockHashRes.Value.Blockhash
			tx, err := solana.NewTransaction(
				[]solana.Instruction{transferInstruction},
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
			fmt.Println("tx:")
			fmt.Println(base58.Encode(bytes))
			return nil
		},
	}
	cmd.Flags().String(flagConfigPath, defaultConfigPath, "Config file path")
	return cmd
}
