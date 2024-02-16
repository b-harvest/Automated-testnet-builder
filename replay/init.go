package replay

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

			err = ChainInit(count, balAmount, stakeAmount, "", "", "")
			if err != nil {
				return err
			}
			return nil
		},
	}

	return &cmd
}

func ChainInit(count int, balAmount, stakeAmount, homePrefix, exportFilePath, monikerPrefix string) error {
	var (
		err              error
		binary           = "cantod"
		mnemonicList     []string
		rawValidatorList RawValidatorList
	)

	for i := 0; i < count; i++ {
		if monikerPrefix == "" {
			monikerPrefix = "validator-"
		}
		moniker := monikerPrefix + strconv.Itoa(i)

		if homePrefix == "" {
			homePrefix = monikerPrefix
		}
		if exportFilePath == "" {
			exportFilePath = "vali-info.yaml"
		}

		homePath := homePrefix + strconv.Itoa(i)

		var initBuffer bytes.Buffer
		initCmd := exec.Command(binary, "--home", homePath, "init", "--chain-id", "canto_7700-1", moniker)
		initCmd.Stdout = &initBuffer
		initCmd.Stderr = &initBuffer
		if err = initCmd.Run(); err != nil {
			return fmt.Errorf("%s\n", initBuffer.String())
		}

		generateMnemonicBuffer := new(bytes.Buffer)
		// Generate mnemonic
		generateMnemonicCmd := exec.Command(binary, "keys", "mnemonic")
		generateMnemonicCmd.Stdout = generateMnemonicBuffer
		generateMnemonicCmd.Stderr = generateMnemonicBuffer

		if err = generateMnemonicCmd.Run(); err != nil {
			return fmt.Errorf("%s\n", generateMnemonicBuffer.String())
		}

		mnemonic := generateMnemonicBuffer.String()
		mnemonicList = append(mnemonicList, mnemonic)

		var mnemonicBuffer bytes.Buffer
		_, err = mnemonicBuffer.Write([]byte(mnemonic))
		if err != nil {
			return err
		}

		//keysAddBuffer, keysAddCmdW, err := os.Pipe()
		var keysAddBuffer bytes.Buffer
		keysAddCmd := exec.Command(binary, "--home", homePath,
			"keys", "add", moniker, "--recover", "--keyring-backend", "test", "--output", "json")
		keysAddCmd.Stdin = &mnemonicBuffer
		keysAddCmd.Stdout = &keysAddBuffer
		keysAddCmd.Stderr = &keysAddBuffer

		if err = keysAddCmd.Run(); err != nil {
			return fmt.Errorf("%s\n", keysAddBuffer.String())
		}

		var jqAddressBuffer bytes.Buffer
		jqAddress := exec.Command("jq", "-r", ".address")
		jqAddress.Stdin = &keysAddBuffer
		jqAddress.Stdout = &jqAddressBuffer
		jqAddress.Stderr = &jqAddressBuffer

		if err = jqAddress.Run(); err != nil {
			return fmt.Errorf("%s\n", jqAddressBuffer.String())
		}

		address := jqAddressBuffer.String()

		var validatorKeyBuffer bytes.Buffer
		validatorKeyCmd := exec.Command("cantod", "tendermint", "show-validator", "--home", homePath)
		validatorKeyCmd.Stdout = &validatorKeyBuffer
		validatorKeyCmd.Stderr = &validatorKeyBuffer

		if err = validatorKeyCmd.Run(); err != nil {
			return fmt.Errorf("%s\n", validatorKeyBuffer.String())
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
			Moniker: moniker,
			Address: address,
			//BalAmount:    balAmount,
			//StakeAmount:  stakeAmount,
			ValidatorKey: validatorKey,
			Mnemonic:     mnemonic,
		})
	}

	marshal, err := yaml.Marshal(rawValidatorList)
	if err != nil {
		return err
	}

	err = os.WriteFile(exportFilePath, marshal, 0666)
	if err != nil {
		return err
	}

	return nil
}
