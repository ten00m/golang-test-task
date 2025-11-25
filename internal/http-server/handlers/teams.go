package handlers

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/render"
	resp "github.com/ten00m/golang-test-task/internal/lib/api/response"
)

type Team struct {
	Name    string `json:"team_name"`
	Members []User `json:"members"`
}

type teamAdder interface {
	AddTeam(team Team) error
}

func NewAddTeam(log *slog.Logger, ta teamAdder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "router.teams.add"

		log = log.With(slog.String("op", op))

		var req Team

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("Failed to decode rquest body: %s", slog.Any("%s", err))

			w.WriteHeader(http.StatusBadRequest)

			render.JSON(w, r, resp.ErrorResponse("Failed to create team: bad json", resp.StatusError))
			return
		}

		log.Info("Request decoded successfully")

		err = ta.AddTeam(req)
		if err != nil {
			log.Error("Failed to add team: %s", slog.Any("%s", err))

			w.WriteHeader(http.StatusBadRequest)

			render.JSON(w, r, resp.ErrorResponse("team_name already exists", resp.CodeTeamExists))
			return
		}

		w.WriteHeader(http.StatusCreated)
		render.JSON(w, r, req)
	}

}

type teamGetter interface {
	GetTeam(teamName string) (Team, error)
}

func NewGetTeam(log *slog.Logger, tg teamGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "router.teams.get"

		log = log.With(slog.String("op", op))

		q := r.URL.Query()
		teamName := q.Get("team_name")
		if teamName == "" {
			log.Warn("missing team_name query param")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.ErrorResponse("team_name query param required", resp.CodeNotFound))
			return
		}

		team, err := tg.GetTeam(teamName)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				log.Info("team not found", slog.String("team_name", teamName))
				w.WriteHeader(http.StatusNotFound)
				render.JSON(w, r, resp.ErrorResponse("team not found", resp.CodeNotFound))
				return
			}

			log.Error("failed to get team", slog.Any("err", err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.ErrorResponse("internal error", resp.CodeNotFound))
			return
		}

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, team)
	}
}
