// Package handler implements the HTTP API handlers for the port information service.
package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

// New builds and returns the top-level HTTP mux with all routes registered.
// If spiffeMiddleware is non-nil it wraps all API routes except health.
func New(store *Store, logger *slog.Logger, spiffeMiddleware func(http.Handler) http.Handler) http.Handler {
	mux := http.NewServeMux()

	api := func(pattern string, h http.HandlerFunc) {
		var handler http.Handler = h
		if spiffeMiddleware != nil {
			handler = spiffeMiddleware(h)
		}
		mux.Handle(pattern, handler)
	}

	// Health (always unauthenticated)
	mux.HandleFunc("GET /api/v1/health/live", handleLive)
	mux.HandleFunc("GET /api/v1/health/ready", makeReadyHandler(store))

	// Vessel routes
	api("GET /api/v1/vessels", makeVesselsHandler(store, logger))
	api("GET /api/v1/vessels/{id}", makeVesselByIDHandler(store, logger))

	// Event routes
	api("GET /api/v1/arrivals", makeArrivalsHandler(store, logger))
	api("GET /api/v1/departures", makeDeparturesHandler(store, logger))

	// Barge routes
	api("GET /api/v1/barges", makeBargessHandler(store, logger))
	api("GET /api/v1/barges/{id}", makeBargeByIDHandler(store, logger))

	// Tow routes
	api("GET /api/v1/tows", makeTowsHandler(store, logger))
	api("GET /api/v1/tows/{id}", makeTowByIDHandler(store, logger))

	// Cargo routes
	api("GET /api/v1/cargo", makeCargoHandler(store, logger))
	api("GET /api/v1/cargo/{id}", makeCargoByIDHandler(store, logger))

	// Facility routes
	api("GET /api/v1/facilities", makeFacilitiesHandler(store, logger))
	api("GET /api/v1/facilities/{id}", makeFacilityByIDHandler(store, logger))

	// Terminal routes
	api("GET /api/v1/terminals", makeTerminalsHandler(store, logger))
	api("GET /api/v1/terminals/{id}", makeTerminalByIDHandler(store, logger))

	// Berth routes
	api("GET /api/v1/berths", makeBerthsHandler(store, logger))
	api("GET /api/v1/berths/{id}", makeBerthByIDHandler(store, logger))

	return requestLogger(logger, mux)
}

// --- Pagination helpers ---

type pagination struct {
	limit  int
	offset int
}

func parsePagination(r *http.Request) (pagination, error) {
	p := pagination{limit: defaultLimit, offset: 0}
	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return p, fmt.Errorf("limit must be a positive integer")
		}
		if n > maxLimit {
			n = maxLimit
		}
		p.limit = n
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return p, fmt.Errorf("offset must be a non-negative integer")
		}
		p.offset = n
	}
	return p, nil
}

func paginate[T any](items []T, p pagination) []T {
	if p.offset >= len(items) {
		return []T{}
	}
	end := p.offset + p.limit
	if end > len(items) {
		end = len(items)
	}
	return items[p.offset:end]
}

// --- Response helpers ---

type collectionResponse[T any] struct {
	Items  []T `json:"items"`
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
}

func writeProblem(w http.ResponseWriter, status int, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	body := fmt.Sprintf(`{"type":"about:blank","title":%q,"status":%d,"detail":%q}`,
		http.StatusText(status), status, detail)
	_, _ = w.Write([]byte(body))
}

// --- Logging middleware ---

func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote", r.RemoteAddr,
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
