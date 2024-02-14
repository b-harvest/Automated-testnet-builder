package main

import (
	"github.com/b-harvest/Automated-testnet-builder/replay"
	"github.com/spf13/cobra"
	"os"
)

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "testnet builder",
		Short: "testnet builder",
	}

	cmd.AddCommand(replay.GenesisCmd())
	cmd.AddCommand(replay.ChainInitCmd())
	return cmd
}

func main() {
	if err := RootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
