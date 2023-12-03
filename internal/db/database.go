package db

import (
	"context"
	"godex/internal/mangadex"
	"godex/internal/util"
	"time"

	"github.com/glebarez/sqlite"
	gap "github.com/muesli/go-app-paths"
	"gorm.io/gorm"
)

type Database interface {
	IsHealthy() bool
	AddManga(ctx context.Context, newManga mangadex.Manga, mangaDir string) error
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
	dbManga := manga{
		MangadexId:  newManga.ID,
		Title:       newManga.GetTitle(),
		Description: newManga.Attributes.Description.Values["en"],
		MangaPath:   mangaDir,
	}
	return s.db.FirstOrCreate(&dbManga, manga{MangadexId: newManga.ID}).Error
}
