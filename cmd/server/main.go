package main

import (
	"log"
	"net/http"
	"os"

	"forum/internal/db"
	"forum/internal/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "forum.db"
	}
	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	srv, err := server.New(database, "web/templates")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, srv); err != nil {
		log.Fatal(err)
	}
}
