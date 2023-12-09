package db

import "time"

type Manga struct {
	MangadexId   string `gorm:"primaryKey"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Title        string
	Description  string
	MangaPath    string
	CoverArt     *CoverArt
	CoverArtId   *string
	Chapters     []Chapter
	ChapterCount int `gorm:"-:all"`
}

type Chapter struct {
	MangadexId    string `gorm:"primaryKey"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Title         string
	ChapterNumber string
	MangaId       string
	ReadAt        *time.Time
	IsRead        bool
}

type CoverArt struct {
	MangadexId string `gorm:"primaryKey"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	MangaId    string
	Filename   string
}

type Page[T any] struct {
	CurrentPage   int
	NumberOfPages int
	TotalNumber   int
	Data          []T
}
