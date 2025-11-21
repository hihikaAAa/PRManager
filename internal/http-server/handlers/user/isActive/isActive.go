package userhandlerisactive

import (
	"context"
	"log/slog"
	"net/http"
	"errors"

	"github.com/go-chi/render"
	"github.com/hihikaAAa/PRManager/internal/domain/user"
	httpresp "github.com/hihikaAAa/PRManager/internal/lib/api/response"
	"github.com/hihikaAAa/PRManager/internal/repository/postgres/repo_errors"
)

type UserSetIsActive interface{
	SetIsActive(ctx context.Context,userID string, isActive bool) (*user.User,error)
}

type userIsActiveRequest struct{
	UserID string `json:"user_id"`
	IsActive bool `json:"is_active"`
}

type userIsActiveResponce struct{
	User userIsActiveItem `json:"user"`
}

type userIsActiveItem struct {
    UserID   string `json:"user_id"`
    Username string `json:"username"`
    TeamName string `json:"team_name"`
    IsActive bool   `json:"is_active"`
}

func New(log *slog.Logger, userSetIsActive UserSetIsActive) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		const op = "internal.http-server.handlers.user.isActive"

		logger := log.With(slog.String("op", op))

		var req userIsActiveRequest
		if err := render.DecodeJSON(r.Body, &req); err != nil{
			httpresp.WriteError(w, r, http.StatusBadRequest, httpresp.CodeNotFound, "invalid json")
			return
		}
		if req.UserID == ""{
			httpresp.WriteError(w,r,http.StatusBadRequest, httpresp.CodeNotFound, "no users found")
			return
		}

		user, err := userSetIsActive.SetIsActive(r.Context(),req.UserID,req.IsActive)
		if err != nil{
			switch{
			case errors.Is(err, repo_errors.ErrUserNotFound):
				httpresp.WriteError(w, r, http.StatusNotFound, httpresp.CodeNotFound, "user not found")
			default:
				logger.Error("failed to set user is_active", slog.Any("err", err))
				httpresp.WriteError(w, r, http.StatusInternalServerError, httpresp.CodeNotFound, "internal error")
			}
			return
		}
		resp := userIsActiveResponce{
			User: userIsActiveItem{
			UserID: user.ID,
			Username: user.Name,
			TeamName: user.TeamName,
			IsActive: user.IsActive,
			},
		}
		logger.Info("users isActive status updated", slog.String("userID", resp.User.UserID))
		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp)
	}
}

