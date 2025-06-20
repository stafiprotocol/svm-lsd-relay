package utils

import (
	"github.com/gagliardetto/solana-go"
)

var _ solana.Instruction = &UpgradeInstruction{}

type UpgradeInstruction struct {
	programId          solana.PublicKey
	bufferAddress      solana.PublicKey
	upgradeAuthority   solana.PublicKey
	spillAddress       solana.PublicKey
	programDataAddress solana.PublicKey
}

func NewUpgradeInstruction(programId, bufferAddress, upgradeAuthority, spillAddress solana.PublicKey) (*UpgradeInstruction, error) {
	programDataAddress, _, err := solana.FindProgramAddress([][]byte{programId[:]}, solana.BPFLoaderUpgradeableProgramID)
	if err != nil {
		return nil, err
	}
	return &UpgradeInstruction{
		programId:          programId,
		bufferAddress:      bufferAddress,
		upgradeAuthority:   upgradeAuthority,
		spillAddress:       spillAddress,
		programDataAddress: programDataAddress,
	}, nil
}

func (i *UpgradeInstruction) ProgramID() solana.PublicKey {
	return solana.BPFLoaderUpgradeableProgramID
}
func (i *UpgradeInstruction) Data() ([]byte, error) {
	return []byte{3, 0, 0, 0}, nil
}

func (i *UpgradeInstruction) Accounts() []*solana.AccountMeta {
	return []*solana.AccountMeta{
		{
			PublicKey:  i.programDataAddress,
			IsWritable: true,
			IsSigner:   false,
		},
		{
			PublicKey:  i.programId,
			IsWritable: true,
			IsSigner:   false,
		},
		{
			PublicKey:  i.bufferAddress,
			IsWritable: true,
			IsSigner:   false,
		},
		{
			PublicKey:  i.spillAddress,
			IsWritable: true,
			IsSigner:   false,
		},

		{
			PublicKey:  solana.SysVarRentPubkey,
			IsWritable: false,
			IsSigner:   false,
		},
		{
			PublicKey:  solana.SysVarClockPubkey,
			IsWritable: false,
			IsSigner:   false,
		},
		{
			PublicKey:  i.upgradeAuthority,
			IsWritable: false,
			IsSigner:   true,
		},
	}
}
