// Package config handles server startup configuration validation.
package config

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

// Config holds validated server configuration.
type Config struct {
	Listen         string // e.g. "127.0.0.1:8443"
	SPIFFEEnabled  bool
	WorkloadSocket string // SPIFFE_ENDPOINT_SOCKET
	AuthPolicyPath string // path to authorization-policy.json
}

// Load parses flags and environment, validates, and returns a Config.
// Explicit flags take precedence over environment variables.
func Load(args []string) (*Config, error) {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)

	listen := fs.String("listen", "", "Listen address (overrides PORT_SERVER_LISTEN)")
	spiffeFlag := fs.String("spiffe", "", "Enable SPIFFE/mTLS: true|false (overrides PORT_SPIFFE_ENABLED)")
	policy := fs.String("policy", "", "Path to authorization policy JSON (overrides PORT_AUTH_POLICY)")

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("parsing flags: %w", err)
	}

	cfg := &Config{}

	// Listen address
	if *listen != "" {
		cfg.Listen = *listen
	} else if v := os.Getenv("PORT_SERVER_LISTEN"); v != "" {
		cfg.Listen = v
	} else {
		cfg.Listen = "127.0.0.1:8443"
	}

	// SPIFFE mode: default true; flag wins over env
	cfg.SPIFFEEnabled = true
	if v := os.Getenv("PORT_SPIFFE_ENABLED"); v != "" {
		cfg.SPIFFEEnabled = parseBool(v)
	}
	if *spiffeFlag != "" {
		cfg.SPIFFEEnabled = parseBool(*spiffeFlag)
	}

	// Workload API socket
	cfg.WorkloadSocket = os.Getenv("SPIFFE_ENDPOINT_SOCKET")

	// Auth policy
	if *policy != "" {
		cfg.AuthPolicyPath = *policy
	} else {
		cfg.AuthPolicyPath = os.Getenv("PORT_AUTH_POLICY")
	}

	return cfg, cfg.validate()
}

func (c *Config) validate() error {
	var errs []string

	// Validate listen address
	host, port, err := net.SplitHostPort(c.Listen)
	if err != nil {
		errs = append(errs, fmt.Sprintf("invalid listen address %q: %v", c.Listen, err))
	} else if port == "" {
		errs = append(errs, "listen address must include a port")
	} else if c.isNonLoopback(host) && !c.SPIFFEEnabled {
		errs = append(errs, "SPIFFE-off mode requires a loopback listen address (127.x.x.x or ::1)")
	}

	// SPIFFE-on requirements
	if c.SPIFFEEnabled {
		if c.WorkloadSocket == "" {
			errs = append(errs, "SPIFFE_ENDPOINT_SOCKET is required when SPIFFE is enabled")
		}
		if c.AuthPolicyPath == "" {
			errs = append(errs, "PORT_AUTH_POLICY (or --policy) is required when SPIFFE is enabled")
		}
	}

	if len(errs) > 0 {
		return errors.New("server configuration errors:\n  " + strings.Join(errs, "\n  "))
	}
	return nil
}

func (c *Config) isNonLoopback(host string) bool {
	if host == "localhost" {
		return false
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return true // unknown — treat as non-loopback
	}
	return !ip.IsLoopback()
}

func parseBool(s string) bool {
	return strings.EqualFold(s, "true") || s == "1"
}
