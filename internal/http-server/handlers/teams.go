package handlers

import (
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

func TeamGet(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
