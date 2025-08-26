package utils

import (
	"database/sql"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

const SessionTimeout = 1 * time.Hour

type DataBase struct {
	Conn  *sql.DB
	Write sync.Mutex
}

type User struct {
	NotRegistered bool
	ID            int
	Username      string
	Email         string
	Password      string
	UUID          string
	Lastseen      time.Time
	LoggedIn      bool
}

type Post struct {
	ID       int
	Title    string
	Content  string
	Author   User
	Comments []Comment
	Likes    []Interaction
	DisLikes []Interaction
}

type Comment struct {
	ID      int
	Content string
	Author  User
	Post    Post
}

type Reply struct {
	ID      int
	Content string
	Author  User
	Comment Comment
}

type Category struct {
	ID   int
	Name string
}

type Interaction struct {
	ID      int
	User    User
	Post    Post
	Like    bool
	DisLike bool
}

type Filter struct {
	ID         int
	User       User
	Post       Post
	ByCategory []Post
	ByCreated  []Post
	ByLiked    []Post
}

type SubForum struct {
	ID      int
	Name    string
	Posts   []Post
	Creator User
	Admins  []User
}

type HomeData struct {
	UserLoggedIn bool
	Username     string
	IsGuest      bool
}
