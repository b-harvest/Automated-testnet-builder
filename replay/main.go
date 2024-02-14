package main

import (
	"github.com/b-harvest/replay/cmd"
	"os"
)

func main() {
	if err := cmd.RootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
