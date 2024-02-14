package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/spf13/cobra"

	chain "github.com/Canto-Network/Canto/v7/app"
	"github.com/Canto-Network/Canto/v7/cmd/config"
	inflationtypes "github.com/Canto-Network/Canto/v7/x/inflation/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/evmos/ethermint/encoding"
	ethermint "github.com/evmos/ethermint/types"
	tmlog "github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
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
	var (
		newChainId    = "canto_7700-1"
		initialHeight = int64(1)
	)
	cmd := &cobra.Command{
		Use:  "replay [dir]",
		Args: cobra.ExactArgs(2),
		PreRun: func(_ *cobra.Command, args []string) {
			sdkConfig := sdk.GetConfig()
			sdkConfig.SetPurpose(sdk.Purpose)
			sdkConfig.SetCoinType(ethermint.Bip44CoinType)
			sdkConfig.SetBech32PrefixForAccount(config.Bech32PrefixAccAddr, config.Bech32PrefixAccPub)
			sdkConfig.SetBech32PrefixForValidator(config.Bech32PrefixValAddr, config.Bech32PrefixValPub)
			sdkConfig.SetBech32PrefixForConsensusNode(config.Bech32PrefixConsAddr, config.Bech32PrefixConsPub)
			sdkConfig.SetAddressVerifier(AddressVerifier)
			sdkConfig.SetFullFundraiserPath(ethermint.BIP44HDPath)
			if err := sdk.RegisterDenom(DisplayDenom, sdk.OneDec()); err != nil {
				panic(err)
			}

			if err := sdk.RegisterDenom(BaseDenom, sdk.NewDecWithPrec(1, ethermint.BaseDenomUnit)); err != nil {
				panic(err)
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := args[0]
			cmd.SilenceUsage = true

			db, err := sdk.NewLevelDB("application", dir)
			if err != nil {
				panic(err)
			}
			defer db.Close()

			// Load previous height
			app := chain.NewCanto(tmlog.NewNopLogger(), db, nil, false, map[int64]bool{}, "localnet", 0, false, encoding.MakeConfig(chain.ModuleBasics), simapp.EmptyAppOptions{})
			height := app.LastBlockHeight()
			fmt.Printf("LastBlockHeight: %d\n", height)

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

			bal_amoint1, _ := sdk.NewIntFromString("1000000000000000000000000000")
			stake_amount1, _ := sdk.NewIntFromString("500000000000000000000000000")
			val1_coin := sdk.NewCoin(bondDenom, bal_amoint1)
			// Create validator1
			val1 := NewValidator(
				"canto1cr6tg4cjvux00pj6zjqkh6d0jzg7mksapardz2",
				sdk.NewCoins(val1_coin),
				sdk.NewCoin(bondDenom, stake_amount1),
				"{\"@type\": \"/cosmos.crypto.ed25519.PubKey\",\"key\":\"CzUC2BDiSxOBJ4tKxd9flLfZy6nrSKJ8YE7mfiHnhv8=\"}",
				"val1",
			)

			if err := app.InflationKeeper.MintCoins(ctx, val1.VotingPower[0]); err != nil {
				return err
			}
			if err := app.BankKeeper.SendCoinsFromModuleToAccount(ctx, inflationtypes.ModuleName, val1.GetAddress(), val1.VotingPower); err != nil {
				return err
			}
			if err := val1.CreateValidator(ctx, &app.StakingKeeper, app.AppCodec()); err != nil {
				return err
			}
			bal_amoint2, _ := sdk.NewIntFromString("10000000000000000000")
			stake_amount2, _ := sdk.NewIntFromString("10000000000000000000")
			val2_coin := sdk.NewCoin(bondDenom, bal_amoint2)
			val2 := NewValidator(
				"canto1ywps7lrfjm8cww04pt9xad494u8qwhvdsjzzan",
				sdk.NewCoins(val2_coin),
				sdk.NewCoin(bondDenom, stake_amount2),
				"{\"@type\": \"/cosmos.crypto.ed25519.PubKey\",\"key\":\"GmAFwR4Z6iFTv6yzMETDigK38Nh38TDimLGvCaKkzvo=\"}",
				"val2",
			)
			if err := app.InflationKeeper.MintCoins(ctx, val2.VotingPower[0]); err != nil {
				return err
			}
			if err := app.BankKeeper.SendCoinsFromModuleToAccount(ctx, inflationtypes.ModuleName, val2.GetAddress(), val2.VotingPower); err != nil {
				return err
			}
			if err := val2.CreateValidator(ctx, &app.StakingKeeper, app.AppCodec()); err != nil {
				return err
			}

			staking.EndBlocker(ctx, app.StakingKeeper)
			staking.BeginBlocker(ctx, app.StakingKeeper)

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
