package storage

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/ten00m/golang-test-task/internal/http-server/handlers"
)

func (s *DB) AddTeam(team handlers.Team) error {
	const op = "Storage.AddTeam"

	stmt, err := s.conn.Prepare(`INSERT INTO teams (name) VALUES ($1)`)
	if err != nil {
		return fmt.Errorf("%s: failed to prepare statement: %w", op, err)
	}

	defer stmt.Close()

	_, err = stmt.Exec(team.Name)
	if err != nil {
		return fmt.Errorf("%s: failed to execute statement: %w", op, err)
	}

	err = addFKsForTeamsAndUsers(s, team.Members, team.Name)
	if err != nil {
		return fmt.Errorf("%s: failed to add FKs for teams and users: %w", op, err)
	}

	return nil
}

func addFKsForTeamsAndUsers(db *DB, users []handlers.User, teamName string) error {
	const op = "Storage.addFKsForTeamsAndUsers"

	insertUserStmt, err := db.conn.Prepare(`INSERT INTO users (id, username, is_active, team_name) VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET username = EXCLUDED.username, is_active = EXCLUDED.is_active, team_name = EXCLUDED.team_name`)
	if err != nil {
		return fmt.Errorf("%s: failed to prepare insertUser statement: %w", op, err)
	}

	insertFKsStmt, err := db.conn.Prepare(`INSERT INTO team_fk_user (team_name, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`)
	if err != nil {
		return fmt.Errorf("%s: failed to prepare insertFKs statement: %w", op, err)
	}

	defer insertUserStmt.Close()
	defer insertFKsStmt.Close()

	for _, user := range users {
		id := user.ID
		if id == "" {
			id = uuid.New().String()
		}

		tName := user.TeamName
		if tName == "" {
			tName = teamName
		}

		_, err := insertUserStmt.Exec(id, user.Username, user.IsActive, tName)
		if err != nil {
			return fmt.Errorf("%s: failed to execute insertUser statement: %w", op, err)
		}

		_, err = insertFKsStmt.Exec(tName, id)
		if err != nil {
			return fmt.Errorf("%s: failed to execute insertFKs statement: %w", op, err)
		}
	}

	return nil
}

func (s *DB) GetTeam(teamName string) (handlers.Team, error) {
	const op = "Storage.GetTeam"

	var name string
	err := s.conn.QueryRow(`SELECT name FROM teams WHERE name = $1`, teamName).Scan(&name)
	if err != nil {
		return handlers.Team{}, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := s.conn.Query(`SELECT id, username, is_active, team_name FROM users WHERE team_name = $1`, teamName)
	if err != nil {
		return handlers.Team{}, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	members := make([]handlers.User, 0)
	for rows.Next() {
		var u handlers.User
		if err := rows.Scan(&u.ID, &u.Username, &u.IsActive, &u.TeamName); err != nil {
			return handlers.Team{}, fmt.Errorf("%s: %w", op, err)
		}
		members = append(members, u)
	}

	if err := rows.Err(); err != nil {
		return handlers.Team{}, fmt.Errorf("%s: %w", op, err)
	}

	return handlers.Team{Name: name, Members: members}, nil
}
