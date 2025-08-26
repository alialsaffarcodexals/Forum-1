package utils

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// Login checks if a user exists and optionally registers them.
// Returns the fully populated User struct.
func (db *DataBase) Login(w http.ResponseWriter, r *http.Request, username, email, password string) (User, error) {
	var user User

	// 1. Check if cookie already corresponds to a logged-in user
	uuid, err := GetUserFromCookie(r) // nil == cookie exists
	if err == nil && uuid != "" {
		var alreadyLoggedIn bool
		err := db.Conn.QueryRow("SELECT loggedin FROM users WHERE uuid = ?", uuid).Scan(&alreadyLoggedIn)
		if err == nil && alreadyLoggedIn {
			return User{}, errors.New("user already logged in")
		}
	}

	// 2. Query the user by username or email
	row := db.Conn.QueryRow(
		"SELECT uuid, username, email, password, notregistered, loggedin FROM users WHERE username = ? OR email = ?",
		username, email,
	)

	// Scan the result into the User struct
	errScan := row.Scan(&user.UUID, &user.Username, &user.Email, &user.Password, &user.NotRegistered, &user.LoggedIn)
	if errScan != nil {
		if errScan == sql.ErrNoRows {
			return User{}, errors.New("user not found")
		}
		log.Println("Error scanning user:", errScan)
		return User{}, errScan
	}

	// 3. Prevent login if user is already logged in
	if user.LoggedIn {
		return User{}, errors.New("this user is already logged in from another session")
	}

	// 4. Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return User{}, errors.New("invalid password")
	}

	// 5. Refresh session & mark user as logged in
	if err := db.RefreshSession(user.UUID); err != nil {
		return User{}, err
	}

	_, err = db.Conn.Exec("UPDATE users SET loggedin = 1 WHERE uuid = ?", user.UUID)
	if err != nil {
		return User{}, err
	}

	// Update the struct too
	user.LoggedIn = true

	// Login successful
	return user, nil
}

// Logout logs the user out by clearing session and updating DB.
func (db *DataBase) Logout(w http.ResponseWriter, r *http.Request) {
	uuid, err := GetUserFromCookie(r)
	if err != nil || uuid == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Set loggedin = 0
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
		MaxAge:   -1, // delete immediately
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)

	// Redirect to login
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
