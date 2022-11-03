package cmd

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/evmos/ethermint/encoding"
	"github.com/spf13/cobra"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	tmlog "github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"

	chain "github.com/Canto-Network/Canto/v2/app"
)

func NewReplayCmd() *cobra.Command {
	var (
		newChainId    = "canto-testnet-1"
		initialHeight = int64(1)
	)
	cmd := &cobra.Command{
		Use:  "replay [dir] [height]",
		Args: cobra.ExactArgs(2),
		PreRun: func(_ *cobra.Command, args []string) {
			//chaincmd.GetConfig()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := args[0]
			height, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("parse height: %w", err)
			}
			cmd.SilenceUsage = true

			db, err := sdk.NewLevelDB("application", dir)
			if err != nil {
				panic(err)
			}
			defer db.Close()

			//encCfg := chain.MakeEncodingConfig()

			// Load previous height
			app := chain.NewCanto(tmlog.NewNopLogger(), db, nil, false, map[int64]bool{}, "localnet", 0, encoding.MakeConfig(chain.ModuleBasics), simapp.EmptyAppOptions{})
			if err := app.LoadHeight(height - 1); err != nil {
				panic(fmt.Errorf("failed to load height: %w", err))
			}

			ctx := app.BaseApp.NewContext(true, tmproto.Header{})
			ctx = ctx.WithBlockHeight(height)

			// Set governance params
			votingParams := app.GovKeeper.GetVotingParams(ctx)
			votingParams.VotingPeriod = 30 * time.Second
			tallyParams := app.GovKeeper.GetTallyParams(ctx)
			tallyParams.Quorum = sdk.MustNewDecFromStr("0.000001")
			app.GovKeeper.SetVotingParams(ctx, votingParams)
			app.GovKeeper.SetTallyParams(ctx, tallyParams)

			// Get staking bond denom
			bondDenom := app.StakingKeeper.BondDenom(ctx)

			// Create validator1
			val1 := NewValidator(
				"canto1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
				sdk.NewCoins(sdk.NewInt64Coin(bondDenom, 1_000_000_000_000_000_000)),
				sdk.NewInt64Coin(bondDenom, 900_000_000_000_000_000),
				"{\"@type\": \"/cosmos.crypto.ed25519.PubKey\",\"key\":\"CzUC2BDiSxOBJ4tKxd9flLfZy6nrSKJ8YE7mfiHnhv8=\"}",
				"val1",
			)

			if err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, val1.VotingPower); err != nil {
				return err
			}
			if err := app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, val1.GetAddress(), val1.VotingPower); err != nil {
				return err
			}
			if err := val1.CreateValidator(ctx, &app.StakingKeeper, app.AppCodec()); err != nil {
				return err
			}

			val2 := NewValidator(
				"canto1mzgucqnfr2l8cj5apvdpllhzt4zeuh2c5l33n3",
				sdk.NewCoins(sdk.NewInt64Coin(bondDenom, 1_000_000_000_000_000_000)),
				sdk.NewInt64Coin(bondDenom, 50_000_000_000_000_000),
				"{\"@type\": \"/cosmos.crypto.ed25519.PubKey\",\"key\":\"GmAFwR4Z6iFTv6yzMETDigK38Nh38TDimLGvCaKkzvo=\"}",
				"val2",
			)
			if err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, val2.VotingPower); err != nil {
				return err
			}
			if err := app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, val2.GetAddress(), val2.VotingPower); err != nil {
				return err
			}
			if err := val2.CreateValidator(ctx, &app.StakingKeeper, app.AppCodec()); err != nil {
				return err
			}

			staking.EndBlocker(ctx, *&app.StakingKeeper)
			staking.BeginBlocker(ctx, *&app.StakingKeeper)

			log.Println("Exporting app state and validators...")

			exported, err := app.ExportAppStateAndValidators(false, nil)
			if err != nil {
				return fmt.Errorf("failed to export app state and validators: %w", err)
			}

			genDoc := &tmtypes.GenesisDoc{
				GenesisTime:   time.Now(),
				ChainID:       newChainId,
				AppState:      exported.AppState,
				Validators:    exported.Validators,
				InitialHeight: initialHeight,
				ConsensusParams: &tmproto.ConsensusParams{
					Block: tmproto.BlockParams{
						MaxBytes:   exported.ConsensusParams.Block.MaxBytes,
						MaxGas:     exported.ConsensusParams.Block.MaxGas,
						TimeIotaMs: 1000,
					},
					Evidence: tmproto.EvidenceParams{
						MaxAgeNumBlocks: exported.ConsensusParams.Evidence.MaxAgeNumBlocks,
						MaxAgeDuration:  exported.ConsensusParams.Evidence.MaxAgeDuration,
						MaxBytes:        exported.ConsensusParams.Evidence.MaxBytes,
					},
					Validator: tmproto.ValidatorParams{
						PubKeyTypes: exported.ConsensusParams.Validator.PubKeyTypes,
					},
				},
			}

			log.Println("Exporting genesis file...")

			return genutil.ExportGenesisFile(genDoc, "exported-genesis.json")
		},
	}
	return cmd
}
