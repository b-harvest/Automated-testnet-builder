package main

import (
	"github.com/b-harvest/Automated-testnet-builder/replay"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestReadValidatorInfos(t *testing.T) {

	expectedValidator := replay.ValidatorList{
		replay.Validator{
			Moniker: "val1",
			Address: "canto1cr6tg4cjvux00pj6zjqkh6d0jzg7mksapardz2",
			//VotingPower:    "1000000000000000000000000000",
			//SelfDelegation: "500000000000000000000000000",
			PublicKeyStr: "~/.cantod/config/priv_validator_key.json",
		},
	}

	bondDenom := "acanto"

	validatorList, err := replay.ReadValidatorInfosFile("../../example-vali-info.yaml", bondDenom)
	if err != nil {
		return
	}

	require.Equal(t, expectedValidator, validatorList)

}
