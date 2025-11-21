package teamhandleradd

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hihikaAAa/PRManager/internal/domain/user"
	httpresp "github.com/hihikaAAa/PRManager/internal/lib/api/response"
	slogdiscard "github.com/hihikaAAa/PRManager/internal/lib/logger/slogdiscard"
	serviceerrors "github.com/hihikaAAa/PRManager/internal/services/serviceErrors"
)

type teamAdderMock struct {
	lastTeamName string
	lastMembers []*user.User
	err error
}

func (m *teamAdderMock) AddTeam(ctx context.Context, teamName string, members []*user.User) error {
	m.lastTeamName = teamName
	m.lastMembers = members
	return m.err
}

func newTestLogger() *slog.Logger {
	return slogdiscard.NewDiscardLogger()
}

func TestAddTeam_Success(t *testing.T) {
	log := newTestLogger()
	mock := &teamAdderMock{}
	h := New(log, mock)

	body := []byte(`{
		"team_name":"backend",
		"members":[{"user_id":"u1","username":"Alice","is_active":true}]
	}`)

	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}

	if mock.lastTeamName != "backend" {
		t.Fatalf("expected teamName=backend, got %q", mock.lastTeamName)
	}
	if len(mock.lastMembers) != 1 || mock.lastMembers[0].ID != "u1" {
		t.Fatalf("unexpected members: %#v", mock.lastMembers)
	}
}

func TestAddTeam_TeamExists(t *testing.T) {
	log := newTestLogger()
	mock := &teamAdderMock{err: serviceerrors.ErrTeamExists}
	h := New(log, mock)

	body := []byte(`{"team_name":"backend","members":[]}`)
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte(httpresp.CodeTeamExists)) {
		t.Fatalf("error body must contain code %q, got %s", httpresp.CodeTeamExists, rr.Body.String())
	}
}

func TestAddTeam_BadJSON(t *testing.T) {
	log := newTestLogger()
	mock := &teamAdderMock{}
	h := New(log, mock)

	body := []byte(`{invalid json`)
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
