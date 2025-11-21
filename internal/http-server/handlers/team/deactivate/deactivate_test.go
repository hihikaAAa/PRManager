package teamhandlerdeactivate

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	httpresp "github.com/hihikaAAa/PRManager/internal/lib/api/response"
	slogdiscard "github.com/hihikaAAa/PRManager/internal/lib/logger/slogdiscard"
	teamservice "github.com/hihikaAAa/PRManager/internal/services/teamservice"
)

type deactivatorMock struct {
	result teamservice.DeactivateResult
	err error
	calledTeam string
	calledUsers []string
}

func (m *deactivatorMock) DeactivateAndReassign(ctx context.Context,teamName string,userIDs []string,) (teamservice.DeactivateResult, error) {
	m.calledTeam = teamName
	m.calledUsers = userIDs
	return m.result, m.err
}

func newTestLogger() *slog.Logger {
	return slogdiscard.NewDiscardLogger()
}

func TestDeactivate_InvalidJSON(t *testing.T) {
	log := newTestLogger()
	mock := &deactivatorMock{}
	h := New(log, mock)

	body := []byte(`{ "team_name": "backend", `)
	req := httptest.NewRequest(http.MethodPost, "/team/deactivate", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), string(httpresp.CodeNotFound)) {
		t.Fatalf("expected error code NOT_FOUND, got body: %s", rr.Body.String())
	}
	if mock.calledTeam != "" {
		t.Fatalf("service must not be called on invalid json")
	}
}

func TestDeactivate_ValidationError(t *testing.T) {
	log := newTestLogger()
	mock := &deactivatorMock{}
	h := New(log, mock)

	body := []byte(`{
		"team_name": "",
		"user_ids": []
	}`)
	req := httptest.NewRequest(http.MethodPost, "/team/deactivate", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "team_name and user_ids are required") {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
	if mock.calledTeam != "" {
		t.Fatalf("service must not be called on validation error")
	}
}

func TestDeactivate_InternalError(t *testing.T) {
	log := newTestLogger()
	mockErr := assertError{}
	mock := &deactivatorMock{err: mockErr}
	h := New(log, mock)

	body := []byte(`{
		"team_name": "backend",
		"user_ids": ["u2"]
	}`)
	req := httptest.NewRequest(http.MethodPost, "/team/deactivate", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), string(httpresp.CodeNotFound)) {
		t.Fatalf("expected generic NOT_FOUND code in internal error, got body: %s", rr.Body.String())
	}
}

type assertError struct{}

func (assertError) Error() string { return "some internal error" }
