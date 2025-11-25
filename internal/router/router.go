package router

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	mwLogger "github.com/ten00m/golang-test-task/internal/http-server/middleware/logger"
	"github.com/ten00m/golang-test-task/internal/storage"

	handlers "github.com/ten00m/golang-test-task/internal/http-server/handlers"
)

func New(log *slog.Logger, storage *storage.DB) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(mwLogger.New(log))
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Teams
	r.Post("/team/add", handlers.NewAddTeam(log, storage))
	r.Get("/team/get", handlers.NewGetTeam(log, storage))

	// Users
	r.Post("/users/setIsActive", handlers.NewUsersSetIsActive(log, storage))
	r.Get("/users/getReview", handlers.NewUsersGetReview(log, storage))

	// Pull Requests
	r.Post("/pullRequest/create", handlers.NewPullRequestCreate(log, storage))
	r.Post("/pullRequest/merge", handlers.NewPullRequestMerge(log, storage))
	r.Post("/pullRequest/reassign", handlers.NewPullRequestReassign(log, storage))

	// Health
	r.Get("/healthz", handlers.HealthCheck)

	return r
}
