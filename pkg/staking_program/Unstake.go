// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package staking_program

import (
	"errors"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Unstake is the `unstake` instruction.
type Unstake struct {
	UnstakeAmount *uint64

	// [0] = [SIGNER] user
	//
	// [1] = [WRITE, SIGNER] rentPayer
	//
	// [2] = [WRITE] stakingPool
	//
	// [3] = [WRITE] stakeAccount
	//
	// [4] = [WRITE, SIGNER] unstakeAccount
	//
	// [5] = [] systemProgram
	ag_solanago.AccountMetaSlice `bin:"-"`
}

// NewUnstakeInstructionBuilder creates a new `Unstake` instruction builder.
func NewUnstakeInstructionBuilder() *Unstake {
	nd := &Unstake{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 6),
	}
	return nd
}

// SetUnstakeAmount sets the "unstakeAmount" parameter.
func (inst *Unstake) SetUnstakeAmount(unstakeAmount uint64) *Unstake {
	inst.UnstakeAmount = &unstakeAmount
	return inst
}

// SetUserAccount sets the "user" account.
func (inst *Unstake) SetUserAccount(user ag_solanago.PublicKey) *Unstake {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(user).SIGNER()
	return inst
}

// GetUserAccount gets the "user" account.
func (inst *Unstake) GetUserAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(0)
}

// SetRentPayerAccount sets the "rentPayer" account.
func (inst *Unstake) SetRentPayerAccount(rentPayer ag_solanago.PublicKey) *Unstake {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(rentPayer).WRITE().SIGNER()
	return inst
}

// GetRentPayerAccount gets the "rentPayer" account.
func (inst *Unstake) GetRentPayerAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(1)
}

// SetStakingPoolAccount sets the "stakingPool" account.
func (inst *Unstake) SetStakingPoolAccount(stakingPool ag_solanago.PublicKey) *Unstake {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(stakingPool).WRITE()
	return inst
}

// GetStakingPoolAccount gets the "stakingPool" account.
func (inst *Unstake) GetStakingPoolAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(2)
}

// SetStakeAccountAccount sets the "stakeAccount" account.
func (inst *Unstake) SetStakeAccountAccount(stakeAccount ag_solanago.PublicKey) *Unstake {
	inst.AccountMetaSlice[3] = ag_solanago.Meta(stakeAccount).WRITE()
	return inst
}

// GetStakeAccountAccount gets the "stakeAccount" account.
func (inst *Unstake) GetStakeAccountAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(3)
}

// SetUnstakeAccountAccount sets the "unstakeAccount" account.
func (inst *Unstake) SetUnstakeAccountAccount(unstakeAccount ag_solanago.PublicKey) *Unstake {
	inst.AccountMetaSlice[4] = ag_solanago.Meta(unstakeAccount).WRITE().SIGNER()
	return inst
}

// GetUnstakeAccountAccount gets the "unstakeAccount" account.
func (inst *Unstake) GetUnstakeAccountAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(4)
}

// SetSystemProgramAccount sets the "systemProgram" account.
func (inst *Unstake) SetSystemProgramAccount(systemProgram ag_solanago.PublicKey) *Unstake {
	inst.AccountMetaSlice[5] = ag_solanago.Meta(systemProgram)
	return inst
}

// GetSystemProgramAccount gets the "systemProgram" account.
func (inst *Unstake) GetSystemProgramAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(5)
}

func (inst Unstake) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_Unstake,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst Unstake) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *Unstake) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.UnstakeAmount == nil {
			return errors.New("UnstakeAmount parameter is not set")
		}
	}

	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.User is not set")
		}
		if inst.AccountMetaSlice[1] == nil {
			return errors.New("accounts.RentPayer is not set")
		}
		if inst.AccountMetaSlice[2] == nil {
			return errors.New("accounts.StakingPool is not set")
		}
		if inst.AccountMetaSlice[3] == nil {
			return errors.New("accounts.StakeAccount is not set")
		}
		if inst.AccountMetaSlice[4] == nil {
			return errors.New("accounts.UnstakeAccount is not set")
		}
		if inst.AccountMetaSlice[5] == nil {
			return errors.New("accounts.SystemProgram is not set")
		}
	}
	return nil
}

func (inst *Unstake) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("Unstake")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=1]").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("UnstakeAmount", *inst.UnstakeAmount))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=6]").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("         user", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(ag_format.Meta("    rentPayer", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(ag_format.Meta("  stakingPool", inst.AccountMetaSlice.Get(2)))
						accountsBranch.Child(ag_format.Meta("        stake", inst.AccountMetaSlice.Get(3)))
						accountsBranch.Child(ag_format.Meta("      unstake", inst.AccountMetaSlice.Get(4)))
						accountsBranch.Child(ag_format.Meta("systemProgram", inst.AccountMetaSlice.Get(5)))
					})
				})
		})
}

func (obj Unstake) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `UnstakeAmount` param:
	err = encoder.Encode(obj.UnstakeAmount)
	if err != nil {
		return err
	}
	return nil
}
func (obj *Unstake) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `UnstakeAmount`:
	err = decoder.Decode(&obj.UnstakeAmount)
	if err != nil {
		return err
	}
	return nil
}

// NewUnstakeInstruction declares a new Unstake instruction with the provided parameters and accounts.
func NewUnstakeInstruction(
	// Parameters:
	unstakeAmount uint64,
	// Accounts:
	user ag_solanago.PublicKey,
	rentPayer ag_solanago.PublicKey,
	stakingPool ag_solanago.PublicKey,
	stakeAccount ag_solanago.PublicKey,
	unstakeAccount ag_solanago.PublicKey,
	systemProgram ag_solanago.PublicKey) *Unstake {
	return NewUnstakeInstructionBuilder().
		SetUnstakeAmount(unstakeAmount).
		SetUserAccount(user).
		SetRentPayerAccount(rentPayer).
		SetStakingPoolAccount(stakingPool).
		SetStakeAccountAccount(stakeAccount).
		SetUnstakeAccountAccount(unstakeAccount).
		SetSystemProgramAccount(systemProgram)
}
