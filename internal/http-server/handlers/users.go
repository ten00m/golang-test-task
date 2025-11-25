package handlers

import (
	"net/http"
)

type User struct {
	ID       string `json:"user_id,omitempty"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
	TeamName string `json:"team_name,omitempty"`
}

func UsersSetIsActive(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func UsersGetReview(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
