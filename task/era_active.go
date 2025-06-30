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

func (task *Task) EraActive(stakeManager *lsd_program.StakeManager) error {
	if stakeManager.EraStatus != lsd_program.EraStatusBonded && stakeManager.EraStatus != lsd_program.EraStatusUnbonded {
		return nil
	}

	var platformFeeRecipient solana.PublicKey
	var err error
	if task.tokenProgramId == solana.Token2022ProgramID {
		platformFeeRecipient, _, err = utils.FindAssociatedToken2022Address(stakeManager.Admin, task.lsdTokenMint)
		if err != nil {
			return err
		}
	} else {
		platformFeeRecipient, _, err = solana.FindAssociatedTokenAddress(stakeManager.Admin, task.lsdTokenMint)
		if err != nil {
			return err
		}
	}
	var stakeManagerStakingTokenAccount solana.PublicKey
	var stakingPoolStakingTokenAccount solana.PublicKey
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

	instructions := []solana.Instruction{computebudget.NewSetComputeUnitPriceInstruction(20000).Build()}

	instructions = append(instructions,
		lsd_program.NewEraActiveInstruction(
			task.feePayerAccount.PublicKey(),
			stakeManager.Admin,
			task.stakeManager,
			task.lsdTokenMint,
			task.stakingTokenMint,
			stakeManagerStakingTokenAccount,
			platformFeeRecipient,
			stakeManager.StakingPool,
			stakingPoolStakingTokenAccount,
			stakingStakeAccount,
			task.stakingProgram,
			task.tokenProgramId,
			solana.SPLAssociatedTokenAccountProgramID,
			solana.SystemProgramID,
		).Build())

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

	logrus.Infof("EraActive will send tx: %s", tx.Signatures[0])

	_, err = utils.SendAndWaitForConfirmation(task.client, tx, latestBlockHashRes.Value.LastValidBlockHeight)
	if err != nil {
		return fmt.Errorf("waitForConfirmation error, err: %s", err.Error())
	}

	newStakeManager, err := utils.GetSvmLsdStakeManager(task.client, task.stakeManager)
	if err != nil {
		return err
	}

	logrus.Infof("EraActive success, oldActive: %d newActive: %d, oldRate: %d newRate: %d",
		stakeManager.Active, newStakeManager.Active, stakeManager.Rate, newStakeManager.Rate)

	return nil

}
