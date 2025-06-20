// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package lsd_program

import (
	"errors"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// EraNew is the `eraNew` instruction.
type EraNew struct {

	// [0] = [WRITE] stakeManager
	ag_solanago.AccountMetaSlice `bin:"-"`
}

// NewEraNewInstructionBuilder creates a new `EraNew` instruction builder.
func NewEraNewInstructionBuilder() *EraNew {
	nd := &EraNew{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 1),
	}
	return nd
}

// SetStakeManagerAccount sets the "stakeManager" account.
func (inst *EraNew) SetStakeManagerAccount(stakeManager ag_solanago.PublicKey) *EraNew {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(stakeManager).WRITE()
	return inst
}

// GetStakeManagerAccount gets the "stakeManager" account.
func (inst *EraNew) GetStakeManagerAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(0)
}

func (inst EraNew) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_EraNew,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst EraNew) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *EraNew) Validate() error {
	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.StakeManager is not set")
		}
	}
	return nil
}

func (inst *EraNew) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("EraNew")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=0]").ParentFunc(func(paramsBranch ag_treeout.Branches) {})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=1]").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("stakeManager", inst.AccountMetaSlice.Get(0)))
					})
				})
		})
}

func (obj EraNew) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return nil
}
func (obj *EraNew) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	return nil
}

// NewEraNewInstruction declares a new EraNew instruction with the provided parameters and accounts.
func NewEraNewInstruction(
	// Accounts:
	stakeManager ag_solanago.PublicKey) *EraNew {
	return NewEraNewInstructionBuilder().
		SetStakeManagerAccount(stakeManager)
}
