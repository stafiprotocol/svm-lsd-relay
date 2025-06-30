package utils

import (
	"context"
	"fmt"
	"strings"
	"time"

	"svm-lsd-relay/pkg/lsd_program"
	"svm-lsd-relay/pkg/staking_program"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

var StakePoolSeed = []byte("pool_seed")
var StakeAccountSeed = []byte("stake_account_seed")
var StakeMangerSeed = []byte("stake_manager_seed")
var TokenMintSeed = []byte("token_mint_seed")

var ErrExpired = fmt.Errorf("expired")

func SendAndWaitForConfirmation(rpcClient *rpc.Client, tx *solana.Transaction,
	lastValidBlockHeight uint64) (*rpc.GetTransactionResult, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	opts := rpc.TransactionOpts{
		SkipPreflight:       false,
		PreflightCommitment: rpc.CommitmentConfirmed,
	}

	sig, err := rpcClient.SendTransactionWithOpts(ctx, tx, opts)
	if err != nil {
		return nil, fmt.Errorf("SendTransactionWithOpts failed, err: %s, tx: %s", err.Error(), tx.String())
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			nowBlockHeight, err := rpcClient.GetBlockHeight(ctx, rpc.CommitmentConfirmed)
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
			txRes, err := rpcClient.GetTransaction(ctx, sig, &rpc.GetTransactionOpts{
				Commitment:                     rpc.CommitmentConfirmed,
				MaxSupportedTransactionVersion: &rpc.MaxSupportedTransactionVersion1,
			})
			if err != nil {
				if nowBlockHeight > lastValidBlockHeight {
					return nil, ErrExpired
				}

				rpcClient.SendTransactionWithOpts(ctx, tx, opts)

				time.Sleep(time.Second)
				continue
			}

			if txRes.Meta.Err != nil {
				errString := ""
				for _, log := range txRes.Meta.LogMessages {
					if strings.Contains(log, "Error") || strings.Contains(log, "error") {
						errString += fmt.Sprintf(" log: %s", log)
					}
				}

				return nil, fmt.Errorf("tx execute failed: %v, logs: %s", txRes.Meta.Err, errString)
			}
			return txRes, nil
		}
	}
}

func GetSvmLsdStakeManager(rpcClient *rpc.Client, stakeManager solana.PublicKey) (*lsd_program.StakeManager, error) {
	ret := lsd_program.StakeManager{}
	account, err := rpcClient.GetAccountInfoWithOpts(context.Background(), stakeManager, &rpc.GetAccountInfoOpts{
		Encoding:   solana.EncodingBase64,
		Commitment: rpc.CommitmentConfirmed,
	})
	if err != nil {
		return nil, err
	}

	err = bin.NewBorshDecoder(account.Value.Data.GetBinary()).Decode(&ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func GetSvmStakingPool(rpcClient *rpc.Client, stakingPool solana.PublicKey) (*staking_program.StakingPool, error) {
	ret := staking_program.StakingPool{}
	account, err := rpcClient.GetAccountInfoWithOpts(context.Background(), stakingPool, &rpc.GetAccountInfoOpts{
		Encoding:   solana.EncodingBase64,
		Commitment: rpc.CommitmentConfirmed,
	})
	if err != nil {
		return nil, err
	}

	err = bin.NewBorshDecoder(account.Value.Data.GetBinary()).Decode(&ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func MustFindATA_withCorrectProgram(wallet, mint solana.PublicKey, mintAccount *rpc.Account) (solana.PublicKey, uint8) {
	if mintAccount.Owner == solana.TokenProgramID {
		return MustFindAssociatedTokenAddress(wallet, mint)
	} else if mintAccount.Owner == solana.Token2022ProgramID {
		pk, seed, err := FindAssociatedToken2022Address(wallet, mint)
		if err != nil {
			panic(err)
		}
		return pk, seed
	}
	panic("unknown program id " + mintAccount.Owner.String() + " for mint " + mint.String())
}

func FindAssociatedToken2022Address(
	wallet solana.PublicKey,
	mint solana.PublicKey,
) (solana.PublicKey, uint8, error) {
	return findAssociatedToken2022AddressAndBumpSeed(
		wallet,
		mint,
		solana.SPLAssociatedTokenAccountProgramID,
	)
}

func findAssociatedToken2022AddressAndBumpSeed(
	walletAddress solana.PublicKey,
	splTokenMintAddress solana.PublicKey,
	programID solana.PublicKey,
) (solana.PublicKey, uint8, error) {
	return solana.FindProgramAddress([][]byte{
		walletAddress[:],
		solana.Token2022ProgramID[:],
		splTokenMintAddress[:],
	},
		programID,
	)
}

func MustFindAssociatedTokenAddress(
	wallet solana.PublicKey,
	mint solana.PublicKey,
) (solana.PublicKey, uint8) {
	account, seed, err := solana.FindAssociatedTokenAddress(wallet, mint)
	if err != nil {
		panic(err)
	}
	return account, seed
}
