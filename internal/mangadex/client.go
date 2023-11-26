package mangadex

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

const (
	loginEndpoint    = "https://auth.mangadex.org/realms/mangadex/protocol/openid-connect/token"
	followedEndpoint = "https://api.mangadex.org/user/follows/manga/feed"
	getReadEndpoint  = "https://api.mangadex.org/manga/read/?ids[]=%v"
	setReadEndpoint  = "https://api.mangadex.org/manga/%v/read"
	chapterEndpoint  = "https://api.mangadex.org/chapter"
	mangaEndpoint    = "https://api.mangadex.org/manga/%v?translatedLanguage[]=en&includes[]=cover_art"
	coverEndpoint    = "https://uploads.mangadex.org/covers/%v/%v"
)

type Client struct {
	restyClient *resty.Client
	cfg         *Config
	authToken   string
}

func NewClient(cfg *Config, restyClient *resty.Client) *Client {
	return &Client{
		restyClient: restyClient,
		cfg:         cfg,
	}
}

func (c *Client) SetAuthToken(authToken string) {
	c.authToken = authToken
}

// Login Tries to authenticate a user using the provided credentials
func (c *Client) Login(ctx context.Context) (*LoginResponse, error) {
	loginResult := &LoginResponse{}
	_, err := c.restyClient.R().SetContext(ctx).SetResult(loginResult).
		SetFormData(map[string]string{
			"grant_type":    "password",
			"username":      c.cfg.Username,
			"password":      c.cfg.Password,
			"client_id":     c.cfg.ClientId,
			"client_secret": c.cfg.ClientSecret,
		}).
		Post(loginEndpoint)
	if err != nil {
		return nil, fmt.Errorf("login request failed: %v", err)
	}
	log.Printf("Logged in successfully as %v \n", c.cfg.Username)
	return loginResult, nil
}

// GetFollowedMangaFeed retrieves the feed of followed manga from the MangaDex API.
func (c *Client) GetFollowedMangaFeed(ctx context.Context, lastRanAt time.Time) ([]*GodexManga, error) {
	var chapters []*Chapter
	offset := 0
	limit := 100
	for {
		chapterList := &ChapterList{}
		_, err := c.restyClient.R().SetContext(ctx).SetAuthToken(c.authToken).
			SetQueryParams(map[string]string{
				"limit":                fmt.Sprintf("%d", limit),
				"offset":               fmt.Sprintf("%d", offset),
				"order[readableAt]":    "desc",
				"translatedLanguage[]": "en",
				"includes[]":           "manga",
				"createdAtSince":       getMangaDexTimeFormat(lastRanAt),
			}).
			SetResult(chapterList).
			Get(followedEndpoint)
		if err != nil {
			return nil, err
		}

		chapters = append(chapters, chapterList.Data...)

		if len(chapters) >= chapterList.Total {
			break
		}
		offset += limit
	}
	mangaMap := make(map[string]*GodexManga, 0)
	for _, chapter := range chapters {
		mangaID := chapter.GetManga().ID
		godexManga, ok := mangaMap[mangaID]
		if !ok {
			godexManga = &GodexManga{
				Manga: chapter.GetManga(),
				Chapters: []*GodexChapter{
					{
						Chapter: chapter,
					},
				},
			}
			mangaMap[mangaID] = godexManga
		} else {
			godexManga.Chapters = append(godexManga.Chapters,
				&GodexChapter{Chapter: chapter},
			)
		}
	}
	mangaList := make([]*GodexManga, 0, len(mangaMap))
	for _, godexManga := range mangaMap {
		mangaList = append(mangaList, godexManga)
	}
	err := c.setReadStatus(ctx, c.authToken, mangaList)
	mangaList = filterAlreadyRead(mangaList)
	if err != nil {
		return nil, err
	}
	log.Println("Got followed manga feed successfully")
	return mangaList, nil
}

func (c *Client) GetMangaCover(ctx context.Context, mangaID string) (string, error) {
	mangaResp := &MangaResponse{}
	_, err := c.restyClient.R().
		SetContext(ctx).SetResult(mangaResp).SetAuthToken(c.authToken).Get(fmt.Sprintf(mangaEndpoint, mangaID))
	if err != nil {
		return "", fmt.Errorf("can't get manga cover: %v", err)
	}
	return fmt.Sprintf(coverEndpoint, mangaID, mangaResp.Data.GetCover().Attributes.FileName), nil
}

