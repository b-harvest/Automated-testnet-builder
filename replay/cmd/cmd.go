package cmd

import (
	"fmt"
	"log"
	"strconv"
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
		newChainId    = "canto_9911-1"
		initialHeight = int64(1)
	)
	cmd := &cobra.Command{
		Use:  "replay [dir] [height]",
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

			// Load previous height
			app := chain.NewCanto(tmlog.NewNopLogger(), db, nil, false, map[int64]bool{}, "localnet", 0, false, encoding.MakeConfig(chain.ModuleBasics), simapp.EmptyAppOptions{})
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

			var validatorList []Validator
			validatorList = append(validatorList, NewValidator(
				"canto1yw49hwhhds6583tykf9chh64s652udcxezrmdz",
				sdk.NewCoins(val1_coin),
				sdk.NewCoin(bondDenom, stake_amount1),
				"{\"@type\": \"/cosmos.crypto.ed25519.PubKey\",\"key\":\"/iKUVNLseOb8iUiaDVJNZ6GSCu+9DFO1U+Yy/jJjgBU=\"}",
				"us-west-2-validator-0",
			))
			validatorList = append(validatorList, NewValidator(
				"canto1w80matxepagkfnvdg9uh38j4rzeqn9y3p379km",
				sdk.NewCoins(val1_coin),
				sdk.NewCoin(bondDenom, stake_amount1),
				"{\"@type\": \"/cosmos.crypto.ed25519.PubKey\",\"key\":\"sOwux+d5BfU1DN4J+lDbG3uuB+Lg0Ebruv8tREc/UHk=\"}",
				"us-west-2-validator-1",
			))
			validatorList = append(validatorList, NewValidator(
				"canto179079ruwl0emrr6tsm7f6vspsqjqmzu44gsxpn",
				sdk.NewCoins(val1_coin),
				sdk.NewCoin(bondDenom, stake_amount1),
				"{\"@type\": \"/cosmos.crypto.ed25519.PubKey\",\"key\":\"O8cwRYoiupEi0Yk2Htn8eFK7pQxsBeEk5vBJeq9xO9g=\"}",
				"us-west-2-validator-2",
			))

			//Address:
			//BalAmount: "10000000000000000000000"
			//Mnemonic: hero device liar stairs federal february symbol rib call situate issue
			//	bless blossom program brass violin team spy horror abandon supreme match option
			//	annual
			//Moniker: us-west-2-validator-0
			//StakeAmount: "7000000000000000000000"
			//ValidatorKey: '{"@type":"/cosmos.crypto.ed25519.PubKey","key":"/iKUVNLseOb8iUiaDVJNZ6GSCu+9DFO1U+Yy/jJjgBU="}'
			//	- Address: canto1w80matxepagkfnvdg9uh38j4rzeqn9y3p379km
			//BalAmount: "10000000000000000000000"
			//Mnemonic: punch enact ostrich simple motor bargain shield uphold utility domain
			//	two clog pass large describe cross report taste average burst brass custom sense
			//	guard
			//Moniker: us-west-2-validator-1
			//StakeAmount: "7000000000000000000000"
			//ValidatorKey: '{"@type":"/cosmos.crypto.ed25519.PubKey","key":"sOwux+d5BfU1DN4J+lDbG3uuB+Lg0Ebruv8tREc/UHk="}'
			//	- Address: canto179079ruwl0emrr6tsm7f6vspsqjqmzu44gsxpn
			//BalAmount: "10000000000000000000000"
			//Mnemonic: clown boat soap outer twenty remove online elephant bachelor aerobic disease
			//	carpet drum ensure joy clock auto planet web layer album upgrade trap list
			//Moniker: us-west-2-validator-2
			//StakeAmount: "7000000000000000000000"
			//ValidatorKey: '{"@type":"/cosmos.crypto.ed25519.PubKey","key":"O8cwRYoiupEi0Yk2Htn8eFK7pQxsBeEk5vBJeq9xO9g="}'

			for _, v := range validatorList {
				if err := app.InflationKeeper.MintCoins(ctx, v.VotingPower[0]); err != nil {
					return err
				}
				if err := app.BankKeeper.SendCoinsFromModuleToAccount(ctx, inflationtypes.ModuleName, v.GetAddress(), v.VotingPower); err != nil {
					return err
				}
				if err := v.CreateValidator(ctx, &app.StakingKeeper, app.AppCodec()); err != nil {
					return err
				}
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
