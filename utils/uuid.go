package utils

import (
	"github.com/google/uuid"
)

// GenerateUserID creates a new UUID string for a user
func GenerateUserID() (string, error) {
	id, err := uuid.NewUUID() // time+MAC-based UUID v1
	if err != nil {
		return "", err
	}
	return id.String(), nil
}
