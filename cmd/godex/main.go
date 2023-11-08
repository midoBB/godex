package main

import (
	"context"
	"godex/pkg/config"
	"godex/pkg/downloader"
	"godex/pkg/mangadex"
	"log"

	"github.com/go-resty/resty/v2"
)

func main() {
	ctx := context.Background()
	// Initialize Resty client
	httpClient := resty.New()
	// Load environment variables
	env, err := config.LoadEnvVariables()
	if err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

	// Get the last run time
	lastRanAt, err := config.LoadTimestamp()
	if err != nil {
		log.Fatalf("Error loading last run timestamp: %v", err)
	}

	// Create a new MangaDex client
	client := mangadex.NewClient(env, httpClient)

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
	downloader := downloader.NewDownloader(env, httpClient)

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
}
