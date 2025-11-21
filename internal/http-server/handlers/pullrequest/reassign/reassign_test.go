package pullrequesthandlerreassign

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	pullrequest "github.com/hihikaAAa/PRManager/internal/domain/pull-request"
	httpresp "github.com/hihikaAAa/PRManager/internal/lib/api/response"
	slogdiscard "github.com/hihikaAAa/PRManager/internal/lib/logger/slogdiscard"
	"github.com/hihikaAAa/PRManager/internal/repository/postgres/repo_errors"
	serviceerrors "github.com/hihikaAAa/PRManager/internal/services/serviceErrors"
)

type reassignerMock struct {
	pr *pullrequest.PullRequest
	replacedBy string
	err error
}

func (m *reassignerMock) Reassign(ctx context.Context, prID, oldReviewerID string) (*pullrequest.PullRequest, string, error) {
	return m.pr, m.replacedBy, m.err
}

func newTestLogger() *slog.Logger {
	return slogdiscard.NewDiscardLogger() 
}

func TestReassign_Success(t *testing.T) {
	log := newTestLogger()
	mock := &reassignerMock{
		pr: &pullrequest.PullRequest{
			ID:        "pr-1",
			Name:      "Add",
			AuthorID:  "u1",
			Status:    pullrequest.StatusOpen,
			Reviewers: []string{"u3", "u5"},
			MergedAt:  nil,
		},
		replacedBy: "u5",
	}
	h := New(log, mock)

	body := []byte(`{"pull_request_id":"pr-1","old_user_id":"u2"}`)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if body := rr.Body.String(); !strings.Contains(body, `"replaced_by":"u5"`) {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestReassign_NoCandidate(t *testing.T) {
	log := newTestLogger()
	mock := &reassignerMock{err: serviceerrors.ErrNoCandidates}
	h := New(log, mock)

	body := []byte(`{"pull_request_id":"pr-1","old_user_id":"u2"}`)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), string(httpresp.CodeNoCandidate)) {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}

func TestReassign_NotFound(t *testing.T) {
	log := newTestLogger()
	mock := &reassignerMock{err: repo_errors.ErrPRNotFound}
	h := New(log, mock)

	body := []byte(`{"pull_request_id":"missing","old_user_id":"u2"}`)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
