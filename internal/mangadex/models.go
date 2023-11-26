package mangadex

// Most of these definitions are taken from https://github.com/darylhjd/mangodex

import (
	"encoding/json"
	"fmt"
)

// EnvConfigs struct to map env values
type EnvConfigs struct {
	Username     string
	Password     string
	ClientId     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	DownloadPath string `mapstructure:"download_path"`
}

type Config struct {
	Username     string
	Password     string
	ClientId     string
	ClientSecret string
	DownloadPath string
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
}

type MangaResponse struct {
	Result   string `json:"result"`
	Response string `json:"response"`
	Data     Manga  `json:"data"`
}
type Manga struct {
	ID            string          `json:"id"`
	Type          string          `json:"type"`
	Attributes    MangaAttributes `json:"attributes"`
	Relationships []Relationship  `json:"relationships"`
}

type ChapterList struct {
	Result   string     `json:"result"`
	Response string     `json:"response"`
	Data     []*Chapter `json:"data"`
	Limit    int        `json:"limit"`
	Offset   int        `json:"offset"`
	Total    int        `json:"total"`
}

// LocalisedStrings : A struct wrapping around a map containing each localised string.
type LocalisedStrings struct {
	Values map[string]string
}

// MangaAttributes : Attributes for a Manga.
type MangaAttributes struct {
	Title                  LocalisedStrings `json:"title"`
	AltTitles              LocalisedStrings `json:"altTitles"`
	Description            LocalisedStrings `json:"description"`
	IsLocked               bool             `json:"isLocked"`
	Links                  LocalisedStrings `json:"links"`
	OriginalLanguage       string           `json:"originalLanguage"`
	LastVolume             *string          `json:"lastVolume"`
	LastChapter            *string          `json:"lastChapter"`
	PublicationDemographic *string          `json:"publicationDemographic"`
	Status                 *string          `json:"status"`
	Year                   *int             `json:"year"`
	ContentRating          *string          `json:"contentRating"`
	State                  string           `json:"state"`
	Version                int              `json:"version"`
	CreatedAt              string           `json:"createdAt"`
	UpdatedAt              string           `json:"updatedAt"`
}

// Chapter : Struct containing information on a manga.
type Chapter struct {
	ID            string            `json:"id"`
	Type          string            `json:"type"`
	Attributes    ChapterAttributes `json:"attributes"`
	Relationships []Relationship    `json:"relationships"`
}

type GodexChapter struct {
	Chapter *Chapter
	IsRead  bool
}
type GodexManga struct {
	Manga    *Manga
	Chapters []*GodexChapter
}

// ChapterAttributes : Attributes for a Chapter.
type ChapterAttributes struct {
	Title              string  `json:"title"`
	Volume             *string `json:"volume"`
	Chapter            *string `json:"chapter"`
	TranslatedLanguage string  `json:"translatedLanguage"`
	Uploader           string  `json:"uploader"`
	ExternalURL        *string `json:"externalUrl"`
	Version            int     `json:"version"`
	CreatedAt          string  `json:"createdAt"`
	UpdatedAt          string  `json:"updatedAt"`
	PublishAt          string  `json:"publishAt"`
}

