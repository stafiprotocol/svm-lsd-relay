package cmd

import (
	"context"
	"fmt"
	"time"

	"svm-lsd-relay/pkg/config"
	"svm-lsd-relay/pkg/lsd_program"
	"svm-lsd-relay/pkg/utils"
	"svm-lsd-relay/pkg/vault"

	"github.com/gagliardetto/solana-go"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/spf13/cobra"
	"golang.org/x/time/rate"
)

func stakeManagerInitCmd() *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "init",
		Short: "Init stake manager",

		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := cmd.Flags().GetString(flagConfigPath)
			if err != nil {
				return err
			}
			fmt.Printf("config path: %s\n", configPath)

			cfg, err := config.LoadConfig[config.ConfigInitStakeManager](configPath)
			if err != nil {
				return err
			}
			v, err := vault.NewVaultFromWalletFile(cfg.KeystorePath)
			if err != nil {
				return err
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

			rpcClient := rpc.NewWithCustomRPCClient(rpc.NewWithLimiter(
				cfg.RpcEndpoint,
				rate.Every(time.Second), // time frame
				5,                       // limit of requests per time frame
			))

			lsdProgramID := solana.MustPublicKeyFromBase58(cfg.LsdProgramID)
			stakingPool := solana.MustPublicKeyFromBase58(cfg.StakingPoolAddress)

			lsd_program.SetProgramID(lsdProgramID)

			feePayerAccount, exist := privateKeyMap[cfg.FeePayerAccount]
			if !exist {
				return fmt.Errorf("fee payer not exit in vault")
			}
			adminAccount, exist := privateKeyMap[cfg.AdminAccount]
			if !exist {
				return fmt.Errorf("admin not exit in vault")
			}

			stakingPoolDetail, err := utils.GetSvmStakingPool(rpcClient, stakingPool)
			if err != nil {
				return fmt.Errorf("GetSvmStakingPool failed: %s", err.Error())
			}

			stakingTokenMint := stakingPoolDetail.TokenMint
			stakeManager, _, err := solana.FindProgramAddress([][]byte{utils.StakeMangerSeed, adminAccount.PublicKey().Bytes(), []byte{cfg.Index}}, lsdProgramID)
			if err != nil {
				return err
			}
			lsdTokenMint, _, err := solana.FindProgramAddress([][]byte{utils.TokenMintSeed, adminAccount.PublicKey().Bytes(), []byte{cfg.Index}}, lsdProgramID)
			if err != nil {
				return err
			}

			stakingTokenMintDetail, err := rpcClient.GetAccountInfo(context.Background(), stakingTokenMint)
			if err != nil {
				return err
			}

			tokenProgramId := stakingTokenMintDetail.Value.Owner
			var stakeManagerStakingTokenAccount solana.PublicKey
			if tokenProgramId == solana.Token2022ProgramID {
				stakeManagerStakingTokenAccount, _, err = utils.FindAssociatedToken2022Address(stakeManager, stakingTokenMint)
				if err != nil {
					return err
				}
			} else {
				stakeManagerStakingTokenAccount, _, err = solana.FindAssociatedTokenAddress(stakeManager, stakingTokenMint)
				if err != nil {
					return err
				}
			}

			fmt.Println("lsdProgramID:", lsdProgramID.String())
			fmt.Println("stakingPool:", stakingPool.String())
			fmt.Println("stakingTokenMint:", stakingTokenMint.String())
			fmt.Println("stakeManager:", stakeManager.String())
			fmt.Println("lsdTokenMint:", lsdTokenMint.String())
			fmt.Println("admin", adminAccount.PublicKey().String())
			fmt.Println("feePayer:", feePayerAccount.PublicKey().String())
			fmt.Println("eraSeconds:", cfg.EraSeconds)
			fmt.Println("index:", cfg.Index)
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

			initIns := lsd_program.NewInitializeStakeManagerInstruction(lsd_program.InitializeStakeManagerParams{
				EraSeconds: cfg.EraSeconds,
				Index:      cfg.Index,
			}, adminAccount.PublicKey(), feePayerAccount.PublicKey(), stakeManager, stakingPool, lsdTokenMint, stakingTokenMint,
				stakeManagerStakingTokenAccount, tokenProgramId, solana.SPLAssociatedTokenAccountProgramID, solana.SystemProgramID)

			latestBlockHashRes, err := rpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentConfirmed)
			if err != nil {
				fmt.Printf("get recent block hash error, err: %v\n", err)
			}
			tx, err := solana.NewTransaction(
				[]solana.Instruction{
					computebudget.NewSetComputeUnitPriceInstruction(20000).Build(),
					initIns.Build()},
				latestBlockHashRes.Value.Blockhash,
				solana.TransactionPayer(feePayerAccount.PublicKey()))
			if err != nil {
				return fmt.Errorf("NewTransaction failed, err: %s, tx: %s", err.Error(), tx.String())
			}

			_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
				if feePayerAccount.PublicKey().Equals(key) {
					return &feePayerAccount
				}
				if adminAccount.PublicKey().Equals(key) {
					return &adminAccount
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("sign failed, err: %s", err.Error())
			}
			_, err = utils.SendAndWaitForConfirmation(rpcClient, tx, latestBlockHashRes.Value.LastValidBlockHeight)
			if err != nil {
				return fmt.Errorf("waitForConfirmation error, err: %s", err.Error())
			}

			fmt.Println("initializeStakeManager txHash:", tx.Signatures[0].String())

			return nil
		},
	}
	cmd.Flags().String(flagConfigPath, defaultConfigPath, "Config file path")
	return cmd
}
