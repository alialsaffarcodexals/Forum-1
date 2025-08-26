package utils

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
)

func (db *DataBase) UserExists(masterkey string) (bool, error) {
	query := `SELECT 1 FROM users WHERE uuid = ? LIMIT 1`
	row := db.Conn.QueryRow(query, masterkey)

	var exists int
	err := row.Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil // user does not exist
	} else if err != nil {
		return false, err // some other error
	}

	return true, nil // user exists
}

func (db *DataBase) ForceLogout(w http.ResponseWriter, uuid string) error {
	exists, err := db.UserExists(uuid)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("user doesn't exist")
	}

	ClearUserCookie(w)
	return nil
}

func (db *DataBase) DeleteUser(uuid string) error {
	query := `
		DELETE FROM users
		WHERE uuid = ? AND notregistered = true
	`

	res, err := db.Conn.Exec(query, uuid)
	if err != nil {
		return fmt.Errorf("failed to delete guest user: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not check deletion result: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("no guest user found with the provided UUID")
	}

	return nil
}
