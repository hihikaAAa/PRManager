package teamhandleradd

import (
	"context"
	"log/slog"
	"net/http"
	"errors"

	"github.com/go-chi/render"

	"github.com/hihikaAAa/PRManager/internal/domain/user"
	httpresp "github.com/hihikaAAa/PRManager/internal/lib/api/response"
	serviceerrors "github.com/hihikaAAa/PRManager/internal/services/serviceErrors"
)

type TeamAdder interface{
	AddTeam (ctx context.Context, teamName string, members []*user.User) error
}

type teamMemberRequest struct{
	UserID string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool `json:"is_active"`
}

type addTeamRequest struct {
	TeamName string `json:"team_name"`
	Members []teamMemberRequest `json:"members"`
}

type addTeamResponse struct {
	Team struct {
		TeamName string `json:"team_name"`
		Members []teamMemberRequest `json:"members"`
	} `json:"team"`
}

func New(log *slog.Logger, teamAdder TeamAdder) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		const op = "internal.http-server.handlers.team.add"

		logger := log.With(slog.String("op", op))

		var req addTeamRequest
		if err := render.DecodeJSON(r.Body, &req); err != nil{
			httpresp.WriteError(w, r, http.StatusBadRequest, httpresp.CodeNotFound, "invalid json")
			return
		}
		if req.TeamName == "" {
			httpresp.WriteError(w, r, http.StatusBadRequest, httpresp.CodeNotFound, "team_name is required")
			return
		}

		members := make([]*user.User, 0, len(req.Members))
		for _, m := range req.Members{
			members = append(members, &user.User{
				ID: m.UserID,
				Name: m.Username,
				IsActive: m.IsActive,
				TeamName: req.TeamName,
			})
		}
		
		err := teamAdder.AddTeam(r.Context(), req.TeamName, members)
		if err != nil{
			switch{
			case errors.Is(err, serviceerrors.ErrTeamExists):
				httpresp.WriteError(w, r, http.StatusBadRequest, httpresp.CodeTeamExists, "team_name already exists")
			default:
				logger.Error("failed to add team", slog.Any("err", err))
				httpresp.WriteError(w, r, http.StatusInternalServerError, httpresp.CodeNotFound, "internal error")
			}
			return
		}
		resp := addTeamResponse{}
		resp.Team.TeamName = req.TeamName
		resp.Team.Members = req.Members

		logger.Info("team created", slog.String("team_name", req.TeamName))

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, resp)
	}
}