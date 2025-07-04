// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package lsd_program

import (
	"errors"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Withdraw is the `withdraw` instruction.
type Withdraw struct {

	// [0] = [SIGNER] user
	//
	// [1] = [WRITE, SIGNER] rentPayer
	//
	// [2] = [WRITE] stakeManager
	//
	// [3] = [WRITE] unstakeAccount
	//
	// [4] = [] stakingTokenMint
	//
	// [5] = [WRITE] userStakingTokenAccount
	//
	// [6] = [WRITE] stakeManagerStakingTokenAccount
	//
	// [7] = [] tokenProgram
	//
	// [8] = [] associatedTokenProgram
	//
	// [9] = [] systemProgram
	ag_solanago.AccountMetaSlice `bin:"-"`
}

// NewWithdrawInstructionBuilder creates a new `Withdraw` instruction builder.
func NewWithdrawInstructionBuilder() *Withdraw {
	nd := &Withdraw{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 10),
	}
	return nd
}

// SetUserAccount sets the "user" account.
func (inst *Withdraw) SetUserAccount(user ag_solanago.PublicKey) *Withdraw {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(user).SIGNER()
	return inst
}

// GetUserAccount gets the "user" account.
func (inst *Withdraw) GetUserAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(0)
}

// SetRentPayerAccount sets the "rentPayer" account.
func (inst *Withdraw) SetRentPayerAccount(rentPayer ag_solanago.PublicKey) *Withdraw {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(rentPayer).WRITE().SIGNER()
	return inst
}

// GetRentPayerAccount gets the "rentPayer" account.
func (inst *Withdraw) GetRentPayerAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(1)
}

// SetStakeManagerAccount sets the "stakeManager" account.
func (inst *Withdraw) SetStakeManagerAccount(stakeManager ag_solanago.PublicKey) *Withdraw {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(stakeManager).WRITE()
	return inst
}

// GetStakeManagerAccount gets the "stakeManager" account.
func (inst *Withdraw) GetStakeManagerAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(2)
}

// SetUnstakeAccountAccount sets the "unstakeAccount" account.
func (inst *Withdraw) SetUnstakeAccountAccount(unstakeAccount ag_solanago.PublicKey) *Withdraw {
	inst.AccountMetaSlice[3] = ag_solanago.Meta(unstakeAccount).WRITE()
	return inst
}

// GetUnstakeAccountAccount gets the "unstakeAccount" account.
func (inst *Withdraw) GetUnstakeAccountAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(3)
}

// SetStakingTokenMintAccount sets the "stakingTokenMint" account.
func (inst *Withdraw) SetStakingTokenMintAccount(stakingTokenMint ag_solanago.PublicKey) *Withdraw {
	inst.AccountMetaSlice[4] = ag_solanago.Meta(stakingTokenMint)
	return inst
}

// GetStakingTokenMintAccount gets the "stakingTokenMint" account.
func (inst *Withdraw) GetStakingTokenMintAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(4)
}

// SetUserStakingTokenAccountAccount sets the "userStakingTokenAccount" account.
func (inst *Withdraw) SetUserStakingTokenAccountAccount(userStakingTokenAccount ag_solanago.PublicKey) *Withdraw {
	inst.AccountMetaSlice[5] = ag_solanago.Meta(userStakingTokenAccount).WRITE()
	return inst
}

// GetUserStakingTokenAccountAccount gets the "userStakingTokenAccount" account.
func (inst *Withdraw) GetUserStakingTokenAccountAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(5)
}

// SetStakeManagerStakingTokenAccountAccount sets the "stakeManagerStakingTokenAccount" account.
func (inst *Withdraw) SetStakeManagerStakingTokenAccountAccount(stakeManagerStakingTokenAccount ag_solanago.PublicKey) *Withdraw {
	inst.AccountMetaSlice[6] = ag_solanago.Meta(stakeManagerStakingTokenAccount).WRITE()
	return inst
}

// GetStakeManagerStakingTokenAccountAccount gets the "stakeManagerStakingTokenAccount" account.
func (inst *Withdraw) GetStakeManagerStakingTokenAccountAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(6)
}

// SetTokenProgramAccount sets the "tokenProgram" account.
func (inst *Withdraw) SetTokenProgramAccount(tokenProgram ag_solanago.PublicKey) *Withdraw {
	inst.AccountMetaSlice[7] = ag_solanago.Meta(tokenProgram)
	return inst
}

// GetTokenProgramAccount gets the "tokenProgram" account.
func (inst *Withdraw) GetTokenProgramAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(7)
}

