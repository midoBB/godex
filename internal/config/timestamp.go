package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"godex/internal/util"

	gap "github.com/muesli/go-app-paths"
	"github.com/spf13/viper"
)

const (
	timestampKey  = "last_ran_at"
	timestampFile = "timestamp"
)

func timestampFileExists() (bool, error) {
	scope := gap.NewScope(gap.User, "godex")
	configPath, err := scope.DataPath(timestampFile)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(configPath)
	return err == nil, nil
}

func SaveTimestamp() error {
	scope := gap.NewScope(gap.User, "godex")
	dataFile, err := scope.DataPath(timestampFile)
	if err != nil {
		return err
	}

	alreadyExists, err := timestampFileExists()
	if err != nil {
		return err
	}
	if !alreadyExists {
		err = util.CreateFileAndDir(dataFile)
		if err != nil {
			return err
		}
	}
	now := time.Now()

	// Initialize Viper
	v := viper.New()
	v.SetConfigType("json")

	// Set the timestamp value
	v.Set(timestampKey, now.Format(time.RFC3339))

	// Write the configuration to file
	err = v.WriteConfigAs(dataFile)
	if err != nil {
		return fmt.Errorf("error saving timestamp: %v", err)
	}

	return nil
}

func LoadTimestamp() (time.Time, error) {
	scope := gap.NewScope(gap.User, "godex")
	dataFile, err := scope.DataPath(timestampFile)
	if err != nil {
		return time.Now(), err
	}
	// Initialize Viper
	v := viper.New()
	v.SetConfigFile(dataFile)
	v.SetConfigType("json")

	// Read the configuration file
	err = v.ReadInConfig()
	if err != nil {
		return time.Now().AddDate(0, 0, -7), nil
	}

	// Get the timestamp value
	timestampStr := v.GetString(timestampKey)

	// Parse the timestamp in RFC3339 format
	t, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing timestamp: %v", err)
	}
	logTimestamp(t)
	return t, nil
}

func logTimestamp(lastRanAt time.Time) {
	formattedTimestamp := lastRanAt.Format("2006-01-02 15:04")
	log.Printf("Last ran at : %s", formattedTimestamp)
}
