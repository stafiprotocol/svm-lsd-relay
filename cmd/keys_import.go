// Copyright 2020 dfuse Platform Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"svm-lsd-relay/pkg/vault"

	"github.com/spf13/cobra"
)

func vaultImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import private keys taking input from the shell",
		RunE: func(cmd *cobra.Command, args []string) error {
			walletFile, err := cmd.Flags().GetString(flagKeystorePath)
			if err != nil {
				return err
			}

			v, boxer := vault.MustGetWallet(cmd, true)
			if len(v.KeyBag) > 0 {
				v.PrintPublicKeys()
			}

			privateKeys, err := capturePrivateKeys()
			if err != nil {
				fmt.Printf("failed to enter private keys: %s", err)
				return err
			}
			if len(privateKeys) == 0 {
				fmt.Println("quit: no private keys")
				return nil
			}

			var newKeys []vault.PublicKey
			for _, privateKey := range privateKeys {
				v.AddPrivateKey(privateKey)
				newKeys = append(newKeys, privateKey.PublicKey())
			}

			if err = v.Seal(vault.CreateBoxerIfNeeded(boxer)); err != nil {
				fmt.Printf("failed to seal vault: %s", err)
				return err
			}

			err = v.WriteToFile(walletFile)
			if err != nil {
				fmt.Printf("failed to write vault file: %s", err)
				return err
			}

			vault.WrittenReport(walletFile, newKeys, len(v.KeyBag))
			return nil
		},
	}

	cmd.Flags().StringP(flagKeystorePath, "", defaultKeystorePath, "Wallet file that contains encrypted key material")
	return cmd
}

func capturePrivateKeys() (out []vault.PrivateKey, err error) {
	fmt.Println("")
	fmt.Println("PLEASE READ:")
	fmt.Println("We are now going to ask you to paste your private keys, one at a time.")
	fmt.Println("They will not be shown on screen.")
	fmt.Println("Please verify that the public keys printed on screen correspond to what you have noted")
	fmt.Println("")

	first := true
	for {
		privKey, err := capturePrivateKey(first)
		if err != nil {
			return out, fmt.Errorf("capture privkeys: %s", err)
		}
		first = false

		if privKey == nil {
			return out, nil
		}
		out = append(out, privKey)
	}
}

func capturePrivateKey(isFirst bool) (privateKey vault.PrivateKey, err error) {
	prompt := "Paste your first private key: "
	if !isFirst {
		prompt = "Paste your next private key or hit ENTER if you are done: "
	}

	enteredKey, err := vault.GetPassword(prompt)
	if err != nil {
		return nil, fmt.Errorf("get private key: %s", err)
	}

	if enteredKey == "" {
		return nil, nil
	}

	key, err := vault.PrivateKeyFromBase58(enteredKey)
	if err != nil {
		return nil, fmt.Errorf("import private key: %s", err)
	}

	fmt.Printf("- Scanned private key corresponding to %s\n", key.PublicKey().String())

	return key, nil
}
