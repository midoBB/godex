package config

import (
	"log"

	"github.com/spf13/viper"
)

// EnvConfigs struct to map env values
type EnvConfigs struct {
	Username     string
	Password     string
	ClientId     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	DownloadPath string `mapstructure:"download_path"`
}

// LoadEnvVariables loads the environment variables from the .env file
func LoadEnvVariables() (*EnvConfigs, error) {
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	env := &EnvConfigs{}
	if err := viper.Unmarshal(env); err != nil {
		return nil, err
	}

	log.Printf("Loaded environment variables\n")
	log.Printf("Using %v as download folder \n", env.DownloadPath)

	return env, nil
}
