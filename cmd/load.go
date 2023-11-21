package cmd

import (
	"godex/internal/config"
	"log"

	"github.com/spf13/cobra"
)

var (
	envFile string
	loadCmd = &cobra.Command{
		Use:   "load",
		Short: "Load environment variables from a file",
		Run: func(cmd *cobra.Command, args []string) {
			env, err := config.LoadEnvFile(envFile)
			if err != nil {
				log.Fatalf("Error loading environment file: %v", err)
			}
			err = config.SaveConfig(env)
			if err != nil {
				log.Fatalf("Error saving config from provided environment file: %v", err)
			}
			log.Println("Env file loaded successfully")
			log.Println("You can now run Godex to download your manga feed")
		},
	}
)

func init() {
	loadCmd.Flags().StringVarP(&envFile, "env", "e", "", "Path to the environment file")
}
