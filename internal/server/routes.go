package server

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Use(middleware.CORS())
	e.Use(middleware.Gzip())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/api/manga", s.listHandler)
	e.GET("/api/manga/:mangaId", s.mangaHandler)
	e.POST("/api/chapter/:chapterId/read", s.readHandler)
	e.GET("/api/health", s.healthHandler)
	return e
}

func (s *Server) listHandler(c echo.Context) error {
	pageStr := c.QueryParam("page")
	if pageStr == "" {
		pageStr = "1"
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid page number")
	}
	mangaList, err := s.db.GetMangaList(c.Request().Context(), page)
	if err != nil {
		return c.String(http.StatusInternalServerError, "ERROR")
	}
	return c.JSON(http.StatusOK, mangaList)
}

func (s *Server) mangaHandler(c echo.Context) error {
	mangaID := c.Param("mangaId")
	manga, err := s.db.GetManga(c.Request().Context(), mangaID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "ERROR")
	}
	if manga == nil {
		return c.String(http.StatusNotFound, "Manga not found")
	}
	return c.JSON(http.StatusOK, manga)
}

func (s *Server) readHandler(c echo.Context) error {
	chapterId := c.Param("chapterId")
	err := s.db.MarkChapterAsRead(c.Request().Context(), chapterId)
	if err != nil {
		return c.String(http.StatusInternalServerError, "ERROR")
	}
	return c.String(http.StatusOK, "Chapter marked as read")
}

func (s *Server) healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, s.db.IsHealthy())
}
