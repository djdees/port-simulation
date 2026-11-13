// Package config handles client startup configuration validation.
package config

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"
)

// Config holds validated client configuration.
type Config struct {
	ServerURL         string
	SPIFFEEnabled     bool
	WorkloadSocket    string
	ServerSPIFFEID    string
	OutputDir         string
	CollectorInterval time.Duration
	SnapshotListen    string
	ReaderSPIFFEIDs   []string
}

// Load parses flags and environment variables, validates, and returns a Config.
func Load(args []string) (*Config, error) {
	fs := flag.NewFlagSet("client", flag.ContinueOnError)

	server := fs.String("server", "", "Port server base URL (overrides PORT_SERVER_URL)")
	spiffeFlag := fs.String("spiffe", "", "Enable SPIFFE/mTLS: true|false (overrides PORT_SPIFFE_ENABLED)")
	serverID := fs.String("server-id", "", "Expected server SPIFFE ID (overrides PORT_SERVER_SPIFFE_ID)")

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("parsing flags: %w", err)
	}

	cfg := &Config{}

	// Server URL
	if *server != "" {
		cfg.ServerURL = *server
	} else {
		cfg.ServerURL = os.Getenv("PORT_SERVER_URL")
	}

	// SPIFFE mode: default true
	cfg.SPIFFEEnabled = true
	if v := os.Getenv("PORT_SPIFFE_ENABLED"); v != "" {
		cfg.SPIFFEEnabled = parseBool(v)
	}
	if *spiffeFlag != "" {
		cfg.SPIFFEEnabled = parseBool(*spiffeFlag)
	}

	cfg.WorkloadSocket = os.Getenv("SPIFFE_ENDPOINT_SOCKET")

	if *serverID != "" {
		cfg.ServerSPIFFEID = *serverID
	} else {
		cfg.ServerSPIFFEID = os.Getenv("PORT_SERVER_SPIFFE_ID")
	}

	cfg.OutputDir = os.Getenv("PORT_COLLECTOR_OUTPUT_DIR")
	if cfg.OutputDir == "" {
		cfg.OutputDir = "collector-output"
	}

	cfg.CollectorInterval = 30 * time.Second
	if v := os.Getenv("PORT_COLLECTOR_INTERVAL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("invalid PORT_COLLECTOR_INTERVAL: %w", err)
		}
		cfg.CollectorInterval = d
	}

	cfg.SnapshotListen = os.Getenv("PORT_SNAPSHOT_LISTEN")
	if cfg.SnapshotListen == "" {
		cfg.SnapshotListen = "127.0.0.1:8081"
	}

	if v := os.Getenv("PORT_LOCAL_READER_SPIFFE_IDS"); v != "" {
		for _, id := range strings.Split(v, ",") {
			if id = strings.TrimSpace(id); id != "" {
				cfg.ReaderSPIFFEIDs = append(cfg.ReaderSPIFFEIDs, id)
			}
		}
	}

	return cfg, cfg.validate()
}

func (c *Config) validate() error {
	var errs []string

	if c.ServerURL == "" {
		errs = append(errs, "PORT_SERVER_URL (or --server) is required")
	} else {
		u, err := url.Parse(c.ServerURL)
		if err != nil {
			errs = append(errs, fmt.Sprintf("invalid server URL %q: %v", c.ServerURL, err))
		} else {
			if c.SPIFFEEnabled && u.Scheme != "https" {
				errs = append(errs, "SPIFFE mode requires an https:// server URL")
			}
			if !c.SPIFFEEnabled && u.Scheme == "https" {
				errs = append(errs, "SPIFFE-off mode must use http:// (not https://)")
			}
			if !c.SPIFFEEnabled {
				// Reject non-loopback targets in SPIFFE-off mode
				host := u.Hostname()
				if host != "localhost" && host != "127.0.0.1" && host != "::1" {
					ip := net.ParseIP(host)
					if ip == nil || !ip.IsLoopback() {
						errs = append(errs, "SPIFFE-off mode may only target loopback addresses")
					}
				}
			}
		}
	}

	if c.SPIFFEEnabled {
		if c.WorkloadSocket == "" {
			errs = append(errs, "SPIFFE_ENDPOINT_SOCKET is required when SPIFFE is enabled")
		}
		if c.ServerSPIFFEID == "" {
			errs = append(errs, "PORT_SERVER_SPIFFE_ID (or --server-id) is required when SPIFFE is enabled")
		}
	}

	// Snapshot listener must always be loopback
	host, _, err := net.SplitHostPort(c.SnapshotListen)
	if err == nil {
		ip := net.ParseIP(host)
		if host != "localhost" && (ip == nil || !ip.IsLoopback()) {
			errs = append(errs, "PORT_SNAPSHOT_LISTEN must be a loopback address")
		}
	}

	if len(errs) > 0 {
		return errors.New("client configuration errors:\n  " + strings.Join(errs, "\n  "))
	}
	return nil
}

func parseBool(s string) bool {
	return strings.EqualFold(s, "true") || s == "1"
}
