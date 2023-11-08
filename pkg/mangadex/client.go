package mangadex

import (
	"context"
	"encoding/json"
	"fmt"
	"godex/pkg/config"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
	"golang.org/x/sync/errgroup"
)

const (
	loginEndpoint    = "https://auth.mangadex.org/realms/mangadex/protocol/openid-connect/token"
	followedEndpoint = "https://api.mangadex.org/user/follows/manga/feed?translatedLanguage[]=en&includes[]=manga&includes[]=user&order[readableAt]=desc&createdAtSince=%v"
	getReadEndpoint  = "https://api.mangadex.org/manga/read/?ids[]=%v"
	setReadEndpoint  = "https://api.mangadex.org/manga/%v/read"
)

type Client struct {
	restyClient *resty.Client
	env         *config.EnvConfigs
	authToken   string
}

func NewClient(env *config.EnvConfigs, restyClient *resty.Client) *Client {
	return &Client{
		restyClient: restyClient,
		env:         env,
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
			"username":      c.env.Username,
			"password":      c.env.Password,
			"client_id":     c.env.ClientId,
			"client_secret": c.env.ClientSecret,
		}).
		Post(loginEndpoint)
	if err != nil {
		return nil, fmt.Errorf("login request failed: %v", err)
	}
	log.Printf("Logged in successfully as %v \n", c.env.Username)
	return loginResult, nil
}

// GetFollowedMangaFeed retrieves the feed of followed manga from the MangaDex API.
func (c *Client) GetFollowedMangaFeed(ctx context.Context, lastRanAt time.Time) ([]*GodexManga, error) {
	endpoint := fmt.Sprintf(followedEndpoint, getMangaDexTimeFormat(lastRanAt))
	chapterResponse := &ChapterList{}
	_, err := c.restyClient.R().SetContext(ctx).SetResult(chapterResponse).SetAuthToken(c.authToken).Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("Error getting chapter list: %w", err)
	}
	mangaMap := make(map[string]*GodexManga, 0)
	for _, chapter := range chapterResponse.Data {
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
	err = c.setReadStatus(ctx, c.authToken, mangaList)
	mangaList = filterAlreadyRead(mangaList)
	if err != nil {
		return nil, err
	}
	log.Println("Got followed manga feed successfully")
	return mangaList, nil
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
