package models

import "time"

type User struct {
	ID           int
	Email        string
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

type Session struct {
	ID        string
	UserID    int
	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt *time.Time
}

type Category struct {
	ID   int
	Name string
}

type Post struct {
	ID         int
	UserID     int
	Title      string
	Body       string
	CreatedAt  time.Time
	Categories []Category
}

type Comment struct {
	ID        int
	PostID    int
	UserID    int
	Body      string
	CreatedAt time.Time
}