// SetAssociatedTokenProgramAccount sets the "associatedTokenProgram" account.
func (inst *Withdraw) SetAssociatedTokenProgramAccount(associatedTokenProgram ag_solanago.PublicKey) *Withdraw {
	inst.AccountMetaSlice[8] = ag_solanago.Meta(associatedTokenProgram)
	return inst
}

// GetAssociatedTokenProgramAccount gets the "associatedTokenProgram" account.
func (inst *Withdraw) GetAssociatedTokenProgramAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(8)
}

// SetSystemProgramAccount sets the "systemProgram" account.
func (inst *Withdraw) SetSystemProgramAccount(systemProgram ag_solanago.PublicKey) *Withdraw {
	inst.AccountMetaSlice[9] = ag_solanago.Meta(systemProgram)
	return inst
}

// GetSystemProgramAccount gets the "systemProgram" account.
func (inst *Withdraw) GetSystemProgramAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(9)
}

func (inst Withdraw) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_Withdraw,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst Withdraw) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *Withdraw) Validate() error {
	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.User is not set")
		}
		if inst.AccountMetaSlice[1] == nil {
			return errors.New("accounts.RentPayer is not set")
		}
		if inst.AccountMetaSlice[2] == nil {
			return errors.New("accounts.StakeManager is not set")
		}
		if inst.AccountMetaSlice[3] == nil {
			return errors.New("accounts.UnstakeAccount is not set")
		}
		if inst.AccountMetaSlice[4] == nil {
			return errors.New("accounts.StakingTokenMint is not set")
		}
		if inst.AccountMetaSlice[5] == nil {
			return errors.New("accounts.UserStakingTokenAccount is not set")
		}
		if inst.AccountMetaSlice[6] == nil {
			return errors.New("accounts.StakeManagerStakingTokenAccount is not set")
		}
		if inst.AccountMetaSlice[7] == nil {
			return errors.New("accounts.TokenProgram is not set")
		}
		if inst.AccountMetaSlice[8] == nil {
			return errors.New("accounts.AssociatedTokenProgram is not set")
		}
		if inst.AccountMetaSlice[9] == nil {
			return errors.New("accounts.SystemProgram is not set")
		}
	}
	return nil
}

func (inst *Withdraw) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("Withdraw")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=0]").ParentFunc(func(paramsBranch ag_treeout.Branches) {})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=10]").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("                    user", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(ag_format.Meta("               rentPayer", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(ag_format.Meta("            stakeManager", inst.AccountMetaSlice.Get(2)))
						accountsBranch.Child(ag_format.Meta("                 unstake", inst.AccountMetaSlice.Get(3)))
						accountsBranch.Child(ag_format.Meta("        stakingTokenMint", inst.AccountMetaSlice.Get(4)))
						accountsBranch.Child(ag_format.Meta("        userStakingToken", inst.AccountMetaSlice.Get(5)))
						accountsBranch.Child(ag_format.Meta("stakeManagerStakingToken", inst.AccountMetaSlice.Get(6)))
						accountsBranch.Child(ag_format.Meta("            tokenProgram", inst.AccountMetaSlice.Get(7)))
						accountsBranch.Child(ag_format.Meta("  associatedTokenProgram", inst.AccountMetaSlice.Get(8)))
						accountsBranch.Child(ag_format.Meta("           systemProgram", inst.AccountMetaSlice.Get(9)))
					})
				})
		})
}

func (obj Withdraw) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return nil
}
func (obj *Withdraw) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	return nil
}

// NewWithdrawInstruction declares a new Withdraw instruction with the provided parameters and accounts.
func NewWithdrawInstruction(
	// Accounts:
	user ag_solanago.PublicKey,
	rentPayer ag_solanago.PublicKey,
	stakeManager ag_solanago.PublicKey,
	unstakeAccount ag_solanago.PublicKey,
	stakingTokenMint ag_solanago.PublicKey,
	userStakingTokenAccount ag_solanago.PublicKey,
	stakeManagerStakingTokenAccount ag_solanago.PublicKey,
	tokenProgram ag_solanago.PublicKey,
	associatedTokenProgram ag_solanago.PublicKey,
	systemProgram ag_solanago.PublicKey) *Withdraw {
	return NewWithdrawInstructionBuilder().
		SetUserAccount(user).
		SetRentPayerAccount(rentPayer).
		SetStakeManagerAccount(stakeManager).
		SetUnstakeAccountAccount(unstakeAccount).
		SetStakingTokenMintAccount(stakingTokenMint).
		SetUserStakingTokenAccountAccount(userStakingTokenAccount).
		SetStakeManagerStakingTokenAccountAccount(stakeManagerStakingTokenAccount).
		SetTokenProgramAccount(tokenProgram).
		SetAssociatedTokenProgramAccount(associatedTokenProgram).
		SetSystemProgramAccount(systemProgram)
}
