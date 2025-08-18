package server

import (
	"database/sql"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"

	"strings"

	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"forum/internal/models"
)

type Server struct {
	DB         *sql.DB

	tmpl       map[string]*template.Template

	CookieName string
}

func New(db *sql.DB, templateDir string) (*Server, error) {

	templates := map[string]*template.Template{}
	layout := filepath.Join(templateDir, "layout.html")
	pages, err := filepath.Glob(filepath.Join(templateDir, "*.html"))
	if err != nil {
		return nil, err
	}
	for _, page := range pages {
		if filepath.Base(page) == "layout.html" {
			continue
		}
		t, err := template.ParseFiles(layout, page)
		if err != nil {
			return nil, err
		}
		name := strings.TrimSuffix(filepath.Base(page), ".html")
		templates[name] = t
	}
	return &Server{DB: db, tmpl: templates, CookieName: "session_id"}, nil

}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/register", s.handleRegister)
	mux.HandleFunc("/login", s.handleLogin)
	mux.HandleFunc("/logout", s.handleLogout)
	mux.HandleFunc("/post/new", s.requireAuth(s.handleNewPost))
	mux.HandleFunc("/post", s.handlePost)
	mux.HandleFunc("/post/comment", s.requireAuth(s.handleComment))
	mux.HandleFunc("/post/like", s.requireAuth(s.handlePostLike))
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	return mux
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.routes().ServeHTTP(w, r)
}


func (s *Server) render(w http.ResponseWriter, name string, data any) {
	t, ok := s.tmpl[name]
	if !ok {
		http.Error(w, "template not found", http.StatusInternalServerError)
		return
	}
	if err := t.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	posts, err := models.ListPosts(s.DB, nil)
	if err != nil {
		http.Error(w, "error", 500)
		return
	}

	data := map[string]any{
		"Posts": posts,
		"User":  s.currentUser(r),
	}
	s.render(w, "index", data)

}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.render(w, "register", map[string]any{"User": s.currentUser(r)})

	case http.MethodPost:
		email := r.FormValue("email")
		username := r.FormValue("username")
		password := r.FormValue("password")
		if email == "" || username == "" || password == "" {
			http.Error(w, "missing fields", 400)
			return
		}
		hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		err := models.CreateUser(s.DB, email, username, string(hash))
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:

		s.render(w, "login", map[string]any{"User": s.currentUser(r)})

	case http.MethodPost:
		email := r.FormValue("email")
		password := r.FormValue("password")
		user, err := models.GetUserByEmail(s.DB, email)
		if err != nil {
			http.Error(w, "invalid email or password", 400)
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
			http.Error(w, "invalid email or password", 400)
			return
		}
		sid := uuid.NewString()
		expires := time.Now().Add(24 * time.Hour)
		if err := models.CreateSession(s.DB, user.ID, sid, expires); err != nil {
			http.Error(w, "could not create session", 500)
			return
		}
		http.SetCookie(w, &http.Cookie{Name: s.CookieName, Value: sid, Path: "/", Expires: expires, HttpOnly: true})
		http.Redirect(w, r, "/", http.StatusSeeOther)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie(s.CookieName)
	if err == nil {
		s.DB.Exec(`UPDATE sessions SET revoked_at = CURRENT_TIMESTAMP WHERE id = ?`, cookie.Value)
		http.SetCookie(w, &http.Cookie{Name: s.CookieName, Path: "/", MaxAge: -1})
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) handleNewPost(w http.ResponseWriter, r *http.Request, user *models.User) {
	switch r.Method {
	case http.MethodGet:

		s.render(w, "new_post", map[string]any{"User": user})

	case http.MethodPost:
		title := r.FormValue("title")
		body := r.FormValue("body")
		if title == "" || body == "" {
			http.Error(w, "missing fields", 400)
			return
		}
		categoryIDs := []int{}
		for _, v := range r.Form["categories"] {
			// simplistic parse
			if id := atoi(v); id > 0 {
				categoryIDs = append(categoryIDs, id)
			}
		}
		_, err := models.CreatePost(s.DB, user.ID, title, body, categoryIDs)
		if err != nil {
			http.Error(w, "could not create post", 500)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handlePost(w http.ResponseWriter, r *http.Request) {
	id := atoi(r.URL.Query().Get("id"))
	if id == 0 {
		http.NotFound(w, r)
		return
	}
	post, err := models.GetPost(s.DB, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	comments, _ := models.ListComments(s.DB, id)

	data := map[string]any{

		"Post":     post,
		"Comments": comments,
		"User":     s.currentUser(r),
	}
	s.render(w, "post", data)

}

func (s *Server) handleComment(w http.ResponseWriter, r *http.Request, user *models.User) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	postID := atoi(r.FormValue("post_id"))
	body := r.FormValue("body")
	if body == "" {
		http.Error(w, "missing body", 400)
		return
	}
	if err := models.CreateComment(s.DB, postID, user.ID, body); err != nil {
		http.Error(w, "could not create comment", 500)
		return
	}
	http.Redirect(w, r, "/post?id="+itoa(postID), http.StatusSeeOther)
}

func (s *Server) handlePostLike(w http.ResponseWriter, r *http.Request, user *models.User) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	postID := atoi(r.FormValue("post_id"))
	value := atoi(r.FormValue("value"))
	if value != 1 && value != -1 {
		http.Error(w, "invalid value", 400)
		return
	}
	if err := models.TogglePostLike(s.DB, postID, user.ID, value); err != nil {
		http.Error(w, "could not toggle", 500)
		return
	}
	http.Redirect(w, r, "/post?id="+itoa(postID), http.StatusSeeOther)
}

// middleware
func (s *Server) requireAuth(next func(http.ResponseWriter, *http.Request, *models.User)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := s.currentUser(r)
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r, user)
	}
}

func (s *Server) currentUser(r *http.Request) *models.User {
	cookie, err := r.Cookie(s.CookieName)
	if err != nil {
		return nil
	}
	sess, err := models.GetSession(s.DB, cookie.Value)
	if err != nil || sess.RevokedAt != nil || sess.ExpiresAt.Before(time.Now()) {
		return nil
	}
	row := s.DB.QueryRow(`SELECT id, email, username, password_hash, created_at FROM users WHERE id = ?`, sess.UserID)
	var u models.User
	if err := row.Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.CreatedAt); err != nil {
		return nil
	}
	return &u
}

// helpers
func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func itoa(i int) string {
	return strconv.Itoa(i)
}
