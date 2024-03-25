package replay

import (
	"fmt"
	chain "github.com/Canto-Network/Canto/v7/app"
	inflationtypes "github.com/Canto-Network/Canto/v7/x/inflation/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/go-bip39"
	"github.com/evmos/ethermint/encoding"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	tmlog "github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	newChainId     = "canto_7700-1"
	keyringTestDir string
)

func init() {
	userHome, _ := os.UserHomeDir()
	keyringTestDir = filepath.Join(userHome, "keyring-test")
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomString(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)
	for i := 0; i < n; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}
	return sb.String()
}

func GenesisCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "genesis [mainnet-dir] [validator-file-path] [export-genesis-path] [account-count]",

		Args: cobra.ExactArgs(3),
		PreRun: func(cmd *cobra.Command, args []string) {
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				panic(fmt.Errorf("You have to use as \"replay genesis [dir] [validator-file]\". "))
			}

			dir := args[0]
			validatorFile := args[1]
			cmd.SilenceUsage = true
			accountCount, err := strconv.Atoi(args[2])
			if err != nil {
				return err
			}

			_, err = Genesis(dir, validatorFile, "exported-genesis.json", "account.yaml", accountCount)
			return err
		},
	}
	return cmd
}

func Genesis(dir, validatorFile, exportPath, extraAccountExportPath string, accountCnt int) (string, error) {

	rand.Seed(time.Now().UnixNano())

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
	votingParams.VotingPeriod = 300 * time.Second
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
			return "", err
		}
		if err := app.BankKeeper.SendCoinsFromModuleToAccount(ctx, inflationtypes.ModuleName, v.GetAddress(), v.VotingPower); err != nil {
			return "", err
		}

		if err := v.CreateValidator(ctx, &app.StakingKeeper, app.AppCodec()); err != nil {
			return "", err
		}
	}

	fmt.Printf("creating not validator accounts.... %d\n", accountCnt)

	var (
		f           *os.File
		accountList RawValidatorList
		b           []byte
	)
	f, err = os.Create(extraAccountExportPath)

	for i := 0; i < accountCnt; i++ {

		keyName := randomString(29)
		account, mnemonic, err := NewAccount(keyName)
		if err != nil {
			return "", err
		}

		if err := app.BankKeeper.MintCoins(ctx, inflationtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(bondDenom, sdk.NewIntWithDecimal(2, 30)))); err != nil {
			return "", err
		}
		if err := app.BankKeeper.SendCoinsFromModuleToAccount(ctx, inflationtypes.ModuleName, (*account).GetAddress(), sdk.NewCoins(sdk.NewCoin(bondDenom, sdk.NewIntWithDecimal(2, 30)))); err != nil {
			return "", err
		}

		accountList = append(accountList, RawValidator{
			Moniker:      keyName,
			Address:      (*account).GetAddress().String(),
			ValidatorKey: "",
			Mnemonic:     mnemonic,
		})
	}

	b, err = yaml.Marshal(accountList)
	if err != nil {
		return "", err
	}

	_, err = f.Write(b)
	if err != nil {
		return "", err
	}

	err = os.RemoveAll(keyringTestDir)
	if err != nil {
		return "", err
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
		return "", fmt.Errorf("failed to export app state and validators: %w", err)
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
				MaxGas:     1000000000000000000,
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

	return strconv.FormatInt(exported.Height, 10), genutil.ExportGenesisFile(genDoc, exportPath)
}

var (
	mnemonicEntropySize = 256
)

func NewAccount(name string) (*keyring.Info, string, error) {
	kb, err := keyring.New(sdk.KeyringServiceName(), "test", keyringTestDir, nil, []keyring.Option{}...)
	if err != nil {
		return nil, "", err
	}

	keyringAlgos, _ := kb.SupportedAlgorithms()
	algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyringAlgos)
	if err != nil {
		return nil, "", err
	}

	coinType := sdk.GetConfig().GetCoinType()
	account := uint32(0)
	index := uint32(0)

	hdPath := hd.CreateHDPath(coinType, account, index).String()

	// Get bip39 mnemonic
	var mnemonic, bip39Passphrase string

	// read entropy seed straight from tmcrypto.Rand and convert to mnemonic
	entropySeed, err := bip39.NewEntropy(mnemonicEntropySize)
	if err != nil {
		return nil, "", err
	}

	mnemonic, err = bip39.NewMnemonic(entropySeed)
	if err != nil {
		return nil, "", err
	}

	info, err := kb.NewAccount(name, mnemonic, bip39Passphrase, hdPath, algo)
	if err != nil {
		return nil, "", err
	}

	return &info, mnemonic, nil
}
