package cmd

import (
	"godex/internal/db"
	"godex/internal/server"
	"log"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve a web application that allows reading and browsing the downloaded Manga",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := db.New()
		if err != nil {
			log.Fatalf("cannot initialize db: %v", err)
		}
		server := server.NewServer(db)

		err = server.ListenAndServe()
		if err != nil {
			log.Fatalf("cannot start server: %v", err)
		}
	},
}
