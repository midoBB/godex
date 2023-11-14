package config

import (
	"godex/pkg/mangadex"
	"log"

	"github.com/spf13/viper"
)

// LoadEnvVariables loads the environment variables from the .env file
func LoadEnvFile(envFile string) (*mangadex.EnvConfigs, error) {
	viper.SetConfigFile(envFile)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	env := &mangadex.EnvConfigs{}
	if err := viper.Unmarshal(env); err != nil {
		return nil, err
	}

	log.Printf("Loaded environment variables\n")
	log.Printf("Using %v as download folder \n", env.DownloadPath)

	return env, nil
}
