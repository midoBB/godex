package sources

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"godex/internal/mangadex"

	"github.com/go-resty/resty/v2"
)

const (
	downloadEndpoint = "https://api.mangadex.org/at-home/server/%v"
)

type Mangadex struct{}

// IsValid is a function that checks if a given chapter is valid.
// It checks if the ExternalURL attribute of the chapter is nil.
// If the ExternalURL is nil, the function returns true, indicating that the chapter is valid.
// Otherwise, it returns false, indicating that the chapter is not valid.
func (m *Mangadex) IsValid(chapter mangadex.Chapter) bool {
	return chapter.Attributes.ExternalURL == nil
}

// DownloadChapterImages is a function that downloads images for a given chapter.
// It first fetches the chapter data from the server using the provided HTTP client.
// Then it iterates over each image in the chapter data and downloads it.
// If an error occurs during the download, it cancels the context, which stops any ongoing downloads.
// This function returns an error if it fails to fetch the chapter data or if an error occurs during the download.
func (m *Mangadex) DownloadChapterImages(ctx context.Context, httpClient *resty.Client, chapterDir string, chapter mangadex.Chapter) error {
	endpoint := fmt.Sprintf(downloadEndpoint, chapter.ID)
	chapterData := &mangadex.MDHomeServerResponse{}
	_, err := httpClient.R().SetContext(ctx).SetResult(chapterData).Get(endpoint)
	if err != nil {
		return fmt.Errorf("error getting chapter list: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	semaphore := make(chan struct{}, 10) // limit to 10 concurrent downloads

	for i, imageData := range chapterData.Chapter.Data {
		semaphore <- struct{}{} // acquire a token

		url := fmt.Sprintf("%v/data/%v/%v", chapterData.BaseURL, chapterData.Chapter.Hash, imageData)
		dataSaverUrl := fmt.Sprintf("%v/data-saver/%v/%v", chapterData.BaseURL, chapterData.Chapter.Hash, chapterData.Chapter.DataSaver[i])
		go func(i int, url, dataSaverUrl string) {
			defer func() { <-semaphore }() // release the token

			if err := m.downloadImage(ctx, httpClient, chapterDir, url, dataSaverUrl, i); err != nil {
				cancel() // cancel the context on error
			}
		}(i, url, dataSaverUrl)
	}

	// wait for all downloads to finish
	for i := 0; i < cap(semaphore); i++ {
		semaphore <- struct{}{}
	}
	return nil
}

// downloadImage is a helper function that downloads a single image.
// It first attempts to download the image at the provided URL.
// If the download fails, it removes the partially downloaded file and then attempts to download the dataSaver version of the image.
// This function returns an error if it fails to download either version of the image.
func (m *Mangadex) downloadImage(ctx context.Context, httpClient *resty.Client, chapterDir, url, dataSaverUrl string, page int) error {
	imagePath := filepath.Join(chapterDir, strconv.Itoa(page)+filepath.Ext(url))
	_, err := httpClient.R().
		SetContext(ctx).
		SetOutput(imagePath).
		Get(url)
	if err == nil {
		return nil
	}
	// if we can't download the full quality image, just download the dataSaver version
	err = os.Remove(imagePath)
	if err != nil {
		return err
	}
	_, err = httpClient.R().
		SetContext(ctx).
		SetOutput(imagePath).
		Get(url)
	return err
}
