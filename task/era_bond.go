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

func (task *Task) EraBond(stakeManager *lsd_program.StakeManager) error {
	if stakeManager.EraStatus != lsd_program.EraStatusEraUpdated {
		return nil
	}
	if stakeManager.PendingBond <= stakeManager.PendingUnbond {
		return nil
	}
	if stakeManager.PendingBond-stakeManager.PendingUnbond < stakeManager.StakingMinStakeAmount {
		return nil
	}

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

	stakingStakeAccount, _, err := solana.FindProgramAddress([][]byte{utils.StakeAccountSeed, task.stakeManager.Bytes()}, task.stakingProgram)
	if err != nil {
		return err
	}
	instructions := []solana.Instruction{
		computebudget.NewSetComputeUnitPriceInstruction(20000).Build(),
		lsd_program.NewEraBondInstruction(
			task.feePayerAccount.PublicKey(),
			task.stakeManager,
			stakeManager.StakingTokenMint,
			stakeManagerStakingTokenAccount,
			stakeManager.StakingPool,
			stakingPoolStakingTokenAccount,
			stakingStakeAccount,
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

	logrus.Infof("EraBond will send tx: %s, bond amount: %d", tx.Signatures[0], stakeManager.PendingBond)

	_, err = utils.SendAndWaitForConfirmation(task.client, tx, latestBlockHashRes.Value.LastValidBlockHeight)
	if err != nil {
		return fmt.Errorf("waitForConfirmation error, err: %s", err.Error())
	}

	logrus.Info("EraBond success")

	return nil
}
