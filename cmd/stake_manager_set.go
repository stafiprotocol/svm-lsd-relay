package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"svm-lsd-relay/pkg/config"
	"svm-lsd-relay/pkg/lsd_program"
	"svm-lsd-relay/pkg/utils"
	"svm-lsd-relay/pkg/vault"

	"github.com/gagliardetto/solana-go"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/mr-tron/base58"
	"github.com/spf13/cobra"
	"golang.org/x/time/rate"
)

func stakeManagerSetCmd() *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "set",
		Short: "Set stake manager",

		RunE: func(cmd *cobra.Command, args []string) error {
			exportTxMessage, err := cmd.Flags().GetBool(flagExportTx)
			if err != nil {
				return err
			}

			configPath, err := cmd.Flags().GetString(flagConfigPath)
			if err != nil {
				return err
			}
			fmt.Printf("config path: %s\n", configPath)

			cfg, err := config.LoadConfig[config.ConfigSetStakeManager](configPath)
			if err != nil {
				return err
			}

			rpcClient := rpc.NewWithCustomRPCClient(rpc.NewWithLimiter(
				cfg.RpcEndpoint,
				rate.Every(time.Second), // time frame
				5,                       // limit of requests per time frame
			))

			stakeManagerPubkey := solana.MustPublicKeyFromBase58(cfg.StakeManagerAddress)
			stakeManagerDetail, err := rpcClient.GetAccountInfo(context.Background(), stakeManagerPubkey)
			if err != nil {
				return err
			}
			lsdProgramID := stakeManagerDetail.Value.Owner

			lsd_program.SetProgramID(lsdProgramID)

			var signFn func(key solana.PublicKey) *solana.PrivateKey
			feePayerAccountPublicKey := solana.MustPublicKeyFromBase58(cfg.FeePayerAccount)
			adminAccountPublicKey := solana.MustPublicKeyFromBase58(cfg.AdminAccount)
			if !exportTxMessage {
				v, err := vault.NewVaultFromWalletFile(cfg.KeystorePath)
				if err != nil {
					return fmt.Errorf("could not open keystore file '%s': %w.\nWARN: or do you miss --export flag?", cfg.KeystorePath, err)
				}
				boxer, err := vault.SecretBoxerForType(v.SecretBoxWrap)
				if err != nil {
					return fmt.Errorf("secret boxer: %w", err)
				}

				if err := v.Open(boxer); err != nil {
					return fmt.Errorf("opening: %w", err)
				}

				privateKeyMap := make(map[string]solana.PrivateKey)
				for _, privKey := range v.KeyBag {
					privateKeyMap[privKey.PublicKey().String()] = solana.PrivateKey(privKey)
				}
				feePayerAccount, exist := privateKeyMap[cfg.FeePayerAccount]
				if !exist {
					return fmt.Errorf("fee payer not exit in vault")
				}
				adminAccount, exist := privateKeyMap[cfg.AdminAccount]
				if !exist {
					return fmt.Errorf("admin not exit in vault")
				}
				feePayerAccountPublicKey = feePayerAccount.PublicKey()
				adminAccountPublicKey = adminAccount.PublicKey()

				signFn = func(key solana.PublicKey) *solana.PrivateKey {
					if feePayerAccount.PublicKey().Equals(key) {
						return &feePayerAccount
					}
					if adminAccount.PublicKey().Equals(key) {
						return &adminAccount
					}
					return nil
				}
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

			instructions := []solana.Instruction{
				lsd_program.NewConfigStakeManagerInstruction(
					lsd_program.ConfigStakeManagerParams{
						MinStakeAmount:        cfg.MinStakeAmount,
						PlatformFeeCommission: cfg.PlatformFeeCommission,
						RateChangeLimit:       cfg.RateChangeLimit,
					},
					adminAccountPublicKey,
					stakeManagerPubkey).Build(),
			}
			if !exportTxMessage {
				instructions = append([]solana.Instruction{
					computebudget.NewSetComputeUnitPriceInstruction(20000).Build(),
				}, instructions...)
			}

			latestBlockHashRes, err := rpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentConfirmed)
			if err != nil {
				fmt.Printf("get recent block hash error, err: %v\n", err)
			}
			tx, err := solana.NewTransaction(
				instructions,
				latestBlockHashRes.Value.Blockhash,
				solana.TransactionPayer(feePayerAccountPublicKey))
			if err != nil {
				return fmt.Errorf("NewTransaction failed, err: %s, tx: %s", err.Error(), tx.String())
			}

			if exportTxMessage {
				fmt.Println(tx)
				bytes, err := tx.Message.MarshalBinary()
				if err != nil {
					return fmt.Errorf("fail to marshal tx.Message: %w", err)
				}
				fmt.Println("tx:")
				fmt.Println(base58.Encode(bytes))
			} else {
				_, err = tx.Sign(signFn)
				if err != nil {
					return fmt.Errorf("sign failed, err: %s", err.Error())
				}
				_, err = utils.SendAndWaitForConfirmation(rpcClient, tx, latestBlockHashRes.Value.LastValidBlockHeight)
				if err != nil {
					return fmt.Errorf("waitForConfirmation error, err: %s", err.Error())
				}

				fmt.Println("setStakeManager txHash:", tx.Signatures[0].String())
			}

			return nil
		},
	}
	cmd.Flags().String(flagConfigPath, defaultConfigPath, "Config file path")
	cmd.Flags().Bool(flagExportTx, false, "export a base58 encoded transaction message (not a transaction)")
	return cmd
}
