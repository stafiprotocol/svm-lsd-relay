package task

import (
	"context"
	"fmt"
	"svm-lsd-relay/pkg/lsd_program"
	"svm-lsd-relay/pkg/staking_program"
	"svm-lsd-relay/pkg/utils"
	"time"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/sirupsen/logrus"
)

func (task *Task) EraWithdraw(stakeManager *lsd_program.StakeManager) error {
	unstakeAccounts, unstakeAccountPubkey, err := task.GetUnstakeAccounts(context.Background(), task.stakingProgram, stakeManager.StakingPool, task.stakeManager)
	if err != nil {
		return fmt.Errorf("GetUnstakeAccounts failed: %s", err.Error())
	}
	logrus.Debugf("unstakeAccounts: %v", unstakeAccountPubkey)
	for i, unstakeAccount := range unstakeAccounts {
		if unstakeAccount.WithdrawableTimestamp <= uint64(time.Now().Unix()) {
			var stakeManagerStakingTokenAccount solana.PublicKey
			var stakingPoolStakingTokenAccount solana.PublicKey
			var err error
			if task.tokenProgramId == solana.Token2022ProgramID {
				stakeManagerStakingTokenAccount, _, err = utils.FindAssociatedToken2022Address(task.stakeManager, stakeManager.StakingTokenMint)
				if err != nil {
					return err
				}
				stakingPoolStakingTokenAccount, _, err = utils.FindAssociatedToken2022Address(stakeManager.StakingPool, stakeManager.StakingTokenMint)
				if err != nil {
					return err
				}
			} else {
				stakeManagerStakingTokenAccount, _, err = solana.FindAssociatedTokenAddress(task.stakeManager, stakeManager.StakingTokenMint)
				if err != nil {
					return err
				}
				stakingPoolStakingTokenAccount, _, err = solana.FindAssociatedTokenAddress(stakeManager.StakingPool, stakeManager.StakingTokenMint)
				if err != nil {
					return err
				}
			}

			instructions := []solana.Instruction{
				computebudget.NewSetComputeUnitPriceInstruction(20000).Build(),
				lsd_program.NewEraWithdrawInstruction(
					task.feePayerAccount.PublicKey(),
					task.stakeManager,
					stakeManager.StakingPool,
					stakeManager.StakingTokenMint,
					stakeManagerStakingTokenAccount,
					stakingPoolStakingTokenAccount,
					unstakeAccountPubkey[i],
					task.stakingProgram,
					task.tokenProgramId,
					solana.SPLAssociatedTokenAccountProgramID,
					solana.SystemProgramID,
				).Build(),
			}

			latestBlockHashRes, err := task.client.GetLatestBlockhash(context.Background(), rpc.CommitmentConfirmed)
			if err != nil {
				fmt.Printf("get recent block hash error, err: %v\n", err)
			}
			tx, err := solana.NewTransaction(
				instructions,
				latestBlockHashRes.Value.Blockhash,
				solana.TransactionPayer(task.feePayerAccount.PublicKey()))
			if err != nil {
				return fmt.Errorf("NewTransaction failed, err: %s, tx: %s", err.Error(), tx.String())
			}

			_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
				if task.feePayerAccount.PublicKey().Equals(key) {
					return &task.feePayerAccount
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("sign failed, err: %s", err.Error())
			}

			logrus.Infof("EraWithdraw will send tx: %s, withdraw amount: %d",
				tx.Signatures[0], unstakeAccount.Amount)

			_, err = utils.SendAndWaitForConfirmation(task.client, tx, latestBlockHashRes.Value.LastValidBlockHeight)
			if err != nil {
				return fmt.Errorf("waitForConfirmation error, err: %s", err.Error())
			}

			logrus.Info("EraWithdraw success")

			return nil
		}
	}

	return nil

}

func (s *Task) GetUnstakeAccounts(ctx context.Context, programId, stakingPool, user solana.PublicKey) ([]staking_program.UnstakeAccount, []solana.PublicKey, error) {
	hexBts := append(stakingPool.Bytes(), user.Bytes()...)

	accounts, err := s.client.GetProgramAccountsWithOpts(
		context.Background(),
		programId,
		&rpc.GetProgramAccountsOpts{
			Commitment: rpc.CommitmentConfirmed,
			Encoding:   solana.EncodingBase64,
			Filters: []rpc.RPCFilter{
				{
					Memcmp: &rpc.RPCFilterMemcmp{
						Offset: 8,
						Bytes:  solana.Base58(hexBts),
					},
				},
				{
					DataSize: 216,
				},
			},
		})
	if err != nil {
		return nil, nil, err
	}
	ret := make([]staking_program.UnstakeAccount, 0)
	retPubkey := make([]solana.PublicKey, 0)
	for _, accountInfo := range accounts {
		unstakeAccount := staking_program.UnstakeAccount{}
		err = bin.NewBorshDecoder(accountInfo.Account.Data.GetBinary()).Decode(&unstakeAccount)
		if err != nil {
			return nil, nil, fmt.Errorf("deserialize err: %s", err.Error())
		}
		ret = append(ret, unstakeAccount)
		retPubkey = append(retPubkey, accountInfo.Pubkey)
	}
	return ret, retPubkey, nil
}
