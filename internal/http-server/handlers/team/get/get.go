package teamhandlerget

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/render"

	"github.com/hihikaAAa/PRManager/internal/domain/team"
	httpresp "github.com/hihikaAAa/PRManager/internal/lib/api/response"
	"github.com/hihikaAAa/PRManager/internal/repository/postgres/repo_errors"
)

type TeamGetter interface{
	GetTeam(ctx context.Context, teamName string)(*team.Team, error)
}

type teamMemberResponse struct {
	UserID string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool `json:"is_active"`
}

type getTeamResponse struct {
	TeamName string `json:"team_name"`
	Members []teamMemberResponse `json:"members"`
}

func New(log *slog.Logger, teamGetter TeamGetter)http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		const op = "internal.http-server.handlers.team.add"

		logger := log.With(slog.String("op", op))

		teamName := r.URL.Query().Get("team_name")
		if teamName == "" {
			httpresp.WriteError(w, r, http.StatusBadRequest, httpresp.CodeNotFound, "team_name is required")
			return
		}

		t, err := teamGetter.GetTeam(r.Context(), teamName)
		if err != nil {
			switch {
			case errors.Is(err, repo_errors.ErrTeamNotFound):
				httpresp.WriteError(w, r, http.StatusNotFound, httpresp.CodeNotFound, "team not found")
			default:
				logger.Error("failed to get team", slog.Any("err", err))
				httpresp.WriteError(w, r, http.StatusInternalServerError, httpresp.CodeNotFound, "internal error")
			}
			return
		}

		resp := getTeamResponse{
			TeamName: t.TeamName,
			Members: make([]teamMemberResponse, 0, len(t.Members)),
		}

		for _, m := range t.Members {
			resp.Members = append(resp.Members, teamMemberResponse{
				UserID: m.ID,
				Username: m.Name,
				IsActive: m.IsActive,
			})
		}

		logger.Info("team fetched", slog.String("team_name", teamName))

		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp)
	}
}