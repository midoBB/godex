package db

import (
	"context"
	"godex/internal/mangadex"
	"godex/internal/util"
	"math"
	"time"

	"github.com/glebarez/sqlite"
	gap "github.com/muesli/go-app-paths"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database interface {
	IsHealthy() bool
	AddManga(ctx context.Context, newManga mangadex.Manga, mangaDir string) error
	GetCover(ctx context.Context, mangaID string) (*CoverArt, error)
	SaveCover(ctx context.Context, dbCover *CoverArt, mangaID string, newCover *mangadex.CoverArt) error
	AddMangaChapter(ctx context.Context, newChapter mangadex.Chapter, mangaId string) error
	GetMangaList(ctx context.Context, page int) (*Page[Manga], error)
	GetManga(ctx context.Context, mangaID string) (*Manga, error)
	MarkChapterAsRead(ctx context.Context, chapterId string) error
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
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
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
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.db.Save(dbCover).Error; err != nil {
			return err
		}
		return s.db.Model(&Manga{}).Where("mangadex_id = ?", mangaID).Update("cover_art_id", dbCover.MangadexId).Error
	})
}

func (s *database) AddMangaChapter(ctx context.Context, newChapter mangadex.Chapter, mangaId string) error {
	dbChapter := Chapter{
		MangadexId:    newChapter.ID,
		Title:         newChapter.Attributes.Title,
		ChapterNumber: newChapter.GetChapterNum(),
		MangaId:       mangaId,
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.db.FirstOrCreate(&dbChapter, Chapter{MangadexId: newChapter.ID}).Error; err != nil {
			return err
		}
		return s.db.Model(&Manga{}).Where("mangadex_id = ?", mangaId).Update("updated_at", time.Now()).Error
	})
}

func (s *database) GetMangaList(ctx context.Context, page int) (*Page[Manga], error) {
	var mangas []Manga
	offset := (page - 1) * 10
	if err := s.db.Offset(offset).Preload("CoverArt").Limit(10).Order("updated_at desc").Find(&mangas).Error; err != nil {
		return nil, err
	}

	for i, manga := range mangas {
		var count int64
		s.db.Model(&Chapter{}).Where("manga_id = ?", manga.MangadexId).Count(&count)
		mangas[i].ChapterCount = int(count)
	}
	var total int64
	s.db.Model(&Manga{}).Count(&total)
	totalPages := int(math.Ceil(float64(total) / 10))
	result := Page[Manga]{
		CurrentPage:   page,
		NumberOfPages: totalPages,
		TotalNumber:   int(total),
		Data:          mangas,
	}
	return &result, nil
}

func (s *database) GetManga(ctx context.Context, mangaID string) (*Manga, error) {
	var manga Manga

	// Query the database
	if err := s.db.Preload("Chapters").Preload("CoverArt").First(&manga, "mangadex_id = ?", mangaID).Error; err != nil {
		return nil, err
	}
	return &manga, nil
}

func (s *database) MarkChapterAsRead(ctx context.Context, chapterId string) error {
	now := time.Now()
	chapter := Chapter{
		MangadexId: chapterId,
		IsRead:     true,
		ReadAt:     &now,
	}
	return s.db.Model(&Chapter{}).Where("mangadex_id = ?", chapterId).Updates(&chapter).Error
}
