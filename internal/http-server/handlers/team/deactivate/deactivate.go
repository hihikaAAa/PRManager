package teamhandlerdeactivate

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/render"

	httpresp "github.com/hihikaAAa/PRManager/internal/lib/api/response"
	"github.com/hihikaAAa/PRManager/internal/services/teamservice"
	serviceerrors "github.com/hihikaAAa/PRManager/internal/services/serviceErrors"
)

type TeamDeactivator interface {
	DeactivateAndReassign(ctx context.Context, teamName string, userIDs []string) (teamservice.DeactivateResult, error)
}

type deactivateRequest struct {
	TeamName string `json:"team_name"`
	UserIDs []string `json:"user_ids"`
}

type deactivateResponse struct {
	TeamName string `json:"team_name"`
	Deactivated []string `json:"deactivated"`
	ReassignedCount int `json:"reassigned_count"`
	RemovedCount int `json:"removed_count"`
}

func New(log *slog.Logger, svc TeamDeactivator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "internal.http-server.handlers.team.deactivate"

		logger := log.With(slog.String("op", op))

		var req deactivateRequest
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			httpresp.WriteError(w, r, http.StatusBadRequest, httpresp.CodeNotFound, "invalid json")
			return
		}
		if req.TeamName == "" || len(req.UserIDs) == 0 {
			httpresp.WriteError(w, r, http.StatusBadRequest, httpresp.CodeNotFound, "team_name and user_ids are required")
			return
		}

		res, err := svc.DeactivateAndReassign(r.Context(), req.TeamName, req.UserIDs)
		if err != nil {
			switch {
			case errors.Is(err, serviceerrors.ErrTeamNotFound):
				httpresp.WriteError(w, r, http.StatusNotFound, httpresp.CodeNotFound, "team not found")
			case errors.Is(err, serviceerrors.ErrUserNotFound):
				httpresp.WriteError(w, r, http.StatusNotFound, httpresp.CodeNotFound, "user not found")
			default:
				logger.Error("failed to deactivate team users", slog.Any("err", err))
				httpresp.WriteError(w, r, http.StatusInternalServerError, httpresp.CodeNotFound, "internal error")
			}
			return
		}

		resp := deactivateResponse{
			TeamName:        res.TeamName,
			Deactivated:     res.Deactivated,
			ReassignedCount: res.ReassignedCount,
			RemovedCount:    res.RemovedCount,
		}

		logger.Info("team users deactivated and reassigned",
			slog.String("team_name", resp.TeamName),
			slog.Int("deactivated", len(resp.Deactivated)),
			slog.Int("reassigned", resp.ReassignedCount),
			slog.Int("removed", resp.RemovedCount),
		)

		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp)
	}
}