// filterAlreadyRead Filters out any chapters that are marked as read to not redownload them.
func filterAlreadyRead(mangaList []*GodexManga) []*GodexManga {
	for _, godexManga := range mangaList {
		chapters := make([]*GodexChapter, 0, len(godexManga.Chapters))
		for _, godexChapter := range godexManga.Chapters {
			if !godexChapter.IsRead {
				chapters = append(chapters, godexChapter)
			}
		}
		godexManga.Chapters = chapters
	}
	return mangaList
}

// setReadStatus Sets the read status for the chapters collected in GetFollowedMangaFeed
func (c *Client) setReadStatus(ctx context.Context, authToken string, manga []*GodexManga) error {
	g, gCtx := errgroup.WithContext(ctx)
	for _, item := range manga {
		godexManga := item // create a new variable to avoid data race
		g.Go(func() error {
			readMarkers := &ChapterReadMarkers{}
			_, err := c.restyClient.R().SetContext(gCtx).SetResult(readMarkers).SetAuthToken(authToken).Get(fmt.Sprintf(getReadEndpoint, godexManga.Manga.ID))
			if err != nil {
				return fmt.Errorf("Error getting read markers for manga list: %w", err)
			}
			for _, readChapId := range readMarkers.Data {
				for _, chap := range godexManga.Chapters {
					if readChapId == chap.Chapter.ID {
						chap.IsRead = true
					}
				}
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}

// getMangaDexTimeFormat Formats the given timestamp in a format compatible with the mangadex API
func getMangaDexTimeFormat(timestamp time.Time) string {
	return timestamp.In(time.UTC).Format("2006-01-02T15:04:05")
}

func (c *Client) MarkMangaAsRead(ctx context.Context, mangaId string, chaptersToMarkAsRead []string) error {
	payload := ReadPayload{
		ChapterIdsRead:   chaptersToMarkAsRead,
		ChapterIdsUnread: []string{},
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshalling mark as read payload: %v", err)
	}

	_, err = c.restyClient.R().SetAuthToken(c.authToken).SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(jsonPayload).
		Post(fmt.Sprintf(setReadEndpoint, mangaId))
	return err
}

func (c *Client) GetMangaChapters(ctx context.Context, mangaUrl string) (*GodexManga, error) {
	id, err := extractMangaId(mangaUrl)
	if err != nil {
		return nil, err
	}
	var chapters []*Chapter
	offset := 0
	limit := 100

	for {
		chapterList := &ChapterList{}
		_, err := c.restyClient.R().SetContext(ctx).SetAuthToken(c.authToken).
			SetQueryParams(map[string]string{
				"limit":                fmt.Sprintf("%d", limit),
				"offset":               fmt.Sprintf("%d", offset),
				"manga":                id,
				"translatedLanguage[]": "en",
				"includes[]":           "manga",
			}).
			SetResult(chapterList).
			Get(chapterEndpoint)
		if err != nil {
			return nil, err
		}

		chapters = append(chapters, chapterList.Data...)

		if len(chapters) >= chapterList.Total {
			break
		}
		offset += limit
	}
	if len(chapters) == 0 {
		return nil, fmt.Errorf("no chapters found available for the requested manga")
	}
	godexChapters := make([]*GodexChapter, len(chapters))
	for i, chapter := range chapters {
		godexChapters[i] = &GodexChapter{
			Chapter: chapter,
			IsRead:  false,
		}
	}
	return &GodexManga{
		Manga:    chapters[0].GetManga(),
		Chapters: godexChapters,
	}, nil
}

func extractMangaId(mangaUrl string) (string, error) {
	u, err := url.Parse(mangaUrl)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	path := u.Path
	segments := strings.Split(path, "/")

	// Check if the UUID exists
	if len(segments) < 3 {
		return "", fmt.Errorf("UUID not found in URL")
	}

	uuidStr := segments[2]

	// Parse the UUID
	_, err = uuid.Parse(uuidStr)
	if err != nil {
		return "", fmt.Errorf("invalid UUID: %w", err)
	}

	return uuidStr, nil
}
