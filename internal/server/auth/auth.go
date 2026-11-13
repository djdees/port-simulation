// Package auth provides SPIFFE/X.509 workload identity, mTLS configuration,
// and authorization policy enforcement for the port server.
//
// SPIFFE-off mode is a development-only option. It logs a warning, restricts
// the server to loopback addresses, and uses plain HTTP. It must not be used
// for remote access or production deployment.
package auth

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

// Permission is an operation a client may be authorized to perform.
type Permission string

const (
	PermPortRead Permission = "port:read"
)

// Policy maps exact client SPIFFE IDs to sets of permissions.
type Policy struct {
	Clients []ClientPolicy `json:"clients"`
}

// ClientPolicy grants a set of permissions to one SPIFFE ID.
type ClientPolicy struct {
	SPIFFEID    string       `json:"spiffe_id"`
	Permissions []Permission `json:"permissions"`
}

// LoadPolicy reads and parses a JSON authorization policy from path.
func LoadPolicy(path string) (*Policy, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading auth policy %q: %w", path, err)
	}
	var p Policy
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("parsing auth policy %q: %w", path, err)
	}
	return &p, nil
}

// HasPermission reports whether the policy grants perm to the given SPIFFE ID string.
func (p *Policy) HasPermission(id string, perm Permission) bool {
	for _, c := range p.Clients {
		if c.SPIFFEID == id {
			for _, pp := range c.Permissions {
				if pp == perm {
					return true
				}
			}
		}
	}
	return false
}

// X509Source wraps the go-spiffe workload API X.509 source.
// It must remain open for the lifetime of the server to support credential rotation.
type X509Source struct {
	source *workloadapi.X509Source
	policy *Policy
	logger *slog.Logger
}

// NewX509Source creates and starts a live X.509 source from the Workload API.
// socketPath is the SPIFFE Workload API endpoint (e.g. "unix:///tmp/spire-agent.sock").
// The returned source must be closed when the server shuts down.
func NewX509Source(ctx context.Context, socketPath string, policy *Policy, logger *slog.Logger) (*X509Source, error) {
	src, err := workloadapi.NewX509Source(ctx, workloadapi.WithClientOptions(
		workloadapi.WithAddr(socketPath),
	))
	if err != nil {
		return nil, fmt.Errorf("connecting to SPIFFE Workload API at %q: %w", socketPath, err)
	}
	return &X509Source{source: src, policy: policy, logger: logger}, nil
}

// Close releases the Workload API connection.
func (x *X509Source) Close() error {
	return x.source.Close()
}

// ServerTLSConfig returns a *tls.Config that presents the server's X.509-SVID
// and requires clients to present valid SVIDs from the same trust domain.
// Accepted client IDs are those listed in the policy.
func (x *X509Source) ServerTLSConfig() *tls.Config {
	// Build set of authorised client IDs for the TLS layer.
	var allowedIDs []spiffeid.ID
	for _, c := range x.policy.Clients {
		id, err := spiffeid.FromString(c.SPIFFEID)
		if err != nil {
			x.logger.Warn("skipping invalid SPIFFE ID in policy", "spiffe_id", c.SPIFFEID, "err", err)
			continue
		}
		allowedIDs = append(allowedIDs, id)
	}

	authorizer := tlsconfig.AuthorizeOneOf(allowedIDs...)
	return tlsconfig.MTLSServerConfig(x.source, x.source, authorizer)
}

// Middleware returns an http.Handler that enforces that the peer presented a
// valid SPIFFE identity and that the policy grants it perm.
// It must wrap handlers that are protected by SPIFFE authorization.
func (x *X509Source) Middleware(next http.Handler, perm Permission) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
			x.logger.Warn("request missing peer certificate", "remote", r.RemoteAddr)
			writeProblem(w, http.StatusUnauthorized, "unauthorized", "client identity required")
			return
		}

		// Extract SPIFFE ID URI from the peer certificate.
		peerCert := r.TLS.PeerCertificates[0]
		var peerID string
		for _, uri := range peerCert.URIs {
			if uri.Scheme == "spiffe" {
				peerID = uri.String()
				break
			}
		}
		if peerID == "" {
			x.logger.Warn("peer certificate has no SPIFFE URI SAN", "remote", r.RemoteAddr)
			writeProblem(w, http.StatusForbidden, "forbidden", "no SPIFFE identity in client certificate")
			return
		}

		if !x.policy.HasPermission(peerID, perm) {
			x.logger.Warn("authorization denied", "peer_id", peerID, "required_permission", perm)
			writeProblem(w, http.StatusForbidden, "forbidden", "client not authorized for this operation")
			return
		}

		x.logger.Info("request authorized", "peer_id", peerID, "permission", perm, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// writeProblem writes an application/problem+json response.
func writeProblem(w http.ResponseWriter, status int, errCode, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	body := fmt.Sprintf(`{"type":"about:blank","title":%q,"status":%d,"detail":%q}`,
		http.StatusText(status), status, detail)
	_, _ = w.Write([]byte(body))
}
