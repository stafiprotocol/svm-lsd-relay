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

func (task *Task) EraNew(stakeManager *lsd_program.StakeManager) error {
	slotRes, err := task.client.GetSlot(context.Background(), rpc.CommitmentConfirmed)
	if err != nil {
		return fmt.Errorf("GetSlot failed, %s", err.Error())
	}
	blockTimeRes, err := task.client.GetBlockTime(context.Background(), slotRes)
	if err != nil {
		return fmt.Errorf("GetBlockTime failed, %s, height: %d", err.Error(), slotRes)
	}

	currentEra := blockTimeRes.Time().Unix()/stakeManager.EraSeconds + stakeManager.EraOffset
	if stakeManager.LatestEra >= uint64(currentEra) {
		return nil
	}
	if stakeManager.EraStatus != lsd_program.EraStatusActiveUpdated {
		return nil
	}

	instructions := []solana.Instruction{
		computebudget.NewSetComputeUnitPriceInstruction(20000).Build(),
		lsd_program.NewEraNewInstruction(task.stakeManager).Build(),
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

	logrus.Infof("EraNew will send tx: %s, newEra: %d", tx.Signatures[0], stakeManager.LatestEra+1)

	_, err = utils.SendAndWaitForConfirmation(task.client, tx, latestBlockHashRes.Value.LastValidBlockHeight)
	if err != nil {
		return fmt.Errorf("waitForConfirmation error, err: %s", err.Error())
	}

	logrus.Infof("EraNew success")

	return nil
}
