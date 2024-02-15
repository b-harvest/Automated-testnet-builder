package replay

import (
	"fmt"
	chain "github.com/Canto-Network/Canto/v7/app"
	"github.com/Canto-Network/Canto/v7/cmd/config"
	inflationtypes "github.com/Canto-Network/Canto/v7/x/inflation/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/evmos/ethermint/encoding"
	ethermint "github.com/evmos/ethermint/types"
	"github.com/spf13/cobra"
	tmlog "github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"log"
	"time"
)

var (
	newChainId    = "canto_7700-1"
	initialHeight = int64(1)
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

func GenesisCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "genesis",

		Args: cobra.ExactArgs(2),
		PreRun: func(cmd *cobra.Command, args []string) {
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				panic(fmt.Errorf("You have to use as \"replay genesis [dir] [validator-file]\". "))
			}

			dir := args[0]
			validatorFile := args[1]
			cmd.SilenceUsage = true
			return Genesis(dir, validatorFile, "exported-genesis.json")
		},
	}
	return cmd
}

func preReplayGenesis() {
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
}

func Genesis(dir, validatorFile, exportPath string) error {
	preReplayGenesis()

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

	// Read Validators from file
	validatorList, err := ReadValidatorInfosFile(validatorFile, bondDenom)
	if err != nil {
		panic(fmt.Errorf("Failed to read validator file: %s\n%s", validatorFile, err.Error()))
	}

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

	//bal_amoint, _ := sdk.NewIntFromString("1000000000.000000000000000000")
	//stake_amount, _ := sdk.NewIntFromString("500000000.000000000000000000")
	//val_coin := sdk.NewCoin(bondDenom, bal_amoint)
	//
	//val := NewValidator(
	//	"canto1cr6tg4cjvux00pj6zjqkh6d0jzg7mksapardz2",
	//	sdk.NewCoins(val_coin),
	//	sdk.NewCoin(bondDenom, stake_amount),
	//	"{\"@type\": \"/cosmos.crypto.ed25519.PubKey\",\"key\":\"CzUC2BDiSxOBJ4tKxd9flLfZy6nrSKJ8YE7mfiHnhv8=\"}",
	//	"val1",
	//)

	//bal_amoint2, _ := sdk.NewIntFromString("10.000000000000000000")
	//stake_amount2, _ := sdk.NewIntFromString("10.000000000000000000")
	//val2_coin := sdk.NewCoin(bondDenom, bal_amoint2)
	//val2 := NewValidator(
	//	"canto1ywps7lrfjm8cww04pt9xad494u8qwhvdsjzzan",
	//	sdk.NewCoins(val2_coin),
	//	sdk.NewCoin(bondDenom, stake_amount2),
	//	"{\"@type\": \"/cosmos.crypto.ed25519.PubKey\",\"key\":\"GmAFwR4Z6iFTv6yzMETDigK38Nh38TDimLGvCaKkzvo=\"}",
	//	"val2",
	//)

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

	return genutil.ExportGenesisFile(genDoc, exportPath)
}
