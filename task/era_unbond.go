package task

import (
	"context"
	"fmt"

	"svm-lsd-relay/pkg/lsd_program"
	"svm-lsd-relay/pkg/utils"

	"github.com/gagliardetto/solana-go"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/sirupsen/logrus"
)

func (task *Task) EraUnbond(stakeManager *lsd_program.StakeManager) error {
	if stakeManager.EraStatus != lsd_program.EraStatusEraUpdated {
		return nil
	}
	if stakeManager.PendingBond >= stakeManager.PendingUnbond {
		return nil
	}

	stakingStakeAccount, _, err := solana.FindProgramAddress([][]byte{utils.StakeAccountSeed, task.stakeManager.Bytes()}, task.stakingProgram)
	if err != nil {
		return err
	}

	stakingUnstakeAccount := solana.NewWallet().PrivateKey

	instructions := []solana.Instruction{
		computebudget.NewSetComputeUnitPriceInstruction(20000).Build(),
		lsd_program.NewEraUnbondInstruction(
			task.feePayerAccount.PublicKey(),
			task.stakeManager,
			stakeManager.StakingPool,
			stakingStakeAccount,
			stakingUnstakeAccount.PublicKey(),
			task.stakingProgram,
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
		if stakingUnstakeAccount.PublicKey().Equals(key) {
			return &stakingUnstakeAccount
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("sign failed, err: %s", err.Error())
	}

	logrus.Infof("EraUnbond will send tx: %s,unbond amount: %d", tx.Signatures[0], stakeManager.PendingUnbond-stakeManager.PendingBond)

	_, err = utils.SendAndWaitForConfirmation(task.client, tx, latestBlockHashRes.Value.LastValidBlockHeight)
	if err != nil {
		return fmt.Errorf("waitForConfirmation error, err: %s", err.Error())
	}

	logrus.Info("EraUnbond success")

	return nil
}
