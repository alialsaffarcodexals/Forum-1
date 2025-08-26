package utils

import (
	"fmt"
	"net/http"
	"time"
)

// Cookie name we'll use to track the logged-in user
const SessionCookieName = "user"

// SetUserCookie creates a secure cookie with the user UUID
func SetUserCookie(w http.ResponseWriter, uuid string) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    uuid,
		Path:     "/",  // available to all routes
		HttpOnly: true, // JS can't read it
		SameSite: http.SameSiteLaxMode,
		Secure:   false,                         // change to true in production with HTTPS
		Expires:  time.Now().Add(1 * time.Hour), // cookie valid for 1 hour
	})
}

// GetUserFromCookie retrieves the UUID stored in the cookie
func GetUserFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return cookie.Value, nil
}

// ClearUserCookie removes the user cookie (for logout)
func ClearUserCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0), // expired in the past
	})
}
