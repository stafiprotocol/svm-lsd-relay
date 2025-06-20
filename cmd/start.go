package cmd

import (
	"encoding/json"
	"fmt"

	"svm-lsd-relay/pkg/config"
	"svm-lsd-relay/pkg/log"
	"svm-lsd-relay/pkg/utils"
	"svm-lsd-relay/pkg/vault"
	"svm-lsd-relay/task"

	"github.com/gagliardetto/solana-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func startCmd() *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "start",
		Short: "Start sonic lsd relay",

		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := cmd.Flags().GetString(flagConfigPath)
			if err != nil {
				return err
			}
			fmt.Printf("Config path: %s\n", configPath)

			cfg, err := config.LoadConfig[config.ConfigStart](configPath)
			if err != nil {
				return err
			}
			if len(cfg.LogFileDir) == 0 {
				cfg.LogFileDir = "./log_data"
			}

			v, err := vault.NewVaultFromWalletFile(cfg.KeystorePath)
			if err != nil {
				return err
			}
			boxer, err := vault.SecretBoxerForType(v.SecretBoxWrap)
			if err != nil {
				return fmt.Errorf("secret boxer: %w", err)
			}

			if err := v.Open(boxer); err != nil {
				return fmt.Errorf("opening: %w", err)
			}

			privateKeyMap := make(map[string]solana.PrivateKey)
			for _, privKey := range v.KeyBag {
				privateKeyMap[privKey.PublicKey().String()] = solana.PrivateKey(privKey)
			}
			feePayerAccount, exist := privateKeyMap[cfg.FeePayerAccount]
			if !exist {
				return fmt.Errorf("fee payer not exit in vault")
			}

			bts, _ := json.MarshalIndent(cfg, "", "  ")
			fmt.Printf("Config: \n%s\n", string(bts))
		Out:
			for {
				fmt.Println("\nCheck config info, then press (y/n) to continue:")
				var input string
				fmt.Scanln(&input)
				switch input {
				case "y":
					break Out
				case "n":
					return nil
				default:
					fmt.Println("press `y` or `n`")
					continue
				}
			}

			logLevelStr, err := cmd.Flags().GetString(flagLogLevel)
			if err != nil {
				return err
			}
			logLevel, err := logrus.ParseLevel(logLevelStr)
			if err != nil {
				return err
			}
			logrus.SetLevel(logLevel)
			err = log.InitLogFile(cfg.LogFileDir + "/relay")
			if err != nil {
				return fmt.Errorf("InitLogFile failed: %w", err)
			}

			ctx := utils.ShutdownListener()

			t := task.NewTask(*cfg, feePayerAccount)
			err = t.Start()
			if err != nil {
				return err
			}
			defer func() {
				logrus.Infof("shutting down task ...")
				t.Stop()
			}()

			<-ctx.Done()

			return nil
		},
	}
	cmd.Flags().String(flagConfigPath, defaultConfigPath, "Config file path")
	cmd.Flags().String(flagLogLevel, logrus.InfoLevel.String(), "The logging level (trace|debug|info|warn|error|fatal|panic)")
	return cmd
}
