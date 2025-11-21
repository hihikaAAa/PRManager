package userhandlergetreview

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/render"
	pullrequest "github.com/hihikaAAa/PRManager/internal/domain/pull-request"
	httpresp "github.com/hihikaAAa/PRManager/internal/lib/api/response"
	"github.com/hihikaAAa/PRManager/internal/repository/postgres/repo_errors"
)

type UserReviewGetter interface{
	GetReviewPRs(ctx context.Context, userID string)([]pullrequest.PullRequestShort, error)
}

type userGetReviewResponce struct{
	UserID string `json:"user_id"`
	PullRequests []pullrequest.PullRequestShort `json:"pull_requests"`
}

func New(log *slog.Logger, userReviewGetter UserReviewGetter) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		const op = "internal.http-server.handlers.user.getReview"

		logger := log.With(slog.String("op", op))

		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			httpresp.WriteError(w, r, http.StatusBadRequest, httpresp.CodeNotFound, "user_id is required")
			return
		}

		pullrequests, err := userReviewGetter.GetReviewPRs(r.Context(), userID)
		if err != nil{
			switch{
			case errors.Is(err,repo_errors.ErrUserNotFound):
				httpresp.WriteError(w, r, http.StatusNotFound, httpresp.CodeNotFound, "user not found")
			default:
				logger.Error("failed to fetch review PRs", slog.Any("err", err))
				httpresp.WriteError(w, r, http.StatusInternalServerError, httpresp.CodeNotFound, "internal error")
			}
			return
		}

		resp := userGetReviewResponce{
			UserID: userID,
			PullRequests: pullrequests,
		}
		logger.Info("user review PRs fetched", slog.String("userID", resp.UserID))
		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp)
	}
}