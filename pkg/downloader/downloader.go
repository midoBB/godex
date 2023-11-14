package downloader

import (
	"context"
	"errors"
	"fmt"
	"godex/pkg/downloader/sources"
	"godex/pkg/mangadex"
	"godex/pkg/util"
	"log"
	"strings"

	"github.com/go-resty/resty/v2"
)

type Downloader struct {
	httpClient *resty.Client
	cfg        *mangadex.Config
}

func NewDownloader(cfg *mangadex.Config, httpClient *resty.Client) *Downloader {
	return &Downloader{
		httpClient: httpClient,
		cfg:        cfg,
	}
}

var downloadSources = []sources.Source{
	&sources.MangaPlus{},
	&sources.Mangadex{},
}

// DownloadManga downloads a list of manga.
// It takes a context, an authentication token, and a list of manga
// It returns an error if any operation fails.
func (d *Downloader) DownloadManga(ctx context.Context, mangaList []*mangadex.GodexManga, mangadexClient *mangadex.Client) error {
	err := util.CreateDownloadDir(d.cfg.DownloadPath)
	if err != nil {
		return err
	}
	var errs []string

	for _, manga := range mangaList {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			log.Printf("Downloading manga: %v", manga.Manga.Attributes.Title.Values["en"])
			mangaDir, err := util.CreateMangaDir(d.cfg.DownloadPath, manga)
			chaptersToMarkAsRead := make([]string, 0)
			if err != nil {
				errs = append(errs, fmt.Sprintf("failed to create manga directory: %v", err))
				continue
			}
			for _, chapter := range manga.Chapters {
				downloaded, err := d.downloadChapter(ctx, mangaDir, chapter)
				if err != nil {
					errs = append(errs, fmt.Sprintf("failed to download chapter: %v", err))
					continue
				} else {
					chapterNumber := chapter.Chapter.Attributes.Chapter
					if downloaded {
						log.Printf("Downloaded chapter: %v", *chapterNumber)
						chaptersToMarkAsRead = append(chaptersToMarkAsRead, chapter.Chapter.ID)
					} else {
						log.Printf("Skipped chapter: %v", *chapterNumber)
					}
				}
			}
			readErr := mangadexClient.MarkMangaAsRead(ctx, manga.Manga.ID, chaptersToMarkAsRead)
			if readErr != nil {
				errs = append(errs, fmt.Sprintf("failed to mark manga as read: %v", readErr))
				continue
			}
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}

// downloadChapter Downloads a chapter from any of the available sources and compresses it into a cbz in the according folder
// it returns a bool indicating whether the chapter was successfully downloaded and an error indicating if any error happened during download.
func (d *Downloader) downloadChapter(ctx context.Context, mangaDir string, chapter *mangadex.GodexChapter) (bool, error) {
	actualChapter := chapter.Chapter
	if util.CheckChapterAlreadyExists(mangaDir, *actualChapter.Attributes.Chapter) {
		return false, nil
	}
	for _, source := range downloadSources {
		if source.IsValid(actualChapter) {
			chapterDir, err := util.CreateChapterDir(mangaDir, actualChapter)
			if err != nil {
				return false, err
			}
			err = source.DownloadChapterImages(ctx, d.httpClient, chapterDir, actualChapter)
			if err != nil {
				return false, err
			}
			return true, util.CreateCBZ(chapterDir)
		}
	}
	return false, fmt.Errorf("cannot download chapter %v : unknown source %s", *actualChapter.Attributes.Chapter, *actualChapter.Attributes.ExternalURL)
}
