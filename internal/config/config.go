package config

import (
	"godex/internal/mangadex"
	"godex/internal/util"
	"log"
	"os"

	gap "github.com/muesli/go-app-paths"
	"github.com/spf13/viper"
)

func ConfigExists() (bool, error) {
	scope := gap.NewScope(gap.User, "godex")
	configPath, err := scope.ConfigPath("config.json")
	if err != nil {
		return false, err
	}
	_, err = os.Stat(configPath)
	return err == nil, nil
}

func LoadConfig() (*mangadex.Config, error) {
	scope := gap.NewScope(gap.User, "godex")
	configFile, err := scope.ConfigPath("config.json")
	if err != nil {
		return nil, err
	}
	viper.SetConfigFile(configFile)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	env := &mangadex.Config{}
	if err := viper.Unmarshal(env); err != nil {
		return nil, err
	}

	log.Printf("Loaded config\n")
	log.Printf("Using %v as download folder \n", env.DownloadPath)

	return env, nil
}

func SaveConfig(env *mangadex.EnvConfigs) error {
	scope := gap.NewScope(gap.User, "godex")
	configFile, err := scope.ConfigPath("config.json")
	if err != nil {
		return err
	}
	alreadyExists, err := ConfigExists()
	if err != nil {
		return err
	}
	if !alreadyExists {
		err = util.CreateFileAndDir(configFile)
		if err != nil {
			return err
		}
	}

	viper.SetConfigFile(configFile)

	if err := viper.MergeConfigMap(map[string]interface{}{
		"Username":     env.Username,
		"Password":     env.Password,
		"ClientId":     env.ClientId,
		"ClientSecret": env.ClientSecret,
		"DownloadPath": env.DownloadPath,
	}); err != nil {
		return err
	}

	if err := viper.WriteConfig(); err != nil {
		return err
	}

	log.Printf("Saved config\n")

	return nil
}
