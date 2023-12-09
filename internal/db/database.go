package db

import (
	"context"
	"time"

	"godex/internal/mangadex"
	"godex/internal/util"

	"github.com/glebarez/sqlite"
	gap "github.com/muesli/go-app-paths"
	"gorm.io/gorm"
)

type Database interface {
	IsHealthy() bool
	AddManga(ctx context.Context, newManga mangadex.Manga, mangaDir string) error
	GetCover(ctx context.Context, mangaID string) (*CoverArt, error)
	SaveCover(ctx context.Context, dbCover *CoverArt, mangaID string, newCover *mangadex.CoverArt) error
}

type database struct {
	db *gorm.DB
}

func New() (Database, error) {
	exists, err := dbExists()
	if err != nil {
		return nil, err
	}
	scope := gap.NewScope(gap.User, "godex")
	dbPath, err := scope.DataPath("godex.sqlite")
	if err != nil {
		return nil, err
	}
	if !exists {
		err = util.CreateFileAndDir(dbPath)
		if err != nil {
			return nil, err
		}
		err = createDB(dbPath)
		if err != nil {
			return nil, err
		}
	}
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &database{db: db}, nil
}

func (s *database) IsHealthy() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := s.db.WithContext(ctx).Exec("SELECT 1").Error
	return err == nil
}

func (s *database) AddManga(ctx context.Context, newManga mangadex.Manga, mangaDir string) error {
	dbManga := Manga{
		MangadexId:  newManga.ID,
		Title:       newManga.GetTitle(),
		Description: newManga.Attributes.Description.Values["en"],
		MangaPath:   mangaDir,
	}
	return s.db.FirstOrCreate(&dbManga, Manga{MangadexId: newManga.ID}).Error
}

func (s *database) GetCover(ctx context.Context, mangaID string) (*CoverArt, error) {
	var cover CoverArt
	err := s.db.Where("manga_id = ?", mangaID).First(&cover).Error
	if err != nil {
		return nil, err
	}
	return &cover, nil
}

func (s *database) SaveCover(ctx context.Context, dbCover *CoverArt, mangaID string, newCover *mangadex.CoverArt) error {
	if dbCover == nil {
		dbCover = &CoverArt{
			MangadexId: newCover.ID,
			MangaId:    mangaID,
			Filename:   newCover.Attributes.FileName,
		}
	}
	if err := s.db.Save(dbCover).Error; err != nil {
		return err
	}
	return s.db.Model(&Manga{}).Where("mangadex_id = ?", mangaID).Update("cover_art_id", dbCover.ID).Error
}
