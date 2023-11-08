package sources

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"godex/pkg/mangadex"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
)

const (
	API_URL    = "https://jumpg-webapi.tokyo-cdn.com/api"
	USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36"
)

type MangaPlusResponse struct {
	Success struct {
		MangaViewer struct {
			Pages []struct {
				MangaPage struct {
					ImageUrl      string  `json:"imageUrl"`
					Width         int     `json:"width"`
					Height        int     `json:"height"`
					EncryptionKey *string `json:"encryptionKey"`
				} `json:"mangaPage"`
			}
		} `json:"mangaViewer"`
	} `json:"success"`
}

type Page struct {
	Index    int
	Referer  string
	ImageUrl string
}

// Based on the implementation taken from https://github.com/tachiyomiorg/tachiyomi-extensions mostly for de-DRMing the images.
type MangaPlus struct{}

// IsValid checks if the provided URL is valid for mangaplus.
func (p *MangaPlus) IsValid(chapter *mangadex.Chapter) bool {
	return chapter.Attributes.ExternalURL != nil &&
		strings.Contains(*chapter.Attributes.ExternalURL, "mangaplus")
}

// DownloadChapterImages downloads all images of a chapter and saves them in a directory.
// It returns the path to the directory where the images are saved.
func (p *MangaPlus) DownloadChapterImages(ctx context.Context, httpClient *resty.Client, chapterDir string, chapter *mangadex.Chapter) error {
	externalUrl := chapter.Attributes.ExternalURL
	chapterId := getChapterId(*externalUrl)
	pages, err := getPageList(ctx, httpClient, chapterId)
	if err != nil {
		return err
	}
	return downloadImages(ctx, chapterDir, pages)
}

// getChapterId extracts the chapter ID from the chapter's external URL.
func getChapterId(chapter string) string {
	return path.Base(chapter)
}

// getPageList fetches the list of pages for a chapter.
func getPageList(ctx context.Context, httpClient *resty.Client, chapterId string) ([]Page, error) {
	headers := map[string]string{
		"Referer":    API_URL + "/viewer/" + chapterId,
		"User-Agent": USER_AGENT,
	}

	queryParams := map[string]string{
		"chapter_id":  chapterId,
		"split":       "yes",
		"img_quality": "super_high",
		"format":      "json",
	}
	resp, err := httpClient.R().SetContext(ctx).
		SetHeaders(headers).
		SetQueryParams(queryParams).
		Get(API_URL + "/manga_viewer")
	if err != nil {
		return nil, fmt.Errorf("cannot get chapter info from mangaplus: %v", err)
	}
	pages, err := pageListParse(resp)
	if err != nil {
		return nil, fmt.Errorf("cannot parse chapter info from mangaplus: %v", err)
	}
	return pages, nil
}

// downloadImages downloads all images of a chapter concurrently using goroutines.
func downloadImages(ctx context.Context, chapterDir string, pages []Page) error {
	client := &http.Client{}
	sem := make(chan struct{}, 10) // Limit the number of concurrent downloads to 10.
	for i, page := range pages {
		if page.ImageUrl == "" {
			continue
		}
		sem <- struct{}{} // Acquire a token.
		go func(i int, page Page) {
			defer func() { <-sem }() // Release the token when done.
			req := imageRequest(page)
			resp, err := client.Do(req)
			if err != nil {
				// Handle error.
				return
			}
			defer resp.Body.Close()

			resp = imageIntercept(resp)

			imgData, err := io.ReadAll(resp.Body)
			if err != nil {
				// Handle error.
				return
			}

			err = os.WriteFile(fmt.Sprintf("%s/%d.jpg", chapterDir, i), imgData, 0644)
			if err != nil {
				// Handle error.
				return
			}
		}(i, page)
	}
	// Wait for all goroutines to finish.
	for i := 0; i < cap(sem); i++ {
		sem <- struct{}{}
	}
	return nil
}

// pageListParse parses the response from the MangaPlus API.
func pageListParse(response *resty.Response) ([]Page, error) {
	var result MangaPlusResponse
	err := json.Unmarshal(response.Body(), &result)
	if err != nil {
		return nil, err
	}

	referer := response.Request.Header.Get("Referer")

	pages := make([]Page, 0)
	for i, page := range result.Success.MangaViewer.Pages {
		encryptionKey := ""
		if page.MangaPage.EncryptionKey != nil {
			encryptionKey = "#" + *page.MangaPage.EncryptionKey
		}
		pages = append(pages, Page{
			Index:    i,
			Referer:  referer,
			ImageUrl: page.MangaPage.ImageUrl + encryptionKey,
		})
	}

	return pages, nil
}

// imageRequest creates a new HTTP request to download an image.
func imageRequest(page Page) *http.Request {
	req, _ := http.NewRequest("GET", page.ImageUrl, nil)
	req.Header.Set("Referer", page.Referer)
	return req
}

// imageIntercept intercepts the HTTP response to decode the image if necessary.
func imageIntercept(resp *http.Response) *http.Response {
	encryptionKey := resp.Request.URL.Fragment

	if encryptionKey == "" {
		return resp
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	image, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	image = decodeXorCipher(image, encryptionKey)

	body := io.NopCloser(bytes.NewReader(image))
	resp.Body = body
	resp.Header.Set("Content-Type", contentType)

	return resp
}

// decodeXorCipher decodes an image using the XOR cipher.
func decodeXorCipher(image []byte, key string) []byte {
	keyStream := make([]int, 0)
	for i := 0; i < len(key); i += 2 {
		keyByte, _ := strconv.ParseInt(key[i:i+2], 16, 0)
		keyStream = append(keyStream, int(keyByte))
	}

	for i, b := range image {
		image[i] = byte(int(b) ^ keyStream[i%len(keyStream)])
	}

	return image
}
