package httpresp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOK(t *testing.T) {
	data := map[string]string{"foo": "bar"}

	resp := OK(data)
	if resp.Status != "OK" {
		t.Fatalf("expected status OK, got %q", resp.Status)
	}
	m, ok := resp.Data.(map[string]string)
	if !ok || m["foo"] != "bar" {
		t.Fatalf("unexpected data: %#v", resp.Data)
	}
}

func TestError(t *testing.T) {
	resp := Error(CodePRExists, "pr already exists")

	if resp.Status != "ERROR" {
		t.Fatalf("expected status ERROR, got %q", resp.Status)
	}
	if resp.Error.Code != CodePRExists {
		t.Fatalf("expected code %q, got %q", CodePRExists, resp.Error.Code)
	}
	if resp.Error.Message != "pr already exists" {
		t.Fatalf("unexpected message: %q", resp.Error.Message)
	}
}

func TestWriteOK(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	WriteOK(rr, req, map[string]string{"foo": "bar"})

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, `"status":"OK"`) || !strings.Contains(body, `"foo":"bar"`) {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestWriteError(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	WriteError(rr, req, http.StatusNotFound, CodeNotFound, "not found")

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, `"status":"ERROR"`) ||
		!strings.Contains(body, `"code":"NOT_FOUND"`) ||
		!strings.Contains(body, `"message":"not found"`) {
		t.Fatalf("unexpected body: %s", body)
	}
}
