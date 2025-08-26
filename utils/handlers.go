package utils

import (
	"database/sql"
	"errors"
	"html/template"
	"log"
	"net/http"
	"time"
)

var tpl *template.Template

// InitTemplate parses and executes a template
func InitTemplate(w http.ResponseWriter, file string, data interface{}) {
	var err error
	tpl, err = template.ParseFiles(file)
	if err != nil {
		http.Error(w, "Template parsing error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tpl.Execute(w, data); err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// DefaultHandler redirects "/" to "/login"
func DefaultHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func GuestHandler(w http.ResponseWriter, r *http.Request) {
	// ✅ If it's a GET request → create a guest session
	if r.Method == http.MethodGet {
		user, err := db.Guest()
		if err != nil {
			http.Error(w, "Failed to create guest: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// ✅ Check if a user session already exists
		if cookie, err := r.Cookie("user"); err == nil && cookie.Value != "" {
			db.DeleteUser(cookie.Value)
		}

		// ✅ Set cookie manually
		SetUserCookie(w, user.UUID)

		// ✅ Redirect to /home
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	// ❌ Method not allowed
	RenderError(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// LogoutHandler handles POST /logout
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	uuid, err := GetUserFromCookie(r)
	if err != nil || uuid == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Mark user as logged out
	_, err = db.Conn.Exec("UPDATE users SET loggedin = 0 WHERE uuid = ?", uuid)
	if err != nil {
		http.Error(w, "Failed to log out", http.StatusInternalServerError)
		return
	}

	// Clear cookie
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)

	// Redirect to login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")

		// Authenticate
		user, err := db.Login(w, r, username, email, password)
		if err != nil {
			http.Error(w, "Login failed: "+err.Error(), http.StatusBadRequest)
			RenderError(w, "login failed: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Store cookie
		SetUserCookie(w, user.UUID)

		// Redirect (doesn't show POST response to user)
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}
	if r.Method == http.MethodGet {
		// need to kick user out if uuid in cookie exits/////////////// <----------
		// Show login form
		InitTemplate(w, "templates/login.html", nil)
		return
	}

	RenderError(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Get UUID from cookie
	uuid, err := GetUserFromCookie(r)
	if err != nil || uuid == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Check if session is still valid
	if err := db.CheckSession(w, uuid); err != nil {
		log.Println(err)
		RenderError(w, "Session expired. Please log in again.", http.StatusUnauthorized)
		return
	}

	// Refresh session (update lastseen)
	if err := db.RefreshSession(uuid); err != nil {
		log.Printf("Failed to refresh session for uuid %s: %v", uuid, err)
		// You may want to log the user out or ignore silently depending on use-case
	}

	// Render home page
	InitTemplate(w, "templates/home.html", map[string]string{"UUID": uuid})
}

func (db *DataBase) Guest() (*User, error) {
	uuid, err := GenerateUserID()
	if err != nil {
		return nil, err
	}

	user := User{
		UUID:          uuid,
		NotRegistered: true,
		Username:      "guest_" + uuid[:8],
		Email:         "",
		Password:      "",
		Lastseen:      time.Now(),
	}

	if err := db.SafeWriter("users", user); err != nil {
		return nil, err
	}

	return &user, nil
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")

		// Validate form fields
		if username == "" || email == "" || password == "" || confirmPassword == "" {
			http.Error(w, "All fields are required", http.StatusBadRequest)
			RenderError(w, "All fields are required", http.StatusBadRequest)
			return
		}

		if password != confirmPassword {
			http.Error(w, "Passwords do not match", http.StatusBadRequest)
			RenderError(w, "Passwords do not match", http.StatusBadRequest)
			return
		}

		// Register user
		user, err := db.Register(w, username, email, password)
		if err != nil {
			http.Error(w, "Registration failed: "+err.Error(), http.StatusBadRequest)
			RenderError(w, "Registration failed: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Store cookie
		SetUserCookie(w, user.UUID)

		// Redirect to home
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Show registration form
	InitTemplate(w, "templates/register.html", nil)
}

func (db *DataBase) Register(w http.ResponseWriter, username, email, password string) (*User, error) {
	uuid, err := GenerateUserID()
	if err != nil {
		return nil, err
	}

	hash, err := HashPassword(password)
	if err != nil {
		log.Println("Failed to hash password:", err)
		RenderError(w, "Internal server error", http.StatusInternalServerError)
		return nil, err
	}
	password = hash

	// Check if user already exists
	var existing User
	err = db.Conn.QueryRow("SELECT uuid FROM users WHERE username = ? OR email = ?", username, email).Scan(&existing.UUID)
	if err != sql.ErrNoRows {
		if err != nil {
			log.Println("Database error:", err)
			RenderError(w, "Internal server error", http.StatusInternalServerError)
			return nil, err
		}
		return nil, errors.New("user with this username or email already exists")
	}

	// Create new user
	user := User{
		UUID:          uuid,
		NotRegistered: false,
		Username:      username,
		Email:         email,
		Password:      password,
		Lastseen:      time.Now(),
	}

	// Insert safely using SafeWriter
	if err := db.SafeWriter("users", user); err != nil {
		log.Println("Failed to insert user:", err)
		RenderError(w, "Internal server error", http.StatusInternalServerError)
		return nil, err
	}

	return &user, nil
}
