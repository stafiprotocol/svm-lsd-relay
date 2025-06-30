package utils_test

import (
	"context"
	"testing"
	"time"

	svm_lsd_program "svm-lsd-relay/pkg/lsd_program"
	"svm-lsd-relay/pkg/staking_program"
	"svm-lsd-relay/pkg/utils"

	"github.com/gagliardetto/solana-go"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/rpc"
	"golang.org/x/time/rate"
)

func TestStaking(t *testing.T) {
	stakingProgram := solana.MustPublicKeyFromBase58("DjzuM5GR2NLjwcXAvzC5wqj8oTZD71FR5suCDGnx3GmB")
	endpoint := "https://api.devnet.solana.com"
	rpcClient := rpc.NewWithCustomRPCClient(rpc.NewWithLimiter(
		endpoint,
		rate.Every(time.Second), // time frame
		5,                       // limit of requests per time frame
	))
	user, err := solana.PrivateKeyFromSolanaKeygenFile("/Users/tpkeeper/.config/solana/id.json")
	if err != nil {
		t.Fatal(err)
	}
	staking_program.SetProgramID(stakingProgram)

	tokenMint := solana.MustPublicKeyFromBase58("4qVFnsc4WJLo5t4a8guWtbz2K6KatuuMaPWbDFJ9Z2sh")
	stakingPool, _, err := solana.FindProgramAddress([][]byte{utils.StakePoolSeed, tokenMint.Bytes(), user.PublicKey().Bytes()}, stakingProgram)
	if err != nil {
		t.Fatal(err)
	}
	adminTokenAccount, _, err := solana.FindAssociatedTokenAddress(user.PublicKey(), tokenMint)
	if err != nil {
		t.Fatal(err)
	}
	poolTokenAccount, _, err := solana.FindAssociatedTokenAddress(stakingPool, tokenMint)
	if err != nil {
		t.Fatal(err)
	}

	initIns := staking_program.NewInitializeStakingPoolInstruction(
		staking_program.InitializeStakingPoolParams{
			RewardRate:       1,
			TotalReward:      0,
			UnbondingSeconds: 60,
			RewardAlgorithm:  staking_program.RewardAlgorithmFixedPerTokenPerSecond,
			Index:            0,
		}, user.PublicKey(), user.PublicKey(), tokenMint, stakingPool, adminTokenAccount, poolTokenAccount,
		solana.TokenProgramID, solana.SPLAssociatedTokenAccountProgramID, solana.SystemProgramID)
	latestBlockHashRes, err := rpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentConfirmed)
	if err != nil {
		t.Fatal(err)
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{initIns.Build()},
		latestBlockHashRes.Value.Blockhash,
		solana.TransactionPayer(user.PublicKey()))
	if err != nil {
		t.Fatal(err)
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if user.PublicKey().Equals(key) {
			return &user
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Stake will send tx: %s", tx.Signatures[0])

	_, err = utils.SendAndWaitForConfirmation(rpcClient, tx, latestBlockHashRes.Value.LastValidBlockHeight)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSvmLsd(t *testing.T) {
	stakingProgram := solana.MustPublicKeyFromBase58("DjzuM5GR2NLjwcXAvzC5wqj8oTZD71FR5suCDGnx3GmB")
	lsdProgram := solana.MustPublicKeyFromBase58("6UrZH8GHxgSHu13ZqUMxHwiUnezXSqnEKDVNEpY1cAPu")

	endpoint := "https://api.devnet.solana.com"
	rpcClient := rpc.NewWithCustomRPCClient(rpc.NewWithLimiter(
		endpoint,
		rate.Every(time.Second), // time frame
		5,                       // limit of requests per time frame
	))
	user, err := solana.PrivateKeyFromSolanaKeygenFile("/Users/tpkeeper/.config/solana/id.json")
	if err != nil {
		t.Fatal(err)
	}
	svm_lsd_program.SetProgramID(lsdProgram)
	staking_program.SetProgramID(stakingProgram)

	stakingTokenMint := solana.MustPublicKeyFromBase58("7N3HQ8P73rxgZe426XSXGejxtL5KbUvHcPup2KTKvGWf")

	initStakingParamIndex := uint8(2)

	stakingPool, _, err := solana.FindProgramAddress([][]byte{utils.StakePoolSeed, stakingTokenMint.Bytes(), user.PublicKey().Bytes(), {initStakingParamIndex}}, stakingProgram)
	if err != nil {
		t.Fatal(err)
	}
	adminStakingTokenAccount, _, err := solana.FindAssociatedTokenAddress(user.PublicKey(), stakingTokenMint)
	if err != nil {
		t.Fatal(err)
	}
	stakingPoolStakingTokenAccount, _, err := solana.FindAssociatedTokenAddress(stakingPool, stakingTokenMint)
	if err != nil {
		t.Fatal(err)
	}

	initStakingIns := staking_program.NewInitializeStakingPoolInstruction(
		staking_program.InitializeStakingPoolParams{
			RewardRate:       1000,
			TotalReward:      1000000000000,
			UnbondingSeconds: 600,
			RewardAlgorithm:  staking_program.RewardAlgorithmFixedPerTokenPerSecond,
			Index:            initStakingParamIndex,
		}, user.PublicKey(), user.PublicKey(), stakingTokenMint, stakingPool, adminStakingTokenAccount, stakingPoolStakingTokenAccount,
		solana.TokenProgramID, solana.SPLAssociatedTokenAccountProgramID, solana.SystemProgramID)

	initLsdParamIndex := uint8(3)
	eraSeconds := int64(600)

	stakeManager, _, err := solana.FindProgramAddress([][]byte{utils.StakeMangerSeed, user.PublicKey().Bytes(), {initLsdParamIndex}}, lsdProgram)
	if err != nil {
		t.Fatal(err)
	}
	lsdTokenMint, _, err := solana.FindProgramAddress([][]byte{utils.TokenMintSeed, user.PublicKey().Bytes(), {initLsdParamIndex}}, lsdProgram)
	if err != nil {
		t.Fatal(err)
	}
	stakeManagerStakingTokenAccount, _, err := solana.FindAssociatedTokenAddress(stakeManager, stakingTokenMint)
	if err != nil {
		t.Fatal(err)
	}

	initLsdIns := svm_lsd_program.NewInitializeStakeManagerInstruction(svm_lsd_program.InitializeStakeManagerParams{
		EraSeconds: eraSeconds,
		Index:      initLsdParamIndex,
	}, user.PublicKey(), user.PublicKey(), stakeManager, stakingPool, lsdTokenMint, stakingTokenMint,
		stakeManagerStakingTokenAccount, solana.TokenProgramID, solana.SPLAssociatedTokenAccountProgramID, solana.SystemProgramID)

	latestBlockHashRes, err := rpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentConfirmed)
	if err != nil {
		t.Fatal(err)
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			initStakingIns.Build(),
			initLsdIns.Build()},
		latestBlockHashRes.Value.Blockhash,
		solana.TransactionPayer(user.PublicKey()))
	if err != nil {
		t.Fatal(err)
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if user.PublicKey().Equals(key) {
			return &user
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Stake will send tx: %s", tx.Signatures[0])

	_, err = utils.SendAndWaitForConfirmation(rpcClient, tx, latestBlockHashRes.Value.LastValidBlockHeight)
	if err != nil {
		t.Fatal(err)
	}
}

func TestStake(t *testing.T) {

	endpoint := "https://api.devnet.solana.com"
	// endpoint := "https://api.testnet.sonic.game"
	rpcClient := rpc.NewWithCustomRPCClient(rpc.NewWithLimiter(
		endpoint,
		rate.Every(time.Second), // time frame
		5,                       // limit of requests per time frame
	))

	stakeManager := solana.MustPublicKeyFromBase58("9MD8swVgHG11rmAhrD2ipgfqvUbiJBHL148pHrg1WRKK")
	lsdTokenMint := solana.MustPublicKeyFromBase58("ExvXyxrc98hD7Nd8gTNEMHdLXgkvhRkrtjW7io8qEvmw")
	stakingTokenMint := solana.MustPublicKeyFromBase58("7N3HQ8P73rxgZe426XSXGejxtL5KbUvHcPup2KTKvGWf")

	lsdProgram := solana.MustPublicKeyFromBase58("6UrZH8GHxgSHu13ZqUMxHwiUnezXSqnEKDVNEpY1cAPu")
	svm_lsd_program.SetProgramID(lsdProgram)

	user, err := solana.PrivateKeyFromSolanaKeygenFile("/Users/tpkeeper/.config/solana/id.json")
	if err != nil {
		t.Fatal(err)
	}

	userLsdTokenAccount, _, err := solana.FindAssociatedTokenAddress(user.PublicKey(), lsdTokenMint)
	if err != nil {
		t.Fatal(err)
	}
	userStakingTokenAccount, _, err := solana.FindAssociatedTokenAddress(user.PublicKey(), stakingTokenMint)
	if err != nil {
		t.Fatal(err)
	}
	stakeManagerStakingTokenAccount, _, err := solana.FindAssociatedTokenAddress(stakeManager, stakingTokenMint)
	if err != nil {
		t.Fatal(err)
	}

	instructions := []solana.Instruction{
		computebudget.NewSetComputeUnitPriceInstruction(20000).Build(),
		svm_lsd_program.NewStakeInstruction(1000000000, user.PublicKey(), user.PublicKey(), stakeManager, lsdTokenMint,
			stakingTokenMint, userLsdTokenAccount, userStakingTokenAccount, stakeManagerStakingTokenAccount,
			solana.TokenProgramID, solana.SPLAssociatedTokenAccountProgramID, solana.SystemProgramID).Build()}

	latestBlockHashRes, err := rpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentConfirmed)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := solana.NewTransaction(
		instructions,
		latestBlockHashRes.Value.Blockhash,
		solana.TransactionPayer(user.PublicKey()))
	if err != nil {
		t.Fatal(err)
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if user.PublicKey().Equals(key) {
			return &user
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Stake will send tx: %s", tx.Signatures[0])

	_, err = utils.SendAndWaitForConfirmation(rpcClient, tx, latestBlockHashRes.Value.LastValidBlockHeight)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnstake(t *testing.T) {
	endpoint := "https://api.devnet.solana.com"
	// endpoint := "https://api.testnet.sonic.game"
	user, err := solana.PrivateKeyFromSolanaKeygenFile("/Users/tpkeeper/.config/solana/id.json")
	if err != nil {
		t.Fatal(err)
	}

	stakeManager := solana.MustPublicKeyFromBase58("9MD8swVgHG11rmAhrD2ipgfqvUbiJBHL148pHrg1WRKK")
	lsdTokenMint := solana.MustPublicKeyFromBase58("ExvXyxrc98hD7Nd8gTNEMHdLXgkvhRkrtjW7io8qEvmw")

	lsdProgramID := solana.MustPublicKeyFromBase58("6UrZH8GHxgSHu13ZqUMxHwiUnezXSqnEKDVNEpY1cAPu")

	svm_lsd_program.SetProgramID(lsdProgramID)

	userLsdTokenAccount, _, err := solana.FindAssociatedTokenAddress(user.PublicKey(), lsdTokenMint)
	if err != nil {
		t.Fatal(err)
	}

	unstakeAccount := solana.NewWallet().PrivateKey

	t.Logf("unstakeAccount: %s", unstakeAccount.PublicKey().String())

	instructions := []solana.Instruction{
		computebudget.NewSetComputeUnitPriceInstruction(20000).Build(),
		svm_lsd_program.NewUnstakeInstruction(
			1000000000,
			user.PublicKey(),
			user.PublicKey(),
			stakeManager,
			lsdTokenMint,
			userLsdTokenAccount,
			unstakeAccount.PublicKey(),
			solana.TokenProgramID,
			solana.SPLAssociatedTokenAccountProgramID,
			solana.SystemProgramID,
		).Build(),
	}

	rpcClient := rpc.NewWithCustomRPCClient(rpc.NewWithLimiter(
		endpoint,
		rate.Every(time.Second), // time frame
		5,                       // limit of requests per time frame
	))

	latestBlockHashRes, err := rpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentConfirmed)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := solana.NewTransaction(
		instructions,
		latestBlockHashRes.Value.Blockhash,
		solana.TransactionPayer(user.PublicKey()))
	if err != nil {
		t.Fatal(err)
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if user.PublicKey().Equals(key) {
			return &user
		}
		if unstakeAccount.PublicKey().Equals(key) {
			return &unstakeAccount
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Unstake will send tx: %s", tx.Signatures[0])

	_, err = utils.SendAndWaitForConfirmation(rpcClient, tx, latestBlockHashRes.Value.LastValidBlockHeight)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWithdraw(t *testing.T) {
	endpoint := "https://api.devnet.solana.com"
	rpcClient := rpc.NewWithCustomRPCClient(rpc.NewWithLimiter(
		endpoint,
		rate.Every(time.Second), // time frame
		5,                       // limit of requests per time frame
	))
	user, err := solana.PrivateKeyFromSolanaKeygenFile("/Users/tpkeeper/.config/solana/id.json")
	if err != nil {
		t.Fatal(err)
	}
	stakeManager := solana.MustPublicKeyFromBase58("9MD8swVgHG11rmAhrD2ipgfqvUbiJBHL148pHrg1WRKK")
	lsdProgramID := solana.MustPublicKeyFromBase58("6UrZH8GHxgSHu13ZqUMxHwiUnezXSqnEKDVNEpY1cAPu")
	stakingTokenMint := solana.MustPublicKeyFromBase58("7N3HQ8P73rxgZe426XSXGejxtL5KbUvHcPup2KTKvGWf")

	svm_lsd_program.SetProgramID(lsdProgramID)

	hexBts := append(stakeManager.Bytes(), user.PublicKey().Bytes()...)
	accounts, err := rpcClient.GetProgramAccountsWithOpts(
		context.Background(),
		lsdProgramID,
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
		t.Fatal(err)
	}

	if len(accounts) == 0 {
		return
	}
	unstakeAccount := accounts[0].Pubkey

	poolStakingTokenAccount, _, err := solana.FindAssociatedTokenAddress(stakeManager, stakingTokenMint)
	if err != nil {
		t.Fatal(err)
	}
	userStakingTokenAccount, _, err := solana.FindAssociatedTokenAddress(user.PublicKey(), stakingTokenMint)
	if err != nil {
		t.Fatal(err)
	}
	instructions := []solana.Instruction{
		computebudget.NewSetComputeUnitPriceInstruction(20000).Build(),
		svm_lsd_program.NewWithdrawInstruction(
			user.PublicKey(),
			user.PublicKey(),
			stakeManager,
			unstakeAccount,
			stakingTokenMint,
			userStakingTokenAccount,
			poolStakingTokenAccount,
			solana.TokenProgramID,
			solana.SPLAssociatedTokenAccountProgramID,
			solana.SystemProgramID,
		).Build(),
	}

	latestBlockHashRes, err := rpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentConfirmed)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := solana.NewTransaction(
		instructions,
		latestBlockHashRes.Value.Blockhash,
		solana.TransactionPayer(user.PublicKey()))
	if err != nil {
		t.Fatal(err)
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if user.PublicKey().Equals(key) {
			return &user
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Unstake will send tx: %s", tx.Signatures[0])

	_, err = utils.SendAndWaitForConfirmation(rpcClient, tx, latestBlockHashRes.Value.LastValidBlockHeight)
	if err != nil {
		t.Fatal(err)
	}
}
