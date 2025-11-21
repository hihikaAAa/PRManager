package pullrequesthandlersmerge

import(
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/render"
	pullrequest "github.com/hihikaAAa/PRManager/internal/domain/pull-request"
	httpresp "github.com/hihikaAAa/PRManager/internal/lib/api/response"
	"github.com/hihikaAAa/PRManager/internal/repository/postgres/repo_errors"
)

type PrMerger interface{
	Merge(ctx context.Context, id string)(*pullrequest.PullRequest, error)
}

type prMergerRequest struct{
	PullRequestID string `json:"pull_request_id"`
}

type prMergerResponse struct{
	PullRequest pullRequestItem `json:"pr"`
}

type pullRequestItem struct{
	PullRequestID string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
	Status string `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
	MergedAt *time.Time `json:"mergedAt"`
}

func New(log *slog.Logger, prMerger PrMerger) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		const op = "internal.http-server.handlers.pull-request.merge"

		logger := log.With(slog.String("op", op))

		var req prMergerRequest
		if err := render.DecodeJSON(r.Body, &req); err != nil{
			httpresp.WriteError(w,r,http.StatusBadRequest,httpresp.CodeNotFound,"invalid json")
			return
		}
		if req.PullRequestID == ""{
			httpresp.WriteError(w,r,http.StatusBadRequest, httpresp.CodeNotFound,"pr id is required")
			return
		}

		pullreq, err := prMerger.Merge(r.Context(),req.PullRequestID)
		if err != nil{
			switch{
				case errors.Is(err,repo_errors.ErrPRNotFound):
					httpresp.WriteError(w,r,http.StatusNotFound, httpresp.CodeNotFound, "pr is not found")
				case errors.Is(err, repo_errors.ErrPRMerged):
					resp := buildResponse(pullreq)
					render.Status(r, http.StatusOK)
					render.JSON(w, r, resp)
					return
				default: 
					logger.Error("failed to merge PR", slog.Any("err", err))
					httpresp.WriteError(w, r, http.StatusInternalServerError, httpresp.CodeNotFound, "internal error")
			}
			return
		}

		resp := buildResponse(pullreq)

		logger.Info("pr merged", slog.String("prID", resp.PullRequest.PullRequestID))
		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp)
	}
}

func buildResponse(pr *pullrequest.PullRequest) prMergerResponse {
	return prMergerResponse{
		PullRequest: pullRequestItem{
			PullRequestID:     pr.ID,
			PullRequestName:   pr.Name,
			AuthorID:          pr.AuthorID,
			Status:            string(pr.Status),
			AssignedReviewers: pr.Reviewers,
			MergedAt:          pr.MergedAt,
		},
	}
}