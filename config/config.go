package config

import (
	"fmt"
	"github.com/mitchellh/mapstructure" //nolint:goimports
	"github.com/spf13/viper"
	"os"
	"strings" //nolint:goimports
)

var Config *RootConfig

func Init(configPath string) error {
	path := strings.TrimSpace(configPath)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("config file not exists %q", err.Error())
	}
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("read config error %q", err)
	}
	Config = new(RootConfig)
	if err := viper.Unmarshal(Config, func(cfg *mapstructure.DecoderConfig) {
		cfg.TagName = "json"
	}); err != nil {
		return fmt.Errorf("load config error %q", err)
	}
	return nil
}
