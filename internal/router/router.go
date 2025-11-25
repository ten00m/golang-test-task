package router

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	mwLogger "github.com/ten00m/golang-test-task/internal/http-server/middleware/logger"
)

func New(log *slog.Logger) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(mwLogger.New(log))
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	return r
}
