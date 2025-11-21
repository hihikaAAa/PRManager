package pullrequesthandlerreassign

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/render"

	pullrequest "github.com/hihikaAAa/PRManager/internal/domain/pull-request"
	httpresp "github.com/hihikaAAa/PRManager/internal/lib/api/response"
	"github.com/hihikaAAa/PRManager/internal/repository/postgres/repo_errors"
	serviceerrors "github.com/hihikaAAa/PRManager/internal/services/serviceErrors"
)

type prReassigner interface {
	Reassign(ctx context.Context, prID, oldReviewerID string) (*pullrequest.PullRequest, string, error)
}

type prReassignRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID string `json:"old_user_id"`
}

type prReassignResponse struct {
	PullRequest pullRequestItem `json:"pr"`
	ReplacedBy string `json:"replaced_by"`
}

type pullRequestItem struct {
	PullRequestID string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
	Status string `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
	MergedAt *time.Time `json:"merged_at,omitempty"`
}

func New(log *slog.Logger, reassigner prReassigner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "internal.http-server.handlers.pull-request.reassign"

		logger := log.With(slog.String("op", op))

		var req prReassignRequest
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			httpresp.WriteError(w, r, http.StatusBadRequest, httpresp.CodeNotFound, "invalid json")
			return
		}
		if req.PullRequestID == "" || req.OldUserID == "" {
			httpresp.WriteError(w, r, http.StatusBadRequest, httpresp.CodeNotFound, "pull_request_id and old_user_id are required")
			return
		}

		pullreq, replacedBy, err := reassigner.Reassign(r.Context(), req.PullRequestID, req.OldUserID)
		if err != nil {
			switch {
			case errors.Is(err, repo_errors.ErrPRNotFound),errors.Is(err, repo_errors.ErrUserNotFound):
				httpresp.WriteError(w, r, http.StatusNotFound, httpresp.CodeNotFound, "pr or user not found")
			case errors.Is(err, serviceerrors.ErrPRMerged):
				httpresp.WriteError(w, r, http.StatusConflict, httpresp.CodePRMerged, "cannot reassign on merged PR")
			case errors.Is(err, serviceerrors.ErrReviewerNotFound):
				httpresp.WriteError(w, r, http.StatusConflict, httpresp.CodeNotAssigned, "reviewer is not assigned to this PR")
			case errors.Is(err, serviceerrors.ErrNoCandidates):
				httpresp.WriteError(w, r, http.StatusConflict, httpresp.CodeNoCandidate, "no active replacement candidate in team")
			default:
				logger.Error("failed to reassign reviewer", slog.Any("err", err))
				httpresp.WriteError(w, r, http.StatusInternalServerError, httpresp.CodeNotFound, "internal error")
			}
			return
		}

		resp := prReassignResponse{ PullRequest: pullRequestItem{
				PullRequestID:     pullreq.ID,
				PullRequestName:   pullreq.Name,
				AuthorID:          pullreq.AuthorID,
				Status:            string(pullreq.Status),
				AssignedReviewers: pullreq.Reviewers,
				MergedAt:          pullreq.MergedAt,
			},
			ReplacedBy: replacedBy,
		}

		logger.Info("pr reviewer reassigned",slog.String("prID", resp.PullRequest.PullRequestID), slog.String("replaced_by", resp.ReplacedBy))
		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp)
	}
}
