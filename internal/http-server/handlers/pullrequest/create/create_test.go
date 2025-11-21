package pullrequesthandlercreate

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

type prCreatorMock struct {
	pr *pullrequest.PullRequest
	err error
}

func (m *prCreatorMock) Create(ctx context.Context, id, name, authorID string) (*pullrequest.PullRequest, error) {
	return m.pr, m.err
}

func newTestLogger() *slog.Logger {
	 return slogdiscard.NewDiscardLogger() 
	}

func TestCreatePR_Success(t *testing.T) {
	log := newTestLogger()
	mock := &prCreatorMock{
		pr: &pullrequest.PullRequest{
			ID:        "pr-1",
			Name:      "Add search",
			AuthorID:  "u1",
			Status:    pullrequest.StatusOpen,
			Reviewers: []string{"u2", "u3"},
		},
	}
	h := New(log, mock)

	body := []byte(`{
		"pull_request_id":"pr-1",
		"pull_request_name":"Add search",
		"author_id":"u1"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if body := rr.Body.String(); !strings.Contains(body, `"pull_request_id":"pr-1"`) {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestCreatePR_AlreadyExists(t *testing.T) {
	log := newTestLogger()
	mock := &prCreatorMock{err: serviceerrors.ErrPRExists}
	h := New(log, mock)

	body := []byte(`{"pull_request_id":"pr-1","pull_request_name":"Add","author_id":"u1"}`)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), string(httpresp.CodePRExists)) {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}

func TestCreatePR_AuthorOrTeamNotFound(t *testing.T) {
	log := newTestLogger()
	mock := &prCreatorMock{err: repo_errors.ErrUserNotFound}
	h := New(log, mock)

	body := []byte(`{"pull_request_id":"pr-1","pull_request_name":"Add","author_id":"uX"}`)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
