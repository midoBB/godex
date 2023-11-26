package cmd

import (
	"context"
	"fmt"
	"godex/internal/config"
	"godex/internal/db"
	"godex/internal/downloader"
	"godex/internal/mangadex"
	"log"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
)

var (
	mangaUrl    string
	completeCmd = &cobra.Command{
		Use:   "full",
		Short: "Downloads all available chapters of a manga based on a url passed",
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

			// Create a new MangaDex client
			client := mangadex.NewClient(cfg, httpClient)

			loginInfo, err := client.Login(ctx)
			if err != nil {
				log.Fatalf("Error logging in to MangaDex: %v", err)
			}
			// Set the client to use the login info
			client.SetAuthToken(loginInfo.AccessToken)

			manga, err := client.GetMangaChapters(ctx, mangaUrl)
			if err != nil {
				log.Fatalf("Error loading list of manga chapters: %v", err)
			}

			// Create a new downloader
			downloader := downloader.NewDownloader(cfg, httpClient, db)

			// Download the manga
			err = downloader.DownloadManga(ctx, []*mangadex.GodexManga{manga}, client)
			if err != nil {
				log.Fatalf("Error downloading manga: %v", err)
			}
			log.Println("Downloaded manga successfully")
		},
	}
)

func init() {
	completeCmd.Flags().StringVarP(&mangaUrl, "url", "u", "", "Url of the manga to download")
}
