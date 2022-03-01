package types

import (
	"fmt"
	"io/ioutil"
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/forbole/juno/v2/types"
	"github.com/forbole/juno/v2/types/config"
	"github.com/spf13/cobra"

	nodeconfig "github.com/forbole/juno/v2/node/config"
	"gopkg.in/yaml.v3"
)

var (
	Cfg *Config
)

type Config struct {
	Chain ChainConfig       `yaml:"chain"`
	Node  nodeconfig.Config `yaml:"node"`
}

type ChainConfig struct {
	Bech32Prefix string `yaml:"bech32_prefix"`
}

func ReadConfig() types.CobraCmdFunc {
	return func(cmd *cobra.Command, args []string) error {
		file := config.GetConfigFilePath()

		// Make sure the path exists
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fmt.Errorf("config file does not exist. Make sure you have run the init command")
		}

		// Read the config
		cfg, err := ParseConfig(file)
		if err != nil {
			return err
		}
		Cfg = cfg

		// Setup the SDK config
		sdkCfg := sdk.GetConfig()
		sdkCfg.SetBech32PrefixForAccount(cfg.Chain.Bech32Prefix, cfg.Chain.Bech32Prefix+"pub")

		return nil
	}
}

func ParseConfig(path string) (*Config, error) {
	bz, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(bz, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
