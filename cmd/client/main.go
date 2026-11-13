// Command client is the New Orleans Port Information Service CLI and collector.
//
// All data served by the port server is fictional.
// See README.md for configuration, SPIFFE/SPIRE requirements, and usage.
//
// Usage:
//
//	client query arrivals           - list arrival events
//	client query departures         - list departure events
//	client query vessels            - list vessels
//	client get-vessel <id>          - retrieve a vessel by ID
//	client filter-cargo [params]    - filter cargo (e.g. "commodity=petroleum")
//	client collect                  - start the periodic collector
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	portclient "github.com/djdees/port-simulation/internal/client"
	clientconfig "github.com/djdees/port-simulation/internal/client/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if err := run(os.Args[1:], logger); err != nil {
		logger.Error("client error", "err", err)
		os.Exit(1)
	}
}

func run(args []string, logger *slog.Logger) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	cfg, err := clientconfig.Load(args)
	if err != nil {
		return fmt.Errorf("configuration: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	c, err := portclient.New(ctx, cfg)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}
	defer c.Close()

	cmd := args[0]
	cmdArgs := args[1:]

	switch cmd {
	case "query":
		if len(cmdArgs) == 0 {
			return fmt.Errorf("query requires a subcommand: arrivals | departures | vessels")
		}
		switch cmdArgs[0] {
		case "arrivals":
			return portclient.QueryArrivals(ctx, c, os.Stdout)
		case "departures":
			return portclient.QueryDepartures(ctx, c, os.Stdout)
		case "vessels":
			return portclient.ListVessels(ctx, c, os.Stdout)
		default:
			return fmt.Errorf("unknown query subcommand %q", cmdArgs[0])
		}

	case "list-arrivals":
		return portclient.QueryArrivals(ctx, c, os.Stdout)

	case "get-vessel":
		if len(cmdArgs) == 0 {
			return fmt.Errorf("get-vessel requires a vessel ID")
		}
		return portclient.GetVessel(ctx, c, os.Stdout, cmdArgs[0])

	case "filter-cargo":
		params := ""
		if len(cmdArgs) > 0 {
			params = strings.Join(cmdArgs, "&")
		}
		return portclient.FilterCargo(ctx, c, os.Stdout, params)

	case "collect":
		col := portclient.NewCollector(c, cfg.OutputDir, cfg.CollectorInterval, logger)
		col.Run(ctx)
		return nil

	default:
		return fmt.Errorf("unknown command %q; run without arguments for usage", cmd)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `Port Information Service Client

Commands:
  query arrivals              List recent arrivals
  query departures            List recent departures
  query vessels               List vessels
  get-vessel <id>             Get vessel by ID (e.g. VSL-001)
  filter-cargo [param=value]  Filter cargo records
  collect                     Start periodic collector

Configuration (env vars):
  PORT_SERVER_URL             Server base URL (required)
  PORT_SPIFFE_ENABLED         true|false (default: true)
  SPIFFE_ENDPOINT_SOCKET      Workload API socket
  PORT_SERVER_SPIFFE_ID       Expected server SPIFFE ID
  PORT_COLLECTOR_OUTPUT_DIR   CSV output directory (default: collector-output)
  PORT_COLLECTOR_INTERVAL     Collection interval (default: 30s)
  PORT_SNAPSHOT_LISTEN        Snapshot endpoint listen addr (default: 127.0.0.1:8081)
  PORT_LOCAL_READER_SPIFFE_IDS Comma-separated reader SPIFFE IDs

Flags (override env):
  --server <url>              Server URL
  --spiffe <true|false>       SPIFFE mode
  --server-id <spiffe://...>  Server SPIFFE ID`)
}
