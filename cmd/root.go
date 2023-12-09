package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"godex/internal/config"
	"godex/internal/db"
	"godex/internal/downloader"
	"godex/internal/mangadex"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "godex",
	Short: "Godex is a command line tool for downloading manga",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		// Initialize Resty client
		httpClient := resty.New()
		// Load config
		configExists, err := config.ConfigExists()
		if err != nil {
			log.Fatalf("Error checking if godex config exists: %v", err)
		}

		if !configExists {
			fmt.Println("Cannot run godex, there's no configuration available. \nPlease run either godex prompt or godex load")
			os.Exit(0)
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("Cannot run godex, issue when loading configuration :%v", err)
		}

		db, err := db.New()
		if err != nil {
			log.Fatalf("cannot initialize db: %v", err)
		}

		if !db.IsHealthy() {
			log.Fatalf("cannot initialize db: %v", err)
		}

		// Get the last run time
		lastRanAt, err := config.LoadTimestamp()
		if err != nil {
			log.Fatalf("Error loading last run timestamp: %v", err)
		}

		// Create a new MangaDex client
		client := mangadex.NewClient(cfg, httpClient)

		// Login to MangaDex
		loginInfo, err := client.Login(ctx)
		if err != nil {
			log.Fatalf("Error logging in to MangaDex: %v", err)
		}
		// Set the client to use the login info
		client.SetAuthToken(loginInfo.AccessToken)
		// Get the list of followed manga
		mangaList, err := client.GetFollowedMangaFeed(ctx, lastRanAt)
		if err != nil {
			log.Fatalf("Error getting followed manga feed: %v", err)
		}

		// Create a new downloader
		downloader := downloader.NewDownloader(cfg, httpClient, db)

		// Download the manga
		err = downloader.DownloadManga(ctx, mangaList, client)
		if err != nil {
			log.Fatalf("Error downloading manga: %v", err)
		}

		// Save timestamp for future use
		err = config.SaveTimestamp()
		if err != nil {
			log.Fatalf("Error writing the timestamp of this run of Godex: %v", err)
		}
		log.Println("Downloaded manga successfully")
	},
}

func Execute() {
	rootCmd.AddCommand(loadCmd)
	rootCmd.AddCommand(promptCmd)
	rootCmd.AddCommand(completeCmd)
	rootCmd.AddCommand(serveCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
