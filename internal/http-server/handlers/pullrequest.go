package handlers

import (
	"net/http"
)

func PullRequestCreate(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

func PullRequestMerge(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func PullRequestReassign(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
