package statshandler

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	httpresp "github.com/hihikaAAa/PRManager/internal/lib/api/response"
	slogdiscard "github.com/hihikaAAa/PRManager/internal/lib/logger/slogdiscard"
	statsservice "github.com/hihikaAAa/PRManager/internal/services/statsservice"
)

type statsGetterMock struct {
	stats statsservice.Stats
	err error
}

func (m *statsGetterMock) GetStats(ctx context.Context) (statsservice.Stats, error) {
	return m.stats, m.err
}

func newTestLogger() *slog.Logger {
	return slogdiscard.NewDiscardLogger()
}

func TestStatsHandler_Success(t *testing.T) {
	log := newTestLogger()
	mock := &statsGetterMock{
		stats: statsservice.Stats{
			TotalPR:  5,
			OpenPR:   3,
			MergedPR: 2,
			Reviewers: []statsservice.ReviewerStat{
				{UserID: "u1", Count: 3},
			},
		},
	}
	h := New(log, mock)
	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	rr := httptest.NewRecorder()
	h(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, `"total_pr":5`) {
		t.Fatalf("unexpected body: %s", body)
	}
	if !strings.Contains(body, `"status":"OK"`) {
		t.Fatalf("response not wrapped by OK: %s", body)
	}
}

func TestStatsHandler_Error(t *testing.T) {
	log := newTestLogger()
	mock := &statsGetterMock{
		err: context.DeadlineExceeded,
	}
	h := New(log, mock)
	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	rr := httptest.NewRecorder()
	h(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), string(httpresp.CodeNotFound)) {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}
