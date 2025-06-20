package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"svm-lsd-relay/pkg/utils"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/spf13/cobra"
	"golang.org/x/time/rate"
)

func stakeManagerDetailCmd() *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "detail",
		Short: "Get stake manager detail",

		RunE: func(cmd *cobra.Command, args []string) error {

			stakeManager, err := cmd.Flags().GetString(flagStakeManager)
			if err != nil {
				return err
			}

			stakeManagerPubkey := solana.MustPublicKeyFromBase58(stakeManager)

			endpoint, err := cmd.Flags().GetString(flagRpcEndPoint)
			if err != nil {
				return err
			}

			rpcClient := rpc.NewWithCustomRPCClient(rpc.NewWithLimiter(
				endpoint,
				rate.Every(time.Second), // time frame
				5,                       // limit of requests per time frame
			))

			stakeManagerDetail, err := utils.GetSvmLsdStakeManager(rpcClient, stakeManagerPubkey)
			if err != nil {
				return err
			}

			jsonBts, err := json.MarshalIndent(stakeManagerDetail, "", "  ")
			if err != nil {
				return err
			}

			fmt.Printf("stakeManager: \n%s\n", string(jsonBts))
			return nil
		},
	}
	cmd.Flags().String(flagStakeManager, "", "stake manager")
	cmd.Flags().String(flagRpcEndPoint, "", "solana rpc endpoint")
	return cmd
}
