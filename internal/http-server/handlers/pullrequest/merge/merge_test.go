package pullrequesthandlersmerge

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	pullrequest "github.com/hihikaAAa/PRManager/internal/domain/pull-request"
	httpresp "github.com/hihikaAAa/PRManager/internal/lib/api/response"
	slogdiscard "github.com/hihikaAAa/PRManager/internal/lib/logger/slogdiscard"
	"github.com/hihikaAAa/PRManager/internal/repository/postgres/repo_errors"
)

type prMergerMock struct {
	pr *pullrequest.PullRequest
	err error
}

func (m *prMergerMock) Merge(ctx context.Context, id string) (*pullrequest.PullRequest, error) {
	return m.pr, m.err
}

func newTestLogger() *slog.Logger { 
	return slogdiscard.NewDiscardLogger() 
}

func TestMergePR_Success(t *testing.T) {
	log := newTestLogger()
	now := time.Now().UTC()
	mock := &prMergerMock{
		pr: &pullrequest.PullRequest{
			ID:       "pr-1",
			Name:     "Add",
			AuthorID: "u1",
			Status:   pullrequest.StatusMerged,
			MergedAt: &now,
		},
	}
	h := New(log, mock)

	body := []byte(`{"pull_request_id":"pr-1"}`)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if body := rr.Body.String(); !strings.Contains(body, `"status":"MERGED"`) {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestMergePR_NotFound(t *testing.T) {
	log := newTestLogger()
	mock := &prMergerMock{err: repo_errors.ErrPRNotFound}
	h := New(log, mock)

	body := []byte(`{"pull_request_id":"unknown"}`)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), string(httpresp.CodeNotFound)) {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}
