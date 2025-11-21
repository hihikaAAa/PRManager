package userhandlergetreview

import (
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
)

type userReviewGetterMock struct {
	prs []pullrequest.PullRequestShort
	err error
}

func (m *userReviewGetterMock) GetReviewPRs(ctx context.Context, userID string) ([]pullrequest.PullRequestShort, error) {
	return m.prs, m.err
}

func newTestLogger() *slog.Logger {
	return slogdiscard.NewDiscardLogger()
}

func TestGetReview_Success(t *testing.T) {
	log := newTestLogger()
	mock := &userReviewGetterMock{
		prs: []pullrequest.PullRequestShort{
			{ID: "pr-1", Name: "Add search", AuthorID: "u1", Status: pullrequest.StatusOpen},
		},
	}
	h := New(log, mock)
	req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=u2", nil)
	rr := httptest.NewRecorder()
	h(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Add search") && !strings.Contains(body, `"pull_requests"`) {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestGetReview_UserNotFound(t *testing.T) {
	log := newTestLogger()
	mock := &userReviewGetterMock{err: repo_errors.ErrUserNotFound}
	h := New(log, mock)
	req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=u2", nil)
	rr := httptest.NewRecorder()
	h(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), string(httpresp.CodeNotFound)) {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}

func TestGetReview_MissingUserID(t *testing.T) {
	log := newTestLogger()
	mock := &userReviewGetterMock{}
	h := New(log, mock)
	req := httptest.NewRequest(http.MethodGet, "/users/getReview", nil)
	rr := httptest.NewRecorder()
	h(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), string(httpresp.CodeNotFound)) {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}
