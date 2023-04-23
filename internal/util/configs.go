package util

import (
	"time"

	"github.com/spf13/viper"
)

// A Configs defines the expected config values.
type Configs struct {
	Env                  string        `mapstructure:"ENVIRONMENT"`
	DBDriver             string        `mapstructure:"DB_DRIVER"`
	DBSource             string        `mapstructure:"DB_SOURCE"`
	ServerAddress        string        `mapstructure:"SERVER_ADDRESS"`
	SymmetricKey         string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	SessionTokenDuration time.Duration `mapstructure:"SESSION_TOKEN_DURATION"`
}

// ParseConfigs parses the configuration files.
func ParseConfigs(path string) (configs Configs, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("secrets")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&configs)
	return
}
