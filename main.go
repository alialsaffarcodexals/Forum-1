package main

import (
	"log"
	"net/http"

	"forum/utils"

	_ "github.com/mattn/go-sqlite3" // SQLite3 driver
)

func main() {
	_, err := utils.DBInitialize("forum")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", utils.DefaultHandler)
	http.HandleFunc("/home", utils.HomeHandler)
	http.HandleFunc("/login", utils.LoginHandler)
	http.HandleFunc("/logout", utils.LogoutHandler)
	http.HandleFunc("/guest", utils.GuestHandler)
	http.HandleFunc("/register", utils.RegisterHandler)
	http.HandleFunc("/posts", utils.PostsHandler)
	http.HandleFunc("/like", utils.LikeHandler)
	http.HandleFunc("/dislike", utils.DislikeHandler)
	http.HandleFunc("/comment", utils.CommentHandler)

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
