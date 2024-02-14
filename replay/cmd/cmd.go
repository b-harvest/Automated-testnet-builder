package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	// AddressVerifier address verifier
	AddressVerifier = func(bz []byte) error {
		if n := len(bz); n != 20 && n != 32 {
			return fmt.Errorf("incorrect address length %d", n)
		}

		return nil
	}
)

const (
	// DisplayDenom defines the denomination displayed to users in client applications.
	DisplayDenom = "canto"
	// BaseDenom defines to the default denomination used in canto (staking, EVM, governance, etc.)
	BaseDenom = "acanto"
)

func NewReplayCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "genesis",

		Args: cobra.ExactArgs(2),
		PreRun: func(_ *cobra.Command, args []string) {
			//sdkConfig := sdk.GetConfig()
			//sdkConfig.SetPurpose(sdk.Purpose)
			//sdkConfig.SetCoinType(ethermint.Bip44CoinType)
			//sdkConfig.SetBech32PrefixForAccount(config.Bech32PrefixAccAddr, config.Bech32PrefixAccPub)
			//sdkConfig.SetBech32PrefixForValidator(config.Bech32PrefixValAddr, config.Bech32PrefixValPub)
			//sdkConfig.SetBech32PrefixForConsensusNode(config.Bech32PrefixConsAddr, config.Bech32PrefixConsPub)
			//sdkConfig.SetAddressVerifier(AddressVerifier)
			//sdkConfig.SetFullFundraiserPath(ethermint.BIP44HDPath)
			//if err := sdk.RegisterDenom(DisplayDenom, sdk.OneDec()); err != nil {
			//	panic(err)
			//}
			//
			//if err := sdk.RegisterDenom(BaseDenom, sdk.NewDecWithPrec(1, ethermint.BaseDenomUnit)); err != nil {
			//	panic(err)
			//}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				panic(fmt.Errorf("You have to use as \"replay genesis [dir] [validator-file]\". "))
			}

			dir := args[0]
			validatorFile := args[1]
			cmd.SilenceUsage = true
			return ReplayGenesis(dir, validatorFile)
		},
	}
	return cmd
}

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "testnet builder",
		Short: "testnet builder",
	}

	cmd.AddCommand(NewReplayCmd())
	cmd.AddCommand(ChainInitCmd())
	return cmd
}
