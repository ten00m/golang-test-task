package storage

import (
	"database/sql"
	"fmt"

	"github.com/ten00m/golang-test-task/internal/http-server/handlers"
)

func (db *DB) SetUserIsActive(userID string, isActive bool) (*handlers.User, error) {
	const op = "Storage.SetUserIsActive"

	var user handlers.User
	err := db.conn.QueryRow(`SELECT id, username, team_name, is_active FROM users WHERE id = $1`, userID).
		Scan(&user.ID, &user.Username, &user.TeamName, &user.IsActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: user not found", op)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = db.conn.Exec(`UPDATE users SET is_active = $1 WHERE id = $2`, isActive, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	user.IsActive = isActive
	return &user, nil
}

func (db *DB) GetUser(userID string) (*handlers.User, error) {
	const op = "Storage.GetUser"

	var user handlers.User
	err := db.conn.QueryRow(`SELECT id, username, team_name, is_active FROM users WHERE id = $1`, userID).
		Scan(&user.ID, &user.Username, &user.TeamName, &user.IsActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: user not found", op)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}
