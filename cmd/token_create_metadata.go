package cmd

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gagliardetto/solana-go"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/mr-tron/base58"
	"github.com/spf13/cobra"
	"golang.org/x/time/rate"
	"svm-lsd-relay/pkg/config"
	"svm-lsd-relay/pkg/lsd_program"
	"svm-lsd-relay/pkg/utils"
	"svm-lsd-relay/pkg/vault"
)

// TokenMetadataProgramID is the program ID for the token metadata program
var TokenMetadataProgramID = solana.MustPublicKeyFromBase58("metaqbxxUerdq28cj1RbAWkYQm3ybzjb6a8bt518x1s")

func createTokenMetadataCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "create-token-metadata",
		Short: "Create metadata for LSD token",

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

			cfg, err := config.LoadConfig[config.ConfigCreateMetadata](configPath)
			if err != nil {
				return err
			}

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

			rpcClient := rpc.NewWithCustomRPCClient(rpc.NewWithLimiter(
				cfg.RpcEndpoint,
				rate.Every(time.Second), // time frame
				5,                       // limit of requests per time frame
			))

			lsdProgramID := solana.MustPublicKeyFromBase58(cfg.LsdProgramID)
			stakeManagerPubkey := solana.MustPublicKeyFromBase58(cfg.StakeManagerAddress)

			lsd_program.SetProgramID(lsdProgramID)

			stakeManagerDetail, err := utils.GetSvmLsdStakeManager(rpcClient, stakeManagerPubkey)
			if err != nil {
				return err
			}

			lsdTokenMint, _, err := solana.FindProgramAddress([][]byte{utils.TokenMintSeed, stakeManagerDetail.Creator.Bytes(), []byte{0}}, lsdProgramID)
			if err != nil {
				return err
			}
			stakingPoolDetail, err := utils.GetSvmStakingPool(rpcClient, stakeManagerDetail.StakingPool)
			if err != nil {
				return err
			}

			stakingTokenMint := stakingPoolDetail.TokenMint
			stakingTokenMintDetail, err := rpcClient.GetAccountInfo(context.Background(), stakingTokenMint)
			if err != nil {
				return err
			}

			tokenProgramId := stakingTokenMintDetail.Value.Owner

			// Derive metadata account
			metadataAccount, _, err := solana.FindProgramAddress(
				[][]byte{
					[]byte("metadata"),
					TokenMetadataProgramID.Bytes(),
					lsdTokenMint.Bytes(),
				},
				TokenMetadataProgramID,
			)
			if err != nil {
				return fmt.Errorf("failed to derive metadata account: %w", err)
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
				lsd_program.NewCreateMetadataInstruction(
					lsd_program.CreateMetadataParams{
						TokenName:   cfg.TokenName,
						TokenSymbol: cfg.TokenSymbol,
						TokenUri:    cfg.TokenUri,
					},

					feePayerAccountPublicKey,
					adminAccountPublicKey,
					stakeManagerPubkey,
					lsdTokenMint,
					metadataAccount,
					tokenProgramId,
					TokenMetadataProgramID,
					solana.SystemProgramID,
					solana.SysVarInstructionsPubkey,
				).Build(),
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
				tx.Message.SetVersion(solana.MessageVersionLegacy)

				txData, err := tx.MarshalBinary()
				if err != nil {
					return err
				}
				fmt.Println("tx(base64) -- for Inspector")
				fmt.Println(base64.StdEncoding.EncodeToString(txData))

				bytes, err := tx.Message.MarshalBinary()
				if err != nil {
					return fmt.Errorf("fail to marshal tx.Message: %w", err)
				}
				fmt.Println("tx.Message(base58):")
				fmt.Println(base58.Encode(bytes))

				items := []utils.CompiledInstruction{}
				for _, in := range tx.Message.Instructions {
					accounts := []uint8{}
					for _, a := range in.Accounts {
						accounts = append(accounts, uint8(a))
					}

					items = append(items, utils.CompiledInstruction{
						ProgramIDIndex: uint8(in.ProgramIDIndex),
						AccountIndexes: utils.SmallVec[uint8]{
							Items: accounts,
						},
						Data: utils.SmallVec16[uint8]{
							Items: in.Data,
						},
					})
				}

				txMessage := utils.TransactionMessage{
					NumSigners:            tx.Message.Header.NumRequiredSignatures,
					NumWritableSigners:    tx.Message.Header.NumRequiredSignatures - tx.Message.Header.NumReadonlySignedAccounts,
					NumWritableNonSigners: uint8(len(tx.Message.AccountKeys)) - tx.Message.Header.NumRequiredSignatures - tx.Message.Header.NumReadonlyUnsignedAccounts,
					AccountKeys: utils.SmallVec[solana.PublicKey]{
						Items: tx.Message.AccountKeys,
					},
					Instructions: utils.SmallVec[utils.CompiledInstruction]{
						Items: items,
					},
					AddressTableLookups: utils.SmallVec[utils.MessageAddressTableLookup]{},
				}

				txBts, err := txMessage.Serialize()
				if err != nil {
					return err
				}
				fmt.Println("tx.Message(cli):")
				fmt.Println(hex.EncodeToString(txBts))

				return nil
			} else {
				_, err = tx.Sign(signFn)
				if err != nil {
					return fmt.Errorf("sign failed, err: %s", err.Error())
				}
				_, err = utils.SendAndWaitForConfirmation(rpcClient, tx, latestBlockHashRes.Value.LastValidBlockHeight)
				if err != nil {
					return fmt.Errorf("waitForConfirmation error, err: %s", err.Error())
				}

				fmt.Println("createMetadata txHash:", tx.Signatures[0].String())

				return nil
			}
		},
	}
	cmd.Flags().String(flagConfigPath, defaultConfigPath, "Config file path")
	cmd.Flags().Bool(flagExportTx, false, "export a base58 encoded transaction message (not a transaction)")
	return cmd
}
