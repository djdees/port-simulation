package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/djdees/port-simulation/internal/server/data"
	"github.com/djdees/port-simulation/internal/server/handler"
	"log/slog"
	"os"
)

func newTestHandler(t *testing.T) http.Handler {
	t.Helper()
	d, err := data.Load()
	if err != nil {
		t.Fatalf("loading test data: %v", err)
	}
	store := handler.NewStore(d)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	return handler.New(store, logger, nil) // nil = no SPIFFE middleware for unit tests
}

func TestHealthLive(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/v1/health/live", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHealthReady(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/v1/health/ready", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestVesselsList(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/v1/vessels", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	items, ok := resp["items"].([]any)
	if !ok || len(items) == 0 {
		t.Error("expected at least one vessel in items")
	}
	if resp["total"] == nil {
		t.Error("expected total field in response")
	}
}

func TestVesselsFilterByKind(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/v1/vessels?kind=tanker", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	items := resp["items"].([]any)
	for _, item := range items {
		m := item.(map[string]any)
		if m["kind"] != "tanker" {
			t.Errorf("expected only tankers, got kind=%v", m["kind"])
		}
	}
}

func TestVesselsFilterByInvalidKind(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/v1/vessels?kind=zeppelin", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid kind, got %d", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if ct != "application/problem+json" {
		t.Errorf("expected problem+json content type, got %q", ct)
	}
}

func TestVesselByID(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/v1/vessels/VSL-001", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var vessel map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &vessel); err != nil {
		t.Fatalf("decoding vessel: %v", err)
	}
	if vessel["id"] != "VSL-001" {
		t.Errorf("expected id=VSL-001, got %v", vessel["id"])
	}
}

func TestVesselByIDNotFound(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/v1/vessels/VSL-999", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestPagination(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/v1/vessels?limit=2&offset=0", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	items := resp["items"].([]any)
	if len(items) > 2 {
		t.Errorf("expected at most 2 items with limit=2, got %d", len(items))
	}
}

func TestPaginationBadLimit(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/v1/vessels?limit=-1", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for negative limit, got %d", w.Code)
	}
}

func TestArrivalsList(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/v1/arrivals", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	items := resp["items"].([]any)
	for _, item := range items {
		m := item.(map[string]any)
		if m["event_type"] != "arrival" {
			t.Errorf("arrivals endpoint returned event_type=%v", m["event_type"])
		}
	}
}

func TestDeparturesList(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/v1/departures", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestCargoFilter(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/v1/cargo?commodity=petroleum", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	items := resp["items"].([]any)
	for _, item := range items {
		m := item.(map[string]any)
		if m["commodity"] != "petroleum" {
			t.Errorf("expected only petroleum cargo, got commodity=%v", m["commodity"])
		}
	}
}

func TestFacilitiesList(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/v1/facilities", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestBerthsFilterByTerminal(t *testing.T) {
	h := newTestHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/api/v1/berths?terminal_id=TRM-001", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	items := resp["items"].([]any)
	for _, item := range items {
		m := item.(map[string]any)
		if m["terminal_id"] != "TRM-001" {
			t.Errorf("expected only TRM-001 berths, got terminal_id=%v", m["terminal_id"])
		}
	}
}