type Relationship struct {
	ID         string      `json:"id"`
	Type       string      `json:"type"`
	Attributes interface{} `json:"attributes"`
}
type CoverArt struct {
	ID         string             `json:"id"`
	Type       string             `json:"type"`
	Attributes CoverArtAttributes `json:"attributes"`
}
type CoverArtAttributes struct {
	Description string `json:"description"`
	Volume      string `json:"volume"`
	FileName    string `json:"fileName"`
	Locale      string `json:"locale"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	Version     int64  `json:"version"`
}

func (a *Relationship) UnmarshalJSON(data []byte) error {
	// Check for the type of the relationship, then unmarshal accordingly.
	typ := struct {
		ID         string          `json:"id"`
		Type       string          `json:"type"`
		Attributes json.RawMessage `json:"attributes"`
	}{}
	if err := json.Unmarshal(data, &typ); err != nil {
		return err
	}

	var err error
	switch typ.Type {
	case "manga":
		a.Attributes = &MangaAttributes{}
	case "cover_art":
		a.Attributes = &CoverArtAttributes{}
	default:
		a.Attributes = &json.RawMessage{}
	}

	a.ID = typ.ID
	a.Type = typ.Type
	if typ.Attributes != nil {
		if err = json.Unmarshal(typ.Attributes, a.Attributes); err != nil {
			return fmt.Errorf("error unmarshalling relationship of type %s: %s, %s",
				typ.Type, err.Error(), string(data))
		}
	}
	return err
}

func (l *LocalisedStrings) UnmarshalJSON(data []byte) error {
	l.Values = map[string]string{}

	// First try if can unmarshal directly.
	if err := json.Unmarshal(data, &l.Values); err == nil {
		return nil
	}

	// If fail, try to unmarshal to array of maps.
	var locals []map[string]string
	if err := json.Unmarshal(data, &locals); err != nil {
		return fmt.Errorf("error unmarshalling localisedstring: %s", err.Error())
	}

	// If pass, then add each item in the array to flatten to one map.
	for _, entry := range locals {
		for key, value := range entry {
			l.Values[key] = value
		}
	}
	return nil
}

func (c *Chapter) GetManga() *Manga {
	var manga *Manga
	for _, rel := range c.Relationships {
		if rel.Type == "manga" {
			mangaAttr, _ := rel.Attributes.(*MangaAttributes)
			manga = &Manga{
				ID:         rel.ID,
				Type:       rel.Type,
				Attributes: *mangaAttr,
			}
		}
	}
	return manga
}

// GetTitle : Get title of the Manga.
func (m *Manga) GetTitle() string {
	if title := m.Attributes.Title.GetLocalString("en"); title != "" {
		return title
	}
	return m.Attributes.AltTitles.GetLocalString("en")
}

// GetLocalString : Get the localised string for a particular language code.
// If the required string is not found, it will return the first entry, or an empty string otherwise.
func (l *LocalisedStrings) GetLocalString(langCode string) string {
	// If we cannot find the required code, then return first value.
	if s, ok := l.Values[langCode]; !ok {
		v := ""
		for _, value := range l.Values {
			v = value
			break
		}
		return v
	} else {
		return s
	}
}

// GetChapterNum : Get the chapter's chapter number.
func (c *Chapter) GetChapterNum() string {
	if num := c.Attributes.Chapter; num != nil {
		return *num
	}
	return "-"
}

// GetDescription : Get description of the Manga.
func (m *Manga) GetDescription() string {
	return m.Attributes.Description.GetLocalString("en")
}

func (m *Manga) GetCover() *CoverArt {
	var coverArt *CoverArt
	for _, rel := range m.Relationships {
		if rel.Type == "cover_art" {
			coverAttr, _ := rel.Attributes.(*CoverArtAttributes)
			coverArt = &CoverArt{
				ID:         rel.ID,
				Type:       rel.Type,
				Attributes: *coverAttr,
			}
		}
	}
	return coverArt
}

// ChapterReadMarkers : A response for getting a list of read chapters.
type ChapterReadMarkers struct {
	Data []string `json:"data"`
}

// MDHomeServerResponse : A response for getting a server URL to get chapters.
type MDHomeServerResponse struct {
	Result  string       `json:"result"`
	BaseURL string       `json:"baseUrl"`
	Chapter ChaptersData `json:"chapter"`
}

// ChaptersData : Struct containing data for the chapter's pages.
type ChaptersData struct {
	Hash      string   `json:"hash"`
	Data      []string `json:"data"`
	DataSaver []string `json:"dataSaver"`
}

type ReadPayload struct {
	ChapterIdsRead   []string `json:"chapterIdsRead"`
	ChapterIdsUnread []string `json:"chapterIdsUnread"`
}
