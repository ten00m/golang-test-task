package handlers

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/render"
	resp "github.com/ten00m/golang-test-task/internal/lib/api/response"
)

type User struct {
	ID       string `json:"user_id,omitempty"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
	TeamName string `json:"team_name,omitempty"`
}

type userActivationSetter interface {
	SetUserIsActive(userID string, isActive bool) (*User, error)
}

func NewUsersSetIsActive(log *slog.Logger, uas userActivationSetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.setIsActive"

		log = log.With(slog.String("op", op))

		var req struct {
			UserID   string `json:"user_id"`
			IsActive bool   `json:"is_active"`
		}

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("Failed to decode request body", slog.Any("error", err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.ErrorResponse("Failed to decode request", resp.StatusError))
			return
		}

		user, err := uas.SetUserIsActive(req.UserID, req.IsActive)
		if err != nil {
			log.Error("Failed to set user is_active", slog.Any("error", err))

			if strings.Contains(err.Error(), "not found") {
				w.WriteHeader(http.StatusNotFound)
				render.JSON(w, r, resp.ErrorResponse("User not found", resp.CodeNotFound))
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.ErrorResponse("Internal error", resp.StatusError))
			return
		}

		log.Info("User is_active updated successfully", slog.String("user_id", req.UserID))

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, map[string]interface{}{"user": user})
	}
}

type pullRequestsByReviewerGetter interface {
	GetPullRequestsByReviewer(userID string) ([]PullRequestShort, error)
}

func NewUsersGetReview(log *slog.Logger, prg pullRequestsByReviewerGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.getReview"

		log = log.With(slog.String("op", op))

		q := r.URL.Query()
		userID := q.Get("user_id")
		if userID == "" {
			log.Warn("missing user_id query param")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.ErrorResponse("user_id query param required", resp.CodeNotFound))
			return
		}

		prs, err := prg.GetPullRequestsByReviewer(userID)
		if err != nil {
			log.Error("Failed to get pull requests for reviewer", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.ErrorResponse("Internal error", resp.StatusError))
			return
		}

		// Return empty array if no PRs found
		if prs == nil {
			prs = []PullRequestShort{}
		}

		log.Info("Retrieved pull requests for reviewer", slog.String("user_id", userID), slog.Int("count", len(prs)))

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, map[string]interface{}{
			"user_id":       userID,
			"pull_requests": prs,
		})
	}
}
