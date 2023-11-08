package config

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
)

const (
	timestampKey  = "timestamp"
	timestampFile = ".timestamp"
)

func SaveTimestamp() error {
	// Get the current time
	now := time.Now()

	// Initialize Viper
	v := viper.New()
	v.SetConfigType("ini")

	// Set the timestamp value
	v.Set(timestampKey, now.Format(time.RFC3339))

	// Write the configuration to file
	err := v.WriteConfigAs(timestampFile)
	if err != nil {
		return fmt.Errorf("error saving timestamp: %v", err)
	}

	return nil
}

func LoadTimestamp() (time.Time, error) {
	// Initialize Viper
	v := viper.New()
	v.SetConfigType("ini")

	// Read the configuration file
	err := v.ReadInConfig()
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
