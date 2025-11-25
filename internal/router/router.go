package router

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	mwLogger "github.com/ten00m/golang-test-task/internal/http-server/middleware/logger"

	handlers "github.com/ten00m/golang-test-task/internal/http-server/handlers"
)

func New(log *slog.Logger) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(mwLogger.New(log))
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Teams
	r.Post("/team/add", handlers.TeamAdd)
	r.Get("/team/get", handlers.TeamGet)

	// Users
	r.Post("/users/setIsActive", handlers.UsersSetIsActive)
	r.Get("/users/getReview", handlers.UsersGetReview)

	// Pull Requests
	r.Post("/pullRequest/create", handlers.PullRequestCreate)
	r.Post("/pullRequest/merge", handlers.PullRequestMerge)
	r.Post("/pullRequest/reassign", handlers.PullRequestReassign)

	// Health
	r.Get("/healthz", handlers.HealthCheck)

	return r
}
