package cmd

import (
	"bytes"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"strconv"
)

func ChainInitCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:  "init",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			count, err := strconv.Atoi(args[0])
			if err != nil {
				panic(err)
			}

			var (
				binary           = "cantod"
				balAmount        = ""
				stakeAmount      = ""
				mnemonicList     []string
				rawValidatorList RawValidatorList

				exportFilePath = "vali-info.yaml"
			)

			for i := 0; i < count; i++ {
				moniker := "validator-" + strconv.Itoa(i)

				var initCmdBuffer bytes.Buffer
				initCmd := exec.Command(binary, "--home", moniker, "init", "--chain-id", "canto_7700-1", moniker)
				initCmd.Stdout = &initCmdBuffer
				initCmd.Stderr = &initCmdBuffer
				if err = initCmd.Run(); err != nil {
					panic(fmt.Errorf("%s\n", initCmdBuffer.String()))
				}

				generateMnemonicCmdBuffer := new(bytes.Buffer)
				// Generate mnemonic
				generateMnemonicCmd := exec.Command(binary, "keys", "mnemonic")
				generateMnemonicCmd.Stdout = generateMnemonicCmdBuffer
				generateMnemonicCmd.Stderr = generateMnemonicCmdBuffer

				if err = generateMnemonicCmd.Run(); err != nil {
					panic(fmt.Errorf("%s\n", generateMnemonicCmdBuffer.String()))
				}

				mnemonic := generateMnemonicCmdBuffer.String()
				mnemonicList = append(mnemonicList, mnemonic)

				var mnemonicBuffer bytes.Buffer
				_, err = mnemonicBuffer.Write([]byte(mnemonic))
				if err != nil {
					panic(err)
				}

				//keysAddCmdBuffer, keysAddCmdW, err := os.Pipe()
				var keysAddCmdBuffer bytes.Buffer
				if err != nil {
					panic(err)
				}
				keysAddCmd := exec.Command(binary, "--home", moniker,
					"keys", "add", moniker, "--recover", "--keyring-backend", "test", "--output", "json")
				keysAddCmd.Stdin = &mnemonicBuffer
				keysAddCmd.Stdout = &keysAddCmdBuffer
				keysAddCmd.Stderr = &keysAddCmdBuffer

				if err = keysAddCmd.Run(); err != nil {
					panic(fmt.Errorf("%s\n", keysAddCmdBuffer.String()))
				}

				var jqAddressBuffer bytes.Buffer
				jqAddress := exec.Command("jq", "-r", ".address")
				jqAddress.Stdin = &keysAddCmdBuffer
				jqAddress.Stdout = &jqAddressBuffer
				jqAddress.Stderr = &jqAddressBuffer

				if err = jqAddress.Run(); err != nil {
					panic(fmt.Errorf("%s\n", jqAddressBuffer.String()))
				}

				address := jqAddressBuffer.String()
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

				fmt.Printf("Moniker: %s \nmnemonic: %s \naddress: %s\n", moniker, mnemonic, address)

				// Generate RawValidator
				rawValidatorList = append(rawValidatorList, RawValidator{
					Moniker:       moniker,
					Address:       address,
					BalAmount:     balAmount,
					StakeAmount:   stakeAmount,
					PublicKeyPath: fmt.Sprintf("%s/config/priv_validator_key.json", moniker),
					Mnemonic:      mnemonic,
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
