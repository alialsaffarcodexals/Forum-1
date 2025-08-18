package server

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"forum/internal/db"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	srv, err := New(database, "../../web/templates")
	if err != nil {
		t.Fatalf("server: %v", err)
	}
	return srv
}

func TestRegisterLogin(t *testing.T) {
	srv := newTestServer(t)
	// register
	form := url.Values{"email": {"a@b.com"}, "username": {"alice"}, "password": {"secret"}}
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("register code %d", w.Code)
	}
	// login
	form = url.Values{"email": {"a@b.com"}, "password": {"secret"}}
	req = httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("login code %d", w.Code)
	}
	if cookie := w.Result().Cookies(); len(cookie) == 0 {
		t.Fatalf("no cookie set")
	}
}

func TestRequireAuth(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/post/new", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect")
	}
}

func TestPostCommentLike(t *testing.T) {
	srv := newTestServer(t)
	// register and login
	form := url.Values{"email": {"a@b.com"}, "username": {"alice"}, "password": {"secret"}}
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	form = url.Values{"email": {"a@b.com"}, "password": {"secret"}}
	req = httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	cookie := w.Result().Cookies()[0]

	// create post
	form = url.Values{"title": {"hello"}, "body": {"world"}}
	req = httptest.NewRequest(http.MethodPost, "/post/new", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookie)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("post create code %d", w.Code)
	}

	// fetch post ID (id 1)
	// comment
	form = url.Values{"post_id": {"1"}, "body": {"c"}}
	req = httptest.NewRequest(http.MethodPost, "/post/comment", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookie)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("comment code %d", w.Code)
	}

	// like post
	form = url.Values{"post_id": {"1"}, "value": {"1"}}
	req = httptest.NewRequest(http.MethodPost, "/post/like", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookie)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("like code %d", w.Code)
	}
}
