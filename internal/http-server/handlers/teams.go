package handlers

import (
	"net/http"
)

func TeamAdd(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

func TeamGet(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
