package pullrequesthandlercreate

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/render"
	pullrequest "github.com/hihikaAAa/PRManager/internal/domain/pull-request"
	httpresp "github.com/hihikaAAa/PRManager/internal/lib/api/response"
	"github.com/hihikaAAa/PRManager/internal/repository/postgres/repo_errors"
	serviceerrors "github.com/hihikaAAa/PRManager/internal/services/serviceErrors"
)

type PrCreator interface{
	Create(ctx context.Context, id,name,authorID string)(*pullrequest.PullRequest, error)
}

type prCreateRequest struct{
	PullRequestID string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
}

type prCreateResponse struct{
	PullRequest pullRequestItem `json:"pr"`
}

type pullRequestItem struct{
	PullRequestID string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
	Status string `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
}

func New(log *slog.Logger, prCreator PrCreator) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		const op = "internal.http-server.handlers.pull-request.create"

		logger := log.With(slog.String("op",op))

		var req prCreateRequest
		if err := render.DecodeJSON(r.Body, &req); err != nil{
			httpresp.WriteError(w, r, http.StatusBadRequest, httpresp.CodeNotFound, "invalid json")
			return
		}
		if req.PullRequestID == "" || req.AuthorID == "" || req.PullRequestName == ""{
			httpresp.WriteError(w,r,http.StatusBadRequest,httpresp.CodeNotFound, "missing required paramethres")
			return
		}

		pullreq, err := prCreator.Create(r.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
		if err != nil {
			switch {
			case errors.Is(err, serviceerrors.ErrPRExists):
				httpresp.WriteError(w, r, http.StatusConflict, httpresp.CodePRExists, "PR id already exists")
			case errors.Is(err, repo_errors.ErrUserNotFound),errors.Is(err, repo_errors.ErrTeamNotFound):
				httpresp.WriteError(w, r, http.StatusNotFound, httpresp.CodeNotFound, "author or team not found")
			default:
				logger.Error("failed to create PR", slog.Any("err", err))
				httpresp.WriteError(w, r, http.StatusInternalServerError, httpresp.CodeNotFound, "internal error")
			}
			return
		}
		resp := prCreateResponse{PullRequest: pullRequestItem{
			PullRequestID: pullreq.ID,
			PullRequestName: pullreq.Name,
			AuthorID: pullreq.AuthorID,
			Status: string(pullreq.Status),
			AssignedReviewers: pullreq.Reviewers,
		}}

		logger.Info("pr created", slog.String("prID", resp.PullRequest.PullRequestID))
		render.Status(r, http.StatusCreated)
		render.JSON(w, r, resp)
	}
}