package server

import (
	"fmt"
	"net/http"
	"time"

	"godex/internal/db"
)

type Server struct {
	port int
	db   db.Database
}

func NewServer(db db.Database) *http.Server {
	NewServer := &Server{
		port: 9092,
		db:   db,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
