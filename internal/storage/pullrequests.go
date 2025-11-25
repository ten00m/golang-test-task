package storage

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/ten00m/golang-test-task/internal/http-server/handlers"
)

func (db *DB) CreatePullRequest(prID, prName, authorID string) (*handlers.PullRequest, error) {
	const op = "Storage.CreatePullRequest"

	var existingID string
	err := db.conn.QueryRow(`SELECT id FROM pull_requests WHERE id = $1`, prID).Scan(&existingID)
	if err == nil {
		return nil, fmt.Errorf("%s: PR already exists", op)
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var teamName string
	var isActive bool
	err = db.conn.QueryRow(`SELECT team_name, is_active FROM users WHERE id = $1`, authorID).Scan(&teamName, &isActive)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("%s: author not found", op)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = db.conn.Exec(`INSERT INTO pull_requests (id, title, authorId, status) VALUES ($1, $2, $3, 'OPEN')`,
		prID, prName, authorID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := db.conn.Query(`
		SELECT id FROM users 
		WHERE team_name = $1 AND is_active = true AND id != $2
	`, teamName, authorID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var candidates []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		candidates = append(candidates, userID)
	}

	reviewers := selectRandomReviewers(candidates, 2)

	for _, reviewerID := range reviewers {
		_, err := db.conn.Exec(`INSERT INTO pr_fk_reviewer (pr_id, user_id) VALUES ($1, $2)`, prID, reviewerID)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to add reviewer: %w", op, err)
		}
	}

	return &handlers.PullRequest{
		ID:                prID,
		Name:              prName,
		AuthorID:          authorID,
		Status:            "OPEN",
		AssignedReviewers: reviewers,
	}, nil
}

func (db *DB) GetPullRequest(prID string) (*handlers.PullRequest, error) {
	const op = "Storage.GetPullRequest"

	var pr handlers.PullRequest
	err := db.conn.QueryRow(`SELECT id, title, authorId, status FROM pull_requests WHERE id = $1`, prID).
		Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: PR not found", op)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := db.conn.Query(`SELECT user_id FROM pr_fk_reviewer WHERE pr_id = $1`, prID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		reviewers = append(reviewers, userID)
	}
	pr.AssignedReviewers = reviewers

	return &pr, nil
}

func (db *DB) MergePullRequest(prID string) (*handlers.PullRequest, error) {
	const op = "Storage.MergePullRequest"

	pr, err := db.GetPullRequest(prID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = db.conn.Exec(`UPDATE pull_requests SET status = 'MERGED' WHERE id = $1`, prID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	pr.Status = "MERGED"
	return pr, nil
}

func (db *DB) ReassignReviewer(prID, oldReviewerID string) (string, error) {
	const op = "Storage.ReassignReviewer"

	pr, err := db.GetPullRequest(prID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if pr.Status == "MERGED" {
		return "", fmt.Errorf("%s: cannot reassign on merged PR", op)
	}

	isAssigned := false
	for _, reviewerID := range pr.AssignedReviewers {
		if reviewerID == oldReviewerID {
			isAssigned = true
			break
		}
	}
	if !isAssigned {
		return "", fmt.Errorf("%s: reviewer is not assigned to this PR", op)
	}

	var oldReviewerTeam string
	err = db.conn.QueryRow(`SELECT team_name FROM users WHERE id = $1`, oldReviewerID).Scan(&oldReviewerTeam)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("%s: old reviewer not found", op)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	excludeList := append(pr.AssignedReviewers, pr.AuthorID)

	query := `
		SELECT id FROM users 
		WHERE team_name = $1 AND is_active = true
	`

	args := []interface{}{oldReviewerTeam}
	if len(excludeList) > 0 {
		placeholders := ""
		for i, userID := range excludeList {
			if i > 0 {
				placeholders += ", "
			}
			placeholders += fmt.Sprintf("$%d", i+2)
			args = append(args, userID)
		}
		query += fmt.Sprintf(" AND id NOT IN (%s)", placeholders)
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var candidates []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return "", fmt.Errorf("%s: %w", op, err)
		}
		candidates = append(candidates, userID)
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("%s: no active replacement candidate in team", op)
	}

	newReviewerID := selectRandomReviewers(candidates, 1)[0]

	_, err = db.conn.Exec(`UPDATE pr_fk_reviewer SET user_id = $1 WHERE pr_id = $2 AND user_id = $3`,
		newReviewerID, prID, oldReviewerID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return newReviewerID, nil
}

func (db *DB) GetPullRequestsByReviewer(userID string) ([]handlers.PullRequestShort, error) {
	const op = "Storage.GetPullRequestsByReviewer"

	query := `
		SELECT DISTINCT pr.id, pr.title, pr.authorId, pr.status
		FROM pull_requests pr
		JOIN pr_fk_reviewer pfr ON pr.id = pfr.pr_id
		WHERE pfr.user_id = $1
	`

	rows, err := db.conn.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var prs []handlers.PullRequestShort
	for rows.Next() {
		var pr handlers.PullRequestShort
		if err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		prs = append(prs, pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return prs, nil
}

func selectRandomReviewers(candidates []string, maxCount int) []string {
	if len(candidates) == 0 {
		return []string{}
	}

	count := maxCount
	if len(candidates) < maxCount {
		count = len(candidates)
	}

	rand.Seed(time.Now().UnixNano())

	shuffled := make([]string, len(candidates))
	copy(shuffled, candidates)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled[:count]
}
