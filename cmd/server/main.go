// Command server starts the New Orleans Port Information Service HTTP API.
//
// All data served is fictional and does not represent real port operations.
// See README.md for configuration, SPIFFE/SPIRE requirements, and usage.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/djdees/port-simulation/internal/server/auth"
	"github.com/djdees/port-simulation/internal/server/config"
	"github.com/djdees/port-simulation/internal/server/data"
	"github.com/djdees/port-simulation/internal/server/handler"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	if err := run(os.Args[1:], logger); err != nil {
		logger.Error("server exited with error", "err", err)
		os.Exit(1)
	}
}

func run(args []string, logger *slog.Logger) error {
	// --- Configuration ---
	cfg, err := config.Load(args)
	if err != nil {
		return fmt.Errorf("configuration: %w", err)
	}

	// --- Embedded data ---
	portData, err := data.Load()
	if err != nil {
		return fmt.Errorf("loading dataset: %w", err)
	}
	logger.Info("dataset loaded",
		"vessels", len(portData.Vessels),
		"barges", len(portData.Barges),
		"tows", len(portData.Tows),
		"cargo", len(portData.Cargo),
		"facilities", len(portData.Facilities),
		"terminals", len(portData.Terminals),
		"berths", len(portData.Berths),
		"events", len(portData.Events),
	)

	store := handler.NewStore(portData)

	// --- SPIFFE / TLS ---
	var spiffeMiddlewareFn func(http.Handler) http.Handler
	var xsrc *auth.X509Source

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if cfg.SPIFFEEnabled {
		policy, err := auth.LoadPolicy(cfg.AuthPolicyPath)
		if err != nil {
			return fmt.Errorf("loading auth policy: %w", err)
		}
		xsrc, err = auth.NewX509Source(ctx, cfg.WorkloadSocket, policy, logger)
		if err != nil {
			return fmt.Errorf("SPIFFE source: %w", err)
		}
		defer xsrc.Close()
		spiffeMiddlewareFn = func(next http.Handler) http.Handler {
			return xsrc.Middleware(next, auth.PermPortRead)
		}
		logger.Info("SPIFFE/mTLS enabled", "workload_socket", cfg.WorkloadSocket)
	} else {
		logger.Warn("SPIFFE DISABLED — development-only mode; restricted to loopback, plain HTTP only")
	}

	// --- HTTP server ---
	mux := handler.New(store, logger, spiffeMiddlewareFn)

	srv := &http.Server{
		Addr:              cfg.Listen,
		Handler:           mux,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		MaxHeaderBytes:    64 << 10, // 64 KB
	}

	if cfg.SPIFFEEnabled && xsrc != nil {
		srv.TLSConfig = xsrc.ServerTLSConfig()
	}

	// Listen for shutdown signals.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		logger.Info("server listening", "addr", cfg.Listen, "spiffe", cfg.SPIFFEEnabled)
		if cfg.SPIFFEEnabled {
			// TLS config is set on srv; ListenAndServeTLS with empty strings uses srv.TLSConfig.
			errCh <- srv.ListenAndServeTLS("", "")
		} else {
			errCh <- srv.ListenAndServe()
		}
	}()

	select {
	case sig := <-sigCh:
		logger.Info("shutdown signal received", "signal", sig.String())
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server error: %w", err)
		}
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	logger.Info("server stopped cleanly")
	return nil
}
