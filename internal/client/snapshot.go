// Package client - local snapshot HTTP endpoint (GET /api/v1/latest/{query}).
//
// This endpoint is served by the client process, not the port server.
// It is strictly loopback-only and requires mTLS when SPIFFE is enabled.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

// SnapshotServer serves the local latest-snapshot endpoint.
type SnapshotServer struct {
	collector *Collector
	listen    string
	logger    *slog.Logger
	readerIDs []string
	xsrc      *workloadapi.X509Source
}

// NewSnapshotServer creates a SnapshotServer. If xsrc is non-nil, the server
// requires mTLS and enforces the reader SPIFFE ID allowlist.
func NewSnapshotServer(
	collector *Collector,
	listen string,
	readerIDs []string,
	xsrc *workloadapi.X509Source,
	logger *slog.Logger,
) *SnapshotServer {
	return &SnapshotServer{
		collector: collector,
		listen:    listen,
		logger:    logger,
		readerIDs: readerIDs,
		xsrc:      xsrc,
	}
}

// Serve starts the snapshot HTTP server and blocks until ctx is cancelled.
func (s *SnapshotServer) Serve(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/latest/{query}", s.handleLatest)

	srv := &http.Server{
		Addr:              s.listen,
		Handler:           mux,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if s.xsrc != nil {
		// Build set of allowed reader IDs.
		var allowedIDs []spiffeid.ID
		for _, raw := range s.readerIDs {
			id, err := spiffeid.FromString(raw)
			if err != nil {
				s.logger.Warn("skipping invalid reader SPIFFE ID", "id", raw, "err", err)
				continue
			}
			allowedIDs = append(allowedIDs, id)
		}
		srv.TLSConfig = tlsconfig.MTLSServerConfig(s.xsrc, s.xsrc, tlsconfig.AuthorizeOneOf(allowedIDs...))
	}

	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("snapshot server listening", "addr", s.listen)
		if s.xsrc != nil {
			errCh <- srv.ListenAndServeTLS("", "")
		} else {
			errCh <- srv.ListenAndServe()
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			return err
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func (s *SnapshotServer) handleLatest(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimPrefix(r.PathValue("query"), "/")

	// If SPIFFE enabled, verify peer certificate has an allowed reader SPIFFE ID.
	if s.xsrc != nil {
		if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
			writeSnapshotProblem(w, http.StatusUnauthorized, "client identity required")
			return
		}
		peerCert := r.TLS.PeerCertificates[0]
		var peerID string
		for _, uri := range peerCert.URIs {
			if uri.Scheme == "spiffe" {
				peerID = uri.String()
				break
			}
		}
		allowed := false
		for _, id := range s.readerIDs {
			if id == peerID {
				allowed = true
				break
			}
		}
		if !allowed {
			s.logger.Warn("snapshot access denied", "peer_id", peerID)
			writeSnapshotProblem(w, http.StatusForbidden, "not in reader allowlist")
			return
		}
	}

	snap := s.collector.GetSnapshot(query)
	if snap == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(snap)
}

func writeSnapshotProblem(w http.ResponseWriter, status int, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	body := fmt.Sprintf(`{"type":"about:blank","title":%q,"status":%d,"detail":%q}`,
		http.StatusText(status), status, detail)
	_, _ = w.Write([]byte(body))
}
