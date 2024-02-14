package cmd

import (
	"bytes"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func ChainInitCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:  "init",
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			count, err := strconv.Atoi(args[0])
			if err != nil {
				panic(err)
			}

			balAmount := args[1]
			stakeAmount := args[2]

			var (
				binary           = "cantod"
				mnemonicList     []string
				rawValidatorList RawValidatorList

				exportFilePath = "vali-info.yaml"
			)

			for i := 0; i < count; i++ {
				moniker := "validator-" + strconv.Itoa(i)

				var initBuffer bytes.Buffer
				initCmd := exec.Command(binary, "--home", moniker, "init", "--chain-id", "canto_7700-1", moniker)
				initCmd.Stdout = &initBuffer
				initCmd.Stderr = &initBuffer
				if err = initCmd.Run(); err != nil {
					panic(fmt.Errorf("%s\n", initBuffer.String()))
				}

				generateMnemonicBuffer := new(bytes.Buffer)
				// Generate mnemonic
				generateMnemonicCmd := exec.Command(binary, "keys", "mnemonic")
				generateMnemonicCmd.Stdout = generateMnemonicBuffer
				generateMnemonicCmd.Stderr = generateMnemonicBuffer

				if err = generateMnemonicCmd.Run(); err != nil {
					panic(fmt.Errorf("%s\n", generateMnemonicBuffer.String()))
				}

				mnemonic := generateMnemonicBuffer.String()
				mnemonicList = append(mnemonicList, mnemonic)

				var mnemonicBuffer bytes.Buffer
				_, err = mnemonicBuffer.Write([]byte(mnemonic))
				if err != nil {
					panic(err)
				}

				//keysAddBuffer, keysAddCmdW, err := os.Pipe()
				var keysAddBuffer bytes.Buffer
				if err != nil {
					panic(err)
				}
				keysAddCmd := exec.Command(binary, "--home", moniker,
					"keys", "add", moniker, "--recover", "--keyring-backend", "test", "--output", "json")
				keysAddCmd.Stdin = &mnemonicBuffer
				keysAddCmd.Stdout = &keysAddBuffer
				keysAddCmd.Stderr = &keysAddBuffer

				if err = keysAddCmd.Run(); err != nil {
					panic(fmt.Errorf("%s\n", keysAddBuffer.String()))
				}

				var jqAddressBuffer bytes.Buffer
				jqAddress := exec.Command("jq", "-r", ".address")
				jqAddress.Stdin = &keysAddBuffer
				jqAddress.Stdout = &jqAddressBuffer
				jqAddress.Stderr = &jqAddressBuffer

				if err = jqAddress.Run(); err != nil {
					panic(fmt.Errorf("%s\n", jqAddressBuffer.String()))
				}

				address := jqAddressBuffer.String()

				var validatorKeyBuffer bytes.Buffer
				validatorKeyCmd := exec.Command("cantod", "tendermint", "show-validator", "--home", moniker)
				validatorKeyCmd.Stdout = &validatorKeyBuffer
				validatorKeyCmd.Stderr = &validatorKeyBuffer

				if err = validatorKeyCmd.Run(); err != nil {
					panic(fmt.Errorf("%s\n", validatorKeyBuffer.String()))
				}

				validatorKey := validatorKeyBuffer.String()

				//// Read priv_validator_key.json
				//privValidatorKeyBytes, err := os.ReadFile(fmt.Sprintf("%s/config/priv_validator_key.json", moniker))
				//if err != nil {
				//	return err
				//}
				//
				//var privValidatorKey PrivValidatorKey
				//if err = json.Unmarshal(privValidatorKeyBytes, &privValidatorKey); err != nil {
				//	return err
				//}

				// Cut prefix '\n'
				address, _ = strings.CutPrefix(address, "\n")
				address, _ = strings.CutSuffix(address, "\n")
				mnemonic, _ = strings.CutPrefix(mnemonic, "\n")
				mnemonic, _ = strings.CutSuffix(mnemonic, "\n")
				validatorKey, _ = strings.CutPrefix(validatorKey, "\n")
				validatorKey, _ = strings.CutSuffix(validatorKey, "\n")

				fmt.Printf("Moniker: %s \nmnemonic: %s \naddress: %s\n", moniker, mnemonic, address)

				// Generate RawValidator
				rawValidatorList = append(rawValidatorList, RawValidator{
					Moniker:      moniker,
					Address:      address,
					BalAmount:    balAmount,
					StakeAmount:  stakeAmount,
					ValidatorKey: validatorKey,
					Mnemonic:     mnemonic,
				})
			}

			marshal, err := yaml.Marshal(rawValidatorList)
			if err != nil {
				panic(err)
			}

			err = os.WriteFile(exportFilePath, marshal, 755)
			if err != nil {
				panic(err)
			}

			return nil
		},
	}

	return &cmd
}
