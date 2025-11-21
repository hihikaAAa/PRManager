package teamhandlerget

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hihikaAAa/PRManager/internal/domain/team"
	"github.com/hihikaAAa/PRManager/internal/domain/user"
	httpresp "github.com/hihikaAAa/PRManager/internal/lib/api/response"
	slogdiscard "github.com/hihikaAAa/PRManager/internal/lib/logger/slogdiscard"
	"github.com/hihikaAAa/PRManager/internal/repository/postgres/repo_errors"
)

type teamGetterMock struct {
	t   *team.Team
	err error
}

func (m *teamGetterMock) GetTeam(ctx context.Context, name string) (*team.Team, error) {
	return m.t, m.err
}

func newTestLogger() *slog.Logger {
	return slogdiscard.NewDiscardLogger()
}

func TestGetTeam_Success(t *testing.T) {
	log := newTestLogger()
	mock := &teamGetterMock{
		t: &team.Team{
			TeamName: "backend",
			Members: []*user.User{
				{ID: "u1", Name: "Alice", IsActive: true},
			},
		},
	}
	h := New(log, mock)

	req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=backend", nil)
	rr := httptest.NewRecorder()

	h(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if body := rr.Body.String(); !containsAll(body, `"team_name":"backend"`, `"user_id":"u1"`) {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestGetTeam_NotFound(t *testing.T) {
	log := newTestLogger()
	mock := &teamGetterMock{err: repo_errors.ErrTeamNotFound}
	h := New(log, mock)

	req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=unknown", nil)
	rr := httptest.NewRecorder()

	h(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
	if body := rr.Body.String(); !containsAll(body, string(httpresp.CodeNotFound)) {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestGetTeam_MissingQuery(t *testing.T) {
	log := newTestLogger()
	mock := &teamGetterMock{}
	h := New(log, mock)

	req := httptest.NewRequest(http.MethodGet, "/team/get", nil)
	rr := httptest.NewRecorder()

	h(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}
