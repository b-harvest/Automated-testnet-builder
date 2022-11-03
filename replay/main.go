package main

import (
	"os"

	"github.com/cosmosquad-labs/replay/cmd"
)

func main() {
	if err := cmd.NewReplayCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
