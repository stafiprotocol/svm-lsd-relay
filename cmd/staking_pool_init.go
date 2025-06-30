package cmd

import (
	"context"
	"fmt"
	"time"

	"svm-lsd-relay/pkg/config"
	"svm-lsd-relay/pkg/staking_program"
	"svm-lsd-relay/pkg/utils"
	"svm-lsd-relay/pkg/vault"

	"github.com/gagliardetto/solana-go"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/spf13/cobra"
	"golang.org/x/time/rate"
)

func stakingPoolInitCmd() *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "init",
		Short: "Init staking pool",

		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := cmd.Flags().GetString(flagConfigPath)
			if err != nil {
				return err
			}
			fmt.Printf("config path: %s\n", configPath)

			cfg, err := config.LoadConfig[config.ConfigInitStakingPool](configPath)
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

			stakingProgramID := solana.MustPublicKeyFromBase58(cfg.StakingProgramID)
			stakingTokenMint := solana.MustPublicKeyFromBase58(cfg.StakingTokenMint)
			staking_program.SetProgramID(stakingProgramID)

			feePayerAccount, exist := privateKeyMap[cfg.FeePayerAccount]
			if !exist {
				return fmt.Errorf("fee payer not exit in vault")
			}
			adminAccount, exist := privateKeyMap[cfg.AdminAccount]
			if !exist {
				return fmt.Errorf("admin not exit in vault")
			}

			stakingTokenMintDetail, err := rpcClient.GetAccountInfo(context.Background(), stakingTokenMint)
			if err != nil {
				return err
			}

			stakingPool, _, err := solana.FindProgramAddress([][]byte{utils.StakePoolSeed, stakingTokenMint.Bytes(), adminAccount.PublicKey().Bytes(), {cfg.Index}}, stakingProgramID)
			if err != nil {
				return err
			}

			tokenProgramId := stakingTokenMintDetail.Value.Owner

			var adminTokenAccount solana.PublicKey
			var stakingPoolTokenAccount solana.PublicKey
			if tokenProgramId == solana.Token2022ProgramID {
				adminTokenAccount, _, err = utils.FindAssociatedToken2022Address(adminAccount.PublicKey(), stakingTokenMint)
				if err != nil {
					return err
				}
				stakingPoolTokenAccount, _, err = utils.FindAssociatedToken2022Address(stakingPool, stakingTokenMint)
				if err != nil {
					return err
				}
			} else {

				adminTokenAccount, _, err = solana.FindAssociatedTokenAddress(adminAccount.PublicKey(), stakingTokenMint)
				if err != nil {
					return err
				}
				stakingPoolTokenAccount, _, err = solana.FindAssociatedTokenAddress(stakingPool, stakingTokenMint)
				if err != nil {
					return err
				}

			}

			fmt.Println("stakingProgramID:", stakingProgramID.String())
			fmt.Println("stakingPool:", stakingPool.String())
			fmt.Println("stakingTokenMint:", stakingTokenMint.String())
			fmt.Println("admin", adminAccount.PublicKey().String())
			fmt.Println("feePayer:", feePayerAccount.PublicKey().String())
			fmt.Println("rewardRate:", cfg.RewardRate)
			fmt.Println("totalReward:", cfg.TotalReward)
			fmt.Println("unbondingSeconds:", cfg.UnbondingSeconds)
			fmt.Println("rewardAlgorithm:", cfg.RewardAlgorithm)
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

			initIns := staking_program.NewInitializeStakingPoolInstruction(
				staking_program.InitializeStakingPoolParams{
					RewardRate:       cfg.RewardRate,
					TotalReward:      cfg.TotalReward,
					UnbondingSeconds: cfg.UnbondingSeconds,
					RewardAlgorithm:  cfg.RewardAlgorithm,
					Index:            cfg.Index,
				}, adminAccount.PublicKey(), feePayerAccount.PublicKey(), stakingTokenMint, stakingPool, adminTokenAccount, stakingPoolTokenAccount,
				solana.TokenProgramID, solana.SPLAssociatedTokenAccountProgramID, solana.SystemProgramID)

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

			fmt.Println("initializeStakingPool txHash:", tx.Signatures[0].String())

			return nil
		},
	}
	cmd.Flags().String(flagConfigPath, defaultConfigPath, "Config file path")
	return cmd
}
