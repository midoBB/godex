package db

import "time"

type Manga struct {
	ID          uint `gorm:"primarykey"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	MangadexId  string `gorm:"uniqueIndex"`
	Title       string
	Description string
	MangaPath   string
	CoverArt    *CoverArt
	CoverArtId  *string
	Chapters    []Chapter
}

type Chapter struct {
	ID            uint `gorm:"primarykey"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	MangadexId    string `gorm:"uniqueIndex"`
	Title         string
	ChapterNumber string
	MangaId       string
}

type CoverArt struct {
	ID         uint `gorm:"primarykey"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	MangadexId string `gorm:"uniqueIndex"`
	MangaId    string
	Filename   string
}
