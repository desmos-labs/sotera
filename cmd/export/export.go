package export

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/forbole/juno/v2/types/config"
	"github.com/rs/zerolog/log"

	"github.com/cosmos/cosmos-sdk/simapp"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	"github.com/spf13/cobra"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/desmos-labs/soteria/types"
)

const (
	flagLimitHeight = "max-height"
	flagOutput      = "output"
)

// NewCmdExport returns the Cobra command to be used in order to export the correct vesting state querying the chain
func NewCmdExport() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Exports the vesting accounts data reading the original vesting data from the provided genesis file",
		Long: `
Reads the genesis file located inside the home dir, and for each vesting account present performs the following operations: 

1. gets the original vesting data stored inside the genesis;
2. queries all the MsgDelegate and MsgUnbond transactions the account has ever made; 
3. call the TrackDelegation and TrackUndelegation methods accordingly

Finally, it exports the result accounts state as a JSON array. 
`,
		PreRunE: types.ReadConfig(),
		RunE: func(cmd *cobra.Command, args []string) error {
			encodingConfig := simapp.MakeTestEncodingConfig()

			// Build the exporter
			exporter, err := NewExporter(types.Cfg.Node, &encodingConfig)
			if err != nil {
				return err
			}

			// Get the height
			height, err := cmd.Flags().GetInt64(flagLimitHeight)
			if err != nil {
				return err
			}

			err = exporter.SetLimitHeight(height)
			if err != nil {
				return err
			}

			log.Debug().Int64("max height", height).Msg("exporting accounts")

			// Get the accounts
			authState, err := readAuthGenesis(path.Join(config.HomePath, "genesis.json"), encodingConfig.Marshaler)
			if err != nil {
				return err
			}

			accounts, err := getVestingAccounts(authState)
			if err != nil {
				return err
			}

			log.Info().Msgf("fixing %d vesting accounts", len(accounts))

			// Fix the accounts
			startTime := time.Now()
			for i, account := range accounts {
				accountAddress := account.GetAddress().String()

				// TODO: Debug - remove this
				if accountAddress != "desmos172ejkpn9rxnjsr7py8cpjf8dpvqrspwxdexguu" {
					continue
				}

				log.Debug().Int("index", i+1).Str("address", accountAddress).Msg("fixing account")
				err = exporter.FixVestingAccount(account)
				if err != nil {
					return err
				}
			}

			executionTime := time.Since(startTime)
			log.Info().Str("duration", executionTime.String()).Msgf("fixed %d vesting accounts", len(accounts))

			// Print the results
			genesisAccount := make([]authtypes.GenesisAccount, len(accounts))
			for i, account := range accounts {
				genesisAccount[i] = account.(authtypes.GenesisAccount)
			}

			exportedAuthState := authtypes.NewGenesisState(authState.Params, genesisAccount)
			stateBz, err := encodingConfig.Marshaler.MarshalJSON(exportedAuthState)
			if err != nil {
				return err
			}

			outputFile, err := cmd.Flags().GetString(flagOutput)
			if err != nil {
				return err
			}

			if outputFile != "" {
				err = ioutil.WriteFile(outputFile, stateBz, 666)
				if err != nil {
					return err
				}
				return nil
			}

			cmd.SetOut(os.Stdout)
			cmd.Println(string(stateBz))

			return nil
		},
	}

	cmd.Flags().Int64(flagLimitHeight, 0, "Sets the height limit to be used. If 0, the latest height available will be used")
	cmd.Flags().String(flagOutput, "", "Path to the file where the output should be printed. If not provided, the output will be printed to stdout")

	return cmd
}

// readAuthGenesis reads the x/auth genesis state from the genesis file located at the given path
func readAuthGenesis(genesisPath string, cdc codec.Codec) (*authtypes.GenesisState, error) {
	bz, err := ioutil.ReadFile(genesisPath)
	if err != nil {
		return nil, err
	}

	var genesisDoc tmtypes.GenesisDoc
	err = tmjson.Unmarshal(bz, &genesisDoc)
	if err != nil {
		return nil, err
	}

	var genesisState simapp.GenesisState
	err = json.Unmarshal(genesisDoc.AppState, &genesisState)
	if err != nil {
		return nil, err
	}

	var authState authtypes.GenesisState
	err = cdc.UnmarshalJSON(genesisState[authtypes.ModuleName], &authState)
	if err != nil {
		return nil, err
	}

	return &authState, nil
}

// getVestingAccounts returns the vesting accounts present inside the provided genesis state
func getVestingAccounts(genesisState *authtypes.GenesisState) ([]exported.VestingAccount, error) {
	genAccounts, err := authtypes.UnpackAccounts(genesisState.Accounts)
	if err != nil {
		return nil, err
	}

	var accounts []exported.VestingAccount
	for _, account := range genAccounts {
		if vestingAccount, ok := account.(exported.VestingAccount); ok {
			accounts = append(accounts, vestingAccount)
		}
	}

	return accounts, nil
}
