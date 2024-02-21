package replay

import (
	"fmt"
	chain "github.com/Canto-Network/Canto/v7/app"
	inflationtypes "github.com/Canto-Network/Canto/v7/x/inflation/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/evmos/ethermint/encoding"
	"github.com/spf13/cobra"
	tmlog "github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"log"
	"time"
)

var (
	newChainId = "canto_7700-1"
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

func Genesis(dir, validatorFile, exportPath string) error {

	db, err := sdk.NewLevelDB("application", dir)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	encodingConfig := encoding.MakeConfig(chain.ModuleBasics)

	// Load previous height
	app := chain.NewCanto(tmlog.NewNopLogger(), db, nil, false, map[int64]bool{}, "localnet", 0, false, encodingConfig, EmptyAppOptions{})

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

	for _, v := range app.StakingKeeper.GetAllValidators(ctx) {
		if !v.IsJailed() {
			consAddr, _ := v.GetConsAddr()
			app.StakingKeeper.Jail(ctx, consAddr)
		}
	}

	// Get staking bond denom
	bondDenom := app.StakingKeeper.BondDenom(ctx)

	// Read Validators from file
	validatorList, err := ReadValidatorInfosFile(validatorFile, bondDenom)
	if err != nil {
		panic(fmt.Errorf("Failed to read validator file: %s\n%s", validatorFile, err.Error()))
	}

	for _, v := range validatorList {
		if err := app.BankKeeper.MintCoins(ctx, inflationtypes.ModuleName, v.VotingPower); err != nil {
			return err
		}
		if err := app.BankKeeper.SendCoinsFromModuleToAccount(ctx, inflationtypes.ModuleName, v.GetAddress(), v.VotingPower); err != nil {
			return err
		}

		if err := v.CreateValidator(ctx, &app.StakingKeeper, app.AppCodec()); err != nil {
			return err
		}

	}

	// checking account types
	accounts := app.AccountKeeper.GetAllAccounts(ctx)

	// Iterate over the accounts
	for _, acc := range accounts {
		// Check if the account is a Vesting Account
		if vestingAcc, ok := acc.(exported.VestingAccount); ok {
			address := vestingAcc.GetAddress()
			fmt.Printf("Vesting Account Address: %s\n", address.String())

		} else {
			// Handle other account types
		}
	}

	fmt.Printf("Length of accounts: %d\n", len(accounts))

	staking.EndBlocker(ctx, app.StakingKeeper)
	staking.BeginBlocker(ctx, app.StakingKeeper)
	//
	//newSlashingParams := types.NewParams(10000000, types.DefaultMinSignedPerWindow, types.DefaultDowntimeJailDuration, types.DefaultSlashFractionDoubleSign, types.DefaultSlashFractionDowntime)
	//app.SlashingKeeper.SetParams(ctx, newSlashingParams)

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
		InitialHeight: exported.Height,
		ConsensusParams: &tmproto.ConsensusParams{
			Block: tmproto.BlockParams{
				MaxBytes:   exported.ConsensusParams.Block.MaxBytes,
				MaxGas:     -1,
				TimeIotaMs: 1,
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
