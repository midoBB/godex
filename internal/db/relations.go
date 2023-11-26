package db

import "time"

type manga struct {
	ID          uint `gorm:"primarykey"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	MangadexId  string `gorm:"uniqueIndex"`
	Title       string
	Description string
	MangaPath   string
	CoverArt    *coverArt
	CoverArtId  *string
	Chapters    []chapter
}

type chapter struct {
	ID         uint `gorm:"primarykey"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	MangadexId string `gorm:"uniqueIndex"`
	Title      string
	Chapter    string
	MangaId    string
}

type coverArt struct {
	ID         uint `gorm:"primarykey"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	MangadexId string `gorm:"uniqueIndex"`
	MangaId    string
	Filename   string
}
