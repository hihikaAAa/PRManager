package userhandlerisactive

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hihikaAAa/PRManager/internal/domain/user"
	httpresp "github.com/hihikaAAa/PRManager/internal/lib/api/response"
	slogdiscard "github.com/hihikaAAa/PRManager/internal/lib/logger/slogdiscard"
	"github.com/hihikaAAa/PRManager/internal/repository/postgres/repo_errors"
)

type userSetIsActiveMock struct {
	user *user.User
	err  error
}

func (m *userSetIsActiveMock) SetIsActive(ctx context.Context, userID string, active bool) (*user.User, error) {
	return m.user, m.err
}

func newTestLogger() *slog.Logger {
	return slogdiscard.NewDiscardLogger()
}

func TestSetIsActive_Success(t *testing.T) {
	log := newTestLogger()
	mock := &userSetIsActiveMock{
		user: &user.User{ID: "u1", Name: "Alice", TeamName: "backend", IsActive: false},
	}
	h := New(log, mock)

	body := []byte(`{"user_id":"u1","is_active":false}`)
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if body := rr.Body.String(); !strings.Contains(body, `"user_id":"u1"`) {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestSetIsActive_UserNotFound(t *testing.T) {
	log := newTestLogger()
	mock := &userSetIsActiveMock{err: repo_errors.ErrUserNotFound}
	h := New(log, mock)

	body := []byte(`{"user_id":"u1","is_active":true}`)
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), string(httpresp.CodeNotFound)) {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}
