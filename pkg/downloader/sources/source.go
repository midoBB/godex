package sources

import (
	"context"
	"godex/pkg/mangadex"

	"github.com/go-resty/resty/v2"
)

type Source interface {
	// IsValid checks if the provided URL is valid for an external source.
	IsValid(chapter *mangadex.Chapter) bool
	// DownloadChapterImages downloads all images of a chapter and saves them in a directory.
	// It returns the path to the directory where the images are saved.
	DownloadChapterImages(
		ctx context.Context,
		httpClient *resty.Client,
		chapterDir string,
		chapter *mangadex.Chapter,
	) error
}
