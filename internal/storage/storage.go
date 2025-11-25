package storage

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/ten00m/golang-test-task/internal/config"

	_ "github.com/lib/pq"
)

type DB struct {
	conn *sql.DB
	log  *slog.Logger
}

func New(cfg *config.PostgreSQLConfig, log *slog.Logger) (*DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Database,
	)

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("successfully connected to database")

	db := &DB{
		conn: conn,
		log:  log,
	}

	if err := db.initializeTables(); err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) initializeTables() error {
	const op = "Storage.initializeTables"

	if err := db.createTeamsTable(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := db.createUsersTable(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := db.createPullRequestsTable(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := db.createTeamFkUserTable(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := db.createPrFkReviewerTable(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (db *DB) createUsersTable() error {
	const op = "Storage.createUsersTable"

	query := `
		CREATE TABLE IF NOT EXISTS users(
			id TEXT PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			is_active BOOLEAN NOT NULL,
			team_name TEXT,
			FOREIGN KEY (team_name) REFERENCES teams(name)
		);
	`

	stmt, err := db.conn.Prepare(query)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (db *DB) createTeamsTable() error {
	const op = "Storage.createTeamsTable"

	query := `
		CREATE TABLE IF NOT EXISTS teams(
			name TEXT PRIMARY KEY
		);
	`

	stmt, err := db.conn.Prepare(query)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (db *DB) createPullRequestsTable() error {
	const op = "Storage.createPullRequestsTable"

	query := `
		CREATE TABLE IF NOT EXISTS pull_requests(
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			authorId TEXT NOT NULL,
			status VARCHAR(10) NOT NULL CHECK (status IN ('OPEN', 'MERGED')),
			FOREIGN KEY (authorId) REFERENCES users(id)
		);
	`

	stmt, err := db.conn.Prepare(query)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (db *DB) createTeamFkUserTable() error {
	const op = "Storage.createTeamFkUserTable"

	query := `
		CREATE TABLE IF NOT EXISTS team_fk_user(
			id SERIAL PRIMARY KEY,
			team_name TEXT NOT NULL,
			user_id TEXT NOT NULL,
			FOREIGN KEY (team_name) REFERENCES teams(name),
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
	`

	stmt, err := db.conn.Prepare(query)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// createPrFkReviewerTable creates the pr_fk_reviewer table
func (db *DB) createPrFkReviewerTable() error {
	const op = "Storage.createPrFkReviewerTable"

	query := `
		CREATE TABLE IF NOT EXISTS pr_fk_reviewer(
			id SERIAL PRIMARY KEY,
			pr_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			FOREIGN KEY (pr_id) REFERENCES pull_requests(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
	`

	stmt, err := db.conn.Prepare(query)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (db *DB) Close() error {
	if err := db.conn.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}
	db.log.Info("database connection closed")
	return nil
}

func (db *DB) GetConnection() *sql.DB {
	return db.conn
}
