// Package client provides HTTP transport and API access for the port server.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"

	clientconfig "github.com/djdees/port-simulation/internal/client/config"
)

// Client is an authenticated HTTP client for the port API.
type Client struct {
	httpClient *http.Client
	baseURL    string
	xsrc       *workloadapi.X509Source // nil in SPIFFE-off mode
}

// New creates a Client from the provided config. If SPIFFE is enabled, it
// connects to the Workload API and builds an mTLS transport. The returned
// Client should be closed when no longer needed.
func New(ctx context.Context, cfg *clientconfig.Config) (*Client, error) {
	c := &Client{baseURL: cfg.ServerURL}

	if cfg.SPIFFEEnabled {
		src, err := workloadapi.NewX509Source(ctx, workloadapi.WithClientOptions(
			workloadapi.WithAddr(cfg.WorkloadSocket),
		))
		if err != nil {
			return nil, fmt.Errorf("connecting to SPIFFE Workload API: %w", err)
		}
		c.xsrc = src

		serverID, err := spiffeid.FromString(cfg.ServerSPIFFEID)
		if err != nil {
			return nil, fmt.Errorf("invalid server SPIFFE ID %q: %w", cfg.ServerSPIFFEID, err)
		}

		tlsCfg := tlsconfig.MTLSClientConfig(src, src, tlsconfig.AuthorizeID(serverID))
		c.httpClient = &http.Client{
			Transport: &http.Transport{TLSClientConfig: tlsCfg},
			Timeout:   30 * time.Second,
		}
	} else {
		c.httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	return c, nil
}

// Close releases SPIFFE Workload API resources.
func (c *Client) Close() error {
	if c.xsrc != nil {
		return c.xsrc.Close()
	}
	return nil
}

// Get issues an authenticated GET to path and decodes the JSON response into v.
func (c *Client) Get(ctx context.Context, path string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var prob map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&prob)
		detail, _ := prob["detail"].(string)
		if detail == "" {
			detail = resp.Status
		}
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, detail)
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}
	return nil
}

// NewInsecure creates a plain-HTTP client for unit testing only.
// SPIFFE must be disabled in cfg. This function must not be used for
// production or remote access.
func NewInsecure(cfg *clientconfig.Config) (*Client, error) {
	if cfg.SPIFFEEnabled {
		return nil, fmt.Errorf("NewInsecure requires SPIFFEEnabled=false")
	}
	return &Client{
		baseURL:    cfg.ServerURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}
