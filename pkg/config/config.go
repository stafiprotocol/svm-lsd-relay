package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type ConfigInitStakeManager struct {
	RpcEndpoint  string // rpc endpoint
	KeystorePath string

	LsdProgramID       string
	StakingPoolAddress string

	FeePayerAccount string
	AdminAccount    string

	EraSeconds int64
}

type ConfigTransferAdmin struct {
	RpcEndpoint  string // rpc endpoint
	KeystorePath string

	StakeManagerAddress string

	FeePayerAccount string
	AdminAccount    string
	NewAdminAccount string
}

type ConfigBuildUpgradeTx struct {
	RpcEndpoint string

	LsdProgramID    string
	FeePayerAccount string

	BufferAccount    string
	UpgradeAuthority string
}

type ConfigBuildTransferTokenTx struct {
	RpcEndpoint     string
	FeePayerAccount string

	Mint        string
	FromAddress string
	ToAddress   string
	Amount      float64
}

type ConfigSetStakeManager struct {
	RpcEndpoint  string // rpc endpoint
	KeystorePath string

	StakeManagerAddress string

	FeePayerAccount string
	AdminAccount    string

	MinStakeAmount        *uint64
	PlatformFeeCommission *uint64
	RateChangeLimit       *uint64
}

type ConfigStart struct {
	LogFileDir   string
	RpcEndpoint  string // rpc endpoint
	KeystorePath string

	StakeManagerAddress string

	FeePayerAccount string
}

type ConfigCreateMetadata struct {
	RpcEndpoint  string // rpc endpoint
	KeystorePath string

	StakeManagerAddress string

	FeePayerAccount string
	AdminAccount    string

	// Metadata parameters
	TokenName   string
	TokenSymbol string
	TokenUri    string
}

func LoadConfig[config any](path string) (*config, error) {
	cfg := new(config)
	if err := loadConfig(path, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func loadConfig(path string, config any) error {
	_, err := os.Open(path)
	if err != nil {
		return err
	}
	if _, err := toml.DecodeFile(path, config); err != nil {
		return err
	}
	fmt.Println("load config success")
	return nil
}
