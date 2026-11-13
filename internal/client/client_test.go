package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/djdees/port-simulation/internal/client"
	clientconfig "github.com/djdees/port-simulation/internal/client/config"
)

// newTestClient creates a plain-HTTP (SPIFFE-off) client targeting the given test server URL.
func newTestClient(t *testing.T, serverURL string) *client.Client {
	t.Helper()
	cfg := &clientconfig.Config{
		ServerURL:     serverURL,
		SPIFFEEnabled: false,
		OutputDir:     t.TempDir(),
	}
	c, err := client.NewInsecure(cfg)
	if err != nil {
		t.Fatalf("creating test client: %v", err)
	}
	return c
}

func TestGetDecodeSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[],"total":0,"limit":20,"offset":0}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	defer c.Close()

	var resp map[string]any
	if err := c.Get(context.Background(), "/api/v1/vessels", &resp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := resp["total"]; !ok {
		t.Error("expected total field in response")
	}
}

func TestGetHandlesServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"detail":"vessel not found"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	defer c.Close()

	var resp any
	err := c.Get(context.Background(), "/api/v1/vessels/VSL-999", &resp)
	if err == nil {
		t.Fatal("expected error for 404 response, got nil")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("expected error to mention 404, got: %v", err)
	}
}

func TestGetHandlesMalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`NOT VALID JSON`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	defer c.Close()

	var resp map[string]any
	err := c.Get(context.Background(), "/api/v1/vessels", &resp)
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}

func TestQueryArrivals(t *testing.T) {
	payload := map[string]any{
		"items":  []any{map[string]any{"id": "EVT-001", "event_type": "arrival"}},
		"total":  float64(1),
		"limit":  float64(20),
		"offset": float64(0),
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/arrivals" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	defer c.Close()

	var sb strings.Builder
	if err := client.QueryArrivals(context.Background(), c, &sb); err != nil {
		t.Fatalf("QueryArrivals failed: %v", err)
	}
	if !strings.Contains(sb.String(), "EVT-001") {
		t.Errorf("expected EVT-001 in output, got: %s", sb.String())
	}
}

func TestCollectorSnapshotNilBeforeCollection(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []any{}, "total": 0, "limit": 20, "offset": 0,
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	defer c.Close()

	col := client.NewCollector(c, t.TempDir(), 0, nil)
	if col.GetSnapshot("arrivals") != nil {
		t.Error("expected nil snapshot before first collection")
	}
}
