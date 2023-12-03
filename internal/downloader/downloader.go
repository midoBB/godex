package downloader

import (
	"context"
	"errors"
	"fmt"
	"godex/internal/db"
	"godex/internal/downloader/sources"
	"godex/internal/mangadex"
	"godex/internal/util"
	"log"
	"path/filepath"
	"strings"

	"github.com/go-resty/resty/v2"
)

type Downloader struct {
	httpClient *resty.Client
	cfg        *mangadex.Config
	db         db.Database
}

func NewDownloader(cfg *mangadex.Config, httpClient *resty.Client, db db.Database) *Downloader {
	return &Downloader{
		httpClient: httpClient,
		cfg:        cfg,
		db:         db,
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
			log.Printf("Downloading manga: %v", manga.Manga.GetTitle())
			mangaDir, err := util.CreateMangaDir(d.cfg.DownloadPath, manga)
			if err != nil {
				errs = append(errs, fmt.Sprintf("failed to create manga directory: %v", err))
				continue
			}
			err = d.db.AddManga(ctx, *manga.Manga, mangaDir)
			if err != nil {
				errs = append(errs, fmt.Sprintf("failed to create manga: %v", err))
				continue
			}
			if !util.MangaCoverExists(mangaDir) {
				err := d.downloadCover(ctx, mangadexClient, mangaDir, manga)
				if err != nil {
					errs = append(errs, fmt.Sprintf("failed to download manga cover: %v", err))
					continue
				}
			}
			chaptersToMarkAsRead := make([]string, 0)
			for _, chapter := range manga.Chapters {
				downloaded, err := d.downloadChapter(ctx, mangaDir, chapter)
				if err != nil {
					errs = append(errs, fmt.Sprintf("failed to download chapter: %v", err))
					continue
				} else {
					chapterNumber := chapter.Chapter.GetChapterNum()
					if downloaded {
						log.Printf("Downloaded chapter: %v", chapterNumber)
						chaptersToMarkAsRead = append(chaptersToMarkAsRead, chapter.Chapter.ID)
					} else {
						log.Printf("Skipped chapter: %v", chapterNumber)
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
	actualChapter := *chapter.Chapter
	if util.CheckChapterAlreadyExists(mangaDir, actualChapter.GetChapterNum()) {
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
	return false, fmt.Errorf("cannot download chapter %v : unknown source %s", actualChapter.GetChapterNum(), util.GetHostname(*actualChapter.Attributes.ExternalURL))
}

func (d *Downloader) downloadCover(
	ctx context.Context,
	mangadexClient *mangadex.Client,
	mangaDir string,
	manga *mangadex.GodexManga,
) error {
	coverUrl, err := mangadexClient.GetMangaCover(ctx, manga.Manga.ID)
	if err != nil {
		return err
	}
	imagePath := filepath.Join(mangaDir, "cover.jpg")
	_, err = d.httpClient.R().
		SetContext(ctx).
		SetOutput(imagePath).
		Get(coverUrl)
	return err
}
