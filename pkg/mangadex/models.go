package mangadex

// Most of these definitions are taken from https://github.com/darylhjd/mangodex

import (
	"encoding/json"
	"fmt"
)

type LoginResponse struct {
	AccessToken string `json:"access_token"`
}

type Manga struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	Attributes MangaAttributes `json:"attributes"`
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
