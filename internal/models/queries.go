package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"
)

var (
	ErrDuplicateEmail     = errors.New("email already exists")
	ErrDuplicateUsername  = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

func CreateUser(db *sql.DB, email, username, passwordHash string) error {
	_, err := db.Exec(`INSERT INTO users (email, username, password_hash) VALUES (?, ?, ?)`, email, username, passwordHash)
	if err != nil {
		if sqliteErr, ok := err.(interface{ Error() string }); ok {
			if str := sqliteErr.Error(); str != "" {
				if strings.Contains(str, "UNIQUE constraint failed: users.email") {
					return ErrDuplicateEmail
				}
				if strings.Contains(str, "UNIQUE constraint failed: users.username") {
					return ErrDuplicateUsername
				}
			}
		}
	}
	return err
}

func GetUserByEmail(db *sql.DB, email string) (*User, error) {
	row := db.QueryRow(`SELECT id, email, username, password_hash, created_at FROM users WHERE email = ?`, email)
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func CreateSession(db *sql.DB, userID int, sessionID string, expires time.Time) error {
	// revoke existing
	_, err := db.Exec(`UPDATE sessions SET revoked_at = CURRENT_TIMESTAMP WHERE user_id = ? AND revoked_at IS NULL`, userID)
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)`, sessionID, userID, expires)
	return err
}

func GetSession(db *sql.DB, id string) (*Session, error) {
	row := db.QueryRow(`SELECT id, user_id, created_at, expires_at, revoked_at FROM sessions WHERE id = ?`, id)
	var s Session
	var revoked sql.NullTime
	err := row.Scan(&s.ID, &s.UserID, &s.CreatedAt, &s.ExpiresAt, &revoked)
	if err != nil {
		return nil, err
	}
	if revoked.Valid {
		s.RevokedAt = &revoked.Time
	}
	return &s, nil
}

func CreatePost(db *sql.DB, userID int, title, body string, categoryIDs []int) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	res, err := tx.Exec(`INSERT INTO posts (user_id, title, body) VALUES (?, ?, ?)`, userID, title, body)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	postID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	for _, cid := range categoryIDs {
		if _, err := tx.Exec(`INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)`, postID, cid); err != nil {
			tx.Rollback()
			return 0, err
		}
	}
	return postID, tx.Commit()
}

func ListPosts(db *sql.DB, categoryID *int) ([]Post, error) {
	q := `SELECT p.id, p.user_id, p.title, p.body, p.created_at FROM posts p`
	args := []any{}
	if categoryID != nil {
		q += ` JOIN post_categories pc ON pc.post_id = p.id WHERE pc.category_id = ?`
		args = append(args, *categoryID)
	}
	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []Post
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.UserID, &p.Title, &p.Body, &p.CreatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func GetPost(db *sql.DB, id int) (*Post, error) {
	row := db.QueryRow(`SELECT id, user_id, title, body, created_at FROM posts WHERE id = ?`, id)
	var p Post
	if err := row.Scan(&p.ID, &p.UserID, &p.Title, &p.Body, &p.CreatedAt); err != nil {
		return nil, err
	}
	// categories
	rows, err := db.Query(`SELECT c.id, c.name FROM categories c JOIN post_categories pc ON pc.category_id = c.id WHERE pc.post_id = ?`, id)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var c Category
			if err := rows.Scan(&c.ID, &c.Name); err == nil {
				p.Categories = append(p.Categories, c)
			}
		}
	}
	return &p, nil
}

func CreateComment(db *sql.DB, postID, userID int, body string) error {
	_, err := db.Exec(`INSERT INTO comments (post_id, user_id, body) VALUES (?, ?, ?)`, postID, userID, body)
	return err
}

func ListComments(db *sql.DB, postID int) ([]Comment, error) {
	rows, err := db.Query(`SELECT id, post_id, user_id, body, created_at FROM comments WHERE post_id = ?`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cs []Comment
	for rows.Next() {
		var c Comment
		if err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Body, &c.CreatedAt); err != nil {
			return nil, err
		}
		cs = append(cs, c)
	}
	return cs, rows.Err()
}

func TogglePostLike(db *sql.DB, postID, userID, value int) error {
	_, err := db.Exec(`INSERT INTO post_likes (post_id, user_id, value) VALUES (?, ?, ?)
        ON CONFLICT(post_id, user_id) DO UPDATE SET value = CASE
            WHEN post_likes.value = ? THEN 0 ELSE ? END`, postID, userID, value, value, value)
	return err
}

func ToggleCommentLike(db *sql.DB, commentID, userID, value int) error {
	_, err := db.Exec(`INSERT INTO comment_likes (comment_id, user_id, value) VALUES (?, ?, ?)
        ON CONFLICT(comment_id, user_id) DO UPDATE SET value = CASE
            WHEN comment_likes.value = ? THEN 0 ELSE ? END`, commentID, userID, value, value, value)
	return err
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
