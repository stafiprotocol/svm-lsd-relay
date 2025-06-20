package task

import (
	"svm-lsd-relay/pkg/lsd_program"
)

func (task *Task) EraWithdraw(stakeManager *lsd_program.StakeManager) error {
	if stakeManager.EraStatus != lsd_program.EraStatusEraUpdated {
		return nil
	}
	return nil
	// var poolTokenAccount solana.PublicKey
	// if task.tokenProgramId == solana.Token2022ProgramID {
	// 	sspPoolTokenAccount, _, err = utils.FindAssociatedToken2022Address(sspStakePool, task.sonicTokenMint)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	poolTokenAccount, _, err = utils.FindAssociatedToken2022Address(task.stakePool, task.sonicTokenMint)
	// 	if err != nil {
	// 		return err
	// 	}
	// } else {
	// 	sspPoolTokenAccount, _, err = solana.FindAssociatedTokenAddress(sspStakePool, task.sonicTokenMint)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	poolTokenAccount, _, err = solana.FindAssociatedTokenAddress(task.stakePool, task.sonicTokenMint)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	// instructions := []solana.Instruction{
	// 	computebudget.NewSetComputeUnitPriceInstruction(20000).Build(),
	// 	lsd_program.NewEraWithdrawInstruction(
	// 		lsd_program.EraWithdrawParams{
	// 			StakingPoolBump:       sspStakePoolBump,
	// 			PublicStakeConfigBump: sspPublicStakeConfigBump,
	// 			UserStakeBump:         sspUserStakeBump,
	// 			UserStakeIndexBump:    sspUserStakeIndexBump_,
	// 			Index:                 sspStakeIndex,
	// 			SignatureCreatedAt:    signatureCreatedAt,
	// 			Signature:             signature,
	// 			SignatureRecoveryId:   uint8(signatureRecoveryID),
	// 		},
	// 		task.feePayerAccount.PublicKey(),
	// 		sspStakePool,
	// 		sspPublicStakeConfig,
	// 		sspUserStake,
	// 		sspUserStakeIndex,
	// 		sspPoolTokenAccount,
	// 		task.sonicTokenMint,
	// 		poolTokenAccount,
	// 		task.stakeManager,
	// 		task.stakePool,
	// 		task.ssp,
	// 		task.sonicTokenProgramId,
	// 		solana.SPLAssociatedTokenAccountProgramID,
	// 		solana.SystemProgramID,
	// 		solana.SysVarRentPubkey,
	// 	).Build(),
	// }

	// latestBlockHashRes, err := task.client.GetLatestBlockhash(context.Background(), rpc.CommitmentConfirmed)
	// if err != nil {
	// 	fmt.Printf("get recent block hash error, err: %v\n", err)
	// }
	// tx, err := solana.NewTransaction(
	// 	instructions,
	// 	latestBlockHashRes.Value.Blockhash,
	// 	solana.TransactionPayer(task.feePayerAccount.PublicKey()))
	// if err != nil {
	// 	return fmt.Errorf("NewTransaction failed, err: %s, tx: %s", err.Error(), tx.String())
	// }

	// _, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
	// 	if task.feePayerAccount.PublicKey().Equals(key) {
	// 		return &task.feePayerAccount
	// 	}
	// 	return nil
	// })
	// if err != nil {
	// 	return fmt.Errorf("sign failed, err: %s", err.Error())
	// }

	// logrus.Infof("EraWithdraw will send tx: %s, withdraw amount: %d",
	// 	tx.Signatures[0], willUseStakeInfo.Amount+willUseStakeInfo.Reward)

	// _, err = utils.SendAndWaitForConfirmation(task.client, tx, latestBlockHashRes.Value.LastValidBlockHeight)
	// if err != nil {
	// 	return fmt.Errorf("waitForConfirmation error, err: %s", err.Error())
	// }

	// logrus.Info("EraWithdraw success")

	// return nil

}
