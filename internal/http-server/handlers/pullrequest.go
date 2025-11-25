package handlers

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/render"
	resp "github.com/ten00m/golang-test-task/internal/lib/api/response"
)

type PullRequest struct {
	ID                string   `json:"pull_request_id"`
	Name              string   `json:"pull_request_name"`
	AuthorID          string   `json:"author_id"`
	Status            string   `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
}

type PullRequestShort struct {
	ID       string `json:"pull_request_id"`
	Name     string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
	Status   string `json:"status"`
}

type pullRequestCreator interface {
	CreatePullRequest(prID, prName, authorID string) (*PullRequest, error)
}

func NewPullRequestCreate(log *slog.Logger, prc pullRequestCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.pullrequest.create"

		log = log.With(slog.String("op", op))

		var req struct {
			PullRequestID   string `json:"pull_request_id"`
			PullRequestName string `json:"pull_request_name"`
			AuthorID        string `json:"author_id"`
		}

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("Failed to decode request body", slog.Any("error", err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.ErrorResponse("Failed to decode request", resp.StatusError))
			return
		}

		pr, err := prc.CreatePullRequest(req.PullRequestID, req.PullRequestName, req.AuthorID)
		if err != nil {
			log.Error("Failed to create PR", slog.Any("error", err))

			if strings.Contains(err.Error(), "PR already exists") {
				w.WriteHeader(http.StatusConflict)
				render.JSON(w, r, resp.ErrorResponse("PR id already exists", resp.CodePRExists))
				return
			}

			if strings.Contains(err.Error(), "not found") {
				w.WriteHeader(http.StatusNotFound)
				render.JSON(w, r, resp.ErrorResponse("Author or team not found", resp.CodeNotFound))
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.ErrorResponse("Internal error", resp.StatusError))
			return
		}

		log.Info("PR created successfully", slog.String("pr_id", req.PullRequestID))

		w.WriteHeader(http.StatusCreated)
		render.JSON(w, r, map[string]interface{}{"pr": pr})
	}
}

type pullRequestMerger interface {
	MergePullRequest(prID string) (*PullRequest, error)
}

func NewPullRequestMerge(log *slog.Logger, prm pullRequestMerger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.pullrequest.merge"

		log = log.With(slog.String("op", op))

		var req struct {
			PullRequestID string `json:"pull_request_id"`
		}

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("Failed to decode request body", slog.Any("error", err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.ErrorResponse("Failed to decode request", resp.StatusError))
			return
		}

		pr, err := prm.MergePullRequest(req.PullRequestID)
		if err != nil {
			log.Error("Failed to merge PR", slog.Any("error", err))

			if strings.Contains(err.Error(), "not found") {
				w.WriteHeader(http.StatusNotFound)
				render.JSON(w, r, resp.ErrorResponse("PR not found", resp.CodeNotFound))
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.ErrorResponse("Internal error", resp.StatusError))
			return
		}

		log.Info("PR merged successfully", slog.String("pr_id", req.PullRequestID))

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, map[string]interface{}{"pr": pr})
	}
}

type pullRequestReassigner interface {
	ReassignReviewer(prID, oldReviewerID string) (string, error)
	GetPullRequest(prID string) (*PullRequest, error)
}

func NewPullRequestReassign(log *slog.Logger, prr pullRequestReassigner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.pullrequest.reassign"

		log = log.With(slog.String("op", op))

		var req struct {
			PullRequestID string `json:"pull_request_id"`
			OldUserID     string `json:"old_user_id"`
		}

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("Failed to decode request body", slog.Any("error", err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.ErrorResponse("Failed to decode request", resp.StatusError))
			return
		}

		newReviewerID, err := prr.ReassignReviewer(req.PullRequestID, req.OldUserID)
		if err != nil {
			log.Error("Failed to reassign reviewer", slog.Any("error", err))

			if strings.Contains(err.Error(), "not found") {
				w.WriteHeader(http.StatusNotFound)
				render.JSON(w, r, resp.ErrorResponse("PR or user not found", resp.CodeNotFound))
				return
			}

			if strings.Contains(err.Error(), "cannot reassign on merged PR") {
				w.WriteHeader(http.StatusConflict)
				render.JSON(w, r, resp.ErrorResponse("cannot reassign on merged PR", resp.CodePRMerged))
				return
			}

			if strings.Contains(err.Error(), "reviewer is not assigned") {
				w.WriteHeader(http.StatusConflict)
				render.JSON(w, r, resp.ErrorResponse("reviewer is not assigned to this PR", resp.CodeNotAssigned))
				return
			}

			if strings.Contains(err.Error(), "no active replacement candidate") {
				w.WriteHeader(http.StatusConflict)
				render.JSON(w, r, resp.ErrorResponse("no active replacement candidate in team", resp.CodeNoCandidate))
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.ErrorResponse("Internal error", resp.StatusError))
			return
		}

		// Get updated PR
		pr, err := prr.GetPullRequest(req.PullRequestID)
		if err != nil {
			log.Error("Failed to get updated PR", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.ErrorResponse("Internal error", resp.StatusError))
			return
		}

		log.Info("Reviewer reassigned successfully",
			slog.String("pr_id", req.PullRequestID),
			slog.String("old_reviewer", req.OldUserID),
			slog.String("new_reviewer", newReviewerID))

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, map[string]interface{}{
			"pr":          pr,
			"replaced_by": newReviewerID,
		})
	}
}
