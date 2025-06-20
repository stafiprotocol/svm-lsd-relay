package utils_test

import (
	"context"
	"testing"
	"time"

	svm_lsd_program "svm-lsd-relay/pkg/lsd_program"
	"svm-lsd-relay/pkg/staking_program"
	"svm-lsd-relay/pkg/utils"

	"github.com/gagliardetto/solana-go"
	// associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	// computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
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
			RewardAlgorithm:  staking_program.RewardAlgorithmFixedRate,
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
	lsdProgram := solana.MustPublicKeyFromBase58("6UrZH8GHxgSHu13ZqUMxHwiUnezXSqnEKDVNEpY1cAPu")
	stakingPool := solana.MustPublicKeyFromBase58("EAq1c1pqkvNwcggxrK7SKwxzwqmf95gPgCjcFEgv8H1Y")
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

	stakingTokenMint := solana.MustPublicKeyFromBase58("4qVFnsc4WJLo5t4a8guWtbz2K6KatuuMaPWbDFJ9Z2sh")
	var stakeMangerSeed = []byte("stake_manager_seed")
	var tokenMintSeed = []byte("token_mint_seed")
	stakingManager, _, err := solana.FindProgramAddress([][]byte{stakeMangerSeed, user.PublicKey().Bytes(), {0}}, lsdProgram)
	if err != nil {
		t.Fatal(err)
	}
	lsdTokenMint, _, err := solana.FindProgramAddress([][]byte{tokenMintSeed, user.PublicKey().Bytes(), {0}}, lsdProgram)
	if err != nil {
		t.Fatal(err)
	}
	// adminTokenAccount, _, err := solana.FindAssociatedTokenAddress(user.PublicKey(), tokenMint)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	poolTokenAccount, _, err := solana.FindAssociatedTokenAddress(stakingManager, stakingTokenMint)
	if err != nil {
		t.Fatal(err)
	}

	initIns := svm_lsd_program.NewInitializeStakeManagerInstruction(svm_lsd_program.InitializeStakeManagerParams{
		EraSeconds: 60,
		Index:      0,
	}, user.PublicKey(), user.PublicKey(), stakingManager, stakingPool, lsdTokenMint, stakingTokenMint,
		poolTokenAccount, solana.TokenProgramID, solana.SPLAssociatedTokenAccountProgramID, solana.SystemProgramID)
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

// func TestStake(t *testing.T) {
// 	// 	LsdProgramID = "D7reNAipNYYfpEcBaNFiZKBtQmfUvh3FATyB7Qv3tUt9"
// 	// StakeManagerAddress = "9imjEWKsif5mwSdP8HjmyjizSJZ44dx2quuZyVTYCwdz"

// 	// LsdTokenMintAddress = "H43FkMYZ1MBJETxHEzYwZAg8pQU8dESRaF8jAVnUUcow"
// 	// SonicTokenMintAddress = "FxUoL2StchQJhZGhERmG1o335TcVXm7AtHN5QRtbynXU"

// 	// endpoint := "https://api.devnet.solana.com"
// 	endpoint := "https://api.testnet.sonic.game"
// 	rpcClient := rpc.NewWithCustomRPCClient(rpc.NewWithLimiter(
// 		endpoint,
// 		rate.Every(time.Second), // time frame
// 		5,                       // limit of requests per time frame
// 	))

// 	// 	LsdProgramID = "CEftts4KkpMPiJg9hccAgqZHvUc3t1hx9VssMUGkX3ec"
// 	// SonicStakingProgramID = "CjZyANM4ZgPTa3cUfck3DGHvFZdyHmoHTSU2M9P7KKbA"
// 	// SonicStakingPool = "E161foYz3Cn3sbbV3C9nFdS9WjwHFapNZdT5sAvYakxJ"
// 	// StakeManagerAddress = "FrTfbJVT7uHgt4ejMqgCD5vLCgJCk5ewkuJghtzGRGA1"

// 	// LsdTokenMintAddress = "DqXCrewJCUYJ2V8PqwVV18cgn8qm3yJVBmZbtHUEmW7P"

// 	// stakeManager := solana.MustPublicKeyFromBase58("FrTfbJVT7uHgt4ejMqgCD5vLCgJCk5ewkuJghtzGRGA1")
// 	// lsdTokenMint := solana.MustPublicKeyFromBase58("DqXCrewJCUYJ2V8PqwVV18cgn8qm3yJVBmZbtHUEmW7P")
// 	sonicTokenMint := solana.MustPublicKeyFromBase58("3P3vnLByiXcY1aG3vNs888gQFYyFLN5cm6xkCMZsvYio")

// 	lsdProgramID := solana.MustPublicKeyFromBase58("CEftts4KkpMPiJg9hccAgqZHvUc3t1hx9VssMUGkX3ec")

// 	user, err := solana.PrivateKeyFromSolanaKeygenFile("/Users/tpkeeper/.config/solana/staker.json")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	lsd_program.SetProgramID(lsdProgramID)

// 	// stakePool, _, err := solana.FindProgramAddress([][]byte{utils.StakePoolSeed, stakeManager.Bytes()}, lsdProgramID)
// 	// if err != nil {
// 	// 	t.Fatal(err)
// 	// }

// 	// userLsdTokenAccount, _, err := solana.FindAssociatedTokenAddress(user.PublicKey(), lsdTokenMint)
// 	// if err != nil {
// 	// 	t.Fatal(err)
// 	// }

// 	insCreatePoolTokenAccount := associatedtokenaccount.NewCreateInstruction(
// 		user.PublicKey(),
// 		user.PublicKey(),
// 		sonicTokenMint).Build()
// 	insCreatePoolTokenAccount.BaseVariant.Impl.(associatedtokenaccount.Create).AccountMetaSlice[5] = &solana.AccountMeta{
// 		PublicKey:  solana.Token2022ProgramID,
// 		IsSigner:   false,
// 		IsWritable: false,
// 	}
// 	poolAssociatedToken2022Address, _, err := utils.FindAssociatedToken2022Address(user.PublicKey(), sonicTokenMint)
// 	if err != nil {
// 		t.Fatal()
// 	}
// 	insCreatePoolTokenAccount.BaseVariant.Impl.(associatedtokenaccount.Create).AccountMetaSlice[1] = &solana.AccountMeta{
// 		PublicKey:  poolAssociatedToken2022Address,
// 		IsSigner:   false,
// 		IsWritable: true,
// 	}

// 	// userSonicTokenAccount, _, err := solana.FindAssociatedTokenAddress(user.PublicKey(), sonicTokenMint)
// 	// if err != nil {
// 	// 	t.Fatal(err)
// 	// }

// 	// poolSonicTokenAccount, _, err := solana.FindAssociatedTokenAddress(stakePool, sonicTokenMint)
// 	// if err != nil {
// 	// 	t.Fatal(err)
// 	// }

// 	instructions := []solana.Instruction{
// 		computebudget.NewSetComputeUnitPriceInstruction(20000).Build(), insCreatePoolTokenAccount}

// 	// _, err = rpcClient.GetAccountInfo(context.Background(), userLsdTokenAccount)
// 	// if err != nil {
// 	// 	if err == rpc.ErrNotFound {
// 	// 		instructions = append(instructions,
// 	// 			associatedtokenaccount.NewCreateInstruction(user.PublicKey(), user.PublicKey(), lsdTokenMint).Build())
// 	// 	} else {
// 	// 		t.Fatal(err)
// 	// 	}
// 	// }

// 	// instructions = append(instructions, lsd_program.NewStakeInstruction(
// 	// 	20000000000,
// 	// 	user.PublicKey(),
// 	// 	stakeManager,
// 	// 	stakePool,
// 	// 	lsdTokenMint,
// 	// 	sonicTokenMint,
// 	// 	userLsdTokenAccount,
// 	// 	userSonicTokenAccount,
// 	// 	poolSonicTokenAccount,
// 	// 	solana.TokenProgramID,
// 	// 	solana.SPLAssociatedTokenAccountProgramID,
// 	// 	solana.SystemProgramID,
// 	// ).Build())

// 	latestBlockHashRes, err := rpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentConfirmed)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	tx, err := solana.NewTransaction(
// 		instructions,
// 		latestBlockHashRes.Value.Blockhash,
// 		solana.TransactionPayer(user.PublicKey()))
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
// 		if user.PublicKey().Equals(key) {
// 			return &user
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	t.Logf("Stake will send tx: %s", tx.Signatures[0])

// 	_, err = utils.SendAndWaitForConfirmation(rpcClient, tx, latestBlockHashRes.Value.LastValidBlockHeight)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }

// func TestUnStake(t *testing.T) {
// 	// endpoint := "https://api.devnet.solana.com"
// 	endpoint := "https://api.testnet.sonic.game"
// 	user, err := solana.PrivateKeyFromSolanaKeygenFile("/Users/tpkeeper/.config/solana/staker.json")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	stakeManager := solana.MustPublicKeyFromBase58("9imjEWKsif5mwSdP8HjmyjizSJZ44dx2quuZyVTYCwdz")
// 	lsdTokenMint := solana.MustPublicKeyFromBase58("H43FkMYZ1MBJETxHEzYwZAg8pQU8dESRaF8jAVnUUcow")
// 	// sonicTokenMint := solana.MustPublicKeyFromBase58("FxUoL2StchQJhZGhERmG1o335TcVXm7AtHN5QRtbynXU")

// 	lsdProgramID := solana.MustPublicKeyFromBase58("D7reNAipNYYfpEcBaNFiZKBtQmfUvh3FATyB7Qv3tUt9")

// 	// lsdProgramID := solana.MustPublicKeyFromBase58("CEftts4KkpMPiJg9hccAgqZHvUc3t1hx9VssMUGkX3ec")
// 	lsd_program.SetProgramID(lsdProgramID)
// 	// stakeManager := solana.MustPublicKeyFromBase58("921ksHLP7Hc4tENgziqz4xZFCiwzzuii6HEwUm9EnB1H")
// 	// lsdTokenMint := solana.MustPublicKeyFromBase58("6HbNJxx813dzV7or2y8xRjVf2kLZnrPCXaFunzGeeofW")

// 	userLsdTokenAccount, _, err := solana.FindAssociatedTokenAddress(user.PublicKey(), lsdTokenMint)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	unstakeAccount := solana.NewWallet().PrivateKey

// 	t.Logf("unstakeAccount: %s", unstakeAccount.String())

// 	instructions := []solana.Instruction{
// 		computebudget.NewSetComputeUnitPriceInstruction(20000).Build(),
// 		lsd_program.NewUnstakeInstruction(
// 			100000000000,
// 			user.PublicKey(),
// 			user.PublicKey(),
// 			stakeManager,
// 			lsdTokenMint,
// 			userLsdTokenAccount,
// 			unstakeAccount.PublicKey(),
// 			solana.TokenProgramID,
// 			solana.SPLAssociatedTokenAccountProgramID,
// 			solana.SystemProgramID,
// 			solana.SysVarRentPubkey,
// 		).Build(),
// 	}

// 	rpcClient := rpc.NewWithCustomRPCClient(rpc.NewWithLimiter(
// 		endpoint,
// 		rate.Every(time.Second), // time frame
// 		5,                       // limit of requests per time frame
// 	))

// 	latestBlockHashRes, err := rpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentConfirmed)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	tx, err := solana.NewTransaction(
// 		instructions,
// 		latestBlockHashRes.Value.Blockhash,
// 		solana.TransactionPayer(user.PublicKey()))
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
// 		if user.PublicKey().Equals(key) {
// 			return &user
// 		}
// 		if unstakeAccount.PublicKey().Equals(key) {
// 			return &unstakeAccount
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	t.Logf("Unstake will send tx: %s", tx.Signatures[0])

// 	_, err = utils.SendAndWaitForConfirmation(rpcClient, tx, latestBlockHashRes.Value.LastValidBlockHeight)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }

// func TestWithdraw(t *testing.T) {
// 	endpoint := "https://api.devnet.solana.com"
// 	user, err := solana.PrivateKeyFromSolanaKeygenFile("/Users/tpkeeper/.config/solana/staker.json")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	lsdProgramID := solana.MustPublicKeyFromBase58("CEftts4KkpMPiJg9hccAgqZHvUc3t1hx9VssMUGkX3ec")
// 	lsd_program.SetProgramID(lsdProgramID)
// 	stakeManager := solana.MustPublicKeyFromBase58("921ksHLP7Hc4tENgziqz4xZFCiwzzuii6HEwUm9EnB1H")
// 	stakePool, _, err := solana.FindProgramAddress([][]byte{utils.StakePoolSeed, stakeManager.Bytes()}, lsdProgramID)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	userSonicTokenAccount := solana.MustPublicKeyFromBase58("ESFVYpoy9bnXmKzyFdmaCZTpc3WPcKQA8R7Tg8J79s8V")
// 	sonicTokenMint := solana.MustPublicKeyFromBase58("2PtjgnsDgzTCbaD9rRYFNeVTRREPLqk6cwsFm8pbncXe")

// 	unstakeAccount := solana.MustPublicKeyFromBase58("z1RkpMaCTAgSsow6ad5kKrwWFe9xnh2EJT4ZXjJtkPa")
// 	poolSonicTokenAccount, _, err := solana.FindAssociatedTokenAddress(stakePool, sonicTokenMint)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	instructions := []solana.Instruction{
// 		computebudget.NewSetComputeUnitPriceInstruction(20000).Build(),
// 		lsd_program.NewWithdrawInstruction(
// 			user.PublicKey(),
// 			stakeManager,
// 			stakePool,
// 			unstakeAccount,
// 			sonicTokenMint,
// 			userSonicTokenAccount,
// 			poolSonicTokenAccount,
// 			solana.TokenProgramID,
// 			solana.SPLAssociatedTokenAccountProgramID,
// 			solana.SystemProgramID,
// 		).Build(),
// 	}

// 	rpcClient := rpc.NewWithCustomRPCClient(rpc.NewWithLimiter(
// 		endpoint,
// 		rate.Every(time.Second), // time frame
// 		5,                       // limit of requests per time frame
// 	))

// 	latestBlockHashRes, err := rpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentConfirmed)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	tx, err := solana.NewTransaction(
// 		instructions,
// 		latestBlockHashRes.Value.Blockhash,
// 		solana.TransactionPayer(user.PublicKey()))
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
// 		if user.PublicKey().Equals(key) {
// 			return &user
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	t.Logf("Unstake will send tx: %s", tx.Signatures[0])

// 	_, err = utils.SendAndWaitForConfirmation(rpcClient, tx, latestBlockHashRes.Value.LastValidBlockHeight)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }
