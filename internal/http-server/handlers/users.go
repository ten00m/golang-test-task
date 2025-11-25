package handlers

import (
	"net/http"
)

func UsersSetIsActive(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func UsersGetReview(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
